# For teachers and course authors

> 🇬🇧 English · [🇸🇪 Svenska](../sv/TEACHERS.md) · [🇯🇵 日本語](../ja/TEACHERS.md)

How the course is built, and how to add or change lessons.

## Directory layout

```
mudschool                    script that builds and runs the course tool
tools/mudschool/             the course tool (Go, standard library only)
starter/                     the code the student starts with (copied to workspace/)
  server/                    Go module "mudserver"
  client/                    Cargo package "mud-client" (the library is called mud_client)
starter.<lang>/              language-dependent starter files: messages.go and messages.rs
lessons/
  _shared/go/...             files copied in for EVERY Go lesson (test helpers)
  _shared/rust/...           the same for Rust (tests/mudschool_common/)
  NN-name/
    lesson.json              title and summary in every language, track ("go"/"rust")
    README.md                the instructions in English (README.sv.md, README.ja.md)
    HINTS.md                 hints, one per "## " heading (HINTS.sv.md, HINTS.ja.md)
    scaffold/                NEW files the student gets when the lesson starts
    scaffold.<lang>/         language-dependent new files (e.g. the world)
    tests/                   tests copied in by `mudschool check`
    solution/                the reference solution: files as they look AFTER the lesson
    solution.<lang>/         language-dependent parts of the solution (rarely needed)
workspace/                   the student's code (ignored by git)
docs/<lang>/                 curriculum, protocol, setup, ideas, this document
```

All paths under `scaffold/`, `tests/` and `solution/` are relative to
`workspace/`, e.g. `tests/server/mud/mudschool_04_test.go`.

## Languages

The course exists in Swedish (`sv`), English (`en`) and Japanese (`ja`).

| What | How it is translated |
|---|---|
| Instructions and hints | `README.<lang>.md`, `HINTS.<lang>.md` (English, the default, is `README.md`) |
| Title and summary | objects in `lesson.json`: `{"sv": ..., "en": ..., "ja": ...}` |
| The course tool's texts | `L(sv, en, ja)` in `tools/mudschool` |
| The game's texts | `starter.<lang>/server/mud/messages.go` and `.../client/src/messages.rs` |
| The game world | `scaffold.<lang>/` in lessons 08 and 09 (rooms), 15 (`items.json`) and 16 (`npcs.json`) |
| The tests' failure messages | `L(sv, en, ja)` (Go) and `l(sv, en, ja)` (Rust) |
| Documentation | `docs/<lang>/` |
| Code and code comments | **English** – shared by everyone |

The language is chosen with `./mudschool init --lang en|sv|ja` (otherwise it
is guessed from `$LANG`) and can be changed with `./mudschool lang <lang>`. It
is stored in `workspace/.mudschool.json` and passed to the tests as `MUDSCHOOL_LANG`.

The tests never compare against hard-coded text. They use the same constants
as the student's code (`MsgAskName`, `fmt.Sprintf(MsgSays, "Anna", "hi")`), so
they work in every language – and the student is welcome to change the texts.

If the student switches languages with `./mudschool lang`, an edited
`messages.go`/`messages.rs` is first saved as `.bak`.

## How the course tool works

- **`init`** copies `starter/` and `starter.<lang>/` to `workspace/` and
  enters the first lesson.
- **Entering a lesson** copies `scaffold/` and `scaffold.<lang>/` – but only
  files that don't exist yet. The student's code is never overwritten.
- **`check`**:
  1. removes every file and directory in `workspace/` whose name starts with
     `mudschool_` (old course tests),
  2. copies in `lessons/_shared/<track>/` and the lesson's `tests/`,
  3. runs `go vet ./... && go test -count=1 ./...` in `workspace/server`
     (Go) or `cargo test` in `workspace/client` (Rust),
  4. when green: marks the lesson done and enters the next one.
- **`doctor`** checks that Go, Rust and the tools are present.

Only the current lesson's tests run. That is deliberate: the behaviour changes
between lessons (the echo from lesson 01 disappears in lesson 04).

### Naming convention for tests

Every test file **must** start with `mudschool_` – otherwise it is not cleaned
up and old tests start failing. Go: `mudschool_NN_test.go` in package `mud`.
Rust: `tests/mudschool_NN.rs` (integration tests against the library
`mud_client`), starting with `mod mudschool_common;` to get `l(...)`.

## Updating students' workspaces

The tool records a checksum for every file it copies in from `starter/` and
`scaffold/` (`files` in `.mudschool.json`). `./mudschool update` compares:
unchanged files are replaced with the course's current version, files that
were never in the workspace are added (a file the student deleted or renamed
is left alone), a file changed by both the student and the course gets the
course's version as `<file>.new`, and an edited `messages.go`/`messages.rs` is
backed up as `.bak` and replaced (new lessons need its new constants).
Workspaces created before the records existed are matched against every
version of the file in git history; a file that matches none is the
student's own work and is left alone. `check` applies the safe updates by itself.
So: fixing a scaffold or starter file is fine – students get it with `git pull`.

## Verification

```
./mudschool verify              # every lesson, every language
./mudschool verify --lang ja    # only one language
./mudschool verify 07 08        # only some lessons (all are still built up in order)
```

`verify` builds up a workspace in a temporary directory: `starter/`, then for
every lesson `scaffold/` and `solution/` in turn – once per language. For every
lesson it checks that

1. the tests **fail** with the code *before* the lesson (previous solution +
   scaffold) – otherwise the lesson tests nothing new, and
2. the tests **pass** with the lesson's solution.

`./mudschool reference <dir> [id]` builds the whole reference solution (up to
a given lesson) so you can run it. Run `verify` before you commit; CI
(`.github/workflows/verify.yml`) does the same.

## Screenshots

`docs/images/client-<lang>.png` are generated from the reference solution with
`tools/screenshots/make.sh` (needs tmux, nc, Node with Playwright and a Japanese
font such as Noto CJK). Run it again when the client's look changes. Use
images rather than text diagrams when Japanese text and box-drawing characters
must line up – GitHub's fonts don't draw CJK characters exactly two columns wide.

## Writing a new lesson

1. Create `lessons/NN-track-name/` with `lesson.json`, `README*.md`, `HINTS*.md`.
2. Write `solution/`: **whole** files, as they should look after the lesson.
   Files that don't change need not be there – they are inherited from earlier lessons.
3. A new player-facing text? Add the constant to **all three** `messages.go`
   (or `messages.rs`) in `starter.<lang>/`.
4. Write `tests/`. Keep in mind:
   - Failure messages with `L(sv, en, ja)` that say what was expected.
   - Compare with the constants in `messages.go`, never with hard-coded text.
   - Test behaviour, not implementation. Server tests preferably go over TCP
     with the helpers in `_shared/go/server/mud/mudschool_helpers_test.go`
     (`startServer`, `login`, `expect`, `expectContains`, `expectSilence`, ...).
   - Just strict enough: test what the lesson teaches, leave freedom in the rest.
   - No sleeps where they can be avoided; use timeouts.
5. Write `scaffold/` for new files: types and signatures with `TODO(lesson NN)`
   and `todo!()`/empty return values so that the code compiles.
6. Run `./mudschool verify NN`.

### The README template

```
# Lesson NN – Title
(Motivation: which problem in the GAME are we solving?)
## Goals
## <Language> concepts (short examples, ideally a "Go ↔ Rust" comparison)
## Task                (numbered, file by file, with the constants the tests require)
## Try it yourself
## Think about         (2–3 reflection questions)
## Challenges          (optional, do not affect the tests)
```

### Hints

Every `## ` heading in `HINTS.md` is one hint. Order them from vague to
concrete; the last hint may well be almost-code.
