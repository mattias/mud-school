# Lektion 04 – Kommandon

En ekoserver är ingen MUD. Nu ska servern förstå vad spelaren skriver:
`look`, `north`, `get sword`, `say hej`. Vi börjar med en **tolk** och ett
**kommandoregister** – och registret är frö till något stort: i fas 5 ska
skript och plugins kunna lägga till kommandon medan servern kör.

## Mål

- Arbeta med strängar (`strings.Fields`, `strings.Cut`, `strings.ToLower`).
- Använda **maps** och **slices**.
- Förstå att funktioner är värden: `map[string]CommandFunc`.
- Skriva och läsa **tabelldrivna tester** – Go-världens favoritmönster.

## Nya filer

Kursverktyget har skapat tre filer åt dig i `workspace/server/mud/`:

| Fil | Innehåll |
|---|---|
| `command.go` | `Command`, `ParseCommand` och förkortningar – **du fyller i** |
| `commands.go` | kommandoregistret och de inbyggda kommandona – **du fyller i** |
| `player.go` | `Player`: en ansluten spelare med `Send`/`Sendf` – färdig, läs igenom den |

## Go-koncept

### Maps

```go
ages := map[string]int{"anna": 31}
ages["bertil"] = 40
age, ok := ages["cecilia"]   // ok == false om nyckeln saknas, age blir 0
for name, age := range ages { ... }  // ordningen är SLUMPMÄSSIG!
```

### Slices

```go
words := strings.Fields("  get  the sword ")  // []string{"get", "the", "sword"}
first, rest := words[0], words[1:]
names := make([]string, 0, 10)                // längd 0, kapacitet 10
names = append(names, "anna")
sort.Strings(names)
```

### Funktioner som värden

```go
type CommandFunc func(s *Server, p *Player, cmd Command)

s.commands["dance"] = func(s *Server, p *Player, cmd Command) {
    p.Send("Du dansar!")
}
```

### Tabelldrivna tester

Öppna `mudschool_04_test.go` och titta på `TestParseCommand`. Varje rad i
tabellen är ett testfall. Vill du testa något nytt lägger du bara till en
rad. Du kan köra ett enskilt test direkt med Go:

```
cd workspace/server && go test ./mud -run TestParseCommand -v
```

## Uppgift

1. **`command.go`**: fyll i `aliases` och skriv `ParseCommand`:
   - ta bort mellanslag i början och slutet; tom rad ger `Command{}`,
   - första ordet är verbet, med små bokstäver, och förkortningar expanderas,
   - `Args` är resten av orden, `Rest` är resten av raden (trimmad),
   - en rad som börjar med `'` betyder `say`: `'hej` → `say` med `Rest` `"hej"`.
2. **`server.go`**: ge `Server` ett fält `commands map[string]CommandFunc`.
   Skapa mappen i `NewServer` och anropa `s.registerBuiltins()`.
   Skriv om `handle`: skapa en `Player` med `newPlayer(conn)`, skicka
   `Welcome`, och läs sedan rader och anropa `s.dispatch(p, line)` tills
   spelaren kopplar ner eller `p.quit` blir sann.
3. **`commands.go`**: skriv `dispatch`, `cmdHelp` och `cmdQuit`:
   - `help` skriver `MsgCommands` med alla registrerade verb, sorterade och
     åtskilda med `", "` – t.ex. `Kommandon: echo, help, quit`,
   - `quit` skriver `MsgGoodbye` (`Hej då!`) och avslutar,
   - okänt verb ger `MsgUnknownCommand`:
     `p.Sendf(MsgUnknownCommand, cmd.Verb)` → `Okänt kommando "flyg". Skriv help för hjälp.`
     (`%q` i formatsträngen lägger till citattecknen).

## Testa själv

```
telnet localhost 4000
help
echo   hej  på  dig
L
q
```

## Fundera på

- Varför sorterar vi verben i `help`? (Kör utan sortering några gånger.)
- `Server.commands` läses av många goroutines samtidigt. Är det säkert?
  (Ja – så länge ingen *skriver* till mappen efter att servern startat.
  Kom ihåg den här frågan till fas 5, då vi vill lägga till kommandon i
  en körande server.)

## Utmaningar

- Låt `help <kommando>` visa en beskrivning. Du kan byta `CommandFunc` mot
  en struct: `type commandDef struct { fn CommandFunc; help string }`.
- Stöd prefix: `hel` ska matcha `help` om inget annat kommando börjar så.
