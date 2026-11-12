# Lektion 42 – Ut på nätet

Dags att släppa in världen. I den här lektionen får servern det den behöver
för att leva på en dator du inte sitter framför – och du lägger den på en:
en liten **VPS** (en virtuell server, några euro i månaden), med **Docker**
för att köra den och **Caddy** framför för att ge den HTTPS.

```
                          din VPS
                ┌────────────────────────────────────┐
https://mud.… ──┼─▶ Caddy :443 ──▶ mud :8080 (webb)  │   /data/mud.db
telnet  …:4000 ─┼──────────────▶ mud :4000 (telnet)  │   (en volym)
                └────────────────────────────────────┘
```

## Mål

- Konfigurera ett program med **miljövariabler** – och håll `main` testbar.
- **Strukturerade loggar** med `log/slog`: text för dig, JSON för maskiner.
- En **hälsokontroll** (`/healthz`) och ett snyggt stopp vid **SIGTERM**.
- Vet vem som verkligen ansluter bakom en **proxy** (`X-Forwarded-For`) – och
  när man inte ska tro på den.
- Bygg en liten, säker containeravbild, och kör den på en VPS.

## Go-koncept

### En testbar `main`

`main` startar servrar och returnerar aldrig – svår att testa. Så det som
går att testa flyttar ut ur den:

```go
func loadConfig(args []string, lookup func(string) (string, bool)) (Config, error)
```

`main` anropar `loadConfig(os.Args[1:], os.LookupEnv)`; testerna anropar den
med en map. Varje flaggas standardvärde kommer från dess miljövariabel
(`MUD_ADDR`, `MUD_WEB`, `MUD_DB`, …), så en flagga vinner över miljön, som
vinner över standardvärdet. `os.LookupEnv` – inte `os.Getenv` – skiljer på
"satt men tom" och "inte satt": `MUD_DB=` betyder "spara ingenting".

### `log/slog`

```go
slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))
```

Efter `SetDefault` går även de gamla `log.Printf`-anropen genom slog – så med
`-log json` är varje rad ett JSON-objekt som en loggtjänst kan söka i.

### SIGTERM

`docker stop` trycker inte Ctrl+C: den skickar **SIGTERM**, väntar tio
sekunder och dödar sedan. `signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)`
– som det första i `main`, så att en signal under starten inte heller går
förlorad – låter servern säga hej då och spara alla först.

### `/healthz`

```go
mux.Handle("/healthz", srv.HealthHandler())
mux.Handle("/", gw)
```

Handlern frågar spelloopen efter antalet spelare, med en tidsgräns: en server
vars loop har fastnat är inte frisk, även om HTTP fortfarande svarar.
Containeravbilden har ingen `curl`, så servern kan kontrollera sig själv:
`mudserver -healthcheck` frågar `/healthz` och avslutar med 0 eller 1 – det
är Dockers `HEALTHCHECK`.

### Bakom en proxy

Bakom Caddy kommer varje WebSocket *från Caddy*. Caddy berättar vem den
riktiga spelaren är i rubriken `X-Forwarded-For`. Men vem som helst kan
skicka den rubriken – så gatewayen tror bara på den med `TrustProxy`, som
bara sätts när servern står bakom en proxy. Lektion 43 räknar anslutningar
per adress, så det här spelar roll.

## Driftsfilerna (givna)

- `Dockerfile` – tre steg: Rust bygger `mud_web.wasm`, Go bygger servern med
  webbklienten inuti (`CGO_ENABLED=0`: SQLite-drivrutinen är ren Go), och den
  sista avbilden är *distroless*: inget skal, inte root, ~30 MB.
- `compose.yaml` – MUD:en och Caddy. Databasen ligger i en *volym*, så den
  överlever nya versioner.
- `Caddyfile` – tre rader: din domän, komprimering, `reverse_proxy mud:8080`.
  Caddy hämtar och förnyar certifikatet från Let's Encrypt själv.
- `.env.example` – din domän och dina administratörer.

## Kontraktet

