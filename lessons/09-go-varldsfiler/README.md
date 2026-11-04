# Lesson 09 – World files

Writing rooms in Go code works for six rooms, but not for six hundred. And
you cannot change the world without recompiling. Real MUDs have *builders*
who create areas in data files. It is also the first step towards a server
that can **reload content while it is running** (phase 5).

The course tool has put `workspace/server/mud/world.json` in place with 21
rooms: Mudby, the country road, the lake, the forest and a cave. Every room
now also has a **terrain** and **coordinates** – we need those for the map in
the next lesson.

## Goals

- Read JSON with `encoding/json` and control field names with **struct tags**.
- Program against interfaces: `io.Reader` instead of file names.
- Validate data properly and report *all* errors with `errors.Join`.
- Bake files into the program with `//go:embed`.

## Go concepts

### JSON and struct tags

```go
type Room struct {
    ID    string            `json:"id"`
    Exits map[string]string `json:"exits,omitempty"`
}

var r Room
err := json.Unmarshal(data, &r)        // from []byte

dec := json.NewDecoder(reader)         // from any io.Reader
dec.DisallowUnknownFields()            // "nmae" instead of "name" -> error
err = dec.Decode(&r)
```

### io.Reader

`LoadWorld(r io.Reader)` can read from a file (`os.Open`), a string
(`strings.NewReader` – that's what the tests do), a byte slice
(`bytes.NewReader`), a network ... Interfaces in Go are satisfied
*implicitly*: every type with the method `Read([]byte) (int, error)` is an
`io.Reader`.

### errors.Join

```go
var errs []error
errs = append(errs, fmt.Errorf(MsgErrExitMissing, id, dir, to))
return errors.Join(errs...)   // nil if the list is empty
```

A builder who has made five mistakes wants to see all five at once.

### go:embed

```go
import _ "embed"

//go:embed world.json
var defaultWorldJSON []byte   // the file's contents, baked in at compile time
```

## The coordinate system

```
        y-1 (north)
 x-1 (west) · x+1 (east)        z+1 = up, z-1 = down
        y+1 (south)
```

## Task

1. **`world.go`**:
   - add the fields `Terrain string`, `X`, `Y`, `Z int` to `Room` and JSON
     tags on every field (look in `world.json` for the names),
   - `LoadWorld(r io.Reader) (*World, error)` – read the format
     `{"start": "...", "rooms": [...]}`, with `DisallowUnknownFields`, return
     an error if the same ID appears twice, and validate,
   - `LoadWorldFile(path string) (*World, error)`,
   - `(w *World) Validate() error`, which finds **all** of these mistakes:
     the start room is missing · a room without id or name · unknown terrain
     (not in `Terrains`, see below) · unknown direction · an exit to a room
     that doesn't exist · two rooms with the same coordinates · an exit whose
     coordinates don't match (walking `north` the target must be at `y-1`).
     The error messages are ready in `messages.go` (`MsgErrStartMissing`,
     `MsgErrExitMissing`, ...); use them with `fmt.Errorf`,
   - `var Terrains = []string{"town", "inside", "road", "field", "forest", "water", "mountain", "cave"}`.
2. **`defaultworld.go`**: rewrite `DefaultWorld()` so that it reads the
   embedded `world.json` (and may `panic` if it is broken – that is a
   programming error).
3. **`main.go`**: add a flag `-world file.json` that loads a different world.

## Try it yourself

```
cd workspace/server
cp mud/world.json /tmp/my-world.json   # change something, make mistakes on purpose
go run . -world /tmp/my-world.json
```

## Think about

- Why is `[3]int` a valid map key but `[]int` is not?
- Why may `DefaultWorld` `panic` but not `LoadWorldFile`?

## Challenges

- Build an area of your own with at least five rooms and add it to `world.json`.
- Warn (log, but don't fail) about one-way exits.
- Write a small tool `go run ./cmd/worldcheck file.json` that only validates.
