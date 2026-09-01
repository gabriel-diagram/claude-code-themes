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

# `python3 -c` mete el directorio actual en sys.path como "". La statusline
# corre con el cwd puesto en el repo que tengas abierto, asi que sin purgarlo
# un bicho.py cualquiera de ese repo se importaria —y se ejecutaria— una vez
# por refresco, con la excepcion tragada por el try de abajo.
sys.path[:] = [_p for _p in sys.path if _p not in ("", ".", os.getcwd())]
_SL_DIR = os.environ.get("SL_DIR", "")
if _SL_DIR:
    sys.path.insert(0, _SL_DIR)
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
FONDO = (_hx("#0a0d0f"), 233)   # el pie: un tono por encima del negro
BORDE = (_hx("#161c21"), 234)   # la raya fina que lo separa del hilo
DIRT  = (_hx("#4a545f"), 240)   # el "~/" del directorio, mas apagado
COLA  = (_hx("#3a444e"), 238)   # la cola del bocadillo
TEXTO = (_hx("#c9d1d9"), 252)   # lo que dice el bicho

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
# El "~/" se pinta mas apagado que el resto de la ruta. Si la ruta se recorta
# desaparece con lo demas: fingirlo dejaria entender que "srv" cuelga de tu casa
# cuando puede colgar de tres carpetas mas.
dircola = "/".join(partes[-3:]) if len(partes) > 3 else ddir
dirpre = ""
if dircola.startswith("~/"):
    dirpre, dircola = "~/", dircola[2:]
dircorto = dirpre + dircola

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
salida_tok = num(g(d, "context_window", "total_output_tokens"))
api_ms = num(g(d, "cost", "total_api_duration_ms"))
prompt_id = d.get("prompt_id")
transcripcion = d.get("transcript_path")


def modo_permisos(ruta):
    """El modo de permisos no viene en el payload, pero si en el transcript, que
    si llega. Se lee solo la COLA del fichero: es un jsonl de megas y esto corre
    en cada refresco."""
    if not ruta:
        return None
    try:
        tam = os.path.getsize(ruta)
        with open(ruta, "rb") as fh:
            fh.seek(max(0, tam - 32768))
            cola = fh.read().decode("utf-8", "replace")
    except Exception:
        return None
    ult = None
    for ult in re.finditer(r'"permissionMode"\s*:\s*"([A-Za-z]+)"', cola):
        pass
    return ult.group(1) if ult else None


MODOS = {"bypassPermissions": "bypass", "acceptEdits": "auto-edit",
         "plan": "plan", "dangerouslySkipPermissions": "bypass"}
permisos = MODOS.get(modo_permisos(transcripcion) or "")


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


# La identidad del repo la manda el propio payload cuando hay remoto:
# owner/nombre, como en el diseño. Sin remoto no hay owner que enseñar, y la
# carpeta de al lado no lo es: se queda el nombre a secas.
_rjson = g(d, "workspace", "repo")
repo = None
if isinstance(_rjson, dict) and _rjson.get("name"):
    _own = _rjson.get("owner")
    repo = ("%s/%s" % (_own, _rjson["name"])) if _own else str(_rjson["name"])
rama = None
sucio = ""
try:
    p = subprocess.run(["git", "--no-optional-locks", "-C", cwd,
                        "rev-parse", "--show-toplevel", "--abbrev-ref", "HEAD"],
                       capture_output=True, text=True, timeout=1).stdout.split("\n")
    if len(p) >= 2 and p[0].strip():
        raiz = p[0].strip()
        trozos = raiz.rstrip("/").split("/")
        repo = repo or trozos[-1]
        rama = p[1].strip() or None
        est = subprocess.run(["git", "--no-optional-locks", "-C", cwd,
                              "status", "--porcelain"],
                             capture_output=True, text=True, timeout=1).stdout.strip()
        if est:
            sucio = " " + c(NUM) + "✳" + R
except Exception:
    pass

# ---------------------------------------------------------------------------
# SESION — lo que solo se sabe comparando este refresco con el anterior.
# Vive en $TMPDIR/claude-statusline-<session_id>. De aqui salen el ritmo de
# tokens, el pico de contexto y los turnos en bypass, y el hook de SessionEnd
# lo convierte en contadores de comportamiento.
# ---------------------------------------------------------------------------
_sid = d.get("session_id")
_sid = str(_sid) if _sid and re.match(r"^[A-Za-z0-9_-]{1,64}$", str(_sid)) else None
_tmp = os.environ.get("TMPDIR", "/tmp")
_f = os.path.join(_tmp, "claude-statusline-" + _sid) if _sid else None
_hechos = {}
if _f:
    try:
        with open(_f) as _fh:
            _crudo = _fh.read().strip()
        if _crudo.startswith("{"):
            _hechos = json.loads(_crudo)
            if not isinstance(_hechos, dict):
                _hechos = {}
        elif _crudo:
            _hechos = {"etq": _crudo}
    except Exception:
        _hechos = {}

