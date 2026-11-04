# 環境構築

必要なのは **Go**、**Rust**、UTF-8 と色を扱える **端末**、そしてできれば **telnet** か
**nc** です。方法は 3 つあります。どれか 1 つを選んでください。

| 方法 | こんな人に | 所要時間 |
|---|---|---|
| [A. Dev Container / Codespaces](#a-dev-container-または-github-codespaces) | パソコンに何もインストールせず、すぐに始めたい | 5 分 |
| [B. 自分でインストール](#b-自分でインストール) | すべてを手元に置きたい（長い目で見るとおすすめ） | 15〜30 分 |
| [C. Windows と WSL](#c-windows) | Windows を使っている | 20〜30 分 |

準備ができたら、必ず次を実行します:

```
./mudschool doctor             # Go、Rust、telnet/nc、gcc、UTF-8、色を確認
./mudschool init --lang ja     # 日本語で workspace/ を作る（sv = スウェーデン語、en = 英語）
./mudschool show               # 最初のレッスン
```

---

## A. Dev Container または GitHub Codespaces

リポジトリには `.devcontainer/devcontainer.json` に完成済みの環境があります
（Go 1.24、Rust stable、telnet、nc 入りの Ubuntu と、Go・Rust 用の VS Code 拡張）。

- **GitHub Codespaces:** GitHub のリポジトリのページで **Code → Codespaces →
  Create codespace** をクリック。すべてブラウザの中で動きます。
- **手元の VS Code:** [Docker](https://www.docker.com/) と拡張機能 *Dev Containers* を
  インストールし、フォルダを開いて **Reopen in Container** を選びます。

ポート 4000 は自動で転送されるので、自分のパソコンからもサーバーに接続できます。

## B. 自分でインストール

### 1. Go（1.22 以降）

| システム | コマンド |
|---|---|
| macOS | `brew install go`、または <https://go.dev/dl/> のインストーラー |
| Linux | <https://go.dev/dl/> からダウンロードして手順に従う（ディストリビューションのパッケージは古いことがあります。`go version` で確認） |
| Windows | `winget install GoLang.Go`、または <https://go.dev/dl/> の MSI ファイル |

確認: `go version` が `go1.22` 以降を表示すること。

### 2. Rust（stable）

公式のインストーラー **rustup** で入れます:

```
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
```

Windows では <https://rustup.rs> の `rustup-init.exe` を実行します（または
`winget install Rustlang.Rustup`）。Rust には「C++ によるデスクトップ開発」を含む
*Visual Studio Build Tools* が必要で、インストーラーが確認してくれます。

確認: `rustc --version` と `cargo --version`。

### 3. telnet と nc（任意だがおすすめ）

| システム | コマンド |
|---|---|
| macOS | `brew install telnet`（`nc` は最初から入っています） |
| Ubuntu/Debian | `sudo apt install telnet netcat-openbsd` |
| Fedora | `sudo dnf install telnet nmap-ncat` |
| Windows | 「Windows の機能の有効化または無効化」で *Telnet クライアント* を有効にするか、WSL を使う |

### 4. C コンパイラ（任意）

`./mudschool check --race` を使うときだけ必要です（Go のレース検出器は cgo を使います）。

| システム | コマンド |
|---|---|
| macOS | `xcode-select --install` |
| Ubuntu/Debian | `sudo apt install build-essential` |

### 5. エディタ

おすすめは **VS Code** と拡張機能 **Go**（`golang.go`）、**rust-analyzer**
（`rust-lang.rust-analyzer`）です。補完、コード上のエラー表示、「定義へ移動」が
使えます。JetBrains の GoLand/RustRover や、LSP を設定した Neovim もとても快適です。

ツールが `go.mod` と `Cargo.toml` を見つけられるよう、`workspace/server` と
`workspace/client` を別々のウィンドウで（またはリポジトリ全体を）開きましょう。

### 6. 端末

クライアントは `♣ ≈ ▲ ■ ─ │` などの文字と 24 ビットカラーでマップを描きます。

- **macOS:** **iTerm2**、**WezTerm**、**Ghostty**、**kitty** を使ってください。
  *ターミナル.app はトゥルーカラー非対応* で、色がおかしくなります。
- **Windows:** **Windows Terminal**（Windows 11 には最初から入っています）。
- **Linux:** GNOME Terminal、Konsole、kitty、Alacritty、WezTerm、どれでも動きます。
- **フォント:** Unicode を広くカバーするもの。例: JetBrains Mono、Cascadia Code、
  Fira Code、DejaVu Sans Mono。日本語も表示するなら、HackGen や UDEV Gothic などの
  日本語プログラミングフォントが便利です。
- **ロケール:** UTF-8。例: `export LANG=ja_JP.UTF-8`。
- **文字幅（重要）:** `≈ ▲ ■ ─ │ ♣` などは「East Asian Ambiguous Width」の文字です。
  端末の設定で曖昧幅を **全角（2）** にしているとマップがずれます。**半角（1）** に
  してください（iTerm2: Profiles → Text → "Ambiguous characters are double-width" を
  オフ、WezTerm: `treat_east_asian_ambiguous_width_as_wide = false`）。

## C. Windows

いちばん簡単なのは **WSL 2**（Windows Subsystem for Linux）です:

```
wsl --install -d Ubuntu
```

再起動し、Windows Terminal で *Ubuntu* を開いたら、Linux のコマンドで
[B](#b-自分でインストール) の手順に従ってください。VS Code の *WSL* 拡張を使えば、
Windows 側からファイルを編集できます。

Windows（PowerShell）で直接動かすこともできますが、`./mudschool` は sh スクリプト
です。コースツールを一度だけビルドしましょう:

```
cd tools\mudschool
go build -o ..\..\mudschool.exe .
cd ..\..
.\mudschool.exe doctor
```

以後は `./mudschool` の代わりに `.\mudschool.exe` を使います。

---

## コースの入手

```
git clone https://github.com/mattias/mud-school.git
cd mud-school
./mudschool doctor
./mudschool init --lang ja
```

Rust のレッスンを初めて実行するときは、ライブラリ（クレート）をダウンロードする
ので少し時間がかかります。初回だけです。

## トラブルシューティング

| 問題 | 解決方法 |
|---|---|
| `go: command not found` / `cargo: command not found` | 端末を開き直す。`~/go/bin`、`/usr/local/go/bin`、`~/.cargo/bin` が `PATH` にあるか確認する。 |
| プロキシの内側で `cargo` がクレートをダウンロードできない | `HTTPS_PROXY`（必要なら `CARGO_HTTP_CAINFO` も）を設定する。 |
| クライアントが落ちた後、端末の様子がおかしい | `reset` と入力して Enter（クライアントが端末を「raw モード」のままにしました）。 |
| `address already in use` | ポート 4000 ですでに何かが動いています。止めるか、`./mudschool run server -addr :4001` と `./mudschool run client 127.0.0.1:4001` を使う。 |
| マップに変な文字や四角が出る | 端末かフォントが UTF-8 に対応していません。[端末](#6-端末) を参照。 |
| マップがずれる | 曖昧幅の文字が全角で表示されています。[端末](#6-端末) の「文字幅」を参照。 |
| 色が出ない・おかしい | 端末がトゥルーカラー非対応です（例: macOS のターミナル.app）。端末を変えてください。 |
| `-race requires cgo` | C コンパイラを入れる（手順 4）か、`--race` なしで実行する。 |
| Windows ファイアウォールがネットワークアクセスを尋ねてくる | プライベートネットワークで許可する。サーバーはポート 4000 で待ち受けます。 |

それでも動かない？ `./mudschool doctor` を実行し、赤い項目を読んでください。
