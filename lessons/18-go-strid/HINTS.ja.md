# ヒント

## roll

```go
if d[1] <= d[0] {
    return d[0]
}
return d[0] + s.rng.IntN(d[1]-d[0]+1)
```

## ラウンド

```go
for _, name := range s.playerNames() {
    p := s.players[name]
    n := p.target
    if n == nil {
        continue
    }
    if n.dead || n.Location != p.Room {
        p.target = nil
        continue
    }
    // プレイヤーが攻撃、倒せたら勝利、そうでなければ生き物が反撃、倒れることも
}
```

## 倒された生き物

`World.NPCs` からは消さず、印を付け（`dead = true`）、部屋の `NPCs` から取り除き、
`respawnAt = s.ticks + respawn` を覚えておきます。時が来たら `respawnCreatures` が戻します。

## 倒れたとき

プレイヤーを `s.world.Room(s.world.Start)` に移し、`HP = MaxHP` にし、`target` を消し、
`MsgYouDie` を送って `s.look(p)` を呼びます。