_e = _hechos.get("etq")
_antes = _e if isinstance(_e, str) else None
_pico_ant = num(_hechos.get("pico")) or 0.0
_t0 = num(_hechos.get("t0"))
_ahora = time.time()

# Ritmo de salida. Los dos campos NO miden lo mismo, y ahi esta el truco:
# `total_output_tokens` es lo que sacó la ULTIMA respuesta —- se reinicia en cada
# turno, no es un contador que sube—- y `total_api_duration_ms` es el tiempo de
# API ACUMULADO de la sesion. El ritmo de la ultima respuesta es lo primero
# entre lo que ha crecido lo segundo. Restar dos `total_output_tokens` seguidos
# no mide nada: son dos respuestas distintas.
_api_ant = num(_hechos.get("api"))
_api_t = _api_ant
tps = num(_hechos.get("tps"))
_tps_t = num(_hechos.get("tps_t")) or 0.0
if salida_tok and api_ms is not None:
    if _api_ant is not None and api_ms > _api_ant:
        _dt = (api_ms - _api_ant) / 1000.0
        # Menos de 300 ms de API no da una medida: seria ruido dividido.
        if 0.3 <= _dt <= 600:
            tps = salida_tok / _dt
            _tps_t = _ahora
    _api_t = api_ms
# Cuando el modelo lleva un rato callado el ultimo ritmo ya no dice nada.
if _ahora - _tps_t > 120:
    tps = None

# "30 turnos con permisos en bypass": un turno es un prompt, y el payload trae
# su id. Se cuenta al cambiar de prompt, no en cada refresco.
_turno_nuevo = bool(prompt_id) and _hechos.get("pid") != prompt_id

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
if tps is not None:
    _t = "%.1f" % tps if tps < 100 else "%d" % round(tps)
    b1.append(seg(4, c(NUM) + _t + R + c(TENUE) + " tok/s" + R))
if cache is not None and tps is None:
    # Releva al ritmo, no lo acompaña: el diseño da un hueco de velocidad en la
    # banda, y mientras el modelo habla lo ocupa el tok/s.
    _t = "%d%%" % round(cache * 100)
    b1.append(seg(5, c(NUM) + _t + R + c(TENUE) + " caché" + R))
if permisos:
    # El CLI ya pinta su propio pie de "bypass permissions"; esto es un distintivo
    # en la banda, no una copia de aquella linea, que no es nuestra.
    _pc = c(MAL) if permisos == "bypass" else c(MODO)
    b1.append(seg(3, _pc + B + permisos + R, color=_pc + B, plain=permisos))
if vim:
    vc = c(MODO) if vim == "INSERT" else c(TENUE)
    b1.append(seg(6, vc + B + str(vim) + R, color=vc + B, plain=str(vim)))

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
_dirtxt = (c(DIRT) + dirpre + R if dirpre else "") + c(DIR) + dircola + R
b3 = [seg(1, _dirtxt, color=c(DIR), plain=dircorto)]
_primera = True
for _v, _et in ((rl5, "5h"), (rl7, "7d")):
    if _v is None:
        continue
    # La primera barra va detras del separador; la de 7d, pegada a la de 5h.
    b3.append(seg(2, c(TENUE) + _et + " " + R + barra(_v, 10, CUOTA, VACIO) +
                     c(TENUE) + " %d%%" % round(_v) + R,
                  sep=SEPX if _primera else "  "))
    _primera = False
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
    E = BI.estado_de(uso, ctx=pct)
    etq = E[1]

    pet = BI.leer_pet()
    BI.pasar_hambre(pet)
    criatura, nivel = BI.evolucion(pet)
    hambre = int(pet.get("hambre", 0))
    # La subida de nivel se detecta comparando con el ultimo nivel anunciado,
    # que vive en el propio pet.json. (Antes se miraba una clave que `alimentar`
    # ponia y `leer_pet` filtraba, o sea nunca.)
    subio = nivel > int(pet.get("nivel_visto", 0))

    paso = int(time.time())
    filas = BI.dibujar(criatura, E, paso=paso, hambre=hambre, veinticuatro=_TRUE)

    # Al cruzar un umbral la etiqueta sale en negrita un refresco: para eso hace
    # falta el estado anterior, que ya viene leido de arriba.
    salta = _antes is not None and _antes != etq
    _pico = max(_pico_ant, uso)
    _cambia = (_antes != etq or _pico >= _pico_ant + 1.0 or _turno_nuevo
               or (api_ms is not None and api_ms != _api_ant))
    if _f and _cambia:
        # Cosas que hay que cobrarle al bicho y que ningun hook puede ver, porque
        # los hooks no reciben ni el uso de contexto ni el modo de permisos.
        _pet_cambia = False
        _p2 = None
        try:
            if etq == "k.o." and _antes is not None and _antes != "k.o.":
                _p2 = BI.leer_pet()
                _p2, _ok = BI.alimentar(_p2, "reventon")
                if _ok:
                    BI.contar(_p2, "impulsivo")
                    BI.contar(_p2, "ctx_limite")
                    # sesiones_ctx100 NO se toca aqui: lo cuenta `pet sesion` al
                    # cerrar, que es una vez por sesion.
                    _p2["contadores"]["racha_tests"] = 0
                    _p2["contadores"]["racha_diffs"] = 0
                    _pet_cambia = True
            if _turno_nuevo and permisos == "bypass":
                _p2 = _p2 or BI.leer_pet()
                BI.contar(_p2, "turnos_bypass")
                _pet_cambia = True
            if _pet_cambia and _p2 is not None:
                BI.escribir_pet(_p2)
        except Exception:
            pass
        try:
            with open(_f, "w") as _fh:
                json.dump({"etq": etq, "pico": round(_pico, 1),
                           "t0": int(_t0) if _t0 and _t0 > 1e9 else int(_ahora),
                           "repo": repo or "",
                           "pid": prompt_id or _hechos.get("pid"),
                           "api": _api_t,
                           "tps": round(tps, 2) if tps else None,
                           "tps_t": _tps_t}, _fh)
        except Exception:
            pass
        # De paso, barrer los huerfanos de sesiones muertas (mas de un dia).
        try:
            _limite = _ahora - 86400
            for _n in os.listdir(_tmp):
                if _n.startswith("claude-statusline-"):   # incluye los -todos-
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
        if subio:
            bocadillo = "nivel %d. %s" % (nivel, BI.BONITO.get(criatura, criatura))
            try:
                _p3 = BI.leer_pet()
                _p3["nivel_visto"] = nivel
                BI.escribir_pet(_p3)
            except Exception:
                pass
        elif hambre >= BI.HAMBRE_AVISA:
            bocadillo = "llevo %dh sin comer. /feed" % hambre
        elif etq == "k.o.":
            bocadillo = "100% de contexto. avisé. no pasa nada."
        elif etq == "cansada" and durms and durms > 4 * 3600 * 1000:
            bocadillo = "llevas 4h. yo también estaría espeso"
        elif pet.get("racha", 0) >= 3:
            bocadillo = "racha de %d días. no la rompas hoy" % pet["racha"]

