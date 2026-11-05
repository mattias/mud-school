# Lesson 10 – Protocol and map

So far the server has only sent text. A telnet client can't do much more than
display it – but our Rust client is going to draw a colour map, a status bar
and (later) tile graphics. For that it needs **structured data**, not text
whose meaning it has to guess.

At the same time we don't want to lose telnet. The solution: the client
*negotiates*.

```
server: Welcome to MUD School!
server: What is your name?
client: @hello mudschool 1                      ← "I understand JSON"
server: {"type":"hello","text":"mudschool 1"}
client: Anna
server: {"type":"text","text":"Welcome, Anna!"}
server: {"type":"text","text":"Mudby Square\nA cobbled square..."}
server: {"type":"map","map":{"width":13,"height":13,"rows":[...],"legend":{"#":"town",...}}}
server: {"type":"status","status":{"name":"Anna","room":"Mudby Square","hp":10,...}}
```

The whole protocol is described in `docs/en/PROTOCOL.md`. It is the
**contract** between your Go server and your Rust client.

## The most important idea of the lesson

> **The server sends meaning – the client decides appearance.**

The map is sent as codes (`T`) with an explanation (`"T": "forest"`). Telnet
shows the code as it is. The Rust client draws a green `♣` in lesson 12, and
in phase 4 a nice 16×16-pixel spruce image – **without changing the server at all**.

## Goals

- Let the *same* game logic have two different encodings (text and JSON).
- Use `json.Marshal` and `omitempty`.
- Use `sync/atomic` for a flag that is read by several goroutines.
- Draw a 2D grid with `[][]byte`.

## The map

Rooms go on every other row/column so the exits fit in between:

```
    +          tavern            (radius 1 -> 5×5 characters)
    |
    @-+        square – smithy   ('@' = you)

"              meadow (no exit there from here)
```

## Task

New files: `frame.go` (the types are finished – they *are* the protocol) and `mapview.go`.

1. **`frame.go`**: write `encodeText` and `encodeJSON`.
2. **`mapview.go`**: write `RenderMap` and `WithSelf` as the comments say.
   The tables `terrainCodes` and `mapLegend` are given.
3. **`player.go`**:
   - replace `outbox chan string` with `outbox chan Frame` and add
     `SendFrame(f Frame)`; `Send(text)` becomes `SendFrame(TextFrame(text))`,
   - give `Player` a field `structured atomic.Bool`; `writeLoop` encodes each
     Frame with `encodeJSON` if it is set, otherwise `encodeText` (and skips
     empty strings),
   - in `readLine`: a line starting with `HelloPrefix` (`@hello`) sets
     `structured`, is answered with `Frame{Type: "hello", Text: "mudschool 1"}`
     and is **not** passed on to the commands.
4. **The room commands**: after every `look` (and therefore every move and
   login) a structured client also gets a `map` frame (radius 3, with
   `WithSelf`) and a `status` frame. A new command `map` sends the map to
   everyone – in telnet that becomes the ASCII map.

## Try it yourself

```
telnet 127.0.0.1 4000
map
```

and to see what the Rust client is going to get:

```
nc 127.0.0.1 4000
@hello mudschool 1
Anna
```

## Think about

- Why does `writeLoop` read the flag for *every* Frame instead of once?
- Why `atomic.Bool` and not an ordinary `bool`? Try `--race` with an ordinary `bool`.
- What happens if a future client sends `@hello mudschool 2`?

## Challenges

- Show other players on the map as `&` (add `"&": "player"` to the legend).
- Send a new status to everyone when someone logs in or out (`online` changes).
