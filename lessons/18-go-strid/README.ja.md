# レッスン 18 – 戦闘

洞窟のネズミが歯をむき、暗い森で狼が遠吠えする。戦いの時です。MUD の戦闘は
ゲームループの典型的な仕事です。いったん戦いが始まれば、誰かが倒れるか逃げるまで
**1 ティックに 1 ラウンド** 進みます。

```
> kill ネズミ
洞窟ネズミ に攻撃を仕掛けた！
洞窟ネズミ に命中（3 ダメージ）。     ← ティック
洞窟ネズミ の攻撃が命中（1 ダメージ）。
洞窟ネズミ に命中（2 ダメージ）。     ← ティック
洞窟ネズミ を倒した！
```

難しいのは戦闘そのものではなく、ダイスと時間に左右されるものを **テスト** する
ことです。このレッスンのテストは世界を止め（`Tick: time.Hour`）、自分で 1 ティック
ずつ進めます:

```go
srv.call(func() { srv.ticks++; srv.onTick() })   // ちょうど 1 ティック
```

ゲームループの状態が 1 か所にあるのはこのためです。テストはデバッガのように
ループを操作できます。

## 目標

- 戦闘をゲームループ内の状態（誰が誰と戦っているか）として表し、毎ティック進める。
- 公平でテストしやすい乱数を書く（注入した `rand.Rand` を使う `roll`）。
- 時間に依存する振る舞いの決定的なテストを書く。
- 再びプロトコルを拡張する: `status.target`。

## 契約

- `const PlayerMaxHP = 20`、`const DefaultRespawn = 50`、`var PlayerDamage = [2]int{1, 4}`、
  `Options.Respawn`（ティック数、0 ならデフォルト）。
- `Player.HP`、`Player.MaxHP`（ログイン時は満タン）と `Player.target *NPC`。
- レッスン 16 の生き物のフィールド: `HP`（最大体力）、`Damage [2]int`、`Attackable`。
  実行時は現在の体力、倒されたかどうか、いつ戻ってくるか。
- コマンド:
  - `kill <生き物>`（`k`、`attack`）: `MsgAttackWhat`、`MsgNoSuchTarget`、`MsgCantAttack`。
    それ以外なら自分に `MsgYouAttack`、部屋に `MsgAttacks`。
  - `flee`: `MsgNotFighting`、`MsgFleeNowhere`。それ以外なら `MsgYouFlee` を送り、
    ランダムな出口から出ていく。戦闘は終わり。
- 毎ティック、戦っているすべてのプレイヤーについて **名前順** に:
  1. プレイヤーが `roll(PlayerDamage)` で攻撃: `MsgYouHit`、
  2. 生き物の体力が 0 以下なら: 自分に `MsgYouKill`、部屋に `MsgKills`。その生き物と
     戦っている全員が戦闘をやめ、生き物は消え、`Respawn` ティック後に体力満タンで
     開始部屋に戻る（`MsgAppears`）、
  3. そうでなければ生き物が `roll(n.Damage)` で反撃: `MsgHitsYou`、
  4. 自分の体力が 0 以下なら: 部屋に `MsgDies`。体力満タンで開始部屋で目を覚まし
     （部屋の名前付きの `MsgYouDie`）、`look` を受け取る。
- 誰かと戦っている生き物は歩き回らない。
- 戦闘メッセージの kind はすべて `combat`。
- `Status` はプレイヤーの `HP`/`MaxHP` と、戦闘中は `Target *Target{Name, HP, MaxHP}`
  （json `target,omitempty`）を持つ。構造化クライアントにはラウンドごとに status を送る。

## 課題

1. **`combat.go`**（新規）: 関数を埋める。
2. **`loop.go`**: `onTick` は倒された生き物を飛ばし、その後 `combatRound()` と
   `respawnCreatures()` を実行する。
3. **`npc.go`**: 実行時のフィールド（現在の HP、倒されたか、復活ティック、開始部屋）を
   `AddNPCs` で設定する。戦闘中の `Wanderer` は動かない。
4. **`server.go`**/**`player.go`**: `Options.Respawn`、プレイヤーの HP。
5. **`frame.go`**/**`rooms.go`**: status の `Target`。
6. `kill` と `flee`、省略形 `k` と `attack` を登録する。

## 自分で試す

```
./mudschool run server
```

洞窟まで歩き（s, s, s, s, s, s, e, s）、`kill ネズミ`。次は暗い森の狼に挑戦――
ただし早めに `flee` を！

## 考えてみよう

- 十分なティックが過ぎたことを祈って待つのではなく、テストが自分で世界を進めるのはなぜ？
- 戦闘ラウンドがプレイヤーを決まった順序で処理しなければならないのはなぜ？
- 2 人が同じ生き物と戦ったら？ とどめを刺すのは誰？

## チャレンジ

- 経験値: 倒した生き物から XP を得て、XP がたまると最大 HP が上がる。
- 武器: 錆びた剣（レッスン 15）を持っていると与えるダメージが増える。
- 攻撃的な生き物: 部屋に入ってきたプレイヤーに襲いかかる。
