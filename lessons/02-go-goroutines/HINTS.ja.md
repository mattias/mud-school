# ヒント

## 1 語だけ

`Serve` の中の `s.handle(conn)` の前に `go` を付けます。

## ログ

```go
log.Printf("new connection from %s", conn.RemoteAddr())
defer log.Printf("%s disconnected", conn.RemoteAddr())
```

同じ関数内の 2 つの `defer` は逆順（後入れ先出し）で実行されます。
