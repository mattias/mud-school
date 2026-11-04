package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
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

	// A way to connect by hand
	var client string
	for _, t := range []string{"telnet", "nc", "ncat"} {
		if _, err := exec.LookPath(t); err == nil {
			client = t
			break
		}
	}
	if client != "" {
		ok("telnet/nc", client)
	} else {
		warn("telnet/nc", L("hittades inte (frivilligt)", "not found (optional)", "見つかりません（任意）"), L(
			"Bra för att prova servern för hand: t.ex. `sudo apt install telnet` eller `brew install telnet`",
			"Handy for trying the server by hand: e.g. `sudo apt install telnet` or `brew install telnet`",
			"サーバーを手で試すのに便利です: 例 `sudo apt install telnet` または `brew install telnet`"))
	}

	// C compiler for -race
	if _, err := exec.LookPath("gcc"); err == nil {
		ok("gcc", L("finns (behövs för --race)", "found (needed for --race)", "あり（--race に必要）"))
	} else if _, err := exec.LookPath("clang"); err == nil {
		ok("clang", L("finns (behövs för --race)", "found (needed for --race)", "あり（--race に必要）"))
	} else {
		warn("gcc", L("hittades inte (frivilligt)", "not found (optional)", "見つかりません（任意）"), L(
			"Behövs bara för `./mudschool check --race`",
			"Only needed for `./mudschool check --race`",
			"`./mudschool check --race` を使うときだけ必要です"))
	}

	// Terminal
	loc := strings.ToUpper(os.Getenv("LC_ALL") + os.Getenv("LC_CTYPE") + os.Getenv("LANG"))
	if strings.Contains(loc, "UTF-8") || strings.Contains(loc, "UTF8") {
		ok("UTF-8", loc)
	} else {
		warn("UTF-8", L("okänt", "unknown", "不明"), L(
			"Kartan använder tecken som ♣ ≈ ▲. Använd en terminal med UTF-8 (t.ex. LANG=sv_SE.UTF-8)",
			"The map uses characters like ♣ ≈ ▲. Use a UTF-8 terminal (e.g. LANG=en_US.UTF-8)",
			"マップは ♣ ≈ ▲ などの文字を使います。UTF-8 の端末を使ってください（例: LANG=ja_JP.UTF-8）"))
	}
	if ct := os.Getenv("COLORTERM"); ct == "truecolor" || ct == "24bit" {
		ok(L("färg", "color", "色"), "truecolor")
	} else {
		warn(L("färg", "color", "色"), "COLORTERM="+ct, L(
			"Kartan använder 24-bitars färg. De flesta moderna terminaler klarar det även om COLORTERM inte är satt.",
			"The map uses 24-bit color. Most modern terminals support it even if COLORTERM is not set.",
			"マップは 24 ビットカラーを使います。COLORTERM が未設定でも、最近の端末ならたいてい対応しています。"))
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
