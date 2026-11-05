package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// FileRecord remembers a file the course copied into the workspace, so that
// `update` can tell whether the student has changed it since.
type FileRecord struct {
	Hash   string `json:"hash"`             // sha256 of what the course wrote
	Source string `json:"source"`           // "starter" or "scaffold"
	Lesson string `json:"lesson,omitempty"` // lesson ID for "scaffold"
	Lang   string `json:"lang"`             // course language when it was written
}

func hashOf(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// record notes that the course wrote rels (relative to the workspace) from
// the given source.
func (c *course) record(p *Progress, rels []string, source, lesson string) {
	if p.Files == nil {
		p.Files = map[string]FileRecord{}
	}
	for _, rel := range rels {
		data, err := os.ReadFile(filepath.Join(c.workspace(), rel))
		if err != nil {
			continue
		}
		p.Files[filepath.ToSlash(rel)] = FileRecord{Hash: hashOf(data), Source: source, Lesson: lesson, Lang: lang}
	}
}

// enterWorkspace enters lesson l in the student's workspace and records the
// files it created.
func (c *course) enterWorkspace(p *Progress, l *Lesson) ([]string, error) {
	written, err := c.enter(c.workspace(), l)
	c.record(p, written, "scaffold", l.ID)
	return written, err
}

// sourceDirs returns the directories that hold the course's version of files
// from source/lesson, language-specific first.
func (c *course) sourceDirs(source, lessonID, language string) []string {
	if source == "starter" {
		return []string{filepath.Join(c.root, "starter."+language), filepath.Join(c.root, "starter")}
	}
	l, err := c.lesson(lessonID)
	if err != nil {
		return nil
	}
	return []string{filepath.Join(l.Dir, "scaffold."+language), filepath.Join(l.Dir, "scaffold")}
}

// courseVersion reads the course's current version of rel from source.
func (c *course) courseVersion(source, lessonID, language, rel string) ([]byte, bool) {
	for _, dir := range c.sourceDirs(source, lessonID, language) {
		if data, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(rel))); err == nil {
			return data, true
		}
	}
	return nil, false
}

type updateKind int

const (
	updAdd     updateKind = iota // the file is missing in the workspace
	updReplace                   // unchanged by the student: safe to replace
	updBackup                    // messages file changed by the student: back up, then replace
	updNew                       // changed by the student: write <file>.new next to it
)

type updateAction struct {
	kind   updateKind
	rel    string // slash-separated, relative to the workspace
	data   []byte
	record FileRecord
}

func (a updateAction) safe() bool { return a.kind == updAdd || a.kind == updReplace }

func isMessagesFile(rel string) bool {
	base := filepath.Base(rel)
	return base == "messages.go" || base == "messages.rs"
}

// planUpdate compares the workspace with the course. It also records
// ("bootstraps") files that are identical to the course's current version
// but were copied in before the course started keeping records.
func (c *course) planUpdate(p *Progress) []updateAction {
	cur, err := c.current(p)
	if err != nil {
		return nil
	}
	if p.Files == nil {
		p.Files = map[string]FileRecord{}
	}
	type candidate struct{ source, lesson string }
	cands := map[string]candidate{}
	addTree := func(dir, source, lesson string) {
		_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err == nil && !d.IsDir() {
				rel, _ := filepath.Rel(dir, path)
				cands[filepath.ToSlash(rel)] = candidate{source, lesson}
			}
			return nil
		})
	}
	addTree(filepath.Join(c.root, "starter"), "starter", "")
	addTree(filepath.Join(c.root, "starter."+p.Lang), "starter", "")
	for _, l := range c.lessons {
		addTree(filepath.Join(l.Dir, "scaffold"), "scaffold", l.ID)
		addTree(filepath.Join(l.Dir, "scaffold."+p.Lang), "scaffold", l.ID)
		if l == cur {
			break
		}
	}

	rels := make([]string, 0, len(cands))
	for rel := range cands {
		rels = append(rels, rel)
	}
	sort.Strings(rels)

	var plan []updateAction
	for _, rel := range rels {
		cand := cands[rel]
		rec, hasRec := p.Files[rel]
		language := p.Lang
		if hasRec {
			language = rec.Lang
			cand = candidate{rec.Source, rec.Lesson}
		}
		want, ok := c.courseVersion(cand.source, cand.lesson, language, rel)
		if !ok {
			continue // no longer part of the course
		}
		newRec := FileRecord{Hash: hashOf(want), Source: cand.source, Lesson: cand.lesson, Lang: language}
		disk, err := os.ReadFile(filepath.Join(c.workspace(), filepath.FromSlash(rel)))
		switch {
		case errors.Is(err, fs.ErrNotExist):
			plan = append(plan, updateAction{updAdd, rel, want, newRec})
		case err != nil:
			continue
		case !hasRec:
			switch {
			case hashOf(disk) == newRec.Hash:
				p.Files[rel] = newRec // bootstrap: identical to the course's version
			case c.wasCourseVersion(cand.source, cand.lesson, language, rel, hashOf(disk)):
				// an older, untouched course version from before records existed
				plan = append(plan, updateAction{updReplace, rel, want, newRec})
			}
			// otherwise it is the student's own work: leave it alone
		case rec.Hash == newRec.Hash:
			// up to date
		case hashOf(disk) == rec.Hash:
			plan = append(plan, updateAction{updReplace, rel, want, newRec})
		case isMessagesFile(rel):
			plan = append(plan, updateAction{updBackup, rel, want, newRec})
		default:
			plan = append(plan, updateAction{updNew, rel, want, newRec})
		}
	}
	return plan
}

