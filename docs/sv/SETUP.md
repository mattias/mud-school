# Installation och miljö

Du behöver **Go**, **Rust**, en **terminal** som klarar UTF-8 och färg, och
gärna **telnet** eller **nc**. Här finns tre vägar – välj en.

| Väg | Passar dig som ... | Tid |
|---|---|---|
| [A. Dev Container / Codespaces](#a-dev-container-eller-github-codespaces) | vill komma igång direkt utan att installera något på datorn | 5 min |
| [B. Installera själv](#b-installera-själv) | vill ha allt lokalt (rekommenderas på sikt) | 15–30 min |
| [C. Windows med WSL](#c-windows) | kör Windows | 20–30 min |

När du är klar, kör alltid:

```
./mudschool doctor             # kontrollerar Go, Rust, telnet/nc, gcc, UTF-8, färg
./mudschool init --lang sv     # skapar workspace/ på svenska (en = engelska, ja = japanska)
./mudschool show               # första lektionen
```

---

## A. Dev Container eller GitHub Codespaces

Repot har en färdig miljö i `.devcontainer/devcontainer.json` (Ubuntu med Go
1.24, Rust stable, telnet och nc, plus VS Code-tilläggen för Go och Rust).

- **GitHub Codespaces:** på repots sida på GitHub, klicka **Code → Codespaces →
  Create codespace**. Allt startar i webbläsaren.
- **VS Code lokalt:** installera [Docker](https://www.docker.com/) och tillägget
  *Dev Containers*, öppna mappen och välj **Reopen in Container**.

Port 4000 vidarebefordras automatiskt, så du kan ansluta till servern från
din egen dator också.

## B. Installera själv

### 1. Go (1.22 eller senare)

| System | Kommando |
|---|---|
| macOS | `brew install go` – eller installationspaketet från <https://go.dev/dl/> |
| Linux | ladda ner från <https://go.dev/dl/> och följ instruktionerna (distributionens paket kan vara för gamla – kontrollera med `go version`) |
| Windows | `winget install GoLang.Go` – eller MSI-filen från <https://go.dev/dl/> |

Kontrollera: `go version` ska visa `go1.22` eller senare.

### 2. Rust (stable)

Installera med **rustup**, det officiella installationsverktyget:

```
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
```

På Windows: kör `rustup-init.exe` från <https://rustup.rs> (eller
`winget install Rustlang.Rustup`). Rust behöver då *Visual Studio Build Tools*
med "Desktop development with C++" – installationsprogrammet frågar.

Kontrollera: `rustc --version` och `cargo --version`.

### 3. telnet och nc (frivilligt men rekommenderat)

| System | Kommando |
|---|---|
| macOS | `brew install telnet` (`nc` finns redan) |
| Ubuntu/Debian | `sudo apt install telnet netcat-openbsd` |
| Fedora | `sudo dnf install telnet nmap-ncat` |
| Windows | aktivera *Telnet Client* under "Turn Windows features on or off", eller använd WSL |

### 4. En C-kompilator (frivilligt)

Behövs bara för `./mudschool check --race` (Gos race detector använder cgo).

| System | Kommando |
|---|---|
| macOS | `xcode-select --install` |
| Ubuntu/Debian | `sudo apt install build-essential` |

### 5. Editor

Rekommendation: **VS Code** med tilläggen **Go** (`golang.go`) och
**rust-analyzer** (`rust-lang.rust-analyzer`). De ger autokomplettering,
felmarkering direkt i koden och "gå till definition". JetBrains GoLand/RustRover
och Neovim med LSP fungerar också utmärkt.

Öppna gärna `workspace/server` och `workspace/client` som separata fönster
(eller hela repot) så att verktygen hittar `go.mod` och `Cargo.toml`.

### 6. Terminalen

Klienten ritar kartan med tecken som `♣ ≈ ▲ ■ ─ │` i 24-bitars färg.

- **macOS:** använd **iTerm2**, **WezTerm**, **Ghostty** eller **kitty**.
  *Terminal.app saknar truecolor* – färgerna blir fel.
- **Windows:** **Windows Terminal** (förinstallerat i Windows 11).
- **Linux:** GNOME Terminal, Konsole, kitty, Alacritty, WezTerm – alla fungerar.
- **Typsnitt:** ett med bra Unicode-täckning, t.ex. JetBrains Mono, Cascadia
  Code, Fira Code eller DejaVu Sans Mono.
- **Språkinställning:** UTF-8, t.ex. `export LANG=sv_SE.UTF-8`.

## C. Windows

Det enklaste är **WSL 2** (Windows Subsystem for Linux):

```
wsl --install -d Ubuntu
```

Starta om, öppna *Ubuntu* i Windows Terminal och följ sedan [B](#b-installera-själv)
med Linux-kommandona. VS Code har tillägget *WSL* så att du kan redigera filerna
från Windows.

Vill du köra direkt i Windows (PowerShell) fungerar det också, men skriptet
`./mudschool` är ett sh-skript. Bygg kursverktyget en gång:

```
cd tools\mudschool
go build -o ..\..\mudschool.exe .
cd ..\..
.\mudschool.exe doctor
```

Använd sedan `.\mudschool.exe` istället för `./mudschool`.

---

## Hämta kursen

```
git clone https://github.com/mattias/mud-school.git
cd mud-school
./mudschool doctor
./mudschool init --lang sv
```

Första körningen av en Rust-lektion laddar ner bibliotek (crates) och tar en
stund. Det är bara första gången.

## Felsökning

| Problem | Lösning |
|---|---|
| `go: command not found` / `cargo: command not found` | Starta om terminalen. Kontrollera att `~/go/bin`, `/usr/local/go/bin` och `~/.cargo/bin` finns i `PATH`. |
| `cargo` kan inte ladda ner crates bakom en proxy | Sätt `HTTPS_PROXY` (och vid behov `CARGO_HTTP_CAINFO`). |
| Terminalen beter sig konstigt efter att klienten kraschat | Skriv `reset` och tryck Enter (klienten lämnade terminalen i "raw mode"). |
| `address already in use` | Något kör redan på port 4000. Stoppa det, eller kör `./mudschool run server -addr :4001` och `./mudschool run client 127.0.0.1:4001`. |
| Konstiga tecken eller rutor på kartan | Terminalen eller typsnittet saknar UTF-8 – se [Terminalen](#6-terminalen). |
| Inga eller fel färger | Terminalen saknar truecolor (t.ex. macOS Terminal.app). Byt terminal. |
| `-race requires cgo` | Installera en C-kompilator (steg 4), eller kör utan `--race`. |
| Windows-brandväggen frågar om nätverksåtkomst | Tillåt för privata nätverk – servern lyssnar på port 4000. |

Fungerar det fortfarande inte? Kör `./mudschool doctor` och läs vad som är rött.
