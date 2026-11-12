package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"mudserver/mud"
)

// Configuration (lesson 42). On your own computer you use flags; in a
// container it is easier to use environment variables (MUD_ADDR, MUD_WEB,
// ...). A flag wins over the environment, which wins over the default.
//
// main is hard to test – it starts servers and never returns – so the part
// that can be tested lives here, as ordinary functions that main calls.

// Config is everything main needs to know.
type Config struct {
	Addr        string        // telnet
	Web         string        // the web client; "" = none
	World       string        // the world file; "" = the built-in world
	DB          string        // the database; "" = save nothing
	Tick        time.Duration // how often the world advances
	Reload      time.Duration // how often the world file is checked
	Admins      []string      // administrators' names
	Scripts     string        // Lua scripts (lesson 32)
	Plugins     string        // WebAssembly plugins (lesson 33)
	Log         string        // "text" or "json"
	TrustProxy  bool          // believe X-Forwarded-For (behind Caddy)
	HealthCheck bool          // only ask a running server whether it is healthy
}

// loadConfig reads the flags in args (without the program name) on top of
// the environment variables that lookup finds – os.LookupEnv, or a map in
// the tests. A variable that is set but empty counts: MUD_DB= saves nothing.
func loadConfig(args []string, lookup func(string) (string, bool)) (Config, error) {
	// TODO(lesson 42): a flag.FlagSet ("mudserver", flag.ContinueOnError) with
	// one flag per field; each default comes from its MUD_* variable (lookup)
	// or the old default. A variable that is set but empty counts. Bad values
	// (MUD_TICK, MUD_RELOAD, MUD_TRUST_PROXY) and a -log other than text/json
	// are errors – all of them at once (errors.Join). Split -admin at commas.
	_, _, _ = flag.NewFlagSet, strconv.ParseBool, errors.Join
	_, _ = mud.DefaultTick, strings.Split
	return Config{}, errors.New("TODO: loadConfig")
}

// newLogger makes the server's logger: text for people, JSON for machines
// (a log service can search JSON by field).
func newLogger(format string, w io.Writer) *slog.Logger {
	// TODO(lesson 42): slog.NewJSONHandler for "json", else slog.NewTextHandler.
	return slog.New(slog.NewTextHandler(w, nil))
}

// healthCheck asks the server at addr (host:port of the web client) for
// /healthz. nil means healthy. A container image without curl can still
// check itself: `mudserver -healthcheck`.
func healthCheck(addr string) error {
	// TODO(lesson 42): GET http://addr/healthz (":8080" means 127.0.0.1:8080)
	// with a timeout; nil only for 200.
	_, _ = http.Client{}, time.Second
	return fmt.Errorf("TODO: healthCheck %s", addr)
}
