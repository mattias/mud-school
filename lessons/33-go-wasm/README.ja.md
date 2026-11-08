# レッスン 33 – WebAssembly のプラグイン

ここでコースの 2 つの言語が本当に出会います。**Rust** で書き、**WebAssembly** にコンパイルし、
**Go サーバーの中で** 動くプラグインです。

```
workspace/plugins/greeter/src/lib.rs   (Rust)
        │  cargo build --release --target wasm32-unknown-unknown
        ▼
server/plugins/greeter.wasm            (WebAssembly: 移植可能で隔離されたバイトコード)
        │  wazero（純粋な Go で書かれた WebAssembly ランタイム）が読み込む
        ▼
"greet"  ->  Hello, Anna! Greetings from a plugin written in Rust.
```

なぜ Lua（レッスン 32）ではなく WebAssembly？ 速く（コンパイル済み）、WebAssembly に
コンパイルできるどんな言語でも書け（Rust、C、Zig、Go など）、そして **隔離されて** います。
プラグインが触れられるのは自分のメモリと、サーバーが与えると決めた関数だけ。クラッシュしたり
終わらなかったりするプラグインがサーバーを落とすことはできません。

## 目標

- 数値と共有メモリのブロックだけで **ABI**（ホストとプラグインの契約）を設計する。
- 境界を越えて文字列を渡す: プラグインで確保し、ホストから書き、ポインター + 長さで読む。
- 信頼できないコードを制限する: メモリページ、制限時間、バージョンチェック。
- 別のターゲット（`wasm32-unknown-unknown`）向けに `crate-type = ["cdylib"]` で Rust の
  ライブラリをビルドする。

## ABI（バージョン 1）

WebAssembly の関数は数値しか受け取れず、返せません。テキストはプラグインのメモリを通します:

```
ホスト（Go）                                 プラグイン（Rust）
────────────                                 ──────────────────
ptr := alloc(len)                    ──────▶ Vec::with_capacity(len)、forget してポインターを返す
memory.Write(ptr, "Anna\tgreet\t")
on_command(ptr, len)                 ──────▶ バイトを読み、タブで分けて判断する
                                     ◀────── mud.send(ptr, len)   （インポート: ホストがテキストを読む）
                                     ◀────── 1（処理した）か 0（自分のではない）を返す
dealloc(ptr, len)                    ──────▶ Vec を解放する
```

| | 関数 | 意味 |
|---|---|---|
| プラグインがエクスポート | `abi_version() -> i32` | 1 でなければならない |
| | `alloc(len) -> ptr`、`dealloc(ptr, len)` | ホストが書き込むためのメモリ |
| | `on_command(ptr, len) -> i32` | `"player\tverb\targs"`。1 = 処理した |
| ホストが提供（モジュール `mud`） | `send(ptr, len)` | コマンドを打ったプレイヤーへのテキスト |
| | `log(ptr, len)` | サーバーのログの 1 行 |

バージョン番号があるので、サーバーは別の契約向けにビルドされたプラグインで落ちるのではなく、
それを拒否できます。

## Go の概念

### wazero

```go
rt := wazero.NewRuntimeWithConfig(ctx, wazero.NewRuntimeConfig().
	WithMemoryLimitPages(PluginMemoryPages).   // 最大 256 × 64 KiB
	WithCloseOnContextDone(true))              // タイムアウトで `loop {}` も止まる

rt.NewHostModuleBuilder("mud").
	NewFunctionBuilder().WithFunc(func(ctx context.Context, m api.Module, ptr, length uint32) {
		b, _ := m.Memory().Read(ptr, length)   // プラグインのメモリ
		// ...
	}).Export("send").
	Instantiate(ctx)

mod, err := rt.InstantiateWithConfig(ctx, code, wazero.NewModuleConfig().WithName(name))
res, err := mod.ExportedFunction("abi_version").Call(ctx)   // []uint64
```

`WithFunc` はリフレクションを使って、ふつうの Go 関数を WebAssembly のインポートに変えます。
呼び出しは `context` を受け取ります。タイムアウトがあれば、返ってこないプラグインは止められ
（モジュールは閉じられ、フォルダーが次に変わったときに再読み込みされます）。

### コマンドの順番

組み込みのコマンドが先、次にスクリプト（レッスン 32）、次にプラグイン、最後に「不明な
コマンド」。0 を返すプラグインは、次のプラグインに譲ります。

## 契約

- `Options.PluginDir` – `.wasm` ファイルのフォルダー。サーバーを作るときに名前順に読み込み、
  ファイルが変われば再読み込みする（`ReloadEvery` ごと）。
- `LoadPlugins(dir) (*Plugins, error)` – エクスポートが足りない、ABI バージョンが違う、
  WebAssembly でない、のいずれかのプラグインは飛ばす。エラーにはファイル名（とバージョン）を
  含める。他は読み込む。`Plugins.Close()`。
- 組み込みのコマンドもスクリプトも引き受けなかったコマンドはプラグインに渡す
  （`"player\tverb\targs"`）。最初に 1 を返したプラグインが処理したことになる。
- `mud.send` はコマンドを打ったプレイヤーにテキストを送る。
- プラグインの呼び出しには制限時間（`PluginTimeout`、100 ms）があり、メモリは最大
  `PluginMemoryPages`。失敗したプラグインはプレイヤーに `MsgPluginError`（管理者には
  `MsgPluginLoadFailed`）を送り、壊れたものとして印を付ける。サーバーは動き続ける。
- `main.go`: `-plugins plugins`。

## 課題

1. `go get github.com/tetratelabs/wazero@v1.8.2`。
2. `plugins.go` は一部用意されています（ホストモジュール付きの `LoadPlugins`、`loadPlugins`、
   `runPluginCommand`、`watchPlugins`）。`load`、`handle`、`call` を書きましょう。次に
   オプション、フィールド、`dispatch`、フラグ。
3. プラグインをビルドする（最初の 1 回: `rustup target add wasm32-unknown-unknown`）:

   ```
   cd workspace/plugins/greeter
   cargo build --release --target wasm32-unknown-unknown
   cp target/wasm32-unknown-unknown/release/greeter.wasm ../../server/plugins/
   ```

   `-plugins plugins` でサーバーを起動し、ログインして `greet` と打ちます。`lib.rs` の
   あいさつを変え、ビルドし、コピーして、もう一度 `greet`。再起動なしで。

## 考えてみよう

- プラグインはサーバーに何ができる？ 何ができない？ それはなぜ？
- なぜホストは自分で選んだアドレスに書かず、`alloc` を呼ぶのでしょう？
- `on_command` はタブ区切りの文字列を 1 つ受け取ります。構造化されたデータならどうする？
  （JSON？ バイナリ形式？）
- バージョン 1 向けのプラグインを壊さずに、レッスン 34 のための `on_event` を加えるには？

## チャレンジ

- 別の言語のプラグイン: TinyGo、Zig、C。ABI は同じ。
- `mud.player(name_ptr, name_len, out_ptr) -> len`: プラグインがプレイヤーについて問い合わせ
  られるようにする。
- 「燃料」: 実時間の代わりに命令数を数える（wazero にはそのためのリスナー API があります）。
- 読み込まれているものと壊れたプラグインを一覧にするコマンド `@plugins`。
