# ヒント

## Accept

```go
select {
case c := <-l.conns:
	return c, nil
case <-l.closed:
	return nil, &net.OpError{Op: "accept", Net: "websocket", Addr: l.addr, Err: net.ErrClosed}
}
```

`*net.OpError` には `Unwrap` メソッドがあるので、`errors.Is(err, net.ErrClosed)` で見つかります。

## Close – 一度だけ

```go
l.once.Do(func() { close(l.closed) })
return nil
```

閉じたチャネルをもう一度閉じるとパニックします。`sync.Once` で 2 回目の `Close` は何もしません。

## ServeHTTP

`websocket.Accept` が失敗したら、もうエラーの応答（よそのオリジンなら 403）を書いています。
ログを残して return するだけです。`Offer` が false を返したらサーバーは停止中です:
`c.Close(websocket.StatusGoingAway, "...")`。

## main.go

```go
//go:embed web
var webFiles embed.FS

files, _ := fs.Sub(webFiles, "web") // "/" が web/index.html になるように
gw := mud.NewGateway(files)
web := &http.Server{Addr: *webAddr, Handler: gw, ReadHeaderTimeout: 10 * time.Second}
wl, err := net.Listen("tcp", *webAddr)
...
go web.Serve(wl)
go srv.Serve(gw.Listener())
```

`ReadHeaderTimeout` は、接続を開いたままリクエストを送り終えないクライアントを止めます。
そして Web がオンの間は `OnCopyover` を nil に。

## 手で試す

ページを開いて F12 を押し、コンソールとネットワークのタブを見ましょう。`/ws` への
リクエストは `101 Switching Protocols` になっているはずです。
