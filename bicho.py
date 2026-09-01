# -*- coding: utf-8 -*-
"""El bicho: plantillas, estados, evoluciones y el fichero de vida.

Lo comparten la statusline y `pet`. Va en un modulo y no incrustado en el .sh
por dos razones: una sola copia de las 27 plantillas, y un modulo importado si
usa cache de bytecode (un `python3 -c` recompila su fuente en cada refresco).

Dos capas que no se mezclan (ver seccion 05 del diseno):

  vida      del momento. Sale del uso de contexto y cuota. Sube y baja.
  progreso  permanente. Sale de la XP acumulada. El nivel nunca baja.

La vida elige ojos, patas y color. El progreso elige la silueta.
"""

import json
import os
import time

# ---------------------------------------------------------------------------
# PALETA — un color por tipo de dato. (rgb, caida a 256)
# ---------------------------------------------------------------------------
def _hx(s):
    return (int(s[1:3], 16), int(s[3:5], 16), int(s[5:7], 16))

RUTA  = (_hx("#4dd6c1"),  79)
IDENT = (_hx("#57e389"),  78)
LINK  = (_hx("#6fb6ff"),  75)
NUM   = (_hx("#e8c46a"), 179)
MODO  = (_hx("#b07cf0"), 141)
MAL   = (_hx("#f2777a"), 210)
ENF   = (_hx("#eceff4"), 255)
TENUE = (_hx("#6b7683"), 243)
RAYA  = (_hx("#2c343c"), 236)
DIR   = (_hx("#8d99a6"), 246)
CUOTA = (_hx("#4ea3f5"),  75)
VACIO = (_hx("#1d2b38"), 235)
CTXV  = (_hx("#24382c"), 235)

PALE_V = (_hx("#d8ffe9"), 194); OSC_V = (_hx("#2f7a52"),  29)
PALE_T = (_hx("#d6fffa"), 195); OSC_T = (_hx("#2b7d74"),  30)
PALE_A = (_hx("#d5e9ff"), 189); OSC_A = (_hx("#2a5c8a"),  24)
PALE_I = (_hx("#cdd2ff"), 189); OSC_I = (_hx("#333c9e"),  61)
GRIS   = (_hx("#5a6270"), 242)
PALE_G = (_hx("#8891a0"), 245); OSC_G = (_hx("#3d434d"), 238)

R = "\033[0m"
B = "\033[1m"


def color(par, fg=True, veinticuatro=True):
    """Secuencia ANSI. 24 bits si el terminal lo admite, si no el 256 mas cercano."""
    lead = 38 if fg else 48
    if veinticuatro:
        return "\033[%d;2;%d;%d;%dm" % ((lead,) + par[0])
    return "\033[%d;5;%dm" % (lead, par[1])


# ---------------------------------------------------------------------------
# VIDA — los siete estados. Un solo numero (0-100) los elige.
# ---------------------------------------------------------------------------
# tope, etiqueta, color, ojo_claro, ojo_oscuro, ojos, cabeza_alta, anda, chispa
ESTADOS = [
    (22,     "fresca",   IDENT, PALE_V, OSC_V, (">", "<"), True,  True,  True),
    (45,     "vibrante", IDENT, PALE_V, OSC_V, (">", "<"), True,  True,  False),
    (63,     "a gusto",  RUTA,  PALE_T, OSC_T, ("o", "o"),       True,  True,  False),
    (78,     "espesa",   CUOTA, PALE_A, OSC_A, ("▬", "▬"), True, False, False),
    (89,     "cansada",  CUOTA, PALE_A, OSC_A, ("_", "_"),       False, False, False),
    (99.999, "ahogada",  (_hx("#5865f2"), 63), PALE_I, OSC_I, ("x", "x"), False, False, False),
    (100,    "k.o.",     GRIS,  PALE_G, OSC_G, ("x", "x"),       False, False, False),
]

