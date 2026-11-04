# Lesson 05 – Players

Time to know *who* is playing. Players choose a name when they connect, and
the server keeps track of everyone logged in. It is the first time several
goroutines share the same data – and therefore the first time you can write a
truly sneaky bug: a **data race**.

## Goals

- Create and compare errors with `errors.New`, `errors.Is` and `%w`.
- Understand the difference between *bytes* and *runes* (Unicode characters) in Go strings.
- Protect shared data with `sync.Mutex`.
- See the race detector find a bug.

## The login flow

```
Welcome to MUD School!
What is your name?
> x
Invalid name: the name must be at least 2 characters.
What is your name?
> anna
Welcome, Anna!
> who
Players online (2): Anna, Bertil
```

## Go concepts

### Errors are values – your own ones too

```go
var ErrNameTooShort = errors.New(MsgErrNameTooShort) // the text is in messages.go

func ValidateName(name string) error {
    if len(name) < 2 { return ErrNameTooShort }  // (but len counts bytes, see below!)
    return nil
}

err := fmt.Errorf("%w: %s", ErrNameTaken, name) // "wrap" an error with more information
errors.Is(err, ErrNameTaken)                    // true – Is looks through the wrapping
```

### Bytes and runes

Go strings are UTF-8 bytes. `len("Ö")` is **2**, and `len("空")` is **3**. To
count characters use `utf8.RuneCountInString`, and `for _, r := range s` loops
over runes (`rune` = one Unicode character), not bytes.

```go
for i, r := range "Åsa" {
    fmt.Println(i, string(r), unicode.IsLetter(r)) // 0 Å true, 2 s true, 3 a true
}
```

`unicode.IsLetter` also says yes to Japanese kana and kanji – so `さくら` is a
valid name.

### Mutex

A `sync.Mutex` makes sure only one goroutine at a time runs the code between
`Lock` and `Unlock`:

```go
type Server struct {
    mu      sync.Mutex          // protects players
    players map[string]*Player
}

func (s *Server) addPlayer(p *Player) error {
    s.mu.Lock()
    defer s.mu.Unlock()        // always unlock, even on an early return
    ...
}
```

Maps in Go are **not** safe to use from several goroutines at once. If you do
it without a lock the program can crash with `fatal error: concurrent map writes`.

## Task

1. **`names.go`** (new file): write `ValidateName` and `NormalizeName`.
2. **`server.go`**:
   - add `mu sync.Mutex` and `players map[string]*Player` to `Server`,
   - write `addPlayer(p) error` (returns `ErrNameTaken` if the name exists),
     `removePlayer(p)` and `PlayerNames() []string` (sorted),
   - in `handle`: after `Welcome`, ask `MsgAskName` until the player gives a
     valid, free name. On error send `MsgInvalidName` with the error
     (`p.Sendf(MsgInvalidName, err)`) and ask again. Then send
     `MsgWelcomeName` with the name,
   - remove the player when they disconnect (hint: `defer`).
3. **`commands.go`**: add `who`, which sends `MsgWho` with the count and the
   names, sorted and separated by `MsgListSep`:
   `Players online (2): Anna, Bertil`.

All the texts (`MsgAskName`, `MsgErrNameTooShort`, ...) are in `messages.go`.
The error variables in `names.go` are already created from them.

## Try it yourself: find a data race

Temporarily remove `s.mu.Lock()` and `defer s.mu.Unlock()` in `addPlayer` and run

```
./mudschool check --race
```

Read the report – it shows exactly which two goroutines collided and where.
Then put the lock back.

## Think about

- Why do we normalize the name *before* checking whether it is taken?
- What happens if a player disconnects in the middle of logging in?

## Challenges

- Store passwords (hashed with `golang.org/x/crypto/bcrypt`) so nobody else
  can take your name. This comes back in phase 3 when we save players.
- Forbid some names (`admin`, `god`, ...) with a new error `ErrNameReserved`.
