# Lektion 36 – Bygga i spelet

Sista lektionen. Sedan lektion 31 kan du ändra världen genom att redigera en
JSON-fil. MUD:ar har länge haft något bättre: **online creation**. Byggare går
omkring inne i världen och formar den därifrån:

```
> @dig north Mörk grotta
Du gräver åt norr och skapar "Mörk grotta" (m-rk-grotta).
> go north
Mörk grotta
Ett tomt rum. Beskriv det med @desc.
> @desc Vatten droppar från taket. Något glittrar i mörkret.
Klart.
```

Varje ändring **valideras** med reglerna från lektion 9, **sparas** i
världsfilen och **syns för alla direkt** – ingen omstart, ingen omladdning,
ingen editor.

## Mål

- Ändra levande tillstånd säkert: genomför, validera och **ångra** om resultatet är ogiltigt.
- Skriva en fil **atomiskt**: en temporär fil och ett namnbyte, så att en krasch
  eller en läsare aldrig ser en halv fil.
- Hålla långsam I/O utanför spelloopen – och hålla samtidiga sparningar i ordning.
- Känna igen dina egna skrivningar, så att filbevakningen från lektion 31 inte laddar om dem.

## Go-koncept

### Genomför, validera, ångra

```go
func (s *Server) change(fn, undo func()) error {
	fn()
	if err := s.world.Validate(); err != nil {
		undo()
		return err
	}
	return nil
}
```

Två closures: en som gör ändringen och en som tar tillbaka den. Världens
egen validering (utgångar som stämmer med koordinaterna, inga två rum på samma
plats, …) bestämmer – byggkommandona behöver inga egna regler.

### Atomiska skrivningar

```go
tmp, _ := os.CreateTemp(filepath.Dir(path), ".world.json.*")
tmp.Write(data)
tmp.Close()
os.Rename(tmp.Name(), path) // atomiskt i samma filsystem
```

Den som öppnar filen ser antingen den gamla versionen eller den nya, aldrig
något mitt emellan.

### Sparningar i ordning

JSON:en görs **inne i** spelloopen (en konsekvent ögonblicksbild); disken
skrivs **utanför** den, i en goroutine. Två snabba ändringar ger två
goroutiner – och den äldre ögonblicksbilden skulle kunna skrivas sist. Ett
löpnummer fixar det: varje ögonblicksbild får nästa nummer, och en goroutine
som märker att en nyare redan är skriven ger upp.

### Inte ladda om sig själv

Bevakningen från lektion 31 ser varje ändring av filen – också våra. Därför
minns servern filens ändringstid och storlek efter skrivningen (`ownWrite`),
och bevakningen hoppar över en ändring som stämmer med den. En mutex
(`saveMu`) ser till att bevakningen aldrig tittar mitt i en sparning.

## Kontraktet

Allt bara för administratörer (annars `MsgNotAdmin`), i rummet du står i:

- `@dig <riktning> <namn>` – ett nytt rum ett steg åt det hållet
  (koordinaterna från lektion 9), med samma terräng, beskrivningen
  `MsgNewRoomDesc` och utgångar åt båda hållen; ID:t görs av namnet
  (`"Dark cave"` → `dark-cave`, unikt). `MsgDug`. Finns det redan en utgång
  åt det hållet: `MsgExitExists`; utan namn: `MsgDigUsage`.
- `@desc <text>` och `@name <text>` – rummets beskrivning och namn; `MsgBuilt`.
- `@exit <riktning> <rums-id>` – en utgång (enkelriktad); `@exit <riktning> none`
  tar bort den; `MsgBuilt`.
- En ändring som skulle göra världen ogiltig ångras; byggaren får
  `MsgBuildInvalid` med valideringsfelen.
- Rum vars utgångar ändras får en ny zon; allt i dem behåller en gångbar position.
- Med en världsfil (`Options.WorldPath`) sparas hela världen i den efter
  varje ändring – atomiskt, i ordning, utanför spelloopen; bevakningen laddar
  inte om serverns egna sparningar. Utan världsfil: `MsgNotSaved`.

## Uppgift

`olc.go` är delvis given (`slug`, `newRoomID`, `change`, `rezone`,
`cmdDesc`, `cmdName`, `worldSnapshot`, `writeFileAtomic`): skriv `cmdDig`,
`cmdExit` och `saveWorld`. Sedan fälten på `Server`, kommandona och
kontrollen i `watchWorld`.

Bygg ett litet eget område: starta servern med `-world min-värld.json -admin
DittNamn`, gräv, beskriv och titta i filen efteråt.

(ID:n görs av latinska bokstäver och siffror – "Mörk grotta" blir
`m-rk-grotta`. Det fungerar, men en utmaning är att göra det snyggare.)

## Fundera på

- Varför görs JSON:en i spelloopen men skrivs utanför?
- `@dig` använder världens validering i stället för att kontrollera
  koordinaterna själv. Vilka är fördelarna? Finns det nackdelar?
- Vad händer med den procedurgenererade zonen i ett rum när du lägger till en utgång?
- Två administratörer bygger samtidigt. Vad kan gå fel? Vad kan inte gå fel?

## Utmaningar

- `@undo`: spara de tio senaste ändringarna (ångra-closures) och ta tillbaka dem.
- `@item` och `@npc`: skapa saker och varelser inifrån spelet.
- Byggare: en roll mellan spelare och administratörer som bara får bygga i sitt eget område.
- `@zone`: rita en handgjord zon i spelet, rad för rad.
- Snyggare ID:n: `å` → `a`, `ö` → `o` (paketet `golang.org/x/text/unicode/norm`).

**Det var sista lektionen.** Du har byggt en spelserver för många spelare i
Go och en klient i Rust – från en ekoserver till en tile-värld med ljus, mus,
sparande, skript, plugins i ett annat språk och en server som kan ändras på
alla sätt utan att någon kopplas ner. Fas 6 i `docs/*/CURRICULUM.md` har idéer
om vart du kan gå härnäst.
