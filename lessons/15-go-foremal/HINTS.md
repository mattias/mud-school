# Hints

## Matches

```go
word = strings.ToLower(word)
if strings.ToLower(i.ID) == word {
    return true
}
for _, k := range i.Keywords { ... }
```

`strings.EqualFold(a, b)` compares without caring about case, too.

## AddItems

Use a `switch` with one `case` per problem and a `default` that adds the
item. Collect the errors in a slice and `return errors.Join(errs...)`.

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

`drop` is the same the other way round.

## Embedding a second file

```go
//go:embed items.json
var defaultItemsJSON []byte
```

and in `DefaultWorld`, after `LoadWorld`: `w.LoadItems(bytes.NewReader(defaultItemsJSON))`.
