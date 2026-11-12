# レッスン 38 – プレイヤーごとの言語

このコースは 3 つの言語を話しますが、あなたのサーバーが話すのは 1 つ、ビルドされたときの
言語だけです。Web に出たら（レッスン 42）、日本のプレイヤーとスウェーデンのプレイヤーが
同じ部屋で会うかもしれません。このレッスンでは、プレイヤーがそれぞれ言語を選び、
世界のテキストがそれに従います:

```
> language
あなたの言語: ja。あるのは: en, ja, sv。例えば language ja で選べます
> language en
言語: en。
The Hall of Beginnings
A quiet hall with whitewashed walls. ...
[チュートリアル 1/6] Welcome, Anna! This short tutorial teaches you the basics ...
```

訳されるのは **世界** です。チュートリアルと、翻訳のある部屋。サーバー自身のメッセージ
（`messages.go`）はビルドされた言語のままです。それを訳すのはもっと大きな仕事です
（チャレンジを見てください）。

## 目標

- JSON の値に **2 つの形** を持たせる: 文字列か、言語ごとに 1 つの文字列を持つ
  オブジェクト。自分の `UnmarshalJSON` で。
- うまく代わりを選ぶ: プレイヤーの言語、世界の言語、普通のテキスト、とにかく何か。
- `text/template` でテキストを埋める: `ようこそ、{{.Name}}！`。
- 古いファイル形式も動かす: 文字列だったものは今も文字列でいい。

## Go の概念

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

`encoding/json` は型に `UnmarshalJSON` メソッド（インターフェース `json.Unmarshaler`）が
あるかを確かめ、あれば自分で読ませます。普通の文字列はキー `""` に入るので、
`"text": "look と打ちましょう。"` はそのまま動き、`"text": {"en": "Type look.", "ja": "look と打ちましょう。"}`
も動きます。`MarshalJSON` はその逆で、世界を保存するとき（レッスン 36）、普通の文字列は
文字列として書き戻されます。

### map 型のメソッド

`Text` はメソッドを持つ map です。`t.In("ja", "en")` は `ja`、次に世界の言語、次に `""`、
最後にアルファベット順で最初の言語を探します。何かの言語のテキストは、ないよりましです。

### `text/template`

```go
tmpl, err := template.New("text").Option("missingkey=error").Parse(s)
...
tmpl.Execute(&b, TextData{Name: "Anna", Step: 2, Steps: 6})
```

テンプレートは世界を読み込むときに確かめます。空の `TextData` で一度実行すると、
`{{.Nmae}}` や壊れた括弧が見つかります。だから打ち間違いはプレイヤーへの驚きではなく、
ファイルのエラーになります。

## 契約

- `Text`（`map[string]string`）と `UnmarshalJSON`（文字列 → `{"": s}`、オブジェクト、
  それ以外は `MsgErrText` のエラー）、`MarshalJSON`、`In(lang, fallback)`、
  `Render(lang, fallback, TextData)`、`check()`（言語付きの `MsgErrTemplate`）。
  `TextData{Name, Step, Steps}`。
- `TutorialStep.Text` と `Tutorial.Done` は `Text`。正しいテンプレートでないテキストは
  読み込みエラー: ステップ番号付きの `MsgErrTutorialText`（`done` は `len(Steps)+1`）。
- ワールドファイルには `"language"`（`World.Language`: 普通のテキストの言語）を置け、
  部屋には `"translations": {"ja": {"name": ..., "description": ...}}`
  （`Room.Translations`、`RoomText`）を置ける。
- `(w *World) Languages()` – 世界がテキストを持つすべての言語、ソート済み。
- `language` は `MsgLanguage`（プレイヤーの言語か世界の言語、と一覧）を見せ、
  `language xx` でそれを選ぶ（`MsgLanguageSet`、続けて look とチュートリアルのステップ）か、
  `MsgLanguageUnknown`。
- 部屋の説明と status フレームの部屋名はプレイヤーの言語を使い、チュートリアルの
  テキストはプレイヤーの名前とステップで描画される。
- `SavedPlayer.Language`、マイグレーション 5 で保存。
- `DefaultWorld` は `tutorial.all.json`（用意済み: sv・en・ja のチュートリアル）を使い、
  `Language` を `CourseLanguage`（`messages.go` にある）にする。
- `worldSnapshot` は `Language` を残し、翻訳と `Text` は読んだとおりに保存される。

## 課題

1. `i18n.go` は一部用意されています（型、`sortedKeys`）。`Text` のメソッド、`Languages`、
   `roomText`、`render`、`cmdLanguage` を書きましょう。
2. `tutorial.go`: `string` の代わりに `Text`、テンプレートの確認、テキストを見せる所で `s.render`。
3. `world.go`: `Room.Translations`、`World.Language`、`worldFile.Language`。
   `rooms.go`: `describeRoom` と `sendStatus` で `roomText`。
4. `player.go`: `lang string`。`store.go`: マイグレーション 5 と `SavedPlayer.Language`。
   `account.go`: 保存と復元。
5. `defaultworld.go`: `tutorial.all.json` と `w.Language = CourseLanguage`。
   `commands.go`: `language`。`olc.go`: `worldSnapshot` に `Language`。

## 考えてみよう

- `@desc`（レッスン 36）は部屋自身の説明を変えます。翻訳はどうなる？ どうなるべき？
- テンプレートにないフィールドが世界の読み込み時にはエラーなのに、実行時に失敗した
  テンプレートは普通のテキストとして見せるのはなぜ？
- Web クライアント（レッスン 39）はブラウザの言語から自分で `language` を送れます。
  それはどこに入れる？
- 訳しやすいのはどちら: `%s` で組み立てた文、それともテンプレート？

## チャレンジ

- サーバーのメッセージを訳す: 言語ごとのカタログ（3 つの `messages.go` が良い出発点）と
  `p.Msg("MsgWelcomeName", name)`。
- 物やクリーチャーの名前を訳す。
- `@translate ja <テキスト>`: ビルダーが今いる部屋を訳す。
