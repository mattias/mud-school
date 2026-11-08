# レッスン 30 – マウス

フェーズ 4 の最後のレッスンで、ウィンドウにマウスが加わります:

- どこかを **クリック** するとそこへ歩きます。クライアントが自分で道を探し（レッスン 27 の
  サーバーと同じ A* を、今度は Rust で）、1 歩ずつ送ります。
- 何かを **指す** と、ツールチップがそれが何かを教えてくれます: *「大きなネズミ」*、*「木」*。
- **右クリック** すると、できることのメニューが出ます: *大きなネズミを調べる*、
  *大きなネズミを攻撃する*、*ここへ歩く*。

```
#########
#@..#...#        ネズミを右クリック:
#...#.r.#           ┌──────────────────────────┐
#.......#           │ 調べる 大きなネズミ      │  -> "examine rat"
#.....*.#           │ 攻撃する 大きなネズミ    │  -> "kill rat"
#########           │ ここへ歩く               │  -> 歩く（クライアント内）
                    └──────────────────────────┘
```

## 目標

- もう一度 A* を実装する。今度は Rust で、`BinaryHeap` と `Reverse` を使う。
- 座標系を変換する: ピクセル → 画面のマス → ゾーンのマス。
- 一連の行動を、時間をかけて送るコマンドに変える（キュー + 時計）。
- 「これで何ができる？」をデータ（`MenuItem` のメニュー）として持つ。テストしやすく、
  どの renderer でも描きやすい。

## Rust の概念

### `BinaryHeap` と `Reverse`

Rust の優先度付きキューは **最大** の要素を先に返します。A* には最小の `f` が必要なので、
キーを `std::cmp::Reverse` で包んで順序を逆にします:

```rust
let mut open = BinaryHeap::new();
open.push(Reverse((f, g, cell)));            // タプルはフィールドごとに比較される
while let Some(Reverse((_, g, cur))) = open.pop() { ... }
```

レッスン 27 の Go の `container/heap` と比べてみましょう。あちらでは `Less` を書きましたが、
こちらでは数値のタプルなら `Ord` が最初から付いています。

### 歩みのキュー

`walk_to` は経路をコマンド（`"southeast"`、`"east"` など）にして `VecDeque` に入れます。
`next_walk_step(now)` は最大でも `WALK_STEP` ごとに 1 つずつ渡します。レッスン 19 のように
時計は引数で渡すので、テストが時刻を決められます。キーを押すとキューは空になります。
プレイヤーが操作を引き継いだのです。

### 座標

```
ピクセル（mouse_position） --screen_to_cell-->  画面のマス (x, y)
画面のマス                 --Scene::zone_cell-->  ゾーンのマス（origin + (x, y − 1)）
```

シーンは左上に表示したゾーンのマス（`origin`、レッスン 25 のカメラ）を覚えているので、
クリックをゾーンのマスに戻せます。

## 契約

`mouse.rs`（新規）:

- `walkable(meaning)` – 床、草、扉、階段。`STEPS` – 8 方向とそのコマンド。
- `step_command(from, to)` – 隣のマスへの 1 歩のコマンド。隣でなければ `None`。
- `find_path(zone, from, to)` – クライアントが歩けると知っているタイル上の最短経路（8 方向）。
  `from` は含まず `to` は含む。`to` が歩けないか到達できなければ `None`、`from == to` なら
  `Some(vec![])`。
- `describe_cell(zone, x, y)` – エンティティの名前、なければ `messages::tile_name(meaning)`、
  それもなければ `None`（見たことがない）。
- `context_menu(zone, x, y) -> Vec<MenuItem {label, command}>`: 生き物 → `LOOK_AT` + 名前 /
  `examine <id>` と `ATTACK` + 名前 / `kill <id>`。アイテム → `LOOK_AT` / `examine <id>` と
  `PICK_UP` / `get <id>`。その後、マスが歩けるなら `WALK_HERE` / `WALK`（`"@walk"`、
  クライアントが処理）。`<id>` は接頭辞を除いたエンティティ ID（`n:rat` → `rat`）で、
  サーバーは必ず理解します。
- `screen_to_cell(mx, my, cell_px)`。

`messages.rs`: `LOOK_AT`、`ATTACK`、`PICK_UP`、`WALK_HERE`、`tile_name(meaning)`。

`app.rs`: `walk: VecDeque<String>`、`next_step`、`hover`、`menu: Option<Menu {cell, items}>`、
`WALK_STEP`（150 ms）、そして:
- `walk_to(cell)` – 自分からの経路を、1 歩ずつのコマンドとして。
- `next_walk_step(now)` – 次の 1 歩。最大でも `WALK_STEP` ごとに 1 つ。
- `left_click(cell, menu_row)` – メニューが開いていれば `menu_row` 行目を選び（そのコマンドを
  返すか、`WALK` なら歩き始める）、メニューを閉じる。開いていなければそのマスへ歩く。
- `right_click(cell)` – そのマスのメニューを開く（空なら開かない）。
- キーを押すと歩行が止まり、メニューが閉じる。

`scene.rs`: `Scene::origin`、`Scene::overlay`（`hover` のツールチップ、またはメニューの項目を
マスから下へ 1 行に 1 つ）、`Scene::zone_cell(x, y)`、そしてデフォルト本体（ふつうのテキスト）を
持つ `Renderer::overlay`。

ウィンドウ（`gfx/main.rs`、`gfx.rs`）: ホバー、左右のクリック、`WALK_STEP` ごとに 1 歩、
テキストの後ろに暗い箱を描く overlay。

## 課題

まず `mouse.rs`。テストはそれ単体を確かめます。次に App、シーン、ウィンドウ。
`./mudschool run gfx` でマウスを使って歩き回ってみましょう。

## 考えてみよう

- クライアントは *自分が知っている* タイルの上で経路を探します。経路が間違うのはどんな
  とき？ そのとき何が起きる？（ヒント: 未探索のマス、道をふさぐ生き物。）
- なぜサーバーが経路を送るのではなく、クライアントで計算するのでしょう？
- メニューのコマンドは `"kill rat"` のような文字列です。行動専用のプロトコルと比べた長所と
  短所は？
- 同じマウスを端末クライアントでも動かすには？（crossterm には `EnableMouseCapture` と
  `Event::Mouse` があります。）

## チャレンジ

- 端末クライアントでもマウスを。App のロジックはもうあります。
- ホバー中に予定の経路を表示する（かすかな点の線）。
- Shift + クリックで直接攻撃、ダブルクリックで拾う。
- 生き物が見えたら歩行を止める。

**フェーズ 4 の完了です。** MUD には、シードから生成したタイルの世界、視界、狩りをする
生き物、そして光とマウスのあるグラフィカルクライアントがそろいました。フェーズ 5 では、
再起動せずにサーバーを拡張できるようにします。
