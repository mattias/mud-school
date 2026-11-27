"""Generates the tutorial files from ONE source, so the three languages stay in step:

  lessons/37-*/scaffold.<lang>/server/mud/tutorial.json      (lesson 37: one language)
  lessons/38-*/scaffold.<lang>/server/mud/tutorial.all.json  (lesson 38: all three)

Run:  python3 tools/gen/tutorial.py   (from anywhere; it finds the repository)
Edit the texts below, never the generated JSON files by hand."""
import json, os, sys

LANGS = ["en", "sv", "ja"]
ROOMS = {
  "tutorial-hall": {
    "en": ("The Hall of Beginnings", "A quiet hall with whitewashed walls. Morning light falls through\na high window. A wide doorway leads north, out into a courtyard."),
    "sv": ("Begynnelsens sal", "En stilla sal med vitkalkade väggar. Morgonljus faller in genom\nett högt fönster. En bred dörröppning leder norrut, ut på en gård."),
    "ja": ("はじまりの広間", "白い漆喰の壁の静かな広間。高い窓から朝の光が差しこむ。\n北の広い戸口から中庭に出られる。"),
  },
  "tutorial-yard": {
    "en": ("The Courtyard", "A small courtyard behind high walls. Beyond the gate in the north\nwall you hear the sounds of a town."),
    "sv": ("Gården", "En liten gård bakom höga murar. Bortom porten i norra muren\nhör du ljuden från en stad."),
    "ja": ("中庭", "高い壁に囲まれた小さな中庭。北の壁の門の向こうから、\n町のざわめきが聞こえる。"),
  },
}
ITEM = {
  "en": ("an unlit torch", "A simple torch of pine wood, wrapped in tarred cloth.", ["torch"]),
  "sv": ("en otänd fackla", "En enkel fackla av furu, lindad i tjärat tyg.", ["fackla", "torch"]),
  "ja": ("火のついていないたいまつ", "松の木に、タールを染みこませた布を巻いた素朴なたいまつ。", ["たいまつ", "torch"]),
}
# The keyword used in "get ..." in each language.
GET = {"en": "torch", "sv": "fackla", "ja": "たいまつ"}
STEPS = [
  ("look", None, None, {
    "en": "Welcome, {{.Name}}! This short tutorial teaches you the basics (tutorial skip leaves it). Type look to see where you are.",
    "sv": "Välkommen, {{.Name}}! Den här korta introduktionen lär dig grunderna (tutorial skip lämnar den). Skriv look för att se var du är.",
    "ja": "ようこそ、{{.Name}}！ この短いチュートリアルで基本を覚えましょう（tutorial skip で抜けられます）。look と打って、今いる場所を見てみましょう。",
  }),
  (None, "tutorial-yard", None, {
    "en": "There is a doorway to the north. Go through it with: go north (north, or just n, takes one step at a time).",
    "sv": "Det finns en dörröppning norrut. Gå igenom den med: go north (north, eller bara n, tar ett steg i taget).",
    "ja": "北に戸口があります。go north で通り抜けましょう（north、または n だけなら 1 歩ずつ進みます）。",
  }),
  (None, None, "tutorial-torch", {
    "en": "Something lies on the ground here. Pick it up with: get {kw}",
    "sv": "Något ligger på marken här. Plocka upp det med: get {kw}",
    "ja": "地面に何か落ちています。get {kw} で拾いましょう。",
  }),
  ("inventory", None, None, {
    "en": "Check what you are carrying with: inventory (or just i)",
    "sv": "Se vad du bär på med: inventory (eller bara i)",
    "ja": "inventory（または i だけ）で持ち物を確かめましょう。",
  }),
  ("say", None, None, {
    "en": "Other players hear what you say in the same room. Try: say hello",
    "sv": "Andra spelare i samma rum hör vad du säger. Prova: say hej",
    "ja": "同じ部屋にいる他のプレイヤーには、あなたの言葉が聞こえます。say こんにちは と打ってみましょう。",
  }),
  ("help", None, None, {
    "en": "Last step ({{.Step}} of {{.Steps}}): help lists every command there is. Type help.",
    "sv": "Sista steget ({{.Step}} av {{.Steps}}): help visar alla kommandon som finns. Skriv help.",
    "ja": "最後のステップです（{{.Steps}} 中 {{.Step}}）。help ですべてのコマンドが見られます。help と打ちましょう。",
  }),
]
DONE = {
  "en": "Well done, {{.Name}}! You hand the torch to the gatekeeper, the gate opens, and you step out into Mudby.",
  "sv": "Bra gjort, {{.Name}}! Du lämnar facklan till portvakten, porten öppnas och du kliver ut i Mudby.",
  "ja": "よくできました、{{.Name}}！ たいまつを門番に返すと門が開き、あなたはマッドビーの町へ踏み出した。",
}
COORDS = {"tutorial-hall": (0, 0, {"north": "tutorial-yard"}, "inside"), "tutorial-yard": (0, -1, {"south": "tutorial-hall"}, "field")}