- `Config` och `loadConfig(args, lookup)`: flaggorna `-addr`, `-web`, `-world`,
  `-db`, `-tick`, `-reload`, `-admin`, `-scripts`, `-plugins`, `-log`
  (`text`|`json`), `-trust-proxy` och `-healthcheck`, med standardvärden från
  `MUD_ADDR`, `MUD_WEB`, `MUD_WORLD`, `MUD_DB`, `MUD_TICK`, `MUD_RELOAD`,
  `MUD_ADMINS`, `MUD_SCRIPTS`, `MUD_PLUGINS`, `MUD_LOG`, `MUD_TRUST_PROXY`.
  Satt men tom räknas. Felaktiga värden ger ett fel som nämner varje variabel
  som var fel. `-admin` delas vid komman och trimmas.
- `newLogger(format, w) *slog.Logger`; `healthCheck(addr) error` (nil bara för
  200 från `http://addr/healthz`; `:8080` betyder `127.0.0.1:8080`).
- `(s *Server) HealthHandler()`: 200 `{"status":"ok","players":N,"uptime_seconds":S}`;
  503 `{"status":"stopping",...}` under avstängning eller om loopen inte svarar.
- `Gateway.TrustProxy`: en webbspelares adress är den sista i
  `X-Forwarded-For`; utan `TrustProxy` ignoreras rubriken.
- `main`: SIGTERM som Ctrl+C, slutkod 0; `/healthz` bredvid gatewayen.

## Uppgift

1. `config.go` är delvis given (`Config`): skriv `loadConfig`, `newLogger` och
   `healthCheck`, och låt `main` använda dem.
2. `mud/health.go`: `HealthHandler`. `mud/web.go`: `TrustProxy`.
3. Prova lokalt: `MUD_LOG=json ./mudschool run server`, sedan `curl localhost:8080/healthz`.
4. Har du Docker: `docker build -t mud .` i `workspace/`, och
   `docker run -p 4000:4000 -p 8080:8080 -v mud-data:/data mud`.

## Ut på en VPS

1. Hyr den minsta VPS:en med Linux (Ubuntu eller Debian) – 1 CPU och 1 GB räcker gott.
2. Peka ett domännamn mot den: en **A-post** för t.ex. `mud.example.com` med
   VPS:ens IP-adress. (Många DNS-leverantörer har gratis underdomäner.)
3. Logga in med `ssh`, installera Docker (`curl -fsSL https://get.docker.com | sh`
   – eller din distributions paket), och öppna portarna 80, 443 och 4000 i
   brandväggen om det finns en.
4. Kopiera dit din `workspace/` (`git push`/`git clone`, eller `rsync -a --exclude target`).
5. `cp .env.example .env`, fyll i `MUD_DOMAIN` och `MUD_ADMINS`, och
   `docker compose up -d --build`.
6. Öppna `https://mud.example.com`. **Registrera sedan ditt administratörsnamn
   direkt** (lektion 31: en administratör är bara ett namn).

Bra att veta:

- Docker öppnar sina publicerade portar själv, förbi `ufw`: en brandväggsregel
  stänger dem inte. Ta bort raden med `4000` i `compose.yaml` om du inte vill ha telnet.
- Telnetspelare som når port 4000 genom Dockers proxy (från VPS:en själv, eller
  över IPv6) får alla Docker-nätverkets adress, och delar på de fem
  anslutningarna per adress från lektion 43.
- Datan ligger i en namngiven volym. Monterar du en egen mapp på `/data` i
  stället måste användaren i avbilden (uid 65532) kunna skriva i den.

I vardagen: `docker compose logs -f mud` visar loggen, `docker compose up -d --build`
lägger ut en ny version, och `docker compose cp mud:/data/mud.db mud-backup.db`
kopierar databasen (helst när ingen spelar – se utmaningarna).

## Fundera på

- Varför har sista steget i avbilden inget skal? Vad kan en angripare som tar
  sig in i programmet göra med ett skal, och utan?
- Telnet på port 4000 har ingen TLS. Lösenord skickas över den i klartext.
  Vad skulle du göra?
- Vad händer med spelarna om `docker compose up -d --build` byter containern
  medan de spelar? (Lektion 35 hade ett svar – varför fungerar det inte här?)
- Varför hör `MUD_TRUST_PROXY=true` hemma i `compose.yaml` och inte i `Dockerfile`?

## Utmaningar

- Kör servern som en systemd-tjänst utan Docker – vad gör systemd som Docker gjorde?
- Automatiska säkerhetskopior av `mud.db` varje natt (SQLites `VACUUM INTO`).
- En statussida: spelare online och drifttid, från `/healthz`.
