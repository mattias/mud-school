# レッスン 42 – インターネットへ

世界を招き入れる時です。このレッスンでサーバーは、目の前にないコンピューターで生きるために
必要なものを手に入れます。そして実際にそこへ置きます。小さな **VPS**（仮想サーバー、
月に数ユーロ）で、**Docker** で動かし、前に **Caddy** を置いて HTTPS にします。

```
                          あなたの VPS
                ┌────────────────────────────────────┐
https://mud.… ──┼─▶ Caddy :443 ──▶ mud :8080 (web)   │   /data/mud.db
telnet  …:4000 ─┼──────────────▶ mud :4000 (telnet)  │   （ボリューム）
                └────────────────────────────────────┘
```

## 目標

- **環境変数** でプログラムを設定し、それでも `main` をテストしやすく保つ。
- `log/slog` の **構造化ログ**: 人にはテキスト、機械には JSON。
- **ヘルスチェック**（`/healthz`）と、**SIGTERM** できれいに止まること。
- **プロキシ** の後ろで本当に誰がつないでいるかを知る（`X-Forwarded-For`）、そして
  信じてはいけないときを知る。
- 小さく安全なコンテナイメージを作り、VPS で動かす。

## Go の概念

### テストできる `main`

`main` はサーバーを起動して戻ってきません。テストしにくいのです。だからテストできる部分を
外に出します:

```go
func loadConfig(args []string, lookup func(string) (string, bool)) (Config, error)
```

`main` は `loadConfig(os.Args[1:], os.LookupEnv)` を呼び、テストは map で呼びます。
各フラグの既定値はその環境変数（`MUD_ADDR`、`MUD_WEB`、`MUD_DB` …）から来るので、
フラグは環境変数に、環境変数は既定値に勝ちます。`os.Getenv` ではなく `os.LookupEnv` なら
「空で設定」と「未設定」を区別できます。`MUD_DB=` は「何も保存しない」です。

### `log/slog`

```go
slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))
```

`SetDefault` の後は古い `log.Printf` も slog を通るので、`-log json` なら各行がログサービスで
検索できる JSON オブジェクトになります。

### SIGTERM

`docker stop` は Ctrl+C を押しません。**SIGTERM** を送り、10 秒待ってから強制終了します。
`signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)` を `main` の最初に置けば
（起動中のシグナルも失われません）、サーバーは先に全員に別れを告げて保存できます。

### `/healthz`

```go
mux.Handle("/healthz", srv.HealthHandler())
mux.Handle("/", gw)
```

ハンドラーは時間制限付きでゲームループにプレイヤー数を尋ねます。ループが止まっている
サーバーは、HTTP が答えても健康ではありません。コンテナイメージには `curl` がないので、
サーバーが自分で確かめます: `mudserver -healthcheck` は `/healthz` を尋ねて 0 か 1 で終わります。
それが Docker の `HEALTHCHECK` です。

### プロキシの後ろで

Caddy の後ろでは、どの WebSocket も *Caddy から* 来ます。Caddy は本当のプレイヤーが誰かを
ヘッダー `X-Forwarded-For` で伝えます。でもそのヘッダーは誰でも送れます。だからゲートウェイが
それを信じるのは `TrustProxy` のときだけで、それはサーバーがプロキシの後ろにあるときだけ
設定します。レッスン 43 はアドレスごとに接続を数えるので、これは大事です。

## デプロイのファイル（用意済み）

- `Dockerfile` – 3 段階: Rust が `mud_web.wasm` を作り、Go が Web クライアント入りの
  サーバーを作り（`CGO_ENABLED=0`: SQLite のドライバーも純粋な Go）、最後のイメージは
  *distroless*: シェルなし、root でない、約 30 MB。
- `compose.yaml` – MUD と Caddy。データベースは *ボリューム* にあるので、新しい版になっても残る。
- `Caddyfile` – 3 行: あなたのドメイン、圧縮、`reverse_proxy mud:8080`。
  証明書は Caddy が Let's Encrypt から自分で取って更新する。
- `.env.example` – あなたのドメインと管理者。

## 契約

