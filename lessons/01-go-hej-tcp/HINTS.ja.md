# ヒント

## 無限ループ

Go のループ文は `for` だけです。無限ループは `for { ... }` と書きます。

```go
for {
    conn, err := l.Accept()
    if err != nil {
        return err
    }
    // ...
}
```

## 受け取ったリスナーを使う

`Serve` はすでに待ち受けているリスナー `l` を受け取ります。`Serve` の中で
`net.Listen` を呼ばないでください。ループの中で `l, err := ...` と書くのもだめです。
受け取った `l` を隠す新しい `l` ができてしまいます。`conn, err := l.Accept()` と書きましょう。

## 接続への書き込み

`net.Conn` は `io.Writer` インターフェースを満たすので、
`fmt.Fprint(conn, "text\r\n")` や `conn.Write([]byte("text\r\n"))` が使えます。
ファイルの先頭で `"fmt"` と `"bufio"` をインポートするのを忘れずに。

## 行の読み込み

```go
scanner := bufio.NewScanner(conn)
for scanner.Scan() {
    line := scanner.Text()
    fmt.Fprint(conn, line+"\r\n")
}
```

クライアントが切断すると `scanner.Scan()` は `false` を返します。するとループが
終わり、関数が戻り、`defer conn.Close()` が実行されます。
