package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func (c *course) cmdInit() error {
	ws := c.workspace()
	if exists(c.progressPath()) {
		fmt.Println(L("Du har redan en arbetsyta i", "You already have a workspace in", "作業領域はすでにあります:"), ws)
		fmt.Println(L("Kör", "Run", "実行:"), cyan("./mudschool show"))
		return nil
	}
	written, err := overlay(c.root, "starter", ws, false)
	if err != nil {
		return err
	}
	first := c.lessons[0]
	p := &Progress{Lang: lang, Current: first.ID, Done: map[string]bool{}, Hints: map[string]int{}}
	c.record(p, written, "starter", "")
	if _, err := c.enterWorkspace(p, first); err != nil {
		return err
	}
	if err := c.saveProgress(p); err != nil {
		return err
	}
	fmt.Println(green(L("✔ Arbetsytan är skapad:", "✔ Workspace created:", "✔ 作業領域を作成しました:")), ws)
	fmt.Println()
	fmt.Println(L("  workspace/server/   din MUD-server i Go", "  workspace/server/   your MUD server in Go", "  workspace/server/   Go で書く MUD サーバー"))
	fmt.Println(L("  workspace/client/   din MUD-klient i Rust", "  workspace/client/   your MUD client in Rust", "  workspace/client/   Rust で書く MUD クライアント"))
	fmt.Println()
	fmt.Println(L("Språk:", "Language:", "言語:"), lang, dim(L("(byt med ./mudschool lang en|sv|ja)", "(change with ./mudschool lang en|sv|ja)", "（./mudschool lang en|sv|ja で変更）")))
	fmt.Println(L("Nästa steg:", "Next step:", "次のステップ:"), cyan("./mudschool show"))
	return nil
}

func (c *course) cmdLang(args []string) error {
	if len(args) != 1 || !validLang(args[0]) {
		return errors.New(L("användning: mudschool lang en|sv|ja", "usage: mudschool lang en|sv|ja", "使い方: mudschool lang en|sv|ja"))
	}
	p, err := c.loadProgress()
	if err != nil {
		return err
	}
	lang = args[0]
	p.Lang = lang

	// The game texts (messages.go / messages.rs) follow the course language.
	// If the student has edited one of them, keep a backup before replacing it.
	src := filepath.Join(c.root, "starter."+lang)
	err = filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		relPath, _ := filepath.Rel(src, path)
		target := filepath.Join(c.workspace(), relPath)
		shown := filepath.ToSlash(relPath)
		if current, err := os.ReadFile(target); err == nil && !c.isStarterVersion(relPath, current) {
			backup := freeBackupName(target)
			if err := os.WriteFile(backup, current, 0o644); err != nil {
				return err
			}
			shownBackup := filepath.ToSlash(filepath.Join(filepath.Dir(relPath), filepath.Base(backup)))
			fmt.Println(yellow(fmt.Sprintf(L(
				"  du hade ändrat workspace/%s – din version är sparad som workspace/%s",
				"  you had edited workspace/%s – your version is saved as workspace/%s",
				"  workspace/%s は変更されていたので、あなたの版を workspace/%s として保存しました"), shown, shownBackup)))
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(target, data, 0o644); err != nil {
			return err
		}
		fmt.Println(dim(L("  skrev workspace/", "  wrote workspace/", "  書き込み: workspace/") + shown))
		c.record(p, []string{relPath}, "starter", "")
		return nil
	})
	if err != nil {
		return err
	}
	fmt.Println(green("✔"), L("Kursen är nu på svenska.", "The course is now in English.", "コースを日本語に切り替えました。"))
	fmt.Println(dim(L(
		"Spelvärlden du redan fått (defaultworld.go i lektion 08, world.json från lektion 09) byts inte ut – ändra den själv om du vill.",
		"The game world you already have (defaultworld.go in lesson 08, world.json from lesson 09) is not replaced – edit it yourself if you like.",
		"すでに受け取った世界データ（レッスン 08 の defaultworld.go、レッスン 09 以降の world.json）は置き換えません。必要なら自分で編集してください。")))
	return c.saveProgress(p)
}

// freeBackupName returns path+".bak", or path+".bak.1", ".bak.2", ... if
// earlier backups exist, so that no backup is ever overwritten.
func freeBackupName(path string) string {
	name := path + ".bak"
	for i := 1; exists(name); i++ {
		name = fmt.Sprintf("%s.bak.%d", path, i)
	}
	return name
}

