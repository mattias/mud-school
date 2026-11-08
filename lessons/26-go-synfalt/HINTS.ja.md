# ヒント

## castLight の全体

古典的な版（Björn Bergström、RogueBasin）を `Zone` 向けにしたものです:

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

紙を用意して読みましょう。オクタントと壁を描き、`start`、`end`、`newStart` を追いかけます。

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

## syncPlayer の落とし穴

`updateFOV` は `p.explored` を変更します。`before := p.explored[room]` としてから
`updateFOV` を呼ぶと、`before` は *同じマップ* です（マップは参照）。だから `updateFOV` は
代わりに新しいマスを返すのです。
