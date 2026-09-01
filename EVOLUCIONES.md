# Las evoluciones del bicho

El bicho tiene **dos capas que no se mezclan**:

| | qué mide | de dónde sale | sube y baja |
| --- | --- | --- | --- |
| **vida** | cómo está *ahora* | uso de contexto y cuota | sí, todo el rato |
| **progreso** | lo que llevas hecho | XP acumulada en `~/.claude/pet.json` | el nivel nunca baja |

La **vida** elige los ojos, las patas y el color — está en [VIDA.md](VIDA.md).
La **progreso** elige la silueta, que es de lo que va este documento.

## Una plantilla, siete estados

Cada evolución es una **plantilla de 5 filas y 9 columnas** con dos huecos de
ojos y una fila de patas:

```
  |   |     <- marca de la ramificación
 ▗█┼█┼█▖    <- cuerpo de la evolución
▐█ o o █▌   <- los ojos los pone el estado
 ▝█┼█┼█▘
 ▘▘   ▝▝    <- las patas las pone el estado
```

El estado de vida **no cambia la silueta**: rellena los huecos y elige el color.
Por eso las 27 evoluciones tienen sus siete estados sin dibujar 189 sprites.

Los ojos siguen una regla: la evolución pone los suyos mientras está entera
(*fresca* y *vibrante*), y de *a gusto* para abajo manda el estado
(`o o` → `▬ ▬` → `_ _` → `x x`). Así el cansancio se lee de un vistazo aunque
no sepas qué bicho es.

## El árbol

```
chispa
├─ pauta  compacta y commitea corto
│  ├─ refactor  muchos diffs pequeños
│  │  ├─ cirujano......  20 diffs seguidos sin un rechazo
│  │  └─ tejedor.......  un refactor que toca 10+ ficheros
│  └─ pulcro  jamás pasa del 60%
│     ├─ monje.........  5 sesiones sin pasar del 40% de contexto
│     └─ jardinero.....  docs y limpieza dos días seguidos
├─ sonda  lee, planifica, testea
│  ├─ cazabugs  tests y fixes en cadena
│  │  ├─ sabueso.......  repro antes del fix, 10 veces
│  │  └─ exterminador..  15 tests verdes sin uno rojo
│  └─ arquitecto  planes largos, docs
│     ├─ cartógrafo....  un plan de 10 tareas cerrado entero
│     └─ oráculo.......  5 planes escritos antes de tocar código
└─ brasa  tira al límite sin frenar
   ├─ velocista  sesiones cortas, rápidas
   │  ├─ relámpago.....  10 sesiones de menos de 15 minutos
   │  └─ francotirador.  8 tareas cerradas con una sola herramienta
   ├─ maratón  sesiones de horas
   │  ├─ buey..........  3 sesiones de más de 4 horas
   │  └─ topo..........  5 días seguidos en el mismo repo
   └─ salvaje  al límite, sin compactar
      ├─ gremlin.......  30 turnos con permisos en bypass
      └─ kraken........  3 sesiones tocando el 100% de contexto
```

**La rama no la eliges.** En cada bifurcación gana el contador de comportamiento
que va más alto en ese momento. Si cambias de hábitos antes de subir de nivel,
cambias de rama.

Las de nivel 5 no dependen de la XP sino del **hábito**: se desbloquean al
cumplir su condición estando en la evolución padre.

| Nivel | XP | Qué eres |
| --- | --- | --- |
| 1 | 0 | larva — `chispa`, sin patas todavía |
| 2 | 60 | temperamento — cómo trabajas |
| 3 | 180 | oficio — en qué eres bueno |
| 4 | 400 | el mismo oficio, asentado |
| 5 | 900 | marca — más la condición de hábito |

Y dos secretas fuera del árbol: **fénix** (llegar a hambre 10 y remontar a 0 en
la misma sesión, solo desde `salvaje` o `maratón`) y **quimera** (dos
temperamentos empatados al subir a nivel 4; hereda ojos de uno y cuerpo del
otro).

## La comida

| Comida | XP | Hambre | Tope |
| --- | --- | --- | --- |
| tests en verde | **+15** | −4 | — |
| commit hecho | **+12** | −3 | — |
| `/compact` | **+8** | −3 | — |
| tarea del plan cerrada | **+6** | −1 | — |
| `/feed` | **+3** | −2 | 4 al día |
| contexto al 100% | **−15** | — | — |

