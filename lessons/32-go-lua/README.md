# Lesson 32 – Scripts in Lua

Reloading data (lesson 31) changes *what* the world contains. To change *what
happens* – a new command, a creature that behaves differently – you have had
to write Go and restart. Now the server gets an **embedded scripting
language**: commands and behaviours in `.lua` files, reloaded the moment you
save them.

```lua
-- scripts/dance.lua
mud.command("dance", function(player, args)
  mud.send(player, player .. " dances " .. args .. "!")
end)

mud.behavior("bark", function(npc, room)
  if mud.roll(1, 20) == 1 then mud.broadcast(room, "The dog barks.") end
end)
```

[Lua](https://www.lua.org) is *the* classic language for scripting games
(World of Warcraft, Roblox, Factorio, …). We embed it with
**gopher-lua** – a Lua interpreter written in pure Go, so no C compiler is needed.

## Goals

- Embed an interpreter and expose a Go API to it (`mud.send`, `mud.player`, …).
- **Sandbox** untrusted code: no files, no programs, a time limit.
- Use `context.Context` to stop code that runs too long.
- Think about API design: what should a script be able to do – and not?

## Go concepts

### Calling Go from Lua

A Go function with the signature `func(L *lua.LState) int` can be called from
Lua: it reads its arguments from the Lua stack (`L.CheckString(1)`), pushes
its results (`L.Push(...)`), and returns how many it pushed.

```go
"send": func(L *lua.LState) int {
	if p := s.players[NormalizeName(L.CheckString(1))]; p != nil {
		p.SendKind(KindInfo, L.CheckString(2))
	}
	return 0
},
```

`L.SetFuncs(table, map)` puts a whole map of them in a table, and
`L.SetGlobal("mud", table)` makes it the global `mud`.

### The sandbox

`lua.NewState(lua.Options{SkipOpenLibs: true})` starts with *nothing*. Then
open only what is safe – `base`, `table`, `string`, `math` – and remove the
functions in `base` that read files or load code (`dofile`, `loadfile`,
`load`, `loadstring`, `require`). No `os`, no `io`: a script can't delete
files or start programs.

### A time limit

```go
ctx, cancel := context.WithTimeout(context.Background(), ScriptTimeout)
defer cancel()
L.SetContext(ctx)
defer L.RemoveContext()
err := L.CallByParam(lua.P{Fn: fn, NRet: 0, Protect: true}, args...)
```

With a context, gopher-lua checks between instructions whether time is up and
stops the script with an error – even `while true do end`. `Protect: true`
turns Lua errors into a Go `error` instead of a panic.

### Who owns the Lua state?

An `LState` is not safe for several goroutines – just like the world. So it
belongs to the game loop: scripts are loaded and run only inside it. A script
can then call `mud.send` and read players without any locks.

## The contract

- `Options.ScriptDir` – a folder with `.lua` files (`""` = no scripts). They are
  loaded when the server is created, in name order, and reloaded (in the game
  loop) when a file is added, removed or changed – checked every `ReloadEvery`.
- A broken file is skipped; the others still load; logged-in administrators get
  `MsgScriptLoadFailed` with the error, which names the file.
- The sandbox: `os`, `io`, `dofile`, `loadfile`, `load`, `loadstring` and
  `require` are `nil`.
- The API (a global table `mud`):
  - `mud.command(name, function(player, args))` – a new command;
  - `mud.behavior(name, function(npc, room))` – a behaviour, used in the NPC
    files as `{"type": "script", "script": "name"}` (`BehaviorSpec.Script`);
  - `mud.send(player, text)`, `mud.broadcast(room, text)`;
  - `mud.player(name)` → `{name, room, hp, max_hp}` or `nil`;
  - `mud.players(room)` → a list of names; `mud.roll(a, b)` → a random number from a to b.
- Built-in commands come first: a script can *add* commands, not replace them.
- Every call has a time limit of `ScriptTimeout` (50 ms); a script that fails or
  runs too long gives the player `MsgScriptError`, and the server keeps running.
- `main.go`: `-scripts scripts`.

## Task

`go get github.com/yuin/gopher-lua@v1.1.1` first. `scripts.go` is partly given
(`newScripts` with the API, `loadScripts`, `ScriptBehavior`, `watchScripts`):
write `run`, `LoadScripts` and `runScriptCommand`. Then the option, the fields,
`dispatch`, the behaviour type and the flag. `server/scripts/hello.lua` is an
example.

## Think about

- Why can scripts add commands but not replace `look`? When would replacing be useful?
- What could a malicious script still do? (Memory? `string.rep("x", 1e9)`?)
- The API gives scripts names (strings), not Go pointers. Why?
- Lua runs in the game loop. A script that takes 40 ms – what does that do to
  the tick rate? What would you do about it?
- A script can register `@anything` – and `adminOnly` only guards the built-in
  commands. Should scripts be allowed `@` commands at all?

## Challenges

- `mud.on_tick(function() ... end)` – scripts that do something every tick.
- `mud.give(player, item)` and `mud.move(player, room)` – what checks do they need?
- Let builders (not only administrators) upload scripts with an `@script` command.
- A memory limit: count the size of tables a script creates.
