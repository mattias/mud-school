# レッスン 35 – copyover

データ、スクリプト、プラグインは動かしたまま再読み込みできます。でも変更によっては新しい
Go のコードが必要です。ゲームループの新機能、戦闘のバグ修正。それは新しいバイナリを意味し、
新しいバイナリは再起動を意味します。本当に？

**copyover** は 90 年代から MUD が使ってきたトリックです。動いているサーバーが自分の
**新しい** 版を起動し、生き残るべきものをすべて引き渡します。待ち受けソケット、各プレイヤーの
開いた接続、そしてプレイヤーの状態。それから古いプロセスは終了します。プレイヤーが見るのは
1 行だけ:

```
一瞬、世界が震えた……すべては元のまま、それでいて新しい。
```

そして一度も切断されることなく、新しいコードの上で遊び続けます。

```
古いプロセス                                 新しいプロセス（新しいバイナリ）
────────────                                 ────────────────────────────────
リスナーと各接続を複製する          ──▶      fd 3 = リスナー、fd 4, 5, ... = プレイヤー
状態（部屋、HP、アイテム）を JSON で書く ─▶  MUD_COPYOVER=/tmp/state.json
新しいバイナリを起動する（ExtraFiles）
自分のコピーを閉じ、静かに終了する            net.FileListener(fd 3)、net.FileConn(fd 4) ...
                                              プレイヤーを引き取り、リスナーで Serve する
```

## 目標

- **ファイルディスクリプター** を理解する: 接続は OS がプロセスに与える番号で、その番号は
  子プロセスに渡せる。
- `exec.Cmd.ExtraFiles`、`(*net.TCPConn).File()`、`net.FileConn`、`net.FileListener` を使う。
- 1 つの責任を 2 つに分ける: 「ログイン」と「遊ぶ」（`handle` → `play`）。プレイヤーが
  「遊ぶ」から始められるように。
- プロセス間で状態を引き渡す。JSON にシリアライズして。

## Go の概念

### ファイルディスクリプター

`conn.(*net.TCPConn).File()` は、接続のディスクリプターの **複製**（同じソケットを指す
2 つ目の番号）を持つ `*os.File` を返します。ソケットが閉じるのは、それを指す *すべての*
ディスクリプターが閉じたときだけです。だから:

1. 古いプロセスが複製する（`File()`）。
2. 複製を新しいプロセスに渡す（`cmd.ExtraFiles`。向こうでは fd 3, 4, 5… になる）。
3. 自分のディスクリプターを閉じる。新しいプロセスがコピーを持っているので、接続は開いたまま。
4. 新しいプロセスが番号を Go の値に戻す: `net.FileConn(os.NewFile(4, "anna"))`。

### 2 種類の別れ

`Shutdown`（レッスン 14）は別れを告げ、全員をログアウトさせます。`HandOver` はどちらもしては
いけません。メッセージなし、保存なし、「アンナが去った」もなし。フラグ `s.handingOver` が、
プレイヤーのゴルーチンに後片付けを飛ばすよう伝えます。

### 本物の exec

OS の部分は `main.go` が行います（用意済み）。`startNewServer` は状態を一時ファイルに書き、
`ExtraFiles` と環境変数 `MUD_COPYOVER` を付けて `os.Executable()`（ディスク上のバイナリ。
ビルドし直したばかりなら新しいコード）を起動します。`main` はその変数があると、新しい
リスナーを開く代わりにファイルを引き取ります。

### Windows

Windows には、このやり方でソケットを新しいプロセスに渡す方法がありません。そこでは
`(*net.TCPConn).File()` はエラーを返し、`@copyover` はそれを伝えます。copyover は WSL で
試してください。それが必要なテストは Windows では飛ばされます。

## 契約

- `CopyoverPlayer {Name, Room, X, Y, HP, Items, Structured, File}` と
  `CopyoverState {Ticks, Players}`。JSON の名前は `name`、`room`、`x`、`y`、`hp`、`items`、
  `structured`、`file`、`ticks`、`players`。
- `(s *Server) PrepareCopyover() (*CopyoverState, []*os.File, error)` – ゲームループの外から
  呼ぶ。`files[0]` がリスナーで、各プレイヤーの `File` はその接続のファイルの添え字。
- `(s *Server) HandOver(ctx) error` – 別れも告げず、誰もログアウトさせず、保存もせずに、この
  サーバーを止める。接続は新しいプロセスのために開いたまま。
- `(s *Server) Adopt(state, files) (net.Listener, error)` – 新しいサーバーがプレイヤーを
  引き取る: 同じ部屋（なければ開始部屋）、位置（まだ歩けるなら）、HP、アイテム。それぞれに
  `MsgCopyoverDone` と `look` を送り、ログインなしで遊び続けさせる。`Serve` 用のリスナーを返す。
- `(s *Server) Done() <-chan struct{}` – サーバーが止まると閉じる。
- `@copyover`（管理者）: `Options.OnCopyover` があれば、準備 → 新しいサーバーを起動 →
  引き渡し。エラーなら管理者に `MsgCopyoverFailed`。
- `handle` を分ける: ログインは `handle`、その後のすべては `play(p)`。

## 課題

1. `handle` を `handle` と `play` に分け、`s.handingOver` が立っている間は `play` の後片付けが
   何もしないようにする。
2. `copyover.go` は一部用意されています（型、`fileOf`、`Done`、`cmdCopyover`、`copyover`）。
   `PrepareCopyover`、`HandOver`、`Adopt`、`adopt` を書きましょう。
3. `Options.OnCopyover`、フィールド、コマンド。`main.go` は用意済み。

本物で試してみましょう（Linux、macOS、WSL）:

```
cd workspace/server
go build -o mudserver . && ./mudserver -admin Anna
# 別の端末で Anna としてログインし、messages.go のメッセージを変えてから:
go build -o mudserver .
# ゲームの中で: @copyover
```

（`go run` ではなくビルドしたバイナリを使います。`go run` は一時的なバイナリを作って
消すので、新しいプロセスが *古い* コードを起動してしまいます。）

## 考えてみよう

- copyover の *最中に* プレイヤーが打ったコマンドはどうなる？
- 引き渡されない状態は？（戦闘、探索済みのゾーン、生き物の位置など）そのうちプレイヤーが
  気づくのはどれ？
- 新しいバイナリが起動時にクラッシュしたら？ 古いほうがそれに気づいて続けるには？
- なぜ `os.Args[0]` ではなく `os.Executable()` なのでしょう？

## チャレンジ

- もっと引き渡す: 戦闘、生き物の位置、時計。
- ヘルスチェック: 古いプロセスが終わる前に、新しいプロセスがパイプで「起動した」と伝える。
  そうでなければ古いほうが動き続ける。
- Erlang 流のホットコードリローディングはまったくの別物です。調べて比べてみよう。
