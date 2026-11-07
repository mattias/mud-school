package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

// cmdDoctor checks that the tools the course needs are installed.
func (c *course) cmdDoctor() error {
	problems := 0
	ok := func(name, detail string) { fmt.Printf("  %s %-10s %s\n", green("✔"), name, dim(detail)) }
	bad := func(name, detail, fix string) {
		problems++
		fmt.Printf("  %s %-10s %s\n", red("✘"), name, detail)
		fmt.Printf("    %s %s\n", yellow("→"), fix)
	}
	warn := func(name, detail, fix string) {
		fmt.Printf("  %s %-10s %s\n", yellow("!"), name, detail)
		fmt.Printf("    %s %s\n", dim("→"), dim(fix))
	}
	output := func(argv ...string) (string, bool) {
		out, err := exec.Command(argv[0], argv[1:]...).CombinedOutput()
		return strings.TrimSpace(string(out)), err == nil
	}

	fmt.Println(bold(L("Kontrollerar din miljö", "Checking your environment", "環境をチェックしています")))

	// Go >= 1.22
	if out, found := output("go", "version"); !found {
		bad("go", L("hittades inte", "not found", "見つかりません"), L(
			"Installera Go 1.22+ från https://go.dev/dl/",
			"Install Go 1.22+ from https://go.dev/dl/",
			"https://go.dev/dl/ から Go 1.22 以降をインストールしてください"))
	} else if m := regexp.MustCompile(`go1\.(\d+)`).FindStringSubmatch(out); m != nil {
		if minor, _ := strconv.Atoi(m[1]); minor < 22 {
			bad("go", out, L("Uppdatera till Go 1.22 eller senare", "Upgrade to Go 1.22 or later", "Go 1.22 以降に更新してください"))
		} else {
			ok("go", out)
		}
	} else {
		ok("go", out)
	}

	// Rust
	for _, tool := range []string{"rustc", "cargo"} {
		if out, found := output(tool, "--version"); found {
			ok(tool, out)
		} else {
			bad(tool, L("hittades inte", "not found", "見つかりません"), L(
				"Installera Rust med rustup: https://rustup.rs",
				"Install Rust with rustup: https://rustup.rs",
				"rustup で Rust をインストールしてください: https://rustup.rs"))
		}
	}

	// A way to connect by hand. On Windows ncat first: the Windows telnet
	// client can't show UTF-8.
	tools := []string{"telnet", "nc", "ncat"}
	if runtime.GOOS == "windows" {
		tools = []string{"ncat", "nc", "telnet"}
	}
	var client string
	for _, t := range tools {
		if _, err := exec.LookPath(t); err == nil {
			client = t
			break
		}
	}
	if client == "telnet" && runtime.GOOS == "windows" {
		warn("telnet/nc", "telnet", L(
			"Windows telnet visar å, ä, ö och japanska fel. Installera ncat i stället: `winget install Insecure.Nmap`, öppna en ny terminal, sedan `ncat 127.0.0.1 4000`.",
			"Windows telnet shows non-English characters wrong. Install ncat instead: `winget install Insecure.Nmap`, open a new terminal, then `ncat 127.0.0.1 4000`.",
			"Windows の telnet では日本語が文字化けします。代わりに ncat を入れてください: `winget install Insecure.Nmap`、新しい端末を開いてから `ncat 127.0.0.1 4000`。"))
	} else if client != "" {
		ok("telnet/nc", client)
	} else {
		warn("telnet/nc", L("hittades inte (frivilligt)", "not found (optional)", "見つかりません（任意）"), telnetHint())
	}

	// C compiler for -race
	if _, err := exec.LookPath("gcc"); err == nil {
		ok("gcc", L("finns (behövs för --race)", "found (needed for --race)", "あり（--race に必要）"))
	} else if _, err := exec.LookPath("clang"); err == nil {
		ok("clang", L("finns (behövs för --race)", "found (needed for --race)", "あり（--race に必要）"))
	} else {
		warn("gcc", L("hittades inte (frivilligt)", "not found (optional)", "見つかりません（任意）"), gccHint())
	}

	// Terminal
	t := detectTerminal()
	label := L("terminal", "terminal", "端末")
	switch {
	case t.name != "" && t.utf8 && t.truecolor:
		ok(label, t.name+L(" (UTF-8, 24-bitars färg)", " (UTF-8, 24-bit colour)", "（UTF-8、24 ビットカラー）"))
	case t.name != "" && !t.truecolor:
		warn(label, t.name+L(" har ofta saknat 24-bitars färg", " has often lacked 24-bit colour", " は 24 ビットカラーに対応していないことがあります"), L(
			"Ser kartans färger fel ut, byt terminal – se SETUP.md.",
			"If the map's colours look wrong, switch terminals – see SETUP.md.",
			"マップの色がおかしければ端末を変えてください。SETUP.md を参照。"))
	case runtime.GOOS == "windows":
		warn(label, L("inte Windows Terminal", "not Windows Terminal", "Windows Terminal ではありません"), L(
			"Den gamla Windows-konsolen visar varken färgerna eller tecknen rätt. Använd Windows Terminal (`winget install Microsoft.WindowsTerminal`).",
			"The old Windows console shows neither the colours nor the characters correctly. Use Windows Terminal (`winget install Microsoft.WindowsTerminal`).",
			"古い Windows コンソールでは色も文字も正しく表示されません。Windows Terminal を使ってください（`winget install Microsoft.WindowsTerminal`）。"))
	default:
		if t.utf8 {
			ok("UTF-8", t.locale)
		} else {
			warn("UTF-8", L("okänt", "unknown", "不明"), L(
				"Kartan använder tecken som ♣ ≈ ▲. Använd UTF-8 (t.ex. LANG=sv_SE.UTF-8)",
				"The map uses characters like ♣ ≈ ▲. Use UTF-8 (e.g. LANG=en_US.UTF-8)",
				"マップは ♣ ≈ ▲ などの文字を使います。UTF-8 を使ってください（例: LANG=ja_JP.UTF-8）"))
		}
		if t.truecolor {
			ok(L("färg", "colour", "色"), "truecolor")
		} else {
			warn(L("färg", "colour", "色"), L("okänt", "unknown", "不明"), L(
				"Kartan använder 24-bitars färg. De flesta moderna terminaler klarar det även om COLORTERM inte är satt.",
				"The map uses 24-bit colour. Most modern terminals support it even if COLORTERM is not set.",
				"マップは 24 ビットカラーを使います。COLORTERM が未設定でも、最近の端末ならたいてい対応しています。"))
		}
	}

	// The Windows console's code page: telnet/ncat print the server's UTF-8
	// bytes as they are, so the console must expect UTF-8.
	if cp, found := consoleCodePage(); found && cp != 65001 {
		warn(L("kodsida", "code page", "コードページ"), fmt.Sprintf(L("konsolen använder kodsida %d, inte UTF-8 (65001)", "the console uses code page %d, not UTF-8 (65001)", "コンソールのコードページが %d で、UTF-8（65001）ではありません"), cp), L(
			"ncat/telnet visar då å, ä, ö och japanska fel. Gör UTF-8 permanent i PowerShell – se docs/sv/SETUP.md, \"UTF-8 i varje ny terminal\".",
			"ncat/telnet will then show non-English characters wrong. Make UTF-8 permanent in PowerShell – see docs/en/SETUP.md, \"UTF-8 in every new terminal\".",
			"そのままでは ncat/telnet で日本語が文字化けします。PowerShell で UTF-8 を常に使うように設定してください。docs/ja/SETUP.md の「新しい端末でも毎回 UTF-8 に」を参照。"))
	}

	// Workspace
	if p, err := c.loadProgress(); err == nil {
		if n, names, err := c.describePending(p); err != nil {
			warn("workspace", err.Error(), L("Kör ./mudschool goto <id> eller ./mudschool init", "Run ./mudschool goto <id> or ./mudschool init", "./mudschool goto <id> か ./mudschool init を実行してください"))
		} else if n > 0 {
			warn("workspace", fmt.Sprintf(L("%d fil(er) kan uppdateras", "%d file(s) can be updated", "%d 個のファイルを更新できます"), n), L(
				"Kör ./mudschool update ("+names+")",
				"Run ./mudschool update ("+names+")",
				"./mudschool update を実行してください（"+names+"）"))
		} else {
			ok("workspace", L("uppdaterad", "up to date", "最新"))
		}
	}

	fmt.Println()
	if problems > 0 {
		return fmt.Errorf(L("%d problem – se docs/sv/SETUP.md", "%d problem(s) – see docs/en/SETUP.md", "%d 件の問題があります。docs/ja/SETUP.md を参照してください"), problems)
	}
	next := "./mudschool init"
	if exists(c.progressPath()) {
		next = "./mudschool show"
	}
	fmt.Println(green(L("Allt ser bra ut!", "Everything looks good!", "準備完了です！")), L("Nästa steg:", "Next step:", "次のステップ:"), cyan(next))
	return nil
}

