# Hints

## Parsing "@dig north Dark cave"

```go
dir := cmd.Args[0]
if full, ok := aliases[dir]; ok {
	dir = full // "n" -> "north"
}
name := strings.TrimSpace(strings.TrimPrefix(cmd.Rest, cmd.Args[0]))
d, ok := directionDelta[dir]
```

## cmdDig with change

```go
err := s.change(func() {
	s.world.Rooms[room.ID] = room
	here.Exits[dir] = room.ID
}, func() {
	delete(s.world.Rooms, room.ID)
	delete(here.Exits, dir)
})
```

`here.Exits` may be `nil` for a room without exits – make it first.

## saveWorld, the goroutine

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

Read the file's stamp and `s.ownWrite` under `s.saveMu`, and skip the reload
when they are the same.
