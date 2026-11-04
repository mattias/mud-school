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

// Text is a string in several languages: {"sv": "...", "en": "...", "ja": "..."}.
type Text map[string]string

// String returns the text in the active language (falling back to en, then sv).
func (t Text) String() string {
	for _, l := range []string{lang, "en", "sv"} {
		if s := t[l]; s != "" {
			return s
		}
	}
	return ""
}

// Lesson describes a lesson, read from lessons/NN-name/lesson.json.
type Lesson struct {
	ID      string `json:"-"`
	Dir     string `json:"-"`
	Title   Text   `json:"title"`
	Track   string `json:"track"` // "go" or "rust"
	Summary Text   `json:"summary"`
}

// Progress is stored in workspace/.mudschool.json.
type Progress struct {
	Lang    string          `json:"lang"`
	Current string          `json:"current"`
	Done    map[string]bool `json:"done"`
	Hints   map[string]int  `json:"hints"`
}

type course struct {
	root    string
	lessons []*Lesson
}

var lessonDirRe = regexp.MustCompile(`^(\d\d)-`)

// findRoot walks upwards looking for the course root (a directory with lessons/ and starter/).
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
			return "", errors.New(L(
				"hittar inte kursens rotkatalog (en katalog med lessons/ och starter/)",
				"cannot find the course root (a directory containing lessons/ and starter/)",
				"コースのルートディレクトリ（lessons/ と starter/ を含むディレクトリ）が見つかりません"))
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
			return fmt.Errorf("%s: %w", e.Name(), err)
		}
		l := &Lesson{}
		if err := json.Unmarshal(data, l); err != nil {
			return fmt.Errorf("%s: lesson.json: %w", e.Name(), err)
		}
		l.ID, l.Dir = m[1], dir
		if l.Track != "go" && l.Track != "rust" {
			return fmt.Errorf("%s: track must be \"go\" or \"rust\"", e.Name())
		}
		c.lessons = append(c.lessons, l)
	}
	sort.Slice(c.lessons, func(i, j int) bool { return c.lessons[i].ID < c.lessons[j].ID })
	if len(c.lessons) == 0 {
		return errors.New("no lessons found")
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
	return nil, fmt.Errorf(L("ingen lektion med id %q (se `mudschool list`)",
		"no lesson with id %q (see `mudschool list`)",
		"ID %q のレッスンはありません（`mudschool list` を参照）"), id)
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
		return nil, errors.New(L(
			"du har ingen arbetsyta än – kör `./mudschool init` först",
			"you don't have a workspace yet – run `./mudschool init` first",
			"作業領域がまだありません。先に `./mudschool init` を実行してください"))
	}
	if err != nil {
		return nil, err
	}
	p := &Progress{}
	if err := json.Unmarshal(data, p); err != nil {
		return nil, fmt.Errorf("%s: %w", c.progressPath(), err)
	}
	if p.Done == nil {
		p.Done = map[string]bool{}
	}
	if p.Hints == nil {
		p.Hints = map[string]int{}
	}
	if !validLang(p.Lang) {
		p.Lang = lang
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

// localized returns the path to name (e.g. "README") in the active language:
// README.<lang>.md, falling back to README.md (Swedish, the original).
func localized(dir, name string) string {
	p := filepath.Join(dir, name+"."+lang+".md")
	if exists(p) {
		return p
	}
	return filepath.Join(dir, name+".md")
}

// overlay copies tree/ and then tree.<lang>/ from dir into dst.
func overlay(dir, tree, dst string, overwrite bool) ([]string, error) {
	a, err := copyTree(filepath.Join(dir, tree), dst, overwrite)
	if err != nil {
		return a, err
	}
	b, err := copyTree(filepath.Join(dir, tree+"."+lang), dst, overwrite)
	return append(a, b...), err
}

// enter prepares the workspace for a lesson: new files from scaffold/ (and
// scaffold.<lang>/) are copied in, but existing files are never touched.
func (c *course) enter(ws string, l *Lesson) ([]string, error) {
	return overlay(l.Dir, "scaffold", ws, false)
}

// placeTests removes old course tests (files/directories starting with
// "mudschool_") and copies in the current lesson's tests.
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

// copyTree copies src to dst. With overwrite=false existing files are skipped.
// Returns the files written (relative to dst).
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
