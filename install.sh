#!/bin/bash
# Instalador del tema Terminal y del bicho.
#
#   ./install.sh              tema + statusline + panel /pet   (no toca settings.json)
#   ./install.sh --hooks      lo anterior y ademas engancha los hooks de comida
#   ./install.sh --uninstall  deshace los hooks y la statusline
#
# Los hooks van aparte a proposito: viven en ~/.claude/settings.json, que es
# global, asi que corren en TODOS tus repos. Sin ellos el bicho existe y se ve,
# pero no come solo: se le da con `pet feed`.
set -euo pipefail

ORIGEN=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)
DESTINO="${CLAUDE_DIR:-$HOME/.claude}"
AJUSTES="$DESTINO/settings.json"
HOOKS=0
QUITAR=0

for a in "$@"; do
  case "$a" in
    --hooks) HOOKS=1 ;;
    --uninstall) QUITAR=1 ;;
    -h|--help) sed -n '2,12p' "$0"; exit 0 ;;
    *) echo "opcion desconocida: $a" >&2; exit 2 ;;
  esac
done

command -v python3 >/dev/null || { echo "hace falta python3" >&2; exit 1; }

respaldo() {
  [ -f "$AJUSTES" ] || return 0
  local b="$AJUSTES.bak.$(date +%Y%m%d-%H%M%S)"
  cp "$AJUSTES" "$b"
  echo "  respaldo: $b"
}

# ---------------------------------------------------------------------------
if [ "$QUITAR" = 1 ]; then
  echo "Desinstalando..."
  respaldo
  python3 - "$AJUSTES" <<'PY' || echo "  settings.json ilegible: sigo y borro los ficheros"
import json, sys, os
p = sys.argv[1]
if not os.path.exists(p):
    raise SystemExit(0)
try:
    d = json.load(open(p, encoding="utf-8"))
except Exception as e:
    # Se avisa y se sigue: si esto abortara, con `set -e` no se borraria
    # ningun fichero y no habria forma de desinstalar sin editar json a mano.
    print("  no puedo tocar settings.json (%s)" % e, file=sys.stderr)
    raise SystemExit(1)
d.pop("statusLine", None)
h = d.get("hooks") or {}
for ev in list(h):
    h[ev] = [g for g in h[ev]
             if not any("pet-hook.sh" in (c.get("command") or "")
                        for c in (g.get("hooks") or []))]
    if not h[ev]:
        del h[ev]
if h:
    d["hooks"] = h
else:
    d.pop("hooks", None)

def guardar(d, p):
    """Escritura atomica. Este fichero lleva tu modelo, tus permisos y tus MCP:
    un `open(p,"w")` lo trunca ANTES de escribir, asi que un fallo a media
    escritura lo deja vacio. Temporal al lado y rename, que es atomico."""
    import tempfile
    dirn = os.path.dirname(os.path.abspath(p)) or "."
    fd, tmp = tempfile.mkstemp(dir=dirn, prefix=".settings-", suffix=".json")
    try:
        with os.fdopen(fd, "w", encoding="utf-8") as fh:
            json.dump(d, fh, indent=2, ensure_ascii=False)
            fh.write("\n")
            fh.flush()
            os.fsync(fh.fileno())
        os.replace(tmp, p)
    except Exception:
        try:
            os.unlink(tmp)
        except OSError:
            pass
        raise

guardar(d, p)
print("  settings.json limpio (el tema no se toca: cambialo con /theme)")
PY
  rm -f "$DESTINO/statusline.sh" "$DESTINO/bicho.py" "$DESTINO/pet" \
        "$DESTINO/pet-hook.sh" "$DESTINO/commands/pet.md" "$DESTINO/commands/feed.md"
  rm -rf "$DESTINO/__pycache__"
  echo "Hecho. El ~/.claude/pet.json se queda: bórralo tú si quieres empezar de cero."
  exit 0
fi

# ---------------------------------------------------------------------------
echo "Instalando en $DESTINO"
mkdir -p "$DESTINO/themes" "$DESTINO/commands"

