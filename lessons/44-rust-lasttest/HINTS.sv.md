# Tips

## percentile

```rust
let rank = ((p / 100.0) * sorted.len() as f64).ceil() as usize;
sorted[rank.clamp(1, sorted.len()) - 1]
```

## bot_name

Skriv `id` i bas 26 med bokstäverna `a`–`z` (`insert(0, …)` lägger varje
siffra först), sedan `format!("Bot{letters}")`. 0 → `Bota`, 26 → `Botba`.

## run_bot – att ansluta med en tidsgräns

```rust
let stream = match timeout(Duration::from_secs(5), TcpStream::connect(&plan.addr)).await {
    Ok(Ok(s)) => s,
    Ok(Err(e)) => return drop(events.send(Event::Failed(e.to_string())).await),
    Err(_) => return drop(events.send(Event::Failed("connect: timeout".into())).await),
};
let (read, mut write) = stream.into_split();
let mut lines = BufReader::new(read).lines();
```

`into_split` ger en läsare och en skrivare som kan användas samtidigt.

## run_bot – echot

Kom ihåg varje echo du skickar, `HashMap<String, Instant>`; när en
textframes text är en av dem, `remove` den och rapportera `sent.elapsed()`.
Med en långsam server är flera på väg samtidigt.

## run – samla in till en tidsgräns

```rust
let deadline = Instant::now() + plan.duration;
loop {
    tokio::select! {
        Some(ev) = events.recv() => record(&mut report, ev),
        _ = sleep_until(deadline) => break,
    }
}
```

`Some(ev) = …` är ett mönster: om `recv()` returnerar `None` (alla sändare
borta) hoppas den grenen bara över.

## run – att stoppa

`stop_tx.send(true)` väcker varje bots `stop.changed()`. Vänta på tasken med
`timeout(Duration::from_secs(5), handle).await` – en bot som fastnat på en
död anslutning får inte låsa hela testet.

## "För många anslutningar från din adress"

Det är lektion 43 som gör sitt jobb. Starta servern med `-max-per-ip 0 -rate 0`.
