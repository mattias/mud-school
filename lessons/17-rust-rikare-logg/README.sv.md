# Lektion 17 – En rikare logg

Sedan lektion 15 säger servern *vilken sorts* text den skickar:
`{"type":"text","text":"Bertil säger: hej","kind":"say"}`. Nu använder
klienten det: tal i blått, fel i rött, rumsbeskrivningen ljus, folk som kommer
och går i grått. Dessutom två saker som alla MUD-klienter har: en
**kommandohistorik** (pil upp/ner) och **scrollning** (PageUp/PageDown).

## Mål

- Utöka en enum och ett protokoll bakåtkompatibelt (`#[serde(default)]`, `#[serde(other)]`).
- Använda `#[derive(Default)]` med `#[default]` på en enum-variant.
- Modellera små tillståndsmaskiner (historikposition, scrollning) och testa dem.
- Arbeta säkert med slices och index: `saturating_sub`, `min`.

## Rust-koncept

### En enum med standardvärde

```rust
#[derive(Debug, Clone, Copy, PartialEq, Eq, Default, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum Kind {
    #[default]
    Info,          // Kind::default() == Kind::Info
    Room, Say, Move, Error, Combat,
    #[serde(other)]
    Other,         // "kind":"poetry" från en nyare server
}
```

Med `#[serde(default)] kind: Kind` på text-varianten fungerar en server från
lektion 14 (som inte skickar något `kind`) fortfarande. Det är samma
kompatibilitetsregel som du följde på serversidan.

### Säker indexaritmetik

```rust
let end = app.log.len().saturating_sub(app.scroll);   // aldrig under 0
let start = end.saturating_sub(visible);
let lines = &app.log[start..end];
```

`usize` kan inte vara negativ: `3 - 5` ger panik i debug-byggen.
`saturating_sub` stannar på 0 istället.

## Kontraktet

- `protocol.rs`: `pub enum Kind { Info, Room, Say, Move, Error, Combat, Other }`
  (standard `Info`, gemener i JSON, okänd → `Other`), och
  `ServerMsg::Text { text: String, kind: Kind }` (`kind` är som standard `Info`).
  `parse_line` av vanlig text ger `Kind::Info`.
- `app.rs`:
  - `pub struct LogLine { pub kind: Kind, pub text: String }`, och
    `App::log: Vec<LogLine>`; `push_log(&mut self, kind: Kind, text: &str)`.
  - `pub fn style_for(kind: Kind) -> Style` – en egen stil per kind (`Room`,
    `Say`, `Move`, `Error`, `Combat`); `Other` ser ut som `Info`. Loggen
    ritar varje rad med sin stil.
  - Historik: `pub history: Vec<String>` (äldst först, högst `MAX_HISTORY`,
    samma kommando två gånger i rad sparas en gång). Upp visar äldre
    kommandon, ner nyare, och efter det nyaste får du tillbaka det du höll på
    att skriva.
  - Scrollning: `pub scroll: usize` (rader upp från botten), `pub const PAGE: usize = 10`.
    PageUp/PageDown flyttar `PAGE` rader, End och Enter går tillbaka till
    botten. Medan du har scrollat upp flyttar nya meddelanden inte vyn.

## Uppgift

1. Uppdatera `protocol.rs` och `app.rs` enligt ovan. (`main.rs` behöver också
   en liten ändring: `push_log` tar nu ett kind.)
2. Visa i loggens rubrik att du har scrollat upp, t.ex. `Logg ↑10`.

## Testa själv

Kör servern från lektion 16 och din klient. Gå runt, prata, skriv nonsens –
färgerna ska berätta vad som händer. Tryck pil upp några gånger.

## Fundera på

- Varför är `Kind` `Copy` men inte `LogLine`?
- Vad händer med `Kind::Other` om en senare serverversion lägger till `"kind":"whisper"`?
- Historik och scrollning testas utan terminal. Vad mer i klienten skulle
  kunna testas så?

## Utmaningar

- Markera ditt eget namn i loggen.
- Tab-komplettering av kommandon (`lo<Tab>` → `look`).
- Spara historiken till en fil när klienten avslutas och läs in den vid start.
