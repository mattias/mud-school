# レッスン 22 – 本気のテスト

`ParseCommand` はレッスン 4 からずっと一緒です。テストもあります。動いています。本当に…？

telnet で `look`、**タブ**、`sword` と入力してみましょう。サーバーは `look	sword` という
コマンドは知らないと答えます。タブが動詞の一部になってしまったのです。これまでのどのテストも
これを見つけられませんでした。その例を書こうと誰も思いつかなかったからです。

それが例に基づくテストの限界です。思いついたことしかテストできません。
このレッスンでは、例を考えるのをコンピューターに任せます。

- **ファジング**: Go が到達したコード経路を手がかりに何百万もの入力を生成し、ルールを破った
  入力をすべて保存する。
- **プロパティベーステスト**: 「この入力ならこの出力」ではなく、*あらゆる* 入力で成り立つ
  ルールを書く。

## 目標

- ファズテストを書く: `func FuzzXxx(f *testing.F)`、`f.Add` でシード、`f.Fuzz` でテスト本体。
- ファザーを実行する: `go test -fuzz=...` と、見つかったものの読み方。
- **不変条件**（プロパティ）で考える: 常に成り立たなければならないことは？
- `testdata/fuzz/` を理解する: 一度失敗した入力は永遠に回帰テストになる。
- `unicode.IsSpace` でバグを直す。空白文字は `' '` だけではない。

## Go の概念

### ファズテスト

```go
func FuzzParseCommand(f *testing.F) {
	f.Add("look")                 // シード: 出発点となる普通の例
	f.Add("Get Sword from chest")
	f.Fuzz(func(t *testing.T, line string) {
		c := ParseCommand(line)
		if strings.IndexFunc(c.Verb, unicode.IsSpace) >= 0 {
			t.Errorf("the verb %q contains whitespace", c.Verb)
		}
		// ... さらにルール ...
	})
}
```

普通の `_test.go` ファイルに書きます。ふつうの `go test` はシードだけを実行するので、
これは通常のテストでもあります。`-fuzz` を付けると Go が入力の変異を始めます:

```
cd workspace/server
go test -fuzz=FuzzParseCommand -fuzztime=30s ./mud/
```

何かを見つけると、まだ失敗する最小の入力まで **縮小** し、
`mud/testdata/fuzz/FuzzParseCommand/` に保存します。このフォルダーをコミットしましょう。
以後は毎回の `go test` がその入力も実行します。

### どんなルール？

良い不変条件とは、入力を知らなくても *確信* できることです:

- 動詞に空白文字は決して含まれない。
- `Rest` の両端に空白文字はない。
- `Args` はちょうど `Rest` の単語（`strings.Fields(Rest)`）である。
- 動詞が空になるのは、行が空白だけのときに限る。
- 動詞は小文字である（正しい UTF-8 の場合）。

リストに *ない* ものに注目してください。「動詞は既知のコマンドである」。パーサーはコマンドを
知りません。それはサーバーの仕事です。

### 空白文字

```go
unicode.IsSpace(' ')      // true
unicode.IsSpace('\t')     // true
unicode.IsSpace(' ') // true: ノーブレークスペース
unicode.IsSpace('　') // true: 日本語キーボードで入力される全角スペース
```

`strings.Fields` と `strings.TrimSpace` はすでに `unicode.IsSpace` を使っています。
`strings.Cut(line, " ")` は使いません。`strings.IndexFunc(line, unicode.IsSpace)` なら
あらゆる種類の最初の空白文字を見つけられます。

## 契約

- 自分で書く `server/mud/command_fuzz_test.go` に `FuzzParseCommand` を置く。シード
  （`f.Add`）と、上のルールを確かめる `f.Fuzz`。（`mudschool_` のファイルには書かないこと。
  それらはコースのもので、置き換えられます。）
- `ParseCommand` は *あらゆる* 種類の最初の空白文字で動詞を切り分ける。

コースのテストは、あなたのファズテストがあることを確かめ、いくつかの厄介な行を試し、
5000 行のランダムな行で独自の小さなプロパティベーステストを実行します。

## 課題

1. 普通のシードをいくつか入れて `FuzzParseCommand` を書く（タブは *入れない*。バグは
   ファザー自身に見つけさせる）。
2. ファザーを実行する。どれくらいかかった？ 縮小された入力はどんな形だった？
3. `testdata/fuzz/FuzzParseCommand/` を見てみる。
4. `ParseCommand` を直す。もう一度 1 分間ファザーを回す。

## 考えてみよう

- あなたのファズテストは、レッスン 4 のテストが決して見つけなかったバグを数秒で見つけました。
  では例に基づくテストは無意味だった？
- サーバーで見知らぬ人からの入力を受け取るところは他にどこがある？（`ValidateName`、
  ワールドの読み込み、グラフィカルクライアントからの JSON……）どんなルールでファジングする？
- インターネット上のプログラムにとって、ファジングが特に大切なのはなぜ？
- ファジングはクラッシュや破られたルールを見つけます。では `look` が *正しい* 部屋を
  表示することまで分かる？

## チャレンジ

- `ValidateName` と `NormalizeName` をファジングする: `NormalizeName(NormalizeName(x)) == NormalizeName(x)` は成り立つ？
- ランダムな JSON で `LoadWorld` をファジングする: エラーを返すのはよいが、panic してはいけない。
- Rust で: dev-dependency に `proptest` を加え、`parse_line` のプロパティテストを書く。
  どんな文字列も panic せずに *何らかの* `ServerMsg` になる。
- ボット: ログインしてランダムに歩き回る小さな Go プログラム。10 体を `-race` 付きの
  サーバーに 1 分間ぶつけてみる。