for t in "$ORIGEN"/themes/*.json; do
  cp "$t" "$DESTINO/themes/"
  echo "  tema: $(basename "$t")"
done

install -m 755 "$ORIGEN/statusline.sh" "$DESTINO/statusline.sh"
install -m 644 "$ORIGEN/bicho.py"      "$DESTINO/bicho.py"
install -m 755 "$ORIGEN/pet"           "$DESTINO/pet"
install -m 755 "$ORIGEN/hooks/pet-hook.sh" "$DESTINO/pet-hook.sh"
rm -rf "$DESTINO/__pycache__"
echo "  statusline.sh, bicho.py, pet, pet-hook.sh"

cat > "$DESTINO/commands/pet.md" <<'MD'
---
description: El panel del bicho — nivel, evolución, xp, hambre y racha
---
Ejecuta `~/.claude/pet` con Bash y muestra su salida tal cual, sin resumirla
ni reformatearla. Es un panel ya compuesto y coloreado.
MD

cat > "$DESTINO/commands/feed.md" <<'MD'
---
description: Dale de comer al bicho (+3 xp, hambre −2, máximo 4 al día)
---
Ejecuta `~/.claude/pet feed` con Bash y muestra su salida tal cual.
MD
echo "  comandos: /pet, /feed"

respaldo
python3 - "$AJUSTES" "$DESTINO" "$HOOKS" <<'PY'
import json, os, sys

p, destino, con_hooks = sys.argv[1], sys.argv[2], sys.argv[3] == "1"
d = {}
if os.path.exists(p):
    try:
        d = json.load(open(p, encoding="utf-8"))
    except Exception:
        print("  settings.json ilegible: no lo toco", file=sys.stderr)
        raise SystemExit(1)

d["statusLine"] = {"type": "command",
                   "command": os.path.join(destino, "statusline.sh").replace(
                       os.path.expanduser("~"), "~", 1),
                   "hideVimModeIndicator": True,
                   "refreshInterval": 1}
print("  statusLine enganchada")

if con_hooks:
    cmd = os.path.join(destino, "pet-hook.sh")
    # PostToolUse va sin matcher a proposito: el francotirador cuenta CUANTAS
    # herramientas distintas se usan entre dos tareas cerradas, o sea que hay que
    # verlas todas. Para las que no interesan el hook sale en bash, sin arrancar
    # python: unos 3 ms, contra los 15 que costaria el interprete.
    quiero = {"PostToolUse": None, "PreCompact": None, "SessionEnd": None}
    h = d.setdefault("hooks", {})
    for ev, matcher in quiero.items():
        grupos = h.setdefault(ev, [])
        grupos[:] = [g for g in grupos
                     if not any("pet-hook.sh" in (c.get("command") or "")
                                for c in (g.get("hooks") or []))]
        g = {"hooks": [{"type": "command", "command": cmd, "timeout": 5}]}
        if matcher:
            g["matcher"] = matcher
        grupos.append(g)
    print("  hooks enganchados: PostToolUse (todas), PreCompact, SessionEnd")
else:
    print("  hooks NO instalados (pásale --hooks si los quieres)")


def guardar(d, p):
    """Escritura atomica. Este fichero lleva tu modelo, tus permisos y tus MCP:
    un `open(p,"w")` lo trunca ANTES de escribir, asi que un fallo a media
    escritura lo deja vacio. Temporal al lado y rename, que es atomico."""
    import tempfile
    dirn = os.path.dirname(os.path.abspath(p)) or "."
    fd, tmp = tempfile.mkstemp(dir=dirn, prefix=".settings-", suffix=".json")
    try:
        with os.fdopen(fd, "w", encoding="utf-8") as fh:
            json.dump(d, fh, indent=2, ensure_ascii=False)
            fh.write("\n")
            fh.flush()
            os.fsync(fh.fileno())
        os.replace(tmp, p)
    except Exception:
        try:
            os.unlink(tmp)
        except OSError:
            pass
        raise

guardar(d, p)
PY

echo
echo "Listo. Falta elegir el tema: /theme -> Terminal"
echo "Abre una terminal NUEVA para que llegue COLORTERM; sin eso lo ves cuantizado."
