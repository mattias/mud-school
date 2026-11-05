# ヒント

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

## アップサート

「なければ挿入、あれば更新」:

```sql
INSERT INTO players (name, room, hp) VALUES (?, ?, ?)
ON CONFLICT (name) DO UPDATE SET room = excluded.room, hp = excluded.hp
```

## 重複した名前

SQLite ドライバーのエラーメッセージには `UNIQUE constraint failed` が含まれます:

```go
if err != nil && strings.Contains(err.Error(), "UNIQUE") {
	return fmt.Errorf("%w: %s", ErrNameTaken, name)
}
```

## login

`login` が保存内容を返せば、`handle` がそれをゲームループの中で `restore` に渡せます:

```go
func (s *Server) login(p *Player) (saved *SavedPlayer, ok bool)
```

名前の後: `s.store != nil` なら `s.authenticate(p)` を呼ぶ。`retry` は「名前をもう一度尋ねる」
（`continue`）という意味。その後 `s.store.LoadPlayer(p.Name)`。

## ログアウト

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

`call` が実行されなかった（ゲームループが止まっていた）場合、`snap.Name` は空です。
