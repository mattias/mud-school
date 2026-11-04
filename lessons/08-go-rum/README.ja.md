# レッスン 08 – 部屋

MUD は **出口** でつながった **部屋** の世界です。数学的には *グラフ* で、部屋が
ノード、出口が有向辺です。このレッスンでは、プレイヤーが世界の中に居場所を持ち、
`look` で見回し、`north`/`south`/... と歩けるようになります。そして `say` は同じ
部屋の中でしか聞こえなくなります。

```
              [客室]
                | 上/下
 [神殿] — [広場] — [鍛冶屋]          ← マドビーの村
            |
         [町の門]
```

## 目標

- **ポインタ** を扱う: 同じ部屋の全プレイヤーは *同じ* `*Room` を指す。
- 構造体リテラルで構造体を作る: `&Room{ID: "square", ...}`。
- **可変長引数** の関数を書く: `NewWorld(start string, rooms ...*Room)`。
- **クロージャ** でコマンドを作る: `moveCommand("north")`。
- `strings.Builder` で効率よくテキストを組み立てる。

## Go の概念

### ポインタ

```go
r := &Room{Name: "広場"}   // r は *Room – ポインタ
p.Room = r                 // p と r は同じ部屋を指す
p.Room == other.Room       // アドレスを比較: 同じ部屋にいるか？
```

### 「覚えている」クロージャ

```go
func moveCommand(dir string) CommandFunc {
    return func(s *Server, p *Player, cmd Command) {
        // moveCommand が戻ったずっと後でも、dir はここにある
    }
}

for _, dir := range Directions {
    s.Register(dir, moveCommand(dir))   // コマンド 6 つ、関数 1 つ
}
```

## 課題

1. **`world.go`**（新しいファイル）: `NewWorld` と `World.Room` を書く。
   **`defaultworld.go`**（新しいファイル、完成済み）にはマドビーの村
   `DefaultWorld` があります。部屋を増やしてもかまいません！
2. **`player.go`**: `Player` に `Room *Room` フィールドを追加する。これは `s.mu` で
   守られます（ロックを握っているときだけ読み書きする）。
3. **`server.go`**:
   - `Server` に `world *World` フィールドを追加する、
   - `NewServerWithWorld(w *World) *Server` を作り、`NewServer()` からは
     `DefaultWorld()` で呼ぶ、
   - 新しいプレイヤーを開始部屋に置き、`MsgWelcomeName` の直後に `look` を実行する。
4. **コマンド**（新しいファイル `rooms.go` にまとめるとよいでしょう）:
   - `look` は部屋の名前、説明、出口、そこにいる他の人を送る:

     ```
     マドビー広場              ← Name
     石畳の広場。……           ← Description
     出口: 北、東              ← MsgExits
     ここにいる人: Bertil      ← MsgHere
     ```

     出口は `Directions` の順で、名前は `directionNames`（`messages.go` 内）を使い、
     `MsgListSep` で区切ります。出口がなければ `MsgNoExits`。`MsgHere` の行
     （他のプレイヤー、ソート済み）は、誰か他の人がいるときだけ送ります。
   - `north`、`east`、`south`、`west`、`up`、`down`: 出口がなければ `MsgCantGo` を送る。
     あれば、元の部屋の人には `directionTravel[dir]` を使った `MsgLeaves`
     （`Anna は北へ去った。`）、新しい部屋の人には `MsgArrives`
     （`Anna がやって来た。`）、Anna 本人には新しい部屋の `look`。
   - `say` は同じ部屋でしか聞こえなくなります。新しいコマンド `shout` はどこでも
     聞こえます: `MsgYouShout` / `MsgShouts`。

## ロックの約束事

`s.mu` を必要とする関数が増えると、うっかり二重にロックしやすくなります
（Go のミューテックスは再入可能では **ありません**。デッドロックします）。よくある
工夫は、ロックが *すでに取られている前提* の関数の名前を `Locked` で終えることです:

```go
func cmdLook(s *Server, p *Player, cmd Command) {
    s.mu.Lock()
    defer s.mu.Unlock()
    p.Send(s.describeRoomLocked(p))   // 自分ではロックしない
}
```

## 自分で試す

2 つのクライアントでログインし、歩き回り、話し、叫んでみましょう。

## 考えてみよう

- `Room.Exits` が `map[string]*Room` ではなく `map[string]string`（方向 → ID）なのは
  なぜ？（ヒント: 世界をファイルから読み込む次のレッスンを考えてみて）
- 出口は一方通行です。対応する `south` が戻ってこない `north` の出口を、どうやって
  見つけますか？

## チャレンジ

- `go <方向>` と、日本語の方向名（`北` など）に対応する。
- 各出口の行き先を表示する `exits`: `北 – 酒場「陽気な竜亭」`。