// applyUpdate carries out the actions (only the safe ones if safeOnly) and
// prints what it did. It returns how many actions were left for `update`.
func (c *course) applyUpdate(p *Progress, plan []updateAction, safeOnly bool) (pending int, err error) {
	for _, a := range plan {
		if safeOnly && !a.safe() {
			pending++
			continue
		}
		target := filepath.Join(c.workspace(), filepath.FromSlash(a.rel))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return pending, err
		}
		switch a.kind {
		case updAdd, updReplace:
			if err := os.WriteFile(target, a.data, 0o644); err != nil {
				return pending, err
			}
			if a.kind == updAdd {
				fmt.Println(green("  +"), L("lade till", "added", "追加:"), "workspace/"+a.rel)
			} else {
				fmt.Println(green("  ✔"), L("uppdaterade", "updated", "更新:"), "workspace/"+a.rel)
			}
			p.Files[a.rel] = a.record
		case updBackup:
			old, _ := os.ReadFile(target)
			backup := freeBackupName(target)
			if err := os.WriteFile(backup, old, 0o644); err != nil {
				return pending, err
			}
			if err := os.WriteFile(target, a.data, 0o644); err != nil {
				return pending, err
			}
			shown := filepath.ToSlash(filepath.Join(filepath.Dir(a.rel), filepath.Base(backup)))
			fmt.Println(yellow("  !"), fmt.Sprintf(L(
				"uppdaterade workspace/%s – dina ändringar finns i workspace/%s; för över dem igen om du vill",
				"updated workspace/%s – your edits are in workspace/%s; copy them back if you want them",
				"workspace/%s を更新しました。あなたの変更は workspace/%s にあります。必要なら書き戻してください"), a.rel, shown))
			p.Files[a.rel] = a.record
		case updNew:
			if err := os.WriteFile(target+".new", a.data, 0o644); err != nil {
				return pending, err
			}
			fmt.Println(yellow("  !"), fmt.Sprintf(L(
				"workspace/%s har ändrats av dig och av kursen – kursens nya version finns i workspace/%s.new",
				"workspace/%s was changed by you and by the course – the course's new version is in workspace/%s.new",
				"workspace/%s はあなたとコースの両方で変更されています。コースの新しい版は workspace/%s.new にあります"), a.rel, a.rel))
			// Remember the new course version so that the same .new is not
			// written again on the next update.
			p.Files[a.rel] = a.record
		}
	}
	return pending, nil
}

func (c *course) cmdUpdate() error {
	p, err := c.loadProgress()
	if err != nil {
		return err
	}
	plan := c.planUpdate(p)
	if len(plan) == 0 {
		fmt.Println(green("✔"), L("Arbetsytan är redan uppdaterad.", "Your workspace is already up to date.", "作業領域はすでに最新です。"))
		return c.saveProgress(p)
	}
	fmt.Println(bold(L("Uppdaterar arbetsytan", "Updating your workspace", "作業領域を更新しています")))
	if _, err := c.applyUpdate(p, plan, false); err != nil {
		return err
	}
	fmt.Println(dim(L(
		"Din egen kod rörs aldrig: bara filer du inte har ändrat byts ut.",
		"Your own code is never touched: only files you haven't changed are replaced.",
		"あなたのコードには触れません。置き換えるのは、あなたが変更していないファイルだけです。")))
	return c.saveProgress(p)
}

// describePending is used by doctor.
func (c *course) describePending(p *Progress) (int, string) {
	plan := c.planUpdate(p)
	var names []string
	for _, a := range plan {
		names = append(names, a.rel)
	}
	return len(plan), strings.Join(names, ", ")
}

// wasCourseVersion reports whether hash matches any version of rel that the
// course has ever had in git (for workspaces created before records existed).
// Without git it reports false, which only means "leave the file alone".
func (c *course) wasCourseVersion(source, lessonID, language, rel, hash string) bool {
	for _, dir := range c.sourceDirs(source, lessonID, language) {
		path, err := filepath.Rel(c.root, filepath.Join(dir, filepath.FromSlash(rel)))
		if err != nil {
			continue
		}
		path = filepath.ToSlash(path)
		out, err := exec.Command("git", "-C", c.root, "log", "--format=%H", "--", path).Output()
		if err != nil {
			continue
		}
		for _, rev := range strings.Fields(string(out)) {
			data, err := exec.Command("git", "-C", c.root, "show", rev+":"+path).Output()
			if err == nil && hashOf(data) == hash {
				return true
			}
			// Git may have converted line endings on checkout (Windows).
			if err == nil && hashOf(bytes.ReplaceAll(data, []byte("\n"), []byte("\r\n"))) == hash {
				return true
			}
		}
	}
	return false
}
