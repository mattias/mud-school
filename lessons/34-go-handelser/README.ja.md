# レッスン 34 – イベントとフック

スクリプト（レッスン 32）とプラグイン（レッスン 33）は *新しい* コマンドを加えられます。でも
多くのアイデアは、*既存の* 行動の意味を変えることです。銀行家以外には鍵のかかった扉、
言葉が逆さまになる呪い、今日死んだネズミの数。そのために、サーバーはこれから起きることを
**イベント** として知らせ、**フック** がそれに反応したり、「否」と言ったりできるようにします。

```
"go north"  ─▶  travel  ─▶  emit(Event{Kind: "enter", Player: Anna, To: vault})
                                │
                 ┌──────────────┼───────────────────┐
             Go のフック    Lua のフック       （プラグインのフック – チャレンジ）
                 │              │
         e.Cancel("金庫室の扉は鍵がかかっている。")  ──▶  アンナはとどまり、理由を知らされる
```

## 目標

- **イベントバス** を作る: イベントの種類ごとの購読者を、順番に呼ぶ。
- フックが行動を **取り消し**（理由も伝え）たり、**変えたり**（言う内容）できるようにする。
- テーブルを渡して Lua を呼び、戻り値を読む。
- 疎結合にする: プレイヤーを動かすコードは金庫室のことを知らない。

## Go の概念

### 購読者としての関数

```go
type Hook func(e *Event)

func (s *Server) On(kind string, h Hook) {
	s.hooks[kind] = append(s.hooks[kind], h)
}
```

フックはイベントへの *ポインター* を受け取るので、イベントを変えられます。`e.Text` を
書き換えたり、`e.Cancel(reason)` を呼んだり。イベントを出したコードは結果を見ます:

```go
e := &Event{Kind: EventEnter, Player: p, Room: p.Room, To: to, Dir: dir}
if !s.emit(e) {
	e.refuse()   // e の理由をプレイヤーに伝える
	return
}
```

### 戻り値のある Lua のフック

```lua
mud.on("enter", function(ev)
  if ev.to == "vault" then return "A voice says: not today." end
end)
```

イベントは Lua のテーブル（`kind`、`player`、`npc`、`room`、`to`、`dir`、`text`）になり、
文字列を返すと取り消しになります。`NRet: 1` の `CallByParam` は結果を Lua のスタックに
残すので、`L.Get(-1)` で読み、`L.Pop(1)` で取り除きます。

## 契約

- `Cancel(reason)` と `Cancelled()` を持つ `Event {Kind, Player, NPC, Room, To, Dir, Text}`、
  `Hook func(*Event)`、`(s *Server) On(kind, hook)`（サーバーの起動前か、ゲームループの中で
  呼ぶ）。
- 種類:
  - `EventEnter` – `travel` の前（`go`、逃走、ゾーンの外へ歩き出すときも）: `Player`、`Room`、
    `To`、`Dir`。取り消されたら、プレイヤーはとどまり、理由を受け取る。
  - `EventSay` – `say` の前: `Player`、`Room`、`Text`。フックは `Text`（実際に言われる内容）を
    変えてよい。取り消されたら何も言われず、プレイヤーは理由を受け取る。
  - `EventTick` – 毎ティックの後（戦闘と復活の後）。
  - `EventDeath` – 生き物が倒されたとき（`NPC`、`Player` = 倒した人、`Room`）と、プレイヤーが
    死んだとき（`Player`、`Room`）。
- `emit(e) bool` は Go のフック、次に Lua のフックを、加えられた順に実行し、最初に取り消した
  ところで止まる。`false` = 取り消された。
- 理由なしの取り消しでは、プレイヤーに `MsgCancelled`。
- Lua: `mud.on(kind, function(event) ... end)`。文字列を返すと、それを理由に取り消す。Lua の
  フックはスクリプトと一緒に再読み込みされる。

## 課題

`events.go` は一部用意されています（`Event`、`Cancel`、`On`、`refuse`）。`emit` と `runHook` を
書きましょう。`scripts.go` に `mud.on` と `Scripts.hooks` を加え、`run` を値を返せる `call` に
分けます。そして `travel`、`cmdSay`、`onTick`、`defeat`、`die` でイベントを出します。

スクリプトで試してみましょう:

```lua
-- scripts/curse.lua
mud.on("say", function(ev)
  if ev.player == "Bertil" then return "舌がもつれて話せない！" end
end)
```

## 考えてみよう

- フックは加えられた順に動きます。順番が大事になるのはいつ？ フックに優先度は必要？
- `EventSay` では Go のフックはテキストを変えられますが、Lua のフックは取り消すことしか
  できません。Lua でも変えられるようにするには？
- `enter` のフック自身がプレイヤーを動かしたら（つまり別の `enter` を出したら）どうなる？
- 他にどんなイベントが役立つ？（`get`、`drop`、`attack`、`login` など）

## チャレンジ

- WebAssembly プラグインのための `on_event`（ABI バージョン 2。バージョン 1 も引き続き
  読み込む）。
- 取り消せる `EventGet` で「剣は石に刺さって抜けない」。
- 統計（プレイヤーごとに倒したネズミの数）をとるフックと、それを見せる `stats` コマンド。
