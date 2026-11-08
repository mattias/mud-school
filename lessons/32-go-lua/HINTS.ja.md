# ヒント

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

ファイルを読み込むとそれが *実行* されます。`mud.command` が呼ばれるのはそのときです。
だから読み込みにも制限時間が必要です。

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

## 起動時の読み込み

`NewServerWithOptions` の中ではまだゲームループが動いていないので、そこで直接
`s.loadScripts()` を呼べます。その後の再読み込みは `s.do(s.loadScripts)` を通します。
