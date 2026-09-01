#!/bin/bash
# Statusline del tema "Terminal" — tres bandas y el bicho a la derecha.
#
#   banda 1 · MOTOR    modelo, contexto, razonamiento, caché   -> lo que cambia cada turno
#   banda 2 · TRABAJO  repo, rama, diff, coste                 -> lo que se lleva a un commit
#   banda 3 · CUOTA    directorio, límites 5h/7d, tiempo       -> se lee de reojo, va en gris
#
# Un color por tipo de dato, siempre el mismo (ver README). 24 bits si hay COLORTERM,
# si no cae al 256 más cercano. Solo necesita python3; git es opcional.
#
# Las plantillas del bicho y la lógica de progreso viven en bicho.py, al lado de
# este fichero. Si no está, la statusline sigue funcionando sin bicho.
SL_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)
export SL_DIR
PY_SRC=$(cat <<'PYEOF'
import sys, json, os, re, subprocess, time

sys.path.insert(0, os.environ.get("SL_DIR", ""))
try:
    import bicho as BI
except Exception:
    BI = None

try:
    d = json.load(sys.stdin)
except Exception:
    d = {}

R = "\033[0m"
B = "\033[1m"
NEGRO = "\033[38;5;16m"

_TRUE = os.environ.get("COLORTERM", "").lower() in ("truecolor", "24bit")


def _hx(s):
    return (int(s[1:3], 16), int(s[3:5], 16), int(s[5:7], 16))


def c(par, fg=True):
    lead = 38 if fg else 48
    if _TRUE:
        return "\033[%d;2;%d;%d;%dm" % ((lead,) + par[0])
    return "\033[%d;5;%dm" % (lead, par[1])


# ---------------------------------------------------------------------------
# PALETA — un color por tipo de dato. La comparte bicho.py; aqui se replica lo
# justo para que la statusline pinte aunque el modulo no este instalado.
# ---------------------------------------------------------------------------
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


def num(v):
    """Todo numero del JSON pasa por aqui. Un valor raro vale None, no un crash:
    la statusline entera desaparece si esto lanza."""
    if v is None or isinstance(v, bool):
        return None
    try:
        f = float(v)
    except (TypeError, ValueError):
        return None
    return None if f != f or f in (float("inf"), float("-inf")) else f


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
    fd = None
    try:
        fd = os.open("/dev/tty", os.O_RDONLY)
        return os.get_terminal_size(fd).columns
    except Exception:
        return 80
    finally:
        if fd is not None:
            try:
                os.close(fd)
            except OSError:
                pass


try:
    _rp = max(1, int(os.environ.get("STATUSLINE_RIGHT_PAD", "6")))
except ValueError:
    _rp = 6
COLS = term_width()
WIDTH = max(20, COLS - _rp)

BICHO_W = 9
BICHO_GAP = 2
BICHO = (BI is not None
         and os.environ.get("STATUSLINE_BICHO", "1").lower() not in ("0", "off", "no")
         and WIDTH >= 55)
CW = WIDTH - (BICHO_W + BICHO_GAP) if BICHO else WIDTH
BOCADILLO_MIN = 100     # el diseño: con menos columnas el bocadillo se cae solo.
                        # Se mide contra las columnas del terminal, no contra el
                        # ancho util, que es lo que dice el diseño.

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


def padl(s, w):
    n = w - vis(s)
    return (" " * n) + s if n > 0 else s


def ctr(plain, coloreado, w=BICHO_W):
    n = w - len(plain)
    if n < 0:
        return coloreado
    return " " * (n // 2) + coloreado + " " * (n - n // 2)


def barra(p, w, lleno, hueco):
    p = max(0.0, min(100.0, p))
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
coste = num(g(d, "cost", "total_cost_usd"))
altas = g(d, "cost", "total_lines_added")
bajas = g(d, "cost", "total_lines_removed")
durms = num(g(d, "cost", "total_duration_ms"))
pct = num(g(d, "context_window", "used_percentage"))
ctxsz = num(g(d, "context_window", "context_window_size"))
cache = num(g(d, "prompt_cache", "hit_ratio"))
rl5 = num(g(d, "rate_limits", "five_hour", "used_percentage"))
rl7 = num(g(d, "rate_limits", "seven_day", "used_percentage"))


def ctxlabel(n):
    if n is None:
        return None
    n = int(n)
    if n >= 1000000:
        return "%gM ctx" % (n / 1000000.0)
    return "%gk ctx" % (n / 1000.0)


def duracion(ms):
    if ms is None:
        return None
    s = int(ms / 1000)
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
    p = subprocess.run(["git", "--no-optional-locks", "-C", cwd,
                        "rev-parse", "--show-toplevel", "--abbrev-ref", "HEAD"],
                       capture_output=True, text=True, timeout=1).stdout.split("\n")
    if len(p) >= 2 and p[0].strip():
        raiz = p[0].strip()
        trozos = raiz.rstrip("/").split("/")
        repo = "/".join(trozos[-2:]) if len(trozos) > 1 else trozos[-1]
        rama = p[1].strip() or None
        est = subprocess.run(["git", "--no-optional-locks", "-C", cwd,
                              "status", "--porcelain"],
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
    niv = IDENT if pct < 60 else (NUM if pct < 85 else MAL)
    b1.append(seg(1, barra(pct, 16, niv, CTXV) + " " + c(ENF) + B + "%d%%" % round(pct) + R,
                  sep=" "))
_cl = ctxlabel(ctxsz)
if _cl:
    b1.append(seg(3, c(TENUE) + _cl + R, color=c(TENUE), plain=_cl,
                  sep=" " + c(RAYA) + "·" + R + " "))
if esfuerzo:
    b1.append(seg(2, c(MODO) + str(esfuerzo) + R, color=c(MODO), plain=str(esfuerzo)))
if cache is not None:
    _t = "%d%%" % round(cache * 100)
    b1.append(seg(4, c(NUM) + _t + R + c(TENUE) + " caché" + R))
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
    _t = "$" + format(coste, ".2f")
    b2.append(seg(3, c(NUM) + _t + R, color=c(NUM), plain=_t))

# ---------------------------------------------------------------------------
# BANDA 3 · CUOTA
# ---------------------------------------------------------------------------
b3 = [seg(1, c(DIR) + dircorto + R, color=c(DIR), plain=dircorto)]
for _v, _et in ((rl5, "5h"), (rl7, "7d")):
    if _v is None:
        continue
    b3.append(seg(2, c(TENUE) + _et + " " + R + barra(_v, 10, CUOTA, VACIO) +
                     c(TENUE) + " %d%%" % round(_v) + R, sep="  "))
_dur = duracion(durms)
if _dur:
    b3.append(seg(3, c(TENUE) + _dur + R, color=c(TENUE), plain=_dur))

# ---------------------------------------------------------------------------
# EL BICHO — dos capas que no se mezclan.
#   vida      del momento, del uso de contexto y cuota: ojos, patas, color.
#   progreso  permanente, de la XP de ~/.claude/pet.json: la silueta.
# ---------------------------------------------------------------------------
tarjeta = None
bocadillo = None
if BICHO:
    uso = BI.uso_ponderado(pct, rl5, rl7)
    E = BI.estado_de(uso)
    etq = E[1]

    pet = BI.leer_pet()
    BI.pasar_hambre(pet)
    criatura, nivel = BI.evolucion(pet)
    hambre = int(pet.get("hambre", 0))

    paso = int(time.time())
    filas = BI.dibujar(criatura, E, paso=paso, hambre=hambre, veinticuatro=_TRUE)

    # Al cruzar un umbral la etiqueta sale en negrita un refresco. Hace falta
    # recordar el estado anterior, y se guarda en un fichero por sesion.
    salta = False
    _sid = d.get("session_id")
    if _sid and re.match(r"^[A-Za-z0-9_-]{1,64}$", str(_sid)):
        _tmp = os.environ.get("TMPDIR", "/tmp")
        _f = os.path.join(_tmp, "claude-statusline-" + str(_sid))
        # El fichero de sesion guarda la etiqueta anterior y, de paso, los hechos
        # que solo la statusline ve: el pico de uso, cuando empezo y en que repo.
        # Un hook de SessionEnd los convierte en contadores de comportamiento.
        _antes = None
        _hechos = {}
        try:
            with open(_f) as _fh:
                _crudo = _fh.read().strip()
            if _crudo.startswith("{"):
                _hechos = json.loads(_crudo)
                _antes = _hechos.get("etq")
            else:
                _antes = _crudo or None
        except Exception:
            pass
        _pico = max(float(_hechos.get("pico", 0)), uso)
        _sube_pico = _pico >= float(_hechos.get("pico", 0)) + 1.0
        salta = _antes is not None and _antes != etq
        # Solo se escribe cuando cambia: esto corre en cada refresco.
        if _antes != etq or _sube_pico:
            # El reventon del contexto no lo puede ver ningun hook —- los hooks
            # no reciben el uso— asi que se cobra aqui, y solo en el cruce a
            # k.o., que pasa una vez por sesion como mucho.
            if etq == "k.o." and _antes is not None:
                try:
                    _p2 = BI.leer_pet()
                    _p2, _ok = BI.alimentar(_p2, "reventon")
                    if _ok:
                        BI.contar(_p2, "impulsivo")
                        BI.contar(_p2, "ctx_limite")
                        BI.contar(_p2, "sesiones_ctx100")
                        _p2["contadores"]["racha_tests"] = 0
                        _p2["contadores"]["racha_diffs"] = 0
                        BI.escribir_pet(_p2)
                except Exception:
                    pass
            try:
                with open(_f, "w") as _fh:
                    json.dump({"etq": etq, "pico": round(_pico, 1),
                               "t0": _hechos.get("t0") or int(time.time()),
                               "repo": repo or ""}, _fh)
            except Exception:
                pass
            # De paso, barrer los huerfanos de sesiones muertas (mas de un dia).
            try:
                _limite = time.time() - 86400
                for _n in os.listdir(_tmp):
                    if _n.startswith("claude-statusline-"):
                        _p = os.path.join(_tmp, _n)
                        if os.path.getmtime(_p) < _limite:
                            os.unlink(_p)
            except Exception:
                pass

    COL = c(E[2])
    _e = etq + (" ✦" if E[8] else "")
    tarjeta = [ctr(_e, COL + (B if salta else "") + _e + R)] + filas

    # El bocadillo sale solo cuando hay algo que decir, y solo si cabe.
    if COLS >= BOCADILLO_MIN:
        if pet.get("_subio"):
            bocadillo = "nivel %d. %s" % (nivel, BI.BONITO.get(criatura, criatura))
        elif hambre >= BI.HAMBRE_AVISA:
            bocadillo = "llevo %dh sin comer. /feed" % hambre
        elif etq == "k.o.":
            bocadillo = "100% de contexto. avisé. no pasa nada."
        elif etq == "cansada" and durms and durms > 4 * 3600 * 1000:
            bocadillo = "llevas 4h. yo también estaría espeso"
        elif pet.get("racha", 0) >= 3:
            bocadillo = "racha de %d días. no la rompas hoy" % pet["racha"]

# ---------------------------------------------------------------------------
# SALIDA
# ---------------------------------------------------------------------------
linea1 = assemble(b1, WIDTH)
izq = [assemble(b2, CW), assemble(b3, CW), "", "", "", ""]
if bocadillo:
    # A la izquierda del bicho y a la altura de la cara, con la comilla
    # apuntando a quien habla.
    _t = c(TENUE) + bocadillo + " " + c(RAYA) + "«" + R
    if vis(_t) <= CW:
        izq[3] = padl(_t, CW)

if tarjeta and all(vis(x) <= CW for x in izq):
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
exec python3 -S -c "$PY_SRC"
