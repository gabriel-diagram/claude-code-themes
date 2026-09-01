#!/bin/bash
# Statusline del tema "Terminal" — tres bandas y el bicho a la derecha.
#
#   banda 1 · MOTOR    modelo, contexto, razonamiento, caché   -> lo que cambia cada turno
#   banda 2 · TRABAJO  repo, rama, diff, coste                 -> lo que se lleva a un commit
#   banda 3 · CUOTA    directorio, límites 5h/7d, tiempo       -> se lee de reojo, va en gris
#
# Un color por tipo de dato, siempre el mismo (ver README). 24 bits si hay COLORTERM,
# si no cae al 256 más cercano. Solo necesita python3; git es opcional.
PY_SRC=$(cat <<'PYEOF'
import sys, json, os, re, subprocess, time

try:
    d = json.load(sys.stdin)
except Exception:
    d = {}

R = "\033[0m"
B = "\033[1m"
NEGRO = "\033[38;5;16m"

# ---------------------------------------------------------------------------
# PALETA — un color por tipo de dato. (rgb, caida a 256)
# ---------------------------------------------------------------------------
_TRUE = os.environ.get("COLORTERM", "").lower() in ("truecolor", "24bit")

def _hx(s):
    return (int(s[1:3], 16), int(s[3:5], 16), int(s[5:7], 16))

def c(par, fg=True):
    lead = 38 if fg else 48
    if _TRUE:
        return "\033[%d;2;%d;%d;%dm" % ((lead,) + par[0])
    return "\033[%d;5;%dm" % (lead, par[1])

RUTA  = (_hx("#4dd6c1"),  79)   # rutas, ficheros, repos
IDENT = (_hx("#57e389"),  78)   # identificadores, altas
LINK  = (_hx("#6fb6ff"),  75)   # urls, ramas, enlaces
NUM   = (_hx("#e8c46a"), 179)   # números, dinero, métricas, avisos
MODO  = (_hx("#b07cf0"), 141)   # modos y ajustes del propio CLI
MAL   = (_hx("#f2777a"), 210)   # bajas, errores, riesgo
ENF   = (_hx("#eceff4"), 255)   # énfasis
TENUE = (_hx("#6b7683"), 243)   # flechas, separadores, unidades
RAYA  = (_hx("#2c343c"), 236)   # el separador vertical
DIR   = (_hx("#8d99a6"), 246)   # el directorio, un gris por encima del tenue
CUOTA = (_hx("#4ea3f5"),  75)   # barras de límite
VACIO = (_hx("#1d2b38"), 235)   # hueco de las barras de límite
CTXV  = (_hx("#24382c"), 235)   # hueco de la barra de contexto

SEP = c(RAYA) + "│" + R
ANSI = re.compile("\033\\[[0-9;]*m")

def g(o, *ks):
    for k in ks:
        o = (o or {}).get(k) if isinstance(o, dict) else None
    return o

def vis(s):
    return len(ANSI.sub("", s))

# ---------------------------------------------------------------------------
# ANCHO — no llega en el JSON, sale de COLUMNS. Claude Code recorta la linea
# unos 5 caracteres antes de ese valor, de ahi el margen. Ver README.
# ---------------------------------------------------------------------------
def term_width():
    cols = os.environ.get("COLUMNS")
    if cols and cols.isdigit() and int(cols) > 0:
        return int(cols)
    try:
        return os.get_terminal_size(os.open("/dev/tty", os.O_RDONLY)).columns
    except Exception:
        return 80

try:
    _rp = max(1, int(os.environ.get("STATUSLINE_RIGHT_PAD", "6")))
except ValueError:
    _rp = 6
WIDTH = max(20, term_width() - _rp)

BICHO_W = 9
BICHO_GAP = 2
BICHO = (os.environ.get("STATUSLINE_BICHO", "1").lower() not in ("0", "off", "no")
         and WIDTH >= 55)
CW = WIDTH - (BICHO_W + BICHO_GAP) if BICHO else WIDTH

# ---------------------------------------------------------------------------
# ENSAMBLADO — cada banda suelta los elementos de menor prioridad antes que
# hacer wrap, que descuadra la caja del prompt.
# ---------------------------------------------------------------------------
SEPX = " " + SEP + " "

def assemble(segs, width):
    items = [s for s in segs if s]
    def total(its):
        return sum(it["vis"] + (0 if i == 0 else it["sepw"]) for i, it in enumerate(its))
    while len(items) > 1 and total(items) > width:
        v = max(range(len(items)), key=lambda i: (items[i]["p"], i))
        items.pop(v)
    if len(items) == 1 and items[0]["vis"] > width:
        it = items[0]
        if it.get("color") is not None and it.get("plain") is not None:
            keep = max(1, width - 1)
            t = it["plain"][:keep] + "…"
            it = dict(it, txt=it["color"] + t + R, vis=vis(it["color"] + t + R))
        else:
            t = ANSI.sub("", it["txt"])[:max(1, width - 1)] + "…"
            it = dict(it, txt=t, vis=len(t))
        items = [it]
    out = ""
    for i, it in enumerate(items):
        out += ("" if i == 0 else it["sep"]) + it["txt"]
    return out

def seg(p, txt, color=None, plain=None, sep=SEPX):
    return {"p": p, "txt": txt, "vis": vis(txt), "sep": sep,
            "sepw": vis(sep), "color": color, "plain": plain}

def padr(s, w):
    n = w - vis(s)
    return s + (" " * n) if n > 0 else s

def ctr(plain, coloreado, w=BICHO_W):
    n = w - len(plain)
    if n < 0:
        return coloreado
    return " " * (n // 2) + coloreado + " " * (n - n // 2)

def barra(p, w, lleno, hueco):
    p = max(0.0, min(100.0, float(p)))
    f = int(round(w * p / 100.0))
    return c(lleno) + ("█" * f) + c(hueco) + ("░" * (w - f)) + R

# ---------------------------------------------------------------------------
# DATOS
# ---------------------------------------------------------------------------
modelo = g(d, "model", "display_name") or "?"
modelo = re.sub(r"\s*\(.*\)\s*$", "", modelo)          # "Opus 5 (1M context)" -> "Opus 5"
cwd = g(d, "workspace", "current_dir") or d.get("cwd") or os.getcwd()
home = os.path.expanduser("~")
ddir = "~" + cwd[len(home):] if cwd.startswith(home) else cwd
partes = ddir.rstrip("/").split("/")
dircorto = "/".join(partes[-3:]) if len(partes) > 3 else ddir

esfuerzo = g(d, "effort", "level")
vim = g(d, "vim", "mode")
coste = g(d, "cost", "total_cost_usd")
altas = g(d, "cost", "total_lines_added")
bajas = g(d, "cost", "total_lines_removed")
durms = g(d, "cost", "total_duration_ms")
pct = g(d, "context_window", "used_percentage")
ctxsz = g(d, "context_window", "context_window_size")
cache = g(d, "prompt_cache", "hit_ratio")
rl5 = g(d, "rate_limits", "five_hour", "used_percentage")
rl7 = g(d, "rate_limits", "seven_day", "used_percentage")

def ctxlabel(n):
    try:
        n = int(n)
    except (TypeError, ValueError):
        return None
    if n >= 1000000:
        return "%gM ctx" % (n / 1000000.0)
    return "%gk ctx" % (n / 1000.0)

def duracion(ms):
    try:
        s = int(float(ms) / 1000)
    except (TypeError, ValueError):
        return None
    m, s = divmod(s, 60)
    h, m = divmod(m, 60)
    if h:
        return "%dh %02dm" % (h, m)
    if m:
        return "%dm %02ds" % (m, s)
    return "%ds" % s

repo = None
rama = None
sucio = ""
try:
    p = subprocess.run(["git", "-C", cwd, "rev-parse", "--show-toplevel", "--abbrev-ref", "HEAD"],
                       capture_output=True, text=True, timeout=1).stdout.split("\n")
    if len(p) >= 2 and p[0].strip():
        raiz = p[0].strip()
        trozos = raiz.rstrip("/").split("/")
        repo = "/".join(trozos[-2:]) if len(trozos) > 1 else trozos[-1]
        rama = p[1].strip() or None
        est = subprocess.run(["git", "-C", cwd, "status", "--porcelain"],
                             capture_output=True, text=True, timeout=1).stdout.strip()
        if est:
            sucio = " " + c(NUM) + "✳" + R
except Exception:
    pass

# ---------------------------------------------------------------------------
# BANDA 1 · MOTOR
# ---------------------------------------------------------------------------
b1 = [seg(0, c(IDENT, False) + NEGRO + B + " " + modelo + " " + R,
          color=c(IDENT, False) + NEGRO + B, plain=" " + modelo + " ")]

if pct is not None:
    niv = IDENT if float(pct) < 60 else (NUM if float(pct) < 85 else MAL)
    b1.append(seg(1, barra(pct, 16, niv, CTXV) + " " + c(ENF) + B + "%d%%" % round(float(pct)) + R,
                  sep=" "))
_cl = ctxlabel(ctxsz)
if _cl:
    b1.append(seg(3, c(TENUE) + _cl + R, color=c(TENUE), plain=_cl,
                  sep=" " + c(RAYA) + "·" + R + " "))
if esfuerzo:
    b1.append(seg(2, c(MODO) + str(esfuerzo) + R, color=c(MODO), plain=str(esfuerzo)))
if cache is not None:
    try:
        _t = "%d%%" % round(float(cache) * 100)
        b1.append(seg(4, c(NUM) + _t + R + c(TENUE) + " caché" + R))
    except (TypeError, ValueError):
        pass
if vim:
    vc = c(MODO) if vim == "INSERT" else c(TENUE)
    b1.append(seg(5, vc + B + str(vim) + R, color=vc + B, plain=str(vim)))

# ---------------------------------------------------------------------------
# BANDA 2 · TRABAJO
# ---------------------------------------------------------------------------
b2 = []
_nom = repo or partes[-1]
b2.append(seg(0, c(RUTA) + _nom + R, color=c(RUTA), plain=_nom))
if rama:
    b2.append(seg(1, c(TENUE) + "(" + R + c(LINK) + rama + R + sucio + c(TENUE) + ")" + R, sep=" "))
if altas is not None or bajas is not None:
    b2.append(seg(2, c(IDENT) + "+" + str(altas or 0) + R + c(RAYA) + "/" + R +
                     c(MAL) + "−" + str(bajas or 0) + R))
if coste is not None:
    _t = "$" + format(float(coste), ".2f")
    b2.append(seg(3, c(NUM) + _t + R, color=c(NUM), plain=_t))

# ---------------------------------------------------------------------------
# BANDA 3 · CUOTA
# ---------------------------------------------------------------------------
b3 = [seg(1, c(DIR) + dircorto + R, color=c(DIR), plain=dircorto)]
for _v, _et in ((rl5, "5h"), (rl7, "7d")):
    if _v is None:
        continue
    b3.append(seg(2, c(TENUE) + _et + " " + R + barra(_v, 10, CUOTA, VACIO) +
                     c(TENUE) + " %d%%" % round(float(_v)) + R, sep="  "))
_dur = duracion(durms)
if _dur:
    b3.append(seg(3, c(TENUE) + _dur + R, color=c(TENUE), plain=_dur))

# ---------------------------------------------------------------------------
# EL BICHO — 9x5, la silueta con antenas, orejas y patas.
# El uso es una MEDIA PONDERADA: el contexto pesa 50, el limite de 5h 30 y el
# de 7d 20, porque el contexto es lo unico que puedes gestionar en el momento;
# los limites solo avisan. Si falta alguno, se reparte su peso entre los otros.
# ---------------------------------------------------------------------------
PALE_V = (_hx("#d8ffe9"), 194); OSC_V = (_hx("#2f7a52"),  29)
PALE_T = (_hx("#d6fffa"), 195); OSC_T = (_hx("#2b7d74"),  30)
PALE_A = (_hx("#d5e9ff"), 189); OSC_A = (_hx("#2a5c8a"),  24)
PALE_I = (_hx("#cdd2ff"), 189); OSC_I = (_hx("#333c9e"),  61)
GRIS   = (_hx("#5a6270"), 242)
PALE_G = (_hx("#8891a0"), 245); OSC_G = (_hx("#3d434d"), 238)

ESTADOS = [
    dict(tope=22,     et="fresca",   col=IDENT,               ojoc=PALE_V, patc=OSC_V,
         ojos=(">", "<"),                 cabeza=True, anda=True, chispa=True),
    dict(tope=45,     et="vibrante", col=IDENT,               ojoc=PALE_V, patc=OSC_V,
         ojos=(">", "<"),                 cabeza=True, anda=True),
    dict(tope=63,     et="a gusto",  col=RUTA,                ojoc=PALE_T, patc=OSC_T,
         ojos=("●", "●"),       cabeza=True, anda=True),
    dict(tope=78,     et="espesa",   col=CUOTA,               ojoc=PALE_A, patc=OSC_A,
         ojos=("▬", "▬"),       cabeza=True),
    dict(tope=89,     et="cansada",  col=CUOTA,               ojoc=PALE_A, patc=OSC_A,
         ojos=("◠", "◠")),
    dict(tope=99.999, et="ahogada",  col=(_hx("#5865f2"), 63), ojoc=PALE_I, patc=OSC_I,
         ojos=("✕", "✕")),
    dict(tope=100,    et="k.o.",     col=GRIS,                ojoc=PALE_G, patc=OSC_G,
         ojos=("✕", "✕"),       ko=True),
]

num = 0.0
den = 0.0
for _v, _w in ((pct, 0.5), (rl5, 0.3), (rl7, 0.2)):
    try:
        if _v is not None:
            num += float(_v) * _w
            den += _w
    except (TypeError, ValueError):
        pass
uso = (num / den) if den else 0.0
uso = max(0.0, min(100.0, uso))

E = next((e for e in ESTADOS if uso <= e["tope"]), ESTADOS[-1])
paso = int(time.time())
# El diseno pide que las patas alternen en CADA refresco. A 1 fps eso es un
# baile continuo en la esquina del ojo: STATUSLINE_BICHO_CALMA=1 lo deja
# andando solo 4 segundos de cada 12.
_calma = os.environ.get("STATUSLINE_BICHO_CALMA", "").lower() in ("1", "on", "yes")
anda = bool(E.get("anda")) and (paso % 12 < 4 if _calma else True)

if E.get("ko"):
    patas = "  ▖ ▗ ▖ ▗"
elif anda:
    patas = "  ▘   ▝  " if paso % 2 == 0 else "  ▝   ▘  "
else:
    patas = "  ▖   ▗  "

ojos = ("◠", "◠") if (E.get("anda") and paso % 7 == 3) else E["ojos"]

COL = c(E["col"])
PAT = c(E["patc"])
OJO = c(E["ojoc"])

if E.get("ko"):
    fila1 = PAT + patas + R
elif E.get("cabeza"):
    fila1 = COL + "  ╲   ╱  " + R
else:
    fila1 = " " * 9
fila2 = COL + (" ▗█████▖ " if E.get("cabeza")
               else " ▗▄▄▄▄▄▖ ") + R
fila3 = (COL + "▐█ " + OJO + ojos[0] + COL + " " + OJO + ojos[1] +
         COL + " █▌" + R)
fila4 = COL + (" ▀▀▀▀▀▀▀ " if E.get("ko")
               else " ▝█████▘ ") + R
fila5 = (" " * 9) if E.get("ko") else (PAT + patas + R)

# Al cruzar un umbral la etiqueta sale en negrita un solo refresco. Hace falta
# recordar el estado anterior, asi que se guarda en un fichero por sesion.
salta = False
_sid = d.get("session_id")
if _sid and re.match(r"^[A-Za-z0-9_-]{1,64}$", str(_sid)):
    _f = os.path.join(os.environ.get("TMPDIR", "/tmp"), "claude-statusline-" + str(_sid))
    try:
        with open(_f) as _fh:
            salta = _fh.read().strip() != E["et"]
    except Exception:
        pass
    try:
        with open(_f, "w") as _fh:
            _fh.write(E["et"])
    except Exception:
        pass

etq = E["et"] + (" ✦" if E.get("chispa") else "")
tarjeta = [ctr(etq, COL + (B if salta else "") + etq + R), fila1, fila2, fila3, fila4, fila5]

# ---------------------------------------------------------------------------
# SALIDA
# ---------------------------------------------------------------------------
linea1 = assemble(b1, WIDTH)
izq = [assemble(b2, CW), assemble(b3, CW), "", "", "", ""]

if BICHO and all(vis(x) <= CW for x in izq):
    # Claude Code recorta los espacios del principio de la linea: si la izquierda
    # va vacia, el sangrado desaparece y el bicho se cae al borde. Se ancla con un
    # braille en blanco (U+2800), que se pinta vacio pero no es un espacio.
    print(linea1)
    for _l, _t in zip(izq, tarjeta):
        if not ANSI.sub("", _l).strip():
            _l = "⠀"
        print(padr(_l, CW) + " " * BICHO_GAP + _t)
else:
    print(linea1)
    print(assemble(b2, WIDTH))
    print(assemble(b3, WIDTH))
PYEOF
)
exec python3 -c "$PY_SRC"
