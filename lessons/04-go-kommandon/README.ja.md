# レッスン 04 – コマンド

エコーサーバーは MUD ではありません。サーバーがプレイヤーの入力を理解する
必要があります: `look`、`north`、`get sword`、`say hello`。まずは **パーサー** と
**コマンド登録表** から始めます。この登録表は大きなものの種です。フェーズ 5 では、
スクリプトやプラグインがサーバーの稼働中にコマンドを追加できるようになります。

## 目標

- 文字列を扱う（`strings.Fields`、`strings.Cut`、`strings.ToLower`）。
- **マップ** と **スライス** を使う。
- 関数が値であることを理解する: `map[string]CommandFunc`。
- **テーブル駆動テスト** を読み書きする。Go の世界で一番人気のパターンです。

## 新しいファイル

コースツールが `workspace/server/mud/` に 3 つのファイルを作りました:

| ファイル | 内容 |
|---|---|
| `command.go` | `Command`、`ParseCommand`、省略形 – **あなたが書く** |
| `commands.go` | コマンド登録表と組み込みコマンド – **あなたが書く** |
| `player.go` | `Player`: `Send`/`Sendf` を持つ接続中のプレイヤー – 完成済み、読んでおく |

## Go の概念

### マップ

```go
ages := map[string]int{"anna": 31}
ages["bertil"] = 40
age, ok := ages["cecilia"]   // キーがなければ ok == false、age は 0
for name, age := range ages { ... }  // 順序はランダム！
```

### スライス

```go
words := strings.Fields("  get  the sword ")  // []string{"get", "the", "sword"}
first, rest := words[0], words[1:]
names := make([]string, 0, 10)                // 長さ 0、容量 10
names = append(names, "anna")
sort.Strings(names)
```

### 値としての関数

```go
type CommandFunc func(s *Server, p *Player, cmd Command)

s.commands["dance"] = func(s *Server, p *Player, cmd Command) {
    p.Send("あなたは踊った！")
}
```

### テーブル駆動テスト

`mudschool_04_test.go` を開いて `TestParseCommand` を見てください。表の各行が
1 つのテストケースです。新しいことをテストしたければ行を足すだけです。
1 つのテストだけを Go で直接実行することもできます:

```
cd workspace/server && go test ./mud -run TestParseCommand -v
```

## 課題

1. **`command.go`**: `aliases` を埋め、`ParseCommand` を書く:
   - 前後の空白を取り除く。空行なら `Command{}`、
   - 最初の単語が動詞。小文字にし、省略形を展開する、
   - `Args` は残りの単語、`Rest` は行の残り（前後の空白なし）、
   - `'` で始まる行は `say` を意味する: `'hi` → `Rest` が `"hi"` の `say`。
2. **`server.go`**: `Server` に `commands map[string]CommandFunc` フィールドを
   追加する。`NewServer` でマップを作り、`s.registerBuiltins()` を呼ぶ。
   `handle` を書き直す: `newPlayer(conn)` で `Player` を作り、`Welcome` を送り、
   プレイヤーが切断するか `p.quit` が true になるまで行を読んで
   `s.dispatch(p, line)` を呼ぶ。
3. **`commands.go`**: `dispatch`、`cmdHelp`、`cmdQuit` を書く:
   - `help` は登録済みの全動詞をソートして `MsgListSep` で区切り、`MsgCommands` で送る
     （例: `コマンド: echo、help、quit`）、
   - `quit` は `MsgGoodbye`（`さようなら！`）を送って終了する、
   - 不明な動詞には `MsgUnknownCommand`:
     `p.Sendf(MsgUnknownCommand, cmd.Verb)` → `不明なコマンド "fly" です。…`
     （書式文字列の `%q` が引用符を付けます）。

## 自分で試す

```
telnet localhost 4000
help
echo   hello  there
L
q
```

## 考えてみよう

- なぜ `help` で動詞をソートするのでしょう？（ソートなしで何度か実行してみて）
- `Server.commands` は多数のゴルーチンから同時に読まれます。安全でしょうか？
  （はい。サーバー起動後に誰もマップに *書き込まない* 限りは。この問いは
  フェーズ 5 まで覚えておいてください。稼働中のサーバーにコマンドを追加したく
  なるときです。）

## チャレンジ

- `help <コマンド>` で説明を表示する。`CommandFunc` を構造体に置き換えても
  よいでしょう: `type commandDef struct { fn CommandFunc; help string }`。
- 前方一致に対応する: 他に `hel` で始まるコマンドがなければ `hel` を `help` と見なす。
