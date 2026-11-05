# Hints

## The skeleton

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

## Args and Rest

`strings.Fields("")` returns an empty slice, while `Args` may be `nil` –
`slices.Equal` treats nil and empty as equal, so it is fine:

```go
if !slices.Equal(c.Args, strings.Fields(c.Rest)) {
	t.Errorf("Args %q are not the words of Rest %q", c.Args, c.Rest)
}
```

## The fix

```go
verb, rest := line, ""
if i := strings.IndexFunc(line, unicode.IsSpace); i >= 0 {
	verb, rest = line[:i], line[i:]
}
```

`rest` starts with the whitespace; the `strings.TrimSpace(rest)` that is
already there takes care of it.

## "fuzzing is not supported"

Fuzzing works on Linux, macOS and Windows on amd64 and arm64. If `-race` is
also given it is slower; leave it out while fuzzing.
