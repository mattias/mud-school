# ヒント

## emit

```go
for _, h := range s.hooks[e.Kind] {
	h(e)
	if e.cancelled {
		return false
	}
}
if s.scripts != nil {
	for _, fn := range s.scripts.hooks[e.Kind] {
		reason, err := s.scripts.runHook(fn, e)
		// log errors and carry on; a non-empty reason cancels
	}
}
return true
```

## 戻り値のある call

```go
if err := sc.L.CallByParam(lua.P{Fn: fn, NRet: nret, Protect: true}, args...); err != nil {
	return lua.LNil, err
}
if nret == 0 {
	return lua.LNil, nil
}
ret := sc.L.Get(-1) // the top of the stack
sc.L.Pop(1)
return ret, nil
```

## cmdSay の中で

`emit` の後のメッセージには `cmd.Rest` ではなく `e.Text` を使います。そうすればフックが
言う内容を変えられます。
