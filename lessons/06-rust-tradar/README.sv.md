# Lektion 06 – Trådar och kanaler

Din klient från lektion 03 har ett problem: den läser *exakt en* rad från
servern efter varje rad du skickar. Men i en MUD pratar servern när den vill
– någon annan spelare säger något, ett troll kliver in i rummet. Klienten
måste lyssna och låta dig skriva **samtidigt**.

## Mål

- Starta trådar med `std::thread::spawn` och förstå `move`-closures.
- Skicka data mellan trådar med kanaler (`std::sync::mpsc`).
- Skriva en **generisk** funktion med trait bounds: `R: BufRead + Send + 'static`.
- Förstå varför Rust-kompilatorn kan garantera att trådprogram saknar data races.

## Go ↔ Rust

Du har redan sett samma idé i Go:

| Go | Rust |
|---|---|
| `go f()` | `thread::spawn(move \|\| f())` |
| `ch := make(chan string)` | `let (tx, rx) = mpsc::channel::<String>();` |
| `ch <- "hej"` | `tx.send("hej".to_string())` |
| `v, ok := <-ch` | `rx.recv()` → `Ok(v)` eller `Err(RecvError)` |
| `close(ch)` | släpp (drop) alla `tx` |
| `for v := range ch` | `for v in rx` |

Skillnaden: Rust-trådar är riktiga OS-trådar (dyrare än goroutines), och
kompilatorn **vägrar** kompilera kod där två trådar kan ändra samma data
utan synkronisering. Det är det som traiten `Send` handlar om.

## Rust-koncept

### move-closures

En closure som ska köras i en annan tråd måste **äga** allt den använder.
`move` flyttar in variablerna:

```rust
let name = String::from("Anna");
let handle = thread::spawn(move || {
    println!("hej {name}");     // name ägs nu av tråden
});
// println!("{name}");          // fel! name har flyttats
handle.join().unwrap();
```

### Generics och trait bounds

```rust
fn print_all<R: BufRead>(reader: R) {    // fungerar för ALLA typer som är BufRead
    for line in reader.lines() { ... }
}
```

`BufRead` implementeras av `BufReader<TcpStream>`, `Cursor<Vec<u8>>`, `&[u8]`,
`StdinLock` ... Testerna utnyttjar det: de matar din funktion med text från
minnet istället för ett nätverk. Så kan man testa nätverkskod utan nätverk!

- `Send` – värdet får flyttas till en annan tråd.
- `'static` – värdet lånar inget som kan försvinna medan tråden lever.

## Uppgift

1. **`src/reader.rs`** (ny fil): implementera `spawn_reader`. Lägg till
   `pub mod reader;` i `lib.rs`.
2. **`src/net.rs`**: lägg till en metod `split(self)` som delar upp
   anslutningen i en läshalva och en skrivhalva:

   ```rust
   pub fn split(self) -> (BufReader<TcpStream>, Outgoing)
   ```

   `Outgoing` är en ny `pub struct` som äger skriv-`TcpStream`:en och har
   metoden `send_line(&mut self, line: &str) -> io::Result<()>`. Låt gärna
   `Connection` använda en `Outgoing` internt så slipper du duplicerad kod.
3. **`src/main.rs`**: anslut, `split`, starta `spawn_reader` och en tråd som
   skriver ut allt som kommer på kanalen. Huvudtråden läser tangentbordet och
   skickar.

## Testa själv

Starta servern och två klienter. Logga in med olika namn och skriv `who` i
båda. (I nästa lektion kan de prata med varandra – då märks skillnaden.)

## Fundera på

- `split` tar `self` och inte `&self`. Vad betyder det för den som anropar?
  Vad skulle hända om du försökte använda `conn` efter `conn.split()`?
- Varför avslutas läs-tråden av sig själv när servern stänger anslutningen?
- Varför behövs inget `Mutex` här, när vi behövde ett i Go-servern?

## Utmaningar

- Byt utskriftstråden mot en enda loop i huvudtråden som tar emot *både*
  tangentbordsrader och serverrader från samma kanal, med en
  `enum Event { Input(String), Server(String), Disconnected }`.
  Det är precis den arkitektur vi behöver för det grafiska gränssnittet i
  lektion 13.
