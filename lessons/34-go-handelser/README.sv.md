# Lektion 34 – Händelser och hooks

Skript (lektion 32) och plugins (lektion 33) kan lägga till *nya*
kommandon. Men många idéer handlar om att ändra vad *befintliga* handlingar
gör: en dörr som är låst för alla utom bankiren, en förbannelse som får dig
att tala baklänges, en räknare över hur många råttor som dött i dag. För det
talar servern om vad som är på väg att hända – en **händelse** – och
**hooks** kan reagera, eller säga nej.

```
"go north"  ─▶  travel  ─▶  emit(Event{Kind: "enter", Player: Anna, To: vault})
                                │
                 ┌──────────────┼───────────────────┐
             Go-hook        Lua-hook             (plugin-hook – en utmaning)
                 │              │
         e.Cancel("Valvdörren är låst.")  ──▶  Anna stannar; hon får veta varför
```

## Mål

- Bygga en **händelsebuss**: prenumeranter per sorts händelse, anropade i ordning.
- Låta en hook **avbryta** en handling (och säga varför) eller **ändra** den (det som sägs).
- Anropa Lua med en tabell och läsa returvärdet.
- Frikoppla: koden som flyttar spelare vet inget om valv.

## Go-koncept

### Funktioner som prenumeranter

```go
type Hook func(e *Event)

func (s *Server) On(kind string, h Hook) {
	s.hooks[kind] = append(s.hooks[kind], h)
}
```

Eftersom hooks får en *pekare* till händelsen kan de ändra den: sätta
`e.Text`, eller anropa `e.Cancel(reason)`. Koden som skickade händelsen
tittar på resultatet:

```go
e := &Event{Kind: EventEnter, Player: p, Room: p.Room, To: to, Dir: dir}
if !s.emit(e) {
	e.refuse()   // berättar e:s anledning för spelaren
	return
}
```

### Lua-hooks med returvärde

```lua
mud.on("enter", function(ev)
  if ev.to == "vault" then return "A voice says: not today." end
end)
```

Händelsen blir en Lua-tabell (`kind`, `player`, `npc`, `room`, `to`, `dir`,
`text`); en returnerad sträng avbryter. `CallByParam` med `NRet: 1` lämnar
resultatet på Lua-stacken: `L.Get(-1)` läser det, `L.Pop(1)` tar bort det.

## Kontraktet

- `Event {Kind, Player, NPC, Room, To, Dir, Text}` med `Cancel(reason)` och
  `Cancelled()`; `Hook func(*Event)`; `(s *Server) On(kind, hook)` – anropa
  den innan servern startar eller inne i spelloopen.
- Sorter:
  - `EventEnter` – före `travel` (även `go`, flykt och att gå ut ur en zon):
    `Player`, `Room`, `To`, `Dir`. Avbruten: spelaren stannar och får anledningen.
  - `EventSay` – före `say`: `Player`, `Room`, `Text`. Hooks får ändra
    `Text` (det som då sägs); avbruten: inget sägs, spelaren får anledningen.
  - `EventTick` – efter varje tick (efter strid och återuppståndelse).
  - `EventDeath` – när en varelse besegras (`NPC`, `Player` = den som
    dödade, `Room`) och när en spelare dör (`Player`, `Room`).
- `emit(e) bool` kör Go-hooks, sedan Lua-hooks, i den ordning de lades till,
  och slutar vid den första som avbryter; `false` = avbruten.
- Ett avbrott utan anledning ger spelaren `MsgCancelled`.
- Lua: `mud.on(kind, function(event) ... end)`; en returnerad sträng avbryter
  med den anledningen. Lua-hooks laddas om med skripten.

## Uppgift

`events.go` är delvis given (`Event`, `Cancel`, `On`, `refuse`): skriv `emit`
och `runHook`. Lägg till `mud.on` och `Scripts.hooks` i `scripts.go` och dela
upp `run` i en `call` som kan returnera ett värde. Skicka sedan händelserna
i `travel`, `cmdSay`, `onTick`, `defeat` och `die`.

Prova med ett skript:

```lua
-- scripts/curse.lua
mud.on("say", function(ev)
  if ev.player == "Bertil" then return "Din tunga är bunden!" end
end)
```

## Fundera på

- Hooks körs i den ordning de lades till. När spelar ordningen roll? Borde
  hooks ha prioriteter?
- `EventSay` låter Go-hooks ändra texten, men Lua-hooks kan bara avbryta. Hur
  skulle du låta Lua ändra den också?
- Vad händer om en hook för `enter` själv flyttar spelaren (och alltså
  skickar ännu en `enter`)?
- Vilka andra händelser vore användbara? (`get`, `drop`, `attack`, `login` …)

## Utmaningar

- `on_event` för WebAssembly-plugins (ABI version 2 – och fortsätt ladda version 1).
- `EventGet` med ett avbrytbart "svärdet sitter fast i stenen".
- En hook som för statistik (dödade råttor per spelare) och ett kommando `stats` som visar den.
