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
    // TODO(lesson 44): "nearest rank": the value at rank ceil(p/100 × n),
    // counted from 1 and kept within 1..=n. 0.0 for no values.
    let _ = (sorted, p);
    0.0
}

/// A name for bot number id: "Bot" and letters (the server only accepts
/// letters), different for every id.
pub fn bot_name(id: usize) -> String {
    // TODO(lesson 44): "Bot" and id written with the letters a–z (base 26).
    format!("Bot{id}")
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
    // TODO(lesson 44): connect (timeout 5 s; Event::Failed on errors), Event::Connected,
    // send HELLO and the name, then loop with tokio::select! over: a line from the
    // server (zone → LoggedIn + "tutorial skip", delta, prompt → a password, the
    // text of an echo → Event::Reply), the time for the next command (only when
    // logged in: next_command, Event::Sent, think + jitter) and stop.changed()
    // (send quit). A closed connection: Event::Disconnected.
    let _ = (id, plan, events, &mut stop);
    todo!("lesson 44: run_bot")
}

/// Runs a load test and collects what the bots report.
pub async fn run(plan: Plan) -> Report {
    // TODO(lesson 44): an mpsc channel for events and a watch channel for "stop";
    // spawn the bots spread over plan.ramp (in a task of its own); collect events
    // with tokio::select! until plan.duration has passed; then stop the bots,
    // wait for them (with a timeout), and sort the answer times.
    let _ = plan;
    todo!("lesson 44: run")
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
