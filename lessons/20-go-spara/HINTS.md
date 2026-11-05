# Hints

## migrate

```go
var version int
if err := st.db.QueryRow(`PRAGMA user_version`).Scan(&version); err != nil {
	return err
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
	// PRAGMA can't take ? parameters. i is our own int, so Sprintf is safe here.
	if _, err := tx.Exec(fmt.Sprintf(`PRAGMA user_version = %d`, i+1)); err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
}
```

## Upsert

"Insert, or update if it is already there":

```sql
INSERT INTO players (name, room, hp) VALUES (?, ?, ?)
ON CONFLICT (name) DO UPDATE SET room = excluded.room, hp = excluded.hp
```

## A duplicate name

The SQLite driver's error message contains `UNIQUE constraint failed`:

```go
if err != nil && strings.Contains(err.Error(), "UNIQUE") {
	return fmt.Errorf("%w: %s", ErrNameTaken, name)
}
```

## login

`login` can return what was saved, so that `handle` can give it to `restore`
inside the game loop:

```go
func (s *Server) login(p *Player) (saved *SavedPlayer, ok bool)
```

After the name: if `s.store != nil`, call `s.authenticate(p)`. `retry` means
"ask for the name again" (`continue`). Then `s.store.LoadPlayer(p.Name)`.

## Logout

```go
defer func() {
	var snap SavedPlayer
	s.call(func() { snap = p.snapshot() })
	if s.store != nil && snap.Name != "" {
		if err := s.store.SavePlayer(snap); err != nil {
			log.Printf("cannot save %s: %v", p.Name, err)
		}
	}
	s.call(func() {
		s.removePlayer(p)
		// ... MsgLeft ...
	})
}()
```

`snap.Name` is empty if `call` didn't run (the game loop had stopped).
