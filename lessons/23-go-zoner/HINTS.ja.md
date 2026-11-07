# ヒント

## opening

その辺のマスを集め、中央からの距離で並べ替え、最初の歩けるマスを返します:

```go
mid := z.Center()
sort.SliceStable(cells, func(i, j int) bool { return dist2(cells[i], mid) < dist2(cells[j], mid) })
for _, c := range cells {
	if Walkable(z.At(c)) {
		return c, true
	}
}
```

上下の場合は、代わりに最初の `<` か `>` のタイルを探します。

## arrival

`from` は入ってきた側です。*内側への* 1 歩は反対の方向です。南から入ってきたなら北へ 1 歩。

```go
p, ok := z.opening(from)
in := p.Add(stepDelta[opposite[from]])
```

## syncPlayer

```go
current := make(map[string]Entity, len(now))
for _, e := range now {
	current[e.ID] = e
	old, known := p.seen[e.ID]
	switch {
	case !known:
		d.Added = append(d.Added, e)
	case old != e:
		d.Moved = append(d.Moved, e)
	}
}
for id := range p.seen {
	if _, still := current[id]; !still {
		d.Removed = append(d.Removed, id)
	}
}
sort.Strings(d.Removed) // maps have no order – keep frames predictable
```

## step

```go
to := p.Pos.Add(stepDelta[dir])
if !z.In(to) {
	if _, ok := p.Room.Exits[dir]; ok {
		s.travel(p, dir)
		return
	}
	// MsgBlocked
}
```

## ログイン時の同期

`MsgJoined` を送る `s.call` の中で `defer s.syncAll()` とすると、同じ呼び出しの中で
`look` の後に同期が走ります。
