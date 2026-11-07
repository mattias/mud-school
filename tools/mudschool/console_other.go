//go:build !windows

package main

// consoleCodePage only means something on Windows.
func consoleCodePage() (cp uint32, ok bool) { return 0, false }
