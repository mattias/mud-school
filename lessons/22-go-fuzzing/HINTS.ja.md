# ヒント

## 骨組み

```go
package mud

import (
	"slices"
	"strings"
	"testing"
	"unicode"
)

func FuzzParseCommand(f *testing.F) {
	for _, seed := range []string{"", "look", "n", "Get Sword from chest", "'hello there"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, line string) {
		c := ParseCommand(line)
		// the rules
	})
}
```

## Args と Rest

`strings.Fields("")` は空のスライスを返し、`Args` は `nil` のこともあります。
`slices.Equal` は nil と空を等しいとみなすので問題ありません:

```go
if !slices.Equal(c.Args, strings.Fields(c.Rest)) {
	t.Errorf("Args %q are not the words of Rest %q", c.Args, c.Rest)
}
```

## 修正

```go
verb, rest := line, ""
if i := strings.IndexFunc(line, unicode.IsSpace); i >= 0 {
	verb, rest = line[:i], line[i:]
}
```

`rest` は空白文字で始まりますが、すでにある `strings.TrimSpace(rest)` が処理してくれます。

## "fuzzing is not supported"

ファジングは amd64 と arm64 の Linux、macOS、Windows で動きます。`-race` を付けると
遅くなるので、ファジング中は外しておきましょう。
