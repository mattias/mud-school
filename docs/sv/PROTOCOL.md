# Protokollet – mudschool 1

> [🇬🇧 English](../en/PROTOCOL.md) · 🇸🇪 Svenska · [🇯🇵 日本語](../ja/PROTOCOL.md)

Det här är kontraktet mellan servern (Go) och klienten (Rust). Det införs i
lektion 10 (servern) och 11 (klienten).

## Grundregler

- Transport: TCP – eller en WebSocket på `/ws` på webbklientens adress
  (lektion 39), vars textmeddelanden bär samma rader (ett meddelande kan
  innehålla en del av en rad, eller flera). All text är **UTF-8**.
- Varje meddelande är **en rad** som avslutas med `\r\n` (CR LF), åt båda hållen.
- Klienten skickar alltid vanlig text: kommandon som `look` eller `say hej`.
- Servern skickar **vanlig text** tills klienten har skickat en hälsning.
  Därefter skickar servern **en JSON-frame per rad**.

Det gör att `telnet` och `nc` fungerar precis som förut, medan en klient som
förstår protokollet får strukturerad data.

## Handskakning

Klienten skickar, när som helst (vanligen direkt efter anslutningen):

```
@hello mudschool 1
```

Servern svarar med en `hello`-frame och skickar från och med nu JSON:

```json
{"type":"hello","text":"mudschool 1"}
```

Rader som börjar med `@hello` tolkas aldrig som spelkommandon.

## Frames

Varje frame är ett JSON-objekt med fältet `type`.

### `text`

```json
{"type":"text","text":"Torget i Mudby\nEtt kullerstensbelagt torg."}
```

| Fält | Typ | Beskrivning |
|---|---|---|
| `text` | sträng | Text att visa. Kan innehålla `\n`. Utelämnas om texten är tom. |
| `kind` | sträng | Vilken sorts text det är, så att en klient kan färga den (servern skickar den från lektion 15; klienten använder den från lektion 17): `info`, `room`, `say`, `move`, `error`, `combat`. Saknas = `info`. **En okänd sort ska visas som `info`** – nya sorter kan komma. |

```json
{"type":"text","text":"Bertil säger: hej!","kind":"say"}
```

Texten är på serverns språk (se `messages.go`).

### `map`

Skickas efter varje `look` (alltså också efter förflyttning och inloggning)
och som svar på kommandot `map`.

```json
{"type":"map","map":{
  "width":5,"height":5,
  "rows":["  +  ","  |  ","  @-+","     ","\"    "],
  "legend":{"+":"inside","|":"path_ns","@":"self","-":"path_ew","\"":"field"}
}}
```

| Fält | Typ | Beskrivning |
|---|---|---|
| `width`, `height` | heltal | Kartans storlek i tecken. |
| `rows` | lista av strängar | `height` rader med `width` tecken. Varje tecken är en **kod**. Mellanslag = inget. |
| `legend` | objekt | Kod → betydelse, för de koder som används i `rows`. |

Rummen ligger på jämna rader och kolumner; däremellan ritas utgångar. Spelaren
står alltid i mitten.

Betydelser i version 1:

| Betydelse | Kod | Förklaring |
|---|---|---|
| `town` | `#` | gata, torg |
| `inside` | `+` | inomhus |
| `road` | `.` | väg |
| `field` | `"` | äng, fält |
| `forest` | `T` | skog |
| `water` | `~` | vatten |
| `mountain` | `^` | berg |
| `cave` | `o` | grotta |
| `stairs_up` | `<` | rum med utgång uppåt |
| `stairs_down` | `>` | rum med utgång nedåt |
| `path_ew` | `-` | utgång öster–väster |
| `path_ns` | `\|` | utgång norr–söder |
| `self` | `@` | du |

> **Viktigt:** klienten ska rita utifrån **betydelsen** (`legend`), inte koden.
> Koderna är valda för att se vettiga ut i telnet, men kan ändras.

### `status`

Skickas efter varje `look` och efter varje stridsrunda.

```json
{"type":"status","status":{"name":"Anna","room":"Torget i Mudby","terrain":"town","hp":10,"max_hp":10,"online":2}}
```

| Fält | Typ | Beskrivning |
|---|---|---|
| `name` | sträng | Spelarens namn |
| `room` | sträng | Rummets namn |
| `terrain` | sträng | Rummets terräng (en av betydelserna ovan) |
| `hp`, `max_hp` | heltal | Hälsa |
| `online` | heltal | Antal inloggade spelare |
| `target` | objekt | Vem du slåss mot (från lektion 18): `{"name":"en råtta","hp":3,"max_hp":5}`. Saknas = ingen. |