# nombre: (marca, cuerpo_sup, cara, cuerpo_inf, patas, ojos_propios, (ojo_izq, ojo_der))
# Las 5 filas son de 9 columnas clavadas. La cara lleva los huecos de ojos
# en blanco y los indices dicen donde van.
PLANTILLAS = {
    # --- nivel 1 · larva
    'chispa'       : ('         ', '  ▗▄▄▖   ', '  ▐    ▌ ', '  ▝▀▀▘   ', '   ▘▝    ', '..', (4, 6)),
    # --- nivel 2 · temperamentos
    'pauta'        : ('  |   |  ', ' ▗█▀█▀█▖ ', '▐█     █▌', ' ▝█▄█▄█▘ ', '  ▘   ▝  ', '><', (3, 5)),
    'sonda'        : ('  \\ o    ', ' ▗█████▖ ', '▐█     █▌', ' ▝█████▘ ', '  ▘   ▝  ', 'OO', (3, 5)),
    'brasa'        : ('  /   /  ', '▗██████▖ ', '█     ██▌', ' ▝████▘  ', ' ▘ ▘  ▝  ', '>>', (2, 4)),
    # --- niveles 3-4 · oficios
    'refactor'     : ('  |   |  ', ' ▗█┼█┼█▖ ', '▐█     █▌', ' ▝█┼█┼█▘ ', ' ▘▘   ▝▝ ', '><', (3, 5)),
    'pulcro'       : ('  \\ * /  ', ' ▗█████▖ ', '▐█     █▌', ' ▝█████▘ ', '  ▘   ▝  ', '^^', (3, 5)),
    'cazabugs'     : (' \\\\   // ', ' ▗█████▖ ', '▐█     █▌', ' ▝█▀▀▀█▘ ', ' \\▘   ▝/ ', '\\/', (3, 5)),
    'arquitecto'   : ('  ╔═══╗  ', ' ▗█████▖ ', '▐█     █▌', ' ▝█████▘ ', '  ▘   ▝  ', 'OO', (3, 5)),
    'velocista'    : ('   / /   ', '=▗█████▖ ', '=█     █▌', ' ▝████▘  ', '   ▘▘    ', '>>', (3, 5)),
    'maraton'      : ('  \\   /  ', '▗███████▖', '██     ██', '▝███████▘', ' ▘ ▘ ▝ ▝ ', '▬▬', (3, 5)),
    'salvaje'      : (' ^ \\ / ^ ', ' ▗█▀█▀█▖ ', '▐█     █▌', ' ▝█▀█▀█▘ ', ' ▘  ▝  ▝ ', 'xo', (3, 5)),
    # --- nivel 5 · marcas y secretas
    'cirujano'     : ('   -+-   ', ' ▗█┼█┼█▖ ', '▐█     █▌', ' ▝█┼█┼█▘ ', ' ▘▘   ▝▝ ', '><', (3, 5)),
    'tejedor'      : (' \\ | | / ', ' ▗█┼█┼█▖ ', '▐█     █▌', ' ▝█┼█┼█▘ ', ' ▘▘   ▝▝ ', '><', (3, 5)),
    'monje'        : ('  ▗▄▄▄▖  ', ' ▗█████▖ ', '▐█     █▌', ' ▝█████▘ ', '  ▘   ▝  ', '><', (3, 5)),
    'jardinero'    : ('  * . *  ', ' ▗█████▖ ', '▐█     █▌', ' ▝█████▘ ', '  ▘   ▝  ', '><', (3, 5)),
    'sabueso'      : (' \\\\   // ', ' ▗█████▖ ', '▐█     █▌', ' ▝█▀▀▀█▘ ', ' ▘ ▘ ▝ ▝ ', '><', (3, 5)),
    'exterminador' : ('  ^^ ^^  ', ' ▗█████▖ ', '▐█     █▌', ' ▝█▀▀▀█▘ ', ' ▘▘   ▝▝ ', '><', (3, 5)),
    'cartografo'   : ('  ╔═══╗  ', ' ▗█████▖ ', '▐█     █▌', ' ▝█████▘ ', '  ▘   ▝  ', '><', (3, 5)),
    'oraculo'      : ('  o - o  ', ' ▗█████▖ ', '▐█     █▌', ' ▝█████▘ ', '  ▘   ▝  ', '><', (3, 5)),
    'relampago'    : ('  / / /  ', '▗██████▖ ', '▐█     █▌', ' ▝████▘  ', '   ▘▘    ', '><', (3, 5)),
    'francotirador': ('  =[+]=  ', '▗██████▖ ', '▐█     █▌', ' ▝████▘  ', '   ▘▘    ', '><', (3, 5)),
    'buey'         : (' \\_   _/ ', '▗███████▖', '▐█     █▌', '▝███████▘', ' ▘ ▘ ▝ ▝ ', '><', (3, 5)),
    'topo'         : ('  \\\\ //  ', '▗███████▖', '▐█     █▌', '▝███████▘', ' ▘ ▘ ▝ ▝ ', '><', (3, 5)),
    'gremlin'      : (' ^ \\ / ^ ', ' ▗█▀█▀█▖ ', '▐█     █▌', ' ▝█▀█▀█▘ ', ' ▘  ▝  ▝ ', '><', (3, 5)),
    'kraken'       : (' ~ ~ ~ ~ ', ' ▗█▀█▀█▖ ', '▐█     █▌', ' ▝█▀█▀█▘ ', ' ▘  ▝  ▝ ', '><', (3, 5)),
    'fenix'        : ('  \\ * /  ', ' ▗█████▖>', '▐█     █▌', ' ▝█████▘ ', '  ▘   ▝  ', '##', (3, 5)),
    'quimera'      : (' \\ \\ / / ', ' ▗█████▖ ', '▐█     █▌', ' ▝█▀█▀█▘ ', ' ▘▘   ▝▝ ', 'O#', (3, 5)),
}

