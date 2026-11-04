# Hints

## Counting characters

```go
n := utf8.RuneCountInString(name)
switch {
case n < 2:
    return ErrNameTooShort
case n > 16:
    return ErrNameTooLong
}
```

## Letters only

```go
for _, r := range name {
    if !unicode.IsLetter(r) {
        return ErrNameInvalid
    }
}
```

## Upper-case first letter

`utf8.DecodeRuneInString(name)` gives the first rune and how many bytes it
takes. Then: `string(unicode.ToUpper(first)) + strings.ToLower(name[size:])`.

## The login loop

Write a method `func (s *Server) login(p *Player) bool` that loops: send the
question, read a line (return false if `!ok`), validate, normalize, try
`addPlayer`. If everything works: `return true`.

In `handle`:

```go
if !s.login(p) {
    return
}
defer s.removePlayer(p)
p.Sendf(MsgWelcomeName, p.Name)
```
