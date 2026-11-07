# Setup and environment

> 🇬🇧 English · [🇸🇪 Svenska](../sv/SETUP.md) · [🇯🇵 日本語](../ja/SETUP.md)

You need **Go 1.22+**, **Rust** (stable), **Git**, a **terminal** with UTF-8
and 24-bit colour. (telnet or nc are optional – the course tool has a
built-in client, `./mudschool connect`.) Pick the section for
your system and follow it from top to bottom – each one is complete on its own.

| Your system | Section | Time |
|---|---|---|
| Windows 10/11 | [Windows (PowerShell)](#windows-powershell) – recommended on Windows | 20–30 min |
| Windows, if you prefer Linux tools | [Windows with WSL](#windows-with-wsl) | 20–30 min |
| macOS | [macOS](#macos) | 15 min |
| Linux | [Linux](#linux) | 15 min |
| Nothing installed, just a browser | [Codespaces / Dev Container](#github-codespaces-or-a-dev-container) | 5 min |

Then: [Editor](#editor) · [Fonts and terminals](#fonts-and-terminals) · [Troubleshooting](#troubleshooting)

> **About the commands in the lessons.** The lessons write `./mudschool show`
> etc. On Windows (PowerShell or cmd) you type **`.\mudschool.cmd show`**
> instead. The course tool itself prints its commands the right way for your system.

---

## Windows (PowerShell)

Everything below is typed in **PowerShell** inside **Windows Terminal**.

### 1. Windows Terminal

Windows 11 has it already. On Windows 10:

```powershell
winget install Microsoft.WindowsTerminal
```

Use Windows Terminal for the whole course – the old console window
(`conhost`) shows neither the colours nor the map characters correctly.
Choose a font right away: see [Fonts and terminals](#fonts-and-terminals).

### 2. Git, Go and Rust

```powershell
winget install Git.Git
winget install GoLang.Go
winget install Rustlang.Rustup
```

Rust on Windows also needs Microsoft's C++ build tools (the linker):

```powershell
winget install Microsoft.VisualStudio.2022.BuildTools --override "--wait --passive --add Microsoft.VisualStudio.Workload.VCTools --includeRecommended"
```

**Close and reopen Windows Terminal** so that `PATH` is updated, then check:

```powershell
go version        # go1.22 or later
rustc --version
cargo --version
```

> **Microsoft Defender quarantined `cargo.exe`?** It happens now and then:
> Defender's heuristics flag the small "proxy" programs rustup installs in
> `%USERPROFILE%\.cargo\bin` (`cargo.exe`, `rustc.exe`, … are all copies of
> `rustup.exe`). Don't just click *Allow* – check first:
>
> 1. Update Defender (*Windows Security → Virus & threat protection →
>    Protection updates → Check for updates*).
> 2. Reinstall Rust from the official source: `winget uninstall Rustlang.Rustup`,
>    then `winget install Rustlang.Rustup` (or `rustup-init.exe` from
>    <https://rustup.rs>).
> 3. If the fresh file is quarantined again, check that rustup is genuine.
>    `cargo.exe` is gone, but it is a hard link to `rustup.exe`, which is a
>    copy of the official `rustup-init.exe`. Run `rustup self update`, then
>    compare `(Get-FileHash $env:USERPROFILE\.cargo\bin\rustup.exe).Hash` with
>    the first word of <https://static.rust-lang.org/rustup/dist/x86_64-pc-windows-msvc/rustup-init.exe.sha256> (ignore upper/lower case; on ARM use
>    `aarch64-pc-windows-msvc` in the address).
> 4. If they match, it is a false positive: report the file to Microsoft at
>    <https://www.microsoft.com/wdsi/filesubmission>, then in *Protection
>    history* choose the item → *Actions* → *Allow on device*. A plain
>    *Restore* may be quarantined again until Microsoft updates its
>    definitions. If they don't match – don't allow anything; ask for help.
>
> Programs you build yourself (`target\debug\*.exe`) can be flagged the
> same way; report them as in step 4.

### 3. Connecting by hand – nothing to install

In lessons 01–02, before your own client exists, you try the server by hand.
The course tool has a little client built in:

```powershell
.\mudschool.cmd connect          # = 127.0.0.1:4000
```

It shows å, ä, ö and Japanese correctly in every Windows terminal, without
any settings. The lessons use `./mudschool connect`; where they mention
telnet or nc, this works the same.

#### If you want telnet or ncat anyway

Windows' own telnet doesn't understand UTF-8 at all. ncat (`winget install
Insecure.Nmap`, then reopen the terminal) works if the console expects UTF-8
– which PowerShell doesn't by default (Japanese Windows uses Shift-JIS, code
page 932). The simplest way to fix that for every new tab, without touching
any files: in *Windows Terminal → Settings → (your PowerShell profile) →
Command line*, use

```
pwsh.exe -NoLogo -NoExit -Command "[Console]::InputEncoding = [Console]::OutputEncoding = $OutputEncoding = [System.Text.UTF8Encoding]::new()"
```

(Use `powershell.exe` instead of `pwsh.exe` for Windows PowerShell 5.1.)
(If you edit `settings.json` by hand instead, write each `"` inside the
command as `\"`.) Check in a new tab with `chcp` – it should say `65001`.

The PowerShell profile (`$PROFILE`) would also work, but it lives in
*Documents*, which Defender's **controlled folder access** often protects
("Protected folder access blocked … pwsh.exe"), and it needs
`Set-ExecutionPolicy`. The Windows Terminal setting avoids both.

### 4. A C compiler (optional)

Only needed for `.\mudschool.cmd check --race` (Go's race detector uses cgo).
The simplest route is GCC from MSYS2:

```powershell
winget install MSYS2.MSYS2
```

Start **"MSYS2 UCRT64"** from the Start menu and run:

```sh
pacman -S --needed mingw-w64-ucrt-x86_64-gcc
```

Add `C:\msys64\ucrt64\bin` to your `PATH` (Start → "Edit environment
variables for your account" → `Path` → New), reopen Windows Terminal and
check with `gcc --version`.

### 5. Get the course and start

```powershell
cd $HOME
git clone https://github.com/mattias/mud-school.git
cd mud-school
.\mudschool.cmd doctor            # checks Go, Rust, gcc and the terminal
.\mudschool.cmd init --lang en    # creates workspace\ (en, sv or ja)
.\mudschool.cmd show              # the first lesson
```

`mudschool.cmd` builds the course tool automatically the first time. The
first time a Rust lesson runs, Cargo downloads libraries, which takes a while.
The server only listens on `127.0.0.1` (your own computer), so the firewall
doesn't need to ask. To let others on your network connect, start it with
`.\mudschool.cmd run server -addr :4000`; then Windows Defender Firewall asks
once – allow it for private networks.

---

## Windows with WSL

WSL gives you a real Linux inside Windows. Choose it if you prefer Linux
tools or already use WSL.

```powershell
wsl --install -d Ubuntu
```

Restart the computer, open the **Ubuntu** profile in Windows Terminal and then
follow the [Linux](#linux) section inside Ubuntu. Clone the course into your
Linux home directory (`~/mud-school`), not under `/mnt/c/` – it is much faster.
VS Code's *WSL* extension lets you edit the files from Windows.

---

## macOS

### 1. Tools

With [Homebrew](https://brew.sh):

```sh
brew install go git telnet
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
xcode-select --install            # C compiler, needed for --race
```

Reopen the terminal, then check `go version`, `rustc --version` and `cargo --version`.

### 2. A terminal with 24-bit colour

**Terminal.app has long lacked truecolor** (at least before macOS 26) – the map's colours may come out wrong. Use
**iTerm2**, **WezTerm**, **Ghostty** or **kitty** (e.g. `brew install --cask iterm2`).

### 3. Get the course and start

```sh
git clone https://github.com/mattias/mud-school.git
cd mud-school
./mudschool doctor
./mudschool init --lang en
./mudschool show
```

---

## Linux

### 1. Tools

Go: the distribution's package is often too old. Install the latest from
<https://go.dev/dl/> as described there, or with snap: `sudo snap install go --classic`.

Rust:

```sh
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
```

The rest (Ubuntu/Debian; on Fedora use `sudo dnf install git telnet nmap-ncat gcc`):

```sh
sudo apt install git telnet netcat-openbsd build-essential
```

Reopen the terminal, then check `go version`, `rustc --version` and `cargo --version`.

### 2. Get the course and start

```sh
git clone https://github.com/mattias/mud-school.git
cd mud-school
./mudschool doctor
./mudschool init --lang en
./mudschool show
```

GNOME Terminal, Konsole, kitty, Alacritty and WezTerm all support UTF-8 and
24-bit colour. Make sure your locale is UTF-8 (e.g. `LANG=en_US.UTF-8`).

---

## GitHub Codespaces or a Dev Container

The repository has a ready-made environment in `.devcontainer/` (Ubuntu with
Go 1.24, Rust, telnet and nc, plus the VS Code extensions for Go and Rust).

- **GitHub Codespaces:** on the repository's page, click **Code → Codespaces →
  Create codespace**. Everything runs in the browser.
- **VS Code locally:** install [Docker](https://www.docker.com/) and the *Dev
  Containers* extension, open the folder and choose **Reopen in Container**.

Then follow [Linux, step 2](#2-get-the-course-and-start) (the course is
already cloned). Port 4000 is forwarded automatically.

---

## Editor

Recommendation: **VS Code** with the extensions **Go** (`golang.go`) and
**rust-analyzer** (`rust-lang.rust-analyzer`): completion, errors right in the
code and "go to definition". JetBrains GoLand/RustRover and Neovim with LSP
work very well too. Open the whole `mud-school` folder (or `workspace/server`
and `workspace/client` separately) so the tools find `go.mod` and `Cargo.toml`.

## Fonts and terminals

The client draws the map with characters such as `♣ ≈ ▲ ■ Ω ─ │ ♥` in 24-bit colour.

| Course language | Recommended fonts |
|---|---|
| English, Swedish | **Cascadia Mono** (the default in Windows Terminal), JetBrains Mono, Fira Code, DejaVu Sans Mono |
| Japanese | **UDEV Gothic** or **HackGen** (free programming fonts with both Latin and Japanese characters); built into Windows: **BIZ UDGothic** |

**Windows Terminal:** Settings (`Ctrl+,`) → *Profiles* → *Defaults* →
*Appearance* → *Font face*. Install a downloaded font first (right-click the
`.ttf` file → *Install*) and restart Windows Terminal. BIZ UDGothic is preinstalled on Japanese Windows; elsewhere add it under Settings → System → Optional features → *Japanese Supplemental Fonts*. If a font lacks a
character, Windows Terminal borrows it from another font automatically.

**Japanese text and the map.** Characters such as `≈ ▲ ■ ─ │ ♣` are "East
Asian ambiguous width". The map assumes they are **one cell wide**:

- Windows Terminal always draws them one cell wide – nothing to change.
- iTerm2: Profiles → Text → turn off *Ambiguous characters are double-width*.
- WezTerm: `treat_east_asian_ambiguous_width_as_wide = false` (the default).
- Choose the normal variant of UDEV Gothic/HackGen, not the variants that make
  ambiguous characters full-width (e.g. "UDEV Gothic JPDOC").

Japanese room names and descriptions themselves are two cells wide, which the
client handles correctly.

## Troubleshooting

| Problem | Solution |
|---|---|
| `go`/`cargo` is "not recognized" / "command not found" | Reopen the terminal. Check that Go's `bin` folder and `~/.cargo/bin` (Windows: `%USERPROFILE%\.cargo\bin`) are in `PATH`. |
| Windows: `link.exe not found` when building Rust | Install the C++ build tools (Windows, step 2). |
| `cargo` cannot download crates behind a proxy | Set `HTTPS_PROXY` (and `CARGO_HTTP_CAINFO` if needed). |
| The terminal behaves strangely after the client crashed | Type `reset` (Windows: close the tab) – the client left the terminal in "raw mode". |
| `address already in use` | Something is already on port 4000. Stop it, or run `./mudschool run server -addr 127.0.0.1:4001` and `./mudschool run client 127.0.0.1:4001`. |
| Boxes or odd characters on the map | The font lacks the characters or the terminal isn't UTF-8 – see [Fonts and terminals](#fonts-and-terminals). |
| Windows: telnet/ncat shows mojibake (e.g. Japanese) | Use `.\mudschool.cmd connect` instead – see [Connecting by hand](#3-connecting-by-hand--nothing-to-install). |
| Windows: Defender quarantined `cargo.exe` | See the note under [Git, Go and Rust](#2-git-go-and-rust). |
| The map's columns don't line up (Japanese) | Ambiguous-width characters are drawn double-width – see [Fonts and terminals](#fonts-and-terminals). |
| No colours or wrong colours | The terminal lacks 24-bit colour (e.g. macOS Terminal.app, the old Windows console). |
| `-race requires cgo` | Install a C compiler (see your system's section) or run without `--race`. |
| Windows: Ctrl+C asks "Terminate batch job (Y/N)?" | That is cmd.exe asking because `mudschool.cmd` is a batch file. Answer `Y`. |
| `doctor` warns about UTF-8/colour although everything looks fine | Your terminal wasn't recognised. If the map looks right, you can ignore it. |

Still not working? Run `doctor` and read what is red or yellow.
