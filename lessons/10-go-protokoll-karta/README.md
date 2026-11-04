# Lektion 10 – Protokoll och karta

Hittills har servern bara skickat text. En telnet-klient kan inte göra så
mycket mer än att visa den – men vår Rust-klient ska rita en karta i färg,
en statusrad och (senare) grafik med tiles. Då behöver den **strukturerad
data**, inte text som den måste gissa betydelsen av.

Samtidigt vill vi inte tappa telnet. Lösningen: klienten *förhandlar*.

```
server: Välkommen till MUD School!
server: Vad heter du?
klient: @hello mudschool 1                      ← "jag förstår JSON"
server: {"type":"hello","text":"mudschool 1"}
klient: Anna
server: {"type":"text","text":"Välkommen, Anna!"}
server: {"type":"text","text":"Torget i Mudby\nEtt kullerstensbelagt torg..."}
server: {"type":"map","map":{"width":13,"height":13,"rows":[...],"legend":{"#":"town",...}}}
server: {"type":"status","status":{"name":"Anna","room":"Torget i Mudby","hp":10,...}}
```

Hela protokollet beskrivs i `docs/PROTOKOLL.md`. Det är **kontraktet**
mellan din Go-server och din Rust-klient.

## Den viktigaste idén i lektionen

> **Servern skickar betydelse – klienten bestämmer utseende.**

Kartan skickas som koder (`T`) med en förklaring (`"T": "forest"`). Telnet
visar koden rakt av. Rust-klienten ritar en grön `♣` i lektion 12, och i fas 4
en snygg 16×16-pixels granbild – **utan att servern ändras alls**.

## Mål

- Låta *samma* spellogik ha två olika kodningar (text och JSON).
- Använda `json.Marshal` och `omitempty`.
- Använda `sync/atomic` för en flagga som läses av flera goroutines.
- Rita ett 2D-rutnät med `[][]byte`.

## Kartan

Rummen hamnar på varannan rad/kolumn så att utgångarna får plats emellan:

```
    +          krogen            (radius 1 -> 5×5 tecken)
    |
    @-+        torget – smedjan  ('@' = du)

"              ängen (ingen utgång dit härifrån)
```

## Uppgift

Nya filer: `frame.go` (typerna är klara – de *är* protokollet) och `mapview.go`.

1. **`frame.go`**: skriv `encodeText` och `encodeJSON`.
2. **`mapview.go`**: skriv `RenderMap` och `WithSelf` enligt kommentarerna.
   Tabellerna `terrainCodes` och `mapLegend` är givna.
3. **`player.go`**:
   - byt `outbox chan string` mot `outbox chan Frame` och lägg till
     `SendFrame(f Frame)`; `Send(text)` blir `SendFrame(TextFrame(text))`,
   - ge `Player` ett fält `structured atomic.Bool`; `writeLoop` kodar varje
     Frame med `encodeJSON` om den är satt, annars `encodeText` (och hoppar
     över tomma strängar),
   - i `readLine`: en rad som börjar med `@hello` sätter `structured`,
     besvaras med `Frame{Type: "hello", Text: "mudschool 1"}` och lämnas
     **inte** vidare till kommandona.
4. **Rum-kommandona**: efter varje `look` (och alltså varje förflyttning och
   inloggning) får en strukturerad klient också en `map`-frame (radius 3, med
   `WithSelf`) och en `status`-frame. Ett nytt kommando `map` skickar kartan
   till alla – i telnet blir det ASCII-kartan.

## Testa själv

```
telnet localhost 4000
map
```

och för att se vad Rust-klienten kommer att få:

```
nc localhost 4000
@hello mudschool 1
Anna
```

## Fundera på

- Varför läser `writeLoop` flaggan för *varje* Frame istället för en gång?
- Varför `atomic.Bool` och inte en vanlig `bool`? Prova `--race` med en
  vanlig `bool`.
- Vad händer om en framtida klient skickar `@hello mudschool 2`?

## Utmaningar

- Visa andra spelare på kartan som `&` (lägg till `"&": "player"` i legenden).
- Skicka en ny status till alla när någon loggar in eller ut (`online` ändras).
