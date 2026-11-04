# ヒント

## NewWorld

```go
w := &World{Start: start, Rooms: make(map[string]*Room)}
for _, r := range rooms {
    w.Rooms[r.ID] = r
}
return w
```

## プレイヤーを置く

すでにロックを握っている `addPlayer` の中で: `p.Room = s.world.Room(s.world.Start)`。

## 部屋にいるプレイヤー

```go
func (s *Server) playersInRoomLocked(r *Room, except *Player) []*Player {
    var list []*Player
    for _, other := range s.players {
        if other.Room == r && other != except {
            list = append(list, other)
        }
    }
    // 名前でソート: sort.Slice(list, func(i, j int) bool { return list[i].Name < list[j].Name })
    return list
}
```

## 歩く

```go
to := s.world.Room(p.Room.Exits[dir])  // 出口がなければ ID は "" -> nil
if to == nil {
    p.Send(MsgCantGo)
    return
}
```

## デッドロック？

テストが固まりますか？ ロックを握ったまま `s.mu` をロックしている関数がないか
確認してください。例えば、ロックを握った `cmdSay` が（自分でロックする）
`s.broadcast` を呼んでいないか。
