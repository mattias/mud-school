# Lektion 20 – Spara

Starta om servern och alla är borta: namn, rum, svärdet du hittade. Nu får
världen ett minne. Konton med **lösenord**, och spelare vars rum, hälsa och
ryggsäck sparas i en **SQLite**-databas.

SQLite är en hel SQL-databas i en fil – ingen databasserver att installera.
Vi använder `modernc.org/sqlite`, en översättning av SQLite till ren Go, så du
behöver ingen C-kompilator.

## Först: beroendena

Hittills har kursen bara använt standardbiblioteket. Nu hämtar du två moduler
(i `workspace/server`):

```
go get modernc.org/sqlite@v1.36.0 golang.org/x/crypto@v0.31.0
```

Titta i `go.mod` efteråt: versionerna står där, och `go.sum` innehåller
kontrollsummor så att alla bygger exakt samma kod. (De här versionerna
fungerar med Go 1.22; nyare kan kräva en nyare Go.)

## Mål

- Använda `database/sql`: `Open`, `Exec`, `QueryRow`/`Scan`, `Query`/`rows.Next`, `?`-parametrar.
- Köra flera satser som en enhet med **transaktioner** (`Begin`, `Commit`, `Rollback`).
- Utveckla ett databasschema säkert med **migreringar**.
- Spara lösenord rätt: aldrig i klartext, utan som en långsam **bcrypt**-hash.
- Hålla långsamt arbete (disk, bcrypt) **utanför spelloopen**.

## Go-koncept

### database/sql

```go
db, err := sql.Open("sqlite", "mud.db") // drivrutinen registrerade sig via import _
_, err = db.Exec(`INSERT INTO accounts (name, password_hash) VALUES (?, ?)`, name, hash)
err = db.QueryRow(`SELECT hp FROM players WHERE name = ?`, name).Scan(&hp)
if errors.Is(err, sql.ErrNoRows) { /* ingen */ }
```

Använd **alltid** `?`-parametrar – klistra aldrig in värden i SQL med `+`
eller `Sprintf`. En spelare som heter `x'; DROP TABLE accounts; --` skulle
annars bli mycket obehaglig (SQL-injektion).

### Transaktioner

Att spara en spelare är tre satser: uppdatera spelaren, ta bort den gamla
ryggsäcken, lägg in den nya. Dör servern mitt i får spelaren inte förlora
allt. En transaktion gör det till allt eller inget:

```go
tx, err := db.Begin()
// ... tx.Exec(...) ...
if err != nil { tx.Rollback(); return err }
return tx.Commit()
```

Tips: ett *namngivet* returvärde `(err error)` plus en `defer` som rullar
tillbaka när `err != nil` gör att du slipper skriva `Rollback` överallt.

### Migreringar

Databasen kommer att ändras med tiden – i fas 4 får spelarna koordinater. En
migrering är ett steg från ett schema till nästa. De körs i ordning, och
databasen minns hur långt den har kommit:

```go
var migrations = []string{
	`CREATE TABLE accounts (...)`, // 1
	`CREATE TABLE players (...)`,  // 2
	`CREATE TABLE inventory (...)`,// 3
}
```

SQLite har ett ledigt heltal för just detta: `PRAGMA user_version`. Läs det,
kör `migrations[version:]` – var och en i en transaktion som också sätter den
nya versionen – och du är klar. **Ändra aldrig en släppt migrering**; lägg
till en ny i slutet.

### Lösenord

```go
hash, err := bcrypt.GenerateFromPassword([]byte(pw), PasswordCost)
err = bcrypt.CompareHashAndPassword(hash, []byte(pw)) // nil = rätt
```

bcrypt är *tänkt* att vara långsam (ungefär 50–100 ms) så att den som stjäl
databasen inte kan prova miljarder lösenord. Saltet är inbyggt i hashen.

### Långsamt arbete utanför spelloopen

