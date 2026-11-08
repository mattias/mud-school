# ヒント

## play

`handle` のうち `MsgWelcomeName` と入場の後のすべては `play` のものです。遅延実行される
ログアウト処理とコマンドのループ。遅延関数の先頭で:

```go
if s.handingOver.Load() {
	return // the new server has the player now
}
```

## PrepareCopyover

```go
s.lmu.Lock()
lf, err := fileOf(s.listeners[0])
s.lmu.Unlock()
files := []*os.File{lf}
s.call(func() {
	state.Ticks = s.ticks
	for _, name := range s.playerNames() {
		p := s.players[name]
		f, ferr := fileOf(p.conn)
		// ...
		files = append(files, f)
		cp := CopyoverPlayer{Name: p.Name, /* ... */, File: len(files) - 1}
	}
})
```

## HandOver

`Shutdown` をもとにします。リスナーを閉じ、ハンドラーを待ち、ループを止めるのは同じ。ただし
最初に `s.handingOver.Store(true)` とし、ゲームループの中では別れと期限の代わりに、各接続を
`p.conn.Close()` します。

## adopt のゴルーチン

```go
s.lmu.Lock()
s.handlers.Add(1)
s.lmu.Unlock()
go func() {
	defer s.handlers.Done()
	defer conn.Close()
	defer p.close()
	defer s.call(func() { delete(s.conns, p) })
	s.play(p)
}()
```
