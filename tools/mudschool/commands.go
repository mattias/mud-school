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
		fmt.Println("Du har redan en arbetsyta i", ws)
		fmt.Println("Kör", cyan("./mudschool show"), "för att se din aktuella lektion.")
		return nil
	}
	if _, err := copyTree(filepath.Join(c.root, "starter"), ws, false); err != nil {
		return err
	}
	first := c.lessons[0]
	if _, err := c.enter(ws, first); err != nil {
		return err
	}
	p := &Progress{Current: first.ID, Done: map[string]bool{}, Hints: map[string]int{}}
	if err := c.saveProgress(p); err != nil {
		return err
	}
	fmt.Println(green("✔ Arbetsytan är skapad:"), ws)
	fmt.Println()
	fmt.Println("  workspace/server/   din MUD-server i Go")
	fmt.Println("  workspace/client/   din MUD-klient i Rust")
	fmt.Println()
	fmt.Println("Nästa steg:", cyan("./mudschool show"))
	return nil
}

func (c *course) cmdList() error {
	p, err := c.loadProgress()
	if err != nil {
		p = &Progress{Done: map[string]bool{}}
	}
	fmt.Println(bold("Lektioner"))
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
			fmt.Printf("           %s\n", dim(l.Summary))
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
	data, err := os.ReadFile(filepath.Join(l.Dir, "README.md"))
	if err != nil {
		return err
	}
	fmt.Print(renderMarkdown(string(data)))
	fmt.Println(dim("─────────────────────────────────────────────"))
	fmt.Println(dim("Kör ") + cyan("./mudschool check") + dim(" när du är klar, eller ") + cyan("./mudschool hint") + dim(" om du kör fast."))
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
	data, err := os.ReadFile(filepath.Join(l.Dir, "HINTS.md"))
	if errors.Is(err, fs.ErrNotExist) {
		fmt.Println("Den här lektionen har inga ledtrådar.")
		return nil
	} else if err != nil {
		return err
	}
	hints := splitHints(string(data))
	n := p.Hints[l.ID]
	if n >= len(hints) {
		fmt.Println("Du har sett alla ledtrådar. Fortfarande fast? Prova", cyan("./mudschool solution"))
		n = len(hints) - 1
	} else {
		p.Hints[l.ID] = n + 1
		if err := c.saveProgress(p); err != nil {
			return err
		}
	}
	fmt.Printf(dim("Ledtråd %d av %d\n\n"), n+1, len(hints))
	fmt.Print(renderMarkdown(hints[n]))
	return nil
}

// splitHints delar upp HINTS.md vid varje "## "-rubrik.
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
		return errors.New("användning: mudschool goto <id>")
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
	if err := c.saveProgress(p); err != nil {
		return err
	}
	if _, err := c.enter(c.workspace(), l); err != nil {
		return err
	}
	fmt.Printf("Aktuell lektion: %s %s\n", l.ID, bold(l.Title))
	return nil
}

