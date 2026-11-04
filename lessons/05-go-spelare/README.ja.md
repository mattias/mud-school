# レッスン 05 – プレイヤー

*誰が* 遊んでいるのかを知るときが来ました。プレイヤーは接続時に名前を選び、
サーバーはログイン中の全員を管理します。複数のゴルーチンが同じデータを共有する
のはこれが初めてです。つまり、本当にやっかいなバグ **データ競合（data race）** を
初めて書けるようになったということです。

## 目標

- `errors.New`、`errors.Is`、`%w` でエラーを作り、比較する。
- Go の文字列における *バイト* と *ルーン*（Unicode 文字）の違いを理解する。
- 共有データを `sync.Mutex` で守る。
- レース検出器がバグを見つけるところを見る。

## ログインの流れ

```
MUD School へようこそ！
お名前は？
> x
その名前は使えません: 名前は 2 文字以上にしてください。
お名前は？
> anna
ようこそ、Anna さん！
> who
オンラインのプレイヤー (2): Anna、Bertil
```

## Go の概念

### エラーは値 – 自作のものも

```go
var ErrNameTooShort = errors.New(MsgErrNameTooShort) // テキストは messages.go に

func ValidateName(name string) error {
    if len(name) < 2 { return ErrNameTooShort }  // （ただし len はバイト数。下記参照！）
    return nil
}

err := fmt.Errorf("%w: %s", ErrNameTaken, name) // エラーを情報付きで「ラップ」する
errors.Is(err, ErrNameTaken)                    // true – Is はラップの中まで見る
```

### バイトとルーン

Go の文字列は UTF-8 のバイト列です。`len("Ö")` は **2**、`len("空")` は **3** です。
文字数を数えるには `utf8.RuneCountInString` を使い、`for _, r := range s` は
バイトではなくルーン（`rune` = Unicode の 1 文字）ごとにループします。

```go
for i, r := range "さくら" {
    fmt.Println(i, string(r), unicode.IsLetter(r)) // 0 さ true, 3 く true, 6 ら true
}
```

`unicode.IsLetter` はひらがなや漢字にも true を返すので、`さくら` は有効な名前です。

### ミューテックス

`sync.Mutex` は `Lock` と `Unlock` の間のコードを一度に 1 つのゴルーチンだけが
実行するようにします:

```go
type Server struct {
    mu      sync.Mutex          // players を守る
    players map[string]*Player
}

func (s *Server) addPlayer(p *Player) error {
    s.mu.Lock()
    defer s.mu.Unlock()        // 途中で return しても必ずロックを外す
    ...
}
```

Go のマップは複数のゴルーチンから同時に使うと **安全ではありません**。ロックなしで
使うと、`fatal error: concurrent map writes` でプログラムが落ちることがあります。

## 課題

1. **`names.go`**（新しいファイル）: `ValidateName` と `NormalizeName` を書く。
2. **`server.go`**:
   - `Server` に `mu sync.Mutex` と `players map[string]*Player` を追加する、
   - `addPlayer(p) error`（名前があれば `ErrNameTaken`）、`removePlayer(p)`、
     `PlayerNames() []string`（ソート済み）を書く、
   - `handle` では `Welcome` の後、有効で空いている名前が入力されるまで
     `MsgAskName` を尋ねる。エラーなら `MsgInvalidName` でエラーを送り
     （`p.Sendf(MsgInvalidName, err)`）、もう一度尋ねる。その後 `MsgWelcomeName` を送る、
   - 切断時にプレイヤーを削除する（ヒント: `defer`）。
3. **`commands.go`**: `who` を追加する。`MsgWho` で人数と、ソートして
   `MsgListSep` で区切った名前を送る: `オンラインのプレイヤー (2): Anna、Bertil`。

テキスト（`MsgAskName`、`MsgErrNameTooShort` など）はすべて `messages.go` に
あります。`names.go` のエラー変数はすでにそれらから作られています。

## 自分で試す: データ競合を見つける

`addPlayer` の `s.mu.Lock()` と `defer s.mu.Unlock()` を一時的に消して、次を実行します

```
./mudschool check --race
```

レポートを読んでください。どの 2 つのゴルーチンがどこで衝突したかが正確に
わかります。そのあとロックを元に戻しましょう。

## 考えてみよう

- 名前が使われているか確認する *前に* 正規化するのはなぜ？
- ログインの途中でプレイヤーが切断したらどうなる？

## チャレンジ

- パスワードを（`golang.org/x/crypto/bcrypt` でハッシュ化して）保存し、他人に
  名前を取られないようにする。フェーズ 3 でプレイヤーを保存するときに再登場します。
- いくつかの名前（`admin`、`god` など）を新しいエラー `ErrNameReserved` で禁止する。
