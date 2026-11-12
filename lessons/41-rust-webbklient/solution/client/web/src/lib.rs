//! The browser client (lesson 41): the core, compiled to WebAssembly, draws
//! the zone on a `<canvas>`. The page (server/web/mud.js) owns the WebSocket
//! and the canvas; this crate owns the game state and decides what every
//! cell looks like.
//!
//! The contract between them – the ABI, version 1 – is the same idea as the
//! plugins in lesson 33, with the browser as the host:
//!
//! | exports                        | meaning                                  |
//! |--------------------------------|------------------------------------------|
//! | `abi_version() -> i32`         | 1                                        |
//! | `alloc(len) -> ptr`, `dealloc` | memory for the page to write lines into  |
//! | `on_line(ptr, len)`            | a line from the server                   |
//! | `frame(cols, rows, t) -> ptr`  | cols×rows cells of CELL_BYTES each       |
//! | `click(col, row)`              | a click on the canvas, in cells          |
//! | `tick(t)`                      | time passes (t in ms); walks take steps  |
//!
//! and it imports `mud.send(ptr, len)`: a command for the server.

use std::collections::VecDeque;

use mud_core::light::{shade, zone_light};
use mud_core::path::{find_path, step_command};
use mud_core::protocol::{parse_line, ServerMsg, Status};
use mud_core::render::{render_zone, Rgb};
use mud_core::zone::{camera, ZoneView};

/// The version of the contract with mud.js.
pub const ABI_VERSION: i32 = 1;

/// One cell in a frame: the character (u32, little endian), the foreground
/// r, g, b, the background r, g, b, 1 if there is a background, and a
/// spare byte – 12 bytes, so that every cell starts at a multiple of 4.
pub const CELL_BYTES: usize = 12;

/// How long a clicked walk waits between steps, in milliseconds.
pub const WALK_MS: f64 = 150.0;

/// Everything the browser client knows.
#[derive(Debug, Default)]
pub struct WebClient {
    pub zone: Option<ZoneView>,
    pub status: Option<Status>,
    /// The zone cell drawn first in the last frame, and how many cells
    /// from the canvas' top left it was drawn (a small zone is centred).
    origin: (usize, usize),
    pad: (usize, usize),
    /// Steps still to walk (commands), and when the last one was sent.
    walk: VecDeque<String>,
    last_step: f64,
    /// Commands for the server that the page hasn't taken yet.
    outbox: Vec<String>,
    /// The last frame's bytes – they stay here so the page can read them.
    frame: Vec<u8>,
}

impl WebClient {
    pub fn new() -> WebClient {
        WebClient::default()
    }

    /// A line from the server. Text and prompts are the page's business; the
    /// zone, its changes and the status are ours.
    pub fn on_line(&mut self, line: &str) {
        match parse_line(line) {
            ServerMsg::Zone { zone } => {
                if self.zone.as_ref().is_some_and(|z| z.room != zone.room) {
                    self.walk.clear(); // the walk belonged to the old room
                }
                self.zone = Some(ZoneView::from(zone));
            }
            ServerMsg::Delta { delta } => {
                if let Some(z) = &mut self.zone {
                    z.apply(delta);
                }
            }
            ServerMsg::Status { status } => self.status = Some(status),
            _ => {}
        }
    }

    /// Draws cols×rows cells, the player in the middle, at time t_ms, and
    /// returns them as bytes (CELL_BYTES per cell, row by row). Cells with
    /// nothing in them are all zero bytes – the page skips character 0.
    pub fn frame(&mut self, cols: usize, rows: usize, t_ms: u64) -> &[u8] {
        self.frame.clear();
        self.frame.resize(cols * rows * CELL_BYTES, 0);
        let Some(z) = &self.zone else {
            return &self.frame;
        };
        let (sx, sy) = z.self_pos().unwrap_or((z.width / 2, z.height / 2));
        let (x0, y0) = (camera(sx, cols, z.width), camera(sy, rows, z.height));
        // A zone smaller than the canvas is drawn in the middle of it.
        let pad = (cols.saturating_sub(z.width) / 2, rows.saturating_sub(z.height) / 2);
        self.origin = (x0, y0);
        self.pad = pad;
        let terrain = self.status.as_ref().map_or("", |s| s.terrain.as_str());
        let grid = render_zone(z, x0, y0, cols - pad.0, rows - pad.1);
        for (y, row) in grid.iter().enumerate() {
            for (x, cell) in row.iter().enumerate() {
                let (zx, zy) = (x0 + x, y0 + y);
                let light = if zx < z.width && zy < z.height {
                    zone_light(z, terrain, zx, zy, t_ms)
                } else {
                    Rgb(255, 255, 255)
                };
                let o = ((y + pad.1) * cols + x + pad.0) * CELL_BYTES;
                let b = &mut self.frame[o..o + CELL_BYTES];
                b[0..4].copy_from_slice(&(cell.glyph as u32).to_le_bytes());
                let fg = shade(cell.fg, light);
                b[4..7].copy_from_slice(&[fg.0, fg.1, fg.2]);
                if let Some(bg) = cell.bg {
                    let bg = shade(bg, light);
                    b[7..10].copy_from_slice(&[bg.0, bg.1, bg.2]);
                    b[10] = 1;
                }
            }
        }
        &self.frame
    }