# ---------------------------------------------------------------------------
# PROGRESO — el arbol. La forma la da la rama; el estado, los ojos.
# ---------------------------------------------------------------------------
# Cada evolucion es una plantilla de 5 filas con dos huecos de ojos y una fila
# de patas. El estado de vida no cambia la silueta: rellena los huecos y elige
# el color. Asi las 27 evoluciones tienen sus 7 estados sin dibujar 189 sprites.

ARBOL = {
    "chispa":     ["pauta", "sonda", "brasa"],
    "pauta":      ["refactor", "pulcro"],
    "sonda":      ["cazabugs", "arquitecto"],
    "brasa":      ["velocista", "maraton", "salvaje"],
    "refactor":   ["cirujano", "tejedor"],
    "pulcro":     ["monje", "jardinero"],
    "cazabugs":   ["sabueso", "exterminador"],
    "arquitecto": ["cartografo", "oraculo"],
    "velocista":  ["relampago", "francotirador"],
    "maraton":    ["buey", "topo"],
    "salvaje":    ["gremlin", "kraken"],
}

# Nombre con tilde para pintar, sin tilde para las claves del arbol y del json.
BONITO = {"maraton": "maratón", "cartografo": "cartógrafo", "oraculo": "oráculo",
          "relampago": "relámpago", "fenix": "fénix"}

PADRE = {h: p for p, hs in ARBOL.items() for h in hs}

# XP acumulada -> nivel. El nivel nunca baja.
NIVELES = [(0, 1), (60, 2), (180, 3), (400, 4), (900, 5)]

# El contador que decide cada bifurcacion. Gana el mas alto al subir de nivel.
DECIDE = {
    "pauta": "metodico", "sonda": "inquisitivo", "brasa": "impulsivo",
    "refactor": "diffs", "pulcro": "ctx_bajo",
    "cazabugs": "tests", "arquitecto": "planes",
    "velocista": "sesiones_cortas", "maraton": "sesiones_largas", "salvaje": "ctx_limite",
}

