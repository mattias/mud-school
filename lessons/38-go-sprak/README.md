# Lesson 38 – Languages per player

The course speaks three languages, but your server speaks one: the language
it was built in. Once it is on the web (lesson 42) a Japanese and a Swedish
player may meet in the same room. In this lesson each player chooses a
language, and the world's texts follow:

```
> language
Your language: en. There are: en, ja, sv. Choose with e.g. language en
> language sv
Language: sv.
Begynnelsens sal
En stilla sal med vitkalkade väggar. ...
[Tutorial 1/6] Välkommen, Anna! Den här korta introduktionen lär dig grunderna ...
```

What is translated is the **world**: the tutorial, and rooms that have
translations. The server's own messages (`messages.go`) stay in the language
it was built in – translating those is a bigger job (see the challenges).

## Goals

- Let a value in JSON have **two shapes**: a string, or an object with one
  string per language – with your own `UnmarshalJSON`.
- Fall back sensibly: the player's language, the world's, the plain text, anything.
- Fill in texts with `text/template`: `Welcome, {{.Name}}!`.
- Keep the old file format working: what was a string still is.

## Go concepts

### `json.Unmarshaler`

```go
type Text map[string]string

func (t *Text) UnmarshalJSON(data []byte) error {
	var plain string
	if err := json.Unmarshal(data, &plain); err == nil {
		*t = Text{"": plain}
		return nil
	}
	...
}
```

`encoding/json` checks whether a type has an `UnmarshalJSON` method – an
interface, `json.Unmarshaler` – and lets it decode itself. A plain string is
kept under the key `""`, so `"text": "Type look."` still works, and
`"text": {"en": "Type look.", "sv": "Skriv look."}` works too.
`MarshalJSON` does the opposite, so saving the world (lesson 36) writes a
plain string back as a string.

### A method on a map type

`Text` is a map with methods. `t.In("sv", "en")` looks for `sv`, then the
world's language, then `""`, and finally the first language in alphabetical
order – a text in some language is better than none.

### `text/template`

```go
tmpl, err := template.New("text").Option("missingkey=error").Parse(s)
...
tmpl.Execute(&b, TextData{Name: "Anna", Step: 2, Steps: 6})
```

Templates are checked when the world is loaded – running them once with an
empty `TextData` finds `{{.Nmae}}` and broken braces – so a typo is an error
in the file, not a surprise for a player.

## The contract

- `Text` (`map[string]string`) with `UnmarshalJSON` (a string → `{"": s}`; an
  object; otherwise an error with `MsgErrText`), `MarshalJSON`, `In(lang, fallback)`,
  `Render(lang, fallback, TextData)` and `check()` (`MsgErrTemplate` with the
  language). `TextData{Name, Step, Steps}`.
- `TutorialStep.Text` and `Tutorial.Done` are `Text`. A text that is not a valid
  template is a load error: `MsgErrTutorialText` with the step number
  (`len(Steps)+1` for `done`). A step whose text is empty in every language
  (`""`, `{"en": ""}`, `null`) has no text: `MsgErrTutorialStep`, as in lesson 37.
- A world file may have `"language"` (`World.Language`: the language of its
  plain texts), and rooms may have `"translations": {"sv": {"name": ..., "description": ...}}`
  (`Room.Translations`, `RoomText`).
- `(w *World) Languages()` – every language the world has texts in, sorted.
- `language` shows `MsgLanguage` (the player's language, or the world's, and the
  list); `language xx` sets it (`MsgLanguageSet`, then a look and the tutorial
  step again), or `MsgLanguageUnknown`.
- The room description and the status frame's room name use the player's
  language; tutorial texts are rendered with the player's name and step.
- `SavedPlayer.Language`, saved through migration 5.
- `DefaultWorld` uses `tutorial.all.json` (given: the tutorial in sv, en and ja)
  and sets `Language` to `CourseLanguage` (in `messages.go`).
- `worldSnapshot` keeps `Language`; translations and `Text` are saved as they were read.

## Task

1. `i18n.go` is partly given (the types, `empty`, `sortedKeys`): write the other `Text`
   methods, `Languages`, `roomText`, `render` and `cmdLanguage`.
2. `tutorial.go`: `Text` instead of `string`, the template check, and
   `s.render` where the texts are shown.
3. `world.go`: `Room.Translations`, `World.Language`, `worldFile.Language`.
   `rooms.go`: `roomText` in `describeRoom` and `sendStatus`.
4. `player.go`: `lang string`. `store.go`: migration 5 and `SavedPlayer.Language`.
   `account.go`: snapshot and restore it.
5. `defaultworld.go`: `tutorial.all.json` and `w.Language = CourseLanguage`.
   `commands.go`: `language`. `olc.go`: `Language` in `worldSnapshot`.
   `copyover.go`: `Lang` in `CopyoverPlayer` (JSON `lang`), `PrepareCopyover` and `adopt`.

## Think about

- `@desc` (lesson 36) changes a room's own description. What happens to its
  translations? What should?
- Why is a missing template field an error when the world loads, but a
  failing template at runtime is shown as plain text?
- The web client (lesson 39) could send `language` by itself, from the
  browser's language. Where would that go?
- Which is easier to translate: a sentence put together with `%s`, or a template?

## Challenges

- Translate the server's messages: a catalog per language (the three
  `messages.go` are a good start) and `p.Msg("MsgWelcomeName", name)`.
- Items and creatures with translated names.
- `@translate sv <text>`: builders translate the room they stand in.
