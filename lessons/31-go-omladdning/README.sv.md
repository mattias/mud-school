# Lektion 31 – Ladda om världen

Fas 5 börjar. Målet: **servern ska aldrig behöva startas om** för att få nytt
innehåll eller nya funktioner. Spelarna är kvar medan världen förändras
runt dem. Det finns fyra nivåer, från den enklaste till den mest kraftfulla:

```
 data (JSON, laddas om live)  ->  skript (Lua)  ->  plugins (WASM)  ->  ett nytt program (copyover)
   lektion 31, 36                  lektion 32, 34    lektion 33         lektion 35
```

Den här lektionen är den första: ändra i `world.json` – rätta ett stavfel,
lägg till ett rum – spara, och en sekund senare skimrar världen till och alla
står i den nya.

## Mål

- Bevaka en fil utan bibliotek: **pollning** med `os.Stat` och en `time.Ticker`.
- Göra långsamt arbete (läsa och validera en fil) **utanför** spelloopen och
  bytet **inne i** den – atomiskt, så att ingen någonsin ser en halv värld.
- Flytta levande tillstånd från en gammal struktur till en ny (spelare, positioner, saker).
- Skydda kommandon med en enkel behörighetskontroll: **administratörer**.

## Go-koncept

### Pollning

```go
t := time.NewTicker(s.reloadEvery)
defer t.Stop()
for {
	select {
	case <-s.stop:
		return
	case <-t.C:
		fi, err := os.Stat(path)  // har ModTime eller Size ändrats?
	}
}
```

Bibliotek som `fsnotify` får besked av operativsystemet i stället för att
fråga, men att polla var eller varannan sekund är enkelt, portabelt och
billigt nog för en fil. Lägg märke till `select` på `s.stop`: goroutinen
slutar med servern – inga läckor.

### Validera först, byt sedan

```go
nw, err := LoadWorldFile(path)   // utanför loopen: disk + validering (lektion 9)
s.do(func() {                    // inne i loopen: ett snabbt byte
	if err != nil { /* säg till administratörerna, behåll den gamla världen */ return }
	s.swapWorld(nw)
})
```

Är filen trasig ändras *ingenting* – administratörerna får felmeddelandena
från lektion 9 och kan rätta filen.

### Flytta tillstånd

Den nya världen har nya värden av `*Room`, `*Item` och `*NPC`. Allt som
pekade in i den gamla världen måste peka in i den nya:

- spelare: rummet med samma ID (eller startrummet om det är borta), en
  position som fortfarande är gångbar;
- deras saker: den nya världens sak med samma ID, tagen ur sitt rum;
- strider: avbryts – varelserna är nya;
- zonen: `seenRoom` och `explored` glöms, så att klienterna får en ny zone-frame.

### Kommandon med `@`

Administratörernas kommandon börjar med `@`. Ett omslag kontrollerar
behörigheten så att kommandona själva slipper:

```go
func adminOnly(fn CommandFunc) CommandFunc {
	return func(s *Server, p *Player, cmd Command) {
		if !s.isAdmin(p) { p.SendKind(KindError, MsgNotAdmin); return }
		fn(s, p, cmd)
	}
}
```

En funktion som tar en funktion och returnerar en funktion – samma idé som
`moveCommand` i lektion 8.

## Kontraktet

- `Options.WorldPath` (världsfilen, `""` = ingen), `Options.ReloadEvery`
  (0 = bevaka inte), `Options.Admins` (namn, normaliserade som spelarnamn).
- När filen ändras (ModTime eller Size) läses och valideras den utanför spelloopen:
  - giltig: världen byts; varje spelare får `MsgReloaded` och en `look`; en
    spelare vars rum är borta får `MsgRoomGone` och hamnar i startrummet; saker
    man bär byts mot den nya världens saker med samma ID (som inte längre ligger
    i sina rum); strider avbryts;
  - ogiltig: ingenting ändras; inloggade administratörer får `MsgReloadFailed` med felet.
- `@reload` (bara administratörer, annars `MsgNotAdmin`; `MsgNoWorldFile` utan
  världsfil) gör samma sak direkt.
- `main.go`: `-reload 2s` och `-admin Anna,Bertil`.

## Uppgift

`reload.go` är delvis given (`isAdmin`, `adminOnly`, `notifyAdmins`,
`cmdReload`, `normalizeAdmins`): skriv `reloadFromFile`, `watchWorld` och
`swapWorld`. Sedan alternativen, fälten på `Server`, att starta bevakningen i
`start`, kommandot och flaggorna.

Prova: kopiera världen (`cp server/mud/world.json min-värld.json`), starta
servern med `-world min-värld.json -admin DittNamn`, logga in och ändra en
beskrivning i filen.

## Fundera på

- Varför måste bytet ske i spelloopen, men läsningen utanför?
- En spelare står på en ruta som är en vägg i den nya zonen. Vad händer? Vad *borde* hända?
- Vad händer om filen sparas mitt i pollningen – halvskriven? (Testerna skriver
  en temporär fil och byter namn på den. Varför hjälper det?)
- Två omladdningar tätt efter varandra – kan de tillämpas i fel ordning?
- En administratör är bara ett *namn*. Utan databas (`-db ""`), eller innan
  Anna har registrerat sig, kan vem som helst logga in som Anna – och lektion
  35–36 ger administratörer `@copyover` och rätten att skriva världsfilen. Kör
  med lösenord (lektion 20) och registrera administratörsnamnen först. Vad mer skulle du vilja ha?

## Utmaningar

- Ladda om `items.json` och `npcs.json` var för sig, så att varelser som håller
  på med något kan fortsätta.
- `@reload` visar en diff: vilka rum som lagts till, tagits bort och ändrats.
- Använd `fsnotify` i stället för pollning och jämför.
