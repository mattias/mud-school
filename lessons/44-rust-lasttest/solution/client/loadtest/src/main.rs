//! The load tester's command line (lesson 44).
//!
//! Start the server without limits and without saving first:
//!   ./mudschool run server -db "" -max-per-ip 0 -rate 0
//! then, in client/:
//!   cargo run --release -p mud-loadtest -- --bots 500 --seconds 30

use std::time::Duration;

use clap::Parser;
use mud_loadtest::{run, Plan};

#[derive(Parser)]
#[command(about = "Many bots play at once; how quickly does the server answer?")]
struct Args {
    /// The server's telnet address.
    #[arg(long, default_value = "127.0.0.1:4000")]
    addr: String,
    /// How many bots.
    #[arg(long, default_value_t = 100)]
    bots: usize,
    /// How long the test runs.
    #[arg(long, default_value_t = 20)]
    seconds: u64,
    /// Milliseconds a bot waits between commands.
    #[arg(long, default_value_t = 500)]
    think_ms: u64,
    /// Seconds over which the bots connect.
    #[arg(long, default_value_t = 5)]
    ramp: u64,
}

#[tokio::main]
async fn main() {
    let a = Args::parse();
    println!("{} bots against {} for {} s ...", a.bots, a.addr, a.seconds);
    let report = run(Plan {
        addr: a.addr,
        bots: a.bots,
        duration: Duration::from_secs(a.seconds),
        think: Duration::from_millis(a.think_ms),
        ramp: Duration::from_secs(a.ramp),
    })
    .await;
    println!("{report}");
}
