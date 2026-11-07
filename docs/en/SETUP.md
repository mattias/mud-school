# Setup and environment

> 🇬🇧 English · [🇸🇪 Svenska](../sv/SETUP.md) · [🇯🇵 日本語](../ja/SETUP.md)

You need **Go 1.22+**, **Rust** (stable), **Git**, a **terminal** with UTF-8
and 24-bit colour, and preferably **telnet** or **nc**. Pick the section for
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

### 3. ncat (optional but recommended)

To try your server by hand before your own client exists (lessons 01–02),
you need a simple line-based client. On Windows, use **ncat** (from Nmap):

```powershell
winget install Insecure.Nmap
```

Accept the installer's defaults (Npcap is optional – you don't need it), then
**close and reopen Windows Terminal** so that `ncat` is in `PATH`. Then:

```powershell
ncat 127.0.0.1 4000
```

> Windows' own telnet client (`dism /online /Enable-Feature /FeatureName:TelnetClient`)
> works too, but it shows å, ä, ö and Japanese characters as mojibake – it
> doesn't understand UTF-8.

#### UTF-8 in every new terminal

ncat writes what the server sends straight to the console, so the console
must expect UTF-8. PowerShell doesn't by default (Japanese Windows uses
Shift-JIS, code page 932). Add this to your PowerShell profile **once**, and
every new terminal uses UTF-8:

```powershell
Set-ExecutionPolicy -Scope CurrentUser RemoteSigned
if (!(Test-Path $PROFILE)) { New-Item -ItemType File -Path $PROFILE -Force | Out-Null }
Add-Content $PROFILE '[Console]::InputEncoding = [Console]::OutputEncoding = $OutputEncoding = [System.Text.UTF8Encoding]::new()'
```

The first line lets PowerShell run your own profile at all – by default
Windows refuses to run scripts, and the profile is a script (no
administrator needed). Run the commands in the shell you actually use in
Windows Terminal: Windows PowerShell and PowerShell 7 have different profiles.

Open a new tab and check with `chcp` – it should say `65001`. `doctor` warns
if the code page is something else.

(The alternative is system-wide: *Settings → Time & language → Language &
region → Administrative language settings → Change system locale → "Beta: Use
Unicode UTF-8 for worldwide language support"*, then restart. It also affects
older programs, which may then show other characters wrong.)

From lesson 03 you have your own Rust client, which shows everything
correctly without any of this.

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
.\mudschool.cmd doctor            # checks Go, Rust, ncat/telnet, gcc, the terminal and the code page
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
| Windows: telnet/ncat shows mojibake (e.g. Japanese) | Use ncat and make the console UTF-8 – see [UTF-8 in every new terminal](#utf-8-in-every-new-terminal). |
| The map's columns don't line up (Japanese) | Ambiguous-width characters are drawn double-width – see [Fonts and terminals](#fonts-and-terminals). |
| No colours or wrong colours | The terminal lacks 24-bit colour (e.g. macOS Terminal.app, the old Windows console). |
| `-race requires cgo` | Install a C compiler (see your system's section) or run without `--race`. |
| Windows: Ctrl+C asks "Terminate batch job (Y/N)?" | That is cmd.exe asking because `mudschool.cmd` is a batch file. Answer `Y`. |
| `doctor` warns about UTF-8/colour although everything looks fine | Your terminal wasn't recognised. If the map looks right, you can ignore it. |

Still not working? Run `doctor` and read what is red or yellow.
