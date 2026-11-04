# Converts a `tmux capture-pane -e` dump (ANSI colours) into an HTML page that
# draws every terminal cell at a fixed position, so wide (CJK) characters always
# take exactly two cells. Usage: ansi2html.py <in.ansi> <out.html> <title>
import sys, re, unicodedata, html
src, dst, title = sys.argv[1], sys.argv[2], sys.argv[3]
COLS, ROWS = 80, 24
CW, CH = 9.33, 18
PALETTE16 = ['#000000','#cd3131','#0dbc79','#e5e510','#2472c8','#bc3fbc','#11a8cd','#e5e5e5',
             '#666666','#f14c4c','#23d18b','#f5f543','#3b8eea','#d670d6','#29b8db','#ffffff']
DEF_FG, DEF_BG = '#d4d4d4', '#14141c'
def c256(n):
    if n < 16: return PALETTE16[n]
    if n < 232:
        n -= 16; r, g, b = n // 36, (n // 6) % 6, n % 6
        f = lambda v: 0 if v == 0 else 55 + v * 40
        return '#%02x%02x%02x' % (f(r), f(g), f(b))
    v = 8 + (n - 232) * 10
    return '#%02x%02x%02x' % (v, v, v)
def width(ch):
    return 2 if unicodedata.east_asian_width(ch) in ('W', 'F') else 1
cells = []
for line in open(src, encoding='utf-8').read().split('\n')[:ROWS]:
    fg, bg, bold = None, None, False
    row, col, i = [], 0, 0
    while i < len(line):
        m = re.match(r'\x1b\[([0-9;]*)m', line[i:])
        if m:
            ps = [int(p) if p else 0 for p in m.group(1).split(';')] or [0]
            j = 0
            while j < len(ps):
                p = ps[j]
                if p == 0: fg, bg, bold = None, None, False
                elif p == 1: bold = True
                elif p == 22: bold = False
                elif p == 39: fg = None
                elif p == 49: bg = None
                elif 30 <= p <= 37: fg = PALETTE16[p - 30]
                elif 90 <= p <= 97: fg = PALETTE16[p - 90 + 8]
                elif 40 <= p <= 47: bg = PALETTE16[p - 40]
                elif p in (38, 48):
                    if ps[j + 1] == 2:
                        c = '#%02x%02x%02x' % tuple(ps[j + 2:j + 5]); j += 4
                    else:
                        c = c256(ps[j + 2]); j += 2
                    if p == 38: fg = c
                    else: bg = c
                j += 1
            i += m.end()
            continue
        ch = line[i]; i += 1
        w = width(ch)
        row.append((col, w, ch, fg, bg, bold))
        col += w
    cells.append(row)
out = []
for y, row in enumerate(cells):
    for (x, w, ch, fg, bg, bold) in row:
        style = 'left:%.1fpx;top:%dpx;width:%.1fpx;' % (x * CW, y * CH, w * CW + 0.6)
        if bg: style += 'background:%s;' % bg
        style += 'color:%s;' % (fg or DEF_FG)
        if bold: style += 'font-weight:bold;'
        if ch == ' ' and not bg: continue
        out.append('<span style="%s">%s</span>' % (style, html.escape(ch)))
W, H = COLS * CW, ROWS * CH
page = f'''<!doctype html><meta charset="utf-8"><style>
body{{margin:0;background:#0b0b10}}
.win{{margin:0;padding:0;background:#0b0b10;display:inline-block}}
.bar{{height:30px;background:#2a2a36;border-radius:10px 10px 0 0;display:flex;align-items:center;padding:0 12px;gap:8px;font:13px "DejaVu Sans",sans-serif;color:#aaa}}
.dot{{width:12px;height:12px;border-radius:6px;display:inline-block}}
.term{{position:relative;width:{W}px;height:{H}px;background:{DEF_BG};padding:12px 14px;border-radius:0 0 10px 10px}}
.term span{{position:absolute;display:inline-block;height:{CH}px;line-height:{CH}px;text-align:center;white-space:pre;
font-family:"DejaVu Sans Mono","Noto Sans Mono CJK JP","Noto Sans CJK JP",monospace;font-size:15.5px;margin-left:14px;margin-top:12px}}
</style><div class="win"><div class="bar"><span class="dot" style="background:#ff5f57"></span><span class="dot" style="background:#febc2e"></span><span class="dot" style="background:#28c840"></span><span style="margin-left:10px">{html.escape(title)}</span></div><div class="term">{''.join(out)}</div></div>'''
open(dst, 'w', encoding='utf-8').write(page)
