# 環境構築

> [🇬🇧 English](../en/SETUP.md) · [🇸🇪 Svenska](../sv/SETUP.md) · 🇯🇵 日本語

必要なのは **Go 1.22 以降**、**Rust**（stable）、**Git**、UTF-8 と 24 ビットカラーに対応した
**端末** です。（telnet や nc は任意です。コースツールに内蔵クライアント `./mudschool connect` があります。）自分のシステムの節を選び、上から順に
進めてください。どの節もそれだけで完結しています。

| システム | 節 | 所要時間 |
|---|---|---|
| Windows 10/11 | [Windows（PowerShell）](#windowspowershell) – Windows ではこれがおすすめ | 20〜30 分 |
| Windows で Linux のツールを使いたい | [Windows と WSL](#windows-と-wsl) | 20〜30 分 |
| macOS | [macOS](#macos) | 15 分 |
| Linux | [Linux](#linux) | 15 分 |
| 何もインストールせずブラウザだけで | [Codespaces / Dev Container](#github-codespaces-または-dev-container) | 5 分 |

その後: [エディタ](#エディタ) · [フォントと端末](#フォントと端末) · [トラブルシューティング](#トラブルシューティング)

> **レッスン中のコマンドについて。** レッスンでは `./mudschool show` のように書いています。
> Windows（PowerShell または cmd）では代わりに **`.\mudschool.cmd show`** と入力して
> ください。コースツール自身が表示するコマンドは、あなたのシステムに合った形になります。

---

## Windows（PowerShell）

以下はすべて **Windows Terminal** の中の **PowerShell** で入力します。

### 1. Windows Terminal

Windows 11 には最初から入っています。Windows 10 では:

```powershell
winget install Microsoft.WindowsTerminal
```

コースでは常に Windows Terminal を使ってください。古いコンソールウィンドウ（`conhost`）では
色もマップの文字も正しく表示されません。フォントもすぐに選んでおきましょう:
[フォントと端末](#フォントと端末) を参照。

### 2. Git、Go、Rust

```powershell
winget install Git.Git
winget install GoLang.Go
winget install Rustlang.Rustup
```

Windows の Rust には Microsoft の C++ ビルドツール（リンカー）も必要です:

```powershell
winget install Microsoft.VisualStudio.2022.BuildTools --override "--wait --passive --add Microsoft.VisualStudio.Workload.VCTools --includeRecommended"
```

`PATH` を反映させるため **Windows Terminal をいったん閉じて開き直し**、確認します:

```powershell
go version        # go1.22 以降
rustc --version
cargo --version
```

> **Microsoft Defender が `cargo.exe` を検疫した？** ときどき起こります。rustup が
> `%USERPROFILE%\.cargo\bin` に置く小さな「プロキシ」プログラム（`cargo.exe`、`rustc.exe`
> などはすべて `rustup.exe` のコピー）を、Defender のヒューリスティックが検出することが
> あるのです。すぐに *許可* を押さず、まず確認しましょう:
>
> 1. Defender を更新する（*Windows セキュリティ → ウイルスと脅威の防止 → 保護の更新 →
>    更新プログラムのチェック*）。
> 2. 公式の入手元から Rust を入れ直す: `winget uninstall Rustlang.Rustup` の後に
>    `winget install Rustlang.Rustup`（または <https://rustup.rs> の `rustup-init.exe`）。
> 3. 新しいファイルもまた検疫されたら、rustup が本物か確かめます。`cargo.exe` は消えて
>    いますが、これは `rustup.exe` へのハードリンクで、`rustup.exe` は公式の
>    `rustup-init.exe` のコピーです。`rustup self update` を実行してから、
>    `(Get-FileHash $env:USERPROFILE\.cargo\bin\rustup.exe).Hash` を
>    <https://static.rust-lang.org/rustup/dist/x86_64-pc-windows-msvc/rustup-init.exe.sha256> の最初の語と比べます（大文字・小文字は無視）。
> 4. 一致すれば誤検出です: <https://www.microsoft.com/wdsi/filesubmission> から
>    Microsoft に報告し、*保護の履歴* でその項目 → *操作* → *デバイスで許可* を選びます。
>    単なる *復元* では、Microsoft が定義を更新するまで再び検疫されることがあります。
>    一致しなければ何も許可せず、助けを求めてください。
>
> 自分でビルドしたプログラム（`target\debug\*.exe`）も同じように検出されることがあります。
> 手順 4 のように報告してください。

### 3. 手で接続する – インストール不要

レッスン 01〜02 では、自分のクライアントができる前にサーバーを手で試します。コースツールには
小さなクライアントが内蔵されています:

```powershell
.\mudschool.cmd connect          # = 127.0.0.1:4000
```

設定なしで、どの Windows 端末でも日本語が正しく表示されます。レッスンでは
`./mudschool connect` を使います。telnet や nc とあるところでも同じように使えます。

#### それでも telnet や ncat を使いたい場合

Windows 標準の telnet は UTF-8 をまったく理解しません。ncat（`winget install
Insecure.Nmap` の後、端末を開き直す）は、コンソールが UTF-8 を前提にしていれば動きますが、
PowerShell は最初はそうなっていません（日本語版 Windows は Shift-JIS、コードページ 932）。
ファイルに触れずに、新しいタブすべてでこれを直す一番簡単な方法: *Windows Terminal → 設定 →
（使っている PowerShell のプロファイル）→ コマンド ライン* を次にします

```
pwsh.exe -NoLogo -NoExit -Command "[Console]::InputEncoding = [Console]::OutputEncoding = $OutputEncoding = [System.Text.UTF8Encoding]::new()"
```

（Windows PowerShell 5.1 なら `pwsh.exe` の代わりに `powershell.exe`。）（`settings.json` を直接編集する場合は、コマンド内の `"` を `\"` と書きます。）新しいタブで `chcp`
を実行し、`65001` と表示されれば OK です。

PowerShell のプロファイル（`$PROFILE`）でも可能ですが、プロファイルは *ドキュメント* にあり、
Defender の **コントロールされたフォルダー アクセス** がよく保護しています（「保護されたフォルダーへの
アクセスがブロックされました … pwsh.exe」）。さらに `Set-ExecutionPolicy` も必要です。
Windows Terminal の設定ならどちらも避けられます。

### 4. C コンパイラ（任意）

`.\mudschool.cmd check --race` を使うときだけ必要です（Go のレース検出器は cgo を使います）。
いちばん簡単なのは MSYS2 の GCC です:

```powershell
winget install MSYS2.MSYS2
```

スタートメニューから **「MSYS2 UCRT64」** を起動して実行します:

```sh
pacman -S --needed mingw-w64-ucrt-x86_64-gcc
```

`C:\msys64\ucrt64\bin` を `PATH` に追加し（スタート →「アカウントの環境変数を編集」→
`Path` → 新規）、Windows Terminal を開き直して `gcc --version` で確認します。

### 5. コースを入手して始める

```powershell
cd $HOME
git clone https://github.com/mattias/mud-school.git
cd mud-school
.\mudschool.cmd doctor            # Go、Rust、gcc、端末を確認
.\mudschool.cmd init --lang ja    # workspace\ を作る（en、sv、ja）
.\mudschool.cmd show              # 最初のレッスン
```

`mudschool.cmd` は初回にコースツールを自動でビルドします。Rust のレッスンを初めて
実行するときは Cargo がライブラリをダウンロードするので、少し時間がかかります。
サーバーは `127.0.0.1`（自分のパソコン）だけで待ち受けるので、ファイアウォールの確認は
出ません。同じネットワークの他の人にも接続してもらうには
`.\mudschool.cmd run server -addr :4000` で起動します。このとき Windows Defender
ファイアウォールが一度だけ尋ねてくるので、プライベートネットワークで許可してください。

---

## Windows と WSL

WSL を使うと Windows の中で本物の Linux が動きます。Linux のツールが好きな人や、
すでに WSL を使っている人向けです。

```powershell
wsl --install -d Ubuntu
```

再起動したら Windows Terminal で **Ubuntu** のプロファイルを開き、Ubuntu の中で
[Linux](#linux) の節に従ってください。コースは `/mnt/c/` の下ではなく Linux の
ホームディレクトリ（`~/mud-school`）にクローンしましょう。そのほうがずっと速く
動きます。VS Code の *WSL* 拡張を使えば Windows 側から編集できます。

---

## macOS

### 1. ツール

[Homebrew](https://brew.sh) で:

```sh
brew install go git telnet
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
xcode-select --install            # C コンパイラ（--race に必要）
```

端末を開き直し、`go version`、`rustc --version`、`cargo --version` を確認します。

### 2. 24 ビットカラー対応の端末

**ターミナル.app は長らくトゥルーカラー非対応** でした（少なくとも macOS 26 より前）。マップの色がおかしくなることがあります。
**iTerm2**、**WezTerm**、**Ghostty**、**kitty** を使ってください（例: `brew install --cask iterm2`）。

### 3. コースを入手して始める

```sh
git clone https://github.com/mattias/mud-school.git
cd mud-school
./mudschool doctor
./mudschool init --lang ja
./mudschool show
```

---

## Linux

### 1. ツール

Go: ディストリビューションのパッケージは古いことが多いです。<https://go.dev/dl/> の
手順で最新版を入れるか、snap を使います: `sudo snap install go --classic`。

Rust:

```sh
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
```

その他（Ubuntu/Debian。Fedora では `sudo dnf install git telnet nmap-ncat gcc`）:

```sh
sudo apt install git telnet netcat-openbsd build-essential
```

端末を開き直し、`go version`、`rustc --version`、`cargo --version` を確認します。

### 2. コースを入手して始める

```sh
git clone https://github.com/mattias/mud-school.git
cd mud-school
./mudschool doctor
./mudschool init --lang ja
./mudschool show
```

GNOME Terminal、Konsole、kitty、Alacritty、WezTerm はどれも UTF-8 と 24 ビットカラーに
対応しています。ロケールが UTF-8 になっているか確認してください（例: `LANG=ja_JP.UTF-8`）。

---

## GitHub Codespaces または Dev Container

リポジトリの `.devcontainer/` に完成済みの環境があります（Go 1.24、Rust、telnet、nc 入りの
Ubuntu と、Go・Rust 用の VS Code 拡張）。

- **GitHub Codespaces:** リポジトリのページで **Code → Codespaces → Create codespace** を
  クリック。すべてブラウザの中で動きます。
- **手元の VS Code:** [Docker](https://www.docker.com/) と拡張機能 *Dev Containers* を
  インストールし、フォルダを開いて **Reopen in Container** を選びます。

その後 [Linux の手順 2](#2-コースを入手して始める) に従ってください（クローンは済んで
います）。ポート 4000 は自動で転送されます。

---

## エディタ

おすすめは **VS Code** と拡張機能 **Go**（`golang.go`）、**rust-analyzer**
（`rust-lang.rust-analyzer`）です。補完、コード上のエラー表示、「定義へ移動」が使えます。
JetBrains の GoLand/RustRover や、LSP を設定した Neovim も快適です。ツールが `go.mod` と
`Cargo.toml` を見つけられるよう、`mud-school` フォルダ全体（または `workspace/server` と
`workspace/client` を別々に）開いてください。

## フォントと端末

クライアントは `♣ ≈ ▲ ■ Ω ─ │ ♥` などの文字と 24 ビットカラーでマップを描きます。

| コースの言語 | おすすめのフォント |
|---|---|
| 英語、スウェーデン語 | **Cascadia Mono**（Windows Terminal の標準）、JetBrains Mono、Fira Code、DejaVu Sans Mono |
| 日本語 | **UDEV Gothic** または **HackGen**（英字と日本語の両方を含む無料のプログラミングフォント）。Windows 標準なら **BIZ UDゴシック** |

**Windows Terminal:** 設定（`Ctrl+,`）→ *プロファイル* → *既定値* → *外観* → *フォント フェイス*。
ダウンロードしたフォントは先にインストールし（`.ttf` を右クリック → *インストール*）、
Windows Terminal を再起動してください。BIZ UDゴシックは日本語版 Windows には最初から入っています。それ以外の言語の Windows では、設定 → システム → オプション機能 → *日本語の追加フォント*（Japanese Supplemental Fonts）で追加できます。フォントにない文字は、Windows Terminal が自動で
別のフォントから補います。

**日本語のテキストとマップ。** `≈ ▲ ■ ─ │ ♣` などは「East Asian Ambiguous Width（曖昧幅）」
の文字です。マップはこれらを **1 マス幅** として描きます:

- Windows Terminal は常に 1 マス幅で描くので、設定変更は不要です。
- iTerm2: Profiles → Text → *Ambiguous characters are double-width* をオフ。
- WezTerm: `treat_east_asian_ambiguous_width_as_wide = false`（既定値）。
- UDEV Gothic/HackGen は通常版を選び、曖昧幅の文字を全角にする版
  （例: 「UDEV Gothic JPDOC」）は避けてください。

日本語の部屋名や説明そのものは 2 マス幅で、クライアントはそれを正しく扱います。

## トラブルシューティング

| 問題 | 解決方法 |
|---|---|
| `go`/`cargo` が「認識されません」/「command not found」 | 端末を開き直す。Go の `bin` フォルダと `~/.cargo/bin`（Windows: `%USERPROFILE%\.cargo\bin`）が `PATH` にあるか確認する。 |
| Windows: Rust のビルドで `link.exe not found` | C++ ビルドツールを入れる（Windows の手順 2）。 |
| プロキシの内側で `cargo` がクレートをダウンロードできない | `HTTPS_PROXY`（必要なら `CARGO_HTTP_CAINFO` も）を設定する。 |
| クライアントが落ちた後、端末の様子がおかしい | `reset` と入力（Windows ではタブを閉じる）。クライアントが端末を「raw モード」のままにしました。 |
| `address already in use` | ポート 4000 ですでに何かが動いています。止めるか、`./mudschool run server -addr 127.0.0.1:4001` と `./mudschool run client 127.0.0.1:4001` を使う。 |
| マップに四角や変な文字が出る | フォントに文字がないか、端末が UTF-8 ではありません。[フォントと端末](#フォントと端末) を参照。 |
| Windows: telnet/ncat で日本語が文字化けする | 代わりに `.\mudschool.cmd connect` を使います。[手で接続する](#3-手で接続する--インストール不要) を参照。 |
| Windows: Defender が `cargo.exe` を検疫した | [Git、Go、Rust](#2-gitgorust) の注意書きを参照。 |
| マップの列がずれる（日本語） | 曖昧幅の文字が全角で描かれています。[フォントと端末](#フォントと端末) を参照。 |
| 色が出ない・おかしい | 端末が 24 ビットカラー非対応です（例: macOS のターミナル.app、古い Windows コンソール）。 |
| `-race requires cgo` | C コンパイラを入れる（各システムの節を参照）か、`--race` なしで実行する。 |
| Windows: Ctrl+C で「Terminate batch job (Y/N)?」と聞かれる | `mudschool.cmd` がバッチファイルなので cmd.exe が確認しています。`Y` と答えてください。 |
| 問題なく見えるのに `doctor` が UTF-8／色の警告を出す | 端末を判別できませんでした。マップが正しく見えるなら無視してかまいません。 |

それでも動かない？ `doctor` を実行し、赤や黄色の項目を読んでください。
