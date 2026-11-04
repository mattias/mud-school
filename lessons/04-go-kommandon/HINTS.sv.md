# Ledtrådar

## Dela upp raden

`strings.Cut(line, " ")` delar vid första mellanslaget och returnerar
`(före, efter, hittades)`:

```go
verb, rest, _ := strings.Cut("get sword from chest", " ")
// verb = "get", rest = "sword from chest"
```

## Förkortningar

```go
if full, ok := aliases[verb]; ok {
    verb = full
}
```

## Apostrof-genvägen

`strings.CutPrefix(line, "'")` returnerar resten och om prefixet fanns.

## help

Samla nycklarna ur mappen i en slice, sortera med `sort.Strings` och sätt
ihop dem med `strings.Join(verbs, MsgListSep)`. Sedan `p.Sendf(MsgCommands, lista)`.

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
