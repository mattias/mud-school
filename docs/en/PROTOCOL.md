# The protocol – mudschool 1

> 🇬🇧 English · [🇸🇪 Svenska](../sv/PROTOCOL.md) · [🇯🇵 日本語](../ja/PROTOCOL.md)

This is the contract between the server (Go) and the client (Rust). It is
introduced in lesson 10 (the server) and 11 (the client).

## Ground rules

- Transport: TCP. All text is **UTF-8**.
- Every message is **one line** ending in `\r\n` (CR LF), in both directions.
- The client always sends plain text: commands like `look` or `say hello`.
- The server sends **plain text** until the client has sent a greeting.
  After that the server sends **one JSON frame per line**.

That way `telnet` and `nc` work just as before, while a client that
understands the protocol gets structured data.

## Handshake

The client sends, at any time (usually right after connecting):

```
@hello mudschool 1
```

The server replies with a `hello` frame and sends JSON from then on:

```json
{"type":"hello","text":"mudschool 1"}
```

Lines starting with `@hello` are never interpreted as game commands.

## Frames

Every frame is a JSON object with the field `type`.

### `text`

```json
{"type":"text","text":"Mudby Square\nA cobbled square with an old well in the middle."}
```

| Field | Type | Description |
|---|---|---|
| `text` | string | Text to show. May contain `\n`. Left out if the text is empty. |
| `kind` | string | What kind of text it is, so a client can colour it (the server sends it from lesson 15; the client uses it from lesson 17): `info`, `room`, `say`, `move`, `error`, `combat`. Missing = `info`. **An unknown kind must be shown like `info`** – new kinds may come. |

```json
{"type":"text","text":"Bertil says: hi!","kind":"say"}
```

The text is in the server's language (see `messages.go`).

### `map`

Sent after every `look` (and therefore also after moving and logging in) and
in reply to the command `map`.

```json
{"type":"map","map":{
  "width":5,"height":5,
  "rows":["  +  ","  |  ","  @-+","     ","\"    "],
  "legend":{"+":"inside","|":"path_ns","@":"self","-":"path_ew","\"":"field"}
}}
```

| Field | Type | Description |
|---|---|---|
| `width`, `height` | integer | The size of the map in characters. |
| `rows` | list of strings | `height` rows of `width` characters. Every character is a **code**. Space = nothing. |
| `legend` | object | Code → meaning, for the codes used in `rows`. |

Rooms are on even rows and columns; exits are drawn between them. The player
is always in the middle.

Meanings in version 1:

| Meaning | Code | Explanation |
|---|---|---|
| `town` | `#` | street, square |
| `inside` | `+` | indoors |
| `road` | `.` | road |
| `field` | `"` | meadow, field |
| `forest` | `T` | forest |
| `water` | `~` | water |
| `mountain` | `^` | mountain |
| `cave` | `o` | cave |
| `stairs_up` | `<` | a room with an exit up |
| `stairs_down` | `>` | a room with an exit down |
| `path_ew` | `-` | an east–west exit |
| `path_ns` | `\|` | a north–south exit |
| `self` | `@` | you |

> **Important:** the client should draw based on the **meaning** (`legend`),
> not the code. The codes are chosen to look sensible in telnet, but may change.

### `status`

Sent after every `look` and after every combat round.

```json
{"type":"status","status":{"name":"Anna","room":"Mudby Square","terrain":"town","hp":10,"max_hp":10,"online":2}}
```

| Field | Type | Description |
|---|---|---|
| `name` | string | The player's name |
| `room` | string | The room's name |
| `terrain` | string | The room's terrain (one of the meanings above) |
| `hp`, `max_hp` | integer | Health |
| `online` | integer | Number of players logged in |
| `target` | object | Who you are fighting (from lesson 18): `{"name":"a rat","hp":3,"max_hp":5}`. Missing = nobody. |

### `prompt`

The server is waiting for something special (from lesson 20). The question
itself comes as an ordinary `text` frame just before.

```json
{"type":"prompt","prompt":"password"}
```

| Field | Type | Description |
|---|---|---|
| `prompt` | string | `password` (an existing account) or `new_password` (choose one). The client should **hide** what is typed and not save it in its history. Other values may come; treat them as ordinary input. |

Telnet never sees `prompt` frames.

### `zone`

The whole zone the player is in (from lesson 23). Sent with every `look` –
so after logging in and in every new room – and in reply to `zone`.

```json
{"type":"zone","zone":{"room":"torget","width":9,"height":5,
  "rows":["####.####","#.......#","#..~....#","#.......#","#########"],
  "legend":{"#":"wall",".":"floor","~":"water"},
  "entities":[{"id":"p:Anna","kind":"player","name":"Anna","glyph":"@","x":4,"y":2},
              {"id":"i:lamp","kind":"item","name":"a lamp","glyph":"*","x":5,"y":2}],
  "self":"p:Anna",
  "visible":["111111111","111111111","111111111","111111111","111111111"]}}
```

| Field | Type | Description |
|---|---|---|
| `room` | string | The room's ID |
| `width`, `height` | integer | The size in cells |
| `rows` | list of strings | `height` rows of `width` tile codes. A space = not explored yet (from lesson 26). |
| `legend` | object | Code → meaning: `floor`, `wall`, `grass`, `tree`, `water`, `rock`, `door`, `stairs_up`, `stairs_down`. The client draws by meaning. |
| `entities` | list | Everything with a position that the player sees: `id` (`p:<name>`, `n:<id>`, `i:<id>`), `kind` (`player`, `creature`, `item`), `name`, `glyph` (for text clients), `x`, `y`. |
| `self` | string | The receiver's own entity ID |
| `visible` | list of strings | The field of view: `'1'` visible now, `'0'` not (from lesson 26). Missing = everything visible. |

Telnet gets the zone drawn as text (with the entities' glyphs) after every step.

### `delta`

What changed in the zone since the last frame (from lesson 23) – sent after a
command or a tick if anything changed. Much smaller than a whole zone.

```json
{"type":"delta","delta":{"room":"torget",
  "moved":[{"id":"p:Anna","kind":"player","name":"Anna","glyph":"@","x":5,"y":2}],
  "removed":["n:cat"]}}
```

| Field | Type | Description |
|---|---|---|
| `room` | string | The room it applies to. **A delta for another room must be ignored.** |
| `added`, `moved` | list of entities | New and moved entities (whole entities, as in `zone`) |
| `removed` | list of strings | IDs that are gone (or out of sight) |
| `tiles` | list | Tiles seen for the first time: `x`, `y`, `code`, `meaning` (from lesson 26) |
| `visible` | list of strings | The whole field of view, when it changed (from lesson 26) |

All lists may be missing. Telnet never sees `delta` frames.

## Compatibility

The protocol must be able to grow without breaking old clients:

1. **Clients must ignore frames with an unknown `type`** (in Rust: `#[serde(other)] Unknown`).
2. **Clients must ignore unknown fields.** New fields are only ever added, never changed.
3. **Clients must cope with optional fields being missing** (e.g. `text`).
4. A line that is not valid JSON is shown as plain text.
5. Incompatible changes require a new version: `@hello mudschool 2`.

## Planned extensions

| Version | Frame / field | Phase |
|---|---|---|
| 1 | `hello.lang` – the player's language, so the server can choose texts per player | 6 |
| 1 | `sound`, `effect` – sounds and visual effects | 6 |
