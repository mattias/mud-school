# Lektion 33 – Plugins i WebAssembly

Här möts kursens två språk på riktigt: ett plugin skrivet i **Rust**,
kompilerat till **WebAssembly**, som körs **inne i Go-servern**.

```
workspace/plugins/greeter/src/lib.rs   (Rust)
        │  cargo build --release --target wasm32-unknown-unknown
        ▼
server/plugins/greeter.wasm            (WebAssembly: portabel, inkapslad bytekod)
        │  laddas av wazero – en WebAssembly-körmiljö skriven i ren Go
        ▼
"greet"  ->  Hello, Anna! Greetings from a plugin written in Rust.
```

Varför WebAssembly i stället för Lua (lektion 32)? Det är snabbt (kompilerat),
det kan skrivas i vilket språk som helst som kompilerar till det (Rust, C,
Zig, Go, …), och det är **isolerat**: ett plugin kommer bara åt sitt eget
minne och de funktioner servern väljer att ge det. Ett plugin som kraschar
eller aldrig blir klart kan inte ta ner servern.

## Mål

- Designa ett **ABI** – kontraktet mellan värd och plugin – av ingenting annat
  än tal och ett delat minnesblock.
- Skicka strängar över gränsen: allokera i pluginet, skriv från värden, läs med pekare + längd.
- Begränsa kod man inte litar på: minnessidor, en tidsgräns, en versionskontroll.
- Bygga ett Rust-bibliotek för ett annat mål (`wasm32-unknown-unknown`) med `crate-type = ["cdylib"]`.

## ABI:t (version 1)

WebAssembly-funktioner tar och returnerar bara tal. Text går via pluginets minne:

```
värden (Go)                                  pluginet (Rust)
───────────                                  ───────────────
ptr := alloc(len)                    ──────▶ Vec::with_capacity(len), glöm den, returnera pekaren
memory.Write(ptr, "Anna\tgreet\t")
on_command(ptr, len)                 ──────▶ läs byten, dela vid tabbar, bestäm
                                     ◀────── mud.send(ptr, len)   (en import: värden läser texten)
                                     ◀────── returnerar 1 (hanterat) eller 0 (inte mitt)
dealloc(ptr, len)                    ──────▶ släpp Vec:en
```

| | Funktion | Betydelse |
|---|---|---|
| pluginet exporterar | `abi_version() -> i32` | måste vara 1 |
| | `alloc(len) -> ptr`, `dealloc(ptr, len)` | minne som värden får skriva i |
| | `on_command(ptr, len) -> i32` | `"player\tverb\targs"`; 1 = hanterat |
| värden erbjuder (modulen `mud`) | `send(ptr, len)` | text till spelaren som skrev kommandot |
| | `log(ptr, len)` | en rad i serverns logg |

Versionsnumret gör att servern kan vägra ett plugin byggt för ett annat
kontrakt i stället för att krascha på det.

## Go-koncept

### wazero

```go
rt := wazero.NewRuntimeWithConfig(ctx, wazero.NewRuntimeConfig().
	WithMemoryLimitPages(PluginMemoryPages).   // högst 256 × 64 KiB
	WithCloseOnContextDone(true))              // en timeout stoppar även `loop {}`

rt.NewHostModuleBuilder("mud").
	NewFunctionBuilder().WithFunc(func(ctx context.Context, m api.Module, ptr, length uint32) {
		b, _ := m.Memory().Read(ptr, length)   // pluginets minne
		// ...
	}).Export("send").
	Instantiate(ctx)

mod, err := rt.InstantiateWithConfig(ctx, code, wazero.NewModuleConfig().WithName(name))
res, err := mod.ExportedFunction("abi_version").Call(ctx)   // []uint64
```

`WithFunc` använder reflektion för att göra en vanlig Go-funktion till en
WebAssembly-import. Ett anrop tar en `context`: med en timeout stoppas ett
plugin som aldrig returnerar (och dess modul stängs – den laddas om vid nästa
ändring i mappen).

### Kommandonas ordning

Inbyggda kommandon först, sedan skript (lektion 32), sedan plugins, sedan
"okänt kommando". Ett plugin som svarar 0 låter nästa försöka.

## Kontraktet

- `Options.PluginDir` – en mapp med `.wasm`-filer, laddade i namnordning när
  servern skapas och omladdade när en fil ändras (var `ReloadEvery`).
- `LoadPlugins(dir) (*Plugins, error)` – ett plugin som saknar exporter, har
  fel ABI-version eller inte är WebAssembly hoppas över; felet nämner filen
  (och versionen); de andra laddas. `Plugins.Close()`.
- Ett kommando som inget inbyggt kommando eller skript tog erbjuds pluginen
  (`"player\tverb\targs"`); det första som returnerar 1 har hanterat det.
- `mud.send` skickar texten till spelaren som skrev kommandot.
- Ett pluginanrop har en tidsgräns (`PluginTimeout`, 100 ms) och högst
  `PluginMemoryPages` minne. Ett plugin som får fel ger spelaren
  `MsgPluginError` (och administratörerna `MsgPluginLoadFailed`), markeras
  som trasigt, och servern fortsätter.
- `main.go`: `-plugins plugins`.

## Uppgift

1. `go get github.com/tetratelabs/wazero@v1.8.2`.
2. `plugins.go` är delvis given (`LoadPlugins` med värdmodulen,
   `loadPlugins`, `runPluginCommand`, `watchPlugins`): skriv `load`, `handle`
   och `call`. Sedan alternativet, fälten, `dispatch` och flaggan.
3. Bygg pluginet (en gång: `rustup target add wasm32-unknown-unknown`):

   ```
   cd workspace/plugins/greeter
   cargo build --release --target wasm32-unknown-unknown
   cp target/wasm32-unknown-unknown/release/greeter.wasm ../../server/plugins/
   ```

   Starta servern med `-plugins plugins`, logga in och skriv `greet`. Ändra
   hälsningen i `lib.rs`, bygg igen, kopiera – och `greet` igen, utan omstart.

## Fundera på

- Vad kan ett plugin göra med servern? Vad kan det inte – och varför?
- Varför anropar värden `alloc` i stället för att skriva på en adress den väljer själv?
- `on_command` får en sträng med tabbar. Vad skulle du göra för strukturerad data (JSON? ett binärformat?)?
- Hur skulle du lägga till `on_event` för lektion 34 utan att förstöra plugins byggda för version 1?

## Utmaningar

- Ett plugin i ett annat språk: TinyGo, Zig eller C – samma ABI.
- `mud.player(name_ptr, name_len, out_ptr) -> len`: låt plugins fråga om spelare.
- "Bränsle": räkna instruktioner i stället för väggklockans tid (wazero har ett lyssnar-API för det).
- Ett kommando `@plugins` som listar vad som är laddat och vilka plugins som är trasiga.
