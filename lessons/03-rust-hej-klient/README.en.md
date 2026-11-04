# Lesson 03 – Hello, Rust!

You have a server. Time to switch languages and build the **client** – the
program the player will eventually see: a colour map, a status bar, a log
window. All of that is coming, but first we have to be able to talk to the
server.

## Why two languages?

| | Go (the server) | Rust (the client) |
|---|---|---|
| Strength | simple, quick to write, fantastic for networking and concurrency | full control, no GC pauses, the compiler catches bugs early |
| Memory | garbage collector (GC) | ownership and borrowing – checked at compile time |
| Errors | `error` as an ordinary return value | `Result<T, E>` and `?` |
| Missing value | `nil` | `Option<T>` – `Some(x)` or `None` |

You have already seen how Go handles errors. Rust does something similar –
but the compiler **forces** you to handle them.

## Rust in five minutes

```rust
pub struct Point { x: i32, y: i32 }        // struct

impl Point {                               // methods live in an impl block
    pub fn new(x: i32, y: i32) -> Point {  // "associated function" (like a constructor)
        Point { x, y }
    }
    pub fn len2(&self) -> i32 {            // &self = borrows the point (reading)
        self.x * self.x + self.y * self.y  // the last expression without ; is returned
    }
    pub fn move_right(&mut self) {         // &mut self = borrows it for changing
        self.x += 1;
    }
}

fn parse_port(s: &str) -> Result<u16, std::num::ParseIntError> {
    let port: u16 = s.parse()?;            // ? = return the error at once if it was Err
    Ok(port)
}
```

**Go ↔ Rust:** `x, err := f(); if err != nil { return err }` in Go
corresponds to `let x = f()?;` in Rust.

### Ownership in one sentence

Every value has exactly one owner; you can *lend* it out, either to many
readers (`&T`) or to one single writer (`&mut T`) – never both at once.

### `String` and `&str`

`String` is an owned, growable text buffer. `&str` is a borrowed piece of
text. Functions usually take `&str` as an argument (works for both) and
return `String`.

## Task

The course tool has created `workspace/client/src/net.rs` for you.

1. Add `pub mod net;` to `src/lib.rs` so that the module becomes part of the
   library. (Rust does not find files automatically – you have to declare the
   module.)
2. Implement the three methods in `net.rs` (replace `todo!()`).
3. Write `src/main.rs` so that the client:
   - connects to the address in the first argument, or `127.0.0.1:4000`,
   - prints the welcome line,
   - in a loop reads a line from the keyboard, sends it and prints the reply.

The tests (`tests/mudschool_03.rs`) start a fake server and only test
`Connection`, but do `main.rs` too – otherwise you have no client to run!

> The client's own texts (e.g. "Connected to") are in `src/messages.rs` in
> the course language: `println!("{} {addr}", messages::CONNECTED_TO);`

## Try it yourself

```
./mudschool run server                    # terminal 1
./mudschool run client                    # terminal 2
./mudschool run client 127.0.0.1:4000     # the same thing
```

## Think about

- Why does `read_line` have to take `&mut self` and not `&self`?
- `TcpStream::try_clone` gives you two handles to the *same* socket. Why do
  we need two? (Hint: `BufReader` *owns* the stream it reads from.)
- What happens in the client if the server sends two lines but we only read
  one? We solve that in lesson 06.

## Challenges

- Change the return type of `main` to `Result<(), Box<dyn std::error::Error>>`
  and read about why that lets you use `?` on different error types.
- Run `cargo clippy` and fix what it finds.
