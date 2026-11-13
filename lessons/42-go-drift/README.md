# Lesson 42 – Out on the internet

Time to let the world in. In this lesson the server gets what it needs to
live on a computer you don't sit in front of – and you put it on one: a
small **VPS** (a virtual server, a few euros a month), with **Docker** to
run it and **Caddy** in front to give it HTTPS.

```
                          your VPS
                ┌────────────────────────────────────┐
https://mud.… ──┼─▶ Caddy :443 ──▶ mud :8080 (web)   │   /data/mud.db
telnet  …:4000 ─┼──────────────▶ mud :4000 (telnet)  │   (a volume)
                └────────────────────────────────────┘
```

## Goals

- Configure a program from **environment variables** – and keep `main` testable.
- **Structured logs** with `log/slog`: text for you, JSON for machines.
- A **health check** (`/healthz`) and a clean stop on **SIGTERM**.
- Know who is really connecting behind a **proxy** (`X-Forwarded-For`) – and
  when not to believe it.
- Build a small, safe container image, and run it on a VPS.

## Go concepts

### A testable `main`

`main` starts servers and never returns – hard to test. So the parts that
can be tested move out of it:

```go
func loadConfig(args []string, lookup func(string) (string, bool)) (Config, error)
```

`main` calls `loadConfig(os.Args[1:], os.LookupEnv)`; the tests call it with
a map. Each flag's default comes from its environment variable
(`MUD_ADDR`, `MUD_WEB`, `MUD_DB`, …), so a flag wins over the environment,
which wins over the default. `os.LookupEnv` – not `os.Getenv` – tells "set
but empty" from "not set": `MUD_DB=` means "save nothing".

### `log/slog`

```go
slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))
```

After `SetDefault` even the old `log.Printf` calls go through slog – so with
`-log json` every line is a JSON object a log service can search.

### SIGTERM

`docker stop` doesn't press Ctrl+C: it sends **SIGTERM**, waits ten seconds,
and then kills. `signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)` –
set up as the first thing in `main`, so a signal during the start isn't lost
either – lets the server say goodbye and save everyone first.

### `/healthz`

```go
mux.Handle("/healthz", srv.HealthHandler())
mux.Handle("/", gw)
```

The handler asks the game loop for the number of players, with a time limit:
a server whose loop is stuck is not healthy, even if HTTP still answers. The
container image has no `curl`, so the server can check itself:
`mudserver -healthcheck` asks `/healthz` and exits with 0 or 1 – that is
Docker's `HEALTHCHECK`.

### Behind a proxy

Behind Caddy every WebSocket comes *from Caddy*. Caddy tells who the real
player is in the header `X-Forwarded-For`. But anyone can send that header –
so the gateway only believes it with `TrustProxy`, which is only set when
the server is behind a proxy. Lesson 43 counts connections per address, so
this matters.

## The deployment files (given)

- `Dockerfile` – three stages: Rust builds `mud_web.wasm`, Go builds the
  server with the web client inside (`CGO_ENABLED=0`: the SQLite driver is
  pure Go), and the last image is *distroless*: no shell, not root, ~30 MB.
- `compose.yaml` – the MUD and Caddy. The database lives in a *volume*, so it
  survives new versions.
- `Caddyfile` – three lines: your domain, compression, `reverse_proxy mud:8080`.
  Caddy fetches and renews the certificate from Let's Encrypt by itself.
- `.env.example` – your domain and administrators.

## The contract

- `Config` and `loadConfig(args, lookup)`: the flags `-addr`, `-web`, `-world`,
  `-db`, `-tick`, `-reload`, `-admin`, `-scripts`, `-plugins`, `-log`
  (`text`|`json`), `-trust-proxy` and `-healthcheck`, with the defaults from
  `MUD_ADDR`, `MUD_WEB`, `MUD_WORLD`, `MUD_DB`, `MUD_TICK`, `MUD_RELOAD`,
  `MUD_ADMINS`, `MUD_SCRIPTS`, `MUD_PLUGINS`, `MUD_LOG`, `MUD_TRUST_PROXY`.
  Set-but-empty counts. Bad values give one error that names every variable
  that was wrong. `-admin` is split at commas and trimmed.
- `newLogger(format, w) *slog.Logger`; `healthCheck(addr) error` (nil only for
  200 from `http://addr/healthz`; `:8080` means `127.0.0.1:8080`).
- `(s *Server) HealthHandler()`: 200 `{"status":"ok","players":N,"uptime_seconds":S}`;
  503 `{"status":"stopping",...}` while shutting down or if the loop doesn't answer.
- `Gateway.TrustProxy`: a web player's address is the last one in
  `X-Forwarded-For`; without `TrustProxy` the header is ignored.
- `main`: SIGTERM like Ctrl+C, exit code 0; `/healthz` next to the gateway.

## Task

1. `config.go` is partly given (`Config`): write `loadConfig`, `newLogger` and
   `healthCheck`, and make `main` use them.
2. `mud/health.go`: `HealthHandler`. `mud/web.go`: `TrustProxy`.
3. Try it locally: `MUD_LOG=json ./mudschool run server`, then `curl localhost:8080/healthz`.
4. If you have Docker: `docker build -t mud .` in `workspace/`, and
   `docker run -p 4000:4000 -p 8080:8080 -v mud-data:/data mud`.

## Out on a VPS

1. Rent the smallest VPS with Linux (Ubuntu or Debian) – 1 CPU and 1 GB is plenty.
2. Point a domain name at it: an **A record** for e.g. `mud.example.com` with
   the VPS's IP address. (Many DNS providers have free subdomains.)
3. Log in with `ssh`, install Docker (`curl -fsSL https://get.docker.com | sh`
   – or your distribution's package), and open ports 80, 443 and 4000 in the
   firewall if there is one.
4. Copy your `workspace/` there (`git push`/`git clone`, or `rsync -a --exclude target`).
5. `cp .env.example .env`, fill in `MUD_DOMAIN` and `MUD_ADMINS`, and
   `docker compose up -d --build`.
6. Open `https://mud.example.com`. Then **register your admin name at once**
   (lesson 31: an administrator is only a name).

Good to know:

- Docker opens its published ports itself, past `ufw`: a firewall rule doesn't
  close them. Remove the `4000` line in `compose.yaml` if you don't want telnet.
- Telnet players that reach port 4000 through Docker's proxy (from the VPS
  itself, or over IPv6) all get the address of the Docker network, and share
  the five connections per address of lesson 43.
- The data lives in a named volume. If you mount a folder of your own on
  `/data` instead, it must be writable by the user in the image (uid 65532).

Day to day: `docker compose logs -f mud` shows the log, `docker compose up -d --build`
puts out a new version, and `docker compose cp mud:/data/mud.db mud-backup.db`
copies the database (best while nobody plays – see the challenges).

## Think about

- Why does the last image stage have no shell? What can an attacker who gets
  into the program do with one, and without?
- Telnet on port 4000 has no TLS. Passwords go over it in plain text. What
  would you do?
- What happens to players if `docker compose up -d --build` replaces the
  container while they play? (Lesson 35 had an answer – why doesn't it work here?)
- Why does `MUD_TRUST_PROXY=true` belong in `compose.yaml` and not in the `Dockerfile`?

## Challenges

- Run the server as a systemd service without Docker – what does systemd do
  that Docker did?
- Automatic backups of `mud.db` every night (SQLite's `VACUUM INTO`).
- A status page: players online and uptime, from `/healthz`.
