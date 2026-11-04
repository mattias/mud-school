# Lektion 03 – Hej, Rust!

Nu har du en server. Dags att byta språk och bygga **klienten** – det
program som spelaren till slut kommer att se: karta i färg, statusrad,
loggfönster. Allt det kommer, men först måste vi kunna prata med servern.

## Varför två språk?

| | Go (servern) | Rust (klienten) |
|---|---|---|
| Styrka | enkelt, snabbt att skriva, fantastiskt för nätverk och samtidighet | full kontroll, inga GC-pauser, kompilatorn fångar buggar tidigt |
| Minne | skräpsamlare (GC) | ägarskap och lån – kontrolleras vid kompilering |
| Fel | `error` som vanligt returvärde | `Result<T, E>` och `?` |
| Saknat värde | `nil` | `Option<T>` – `Some(x)` eller `None` |

Du har redan sett hur Go hanterar fel. Rust gör liknande – men kompilatorn
**tvingar** dig att hantera dem.

## Rust på fem minuter

```rust
pub struct Point { x: i32, y: i32 }        // struct

impl Point {                               // metoder ligger i ett impl-block
    pub fn new(x: i32, y: i32) -> Point {  // "associerad funktion" (som en konstruktor)
        Point { x, y }
    }
    pub fn len2(&self) -> i32 {            // &self = lånar punkten (läsning)
        self.x * self.x + self.y * self.y  // sista uttrycket utan ; returneras
    }
    pub fn move_right(&mut self) {         // &mut self = lånar för ändring
        self.x += 1;
    }
}

fn parse_port(s: &str) -> Result<u16, std::num::ParseIntError> {
    let port: u16 = s.parse()?;            // ? = returnera felet direkt om det blev Err
    Ok(port)
}
```

**Go ↔ Rust:** `x, err := f(); if err != nil { return err }` i Go motsvaras
av `let x = f()?;` i Rust.

### Ägarskap i en mening

Varje värde har exakt en ägare; du kan *låna* ut det, antingen till många
läsare (`&T`) eller till en enda skrivare (`&mut T`) – aldrig båda samtidigt.

### `String` och `&str`

`String` är en ägd, växbar textbuffert. `&str` är en lånad bit text. Funktioner
tar oftast `&str` som argument (fungerar för båda) och returnerar `String`.

## Uppgift

Kursverktyget har skapat `workspace/client/src/net.rs` åt dig.

1. Lägg till `pub mod net;` i `src/lib.rs` så att modulen blir en del av
   biblioteket. (I Rust hittas filer inte automatiskt – du måste deklarera
   modulen.)
2. Implementera de tre metoderna i `net.rs` (byt ut `todo!()`).
3. Skriv `src/main.rs` så att klienten:
   - ansluter till adressen i första argumentet, eller `127.0.0.1:4000`,
   - skriver ut välkomstraden,
   - i en loop läser en rad från tangentbordet, skickar den och skriver ut
     svaret.

Testerna (`tests/mudschool_03.rs`) startar en låtsasserver och testar bara
`Connection`, men gör `main.rs` också – annars har du ingen klient att köra!

## Testa själv

```
./mudschool run server                    # terminal 1
./mudschool run client                    # terminal 2
./mudschool run client 127.0.0.1:4000     # samma sak
```

## Fundera på

- Varför måste `read_line` ta `&mut self` och inte `&self`?
- `TcpStream::try_clone` ger dig två handtag till *samma* socket. Varför
  behöver vi två? (Tips: `BufReader` *äger* strömmen den läser från.)
- Vad händer i klienten om servern skickar två rader men vi bara läser en?
  Vi löser det i lektion 06.

## Utmaningar

- Byt `main`s returtyp till `Result<(), Box<dyn std::error::Error>>` och läs
  om varför det låter dig använda `?` på olika feltyper.
- Kör `cargo clippy` och åtgärda det den hittar.
