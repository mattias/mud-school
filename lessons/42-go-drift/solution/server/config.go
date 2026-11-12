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
	var errs []error
	str := func(env, def string) string {
		if v, ok := lookup(env); ok {
			return v
		}
		return def
	}
	dur := func(env string, def time.Duration) time.Duration {
		v, ok := lookup(env)
		if !ok {
			return def
		}
		d, err := time.ParseDuration(v)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s=%q: %w", env, v, err))
			return def
		}
		return d
	}
	boolean := func(env string) bool {
		v, ok := lookup(env)
		if !ok || v == "" {
			return false
		}
		b, err := strconv.ParseBool(v)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s=%q: %w", env, v, err))
		}
		return b
	}

	var c Config
	var admins string
	fs := flag.NewFlagSet("mudserver", flag.ContinueOnError)
	fs.StringVar(&c.Addr, "addr", str("MUD_ADDR", "127.0.0.1:4000"), "address to listen on for telnet (use :4000 to let other computers connect) [MUD_ADDR]")
	fs.StringVar(&c.Web, "web", str("MUD_WEB", "127.0.0.1:8080"), "address of the web client (lesson 39); empty = no web client [MUD_WEB]")
	fs.StringVar(&c.World, "world", str("MUD_WORLD", ""), "world file (JSON); empty = the built-in world [MUD_WORLD]")
	fs.StringVar(&c.DB, "db", str("MUD_DB", "mud.db"), "the SQLite database where players are saved; empty = save nothing [MUD_DB]")
	fs.DurationVar(&c.Tick, "tick", dur("MUD_TICK", mud.DefaultTick), "how often the world advances [MUD_TICK]")
	fs.DurationVar(&c.Reload, "reload", dur("MUD_RELOAD", 2*time.Second), "how often the -world file is checked for changes (0 = never) [MUD_RELOAD]")
	fs.StringVar(&admins, "admin", str("MUD_ADMINS", ""), "comma-separated names of administrators, e.g. -admin Anna,Bertil [MUD_ADMINS]")
	fs.StringVar(&c.Scripts, "scripts", str("MUD_SCRIPTS", ""), "a folder with .lua scripts (lesson 32) [MUD_SCRIPTS]")
	fs.StringVar(&c.Plugins, "plugins", str("MUD_PLUGINS", ""), "a folder with .wasm plugins (lesson 33) [MUD_PLUGINS]")
	fs.StringVar(&c.Log, "log", str("MUD_LOG", "text"), "log format: text or json [MUD_LOG]")
	fs.BoolVar(&c.TrustProxy, "trust-proxy", boolean("MUD_TRUST_PROXY"), "believe X-Forwarded-For – only behind a proxy like Caddy [MUD_TRUST_PROXY]")
	fs.BoolVar(&c.HealthCheck, "healthcheck", false, "ask the server running on -web whether it is healthy, and exit (for Docker)")
	if err := fs.Parse(args); err != nil {
		return c, err
	}
	if c.Log != "text" && c.Log != "json" {
		errs = append(errs, fmt.Errorf("-log %q: text or json", c.Log))
	}
	for _, a := range strings.Split(admins, ",") {
		if a = strings.TrimSpace(a); a != "" {
			c.Admins = append(c.Admins, a)
		}
	}
	return c, errors.Join(errs...)
}

// newLogger makes the server's logger: text for people, JSON for machines
// (a log service can search JSON by field).
func newLogger(format string, w io.Writer) *slog.Logger {
	if format == "json" {
		return slog.New(slog.NewJSONHandler(w, nil))
	}
	return slog.New(slog.NewTextHandler(w, nil))
}

// healthCheck asks the server at addr (host:port of the web client) for
// /healthz. nil means healthy. A container image without curl can still
// check itself: `mudserver -healthcheck`.
func healthCheck(addr string) error {
	if strings.HasPrefix(addr, ":") {
		addr = "127.0.0.1" + addr
	}
	client := http.Client{Timeout: 3 * time.Second}
	res, err := client.Get("http://" + addr + "/healthz")
	if err != nil {
		return err
	}
	res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("/healthz: %s", res.Status)
	}
	return nil
}
