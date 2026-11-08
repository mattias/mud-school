# レッスン 32 – Lua のスクリプト

データの再読み込み（レッスン 31）で変わるのは、世界に *何があるか* です。*何が起きるか*
（新しいコマンド、違う振る舞いをする生き物）を変えるには、これまでは Go を書いて再起動する
必要がありました。ここでサーバーに **組み込みのスクリプト言語** が加わります。コマンドと
振る舞いを `.lua` ファイルに書き、保存した瞬間に再読み込みされます。

```lua
-- scripts/dance.lua
mud.command("dance", function(player, args)
  mud.send(player, player .. " dances " .. args .. "!")
end)

mud.behavior("bark", function(npc, room)
  if mud.roll(1, 20) == 1 then mud.broadcast(room, "The dog barks.") end
end)
```

[Lua](https://www.lua.org) はゲームのスクリプトの定番言語です（World of Warcraft、Roblox、
Factorio など）。純粋な Go で書かれた Lua インタープリター **gopher-lua** で組み込むので、
C コンパイラは不要です。

## 目標

- インタープリターを組み込み、Go の API（`mud.send`、`mud.player` など）を公開する。
- 信頼できないコードを **サンドボックス** に入れる: ファイルなし、プログラムなし、制限時間。
- `context.Context` で、長く走りすぎるコードを止める。
- API の設計を考える: スクリプトに何ができるべきで、何ができるべきでないか。

## Go の概念

### Lua から Go を呼ぶ

`func(L *lua.LState) int` という形の Go 関数は Lua から呼べます。引数を Lua のスタックから
読み（`L.CheckString(1)`）、結果を積み（`L.Push(...)`）、積んだ数を返します。

```go
"send": func(L *lua.LState) int {
	if p := s.players[NormalizeName(L.CheckString(1))]; p != nil {
		p.SendKind(KindInfo, L.CheckString(2))
	}
	return 0
},
```

`L.SetFuncs(テーブル, マップ)` はそれらのマップをまとめてテーブルに入れ、
`L.SetGlobal("mud", テーブル)` でそれをグローバルの `mud` にします。

### サンドボックス

`lua.NewState(lua.Options{SkipOpenLibs: true})` は *何もない* 状態から始まります。安全な
ものだけ（`base`、`table`、`string`、`math`）を開き、`base` のうちファイルを読んだりコードを
読み込んだりする関数（`dofile`、`loadfile`、`load`、`loadstring`、`require`）を取り除きます。
`os` も `io` もないので、スクリプトはファイルを消したりプログラムを起動したりできません。

### 制限時間

```go
ctx, cancel := context.WithTimeout(context.Background(), ScriptTimeout)
defer cancel()
L.SetContext(ctx)
defer L.RemoveContext()
err := L.CallByParam(lua.P{Fn: fn, NRet: 0, Protect: true}, args...)
```

context があると、gopher-lua は命令の合間に時間切れかどうかを確かめ、エラーでスクリプトを
止めます。`while true do end` でもです。`Protect: true` にすると、Lua のエラーは panic では
なく Go の `error` になります。

### Lua の状態は誰のもの？

`LState` は複数のゴルーチンから使えません。世界と同じです。だからゲームループのものです。
スクリプトの読み込みも実行もループの中だけ。するとスクリプトはロックなしで `mud.send` を呼び、
プレイヤーを読めます。

## 契約

- `Options.ScriptDir` – `.lua` ファイルのフォルダー（`""` = スクリプトなし）。サーバーを作る
  ときに名前順に読み込み、ファイルが加わったり消えたり変わったりしたら（ゲームループの中で）
  再読み込みする。確認は `ReloadEvery` ごと。
- 壊れたファイルは飛ばし、他は読み込む。ログイン中の管理者には、ファイル名を含むエラー付きの
  `MsgScriptLoadFailed`。
- サンドボックス: `os`、`io`、`dofile`、`loadfile`、`load`、`loadstring`、`require` は `nil`。
- API（グローバルなテーブル `mud`）:
  - `mud.command(名前, function(player, args))` – 新しいコマンド。
  - `mud.behavior(名前, function(npc, room))` – 振る舞い。NPC ファイルで
    `{"type": "script", "script": "名前"}` として使う（`BehaviorSpec.Script`）。
  - `mud.send(player, text)`、`mud.broadcast(room, text)`。
  - `mud.player(名前)` → `{name, room, hp, max_hp}` または `nil`。
  - `mud.players(room)` → 名前のリスト。`mud.roll(a, b)` → a から b までの乱数。
- 組み込みのコマンドが先: スクリプトはコマンドを *追加* できるが、置き換えはできない。
- 呼び出しごとに `ScriptTimeout`（50 ms）の制限時間。失敗したり長く走りすぎたりした
  スクリプトはプレイヤーに `MsgScriptError` を送り、サーバーは動き続ける。
- `main.go`: `-scripts scripts`。

## 課題

まず `go get github.com/yuin/gopher-lua@v1.1.1`。`scripts.go` は一部用意されています
（API 付きの `newScripts`、`loadScripts`、`ScriptBehavior`、`watchScripts`）。`run`、
`LoadScripts`、`runScriptCommand` を書きましょう。次にオプション、フィールド、`dispatch`、
振る舞いの種類、フラグ。`server/scripts/hello.lua` が例です。

## 考えてみよう

- なぜスクリプトはコマンドを追加できても `look` を置き換えられないのでしょう？ 置き換えが
  役立つのはどんなとき？
- 悪意のあるスクリプトに、まだできることは？（メモリ？ `string.rep("x", 1e9)`？）
- API はスクリプトに Go のポインターではなく名前（文字列）を渡します。なぜ？
- Lua はゲームループの中で動きます。40 ms かかるスクリプトはティックの速さにどう響く？
  どうする？
- スクリプトは `@なんでも` を登録できます。そして `adminOnly` が守るのは組み込みの
  コマンドだけです。スクリプトに `@` コマンドを許すべきでしょうか？

## チャレンジ

- `mud.on_tick(function() ... end)` – 毎ティック何かをするスクリプト。
- `mud.give(player, item)` と `mud.move(player, room)` – どんなチェックが必要？
- 管理者だけでなくビルダーも `@script` コマンドでスクリプトをアップロードできるようにする。
- メモリの上限: スクリプトが作るテーブルの大きさを数える。