    /// A click on the canvas at (col, row): walk there, the way the terminal
    /// client does (lesson 30) – A* here, one step at a time in `tick`.
    pub fn click(&mut self, col: usize, row: usize) {
        self.walk.clear();
        let Some(z) = &self.zone else { return };
        let Some(start) = z.self_pos() else { return };
        if col < self.pad.0 || row < self.pad.1 {
            return; // beside the zone
        }
        let target = (self.origin.0 + col - self.pad.0, self.origin.1 + row - self.pad.1);
        let Some(path) = find_path(z, start, target) else { return };
        let mut prev = start;
        for cell in path {
            if let Some(cmd) = step_command(prev, cell) {
                self.walk.push_back(cmd.to_string());
            }
            prev = cell;
        }
    }

    /// Time passes: sends the next step of a walk when it is time.
    pub fn tick(&mut self, t_ms: f64) {
        if self.walk.is_empty() || t_ms - self.last_step < WALK_MS {
            return;
        }
        self.last_step = t_ms;
        if let Some(cmd) = self.walk.pop_front() {
            self.outbox.push(cmd);
        }
    }

    /// The commands to send to the server, oldest first.
    pub fn take_outbox(&mut self) -> Vec<String> {
        std::mem::take(&mut self.outbox)
    }
}

// ---- the ABI: only when compiled to WebAssembly ----------------------------

#[cfg(target_arch = "wasm32")]
mod abi {
    use super::*;
    use std::cell::RefCell;

    #[link(wasm_import_module = "mud")]
    extern "C" {
        fn send(ptr: *const u8, len: usize);
    }

    thread_local! {
        // One client per page. WebAssembly in the browser has one thread,
        // so a thread_local is a safe home for it.
        static CLIENT: RefCell<WebClient> = RefCell::new(WebClient::new());
    }

    fn flush(c: &mut WebClient) {
        for cmd in c.take_outbox() {
            // SAFETY: the page reads exactly len bytes from ptr during the call.
            unsafe { send(cmd.as_ptr(), cmd.len()) };
        }
    }

    #[no_mangle]
    pub extern "C" fn abi_version() -> i32 {
        ABI_VERSION
    }

    #[no_mangle]
    pub extern "C" fn alloc(len: usize) -> *mut u8 {
        let mut buf = Vec::<u8>::with_capacity(len);
        let ptr = buf.as_mut_ptr();
        std::mem::forget(buf); // the page owns it until dealloc
        ptr
    }

    /// # Safety
    /// ptr and len must come from `alloc`.
    #[no_mangle]
    pub unsafe extern "C" fn dealloc(ptr: *mut u8, len: usize) {
        drop(Vec::from_raw_parts(ptr, 0, len));
    }

    /// # Safety
    /// The page has written len bytes of UTF-8 at ptr (from `alloc`).
    #[no_mangle]
    pub unsafe extern "C" fn on_line(ptr: *const u8, len: usize) {
        let bytes = std::slice::from_raw_parts(ptr, len);
        let line = String::from_utf8_lossy(bytes);
        CLIENT.with_borrow_mut(|c| c.on_line(&line));
    }

    #[no_mangle]
    pub extern "C" fn frame(cols: usize, rows: usize, t_ms: f64) -> *const u8 {
        CLIENT.with_borrow_mut(|c| c.frame(cols, rows, t_ms as u64).as_ptr())
    }

    #[no_mangle]
    pub extern "C" fn click(col: usize, row: usize) {
        CLIENT.with_borrow_mut(|c| {
            c.click(col, row);
            flush(c);
        });
    }

    #[no_mangle]
    pub extern "C" fn tick(t_ms: f64) {
        CLIENT.with_borrow_mut(|c| {
            c.tick(t_ms);
            flush(c);
        });
    }
}
