# Lesson 30 – The mouse

The last lesson of phase 4 gives the window a mouse:

- **click** somewhere and you walk there – the client finds the way itself
  (A*, like the server's in lesson 27, now in Rust) and sends one step at a time;
- **point** at something and a tooltip says what it is: *"a big rat"*, *"A tree"*;
- **right-click** and a menu shows what you can do: *Look at a big rat*,
  *Attack a big rat*, *Walk here*.

```
#########
#@..#...#        right-click on the rat:
#...#.r.#           ┌──────────────────────┐
#.......#           │ Look at a big rat    │  -> "examine rat"
#.....*.#           │ Attack a big rat     │  -> "kill rat"
#########           │ Walk here            │  -> walk (in the client)
                    └──────────────────────┘
```

## Goals

- Implement A* again – in Rust, with `BinaryHeap` and `Reverse`.
- Convert between coordinate systems: pixels → screen cells → zone cells.
- Turn a sequence of actions into commands sent over time (a queue + a clock).
- Keep "what can I do with this?" as data (a menu of `MenuItem`) – easy to test,
  easy to draw in any renderer.

## Rust concepts

### `BinaryHeap` and `Reverse`

Rust's priority queue gives you the **largest** element first. A* needs the
smallest `f`, so wrap the key in `std::cmp::Reverse`, which turns the
ordering around:

```rust
let mut open = BinaryHeap::new();
open.push(Reverse((f, g, cell)));            // tuples compare field by field
while let Some(Reverse((_, g, cur))) = open.pop() { ... }
```

Compare with Go's `container/heap` in lesson 27: there you wrote `Less`; here
`Ord` comes for free with tuples of numbers.

### A queue of steps

`walk_to` turns a path into commands (`"southeast"`, `"east"`, …) in a
`VecDeque`. `next_walk_step(now)` hands out one at a time, at most every
`WALK_STEP` – the clock is passed in, as in lesson 19, so the tests decide what
time it is. Any key press empties the queue: the player has taken over.

### Coordinates

```
pixels (mouse_position)  --screen_to_cell-->  screen cell (x, y)
screen cell              --Scene::zone_cell-->  zone cell  (origin + (x, y − 1))
```

The scene remembers which zone cell it showed at the top left (`origin`) – the
camera from lesson 25 – so a click can be turned back into a zone cell.

## The contract

`mouse.rs` (new):

- `walkable(meaning)` – floor, grass, door, stairs; `STEPS` – the eight
  directions with their commands.
- `step_command(from, to)` – the command for one step to a neighbouring cell,
  `None` if they are not neighbours.
- `find_path(zone, from, to)` – the shortest path (8 directions) over tiles
  the client knows are walkable, without `from`, with `to`; `None` if `to`
  isn't walkable or can't be reached; `Some(vec![])` when `from == to`.
- `describe_cell(zone, x, y)` – the entity's name, else
  `messages::tile_name(meaning)`, else `None` (never seen).
- `context_menu(zone, x, y) -> Vec<MenuItem {label, command}>`: a creature →
  `LOOK_AT` + name / `examine <id>` and `ATTACK` + name / `kill <id>`; an item →
  `LOOK_AT` / `examine <id>` and `PICK_UP` / `get <id>`; then, if the cell is
  walkable, `WALK_HERE` / `WALK` (`"@walk"`, handled by the client). `<id>` is
  the entity ID without its prefix (`n:rat` → `rat`) – the server always
  understands it.
- `screen_to_cell(mx, my, cell_px)`.

`messages.rs`: `LOOK_AT`, `ATTACK`, `PICK_UP`, `WALK_HERE`, `tile_name(meaning)`.

`app.rs`: `walk: VecDeque<String>`, `next_step`, `hover`, `menu: Option<Menu
{cell, items}>`, `WALK_STEP` (150 ms), and:
- `walk_to(cell)` – the path from yourself, as step commands;
- `next_walk_step(now)` – the next step, at most every `WALK_STEP`;
- `left_click(cell, menu_row)` – with an open menu: choose row `menu_row`
  (returns its command, or starts walking for `WALK`) and close the menu;
  without: walk to the cell;
- `right_click(cell)` – opens the cell's menu (none if it is empty);
- a key press stops the walk and closes the menu.

`scene.rs`: `Scene::origin`, `Scene::overlay` (the tooltip for `hover`, or the
menu's labels, one per row from the cell downwards), `Scene::zone_cell(x, y)`,
and `Renderer::overlay` with a default body (ordinary text).

The window (`mud-gfx.rs`, `gfx.rs`): hover, left and right click, a step per
`WALK_STEP`, and an overlay with a dark box behind the text.

## Task

`mouse.rs` first – the tests check it on its own. Then the App, the scene and
the window. Walk around with the mouse in `./mudschool run gfx`.

## Think about

- The client finds paths over the tiles *it knows*. When can the path be wrong,
  and what happens then? (Hint: unexplored cells, a creature in the way.)
- Why is the path computed in the client instead of the server sending it?
- The menu's commands are strings like `"kill rat"`. What are the advantages
  and disadvantages compared with a separate protocol for actions?
- How would you make the same mouse work in the terminal client? (crossterm has
  `EnableMouseCapture` and `Event::Mouse`.)

## Challenges

- The mouse in the terminal client too – the App logic is already there.
- Show the planned path while you hover (a faint line of dots).
- Shift-click to attack directly, double-click to pick up.
- Stop the walk when a creature comes into sight.

**Phase 4 is done.** The MUD now has a tile world generated from seeds, a
field of view, hunting creatures and a graphical client with light and a
mouse. Phase 5 makes the server extensible without restarting it.
