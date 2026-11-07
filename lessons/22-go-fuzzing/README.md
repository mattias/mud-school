# Lesson 22 – Testing for real

`ParseCommand` has been with you since lesson 4. It has tests. It works. Or…?

Type `look`, a **tab** and `sword` in `./mudschool connect` (or telnet). The server answers that it
doesn't know the command `look	sword` – the tab became part of the verb. None
of the tests you have seen caught that, because nobody thought of writing that
example.

That is the limit of example-based tests: they only test what you think of.
In this lesson you let the computer think of the examples instead.

- **Fuzzing**: Go generates millions of inputs, guided by which code paths
  they reach, and saves every input that breaks a rule.
- **Property-based tests**: instead of "this input gives that output" you
  write rules that hold for *every* input.

## Goals

- Write a fuzz test: `func FuzzXxx(f *testing.F)`, seeds with `f.Add`, the test with `f.Fuzz`.
- Run the fuzzer: `go test -fuzz=...` and read what it found.
- Think in **invariants** (properties): what must always be true?
- Understand `testdata/fuzz/`: inputs that once failed become regression tests forever.
- Fix the bug with `unicode.IsSpace` – whitespace is more than `' '`.

## Go concepts

### A fuzz test

```go
func FuzzParseCommand(f *testing.F) {
	f.Add("look")                 // seeds: ordinary examples to start from
	f.Add("Get Sword from chest")
	f.Fuzz(func(t *testing.T, line string) {
		c := ParseCommand(line)
		if strings.IndexFunc(c.Verb, unicode.IsSpace) >= 0 {
			t.Errorf("the verb %q contains whitespace", c.Verb)
		}
		// ... more rules ...
	})
}
```

It lives in an ordinary `_test.go` file. A plain `go test` runs only the
seeds – so it is also a normal test. With `-fuzz` Go starts mutating:

```
cd workspace/server
go test -fuzz=FuzzParseCommand -fuzztime=30s ./mud/
```

When it finds something, it **shrinks** the input to the smallest one that
still fails and saves it in `mud/testdata/fuzz/FuzzParseCommand/`. Commit that
folder: from now on every `go test` runs that input too.

### Which rules?

A good invariant is something you are *sure* of without knowing the input:

- the verb never contains whitespace;
- `Rest` has no whitespace at the ends;
- `Args` are exactly the words of `Rest` (`strings.Fields(Rest)`);
- the verb is empty if and only if the line is blank;
- the verb is lower case (for valid UTF-8).

Notice what is *not* on the list: "the verb is a known command". The parser
doesn't know the commands – that is the server's business.

### Whitespace

```go
unicode.IsSpace(' ')      // true
unicode.IsSpace('\t')     // true
unicode.IsSpace(' ') // true: a no-break space
unicode.IsSpace('　') // true: the ideographic space a Japanese keyboard types
```

`strings.Fields` and `strings.TrimSpace` already use `unicode.IsSpace`.
`strings.Cut(line, " ")` does not. `strings.IndexFunc(line, unicode.IsSpace)`
finds the first whitespace of any kind.

## The contract

- Your own `server/mud/command_fuzz_test.go` with `FuzzParseCommand` – seeds
  (`f.Add`) and `f.Fuzz` with the rules above. (Not in a `mudschool_` file:
  those belong to the course and are replaced.)
- `ParseCommand` splits the verb at the first whitespace of *any* kind.

The course's test checks that your fuzz test exists, tries some tricky lines,
and runs its own small property-based test with 5000 random lines.

## Task

1. Write `FuzzParseCommand` with a few ordinary seeds (*not* the tab – let
   the fuzzer find the bug itself).
2. Run the fuzzer. How long did it take? What did the shrunk input look like?
3. Look in `testdata/fuzz/FuzzParseCommand/`.
4. Fix `ParseCommand`. Run the fuzzer again for a minute.

## Think about

- Your fuzz test found a bug in seconds that the tests from lesson 4 never
  found. Does that mean the example tests were useless?
- What else in the server takes input from strangers? (`ValidateName`, the
  world loader, the JSON from a graphical client …) Which rules would you
  fuzz them with?
- Why is fuzzing especially important for a program on the internet?
- Fuzzing finds crashes and broken rules – but can it tell you that `look`
  shows the *right* room?

## Challenges

- Fuzz `ValidateName` and `NormalizeName`: is `NormalizeName(NormalizeName(x)) == NormalizeName(x)`?
- Fuzz `LoadWorld` with random JSON: it may return an error, but never panic.
- In Rust: add `proptest` as a dev-dependency and write a property test for
  `parse_line` – any string becomes *some* `ServerMsg`, without panicking.
- A bot: a little Go program that logs in and walks around at random. Run ten
  of them against your server with `-race` for a minute.
