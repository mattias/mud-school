//! The load tester (lesson 44): many bots play at once, and we measure how
//! quickly the server answers. Each bot is an async *task* – cheap like a
//! goroutine – and `tokio::select!` waits for whichever comes first, just
//! like Go's `select`.

use std::collections::HashMap;
use std::fmt;
use std::sync::Arc;
use std::time::Duration;

use mud_core::path::{find_path, step_command};
use mud_core::protocol::{parse_line, ServerMsg, HELLO};
use mud_core::zone::ZoneView;
use tokio::io::{AsyncBufReadExt, AsyncWriteExt, BufReader};
use tokio::net::TcpStream;
use tokio::sync::{mpsc, watch};
use tokio::time::{sleep, sleep_until, timeout, Instant};

/// What to do.
#[derive(Debug, Clone)]
pub struct Plan {
    pub addr: String,
    pub bots: usize,
    pub duration: Duration,
    /// How long a bot waits between commands.
    pub think: Duration,
    /// The bots don't all connect at once: they start over this long.
    pub ramp: Duration,
}

/// What a bot tells the collector.
#[derive(Debug, Clone, PartialEq)]
pub enum Event {
    Connected,
    Failed(String),
    LoggedIn,
    Sent,
    /// An answer to an "echo", and how long it took.
    Reply(Duration),
    Disconnected,
}

/// The result of a load test.
#[derive(Debug, Clone, Default, PartialEq)]
pub struct Report {
    pub bots: usize,
    pub connected: usize,
    pub failed: usize,
    pub logged_in: usize,
    pub commands: usize,
    pub replies: usize,
    pub disconnected: usize,
    /// Answer times in milliseconds, sorted.
    pub latencies_ms: Vec<f64>,
    pub seconds: f64,
    /// Why connections failed (the first few).
    pub errors: Vec<String>,
}

/// The value below which p percent (0–100) of the sorted values lie
/// ("nearest rank"). 0 for no values.
pub fn percentile(sorted: &[f64], p: f64) -> f64 {
    if sorted.is_empty() {
        return 0.0;
    }
    let rank = ((p / 100.0) * sorted.len() as f64).ceil() as usize;
    sorted[rank.clamp(1, sorted.len()) - 1]
}

/// A name for bot number id: "Bot" and letters (the server only accepts
/// letters), different for every id.
pub fn bot_name(id: usize) -> String {
    let mut letters = String::new();
    let mut n = id;
    loop {
        letters.insert(0, (b'a' + (n % 26) as u8) as char);
        n /= 26;
        if n == 0 {
            break;
        }
    }
    format!("Bot{letters}")
}

impl fmt::Display for Report {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        let l = &self.latencies_ms;
        writeln!(f, "bots:          {} ({} connected, {} failed, {} logged in)", self.bots, self.connected, self.failed, self.logged_in)?;
        writeln!(f, "commands:      {} in {:.1} s = {:.0}/s", self.commands, self.seconds, self.commands as f64 / self.seconds.max(0.001))?;
        writeln!(f, "disconnected:  {}", self.disconnected)?;
        write!(
            f,
            "echo answers:  {} – p50 {:.1} ms, p95 {:.1} ms, p99 {:.1} ms, max {:.1} ms",
            self.replies,
            percentile(l, 50.0),
            percentile(l, 95.0),
            percentile(l, 99.0),
            l.last().copied().unwrap_or(0.0)
        )?;
        for e in &self.errors {
            write!(f, "\nerror:         {e}")?;
        }
        Ok(())
    }
}

/// A small, fast random number generator (xorshift) – each bot has its own,
/// so no locking and no extra crate.
struct Rng(u64);

impl Rng {
    fn next(&mut self) -> u64 {
        self.0 ^= self.0 << 13;
        self.0 ^= self.0 >> 7;
        self.0 ^= self.0 << 17;
        self.0
    }
    fn below(&mut self, n: usize) -> usize {
        (self.next() % n.max(1) as u64) as usize
    }
}

/// The next command for a bot: every third an "echo" (to measure the answer
/// time), otherwise a step towards a random place in the zone (the core's
/// A*, lesson 30), or "look".
fn next_command(rng: &mut Rng, zone: Option<&ZoneView>, count: u64) -> String {
    if count.is_multiple_of(3) {
        return format!("echo ping-{count}");
    }
    if let Some(z) = zone {
        if let Some(me) = z.self_pos() {
            let target = (rng.below(z.width), rng.below(z.height));
            if let Some(path) = find_path(z, me, target) {
                if let Some(cmd) = path.first().and_then(|&next| step_command(me, next)) {
                    return cmd.to_string();
                }
            }
        }
    }
    "look".to_string()
}

