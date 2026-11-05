# レッスン 20 – 保存

サーバーを再起動すると、すべてが消えてしまいます。名前も、部屋も、見つけた剣も。
ここで世界に記憶を持たせます。**パスワード** 付きのアカウントを作り、プレイヤーの部屋・体力・
持ち物を **SQLite** データベースに保存します。

SQLite は 1 つのファイルに収まる完全な SQL データベースで、データベースサーバーを
インストールする必要はありません。SQLite を純粋な Go に翻訳した `modernc.org/sqlite` を
使うので、C コンパイラーも不要です。

## まず: 依存関係

これまでコースは標準ライブラリだけを使ってきました。ここで 2 つのモジュールを取得します
（`workspace/server` で）:

```
go get modernc.org/sqlite@v1.36.0 golang.org/x/crypto@v0.31.0
```

その後 `go.mod` を見てみましょう。バージョンが記録され、`go.sum` にはチェックサムが入って
いるので、誰もがまったく同じコードをビルドできます。（これらのバージョンは Go 1.22 で
動きます。より新しいものは新しい Go を必要とすることがあります。）

## 目標

- `database/sql` を使う: `Open`、`Exec`、`QueryRow`/`Scan`、`Query`/`rows.Next`、`?` パラメーター。
- **トランザクション**（`Begin`、`Commit`、`Rollback`）で複数の文を 1 つの単位として実行する。
- **マイグレーション** でデータベースのスキーマを安全に育てる。
- パスワードを正しく保存する: 平文ではなく、遅い **bcrypt** ハッシュとして。
- 遅い処理（ディスク、bcrypt）を **ゲームループの外** に置く。

## Go の概念

### database/sql

```go
db, err := sql.Open("sqlite", "mud.db") // ドライバーは import _ で自分を登録した
_, err = db.Exec(`INSERT INTO accounts (name, password_hash) VALUES (?, ?)`, name, hash)
err = db.QueryRow(`SELECT hp FROM players WHERE name = ?`, name).Scan(&hp)
if errors.Is(err, sql.ErrNoRows) { /* 該当なし */ }
```

**必ず** `?` パラメーターを使ってください。`+` や `Sprintf` で値を SQL に貼り付けては
いけません。そうしないと `x'; DROP TABLE accounts; --` という名前のプレイヤーが大惨事を
起こします（SQL インジェクション）。

### トランザクション

プレイヤーの保存は 3 つの文です。プレイヤーを更新し、古い持ち物を削除し、新しい持ち物を
挿入する。途中でサーバーが落ちても、プレイヤーがすべてを失ってはいけません。
トランザクションで「全部か、何もなしか」にします:

```go
tx, err := db.Begin()
// ... tx.Exec(...) ...
if err != nil { tx.Rollback(); return err }
return tx.Commit()
```

ヒント: *名前付き* 戻り値 `(err error)` と、`err != nil` のときにロールバックする `defer`
を組み合わせれば、あちこちに `Rollback` を書かずに済みます。

### マイグレーション

データベースは時間とともに変わります。フェーズ 4 ではプレイヤーに座標が加わります。
マイグレーションは、あるスキーマから次のスキーマへの 1 ステップです。順番に実行され、
データベースはどこまで進んだかを覚えています:

```go
var migrations = []string{
	`CREATE TABLE accounts (...)`, // 1
	`CREATE TABLE players (...)`,  // 2
	`CREATE TABLE inventory (...)`,// 3
}
```

SQLite にはこのための空き整数 `PRAGMA user_version` があります。それを読み、
`migrations[version:]` を実行する（それぞれ、新しいバージョンも設定するトランザクションの中で）。
それで完了です。**リリース済みのマイグレーションは決して変更しない** こと。末尾に新しいものを
追加します。

### パスワード

```go
hash, err := bcrypt.GenerateFromPassword([]byte(pw), PasswordCost)
err = bcrypt.CompareHashAndPassword(hash, []byte(pw)) // nil = 正しい
```

bcrypt は *わざと* 遅く（約 50〜100 ms）作られています。データベースを盗んだ人が何十億もの
パスワードを試せないようにするためです。ソルトはハッシュに組み込まれています。

### 遅い処理はゲームループの外で

