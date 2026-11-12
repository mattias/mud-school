# レッスン 37 – チュートリアル

フェーズ 6 では、あなたの MUD を世界に出します。そこにいるプレイヤーのほとんどは初心者です。
MUD を遊んだことのない人がログインし、部屋の説明と点滅するカーソルを見て――去っていきます。
その対策は MUD と同じくらい古いもの、**チュートリアル** です。新しいプレイヤーは専用の
小さな区域から始め、一歩ずつ案内されます:

```
はじまりの広間
白い漆喰の壁の静かな広間。...
[チュートリアル 1/6] ようこそ！ この短いチュートリアルで基本を覚えましょう
（tutorial skip で抜けられます）。look と打って、今いる場所を見てみましょう。
> look
...
[チュートリアル 2/6] 北に戸口があります。go north で通り抜けましょう
> go north
中庭
...
[チュートリアル 3/6] 地面に何か落ちています。get たいまつ で拾いましょう。
```

最後のステップが終わると、本物の世界へ踏み出します。

## 目標

- 振る舞いを **データ** として書く: ステップはコードではなくワールドファイルにある。
- レッスン 34 のイベントの仕組みを新しい用途に使う: プレイヤーを止めるのではなく、
  プレイヤーが *したこと* を見るフック。
- 使用中のデータベースを変える: 新しい **マイグレーション**。
- ある部屋から行ける部屋を見つける: **幅優先探索**。

## Go の概念

### データとしてのステップ

```json
"tutorial": {
  "start": "tutorial-hall",
  "steps": [
    {"text": "look と打って、今いる場所を見てみましょう。", "command": "look"},
    {"text": "北へ行きましょう。", "room": "tutorial-yard"},
    {"text": "get たいまつ で拾いましょう。", "item": "tutorial-torch"}
  ],
  "done": "よくできました！ ..."
}
```

各ステップは *プレイヤーに伝えること* と *いつ終わるか* を言います。コマンドの後、ある部屋で、
または持ち物の中のある物で。サーバーはたいまつのことも中庭のことも何も知りません。
ステップを足すのも、訳すのも、まったく新しいチュートリアルを作るのも、コードではなく
ファイルの変更です。

### 見ているフック

レッスン 34 のイベントは何かが起きる *前* に来るので、フックは「否」と言えます。
チュートリアルに必要なのはその逆、今起きたことを知ることです。新しいイベント
`EventCommand` は、存在するコマンドの後に毎回来ます:

```go
s.emit(&Event{Kind: EventCommand, Player: p, Room: p.Room, Text: cmd.Verb})
```

そしてチュートリアルはただのフックです: `s.On(EventCommand, s.onTutorialCommand)`。
スクリプトもおまけで使えます: `mud.on("command", ...)`。

### マイグレーション

レッスン 20 は、データベースを捨てずに育てられると約束しました:

```go
// 4: how far they have come in the tutorial (lesson 37); 0 = not in it
`ALTER TABLE players ADD COLUMN tutorial INTEGER NOT NULL DEFAULT 0`,
```

`DEFAULT 0` が安全の鍵です。チュートリアルができる前に保存されたプレイヤーは全員 0
（「チュートリアル中ではない」）になり、学校に送り返されることはありません。

### チュートリアルの区域

チュートリアルの部屋は閉じていなければなりません。チュートリアルから歩いて出られる新人は、
うっかり飛ばしてしまったことになります。区域とは開始の部屋から行けるすべての部屋で、
出口をたどる幅優先探索（キューと、訪れた部屋の `map[string]bool`）で求めます。
同じ区域を使って、プレイヤーが抜けるときにどの物を戻すかも決めます。次の新人も
たいまつを見つけられるように。

## 契約

- ワールドファイルには `"tutorial"` オブジェクトを置ける: `start`（部屋）、`steps`
  （それぞれ `text` と、`command`・`room`・`item` のちょうど 1 つ）と `done`。
  `Tutorial`、`TutorialStep`、`World.Tutorial`、`worldFile.Tutorial`。
