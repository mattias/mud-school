# Ledtrådar

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

`ExportedFunction` returnerar `nil` för en export som saknas – kontrollera
alla fyra (också `abi_version`) innan du anropar någon av dem.

Ett plugin kan också exportera rätt namn med fel typer. Jämför
`mod.ExportedFunction(name).Definition().ParamTypes()` och `.ResultTypes()`
med `[]api.ValueType{api.ValueTypeI32}` och liknande (`slices.Equal`), och
kontrollera `len(mod.ExportedMemoryDefinitions()) > 0` (`mod.Memory()` är
aldrig ett vanligt `nil`, inte ens utan minne) – annars får `res[0]` eller `Memory().Write`
panik inne i spelloopen.

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

Resultaten kommer som `[]uint64` – en `i32` är `int32(res[0])`.

## "unknown import: mud.send"

Värdmodulen måste instansieras **före** pluginen: det är det `LoadPlugins`
gör först.
