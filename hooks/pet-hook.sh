#!/bin/bash
# Hook del bicho: convierte lo que haces en comida y en contadores de habito.
#
# Un solo script para todos los eventos; el evento llega en el JSON de stdin.
# Instalalo con ./install.sh --hooks.
#
#   PostToolUse  Bash      -> commit / tests / repro-antes-del-fix / docs
#   PostToolUse  TodoWrite -> tarea del plan / plan antes de codigo / una herramienta
#   PostToolUse  (resto)   -> solo se anota el nombre, sin arrancar python
#   PreCompact             -> compact
#   SessionEnd             -> forma de la sesion (duracion, pico de contexto, repo)
#
# SOBRE LAS HEURISTICAS. "Tests en verde" no es un hecho que el CLI exponga: hay
# que deducirlo. La regla es preferir perderse una comida a inventarse una.
#   - Un `is_error` del propio CLI manda sobre cualquier patron: es el unico
#     dato duro que hay, y equivale al codigo de salida.
#   - Los patrones de rojo se buscan SOLO en las ultimas lineas, que es donde va
#     el resumen. Buscarlos en toda la salida hacia que un test llamado
#     `test_login_failed` tiñera de rojo una suite verde.
#   - No hace falta un patron de verde: si el comando era un runner y salio con
#     bien, cuenta. Asi entran los runners que no estan en la lista.
#   - PET_TEST_RUNNERS admite un regex extra para lo tuyo.
SL_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)
PET="$SL_DIR/pet"
[ -x "$PET" ] || PET="$(dirname "$SL_DIR")/pet"
[ -x "$PET" ] || exit 0

ENTRADA=$(cat)
saca() { printf '%s' "$ENTRADA" | grep -o "\"$1\"[[:space:]]*:[[:space:]]*\"[^\"]*\"" | head -1 | sed 's/.*"\([^"]*\)"$/\1/'; }
HERR=$(saca tool_name)
SID=$(saca session_id)
EV=$(saca hook_event_name)

# El francotirador necesita saber CUANTAS herramientas distintas se han usado
# entre dos tareas cerradas, o sea todas. Anotar el nombre cuesta un `echo`;
# arrancar python para cada herramienta costaria 15 ms por llamada.
case "$SID" in
  ''|*[!A-Za-z0-9_-]*) SID='' ;;
  *) [ -n "$HERR" ] && echo "$HERR" >> "${TMPDIR:-/tmp}/claude-statusline-herr-$SID" 2>/dev/null ;;
esac

case "$EV:$HERR" in
  PreCompact:*|SessionEnd:*|*:Bash|*:TodoWrite|*:Edit|*:Write|*:NotebookEdit|*:MultiEdit) ;;
  *) exit 0 ;;
esac

