# Ledtrådar

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

## call med returvärde

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

## I cmdSay

Använd `e.Text` – inte `cmd.Rest` – i meddelandena efter `emit`, så att
hooks kan ändra det som sägs.
