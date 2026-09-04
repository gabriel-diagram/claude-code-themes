# El árbol de 97 formas: puertas y umbrales

> **Trabajo de diseño, no implementado.** El código sigue con sus 41 formas. Esto es
> el árbol del lienzo de Claude Design contrastado contra el runtime, con su defecto
> encontrado y el arreglo verificado. Se guarda aquí porque la calibración costó
> medir y no debe perderse con la sesión.
>
> Traspaso completo: <https://claude.ai/code/artifact/e15cd05f-cb01-4e68-a336-85c61100faee>

## El defecto

El lienzo elige la marca del nivel 5 con una carrera sin umbrales: *gana el contador
más alto de los nueve que no ganaron el oficio*. Eso supone diez señales
independientes. No lo son — `internal/pet/feeding.go` las ata:

```
comida "tests"   -> inquisitive, tests      =>  inquisitive = tests + plans
comida "task"    -> inquisitive, plans
comida "commit"  -> methodical, diffs       =>  methodical  = diffs + compacts
comida "compact" -> methodical
pico >=85%       -> impulsive               =>  impulsive  >= ctx_maxed
pico >=95%       -> impulsive, ctx_maxed
```

Una carrera entre una suma y su sumando tiene ganador antes de empezar, y el lienzo
enfrenta `inquisitive` con `tests`/`plans` en cinco oficios y `methodical` con
`diffs` en seis.

**Resultado: 21 de las 42 marcas son inalcanzables**, y con ellas sus 21 títulos —
42 de las 97 formas. Verificado por rejilla exhaustiva sobre las invariantes (~19 M
estados) y por muestreo aleatorio de 600 k, con lista idéntica.

Las 21 muertas: `andamio`, `avalancha`, `cepo`, `cimiento`, `cristal`, `erizo`,
`flecha`, `francotirador`, `fuente`, `grieta`, `incendio`, `injerto`, `jardinero`,
`kraken`, `lienzo`, `lima`, `linterna`, `muelle`, `oráculo`, `relámpago`, `sabueso`.

## El arreglo: dos piezas, ninguna basta sola

Con ritmos medidos, solo umbrales da 34/42 y solo puertas da 39/42. Juntas, 42/42.

**Uno — catorce puertas cambian**, dos por oficio, para que ningún oficio tenga una
suma junto a sus sumandos. No depende de ningún ritmo: quita la dominación por
álgebra.

**Dos — cada marca pide un umbral**, y la carrera pasa a ser de ratios
(`contador / lo que pide`). El umbral no es una puerta que cruzar: es el conversor de
unidades que permite comparar `bypass_turns`, que va por turno, con
`ctx100_sessions`, que sube una vez por sesión reventada.

El lienzo rechaza los umbrales porque «un bicho que no cumpliera ninguno se quedaría
sin oficio». Eso es cierto de una puerta dura y falso de una carrera de ratios: el
máximo de un ratio existe siempre, igual que el de una cuenta.

## Los ritmos, medidos

366 transcripts de `~/.claude/projects`, 31 días, con los patrones de detección del
propio hook y sus cooldowns aplicados.

| Señal | Por semana | Fuente | Alimenta |
| --- | ---: | --- | --- |
| turnos de usuario | 682 | medido · transcripts | — |
| turnos en bypass | 651 | medido · pet.json, 95% del total | `bypass_turns` |
| sesiones | 34 | medido · 134 con marcas de tiempo | — |
| suites verdes (tras cooldown) | 53,3 | medido | `tests` `inquisitive` |
| commits | 46,3 | medido | `diffs` `methodical` |
| sesiones >90 min | 22,8 | medido · 68% | `long_sessions` |
| compactados | 5,4 | medido | `methodical` |
| sesiones <15 min | 5,0 | medido · 15% | `short_sessions` |
| tareas de plan cerradas | 0 | medido · cero TodoWrite en 366 ficheros | `plans` |
| sesiones con pico >=85% | 7 | **estimado** · el pet está a 0 | `impulsive` |
| sesiones con pico >=95% | 3 | **estimado** · el pet está a 0 | `ctx_maxed` |
| sesiones al 100% | 1 | **estimado** · el pet está a 0 | `ctx100_sessions` |

