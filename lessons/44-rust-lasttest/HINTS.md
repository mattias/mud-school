# Hints

## percentile

```rust
let rank = ((p / 100.0) * sorted.len() as f64).ceil() as usize;
sorted[rank.clamp(1, sorted.len()) - 1]
```

## bot_name

Write `id` in base 26 with the letters `a`–`z` (`insert(0, …)` puts each
digit in front), then `format!("Bot{letters}")`. 0 → `Bota`, 26 → `Botba`.

## run_bot – connecting with a time limit

```rust
let stream = match timeout(Duration::from_secs(5), TcpStream::connect(&plan.addr)).await {
    Ok(Ok(s)) => s,
    Ok(Err(e)) => return drop(events.send(Event::Failed(e.to_string())).await),
    Err(_) => return drop(events.send(Event::Failed("connect: timeout".into())).await),
};
let (read, mut write) = stream.into_split();
let mut lines = BufReader::new(read).lines();
```

`into_split` gives a reader and a writer that can be used at the same time.

## run_bot – the echo

Remember every echo you send, `HashMap<String, Instant>`; when a text frame's
text is one of them, `remove` it and report `sent.elapsed()`. With a slow
server several are on their way at once.

## run – collecting until a deadline

```rust
let deadline = Instant::now() + plan.duration;
loop {
    tokio::select! {
        Some(ev) = events.recv() => record(&mut report, ev),
        _ = sleep_until(deadline) => break,
    }
}
```

`Some(ev) = …` is a pattern: if `recv()` returns `None` (every sender gone),
that branch is just skipped.

## run – stopping

`stop_tx.send(true)` wakes every bot's `stop.changed()`. Wait for the tasks
with `timeout(Duration::from_secs(5), handle).await` – a bot stuck on a dead
connection must not hang the whole test.

## "Too many connections from your address"

That is lesson 43 doing its job. Start the server with `-max-per-ip 0 -rate 0`.
