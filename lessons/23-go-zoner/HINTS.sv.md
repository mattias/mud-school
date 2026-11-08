# Ledtrådar

## opening

Samla kantens rutor, sortera dem efter avståndet till mitten och
returnera den första gångbara:

```go
mid := z.Center()
sort.SliceStable(cells, func(i, j int) bool { return dist2(cells[i], mid) < dist2(cells[j], mid) })
for _, c := range cells {
	if Walkable(z.At(c)) {
		return c, true
	}
}
```

För upp/ner letar du i stället efter den första rutan `<` eller `>`.

## arrival

`from` är sidan du kom in från. Steget *inåt* är åt motsatt håll: kommer
du in från söder tar du ett steg norrut.

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

## Synk vid inloggning

`defer s.syncAll()` inne i det `s.call` som skickar `MsgJoined` kör synken
efter `look` – i samma anrop.