Spelloopen måste svara varje tick. En bcrypt-kontroll eller en diskskrivning
där inne skulle frysa hela världen för alla. Därför:

- inloggningen (bcrypt, `LoadPlayer`) körs i spelarens egen goroutine,
  *innan* spelaren kliver in i världen;
- vid utloggning tar spelloopen bara en **ögonblicksbild** (`p.snapshot()`),
  och goroutinen skriver den till disk efteråt.

`*sql.DB` går att använda från många goroutiner samtidigt.

## Kontraktet

`store.go` (ramen är given):

- `OpenStore(path) (*Store, error)` – öppnar databasen och kör migreringarna. `Close()`.
- `CreateAccount(name, hash) error` – ett namn som redan finns ger ett fel som wrappar `ErrNameTaken`.
- `PasswordHash(name) ([]byte, error)` – `ErrNoAccount` om namnet är ledigt.
- `SavePlayer(SavedPlayer) error` – i en transaktion; att spara igen ersätter.
- `LoadPlayer(name) (SavedPlayer, found bool, error)`.

`account.go` (delvis given): `authenticate`, `register` och `restore`.

Servern:

- `Options.Store *Store` – `nil` betyder ingen sparning och inget lösenord, precis som förut.
- Inloggning med en store: efter namnet får ett nytt namn `MsgChoosePassword`
  (minst `MinPasswordLen` tecken – annars `MsgPasswordTooShort` och frågan
  igen); ett befintligt namn får `MsgAskPassword`. Fel lösenord ger
  `MsgWrongPassword` och namnfrågan igen.
- Efter varje lösenordsfråga skickar servern också en **prompt-frame**:
  `{"type":"prompt","prompt":"password"}` (eller `"new_password"`). Lägg till
  `Prompt string` (`json:"prompt,omitempty"`) i `Frame`. Telnet får ingenting
  (`encodeText` returnerar `""` för okända typer).
- En sparad spelare kommer tillbaka i samma rum, med samma HP och samma saker
  – som då inte längre ligger kvar i rummet.
- Även efter att ha loggat ut och in igen *på samma server*: sakerna gick
  tillbaka till sina startrum vid utloggningen (lektion 15), och `restore`
  tar dem därifrån.
- Vid utloggning: först ögonblicksbilden, sedan `SavePlayer`, och först
  **därefter** `removePlayer` – så att en snabb återanslutning inte kan läsa
  in den gamla sparningen.
- `main.go`: en `-db`-flagga (standard `mud.db`; tom = ingen sparning).

## Uppgift

1. `go get` enligt ovan.
2. Lägg till `Prompt` i `Frame` (då kompilerar `account.go`).
3. Fyll i `store.go` – `TestMigrations`, `TestAccounts` och `TestSaveAndLoadPlayer` testar den för sig.
4. `authenticate`, `register` och `restore`; ändra `login` och utloggningen i `server.go`.
5. `./mudschool run server`, logga in, plocka upp något, Ctrl+C, starta igen.

Lägg till `mud.db` i din `.gitignore`.

## Fundera på

- Varför sätter testet `PasswordCost = 4`? Vad skulle hända med testerna med standardkostnaden?
- `Store` är en konkret typ. Vad skulle du vinna på att låta servern använda
  ett interface i stället (`type Store interface { ... }`)? Hur skulle testerna se ut då?
- Servern sparar bara vid utloggning. Vad går förlorat om strömmen går? Hur
  skulle du lägga till en autosparning varje minut utan att stoppa spelloopen?
- Vad händer om två personer registrerar samma namn samtidigt?

## Utmaningar

- Autosparning: var N:e tick tar spelloopen ögonblicksbilder av alla och en
  separat goroutine skriver dem.
- Dölj lösenordet i telnet också: skicka `IAC WILL ECHO` (byte 255 251 1)
  före frågan och `IAC WONT ECHO` efter.
- Lås ett konto en minut efter tre felaktiga lösenord.
- Spara även saker som ligger på marken, så att världen minns var saker släpptes.
