# Las evoluciones del bicho

El bicho tiene **dos capas que no se mezclan**:

| | qué mide | de dónde sale | sube y baja |
| --- | --- | --- | --- |
| **vida** | cómo está *ahora* | uso de contexto y cuota | sí, todo el rato |
| **progreso** | lo que llevas hecho | XP acumulada en `~/.claude/pet.json` | el nivel nunca baja |

La **vida** elige los ojos, las patas y el color — está en [vitals.md](vitals.md).
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
(*fresh* y *lively*), y de *easy* para abajo manda el estado
(`o o` → `▬ ▬` → `_ _` → `x x`). Así el cansancio se lee de un vistazo aunque
no sepas qué bicho es.

## El árbol

```
spark
├─ pattern  compacta y commitea corto
│  ├─ refactor  muchos diffs pequeños
│  │  ├─ surgeon.......  20 diffs seguidos sin un rechazo
│  │  └─ weaver........  un refactor que toca 10+ ficheros
│  └─ tidy  jamás pasa del 60%
│     ├─ monk..........  5 sesiones sin pasar del 40% de contexto
│     └─ gardener......  docs y limpieza dos días seguidos
├─ probe  lee, planifica, testea
│  ├─ bughunter  tests y fixes en cadena
│  │  ├─ bloodhound....  repro antes del fix, 10 veces
│  │  └─ exterminator..  15 tests verdes sin uno rojo
│  └─ architect  planes largos, docs
│     ├─ cartographer..  un plan de 10 tareas cerrado entero
│     └─ oracle........  5 planes escritos antes de tocar código
└─ ember  tira al límite sin frenar
   ├─ sprinter  sesiones cortas, rápidas
   │  ├─ bolt..........  10 sesiones de menos de 15 minutos
   │  └─ sniper........  8 tareas cerradas con una sola herramienta
   ├─ marathon  sesiones de horas
   │  ├─ ox............  3 sesiones de más de 4 horas
   │  └─ mole..........  5 días seguidos en el mismo repo
   └─ feral  al límite, sin compactar
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
| 1 | 0 | larva — `spark`, sin patas todavía |
| 2 | 60 | temperamento — cómo trabajas |
| 3 | 180 | oficio — en qué eres bueno |
| 4 | 400 | el mismo oficio, asentado |
| 5 | 900 | marca — más la condición de hábito |

Y dos secretas fuera del árbol: **phoenix** (llegar a hambre 10 y remontar a 0 en
la misma sesión, solo desde `feral` o `marathon`) y **chimera** (dos
temperamentos empatados al subir a nivel 4; hereda ojos de uno y cuerpo del
otro).

## La comida

| Comida | XP | Hambre | Tope |
| --- | --- | --- | --- |
| tests en verde | **+15** | −4 | — |
| commit hecho | **+12** | −3 | — |
| `/compact` | **+8** | −3 | — |
| tarea del plan cerrada | **+6** | −1 | — |
| `/feed` | **+3** | −2 | uno cada 4 h |
| contexto al 100% | **−15** | — | — |

El **hambre** sube +1 por hora sin comer, hasta 10. A partir de 7 los ojos se
apagan y el bicho pide comida en la statusline. **Nunca muere**: el hambre no
resta XP ni baja el nivel, solo se le nota en la cara.

Reventar el contexto resta 15 XP y rompe las rachas limpias, pero **el nivel
nunca baja**. Lo que se pierde es el camino hacia la siguiente evolución, no lo
andado.

## Qué alimenta cada contador

Los hooks (`bin/pet-hook`) y la propia statusline traducen lo que haces en
contadores. **Las 27 evoluciones son alcanzables**: los siete oficios, las
catorce marcas y las dos secretas.

| Contador | Se llena con | Quién lo ve |
| --- | --- | --- |
| `metodico` | commits y `/compact` | hook |
| `inquisitivo` | tests y tareas del plan | hook |
| `impulsivo`, `ctx_maxed` | reventones de contexto | statusline |
| `diffs`, `diff_streak` | commits (la racha se rompe al reventar) | hook |
| `tests`, `test_streak` | tests en verde (íd.) | hook |
| `widest_commit` | el `N files changed` más alto | hook |
| `longest_plan` | el plan más largo cerrado del todo | hook |
| `plans_before_code` | un plan de 3+ tareas escrito antes de editar nada | hook |
| `single_tool_tasks` | tareas cerradas usando una sola herramienta | hook |
| `repro_before_fix` | un test rojo seguido de uno verde | hook |
| `docs_days` | días seguidos con un commit de docs o de limpieza | hook + `git show --numstat` |
| `bypass_turns` | prompts nuevos con los permisos en bypass | statusline + transcript |
| `ctx_low`, `sessions_under_40` | el pico de contexto de la sesión | statusline → `SessionEnd` |
| `short_sessions`, `sessions_15min`, `long_sessions`, `sessions_4h` | la duración de la sesión | íd. |
| `ctx100_sessions` | tocar el 100% de contexto | íd. |
| `same_repo_days` | días seguidos cerrando sesión en el mismo repo | íd. |

Tres datos **solo los ve la statusline**, porque no llegan a ningún hook: el uso
de contexto, el modo de permisos y el ritmo de tokens. Los va dejando en su
fichero temporal y el hook de `SessionEnd` los convierte en contadores.

El **modo de permisos** merece una nota: no viene en el payload de la statusline,
pero sí en el transcript, cuya ruta sí llega. Se lee la cola del fichero (32 KB,
0,02 ms) buscando el último `permissionMode`. Es la única forma de que `gremlin`
—«30 turnos con permisos en bypass»— sea alcanzable.

## Cómo se deducen las cosas que el CLI no dice

Dos comidas y cuatro marcas salen de **deducir**, no de un dato que el CLI
exponga. La regla al escribirlas ha sido siempre la misma: **preferir perderse
una comida a inventarse una.**

**Tests en verde.** Tres capas, de más dura a más blanda:

1. Un `is_error` del CLI manda sobre todo: equivale al código de salida y es el
   único dato duro que hay.
2. Los patrones de rojo se buscan **solo en las últimas doce líneas**, que es
   donde va el resumen. Buscarlos en toda la salida hacía que un test llamado
   `test_login_failed` tiñera de rojo una suite verde.
3. No hace falta un patrón de verde. Si el comando era un runner y salió bien,
   cuenta — así entran los runners que no están en la lista.

Para reconocer el runner hay una lista larga (pytest, jest, vitest, go, cargo,
phpunit, rspec, mvn, gradle, dotnet, swift, flutter, mix, make/just/task…), un
último recurso por el **nombre del ejecutable** (`run-tests.sh`, `testear.sh`,
`bin/spec` cuentan; el `test` de shell no, que es una comparación de ficheros), y
`PET_TEST_RUNNERS` para meter un regex propio.

Y lo más importante: el runner tiene que estar **en posición de comando**, no en
cualquier parte del texto. El comando se parte por los operadores de shell
(`;`, `&&`, `||`, `|`) y cada trozo se mira por su principio, saltándose las
asignaciones de entorno y el `sudo`. Sin eso, un `echo "lanza pytest"` o un
`grep -rn "go test"` contaban como suite verde — el mismo fallo que ya tenía la
detección de commits, arreglado aquí de la misma manera.

**Commit hecho.** `git … commit` tiene que estar al principio del comando o
detrás de un operador de shell. Sin ese anclaje, un `grep -rn "git commit"`
contaba como commit.

**Las cuatro marcas deducidas.** Ninguna adivina intenciones: todas miran un
hecho comprobable que se le parece mucho.

| Marca | Lo que pide el diseño | Lo que se mira de verdad |
| --- | --- | --- |
| `bloodhound` | repro antes del fix | un test rojo seguido de uno verde en la misma sesión |
| `oraculo` | planes escritos antes de tocar código | un `TodoWrite` de 3+ tareas antes del primer `Edit`/`Write` |
| `gardener` | docs y limpieza dos días seguidos | `git show --numstat` del commit: mayoría de `.md`/`docs/`, o un borrado grande |
| `sniper` | tareas cerradas con una sola herramienta | herramientas distintas usadas entre dos tareas cerradas |

Son aproximaciones y se equivocan: un rojo por un fallo de red y un verde después
cuentan como repro→fix aunque no arreglaras nada. Pero se equivocan **por lo
bajo** casi siempre, y ninguna se inventa un hecho que no haya ocurrido.

## Los mandos

```bash
pet                    # el panel: nivel, evolución, xp, hambre, racha y la comida de hoy
pet feed               # +3 xp, hambre −2, uno cada cuatro horas
pet cuenta <c> [n]     # suma a un contador de comportamiento
pet record <c> <v>     # guarda el máximo de un contador
```

Instalados como `/pet` y `/feed` desde `scripts/install.sh`.

Para empezar de cero: `rm ~/.claude/pet.json`.

---

Ver también el [README](../../README.md) para las bandas y la paleta, y
[vitals.md](vitals.md) para la otra capa, la del momento.
