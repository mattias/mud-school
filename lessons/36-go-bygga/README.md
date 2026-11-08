# Lesson 36 – Building in the game

The last lesson. Since lesson 31 you can change the world by editing a JSON
file. MUDs have long had something better: **online creation**. Builders
walk around inside the world and shape it from there:

```
> @dig north Dark cave
You dig north and create "Dark cave" (dark-cave).
> go north
Dark cave
An empty room. Describe it with @desc.
> @desc Water drips from the ceiling. Something glitters in the dark.
Done.
```

Every change is **validated** with the rules from lesson 9, **saved** to the
world file, and **visible to everyone at once** – no restart, no reload, no
editor.

## Goals

- Change live state safely: apply, validate, and **undo** if the result is invalid.
- Write a file **atomically**: a temporary file and a rename, so a crash or a
  reader never sees half a file.
- Keep slow I/O out of the game loop – and keep concurrent saves in order.
- Recognise your own writes, so the file watcher from lesson 31 doesn't reload them.

## Go concepts

### Apply, validate, undo

```go
func (s *Server) change(fn, undo func()) error {
	fn()
	if err := s.world.Validate(); err != nil {
		undo()
		return err
	}
	return nil
}
```

Two closures: one that makes the change and one that takes it back. The
world's own validation (exits that match coordinates, no two rooms in the
same place, …) decides – the building commands don't need rules of their own.

### Atomic writes

```go
tmp, _ := os.CreateTemp(filepath.Dir(path), ".world.json.*")
tmp.Write(data)
tmp.Close()
os.Rename(tmp.Name(), path) // atomic on the same file system
```

Anyone who opens the file sees either the old version or the new one,
never something in between.

### Saves in order

The JSON is made **in** the game loop (a consistent snapshot); the disk is
written **outside** it, in a goroutine. Two quick changes give two goroutines –
and the older snapshot could be written last. A sequence number fixes that:
each snapshot gets the next number, and a goroutine that finds that a newer
one has already been written gives up.

### Not reloading yourself

The watcher from lesson 31 sees every change to the file – including ours. So
after writing, the server remembers the file's modification time and size
(`ownWrite`), and the watcher skips a change that matches it. A mutex
(`saveMu`) makes sure the watcher never looks in the middle of a save.

## The contract

All for administrators only (otherwise `MsgNotAdmin`), in the room you stand in:

- `@dig <direction> <name>` – a new room one step in that direction (the
  coordinates from lesson 9), with the same terrain, the description
  `MsgNewRoomDesc` and exits both ways; its ID is made from the name
  (`"Dark cave"` → `dark-cave`, unique). `MsgDug`. If there already is an exit
  that way: `MsgExitExists`; without a name: `MsgDigUsage`.
- `@desc <text>` and `@name <text>` – the room's description and name; `MsgBuilt`.
- `@exit <direction> <room id>` – an exit (one way); `@exit <direction> none`
  removes it; `MsgBuilt`.
- A change that would make the world invalid is undone; the builder gets
  `MsgBuildInvalid` with the validation errors.
- Rooms whose exits change get a new zone; everything in them keeps a walkable position.
- With a world file (`Options.WorldPath`) the whole world is saved to it after
  every change – atomically, in order, outside the game loop; the watcher doesn't
  reload the server's own saves; a stopping server waits for the saves still
  being written. Without one: `MsgNotSaved`.

## Task

`olc.go` is partly given (`slug`, `newRoomID`, `change`, `rezone`, `cmdDesc`,
`cmdName`, `worldSnapshot`, `finishSaves`, `writeFileAtomic`): write `cmdDig`,
`cmdExit` and `saveWorld`. Then the fields on `Server` (`saves` is a
`sync.WaitGroup`), the commands, the check in `watchWorld`, and in `run`: call
`s.finishSaves()` when the loop stops, so that a change made just before Ctrl+C
(or a copyover) is not lost. While you are there, close the Lua state
(`s.scripts.L.Close()`) and the plugins (`s.plugins.Close()`) too.

Build a little area of your own: start the server with `-world my-world.json
-admin YourName`, dig, describe, and look at the file afterwards.

## Think about

- Why is the JSON made in the game loop but written outside it?
- `@dig` uses the world's validation instead of checking the coordinates
  itself. What are the advantages? Are there disadvantages?
- What happens to the procedurally generated zone of a room when you add an exit?
- Two administrators build at the same time. What can go wrong? What can't?

## Challenges

- `@undo`: keep the last ten changes (the undo closures) and take them back.
- `@item` and `@npc`: create items and creatures from inside the game.
- Builders: a role between players and administrators who may only build in their own area.
- `@zone`: draw a hand-made zone in the game, row by row.

**That was the last lesson.** You have built a multiplayer game server in Go
and a client in Rust – from an echo server to a tile world with light, a mouse,
saving, scripts, plugins in another language, and a server that can be
changed in every way without anyone being disconnected. Phase 6 in
`docs/*/CURRICULUM.md` has ideas for where to go next.
