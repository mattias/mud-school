# Lektion 44 – Lasttestaren

Den sista lektionen. Din MUD ligger på internet, skyddad, översatt och
spelbar i en webbläsare. Men hur många spelare klarar den? Gissa inte –
**mät**. I den här lektionen skriver du en lasttestare i **asynkron Rust**:
hundratals botar som ansluter, loggar in, går omkring och pingar servern,
och en rapport om hur snabbt den svarade.

```
$ cargo run --release -p mud-loadtest -- --bots 200 --seconds 10
200 bots against 127.0.0.1:4000 for 10 s ...
bots:          200 (200 connected, 0 failed, 200 logged in)
commands:      2416 in 10.0 s = 242/s
disconnected:  0
echo answers:  79 – p50 1635.9 ms, p95 3746.5 ms, p99 6643.9 ms, max 6643.9 ms
```

De siffrorna är riktiga – och dåliga. Läs vidare.

## Mål

- Skriv asynkron Rust med **tokio**: tasks, `async`/`.await`, kanaler.
- Vänta på det som kommer först med `tokio::select!` – och jämför med Gos `select`.
- Mät **latens** ordentligt: percentiler, inte medelvärden.
- Använd det du mäter: hitta flaskhalsen med `pprof` (lektion 43).

## Rust-koncept

### Tasks

```rust
tokio::spawn(run_bot(id, plan.clone(), events.clone(), stop.clone()));
```

En tokio-*task* är som en goroutine: billig (några hundra byte), tusentals
samtidigt, schemalagda på några få trådar. En `async fn` körs inte förrän
den awaitas eller spawnas; `.await` är där en task kan pausa och låta andra köra.

### `tokio::select!`

```rust
loop {
    tokio::select! {
        line = lines.next_line() => { /* en rad från servern */ }
        _ = sleep_until(next), if logged_in => { /* dags för ett kommando */ }
        _ = stop.changed() => { /* testet är slut */ }
    }
}
```

Det är Gos `select` – med en twist: grenarna är *futures*, och de som inte
vann släpps (avbryts). `, if logged_in` stänger av en gren, som en nil-kanal i Go.

### Kanaler

Botarna rapporterar till insamlaren över en `mpsc`-kanal (många sändare, en
mottagare); insamlaren säger åt dem att sluta med en `watch`-kanal (ett värde,
många mottagare som ser när det ändras). Go har en sorts kanal för båda;
tokio har en per mönster.

### Percentiler

Ett medelvärde döljer spelarna som väntar. "p95 = 40 ms" betyder att 95 % av
svaren kom inom 40 ms – och 5 % inte gjorde det. Botarna mäter tiden från
`echo ping-7` till svaret `ping-7`.

## Kontraktet

- `client/loadtest/` är ett paket `mud-loadtest` (givet: `Cargo.toml` och
  `src/main.rs`), en medlem i workspacen och ett dev-beroende för klienten.
- `percentile(sorted, p)` – "nearest rank": värdet på plats ⌈p/100 × n⌉
  (räknat från 1, inom 1..n); 0 utan värden.
- `bot_name(id)` – "Bot" och bara bokstäver (servern tar inget annat), olika
  för varje id.
- `run_bot(id, plan, events, stop)` – ansluter (5 s tidsgräns; `Event::Failed`
  med felet), `Event::Connected`, skickar `HELLO` och sitt namn, och sedan,
  med `tokio::select!`: vid första zonen `Event::LoggedIn` och `tutorial skip`;
  deltan appliceras; en lösenordsfråga besvaras; svaret på ett echo ger
  `Event::Reply` med tiden; när den är inloggad ett kommando var `think`
  (+ lite slump) med `Event::Sent`; vid stopp `quit`. En stängd anslutning:
  `Event::Disconnected`.
- `run(plan) -> Report` – startar botarna utspridda över `plan.ramp`, samlar
  händelser tills `plan.duration` har gått, stoppar botarna och väntar på dem,
  och sorterar `latencies_ms`. `run_blocking` (given) för kod som inte är asynkron.

## Uppgift

1. Lägg till `"loadtest"` bland workspacens medlemmar och
   `mud-loadtest = { path = "loadtest" }` under `[dev-dependencies]`.
2. `loadtest/src/lib.rs` är delvis given (`Plan`, `Event`, `Report` och dess
   `Display`, `Rng`, `next_command`, `run_blocking`): skriv `percentile`,
   `bot_name`, `run_bot` och `run`.
3. Starta servern utan gränser och utan att spara:
   `./mudschool run server -db "" -max-per-ip 0 -rate 0`, och i `client/`:
   `cargo run --release -p mud-loadtest -- --bots 50`. Sedan 200. Sedan 500.

## Vad siffrorna säger

Med 50 botar kommer svaren inom en millisekund eller två. Med 200 tar de
sekunder: servern hinner inte med. Starta servern med `-debug 127.0.0.1:6060`
också, kör botarna, och i en annan terminal:

```
go tool pprof -top http://127.0.0.1:6060/debug/pprof/profile?seconds=10
```

Alla botar står i samma rum. Efter varje kommando bygger `syncAll` varje
spelares bild av rummet – och listan över spelare i rummet görs genom att
sortera *alla* spelares namn, för varje spelare. Det är n² log n jobb per
kommando: bra för tio spelare, inte för tvåhundra. Att rätta det är kursens
sista utmaning.

## Go ↔ Rust

| | Go | Rust (tokio) |
|---|---|---|
| en billig tråd | `go f()` | `tokio::spawn(f())` |
| vänta på den första av flera | `select { case …: }` | `tokio::select! { … => … }` |
| stäng av ett fall | en nil-kanal | `, if villkor` |
| en kanal | `chan T` | `mpsc` (många → en), `watch` (ett värde → många), `broadcast`, `oneshot` |
| en timer | `time.After(d)` | `sleep(d)`, `sleep_until(t)` |

## Fundera på

- Botarna svarar på en lösenordsfråga. När kommer en – och vad finns kvar i
  databasen efter ett lasttest utan `-db ""`?
- Botarnas klocka startar när de *skickar*. Vad missar det? (Tänk på vad som
  händer innan servern läser raden.)
- 500 botar är 500 tasks på några få trådar. Hur många trådar skulle 500
  botar med `std::thread` behöva – och varför spelar det roll?
- Är ett rum fullt av botar ett rättvist test? Hur skulle ett rättvist se ut?

## Utmaningar

- Få servern att klara 500 botar i ett rum: håll en lista med spelare per rum,
  synka bara rummen där något hänt – och mät igen.
- Botar som loggar in över WebSocket i stället (`tokio-tungstenite`) – genom Caddy.
- Ett histogram över svarstiderna i stället för tre percentiler.

**Det var sista lektionen.** Du har byggt en spelserver för många spelare i
Go och en klient i Rust – från en ekoserver till en tile-värld med ljus och
mus, sparande, skript och plugins, en server som kan ändras utan att någon
kopplas ner, och till sist en MUD på internet som vem som helst kan spela i
en webbläsare. `docs/*/IDEAS.md` har idéer om vart du kan gå härnäst.