- `LoadWorld` は物の後でチュートリアルを確かめ、すべての誤りを一度に報告する:
  `MsgErrTutorialStart`（その後は何もしない）、`MsgErrTutorialOpen`（チュートリアルから
  世界の開始の部屋に行ける）、`MsgErrTutorialEmpty`、そしてステップごと（1 から番号付け）に
  `MsgErrTutorialStep`、`MsgErrTutorialRoom`（区域にない）、`MsgErrTutorialItem`（区域にない）。
- `(w *World) LoadTutorial(r)` はチュートリアルのファイル（`rooms`、`items`、`tutorial`）を
  世界に加える。`DefaultWorld` はそれで `tutorial.json`（用意済み）を読み込む。
- 保存のない（またはデータベースのない）プレイヤーは、チュートリアルの開始の部屋の
  ステップ 1 から始め、部屋の説明の後に `MsgTutorialStep`（番号、数、テキスト）を受け取る。
- `EventCommand` は存在するコマンド（組み込み、スクリプト、プラグイン。未知のものは除く）の
  後に毎回来て、`Text` に動詞（別名は展開済み）が入る。
- コマンドの後、終わったステップは必要なら一度に複数進み、次のステップを見せる。
  最後の後は `done` のテキスト、区域の物はそれぞれの部屋に戻り、プレイヤーは世界の
  開始の部屋へ移る（`MsgDisappears`、`MsgArrives`、look）。
- `tutorial` は今のステップを繰り返し、`tutorial skip` で抜ける（`MsgTutorialSkipped`）。
  チュートリアル外では `MsgTutorialNotIn`。
- `SavedPlayer.Tutorial`（ステップ。0 = チュートリアル中ではない）はマイグレーション 4 の
  新しい列に保存され、再ログインしたプレイヤーは続きから始める。
- `worldSnapshot`（レッスン 36）はチュートリアルも残す。

## 課題

1. `tutorial.go` は一部用意されています（型、`tutorialFile`）。`LoadTutorial`、
   `tutorialArea`、`validateTutorial`、`startTutorial`、`showTutorialStep`、`stepDone`、
   `onTutorialCommand`、`leaveTutorial`、`cmdTutorial` を書きましょう。
2. `world.go`: `World` と `worldFile` に `Tutorial`、そして `LoadWorld` での確認。
3. `events.go` と `dispatch`: `EventCommand`。`server.go`: フックを登録し、新しい
   プレイヤーをチュートリアルから始め、最初の look の後にステップを見せる。
4. `player.go`: `tutorial int`。`store.go`: マイグレーション 4、`SavedPlayer.Tutorial`、
   `SavePlayer`/`LoadPlayer` の列。`account.go`: 保存と復元。
5. `defaultworld.go`: `tutorial.json` を埋め込み、`LoadTutorial` する。
   `commands.go`: `tutorial`。`olc.go`: `worldSnapshot` に `Tutorial`。

それから新しい名前でログインして、自分でチュートリアルをやってみましょう。

## 考えてみよう

- 新しいプレイヤーが 2 人同時に来て、たいまつは 1 本。何が起きる？ 本物の MUD はどう
  解決している？（「インスタンス」で調べてみましょう。）
- なぜフックはコマンドの *後* に来るのでしょう？ チュートリアルがレッスン 34 の
  取り消せるイベントを使ったら、何がまずい？
- `DEFAULT 0` のおかげで古いプレイヤーはチュートリアルを飛ばします。逆にしたいのはどんなとき？
- あなたの MUD で、コードではなくデータにできるものは他に何がある？

## チャレンジ

- プレイヤーが特定のことを *言った* とき、またはクリーチャーを倒したとき（`EventDeath`）に
  終わるステップ。
- Lua のチュートリアル: スクリプトが `mud.on("command", ...)` でステップを足せるように。
- ヒント: 1 分たってもステップが終わらなければ、別の言葉で繰り返す。
