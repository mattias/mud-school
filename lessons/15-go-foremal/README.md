# Lesson 15 – Items

A world needs *things*: a rusty sword in the smithy, bread in the tavern, a
mysterious key in the temple. In this lesson items lie in the rooms, players
can pick them up, carry them, drop them and examine them – and you write your
first **generic** function.

We also make the protocol a little richer: every text frame gets a `kind`
(`room`, `say`, `move`, `error`, …) so that the client can colour its log in
lesson 17. Old clients ignore the new field – that is the compatibility rule
from `docs/en/PROTOCOL.md` in practice.

## Goals

- Model things with structs and slices of pointers, and move them between owners.
- Write a generic function with a type constraint: `func FindMatch[T Matcher](...)`.
- Use an interface (`Matcher`) as a constraint.
- Extend a protocol without breaking old clients.

## Go concepts

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
    var zero T          // the zero value of whatever T is (nil for pointers)
    return zero, false
}

it, ok := FindMatch(p.Inventory, "sword")   // T = *Item is inferred
```

The same function will find creatures in lesson 16 – any type with a
`Matches` method works. Before Go 1.18 you would have used `[]Matcher` and
type assertions instead.

### Removing from a slice

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

`list[:i:i]` caps the capacity so that `append` makes a new array instead of
overwriting the original – a classic slice gotcha.

## The contract

- `type Item struct { ID, Name, Description string; Keywords []string; Room string }`
  with JSON tags `id`, `name`, `description`, `keywords`, `room`.
- `func (i *Item) Matches(word string) bool` – the ID or a keyword, ignoring case.
- `type Matcher interface { Matches(word string) bool }` and `FindMatch` as above.
- `Room.Items []*Item` (json `"-"`), `World.Items map[string]*Item`,
  `Player.Inventory []*Item`.
- `(w *World) AddItems(items ...*Item) error` and `(w *World) LoadItems(r io.Reader) error`
  (format `{"items": [...]}`), reporting all problems with
  `MsgErrItemFields`, `MsgErrDuplicateItem`, `MsgErrItemRoom`.
- A world file may contain `"items"` too: `LoadWorld` adds them.
- `DefaultWorld` also loads the new `items.json` (embed it like `world.json`).
- `look` shows `MsgItemsHere` after the exits (and before `MsgHere`).
- When a player leaves, everything they carried goes back to the room each
  item started in (`Item.Room`) – nothing disappears from the world.
- Commands: `get`/`take`, `drop`, `inventory`/`i`, `examine`/`x` – see the
  messages `MsgGetWhat`, `MsgNoSuchItem`, `MsgYouGet`, `MsgGets`, `MsgDropWhat`,
  `MsgNotCarrying`, `MsgYouDrop`, `MsgDrops`, `MsgInventory`,
  `MsgInventoryEmpty`, `MsgExamineWhat` in `messages.go`.
- `Frame.Kind` (json `kind,omitempty`) with the constants `KindInfo`,
  `KindRoom`, `KindSay`, `KindMove`, `KindError`, `KindCombat`, and
  `Player.SendKind(kind, text)`. `look` → room, `say`/`shout` → say,
  coming/going → move, failed commands → error, the rest → info.

## Task

1. **`items.go`** (new): fill in `Matches`, `FindMatch`, `AddItems` and the
   four commands. Register them, and add the aliases `i`, `x` and `take` in
   `command.go`.
2. **`world.go`** and **`defaultworld.go`**: the new fields, `"items"` in the
   world file, and `items.json` in the default world.
3. **`frame.go`/`player.go`**: `Kind`, the constants, `KindFrame` and
   `SendKind`. Give `broadcast` and `broadcastRoom` a `kind` parameter and
   use the right kind everywhere.
4. **`rooms.go`**: the `MsgItemsHere` line in `describeRoom`.

All of this runs in the game loop from lesson 14 – no locks.

## Try it yourself

```
./mudschool connect
e
get sword
i
w
drop sword
look
```

## Think about

- Who "owns" an item at any moment? What would happen if two players typed
  `get sword` at exactly the same time – and why is that no problem now?
- Why does `Matches` live on `*Item` and not on `Item`?

## Challenges

- `get all` and `drop all`.
- Containers: a chest that holds items (`get key from chest`). `Command.Args`
  helps here.
- `give <item> <player>`.
