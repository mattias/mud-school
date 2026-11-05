package mud

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const saveWorldJSON = `{
  "start": "torg",
  "rooms": [
    {"id": "torg", "name": "Square", "description": "A square.", "terrain": "town", "exits": {"north": "hall"}},
    {"id": "hall", "name": "Hall", "description": "A hall.", "terrain": "town", "y": -1, "exits": {"south": "torg"}}
  ],
  "items": [
    {"id": "lamp", "name": "a lamp", "description": "Brass.", "keywords": ["lamp"], "room": "torg"}
  ]
}`

func init() {
	// bcrypt is slow on purpose. The tests use the lowest cost to run quickly.
	PasswordCost = 4
}

// openTestStore opens a store that is closed when the test ends (after the
// servers using it have been shut down – Cleanup runs in reverse order).
func openTestStore(t *testing.T, path string) *Store {
	t.Helper()
	st, err := OpenStore(path)
	if err != nil {
		t.Fatalf("OpenStore(%q): %v", path, err)
	}
	t.Cleanup(func() { st.Close() })
	return st
}

// newSaveServer starts a server with a frozen world and the store st.
func newSaveServer(t *testing.T, st *Store) (*Server, string) {
	t.Helper()
	w, err := LoadWorld(strings.NewReader(saveWorldJSON))
	if err != nil {
		t.Fatal(err)
	}
	srv := NewServerWithOptions(Options{World: w, Tick: time.Hour, Store: st})
	addr := startServer(t, srv)
	t.Cleanup(func() { shutdown(t, srv) })
	return srv, addr
}

func shutdown(t *testing.T, srv *Server) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}
}

// loginWithPassword logs in; first is true for a name that has no account yet.
func loginWithPassword(t *testing.T, addr, name, password string, first bool) *testClient {
	t.Helper()
	c := connect(t, addr)
	c.name = name
	c.expectContains(MsgAskName)
	c.send(name)
	if first {
		c.expectContains(fmt.Sprintf(MsgChoosePassword, name))
	} else {
		c.expectContains(MsgAskPassword)
	}
	c.send(password)
	c.expectContains(fmt.Sprintf(MsgWelcomeName, name))
	return c
}

func TestMigrations(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mud.db")
	st := openTestStore(t, path)
	var version int
	if err := st.db.QueryRow(`PRAGMA user_version`).Scan(&version); err != nil {
		t.Fatal(err)
	}
	if version != len(migrations) {
		t.Errorf(L("user_version är %d, men det finns %d migreringar", "user_version is %d, but there are %d migrations", "user_version が %d ですが、マイグレーションは %d 個あります"), version, len(migrations))
	}
	st.Close()
	// Opening again must not run anything twice ("table already exists").
	st = openTestStore(t, path)
	st.Close()
}

func TestAccounts(t *testing.T) {
	st := openTestStore(t, filepath.Join(t.TempDir(), "mud.db"))

	if _, err := st.PasswordHash("Anna"); err != ErrNoAccount {
		t.Errorf(L("PasswordHash för ett okänt namn: fick %v, väntade ErrNoAccount", "PasswordHash for an unknown name: got %v, expected ErrNoAccount", "未登録の名前の PasswordHash: %v が返りました。ErrNoAccount を期待しました"), err)
	}
	if err := st.CreateAccount("Anna", []byte("hash1")); err != nil {
		t.Fatalf("CreateAccount: %v", err)
	}
	hash, err := st.PasswordHash("Anna")
	if err != nil || string(hash) != "hash1" {
		t.Errorf("PasswordHash(Anna) = %q, %v", hash, err)
	}
	if err := st.CreateAccount("Anna", []byte("hash2")); err == nil || !strings.Contains(err.Error(), MsgErrNameTaken) {
		t.Errorf(L("ett andra konto med samma namn: fick %v, väntade ett fel som wrappar ErrNameTaken", "a second account with the same name: got %v, expected an error wrapping ErrNameTaken", "同じ名前の 2 つ目のアカウント: %v が返りました。ErrNameTaken をラップしたエラーを期待しました"), err)
	}
}

func TestSaveAndLoadPlayer(t *testing.T) {
	st := openTestStore(t, filepath.Join(t.TempDir(), "mud.db"))
	if err := st.CreateAccount("Anna", []byte("x")); err != nil {
		t.Fatal(err)
	}
	if _, found, err := st.LoadPlayer("Anna"); found || err != nil {
		t.Errorf(L("LoadPlayer före första sparningen: found=%v err=%v", "LoadPlayer before the first save: found=%v err=%v", "初回保存前の LoadPlayer: found=%v err=%v"), found, err)
	}
	want := SavedPlayer{Name: "Anna", Room: "hall", HP: 7, Items: []string{"lamp", "key"}}
	for i := 0; i < 2; i++ { // saving twice must replace, not add
		if err := st.SavePlayer(want); err != nil {
			t.Fatalf("SavePlayer: %v", err)
		}
	}
	got, found, err := st.LoadPlayer("Anna")
	if err != nil || !found {
		t.Fatalf("LoadPlayer: found=%v err=%v", found, err)
	}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Errorf("LoadPlayer:\n  got:      %+v\n  expected: %+v", got, want)
	}
	want.Items = nil
	if err := st.SavePlayer(want); err != nil {
		t.Fatal(err)
	}
	if got, _, _ := st.LoadPlayer("Anna"); len(got.Items) != 0 {
		t.Errorf(L("efter att ha sparat en tom ryggsäck: %v", "after saving an empty inventory: %v", "空の持ち物を保存した後: %v"), got.Items)
	}
}