# ---------------------------------------------------------------------------
# SALIDA — el pie, no una banda mas. Fondo un tono por encima del negro y una
# raya fina arriba: pertenece a la ventana, no al hilo. Las tres bandas van en
# la columna izquierda, contra la tarjeta del bicho, en seis filas.
#
#   fila 0   banda 1 · motor      etiqueta del estado
#   fila 1   banda 2 · trabajo    marca de la evolucion
#   fila 2   banda 3 · cuota      cuerpo
#   fila 3                        cara
#   fila 4                        cuerpo
#   fila 5   bocadillo            patas
#
# La raya cuesta una fila. STATUSLINE_REGLA=0 la quita y deja seis.
# ---------------------------------------------------------------------------
def _on(var, defecto="1"):
    return os.environ.get(var, defecto).lower() not in ("0", "off", "no")


FONDO_ON = _on("STATUSLINE_FONDO")
REGLA_ON = _on("STATUSLINE_REGLA")
BG = c(FONDO, False) if FONDO_ON else ""
# Un reset entero se lleva por delante el fondo del pie. Al pintar se cambia por
# "quita la negrita, vuelve al color de texto de la terminal" y se reafirma el
# fondo, que asi sobrevive incluso a la pastilla del modelo, que trae el suyo.
_SUAVE = "\033[22;39m"


def pinta(linea):
    if not FONDO_ON:
        return linea
    hueco = " " * max(0, WIDTH - vis(linea))
    return BG + linea.replace(R, _SUAVE + BG) + hueco + R


salida = []
if REGLA_ON:
    salida.append(c(BORDE) + "─" * WIDTH + R)

izq = None
if tarjeta:
    izq = [assemble(b1, CW), assemble(b2, CW), assemble(b3, CW), "", "", ""]
    if bocadillo:
        # Quien habla lo dice la cara, no una comilla: la fila de los ojos del
        # propio bicho y la cola apuntando al texto.
        _t = filas[2] + c(COLA) + "◗" + R + " " + c(TEXTO) + bocadillo + R
        if vis(_t) <= CW:
            izq[5] = _t
    if any(vis(x) > CW for x in izq):
        izq = None

if izq is not None:
    # Claude Code recorta los espacios del principio de la linea: si la izquierda
    # va vacia, el sangrado desaparece y el bicho se cae al borde. Se ancla con un
    # braille en blanco (U+2800), que se pinta vacio pero no es un espacio.
    for _l, _t in zip(izq, tarjeta):
        if not ANSI.sub("", _l).strip():
            _l = "⠀"
        salida.append(padr(_l, CW) + " " * BICHO_GAP + _t)
else:
    salida.append(assemble(b1, WIDTH))
    salida.append(assemble(b2, WIDTH))
    salida.append(assemble(b3, WIDTH))

print("\n".join(pinta(l) for l in salida))
PYEOF
)
exec python3 -S -c "$PY_SRC"