// checkCommands returnerar kommandona som testar ett spår.
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
	cmd.Env = append(os.Environ(), env...)
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
	if _, err := c.enter(ws, l); err != nil {
		return false, err
	}
	if err := c.placeTests(ws, l); err != nil {
		return false, err
	}
	fmt.Printf("%s Lektion %s: %s\n\n", cyan("▶"), l.ID, bold(l.Title))
	dir, cmds := checkCommands(l.Track, race)
	for _, argv := range cmds {
		fmt.Println(dim("$ " + strings.Join(argv, " ")))
		if _, err := runIn(filepath.Join(ws, dir), false, nil, argv...); err != nil {
			fmt.Println()
			fmt.Println(red("✘ Inte riktigt än."), "Läs felmeddelandena ovan – de berättar vad testerna förväntade sig.")
			fmt.Println("  Instruktioner:", cyan("./mudschool show"), "  Ledtråd:", cyan("./mudschool hint"))
			return false, nil
		}
	}
	fmt.Println()
	fmt.Println(green("✔ Alla tester gick igenom!"), "Lektion", l.ID, "är klar.")
	if !advance {
		return true, nil
	}
	p.Done[l.ID] = true
	if nxt := c.next(l); nxt != nil {
		p.Current = nxt.ID
		created, err := c.enter(ws, nxt)
		if err != nil {
			return true, err
		}
		fmt.Printf("\nNästa lektion: %s %s\n", nxt.ID, bold(nxt.Title))
		for _, f := range created {
			fmt.Println(dim("  ny fil: workspace/" + f))
		}
		fmt.Println("Kör", cyan("./mudschool show"), "för att läsa instruktionerna.")
	} else {
		fmt.Println("\n" + bold("🎉 Du har klarat alla lektioner som finns hittills!"))
		fmt.Println("Läs docs/KURSPLAN.md för nästa fas och idéer att bygga vidare på.")
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
			changed() // ignorera filer som testkörningen själv skapade
			fmt.Println(dim("\n(bevakar workspace/ – spara en fil för att köra igen, Ctrl+C avslutar)"))
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
	sol := filepath.Join(l.Dir, "solution")
	if !apply {
		fmt.Println(yellow("Referenslösningen för lektion " + l.ID + " finns i:"))
		_ = filepath.WalkDir(sol, func(path string, d fs.DirEntry, err error) error {
			if err == nil && !d.IsDir() {
				rel, _ := filepath.Rel(c.root, path)
				fmt.Println("  " + rel)
			}
			return nil
		})
		fmt.Println()
		fmt.Println("Tips: jämför med din egen kod, t.ex.")
		fmt.Println(dim("  diff -ru " + rel(c.root, sol) + "/server workspace/server"))
		fmt.Println("Vill du skriva över din kod med lösningen? Kör", cyan("./mudschool solution --apply"))
		return nil
	}
	written, err := copyTree(sol, c.workspace(), true)
	for _, f := range written {
		fmt.Println(dim("  skrev workspace/" + f))
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
		return errors.New("användning: mudschool run server|client [argument...]")
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
	return fmt.Errorf("okänt mål %q – välj server eller client", args[0])
}

// cmdVerify bygger upp referenslösningen lektion för lektion i en temporär
// katalog och kontrollerar för varje lektion att
//  1. koden *före* lektionen (förra lösningen + scaffold) INTE klarar testerna, och
//  2. lektionens lösning klarar testerna.
func (c *course) cmdVerify(only []string) error {
	ws, err := os.MkdirTemp("", "mudschool-verify-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(ws)
	env := []string{"CARGO_TARGET_DIR=" + filepath.Join(c.root, ".verify-target")}
	if _, err := copyTree(filepath.Join(c.root, "starter"), ws, true); err != nil {
		return err
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
	for _, l := range c.lessons {
		if _, err := c.enter(ws, l); err != nil {
			return err
		}
		run := len(selected) == 0 || selected[l.ID]
		dir, cmds := checkCommands(l.Track, false)
		check := func() (string, bool) {
			if err := c.placeTests(ws, l); err != nil {
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
				fmt.Printf("%s %s %s: testerna går igenom redan FÖRE lösningen\n", red("✘"), l.ID, l.Title)
				failures++
			}
		}
		if _, err := copyTree(filepath.Join(l.Dir, "solution"), ws, true); err != nil {
			return err
		}
		if !run {
			continue
		}
		if out, ok := check(); ok {
			fmt.Printf("%s %s %s\n", green("✔"), l.ID, l.Title)
		} else {
			fmt.Printf("%s %s %s: lösningen klarar inte testerna\n%s\n", red("✘"), l.ID, l.Title, out)
			failures++
		}
	}
	_ = cleanTests(ws)
	if failures > 0 {
		return fmt.Errorf("%d lektion(er) misslyckades", failures)
	}
	fmt.Println(green("Alla lektioner verifierade."))
	return nil
}

// cmdReference bygger referenslösningen (till och med en viss lektion) i en katalog.
func (c *course) cmdReference(args []string) error {
	if len(args) < 1 {
		return errors.New("användning: mudschool reference <katalog> [id]")
	}
	dst := args[0]
	if exists(dst) {
		return fmt.Errorf("%s finns redan", dst)
	}
	last := c.lessons[len(c.lessons)-1]
	if len(args) > 1 {
		var err error
		if last, err = c.lesson(args[1]); err != nil {
			return err
		}
	}
	if _, err := copyTree(filepath.Join(c.root, "starter"), dst, true); err != nil {
		return err
	}
	for _, l := range c.lessons {
		if _, err := c.enter(dst, l); err != nil {
			return err
		}
		if _, err := copyTree(filepath.Join(l.Dir, "solution"), dst, true); err != nil {
			return err
		}
		if l == last {
			break
		}
	}
	fmt.Printf("Referenslösningen till och med lektion %s finns i %s\n", last.ID, dst)
	return nil
}
