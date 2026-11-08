# Hints

## load

```go
code, err := os.ReadFile(path)
// ...
name := strings.TrimSuffix(filepath.Base(path), ".wasm")
mod, err := pl.rt.InstantiateWithConfig(ctx, code, wazero.NewModuleConfig().WithName(name))
// ...
p := &Plugin{Name: name, mod: mod,
	alloc: mod.ExportedFunction("alloc"), dealloc: mod.ExportedFunction("dealloc"),
	onCommand: mod.ExportedFunction("on_command")}
```

`ExportedFunction` returns `nil` for a missing export – check all four
(including `abi_version`) before you call any of them.

## call

```go
ctx, cancel := context.WithTimeout(context.Background(), PluginTimeout)
defer cancel()
res, err := plug.alloc.Call(ctx, uint64(len(input)))
if err != nil {
	return false, err
}
ptr := uint32(res[0])
defer plug.dealloc.Call(ctx, uint64(ptr), uint64(len(input)))
if !plug.mod.Memory().Write(ptr, input) {
	return false, errors.New("alloc returned memory outside the plugin")
}
pl.current = p
defer func() { pl.current = nil }()
res, err = plug.onCommand.Call(ctx, uint64(ptr), uint64(len(input)))
```

Results come back as `[]uint64` – an `i32` is `int32(res[0])`.

## "unknown import: mud.send"

The host module must be instantiated **before** the plugins: that is what
`LoadPlugins` does first.
