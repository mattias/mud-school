# Tips

## tutorialArea – bredden först

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

## validateTutorial – precis ett villkor

```go
n := 0
for _, set := range []bool{st.Command != "", st.Room != "", st.Item != ""} {
	if set {
		n++
	}
}
```

En `switch` med ett fall per fel ger ett fel per steg.

## onTutorialCommand – flera steg på en gång

```go
for p.tutorial > 0 && p.tutorial <= len(t.Steps) && s.stepDone(p, t.Steps[p.tutorial-1], e.Text) {
	p.tutorial++
	advanced = true
}
```

En spelare som redan står i rummet som nästa steg ber om klarar det direkt.
Efter loopen betyder `p.tutorial > len(t.Steps)` att hen är klar.

## leaveTutorial – behåll vissa, lämna tillbaka andra

```go
kept := p.Inventory[:0] // filtrera på plats
for _, it := range p.Inventory {
	if area[it.Room] { // it.Room är där saken låg från början
		s.world.placeItem(s.world.Room(it.Room), it)
	} else {
		kept = append(kept, it)
	}
}
p.Inventory = kept
```

## Migreringen

`ALTER TABLE ... ADD COLUMN ... NOT NULL` behöver en `DEFAULT` – SQLite måste
fylla i kolumnen för raderna som redan finns. Lägg sedan till `tutorial` i
både `INSERT ... ON CONFLICT` och `SELECT`.

## dispatch

`dispatch` returnerar tidigt på flera ställen. Låt fallen för inbyggda
kommandon, skript och plugins gå vidare till slutet, returnera bara för ett
okänt kommando, och `emit` sist.
