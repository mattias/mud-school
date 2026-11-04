# Ideas to build on

> 🇬🇧 English · [🇸🇪 Svenska](../sv/IDEAS.md) · [🇯🇵 日本語](../ja/IDEAS.md)

A collection of ideas: for the game, for the client and for the course itself.
Many of them work as challenges, projects of your own or future lessons.

## The game

**Inspired by Dwarf Fortress and Caves of Qud**

- **History generation:** generate the village's history before the game
  starts – founders, wars, lost artefacts – and let NPCs tell it. Books in the
  world that contain generated history.
- **Body parts instead of HP:** a hit to the arm makes you drop the sword.
- **Mutations and factions** (Qud): reputation with different groups that
  affects how creatures react to you.
- **Materials and temperature:** wood burns, water freezes in winter, lava
  melts stone. A simulation system that runs in the game loop.
- **Ecology:** wolves hunt deer, deer eat grass. If you hunt away all the
  deer the wolves get hungry ...

**Classic MUD**

- Day and weather: night makes the map darker, rain gives sound and worse sight.
- Quests written in Lua (phase 5), with per-player state.
- Crafting: recipes in data files; the smithy in Mudby becomes useful.
- Economy: shops, prices affected by supply and demand.
- Pets and companions that use the pathfinding from phase 4.
- Mail and a notice board that persist across restarts.
- Channels (`ooc`, `newbie`) and an ignore list.
- Groups: share experience and see each other on the map.

**Technical game features**

- **Replays:** because the protocol is JSON lines, the client can record a
  session to a file and play it back – without any server at all.
- **Spectator mode:** follow another player and see what they see.
- **Bots:** write AI players in Rust that explore the world (good for load
  testing and for finding bugs).
- **Achievements and leaderboards.**

## The client

- A minimap of the whole explored world next to the local map.
- Themes: selectable colour palettes, including **colour-blind friendly** ones
  (don't distinguish only green and red) and a monochrome "phosphor screen" variant.
- A screen-reader mode: text only, no borders – accessibility is an important
  reason why MUDs are still played.
- Macros and aliases in the client (`alias k kill`), triggers ("when the text
  'You are hungry' appears, send 'eat bread'").
- Clickable exits and names in the log.
- Sound: a simple sound engine (e.g. `rodio`) that plays effects from `sound` frames.
- Tiles: support for several tilesets that can be switched with a key,
  including an "isometric" mode as a challenge.

## The course

- **Bug-hunt lessons:** a lesson where the course tool *plants* a bug (a data
  race, a deadlock, a leaking goroutine) that you have to find with `-race`,
  `pprof` or logging.
- **Refactoring lessons:** "replace the mutex with an actor" (phase 3) is one –
  the tests are unchanged, only the code changes.
- **Read real code:** short guided tours of the source code of e.g.
  `net/http`, `bufio` or `ratatui` when they are used.
- **Pair programming:** one person builds the server, one the client, and the
  protocol document is the only thing they share. Swap roles every phase.
- **Code review:** a checklist for reviewing each other's solutions.
- **Katas:** small stand-alone exercises per language (e.g. "implement
  `ParseCommand` in Rust") for those who want more practice with a concept.
- **A web version of the course tool:** show lessons and test results in a
  browser (Go: `net/http` + `html/template`), runnable e.g. in GitHub
  Codespaces without installation.
- **Achievement badges** in `mudschool list` for completed challenges.
- **The game in several languages at once:** every player picks a language at
  login and the server chooses `messages` per player. A good exercise in
  redesigning when global constants are no longer enough.
- **More languages:** the same protocol makes it easy to add a track where the
  client is written in a third language – and the course itself can be
  translated into more languages in the same way as Swedish, English and Japanese.
