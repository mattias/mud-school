# Hints

## Splitting the line

`strings.Cut(line, " ")` splits at the first space and returns
`(before, after, found)`:

```go
verb, rest, _ := strings.Cut("get sword from chest", " ")
// verb = "get", rest = "sword from chest"
```

## Abbreviations

```go
if full, ok := aliases[verb]; ok {
    verb = full
}
```

## The apostrophe shortcut

`strings.CutPrefix(line, "'")` returns the rest and whether the prefix was there.

## help

Collect the keys of the map in a slice, sort it with `sort.Strings` and join
it with `strings.Join(verbs, MsgListSep)`. Then `p.Sendf(MsgCommands, list)`.

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
