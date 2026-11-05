package main

import (
	"bytes"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
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
		key := filepath.ToSlash(rel)
		p.Files[key] = FileRecord{Hash: hashOf(data), Source: source, Lesson: lesson, Lang: lang}
		delete(p.Own, key)
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
	updReplace updateKind = iota // unchanged by the student: safe to replace
	updAdd                       // missing and never recorded: added by `update` only
	updBackup                    // messages file changed by the student: back up, then replace
	updNew                       // changed by the student: write <file>.new next to it
)

type updateAction struct {
	kind   updateKind
	rel    string // slash-separated, relative to the workspace
	data   []byte
	record FileRecord
}

// safe actions are applied automatically by `check`.
func (a updateAction) safe() bool { return a.kind == updReplace }

func isMessagesFile(rel string) bool {
	base := filepath.Base(rel)
	return base == "messages.go" || base == "messages.rs"
}

// planUpdate compares the workspace with the course. It also updates the
// records ("bootstraps") for files that already match the course.
//
// The rules, per file the course provides (starter, and the scaffolds of all
// lessons up to the current one):
//   - missing, recorded:     the student deleted or renamed it – leave it alone
//   - missing, not recorded: offer to add it if it is a starter file or from
//     the current lesson (only `update` does that)
//   - equal to the course:   up to date (record it)
//   - recorded, unchanged by the student, course changed: replace
//   - recorded, changed by both: messages.go/.rs are backed up and replaced,
//     everything else gets <file>.new
//   - not recorded: replace if it equals an older course version from git
//     history, otherwise it is the student's own work – leave it alone
func (c *course) planUpdate(p *Progress) ([]updateAction, error) {
	cur, err := c.current(p)
	if err != nil {
		return nil, err
	}
	if p.Files == nil {
		p.Files = map[string]FileRecord{}
	}
	if p.Own == nil {
		p.Own = map[string]string{}
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

	hist := &gitHistory{c: c}
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
		if errors.Is(err, fs.ErrNotExist) {
			// Only add files that were never there. A recorded file the
			// student deleted or renamed stays deleted, and so does a missing
			// file from an earlier lesson (we can't tell if it was renamed).
			if !hasRec && (cand.source == "starter" || cand.lesson == cur.ID) {
				plan = append(plan, updateAction{updAdd, rel, want, newRec})
			}
			continue
		}
		if err != nil {
			return nil, err
		}
		diskHash := hashOf(disk)
		switch {
		case diskHash == newRec.Hash:
			p.Files[rel] = newRec // already the course's current version
			delete(p.Own, rel)
		case !hasRec:
			if p.Own[rel] == diskHash {
				continue // known to be the student's own work
			}
			if hist.contains(cand.source, cand.lesson, language, rel, disk) {
				plan = append(plan, updateAction{updReplace, rel, want, newRec})
			} else {
				p.Own[rel] = diskHash // remember, so that git is not asked again
			}
		case rec.Hash == newRec.Hash:
			// the student changed it, the course didn't: nothing to do
		case diskHash == rec.Hash:
			plan = append(plan, updateAction{updReplace, rel, want, newRec})
		case isMessagesFile(rel):
			plan = append(plan, updateAction{updBackup, rel, want, newRec})
		default:
			plan = append(plan, updateAction{updNew, rel, want, newRec})
		}
	}
	return plan, nil
}

// applyUpdate carries out the actions (only the safe ones if safeOnly) and
// prints what it did. It returns how many actions were left for `update` and
// how many files were backed up.
func (c *course) applyUpdate(p *Progress, plan []updateAction, safeOnly bool) (pending, backups int, err error) {
	for _, a := range plan {
		if safeOnly && !a.safe() {
			pending++
			continue
		}
		target := filepath.Join(c.workspace(), filepath.FromSlash(a.rel))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return pending, backups, err
		}
		switch a.kind {
		case updAdd, updReplace:
			if err := os.WriteFile(target, a.data, 0o644); err != nil {
				return pending, backups, err
			}
			if a.kind == updAdd {
				fmt.Println(green("  +"), L("lade till", "added", "追加:"), "workspace/"+a.rel)
			} else {
				fmt.Println(green("  ✔"), L("uppdaterade", "updated", "更新:"), "workspace/"+a.rel)
			}
		case updBackup:
			old, _ := os.ReadFile(target)
			backup := freeBackupName(target)
			if err := os.WriteFile(backup, old, 0o644); err != nil {
				return pending, backups, err
			}
			if err := os.WriteFile(target, a.data, 0o644); err != nil {
				return pending, backups, err
			}
			backups++
			shown := filepath.ToSlash(filepath.Join(filepath.Dir(a.rel), filepath.Base(backup)))
			fmt.Println(yellow("  !"), fmt.Sprintf(L(
				"uppdaterade workspace/%s – dina ändringar finns i workspace/%s; för över dem igen om du vill",
				"updated workspace/%s – your edits are in workspace/%s; copy them back if you want them",
				"workspace/%s を更新しました。あなたの変更は workspace/%s にあります。必要なら書き戻してください"), a.rel, shown))
		case updNew:
			if err := os.WriteFile(target+".new", a.data, 0o644); err != nil {
				return pending, backups, err
			}
			fmt.Println(yellow("  !"), fmt.Sprintf(L(
				"workspace/%s har ändrats av dig och av kursen – kursens version finns i workspace/%s.new.\n    Jämför dem (t.ex. med git diff --no-index) och för över det du vill ha. Din fil är orörd.",
				"workspace/%s was changed by you and by the course – the course's version is in workspace/%s.new.\n    Compare them (e.g. with git diff --no-index) and copy over what you want. Your file is untouched.",
				"workspace/%s はあなたとコースの両方で変更されています。コースの版は workspace/%s.new にあります。\n    （git diff --no-index などで）比べて、必要な部分を取り込んでください。あなたのファイルはそのままです。"), a.rel, a.rel))
		}
		// Remember the new course version (for .new too, so that the same
		// .new is not written again next time).
		p.Files[a.rel] = a.record
		delete(p.Own, a.rel)
	}
	return pending, backups, nil
}

