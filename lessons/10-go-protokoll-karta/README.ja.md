# レッスン 10 – プロトコルとマップ

これまでサーバーはテキストしか送っていませんでした。telnet クライアントは表示する
くらいしかできませんが、私たちの Rust クライアントはカラーマップ、ステータスバー、
そして（後で）タイルグラフィックを描きます。そのためには、意味を推測しなければ
ならないテキストではなく **構造化されたデータ** が必要です。

一方で telnet を失いたくはありません。解決策: クライアントが *交渉* します。

```
server: MUD School へようこそ！
server: お名前は？
client: @hello mudschool 1                      ← 「JSON がわかります」
server: {"type":"hello","text":"mudschool 1"}
client: Anna
server: {"type":"text","text":"ようこそ、Anna さん！"}
server: {"type":"text","text":"マドビー広場\n石畳の広場。……"}
server: {"type":"map","map":{"width":13,"height":13,"rows":[...],"legend":{"#":"town",...}}}
server: {"type":"status","status":{"name":"Anna","room":"マドビー広場","hp":10,...}}
```

プロトコル全体は `docs/ja/PROTOCOL.md` に書かれています。これが Go のサーバーと
Rust のクライアントの間の **契約** です。

## このレッスンで一番大切な考え

> **サーバーは「意味」を送り、クライアントが「見た目」を決める。**

マップはコード（`T`）と説明（`"T": "forest"`）として送られます。telnet はコードを
そのまま表示します。Rust クライアントはレッスン 12 で緑の `♣` を描き、フェーズ 4 では
16×16 ピクセルのきれいなトウヒの画像を描きます。**サーバーは一切変えずに**。

## 目標

- *同じ* ゲームロジックに 2 つの異なるエンコード（テキストと JSON）を持たせる。
- `json.Marshal` と `omitempty` を使う。
- 複数のゴルーチンから読まれるフラグに `sync/atomic` を使う。
- `[][]byte` で 2 次元のグリッドを描く。

## マップ

出口を間に描けるように、部屋は 1 行・1 列おきに置きます:

```
    +          酒場              （半径 1 -> 5×5 文字）
    |
    @-+        広場 – 鍛冶屋     （'@' = あなた）

"              草原（ここからは出口なし）
```

## 課題

新しいファイル: `frame.go`（型は完成済み。これが *プロトコルそのもの* です）と `mapview.go`。

1. **`frame.go`**: `encodeText` と `encodeJSON` を書く。
2. **`mapview.go`**: コメントのとおりに `RenderMap` と `WithSelf` を書く。
   表 `terrainCodes` と `mapLegend` は用意してあります。
3. **`player.go`**:
   - `outbox chan string` を `outbox chan Frame` に変え、`SendFrame(f Frame)` を追加する。
     `Send(text)` は `SendFrame(TextFrame(text))` になる、
   - `Player` に `structured atomic.Bool` フィールドを追加する。`writeLoop` は、それが
     立っていれば各 Frame を `encodeJSON` で、そうでなければ `encodeText` でエンコードする
     （空文字列は送らない）、
   - `readLine` では、`HelloPrefix`（`@hello`）で始まる行は `structured` を立て、
     `Frame{Type: "hello", Text: "mudschool 1"}` で応答し、コマンドには **渡さない**。
4. **部屋のコマンド**: すべての `look` の後（つまり移動やログインの後も）、構造化
   クライアントには `map` フレーム（半径 3、`WithSelf` 付き）と `status` フレームも
   送る。新しいコマンド `map` は全員にマップを送る。telnet では ASCII のマップになる。

## 自分で試す

```
telnet localhost 4000
map
```

Rust クライアントが受け取るものを見るには:

```
nc localhost 4000
@hello mudschool 1
Anna
```

## 考えてみよう

- `writeLoop` がフラグを一度だけでなく *Frame ごとに* 読むのはなぜ？
- 普通の `bool` ではなく `atomic.Bool` なのはなぜ？ 普通の `bool` で `--race` を試してみて。
- 将来のクライアントが `@hello mudschool 2` を送ってきたらどうなる？

## チャレンジ

- 他のプレイヤーを `&` としてマップに表示する（凡例に `"&": "player"` を追加）。
- 誰かがログイン・ログアウトしたら全員に新しいステータスを送る（`online` が変わる）。
