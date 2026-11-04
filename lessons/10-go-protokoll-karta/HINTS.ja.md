# ヒント

## encodeJSON

```go
data, err := json.Marshal(f)
if err != nil {
    return ""
}
return string(data) + "\r\n"
```

## 空白のグリッド

```go
size := 4*radius + 1
grid := make([][]byte, size)
for i := range grid {
    grid[i] = []byte(strings.Repeat(" ", size))
}
```

## 部屋はどこに置く？

```go
dx, dy := r.X-center.X, r.Y-center.Y
// 別の階（Z）や -radius..radius の外の部屋は飛ばす
col, row := 2*(dx+radius), 2*(dy+radius)
grid[row][col] = terrainCodes[r.Terrain]
```

出口 `east`: `grid[row][col+1] = '-'`。出口 `south`: `grid[row+1][col] = '|'`。
グリッドの外に出ないよう気をつけて！

## 変更する前にコピーする

```go
rows := append([]string(nil), m.Rows...)   // 同じ文字列を持つ新しいスライス
line := []byte(rows[mid])                  // 文字列は変更できないがバイト列はできる
line[m.Width/2] = '@'
rows[mid] = string(line)
```

`Legend` も同じです。新しいマップを作ってコピーしましょう。

## @hello を読む

```go
for p.scanner.Scan() {
    line := p.scanner.Text()
    if strings.HasPrefix(line, HelloPrefix) {
        p.structured.Store(true)
        p.SendFrame(Frame{Type: "hello", Text: "mudschool 1"})
        continue
    }
    return line, true
}
return "", false
```