PY_SRC=$(cat <<'PYEOF'
import json, os, re, subprocess, sys

PET = os.environ["PET"]
SID = os.environ.get("SID", "")
TMP = os.environ.get("TMPDIR", "/tmp")

try:
    h = json.loads(os.environ["ENTRADA"])
    if not isinstance(h, dict):
        raise ValueError
except Exception:
    raise SystemExit(0)


def pet(*args):
    try:
        subprocess.run([PET] + [str(a) for a in args], capture_output=True, timeout=5)
    except Exception:
        pass


def dar(evento, nota=""):
    pet(evento, nota)
    raise SystemExit(0)


# Estado por sesion. Mismo prefijo que el fichero de la statusline para que lo
# alcance su barrido de huerfanos.
_EST = os.path.join(TMP, "claude-statusline-sesion-" + SID) if SID else None
_HERR = os.path.join(TMP, "claude-statusline-herr-" + SID) if SID else None


def estado():
    try:
        with open(_EST) as fh:
            d = json.load(fh)
        return d if isinstance(d, dict) else {}
    except Exception:
        return {}


def guardar(d):
    if not _EST:
        return
    try:
        with open(_EST, "w") as fh:
            json.dump(d, fh)
    except Exception:
        pass


evento = h.get("hook_event_name") or ""

if evento == "PreCompact":
    dar("compact")

if evento == "SessionEnd":
    if SID:
        pet("sesion", os.path.join(TMP, "claude-statusline-" + SID))
        for f in (_EST, _HERR):
            try:
                os.unlink(f)
            except OSError:
                pass
    raise SystemExit(0)

herr = h.get("tool_name") or ""
ent = h.get("tool_input") or {}
sal = h.get("tool_response") or {}
if not isinstance(ent, dict):
    ent = {}
if not isinstance(sal, dict):
    sal = {}

# --- tocar codigo cierra la puerta al oraculo -------------------------------
if herr in ("Edit", "Write", "NotebookEdit", "MultiEdit"):
    e = estado()
    if not e.get("codigo"):
        e["codigo"] = 1
        guardar(e)
    raise SystemExit(0)

if herr == "TodoWrite":
    todos = ent.get("todos") or []
    todos = [t for t in todos if isinstance(t, dict)]
    hechas = sum(1 for t in todos if t.get("status") == "completed")
    pend = sum(1 for t in todos if t.get("status") == "pending")
    e = estado()

    # oraculo: "5 planes escritos antes de tocar codigo". Un plan es un TodoWrite
    # con varias tareas y ninguna cerrada; solo cuenta si aun no se ha editado.
    if (pend >= 3 and hechas == 0 and not e.get("codigo")
            and not e.get("plan_contado")):
        e["plan_contado"] = 1
        pet("cuenta", "planes_antes_codigo")

    # cartografo: el plan mas largo cerrado del todo.
    if todos and hechas == len(todos):
        pet("record", "plan_entero", len(todos))

    antes = int(e.get("hechas", 0))
    e["hechas"] = hechas
    if hechas > antes:
        # francotirador: "8 tareas cerradas con una sola herramienta". Se miran
        # las herramientas distintas usadas desde la tarea anterior.
        distintas = set()
        if _HERR:
            try:
                with open(_HERR) as fh:
                    distintas = {l.strip() for l in fh if l.strip()} - {"TodoWrite"}
                os.unlink(_HERR)
            except Exception:
                pass
        if len(distintas) == 1:
            pet("cuenta", "tarea_una_herramienta")
        guardar(e)
        titulo = next((t.get("content", "") for t in reversed(todos)
                       if t.get("status") == "completed"), "")
        dar("tarea", str(titulo)[:40])
    guardar(e)
    raise SystemExit(0)

if herr != "Bash":
    raise SystemExit(0)

cmd = str(ent.get("command") or "")
salida = "".join(v for k in ("stdout", "stderr", "output", "content")
                 for v in [sal.get(k)] if isinstance(v, str))
fallo = bool(sal.get("is_error"))

# --- commit -----------------------------------------------------------------
# Anclado al principio del comando o detras de un operador: si no, un
# `grep -rn "git commit"` contaba como commit.
GIT_COMMIT = (r"(?:^|[;&|(]|&&|\|\|)\s*(?:\w+=\S+\s+)*(?:sudo\s+)?"
              r"git\b(?:\s+-[^\s]+(?:\s+\S+)?)*\s+commit\b")
if not fallo and re.search(GIT_COMMIT, cmd) and "--dry-run" not in cmd:
    if not re.search(r"nothing to commit|nada que (?:hacer|confirmar)|"
                     r"no changes added", salida, re.I):
        anchos = re.search(r"(\d+)\s+files?\s+changed", salida)
        if anchos:
            pet("record", "commit_ancho", anchos.group(1))
        # jardinero: "docs y limpieza dos dias seguidos". El commit ya esta
        # hecho, asi que se le pregunta a git en vez de adivinarlo del mensaje.
        cwd = h.get("cwd") or ent.get("cwd") or "."
        try:
            ns = subprocess.run(["git", "--no-optional-locks", "-C", cwd,
                                 "show", "--numstat", "--format=", "HEAD"],
                                capture_output=True, text=True, timeout=3).stdout
            mas = menos = 0
            docs = otros = 0
            for linea in ns.splitlines():
                p = linea.split("\t")
                if len(p) != 3:
                    continue
                try:
                    mas += int(p[0]); menos += int(p[1])
                except ValueError:
                    pass
                f = p[2].lower()
                if (f.endswith((".md", ".rst", ".txt", ".adoc")) or "/docs/" in f
                        or f.startswith("docs/")):
                    docs += 1
                else:
                    otros += 1
            if (docs and docs >= otros) or (menos > mas * 2 and menos > 20):
                pet("dia", "docs")
        except Exception:
            pass
        m = re.search(r"\[([^\]\s]+)", salida)      # "[main a1b2c3] mensaje"
        dar("commit", m.group(1) if m else "")

# --- tests ------------------------------------------------------------------
RUNNERS = (r"\bpytest\b|\bpy\.test\b|\bunittest\b|\btox\b|\bnox\b|\bbehave\b|"
           r"\bnpm\s+(?:run\s+)?test\b|\byarn\s+test\b|\bpnpm\s+(?:run\s+)?test\b|"
           r"\bbun\s+test\b|\bdeno\s+test\b|\bjest\b|\bvitest\b|\bmocha\b|\bava\b|"
           r"\bplaywright\s+test\b|\bcypress\s+run\b|"
           r"\bgo\s+test\b|\bcargo\s+(?:test|nextest)\b|\bctest\b|\bzig\s+test\b|"
           r"\bphpunit\b|\bartisan\s+test\b|\bpest\b|"
           r"\brspec\b|\bcucumber\b|\brake\s+test\b|\bminitest\b|"
           r"\bmvn\s+(?:test|verify)\b|\bgradle\w*\s+\w*test\b|\bsbt\s+test\b|"
           r"\bdotnet\s+test\b|\bswift\s+test\b|\bflutter\s+test\b|\bmix\s+test\b|"
           r"\b(?:make|just|task)\s+\w*test\w*\b|\bnix\s+flake\s+check\b")
_extra = os.environ.get("PET_TEST_RUNNERS", "").strip()
if _extra:
    try:
        re.compile(_extra)
        RUNNERS += "|" + _extra
    except re.error:
        pass
def parece_runner(orden):
    """Ultimo recurso para los runners que no estan en la lista: que el propio
    ejecutable se llame algo con "test" o "spec" (run-tests.sh, testear.sh,
    bin/spec). Se excluye el `test` de shell, que es una comparacion de ficheros
    y no una suite."""
    trozo = orden.strip().split()
    if not trozo:
        return False
    base = os.path.basename(trozo[0]).lower()
    base = re.sub(r"\.(sh|bash|zsh|py|rb|js|ts|pl)$", "", base)
    if base in ("test", "[", "[["):
        return False
    return bool(re.search(r"test|spec", base))


if re.search(RUNNERS, cmd, re.I) or parece_runner(cmd):
    # El rojo se busca solo en la cola, que es donde va el resumen.
    cola = "\n".join(salida.splitlines()[-12:])
    ROJO = (r"\bFAILED\b|\bFAILURES\b|\bfailures?[:=]\s*[1-9]|\berrors?[:=]\s*[1-9]|"
            r"\b[1-9]\d*\s+(?:failed|failing|errors?)\b|\bpanic:|"
            r"\bTests?\s+failed\b|\bFAIL\b")
    rojo = fallo or bool(re.search(ROJO, cola, re.I))
    e = estado()
    if rojo:
        # sabueso: "repro antes del fix". Se recuerda que hubo un rojo; cuando
        # despues venga un verde, eso es un ciclo reproducir -> arreglar.
        if not e.get("rojo"):
            e["rojo"] = 1
            guardar(e)
    else:
        if e.get("rojo"):
            e["rojo"] = 0
            guardar(e)
            pet("cuenta", "repro_antes_fix")
        dar("tests")
PYEOF
)
export PET SID ENTRADA
exec python3 -S -c "$PY_SRC"
