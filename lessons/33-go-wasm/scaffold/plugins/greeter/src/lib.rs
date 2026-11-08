//! A plugin for the MUD School server, written in Rust and compiled to
//! WebAssembly (lesson 33). The server is written in Go – the two only meet
//! through the ABI below: a few functions with numbers as arguments, and a
//! shared block of memory.
//!
//! ABI version 1:
//!   the plugin exports  abi_version() -> i32           always 1
//!                       alloc(len) -> ptr              memory the host may write into
//!                       dealloc(ptr, len)              give it back
//!                       on_command(ptr, len) -> i32    1 = handled, 0 = not mine
//!   the host offers     mud.send(ptr, len)             text to the player who typed the command
//!                       mud.log(ptr, len)              a line in the server's log
//! on_command gets "player<TAB>verb<TAB>args" as UTF-8.

#[link(wasm_import_module = "mud")]
extern "C" {
    fn send(ptr: *const u8, len: usize);
    fn log(ptr: *const u8, len: usize);
}

/// Sends text to the player who typed the command.
fn say(text: &str) {
    unsafe { send(text.as_ptr(), text.len()) }
}

/// Writes a line in the server's log.
fn note(text: &str) {
    unsafe { log(text.as_ptr(), text.len()) }
}

#[no_mangle]
pub extern "C" fn abi_version() -> i32 {
    1
}

#[no_mangle]
pub extern "C" fn alloc(len: usize) -> *mut u8 {
    let mut buf = Vec::<u8>::with_capacity(len);
    let ptr = buf.as_mut_ptr();
    std::mem::forget(buf); // the host owns it until dealloc
    ptr
}

#[no_mangle]
pub extern "C" fn dealloc(ptr: *mut u8, len: usize) {
    unsafe { drop(Vec::from_raw_parts(ptr, 0, len)) }
}

#[no_mangle]
pub extern "C" fn on_command(ptr: *const u8, len: usize) -> i32 {
    let bytes = unsafe { std::slice::from_raw_parts(ptr, len) };
    let input = std::str::from_utf8(bytes).unwrap_or("");
    let mut parts = input.splitn(3, '\t');
    let player = parts.next().unwrap_or("");
    let verb = parts.next().unwrap_or("");
    let args = parts.next().unwrap_or("");

    match verb {
        "greet" => {
            say(&format!("Hello, {player}! Greetings from a plugin written in Rust."));
            note(&format!("greeted {player}"));
            1
        }
        "reverse" => {
            let reversed: String = args.chars().rev().collect();
            say(&reversed);
            1
        }
        // A plugin that never finishes – the server must stop it (a time limit).
        "forever" => loop {
            std::hint::spin_loop();
        },
        _ => 0, // not a command of ours
    }
}