// isStarterVersion reports whether data is identical to relPath in any of the
// starter.<lang>/ trees, i.e. the student has not edited it.
func (c *course) isStarterVersion(relPath string, data []byte) bool {
	for _, l := range Languages {
		orig, err := os.ReadFile(filepath.Join(c.root, "starter."+l, relPath))
		if err == nil && string(orig) == string(data) {
			return true
		}
	}
	return false
}

func (c *course) cmdList() error {
	p, err := c.loadProgress()
	if err != nil {
		p = &Progress{Done: map[string]bool{}}
	}
	fmt.Println(bold(L("Lektioner", "Lessons", "レッスン")))
	for _, l := range c.lessons {
		mark := dim("·")
		switch {
		case p.Done[l.ID]:
			mark = green("✔")
		case l.ID == p.Current:
			mark = yellow("▶")
		}
		track := cyan("Go  ")
		if l.Track == "rust" {
			track = yellow("Rust")
		}
		fmt.Printf("  %s %s  %s  %s\n", mark, l.ID, track, l.Title)
		if l.ID == p.Current {
			fmt.Printf("           %s\n", dim(l.Summary.String()))
		}
	}
	return nil
}

func (c *course) cmdShow(args []string) error {
	var l *Lesson
	var err error
	if len(args) > 0 {
		l, err = c.lesson(args[0])
	} else {
		var p *Progress
		if p, err = c.loadProgress(); err == nil {
			l, err = c.current(p)
		}
	}
	if err != nil {
		return err
	}
	data, err := os.ReadFile(localized(l.Dir, "README"))
	if err != nil {
		return err
	}
	fmt.Print(renderMarkdown(string(data)))
	fmt.Println(dim("─────────────────────────────────────────────"))
	fmt.Println(L(
		dim("Kör ")+cyan("./mudschool check")+dim(" när du är klar, eller ")+cyan("./mudschool hint")+dim(" om du kör fast."),
		dim("Run ")+cyan("./mudschool check")+dim(" when you are done, or ")+cyan("./mudschool hint")+dim(" if you get stuck."),
		dim("終わったら ")+cyan("./mudschool check")+dim("、行き詰まったら ")+cyan("./mudschool hint")+dim(" を実行しましょう。")))
	return nil
}

func (c *course) cmdHint() error {
	p, err := c.loadProgress()
	if err != nil {
		return err
	}
	l, err := c.current(p)
	if err != nil {
		return err
	}
	data, err := os.ReadFile(localized(l.Dir, "HINTS"))
	if errors.Is(err, fs.ErrNotExist) {
		fmt.Println(L("Den här lektionen har inga ledtrådar.", "This lesson has no hints.", "このレッスンにはヒントがありません。"))
		return nil
	} else if err != nil {
		return err
	}
	hints := splitHints(string(data))
	n := p.Hints[l.ID]
	if n >= len(hints) {
		fmt.Println(L("Du har sett alla ledtrådar. Fortfarande fast? Prova",
			"You have seen all the hints. Still stuck? Try",
			"ヒントはこれで全部です。まだ解けなければ次を試してください:"), cyan("./mudschool solution"))
		n = len(hints) - 1
	} else {
		p.Hints[l.ID] = n + 1
		if err := c.saveProgress(p); err != nil {
			return err
		}
	}
	fmt.Printf(dim(L("Ledtråd %d av %d\n\n", "Hint %d of %d\n\n", "ヒント %d / %d\n\n")), n+1, len(hints))
	fmt.Print(renderMarkdown(hints[n]))
	return nil
}

// splitHints splits HINTS.md at every "## " heading.
func splitHints(md string) []string {
	var hints []string
	for _, part := range strings.Split("\n"+md, "\n## ") {
		part = strings.TrimSpace(part)
		if part == "" || strings.HasPrefix(part, "# ") {
			continue
		}
		hints = append(hints, "## "+part+"\n")
	}
	return hints
}

