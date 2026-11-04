//! Läser rader i en bakgrundstråd.

use std::io::BufRead;
use std::sync::mpsc::{self, Receiver};
use std::thread;

/// Startar en tråd som läser rader från `reader` och skickar dem (utan
/// radbrytning) på kanalen som returneras. När `reader` tar slut, eller ett
/// fel uppstår, avslutas tråden och kanalen stängs.
///
/// `R` kan vara vad som helst som går att läsa rader från (`BufRead`), som går
/// att skicka till en annan tråd (`Send`) och som inte lånar något som kan
/// försvinna under tiden (`'static`).
pub fn spawn_reader<R>(reader: R) -> Receiver<String>
where
    R: BufRead + Send + 'static,
{
    let (tx, rx) = mpsc::channel();
    // TODO(lektion 6): starta en tråd med thread::spawn(move || { ... })
    // som loopar över reader.lines() och skickar varje rad med tx.send(...).
    rx
}
