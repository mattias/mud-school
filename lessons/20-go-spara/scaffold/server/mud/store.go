package mud

import (
	"database/sql"
	"errors"

	_ "modernc.org/sqlite" // registers the "sqlite" driver – pure Go, no C compiler
)

// ErrNoAccount is returned when nobody has registered a name yet.
var ErrNoAccount = errors.New("no such account")

// migrations bring an empty database up to the newest schema, one step at a
// time. NEVER change a migration that has been released – add a new one at
// the end instead. The database remembers how many have run (PRAGMA user_version).
var migrations = []string{
	// 1: accounts
	`CREATE TABLE accounts (
		name          TEXT PRIMARY KEY,
		password_hash BLOB NOT NULL
	)`,
	// 2: where players are and how they feel
	`CREATE TABLE players (
		name TEXT PRIMARY KEY REFERENCES accounts(name),
		room TEXT NOT NULL,
		hp   INTEGER NOT NULL
	)`,
	// 3: what they carry, in order
	`CREATE TABLE inventory (
		player TEXT NOT NULL REFERENCES players(name),
		slot   INTEGER NOT NULL,
		item   TEXT NOT NULL,
		PRIMARY KEY (player, slot)
	)`,
}

// Store saves accounts and players in an SQLite database.
// A *sql.DB is safe to use from many goroutines at once.
type Store struct {
	db *sql.DB
}

// SavedPlayer is what is remembered about a player between visits.
type SavedPlayer struct {
	Name  string
	Room  string   // room ID
	HP    int      // 0 = full health
	Items []string // item IDs, in inventory order
}

// OpenStore opens (or creates) the database at path and runs the migrations
// that haven't run yet.
func OpenStore(path string) (*Store, error) {
	// TODO: sql.Open("sqlite", path), SetMaxOpenConns(1), then st.migrate().
	return nil, errors.New("TODO: OpenStore")
}

// Close closes the database.
func (st *Store) Close() error { return st.db.Close() }

// migrate runs every migration after the database's user_version, each in a transaction.
func (st *Store) migrate() error {
	// TODO: read PRAGMA user_version, run migrations[version:] one by one,
	// each in a transaction that also sets PRAGMA user_version = i+1.
	return errors.New("TODO: migrate")
}

// CreateAccount registers name with a password hash. If the name is already
// registered, the error wraps ErrNameTaken.
func (st *Store) CreateAccount(name string, hash []byte) error {
	// TODO: INSERT INTO accounts ... – a UNIQUE error becomes ErrNameTaken.
	return errors.New("TODO: CreateAccount")
}

// PasswordHash returns the password hash for name, or ErrNoAccount.
func (st *Store) PasswordHash(name string) ([]byte, error) {
	// TODO: SELECT password_hash ... – sql.ErrNoRows becomes ErrNoAccount.
	return nil, errors.New("TODO: PasswordHash")
}

// SavePlayer saves a player – all or nothing, in one transaction.
func (st *Store) SavePlayer(sp SavedPlayer) (err error) {
	// TODO: in ONE transaction: upsert the players row, delete the old
	// inventory rows and insert the new ones.
	return errors.New("TODO: SavePlayer")
}

// LoadPlayer loads a saved player. found is false if the player has never been saved.
func (st *Store) LoadPlayer(name string) (sp SavedPlayer, found bool, err error) {
	// TODO: SELECT room, hp ... (sql.ErrNoRows = not found), then the
	// inventory ORDER BY slot.
	return sp, false, errors.New("TODO: LoadPlayer")
}
