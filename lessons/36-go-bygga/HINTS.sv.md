# Ledtrådar

## Tolka "@dig north Dark cave"

```go
dir := cmd.Args[0]
if full, ok := aliases[dir]; ok {
	dir = full // "n" -> "north"
}
name := strings.TrimSpace(strings.TrimPrefix(cmd.Rest, cmd.Args[0]))
d, ok := directionDelta[dir]
```

## cmdDig med change

```go
err := s.change(func() {
	s.world.Rooms[room.ID] = room
	here.Exits[dir] = room.ID
}, func() {
	delete(s.world.Rooms, room.ID)
	delete(here.Exits, dir)
})
```

`here.Exits` kan vara `nil` för ett rum utan utgångar – skapa den först.

## saveWorld, goroutinen

```go
s.saveSeq++
seq := s.saveSeq
go func() {
	s.saveMu.Lock()
	defer s.saveMu.Unlock()
	if seq < s.savedSeq {
		return // a newer snapshot has already been written
	}
	s.savedSeq = seq
	// writeFileAtomic, then os.Stat and s.ownWrite.Store(fileStamp{...})
}()
```

## watchWorld

Läs filens stämpel och `s.ownWrite` under `s.saveMu`, och hoppa över
omladdningen när de är lika.
