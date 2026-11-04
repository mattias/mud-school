# ヒント

## ファイル形式

ファイル用の構造体を別に作りましょう:

```go
type worldFile struct {
    Start string  `json:"start"`
    Rooms []*Room `json:"rooms"`
}
```

これにデコードし、`NewWorld(f.Start, f.Rooms...)` で世界を作ります。
3 つの点はスライスを可変長引数に「展開」します。

## 座標

表にすると簡単です:

```go
var directionDelta = map[string][3]int{
    "north": {0, -1, 0}, "south": {0, 1, 0}, "east": {1, 0, 0},
    "west": {-1, 0, 0}, "up": {0, 0, 1}, "down": {0, 0, -1},
}
```

重複の確認には: `positions := map[[3]int]string{}`。

## その地形はある？

`slices.Contains(Terrains, r.Terrain)`（`slices` パッケージ、Go 1.21 以降）。

## 埋め込みファイル

```go
import (
    "bytes"
    _ "embed"
)

//go:embed world.json
var defaultWorldJSON []byte

func DefaultWorld() *World {
    w, err := LoadWorld(bytes.NewReader(defaultWorldJSON))
    if err != nil {
        panic(err)
    }
    return w
}
```
