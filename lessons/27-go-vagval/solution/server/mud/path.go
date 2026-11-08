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
	if start == goal {
		return nil, true
	}
	open := &queue[N]{}
	heap.Push(open, &entry[N]{node: start, g: 0, f: h(start)})
	came := map[N]N{}
	best := map[N]int{start: 0}
	for open.Len() > 0 {
		cur := heap.Pop(open).(*entry[N])
		if cur.node == goal {
			var path []N
			for n := goal; n != start; n = came[n] {
				path = append(path, n)
			}
			for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
				path[i], path[j] = path[j], path[i]
			}
			return path, true
		}
		if cur.g > best[cur.node] {
			continue // an old, more expensive entry for a node we already improved
		}
		for _, next := range neighbors(cur.node) {
			g := cur.g + cost(cur.node, next)
			if old, seen := best[next]; seen && g >= old {
				continue
			}
			best[next] = g
			came[next] = cur.node
			heap.Push(open, &entry[N]{node: next, g: g, f: g + h(next), seq: open.next()})
		}
	}
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
	neighbors := func(p Pos) []Pos {
		var out []Pos
		for _, d := range neighbours8 {
			n := p.Add(d)
			if Walkable(z.At(n)) && (n == to || blocked == nil || !blocked(n)) {
				out = append(out, n)
			}
		}
		return out
	}
	one := func(a, b Pos) int { return 1 }
	return AStar(from, to, neighbors, one, func(p Pos) int { return chebyshev(p, to) })
}

// DijkstraMap gives every walkable cell its distance (in steps, diagonals
// allowed) to the nearest goal. Unreachable cells are missing. A creature
// that wants to get away walks to the neighbour with the highest number;
// one that wants to get closer, to the lowest (Brogue uses this for everything).
func DijkstraMap(z *Zone, goals []Pos) map[Pos]int {
	dist := map[Pos]int{}
	var queue []Pos
	for _, g := range goals {
		if Walkable(z.At(g)) {
			dist[g] = 0
			queue = append(queue, g)
		}
	}
	for len(queue) > 0 {
		p := queue[0]
		queue = queue[1:]
		for _, d := range neighbours8 {
			n := p.Add(d)
			if _, seen := dist[n]; !seen && Walkable(z.At(n)) {
				dist[n] = dist[p] + 1
				queue = append(queue, n)
			}
		}
	}
	return dist
}

// --- behaviours that use it ------------------------------------------------

// Hunter walks towards the nearest player it can see and attacks when it is
// next to them.
type Hunter struct {
	Range  int     // how far it sees
	Chance float64 // probability per tick to take a step (0 = every tick)
}

func (b Hunter) Act(s *Server, n *NPC) {
	if n.dead || (b.Chance > 0 && s.rng.Float64() >= b.Chance) {
		return
	}
	sight := b.Range
	if sight <= 0 {
		sight = 6
	}
	z := s.world.zoneOf(n.Location)
	sees := ComputeFOV(z, n.Pos, sight)
	var prey *Player
	for _, p := range s.playersInRoom(n.Location, nil) {
		if sees[p.Pos] && (prey == nil || chebyshev(n.Pos, p.Pos) < chebyshev(n.Pos, prey.Pos)) {
			prey = p
		}
	}
	if prey == nil {
		return
	}
	if chebyshev(n.Pos, prey.Pos) <= 1 {
		if prey.target == nil && n.Attackable {
			prey.target = n
			prey.SendKind(KindCombat, fmt.Sprintf(MsgAttacksYou, n.Name))
			s.broadcastRoom(n.Location, prey, KindCombat, fmt.Sprintf(MsgAttacks, n.Name, prey.Name))
		}
		return
	}
	others := s.npcCells(n)
	if path, ok := z.Path(n.Pos, prey.Pos, func(p Pos) bool { return others[p] }); ok && len(path) > 0 && path[0] != prey.Pos {
		n.Pos = path[0]
	}
}

// Fleer runs away from players when its health is low.
type Fleer struct {
	Below float64 // flee when hp < Below × max hp
}

func (b Fleer) Act(s *Server, n *NPC) {
	if n.dead || n.HP == 0 || float64(n.hp) >= b.Below*float64(n.HP) {
		return
	}
	var players []Pos
	for _, p := range s.playersInRoom(n.Location, nil) {
		players = append(players, p.Pos)
	}
	if len(players) == 0 {
		return
	}
	z := s.world.zoneOf(n.Location)
	dist := DijkstraMap(z, players)
	others := s.npcCells(n)
	best, bestDist := n.Pos, dist[n.Pos]
	for _, d := range neighbours8 {
		c := n.Pos.Add(d)
		if dc, ok := dist[c]; ok && dc > bestDist && !others[c] {
			best, bestDist = c, dc
		}
	}
	n.Pos = best
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