- `Config` と `loadConfig(args, lookup)`: フラグ `-addr`、`-web`、`-world`、`-db`、`-tick`、
  `-reload`、`-admin`、`-scripts`、`-plugins`、`-log`（`text`|`json`）、`-trust-proxy`、
  `-healthcheck`。既定値は `MUD_ADDR`、`MUD_WEB`、`MUD_WORLD`、`MUD_DB`、`MUD_TICK`、
  `MUD_RELOAD`、`MUD_ADMINS`、`MUD_SCRIPTS`、`MUD_PLUGINS`、`MUD_LOG`、`MUD_TRUST_PROXY` から。
  空で設定も数える。誤った値は、誤っていた変数をすべて挙げた 1 つのエラーになる。
  `-admin` はカンマで分けて空白を除く。
- `newLogger(format, w) *slog.Logger`。`healthCheck(addr) error`（`http://addr/healthz` が
  200 のときだけ nil。`:8080` は `127.0.0.1:8080` の意味）。
- `(s *Server) HealthHandler()`: 200 `{"status":"ok","players":N,"uptime_seconds":S}`、
  停止中やループが答えないときは 503 `{"status":"stopping",...}`。
- `Gateway.TrustProxy`: Web プレイヤーのアドレスは `X-Forwarded-For` の最後のもの。
  `TrustProxy` がなければヘッダーは無視。
- `main`: SIGTERM は Ctrl+C と同じで終了コード 0。ゲートウェイの隣に `/healthz`。

## 課題

1. `config.go` は一部用意されています（`Config`）。`loadConfig`、`newLogger`、`healthCheck` を
   書き、`main` で使いましょう。
2. `mud/health.go`: `HealthHandler`。`mud/web.go`: `TrustProxy`。
3. 手元で試す: `MUD_LOG=json ./mudschool run server`、それから `curl localhost:8080/healthz`。
4. Docker があれば: `workspace/` で `docker build -t mud .`、そして
   `docker run -p 4000:4000 -p 8080:8080 -v mud-data:/data mud`。

## VPS へ

1. Linux（Ubuntu か Debian）のいちばん小さな VPS を借りる。CPU 1 つとメモリ 1 GB で十分です。
2. ドメイン名を向ける: 例えば `mud.example.com` に VPS の IP アドレスの **A レコード**。
   （無料のサブドメインがある DNS サービスもたくさんあります。）
3. `ssh` でログインし、Docker をインストールし（`curl -fsSL https://get.docker.com | sh`、
   またはディストリビューションのパッケージ）、ファイアウォールがあればポート 80・443・4000 を開ける。
4. `workspace/` をそこへコピーする（`git push`/`git clone`、または `rsync -a --exclude target`）。
5. `cp .env.example .env`、`MUD_DOMAIN` と `MUD_ADMINS` を埋めて `docker compose up -d --build`。
6. `https://mud.example.com` を開く。そして **すぐに管理者の名前を登録** する
   （レッスン 31: 管理者は名前にすぎません）。

ふだんは: `docker compose logs -f mud` でログを見て、`docker compose up -d --build` で新しい版を
出し、`docker compose cp mud:/data/mud.db mud-backup.db` でデータベースをコピーします
（できれば誰も遊んでいないとき。チャレンジを見てください）。

## 考えてみよう

- イメージの最後の段階にシェルがないのはなぜ？ プログラムに入り込んだ攻撃者は、
  シェルがあるときとないときで何ができる？
- ポート 4000 の telnet には TLS がありません。パスワードは平文で流れます。どうする？
- 遊んでいる最中に `docker compose up -d --build` がコンテナを入れ替えたら、プレイヤーは
  どうなる？（レッスン 35 には答えがありました。ここではなぜ使えない？）
- `MUD_TRUST_PROXY=true` が `Dockerfile` ではなく `compose.yaml` に入るのはなぜ？

## チャレンジ

- Docker なしで systemd のサービスとして動かす。Docker がしていたことのうち systemd がするのは？
- `mud.db` の毎晩の自動バックアップ（SQLite の `VACUUM INTO`）。
- ステータスページ: `/healthz` からオンラインのプレイヤーと稼働時間。
