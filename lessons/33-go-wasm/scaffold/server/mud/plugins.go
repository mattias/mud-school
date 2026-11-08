package mud

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
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
	// TODO(lesson 33): read the file, pl.rt.InstantiateWithConfig with the name,
	// look up the exports abi_version, alloc, dealloc, on_command (all needed),
	// require an exported memory (len(mod.ExportedMemoryDefinitions()) > 0 –
	// mod.Memory() is a non-nil interface even without one) and the right signatures
	// (Definition().ParamTypes()/ResultTypes(): abi_version ()→i32,
	// alloc (i32)→i32, dealloc (i32,i32)→(), on_command (i32,i32)→i32), then
	// call abi_version and require ABIVersion. Close the module on every error.
	_, _, _ = os.ReadFile, strings.TrimSuffix, filepath.Base
	_, _ = slices.Equal[[]api.ValueType], api.ValueTypeI32
	return nil, errors.New("TODO: Plugins.load")
}

// Close stops every plugin.
func (pl *Plugins) Close() { pl.rt.Close(context.Background()) }

// handle offers a command to the plugins, one at a time, until one handles
// it. handled is false if none did; err is set if a plugin failed.
func (pl *Plugins) handle(p *Player, cmd Command) (handled bool, err error) {
	// TODO(lesson 33): input = "player\tverb\targs"; offer it to every plugin
	// that is not broken (pl.call) until one returns true. An error: mark the
	// plugin broken and return (true, the error with the name of the plugin).
	return false, nil
}

// call copies the input into the plugin's memory and calls on_command.
func (pl *Plugins) call(plug *Plugin, p *Player, input []byte) (bool, error) {
	// TODO(lesson 33): ctx with PluginTimeout; ptr := alloc(len); write the
	// input into the memory of the plugin; set pl.current = p; on_command(ptr, len);
	// dealloc(ptr, len) afterwards. 1 = handled.
	return false, errors.New("TODO: Plugins.call")
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
