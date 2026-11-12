# ヒント

## UnmarshalJSON

形を順に試します。`*t = ...` はポインターの指す map を置き換えます:

```go
var m map[string]string
if err := json.Unmarshal(data, &m); err != nil {
	return errors.New(MsgErrText)
}
*t = m
```

`t` そのものを渡して `json.Unmarshal(data, t)` を呼ばないこと。自分の
`UnmarshalJSON` がまた呼ばれ、永遠に続きます。

## MarshalJSON

`json.Marshal(map[string]string(t))`。普通の map 型に変換することで、逆向きの同じ罠を避けます。

## In

```go
for _, l := range []string{lang, fallback, ""} {
	if s, ok := t[l]; ok {
		return s
	}
}
```

その後は `sortedKeys(t)` の最初のもの（空のテキストなら `""`）。

## check

`template.New("text").Option("missingkey=error").Parse(...)`、続けて
`Execute(new(bytes.Buffer), TextData{})`。解析で壊れた括弧が、実行で存在しない
フィールドが見つかります。

## チュートリアル

`st.Text == ""` の代わりに `len(st.Text) == 0`。`showTutorialStep` では:

```go
text := s.render(p, t.Steps[p.tutorial-1].Text, TextData{Step: p.tutorial, Steps: len(t.Steps)})
```

## マイグレーション 5

```go
`ALTER TABLE players ADD COLUMN language TEXT NOT NULL DEFAULT ''`,
```
