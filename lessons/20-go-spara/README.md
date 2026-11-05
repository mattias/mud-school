# Lesson 20 – Saving

Restart the server and everyone is gone: names, rooms, the sword you found.
Now the world gets a memory. Accounts with **passwords**, and players whose
room, health and inventory are saved in an **SQLite** database.

SQLite is a whole SQL database in one file – no database server to install.
We use `modernc.org/sqlite`, a translation of SQLite to pure Go, so you don't
need a C compiler.

## First: the dependencies

The course has used only the standard library so far. Now you fetch two
modules (in `workspace/server`):

```
go get modernc.org/sqlite@v1.36.0 golang.org/x/crypto@v0.31.0
```

Look at `go.mod` afterwards: the versions are written down there, and
`go.sum` holds checksums so that everyone builds exactly the same code.
(These versions work with Go 1.22; newer ones may require a newer Go.)

## Goals

- Use `database/sql`: `Open`, `Exec`, `QueryRow`/`Scan`, `Query`/`rows.Next`, `?` parameters.
- Run several statements as one unit with **transactions** (`Begin`, `Commit`, `Rollback`).
- Evolve a database schema safely with **migrations**.
- Store passwords the right way: never in plain text, but as a slow **bcrypt** hash.
- Keep slow work (disk, bcrypt) **out of the game loop**.

## Go concepts

### database/sql

```go
db, err := sql.Open("sqlite", "mud.db") // the driver registered itself via import _
_, err = db.Exec(`INSERT INTO accounts (name, password_hash) VALUES (?, ?)`, name, hash)
err = db.QueryRow(`SELECT hp FROM players WHERE name = ?`, name).Scan(&hp)
if errors.Is(err, sql.ErrNoRows) { /* nobody */ }
```

**Always** use `?` parameters – never glue values into SQL with `+` or
`Sprintf`. A player called `x'; DROP TABLE accounts; --` would otherwise be
very unpleasant (SQL injection).

### Transactions

Saving a player is three statements: update the player, delete the old
inventory, insert the new one. If the server dies in the middle, the player
must not lose everything. A transaction makes it all or nothing:

```go
tx, err := db.Begin()
// ... tx.Exec(...) ...
if err != nil { tx.Rollback(); return err }
return tx.Commit()
```

Tip: a *named* return value `(err error)` plus `defer` that rolls back when
`err != nil` saves you writing `Rollback` everywhere.

### Migrations

Your database will change over time – in phase 4 players get coordinates. A
migration is one step from one schema to the next. They are run in order, and
the database remembers how far it has come:

```go
var migrations = []string{
	`CREATE TABLE accounts (...)`, // 1
	`CREATE TABLE players (...)`,  // 2
	`CREATE TABLE inventory (...)`,// 3
}
```

SQLite has a free integer for this: `PRAGMA user_version`. Read it, run
`migrations[version:]` – each one in a transaction that also sets the new
version – and you are done. **Never change a released migration**; add a new
one at the end.

### Passwords

```go
hash, err := bcrypt.GenerateFromPassword([]byte(pw), PasswordCost)
err = bcrypt.CompareHashAndPassword(hash, []byte(pw)) // nil = right
```

bcrypt is *meant* to be slow (about 50–100 ms) so that someone who steals the
database can't try billions of passwords. The salt is built into the hash.

### Slow work outside the game loop

The game loop must answer every tick. A bcrypt check or a disk write in there
would freeze the whole world for everyone. So:

- login (bcrypt, `LoadPlayer`) runs in the player's own goroutine, *before*
  the player enters the world;
- on logout, the game loop only takes a **snapshot** (`p.snapshot()`), and the
  goroutine writes it to disk afterwards.

`*sql.DB` is safe to use from many goroutines at once.

## The contract

`store.go` (the frame is given):

- `OpenStore(path) (*Store, error)` – opens the database and runs the migrations. `Close()`.
- `CreateAccount(name, hash) error` – a name that already exists gives an error that wraps `ErrNameTaken`.
- `PasswordHash(name) ([]byte, error)` – `ErrNoAccount` if the name is free.
- `SavePlayer(SavedPlayer) error` – in one transaction; saving again replaces.
- `LoadPlayer(name) (SavedPlayer, found bool, error)`.

`account.go` (partly given): `authenticate`, `register` and `restore`.

The server:

- `Options.Store *Store` – `nil` means no saving and no password, exactly as before.
- Login with a store: after the name, a new name gets `MsgChoosePassword`
  (at least `MinPasswordLen` characters – else `MsgPasswordTooShort` and the
  question again); an existing name gets `MsgAskPassword`. A wrong password
  gives `MsgWrongPassword` and the name question again.
- After every password question the server also sends a **prompt frame**:
  `{"type":"prompt","prompt":"password"}` (or `"new_password"`). Add
  `Prompt string` (`json:"prompt,omitempty"`) to `Frame`. Telnet gets nothing
  (`encodeText` returns `""` for unknown types).
- A saved player comes back in the same room, with the same HP and the same
  items – which are then no longer lying in the room.
- `main.go`: a `-db` flag (default `mud.db`; empty = no saving).

## Task

1. `go get` as above.
2. Add `Prompt` to `Frame` (then `account.go` compiles).
3. Fill in `store.go` – `TestMigrations`, `TestAccounts` and `TestSaveAndLoadPlayer` test it on its own.
4. `authenticate`, `register` and `restore`; change `login` and the logout in `server.go`.
5. `./mudschool run server`, log in, pick something up, Ctrl+C, start again.

Add `mud.db` to your `.gitignore`.

## Think about

- Why does the test set `PasswordCost = 4`? What would happen to the tests with the default cost?
- `Store` is a concrete type. What would you gain by making the server use an
  interface instead (`type Store interface { ... }`)? What would the tests look like then?
- The server only saves on logout. What is lost if the power goes? How would
  you add an autosave every minute without stopping the game loop?
- What happens if two people register the same name at the same time?

## Challenges

- Autosave: every N ticks the game loop takes snapshots of everyone and a
  separate goroutine writes them.
- Hide the password in telnet too: send `IAC WILL ECHO` (bytes 255 251 1)
  before the question and `IAC WONT ECHO` after.
- Lock an account for a minute after three wrong passwords.
- Save items that lie on the floor too, so the world remembers where things were dropped.
