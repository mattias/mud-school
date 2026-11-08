package mud

import (
	"container/heap"
	"fmt"
)

// Pathfinding: A* finds the shortest way from one place to another, and a
// Dijkstra map tells every cell how far it is to the nearest goal.

// AStar finds the cheapest path from start to goal. It is generic: N can be
// a Pos in a zone, a room ID in the world graph, an int in a test – anything
// comparable. neighbors lists where you can go from a node, cost is the price
// of one step, and h estimates the remaining cost (it must never
// overestimate, or the path may not be the shortest).
// The path starts with the first step after start and ends with goal.
func AStar[N comparable](start, goal N, neighbors func(N) []N, cost func(a, b N) int, h func(N) int) ([]N, bool) {
	// TODO(lesson 27): see the README and HINTS. Use the queue below with
	// container/heap, remember where every node came from (came) and the
	// best cost so far (best), and build the path backwards from goal.
	_ = heap.Push
	return nil, false
}

// entry is a node in the priority queue: f = g (cost so far) + h (estimate).
type entry[N any] struct {
	node N
	g, f int
	seq  int // insertion order: equal f values come out first-in first-out
}

// queue is a priority queue for container/heap: the smallest f first.
type queue[N any] struct {
	items []*entry[N]
	count int
}

func (q *queue[N]) next() int { q.count++; return q.count }
func (q *queue[N]) Len() int  { return len(q.items) }
func (q *queue[N]) Less(i, j int) bool {
	a, b := q.items[i], q.items[j]
	if a.f != b.f {
		return a.f < b.f
	}
	return a.seq < b.seq
}
func (q *queue[N]) Swap(i, j int) { q.items[i], q.items[j] = q.items[j], q.items[i] }
func (q *queue[N]) Push(x any)    { q.items = append(q.items, x.(*entry[N])) }
func (q *queue[N]) Pop() any {
	last := q.items[len(q.items)-1]
	q.items = q.items[:len(q.items)-1]
	return last
}

// neighbours8 are the eight cells around a cell, in a fixed order.
var neighbours8 = []Pos{{0, -1}, {1, -1}, {1, 0}, {1, 1}, {0, 1}, {-1, 1}, {-1, 0}, {-1, -1}}

// chebyshev is the distance when diagonal steps cost the same as straight ones.
func chebyshev(a, b Pos) int { return max(abs(a.X-b.X), abs(a.Y-b.Y)) }

// Path is the shortest walk from `from` to `to` in the zone, with diagonal
// steps. blocked (may be nil) marks extra cells you can't walk through, e.g.
// other creatures; `to` itself is always allowed.
func (z *Zone) Path(from, to Pos, blocked func(Pos) bool) ([]Pos, bool) {
	// TODO(lesson 27): AStar over Pos: neighbours8 that are walkable (and not
	// blocked, except `to`), cost 1, heuristic chebyshev(p, to).
	return nil, false
}

// DijkstraMap gives every walkable cell its distance (in steps, diagonals
// allowed) to the nearest goal. Unreachable cells are missing. A creature
// that wants to get away walks to the neighbour with the highest number;
// one that wants to get closer, to the lowest (Brogue uses this for everything).
func DijkstraMap(z *Zone, goals []Pos) map[Pos]int {
	// TODO(lesson 27): a breadth-first search from all goals at once.
	return map[Pos]int{}
}

// --- behaviours that use it ------------------------------------------------

// Hunter walks towards the nearest player it can see and attacks when it is
// next to them.
type Hunter struct {
	Range  int     // how far it sees
	Chance float64 // probability per tick to take a step (0 = every tick)
}

func (b Hunter) Act(s *Server, n *NPC) {
	// TODO(lesson 27): find the nearest player in the room that the creature
	// sees (ComputeFOV from n.Pos); next to them (chebyshev <= 1): attack
	// (prey.target = n, MsgAttacksYou, MsgAttacks to the room); otherwise take
	// the first step of z.Path towards them, not through other creatures.
	_ = fmt.Sprintf
}

// Fleer runs away from players when its health is low.
type Fleer struct {
	Below float64 // flee when hp < Below × max hp
}

func (b Fleer) Act(s *Server, n *NPC) {
	// TODO(lesson 27): below Below × max health and with players in the room:
	// step to the neighbour with the highest DijkstraMap value (from the
	// players), if it is higher than where it stands.
}

// npcCells is where the other creatures in n's room stand.
func (s *Server) npcCells(n *NPC) map[Pos]bool {
	cells := map[Pos]bool{}
	for _, other := range n.Location.NPCs {
		if other != n {
			cells[other.Pos] = true
		}
	}
	return cells
}
