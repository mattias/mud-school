# Lesson 31 – Reloading the world

Phase 5 begins. Its goal: **the server should never need a restart** to get
new content or new features. Players stay connected while the world changes
around them. There are four levels, from the simplest to the most powerful:

```
 data (JSON, reloaded live)  ->  scripts (Lua)  ->  plugins (WASM)  ->  a new binary (copyover)
   lesson 31, 36                  lessons 32, 34     lesson 33           lesson 35
```

This lesson is the first: edit `world.json` – fix a typo, add a room – save,
and a second later the world shimmers and everyone stands in the new one.

## Goals

- Watch a file without a library: **polling** with `os.Stat` and a `time.Ticker`.
- Do slow work (reading and validating a file) **outside** the game loop and
  the swap **inside** it – atomically, so nobody ever sees half a world.
- Migrate live state from an old structure to a new one (players, positions, items).
- Protect commands with a simple permission check: **administrators**.

## Go concepts

### Polling

```go
t := time.NewTicker(s.reloadEvery)
defer t.Stop()
for {
	select {
	case <-s.stop:
		return
	case <-t.C:
		fi, err := os.Stat(path)  // has ModTime or Size changed?
	}
}
```

Libraries like `fsnotify` get told by the operating system instead of
asking, but polling every second or two is simple, portable and cheap
enough for one file. Note the `select` on `s.stop`: the goroutine ends with
the server – no leaks.

### Validate first, swap later

```go
nw, err := LoadWorldFile(path)   // outside the loop: disk + validation (lesson 9)
s.do(func() {                    // inside the loop: a quick swap
	if err != nil { /* tell the admins, keep the old world */ return }
	s.swapWorld(nw)
})
```

If the file is broken, *nothing* changes – the administrators get the error
messages from lesson 9 and can fix the file.

### Migrating state

The new world has new `*Room`, `*Item` and `*NPC` values. Everything that
pointed into the old world must point into the new one:

- players: the room with the same ID (or the start room if it is gone), a
  position that is still walkable;
- their things: the new world's item with the same ID, taken out of its room;
- fights: cleared – the creatures are new;
- the zone: `seenRoom` and `explored` are forgotten, so clients get a new zone frame.

### Commands with `@`

Administrators' commands start with `@`. A wrapper checks the permission so
the commands themselves don't have to:

```go
func adminOnly(fn CommandFunc) CommandFunc {
	return func(s *Server, p *Player, cmd Command) {
		if !s.isAdmin(p) { p.SendKind(KindError, MsgNotAdmin); return }
		fn(s, p, cmd)
	}
}
```

A function that takes a function and returns a function – the same idea as
`moveCommand` in lesson 8.

## The contract

- `Options.WorldPath` (the world file, `""` = none), `Options.ReloadEvery`
  (0 = don't watch), `Options.Admins` (names, normalized like player names).
- When the file changes (ModTime or Size), it is loaded and validated outside
  the game loop:
  - valid: the world is swapped; every player gets `MsgReloaded` and a `look`;
    a player whose room is gone gets `MsgRoomGone` and ends up in the start room;
    carried items are replaced by the new world's items with the same ID (which
    no longer lie in their rooms); fights stop;
  - invalid: nothing changes; logged-in administrators get `MsgReloadFailed` with the error.
- `@reload` (administrators only, otherwise `MsgNotAdmin`; `MsgNoWorldFile`
  without a world file) does the same at once.
- `main.go`: `-reload 2s` and `-admin Anna,Bertil`.

## Task

`reload.go` is partly given (`isAdmin`, `adminOnly`, `notifyAdmins`,
`cmdReload`, `normalizeAdmins`): write `reloadFromFile`, `watchWorld` and
`swapWorld`. Then the options, the fields on `Server`, starting the watcher
in `start`, the command and the flags.

Try it: copy the world (`cp server/mud/world.json my-world.json`), start the
server with `-world my-world.json -admin YourName`, log in, and change a
description in the file.

## Think about

- Why must the swap happen in the game loop, but the reading outside it?
- A player stands on a cell that is a wall in the new zone. What happens? What
  *should* happen?
- What if the file is saved in the middle of the poll – half written? (The
  tests write a temporary file and rename it. Why does that help?)
- Two reloads in quick succession – can they be applied in the wrong order?
- An administrator is only a *name*. Without a database (`-db ""`), or before
  Anna has registered, anyone can log in as Anna – and lessons 35–36 give
  administrators `@copyover` and the right to write the world file. Run with
  passwords (lesson 20) and register the admin names first. What else would you want?

## Challenges

- Reload `items.json` and `npcs.json` separately, so creatures that are in the
  middle of something keep going.
- `@reload` shows a diff: which rooms were added, removed and changed.
- Use `fsnotify` instead of polling and compare.