# Nivel 5: cada marca pide una condicion de habito, no XP. (contador, umbral)
CONDICION = {
    "cirujano":      ("racha_diffs", 20),
    "tejedor":       ("commit_ancho", 10),
    "monje":         ("sesiones_bajo_40", 5),
    "jardinero":     ("dias_docs", 2),
    "sabueso":       ("repro_antes_fix", 10),
    "exterminador":  ("racha_tests", 15),
    "cartografo":    ("plan_entero", 10),
    "oraculo":       ("planes_antes_codigo", 5),
    "relampago":     ("sesiones_15min", 10),
    "francotirador": ("tarea_una_herramienta", 8),
    "buey":          ("sesiones_4h", 3),
    "topo":          ("dias_mismo_repo", 5),
    "gremlin":       ("turnos_bypass", 30),
    "kraken":        ("sesiones_ctx100", 3),
}

# La comida: (xp, hambre, tope diario o None)
COMIDA = {
    "tests":   ( 15, -4, None),
    "commit":  ( 12, -3, None),
    "compact": (  8, -3, None),
    "tarea":   (  6, -1, None),
    "feed":    (  3, -2, 4),
    "reventon": (-15, 0, None),
}
COMIDA_ETQ = {
    "tests": "tests en verde", "commit": "commit", "compact": "compact",
    "tarea": "tarea del plan", "feed": "/feed", "reventon": "contexto al 100%",
}

HAMBRE_MAX = 10
HAMBRE_AVISA = 7          # a partir de aqui los ojos se apagan y pide comida


def nivel_de(xp):
    n = 1
    for tope, niv in NIVELES:
        if xp >= tope:
            n = niv
    return n


def _mejor(pet, candidatos):
    """De varios hermanos gana el contador mas alto; empate -> orden del diseno."""
    c = pet.get("contadores", {})
    return max(candidatos, key=lambda h: (c.get(DECIDE.get(h, ""), 0), -candidatos.index(h)))


def evolucion(pet):
    """Camina el arbol desde chispa hasta donde llegue la XP y los habitos.

    Devuelve (clave, nivel_alcanzado). La rama no se elige: en cada bifurcacion
    gana el contador de comportamiento que va mas alto en ese momento.
    """
    secreta = pet.get("secreta")
    if secreta in PLANTILLAS:
        return secreta, 5
    xp = pet.get("xp", 0)
    niv = nivel_de(xp)
    aqui = "chispa"
    if niv >= 2:
        aqui = _mejor(pet, ARBOL["chispa"])
    if niv >= 3:
        aqui = _mejor(pet, ARBOL[aqui])
    if niv >= 5:
        c = pet.get("contadores", {})
        listas = [h for h in ARBOL.get(aqui, [])
                  if c.get(CONDICION[h][0], 0) >= CONDICION[h][1]]
        if listas:
            aqui = listas[0]
    return aqui, niv


# ---------------------------------------------------------------------------
# EL FICHERO DE VIDA — ~/.claude/pet.json
# ---------------------------------------------------------------------------
def ruta_pet():
    return os.path.join(os.path.expanduser("~"), ".claude", "pet.json")


# clave -> como se normaliza al leer. Un pet.json lo puede editar cualquiera
# (o corromperlo un disco lleno), asi que NADA se cree sin pasar por aqui.
_ENTERO = "entero"
_TEXTO = "texto"
CAMPOS = {
    "xp": _ENTERO, "hambre": _ENTERO, "comio": _ENTERO, "racha": _ENTERO,
    "mejor_racha": _ENTERO, "nivel_visto": _ENTERO, "hambre_tope": _ENTERO,
    "feed_hoy": _ENTERO,
    "ultimo_dia": _TEXTO, "repo_dia": _TEXTO, "hoy_dia": _TEXTO,
}


