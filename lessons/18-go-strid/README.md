# Lesson 18 – Combat

The rat in the cave shows its teeth. The wolf in the dark wood howls. Time to
fight. Combat in a MUD is a classic job for the game loop: once a fight has
started, it goes **one round per tick** until someone falls or flees.

```
> kill rat
You attack a cave rat!
You hit a cave rat (3 damage).        ← tick
A cave rat hits you (1 damage).
You hit a cave rat (2 damage).        ← tick
You have defeated a cave rat!
```

The hard part isn't the fighting – it is **testing** something that depends
on dice and on time. This lesson's tests freeze the world (`Tick: time.Hour`)
and step it forward themselves, one tick at a time:

```go
srv.call(func() { srv.ticks++; srv.onTick() })   // exactly one tick
```

That is why the game loop's state lives in one place: a test can drive it
like a debugger.

## Goals

- Model a fight as state in the game loop (who fights whom) and advance it every tick.
- Write randomness that is fair and testable (`roll` with an injected `rand.Rand`).
- Write deterministic tests for time-dependent behaviour.
- Extend the protocol again: `status.target`.

## The contract

- `const PlayerMaxHP = 20`, `const DefaultRespawn = 50`,
  `var PlayerDamage = [2]int{1, 4}`; `Options.Respawn` (ticks, 0 = default).
- `Player.HP`, `Player.MaxHP` (full when logging in) and `Player.target *NPC`.
- The creature's fields from lesson 16: `HP` (max health), `Damage [2]int`,
  `Attackable`; at runtime its current health, whether it is defeated and
  when it comes back.
- Commands:
  - `kill <creature>` (`k`, `attack`): `MsgAttackWhat`, `MsgNoSuchTarget`,
    `MsgCantAttack`, otherwise `MsgYouAttack` to you and `MsgAttacks` to the room.
  - `flee`: `MsgNotFighting`, `MsgFleeNowhere`, otherwise `MsgYouFlee` and a
    walk through a random exit. The fight is over.
- Every tick, for every fighting player in **name order**:
  1. the player hits for `roll(PlayerDamage)`: `MsgYouHit`;
  2. if the creature's health is ≤ 0: `MsgYouKill` to you, `MsgKills` to the
     room, everyone fighting it stops, it disappears and comes back after
     `Respawn` ticks in its start room with full health (`MsgAppears`);
  3. otherwise the creature hits back for `roll(n.Damage)`: `MsgHitsYou`;
  4. if your health is ≤ 0: `MsgDies` to the room, you wake up in the start
     room with full health (`MsgYouDie` with the room's name) and get a `look`.
- A creature that someone is fighting doesn't wander.
- All combat messages have kind `combat`; a creature that comes back (`MsgAppears`) is `move`, like anyone else arriving.
- `Status` gets `HP`/`MaxHP` from the player and
  `Target *Target{Name, HP, MaxHP}` (json `target,omitempty`) during a fight.
  A structured client gets a status after every round.

## Task

1. **`combat.go`** (new): fill in the functions.
2. **`loop.go`**: `onTick` skips defeated creatures, then runs
   `combatRound()` and `respawnCreatures()`.
3. **`npc.go`**: the runtime fields (current HP, defeated, respawn tick,
   start room), set up in `AddNPCs`; `Wanderer` stays put during a fight.
4. **`server.go`**/**`player.go`**: `Options.Respawn`, the player's HP.
5. **`frame.go`**/**`rooms.go`**: `Target` in the status.
6. Register `kill` and `flee`, and the aliases `k` and `attack`.

## Try it yourself

```
./mudschool run server
```

Walk to the cave (s, s, s, s, s, s, e, s) and `kill rat`. Then try the wolf in
the dark wood – and `flee` in time!

## Think about

- Why do the tests step the world themselves instead of sleeping and hoping
  that enough ticks have passed?
- Why must the combat round go through the players in a fixed order?
- What if two players fight the same creature? Who gets the kill?

## Challenges

- Experience points: defeated creatures give XP, and enough XP raises your max HP.
- Weapons: holding the rusty sword (lesson 15) adds damage.
- Aggressive creatures that attack players who walk in.
