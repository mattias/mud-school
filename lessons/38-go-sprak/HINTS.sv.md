# Tips

## UnmarshalJSON

Prova formerna i tur och ordning; `*t = ...` ersätter mappen som pekaren pekar på:

```go
var m map[string]string
if err := json.Unmarshal(data, &m); err != nil {
	return errors.New(MsgErrText)
}
*t = m
```

Anropa inte `json.Unmarshal(data, t)` med `t` själv – det skulle anropa din
egen `UnmarshalJSON` igen, för evigt.

## MarshalJSON

`json.Marshal(map[string]string(t))` – konverteringen till en vanlig map-typ
undviker samma fälla åt andra hållet.

## In

```go
for _, l := range []string{lang, fallback, ""} {
	if s, ok := t[l]; ok {
		return s
	}
}
```

och sedan `sortedKeys(t)` och den första (eller `""` för en tom text).

## check

`template.New("text").Option("missingkey=error").Parse(...)`, sedan
`Execute(new(bytes.Buffer), TextData{})` – parsningen hittar trasiga klamrar,
körningen hittar fält som inte finns.

## Introduktionen

`st.Text.empty()` (given i `i18n.go`) ersätter `st.Text == ""`. I `showTutorialStep`:

```go
text := s.render(p, t.Steps[p.tutorial-1].Text, TextData{Step: p.tutorial, Steps: len(t.Steps)})
```

## Migrering 5

```go
`ALTER TABLE players ADD COLUMN language TEXT NOT NULL DEFAULT ''`,
```
