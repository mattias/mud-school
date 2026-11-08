# ヒント

## AStar のループ

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

キューにあるノードを *更新* する（ヒープでは難しい）代わりに、もう一度入れて、古いコピーが
出てきたら飛ばします。

## スライスを逆順にする

```go
for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
	path[i], path[j] = path[j], path[i]
}
```

（または `slices.Reverse(path)`）。

## DijkstraMap

レッスン 24 の `reachable` と同じ塗りつぶしです。ただし *すべての* 目標を距離 0 で
キューに入れて始め、`neighbours8` を使い、`dist[n] = dist[p] + 1` を記録します。

## Hunter

```go
sees := ComputeFOV(z, n.Pos, sight)
for _, p := range s.playersInRoom(n.Location, nil) {
	if sees[p.Pos] && (prey == nil || chebyshev(n.Pos, p.Pos) < chebyshev(n.Pos, prey.Pos)) {
		prey = p
	}
}
```
