# Hints

## watchWorld

```go
stat := func() (time.Time, int64) {
	fi, err := os.Stat(s.worldPath)
	if err != nil {
		return time.Time{}, -1
	}
	return fi.ModTime(), fi.Size()
}
lastMod, lastSize := stat()
```

Then the ticker loop: when the file can be read and `mod`/`size` differ from
the last time, remember them and call `s.reloadFromFile()`.

## swapWorld, the position

```go
z := nw.zoneOf(room)
if room.ID != p.Room.ID || !Walkable(z.At(p.Pos)) {
	p.Pos = z.freeNear(z.Center(), nil)
}
```

## swapWorld, the things

```go
for i, it := range p.Inventory {
	if fresh := s.takeItem(it.ID); fresh != nil {
		p.Inventory[i] = fresh
	}
}
```

`takeItem` (lesson 20) looks in `s.world` – so set `s.world = nw` first.

## Starting the watcher

```go
s.startOnce.Do(func() {
	go s.run()
	if s.worldPath != "" && s.reloadEvery > 0 {
		go s.watchWorld()
	}
})
```