/// One bot: connects, logs in, and plays until `stop` says so.
pub async fn run_bot(id: usize, plan: Arc<Plan>, events: mpsc::Sender<Event>, mut stop: watch::Receiver<bool>) {
    let stream = match timeout(Duration::from_secs(5), TcpStream::connect(&plan.addr)).await {
        Ok(Ok(s)) => s,
        Ok(Err(e)) => return drop(events.send(Event::Failed(e.to_string())).await),
        Err(_) => return drop(events.send(Event::Failed("connect: timeout".into())).await),
    };
    let _ = events.send(Event::Connected).await;
    let (read, mut write) = stream.into_split();
    let mut lines = BufReader::new(read).lines();
    let name = bot_name(id);
    let mut rng = Rng(0x9E37_79B9_7F4A_7C15 ^ (id as u64 + 1).wrapping_mul(0x2545_F491_4F6C_DD1D));
    let mut zone: Option<ZoneView> = None;
    let mut logged_in = false;
    let mut pending: HashMap<String, Instant> = HashMap::new(); // echoes we wait for
    let mut count: u64 = 0;
    let mut next = Instant::now() + plan.think;

    if write.write_all(format!("{HELLO}\r\n{name}\r\n").as_bytes()).await.is_err() {
        let _ = events.send(Event::Disconnected).await;
        return;
    }
    loop {
        // Whichever comes first: a line from the server, time for the next
        // command, or the signal to stop. (Go: select with three cases.)
        tokio::select! {
            line = lines.next_line() => {
                let Ok(Some(line)) = line else {
                    let _ = events.send(Event::Disconnected).await;
                    return;
                };
                match parse_line(&line) {
                    ServerMsg::Zone { zone: z } => {
                        if !logged_in {
                            logged_in = true;
                            let _ = events.send(Event::LoggedIn).await;
                            // New players start in the tutorial (lesson 37).
                            let _ = write.write_all(b"tutorial skip\r\n").await;
                        }
                        zone = Some(ZoneView::from(z));
                    }
                    ServerMsg::Delta { delta } => {
                        if let Some(z) = &mut zone {
                            z.apply(delta);
                        }
                    }
                    ServerMsg::Prompt { .. } => {
                        // The server wants a password: run it with -db "" to avoid
                        // creating accounts – but this works too.
                        let _ = write.write_all(b"loadtest\r\n").await;
                    }
                    ServerMsg::Text { text, .. } => {
                        if let Some(sent) = pending.remove(text.trim()) {
                            let _ = events.send(Event::Reply(sent.elapsed())).await;
                        }
                    }
                    _ => {}
                }
            }
            _ = sleep_until(next), if logged_in => {
                count += 1;
                let cmd = next_command(&mut rng, zone.as_ref(), count);
                if let Some(token) = cmd.strip_prefix("echo ") {
                    pending.insert(token.to_string(), Instant::now());
                }
                if write.write_all(format!("{cmd}\r\n").as_bytes()).await.is_err() {
                    let _ = events.send(Event::Disconnected).await;
                    return;
                }
                let _ = events.send(Event::Sent).await;
                // A little jitter, so the bots don't march in step.
                let jitter = rng.below(plan.think.as_millis() as usize / 2 + 1) as u64;
                next = Instant::now() + plan.think + Duration::from_millis(jitter);
            }
            _ = stop.changed() => {
                let _ = write.write_all(b"quit\r\n").await;
                return;
            }
        }
    }
}

/// Runs a load test and collects what the bots report.
pub async fn run(plan: Plan) -> Report {
    let plan = Arc::new(plan);
    let (events_tx, mut events) = mpsc::channel(1024);
    let (stop_tx, stop_rx) = watch::channel(false);
    let start = Instant::now();
    let deadline = start + plan.duration;

    // Start the bots, spread out over plan.ramp, while collecting events.
    let starter = {
        let plan = plan.clone();
        tokio::spawn(async move {
            let mut bots = Vec::with_capacity(plan.bots);
            let gap = plan.ramp / plan.bots.max(1) as u32;
            for id in 0..plan.bots {
                bots.push(tokio::spawn(run_bot(id, plan.clone(), events_tx.clone(), stop_rx.clone())));
                sleep(gap).await;
            }
            bots
        })
    };

    let mut r = Report { bots: plan.bots, ..Report::default() };
    let record = |r: &mut Report, ev: Event| match ev {
        Event::Connected => r.connected += 1,
        Event::Failed(e) => {
            r.failed += 1;
            if r.errors.len() < 3 && !r.errors.contains(&e) {
                r.errors.push(e);
            }
        }
        Event::LoggedIn => r.logged_in += 1,
        Event::Sent => r.commands += 1,
        Event::Reply(d) => {
            r.replies += 1;
            r.latencies_ms.push(d.as_secs_f64() * 1000.0);
        }
        Event::Disconnected => r.disconnected += 1,
    };
    loop {
        tokio::select! {
            Some(ev) = events.recv() => record(&mut r, ev),
            _ = sleep_until(deadline) => break,
        }
    }
    r.seconds = start.elapsed().as_secs_f64();

    // Time is up: tell every bot, and wait (a little) for them to finish.
    let _ = stop_tx.send(true);
    if let Ok(Ok(bots)) = timeout(Duration::from_secs(5), starter).await {
        for b in bots {
            let _ = timeout(Duration::from_secs(5), b).await;
        }
    }
    while let Ok(ev) = events.try_recv() {
        if matches!(ev, Event::Failed(_) | Event::Connected | Event::LoggedIn) {
            record(&mut r, ev); // late arrivals still count
        }
    }
    r.latencies_ms.sort_by(|a, b| a.total_cmp(b));
    r
}

/// Runs a load test on a new tokio runtime – for code that isn't async
/// itself, like `main` without `#[tokio::main]`, or the tests.
pub fn run_blocking(plan: Plan) -> Report {
    tokio::runtime::Builder::new_multi_thread()
        .enable_all()
        .build()
        .expect("tokio runtime")
        .block_on(run(plan))
}
