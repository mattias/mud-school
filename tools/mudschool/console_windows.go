package main

import "syscall"

// consoleCodePage returns the code page the console uses for output (65001 =
// UTF-8). ok is false if there is no console.
func consoleCodePage() (cp uint32, ok bool) {
	proc := syscall.NewLazyDLL("kernel32.dll").NewProc("GetConsoleOutputCP")
	if proc.Find() != nil {
		return 0, false
	}
	r, _, _ := proc.Call()
	return uint32(r), r != 0
}
