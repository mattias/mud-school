# ヒント

## "@dig north Dark cave" の解析

```go
dir := cmd.Args[0]
if full, ok := aliases[dir]; ok {
	dir = full // "n" -> "north"
}
name := strings.TrimSpace(strings.TrimPrefix(cmd.Rest, cmd.Args[0]))
d, ok := directionDelta[dir]
```

## change を使う cmdDig

```go
err := s.change(func() {
	s.world.Rooms[room.ID] = room
	here.Exits[dir] = room.ID
}, func() {
	delete(s.world.Rooms, room.ID)
	delete(here.Exits, dir)
})
```

出口のない部屋では `here.Exits` が `nil` のことがあります。先に作りましょう。

## saveWorld のゴルーチン

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

`s.saveMu` の下でファイルのスタンプと `s.ownWrite` を読み、同じなら再読み込みを飛ばします。
