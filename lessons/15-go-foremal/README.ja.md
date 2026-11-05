# レッスン 15 – アイテム

世界には *物* が必要です。鍛冶屋の錆びた剣、酒場のパン、神殿の謎の鍵。このレッスンでは
部屋にアイテムが置かれ、プレイヤーはそれを拾い、持ち歩き、置き、調べられるように
なります。そして初めての **ジェネリック** 関数を書きます。

プロトコルも少し豊かにします。テキストフレームごとに `kind`（`room`、`say`、`move`、
`error` など）を付け、レッスン 17 でクライアントがログを色分けできるようにします。
古いクライアントは新しいフィールドを無視します。`docs/ja/PROTOCOL.md` の互換性の
ルールを実践するわけです。

## 目標

- 構造体とポインタのスライスで物を表し、持ち主の間で移動させる。
- 型制約付きのジェネリック関数を書く: `func FindMatch[T Matcher](...)`。
- インターフェース（`Matcher`）を制約として使う。
- 古いクライアントを壊さずにプロトコルを拡張する。

## Go の概念

### ジェネリクス

```go
type Matcher interface {
    Matches(word string) bool
}

func FindMatch[T Matcher](list []T, word string) (T, bool) {
    for _, x := range list {
        if x.Matches(word) {
            return x, true
        }
    }
    var zero T          // T が何であれ、そのゼロ値（ポインタなら nil）
    return zero, false
}

it, ok := FindMatch(p.Inventory, "剣")   // T = *Item は推論される
```

同じ関数がレッスン 16 では生き物を探します。`Matches` メソッドを持つ型なら何でも
動きます。Go 1.18 より前なら、`[]Matcher` と型アサーションを使っていたところです。

### スライスから取り除く

```go
func remove[T comparable](list []T, x T) []T {
    for i, y := range list {
        if y == x {
            return append(list[:i:i], list[i+1:]...)
        }
    }
    return list
}
```

`list[:i:i]` は容量を制限するので、`append` は元の配列を上書きせず新しい配列を作り
ます。スライスの定番の落とし穴です。

## 契約

- `type Item struct { ID, Name, Description string; Keywords []string; Room string }`。
  JSON タグは `id`、`name`、`description`、`keywords`、`room`。
- `func (i *Item) Matches(word string) bool` – ID かキーワードに一致（大文字小文字は無視）。
- `type Matcher interface { Matches(word string) bool }` と上記の `FindMatch`。
- `Room.Items []*Item`（json `"-"`）、`World.Items map[string]*Item`、
  `Player.Inventory []*Item`。
- `(w *World) AddItems(items ...*Item) error` と `(w *World) LoadItems(r io.Reader) error`
  （形式は `{"items": [...]}`）。問題はすべて `MsgErrItemFields`、`MsgErrDuplicateItem`、
  `MsgErrItemRoom` で報告する。
- 世界ファイルにも `"items"` を書ける。`LoadWorld` がそれを追加する。
- `DefaultWorld` は新しい `items.json` も読み込む（`world.json` と同じように埋め込む）。
- `look` は出口の後（`MsgHere` の前）に `MsgItemsHere` を表示する。
- コマンド: `get`/`take`、`drop`、`inventory`/`i`、`examine`/`x`。`messages.go` の
  `MsgGetWhat`、`MsgNoSuchItem`、`MsgYouGet`、`MsgGets`、`MsgDropWhat`、`MsgNotCarrying`、
  `MsgYouDrop`、`MsgDrops`、`MsgInventory`、`MsgInventoryEmpty`、`MsgExamineWhat` を参照。
- 定数 `KindInfo`、`KindRoom`、`KindSay`、`KindMove`、`KindError`、`KindCombat` を持つ
  `Frame.Kind`（json `kind,omitempty`）と `Player.SendKind(kind, text)`。`look` → room、
  `say`/`shout` → say、出入り → move、失敗したコマンド → error、それ以外 → info。

## 課題

1. **`items.go`**（新規）: `Matches`、`FindMatch`、`AddItems`、4 つのコマンドを埋める。
   コマンドを登録し、`command.go` に省略形 `i`、`x`、`take` を追加する。
2. **`world.go`** と **`defaultworld.go`**: 新しいフィールド、世界ファイルの `"items"`、
   標準の世界の `items.json`。
3. **`frame.go`/`player.go`**: `Kind`、定数、`KindFrame`、`SendKind`。`broadcast` と
   `broadcastRoom` に `kind` 引数を加え、どこでも適切な種別を使う。
4. **`rooms.go`**: `describeRoom` に `MsgItemsHere` の行を加える。

これらはすべてレッスン 14 のゲームループの中で動くので、ロックは不要です。

## 自分で試す

```
telnet 127.0.0.1 4000
e
get 剣
i
w
drop 剣
look
```

## 考えてみよう

- アイテムの「持ち主」はその時々で誰でしょう？ 2 人がまったく同時に `get 剣` と
  打ったらどうなる？ そして、今それが問題にならないのはなぜ？
- `Matches` が `Item` ではなく `*Item` にあるのはなぜ？

## チャレンジ

- `get all` と `drop all`。
- 入れ物: アイテムの入った宝箱（`get 鍵 from 宝箱`）。`Command.Args` が役立ちます。
- `give <アイテム> <プレイヤー>`。