El **hambre** sube +1 por hora sin comer, hasta 10. A partir de 7 los ojos se
apagan y el bicho pide comida en la statusline. **Nunca muere**: el hambre no
resta XP ni baja el nivel, solo se le nota en la cara.

Reventar el contexto resta 15 XP y rompe las rachas limpias, pero **el nivel
nunca baja**. Lo que se pierde es el camino hacia la siguiente evolución, no lo
andado.

## Qué alimenta cada contador

Los hooks (`hooks/pet-hook.sh`) traducen lo que haces en contadores:

| Contador | Se llena con |
| --- | --- |
| `metodico` | commits y `/compact` |
| `inquisitivo` | tests y tareas del plan |
| `impulsivo` | reventones de contexto |
| `diffs`, `racha_diffs` | commits (la racha se rompe al reventar) |
| `tests`, `racha_tests` | tests en verde (íd.) |
| `planes`, `plan_entero` | `TodoWrite` — el segundo es el plan más largo cerrado del todo |
| `commit_ancho` | el `N files changed` más alto de un commit |
| `ctx_bajo`, `sesiones_bajo_40` | el pico de contexto de la sesión, al cerrarla |
| `sesiones_cortas`, `sesiones_15min`, `sesiones_largas`, `sesiones_4h` | la duración de la sesión, al cerrarla |
| `sesiones_ctx100`, `ctx_limite` | tocar el 100% de contexto |
| `dias_mismo_repo` | días seguidos cerrando sesión en el mismo repo |

Los datos de forma de sesión —pico de contexto, duración, repo— **solo los ve la
statusline**, que los va dejando en su fichero temporal; el hook de `SessionEnd`
los convierte en contadores y borra el fichero.

## Lo que no se puede contar

**Cinco de las catorce marcas no son alcanzables**, y no por falta de ganas:

| Marca | Condición | Por qué no |
| --- | --- | --- |
| `sabueso` | repro antes del fix, 10 veces | hay que saber que un comando *es* una reproducción del bug |
| `oraculo` | 5 planes escritos antes de tocar código | hay que saber que el plan iba de ese código |
| `jardinero` | docs y limpieza dos días seguidos | hay que saber que un commit *es* limpieza |
| `francotirador` | 8 tareas cerradas con una sola herramienta | hay que atribuir herramientas a una tarea |
| `gremlin` | 30 turnos con permisos en bypass | el modo de permisos no llega al hook |

Las cuatro primeras piden entender la **intención** de lo que haces, no contar
eventos. Un hook ve comandos y salidas; eso no basta. Están cableadas en
`bicho.py` por si algún día llega el dato — su contador existe y se queda a 0.

Las otras nueve sí salen, y los siete oficios también.

## Heurísticas, y lo que fallan

Dos de las comidas son **adivinanzas sobre texto**, no hechos:

- **«commit hecho»** pide que `git … commit` esté al **principio del comando** o
  justo detrás de un operador de shell (`;`, `&&`, `|`, `(`). Sin ese anclaje, un
  `grep -rn "git commit"` contaba como commit. Además la salida no puede decir
  `nothing to commit`. Es la más fiable de las dos, y aun así es texto.
- **«tests en verde»** pide que el comando mencione un runner conocido y que la
  salida tenga una marca de verde (`^ok `, `passed`, `PASS`, `0 failures`…) y
  ninguna de rojo. Un runner exótico no cuenta; un test llamado
  `test_login_failed` puede contar como rojo.

Los dos patrones están arriba del todo en `hooks/pet-hook.sh` y se tocan sin
miedo. **Cualquier heurística sobre texto tiene falsos positivos**: la regla al
escribirlas ha sido preferir perderse una comida a inventarse una. No hay forma
de hacerlo exacto mientras el CLI no exponga el resultado de la herramienta.

## Los mandos

```bash
pet                    # el panel: nivel, evolución, xp, hambre, racha y la comida de hoy
pet feed               # +3 xp, hambre −2, máximo 4 al día
pet cuenta <c> [n]     # suma a un contador de comportamiento
pet record <c> <v>     # guarda el máximo de un contador
```

Instalados como `/pet` y `/feed` desde `./install.sh`.

Para empezar de cero: `rm ~/.claude/pet.json`.

---

Ver también el [README](README.md) para las bandas y la paleta, y
[VIDA.md](VIDA.md) para la otra capa, la del momento.