### `prompt`

Servern väntar på något speciellt (från lektion 20). Själva frågan kommer som
en vanlig `text`-frame strax före.

```json
{"type":"prompt","prompt":"password"}
```

| Fält | Typ | Beskrivning |
|---|---|---|
| `prompt` | sträng | `password` (ett befintligt konto) eller `new_password` (välj ett). Klienten ska **dölja** det som skrivs och inte spara det i historiken. Andra värden kan komma; behandla dem som vanlig inmatning. |

Telnet ser aldrig `prompt`-frames.

### `zone`

Hela zonen spelaren står i (från lektion 23). Skickas vid varje `look` – alltså
efter inloggning och i varje nytt rum – och som svar på `zone`.

```json
{"type":"zone","zone":{"room":"torget","width":9,"height":5,
  "rows":["####.####","#.......#","#..~....#","#.......#","#########"],
  "legend":{"#":"wall",".":"floor","~":"water"},
  "entities":[{"id":"p:Anna","kind":"player","name":"Anna","glyph":"@","x":4,"y":2},
              {"id":"i:lamp","kind":"item","name":"a lamp","glyph":"*","x":5,"y":2}],
  "self":"p:Anna",
  "visible":["111111111","111111111","111111111","111111111","111111111"]}}
```

| Fält | Typ | Beskrivning |
|---|---|---|
| `room` | sträng | Rummets ID |
| `width`, `height` | heltal | Storleken i rutor |
| `rows` | lista av strängar | `height` rader med `width` rutkoder. Ett mellanslag = inte utforskat än (från lektion 26). |
| `legend` | objekt | Kod → betydelse: `floor`, `wall`, `grass`, `tree`, `water`, `rock`, `door`, `stairs_up`, `stairs_down`. Klienten ritar efter betydelsen. |
| `entities` | lista | Allt med en position som spelaren ser: `id` (`p:<namn>`, `n:<id>`, `i:<id>`), `kind` (`player`, `creature`, `item`), `name`, `glyph` (för textklienter), `x`, `y`. |
| `self` | sträng | Mottagarens eget entitets-ID |
| `visible` | lista av strängar | Synfältet: `'1'` synligt nu, `'0'` inte (från lektion 26). Saknas = allt synligt. |

Telnet får zonen ritad som text (med entiteternas tecken) efter varje steg.

### `delta`

Det som ändrats i zonen sedan förra framen (från lektion 23) – skickas efter
ett kommando eller ett tick om något ändrats. Mycket mindre än en hel zon.

```json
{"type":"delta","delta":{"room":"torget",
  "moved":[{"id":"p:Anna","kind":"player","name":"Anna","glyph":"@","x":5,"y":2}],
  "removed":["n:cat"]}}
```

| Fält | Typ | Beskrivning |
|---|---|---|
| `room` | sträng | Rummet den gäller. **En delta för ett annat rum ska ignoreras.** |
| `added`, `moved` | lista av entiteter | Nya och flyttade entiteter (hela entiteter, som i `zone`) |
| `removed` | lista av strängar | ID:n som försvunnit (eller kommit ur sikte) |
| `tiles` | lista | Rutor som syns för första gången: `x`, `y`, `code`, `meaning` (från lektion 26) |
| `visible` | lista av strängar | Hela synfältet, när det ändrats (från lektion 26) |

Alla listor får saknas. Telnet ser aldrig `delta`-frames.

## Kompatibilitet

Protokollet ska kunna växa utan att gamla klienter går sönder:

1. **Klienter ska ignorera frames med okänd `type`** (i Rust: `#[serde(other)] Unknown`).
2. **Klienter ska ignorera okända fält.** Nya fält läggs bara till, aldrig om.
3. **Klienter ska klara att valfria fält saknas** (t.ex. `text`).
4. En rad som inte är giltig JSON visas som vanlig text.
5. Inkompatibla ändringar kräver en ny version: `@hello mudschool 2`.

## Planerade utökningar

| Version | Frame / fält | Fas |
|---|---|---|
| 1 | `hello.lang` – klientens språk redan i handskakningen (i dag gör kommandot `language` det, lektion 38) | – |
| 1 | `sound`, `effect` – ljud och visuella effekter | – |
