# 講師・コース作成者向け

> [🇬🇧 English](../en/TEACHERS.md) · [🇸🇪 Svenska](../sv/TEACHERS.md) · 🇯🇵 日本語

コースの構成と、レッスンの追加・変更方法について。

## ディレクトリ構成

```
mudschool                    コースツールをビルドして実行するスクリプト
tools/mudschool/             コースツール（Go、標準ライブラリのみ）
starter/                     受講者が最初に持つコード（workspace/ にコピーされる）
  server/                    Go モジュール "mudserver"
  client/                    Cargo パッケージ "mud-client"（ライブラリ名は mud_client）
starter.<言語>/              言語ごとのスターターファイル: messages.go と messages.rs
lessons/
  _shared/go/...             すべての Go レッスンでコピーされるファイル（テスト補助）
  _shared/rust/...           Rust 版（tests/mudschool_common/）
  NN-名前/
    lesson.json              各言語のタイトルと要約、トラック（"go"/"rust"）
    README.md                英語の説明（README.sv.md、README.ja.md）
    HINTS.md                 ヒント。"## " 見出しごとに 1 つ（HINTS.sv.md、HINTS.ja.md）
    scaffold/                レッスン開始時に受講者が受け取る新しいファイル
    scaffold.<言語>/         言語ごとの新しいファイル（世界データなど）
    tests/                   レッスン開始時に配置されるテスト（`check` でも再配置）
    solution/                模範解答: レッスン終了「後」のファイル
    solution.<言語>/         模範解答のうち言語に依存する部分（まれ）
workspace/                   受講者のコード（git では無視）
docs/<言語>/                 カリキュラム、プロトコル、環境構築、アイデア、この文書
```

`scaffold/`、`tests/`、`solution/` の下のパスはすべて `workspace/` からの相対パスです。
例: `tests/server/mud/mudschool_04_test.go`。

## 言語

コースはスウェーデン語（`sv`）、英語（`en`）、日本語（`ja`）で提供しています。

| 対象 | 翻訳の方法 |
|---|---|
| 説明とヒント | `README.<言語>.md`、`HINTS.<言語>.md`（標準の英語は `README.md`） |
| タイトルと要約 | `lesson.json` のオブジェクト: `{"sv": ..., "en": ..., "ja": ...}` |
| コースツールのテキスト | `tools/mudschool` の `L(sv, en, ja)` |
| ゲーム内のテキスト | `starter.<言語>/server/mud/messages.go` と `.../client/src/messages.rs` |
| ゲームの世界 | レッスン 08 と 09（部屋）、15（`items.json`）、16（`npcs.json`）の `scaffold.<言語>/` |
| テストの失敗メッセージ | `L(sv, en, ja)`（Go）と `l(sv, en, ja)`（Rust） |
| ドキュメント | `docs/<言語>/` |
| コードとコメント | **英語** – 全員共通 |

言語は `./mudschool init --lang en|sv|ja` で選び（指定しなければ `$LANG` から推測）、
`./mudschool lang <言語>` で変更できます。`workspace/.mudschool.json` に保存され、
テストには `MUDSCHOOL_LANG` として渡されます。

テストは決してハードコードされたテキストと比較しません。受講者のコードと同じ
定数（`MsgAskName`、`fmt.Sprintf(MsgSays, "Anna", "hi")`）を使うので、どの言語でも
動きます。受講者がテキストを変えるのも自由です。

受講者が `./mudschool lang` で言語を切り替えると、変更済みの
`messages.go`/`messages.rs` はまず `.bak` として保存されます。

## コースツールの動作

- **`init`** は `starter/` と `starter.<言語>/` を `workspace/` にコピーし、最初の
  レッスンを開始します。
- **レッスンを開始する** と `scaffold/` と `scaffold.<言語>/` をコピーしますが、
  まだ存在しないファイルだけです。受講者のコードは決して上書きされません。
  README から参照できるよう、レッスンのテストもこの時点で配置されます。
  レッスンを解くまでエディタは新しいテストに（未定義の名前として）エラーを表示しますが、
  想定どおりです。`run server`/`run client` には影響しません。
- **`check`**:
  1. `workspace/` 内の、名前が `mudschool_` で始まるファイルとディレクトリ
     （古いコースのテスト）をすべて削除する、
  2. `lessons/_shared/<トラック>/` とレッスンの `tests/` をコピーする、
  3. `workspace/server` で `go vet ./... && go test -count=1 ./...`（Go）、
     または `workspace/client` で `cargo test`（Rust）を実行する、
  4. 合格したら、そのレッスンの `TODO(lesson NN)` の残りを一覧にし、完了にして次のレッスンを開始する。
- **`doctor`** は Go、Rust、各種ツールがあるかを確認します。

