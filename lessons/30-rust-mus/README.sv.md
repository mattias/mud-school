# Lektion 30 – Musen

Fas 4:s sista lektion ger fönstret en mus:

- **klicka** någonstans så går du dit – klienten hittar vägen själv (A*, som
  serverns i lektion 27, nu i Rust) och skickar ett steg i taget;
- **peka** på något så säger ett verktygstips vad det är: *"en stor råtta"*, *"Ett träd"*;
- **högerklicka** så visar en meny vad du kan göra: *Titta på en stor råtta*,
  *Anfall en stor råtta*, *Gå hit*.

```
#########
#@..#...#        högerklick på råttan:
#...#.r.#           ┌──────────────────────────┐
#.......#           │ Titta på en stor råtta   │  -> "examine rat"
#.....*.#           │ Anfall en stor råtta     │  -> "kill rat"
#########           │ Gå hit                   │  -> gå (i klienten)
                    └──────────────────────────┘
```

## Mål

- Implementera A* igen – i Rust, med `BinaryHeap` och `Reverse`.
- Räkna om mellan koordinatsystem: pixlar → skärmrutor → zonrutor.
- Göra om en följd av handlingar till kommandon som skickas över tid (en kö + en klocka).
- Hålla "vad kan jag göra med det här?" som data (en meny av `MenuItem`) –
  lätt att testa, lätt att rita i vilken renderer som helst.

## Rust-koncept

### `BinaryHeap` och `Reverse`

Rusts prioritetskö ger dig det **största** elementet först. A* behöver det
minsta `f`, så slå in nyckeln i `std::cmp::Reverse`, som vänder på ordningen:

```rust
let mut open = BinaryHeap::new();
open.push(Reverse((f, g, cell)));            // tupler jämförs fält för fält
while let Some(Reverse((_, g, cur))) = open.pop() { ... }
```

Jämför med Go:s `container/heap` i lektion 27: där skrev du `Less`; här
kommer `Ord` gratis med tupler av tal.

### En kö av steg

`walk_to` gör om en väg till kommandon (`"southeast"`, `"east"`, …) i en
`VecDeque`. `next_walk_step(now)` delar ut ett i taget, högst ett per
`WALK_STEP` – klockan skickas in, som i lektion 19, så att testerna bestämmer
vad klockan är. Varje tangenttryckning tömmer kön: spelaren har tagit över.

### Koordinater

```
pixlar (mouse_position)  --screen_to_cell-->  skärmruta (x, y)
skärmruta                --Scene::zone_cell-->  zonruta (origin + (x, y − 1))
```

Scenen minns vilken zonruta den visade längst upp till vänster (`origin`) –
kameran från lektion 25 – så att ett klick kan räknas om till en zonruta.

## Kontraktet

`mouse.rs` (ny):

- `walkable(meaning)` – golv, gräs, dörr, trappor; `STEPS` – de åtta
  riktningarna med sina kommandon.
- `step_command(from, to)` – kommandot för ett steg till en grannruta,
  `None` om de inte är grannar.
- `find_path(zone, from, to)` – den kortaste vägen (8 riktningar) över rutor
  klienten vet är gångbara, utan `from`, med `to`; `None` om `to` inte är
  gångbar eller inte går att nå; `Some(vec![])` när `from == to`.
- `describe_cell(zone, x, y)` – entitetens namn, annars
  `messages::tile_name(meaning)`, annars `None` (aldrig sedd).
- `context_menu(zone, x, y) -> Vec<MenuItem {label, command}>`: en varelse →
  `LOOK_AT` + namn / `examine <id>` och `ATTACK` + namn / `kill <id>`; en sak →
  `LOOK_AT` / `examine <id>` och `PICK_UP` / `get <id>`; sedan, om rutan är
  gångbar, `WALK_HERE` / `WALK` (`"@walk"`, hanteras av klienten). `<id>` är
  entitetens ID utan prefix (`n:rat` → `rat`) – servern förstår det alltid.
- `screen_to_cell(mx, my, cell_px)`.

`messages.rs`: `LOOK_AT`, `ATTACK`, `PICK_UP`, `WALK_HERE`, `tile_name(meaning)`.

`app.rs`: `walk: VecDeque<String>`, `next_step`, `hover`, `menu: Option<Menu
{cell, items}>`, `WALK_STEP` (150 ms), och:
- `walk_to(cell)` – vägen från dig själv, som stegkommandon;
- `next_walk_step(now)` – nästa steg, högst ett per `WALK_STEP`;
- `left_click(cell, menu_row)` – med en öppen meny: välj rad `menu_row`
  (returnerar dess kommando, eller börjar gå för `WALK`) och stäng menyn;
  utan meny: gå till rutan;
- `right_click(cell)` – öppnar rutans meny (ingen om den är tom);
- en tangenttryckning stoppar promenaden och stänger menyn.

`scene.rs`: `Scene::origin`, `Scene::overlay` (verktygstipset för `hover`,
eller menyns texter, en per rad från rutan och nedåt), `Scene::zone_cell(x, y)`,
och `Renderer::overlay` med en standardkropp (vanlig text).

Fönstret (`gfx/main.rs`, `gfx.rs`): hover, vänster- och högerklick, ett steg
per `WALK_STEP`, och en overlay med en mörk ruta bakom texten.

## Uppgift

`mouse.rs` först – testerna kontrollerar den för sig. Sedan App, scenen och
fönstret. Gå omkring med musen i `./mudschool run gfx`.

## Fundera på

- Klienten hittar vägar över de rutor *den känner till*. När kan vägen vara
  fel, och vad händer då? (Tips: outforskade rutor, en varelse i vägen.)
- Varför räknas vägen ut i klienten i stället för att servern skickar den?
- Menyns kommandon är strängar som `"kill rat"`. Vilka är för- och
  nackdelarna jämfört med ett eget protokoll för handlingar?
- Hur skulle du få samma mus att fungera i terminalklienten? (crossterm har
  `EnableMouseCapture` och `Event::Mouse`.)

## Utmaningar

- Musen i terminalklienten också – App-logiken finns redan.
- Visa den planerade vägen medan du pekar (en svag linje av prickar).
- Skift-klick för att anfalla direkt, dubbelklick för att plocka upp.
- Stoppa promenaden när en varelse kommer inom synhåll.

**Fas 4 är klar.** MUD:en har nu en värld av rutor genererad från frön, ett
synfält, jagande varelser och en grafisk klient med ljus och mus. Fas 5 gör
servern utbyggbar utan omstart.