func (c *course) cmdGoto(args []string) error {
	if len(args) != 1 {
		return errors.New(L("användning: mudschool goto <id>", "usage: mudschool goto <id>", "使い方: mudschool goto <id>"))
	}
	p, err := c.loadProgress()
	if err != nil {
		return err
	}
	l, err := c.lesson(args[0])
	if err != nil {
		return err
	}
	p.Current = l.ID
	if _, err := c.enterWorkspace(p, l); err != nil {
		return err
	}
	if err := c.saveProgress(p); err != nil {
		return err
	}
	fmt.Printf("%s %s %s\n", L("Aktuell lektion:", "Current lesson:", "現在のレッスン:"), l.ID, bold(l.Title.String()))
	return nil
}

// checkCommands returns the commands that test a track.
func checkCommands(track string, race bool) (dir string, cmds [][]string) {
	if track == "go" {
		test := []string{"go", "test", "-count=1"}
		if race {
			test = append(test, "-race")
		}
		return "server", [][]string{{"go", "vet", "./..."}, append(test, "./...")}
	}
	return "client", [][]string{{"cargo", "test", "--quiet"}}
}

func runIn(dir string, quiet bool, env []string, argv ...string) (string, error) {
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Dir = dir
	cmd.Env = append(append(os.Environ(), "MUDSCHOOL_LANG="+lang), env...)
	if quiet {
		out, err := cmd.CombinedOutput()
		return string(out), err
	}
	cmd.Stdout, cmd.Stderr, cmd.Stdin = os.Stdout, os.Stderr, os.Stdin
	return "", cmd.Run()
}

func (c *course) cmdCheck(race, advance bool) (bool, error) {
	p, err := c.loadProgress()
	if err != nil {
		return false, err
	}
	l, err := c.current(p)
	if err != nil {
		return false, err
	}
	ws := c.workspace()
	if _, err := c.enterWorkspace(p, l); err != nil {
		return false, err
	}
	// Bring in course fixes the student hasn't touched (git pull + check just works).
	plan, err := c.planUpdate(p)
	if err != nil {
		return false, err
	}
	if pending, _, err := c.applyUpdate(p, plan, true); err != nil {
		return false, err
	} else if pending > 0 {
		fmt.Println(yellow(fmt.Sprintf(L(
			"! %d fil(er) i arbetsytan har nyare versioner i kursen – kör ./mudschool update",
			"! %d file(s) in your workspace have newer versions in the course – run ./mudschool update",
			"! 作業領域の %d 個のファイルにコースの新しい版があります。./mudschool update を実行してください"), pending)))
	}
	if err := c.saveProgress(p); err != nil {
		return false, err
	}
	if _, err := c.placeTests(ws, l); err != nil {
		return false, err
	}
	fmt.Printf("%s %s %s: %s\n\n", cyan("▶"), L("Lektion", "Lesson", "レッスン"), l.ID, bold(l.Title.String()))
	dir, cmds := checkCommands(l.Track, race)
	for _, argv := range cmds {
		fmt.Println(dim("$ " + strings.Join(argv, " ")))
		if _, err := runIn(filepath.Join(ws, dir), false, nil, argv...); err != nil {
			fmt.Println()
			fmt.Println(red(L("✘ Inte riktigt än.", "✘ Not quite yet.", "✘ もう少しです。")),
				L("Läs felmeddelandena ovan – de berättar vad testerna förväntade sig.",
					"Read the messages above – they tell you what the tests expected.",
					"上のメッセージを読んでください。テストが何を期待していたかが書かれています。"))
			fmt.Println(L("  Instruktioner:", "  Instructions:", "  説明:"), cyan("./mudschool show"),
				L("  Ledtråd:", "  Hint:", "  ヒント:"), cyan("./mudschool hint"))
			return false, nil
		}
	}
	fmt.Println()
	fmt.Println(green(L("✔ Alla tester gick igenom!", "✔ All tests passed!", "✔ すべてのテストに合格しました！")),
		fmt.Sprintf(L("Lektion %s är klar.", "Lesson %s is done.", "レッスン %s 完了です。"), l.ID))
	if !advance {
		return true, nil
	}
	if todos := leftoverTODOs(ws, l.ID); len(todos) > 0 {
		fmt.Println(dim(fmt.Sprintf(L(
			"Tips: kommentarerna TODO(lesson %s) är avklarade nu – du kan ta bort dem: %s",
			"Tip: the TODO(lesson %s) comments are done now – you can delete them: %s",
			"ヒント: TODO(lesson %s) のコメントはもう済んでいます。消してかまいません: %s"),
			strings.TrimLeft(l.ID, "0"), strings.Join(todos, ", "))))
	}
	p.Done[l.ID] = true
	if nxt := c.next(l); nxt != nil {
		p.Current = nxt.ID
		created, err := c.enterWorkspace(p, nxt)
		if err != nil {
			return true, err
		}
		fmt.Printf("\n%s %s %s\n", L("Nästa lektion:", "Next lesson:", "次のレッスン:"), nxt.ID, bold(nxt.Title.String()))
		for _, f := range created {
			fmt.Println(dim(L("  ny fil: workspace/", "  new file: workspace/", "  新しいファイル: workspace/") + f))
		}
		fmt.Println(L("Kör", "Run", "次を実行:"), cyan("./mudschool show"))
	} else {
		fmt.Println("\n" + bold(L("🎉 Du har klarat alla lektioner som finns hittills!",
			"🎉 You have finished every lesson that exists so far!",
			"🎉 現在あるレッスンをすべて終えました！")))
		fmt.Println(L("Läs docs/sv/CURRICULUM.md för nästa fas och idéer att bygga vidare på.",
			"Read docs/en/CURRICULUM.md for the next phase and ideas to build on.",
			"次のフェーズと発展のアイデアは docs/ja/CURRICULUM.md を読んでください。"))
	}
	return true, c.saveProgress(p)
}

