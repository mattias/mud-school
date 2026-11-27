"""Generates the web client's page for each course language:

  lessons/39-*/scaffold.<lang>/server/web/index.html

Run:  python3 tools/gen/webpage.py   (from anywhere; it finds the repository)
The texts are in TEXT below (mud.js reads them as window.MUD_TEXT); edit them
here, not in the HTML."""
import glob, json, os, sys
TEXT = {
  "en": {"title": "MUD School", "placeholder": "Type a command and press Enter", "connected": "Connected.",
         "disconnected": "The connection was closed. Reload the page to connect again.", "offline": "Not connected",
         "online": "online", "textOnly": "(Text only – build the Rust web client in lesson 41 to see the map.)",
         "wrongAbi": "mud_web.wasm speaks another ABI version – text only.", "lang": "en",
         "help": "Arrow keys walk when the input field is empty. Click on the map to walk there."},
  "sv": {"title": "MUD-skolan", "placeholder": "Skriv ett kommando och tryck Enter", "connected": "Ansluten.",
         "disconnected": "Anslutningen stängdes. Ladda om sidan för att ansluta igen.", "offline": "Inte ansluten",
         "online": "online", "textOnly": "(Bara text – bygg webbklienten i Rust i lektion 41 för att se kartan.)",
         "wrongAbi": "mud_web.wasm talar en annan ABI-version – bara text.", "lang": "sv",
         "help": "Piltangenterna går när inmatningsfältet är tomt. Klicka på kartan för att gå dit."},
  "ja": {"title": "MUD スクール", "placeholder": "コマンドを打って Enter", "connected": "接続しました。",
         "disconnected": "接続が閉じられました。もう一度つなぐにはページを再読み込みしてください。", "offline": "未接続",
         "online": "人がオンライン", "textOnly": "（テキストのみ – マップを見るにはレッスン 41 で Rust の Web クライアントをビルドしましょう。）",
         "wrongAbi": "mud_web.wasm の ABI バージョンが違います – テキストのみ。", "lang": "ja",
         "help": "入力欄が空のときは矢印キーで歩けます。マップをクリックするとそこまで歩きます。"},
}
PAGE = """<!doctype html>
<!-- The web client's page (lesson 39). The texts are in MUD_TEXT below. -->
<html lang="{lang}">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>{title}</title>
  <link rel="stylesheet" href="style.css">
</head>
<body>
  <main>
    <div id="status">{offline}</div>
    <canvas id="zone" hidden title="{help}"></canvas>
    <div id="log" aria-live="polite"></div>
    <input id="input" autocomplete="off" autocapitalize="off" spellcheck="false" placeholder="{placeholder}">
  </main>
  <script>
    window.MUD_TEXT = {text};
  </script>
  <script src="mud.js"></script>
</body>
</html>
"""
root = sys.argv[1] if len(sys.argv) > 1 else os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "..")
(lesson_dir,) = glob.glob(os.path.join(root, "lessons", "39-*"))
for lang, t in TEXT.items():
    path = os.path.join(lesson_dir, "scaffold." + lang, "server", "web", "index.html")
    os.makedirs(os.path.dirname(path), exist_ok=True)
    js = json.dumps({k: v for k, v in t.items() if k not in ("title", "placeholder", "lang", "help")}, ensure_ascii=False)
    with open(path, "w", encoding="utf-8", newline="\n") as f:
        f.write(PAGE.format(text=js, **t))
    print(path)
