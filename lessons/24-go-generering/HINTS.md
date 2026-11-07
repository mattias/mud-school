# Hints

## caStep

```go
walls := 0
for dy := -1; dy <= 1; dy++ {
	for dx := -1; dx <= 1; dx++ {
		ny, nx := y+dy, x+dx
		if ny < 0 || nx < 0 || ny >= h || nx >= w || grid[ny][nx] {
			walls++
		}
	}
}
next[y][x] = walls >= 5
```

Write to a **new** grid – if you change `grid` while you read it, the cells
you have already changed affect their neighbours.

## ValueNoise

```go
x0, y0 := math.Floor(x), math.Floor(y)
tx, ty := smooth(x-x0), smooth(y-y0)
ix, iy := int(x0), int(y0)
a := lattice(seed, ix, iy)
b := lattice(seed, ix+1, iy)
c := lattice(seed, ix, iy+1)
d := lattice(seed, ix+1, iy+1)
```

`math.Floor` (not `int(x)`) so that negative coordinates also work.

## reachable

```go
queue := []Pos{start}
seen[start] = true
for len(queue) > 0 {
	p := queue[0]
	queue = queue[1:]
	for _, d := range neighbours4 {
		// a walkable neighbour you haven't seen: mark it and add it to the queue
	}
}
```

## drunkardWalk

```go
for steps := 0; p != to && steps < z.W*z.H*8; steps++ {
	d := neighbours4[rng.IntN(4)]
	if rng.Float64() < 0.7 {
		d = towards(p, to)
	}
	n := p.Add(d)
	if n != to && (n.X <= 0 || n.Y <= 0 || n.X >= z.W-1 || n.Y >= z.H-1) {
		continue
	}
	// move, and make the cell floor
}
```

A limit on the steps, and a straight line at the end if it gives up – random
algorithms need a way out.
