// The operating-system half of copyover (lesson 35) – given. main.go wires it
// in: OnCopyover: startNewServer, and adoptCopyover when MUD_COPYOVER is set.
package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"

	"mudserver/mud"
)

// startNewServer starts this program again (lesson 35) – the executable on
// disk, so a freshly built binary means new code – with the listener and the
// players' connections as extra files (fd 3, 4, ...) and the state in a file.
func startNewServer(state []byte, files []*os.File) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	f, err := os.CreateTemp("", "mud-copyover-*.json")
	if err != nil {
		return err
	}
	if _, err := f.Write(state); err != nil {
		f.Close()
		return err
	}
	f.Close()
	cmd := exec.Command(exe, os.Args[1:]...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	cmd.ExtraFiles = files
	cmd.Env = append(os.Environ(), "MUD_COPYOVER="+f.Name())
	return cmd.Start()
}

// adoptCopyover reads the state and takes over the inherited files.
func adoptCopyover(srv *mud.Server, statePath string) (net.Listener, error) {
	data, err := os.ReadFile(statePath)
	os.Remove(statePath)
	if err != nil {
		return nil, err
	}
	var state mud.CopyoverState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}
	// ExtraFiles[k] becomes file descriptor 3+k in the new process.
	n := 1
	for _, p := range state.Players {
		n = max(n, p.File+1)
	}
	files := make([]*os.File, n)
	for k := range files {
		files[k] = os.NewFile(uintptr(3+k), fmt.Sprintf("copyover-%d", k))
	}
	return srv.Adopt(&state, files)
}
