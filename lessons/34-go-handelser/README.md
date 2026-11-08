# Lesson 34 – Events and hooks

Scripts (lesson 32) and plugins (lesson 33) can add *new* commands. But many
ideas are about changing what *existing* actions do: a door that is locked
for everyone except the banker, a curse that makes you speak backwards, a
counter of how many rats have died today. For that the server announces what
is about to happen – an **event** – and **hooks** may react, or say no.

```
"go north"  ─▶  travel  ─▶  emit(Event{Kind: "enter", Player: Anna, To: vault})
                                │
                 ┌──────────────┼───────────────────┐
             Go hook        Lua hook             (plugin hook – a challenge)
                 │              │
         e.Cancel("The vault door is locked.")  ──▶  Anna stays; she is told why
```

## Goals

- Build an **event bus**: subscribers per kind of event, called in order.
- Let a hook **cancel** an action (and say why) or **change** it (what is said).
- Call Lua with a table and read its return value.
- Decouple: the code that moves players doesn't know about vaults.

## Go concepts

### Functions as subscribers

```go
type Hook func(e *Event)

func (s *Server) On(kind string, h Hook) {
	s.hooks[kind] = append(s.hooks[kind], h)
}
```

Because hooks get a *pointer* to the event, they can change it: set
`e.Text`, or call `e.Cancel(reason)`. The code that emitted the event looks
at the result:

```go
e := &Event{Kind: EventEnter, Player: p, Room: p.Room, To: to, Dir: dir}
if !s.emit(e) {
	e.refuse()   // tells the player e's reason
	return
}
```

### Lua hooks with a return value

```lua
mud.on("enter", function(ev)
  if ev.to == "vault" then return "A voice says: not today." end
end)
```

The event becomes a Lua table (`kind`, `player`, `npc`, `room`, `to`, `dir`,
`text`); a returned string cancels. `CallByParam` with `NRet: 1` leaves the
result on the Lua stack: `L.Get(-1)` reads it, `L.Pop(1)` removes it.

## The contract

- `Event {Kind, Player, NPC, Room, To, Dir, Text}` with `Cancel(reason)` and
  `Cancelled()`; `Hook func(*Event)`; `(s *Server) On(kind, hook)` – call it
  before the server starts or inside the game loop.
- Kinds:
  - `EventEnter` – before `travel` (also `go`, flee and walking out of a zone):
    `Player`, `Room`, `To`, `Dir`. Cancelled: the player stays and gets the reason.
  - `EventSay` – before `say`: `Player`, `Room`, `Text`. Hooks may change
    `Text` (what is then said); cancelled: nothing is said, the player gets the reason.
  - `EventTick` – after every tick (after combat and respawn).
  - `EventDeath` – when a creature is defeated (`NPC`, `Player` = the killer,
    `Room`) and when a player dies (`Player`, `Room`).
- `emit(e) bool` runs the Go hooks, then the Lua hooks, in the order they were
  added, and stops at the first that cancels; `false` = cancelled.
- A cancel without a reason gives the player `MsgCancelled`.
- Lua: `mud.on(kind, function(event) ... end)`; a returned string cancels with
  that reason. Lua hooks are reloaded with the scripts.

## Task

`events.go` is partly given (`Event`, `Cancel`, `On`, `refuse`): write `emit`
and `runHook`. Add `mud.on` and `Scripts.hooks` to `scripts.go` and split
`run` into a `call` that can return a value. Then emit the events in `travel`,
`cmdSay`, `onTick`, `defeat` and `die`.

Try it with a script:

```lua
-- scripts/curse.lua
mud.on("say", function(ev)
  if ev.player == "Bertil" then return "Your tongue is tied!" end
end)
```

## Think about

- Hooks run in the order they were added. When does the order matter? Should
  hooks have priorities?
- `EventSay` lets Go hooks change the text, but Lua hooks can only cancel. How
  would you let Lua change it too?
- What happens if a hook for `enter` itself moves the player (and so emits
  another `enter`)?
- Which other events would be useful? (`get`, `drop`, `attack`, `login` …)

## Challenges

- `on_event` for WebAssembly plugins (ABI version 2 – and keep loading version 1).
- `EventGet` with a cancellable "the sword is stuck in the stone".
- A hook that keeps statistics (rats killed per player) and a `stats` command that shows them.
