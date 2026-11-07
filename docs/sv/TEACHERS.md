# För lärare och kursförfattare

> [🇬🇧 English](../en/TEACHERS.md) · 🇸🇪 Svenska · [🇯🇵 日本語](../ja/TEACHERS.md)

Så här är kursen uppbyggd och så lägger du till eller ändrar lektioner.

## Katalogstruktur

```
mudschool                    skript som bygger och kör kursverktyget
tools/mudschool/             kursverktyget (Go, bara standardbiblioteket)
starter/                     koden studenten börjar med (kopieras till workspace/)
  server/                    Go-modul "mudserver"
  client/                    Cargo-paket "mud-client" (biblioteket heter mud_client)
starter.<språk>/             språkberoende startfiler: messages.go och messages.rs
lessons/
  _shared/go/...             filer som kopieras in vid VARJE Go-lektion (testhjälpare)
  _shared/rust/...           motsvarande för Rust (tests/mudschool_common/)
  NN-namn/
    lesson.json              titel och sammanfattning på alla språk, spår ("go"/"rust")
    README.md                instruktionerna på engelska (README.sv.md, README.ja.md)
    HINTS.md                 ledtrådar, en per "## "-rubrik (HINTS.sv.md, HINTS.ja.md)
    scaffold/                NYA filer studenten får när lektionen börjar
    scaffold.<språk>/        språkberoende nya filer (t.ex. världen)
    tests/                   tester, som läggs in när lektionen börjar (och igen vid `check`)
    solution/                referenslösningen: filer som de ser ut EFTER lektionen
    solution.<språk>/        språkberoende delar av lösningen (sällan behövs)
workspace/                   studentens kod (ignoreras av git)
docs/<språk>/                kursplan, protokoll, installation, idéer, detta dokument
```

Alla sökvägar under `scaffold/`, `tests/` och `solution/` är relativa till
`workspace/`, t.ex. `tests/server/mud/mudschool_04_test.go`.

## Språk

Kursen finns på svenska (`sv`), engelska (`en`) och japanska (`ja`).

| Vad | Hur det översätts |
|---|---|
| Instruktioner och ledtrådar | `README.<språk>.md`, `HINTS.<språk>.md` (engelska, standardspråket, är `README.md`) |
| Titel och sammanfattning | objekt i `lesson.json`: `{"sv": ..., "en": ..., "ja": ...}` |
| Kursverktygets texter | `L(sv, en, ja)` i `tools/mudschool` |
| Spelets texter | `starter.<språk>/server/mud/messages.go` och `.../client/src/messages.rs` |
| Spelvärlden | `scaffold.<språk>/` i lektion 08 och 09 (rum), 15 (`items.json`) och 16 (`npcs.json`) |
| Testernas felmeddelanden | `L(sv, en, ja)` (Go) och `l(sv, en, ja)` (Rust) |
| Dokumentation | `docs/<språk>/` |
| Kod och kodkommentarer | **engelska** – gemensamt för alla |

Språket väljs med `./mudschool init --lang en|sv|ja` (annars gissas det
från `$LANG`) och kan bytas med `./mudschool lang <språk>`. Det sparas i
`workspace/.mudschool.json` och skickas till testerna som `MUDSCHOOL_LANG`.

Testerna jämför aldrig med hårdkodad text. De använder samma konstanter som
studentens kod (`MsgAskName`, `fmt.Sprintf(MsgSays, "Anna", "hej")`), så de
fungerar på alla språk – och studenten får gärna ändra texterna.

Texterna kan studenten ändra fritt. Byter hen språk med `./mudschool lang`
sparas en ändrad `messages.go`/`messages.rs` först som `.bak`.

## Hur kursverktyget arbetar

- **`init`** kopierar `starter/` och `starter.<språk>/` till `workspace/` och
  går in i första lektionen.
- **Att gå in i en lektion** kopierar `scaffold/` och `scaffold.<språk>/` –
  men bara filer som inte redan finns. Studentens kod skrivs aldrig över.
  Lektionens tester läggs också in direkt, så att README kan hänvisa till dem.
  Tills lektionen är löst markerar editorn de nya testerna (odefinierade namn)
  – det är väntat; `run server`/`run client` påverkas inte.
- **`check`**:
  1. tar bort alla filer och kataloger i `workspace/` vars namn börjar på
     `mudschool_` (gamla kurstester),
  2. kopierar in `lessons/_shared/<spår>/` och lektionens `tests/`,
  3. kör `go vet ./... && go test -count=1 ./...` i `workspace/server`
     (Go) eller `cargo test` i `workspace/client` (Rust),
  4. vid grönt: listar kvarglömda `TODO(lesson NN)`-markeringar för lektionen,
     markerar den klar och går in i nästa.
