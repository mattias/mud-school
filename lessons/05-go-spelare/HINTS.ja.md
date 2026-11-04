# ヒント

## 文字数を数える

```go
n := utf8.RuneCountInString(name)
switch {
case n < 2:
    return ErrNameTooShort
case n > 16:
    return ErrNameTooLong
}
```

## 文字だけ

```go
for _, r := range name {
    if !unicode.IsLetter(r) {
        return ErrNameInvalid
    }
}
```

## 先頭を大文字に

`utf8.DecodeRuneInString(name)` は最初のルーンと、それが何バイトかを返します。
あとは `string(unicode.ToUpper(first)) + strings.ToLower(name[size:])`。

## ログインのループ

ループするメソッド `func (s *Server) login(p *Player) bool` を書きます: 質問を送り、
1 行読み（`!ok` なら false を返す）、検証し、正規化し、`addPlayer` を試す。
すべてうまくいけば `return true`。

`handle` では:

```go
if !s.login(p) {
    return
}
defer s.removePlayer(p)
p.Sendf(MsgWelcomeName, p.Name)
```