ゲームループは毎ティック応答しなければなりません。その中で bcrypt の照合やディスク書き込みを
すると、全員の世界が止まってしまいます。そこで:

- ログイン（bcrypt、`LoadPlayer`）は、プレイヤーが世界に入る *前に*、そのプレイヤー自身の
  goroutine で実行する。
- ログアウト時、ゲームループは **スナップショット**（`p.snapshot()`）を取るだけで、
  ディスクへの書き込みはその後 goroutine が行う。

`*sql.DB` は多数の goroutine から同時に使っても安全です。

## 契約

`store.go`（枠は用意されています）:

- `OpenStore(path) (*Store, error)` – データベースを開いてマイグレーションを実行する。`Close()`。
- `CreateAccount(name, hash) error` – すでにある名前なら `ErrNameTaken` をラップしたエラー。
- `PasswordHash(name) ([]byte, error)` – 名前が未登録なら `ErrNoAccount`。
- `SavePlayer(SavedPlayer) error` – 1 つのトランザクションで。もう一度保存すると置き換わる。
- `LoadPlayer(name) (SavedPlayer, found bool, error)`。

`account.go`（一部用意済み）: `authenticate`、`register`、`restore`。

サーバー:

- `Options.Store *Store` – `nil` なら保存もパスワードもなし。これまでとまったく同じ。
- store があるときのログイン: 名前の後、新しい名前には `MsgChoosePassword`
  （`MinPasswordLen` 文字以上。足りなければ `MsgPasswordTooShort` と再質問）、既存の名前には
  `MsgAskPassword`。パスワードが違えば `MsgWrongPassword` を送り、名前の質問に戻る。
- パスワードを尋ねるたびに、サーバーは **prompt フレーム** も送る:
  `{"type":"prompt","prompt":"password"}`（または `"new_password"`）。`Frame` に
  `Prompt string`（`json:"prompt,omitempty"`）を追加します。telnet には何も送られない
  （`encodeText` は未知の種類に `""` を返す）。
- 保存されたプレイヤーは同じ部屋、同じ HP、同じ持ち物で戻ってくる。その持ち物はもう部屋には
  落ちていない。
- *同じサーバー* でログアウトして再ログインした場合も同じ。持ち物はログアウト時に最初の部屋へ
  戻っていて（レッスン 15）、`restore` がそこから取り出す。
- ログアウト時は、まずスナップショット、次に `SavePlayer`、そして **その後で**
  `removePlayer`。こうすれば素早く再接続しても古い保存データを読み込むことはない。
- `main.go`: `-db` フラグ（既定値 `mud.db`、空なら保存しない）。

## 課題

1. 上のとおり `go get` する。
2. `Frame` に `Prompt` を追加する（これで `account.go` がコンパイルできる）。
3. `store.go` を埋める。`TestMigrations`、`TestAccounts`、`TestSaveAndLoadPlayer` がそれ単体をテストします。
4. `authenticate`、`register`、`restore`。`server.go` の `login` とログアウト処理を変更する。
5. `./mudschool run server` でログインし、何かを拾い、Ctrl+C、もう一度起動する。

`mud.db` を `.gitignore` に追加しておきましょう。

## 考えてみよう

- テストはなぜ `PasswordCost = 4` にしているのでしょう？ 既定のコストだとテストはどうなる？
- `Store` は具体的な型です。代わりにサーバーがインターフェース（`type Store interface { ... }`）
  を使うようにすると、何が得られる？ そのときテストはどうなる？
- サーバーはログアウト時にしか保存しません。停電したら何が失われる？ ゲームループを止めずに
  1 分ごとの自動保存を加えるには？
- 2 人が同時に同じ名前を登録したら何が起きる？

## チャレンジ

- 自動保存: N ティックごとにゲームループが全員のスナップショットを取り、別の goroutine が書き込む。
- telnet でもパスワードを隠す: 質問の前に `IAC WILL ECHO`（バイト 255 251 1）、後に
  `IAC WONT ECHO` を送る。
- 3 回パスワードを間違えたらアカウントを 1 分間ロックする。
- 床に落ちている物も保存して、どこに物が置かれたかを世界が覚えているようにする。
