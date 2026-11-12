# Hints

## UnmarshalJSON

Try the shapes in order; `*t = ...` replaces the map the pointer points at:

```go
var m map[string]string
if err := json.Unmarshal(data, &m); err != nil {
	return errors.New(MsgErrText)
}
*t = m
```

Don't call `json.Unmarshal(data, t)` with `t` itself – that would call your
own `UnmarshalJSON` again, forever.

## MarshalJSON

`json.Marshal(map[string]string(t))` – the conversion to a plain map type
avoids the same trap the other way round.

## In

```go
for _, l := range []string{lang, fallback, ""} {
	if s, ok := t[l]; ok {
		return s
	}
}
```

then `sortedKeys(t)` and the first one (or `""` for an empty text).

## check

`template.New("text").Option("missingkey=error").Parse(...)`, then
`Execute(new(bytes.Buffer), TextData{})` – parsing finds broken braces,
executing finds fields that don't exist.

## The tutorial

`len(st.Text) == 0` replaces `st.Text == ""`. In `showTutorialStep`:

```go
text := s.render(p, t.Steps[p.tutorial-1].Text, TextData{Step: p.tutorial, Steps: len(t.Steps)})
```

## Migration 5

```go
`ALTER TABLE players ADD COLUMN language TEXT NOT NULL DEFAULT ''`,
```
