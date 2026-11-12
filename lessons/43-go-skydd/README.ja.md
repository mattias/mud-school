# レッスン 43 – 守り

あなたの MUD はインターネットにあります。遅かれ早かれ、誰かが本を丸ごと貼り付け、
毎秒千のコマンドをスクリプトで送り、百の接続を開き、あるいはつないだまま一言も話しません。
そのどれも、ほかの全員にとってゲームを遅くしてはいけません。コマンドはすべて **1 つの**
ゲームループで動くことを思い出しましょう。

```
> say spam
> say spam
...（さらに 500 回）
落ち着いて – コマンドの送り方が速すぎます。
...
コマンドが多すぎます – さようなら。
```

## 目標

- **トークンバケット** で頻度を制限し、待たずにテストする。
- 制限は安いところに置く: コマンドがゲームループに届く *前* に。
- 終わらないクライアントに期限を使う（`SetReadDeadline`）。
- 起きたことを `expvar` で数え、時間がどこで使われているかを `pprof` で見る。

## Go の概念

### トークンバケット

```
burst ┃■■■■■┃  コマンドは 1 つごとにトークンを 1 つ使う。
      ┃■■■  ┃  毎秒 `rate` 個のトークンが戻ってくる。
      ┃     ┃  トークンがなければ、コマンドは捨てられる。
```

```go
func (r *RateLimiter) Allow(now time.Time) bool
```

短いバースト（速い 5 歩）は問題なし、途切れない洪水はだめ。時刻は引数なので、テストは
眠らずに「0.5 秒後」と言えます。リミッターはプレイヤー自身のゴルーチン（`play`）にあり、
捨てられたコマンドはゲームループにまったく届きません。

### 洪水には静かに答える

500 行の洪水に 500 の返事をしてはいけません。洪水が倍になります。プレイヤーが受け取る
`MsgSlowDown` は洪水ごとに 1 回で、`FloodLimit` 回続けて捨てられたら切断されます。

### 読み込みの制限

```go
p.scanner.Buffer(make([]byte, 0, min(1024, max)), max)   // 長すぎる行: ErrTooLong
conn.SetReadDeadline(time.Now().Add(s.loginTimeout))        // 時間内にログイン
```

`bufio.Scanner` はバッファの上限より長い行を拒み、それで接続は `MsgLineTooLong` で終わります。
（上限は `max` とバッファの容量の大きい方。小さく始めましょう。）レッスン 14 の停止で使った
読み込み期限を再利用します: ログインは時間内に済ませなければならず、済んだら期限は外します。
ただし、その間に停止が期限を設定していたら外しません。

### アドレスごと

`p.ip` は `RemoteAddr` のホスト部分です。Caddy の後ろでも、`TrustProxy`（レッスン 42）の
おかげでプレイヤーの本当のアドレスです。ゲームループは新しい接続を入れる前に、`s.conns` の
中で同じ `ip` の接続を数えます。

### `expvar` と `pprof`

```go
var metricCommands = expvar.NewInt("mud_commands_total")
```

`expvar` はカウンターを `/debug/vars` に JSON で公開し、`net/http/pprof` は `/debug/pprof/` を
足します。どちらも `http.DefaultServeMux` に登録され、サーバーはそれを専用のアドレス
`-debug 127.0.0.1:6060` で配信します。公開のアドレスでは決して配信しません:

```
go tool pprof http://127.0.0.1:6060/debug/pprof/profile?seconds=10
```

で CPU がどの関数で忙しいかがわかります。レッスン 44 がその仕事を用意します。

## 契約

- `Options.MaxPerIP`、`CommandRate`、`CommandBurst`、`MaxLineLength`（0 = `DefaultMaxLine`）、
  `LoginTimeout`。0 は制限なし。
- `NewRateLimiter(rate, burst)` は満杯で始まり、`Allow(now)` は前回からの時間で補充する。
  `burst` を超えない。
- 1 つの IP からの接続が多すぎる: `MsgTooManyConnections` で閉じる。1 つ閉じればまた入れる。
- 頻度を超えたコマンドは捨てる。`MsgSlowDown` は洪水ごとに 1 回。`FloodLimit` 回続いたら
  `MsgFlooding` で切断。
- 上限より長い行: `MsgLineTooLong` で閉じる。`LoginTimeout` 以内にログインしない:
  `MsgLoginTimeout` で閉じる。ログイン後は時間制限なし。
- `expvar`: `mud_connections_total`、`mud_connections_refused_total`、`mud_commands_total`、
  `mud_commands_dropped_total`、`mud_players`。
- `main`: `-max-per-ip`（5）、`-rate`（10）、`-burst`（20）、`-login-timeout`（1m）、
  `-debug`（オフ）。`MUD_MAX_PER_IP`、`MUD_RATE`、`MUD_BURST`、`MUD_LOGIN_TIMEOUT`、`MUD_DEBUG` 付き。

## 課題

1. `mud/protect.go` は一部用意されています（定数、カウンター、`hostOf`）。`Allow`、
   `connsFrom`、`allowCommand` を書きましょう。
2. `server.go`: オプションとフィールド。`handle` で IP、スキャナーの上限、IP ごとの確認、
   ログインの期限（とその解除）。`play` でリミッター。`addPlayer`/`removePlayer` で
   プレイヤーを数える。
3. `player.go`: `ip`、`limiter`、`dropped`、そして `readLine` で `MsgLineTooLong`。
4. `config.go` と `main.go`: 新しいフラグとデバッグ用のサーバー。

## 考えてみよう

- なぜ頻度のリミッターはプレイヤーのゴルーチンに、IP ごとの数はゲームループにあるのでしょう？
- 学校やカフェは 1 つの IP アドレスを共有します。`-max-per-ip 5` は彼らにとって何を意味する？
- bcrypt（レッスン 20）はわざと遅くしてあります。千のログインを同時にされたら、攻撃者は
  何ができる？ 何を制限する？
- ダッシュボードにはどんな数字が欲しい？

## チャレンジ

- *全体の* 制限: ゲームループで 1 ティックあたり最大 N コマンド、公平に分ける。
- 洪水の後、そのアドレスを 10 分間締め出す。
- カウンターを Prometheus のテキスト形式で `/metrics` に出す。
