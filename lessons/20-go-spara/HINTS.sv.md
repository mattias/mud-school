# Ledtrådar

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

"Lägg in, eller uppdatera om den redan finns":

```sql
INSERT INTO players (name, room, hp) VALUES (?, ?, ?)
ON CONFLICT (name) DO UPDATE SET room = excluded.room, hp = excluded.hp
```

## Ett namn som redan finns

SQLite-drivrutinens felmeddelande innehåller `UNIQUE constraint failed`:

```go
if err != nil && strings.Contains(err.Error(), "UNIQUE") {
	return fmt.Errorf("%w: %s", ErrNameTaken, name)
}
```

## login

`login` kan returnera det som sparats, så att `handle` kan ge det till `restore`
inne i spelloopen:

```go
func (s *Server) login(p *Player) (saved *SavedPlayer, ok bool)
```

Efter namnet: om `s.store != nil`, anropa `s.authenticate(p)`. `retry` betyder
"fråga efter namnet igen" (`continue`). Sedan `s.store.LoadPlayer(p.Name)`.

## Utloggning

```go
defer func() {
	var snap SavedPlayer
	s.call(func() {
		snap = p.snapshot()
		s.removePlayer(p)
		// ... MsgLeft ...
	})
	if s.store != nil && snap.Name != "" {
		if err := s.store.SavePlayer(snap); err != nil {
			log.Printf("cannot save %s: %v", p.Name, err)
		}
	}
}()
```

`snap.Name` är tom om `call` inte kördes (spelloopen hade stannat).
