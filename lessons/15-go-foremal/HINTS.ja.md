# ヒント

## Matches

```go
word = strings.ToLower(word)
if strings.ToLower(i.ID) == word {
    return true
}
for _, k := range i.Keywords { ... }
```

`strings.EqualFold(a, b)` でも大文字小文字を区別せずに比較できます。

## AddItems

問題ごとに `case` を 1 つ、アイテムを追加する `default` を持つ `switch` を使いましょう。
エラーはスライスに集めて `return errors.Join(errs...)`。

```go
w.Items[it.ID] = it
room := w.Room(it.Room)
room.Items = append(room.Items, it)
```

## get

```go
it, ok := FindMatch(p.Room.Items, cmd.Rest)
if !ok {
    p.SendKind(KindError, fmt.Sprintf(MsgNoSuchItem, cmd.Rest))
    return
}
p.Room.Items = remove(p.Room.Items, it)
p.Inventory = append(p.Inventory, it)
```

`drop` はその逆です。

## 2 つ目のファイルを埋め込む

```go
//go:embed items.json
var defaultItemsJSON []byte
```

そして `DefaultWorld` の `LoadWorld` の後で: `w.LoadItems(bytes.NewReader(defaultItemsJSON))`.
