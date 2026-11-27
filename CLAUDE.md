# CLAUDE.md – working on MUD School

MUD School is a course: students learn **Go** by building a MUD server and
**Rust** by building its client, in 44 lessons and 6 phases, in **English,
Swedish and Japanese**. This repository is the course itself – the lessons,
their tests and reference solutions, and the course tool – not a game.

Read these before changing anything:

- [PLAN.md](PLAN.md) – where the course stands and where it could go next.
- [TODO.md](TODO.md) – known open issues, small and large.
- [docs/en/TEACHERS.md](docs/en/TEACHERS.md) – the directory layout, how the
  course tool works, how a lesson is built. **The reference for conventions.**
- [docs/en/CURRICULUM.md](docs/en/CURRICULUM.md) – every phase and lesson.

## The user

- Writes in **Swedish**; answer in Swedish. Everything in the repository is
  English by default (README.md is English; README.sv.md and README.ja.md are
  the translations).
- Is the course author, works on Windows (PowerShell, Windows Terminal) as well
  as Linux. Windows problems have come up before – keep Windows in mind.

## Commands

```
./mudschool verify                      # every lesson, every language: ~45 min
./mudschool verify 37 38 --lang en      # some lessons (still builds 01–NN in order)
./mudschool reference <dir> [id]        # the reference solution up to a lesson
./mudschool doctor                      # Go, Rust, wasm32 target, terminal
gofmt -l tools lessons starter starter.sv starter.en starter.ja   # CI checks this
(cd tools/mudschool && go vet ./... && go test ./...)             # the course tool
python3 tools/gen/tutorial.py           # regenerate the tutorial JSON (lessons 37/38)
python3 tools/gen/webpage.py            # regenerate web/index.html ×3 (lesson 39)
```

`verify` checks, per lesson, that the tests **fail** before the solution and
**pass** after it. A change to an early lesson can break any later one: run
the later lessons too, in every language, before calling something done.
CI (`.github/workflows/verify.yml`) runs gofmt, vet and the full verify.

## Rules for every change

- **Three languages, always.** A new player-facing text is a constant in all
  three `starter.<lang>/server/mud/messages.go` (or `client/src/messages.rs`).
  README and HINTS exist as `.md`, `.sv.md` and `.ja.md`; `lesson.json` has
  `sv`/`en`/`ja`. Test failure messages use `L(sv, en, ja)` / `l(sv, en, ja)`.
  Code and code comments are English.
- **Tests compare with the constants**, never with hard-coded text.
- **Test files start with `mudschool_`** and are self-contained: only the
  current lesson's tests are in the workspace. Shared helpers live in
  `lessons/_shared/<track>[-from-NN]/`.
- **`solution/` holds whole files** as they look after the lesson.
  **`scaffold/` only adds files** that don't exist yet – it can never change a
  student's file – and must **compile** on the previous lesson's reference
  (stubs with `TODO(lesson NN)`). A scaffold file that later changes in a
  solution must be given in full if the README says "given".
- Rust: tests live in the root `client` package (`client/tests/`). Other crates
  in the workspace (`core`, `web`, `loadtest`) are its `[dev-dependencies]`, so
  `cargo test` in `client/` reaches them.
- The README contract (`## The contract`) must match what the tests check.
- Things learned the hard way:
  - coder/websocket's `NetConn` closes the connection when a read deadline
    expires; web connections go through `newDeadlineConn` (lesson 39).
  - Copyover (lesson 35) has to carry every new piece of player state
    (`CopyoverPlayer`) – lessons 37, 38 and 43 each extend it.
  - Write files as UTF-8 (`encoding="utf-8"` in Python) – mojibake has slipped
    into scaffolds before. Japanese texts use full-width punctuation.
  - Windows: copyover is unsupported there; tests that need Unix skip on it.

## Workflow (agreed with the user)

1. Work on the branch the session names (so far
   `claude/interactive-golang-rust-mud-iks9h5`). Commit with clear messages
   that end with the attribution lines the session asks for. **No model names
   in commits, PRs or files.**
2. Push, open a PR (title and body in English; the body ends with the
   session's attribution).
3. Have a **reviewer subagent** review the diff: correctness and races, tests
   (stress with `-race -count=5 -cpu 1,4`), scaffolds, all three languages,
   docs. Fix every blocking and should-fix finding, then let it re-review –
   until it approves ("PR-dansen").
4. Merge with the PR's head SHA (`expectedHeadSha`), then fast-forward the
   branch to `origin/main` and push. **No force pushes.**
5. Big work has been done two phases at a time; ask before starting a new phase.

## How the big phases were built

For a phase of new lessons it worked well to:

1. `./mudschool reference <scratch>/dev NN` and `git init` there;
2. implement lesson by lesson in that copy, one commit per lesson, with the
   tests next to the code, running them (`-race`) as you go;
3. extract each lesson's files into `lessons/NN-*/`: solution files from that
   lesson's commit, scaffolds made by replacing function bodies with TODO stubs;
4. write README/HINTS in English, then sv and ja; then `verify` in all three.

## Environment notes (cloud sessions)

- Disk is limited. `.verify-target/` (shared cargo target for verify) grows to
  several GB; delete it and other `target/` dirs when space runs out.
- Docker is installed but the daemon is not running: start it with `dockerd &`.
  Image builds need the agent proxy's CA (`/root/.ccr/ca-bundle.crt`) copied in
  for `go mod download`/`cargo` – only in a scratch copy, never in the repo.
- Playwright and Chromium are preinstalled for browser checks of the web client
  (lessons 39 and 41) and for screenshots (`tools/screenshots/`).
- `rustup target add wasm32-unknown-unknown` is needed for lessons 33 and 41.
