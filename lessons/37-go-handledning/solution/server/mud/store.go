package mud

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

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
	// 4: how far they have come in the tutorial (lesson 37); 0 = not in it
	`ALTER TABLE players ADD COLUMN tutorial INTEGER NOT NULL DEFAULT 0`,
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
	// Tutorial is the tutorial step the player is on, from 1; 0 = not in it (lesson 37).
	Tutorial int
}

// OpenStore opens (or creates) the database at path and runs the migrations
// that haven't run yet.
func OpenStore(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	// SQLite allows one writer at a time; one connection avoids "database is locked".
	db.SetMaxOpenConns(1)
	st := &Store{db: db}
	if err := st.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return st, nil
}

// Close closes the database.
func (st *Store) Close() error { return st.db.Close() }

// migrate runs every migration after the database's user_version, each in a transaction.
func (st *Store) migrate() error {
	var version int
	if err := st.db.QueryRow(`PRAGMA user_version`).Scan(&version); err != nil {
		return err
	}
	if version > len(migrations) {
		return fmt.Errorf("the database has schema version %d, but this server only knows %d", version, len(migrations))
	}
	for i := version; i < len(migrations); i++ {
		tx, err := st.db.Begin()
		if err != nil {
			return err
		}
		if _, err := tx.Exec(migrations[i]); err != nil {
			tx.Rollback()
			return fmt.Errorf("migration %d: %w", i+1, err)
		}
		// PRAGMA can't take ? parameters; i is our own int, so Sprintf is safe here.
		if _, err := tx.Exec(fmt.Sprintf(`PRAGMA user_version = %d`, i+1)); err != nil {
			tx.Rollback()
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}

// CreateAccount registers name with a password hash. If the name is already
// registered, the error wraps ErrNameTaken.
func (st *Store) CreateAccount(name string, hash []byte) error {
	_, err := st.db.Exec(`INSERT INTO accounts (name, password_hash) VALUES (?, ?)`, name, hash)
	if err != nil && strings.Contains(err.Error(), "UNIQUE") {
		return fmt.Errorf("%w: %s", ErrNameTaken, name)
	}
	return err
}

// PasswordHash returns the password hash for name, or ErrNoAccount.
func (st *Store) PasswordHash(name string) ([]byte, error) {
	var hash []byte
	err := st.db.QueryRow(`SELECT password_hash FROM accounts WHERE name = ?`, name).Scan(&hash)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNoAccount
	}
	return hash, err
}

// SavePlayer saves a player – all or nothing, in one transaction.
func (st *Store) SavePlayer(sp SavedPlayer) (err error) {
	tx, err := st.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()
	if _, err = tx.Exec(`INSERT INTO players (name, room, hp, tutorial) VALUES (?, ?, ?, ?)
		ON CONFLICT (name) DO UPDATE SET room = excluded.room, hp = excluded.hp, tutorial = excluded.tutorial`,
		sp.Name, sp.Room, sp.HP, sp.Tutorial); err != nil {
		return err
	}
	if _, err = tx.Exec(`DELETE FROM inventory WHERE player = ?`, sp.Name); err != nil {
		return err
	}
	for slot, id := range sp.Items {
		if _, err = tx.Exec(`INSERT INTO inventory (player, slot, item) VALUES (?, ?, ?)`, sp.Name, slot, id); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// LoadPlayer loads a saved player. found is false if the player has never been saved.
func (st *Store) LoadPlayer(name string) (sp SavedPlayer, found bool, err error) {
	sp.Name = name
	err = st.db.QueryRow(`SELECT room, hp, tutorial FROM players WHERE name = ?`, name).Scan(&sp.Room, &sp.HP, &sp.Tutorial)
	if errors.Is(err, sql.ErrNoRows) {
		return sp, false, nil
	}
	if err != nil {
		return sp, false, err
	}
	rows, err := st.db.Query(`SELECT item FROM inventory WHERE player = ? ORDER BY slot`, name)
	if err != nil {
		return sp, false, err
	}
	defer rows.Close()
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return sp, false, err
		}
		sp.Items = append(sp.Items, id)
	}
	return sp, true, rows.Err()
}
