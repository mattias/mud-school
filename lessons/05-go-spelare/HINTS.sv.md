# Ledtrådar

## Räkna tecken

```go
n := utf8.RuneCountInString(name)
switch {
case n < 2:
    return ErrNameTooShort
case n > 16:
    return ErrNameTooLong
}
```

## Bara bokstäver

```go
for _, r := range name {
    if !unicode.IsLetter(r) {
        return ErrNameInvalid
    }
}
```

## Stor första bokstav

`utf8.DecodeRuneInString(name)` ger första runan och hur många bytes den tar.
Sedan: `string(unicode.ToUpper(first)) + strings.ToLower(name[size:])`.

## Inloggningsloopen

Skriv en metod `func (s *Server) login(p *Player) bool` som loopar:
skicka frågan, läs en rad (returnera false om `!ok`), validera, normalisera,
försök `addPlayer`. Lyckas allt: `return true`.

I `handle`:

```go
if !s.login(p) {
    return
}
defer s.removePlayer(p)
p.Sendf(MsgWelcomeName, p.Name)
```
