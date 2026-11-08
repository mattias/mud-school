# Ledtrådar

## castLight i sin helhet

Det här är den klassiska versionen (Björn Bergström, RogueBasin), anpassad till `Zone`:

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

Läs den med ett papper: rita en oktant, en vägg, och följ `start`, `end`
och `newStart`.

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

## En fälla i syncPlayer

`updateFOV` ändrar `p.explored`. Tar du `before := p.explored[room]` och
sedan anropar `updateFOV` är `before` *samma map* – maps är referenser. Därför
returnerar `updateFOV` de nya rutorna i stället.
