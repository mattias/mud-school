# Setup and environment

You need **Go**, **Rust**, a **terminal** that handles UTF-8 and colour, and
preferably **telnet** or **nc**. There are three ways – pick one.

| Way | Suits you if you ... | Time |
|---|---|---|
| [A. Dev Container / Codespaces](#a-dev-container-or-github-codespaces) | want to start right away without installing anything on your computer | 5 min |
| [B. Install it yourself](#b-install-it-yourself) | want everything locally (recommended in the long run) | 15–30 min |
| [C. Windows with WSL](#c-windows) | use Windows | 20–30 min |

When you are done, always run:

```
./mudschool doctor             # checks Go, Rust, telnet/nc, gcc, UTF-8, colour
./mudschool init --lang en     # creates workspace/ in English (sv = Swedish, ja = Japanese)
./mudschool show               # the first lesson
```

---

## A. Dev Container or GitHub Codespaces

The repository has a ready-made environment in `.devcontainer/devcontainer.json`
(Ubuntu with Go 1.24, Rust stable, telnet and nc, plus the VS Code extensions
for Go and Rust).

- **GitHub Codespaces:** on the repository's page on GitHub, click **Code →
  Codespaces → Create codespace**. Everything runs in the browser.
- **VS Code locally:** install [Docker](https://www.docker.com/) and the *Dev
  Containers* extension, open the folder and choose **Reopen in Container**.

Port 4000 is forwarded automatically, so you can connect to the server from
your own computer too.

## B. Install it yourself

### 1. Go (1.22 or later)

| System | Command |
|---|---|
| macOS | `brew install go` – or the installer from <https://go.dev/dl/> |
| Linux | download from <https://go.dev/dl/> and follow the instructions (distribution packages may be too old – check with `go version`) |
| Windows | `winget install GoLang.Go` – or the MSI file from <https://go.dev/dl/> |

Check: `go version` should show `go1.22` or later.

### 2. Rust (stable)

Install with **rustup**, the official installer:

```
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
```

On Windows: run `rustup-init.exe` from <https://rustup.rs> (or
`winget install Rustlang.Rustup`). Rust then needs the *Visual Studio Build
Tools* with "Desktop development with C++" – the installer asks.

Check: `rustc --version` and `cargo --version`.

### 3. telnet and nc (optional but recommended)

| System | Command |
|---|---|
| macOS | `brew install telnet` (`nc` is already there) |
| Ubuntu/Debian | `sudo apt install telnet netcat-openbsd` |
| Fedora | `sudo dnf install telnet nmap-ncat` |
| Windows | enable *Telnet Client* under "Turn Windows features on or off", or use WSL |

### 4. A C compiler (optional)

Only needed for `./mudschool check --race` (Go's race detector uses cgo).

| System | Command |
|---|---|
| macOS | `xcode-select --install` |
| Ubuntu/Debian | `sudo apt install build-essential` |

### 5. Editor

Recommendation: **VS Code** with the extensions **Go** (`golang.go`) and
**rust-analyzer** (`rust-lang.rust-analyzer`). They give you completion,
errors right in the code and "go to definition". JetBrains GoLand/RustRover
and Neovim with LSP also work very well.

Open `workspace/server` and `workspace/client` as separate windows (or the
whole repository) so the tools find `go.mod` and `Cargo.toml`.

### 6. The terminal

The client draws the map with characters like `♣ ≈ ▲ ■ ─ │` in 24-bit colour.

- **macOS:** use **iTerm2**, **WezTerm**, **Ghostty** or **kitty**.
  *Terminal.app lacks truecolor* – the colours come out wrong.
- **Windows:** **Windows Terminal** (preinstalled in Windows 11).
- **Linux:** GNOME Terminal, Konsole, kitty, Alacritty, WezTerm – all work.
- **Font:** one with good Unicode coverage, e.g. JetBrains Mono, Cascadia
  Code, Fira Code or DejaVu Sans Mono.
- **Locale:** UTF-8, e.g. `export LANG=en_US.UTF-8`.

## C. Windows

The easiest way is **WSL 2** (Windows Subsystem for Linux):

```
wsl --install -d Ubuntu
```

Restart, open *Ubuntu* in Windows Terminal and then follow [B](#b-install-it-yourself)
with the Linux commands. VS Code has the *WSL* extension so you can edit the
files from Windows.

Running directly in Windows (PowerShell) works too, but the `./mudschool`
script is a sh script. Build the course tool once:

```
cd tools\mudschool
go build -o ..\..\mudschool.exe .
cd ..\..
.\mudschool.exe doctor
```

Then use `.\mudschool.exe` instead of `./mudschool`.

---

## Get the course

```
git clone https://github.com/mattias/mud-school.git
cd mud-school
./mudschool doctor
./mudschool init --lang en
```

The first run of a Rust lesson downloads libraries (crates) and takes a
while. That only happens the first time.

## Troubleshooting

| Problem | Solution |
|---|---|
| `go: command not found` / `cargo: command not found` | Restart the terminal. Check that `~/go/bin`, `/usr/local/go/bin` and `~/.cargo/bin` are in `PATH`. |
| `cargo` cannot download crates behind a proxy | Set `HTTPS_PROXY` (and `CARGO_HTTP_CAINFO` if needed). |
| The terminal behaves strangely after the client crashed | Type `reset` and press Enter (the client left the terminal in "raw mode"). |
| `address already in use` | Something is already running on port 4000. Stop it, or run `./mudschool run server -addr :4001` and `./mudschool run client 127.0.0.1:4001`. |
| Strange characters or boxes on the map | The terminal or font lacks UTF-8 – see [The terminal](#6-the-terminal). |
| No colours or wrong colours | The terminal lacks truecolor (e.g. macOS Terminal.app). Switch terminals. |
| `-race requires cgo` | Install a C compiler (step 4), or run without `--race`. |
| The Windows firewall asks about network access | Allow it for private networks – the server listens on port 4000. |

Still not working? Run `./mudschool doctor` and read what is red.
