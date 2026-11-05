package mud

import (
	"go/ast"
	"go/parser"
	"go/token"
	"math/rand/v2"
	"os"
	"slices"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"
)

// TestYouWroteAFuzzTest looks for FuzzParseCommand in YOUR test files (not the
// course's mudschool_* files) and checks that it has seeds and calls f.Fuzz.
func TestYouWroteAFuzzTest(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	for _, e := range entries {
		name := e.Name()
		if !strings.HasSuffix(name, "_test.go") || strings.HasPrefix(name, "mudschool_") {
			continue
		}
		file, err := parser.ParseFile(fset, name, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Name.Name != "FuzzParseCommand" {
				continue
			}
			calls := map[string]int{}
			ast.Inspect(fn.Body, func(n ast.Node) bool {
				if call, ok := n.(*ast.CallExpr); ok {
					if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
						calls[sel.Sel.Name]++
					}
				}
				return true
			})
			if calls["Add"] == 0 {
				t.Errorf(L("%s: FuzzParseCommand har inga frön – lägg till några med f.Add(...)",
					"%s: FuzzParseCommand has no seeds – add some with f.Add(...)",
					"%s: FuzzParseCommand にシードがありません。f.Add(...) で追加してください"), name)
			}
			if calls["Fuzz"] == 0 {
				t.Errorf(L("%s: FuzzParseCommand anropar aldrig f.Fuzz(func(t *testing.T, line string) { ... })",
					"%s: FuzzParseCommand never calls f.Fuzz(func(t *testing.T, line string) { ... })",
					"%s: FuzzParseCommand が f.Fuzz(func(t *testing.T, line string) { ... }) を呼んでいません"), name)
			}
			return
		}
	}
	t.Fatal(L("hittar ingen func FuzzParseCommand(f *testing.F) i dina testfiler. Skriv den i server/mud/command_fuzz_test.go (inte i en mudschool_-fil – de skrivs över).",
		"found no func FuzzParseCommand(f *testing.F) in your test files. Write it in server/mud/command_fuzz_test.go (not in a mudschool_ file – those are overwritten).",
		"あなたのテストファイルに func FuzzParseCommand(f *testing.F) が見つかりません。server/mud/command_fuzz_test.go に書いてください（mudschool_ のファイルは上書きされるので不可）。"))
}

func TestParseCommandAnyWhitespace(t *testing.T) {
	tests := []struct {
		line, verb string
		args       []string
	}{
		{"look\tsword", "look", []string{"sword"}},
		{"\tn", "north", nil},
		{"kill\t\trat ", "kill", []string{"rat"}},
		{"say hi", "say", []string{"hi"}},     // a no-break space
		{"get　sword", "get", []string{"sword"}}, // a Japanese (ideographic) space
	}
	for _, tt := range tests {
		got := ParseCommand(tt.line)
		if got.Verb != tt.verb || !slices.Equal(got.Args, tt.args) {
			t.Errorf(L("ParseCommand(%q) = %+v, väntade verbet %q och argumenten %q",
				"ParseCommand(%q) = %+v, expected the verb %q and the args %q",
				"ParseCommand(%q) = %+v。動詞 %q と引数 %q を期待しました"), tt.line, got, tt.verb, tt.args)
		}
	}
}

// TestParseCommandProperties is a small property-based test: thousands of
// random lines made of nasty pieces, and rules that must hold for all of them.
func TestParseCommandProperties(t *testing.T) {
	pieces := []string{"a", "Look", "N", "say", "'", " ", "  ", "\t", " ", "　", "\r", "é", "Å", "語", "\xff", "x\ty"}
	rng := rand.New(rand.NewPCG(22, 22))
	for i := 0; i < 5000; i++ {
		var b strings.Builder
		for n := rng.IntN(8); n > 0; n-- {
			b.WriteString(pieces[rng.IntN(len(pieces))])
		}
		line := b.String()
		c := ParseCommand(line)
		fail := func(why string) {
			t.Fatalf("ParseCommand(%q) = %+v: %s", line, c, why)
		}
		if strings.IndexFunc(c.Verb, unicode.IsSpace) >= 0 {
			fail(L("verbet innehåller blanktecken", "the verb contains whitespace", "動詞に空白文字が含まれています"))
		}
		if c.Rest != strings.TrimSpace(c.Rest) {
			fail(L("Rest är inte trimmad", "Rest is not trimmed", "Rest の前後の空白が除かれていません"))
		}
		if !slices.Equal(c.Args, strings.Fields(c.Rest)) {
			fail(L("Args är inte orden i Rest", "Args are not the words of Rest", "Args が Rest の単語になっていません"))
		}
		if (c.Verb == "") != (strings.TrimSpace(line) == "") {
			fail(L("verbet ska vara tomt om och endast om raden är tom", "the verb should be empty if and only if the line is blank", "動詞が空になるのは行が空白だけのときに限ります"))
		}
		if utf8.ValidString(line) && strings.ToLower(c.Verb) != c.Verb {
			fail(L("verbet är inte gemener", "the verb is not lower case", "動詞が小文字ではありません"))
		}
	}
}
