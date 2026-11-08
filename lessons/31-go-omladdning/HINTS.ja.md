# ヒント

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

あとはティッカーのループです。ファイルが読めて、`mod`/`size` が前回と違えば、それを覚えて
`s.reloadFromFile()` を呼びます。

## swapWorld: 位置

```go
z := nw.zoneOf(room)
if room.ID != p.Room.ID || !Walkable(z.At(p.Pos)) {
	p.Pos = z.freeNear(z.Center(), nil)
}
```

## swapWorld: 持ち物

```go
for i, it := range p.Inventory {
	if fresh := s.takeItem(it.ID); fresh != nil {
		p.Inventory[i] = fresh
	}
}
```

`takeItem`（レッスン 20）は `s.world` を探すので、先に `s.world = nw` とします。

## 監視の開始

```go
s.startOnce.Do(func() {
	go s.run()
	if s.worldPath != "" && s.reloadEvery > 0 {
		go s.watchWorld()
	}
})
```
