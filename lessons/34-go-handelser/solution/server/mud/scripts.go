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

	lua "github.com/yuin/gopher-lua"
)

// Scripts in Lua (lesson 32): commands and creature behaviours in .lua
// files that are loaded – and reloaded – while the server runs.
//
// A Lua state is NOT safe for several goroutines, so it belongs to the game
// loop, like the world: scripts are only loaded and run inside it.

// ScriptTimeout is how long a script may run before it is stopped.
const ScriptTimeout = 50 * time.Millisecond

// Scripts is a loaded set of scripts.
type Scripts struct {
	L         *lua.LState
	commands  map[string]*lua.LFunction   // registered with mud.command
	behaviors map[string]*lua.LFunction   // registered with mud.behavior
	hooks     map[string][]*lua.LFunction // registered with mud.on (lesson 34)
}

// newScripts creates a sandboxed Lua state with the mud API.
func (s *Server) newScripts() *Scripts {
	L := lua.NewState(lua.Options{SkipOpenLibs: true})
	// Only the safe libraries: no os, no io – a script must not delete files
	// or start programs.
	for _, lib := range []struct {
		name string
		open lua.LGFunction
	}{
		{lua.BaseLibName, lua.OpenBase},
		{lua.TabLibName, lua.OpenTable},
		{lua.StringLibName, lua.OpenString},
		{lua.MathLibName, lua.OpenMath},
	} {
		L.Push(L.NewFunction(lib.open))
		L.Push(lua.LString(lib.name))
		L.Call(1, 0)
	}
	// The base library can still read files and load code: take that away.
	for _, name := range []string{"dofile", "loadfile", "load", "loadstring", "require", "module"} {
		L.SetGlobal(name, lua.LNil)
	}

	sc := &Scripts{L: L, commands: map[string]*lua.LFunction{}, behaviors: map[string]*lua.LFunction{},
		hooks: map[string][]*lua.LFunction{}}
	api := map[string]lua.LGFunction{
		// mud.command(name, function(player, args) ... end)
		"command": func(L *lua.LState) int {
			sc.commands[L.CheckString(1)] = L.CheckFunction(2)
			return 0
		},
		// mud.behavior(name, function(npc, room) ... end)
		"behavior": func(L *lua.LState) int {
			sc.behaviors[L.CheckString(1)] = L.CheckFunction(2)
			return 0
		},
		// mud.on(kind, function(event) ... end) – return a string to cancel (lesson 34)
		"on": func(L *lua.LState) int {
			kind := L.CheckString(1)
			sc.hooks[kind] = append(sc.hooks[kind], L.CheckFunction(2))
			return 0
		},
		// mud.send(player, text)
		"send": func(L *lua.LState) int {
			if p := s.players[NormalizeName(L.CheckString(1))]; p != nil {
				p.SendKind(KindInfo, L.CheckString(2))
			}
			return 0
		},
		// mud.broadcast(room, text) – to everyone in a room
		"broadcast": func(L *lua.LState) int {
			if r := s.world.Room(L.CheckString(1)); r != nil {
				s.broadcastRoom(r, nil, KindInfo, L.CheckString(2))
			}
			return 0
		},
		// mud.player(name) -> {name=, room=, hp=, max_hp=} or nil
		"player": func(L *lua.LState) int {
			p := s.players[NormalizeName(L.CheckString(1))]
			if p == nil {
				L.Push(lua.LNil)
				return 1
			}
			t := L.NewTable()
			t.RawSetString("name", lua.LString(p.Name))
			t.RawSetString("room", lua.LString(p.Room.ID))
			t.RawSetString("hp", lua.LNumber(p.HP))
			t.RawSetString("max_hp", lua.LNumber(p.MaxHP))
			L.Push(t)
			return 1
		},
		// mud.players(room) -> a list of the names of the players there
		"players": func(L *lua.LState) int {
			t := L.NewTable()
			if r := s.world.Room(L.CheckString(1)); r != nil {
				for _, p := range s.playersInRoom(r, nil) {
					t.Append(lua.LString(p.Name))
				}
			}
			L.Push(t)
			return 1
		},
		// mud.roll(a, b) -> a random whole number from a to b
		"roll": func(L *lua.LState) int {
			a, b := L.CheckInt(1), L.CheckInt(2)
			if b < a {
				a, b = b, a
			}
			L.Push(lua.LNumber(s.roll([2]int{a, b})))
			return 1
		},
	}
	L.SetGlobal("mud", L.SetFuncs(L.NewTable(), api))
	return sc
}

// run calls a script function with a time limit.
func (sc *Scripts) run(fn *lua.LFunction, args ...lua.LValue) error {
	_, err := sc.call(fn, 0, args...)
	return err
}

// call calls a script function with a time limit and returns its first
// result if nret is 1 (lesson 34).
func (sc *Scripts) call(fn *lua.LFunction, nret int, args ...lua.LValue) (lua.LValue, error) {
	ctx, cancel := context.WithTimeout(context.Background(), ScriptTimeout)
	defer cancel()
	sc.L.SetContext(ctx)
	defer sc.L.RemoveContext()
	if err := sc.L.CallByParam(lua.P{Fn: fn, NRet: nret, Protect: true}, args...); err != nil {
		return lua.LNil, err
	}
	if nret == 0 {
		return lua.LNil, nil
	}
	ret := sc.L.Get(-1)
	sc.L.Pop(1)
	return ret, nil
}

// LoadScripts loads every .lua file in dir, in name order. A file with an
// error is skipped – the others still load – and the errors are returned
// together.
func (s *Server) LoadScripts(dir string) (*Scripts, error) {
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
}

// loadScripts (re)loads the server's scripts. Runs in the game loop (or
// before it has started).
func (s *Server) loadScripts() {
	sc, err := s.LoadScripts(s.scriptDir)
	if err != nil {
		log.Printf("scripts: %v", err)
		s.notifyAdmins(KindError, fmt.Sprintf(MsgScriptLoadFailed, err))
	}
	if s.scripts != nil {
		s.scripts.L.Close()
	}
	s.scripts = sc
	log.Printf("scripts: %d commands, %d behaviours", len(sc.commands), len(sc.behaviors))
}

// runScriptCommand runs a command a script has registered. ok is false if
// there is no such command.
func (s *Server) runScriptCommand(p *Player, cmd Command) (ok bool) {
	if s.scripts == nil {
		return false
	}
	fn := s.scripts.commands[cmd.Verb]
	if fn == nil {
		return false
	}
	if err := s.scripts.run(fn, lua.LString(p.Name), lua.LString(cmd.Rest)); err != nil {
		log.Printf("script command %s: %v", cmd.Verb, err)
		p.SendKind(KindError, MsgScriptError)
	}
	return true
}

// ScriptBehavior runs a behaviour a script has registered (lesson 32):
// {"type": "script", "script": "guard"}.
type ScriptBehavior struct{ Name string }

func (b ScriptBehavior) Act(s *Server, n *NPC) {
	if s.scripts == nil || n.dead {
		return
	}
	fn := s.scripts.behaviors[b.Name]
	if fn == nil {
		return
	}
	if err := s.scripts.run(fn, lua.LString(n.ID), lua.LString(n.Location.ID)); err != nil {
		log.Printf("script behaviour %s (%s): %v", b.Name, n.ID, err)
	}
}

// watchScripts reloads the scripts when a .lua file is added, removed or changed.
func (s *Server) watchScripts() {
	signature := func() string {
		files, _ := filepath.Glob(filepath.Join(s.scriptDir, "*.lua"))
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
				s.do(s.loadScripts)
			}
		}
	}
}
