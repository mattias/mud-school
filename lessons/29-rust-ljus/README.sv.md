# Lektion 29 – Ljus och liv

Fönstret från lektion 28 visar världen, men den är platt: varje ruta lika
ljus, vattnet fruset, inget händer när du blir träffad. Spel som *Brogue*
får sin stämning från **ljus** – en fackla i en mörk grotta, ljus i olika
färger som blandas – och från små rörelser.

```
grotta, omgivningsljus (30, 30, 45)     du bär en fackla (255, 190, 110), radie 6

   . . . . . . . . . .                  . . . ░ ▒ ▓ ▒ ░ . .
   . . . . . . . . . .        ->        . . ░ ▒ ▓ @ ▓ ▒ ░ .
   . . . . . . . . . .                  . . . ░ ▒ ▓ ▒ ░ . .
```

## Mål

- Modellera ljus som tal: färger som **adderas** (ljuskällor) och
  **multipliceras** (ljus på en yta).
- Skriva animation som **rena funktioner av tiden** – `shimmer(x, y, t)` – så
  att den kan testas vid vilket ögonblick som helst utan att vänta.
- Bygga ett litet partikelsystem: skapa, flytta, blekna, ta bort.
- Använda **standardmetoder** (default methods) i en trait.

## Modellen

Ljuset på en ruta är terrängens omgivningsljus plus varje ljuskälla, svagare
med avståndet:

```
ljus = omgivning + Σ färg × max(0, 1 − avstånd / radie)        (varje kanal högst 255)
det du ser = rutans färg × ljus / 255                          (shade, lektion 28)
```

Ett rött och ett grönt ljus blir gult – precis som med riktigt ljus. Rutor du
har utforskat men inte ser är fortfarande `FOG`.

### Animation utan timrar

`shimmer(x, y, t_ms)` är ett tal mellan 0,7 och 1,0 som rör sig som en
långsam våg: `0.85 + 0.15 × sin(t/400 + 0.7x + 0.3y)`. Inget sparas och inget
räknar ner – ge den en tid så svarar den. Det gör den trivial att testa
(`build_scene_at(app, w, h, start + 500 ms)`) och omöjlig att få i otakt.

Tiden kommer från `App::started` (sätts i `App::new()`): scenen använder
`now − started`.

### Partiklar

Partiklar är undantaget: de *rör sig*, så de har tillstånd. `Particles`
har en lista; `burst` lägger till en ring av gnistor, `update(dt)` flyttar
dem, drar dem lite nedåt, räknar ner deras liv, och `retain` tar bort de
döda. `App::tick(now)` anropar `update` med tiden sedan förra tick – fönstret
anropar den varje bildruta.

### Standardmetoder

```rust
pub trait Renderer {
    fn cell(&mut self, x: u16, y: u16, cell: &SceneCell);
    fn text(&mut self, x: u16, y: u16, text: &str, fg: Rgb);
    fn particle(&mut self, _x: f32, _y: f32, _color: Rgb, _alpha: f32) {}
}
```

`particle` har en kropp: renderers som inte kan rita gnistor –
`TextRenderer`, renderers i dina tester – behöver inte implementera den. Att
lägga till en metod i en trait *med* ett standardbeteende förstör ingen
befintlig implementation.

## Kontraktet

`light.rs` (delvis given):

- `Light {x, y, radius, color}`, `TORCH = Rgb(255, 190, 110)`.
- `ambient(terrain)` (given): mörkt i grottor, ljust i städer, fullt ljus (255) för `""`.
- `light_at(x, y, lights, ambient) -> Rgb` – som ovan; (x, y) är en ruta,
  ljusens positioner är också rutor.
- `lights_for(zone)` – en fackla (radie 6) för varje spelarentitet.
- `shimmer(x, y, t_ms) -> f32` – som ovan. `scale(färg, k)` (given).
- `Particle {x, y, vx, vy, life, max_life, color}` med `alpha()` = life /
  max_life; `Particles {list}` med `burst(x, y, n, färg)` (given) och `update(dt)`.
- `shake_offset(flash_until, now) -> (f32, f32)` – `(0, 0)` när det inte
  blinkar; annars en liten förskjutning (högst en halv ruta) som skakar fram och tillbaka.

`scene.rs`:

- `cell_light(app, x, y, now)`: `FOG` om rutan inte syns; annars
  `light_at(x, y, lights_for(zone), ambient(terrängen från statusen))`, och
  vatten multiplicerat med `shimmer`. Utan zon: `FULL_LIGHT`.
- `build_scene_at(app, width, height, now)` – som `build_scene`, med
  `light = cell_light(...)`, `particles` (i vykoordinater, med alfa) och
  `shake`. `build_scene(...)` = `build_scene_at(..., Instant::now())`.
- `Renderer::particle` med en tom standardkropp; `draw_scene` ritar
  partiklarna efter cellerna (en rad ner, som cellerna).

`app.rs`: `App::particles`, `App::last_tick`, `App::started` (sätts i
`new()`), `App::tick(now)`; en träff (lägre `hp`) ger 12 röda gnistor mitt i
din ruta (`x + 0.5, y + 0.5`). Terminalen använder också `cell_light`.

`gfx.rs` och `mud-gfx.rs`: `particle` ritar en liten kvadrat, `shake` flyttar
zonen, och loopen anropar `app.tick(now)` och `build_scene_at(..., now)`.

## Uppgift

`light.rs`, sedan `scene.rs`, sedan `app.rs` och fönstret. Gå in i grottan
(`grottan`) med `./mudschool run gfx` – och låt råttan bita dig.

## Fundera på

- Varför *adderas* ljuskällorna men ljuset på en yta *multipliceras*?
- `light_at` går igenom alla ljus för varje ruta. Med 50 facklor i en 200×200-zon?
- Ljuset går genom väggar just nu. Hur skulle du använda synfältet från
  lektion 26 för att fixa det – och var ska den koden köra?
- Partiklar finns bara i klienten. Tänk om två spelare ska se samma explosion?

## Utmaningar

- Ljus genom väggar: räkna ut `ComputeFOV` från varje ljuskälla (i klienten) och lys bara upp det den ser.
- Facklor som saker: ett fält `glow` på saker i protokollet, så att en släppt fackla fortsätter lysa upp rummet.
- Dag och natt: låt omgivningsljuset följa spelklockan från lektion 14 (`time`).
- Fladder: låt facklans radie variera lite med tiden.
