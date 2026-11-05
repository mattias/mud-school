# Ledtrådar

## Skelettet

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

## Args och Rest

`strings.Fields("")` returnerar en tom slice, medan `Args` kan vara `nil` –
`slices.Equal` ser nil och tom som lika, så det går bra:

```go
if !slices.Equal(c.Args, strings.Fields(c.Rest)) {
	t.Errorf("Args %q are not the words of Rest %q", c.Args, c.Rest)
}
```

## Lagningen

```go
verb, rest := line, ""
if i := strings.IndexFunc(line, unicode.IsSpace); i >= 0 {
	verb, rest = line[:i], line[i:]
}
```

`rest` börjar med blanktecknet; `strings.TrimSpace(rest)` som redan finns
där tar hand om det.

## "fuzzing is not supported"

Fuzzing fungerar på Linux, macOS och Windows på amd64 och arm64. Med `-race`
blir det långsammare; strunta i den medan du fuzzar.
