# Lesson 37 – The tutorial

Phase 6 takes your MUD out into the world, and out there most players are
new. Someone who has never played a MUD logs in, sees a room description and
a blinking cursor – and leaves. The fix is as old as MUDs: a **tutorial**.
New players start in a small area of their own and are guided, step by step:

```
The Hall of Beginnings
A quiet hall with whitewashed walls. ...
[Tutorial 1/6] Welcome! This short tutorial teaches you the basics
(tutorial skip leaves it). Type look to see where you are.
> look
...
[Tutorial 2/6] There is a doorway to the north. Go through it with: go north
> go north
The Courtyard
...
[Tutorial 3/6] Something lies on the ground here. Pick it up with: get torch
```

When the last step is done they step out into the real world.

## Goals

- Describe behaviour as **data**: the steps live in the world file, not in code.
- Reuse the event system from lesson 34 for something new: a hook that
  watches what players *do*, instead of stopping them.
- Change a database that is already in use: a new **migration**.
- Find the rooms you can reach from a room: a **breadth-first search**.

## Go concepts

### Steps as data

```json
"tutorial": {
  "start": "tutorial-hall",
  "steps": [
    {"text": "Type look to see where you are.", "command": "look"},
    {"text": "Go north.", "room": "tutorial-yard"},
    {"text": "Pick it up with: get torch", "item": "tutorial-torch"}
  ],
  "done": "Well done! ..."
}
```

Each step says *what to tell the player* and *when it is done* – after a
command, in a room, or with a thing in your bag. The server knows nothing
about torches or courtyards. Adding a step, translating one or building a
whole new tutorial is a change to a file, not to code.

### A hook that watches

The events from lesson 34 come *before* something happens, so hooks can say
no. The tutorial needs the opposite: to know what has just happened. A new
event, `EventCommand`, comes after every command that exists:

```go
s.emit(&Event{Kind: EventCommand, Player: p, Room: p.Room, Text: cmd.Verb})
```

and the tutorial is just a hook: `s.On(EventCommand, s.onTutorialCommand)`.
Scripts get it for free too: `mud.on("command", ...)`.

### A migration

Lesson 20 promised that the database can grow without being thrown away:

```go
// 4: how far they have come in the tutorial (lesson 37); 0 = not in it
`ALTER TABLE players ADD COLUMN tutorial INTEGER NOT NULL DEFAULT 0`,
```

`DEFAULT 0` is what makes it safe: every player saved before the tutorial
existed gets 0 – "not in the tutorial" – and is never sent back to school.

### The tutorial area

The tutorial's rooms must be closed off: a new player who can walk out of
the tutorial has skipped it by accident. The area is every room you can
reach from the start – a breadth-first search over the exits, with a queue
and a `map[string]bool` of rooms seen. The same area tells the server which
things to put back when a player leaves, so the next newcomer finds the
torch too.

## The contract

- A world file may have a `"tutorial"` object: `start` (a room), `steps`
  (each with `text` and exactly one of `command`, `room`, `item`) and `done`.
  `Tutorial`, `TutorialStep`, `World.Tutorial`; `worldFile.Tutorial`.
- `LoadWorld` checks the tutorial after the items and reports every mistake
  at once: `MsgErrTutorialStart` (then nothing else), `MsgErrTutorialOpen`
  (the world's start can be reached from the tutorial), `MsgErrTutorialEmpty`,
  and per step (numbered from 1) `MsgErrTutorialStep`, `MsgErrTutorialRoom`
  (not in the tutorial area), `MsgErrTutorialItem` (doesn't lie in it).
- `(w *World) LoadTutorial(r)` adds a tutorial file – `rooms`, `items` and
  `tutorial` – to a world; `DefaultWorld` loads `tutorial.json` (given) with it.
- A player with no save (or without a store) starts in the tutorial's start
  room, on step 1, and gets `MsgTutorialStep` (number, count, text) after the
  room description.
- `EventCommand` comes after every command that exists (built-in, script or
  plugin, not an unknown one), with the verb – aliases expanded – in `Text`.
- After every command, steps that are done are passed, several at once if
  need be; the next step is shown. After the last: the `done` text, things
  from the tutorial area go back to their rooms, and the player is moved to
  the world's start room (`MsgDisappears`, `MsgArrives`, a look).
- `tutorial` repeats the current step; `tutorial skip` leaves
  (`MsgTutorialSkipped`); outside the tutorial: `MsgTutorialNotIn`.
- `SavedPlayer.Tutorial` (the step, 0 = not in the tutorial) is saved in a new
  column through migration 4; a player who logs in again continues where they were.
- `worldSnapshot` (lesson 36) keeps the tutorial.

## Task

1. `tutorial.go` is partly given (the types, `tutorialFile`): write
   `LoadTutorial`, `tutorialArea`, `validateTutorial`, `startTutorial`,
   `showTutorialStep`, `stepDone`, `onTutorialCommand`, `leaveTutorial` and
   `cmdTutorial`.
2. `world.go`: `Tutorial` on `World` and `worldFile`, and the check in `LoadWorld`.
3. `events.go` and `dispatch`: `EventCommand`. `server.go`: subscribe the hook,
   start new players in the tutorial, show the step after the first look.
4. `player.go`: `tutorial int`. `store.go`: migration 4, `SavedPlayer.Tutorial`,
   and the column in `SavePlayer`/`LoadPlayer`. `account.go`: snapshot and restore it.
5. `defaultworld.go`: embed `tutorial.json` and `LoadTutorial` it.
   `commands.go`: `tutorial`. `olc.go`: `Tutorial` in `worldSnapshot`.

Then log in with a new name and do the tutorial yourself.

## Think about

- Two new players arrive at the same time, and there is one torch. What
  happens? How do real MUDs solve it? (Look up "instancing".)
- Why does the hook come *after* the command? What would go wrong if the
  tutorial used the cancellable events from lesson 34?
- `DEFAULT 0` makes old players skip the tutorial. When would you want the
  opposite?
- What else could be data instead of code in your MUD?

## Challenges

- A step that is done when the player *says* something specific, or kills a
  creature (`EventDeath`).
- A tutorial in Lua: let scripts add steps with `mud.on("command", ...)`.
- Hints: if a player hasn't finished a step in a minute, repeat it in other words.