def nuevo_pet():
    """Un bicho recien nacido. Funcion y no constante: si fuera un dict de modulo
    todos los pets compartirian sus contadores."""
    d = {k: (0 if t is _ENTERO else "") for k, t in CAMPOS.items()}
    d.update({"secreta": None, "contadores": {}, "hoy": []})
    return d


PET_VACIO = nuevo_pet()      # solo para consultar las claves; no se muta


def _entero(v, por_defecto=0):
    if isinstance(v, bool) or v is None:
        return por_defecto
    try:
        f = float(v)
    except (TypeError, ValueError):
        return por_defecto
    return por_defecto if f != f or f in (float("inf"), float("-inf")) else int(f)


def leer_pet(ruta=None):
    """Nunca lanza, y nunca devuelve un campo con un tipo que no sea el suyo.

    Un pet.json corrupto es un bicho recien nacido, no un crash: si esto deja
    pasar un texto donde va un numero, la statusline entera desaparece.
    """
    pet = nuevo_pet()
    try:
        with open(ruta or ruta_pet(), encoding="utf-8") as fh:
            d = json.load(fh)
        if not isinstance(d, dict):
            return pet
    except Exception:
        return pet
    for k, t in CAMPOS.items():
        if k in d:
            pet[k] = _entero(d[k]) if t is _ENTERO else (
                d[k] if isinstance(d[k], str) else "")
    if isinstance(d.get("secreta"), str) and d["secreta"] in PLANTILLAS:
        pet["secreta"] = d["secreta"]
    if isinstance(d.get("contadores"), dict):
        pet["contadores"] = {k: _entero(v) for k, v in d["contadores"].items()
                             if isinstance(k, str)}
    if isinstance(d.get("hoy"), list):
        pet["hoy"] = [e for e in d["hoy"] if isinstance(e, dict)][-40:]
    return pet


def escribir_pet(pet, ruta=None):
    """Escritura atomica: fichero temporal en el mismo directorio y rename.

    Varias sesiones comparten este fichero. El rename es atomico en POSIX, asi
    que nadie lee un json a medias; lo que si puede perderse es una escritura
    concurrente, por eso `alimentar` hace leer-modificar-escribir de una pieza.
    """
    import tempfile          # solo hace falta al escribir, y cuesta 2 ms de import:
                              # la statusline lee este fichero en cada refresco y
                              # no lo escribe nunca. Ver AUDITORIA.md.
    ruta = ruta or ruta_pet()
    d = os.path.dirname(ruta)
    try:
        os.makedirs(d, exist_ok=True)
        fd, tmp = tempfile.mkstemp(dir=d, prefix=".pet-", suffix=".json")
        try:
            with os.fdopen(fd, "w", encoding="utf-8") as fh:
                json.dump(pet, fh, ensure_ascii=False, indent=1)
            os.replace(tmp, ruta)
        except Exception:
            try:
                os.unlink(tmp)
            except OSError:
                pass
            raise
        return True
    except Exception:
        return False


def _hoy():
    return time.strftime("%Y-%m-%d")


