// The web client (lessons 39 and 41) – given, you don't need to change it.
//
// It speaks the same JSON protocol as the Rust client (PROTOCOL.md) over a
// WebSocket. Text frames go to the log, prompts to the input field. If the
// Rust part has been built (mud_web.wasm, lesson 41), it gets every line too
// and draws the zone on the <canvas>; without it you play with text only.

"use strict";

const T = window.MUD_TEXT; // the texts, set in index.html
const log = document.getElementById("log");
const input = document.getElementById("input");
const statusLine = document.getElementById("status");
const canvas = document.getElementById("zone");

let ws = null;
let wasm = null; // the Rust client's exports, once loaded

// ---- the log -------------------------------------------------------------

function print(kind, text) {
  const atBottom = log.scrollTop + log.clientHeight >= log.scrollHeight - 4;
  const div = document.createElement("div");
  div.className = "line " + (kind || "info");
  div.textContent = text;
  log.appendChild(div);
  while (log.childNodes.length > 500) log.removeChild(log.firstChild);
  if (atBottom) log.scrollTop = log.scrollHeight;
}

// ---- the connection --------------------------------------------------------

function send(line) {
  if (ws && ws.readyState === WebSocket.OPEN) ws.send(line + "\n");
}

function connect() {
  const scheme = location.protocol === "https:" ? "wss:" : "ws:";
  ws = new WebSocket(scheme + "//" + location.host + "/ws");
  let buffered = "";
  ws.onopen = () => {
    print("info", T.connected);
    send("@hello mudschool 1"); // we want JSON (PROTOCOL.md)
  };
  ws.onmessage = (ev) => {
    // A message can hold part of a line, or several lines.
    buffered += ev.data;
    let i;
    while ((i = buffered.indexOf("\n")) >= 0) {
      const line = buffered.slice(0, i).replace(/\r$/, "");
      buffered = buffered.slice(i + 1);
      if (line !== "") onLine(line);
    }
  };
  ws.onclose = () => {
    print("error", T.disconnected);
    statusLine.textContent = T.offline;
    input.type = "text";
  };
}

function onLine(line) {
  let f;
  try {
    f = JSON.parse(line);
  } catch {
    print("info", line); // not JSON: show it as it is
    return;
  }
  if (wasm) feedWasm(line);
  switch (f.type) {
    case "text":
      print(f.kind, f.text);
      if (input.type === "password") input.type = "text";
      break;
    case "prompt":
      input.type = f.prompt === "password" || f.prompt === "new_password" ? "password" : "text";
      break;
    case "status": {
      const s = f.status;
      statusLine.textContent = `${s.name} · ${s.room} · ${s.hp}/${s.max_hp} HP · ${s.online} ${T.online}`;
      break;
    }
  }
}

// ---- input -------------------------------------------------------------------

const arrows = { ArrowUp: "north", ArrowDown: "south", ArrowLeft: "west", ArrowRight: "east" };

input.addEventListener("keydown", (ev) => {
  if (ev.key === "Enter") {
    const line = input.value;
    input.value = "";
    if (input.type !== "password" && line !== "") print("echo", "> " + line);
    send(line);
    return;
  }
  if (arrows[ev.key] && input.value === "") {
    ev.preventDefault();
    send(arrows[ev.key]);
  }
});

// ---- the Rust part (lesson 41) ---------------------------------------------

// The ABI between this page and mud_web.wasm – the same idea as the plugins
// in lesson 33, but now the browser is the host:
//   exports: memory, abi_version() -> 1, alloc(len) -> ptr, dealloc(ptr, len),
//            on_line(ptr, len), frame(cols, rows, t_ms) -> ptr,
//            click(col, row), tick(t_ms)
//   imports: mud.send(ptr, len) – a command to send to the server
const ABI_VERSION = 1;
const CELL_BYTES = 12; // u32 char, fg r g b, bg r g b, has_bg, (unused)
const enc = new TextEncoder();
const dec = new TextDecoder();

function feedWasm(line) {
  const bytes = enc.encode(line);
  const ptr = wasm.alloc(bytes.length);
  new Uint8Array(wasm.memory.buffer, ptr, bytes.length).set(bytes);
  wasm.on_line(ptr, bytes.length);
  wasm.dealloc(ptr, bytes.length);
}

async function loadWasm() {
  let module;
  try {
    const res = await fetch("mud_web.wasm");
    if (!res.ok) return false; // not built yet: text only
    module = await res.arrayBuffer();
  } catch {
    return false;
  }
  const imports = {
    mud: {
      send: (ptr, len) => send(dec.decode(new Uint8Array(wasm.memory.buffer, ptr, len))),
    },
  };
  const { instance } = await WebAssembly.instantiate(module, imports);
  const e = instance.exports;
  if (e.abi_version() !== ABI_VERSION) {
    print("error", T.wrongAbi);
    return false;
  }
  wasm = e;
  return true;
}

// Drawing: one cell is CELL_W × CELL_H pixels.
const FONT_PX = 18;
let CELL_W = 11;
const CELL_H = 20;

function draw(t) {
  const cols = Math.floor(canvas.width / CELL_W);
  const rows = Math.floor(canvas.height / CELL_H);
  const ptr = wasm.frame(cols, rows, t);
  const view = new DataView(wasm.memory.buffer, ptr, cols * rows * CELL_BYTES);
  const g = canvas.getContext("2d");
  g.fillStyle = "#000";
  g.fillRect(0, 0, canvas.width, canvas.height);
  g.font = `${FONT_PX}px ui-monospace, Menlo, Consolas, monospace`;
  g.textBaseline = "top";
  for (let y = 0; y < rows; y++) {
    for (let x = 0; x < cols; x++) {
      const o = (y * cols + x) * CELL_BYTES;
      if (view.getUint8(o + 10)) {
        g.fillStyle = `rgb(${view.getUint8(o + 7)},${view.getUint8(o + 8)},${view.getUint8(o + 9)})`;
        g.fillRect(x * CELL_W, y * CELL_H, CELL_W, CELL_H);
      }
      const ch = view.getUint32(o, true);
      if (ch !== 0 && ch !== 32) {
        g.fillStyle = `rgb(${view.getUint8(o + 4)},${view.getUint8(o + 5)},${view.getUint8(o + 6)})`;
        g.fillText(String.fromCodePoint(ch), x * CELL_W, y * CELL_H + 1);
      }
    }
  }
}

function startDrawing() {
  canvas.hidden = false;
  const resize = () => {
    canvas.width = canvas.clientWidth;
    canvas.height = canvas.clientHeight;
    const g = canvas.getContext("2d");
    g.font = `${FONT_PX}px ui-monospace, Menlo, Consolas, monospace`;
    CELL_W = Math.ceil(g.measureText("M").width);
  };
  resize();
  window.addEventListener("resize", resize);
  canvas.addEventListener("click", (ev) => {
    const r = canvas.getBoundingClientRect();
    wasm.click(Math.floor((ev.clientX - r.left) / CELL_W), Math.floor((ev.clientY - r.top) / CELL_H));
    input.focus();
  });
  const loop = (t) => {
    wasm.tick(t);
    draw(t);
    requestAnimationFrame(loop);
  };
  requestAnimationFrame(loop);
}

// ---- start -------------------------------------------------------------------

loadWasm().then((ok) => {
  if (ok) startDrawing();
  else print("info", T.textOnly);
  connect();
  input.focus();
});
