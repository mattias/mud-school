# ヒント

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

`ExportedFunction` は、ないエクスポートには `nil` を返します。どれかを呼ぶ前に、4 つすべて
（`abi_version` も）を確かめましょう。

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

結果は `[]uint64` で返ります。`i32` なら `int32(res[0])` です。

## "unknown import: mud.send"

ホストモジュールはプラグインより **前に** インスタンス化しなければなりません。
`LoadPlugins` が最初にしているのはそれです。
