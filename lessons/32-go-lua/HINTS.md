# Hints

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

Loading a file *runs* it – that is when it calls `mud.command`. So loading
needs a time limit too.

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

## Loading at start

The game loop isn't running yet in `NewServerWithOptions`, so
`s.loadScripts()` can be called directly there. Later reloads go through
`s.do(s.loadScripts)`.
