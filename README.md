# claude-code-themes

Tres temas de color para la CLI de [Claude Code](https://claude.com/claude-code)
más una **statusline** de tres bandas con un bicho a la derecha que refleja el
estado real de la sesión — y que **evoluciona** según cómo trabajas: 27 formas
en un árbol de cinco niveles, con XP, hambre y racha.

| Tema | Acento | Look |
| --- | --- | --- |
| **Terminal** | `#4dd6c1` turquesa | un color por tipo de dato: rutas, identificadores, urls, números, modos |
| **Blood Red** | `#ff5c47` coral | cálido: coral, terracota, vino sobre fondo oscuro |
| **Electric Blue** | `#2e8bff` azul | frío: cian, azure, azul profundo sobre fondo oscuro |

![Electric Blue (izquierda) y Blood Red (derecha) en Claude Code](assets/preview.png)

*Izquierda: **Electric Blue** (prompts, picker y selección en azul). Derecha:
**Blood Red** (prompts coral y la statusline de 3 líneas con el indicador de vida
de sesión).*

## Contenido

- `themes/terminal.json` — tema por tipo de dato, el que empareja con la statusline
- `themes/blood-red.json` — tema cálido instalable vía `/theme`
- `themes/electric-blue.json` — tema frío instalable vía `/theme`
- `statusline.sh` — statusline de tres bandas con el bicho
- `bicho.py` — las 27 plantillas, los siete estados y el árbol de evoluciones
- `pet` — el panel del bicho (`/pet`) y su comida (`/feed`)
- `hooks/pet-hook.sh` — traduce commits, tests y compactados en comida
- `install.sh` — instala todo lo anterior; con `--hooks`, también los hooks
- [`VIDA.md`](VIDA.md) — la capa del momento: qué hace que pase de fresca a k.o.
- [`EVOLUCIONES.md`](EVOLUCIONES.md) — la capa permanente: XP, comida y las 27 formas
- [`AUDITORIA.md`](AUDITORIA.md) — rendimiento y errores, medido

## Instalación

```bash
./install.sh            # temas + statusline + bicho + /pet y /feed
./install.sh --hooks    # además engancha los hooks que le dan de comer
./install.sh --uninstall
```

Hace copia de seguridad de `~/.claude/settings.json` antes de tocarlo, es
idempotente, y **no cambia el tema activo**: eso se elige con `/theme` → Terminal.

Los **hooks van aparte a propósito**. Viven en `~/.claude/settings.json`, que es
global, así que corren en todos tus repos. Sin ellos el bicho existe y se ve,
pero no come solo: se le da con `/feed`.

### A mano, si prefieres

```bash
mkdir -p ~/.claude/themes
cp themes/*.json ~/.claude/themes/
cp bicho.py ~/.claude/                      # las plantillas y el árbol
install -m 755 statusline.sh ~/.claude/     # necesita bicho.py al lado
install -m 755 pet ~/.claude/
```

Y en `~/.claude/settings.json`:

```json
{
  "theme": "custom:terminal",
  "statusLine": {
    "type": "command",
    "command": "~/.claude/statusline.sh",
    "hideVimModeIndicator": true,
    "refreshInterval": 1
  }
}
```

> El slug (`custom:<slug>`) sale del **nombre del archivo** sin `.json`, no del
> campo `"name"`. `blood-red.json` → `custom:blood-red`.

La statusline solo necesita `python3` (librería estándar; nada de `jq`). `git` es
opcional: si no está, degrada con elegancia. Si falta `bicho.py`, la statusline
sigue funcionando — sin bicho.

## La statusline en detalle

Es un **pie**, no una banda más: fondo un tono por encima del negro y una raya
fina arriba, para que se lea como parte de la ventana y no como una línea más
del hilo. Tres bandas a la izquierda y el bicho anclado a la derecha. Cada banda
agrupa datos que se miran juntos, y suelta los elementos de menor prioridad
antes que hacer *wrap* (que descuadra la caja del prompt):

```
──────────────────────────────────────────────────────────────────────────────
 Opus 5  ██████░░░░░░░░░░ 36% · 1M ctx │ xhigh │ 42.7 tok/s │ bypass   vibrante
rochas/rochas-energy-backend (fix-errors ✳) │ +184/−37 │ $28.29        |   |
api │ 5h ████░░░░░░ 41%  7d █░░░░░░░░░ 13% │ 1h 12m                   ▗█┼█┼█▖
                                                                     ▐█ > < █▌
                                                                      ▝█┼█┼█▘
▐█ > < █▌◗ racha de 5 días. no la rompas hoy                          ▝▝   ▘▘
```

- **Banda 1 · motor** — modelo, contexto, razonamiento y ritmo: lo que cambia
  cada turno.
- **Banda 2 · trabajo** — repo, rama, diff y coste: lo que se lleva a un commit.
  El nombre del repo lo da `workspace.repo` del payload (`owner/nombre`) cuando
  hay remoto; sin remoto se queda el nombre a secas, porque la carpeta de al lado
  no es un owner.
#### Ritmo y permisos

El **`tok/s`** es real, no una estimación, pero hay que mirar bien de dónde sale.
Los dos campos del payload **no miden lo mismo**: `total_output_tokens` es lo que
sacó la *última* respuesta (se reinicia en cada turno, no es un contador que
sube), mientras que `total_api_duration_ms` es el tiempo de API *acumulado* de la
sesión. El ritmo de la última respuesta es lo primero entre lo que ha crecido lo
segundo. Restar dos `total_output_tokens` seguidos no mide nada: son dos
respuestas distintas, y el resultado sale inflado o negativo según cuál fuera más
larga. Se apaga solo a los dos minutos sin moverse — cuando el modelo ya no está
hablando — y entonces el acierto de caché ocupa ese hueco. Nunca salen los dos.

El **modo de permisos** (`bypass`, `auto-edit`, `plan`) no viene en el payload,
pero sí en el transcript, cuya ruta sí llega. Se lee solo la cola del fichero
(0,02 ms) y sale como distintivo en la banda. No es una copia del pie que pinta
el propio Claude Code: aquello es suyo y no se puede tocar.

- **Banda 3 · cuota** — carpeta, límites 5h/7d y tiempo: se lee de reojo, va en
  gris. De la ruta sale **solo la carpeta en la que estás**, no el camino entero;
  y si se llama igual que el repo —- es decir, estás en su raíz—- desaparece,
  porque eso ya lo dice la banda 2.
- **El bocadillo** cierra el pie. Quien habla lo dice la cara del propio bicho,
  con la cola apuntando al texto; sale solo cuando hay algo que decir y solo a
  partir de 100 columnas.

Las tres bandas van dentro de la columna izquierda, contra la tarjeta del bicho:
seis filas en total, más la raya. Por debajo de 55 columnas el bicho desaparece y
quedan las tres bandas solas.

### El bicho

Nueve columnas por cinco filas. La silueta la elige la **evolución**; los ojos,
las patas y el color los elige el **estado**.

```
  |   |     <- marca de la ramificación (se cae al llegar a cansada)
 ▗█┼█┼█▖    <- cuerpo de la evolución
▐█ > < █▌   <- cara, con las orejas fuera
 ▝█┼█┼█▘
 ▘▘   ▝▝    <- patas
```

Siete estados, y **cuatro señales independientes** que cambian en este orden: los
ojos, luego el paso de las patas, luego la cabeza se hunde y al final la silueta
se tumba. A un vistazo se distingue *cansada* de *ahogada* sin leer la etiqueta.

| Uso | Etiqueta | Ojos | Cabeza | Patas |
| --- | --- | --- | --- | --- |
| ≤22% | fresca ✦ | `>` `<` | sí | anda |
| ≤45% | vibrante | `>` `<` | sí | anda |
| ≤63% | a gusto | `o` `o` | sí | anda |
| ≤78% | espesa | `▬` `▬` | sí | quieto |
| ≤89% | cansada | `_` `_` | **hundida** | quieto |
| <100% | ahogada | `x` `x` | hundida | quieto |
| 100% | k.o. | `x` `x` | hundida | **tumbado**, patas al aire |

Los ojos de las dos primeras filas son **los de la evolución**, que tiene los
suyos; de *a gusto* para abajo manda el estado. Así el cansancio se lee igual sea
cual sea el bicho.

**Movimiento.** Las patas alternan `▘ ▝` ↔ `▝ ▘` en cada refresco de la
statusline: anda mientras trabajas y se queda quieto a partir de *espesa*. El
parpadeo (`_ _`) es un solo frame cada siete refrescos. Al cruzar un umbral la
etiqueta sale en negrita un refresco — para eso guarda el estado anterior en
`$TMPDIR/claude-statusline-<session_id>`.

A 1 fps, unas patas alternando sin parar en la esquina del ojo cansan, así que
**por defecto anda cuatro segundos de cada doce**. `STATUSLINE_BICHO_ANDA=1`
devuelve el baile continuo del diseño.

### Evoluciones

La forma no la eliges: sale de cómo trabajas. Los commits y los `/compact` te
llevan por la rama metódica, los tests y los planes por la inquisitiva, y
reventar el contexto por la impulsiva.

```
chispa -> pauta / sonda / brasa -> siete oficios -> catorce marcas
```

```bash
/pet     # nivel, evolución, xp, hambre, racha y la comida de hoy
/feed    # +3 xp, hambre −2 (máximo 4 al día)
```

El árbol entero, la tabla de comida y qué alimenta cada contador están en
[EVOLUCIONES.md](EVOLUCIONES.md). Sin `./install.sh --hooks` el bicho existe y
se ve, pero solo come con `/feed`.

### El uso: media ponderada

Un solo número entre 0 y 100 decide estado, ojos, patas y color:

```
uso = 0.5 · ctx  +  0.3 · límite 5h  +  0.2 · límite 7d
```

**Por qué ponderada y no el peor de los tres.** El contexto es lo único que
puedes gestionar en el momento —compactar, cerrar la sesión, abrir otra—; los
límites solo avisan. Que el de 7 días vaya por el 80% no debería poner al bicho
al borde de la muerte si acabas de abrir la sesión.

Si falta alguno de los tres (las cuentas por API no reciben `rate_limits`), su
peso se reparte entre los que sí llegan. Y **el k.o. es el único caso exacto**:
hace falta el 100% clavado.

### Ajustes por entorno

| Variable | Efecto |
| --- | --- |
| `STATUSLINE_BICHO=0` | apaga el bicho, deja las tres bandas |
| `STATUSLINE_BICHO_ANDA=1` | anda en cada refresco en vez de a ratos |
| `STATUSLINE_FONDO=0` | quita el fondo del pie, deja las líneas transparentes |
| `STATUSLINE_REGLA=0` | quita la raya de arriba y ahorra una fila |
| `STATUSLINE_RIGHT_PAD` | margen derecho, por defecto `6` (ver abajo) |
| `PET_TEST_RUNNERS` | regex extra para reconocer tu runner de tests |

**Sobre los espacios de la izquierda.** Claude Code **recorta los espacios del
principio** de cada línea de la statusline. Las filas cuya mitad izquierda va
vacía son solo "espacios + bicho": al recortarlos, el bicho se cae al borde
izquierdo y acabas con trozos sueltos por la pantalla. Esas filas se anclan con
un braille en blanco (`U+2800`), que se pinta vacío pero no es un espacio.

**Sobre el margen derecho.** La statusline **no recibe el ancho de la terminal**:
no hay campo para eso en el JSON, así que sale de `COLUMNS`. Y Claude Code
recorta la línea unas 5 columnas antes de `COLUMNS`, de modo que alinear a la
derecha sobre `COLUMNS-1` trunca el bicho o lo hace *wrap*. De ahí el margen de 6
por defecto: si queda demasiado despegado del borde, bájalo; si se corta, súbelo.

**Lo que no puede hacer.** La línea de `bypass permissions` y los badges tipo
`/rc active` los pinta Claude Code en su footer, no la statusline. El script
imprime texto; dónde coloca él sus badges no está en su mano.

**El techo de la animación son 1 fps.** La statusline se re-ejecuta por eventos
(con *debounce* de 300 ms) y en reposo solo si defines `refreshInterval`, cuyo
mínimo es `1` segundo. Sin `refreshInterval` el bicho sigue funcionando, pero
solo cambia de frame cuando algo lo re-dibuja.

## Gotcha: truecolor en Docker / SSH

Los **temas** y la statusline usan color de 24 bits (`#rrggbb`). Sin `COLORTERM`
la interfaz cuantiza cada hex al color más cercano de una paleta reducida y los
tonos parecidos colapsan al mismo. **Windows Terminal y WSL no exportan
`COLORTERM`**, así que ponlo en tu `.zshrc`/`.bashrc`:

```bash
export COLORTERM=truecolor
```

`docker run` tampoco lo propaga:

```bash
docker run -e COLORTERM=truecolor -e TERM=xterm-256color ...
```


## Paletas

### Terminal — un color por tipo de dato

La regla es que un tipo de dato siempre lleva el mismo color, en la statusline y
en la prosa. Así no hay que leer para saber qué estás mirando.

| Rol | Hex | |
| --- | --- | --- |
| Rutas, ficheros, tablas, repos | `#4DD6C1` | turquesa |
| Identificadores, código inline, altas | `#57E389` | verde |
| Urls, ramas, enlaces | `#6FB6FF` | azul claro |
| Números, dinero, métricas, avisos | `#E8C46A` | ámbar |
| Modos y ajustes del propio CLI | `#B07CF0` | violeta |
| Bajas, errores, riesgo | `#F2777A` | salmón |
| Énfasis en prosa (negrita) | `#ECEFF4` | casi blanco |
| Flechas, separadores, unidades | `#6B7683` | gris |

Las barras de límite van en `#4EA3F5` sobre `#1D2B38`, y el fondo de los mensajes
en `#0A0D0F`.

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
