# レッスン 06 – スレッドとチャネル

レッスン 03 のクライアントには問題があります。1 行送るたびに、サーバーから
*ちょうど 1 行* しか読みません。でも MUD では、サーバーは好きなときに話します。
他のプレイヤーが何か言ったり、トロルが部屋に入ってきたり。クライアントは
受信と入力を **同時に** できなければなりません。

## 目標

- `std::thread::spawn` でスレッドを起動し、`move` クロージャを理解する。
- チャネル（`std::sync::mpsc`）でスレッド間にデータを送る。
- トレイト境界付きの **ジェネリック** 関数を書く: `R: BufRead + Send + 'static`。
- Rust のコンパイラが、スレッドを使うプログラムにデータ競合がないことを
  保証できる理由を理解する。

## Go ↔ Rust

同じ考え方は Go ですでに見ています:

| Go | Rust |
|---|---|
| `go f()` | `thread::spawn(move \|\| f())` |
| `ch := make(chan string)` | `let (tx, rx) = mpsc::channel::<String>();` |
| `ch <- "hi"` | `tx.send("hi".to_string())` |
| `v, ok := <-ch` | `rx.recv()` → `Ok(v)` または `Err(RecvError)` |
| `close(ch)` | すべての `tx` を破棄（drop）する |
| `for v := range ch` | `for v in rx` |

違い: Rust のスレッドは本物の OS スレッド（ゴルーチンより高価）で、2 つの
スレッドが同期なしに同じデータを変更できるコードは、コンパイラが **コンパイルを
拒否** します。`Send` トレイトはそのためのものです。

## Rust の概念

### move クロージャ

別スレッドで動くクロージャは、使うものをすべて **所有** していなければなりません。
`move` で変数を中に移します:

```rust
let name = String::from("Anna");
let handle = thread::spawn(move || {
    println!("hello {name}");   // name は今やスレッドのもの
});
// println!("{name}");          // エラー！ name は移動済み
handle.join().unwrap();
```

### ジェネリクスとトレイト境界

```rust
fn print_all<R: BufRead>(reader: R) {    // BufRead なら「どんな型でも」動く
    for line in reader.lines() { ... }
}
```

`BufRead` は `BufReader<TcpStream>`、`Cursor<Vec<u8>>`、`&[u8]`、`StdinLock` などが
実装しています。テストはこれを利用し、ネットワークの代わりにメモリ上のテキストを
関数に渡します。ネットワークなしでネットワークのコードをテストする方法です！

- `Send` – 値を別のスレッドへ移動してよい。
- `'static` – スレッドが動いている間に消えるかもしれないものを借りていない。

## 課題

1. **`src/reader.rs`**（新しいファイル）: `spawn_reader` を実装する。
   `lib.rs` に `pub mod reader;` を追加する。
2. **`src/net.rs`**: 接続を読み取り側と書き込み側に分けるメソッド `split(self)` を
   追加する:

   ```rust
   pub fn split(self) -> (BufReader<TcpStream>, Outgoing)
   ```

   `Outgoing` は書き込み用の `TcpStream` を所有する新しい `pub struct` で、
   メソッド `send_line(&mut self, line: &str) -> io::Result<()>` を持ちます。
   コードの重複を避けるため、`Connection` も内部で `Outgoing` を使うとよいでしょう。
3. **`src/main.rs`**: 接続し、`split` し、`spawn_reader` と、チャネルに届いたものを
   すべて表示するスレッドを起動する。メインスレッドはキーボードを読んで送信する。

## 自分で試す

サーバーとクライアントを 2 つ起動します。別々の名前でログインし、両方で `who` と
入力してみましょう。（次のレッスンでは互いに話せるようになり、違いがはっきりします。）

## 考えてみよう

- `split` は `&self` ではなく `self` を取ります。呼び出し側にとってどういう意味？
  `conn.split()` の後に `conn` を使おうとしたらどうなる？
- サーバーが接続を閉じると、読み取りスレッドが自然に終わるのはなぜ？
- Go のサーバーでは `Mutex` が必要だったのに、ここで不要なのはなぜ？

## チャレンジ

- 表示用スレッドをやめ、メインスレッドの 1 つのループで、キーボードの行と
  サーバーの行の *両方* を同じチャネルから受け取るようにする。
  `enum Event { Input(String), Server(String), Disconnected }` を使います。
  これはまさにレッスン 13 の画面に必要な構成です。
