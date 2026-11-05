# Ledtrådar

## Matches

```go
word = strings.ToLower(word)
if strings.ToLower(i.ID) == word {
    return true
}
for _, k := range i.Keywords { ... }
```

`strings.EqualFold(a, b)` jämför också utan att bry sig om skiftläge.

## AddItems

Använd en `switch` med ett `case` per problem och ett `default` som lägger till
föremålet. Samla felen i en slice och `return errors.Join(errs...)`.

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

`drop` är samma sak åt andra hållet.

## Baka in en fil till

```go
//go:embed items.json
var defaultItemsJSON []byte
```

och i `DefaultWorld`, efter `LoadWorld`: `w.LoadItems(bytes.NewReader(defaultItemsJSON))`.