A ese ritmo son **1.398 xp/semana**, así que el nivel 5 (2000 xp) llega en **10
días**. Cada umbral es el valor esperado de su contador al llegar ahí.

**Comprobación cruzada.** `19 diffs / 46,3` = 2,9 días; `23 tests / 53,3` = 3,0;
`12 long_sessions / 22,8` = 3,7. El `pet.json` marca `streak: 3`. Y el bypass sale
por dos caminos independientes que concuerdan al 5%.

## Las 42 marcas

`~~tachado~~` es la puerta del lienzo; en negrita, la nueva.

| oficio | marca | puerta | pide | título |
| --- | --- | --- | ---: | --- |
| refactor | `cirujano` | `ctx_low` | 80 | `bisturí` |
| refactor | `tejedor` | `plans` | 29 | `telar` |
| refactor | `molde` | ~~`methodical`~~ → **`impulsive`** | 10 | `imprenta` |
| refactor | `lima` | `tests` | 76 | `espejo` |
| refactor | `injerto` | ~~`inquisitive`~~ → **`short_sessions`** | 7 | `raíz` |
| refactor | `tijera` | `long_sessions` | 33 | `guillotina` |
| pulcro | `monje` | ~~`methodical`~~ → **`impulsive`** | 10 | `abad` |
| pulcro | `jardinero` | `plans` | 29 | `bosque` |
| pulcro | `fuente` | `diffs` | 66 | `acueducto` |
| pulcro | `cristal` | `tests` | 76 | `prisma` |
| pulcro | `nieve` | `short_sessions` | 7 | `ventisca` |
| pulcro | `lienzo` | ~~`inquisitive`~~ → **`long_sessions`** | 33 | `mural` |
| cazabugs | `sabueso` | `plans` | 29 | `lobo` |
| cazabugs | `exterminador` | ~~`inquisitive`~~ → **`impulsive`** | 10 | `avispa` |
| cazabugs | `cepo` | ~~`diffs`~~ → **`short_sessions`** | 7 | `red` |
| cazabugs | `linterna` | `methodical` | 74 | `faro` |
| cazabugs | `anzuelo` | `ctx_low` | 80 | `arpón` |
| cazabugs | `lupa` | `long_sessions` | 33 | `microscopio` |
| arquitecto | `cartógrafo` | ~~`inquisitive`~~ → **`impulsive`** | 10 | `atlas` |
| arquitecto | `oráculo` | `tests` | 76 | `esfinge` |
| arquitecto | `andamio` | `methodical` | 74 | `catedral` |
| arquitecto | `brújula` | `ctx_low` | 80 | `sextante` |
| arquitecto | `cimiento` | ~~`diffs`~~ → **`short_sessions`** | 7 | `muralla` |
| arquitecto | `maqueta` | `long_sessions` | 33 | `ciudad` |
| velocista | `relámpago` | ~~`diffs`~~ → **`long_sessions`** | 33 | `tormenta` |
| velocista | `francotirador` | `tests` | 76 | `halcón` |
| velocista | `flecha` | `methodical` | 74 | `saeta` |
| velocista | `muelle` | `plans` | 29 | `resorte` |
| velocista | `chispazo` | ~~`impulsive`~~ → **`ctx_maxed`** | 4 | `descarga` |
| velocista | `patín` | `ctx_low` | 80 | `cohete` |
| maratón | `buey` | ~~`diffs`~~ → **`short_sessions`** | 7 | `mamut` |
| maratón | `topo` | `methodical` | 74 | `gusano` |
| maratón | `ancla` | `ctx_low` | 80 | `puerto` |
| maratón | `caravana` | `plans` | 29 | `legión` |
| maratón | `muro` | `tests` | 76 | `bastión` |
| maratón | `reloj` | ~~`inquisitive`~~ → **`ctx_maxed`** | 4 | `calendario` |
| salvaje | `gremlin` | `impulsive` | 10 | `diablo` |
| salvaje | `kraken` | `long_sessions` | 33 | `leviatán` |
| salvaje | `avalancha` | ~~`diffs`~~ → **`ctx100_sessions`** | 2 | `glaciar` |
| salvaje | `erizo` | `short_sessions` | 7 | `espina` |
| salvaje | `incendio` | ~~`methodical`~~ → **`bypass_turns`** | 931 | `volcán` |
| salvaje | `grieta` | `tests` | 76 | `abismo` |