def pasar_hambre(pet, ahora=None):
    """+1 de hambre por hora sin comer, hasta 10. No mata: solo apaga los ojos."""
    ahora = ahora or time.time()
    comio = pet.get("comio") or 0
    if not comio:
        return pet
    horas = int((ahora - comio) // 3600)
    if horas > 0:
        pet["hambre"] = min(HAMBRE_MAX, int(pet.get("hambre", 0)) + horas)
        pet["comio"] = comio + horas * 3600
    # El pico de hambre se anota AQUI, que es donde sube. Anotarlo solo al comer
    # se pierde justo el momento que le interesa al fenix.
    pet["hambre_tope"] = max(int(pet.get("hambre_tope", 0)), int(pet.get("hambre", 0)))
    return pet


def alimentar(pet, evento, nota="", ahora=None):
    """Aplica una comida. Devuelve (pet, aplicado)."""
    if evento not in COMIDA:
        return pet, False
    ahora = ahora if ahora is not None else time.time()
    xp, dh, tope = COMIDA[evento]
    # El dia sale de `ahora` y no del reloj: si no, pasarle otra hora deja la
    # racha en un estado incoherente y no hay forma de probarla.
    dia = time.strftime("%Y-%m-%d", time.localtime(ahora))
    if pet.get("hoy_dia") != dia:
        pet["hoy_dia"] = dia
        pet["hoy"] = []
        pet["feed_hoy"] = 0
    # El tope diario lleva su propio contador: `hoy` es un registro acotado a 40
    # entradas, asi que contar sobre el se salta el tope en cuanto rota.
    if tope is not None and int(pet.get("feed_hoy", 0)) >= tope:
        return pet, False
    pet["xp"] = max(0, int(pet.get("xp", 0)) + xp)
    if dh:
        pet["hambre"] = max(0, min(HAMBRE_MAX, int(pet.get("hambre", 0)) + dh))
    if xp > 0:
        pet["comio"] = int(ahora)
    if tope is not None:
        pet["feed_hoy"] = int(pet.get("feed_hoy", 0)) + 1
    # racha de dias
    ayer = time.strftime("%Y-%m-%d", time.localtime(ahora - 86400))
    ult = pet.get("ultimo_dia") or ""
    if ult != dia:
        pet["racha"] = int(pet.get("racha", 0)) + 1 if ult == ayer else 1
        pet["mejor_racha"] = max(int(pet.get("mejor_racha", 0)), pet["racha"])
        pet["ultimo_dia"] = dia
    pet["hoy"].append({"q": evento, "xp": xp, "t": int(ahora), "n": str(nota)[:40]})
    pet["hoy"] = pet["hoy"][-40:]
    secretas(pet)
    return pet, True


def secretas(pet):
    """Las dos de nivel 5 que no salen del arbol. Se comprueban al comer, que es
    lo unico que mueve hambre y XP."""
    if pet.get("secreta"):
        return pet
    hambre = int(pet.get("hambre", 0))
    pet["hambre_tope"] = max(int(pet.get("hambre_tope", 0)), hambre)

    # fenix: tocar hambre 10 y volver a 0 en la MISMA sesion, y solo desde las
    # dos evoluciones que se dejan llegar ahi. `hambre_tope` lo pone a cero el
    # cierre de sesion.
    if (hambre == 0 and int(pet.get("hambre_tope", 0)) >= HAMBRE_MAX
            and evolucion(pet)[0] in ("salvaje", "maraton")):
        pet["secreta"] = "fenix"
        return pet

    # quimera: dos temperamentos empatados al llegar a nivel 4. Hereda ojos de
    # uno y cuerpo del otro, que es lo que dibuja su plantilla.
    if nivel_de(pet.get("xp", 0)) >= 4:
        c = pet.get("contadores", {})
        temps = sorted((int(c.get(t, 0)) for t in ("metodico", "inquisitivo", "impulsivo")),
                       reverse=True)
        if temps[0] > 0 and temps[0] == temps[1]:
            pet["secreta"] = "quimera"
    return pet


def contar(pet, contador, n=1):
    """Suma. Para los contadores que cuentan veces."""
    c = pet.setdefault("contadores", {})
    c[contador] = int(c.get(contador, 0)) + n
    return pet


def marcar(pet, contador, valor):
    """Guarda el maximo. Para los contadores que miden un record, no un total:
    "un refactor que toca 10+ ficheros" es el mas ancho, no la suma de todos."""
    c = pet.setdefault("contadores", {})
    c[contador] = max(int(c.get(contador, 0)), int(valor))
    return pet

# ---------------------------------------------------------------------------
# DIBUJO — plantilla x estado. Ni un sprite escrito dos veces.
# ---------------------------------------------------------------------------
ANCHO = 9

_RELLENO = "█┼▀▄"          # lo que forma el cuerpo y se puede hundir
_TUMBADO = "▝▘█┼▄▀"        # lo que se aplasta al caer k.o.
_PASO = {"▘": "▝", "▝": "▘"}
_QUIETO = {"▘": "▖", "▝": "▗"}


def _mapear(fila, tabla):
    return "".join(tabla.get(ch, ch) for ch in fila)


def _hundir(sup):
    """La cabeza se hunde: el relleno pasa a media altura, las esquinas quedan."""
    return "".join("▄" if ch in _RELLENO else ch for ch in sup)


def _aplastar(inf):
    """K.o.: la silueta se tumba y solo queda la linea de arriba del bloque."""
    return "".join("▀" if ch in _TUMBADO else ch for ch in inf)


def estado_de(uso):
    """El estado que corresponde a un uso de 0 a 100."""
    uso = max(0.0, min(100.0, float(uso)))
    for e in ESTADOS:
        if uso <= e[0]:
            return e
    return ESTADOS[-1]


def dibujar(criatura, est, paso=0, hambre=0, veinticuatro=True):
    """Devuelve las cinco filas ya coloreadas, de ANCHO columnas visibles cada una.

    `est` es una tupla de ESTADOS, `paso` un contador que avanza en cada
    refresco (de ahi salen el paso de las patas y el parpadeo).
    """
    pl = PLANTILLAS.get(criatura) or PLANTILLAS["chispa"]
    marca, sup, cara, inf, patas, propios, (ia, ib) = pl
    _, etq, col, claro, oscuro, ojos_est, cabeza, anda, _chispa = est

    # Los ojos: la evolucion pone los suyos mientras esta entera (fresca y
    # vibrante); de "a gusto" para abajo manda el estado, que es lo que deja
    # leer el cansancio de un vistazo sin mirar la etiqueta.
    ojos = tuple(propios) if etq in ("fresca", "vibrante") else ojos_est
    # Parpadeo: un frame de cada siete, solo mientras anda.
    if anda and paso % 7 == 3:
        ojos = ("_", "_")

    # A 1 fps unas patas alternando sin parar en la esquina del ojo cansan:
    # STATUSLINE_BICHO_CALMA=1 lo deja andando 4 segundos de cada 12.
    if anda and os.environ.get("STATUSLINE_BICHO_CALMA", "").lower() in ("1", "on", "yes"):
        anda = paso % 12 < 4

    ko = etq == "k.o."
    if ko:
        f_patas = _mapear(patas, _QUIETO)
        f1, f5 = f_patas, " " * ANCHO
        f2, f4 = _hundir(sup), _aplastar(inf)
    else:
        f1 = marca if cabeza else " " * ANCHO
        f2 = sup if cabeza else _hundir(sup)
        f4 = inf
        if anda:
            f5 = patas if paso % 2 == 0 else _mapear(patas, _PASO)
        else:
            f5 = _mapear(patas, _QUIETO)

    C = color(col, veinticuatro=veinticuatro)
    P = color(oscuro, veinticuatro=veinticuatro)
    # Con hambre los ojos se apagan: mismo glifo, color hundido.
    O = color(oscuro if hambre >= HAMBRE_AVISA else claro, veinticuatro=veinticuatro)

    cuerpo = list(cara)
    cuerpo[ia] = O + ojos[0] + C
    cuerpo[ib] = O + ojos[1] + C
    f3 = C + "".join(cuerpo) + R

    return [
        (P if ko else C) + f1 + R,
        C + f2 + R,
        f3,
        C + f4 + R,
        P + f5 + R,
    ]


def uso_ponderado(ctx, cinco_h, siete_d):
    """Media ponderada 50/30/20. Si falta un dato su peso va a los demas."""
    num = den = 0.0
    for v, w in ((ctx, 0.5), (cinco_h, 0.3), (siete_d, 0.2)):
        try:
            if v is not None:
                num += float(v) * w
                den += w
        except (TypeError, ValueError):
            pass
    return max(0.0, min(100.0, num / den)) if den else 0.0