type terminalInfo struct {
	name      string // "" if unknown
	locale    string
	utf8      bool
	truecolor bool
}

// detectTerminal recognises common terminals from their environment variables.
func detectTerminal() terminalInfo {
	t := terminalInfo{locale: firstNonEmpty(os.Getenv("LC_ALL"), os.Getenv("LC_CTYPE"), os.Getenv("LANG"))}
	up := strings.ToUpper(t.locale)
	t.utf8 = strings.Contains(up, "UTF-8") || strings.Contains(up, "UTF8")
	ct := os.Getenv("COLORTERM")
	t.truecolor = ct == "truecolor" || ct == "24bit"

	switch {
	case os.Getenv("WT_SESSION") != "":
		t.name, t.truecolor = "Windows Terminal", true
	case os.Getenv("TERM_PROGRAM") == "vscode":
		t.name, t.truecolor = "VS Code", true
	case os.Getenv("TERM_PROGRAM") == "mintty":
		t.name, t.truecolor = "mintty", true
	case os.Getenv("TERM_PROGRAM") == "Apple_Terminal":
		t.name, t.utf8, t.truecolor = "Terminal.app", true, false
	case os.Getenv("TERM_PROGRAM") == "iTerm.app":
		t.name, t.truecolor = "iTerm2", true
	case os.Getenv("TERM_PROGRAM") == "WezTerm":
		t.name, t.truecolor = "WezTerm", true
	case os.Getenv("TERM_PROGRAM") == "ghostty":
		t.name, t.truecolor = "Ghostty", true
	case os.Getenv("KITTY_WINDOW_ID") != "":
		t.name, t.truecolor = "kitty", true
	}
	if runtime.GOOS == "windows" && t.name != "" {
		t.utf8 = true // Go and Rust write Unicode to the Windows console directly
	}
	if t.name != "" && !t.utf8 && t.locale == "" && runtime.GOOS == "darwin" {
		t.utf8 = true // macOS terminals use UTF-8 even when LANG is unset
	}
	return t
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func telnetHint() string {
	switch runtime.GOOS {
	case "windows":
		return L(
			"Bra för att prova servern för hand: `winget install Insecure.Nmap` ger ncat (öppna en ny terminal, sedan `ncat 127.0.0.1 4000`). Windows egna telnet visar å, ä, ö fel.",
			"Handy for trying the server by hand: `winget install Insecure.Nmap` gives you ncat (open a new terminal, then `ncat 127.0.0.1 4000`). Windows' own telnet shows non-English characters wrong.",
			"サーバーを手で試すのに便利です: `winget install Insecure.Nmap` で ncat が入ります（新しい端末を開いてから `ncat 127.0.0.1 4000`）。Windows 標準の telnet は日本語が文字化けします。")
	case "darwin":
		return L("Bra för att prova servern för hand: `brew install telnet`", "Handy for trying the server by hand: `brew install telnet`", "サーバーを手で試すのに便利です: `brew install telnet`")
	default:
		return L("Bra för att prova servern för hand: t.ex. `sudo apt install telnet`", "Handy for trying the server by hand: e.g. `sudo apt install telnet`", "サーバーを手で試すのに便利です: 例 `sudo apt install telnet`")
	}
}

func gccHint() string {
	switch runtime.GOOS {
	case "windows":
		return L(
			"Behövs bara för `./mudschool check --race`. Installera MSYS2 (`winget install MSYS2.MSYS2`), kör `pacman -S --needed mingw-w64-ucrt-x86_64-gcc` i \"MSYS2 UCRT64\" och lägg till C:\\msys64\\ucrt64\\bin i PATH. Se SETUP.md.",
			"Only needed for `./mudschool check --race`. Install MSYS2 (`winget install MSYS2.MSYS2`), run `pacman -S --needed mingw-w64-ucrt-x86_64-gcc` in \"MSYS2 UCRT64\" and add C:\\msys64\\ucrt64\\bin to PATH. See SETUP.md.",
			"`./mudschool check --race` を使うときだけ必要です。MSYS2 を入れ（`winget install MSYS2.MSYS2`）、「MSYS2 UCRT64」で `pacman -S --needed mingw-w64-ucrt-x86_64-gcc` を実行し、C:\\msys64\\ucrt64\\bin を PATH に追加します。SETUP.md を参照。")
	case "darwin":
		return L("Behövs bara för `./mudschool check --race`: `xcode-select --install`", "Only needed for `./mudschool check --race`: `xcode-select --install`", "`./mudschool check --race` を使うときだけ必要です: `xcode-select --install`")
	default:
		return L("Behövs bara för `./mudschool check --race`: t.ex. `sudo apt install build-essential`", "Only needed for `./mudschool check --race`: e.g. `sudo apt install build-essential`", "`./mudschool check --race` を使うときだけ必要です: 例 `sudo apt install build-essential`")
	}
}
