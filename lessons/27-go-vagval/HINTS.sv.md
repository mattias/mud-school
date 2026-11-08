# Ledtrådar

## AStar, loopen

```go
open := &queue[N]{}
heap.Push(open, &entry[N]{node: start, g: 0, f: h(start)})
came := map[N]N{}
best := map[N]int{start: 0}
for open.Len() > 0 {
	cur := heap.Pop(open).(*entry[N])
	if cur.node == goal {
		// build the path: from goal back to start through came, then reverse it
	}
	if cur.g > best[cur.node] {
		continue // an old entry – we have found a cheaper way since
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
```

I stället för att *uppdatera* en nod som redan ligger i kön (svårt med en
heap) lägger du in den igen och hoppar över den gamla kopian när den kommer ut.

## Vända på en slice

```go
for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
	path[i], path[j] = path[j], path[i]
}
```

(eller `slices.Reverse(path)`).

## DijkstraMap

Samma flood fill som `reachable` i lektion 24 – men börja med *alla* mål i
kön på avstånd 0, använd `neighbours8` och spara `dist[n] = dist[p] + 1`.

## Hunter

```go
sees := ComputeFOV(z, n.Pos, sight)
for _, p := range s.playersInRoom(n.Location, nil) {
	if sees[p.Pos] && (prey == nil || chebyshev(n.Pos, p.Pos) < chebyshev(n.Pos, prey.Pos)) {
		prey = p
	}
}
```