func TestPasswordIsHashed(t *testing.T) {
	st := openTestStore(t, filepath.Join(t.TempDir(), "mud.db"))
	_, addr := newSaveServer(t, st)
	c := loginWithPassword(t, addr, "Anna", "hemligt", true)
	c.send("quit")
	c.expectClosed()

	hash, err := st.PasswordHash("Anna")
	if err != nil {
		t.Fatalf("PasswordHash: %v", err)
	}
	if strings.Contains(string(hash), "hemligt") {
		t.Fatalf(L("lösenordet sparas i klartext! Använd bcrypt.GenerateFromPassword.", "the password is stored in plain text! Use bcrypt.GenerateFromPassword.", "パスワードが平文で保存されています！ bcrypt.GenerateFromPassword を使ってください。"))
	}
	if !strings.HasPrefix(string(hash), "$2") {
		t.Errorf(L("%q ser inte ut som en bcrypt-hash ($2a$…)", "%q does not look like a bcrypt hash ($2a$…)", "%q は bcrypt のハッシュ（$2a$…）に見えません"), hash)
	}
}

func TestLoginPasswords(t *testing.T) {
	st := openTestStore(t, filepath.Join(t.TempDir(), "mud.db"))
	_, addr := newSaveServer(t, st)

	// A new player: a too short password is asked again.
	c := connect(t, addr)
	c.expectContains(MsgAskName)
	c.send("Anna")
	c.expectContains(fmt.Sprintf(MsgChoosePassword, "Anna"))
	c.send("abc")
	c.expectContains(MsgPasswordTooShort)
	c.expectContains(fmt.Sprintf(MsgChoosePassword, "Anna"))
	c.send("åäöü") // four characters (but eight bytes)
	c.expectContains(fmt.Sprintf(MsgWelcomeName, "Anna"))
	c.send("quit")
	c.expectClosed()

	// Wrong password: back to the name.
	c = connect(t, addr)
	c.expectContains(MsgAskName)
	c.send("anna")
	c.expectContains(MsgAskPassword)
	c.send("fel")
	c.expectContains(MsgWrongPassword)
	c.expectContains(MsgAskName)
	c.send("Anna")
	c.expectContains(MsgAskPassword)
	c.send("åäöü")
	c.expectContains(fmt.Sprintf(MsgWelcomeName, "Anna"))
}

func TestPromptFrames(t *testing.T) {
	st := openTestStore(t, filepath.Join(t.TempDir(), "mud.db"))
	_, addr := newSaveServer(t, st)
	c := connect(t, addr)
	c.send(HelloPrefix + " test 1")
	c.expectContains(`"type":"hello"`)
	c.send("Anna")
	f := c.expectFrame("prompt")
	if f["prompt"] != PromptNewPassword {
		t.Errorf(L("prompt-frame för ett nytt konto: %v, väntade prompt=%q", "prompt frame for a new account: %v, expected prompt=%q", "新規アカウントの prompt フレーム: %v。prompt=%q を期待しました"), f, PromptNewPassword)
	}
	c.send("hemligt")
	c.send("quit")
	c.expectClosed()

	c = connect(t, addr)
	c.send(HelloPrefix + " test 1")
	c.expectContains(`"type":"hello"`)
	c.send("Anna")
	if f := c.expectFrame("prompt"); f["prompt"] != PromptPassword {
		t.Errorf(L("prompt-frame för ett befintligt konto: %v, väntade prompt=%q", "prompt frame for an existing account: %v, expected prompt=%q", "既存アカウントの prompt フレーム: %v。prompt=%q を期待しました"), f, PromptPassword)
	}

	// Telnet never sees prompt frames.
	tc := connect(t, addr)
	tc.expectContains(MsgAskName)
	tc.send("Bertil")
	tc.expectContains(fmt.Sprintf(MsgChoosePassword, "Bertil"))
	tc.expectSilence(200 * time.Millisecond)
}

func TestPlayerSurvivesRestart(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mud.db")

	st := openTestStore(t, path)
	srv, addr := newSaveServer(t, st)
	c := loginWithPassword(t, addr, "Anna", "hemligt", true)
	c.drain()
	c.send("get lamp")
	c.expectContains(fmt.Sprintf(MsgYouGet, "a lamp"))
	c.send("north")
	c.expectContains("Hall")
	c.drain()
	shutdown(t, srv)
	st.Close()

	// A brand new server – only the database is left.
	st = openTestStore(t, path)
	_, addr = newSaveServer(t, st)
	c = loginWithPassword(t, addr, "Anna", "hemligt", false)
	c.expectContains("Hall")
	c.drain()
	c.send("inventory")
	c.expectContains("a lamp")
	c.send("south")
	c.drain()
	c.send("look")
	c.expectContains("Square")
	c.conn.SetReadDeadline(time.Now().Add(300 * time.Millisecond))
	for {
		line, err := c.r.ReadString('\n')
		if strings.Contains(line, "a lamp") {
			t.Fatalf(L("lampan ligger kvar på torget fast Anna bär den: %q", "the lamp still lies on the square although Anna carries it: %q", "アンナが持っているのにランプが広場に残っています: %q"), line)
		}
		if err != nil {
			break
		}
	}
}

func TestNoStoreNoPassword(t *testing.T) {
	w, err := LoadWorld(strings.NewReader(saveWorldJSON))
	if err != nil {
		t.Fatal(err)
	}
	srv := NewServerWithOptions(Options{World: w, Tick: time.Hour})
	login(t, startServer(t, srv), "Anna") // straight in, as before
}
