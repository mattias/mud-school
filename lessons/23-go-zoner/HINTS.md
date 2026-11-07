# Hints

## opening

Collect the cells of that edge, sort them by distance to the middle, and
return the first walkable one:

```go
mid := z.Center()
sort.SliceStable(cells, func(i, j int) bool { return dist2(cells[i], mid) < dist2(cells[j], mid) })
for _, c := range cells {
	if Walkable(z.At(c)) {
		return c, true
	}
}
```

For up/down, look for the first `<` or `>` tile instead.

## arrival

`from` is the side you came in by. The step *inwards* is the opposite
direction: coming in from the south, you take a step north.

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

## login sync

`defer s.syncAll()` inside the `s.call` that broadcasts `MsgJoined` runs the
sync after `look` – inside the same call.
