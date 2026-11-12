# Hints

## tutorialArea – breadth-first search

```go
area := map[string]bool{w.Tutorial.Start: true}
queue := []string{w.Tutorial.Start}
for len(queue) > 0 {
	r := w.Room(queue[0])
	queue = queue[1:]
	for _, to := range r.Exits {
		if !area[to] && w.Room(to) != nil {
			area[to] = true
			queue = append(queue, to)
		}
	}
}
```

## validateTutorial – exactly one condition

```go
n := 0
for _, set := range []bool{st.Command != "", st.Room != "", st.Item != ""} {
	if set {
		n++
	}
}
```

A `switch` with one case per mistake keeps one error per step.

## onTutorialCommand – several steps at once

```go
for p.tutorial > 0 && p.tutorial <= len(t.Steps) && s.stepDone(p, t.Steps[p.tutorial-1], e.Text) {
	p.tutorial++
	advanced = true
}
```

A player who already stands in the room the next step asks for passes it at
once. After the loop, `p.tutorial > len(t.Steps)` means they are done.

## leaveTutorial – keep some, give some back

```go
kept := p.Inventory[:0] // filter in place
for _, it := range p.Inventory {
	if area[it.Room] { // it.Room is where the item started
		s.world.placeItem(s.world.Room(it.Room), it)
	} else {
		kept = append(kept, it)
	}
}
p.Inventory = kept
```

## The migration

`ALTER TABLE ... ADD COLUMN ... NOT NULL` needs a `DEFAULT` – SQLite has to
fill in the column for the rows that are already there. Then add `tutorial`
to both the `INSERT ... ON CONFLICT` and the `SELECT`.

## dispatch

`dispatch` returns early in several places. Make the built-in, script and
plugin cases fall through to the end, return only for an unknown command,
and `emit` last.
