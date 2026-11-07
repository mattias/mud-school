# Hints

## castLight in full

This is the classic version (Björn Bergström, RogueBasin), adapted to `Zone`:

```go
func castLight(z *Zone, origin Pos, radius, row int, start, end float64, o [4]int, vis map[Pos]bool) {
	if start < end {
		return
	}
	newStart := 0.0
	for j := row; j <= radius; j++ {
		dy := -j
		blocked := false
		for dx := -j; dx <= 0; dx++ {
			p := Pos{origin.X + dx*o[0] + dy*o[1], origin.Y + dx*o[2] + dy*o[3]}
			left := (float64(dx) - 0.5) / (float64(dy) + 0.5)
			right := (float64(dx) + 0.5) / (float64(dy) - 0.5)
			if start < right {
				continue
			}
			if end > left {
				break
			}
			if dx*dx+dy*dy <= radius*radius && z.In(p) {
				vis[p] = true
			}
			opaque := !Transparent(z.At(p))
			if blocked {
				if opaque {
					newStart = right
					continue
				}
				blocked = false
				start = newStart
			} else if opaque && j < radius {
				blocked = true
				castLight(z, origin, radius, j+1, start, left, o, vis)
				newStart = right
			}
		}
		if blocked {
			break
		}
	}
}
```

Read it with a piece of paper: draw an octant, a wall, and follow `start`,
`end` and `newStart`.

## zoneData

```go
for y, row := range z.Tiles {
	b := make([]byte, z.W)
	for x, t := range row {
		b[x] = ' '
		if ex[Pos{x, y}] {
			b[x] = t
			legend[string(t)] = tileMeanings[t]
		}
	}
	rows[y] = string(b)
}
```

## A trap in syncPlayer

`updateFOV` changes `p.explored`. If you take `before := p.explored[room]` and
then call `updateFOV`, `before` is *the same map* – maps are references. That
is why `updateFOV` returns the new cells instead.