func (c *course) cmdUpdate() error {
	p, err := c.loadProgress()
	if err != nil {
		return err
	}
	plan, err := c.planUpdate(p)
	if err != nil {
		return err
	}
	if len(plan) == 0 {
		fmt.Println(green("✔"), L("Arbetsytan är redan uppdaterad.", "Your workspace is already up to date.", "作業領域はすでに最新です。"))
		return c.saveProgress(p)
	}
	fmt.Println(bold(L("Uppdaterar arbetsytan", "Updating your workspace", "作業領域を更新しています")))
	_, backups, err := c.applyUpdate(p, plan, false)
	if saveErr := c.saveProgress(p); err == nil {
		err = saveErr
	}
	if err != nil {
		return err
	}
	if backups == 0 {
		fmt.Println(dim(L(
			"Bara filer du inte har ändrat byttes ut – din egen kod är orörd.",
			"Only files you hadn't changed were replaced – your own code is untouched.",
			"置き換えたのは、あなたが変更していないファイルだけです。あなたのコードはそのままです。")))
	}
	return nil
}

// describePending is used by doctor.
func (c *course) describePending(p *Progress) (int, string, error) {
	plan, err := c.planUpdate(p)
	if err != nil {
		return 0, "", err
	}
	var names []string
	for _, a := range plan {
		names = append(names, a.rel)
	}
	return len(plan), strings.Join(names, ", "), nil
}

// gitHistory answers "was this ever the course's version of the file?" for
// workspaces created before the course kept records. It asks git once per
// planUpdate and compares git blob ids. Without git (or outside a git
// checkout) it knows nothing, which only means "leave the file alone".
type gitHistory struct {
	c      *course
	loaded bool
	blobs  map[string]map[string]bool // course path -> blob ids it has had
}

func (h *gitHistory) load() {
	h.loaded = true
	h.blobs = map[string]map[string]bool{}
	cmd := exec.Command("git", "log", "--format=", "--raw", "--no-abbrev", "--no-renames", "--",
		"starter", "starter.en", "starter.sv", "starter.ja", "lessons")
	cmd.Dir = h.c.root
	out, err := cmd.Output()
	if err != nil {
		return
	}
	// Lines look like ":100644 100644 <old> <new> M\t<path>" with the path
	// relative to the repository root, which may be above the course directory.
	prefix := ""
	if top, err := exec.Command("git", "-C", h.c.root, "rev-parse", "--show-prefix").Output(); err == nil {
		prefix = strings.TrimSpace(string(top))
	}
	for _, line := range strings.Split(string(out), "\n") {
		meta, path, ok := strings.Cut(line, "\t")
		fields := strings.Fields(meta)
		if !ok || len(fields) < 4 {
			continue
		}
		path = strings.TrimPrefix(path, prefix)
		if h.blobs[path] == nil {
			h.blobs[path] = map[string]bool{}
		}
		h.blobs[path][fields[3]] = true
	}
}

// gitBlobID is what git calls a file's content: sha1("blob <size>\0" + data).
func gitBlobID(data []byte) string {
	sum := sha1.Sum(append([]byte("blob "+strconv.Itoa(len(data))+"\x00"), data...))
	return hex.EncodeToString(sum[:])
}

func (h *gitHistory) contains(source, lessonID, language, rel string, disk []byte) bool {
	if !h.loaded {
		h.load()
	}
	ids := []string{gitBlobID(disk)}
	if bytes.Contains(disk, []byte("\r\n")) {
		// Git may have converted line endings on checkout (Windows).
		ids = append(ids, gitBlobID(bytes.ReplaceAll(disk, []byte("\r\n"), []byte("\n"))))
	}
	for _, dir := range h.c.sourceDirs(source, lessonID, language) {
		path, err := filepath.Rel(h.c.root, filepath.Join(dir, filepath.FromSlash(rel)))
		if err != nil {
			continue
		}
		for _, id := range ids {
			if h.blobs[filepath.ToSlash(path)][id] {
				return true
			}
		}
	}
	return false
}
