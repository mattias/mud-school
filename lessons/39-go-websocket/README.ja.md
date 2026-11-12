# レッスン 39 – WebSocket

telnet とあなたの Rust クライアントは、あなたには十分です。でも、たいていの人があなたの
MUD を試すのは、それがクリック 1 回で、つまりブラウザで遊べるときだけです。ブラウザは
普通の TCP 接続を開けません。でも **WebSocket** なら開けます。HTTP リクエストとして
始まり、そのまま双方向に開きっぱなしになる接続です。

このレッスンでサーバーは **ゲートウェイ** を手に入れます。Web ページ（用意済み:
`web/index.html`、`web/mud.js`、`web/style.css`）を配信し、各 WebSocket を `net.Conn` に
変えます。そして仕掛け: ゲートウェイは自分の *リスナー* を持ち、`Serve`（telnet と
同じ `Serve`）がそこからプレイヤーを受け取ります。

```
ブラウザ ──HTTP──▶ Gateway ──ファイル──▶ index.html, mud.js
ブラウザ ──/ws───▶ Gateway ──net.Conn──▶ ConnListener ──Accept──▶ Serve ──▶ handle
telnet   ──────────────────────────────▶ net.Listener ──Accept──▶ Serve ──▶ handle
```

ログインし、歩き、話す。サーバーの残りの部分は、誰がブラウザを使っているか知りません。

## 目標

- `net/http` でファイルと WebSocket を配信する。
- サーバーがすでに使っている **インターフェースを満たす** 型を作る: `net.Listener`、
  ソケットの代わりにチャネルで。
- プログラムの残りはそのまま: 新しいものを古いインターフェースに **合わせる**。
- WebSocket がなぜページの出どころ（`Origin`）を確かめるのかを知る。

## Go の概念

### `http.Handler`

```go
func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/ws" {
		g.files.ServeHTTP(w, r) // http.FileServerFS
		return
	}
	...
}
```

`ServeHTTP` メソッドを持つものはすべてハンドラーです: `http.ListenAndServe(addr, g)`。

### チャネルでできたリスナー

`net.Listener` は 3 つのメソッド、`Accept`・`Close`・`Addr` です。`ConnListener` は
バッファなしのチャネルから接続を受け取ります:

```go
func (l *ConnListener) Offer(c net.Conn) bool {
	select {
	case l.conns <- c:   // 誰かが Accept で待っている – 渡す
		return true
	case <-l.closed:     // 閉じた: もう誰も受け取らない
		return false
	}
}
```

`Accept` は逆向きの `select` です。`Close` は `closed` を閉じます（`sync.Once` で一度だけ）。
これで待っている `Accept` と `Offer` がすべて起きます。その後 `Accept` は
`errors.Is(err, net.ErrClosed)` になるエラーを返します。TCP リスナーと同じなので、
`Serve` はいつもどおり止まります。

### WebSocket から `net.Conn` へ

ライブラリ `github.com/coder/websocket` が WebSocket のプロトコルを担います:

```go
c, err := websocket.Accept(w, r, &websocket.AcceptOptions{OriginPatterns: g.Origins})
conn := websocket.NetConn(context.Background(), c, websocket.MessageText)
```

`NetConn` はそれを `net.Conn` で包みます。サーバーが書いたものはテキストメッセージになり、
ブラウザが送ったものはストリームとして読めます。接続は HTTP リクエストより長生きなので
`context.Background()` を渡します。`ServeHTTP` が戻ると終わる `r.Context()` ではありません。

### Origin

*別の* サイトのページも、訪問者のブラウザを使ってあなたのサーバーに WebSocket を開けます。
`Accept` は、ページが同じホスト（または `Origins` のどれか）から来たのでなければ拒否します。
telnet にはない問題ですが、Web にはあります。

## 契約

- `NewConnListener(addr) *ConnListener` と `Offer(c) bool`、`Accept`、`Close`（2 回でも安全）、
  `Addr`。`Close` の後の `Accept` は `net.ErrClosed` であるエラー、`Offer` は false。
- `NewGateway(files fs.FS) *Gateway`、`Listener() net.Listener`、そして `ServeHTTP`: `/ws` は
  ゲーム（テキストメッセージで telnet と同じ行。`@hello` の後は JSON）、それ以外はファイル。
  他のサイトのページは拒否する。
- Web プレイヤーの `RemoteAddr` はブラウザのアドレス。
- `Shutdown` は Web プレイヤーにも別れを告げ、`Serve(gw.Listener())` は `ErrServerClosed` を返す。
- `main.go`: `-web`（既定 `127.0.0.1:8080`、空 = オフ）でゲートウェイを配信し、`web/` の
  ファイルを埋め込む（`//go:embed web`）。オンの間は `@copyover` はオフ（copyover は
  WebSocket を引き渡せない）。

## 課題

1. `server/` で `go get github.com/coder/websocket@v1.8.12`。
2. `mud/web.go` は一部用意されています。`Offer`、`Accept`、`Close`、`ServeHTTP` を書きましょう。
3. `main.go`: `-web` フラグ、`//go:embed web`、ゲートウェイを持つ `http.Server`、
   `go srv.Serve(gw.Listener())`、そして `srv.Shutdown` の後に `web.Shutdown`。

それから `./mudschool run server` を実行し、http://127.0.0.1:8080 を開きましょう。
今はまだテキストだけです。マップはレッスン 41 で来ます。

## 考えてみよう

- なぜ `Offer` は `Accept` を待たなければならない（バッファなしのチャネル）のでしょう？
  バッファありのチャネルだと、停止のときに何が変わる？
- `NetConn` はメッセージをストリームにします。ブラウザが `"lo"` を送り、次に `"ok\n"` を
  送ったら何が起きる？
- TCP 接続は引き渡せるのに、copyover が WebSocket を引き渡せないのはなぜ？
- `/../../etc/passwd` へのリクエストを `http.FileServerFS` はどう扱う？

## チャレンジ

- ログイン直後に、ブラウザの言語から `language` を送る（レッスン 38）。
- ブラウザなしの `ws://`: `websocat` で接続し、端末から遊ぶ。
- copyover を Web でも動かす: HTTP のリスナーもファイルとして引き渡し、Web クライアントには
  自分でつなぎ直してもらう。
