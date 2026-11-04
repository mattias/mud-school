# För lärare och kursförfattare

Så här är kursen uppbyggd och så lägger du till eller ändrar lektioner.

## Katalogstruktur

```
mudschool                    skript som bygger och kör kursverktyget
tools/mudschool/             kursverktyget (Go, bara standardbiblioteket)
starter/                     koden studenten börjar med (kopieras till workspace/)
  server/                    Go-modul "mudserver"
  client/                    Cargo-paket "mud-client" (biblioteket heter mud_client)
lessons/
  _shared/go/...             filer som kopieras in vid VARJE Go-lektion (testhjälpare)
  _shared/rust/...           motsvarande för Rust
  NN-namn/
    lesson.json              titel, spår ("go"/"rust"), sammanfattning, begrepp
    README.md                instruktionerna (visas med `mudschool show`)
    HINTS.md                 ledtrådar, en per "## "-rubrik (visas en i taget)
    scaffold/                NYA filer studenten får när lektionen börjar
    tests/                   tester som kopieras in vid `mudschool check`
    solution/                referenslösningen: filer som de ser ut EFTER lektionen
workspace/                   studentens kod (ignoreras av git)
docs/                        kursplan, protokoll, idéer, detta dokument
```

Alla sökvägar under `scaffold/`, `tests/` och `solution/` är relativa till
`workspace/`, t.ex. `tests/server/mud/mudschool_04_test.go`.

## Hur kursverktyget arbetar

- **`init`** kopierar `starter/` till `workspace/` och går in i första lektionen.
- **Att gå in i en lektion** kopierar `scaffold/` – men bara filer som inte
  redan finns. Studentens kod skrivs aldrig över.
- **`check`**:
  1. tar bort alla filer och kataloger i `workspace/` vars namn börjar på
     `mudschool_` (gamla kurstester),
  2. kopierar in `lessons/_shared/<spår>/` och lektionens `tests/`,
  3. kör `go vet ./... && go test -count=1 ./...` i `workspace/server`
     (Go) eller `cargo test` i `workspace/client` (Rust),
  4. vid grönt: markerar lektionen klar och går in i nästa.
- Framsteg sparas i `workspace/.mudschool.json`.

Bara den aktuella lektionens tester körs. Det är avsiktligt: beteendet
ändras mellan lektioner (ekot från lektion 01 försvinner i lektion 04).

### Namnkonvention för tester

Alla testfiler **måste** börja med `mudschool_` – annars städas de inte bort
och gamla tester börjar fallera. Go: `mudschool_NN_test.go` i paketet `mud`.
Rust: `tests/mudschool_NN.rs` (integrationstester mot biblioteket `mud_client`).

## Verifiering

```
./mudschool verify          # alla lektioner
./mudschool verify 07 08    # bara några (men alla byggs upp i ordning)
```

`verify` bygger upp en arbetsyta i en temporär katalog: `starter/`, sedan
för varje lektion `scaffold/` och `solution/` i tur och ordning. För varje
lektion kontrolleras att

1. testerna **misslyckas** med koden *före* lektionen (förra lösningen +
   scaffold) – annars testar lektionen inget nytt, och
2. testerna **går igenom** med lektionens lösning.

Kör `verify` innan du checkar in. CI (`.github/workflows/verify.yml`) gör
samma sak.

## Skriva en ny lektion

1. Skapa `lessons/NN-spår-namn/` med `lesson.json`, `README.md`, `HINTS.md`.
2. Skriv `solution/`: **hela** filer, som de ska se ut efter lektionen.
   Filer som inte ändras behöver inte vara med – de ärvs från tidigare lektioner.
3. Skriv `tests/`. Tänk på:
   - Felmeddelanden på svenska som säger vad som förväntades.
   - Testa beteende, inte implementation. Servertester går helst över TCP med
     hjälparna i `_shared/go/server/mud/mudschool_helpers_test.go`
     (`startServer`, `login`, `expect`, `expectContains`, `expectSilence`, ...).
   - Lagom stränga: testa det lektionen lär ut, ge frihet i resten.
   - Inga sovtider där det går att undvika; använd tidsgränser.
4. Skriv `scaffold/` för nya filer: typer och signaturer med `TODO(lektion NN)`
   och `todo!()`/tomma returvärden så att koden kompilerar.
5. Kör `./mudschool verify NN`.

### README-mallen

```
# Lektion NN – Titel
(Motivation: vilket problem i SPELET löser vi?)
## Mål
## <Språk>-koncept     (korta exempel, gärna en "Go ↔ Rust"-jämförelse)
## Uppgift             (numrerad, fil för fil, med exakta texter testerna kräver)
## Testa själv
## Fundera på          (2–3 reflektionsfrågor)
## Utmaningar          (frivilligt, påverkar inte testerna)
```

### Ledtrådar

Varje `## `-rubrik i `HINTS.md` är en ledtråd. Ordna dem från vag till
konkret; sista ledtråden får gärna vara nästan-kod.
