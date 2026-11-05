# Lektion 15 – Föremål

En värld behöver *saker*: ett rostigt svärd i smedjan, bröd på krogen, en
mystisk nyckel i templet. I den här lektionen ligger föremål i rummen, och
spelare kan plocka upp dem, bära dem, släppa dem och undersöka dem – och du
skriver din första **generiska** funktion.

Vi gör också protokollet lite rikare: varje text-frame får ett `kind`
(`room`, `say`, `move`, `error`, …) så att klienten kan färga loggen i
lektion 17. Gamla klienter ignorerar det nya fältet – det är
kompatibilitetsregeln i `docs/sv/PROTOCOL.md` i praktiken.

## Mål

- Modellera saker med structs och slices av pekare, och flytta dem mellan ägare.
- Skriva en generisk funktion med typvillkor: `func FindMatch[T Matcher](...)`.
- Använda ett gränssnitt (`Matcher`) som villkor.
- Utöka ett protokoll utan att gamla klienter går sönder.

## Go-koncept

### Generics

```go
type Matcher interface {
    Matches(word string) bool
}

func FindMatch[T Matcher](list []T, word string) (T, bool) {
    for _, x := range list {
        if x.Matches(word) {
            return x, true
        }
    }
    var zero T          // nollvärdet för vad T än är (nil för pekare)
    return zero, false
}

it, ok := FindMatch(p.Inventory, "svärd")   // T = *Item härleds
```

Samma funktion hittar varelser i lektion 16 – alla typer med en
`Matches`-metod fungerar. Före Go 1.18 hade du använt `[]Matcher` och
typkontroller istället.

### Ta bort ur en slice

```go
func remove[T comparable](list []T, x T) []T {
    for i, y := range list {
        if y == x {
            return append(list[:i:i], list[i+1:]...)
        }
    }
    return list
}
```

`list[:i:i]` begränsar kapaciteten så att `append` skapar en ny array istället
för att skriva över originalet – en klassisk fälla med slices.

## Kontraktet

- `type Item struct { ID, Name, Description string; Keywords []string; Room string }`
  med JSON-taggarna `id`, `name`, `description`, `keywords`, `room`.
- `func (i *Item) Matches(word string) bool` – ID:t eller ett nyckelord, oavsett skiftläge.
- `type Matcher interface { Matches(word string) bool }` och `FindMatch` som ovan.
- `Room.Items []*Item` (json `"-"`), `World.Items map[string]*Item`,
  `Player.Inventory []*Item`.
- `(w *World) AddItems(items ...*Item) error` och `(w *World) LoadItems(r io.Reader) error`
  (formatet `{"items": [...]}`), som rapporterar alla problem med
  `MsgErrItemFields`, `MsgErrDuplicateItem`, `MsgErrItemRoom`.
- En världsfil får också innehålla `"items"`: `LoadWorld` lägger till dem.
- `DefaultWorld` läser också den nya `items.json` (baka in den som `world.json`).
- `look` visar `MsgItemsHere` efter utgångarna (och före `MsgHere`).
- Kommandon: `get`/`take`, `drop`, `inventory`/`i`, `examine`/`x` – se
  meddelandena `MsgGetWhat`, `MsgNoSuchItem`, `MsgYouGet`, `MsgGets`, `MsgDropWhat`,
  `MsgNotCarrying`, `MsgYouDrop`, `MsgDrops`, `MsgInventory`,
  `MsgInventoryEmpty`, `MsgExamineWhat` i `messages.go`.
- `Frame.Kind` (json `kind,omitempty`) med konstanterna `KindInfo`,
  `KindRoom`, `KindSay`, `KindMove`, `KindError`, `KindCombat`, och
  `Player.SendKind(kind, text)`. `look` → room, `say`/`shout` → say,
  komma/gå → move, misslyckade kommandon → error, resten → info.

## Uppgift

1. **`items.go`** (ny): fyll i `Matches`, `FindMatch`, `AddItems` och de fyra
   kommandona. Registrera dem, och lägg till förkortningarna `i`, `x` och
   `take` i `command.go`.
2. **`world.go`** och **`defaultworld.go`**: de nya fälten, `"items"` i
   världsfilen och `items.json` i standardvärlden.
3. **`frame.go`/`player.go`**: `Kind`, konstanterna, `KindFrame` och
   `SendKind`. Ge `broadcast` och `broadcastRoom` en `kind`-parameter och
   använd rätt kind överallt.
4. **`rooms.go`**: raden `MsgItemsHere` i `describeRoom`.

Allt detta körs i spelloopen från lektion 14 – inga lås.

## Testa själv

```
telnet 127.0.0.1 4000
e
get svärd
i
w
drop svärd
look
```

## Fundera på

- Vem "äger" ett föremål vid varje tidpunkt? Vad skulle hända om två spelare
  skrev `get svärd` exakt samtidigt – och varför är det inget problem nu?
- Varför ligger `Matches` på `*Item` och inte på `Item`?

## Utmaningar

- `get all` och `drop all`.
- Behållare: en kista med föremål i (`get nyckel from kista`). `Command.Args`
  hjälper här.
- `give <föremål> <spelare>`.
