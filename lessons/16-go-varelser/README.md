# Lesson 16 – Creatures

Mudby comes alive: a black cat wanders around the square, Greta the
innkeeper chats in the tavern, a crow hops along the forest edge – and
somewhere in the dark wood there is a wolf. Creatures (*NPCs*, non-player
characters) act on their own, so they live in the game loop's `onTick`.

What a creature does is described by **behaviours**: small values that
implement an interface. The world file says which ones each creature has.

```json
{"id": "cat", "name": "a black cat", "room": "torget",
 "behaviors": [{"type": "wander", "chance": 0.03},
               {"type": "talk", "chance": 0.02, "lines": ["Meow."]}]}
```

## Goals

- Design with a small interface (`Behavior`) and several implementations.
- Build values from data: a "factory" that turns a `BehaviorSpec` into a `Behavior`.
- Make randomness testable: inject a `*rand.Rand` with a fixed seed.
- Reuse the generic `FindMatch` from lesson 15 for a new type.

## Go concepts

### An interface with several implementations

```go
type Behavior interface {
    Act(s *Server, n *NPC)
}

type Wanderer struct{ Chance float64 }
func (b Wanderer) Act(s *Server, n *NPC) { ... }

type Talker struct{ Chance float64; Lines []string }
func (b Talker) Act(s *Server, n *NPC) { ... }
```

A test can add its own `Behavior` – look at `countingBehavior` in
`mudschool_16_test.go`. That is the power of interfaces: the loop doesn't
know or care what a behaviour does.

### Randomness you can test

```go
rng := rand.New(rand.NewPCG(1, 2))  // math/rand/v2, a fixed seed
rng.Float64()                        // the same sequence every run
rng.IntN(3)                          // 0, 1 or 2
```

The server takes `Options.Rand`. Production uses a random seed; tests use a
fixed one. For the same seed to give the same result, everything must happen
in a fixed order – that's why `npcsByID` sorts and `Wanderer` goes through
`Directions` in order instead of ranging over a map.

## The contract

- `type NPC struct` with JSON fields `id`, `name`, `description`,
  `keywords`, `room`, `behaviors`, and (for lesson 18, unused for now) `hp`,
  `damage`, `attackable`. At runtime: `Location *Room`.
- `func (n *NPC) Matches(word string) bool` and `func (n *NPC) AddBehavior(b Behavior)`.
- `type Behavior interface { Act(s *Server, n *NPC) }`, `Wanderer`, `Talker`,
  `BehaviorSpec{Type, Chance, Lines}` with the types `"wander"` and `"talk"`.
- `Room.NPCs []*NPC`, `World.NPCs map[string]*NPC`,
  `(w *World) AddNPCs(...)` and `(w *World) LoadNPCs(r)` (`{"npcs": [...]}`),
  with `MsgErrNPCFields`, `MsgErrDuplicateNPC`, `MsgErrNPCRoom`,
  `MsgErrBehavior`. A world file may contain `"npcs"`; `DefaultWorld` loads
  the new `npcs.json`.
- `Options.Rand *rand.Rand` (`math/rand/v2`; nil = random seed).
- Every tick, every creature's behaviours run (in ID order).
- `look` lists creatures after the players in `MsgHere`. `examine` works on
  creatures. Creatures that come and go use `MsgLeaves`/`MsgArrives` with kind
  `move`; talking uses `MsgSays` with kind `say`.

## Task

1. **`npc.go`** (new): fill in `Matches`, `Wanderer.Act`, `Talker.Act`,
   `moveNPC` and `AddNPCs`.
2. **`world.go`**/**`defaultworld.go`**: the new fields, `"npcs"` in the
   world file, `npcs.json` in the default world.
3. **`server.go`**/**`loop.go`**: `Options.Rand`, `s.rng`, and `onTick`
   runs all behaviours.
4. **`rooms.go`**/**`items.go`**: creatures in `look` and in `examine`.

## Try it yourself

```
./mudschool run server -tick 20ms
```

A fast tick makes the creatures restless. Stand on the square and wait for
the cat. Visit the tavern.

## Think about

- Why does `Wanderer.Act` receive the server and the creature as arguments,
  instead of storing them in the `Wanderer`?
- What would happen with the seeded tests if `onTick` ranged directly over
  the map `s.world.NPCs`?
- Should a creature be allowed to wander anywhere? How would you keep the
  innkeeper in the tavern?

## Challenges

- A `follow` behaviour: the crow follows the nearest player.
- Day and night: the wolf only wanders at night (use `GameTime`).
- `say` to a creature: let the innkeeper answer when someone says "ale".
