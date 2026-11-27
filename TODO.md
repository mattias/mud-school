# TODO.md – known open issues

Small, concrete things found in reviews and left for later. Each one says
where it is. Remove an item when it is done (and verified), add new ones as
they are found. Larger directions are in [PLAN.md](PLAN.md).

## Behaviour

- [ ] **Shutdown's first `s.call` has no time limit** (`Shutdown` in
  `lessons/14-*/solution/server/mud/loop.go` and every later copy of
  `loop.go`: lessons 16, 18, 23, 31–34, 36 – and the lesson 14 scaffold). Under heavy load
  (lesson 44, thousands of bots) the goodbye call queues behind every handler,
  and `docker stop` hits its 10 s SIGKILL. Either select on `ctx.Done()`
  around it, or at least say so in lesson 44's "What the numbers say".
- [ ] **`/healthz` says `"stopping"` when the loop is stuck**
  (`lessons/42-go-drift/solution/server/mud/health.go`). The 503 is right;
  `"stuck"` would be clearer to an operator. Needs README ×3 and the test.
- [ ] **Copyover window** (lesson 35, `copyover.go`): commands typed between
  `PrepareCopyover`'s snapshot and `HandOver` are lost, and `HandOver` doesn't
  save players – if the new binary dies, progress since login is gone. The
  README's "Think about" and the "health check" challenge mention it.
- [ ] **Scripts and plugins load inside the game loop** (lessons 32/33,
  `loadScripts`/`loadPlugins`), against lesson 31's "slow work outside, swap
  inside". Kept simple on purpose; could become a challenge or a fix.
- [ ] **A failing plugin at runtime is reported as `MsgPluginLoadFailed`**
  (lesson 33, `plugins.go`, `runPluginCommand`). A message of its own would be clearer.
- [ ] **`slug` keeps ASCII only** (lesson 36, `olc.go`): "Mörk grotta" →
  `m-rk-grotta`, Japanese names → `room`. Could become a challenge in the
  README ×3 (e.g. with `golang.org/x/text/unicode/norm`), or a fix.
- [ ] **`rezone` doesn't update `n.home`** (lesson 36): creatures respawn at
  their old cell; other players keep stale `explored` data for the room.
- [ ] **Lua scripts can register `@` commands that bypass `adminOnly`**
  (lesson 32). Only documented in "Think about".
- [ ] **One torch, two newcomers** (lesson 37): the tutorial item is shared.
  Documented ("instancing" in Think about).

## Course tool and infrastructure

- [ ] **Files a lesson moves away stay behind** in reference/verify builds and
  student workspaces: after lesson 40, `client/src/{protocol,render,tileset,zone}.rs`
  remain (unused; overlays can't delete). A `"remove": [...]` list in
  `lesson.json`, applied by `enter`/`reference`/`verify` (and offered, not
  forced, in a student's workspace), would fix it.
- [ ] **`doctor` doesn't check the Rust version.** The lessons work on recent
  stable; a minimum version (and Go's, which is checked) should be stated in
  SETUP ×3 and checked.
- [ ] **The later phases haven't been tried on Windows by a person** (the
  first lessons have). Things to
  watch: the macroquad window (28), `run web` and the wasm build (41), Docker
  Desktop (42), `-race` needing a C compiler.
- [ ] **Screenshots** (`docs/images/client-*.png`, `gfx.png`, `web.png`) are
  from the reference at the time; regenerate if the looks change
  (`tools/screenshots/make.sh`; `web.png` was taken with Playwright from a
  lesson-41 reference).

## Done recently (for context)

- Phase 6 (lessons 37–44) merged in PR #10, after two review rounds.
- CI installs the `wasm32-unknown-unknown` target (lesson 41's test needs it).
- The generators for the tutorial JSON and the web page are in `tools/gen/`;
  CI checks that the committed files match them.
