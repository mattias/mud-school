# レッスン 01 – こんにちは、TCP！

1978 年以来、どの MUD も始まりは同じです。サーバーがポートで待ち受け、
プレイヤーが `telnet` で接続する。このレッスンでは、あなたのサーバーの
最初のバージョンを作ります。あいさつをして、入力したものをそのまま
返す **エコーサーバー** です。

## 目標

- Go のプログラムの構成を理解する: `package`、`import`、`func`。
- **メソッド** を持つ `struct` を使う。
- Go 流のエラー処理: `if err != nil { return err }`。
- TCP で待ち受け、接続を受け付け、行を読む。

## 5 分でわかる Go

```go
package mud                 // すべてのファイルはパッケージに属する

import "fmt"                // 標準ライブラリをインポート

type Server struct{}        // 構造体（フィールドを持つレコード。今はまだ空）

func NewServer() *Server {  // Server へのポインタを返す関数
    return &Server{}
}

func (s *Server) Hello() string {  // *Server の「メソッド」。(s *Server) がレシーバー
    return "こんにちは"
}

func divide(a, b int) (int, error) { // 関数は複数の値を返せる
    if b == 0 {
        return 0, fmt.Errorf("%d を 0 で割ることはできません", a)
    }
    return a / b, nil               // nil は「エラーなし」
}
```

Go には例外がありません。エラーは `error` 型の普通の値で、返して確認する
ものです。最初は冗長に感じますが、どこで失敗しうるかがとても明確になります。

`defer` は関数が戻るときに実行される呼び出しを予約します。後片付けに最適です:

```go
func (s *Server) handle(conn net.Conn) {
    defer conn.Close()   // どんな形で関数を抜けても実行される
    ...
}
```

## Go でのネットワーク

```go
l, err := net.Listen("tcp", ":4000") // 待ち受け（main.go ですでに実行済み）
conn, err := l.Accept()              // 次のクライアントを待つ – ブロックする！
fmt.Fprint(conn, "hi\r\n")           // conn は io.Writer であり ...
scanner := bufio.NewScanner(conn)    // ... io.Reader でもある
for scanner.Scan() {                 // 1 行ずつ読む
    line := scanner.Text()           // 改行を除いた行
}
```

> **なぜ `\r\n`？** telnet プロトコル（RFC 854）では改行は CR LF と決まって
> います。多くのクライアントは `\n` でも動きますが、最初から正しくやりましょう。

## 課題

`workspace/server/mud/server.go` を開き、2 つの `TODO` を埋めてください:

1. `Serve` は無限 `for` ループで `l.Accept()` を呼びます。エラーが返ったら
   `Serve` はそれを返します（例えばリスナーが閉じられたとき）。そうでなければ
   `s.handle(conn)` を呼びます。
2. `handle` は `defer` で接続を閉じ、`Welcome` + `"\r\n"` を送り、
   その後読んだ行をすべてそのまま送り返します。

> **`messages.go` について:** ゲームがプレイヤーに表示するテキストはすべて
> `workspace/server/mud/messages.go` にあり、コースで選んだ言語になって
> います。`Welcome` もすでにあります。テストも同じ定数を使うので、名前は
> 変えずに、テキストは自由に変えてかまいません。コードとコメントは業界の
> 慣習に合わせて英語、説明はあなたの言語です。

`./mudschool check` でテストします。

## 自分で試す

```
./mudschool run server            # 端末 1
telnet localhost 4000             # 端末 2（または: nc localhost 4000）
```

何行か入力してみましょう。次に *2 つ目* の telnet を同時に開きます。
何が起きますか？ なぜでしょう？（次のレッスンで直します。）

## チャレンジ（任意）

- 新しい接続を `log.Printf("new connection from %s", conn.RemoteAddr())` で記録する。
- クライアントが `quit` と打ったら `MsgGoodbye` を返して切断する。
- ドキュメントを読む: `go doc net.Listener` と `go doc bufio.Scanner`。