- **`doctor`** kontrollerar att Go, Rust och verktygen finns.

Bara den aktuella lektionens tester körs. Det är avsiktligt: beteendet
ändras mellan lektioner (ekot från lektion 01 försvinner i lektion 04).

### Namnkonvention för tester

Alla testfiler **måste** börja med `mudschool_` – annars städas de inte bort
och gamla tester börjar fallera. Go: `mudschool_NN_test.go` i paketet `mud`.
Rust: `tests/mudschool_NN.rs` (integrationstester mot biblioteket `mud_client`),
som börjar med `mod mudschool_common;` för att få `l(...)`.

## Uppdatera studenternas arbetsytor

Verktyget sparar en checksumma för varje fil det kopierar in från `starter/` och
`scaffold/` (`files` i `.mudschool.json`). `./mudschool update` jämför: oförändrade
filer byts mot kursens nuvarande version, filer som aldrig funnits i arbetsytan
läggs till (en fil som studenten tagit bort eller döpt om lämnas ifred), en fil som både
studenten och kursen har ändrat får kursens version som `<fil>.new`, och en ändrad
`messages.go`/`messages.rs` säkerhetskopieras som `.bak` och byts ut (nya lektioner
behöver dess nya konstanter). Arbetsytor från innan registret fanns jämförs med
alla versioner av filen i git-historiken; en fil som inte matchar någon är
studentens egen kod och lämnas ifred. `check` gör de säkra uppdateringarna
själv. Alltså: att rätta en scaffold- eller startfil går bra – studenterna får
rättelsen med `git pull`.

## Verifiering

```
./mudschool verify              # alla lektioner, alla språk
./mudschool verify --lang ja    # bara ett språk
./mudschool verify 07 08        # bara några lektioner (alla byggs upp i ordning)
```

`verify` bygger upp en arbetsyta i en temporär katalog: `starter/`, sedan
för varje lektion `scaffold/` och `solution/` i tur och ordning – en gång per
språk. För varje lektion kontrolleras att

1. testerna **misslyckas** med koden *före* lektionen (förra lösningen +
   scaffold) – annars testar lektionen inget nytt, och
2. testerna **går igenom** med lektionens lösning.

`./mudschool reference <katalog> [id]` bygger hela referenslösningen (till och
med en viss lektion) så att du kan köra den. Kör `verify` innan du checkar in;
CI (`.github/workflows/verify.yml`) gör samma sak.

## Skärmbilder

`docs/images/client-<språk>.png` genereras från referenslösningen med
`tools/screenshots/make.sh` (kräver tmux, nc, Node med Playwright och ett
japanskt typsnitt som Noto CJK). Kör det igen när klientens utseende ändras.
Använd bilder istället för textdiagram när japansk text och ramtecken måste
ligga i linje – GitHubs typsnitt ritar inte CJK-tecken exakt två kolumner breda.

## Skriva en ny lektion

1. Skapa `lessons/NN-spår-namn/` med `lesson.json`, `README*.md`, `HINTS*.md`.
2. Skriv `solution/`: **hela** filer, som de ska se ut efter lektionen.
   Filer som inte ändras behöver inte vara med – de ärvs från tidigare lektioner.
3. Ny spelartext? Lägg till konstanten i **alla tre** `messages.go`
   (eller `messages.rs`) i `starter.<språk>/`.
4. Skriv `tests/`. Tänk på:
   - Felmeddelanden med `L(sv, en, ja)` som säger vad som förväntades.
   - Jämför med konstanterna i `messages.go`, aldrig med hårdkodad text.
   - Testa beteende, inte implementation. Servertester går helst över TCP med
     hjälparna i `_shared/go/server/mud/mudschool_helpers_test.go`
     (`startServer`, `login`, `expect`, `expectContains`, `expectSilence`, ...).
   - Lagom stränga: testa det lektionen lär ut, ge frihet i resten.
   - Inga sovtider där det går att undvika; använd tidsgränser.
5. Skriv `scaffold/` för nya filer: typer och signaturer med `TODO(lesson NN)`
   och `todo!()`/tomma returvärden så att koden kompilerar.
6. Kör `./mudschool verify NN`.

### README-mallen

```
# Lektion NN – Titel
(Motivation: vilket problem i SPELET löser vi?)
## Mål
## <Språk>-koncept     (korta exempel, gärna en "Go ↔ Rust"-jämförelse)
## Uppgift             (numrerad, fil för fil, med konstanterna testerna kräver)
## Testa själv
## Fundera på          (2–3 reflektionsfrågor)
## Utmaningar          (frivilligt, påverkar inte testerna)
```

### Ledtrådar

Varje `## `-rubrik i `HINTS.md` är en ledtråd. Ordna dem från vag till
konkret; sista ledtråden får gärna vara nästan-kod.
