# ヒント

## tutorialArea – 幅優先探索

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

## validateTutorial – 条件はちょうど 1 つ

```go
n := 0
for _, set := range []bool{st.Command != "", st.Room != "", st.Item != ""} {
	if set {
		n++
	}
}
```

誤りごとに 1 つの case を持つ `switch` にすると、ステップごとのエラーは 1 つになります。

## onTutorialCommand – 一度に複数のステップ

```go
for p.tutorial > 0 && p.tutorial <= len(t.Steps) && s.stepDone(p, t.Steps[p.tutorial-1], e.Text) {
	p.tutorial++
	advanced = true
}
```

次のステップが求める部屋にもう立っているプレイヤーは、すぐにそれを通過します。
ループの後で `p.tutorial > len(t.Steps)` なら終わりです。

## leaveTutorial – 残すものと返すもの

```go
kept := p.Inventory[:0] // その場で絞り込む
for _, it := range p.Inventory {
	if area[it.Room] { // it.Room は物が最初にあった場所
		s.world.placeItem(s.world.Room(it.Room), it)
	} else {
		kept = append(kept, it)
	}
}
p.Inventory = kept
```

## マイグレーション

`ALTER TABLE ... ADD COLUMN ... NOT NULL` には `DEFAULT` が必要です。SQLite は
既存の行の列を埋めなければならないからです。それから `INSERT ... ON CONFLICT` と
`SELECT` の両方に `tutorial` を足しましょう。

## dispatch

`dispatch` はあちこちで早めに return します。組み込み・スクリプト・プラグインの場合は
最後まで流れるようにし、未知のコマンドのときだけ return して、最後に `emit` しましょう。
