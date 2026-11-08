# Ledtrådar

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

Sedan ticker-loopen: när filen går att läsa och `mod`/`size` skiljer sig från
förra gången, kom ihåg dem och anropa `s.reloadFromFile()`.

## swapWorld, positionen

```go
z := nw.zoneOf(room)
if room.ID != p.Room.ID || !Walkable(z.At(p.Pos)) {
	p.Pos = z.freeNear(z.Center(), nil)
}
```

## swapWorld, sakerna

```go
for i, it := range p.Inventory {
	if fresh := s.takeItem(it.ID); fresh != nil {
		p.Inventory[i] = fresh
	}
}
```

`takeItem` (lektion 20) letar i `s.world` – så sätt `s.world = nw` först.

## Starta bevakningen

```go
s.startOnce.Do(func() {
	go s.run()
	if s.worldPath != "" && s.reloadEvery > 0 {
		go s.watchWorld()
	}
})
```
