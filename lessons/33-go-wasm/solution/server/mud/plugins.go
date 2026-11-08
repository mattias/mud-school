package mud

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

// Plugins in WebAssembly (lesson 33): commands written in any language that
// compiles to WebAssembly – the course uses Rust – run inside the Go server
// by wazero, a WebAssembly runtime in pure Go. A plugin can only touch its
// own memory and the two functions the server offers: it is isolated.

// ABIVersion is the version of the contract between server and plugins.
const ABIVersion = 1

// PluginTimeout is how long a plugin may run per call.
const PluginTimeout = 100 * time.Millisecond

// PluginMemoryPages limits a plugin's memory (a page is 64 KiB).
const PluginMemoryPages = 256

// Plugin is one loaded .wasm file.
type Plugin struct {
	Name      string
	mod       api.Module
	alloc     api.Function
	dealloc   api.Function
	onCommand api.Function
	broken    bool // stopped after a crash or a timeout
}

// Plugins is a loaded set of plugins sharing one runtime.
type Plugins struct {
	rt      wazero.Runtime
	list    []*Plugin
	current *Player // who typed the command being handled (for mud.send)
}

// LoadPlugins loads every .wasm file in dir, in name order. A plugin that
// can't be loaded is skipped and its error returned with the others.
func (s *Server) LoadPlugins(dir string) (*Plugins, error) {
	ctx := context.Background()
	rt := wazero.NewRuntimeWithConfig(ctx, wazero.NewRuntimeConfig().
		WithMemoryLimitPages(PluginMemoryPages).
		WithCloseOnContextDone(true)) // a timeout stops even an endless loop
	pl := &Plugins{rt: rt}

	// What the server offers plugins: the module "mud".
	_, err := rt.NewHostModuleBuilder("mud").
		NewFunctionBuilder().WithFunc(func(ctx context.Context, m api.Module, ptr, length uint32) {
		if b, ok := m.Memory().Read(ptr, length); ok && pl.current != nil {
			pl.current.SendKind(KindInfo, string(b))
		}
	}).Export("send").
		NewFunctionBuilder().WithFunc(func(ctx context.Context, m api.Module, ptr, length uint32) {
		if b, ok := m.Memory().Read(ptr, length); ok {
			log.Printf("plugin %s: %s", m.Name(), b)
		}
	}).Export("log").
		Instantiate(ctx)
	if err != nil {
		return pl, err
	}

	files, _ := filepath.Glob(filepath.Join(dir, "*.wasm"))
	sort.Strings(files)
	var errs []error
	for _, f := range files {
		p, err := pl.load(ctx, f)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", filepath.Base(f), err))
			continue
		}
		pl.list = append(pl.list, p)
	}
	return pl, errors.Join(errs...)
}

// load compiles, instantiates and checks one plugin.
func (pl *Plugins) load(ctx context.Context, path string) (*Plugin, error) {
	code, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	name := strings.TrimSuffix(filepath.Base(path), ".wasm")
	mod, err := pl.rt.InstantiateWithConfig(ctx, code, wazero.NewModuleConfig().WithName(name))
	if err != nil {
		return nil, err
	}
	p := &Plugin{Name: name, mod: mod,
		alloc: mod.ExportedFunction("alloc"), dealloc: mod.ExportedFunction("dealloc"),
		onCommand: mod.ExportedFunction("on_command")}
	version := mod.ExportedFunction("abi_version")
	if version == nil || p.alloc == nil || p.dealloc == nil || p.onCommand == nil {
		mod.Close(ctx)
		return nil, errors.New("missing exports (abi_version, alloc, dealloc, on_command)")
	}
	res, err := version.Call(ctx)
	if err != nil {
		mod.Close(ctx)
		return nil, err
	}
	if v := int32(res[0]); v != ABIVersion {
		mod.Close(ctx)
		return nil, fmt.Errorf("ABI version %d, the server speaks %d", v, ABIVersion)
	}
	return p, nil
}

// Close stops every plugin.
func (pl *Plugins) Close() { pl.rt.Close(context.Background()) }

// handle offers a command to the plugins, one at a time, until one handles
// it. handled is false if none did; err is set if a plugin failed.
func (pl *Plugins) handle(p *Player, cmd Command) (handled bool, err error) {
	input := []byte(p.Name + "\t" + cmd.Verb + "\t" + cmd.Rest)
	for _, plug := range pl.list {
		if plug.broken {
			continue
		}
		ok, err := pl.call(plug, p, input)
		if err != nil {
			plug.broken = true // a timeout closes the module; it is reloaded with the next change
			return true, fmt.Errorf("%s: %w", plug.Name, err)
		}
		if ok {
			return true, nil
		}
	}
	return false, nil
}

// call copies the input into the plugin's memory and calls on_command.
func (pl *Plugins) call(plug *Plugin, p *Player, input []byte) (bool, error) {
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
	if err != nil {
		return false, err
	}
	return int32(res[0]) == 1, nil
}

// loadPlugins (re)loads the server's plugins. Runs in the game loop (or
// before it has started).
func (s *Server) loadPlugins() {
	pl, err := s.LoadPlugins(s.pluginDir)
	if err != nil {
		log.Printf("plugins: %v", err)
		s.notifyAdmins(KindError, fmt.Sprintf(MsgPluginLoadFailed, err))
	}
	if s.plugins != nil {
		s.plugins.Close()
	}
	s.plugins = pl
	log.Printf("plugins: %d loaded", len(pl.list))
}

// runPluginCommand lets the plugins try a command. ok is false if none took it.
func (s *Server) runPluginCommand(p *Player, cmd Command) (ok bool) {
	if s.plugins == nil {
		return false
	}
	handled, err := s.plugins.handle(p, cmd)
	if err != nil {
		log.Printf("plugin command %s: %v", cmd.Verb, err)
		p.SendKind(KindError, MsgPluginError)
		s.notifyAdmins(KindError, fmt.Sprintf(MsgPluginLoadFailed, err))
	}
	return handled
}

// watchPlugins reloads the plugins when a .wasm file is added, removed or changed.
func (s *Server) watchPlugins() {
	signature := func() string {
		files, _ := filepath.Glob(filepath.Join(s.pluginDir, "*.wasm"))
		sort.Strings(files)
		var b strings.Builder
		for _, f := range files {
			if fi, err := os.Stat(f); err == nil {
				fmt.Fprintf(&b, "%s %d %d\n", f, fi.ModTime().UnixNano(), fi.Size())
			}
		}
		return b.String()
	}
	last := signature()
	t := time.NewTicker(s.reloadEvery)
	defer t.Stop()
	for {
		select {
		case <-s.stop:
			return
		case <-t.C:
			if now := signature(); now != last {
				last = now
				s.do(s.loadPlugins)
			}
		}
	}
}
