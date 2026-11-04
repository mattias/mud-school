# ヒント

## writeLoop

```go
func (p *Player) writeLoop() {
    defer close(p.done)
    for msg := range p.outbox {
        io.WriteString(p.conn, msg)
    }
}
```

チャネルを作った後、`newPlayer` の中で `go p.writeLoop()` として起動します。

## 決してブロックしない Send

```go
msg := strings.ReplaceAll(text, "\n", "\r\n") + "\r\n"
select {
case p.outbox <- msg:
default:
    p.conn.Close()
}
```

## 安全に閉じる

`close(p.outbox)` の後に `Send` を呼んではいけません。そこで、送信箱を閉じる
*前に*（ロックを取って）プレイヤーを `s.players` から削除し、`broadcast` には送信中
ロックを握らせます。そうすれば、削除後に誰もそのプレイヤーを見つけられません。

## handle での順序

```go
defer conn.Close()   // 最後に実行
p := newPlayer(conn)
defer p.close()      // 最後から 2 番目
...
defer func() {       // 最初に実行
    s.removePlayer(p)
    s.broadcast(p, fmt.Sprintf(MsgLeft, p.Name))
}()
```