func (c *course) cmdWatch(race bool) error {
	last := map[string]time.Time{}
	changed := func() bool {
		now := map[string]time.Time{}
		_ = filepath.WalkDir(c.workspace(), func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			name := d.Name()
			if d.IsDir() && (name == "target" || name == ".git") {
				return filepath.SkipDir
			}
			if d.IsDir() || strings.HasPrefix(name, "mudschool_") || strings.HasPrefix(name, ".") {
				return nil
			}
			if info, err := d.Info(); err == nil {
				now[path] = info.ModTime()
			}
			return nil
		})
		diff := len(now) != len(last)
		for k, v := range now {
			if !last[k].Equal(v) {
				diff = true
			}
		}
		last = now
		return diff
	}
	for {
		if changed() {
			fmt.Print("\x1b[2J\x1b[H")
			if _, err := c.cmdCheck(race, true); err != nil {
				return err
			}
			changed() // ignore files the test run itself created
			fmt.Println(dim(L("\n(bevakar workspace/ – spara en fil för att köra igen, Ctrl+C avslutar)",
				"\n(watching workspace/ – save a file to run again, Ctrl+C to quit)",
				"\n（workspace/ を監視中。保存すると再実行します。Ctrl+C で終了）")))
		}
		time.Sleep(700 * time.Millisecond)
	}
}

func (c *course) cmdSolution(apply bool) error {
	p, err := c.loadProgress()
	if err != nil {
		return err
	}
	l, err := c.current(p)
	if err != nil {
		return err
	}
	if !apply {
		fmt.Println(yellow(fmt.Sprintf(L("Referenslösningen för lektion %s finns i:",
			"The reference solution for lesson %s is in:",
			"レッスン %s の模範解答はここにあります:"), l.ID)))
		for _, tree := range []string{"solution", "solution." + lang} {
			_ = filepath.WalkDir(filepath.Join(l.Dir, tree), func(path string, d fs.DirEntry, err error) error {
				if err == nil && !d.IsDir() {
					fmt.Println("  " + rel(c.root, path))
				}
				return nil
			})
		}
		fmt.Println()
		fmt.Println(L("Tips: jämför med din egen kod, t.ex.", "Tip: compare with your own code, e.g.", "ヒント: 自分のコードと比べてみましょう。例:"))
		fmt.Println(dim("  diff -ru " + rel(c.root, filepath.Join(l.Dir, "solution")) + "/server workspace/server"))
		fmt.Println(L("Vill du skriva över din kod med lösningen? Kör",
			"Want to overwrite your code with the solution? Run",
			"自分のコードを模範解答で上書きするには:"), cyan("./mudschool solution --apply"))
		return nil
	}
	written, err := overlay(l.Dir, "solution", c.workspace(), true)
	for _, f := range written {
		fmt.Println(dim(L("  skrev workspace/", "  wrote workspace/", "  書き込み: workspace/") + f))
	}
	return err
}

func rel(base, p string) string {
	r, err := filepath.Rel(base, p)
	if err != nil {
		return p
	}
	return r
}

