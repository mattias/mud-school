# レッスン 23 – ゾーン

フェーズ 4 の始まりです。部屋のグラフから、*Caves of Qud* のようなタイルの世界へ。
部屋はそれぞれ **ゾーン** になります。1 マスずつ歩き回るグリッドです。端の開口部から
外へ出れば、隣の部屋のゾーンに入ります。レッスン 10 の部屋マップは *ワールドマップ* として
生き続けます。

```
####.####        # 壁     . 床        ~ 水
#.......#        @ あなた  * アイテム  c 生き物
#..~*@..#
#.......#        "n" は北へ 1 歩だけ進むようになる。
#########        "go north" はこれまでどおり隣の部屋へ跳ぶ。
```

グラフィカルクライアントにゾーンを送るのは **一度だけ**。その後サーバーは *変化した*
ものだけを `delta` フレームで送ります。31×13 マスのゾーンを生き物ごと、全プレイヤーに
毎秒 10 回送るのは無駄だからです。

## 目標

- 2D の世界をモデル化する: グリッド（`[][]byte`）、位置、隣接、端。
- 何かが動くあらゆる場所で、ゲームロジックと位置を同期させる。
- 差分を計算する: 前回から何が加わり、動き、消えたか。
- 古いクライアントを壊さずに、プロトコルに 2 つの新しいフレームを追加する。

## Go の概念

### グリッド

```go
type Pos struct{ X, Y int }          // X は東へ、Y は南へ増える
type Zone struct {
	W, H  int
	Tiles [][]byte                   // Tiles[y][x] – 行が先！
}
func (z *Zone) At(p Pos) byte        // ゾーンの外はすべて壁
```

`Pos` は int 2 つの小さな構造体で、**比較可能** です。だからマップのキーにでき
（`map[Pos]bool`）、`==` で比較できます。次のレッスンでたくさん使います。

### 遅延初期化

部屋のゾーンは、最初に求められたときに作られます:

```go
func (w *World) zoneOf(r *Room) *Zone {
	if r.zone == nil {
		r.zone = ... // ParseZone(r.ZoneRows) か、生成したもの
	}
	return r.zone
}
```

### 差分

プレイヤーはそれぞれ、自分のクライアントに伝えた内容を覚えています（`p.seen`、
エンティティ ID から `Entity` へのマップ）。コマンドとティックのたびに:

```go
now := s.entities(p.Room)
// now にあって seen にない        -> added
// 両方にあるが等しくない（!=）    -> moved
// seen にあって now にない        -> removed
```

`Entity` は文字列と int の構造体なので、`old != e` はすべてのフィールドを比較します。

## 契約

**タイル**（`zone.go` の定数）: `.` 床、`#` 壁、`"` 草、`T` 木、`~` 水、`^` 岩、
`+` 扉、`<` 上り階段、`>` 下り階段。`Walkable`: 床、草、扉、階段。`tileMeanings` が凡例です。

**ワールドファイル:**
- 部屋に `"zone": ["####.####", ...]` を書ける。手作りのゾーン（3×3 以上、同じ長さの行、
  既知のタイルのみ。そうでなければ `MsgErrZoneRows`）。出口ごとに開口部が必要
  （その側の端にある歩けるマス、上下なら階段。なければ `MsgErrZoneExit`）。
- `"zone_size": [w, h]` – 生成されるゾーンの大きさ（既定 31×13）。

**手作りのゾーンがない部屋** には簡単なゾーン: 床の周りを壁で囲み、出口のある辺の中央に
開口部、上へは `(w/2, h/2-1)` に `<`、下へは `(w/2, h/2+1)` に `>`。

**位置:** `Player.Pos`、`NPC.Pos`、`Item.Pos`。
- 新しいプレイヤー、復活したプレイヤー、復元したプレイヤー: 中央に一番近い空きマス（`freeNear`）。
- 出口を通ったとき: 入ってきた開口部の 1 歩内側（`arrival`）。上下なら反対側の階段の上。
- アイテムは中央近くに置かれる。落としたアイテムは自分の立っている場所に。

