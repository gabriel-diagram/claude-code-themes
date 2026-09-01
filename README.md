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
    "hideVimModeIndicator": true,
    "refreshInterval": 1
  }
}
```

La statusline solo necesita `python3` (usa la librería estándar; nada de `jq`).
`git` es opcional: si no está, degrada con elegancia.

## La statusline en detalle

Las líneas truncan al ancho real de la terminal y sueltan elementos de menor
prioridad antes que hacer *wrap* (que descuadra la caja del prompt).

En terminal ancha (≥70 columnas) son **seis filas**: los datos a la izquierda y
la **tarjeta del bicho** anclada a la derecha de las filas 2 a 6.

```
▍Opus 5 (1M context) │ default │ projects/claude-code-themes (main✷)
███████░░░░░░░░░ 45% │ $4.56 │ +0/-0                    vibrante
xhigh │ 5h 28% 7d 12%                              ██████████████████
                                                ████████████████████████
                                                   ████▀▀▀████▀▀▀████
                                                     ▰▰▰▰▰▰▰▰▰▰▰▱▱▱
```

- **fila 1** — `▍modelo · estilo · dir (rama✷) · VIM`
- **fila 2** — `[barra de contexto] % · $coste · +/− líneas`
- **fila 3** — `effort · límites 5h/7d`
- **fila 4** — el cuello, solo cuando aprieta: `cuello: ctx`, o qué te mató
- **derecha** — la tarjeta: estado arriba, sprite en medio, barra de vida abajo

*El bicho* refleja el **estado real** de la sesión. La "vida" es `100 − peor
cuello` entre contexto, límite de 5h y límite de 7d, con curva cuadrática (el
margen no "duele" hasta acercarse al tope). No finge emociones: si un límite
llega al 100%, el bicho hace k.o. y te dice qué lo mató.

Por debajo de 70 columnas la tarjeta desaparece y se vuelve al diseño compacto de
**tres filas**, con los datos en una sola línea y la carita en línea: `✦ ◕▿◕
▰▰▰▰▰▰ fresca`.

### La mascota

La tarjeta son cinco filas de 24 celdas, todo centrado al ancho del sprite:

```
        vibrante          <- estado
   ██████████████████     <- cabeza
████████████████████████  <- tronco, con las orejas saliendo 3 a cada lado
   ████▀▀▀████▀▀▀████     <- las tres patas
     ▰▰▰▰▰▰▰▰▰▰▰▱▱▱       <- barra de vida
```

El bicho es una silueta: 24×6 píxeles en tres filas de texto, con medios bloques
para la resolución vertical. El cuerpo apoya en `▀` y cada pata baja a `█`, que es
lo que deja el hueco entre patas.

**Sin ojos, a propósito.** A esta escala unos ojos de texto (`>` `<`) se leen como
caracteres sueltos, no como parte del dibujo. El estado lo cuentan la etiqueta, la
barra de vida y el color del cuerpo:

| Vida | Cuerpo | Patas |
| --- | --- | --- |
| ≥20 | acento del tema | anda |
| <20 | rojo (`error` del tema) | anda |
| k.o. | gris | tumbado, sin patas |

**Por qué la fila 1 se queda libre:** Claude Code pinta sus propios badges
alineados a la derecha de la primera fila de la statusline. La tarjeta arranca en
la fila 2 para no chocar con ellos.

**Animación.** El techo real son **1 fps**: la statusline se re-ejecuta por
eventos (con *debounce* de 300 ms) y en reposo solo si defines `refreshInterval`,
cuyo mínimo es `1` segundo. Con eso la mascota parpadea (1 de cada 6 frames) y da
tres pasos cada 12 segundos — andar sin parar en la esquina del ojo distrae más
que decora. Sin `refreshInterval` la mascota sigue funcionando, pero solo cambia
de frame cuando algo la re-dibuja.

**Ajustes por entorno:**

| Variable | Efecto |
| --- | --- |
| `STATUSLINE_MASCOT=0` | la apaga |
| `STATUSLINE_MASCOT_COLOR` | `#rrggbb` o índice 0-255; por defecto `#2e8bff` en local y el coral `#da7756` del logo en docker |
| `STATUSLINE_MASCOT_WALK=1` | anda sin parar |
| `STATUSLINE_RIGHT_PAD` | margen derecho, por defecto `6` (ver abajo) |

**Sobre el margen derecho.** La statusline **no recibe el ancho de la terminal**:
no hay campo para eso en el JSON, así que el script lo saca de `COLUMNS`. Y Claude
Code recorta la línea unas 5 columnas antes de `COLUMNS`, de modo que alinear a la
derecha sobre `COLUMNS-1` trunca la tarjeta o la hace *wrap* (la barra de vida se
cae a la línea siguiente). De ahí el margen de 6 por defecto. Si en tu terminal la
tarjeta queda demasiado despegada del borde, bájalo; si se corta, súbelo.

Se oculta sola si la terminal baja de 70 columnas, y las líneas de datos se
ensamblan sobre el ancho ya descontado, así que la tarjeta nunca empuja contenido
fuera.

El color usa **24 bits si hay `COLORTERM`**, y si no cae al 256 más cercano. Ojo:
Windows Terminal / WSL no exportan `COLORTERM` por defecto — sin él tanto la
mascota como los *temas* salen cuantizados. Ponlo en tu `.zshrc`/`.bashrc`:

```bash
export COLORTERM=truecolor
```

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