func (c *course) cmdRun(args []string) error {
	if len(args) == 0 {
		return errors.New(L("användning: mudschool run server|client [argument...]",
			"usage: mudschool run server|client [arguments...]",
			"使い方: mudschool run server|client [引数...]"))
	}
	ws := c.workspace()
	switch args[0] {
	case "server":
		_, err := runIn(filepath.Join(ws, "server"), false, nil, append([]string{"go", "run", "."}, args[1:]...)...)
		return err
	case "client":
		_, err := runIn(filepath.Join(ws, "client"), false, nil, append([]string{"cargo", "run", "--quiet", "--"}, args[1:]...)...)
		return err
	}
	return fmt.Errorf("%q: server | client", args[0])
}

// cmdVerify builds the reference solution lesson by lesson in a temporary
// directory, for every language, and checks for each lesson that
//  1. the code *before* the lesson (previous solution + scaffold) does NOT pass, and
//  2. the lesson's solution passes.
func (c *course) cmdVerify(only []string, langFlag string) error {
	langs := Languages
	if langFlag != "" {
		langs = []string{langFlag}
	}
	selected := map[string]bool{}
	for _, id := range only {
		l, err := c.lesson(id)
		if err != nil {
			return err
		}
		selected[l.ID] = true
	}
	failures := 0
	for _, l := range langs {
		lang = l
		n, err := c.verifyLang(selected)
		if err != nil {
			return err
		}
		failures += n
	}
	if failures > 0 {
		return fmt.Errorf("%d failure(s)", failures)
	}
	fmt.Println(green("All lessons verified."))
	return nil
}

func (c *course) verifyLang(selected map[string]bool) (int, error) {
	ws, err := os.MkdirTemp("", "mudschool-verify-")
	if err != nil {
		return 0, err
	}
	defer os.RemoveAll(ws)
	env := []string{"CARGO_TARGET_DIR=" + filepath.Join(c.root, ".verify-target")}
	if _, err := overlay(c.root, "starter", ws, true); err != nil {
		return 0, err
	}
	failures := 0
	for _, l := range c.lessons {
		if _, err := c.enter(ws, l); err != nil {
			return 0, err
		}
		run := len(selected) == 0 || selected[l.ID]
		dir, cmds := checkCommands(l.Track, false)
		check := func() (string, bool) {
			if _, err := c.placeTests(ws, l); err != nil {
				return err.Error(), false
			}
			var all strings.Builder
			for _, argv := range cmds {
				out, err := runIn(filepath.Join(ws, dir), true, env, argv...)
				all.WriteString(out)
				if err != nil {
					return all.String(), false
				}
			}
			return all.String(), true
		}
		if run {
			if _, ok := check(); ok {
				fmt.Printf("%s [%s] %s %s: tests pass already BEFORE the solution\n", red("✘"), lang, l.ID, l.Title)
				failures++
			}
		}
		if _, err := overlay(l.Dir, "solution", ws, true); err != nil {
			return 0, err
		}
		if !run {
			continue
		}
		if out, ok := check(); ok {
			fmt.Printf("%s [%s] %s %s\n", green("✔"), lang, l.ID, l.Title)
		} else {
			fmt.Printf("%s [%s] %s %s: the solution fails its tests\n%s\n", red("✘"), lang, l.ID, l.Title, out)
			failures++
		}
	}
	return failures, nil
}

// cmdReference builds the reference solution (up to a lesson) in a directory.
func (c *course) cmdReference(args []string) error {
	if len(args) < 1 {
		return errors.New("usage: mudschool reference <dir> [id]")
	}
	dst := args[0]
	if exists(dst) {
		return fmt.Errorf("%s already exists", dst)
	}
	last := c.lessons[len(c.lessons)-1]
	if len(args) > 1 {
		var err error
		if last, err = c.lesson(args[1]); err != nil {
			return err
		}
	}
	if _, err := overlay(c.root, "starter", dst, true); err != nil {
		return err
	}
	for _, l := range c.lessons {
		if _, err := c.enter(dst, l); err != nil {
			return err
		}
		if _, err := overlay(l.Dir, "solution", dst, true); err != nil {
			return err
		}
		if l == last {
			break
		}
	}
	fmt.Printf("Reference solution up to lesson %s (%s) is in %s\n", last.ID, lang, dst)
	return nil
}