def room(rid, base, all_langs):
    x, y, exits, terrain = COORDS[rid]
    name, desc = ROOMS[rid][base]
    r = {"id": rid, "name": name, "description": desc, "terrain": terrain, "x": x, "y": y, "z": -10, "exits": exits}
    if all_langs:
        r["translations"] = {l: {"name": ROOMS[rid][l][0], "description": ROOMS[rid][l][1]} for l in LANGS if l != base}
    return r

def item(base, all_langs):
    name, desc, kws = ITEM[base]
    if all_langs:
        kws = sorted(set(k for l in LANGS for k in ITEM[l][2]))
    return {"id": "tutorial-torch", "name": name, "description": desc, "keywords": kws, "room": "tutorial-yard"}

def text(t, base, all_langs):
    if all_langs:
        return {l: t[l].replace("{kw}", GET[l]) for l in LANGS}
    s = t[base].replace("{kw}", GET[base])
    # Lesson 37 has no templates yet: fill them in.
    return s.replace("{{.Name}}", "").replace("Welcome, !", "Welcome!").replace("Välkommen, !", "Välkommen!").replace("ようこそ、！", "ようこそ！").replace("Well done, !", "Well done!").replace("Bra gjort, !", "Bra gjort!").replace("よくできました、！", "よくできました！").replace("({{.Step}} of {{.Steps}})", "").replace("({{.Step}} av {{.Steps}})", "").replace("（{{.Steps}} 中 {{.Step}}）", "").replace("Last step : ", "Last step: ").replace("Sista steget : ", "Sista steget: ")

def build(base, all_langs):
    steps = []
    for cmd, rid, iid, t in STEPS:
        st = {"text": text(t, base, all_langs)}
        if cmd: st["command"] = cmd
        if rid: st["room"] = rid
        if iid: st["item"] = iid
        steps.append(st)
    return {
        "rooms": [room(r, base, all_langs) for r in ["tutorial-hall", "tutorial-yard"]],
        "items": [item(base, all_langs)],
        "tutorial": {"start": "tutorial-hall", "steps": steps, "done": text(DONE, base, all_langs)},
    }

if __name__ == "__main__":
    import glob
    root = sys.argv[1] if len(sys.argv) > 1 else os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "..")
    for lesson, fname, all_langs in [("37", "tutorial.json", False), ("38", "tutorial.all.json", True)]:
        (lesson_dir,) = glob.glob(os.path.join(root, "lessons", lesson + "-*"))
        for base in LANGS:
            path = os.path.join(lesson_dir, "scaffold." + base, "server", "mud", fname)
            os.makedirs(os.path.dirname(path), exist_ok=True)
            with open(path, "w", encoding="utf-8", newline="\n") as f:
                json.dump(build(base, all_langs), f, ensure_ascii=False, indent=2)
                f.write("\n")
            print(path)
