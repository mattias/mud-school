package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Lesson beskriver en lektion, inläst från lessons/NN-namn/lesson.json.
type Lesson struct {
	ID       string   `json:"-"`
	Dir      string   `json:"-"`
	Title    string   `json:"title"`
	Track    string   `json:"track"` // "go" eller "rust"
	Summary  string   `json:"summary"`
	Concepts []string `json:"concepts"`
}

// Progress sparas i workspace/.mudschool.json.
type Progress struct {
	Current string          `json:"current"`
	Done    map[string]bool `json:"done"`
	Hints   map[string]int  `json:"hints"`
}

type course struct {
	root    string
	lessons []*Lesson
}

var lessonDirRe = regexp.MustCompile(`^(\d\d)-`)

// findRoot letar uppåt efter kursens rotkatalog (den som har lessons/ och starter/).
func findRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if isDir(filepath.Join(dir, "lessons")) && isDir(filepath.Join(dir, "starter")) {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("hittar inte kursens rotkatalog (en katalog med lessons/ och starter/)")
		}
		dir = parent
	}
}

func (c *course) load() error {
	entries, err := os.ReadDir(filepath.Join(c.root, "lessons"))
	if err != nil {
		return err
	}
	for _, e := range entries {
		m := lessonDirRe.FindStringSubmatch(e.Name())
		if !e.IsDir() || m == nil {
			continue
		}
		dir := filepath.Join(c.root, "lessons", e.Name())
		data, err := os.ReadFile(filepath.Join(dir, "lesson.json"))
		if err != nil {
			return fmt.Errorf("lektion %s: %w", e.Name(), err)
		}
		l := &Lesson{}
		if err := json.Unmarshal(data, l); err != nil {
			return fmt.Errorf("lektion %s: lesson.json: %w", e.Name(), err)
		}
		l.ID, l.Dir = m[1], dir
		if l.Track != "go" && l.Track != "rust" {
			return fmt.Errorf("lektion %s: track måste vara \"go\" eller \"rust\"", e.Name())
		}
		c.lessons = append(c.lessons, l)
	}
	sort.Slice(c.lessons, func(i, j int) bool { return c.lessons[i].ID < c.lessons[j].ID })
	if len(c.lessons) == 0 {
		return errors.New("inga lektioner hittades")
	}
	return nil
}

func (c *course) workspace() string { return filepath.Join(c.root, "workspace") }

func (c *course) lesson(id string) (*Lesson, error) {
	id = strings.TrimLeft(id, "0")
	for _, l := range c.lessons {
		if strings.TrimLeft(l.ID, "0") == id {
			return l, nil
		}
	}
	return nil, fmt.Errorf("ingen lektion med id %q (se `mudschool list`)", id)
}

func (c *course) next(l *Lesson) *Lesson {
	for i, x := range c.lessons {
		if x == l && i+1 < len(c.lessons) {
			return c.lessons[i+1]
		}
	}
	return nil
}

func (c *course) progressPath() string { return filepath.Join(c.workspace(), ".mudschool.json") }

func (c *course) loadProgress() (*Progress, error) {
	data, err := os.ReadFile(c.progressPath())
	if errors.Is(err, fs.ErrNotExist) {
		return nil, errors.New("du har ingen arbetsyta än – kör `./mudschool init` först")
	}
	if err != nil {
		return nil, err
	}
	p := &Progress{}
	if err := json.Unmarshal(data, p); err != nil {
		return nil, fmt.Errorf("trasig %s: %w", c.progressPath(), err)
	}
	if p.Done == nil {
		p.Done = map[string]bool{}
	}
	if p.Hints == nil {
		p.Hints = map[string]int{}
	}
	return p, nil
}

func (c *course) saveProgress(p *Progress) error {
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(c.progressPath(), append(data, '\n'), 0o644)
}

func (c *course) current(p *Progress) (*Lesson, error) { return c.lesson(p.Current) }

// enter förbereder arbetsytan för en lektion: nya filer från scaffold/
// kopieras in, men befintliga filer rörs aldrig.
func (c *course) enter(ws string, l *Lesson) ([]string, error) {
	return copyTree(filepath.Join(l.Dir, "scaffold"), ws, false)
}

// placeTests tar bort gamla kurstester (filer/kataloger som börjar med
// "mudschool_") och kopierar in den aktuella lektionens tester.
func (c *course) placeTests(ws string, l *Lesson) error {
	if err := cleanTests(ws); err != nil {
		return err
	}
	if _, err := copyTree(filepath.Join(c.root, "lessons", "_shared", l.Track), ws, true); err != nil {
		return err
	}
	_, err := copyTree(filepath.Join(l.Dir, "tests"), ws, true)
	return err
}

func cleanTests(ws string) error {
	var victims []string
	err := filepath.WalkDir(ws, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && (d.Name() == "target" || d.Name() == ".git") {
			return filepath.SkipDir
		}
		if strings.HasPrefix(d.Name(), "mudschool_") {
			victims = append(victims, path)
			if d.IsDir() {
				return filepath.SkipDir
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	for _, v := range victims {
		if err := os.RemoveAll(v); err != nil {
			return err
		}
	}
	return nil
}

// copyTree kopierar src till dst. Med overwrite=false hoppas befintliga filer över.
// Returnerar de filer som skrevs (relativt dst).
func copyTree(src, dst string, overwrite bool) ([]string, error) {
	if !isDir(src) {
		return nil, nil
	}
	var written []string
	err := filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		if !overwrite && exists(target) {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.WriteFile(target, data, 0o644); err != nil {
			return err
		}
		written = append(written, rel)
		return nil
	})
	return written, err
}

func isDir(p string) bool {
	st, err := os.Stat(p)
	return err == nil && st.IsDir()
}

func exists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}
