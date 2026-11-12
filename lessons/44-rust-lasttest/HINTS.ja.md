# ヒント

## percentile

```rust
let rank = ((p / 100.0) * sorted.len() as f64).ceil() as usize;
sorted[rank.clamp(1, sorted.len()) - 1]
```

## bot_name

`id` を `a`〜`z` の 26 進数で書き（`insert(0, …)` で各桁を前に置く）、
`format!("Bot{letters}")`。0 → `Bota`、26 → `Botba`。

## run_bot – 期限付きで接続する

```rust
let stream = match timeout(Duration::from_secs(5), TcpStream::connect(&plan.addr)).await {
    Ok(Ok(s)) => s,
    Ok(Err(e)) => return drop(events.send(Event::Failed(e.to_string())).await),
    Err(_) => return drop(events.send(Event::Failed("connect: timeout".into())).await),
};
let (read, mut write) = stream.into_split();
let mut lines = BufReader::new(read).lines();
```

`into_split` は同時に使える読み手と書き手をくれます。

## run_bot – echo

送った echo はすべて `HashMap<String, Instant>` に覚えておき、テキストフレームの文字列が
そのどれかなら `remove` して `sent.elapsed()` を報告します。サーバーが遅いと、いくつも
同時に飛んでいます。

## run – 期限まで集める

```rust
let deadline = Instant::now() + plan.duration;
loop {
    tokio::select! {
        Some(ev) = events.recv() => record(&mut report, ev),
        _ = sleep_until(deadline) => break,
    }
}
```

`Some(ev) = …` はパターンです。`recv()` が `None`（送り手が全員いなくなった）を返せば、
その枝は飛ばされるだけです。

## run – 止める

`stop_tx.send(true)` で各ボットの `stop.changed()` が起きます。タスクは
`timeout(Duration::from_secs(5), handle).await` で待ちましょう。死んだ接続で
止まったボットに、テスト全体を止めさせてはいけません。

## 「あなたのアドレスからの接続が多すぎます」

レッスン 43 がちゃんと仕事をしています。サーバーを `-max-per-ip 0 -rate 0` で起動しましょう。
