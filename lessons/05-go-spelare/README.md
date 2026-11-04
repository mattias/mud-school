# Lektion 05 – Spelare

Dags att veta *vem* som spelar. Spelaren väljer ett namn när hen ansluter,
och servern håller reda på alla inloggade. Det är första gången flera
goroutines delar på samma data – och därmed första gången du kan skriva en
riktigt lömsk bugg: en **data race**.

## Mål

- Skapa och jämföra fel med `errors.New`, `errors.Is` och `%w`.
- Förstå skillnaden mellan *bytes* och *runor* (Unicode-tecken) i Go-strängar.
- Skydda delad data med `sync.Mutex`.
- Se race detectorn hitta en bugg.

## Inloggningsflödet

```
Välkommen till MUD School!
Vad heter du?
> x
Ogiltigt namn: namnet måste vara minst 2 tecken.
Vad heter du?
> anna
Välkommen, Anna!
> who
Spelare online (2): Anna, Bertil
```

## Go-koncept

### Fel är värden – även dina egna

```go
var ErrNameTooShort = errors.New(MsgErrNameTooShort) // texten finns i messages.go

func ValidateName(name string) error {
    if len(name) < 2 { return ErrNameTooShort }  // (fast len räknar bytes, se nedan!)
    return nil
}

err := fmt.Errorf("%w: %s", ErrNameTaken, name) // "wrappa" ett fel med mer info
errors.Is(err, ErrNameTaken)                    // true – Is tittar igenom wrappningen
```

### Bytes och runor

Go-strängar är UTF-8-bytes. `len("Ö")` är **2**. För att räkna tecken
använder du `utf8.RuneCountInString`, och `for _, r := range s` loopar över
runor (`rune` = ett Unicode-tecken), inte bytes.

```go
for i, r := range "Åsa" {
    fmt.Println(i, string(r), unicode.IsLetter(r)) // 0 Å true, 2 s true, 3 a true
}
```

### Mutex

En `sync.Mutex` ser till att bara en goroutine i taget kör koden mellan
`Lock` och `Unlock`:

```go
type Server struct {
    mu      sync.Mutex          // skyddar players
    players map[string]*Player
}

func (s *Server) addPlayer(p *Player) error {
    s.mu.Lock()
    defer s.mu.Unlock()        // lås alltid upp, även vid tidig return
    ...
}
```

Maps i Go är **inte** säkra att använda från flera goroutines samtidigt. Gör
du det utan lås kan programmet krascha med `fatal error: concurrent map writes`.

## Uppgift

1. **`names.go`** (ny fil): skriv `ValidateName` och `NormalizeName`.
2. **`server.go`**:
   - lägg till `mu sync.Mutex` och `players map[string]*Player` i `Server`,
   - skriv `addPlayer(p) error` (ger `ErrNameTaken` om namnet finns),
     `removePlayer(p)` och `PlayerNames() []string` (sorterade),
   - i `handle`: efter `Welcome`, fråga `MsgAskName` tills spelaren anger
     ett giltigt, ledigt namn. Vid fel skriver du `MsgInvalidName` med felet
     (`p.Sendf(MsgInvalidName, err)`) och frågar igen. Skriv sedan
     `MsgWelcomeName` med namnet,
   - ta bort spelaren när hen kopplar ner (tips: `defer`).
3. **`commands.go`**: lägg till `who` som skriver `MsgWho` med antal och
   namnen, sorterade och åtskilda med `MsgListSep`:
   `Spelare online (2): Anna, Bertil`.

Alla texter (`MsgAskName`, `MsgErrNameTooShort`, ...) finns i `messages.go`.
Felvariablerna i `names.go` är redan skapade från dem.

## Testa själv: hitta en data race

Ta tillfälligt bort `s.mu.Lock()` och `defer s.mu.Unlock()` i `addPlayer` och kör

```
./mudschool check --race
```

Läs rapporten – den visar exakt vilka två goroutines som krockade och var.
Sätt sedan tillbaka låset.

## Fundera på

- Varför normaliserar vi namnet *innan* vi kollar om det är upptaget?
- Vad händer om en spelare kopplar ner mitt i inloggningen?

## Utmaningar

- Spara lösenord (hashade med `golang.org/x/crypto/bcrypt`) så att ingen
  annan kan ta ditt namn. Det kommer tillbaka i fas 3 när vi sparar spelare.
- Förbjud några namn (`admin`, `gud`, ...) med ett nytt fel `ErrNameReserved`.
