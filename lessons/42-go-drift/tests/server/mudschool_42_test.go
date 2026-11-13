// This file belongs to the course tool and is overwritten by `mudschool check`.
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"
)

// L42 picks the message in the course language ($MUDSCHOOL_LANG).
func L42(sv, en, ja string) string {
	switch os.Getenv("MUDSCHOOL_LANG") {
	case "sv":
		return sv
	case "ja":
		return ja
	}
	return en
}

func env(m map[string]string) func(string) (string, bool) {
	return func(k string) (string, bool) {
		v, ok := m[k]
		return v, ok
	}
}

func TestConfigDefaults(t *testing.T) {
	c, err := loadConfig(nil, env(nil))
	if err != nil {
		t.Fatal(err)
	}
	if c.Addr != "127.0.0.1:4000" || c.Web != "127.0.0.1:8080" || c.DB != "mud.db" || c.Log != "text" || c.TrustProxy {
		t.Errorf(L42("standardvärdena ska vara som förut: %+v", "the defaults should be as before: %+v", "デフォルトは以前と同じであるべきです: %+v"), c)
	}
}

func TestConfigFromTheEnvironment(t *testing.T) {
	c, err := loadConfig(nil, env(map[string]string{
		"MUD_ADDR": ":4000", "MUD_WEB": ":8080", "MUD_DB": "", "MUD_TICK": "250ms",
		"MUD_ADMINS": "Anna, Bertil", "MUD_LOG": "json", "MUD_TRUST_PROXY": "true",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if c.Addr != ":4000" || c.Web != ":8080" || c.Tick != 250*time.Millisecond || c.Log != "json" || !c.TrustProxy {
		t.Errorf(L42("miljövariablerna MUD_* ska användas: %+v", "the MUD_* environment variables should be used: %+v", "環境変数 MUD_* を使うべきです: %+v"), c)
	}
	if c.DB != "" {
		t.Errorf(L42("MUD_DB= (satt men tom) betyder att inget sparas, fick %q", "MUD_DB= (set but empty) means nothing is saved, got %q", "MUD_DB=（空で設定）は何も保存しないという意味です。%q になりました"), c.DB)
	}
	if strings.Join(c.Admins, "|") != "Anna|Bertil" {
		t.Errorf(L42("MUD_ADMINS delas vid komman och trimmas: %q", "MUD_ADMINS is split at commas and trimmed: %q", "MUD_ADMINS はカンマで分けて空白を除きます: %q"), c.Admins)
	}
}

func TestFlagsWin(t *testing.T) {
	c, err := loadConfig([]string{"-addr", "127.0.0.1:5000", "-trust-proxy=false"}, env(map[string]string{"MUD_ADDR": ":4000", "MUD_TRUST_PROXY": "1"}))
	if err != nil {
		t.Fatal(err)
	}
	if c.Addr != "127.0.0.1:5000" || c.TrustProxy {
		t.Errorf(L42("en flagga vinner över miljön: %+v", "a flag wins over the environment: %+v", "フラグは環境変数より優先です: %+v"), c)
	}
}

func TestBadConfig(t *testing.T) {
	_, err := loadConfig(nil, env(map[string]string{"MUD_TICK": "fast", "MUD_TRUST_PROXY": "maybe"}))
	if err == nil || !strings.Contains(err.Error(), "MUD_TICK") || !strings.Contains(err.Error(), "MUD_TRUST_PROXY") {
		t.Errorf(L42("felaktiga värden ska ge ett fel som nämner variablerna (alla på en gång): %v", "bad values should give an error naming the variables (all at once): %v", "誤った値は、変数名を挙げたエラーになるべきです（すべて一度に）: %v"), err)
	}
	if _, err := loadConfig([]string{"-log", "xml"}, env(nil)); err == nil {
		t.Errorf(L42("-log xml ska ge ett fel (text eller json)", "-log xml should give an error (text or json)", "-log xml はエラーになるべきです（text か json）"))
	}
}

func TestJSONLogs(t *testing.T) {
	var buf bytes.Buffer
	newLogger("json", &buf).Info("hello", "players", 3)
	var line map[string]any
	if err := json.Unmarshal(buf.Bytes(), &line); err != nil || line["msg"] != "hello" || line["players"] != float64(3) {
		t.Errorf(L42("newLogger(\"json\") ska skriva en JSON-rad med msg och fälten: %q", "newLogger(\"json\") should write a JSON line with msg and the fields: %q", "newLogger(\"json\") は msg とフィールドを含む JSON の行を書くべきです: %q"), buf.String())
	}
}

func TestHealthCheck(t *testing.T) {
	status := http.StatusOK
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/healthz" {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(status)
	}))
	defer ts.Close()
	addr := strings.TrimPrefix(ts.URL, "http://")
	if err := healthCheck(addr); err != nil {
		t.Errorf(L42("healthCheck mot en frisk server ska ge nil: %v", "healthCheck against a healthy server should give nil: %v", "健康なサーバーへの healthCheck は nil を返すべきです: %v"), err)
	}
	status = http.StatusServiceUnavailable
	if err := healthCheck(addr); err == nil {
		t.Errorf(L42("healthCheck ska ge ett fel när /healthz svarar 503", "healthCheck should give an error when /healthz answers 503", "/healthz が 503 なら healthCheck はエラーを返すべきです"))
	}
}