Los otros niveles no cambian: el temperamento sale del más alto de `methodical`,
`inquisitive` e `impulsive`; el oficio, de la carrera entre hermanos del lienzo.
`inquisitive` deja de abrir marcas — era el compuesto que más daño hacía — y entran
dos contadores que el código ya lleva y el lienzo ignoraba, `bypass_turns` y
`ctx100_sessions`, que devuelven a `salvaje` tres puertas con sentido feral.

## Cómo se verificó

- **Alcanzabilidad del defecto** — rejilla exhaustiva (~19 M estados) y muestreo de
  600 k. Misma lista de 21.
- **El arreglo** — 42/42 en dos poblaciones con semillas y modelos de perfil
  distintos. Los repartos coinciden entre ellas (`injerto` 25/25%, `cepo` 22/23%),
  que es lo que separa una calibración de un sobreajuste. Peor reparto: 2,58% de su
  oficio.
- **Alcanzabilidad estricta** — se exigió que cada marca pueda ganar *sin empate*,
  así que el resultado no depende del orden de desempate.

## Lo que no está sólido

- **Cuatro umbrales son estimados**, no medidos: `impulsive`, `ctx_maxed`,
  `ctx100_sessions` y `plans`. Este perfil los tiene a cero y no hay de dónde
  sacarlos. Son los que más se moverán con datos de otro usuario.
- **La dispersión entre usuarios es modelo, no medida.** Los ritmos son sólidos para
  un perfil; cómo varían entre perfiles es una suposición.
- **`gremlin` se lleva el 66% de `salvaje`.** Es inherente: `impulsive` elige la rama
  brasa *y* abre una marca dentro de ella. Se aplana bajándole el umbral, pero
  entonces la rama feral se queda sin marca por defecto. Decisión de diseño abierta.

## El atlas está completo

`internal/pet/testdata/ATLAS-97.json` — las 97 formas con nombre, padre, nota, color
base, rampa de siete pasos y las siete siluetas de cinco filas. Mismo esquema que el
`ATLAS.json` de 41 que usan los tests, que se deja intacto hasta que esto se
implemente.

Extraído del lienzo en cinco trozos (los ficheros completos superan el tope de 256
KiB de `DesignSync.get_file`) y verificado en siete frentes:

| Comprobación | Resultado |
| --- | --- |
| formas | 97 |
| estructura | todas: 7 estados x 5 filas x 9 columnas, rampa de 7 |
| árbol | 1 + 3 temperamentos + 7 oficios + 42 marcas + 42 títulos + 2 secretas |
| contra `ATLAS.json` | 41 comunes, 40 idénticas carácter por carácter |
| siluetas duplicadas | 0 — el lienzo promete «sin dos iguales» |
| rampas distintas | 10 — coincide con «Ten ramps» de `ramps.go:15` |
| variantes | 97 x 7 = 679 = 371 + 308, los dos números del lienzo |

Las tres últimas no se pidieron: son afirmaciones que el lienzo y el código hacían
por su cuenta y que se cumplen sobre datos extraídos por separado.

**La única discrepancia: `diablo` está redibujado.** Misma rampa y mismo color base,
pero tres de sus cinco filas cambian — los cuernos pasan de `^ ╲ ╱ ^` a `^^ ╲ ^^`, la
base de `▝▙▄█▄▟▘` a `▝▙▄▀▄▟▘` y las patas se juntan. Es un cambio de diseño posterior
al código, no un error de extracción.

## Qué queda para implementarlo

Ya no falta ningún dato. Queda el trabajo: `ATLAS.json` pasa a ser el de 97,
`sprites.go`, `ramps.go`, `evolution.go` y `names.go` crecen con las 56 formas
nuevas, los cuatro tests que hoy afirman 41 formas y 287 variantes pasan a 97 y 679,
y hay que migrar los `pet.json` que ya lleven una marca cuya puerta cambia.
