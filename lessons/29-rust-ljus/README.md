# Lesson 29 – Light and life

The window from lesson 28 shows the world, but it is flat: every cell equally
bright, the water frozen, nothing happens when you are hit. Games like
*Brogue* get their atmosphere from **light** – a torch in a dark cave, light
of different colours blending – and from small movements.

```
cave, ambient light (30, 30, 45)     you carry a torch (255, 190, 110), radius 6

   . . . . . . . . . .                  . . . ░ ▒ ▓ ▒ ░ . .
   . . . . . . . . . .        ->        . . ░ ▒ ▓ @ ▓ ▒ ░ .
   . . . . . . . . . .                  . . . ░ ▒ ▓ ▒ ░ . .
```

## Goals

- Model light as numbers: colours that are **added** (light sources) and
  **multiplied** (light on a surface).
- Write animation as **pure functions of time** – `shimmer(x, y, t)` – so it
  can be tested at any moment without waiting.
- Build a small particle system: create, move, fade, remove.
- Use **default methods** in a trait.

## The model

The light on a cell is the ambient light of the terrain plus every light
source, weaker with distance:

```
light = ambient + Σ colour × max(0, 1 − distance / radius)        (each channel capped at 255)
what you see = the cell's colour × light / 255                     (shade, lesson 28)
```

A red and a green light make yellow – exactly as with real light. Cells you
have explored but don't see are still `FOG`.

### Animation without timers

`shimmer(x, y, t_ms)` is a number between 0.7 and 1.0 that moves like a slow
wave: `0.85 + 0.15 × sin(t/400 + 0.7x + 0.3y)`. Nothing is stored and nothing
counts down – give it a time and it answers. That makes it trivial to test
(`build_scene_at(app, w, h, start + 500 ms)`) and impossible to get out of step.

Time comes from `App::started` (set in `App::new()`): the scene uses
`now − started`.

### Particles

Particles are the exception: they *move*, so they have state. `Particles`
holds a list; `burst` adds a ring of sparks, `update(dt)` moves them, pulls
them down a little, counts down their life, and `retain` removes the dead.
`App::tick(now)` calls `update` with the time since the last tick – the
window calls it every frame.

### Default methods

```rust
pub trait Renderer {
    fn cell(&mut self, x: u16, y: u16, cell: &SceneCell);
    fn text(&mut self, x: u16, y: u16, text: &str, fg: Rgb);
    fn particle(&mut self, _x: f32, _y: f32, _color: Rgb, _alpha: f32) {}
}
```

`particle` has a body: renderers that can't draw sparks – `TextRenderer`, the
renderers in your tests – don't need to implement it. Adding a method to a
trait *with* a default doesn't break any existing implementation.

## The contract

`light.rs` (partly given):

- `Light {x, y, radius, color}`, `TORCH = Rgb(255, 190, 110)`.
- `ambient(terrain)` (given): dark in caves, light in towns, full light (255) for `""`.
- `light_at(x, y, lights, ambient) -> Rgb` – as above; (x, y) is a cell, the
  lights' positions are cells too.
- `lights_for(zone)` – a torch (radius 6) for every player entity.
- `shimmer(x, y, t_ms) -> f32` – as above. `scale(colour, k)` (given).
- `Particle {x, y, vx, vy, life, max_life, color}` with `alpha()` = life /
  max_life; `Particles {list}` with `burst(x, y, n, colour)` (given) and
  `update(dt)`.
- `shake_offset(flash_until, now) -> (f32, f32)` – `(0, 0)` when not flashing;
  otherwise a small offset (at most half a cell) that shakes back and forth.

`scene.rs`:

- `cell_light(app, x, y, now)`: `FOG` if not visible; otherwise
  `light_at(x, y, lights_for(zone), ambient(terrain from the status))`, and
  water multiplied by `shimmer`. Without a zone: `FULL_LIGHT`.
- `build_scene_at(app, width, height, now)` – like `build_scene`, with
  `light = cell_light(...)`, `particles` (in view coordinates, with alpha) and
  `shake`. `build_scene(...)` = `build_scene_at(..., Instant::now())`.
- `Renderer::particle` with an empty default body; `draw_scene` draws the
  particles after the cells (one row down, like the cells).

`app.rs`: `App::particles`, `App::last_tick`, `App::started` (set in
`new()`), `App::tick(now)`; a hit (lower `hp`) bursts 12 red sparks in the
middle of your cell (`x + 0.5, y + 0.5`). The terminal uses `cell_light` too.

`gfx.rs` and `mud-gfx.rs`: `particle` draws a small square, `shake` moves the
zone, and the loop calls `app.tick(now)` and `build_scene_at(..., now)`.

## Task

`light.rs`, then `scene.rs`, then `app.rs` and the window. Walk into the cave
(`grottan`) with `./mudschool run gfx` – and let the rat bite you.

## Think about

- Why are the light sources *added* but the light on a surface *multiplied*?
- `light_at` loops over all lights for every cell. With 50 torches in a 200×200 zone?
- Light passes through walls right now. How would you use the field of view
  from lesson 26 to fix that – and where should that code run?
- Particles live in the client only. What if two players should see the same explosion?

## Challenges

- Light through walls: compute `ComputeFOV` from every light source (in the client) and only light what it sees.
- Torches as items: a `glow` field on items in the protocol, so a dropped torch keeps lighting the room.
- Day and night: let the ambient light follow the game clock from lesson 14 (`time`).
- Flickering: let the torch's radius vary a little with time.
