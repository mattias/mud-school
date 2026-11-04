# Lesson 04 – Commands

An echo server is not a MUD. Now the server has to understand what the
player types: `look`, `north`, `get sword`, `say hello`. We start with a
**parser** and a **command registry** – and the registry is the seed of
something big: in phase 5, scripts and plugins will be able to add commands
while the server is running.

## Goals

- Work with strings (`strings.Fields`, `strings.Cut`, `strings.ToLower`).
- Use **maps** and **slices**.
- Understand that functions are values: `map[string]CommandFunc`.
- Write and read **table-driven tests** – the Go world's favourite pattern.

## New files

The course tool has created three files for you in `workspace/server/mud/`:

| File | Contents |
|---|---|
| `command.go` | `Command`, `ParseCommand` and abbreviations – **you fill in** |
| `commands.go` | the command registry and the built-in commands – **you fill in** |
| `player.go` | `Player`: a connected player with `Send`/`Sendf` – done, read through it |

## Go concepts

### Maps

```go
ages := map[string]int{"anna": 31}
ages["bertil"] = 40
age, ok := ages["cecilia"]   // ok == false if the key is missing, age becomes 0
for name, age := range ages { ... }  // the order is RANDOM!
```

### Slices

```go
words := strings.Fields("  get  the sword ")  // []string{"get", "the", "sword"}
first, rest := words[0], words[1:]
names := make([]string, 0, 10)                // length 0, capacity 10
names = append(names, "anna")
sort.Strings(names)
```

### Functions as values

```go
type CommandFunc func(s *Server, p *Player, cmd Command)

s.commands["dance"] = func(s *Server, p *Player, cmd Command) {
    p.Send("You dance!")
}
```

### Table-driven tests

Open `mudschool_04_test.go` and look at `TestParseCommand`. Every row in the
table is a test case. To test something new, you just add a row. You can run
a single test directly with Go:

```
cd workspace/server && go test ./mud -run TestParseCommand -v
```

## Task

1. **`command.go`**: fill in `aliases` and write `ParseCommand`:
   - trim spaces at the start and end; an empty line gives `Command{}`,
   - the first word is the verb, in lower case, with abbreviations expanded,
   - `Args` are the remaining words, `Rest` is the rest of the line (trimmed),
   - a line starting with `'` means `say`: `'hi` → `say` with `Rest` `"hi"`.
2. **`server.go`**: give `Server` a field `commands map[string]CommandFunc`.
   Create the map in `NewServer` and call `s.registerBuiltins()`.
   Rewrite `handle`: create a `Player` with `newPlayer(conn)`, send `Welcome`,
   then read lines and call `s.dispatch(p, line)` until the player
   disconnects or `p.quit` becomes true.
3. **`commands.go`**: write `dispatch`, `cmdHelp` and `cmdQuit`:
   - `help` sends `MsgCommands` with all registered verbs, sorted and separated
     by `MsgListSep` – e.g. `Commands: echo, help, quit`,
   - `quit` sends `MsgGoodbye` (`Goodbye!`) and leaves,
   - an unknown verb gives `MsgUnknownCommand`:
     `p.Sendf(MsgUnknownCommand, cmd.Verb)` → `Unknown command "fly". Type help for help.`
     (`%q` in the format string adds the quotes).

## Try it yourself

```
telnet localhost 4000
help
echo   hello  there
L
q
```

## Think about

- Why do we sort the verbs in `help`? (Run it a few times without sorting.)
- `Server.commands` is read by many goroutines at once. Is that safe?
  (Yes – as long as nobody *writes* to the map after the server has started.
  Remember this question for phase 5, when we want to add commands to a
  running server.)

## Challenges

- Let `help <command>` show a description. You can replace `CommandFunc` with
  a struct: `type commandDef struct { fn CommandFunc; help string }`.
- Support prefixes: `hel` should match `help` if no other command starts that way.
