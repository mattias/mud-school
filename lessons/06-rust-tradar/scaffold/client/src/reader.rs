//! Reads lines in a background thread.

use std::io::BufRead;
use std::sync::mpsc::{self, Receiver};
use std::thread;

/// Starts a thread that reads lines from `reader` and sends them (without
/// line breaks) on the returned channel. When `reader` ends, or an error
/// occurs, the thread finishes and the channel is closed.
///
/// `R` can be anything you can read lines from (`BufRead`), that can be
/// sent to another thread (`Send`) and that does not borrow anything that
/// might disappear meanwhile (`'static`).
pub fn spawn_reader<R>(reader: R) -> Receiver<String>
where
    R: BufRead + Send + 'static,
{
    let (tx, rx) = mpsc::channel();
    // TODO(lesson 6): start a thread with thread::spawn(move || { ... })
    // that loops over reader.lines() and sends every line with tx.send(...).
    rx
}
