# レッスン 16 – 生き物

マドビーに命が宿ります。広場を黒猫がうろつき、酒場では女将のグレタがおしゃべりし、
森のはずれをカラスが跳ねる。そして暗い森のどこかに狼がいます。生き物（*NPC*、
ノンプレイヤーキャラクター）は自分で行動するので、ゲームループの `onTick` の中で
生きています。

生き物が何をするかは **振る舞い** で表します。インターフェースを実装する小さな値です。
どの生き物がどの振る舞いを持つかは、世界ファイルに書きます。

```json
{"id": "cat", "name": "黒猫", "room": "torget",
 "behaviors": [{"type": "wander", "chance": 0.03},
               {"type": "talk", "chance": 0.02, "lines": ["ニャー。"]}]}
```

## 目標

- 小さなインターフェース（`Behavior`）と複数の実装で設計する。
- データから値を作る: `BehaviorSpec` を `Behavior` にする「ファクトリ」。
- 乱数をテスト可能にする: シードを固定した `*rand.Rand` を注入する。
- レッスン 15 のジェネリックな `FindMatch` を新しい型に再利用する。

## Go の概念

### 複数の実装を持つインターフェース

```go
type Behavior interface {
    Act(s *Server, n *NPC)
}

type Wanderer struct{ Chance float64 }
func (b Wanderer) Act(s *Server, n *NPC) { ... }

type Talker struct{ Chance float64; Lines []string }
func (b Talker) Act(s *Server, n *NPC) { ... }
```

テストは独自の `Behavior` を追加できます。`mudschool_16_test.go` の
`countingBehavior` を見てください。ループは振る舞いが何をするかを知らないし、気にも
しません。それがインターフェースの力です。

### テストできる乱数

```go
rng := rand.New(rand.NewPCG(1, 2))  // math/rand/v2、固定シード
rng.Float64()                        // 実行するたびに同じ並び
rng.IntN(3)                          // 0、1、2 のどれか
```

サーバーは `Options.Rand` を受け取ります。本番ではランダムなシード、テストでは固定
シードを使います。同じシードで同じ結果を得るには、すべてを決まった順序で行う必要が
あります。だから `npcsByID` はソートし、`Wanderer` はマップを回すのではなく
`Directions` の順に調べるのです。

## 契約

- JSON フィールド `id`、`name`、`description`、`keywords`、`room`、`behaviors`、そして
  （レッスン 18 用で今は未使用の）`hp`、`damage`、`attackable` を持つ `type NPC struct`。
  実行時は `Location *Room`。
- `func (n *NPC) Matches(word string) bool` と `func (n *NPC) AddBehavior(b Behavior)`。
- `type Behavior interface { Act(s *Server, n *NPC) }`、`Wanderer`、`Talker`、
  種類 `"wander"` と `"talk"` を持つ `BehaviorSpec{Type, Chance, Lines}`。
- `Room.NPCs []*NPC`、`World.NPCs map[string]*NPC`、`(w *World) AddNPCs(...)` と
  `(w *World) LoadNPCs(r)`（`{"npcs": [...]}`）。エラーは `MsgErrNPCFields`、
  `MsgErrDuplicateNPC`、`MsgErrNPCRoom`、`MsgErrBehavior`。世界ファイルに `"npcs"` を
  書ける。`DefaultWorld` は新しい `npcs.json` を読み込む。
- `Options.Rand *rand.Rand`（`math/rand/v2`。nil ならランダムなシード）。
- 毎ティック、すべての生き物の振る舞いを（ID 順に）実行する。
- `look` は `MsgHere` でプレイヤーの後に生き物を並べる。`examine` は生き物にも使える。
  生き物の出入りは kind `move` の `MsgLeaves`/`MsgArrives`、話すときは kind `say` の
  `MsgSays`。

## 課題

1. **`npc.go`**（新規）: `Matches`、`Wanderer.Act`、`Talker.Act`、`moveNPC`、`AddNPCs` を埋める。
2. **`world.go`**/**`defaultworld.go`**: 新しいフィールド、世界ファイルの `"npcs"`、
   標準の世界の `npcs.json`。
3. **`server.go`**/**`loop.go`**: `Options.Rand`、`s.rng`、`onTick` ですべての振る舞いを実行。
4. **`rooms.go`**/**`items.go`**: `look` と `examine` に生き物を加える。

## 自分で試す

```
./mudschool run server -tick 20ms
```

ティックを速くすると生き物が落ち着かなくなります。広場で猫を待ったり、酒場を
訪ねたりしてみましょう。

## 考えてみよう

- `Wanderer.Act` がサーバーと生き物を `Wanderer` の中に持たず、引数として受け取るのはなぜ？
- `onTick` がマップ `s.world.NPCs` を直接回したら、固定シードのテストはどうなる？
- 生き物はどこへでも歩いていいのでしょうか？ 女将を酒場にとどめておくには？

## チャレンジ

- `follow` の振る舞い: カラスがいちばん近いプレイヤーについていく。
- 昼と夜: 狼は夜だけうろつく（`GameTime` を使う）。
- 生き物への `say`: 誰かが「エール」と言ったら女将が答える。
