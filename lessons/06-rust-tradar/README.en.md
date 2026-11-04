# Lesson 06 – Threads and channels

Your client from lesson 03 has a problem: it reads *exactly one* line from the
server after every line you send. But in a MUD the server talks whenever it
wants – another player says something, a troll walks into the room. The client
has to listen and let you type **at the same time**.

## Goals

- Start threads with `std::thread::spawn` and understand `move` closures.
- Send data between threads with channels (`std::sync::mpsc`).
- Write a **generic** function with trait bounds: `R: BufRead + Send + 'static`.
- Understand why the Rust compiler can guarantee that threaded programs have no data races.

## Go ↔ Rust

You have already seen the same idea in Go:

| Go | Rust |
|---|---|
| `go f()` | `thread::spawn(move \|\| f())` |
| `ch := make(chan string)` | `let (tx, rx) = mpsc::channel::<String>();` |
| `ch <- "hi"` | `tx.send("hi".to_string())` |
| `v, ok := <-ch` | `rx.recv()` → `Ok(v)` or `Err(RecvError)` |
| `close(ch)` | drop every `tx` |
| `for v := range ch` | `for v in rx` |

The difference: Rust threads are real OS threads (more expensive than
goroutines), and the compiler **refuses** to compile code where two threads
can modify the same data without synchronization. That is what the `Send`
trait is about.

## Rust concepts

### move closures

A closure that is going to run in another thread has to **own** everything it
uses. `move` moves the variables in:

```rust
let name = String::from("Anna");
let handle = thread::spawn(move || {
    println!("hello {name}");   // name is now owned by the thread
});
// println!("{name}");          // error! name has been moved
handle.join().unwrap();
```

### Generics and trait bounds

```rust
fn print_all<R: BufRead>(reader: R) {    // works for ALL types that are BufRead
    for line in reader.lines() { ... }
}
```

`BufRead` is implemented by `BufReader<TcpStream>`, `Cursor<Vec<u8>>`, `&[u8]`,
`StdinLock` ... The tests take advantage of that: they feed your function text
from memory instead of a network. That is how you test network code without a
network!

- `Send` – the value may be moved to another thread.
- `'static` – the value borrows nothing that could disappear while the thread lives.

## Task

1. **`src/reader.rs`** (new file): implement `spawn_reader`. Add
   `pub mod reader;` to `lib.rs`.
2. **`src/net.rs`**: add a method `split(self)` that splits the connection
   into a reading half and a writing half:

   ```rust
   pub fn split(self) -> (BufReader<TcpStream>, Outgoing)
   ```

   `Outgoing` is a new `pub struct` that owns the writing `TcpStream` and has
   the method `send_line(&mut self, line: &str) -> io::Result<()>`. Let
   `Connection` use an `Outgoing` internally to avoid duplicated code.
3. **`src/main.rs`**: connect, `split`, start `spawn_reader` and a thread that
   prints everything that arrives on the channel. The main thread reads the
   keyboard and sends.

## Try it yourself

Start the server and two clients. Log in with different names and type `who`
in both. (In the next lesson they can talk to each other – then the
difference really shows.)

## Think about

- `split` takes `self` and not `&self`. What does that mean for the caller?
  What would happen if you tried to use `conn` after `conn.split()`?
- Why does the reader thread end by itself when the server closes the connection?
- Why do we need no `Mutex` here, when we needed one in the Go server?

## Challenges

- Replace the printing thread with a single loop in the main thread that
  receives *both* keyboard lines and server lines from the same channel, with
  an `enum Event { Input(String), Server(String), Disconnected }`. That is
  exactly the architecture we need for the user interface in lesson 13.
