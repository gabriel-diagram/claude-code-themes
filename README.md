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

En terminal ancha (≥60 columnas) son **seis filas**: los datos a la izquierda y
la **tarjeta del bicho** anclada a la derecha de las filas 2 a 6.

```
▍Opus 5 (1M context) │ default │ projects/claude-code-themes (main✷)
███████░░░░░░░░░ 45% │ $4.56 │ +0/-0                       vibrante
xhigh │ 5h 28% 7d 12%                                     ██████████
                                                        ███ >    < ███
                                                          ██▀▀██▀▀██
                                                          ▰▰▰▰▰▰▰▰▰▱
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

Por debajo de 60 columnas la tarjeta desaparece y se vuelve al diseño compacto de
**tres filas**, con los datos en una sola línea y la carita en línea: `✦ ◕▿◕
▰▰▰▰▰▰ fresca`.

### La mascota

La tarjeta son cinco filas de 14 celdas, todo centrado al ancho del sprite:

```
   vibrante     <- estado
  ██████████    <- cabeza
███ >    < ███  <- tronco: orejas fuera y ojos recortados
  ██▀▀██▀▀██    <- las tres patas
  ▰▰▰▰▰▰▰▰▰▱    <- barra de vida
```

El bicho son 14×6 píxeles en tres filas de texto, con medios bloques para la
resolución vertical. El cuerpo apoya en `▀` y cada pata baja a `█`, que es lo que
deja el hueco entre patas.

Los ojos van **tallados** en el cuerpo: el tronco se pinta a color de fondo y
encima se dibuja en negro con medios bloques, de modo que cada ojo ocupa **2×2
celdas** repartidas entre la fila de la cabeza y la del tronco. Eso es lo que
permite que el chevron del logo salga de verdad —

```
▀▄        ▄▀
▄▀        ▀▄
```

— en vez de un `>` de una sola celda, que a este tamaño se lee como un carácter
de texto pegado al dibujo y no como parte de él.

**El cuerpo va del color del estado**, el mismo que la etiqueta y la barra: no
tiene sentido un bicho azul brillante con la barra de vida en rojo. Cada estado
tiene además su versión:

| Vida | Ojos | Postura |
| --- | --- | --- |
| ≥95 | chevron `>` `<` | de pie, anda, con destello en la etiqueta |
| ≥80 | chevron `>` `<` | de pie, anda |
| ≥60 | redondos | de pie, anda |
| ≥40 | entrecerrados | de pie, ya no anda |
| ≥20 | párpados caídos | hombros caídos, quieto |
| <20 | aspas | hombros caídos, quieto |
| k.o. | aspas | **patas al aire**, cuerpo tumbado debajo, en gris |

Animaciones, todas al techo de 1 fps:

- **anda** — ciclo de cuatro fotogramas: la onda recorre las tres patas de fuera a
  dentro. Cuatro segundos de cada doce; con `STATUSLINE_MASCOT_WALK=1`, siempre.
- **parpadea** — un fotograma de cada siete. Por debajo de 40 de vida el parpadeo
  se vuelve más lento y pesado (uno de cada cuatro, a párpados caídos).
- **espasmo** — ya muerto, una pata da un tirón cada nueve segundos.

**Por qué la fila 1 se queda libre:** Claude Code pinta sus propios badges
alineados a la derecha de la primera fila de la statusline. La tarjeta arranca en
la fila 2 para no chocar con ellos.

**El techo de la animación son 1 fps.** La statusline se re-ejecuta por eventos
(con *debounce* de 300 ms) y en reposo solo si defines `refreshInterval`, cuyo
mínimo es `1` segundo. De ahí que ande a ratos en vez de sin parar: a un fotograma
por segundo, un baile continuo en la esquina del ojo distrae más que decora. Sin
`refreshInterval` la mascota sigue funcionando, pero solo cambia de frame cuando
algo la re-dibuja.

**Ajustes por entorno:**

| Variable | Efecto |
| --- | --- |
| `STATUSLINE_MASCOT=0` | la apaga |
| `STATUSLINE_MASCOT_COLOR` | `#rrggbb` o índice 0-255: fija el cuerpo a un color y deja de seguir al estado |
| `STATUSLINE_MASCOT_WALK=1` | anda sin parar |
| `STATUSLINE_RIGHT_PAD` | margen derecho, por defecto `6` (ver abajo) |

**Sobre los espacios de la izquierda.** Claude Code **recorta los espacios del
principio** de cada línea de la statusline. Una fila cuya mitad izquierda va vacía
es solo "espacios + tarjeta": al recortarlos, la tarjeta se cae al borde izquierdo
y acabas con trozos del bicho sueltos por la pantalla. Esas filas se anclan con un
braille en blanco (`U+2800`), que se pinta vacío pero no es un espacio, así que el
recorte no lo toca. Si tu fuente no lo tiene y ves un cuadrito, cámbialo por
cualquier otro carácter invisible en la función que arma las filas.

**Sobre el margen derecho.** La statusline **no recibe el ancho de la terminal**:
no hay campo para eso en el JSON, así que el script lo saca de `COLUMNS`. Y Claude
Code recorta la línea unas 5 columnas antes de `COLUMNS`, de modo que alinear a la
derecha sobre `COLUMNS-1` trunca la tarjeta o la hace *wrap* (la barra de vida se
cae a la línea siguiente). De ahí el margen de 6 por defecto. Si en tu terminal la
tarjeta queda demasiado despegada del borde, bájalo; si se corta, súbelo.

Se oculta sola si la terminal baja de 60 columnas, y las líneas de datos se
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
