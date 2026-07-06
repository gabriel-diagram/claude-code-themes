# claude-code-themes

Dos temas de color para la CLI de [Claude Code](https://claude.com/claude-code) —
uno **cálido** (Blood Red, coral/terracota/vino) y uno **frío** (Electric Blue,
azul eléctrico oscuro) — más una **statusline** de 3 líneas adaptativa, con
detección de entorno y un indicador vivo del estado de la sesión.

| Tema | Acento | Look |
| --- | --- | --- |
| **Blood Red** | `#ff5c47` coral | cálido: coral, terracota, vino sobre fondo oscuro |
| **Electric Blue** | `#2e8bff` azul | frío: cian, azure, azul profundo sobre fondo oscuro |

![Electric Blue (izquierda) y Blood Red (derecha) en Claude Code](assets/preview.png)

*Izquierda: **Electric Blue** (prompts, picker y selección en azul). Derecha:
**Blood Red** (prompts coral y la statusline de 3 líneas con el indicador de vida
de sesión).*

## Contenido

- `themes/blood-red.json` — tema cálido instalable vía `/theme`
- `themes/electric-blue.json` — tema frío instalable vía `/theme`
- `statusline.sh` — statusline de 3 líneas (perfil rojizo/azulado según entorno)

## Instalación

### 1. Temas de color

```bash
mkdir -p ~/.claude/themes
cp themes/blood-red.json     ~/.claude/themes/blood-red.json
cp themes/electric-blue.json ~/.claude/themes/electric-blue.json
```

Dentro de una sesión de Claude Code, ejecuta `/theme` y elige **Blood Red** o
**Electric Blue**. Recargan en caliente al editar el JSON (no hace falta
reiniciar).

O actívalo directamente en `~/.claude/settings.json`:

```json
{
  "theme": "custom:electric-blue"
}
```

> El slug (`custom:<slug>`) sale del **nombre del archivo** sin `.json`, no del
> campo `"name"`. `blood-red.json` → `custom:blood-red`.

### 2. Statusline

```bash
cp statusline.sh ~/.claude/statusline.sh
chmod +x ~/.claude/statusline.sh
```

Añade (o mergea) en `~/.claude/settings.json`:

```json
{
  "statusLine": {
    "type": "command",
    "command": "~/.claude/statusline.sh",
    "hideVimModeIndicator": true
  }
}
```

La statusline solo necesita `python3` (usa la librería estándar; nada de `jq`).
`git` es opcional: si no está, degrada con elegancia.

## La statusline en detalle

Tres líneas que truncan al ancho real de la terminal y sueltan elementos de menor
prioridad antes que hacer *wrap* (que descuadra la caja del prompt):

- **L1** — `▍modelo · estilo · dir (rama✷) · VIM`
- **L2** — `[barra de contexto] % · $coste · +/− líneas · effort · límites 5h/7d`
- **L3** — *el bicho*: una carita + barra de vida que refleja el **estado real**
  de la sesión. La "vida" es `100 − peor cuello` entre contexto, límite de 5h y
  límite de 7d, con curva cuadrática (el margen no "duele" hasta acercarse al
  tope). No finge emociones: si un límite llega al 100%, el bicho hace k.o. y te
  dice qué lo mató.

### Perfil por entorno

El script pinta **azulado en local** y **rojizo dentro de un contenedor**
(detecta `/.dockerenv` o la env `container`). Para forzar el perfil a mano:

```bash
export STATUSLINE_PROFILE=docker   # o "local"
```

La statusline usa ANSI de 256 colores, así que se ve bien en cualquier terminal
razonable sin necesidad de truecolor.

## Gotcha: truecolor en Docker / SSH

Los **temas** sí usan color de 24 bits (`#rrggbb`). Si corres Claude Code dentro
de un contenedor o por SSH y los colores del tema salen "planos" o los tonos
sutiles (p. ej. el fondo de selección) no se distinguen del fondo, casi seguro es
que falta `COLORTERM`: sin él, la interfaz cuantiza cada hex al color más cercano
de una paleta reducida y los tonos parecidos colapsan al mismo.

`docker run` **no** propaga `COLORTERM` por defecto. Pásalo explícito:

```bash
docker run -e COLORTERM=truecolor -e TERM=xterm-256color ...
```

Por SSH, asegúrate de que tu terminal exporta `COLORTERM=truecolor` en el host
remoto.

## Paletas

### Blood Red (cálido)

| Rol | Token | Hex |
| --- | --- | --- |
| Acento de marca | `claude` | `#ff5c47` |
| Texto | `text` | `#f0e4e0` |
| Líneas / bordes | `subtle` | `#a5342c` |
| Borde del input | `promptBorder` | `#b83028` |
| Borde en modo `!bash` | `bashBorder` | `#c97b4a` |
| Fondo de selección | `selectionBg` | `#6b2028` |
| Fondo de tus mensajes | `userMessageBackground` | `#2a1416` |
| Error / Éxito / Aviso | `error`/`success`/`warning` | `#ff4b3e` / `#8fae7a` / `#e0a35c` |

### Electric Blue (frío)

| Rol | Token | Hex |
| --- | --- | --- |
| Acento de marca | `claude` | `#2e8bff` |
| Texto | `text` | `#e7eef4` |
| Líneas / bordes | `subtle` | `#2f5aa0` |
| Borde del input | `promptBorder` | `#1f4fcc` |
| Borde en modo `!bash` | `bashBorder` | `#4a7bc9` |
| Fondo de selección | `selectionBg` | `#173a66` |
| Fondo de tus mensajes | `userMessageBackground` | `#0f1e38` |
| Error / Éxito / Aviso | `error`/`success`/`warning` | `#ff4b3e` / `#8fae7a` / `#e0a35c` |

Los tokens no listados en `overrides` heredan del preset `dark` de Claude Code.

## Limitación conocida

El **banner de bienvenida** de arranque ("Claude Code vX.Y.Z", "Tips for getting
started", "What's new") usa acentos de onboarding que **no** forman parte del
sistema de temas — se quedan en el rosa/coral de marca sin importar el tema
activo. No es un fallo del tema.

## Licencia

[MIT](LICENSE).
