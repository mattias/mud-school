#!/bin/sh
# Regenerates docs/images/client-<lang>.png from the reference solution.
#
# Needs: go, cargo, tmux, nc, node with the "playwright" package (and its
# Chromium), and a Japanese font such as Noto CJK (fonts-noto-cjk).
#
# Usage: tools/screenshots/make.sh            (from the repository root)
set -e
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
HERE="$ROOT/tools/screenshots"
WORK="$(mktemp -d)"
cleanup() {
  set +e
  for s in en sv ja; do tmux kill-session -t "shot$s" 2>/dev/null; done
  rm -rf "$WORK"
}
trap cleanup EXIT

port=4500
for L in en sv ja; do
  case $L in
    en) SAY='Hi Anna!' ;;
    sv) SAY='Hej Anna!' ;;
    ja) SAY='やあ、Anna！' ;;
  esac
  port=$((port + 1))
  (cd "$ROOT" && ./mudschool reference "$WORK/ref-$L" --lang "$L" >/dev/null)
  (cd "$WORK/ref-$L/server" && go build -o mudserver .)
  (cd "$WORK/ref-$L/client" && CARGO_TARGET_DIR="$WORK/target" cargo build -q)
  cp "$WORK/target/debug/mud-client" "$WORK/client-$L"

  (cd "$WORK/ref-$L/server" && timeout 25 ./mudserver -addr 127.0.0.1:$port >/dev/null 2>&1 &)
  sleep 0.5
  tmux new-session -d -s "shot$L" -x 80 -y 24 \
    "LANG=C.UTF-8 TERM=xterm-256color COLORTERM=truecolor $WORK/client-$L 127.0.0.1:$port"
  sleep 1
  tmux send-keys -t "shot$L" 'Anna' Enter
  sleep 0.5
  ( printf 'Bertil\r\n'; sleep 0.4; printf 'say %s\r\n' "$SAY"; sleep 4 ) | timeout 6 nc 127.0.0.1 $port >/dev/null &
  sleep 1.5
  for k in s s s s; do tmux send-keys -t "shot$L" $k Enter; sleep 0.4; done
  sleep 0.6
  tmux send-keys -t "shot$L" -l 'say hi'
  sleep 0.5
  tmux capture-pane -e -t "shot$L" -p > "$WORK/cap-$L.ansi"
  tmux send-keys -t "shot$L" Escape
  sleep 0.5

  python3 "$HERE/ansi2html.py" "$WORK/cap-$L.ansi" "$WORK/cap-$L.html" "mud-client — MUD School ($L)"
  (cd "$HERE" && node shot.js "$WORK/cap-$L.html" "$ROOT/docs/images/client-$L.png")
  echo "docs/images/client-$L.png"
done
