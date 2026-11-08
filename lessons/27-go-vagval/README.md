# Lesson 27 – Pathfinding

The wolf in the dark wood has wandered aimlessly since lesson 16. Now it
**hunts**: when it sees you, it comes – round corners, past trees, the
shortest way. And the rat, badly hurt, **runs away**.

```
###############        ###############        ###############
#.............#        #.............#        #.............#
.@.....w......#   ->   .@.w..........#   ->   .@w...........#   "A wolf attacks you!"
#.............#        #.............#        #.............#
```

## Goals

- Implement **A\***, the most used pathfinding algorithm in games.
- Write a **generic** function with a type parameter constrained by `comparable`.
- Use `container/heap` – a priority queue – through its interface.
- Learn the **Dijkstra map**, the Swiss army knife of roguelike AI.

## Go concepts

### `container/heap`

The standard library has no ready-made priority queue – but a package that
turns *your* slice into one, if it implements `heap.Interface`:

```go
type queue[N any] struct{ items []*entry[N] }
func (q *queue[N]) Len() int           { return len(q.items) }
func (q *queue[N]) Less(i, j int) bool { return q.items[i].f < q.items[j].f }
func (q *queue[N]) Swap(i, j int)      { ... }
func (q *queue[N]) Push(x any)         { ... }  // called by heap.Push
func (q *queue[N]) Pop() any           { ... }  // called by heap.Pop: remove the LAST element
```

Then `heap.Push(q, e)` and `heap.Pop(q)` keep the smallest `f` first –
in O(log n). Note: you call `heap.Push`, not `q.Push`.

### Generics with `comparable`

```go
func AStar[N comparable](start, goal N, neighbors func(N) []N,
	cost func(a, b N) int, h func(N) int) ([]N, bool)
```

`N` can be anything that can be compared with `==` and used as a map key:
`Pos` in a zone, `string` for room IDs, `int` in a test. The same A* finds
paths in zones *and* in the world graph.

## The algorithms

### A\*

Explore the most promising node first: the one with the lowest
`f = g + h`, where `g` is the cost so far and `h` an *estimate* of what
remains. With 8-directional steps that all cost 1 the perfect estimate is the
**Chebyshev distance**, `max(|dx|, |dy|)` – it never overestimates, so the
path found is the shortest.

```
open := a priority queue with start
for the queue is not empty:
	cur := the node with the lowest f
	if cur == goal: walk back through came[...] – that is the path
	for every neighbour n:
		g := cost so far + cost(cur, n)
		if n is new or g is cheaper than before: remember g and came[n] = cur, push n
```

### Dijkstra maps

A breadth-first search from **all goals at once** gives every cell its
distance to the nearest goal. Then decisions become trivial:

- go towards the goal: step to the neighbour with the *lowest* number;
- get away: step to the neighbour with the *highest* number.

*Brogue* uses Dijkstra maps for almost all monster behaviour.

## The contract

- `AStar[N comparable](start, goal, neighbors, cost, h) ([]N, bool)` – the
  cheapest path, **without** `start` and **with** `goal`; `nil, true` when
  start is goal; `nil, false` when there is no path.
- `(z *Zone) Path(from, to Pos, blocked func(Pos) bool) ([]Pos, bool)` – in a
  zone, 8 directions, cost 1, walkable tiles only, not through `blocked` cells
  (may be `nil`) – except `to`.
- `DijkstraMap(z, goals []Pos) map[Pos]int` – the steps to the nearest goal
  (8 directions) for every reachable walkable cell.
- New behaviours in the NPC files (`range` and `below` in `BehaviorSpec`):
  - `{"type": "hunt", "range": 5, "chance": 0.3}` – sees players with
    `ComputeFOV(zone, its position, range)` (default range 6); next to a player
    (Chebyshev ≤ 1) it attacks: `prey.target = n`, `MsgAttacksYou` to the prey and
    `MsgAttacks` to the room; otherwise it takes the first step of `Path`
    towards the nearest visible player, not through other creatures. `chance`
    = probability per tick to act (0 = every tick).
  - `{"type": "flee", "below": 0.4}` – below `below` × its health, with players
    in the room: steps to the neighbour with a higher `DijkstraMap` value than
    where it stands.
- The default world: the wolf hunts, the rat flees.

## Task

`path.go`: `AStar` (the queue is given), `Path`, `DijkstraMap`, `Hunter.Act`
and `Fleer.Act`. Then `newBehavior` and `BehaviorSpec`. Try it: walk into the
dark wood (`morkskogen`) and wait.

## Think about

- Why must `h` never overestimate? Construct a case where a too-big `h` gives a
  path that isn't the shortest.
- The wolf computes a whole A* every time it takes a step. With 100 wolves?
  How could a Dijkstra map from the players replace them all?
- `AStar` works on rooms too. Write `w.RoomPath(from, to string)` – what would
  `h` be? (Hint: the room coordinates from lesson 9.)
- What happens if two wolves want the same cell?

## Challenges

- `travel <room>`: walk automatically, one step per tick, to a room far away
  (A* in the world graph, then A* in every zone).
- Packs: wolves that try to surround you (each one picks a different free cell next to you).
- Fleeing creatures that leave the room when they reach an opening.
