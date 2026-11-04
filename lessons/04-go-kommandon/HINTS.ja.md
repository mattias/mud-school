# ヒント

## 行を分割する

`strings.Cut(line, " ")` は最初の空白で分割し、`(前, 後, 見つかったか)` を返します:

```go
verb, rest, _ := strings.Cut("get sword from chest", " ")
// verb = "get", rest = "sword from chest"
```

## 省略形

```go
if full, ok := aliases[verb]; ok {
    verb = full
}
```

## アポストロフィの近道

`strings.CutPrefix(line, "'")` は残りの文字列と、接頭辞があったかどうかを返します。

## help

マップのキーをスライスに集め、`sort.Strings` でソートし、
`strings.Join(verbs, ", ")` でつなげます。それから `p.Sendf(MsgCommands, list)`。

## handle

```go
p := newPlayer(conn)
p.Send(Welcome)
for !p.quit {
    line, ok := p.readLine()
    if !ok {
        return
    }
    s.dispatch(p, line)
}
```