// TestSIGTERM builds the real server and stops it the way `docker stop` does.
func TestSIGTERM(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("no SIGTERM on Windows")
	}
	bin := filepath.Join(t.TempDir(), "mudserver")
	if out, err := exec.Command("go", "build", "-o", bin, ".").CombinedOutput(); err != nil {
		t.Fatalf("go build: %v\n%s", err, out)
	}
	cmd := exec.Command(bin)
	cmd.Env = append(os.Environ(), "MUD_ADDR=127.0.0.1:0", "MUD_WEB=", "MUD_DB=", "MUD_LOG=text")
	stderr, _ := cmd.StderrPipe()
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	lines := make(chan string, 100)
	go func() {
		sc := bufio.NewScanner(stderr)
		for sc.Scan() {
			lines <- sc.Text()
		}
		close(lines)
	}()
	var seen []string
	wait := func(want string) bool {
		timeout := time.After(10 * time.Second)
		for {
			select {
			case line, ok := <-lines:
				if !ok {
					return false
				}
				seen = append(seen, line)
				if strings.Contains(line, want) {
					return true
				}
			case <-timeout:
				return false
			}
		}
	}
	if !wait("listening") {
		cmd.Process.Kill()
		t.Fatalf(L42("servern startade inte: %q", "the server didn't start: %q", "サーバーが起動しませんでした: %q"), seen)
	}
	cmd.Process.Signal(syscall.SIGTERM)
	if !wait("bye") {
		cmd.Process.Kill()
		t.Fatalf(L42("efter SIGTERM (det docker stop skickar) ska servern stänga snyggt och skriva bye: %q",
			"after SIGTERM (what docker stop sends) the server should shut down cleanly and log bye: %q",
			"SIGTERM（docker stop が送るもの）の後、サーバーはきれいに止まって bye を出すべきです: %q"), seen)
	}
	if err := cmd.Wait(); err != nil {
		t.Errorf(L42("servern ska avsluta med 0 efter SIGTERM: %v", "the server should exit with 0 after SIGTERM: %v", "SIGTERM の後、サーバーは 0 で終了するべきです: %v"), err)
	}
}

func TestTheDeploymentFiles(t *testing.T) {
	for _, f := range []string{"../Dockerfile", "../compose.yaml", "../Caddyfile", "../.dockerignore"} {
		if _, err := os.Stat(f); err != nil {
			t.Errorf(L42("%s saknas (lektion 42 ger dig den)", "%s is missing (lesson 42 gives it to you)", "%s がありません（レッスン 42 で用意されます）"), f)
		}
	}
}
