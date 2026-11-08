# Installation och miljö

> [🇬🇧 English](../en/SETUP.md) · 🇸🇪 Svenska · [🇯🇵 日本語](../ja/SETUP.md)

Du behöver **Go 1.22+**, **Rust** (stable), **Git**, en **terminal** med UTF-8
och 24-bitars färg. (telnet eller nc är frivilliga – kursverktyget har en
inbyggd klient, `./mudschool connect`.) Välj avsnittet för ditt
system och följ det uppifrån och ner – varje avsnitt är komplett i sig.

| Ditt system | Avsnitt | Tid |
|---|---|---|
| Windows 10/11 | [Windows (PowerShell)](#windows-powershell) – rekommenderas på Windows | 20–30 min |
| Windows, om du föredrar Linux-verktyg | [Windows med WSL](#windows-med-wsl) | 20–30 min |
| macOS | [macOS](#macos) | 15 min |
| Linux | [Linux](#linux) | 15 min |
| Inget installerat, bara en webbläsare | [Codespaces / Dev Container](#github-codespaces-eller-en-dev-container) | 5 min |

Sedan: [Editor](#editor) · [Typsnitt och terminaler](#typsnitt-och-terminaler) · [Felsökning](#felsökning)

> **Om kommandona i lektionerna.** Lektionerna skriver `./mudschool show` osv.
> På Windows (PowerShell eller cmd) skriver du **`.\mudschool.cmd show`**
> istället. Kursverktyget skriver själv ut sina kommandon rätt för ditt system.

---

## Windows (PowerShell)

Allt nedan skrivs i **PowerShell** i **Windows Terminal**.

### 1. Windows Terminal

Windows 11 har den redan. På Windows 10:

```powershell
winget install Microsoft.WindowsTerminal
```

Använd Windows Terminal genom hela kursen – det gamla konsolfönstret
(`conhost`) visar varken färgerna eller karttecknen rätt. Välj typsnitt direkt:
se [Typsnitt och terminaler](#typsnitt-och-terminaler).

### 2. Git, Go och Rust

```powershell
winget install Git.Git
winget install GoLang.Go
winget install Rustlang.Rustup
```

Rust på Windows behöver också Microsofts C++-byggverktyg (länkaren):

```powershell
winget install Microsoft.VisualStudio.2022.BuildTools --override "--wait --passive --add Microsoft.VisualStudio.Workload.VCTools --includeRecommended"
```

**Stäng och öppna Windows Terminal igen** så att `PATH` uppdateras, och kontrollera:

```powershell
go version        # go1.22 eller senare
rustc --version
cargo --version
```

> **Satte Microsoft Defender `cargo.exe` i karantän?** Det händer ibland:
> Defenders heuristik flaggar de små "proxy"-programmen som rustup lägger i
> `%USERPROFILE%\.cargo\bin` (`cargo.exe`, `rustc.exe`, … är alla kopior av
> `rustup.exe`). Klicka inte bara på *Tillåt* – kontrollera först:
>
> 1. Uppdatera Defender (*Windows-säkerhet → Virus- och hotskydd →
>    Skyddsuppdateringar → Sök efter uppdateringar*).
> 2. Installera om Rust från den officiella källan: `winget uninstall Rustlang.Rustup`,
>    sedan `winget install Rustlang.Rustup` (eller `rustup-init.exe` från
>    <https://rustup.rs>).
> 3. Hamnar den nya filen i karantän igen, kontrollera att rustup är äkta.
>    `cargo.exe` är borta, men den är en hård länk till `rustup.exe`, som är en
>    kopia av den officiella `rustup-init.exe`. Kör `rustup self update` och
>    jämför sedan `(Get-FileHash $env:USERPROFILE\.cargo\bin\rustup.exe).Hash`
>    med första ordet i <https://static.rust-lang.org/rustup/dist/x86_64-pc-windows-msvc/rustup-init.exe.sha256> (stora/små bokstäver spelar ingen roll; på ARM
>    står det `aarch64-pc-windows-msvc` i adressen).
> 4. Stämmer de är det ett falsklarm: rapportera filen till Microsoft på
>    <https://www.microsoft.com/wdsi/filesubmission>, och välj sedan objektet i
>    *Skyddshistorik* → *Åtgärder* → *Tillåt på enhet*. En vanlig *Återställ*
>    kan hamna i karantän igen tills Microsoft har uppdaterat sina definitioner.
>    Stämmer de inte – tillåt ingenting; be om hjälp.
>
> Program du bygger själv (`target\debug\*.exe`) kan flaggas på samma sätt;
> rapportera dem som i steg 4.

### 3. Ansluta för hand – inget att installera

I lektion 01–02, innan din egen klient finns, provar du servern för hand.
Kursverktyget har en liten inbyggd klient:

```powershell
.\mudschool.cmd connect          # = 127.0.0.1:4000
```

Den visar å, ä, ö och japanska rätt i alla Windows-terminaler, utan några
inställningar. Lektionerna använder `./mudschool connect`; där de nämner
telnet eller nc fungerar den här likadant.

#### Om du ändå vill använda telnet eller ncat

Windows egna telnet förstår inte UTF-8 alls. ncat (`winget install
Insecure.Nmap`, öppna sedan terminalen igen) fungerar om konsolen väntar sig
UTF-8 – vilket PowerShell inte gör från början (en japansk Windows använder
Shift-JIS, kodsida 932). Enklaste sättet att fixa det för varje ny flik, utan
att röra några filer: i *Windows Terminal → Inställningar → (din
PowerShell-profil) → Kommandorad*, använd

```
pwsh.exe -NoLogo -NoExit -Command "[Console]::InputEncoding = [Console]::OutputEncoding = $OutputEncoding = [System.Text.UTF8Encoding]::new()"
```

(Använd `powershell.exe` i stället för `pwsh.exe` för Windows PowerShell 5.1.)
(Redigerar du `settings.json` för hand i stället skriver du varje `"` i
kommandot som `\"`.) Kontrollera i en ny flik med `chcp` – det ska stå `65001`.

PowerShell-profilen (`$PROFILE`) skulle också fungera, men den ligger i
*Dokument*, som Defenders **reglerad mappåtkomst** ofta skyddar ("Åtkomst till
skyddad mapp blockerad … pwsh.exe"), och den kräver `Set-ExecutionPolicy`.
Inställningen i Windows Terminal undviker båda.

### 4. En C-kompilator (frivilligt)

Behövs bara för `.\mudschool.cmd check --race` (Gos race detector använder cgo).
Enklast är GCC från MSYS2:

```powershell
winget install MSYS2.MSYS2
```

Starta **"MSYS2 UCRT64"** från Start-menyn och kör:

```sh
pacman -S --needed mingw-w64-ucrt-x86_64-gcc
```

Lägg till `C:\msys64\ucrt64\bin` i `PATH` (Start → "Redigera miljövariabler
för ditt konto" → `Path` → Ny), öppna Windows Terminal igen och kontrollera
med `gcc --version`.

### 5. Hämta kursen och börja

```powershell
cd $HOME
git clone https://github.com/mattias/mud-school.git
cd mud-school
.\mudschool.cmd doctor            # kontrollerar Go, Rust, gcc och terminalen
.\mudschool.cmd init --lang sv    # skapar workspace\ (en, sv eller ja)
.\mudschool.cmd show              # första lektionen
```

`mudschool.cmd` bygger kursverktyget automatiskt första gången. Första gången
en Rust-lektion körs laddar Cargo ner bibliotek, och det tar en stund. Servern
lyssnar bara på `127.0.0.1` (din egen dator), så brandväggen behöver inte
fråga. Vill du att andra i ditt nätverk ska kunna ansluta startar du den med
`.\mudschool.cmd run server -addr :4000`; då frågar Windows Defender-brandväggen
en gång – tillåt för privata nätverk.

---

## Windows med WSL

WSL ger dig ett riktigt Linux inuti Windows. Välj det om du föredrar
Linux-verktyg eller redan använder WSL.

```powershell
wsl --install -d Ubuntu
```

Starta om datorn, öppna profilen **Ubuntu** i Windows Terminal och följ sedan
avsnittet [Linux](#linux) inne i Ubuntu. Klona kursen i din Linux-hemkatalog
(`~/mud-school`), inte under `/mnt/c/` – det går mycket snabbare. VS Code-tillägget
*WSL* låter dig redigera filerna från Windows.

---

## macOS

### 1. Verktyg

Med [Homebrew](https://brew.sh):

```sh
brew install go git telnet
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
xcode-select --install            # C-kompilator, behövs för --race
```

Öppna terminalen igen och kontrollera `go version`, `rustc --version` och `cargo --version`.

### 2. En terminal med 24-bitars färg

**Terminal.app har länge saknat truecolor** (åtminstone före macOS 26) – kartans färger kan bli fel. Använd **iTerm2**,
**WezTerm**, **Ghostty** eller **kitty** (t.ex. `brew install --cask iterm2`).

### 3. Hämta kursen och börja

```sh
git clone https://github.com/mattias/mud-school.git
cd mud-school
./mudschool doctor
./mudschool init --lang sv
./mudschool show
```

---

## Linux

### 1. Verktyg

Go: distributionens paket är ofta för gammalt. Installera senaste versionen
från <https://go.dev/dl/> enligt instruktionerna där, eller med snap:
`sudo snap install go --classic`.

Rust:

```sh
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
```

Resten (Ubuntu/Debian; på Fedora `sudo dnf install git telnet nmap-ncat gcc`):

```sh
sudo apt install git telnet netcat-openbsd build-essential
```

För den grafiska klienten i lektion 28 (ett fönster med tiles) behöver du också
utvecklingsfilerna för X11 och OpenGL: `sudo apt install libx11-dev libxi-dev libgl1-mesa-dev`.

Öppna terminalen igen och kontrollera `go version`, `rustc --version` och `cargo --version`.

### 2. Hämta kursen och börja

```sh
git clone https://github.com/mattias/mud-school.git
cd mud-school
./mudschool doctor
./mudschool init --lang sv
./mudschool show
```

GNOME Terminal, Konsole, kitty, Alacritty och WezTerm klarar alla UTF-8 och
24-bitars färg. Se till att språkinställningen är UTF-8 (t.ex. `LANG=sv_SE.UTF-8`).

---

## GitHub Codespaces eller en Dev Container

Repot har en färdig miljö i `.devcontainer/` (Ubuntu med Go 1.24, Rust,
telnet och nc, plus VS Code-tilläggen för Go och Rust).

- **GitHub Codespaces:** på repots sida, klicka **Code → Codespaces → Create
  codespace**. Allt körs i webbläsaren.
- **VS Code lokalt:** installera [Docker](https://www.docker.com/) och tillägget
  *Dev Containers*, öppna mappen och välj **Reopen in Container**.

Följ sedan [Linux, steg 2](#2-hämta-kursen-och-börja) (kursen är redan klonad).
Port 4000 vidarebefordras automatiskt.

---

## Editor

Rekommendation: **VS Code** med tilläggen **Go** (`golang.go`) och
**rust-analyzer** (`rust-lang.rust-analyzer`): autokomplettering, fel direkt i
koden och "gå till definition". JetBrains GoLand/RustRover och Neovim med LSP
fungerar också utmärkt. Öppna hela mappen `mud-school` (eller `workspace/server`
och `workspace/client` var för sig) så att verktygen hittar `go.mod` och `Cargo.toml`.

## Typsnitt och terminaler

Klienten ritar kartan med tecken som `♣ ≈ ▲ ■ Ω ─ │ ♥` i 24-bitars färg.

| Kursens språk | Rekommenderade typsnitt |
|---|---|
| Engelska, svenska | **Cascadia Mono** (standard i Windows Terminal), JetBrains Mono, Fira Code, DejaVu Sans Mono |
| Japanska | **UDEV Gothic** eller **HackGen** (fria programmeringstypsnitt med både latinska och japanska tecken); inbyggt i Windows: **BIZ UDGothic** |

**Windows Terminal:** Inställningar (`Ctrl+,`) → *Profiler* → *Standardvärden* →
*Utseende* → *Teckensnitt*. Installera ett nedladdat typsnitt först
(högerklicka på `.ttf`-filen → *Installera*) och starta om Windows Terminal. BIZ UDGothic finns förinstallerat i japansk Windows; i övriga lägger du till det under Inställningar → System → Valfria funktioner → *Japanese Supplemental Fonts*.
Saknar ett typsnitt något tecken lånar Windows Terminal det automatiskt från
ett annat typsnitt.

**Japansk text och kartan.** Tecken som `≈ ▲ ■ ─ │ ♣` har "tvetydig östasiatisk
bredd". Kartan förutsätter att de är **en ruta breda**:

- Windows Terminal ritar dem alltid en ruta breda – inget att ändra.
- iTerm2: Profiles → Text → stäng av *Ambiguous characters are double-width*.
- WezTerm: `treat_east_asian_ambiguous_width_as_wide = false` (standard).
- Välj den vanliga varianten av UDEV Gothic/HackGen, inte varianterna som gör
  tvetydiga tecken helbreda (t.ex. "UDEV Gothic JPDOC").

Japanska rumsnamn och beskrivningar är själva två rutor breda, och det
hanterar klienten rätt.

## Felsökning

| Problem | Lösning |
|---|---|
| `go`/`cargo` "känns inte igen" / "command not found" | Öppna terminalen igen. Kontrollera att Gos `bin`-katalog och `~/.cargo/bin` (Windows: `%USERPROFILE%\.cargo\bin`) finns i `PATH`. |
| Windows: `link.exe not found` när Rust bygger | Installera C++-byggverktygen (Windows, steg 2). |
| `cargo` kan inte ladda ner crates bakom en proxy | Sätt `HTTPS_PROXY` (och vid behov `CARGO_HTTP_CAINFO`). |
| Terminalen beter sig konstigt efter att klienten kraschat | Skriv `reset` (Windows: stäng fliken) – klienten lämnade terminalen i "raw mode". |
| `address already in use` | Något kör redan på port 4000. Stoppa det, eller kör `./mudschool run server -addr 127.0.0.1:4001` och `./mudschool run client 127.0.0.1:4001`. |
| Rutor eller konstiga tecken på kartan | Typsnittet saknar tecknen eller terminalen är inte UTF-8 – se [Typsnitt och terminaler](#typsnitt-och-terminaler). |
| Windows: telnet/ncat visar teckengröt (t.ex. å, ä, ö) | Använd `.\mudschool.cmd connect` i stället – se [Ansluta för hand](#3-ansluta-för-hand--inget-att-installera). |
| Windows: Defender satte `cargo.exe` i karantän | Se rutan under [Git, Go och Rust](#2-git-go-och-rust). |
| Kartans kolumner hamnar fel (japanska) | Tecken med tvetydig bredd ritas dubbelt så breda – se [Typsnitt och terminaler](#typsnitt-och-terminaler). |
| Inga eller fel färger | Terminalen saknar 24-bitars färg (t.ex. macOS Terminal.app, den gamla Windows-konsolen). |
| `-race requires cgo` | Installera en C-kompilator (se avsnittet för ditt system) eller kör utan `--race`. |
| Windows: Ctrl+C frågar "Terminate batch job (Y/N)?" | Det är cmd.exe som frågar eftersom `mudschool.cmd` är en batchfil. Svara `Y`. |
| `doctor` varnar för UTF-8/färg fast allt ser bra ut | Terminalen kändes inte igen. Ser kartan rätt ut kan du strunta i det. |

Fungerar det fortfarande inte? Kör `doctor` och läs det som är rött eller gult.
