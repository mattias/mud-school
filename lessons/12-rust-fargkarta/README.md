# Lektion 12 – Kartan i färg

Nu blir det snyggt. Dwarf Fortress, Caves of Qud, NetHack och Brogue visar
hela världar med *ett tecken per ruta* – men med väl valda tecken och färger
blir det otroligt stämningsfullt. Vi gör samma sak: kartans koder blir
färgade **tiles**.

```
    Ω─Ω                  ♣ grön skog, ≈ blått vatten (med mörkblå bakgrund),
    │ │                  ▲ grått berg, # stenhus, · väg, @ du (gul)
    ▲ Ω
    │
≈─"─·─♣─♣
      │ │
      · ♣
```

Det här är också första steget mot riktig tile-grafik i fas 4: där byter vi
`Cell { glyph: '♣', ... }` mot "ruta nummer 5 i spritearket" – resten av
koden blir likadan.

## Mål

- Skapa små **Copy-typer** (`Rgb`, `Cell`) som kopieras istället för att flyttas.
- Använda `match` på strängar med en reservgren `_`.
- Använda **iteratorer**: `iter`, `chars`, `map`, `collect`, `enumerate`.
- Förstå hur terminaler ritar färg med **ANSI escape-koder**.

## Rust-koncept

### Copy

```rust
#[derive(Clone, Copy, PartialEq)]
pub struct Rgb(pub u8, pub u8, pub u8);   // "tuple struct": fält utan namn

let a = Rgb(255, 0, 0);
let b = a;          // KOPIA – a går fortfarande att använda (tack vare Copy)
let Rgb(r, g, b) = a;   // destrukturering
```

Utan `Copy` hade `let b = a;` *flyttat* värdet. Små värden som bara består av
siffror är perfekta som `Copy`. `String` och `Vec` kan aldrig vara det.

### Iteratorer

```rust
let grid: Vec<Vec<Cell>> = map.rows
    .iter()                                  // &String för varje rad
    .map(|row| row.chars().map(to_cell).collect())  // Vec<Cell> per rad
    .collect();                              // Vec<Vec<Cell>>
```

Iteratorer är *lata*: inget händer förrän `collect` (eller en `for`-loop)
drar i dem. Kompilatorn gör om kedjan till lika snabb kod som en handskriven
loop.

**Go ↔ Rust:** i Go skriver man nästan alltid `for`-loopar. I Rust är båda
stilarna vanliga – välj det som är tydligast.

### ANSI-koder

Terminaler tolkar vissa teckensekvenser som kommandon:

| Kod | Betydelse |
|---|---|
| `\x1b[38;2;R;G;Bm` | förgrundsfärg (truecolor) |
| `\x1b[48;2;R;G;Bm` | bakgrundsfärg |
| `\x1b[0m` | återställ |

Prova: `printf '\x1b[38;2;40;170;70m♣\x1b[0m\n'`

Alla terminalbibliotek (crossterm, ratatui, ncurses) gör detta under huven.
Nu gör du det själv en gång – i nästa lektion låter vi ett bibliotek göra det.

## Uppgift

1. **`src/render.rs`** (ny fil): skriv `tile`, `render_map` och `to_ansi`.
   Lägg till `pub mod render;` i `lib.rs`.
   - `tile` ger varje betydelse ett eget tecken och en färg som passar
     (testerna kollar t.ex. att skog är grön och vatten blått).
     Okända betydelser ger en tydlig reservruta, t.ex. magenta `?`.
   - `render_map`: mellanslag blir `BLANK`; en kod som finns i legenden blir
     `tile(betydelse)`; en kod som *inte* finns visas som sig själv i grått.
   - `to_ansi` enligt kommentaren i filen.
2. **`src/main.rs`**: rita kartan med `to_ansi(&render_map(&map))`.

## Testa själv

Kör servern och klienten och gå runt: `s`, `s`, `s`, `s`, `e` ...

> Ser du konstiga tecken? Din terminal måste använda UTF-8 och ett typsnitt
> med tecknen. Ser du inga färger? Terminalen saknar kanske truecolor-stöd
> (de flesta moderna har det).

## Fundera på

- Varför tar `to_ansi` en `&[Vec<Cell>]` och inte en `Vec<Vec<Cell>>`?
- `tile` returnerar ett nytt `Cell` varje gång. Varför är det inget problem
  för prestandan?

## Utmaningar

- Variera skogen: låt `render_map` välja mellan `♣`, `♠` och `↑` baserat på
  positionen (`enumerate` ger dig x och y), så att skogen ser levande ut.
  Det är ett vanligt knep i roguelikes.
- Läs om *färgteori för roguelikes*: mörkare färger för saker längre bort,
  en färgpalett med begränsat antal toner (t.ex. "PICO-8"-paletten).
