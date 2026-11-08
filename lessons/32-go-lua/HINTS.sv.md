# Ledtrådar

## LoadScripts

```go
sc := s.newScripts()
files, err := filepath.Glob(filepath.Join(dir, "*.lua"))
if err != nil {
	return sc, err
}
sort.Strings(files)
var errs []error
for _, f := range files {
	ctx, cancel := context.WithTimeout(context.Background(), ScriptTimeout)
	sc.L.SetContext(ctx)
	if err := sc.L.DoFile(f); err != nil {
		errs = append(errs, fmt.Errorf("%s: %w", filepath.Base(f), err))
	}
	sc.L.RemoveContext()
	cancel()
}
return sc, errors.Join(errs...)
```

Att ladda en fil *kör* den – det är då den anropar `mud.command`. Därför
behöver laddningen också en tidsgräns.

## dispatch

```go
if fn, ok := s.commands[cmd.Verb]; ok {
	fn(s, p, cmd)
	return
}
if s.runScriptCommand(p, cmd) {
	return
}
p.SendKind(KindError, fmt.Sprintf(MsgUnknownCommand, cmd.Verb))
```

## Laddning vid start

Spelloopen kör inte än i `NewServerWithOptions`, så `s.loadScripts()` kan
anropas direkt där. Senare omladdningar går via `s.do(s.loadScripts)`.
