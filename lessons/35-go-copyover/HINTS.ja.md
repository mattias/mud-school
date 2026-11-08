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

## main.go

`main.go` の変更はおおよそ次のとおりです:

```go
// After a copyover the listener is inherited, not opened.
statePath := os.Getenv("MUD_COPYOVER")
os.Unsetenv("MUD_COPYOVER")
var l net.Listener
var err error
if statePath == "" {
	if l, err = net.Listen("tcp", *addr); err != nil { /* as before */ }
}
// ...
srv := mud.NewServerWithOptions(mud.Options{ /* as before */ OnCopyover: startNewServer})
if statePath != "" {
	if l, err = adoptCopyover(srv, statePath); err != nil {
		log.Fatalf("copyover: %v", err)
	}
}
// ...
// the Ctrl+C goroutine: select { case <-ctx.Done(): case <-srv.Done(): return }
// and after Serve (the server stops either way; ctx says which way):
<-srv.Done()
if ctx.Err() != nil { // Ctrl+C
	<-shutdownDone
	log.Print("bye")
} else {
	log.Print("handed over to the new server – bye")
}
```
