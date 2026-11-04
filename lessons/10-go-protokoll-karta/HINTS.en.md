# Hints

## encodeJSON

```go
data, err := json.Marshal(f)
if err != nil {
    return ""
}
return string(data) + "\r\n"
```

## A grid of spaces

```go
size := 4*radius + 1
grid := make([][]byte, size)
for i := range grid {
    grid[i] = []byte(strings.Repeat(" ", size))
}
```

## Where does a room go?

```go
dx, dy := r.X-center.X, r.Y-center.Y
// skip rooms on another floor (Z) or outside -radius..radius
col, row := 2*(dx+radius), 2*(dy+radius)
grid[row][col] = terrainCodes[r.Terrain]
```

Exit `east`: `grid[row][col+1] = '-'`. Exit `south`: `grid[row+1][col] = '|'`.
Make sure you don't go outside the grid!

## Copy before you change

```go
rows := append([]string(nil), m.Rows...)   // a new slice with the same strings
line := []byte(rows[mid])                  // strings can't be changed – bytes can
line[m.Width/2] = '@'
rows[mid] = string(line)
```

The same goes for `Legend`: create a new map and copy over.

## Reading @hello

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
