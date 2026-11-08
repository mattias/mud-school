# Lesson 33 – Plugins in WebAssembly

This is where the course's two languages truly meet: a plugin written in
**Rust**, compiled to **WebAssembly**, running **inside the Go server**.

```
workspace/plugins/greeter/src/lib.rs   (Rust)
        │  cargo build --release --target wasm32-unknown-unknown
        ▼
server/plugins/greeter.wasm            (WebAssembly: portable, sandboxed bytecode)
        │  loaded by wazero – a WebAssembly runtime written in pure Go
        ▼
"greet"  ->  Hello, Anna! Greetings from a plugin written in Rust.
```

Why WebAssembly instead of Lua (lesson 32)? It is fast (compiled), it can be
written in any language that compiles to it (Rust, C, Zig, Go, …), and it is
**isolated**: a plugin can only touch its own memory and the functions the
server chooses to give it. A crashing or endless plugin can't take the server
down.

## Goals

- Design an **ABI** – the contract between host and plugin – out of nothing but
  numbers and a shared block of memory.
- Pass strings across the boundary: allocate in the plugin, write from the host,
  read with pointer + length.
- Limit untrusted code: memory pages, a time limit, a version check.
- Build a Rust library for another target (`wasm32-unknown-unknown`) with `crate-type = ["cdylib"]`.

## The ABI (version 1)

WebAssembly functions only take and return numbers. Text goes through the
plugin's memory:

```
host (Go)                                    plugin (Rust)
──────────                                   ─────────────
ptr := alloc(len)                    ──────▶ Vec::with_capacity(len), forget it, return the pointer
memory.Write(ptr, "Anna\tgreet\t")
on_command(ptr, len)                 ──────▶ read the bytes, split at tabs, decide
                                     ◀────── mud.send(ptr, len)   (an import: the host reads the text)
                                     ◀────── returns 1 (handled) or 0 (not mine)
dealloc(ptr, len)                    ──────▶ drop the Vec
```

| | Function | Meaning |
|---|---|---|
| plugin exports | `abi_version() -> i32` | must be 1 |
| | `alloc(len) -> ptr`, `dealloc(ptr, len)` | memory for the host to write into |
| | `on_command(ptr, len) -> i32` | `"player\tverb\targs"`; 1 = handled |
| host offers (module `mud`) | `send(ptr, len)` | text to the player who typed the command |
| | `log(ptr, len)` | a line in the server's log |

The version number lets the server refuse a plugin built for another
contract instead of crashing on it.

## Go concepts

### wazero

```go
rt := wazero.NewRuntimeWithConfig(ctx, wazero.NewRuntimeConfig().
	WithMemoryLimitPages(PluginMemoryPages).   // at most 256 × 64 KiB
	WithCloseOnContextDone(true))              // a timeout stops even `loop {}`

rt.NewHostModuleBuilder("mud").
	NewFunctionBuilder().WithFunc(func(ctx context.Context, m api.Module, ptr, length uint32) {
		b, _ := m.Memory().Read(ptr, length)   // the plugin's memory
		// ...
	}).Export("send").
	Instantiate(ctx)

mod, err := rt.InstantiateWithConfig(ctx, code, wazero.NewModuleConfig().WithName(name))
res, err := mod.ExportedFunction("abi_version").Call(ctx)   // []uint64
```

`WithFunc` uses reflection to turn an ordinary Go function into a
WebAssembly import. A call takes a `context`: with a timeout, a plugin that
never returns is stopped (and its module closed – it is reloaded with the
next change to the folder).

### The order of commands

Built-in commands first, then scripts (lesson 32), then plugins, then
"unknown command". A plugin that answers 0 lets the next one try.

## The contract

- `Options.PluginDir` – a folder with `.wasm` files, loaded in name order when
  the server is created and reloaded when a file changes (every `ReloadEvery`).
- `LoadPlugins(dir) (*Plugins, error)` – a plugin with missing exports, the
  wrong ABI version or that isn't WebAssembly is skipped; the error names the
  file (and the version); the others load. `Plugins.Close()`.
- A command that no built-in command or script took is offered to the plugins
  (`"player\tverb\targs"`); the first one returning 1 has handled it.
- `mud.send` sends the text to the player who typed the command.
- A plugin call has a time limit (`PluginTimeout`, 100 ms) and at most
  `PluginMemoryPages` of memory. A failing plugin gives the player `MsgPluginError`
  (and the administrators `MsgPluginLoadFailed`), is marked broken, and the
  server keeps running.
- `main.go`: `-plugins plugins`.

## Task

1. `go get github.com/tetratelabs/wazero@v1.8.2`.
2. `plugins.go` is partly given (`LoadPlugins` with the host module,
   `loadPlugins`, `runPluginCommand`, `watchPlugins`): write `load`, `handle`
   and `call`. Then the option, the fields, `dispatch` and the flag.
3. Build the plugin (once: `rustup target add wasm32-unknown-unknown`):

   ```
   cd workspace/plugins/greeter
   cargo build --release --target wasm32-unknown-unknown
   cp target/wasm32-unknown-unknown/release/greeter.wasm ../../server/plugins/
   ```

   Start the server with `-plugins plugins`, log in and type `greet`. Change the
   greeting in `lib.rs`, build again, copy – and `greet` again, without a restart.

## Think about

- What can a plugin do to the server? What can't it do – and why?
- Why does the host call `alloc` instead of writing to some address it chooses itself?
- `on_command` gets one string with tabs. What would you do for structured data (JSON? a binary format?)?
- How would you add `on_event` for lesson 34 without breaking plugins built for version 1?

## Challenges

- A plugin in another language: TinyGo, Zig or C – same ABI.
- `mud.player(name_ptr, name_len, out_ptr) -> len`: let plugins ask about players.
- "Fuel": count instructions instead of wall-clock time (wazero has a listener API for that).
- A command `@plugins` that lists what is loaded and which plugins are broken.