**コマンド:**
- `north`、`south`、`east`、`west`、`northeast` など（`n`、`ne` など）: 1 歩。壁、木、
  水などなら `MsgBlocked`。出口のある方向の開口部からゾーンの外へ出ると隣の部屋へ。
- `up`/`down`: 階段の上でのみ（`MsgNoStairs`）。
- `go <方向>`: これまでどおり隣の部屋へ直接（`MsgGoWhere`、`MsgCantGo`）。
- `zone`: ゾーンをもう一度。

**フレーム**（PROTOCOL.md を参照）:
- `zone` – `{"room", "width", "height", "rows", "legend", "entities", "self"}`。
  エンティティは `{"id", "kind", "name", "glyph", "x", "y"}`。ID は `p:Anna`、`n:cat`、
  `i:lamp`、種類は `player`/`creature`/`item`、glyph は `@`、NPC の `glyph`
  （既定: ID の最初の文字）、アイテムの `glyph`（既定 `*`）。
- `delta` – `{"room", "added", "moved", "removed"}`。`removed` は ID のリスト。
- グラフィカルクライアントは `look` のたびに（つまり新しい部屋ごとに）zone フレームを、
  コマンドとティックの後には何か変わっていれば delta を受け取る。
- telnet には 1 歩ごとにテキストで描いたゾーンが届く（`renderZone`、用意済み）。

## 課題

1. **フィールド:** `Room.ZoneRows`（json `zone`）と `Room.zone`、`World.ZoneSize` と
   `worldFile.ZoneSize`（json `zone_size`）、`Player`・`NPC`・`Item` の `Pos`、`NPC` と
   `Item` の `Glyph`（json `glyph,omitempty`）、`Player.seenRoom` と `Player.seen`、
   `Frame.Zone` と `Frame.Delta`、`encodeText` の `case "zone"`。これで新しいファイルが
   コンパイルできます。
2. **`zone.go`**: `ParseZone`、`opening`、`arrival`、`plainZone`、`openExits`、`zoneOf`、
   `validateZone`（`Validate` から呼ぶ）。
3. **`sync.go`**: `entities`、`zoneData`、`sendZone`、`syncAll`、`syncPlayer`。
4. **移動:** `moveCommand` を `travel(p, dir)` に改名して `p.Pos` を設定する。`step`、
   階段のコマンド、`go`、`zone` を加え、歩く方向と別名 `ne`、`nw`、`se`、`sw` を登録する。
   逃走は `travel` を使う。
5. **部屋が決まるすべての場所で位置も決める:** `addPlayer`、`restore`、`die`、`moveNPC`、
   `respawnCreatures`、`AddNPCs`、`AddItems`、`cmdDrop`、`removePlayer`（補助関数
   `placeItem` と `taken` が便利です）。
6. **同期:** `dispatch` の後、`onTick` の最後、ログイン後、ログアウト後に `s.syncAll()`。
   `look` はグラフィカルクライアントに `sendZone` を送る。

場所はたくさんあります。忘れたところはコンパイラとテストが見つけてくれます。できたら
telnet で歩き回ってみましょう: `./mudschool connect`、`n`、`e`、`zone`。

## 考えてみよう

- `Tiles[y][x]` – なぜ行が先？ `Tiles[x][y]` だと `Rows()` はどうなる？
- delta はゾーン全体と比べてどれくらい安い？ 1 歩分のバイト数を数えてみよう。
- `syncAll` はコマンドのたびに、全プレイヤーの部屋の *すべての* エンティティを比較します。
  いつ遅くなりすぎる？ そのときどうする？
- ゲームの残りの部分（戦闘、`get`、`say`）はまだ部屋単位で考えています。問題になる？
  `get` でアイテムの上に立つ必要があるとしたら何が変わる？

## チャレンジ

- 扉: `open` するまで閉じている（歩けない）`+`。
- 生き物がときどきゾーンの中をランダムに 1 歩動くようにする。
- telnet のプレイヤーにも、`look` で説明の横にゾーンを表示する。