実行されるのは現在のレッスンのテストだけです。これは意図的です。レッスンごとに
振る舞いが変わるからです（レッスン 01 のエコーはレッスン 04 でなくなります）。

### テストの命名規則

テストファイルは **必ず** `mudschool_` で始めてください。そうしないと削除されず、
古いテストが失敗し始めます。Go: パッケージ `mud` 内の `mudschool_NN_test.go`。
Rust: `tests/mudschool_NN.rs`（ライブラリ `mud_client` に対する結合テスト）。
`l(...)` を使うために `mod mudschool_common;` で始めます。

## 受講者の作業領域の更新

ツールは `starter/` と `scaffold/` からコピーしたファイルごとにチェックサムを記録します
（`.mudschool.json` の `files`）。`./mudschool update` はそれを比較し、変更されていない
ファイルはコースの現在の版に置き換え、一度も作業領域になかったファイルは追加し（受講者が削除・改名したファイルはそのまま）、受講者とコースの両方が
変更したファイルにはコースの版を `<ファイル>.new` として置き、変更された
`messages.go`/`messages.rs` は `.bak` に退避してから置き換えます（新しいレッスンには新しい
定数が必要なため）。記録ができる前に作られた作業領域は、git 履歴にあるそのファイルの
全版と照合し、どれとも一致しないファイルは受講者のコードとしてそのままにします。安全な更新は `check` が自動で行います。つまり、scaffold や starter の
ファイルを直してもかまいません。受講者は `git pull` で修正を受け取れます。

## 検証

```
./mudschool verify              # 全レッスン、全言語
./mudschool verify --lang ja    # 1 つの言語だけ
./mudschool verify 07 08        # 一部のレッスンだけ（構築は順番どおり全部行う）
```

`verify` は一時ディレクトリに作業領域を作ります: `starter/`、続いて各レッスンの
`scaffold/` と `solution/` を順に適用します。これを言語ごとに行います。
各レッスンについて次を確認します:

1. レッスン *前* のコード（前の解答 + scaffold）ではテストが **失敗する**。
   そうでなければ、そのレッスンは何も新しいことをテストしていません。
2. レッスンの解答ではテストに **合格する**。

`./mudschool reference <dir> [id]` は（指定したレッスンまでの）模範解答全体を作るので、
実際に動かせます。コミット前に `verify` を実行してください。CI
（`.github/workflows/verify.yml`）も同じことをします。

## スクリーンショット

`docs/images/client-<言語>.png` は `tools/screenshots/make.sh` で模範解答から生成します
（tmux、nc、Playwright 入りの Node、Noto CJK などの日本語フォントが必要）。クライアントの
見た目が変わったら再実行してください。日本語と罫線文字をそろえる必要がある図は、
テキストではなく画像にしましょう。GitHub のフォントでは CJK 文字がちょうど 2 列幅で
描かれないためです。

## 新しいレッスンの書き方

1. `lessons/NN-トラック-名前/` を作り、`lesson.json`、`README*.md`、`HINTS*.md` を置く。
2. `solution/` を書く: レッスン後の姿の **ファイル全体**。変わらないファイルは不要です。
   以前のレッスンから引き継がれます。
3. プレイヤー向けの新しいテキスト？ `starter.<言語>/` の **3 つすべての**
   `messages.go`（または `messages.rs`）に定数を追加する。
4. `tests/` を書く。注意点:
   - 失敗メッセージは `L(sv, en, ja)` で、何が期待されていたかを伝える。
   - ハードコードしたテキストではなく、`messages.go` の定数と比較する。
   - 実装ではなく振る舞いをテストする。サーバーのテストはなるべく
     `_shared/go/server/mud/mudschool_helpers_test.go` の補助関数（`startServer`、
     `login`、`expect`、`expectContains`、`expectSilence` など）で TCP 越しに行う。
   - 厳しすぎず: レッスンで教えることをテストし、それ以外は自由にさせる。
   - 避けられるならスリープは使わず、タイムアウトを使う。
5. 新しいファイルの `scaffold/` を書く: コードがコンパイルできるよう、
   `TODO(lesson NN)` と `todo!()`／空の戻り値を付けた型とシグネチャ。
6. `./mudschool verify NN` を実行する。

### README のテンプレート

```
# レッスン NN – タイトル
（動機: ゲームのどんな問題を解決するのか？）
## 目標
## <言語> の概念      （短い例。できれば「Go ↔ Rust」の比較）
## 課題               （番号付き、ファイルごとに、テストが求める定数を明記）
## 自分で試す
## 考えてみよう       （振り返りの問いを 2〜3 個）
## チャレンジ         （任意。テストには影響しない）
```

### ヒント

`HINTS.md` の `## ` 見出しが 1 つのヒントです。あいまいなものから具体的なものへ
並べましょう。最後のヒントはほぼコードでもかまいません。
