# Ledtrådar

## encodeJSON

```go
data, err := json.Marshal(f)
if err != nil {
    return ""
}
return string(data) + "\r\n"
```

## Ett rutnät av mellanslag

```go
size := 4*radius + 1
grid := make([][]byte, size)
for i := range grid {
    grid[i] = []byte(strings.Repeat(" ", size))
}
```

## Var hamnar ett rum?

```go
dx, dy := r.X-center.X, r.Y-center.Y
// hoppa över rum på annan våning (Z) eller utanför -radius..radius
col, row := 2*(dx+radius), 2*(dy+radius)
grid[row][col] = terrainCodes[r.Terrain]
```

Utgång `east`: `grid[row][col+1] = '-'`. Utgång `south`: `grid[row+1][col] = '|'`.
Kolla att du inte hamnar utanför rutnätet!

## Kopiera innan du ändrar

```go
rows := append([]string(nil), m.Rows...)   // ny slice med samma strängar
line := []byte(rows[mid])                  // strängar går inte att ändra – bytes gör
line[m.Width/2] = '@'
rows[mid] = string(line)
```

Samma sak för `Legend`: skapa en ny map och kopiera över.

## Läsa @hello

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
