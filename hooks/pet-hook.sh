#!/bin/bash
# Hook del bicho: convierte lo que haces en comida.
#
# Un solo script para todos los eventos; el evento llega en el JSON de stdin.
# Instalalo con ./install.sh, que lo engancha en ~/.claude/settings.json.
#
#   PostToolUse  Bash      -> commit / tests
#   PostToolUse  TodoWrite -> tarea del plan
#   PreCompact             -> compact
#   SessionEnd             -> forma de la sesion (duracion, pico de contexto, repo)
#
# HONESTIDAD: "tests en verde" y "commit hecho" son HEURISTICAS. El hook ve el
# comando y su salida, no un resultado estructurado, asi que detectar que una
# suite ha pasado es reconocer patrones sobre texto arbitrario. Falla en los dos
# sentidos: un `pytest` que imprime "failed" en un nombre de test cuenta como
# rojo, y un runner exotico no cuenta como nada. Se puede afinar la lista de
# abajo. No hay forma de hacerlo exacto sin que el CLI exponga el resultado.
SL_DIR=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)
PET="$SL_DIR/pet"
[ -x "$PET" ] || PET="$(dirname "$SL_DIR")/pet"
[ -x "$PET" ] || exit 0

# El fuente va por -c y no por stdin: stdin lo ocupa el JSON del hook.
PY_SRC=$(cat <<'PYEOF'
import json, os, re, subprocess, sys

PET = os.environ["PET"]

try:
    h = json.load(sys.stdin)
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


evento = h.get("hook_event_name") or ""

if evento == "PreCompact":
    dar("compact")

if evento == "SessionEnd":
    # Los hechos que solo la statusline ve los dejo ella en este fichero.
    sid = str(h.get("session_id") or "")
    if re.match(r"^[A-Za-z0-9_-]{1,64}$", sid):
        pet("sesion", os.path.join(os.environ.get("TMPDIR", "/tmp"),
                                   "claude-statusline-" + sid))
    raise SystemExit(0)

herr = h.get("tool_name") or ""
ent = h.get("tool_input") or {}
sal = h.get("tool_response") or {}

if herr == "TodoWrite":
    # Una tarea que pasa a completada es una comida. Solo cuenta el salto, asi
    # que se compara contra el recuento anterior guardado por sesion.
    todos = ent.get("todos") or []
    hechas = sum(1 for t in todos if (t or {}).get("status") == "completed")
    # Mismo prefijo que el fichero de la statusline para que lo alcance su
    # barrido de huerfanos, y mismo filtro de session_id: sin el, un id con
    # ".." escribe fuera de TMPDIR.
    _sid = str(h.get("session_id") or "")
    if not re.match(r"^[A-Za-z0-9_-]{1,64}$", _sid):
        raise SystemExit(0)
    marca = os.path.join(os.environ.get("TMPDIR", "/tmp"),
                         "claude-statusline-todos-" + _sid)
    antes = 0
    try:
        with open(marca) as fh:
            antes = int(fh.read().strip() or 0)
    except Exception:
        pass
    if hechas != antes:
        try:
            with open(marca, "w") as fh:
                fh.write(str(hechas))
        except Exception:
            pass
    # "un plan de 10 tareas cerrado entero": el record es el plan mas largo
    # que se ha terminado del todo.
    if todos and hechas == len(todos):
        pet("record", "plan_entero", len(todos))
    if hechas > antes:
        titulo = next((t.get("content", "") for t in reversed(todos)
                       if (t or {}).get("status") == "completed"), "")
        dar("tarea", titulo[:40])
    raise SystemExit(0)

if herr != "Bash":
    raise SystemExit(0)

cmd = str(ent.get("command") or "")
salida = ""
for k in ("stdout", "stderr", "output", "content"):
    v = sal.get(k) if isinstance(sal, dict) else None
    if isinstance(v, str):
        salida += v
if isinstance(sal, dict) and sal.get("is_error"):
    raise SystemExit(0)

# --- commit -----------------------------------------------------------------
# Anclado al principio del comando o justo detras de un operador de shell: si no,
# un `grep -rn "git commit"` cuenta como commit.
GIT_COMMIT = (r"(?:^|[;&|(]|&&|\|\|)\s*(?:\w+=\S+\s+)*(?:sudo\s+)?"
              r"git\b(?:\s+-[^\s]+(?:\s+\S+)?)*\s+commit\b")
if re.search(GIT_COMMIT, cmd) and "--dry-run" not in cmd:
    if not re.search(r"nothing to commit|nada que (?:hacer|"
                     r"confirmar)|no changes added", salida, re.I):
        # "N files changed" da el ancho del commit, que es lo que mira tejedor.
        anchos = re.search(r"(\d+)\s+files?\s+changed", salida)
        if anchos:
            pet("record", "commit_ancho", anchos.group(1))
        m = re.search(r"\[([^\]\s]+)", salida)      # "[main a1b2c3] mensaje"
        dar("commit", m.group(1) if m else "")

# --- tests ------------------------------------------------------------------
RUNNERS = (r"\bpytest\b|\bpy\.test\b|\bunittest\b|"
           r"\bnpm\s+(?:run\s+)?test\b|\byarn\s+test\b|\bpnpm\s+(?:run\s+)?test\b|"
           r"\bjest\b|\bvitest\b|\bmocha\b|\bplaywright\s+test\b|"
           r"\bgo\s+test\b|\bcargo\s+test\b|\bctest\b|"
           r"\bphpunit\b|\bartisan\s+test\b|\bpest\b|"
           r"\brspec\b|\brake\s+test\b|\bmvn\s+(?:test|verify)\b|\bgradle\s+test\b")
ROJO = (r"\bFAILED\b|\bFAIL\b|\bfailures?[:=]\s*[1-9]|\berrors?[:=]\s*[1-9]|"
        r"\b[1-9]\d*\s+failed\b|\bAssertionError\b|\bpanic:|\bFAILURES\b")
if re.search(RUNNERS, cmd, re.I) and not re.search(ROJO, salida):
    # El `ok` suelto valia para `go test`, pero bajo re.I casa con cualquier
    # texto que diga "ok". Se limita a una linea que EMPIEZA por ok, que es la
    # forma real de `go test`.
    VERDE = (r"(?m)^ok\s|\bpass(?:ed|ing)\b|\b0 failures?\b|\bAll tests pass|"
             r"\bTests?:\s+\d+\s+passed|\bPASS\b|\bOK\b\s*\(\d+ tests?\)")
    if re.search(VERDE, salida):
        dar("tests")
PYEOF
)
export PET
exec python3 -S -c "$PY_SRC"
