# レッスン 09 – 世界ファイル

部屋を Go のコードで書くのは、6 部屋ならいいですが 600 部屋では無理です。それに、
再コンパイルしないと世界を変えられません。本物の MUD には、データファイルで
エリアを作る *ビルダー* がいます。これは **稼働中に内容を再読み込みできる**
サーバー（フェーズ 5）への第一歩でもあります。

コースツールが `workspace/server/mud/world.json` に 21 部屋を用意しました:
マドビー、街道、湖、森、洞窟です。どの部屋にも **地形** と **座標** があります。
次のレッスンのマップに必要です。

## 目標

- `encoding/json` で JSON を読み、**構造体タグ** でフィールド名を指定する。
- インターフェースに対してプログラムする: ファイル名ではなく `io.Reader`。
- データをきちんと検証し、`errors.Join` で *すべての* エラーを報告する。
- `//go:embed` でファイルをプログラムに埋め込む。

## Go の概念

### JSON と構造体タグ

```go
type Room struct {
    ID    string            `json:"id"`
    Exits map[string]string `json:"exits,omitempty"`
}

var r Room
err := json.Unmarshal(data, &r)        // []byte から

dec := json.NewDecoder(reader)         // どんな io.Reader からでも
dec.DisallowUnknownFields()            // "name" ではなく "nmae" -> エラー
err = dec.Decode(&r)
```

### io.Reader

`LoadWorld(r io.Reader)` はファイル（`os.Open`）、文字列（`strings.NewReader` –
テストはこれを使います）、バイトスライス（`bytes.NewReader`）、ネットワーク……
どこからでも読めます。Go のインターフェースは *暗黙的に* 満たされます。
メソッド `Read([]byte) (int, error)` を持つ型はすべて `io.Reader` です。

### errors.Join

```go
var errs []error
errs = append(errs, fmt.Errorf(MsgErrExitMissing, id, dir, to))
return errors.Join(errs...)   // リストが空なら nil
```

5 つ間違えたビルダーは、5 つすべてを一度に見たいものです。

### go:embed

```go
import _ "embed"

//go:embed world.json
var defaultWorldJSON []byte   // コンパイル時に埋め込まれたファイルの中身
```

## 座標系

```
          y-1（北）
 x-1（西） ·  x+1（東）        z+1 = 上、z-1 = 下
          y+1（南）
```

## 課題

1. **`world.go`**:
   - `Room` に `Terrain string`、`X`、`Y`、`Z int` フィールドを追加し、全フィールドに
     JSON タグを付ける（名前は `world.json` を見てください）、
   - `LoadWorld(r io.Reader) (*World, error)` – `{"start": "...", "rooms": [...]}` の
     形式を `DisallowUnknownFields` 付きで読み、同じ ID が 2 回出てきたらエラーにし、
     検証する、
   - `LoadWorldFile(path string) (*World, error)`、
   - 次の間違いを **すべて** 見つける `(w *World) Validate() error`:
     開始部屋がない・id か name のない部屋・不明な地形（`Terrains` にない。下記参照）・
     不明な方向・存在しない部屋への出口・座標が同じ 2 つの部屋・座標が合わない出口
     （`north` に進むなら行き先は `y-1` のはず）。エラーメッセージは `messages.go` に
     用意してあります（`MsgErrStartMissing`、`MsgErrExitMissing` など）。
     `fmt.Errorf` で使ってください、
   - `var Terrains = []string{"town", "inside", "road", "field", "forest", "water", "mountain", "cave"}`。
2. **`defaultworld.go`**: `DefaultWorld()` を書き直し、埋め込んだ `world.json` を
   読むようにする（壊れていたら `panic` してかまいません。それはプログラミングの誤りです）。
3. **`main.go`**: 別の世界を読み込むフラグ `-world file.json` を追加する。

## 自分で試す

```
cd workspace/server
cp mud/world.json /tmp/my-world.json   # 何か変えてみる。わざと間違えるのも◎
go run . -world /tmp/my-world.json
```

## 考えてみよう

- `[3]int` はマップのキーにできるのに `[]int` はできないのはなぜ？
- `DefaultWorld` は `panic` してよいのに、`LoadWorldFile` はだめなのはなぜ？

## チャレンジ

- 5 部屋以上の自分のエリアを作り、`world.json` に追加する。
- 一方通行の出口を警告する（ログに出すがエラーにはしない）。
- 検証だけを行う小さなツール `go run ./cmd/worldcheck file.json` を書く。
