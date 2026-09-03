# El uso del bicho

Qué mide exactamente el bicho de la statusline, y qué hace que pase de *fresh* a
*k.o.*

Esta es **una de las dos capas**. La vida es del momento: sube y baja con el uso
y se recupera al compactar. La otra capa, el progreso —la XP que elige en qué
evoluciona—, está en [evolution.md](evolution.md) y no baja nunca.

## Un solo número

Todo —estado, ojos, patas, cabeza, color— sale de **un número entre 0 y 100**: el
uso. Y el uso es el **cuello más apretado** de tres consumos.

| Cuello | Campo del JSON de la statusline |
| --- | --- |
| contexto | `context_window.used_percentage` |
| límite de 5 horas | `rate_limits.five_hour.used_percentage` |
| límite de 7 días | `rate_limits.seven_day.used_percentage` |

```
uso = max(ctx, 5h, 7d)
```

Ejemplo real: contexto al 36%, cinco horas al 41%, siete días al 13% → **41%** →
`lively`. El que falta no compite: las cuentas por API no reciben `rate_limits`.

## Por qué el peor y no una media ponderada

Aquí hubo una media 50 / 30 / 20 durante varias versiones, y el argumento que la
sostenía era razonable: **el contexto es lo único que puedes gestionar en el
momento** —compactas, cierras la sesión, abres otra—, mientras que los límites
solo avisan. Coger el máximo castiga igual lo accionable y lo que no lo es.

Lo que ese argumento no vio es lo que una media le hace al caso que más importa.
**Una media diluye.** Con la ventana llena del todo y las cuotas ociosas:

```
media:   0.5·100 + 0.3·20 + 0.2·10  =  58   →  "a gusto", turquesa
cuello:  max(100, 20, 10)           = 100   →  k.o.
```

El contexto agotado, sin sitio para trabajar, y el bicho diciendo que está
cómodo. Eso no es una ponderación desafortunada: es el número mintiendo justo
cuando hacía falta que no lo hiciera.

Se notaba en dos sitios más:

- **El k.o. necesitaba una puerta trasera.** `StateFor` llevaba un segundo
  argumento cuyo único trabajo era forzar el k.o. cuando el contexto llegaba al
  100, porque una media de tres números no llega a 100 si no llegan los tres. Con
  el cuello sobra, y se ha ido.
- **Los contadores de la rama brasa estaban muertos por partida doble.**
  `impulsive` se arregló para que lo pagara el *pico* de la sesión y no el
  reventón, pero el pico que leía era el de la media: una ventana llena con las
  cuotas bajas puntuaba 58, y el umbral está en 85. Solo cruzaba quien tenía las
  tres cosas apretadas a la vez.

Esto es lo que medía la **primera versión** del proyecto (`statusline.sh`, commit
`05bf5c7`), y su comentario sigue siendo la frase correcta: *«no finge emociones;
refleja el cuello más apretado»*.

**Lo que se paga a cambio**, y hay que saberlo: si el límite de 7 días va por el
95%, el bicho está `drowning` toda la semana aunque abras la sesión con la
ventana vacía. Deja de responder a lo que estás haciendo ahora mismo y responde a
algo sobre lo que no puedes actuar. Es el precio de no mentir en el otro caso, y
se juzgó el más barato de los dos.

## La curva viene de la primera versión

Los umbrales no son arbitrarios ni se han tocado nunca. Son la curva de comodidad
que dibujaba `statusline.sh`, **cuadrática** —alta y plana abajo, se desploma solo
cerca del tope, porque *un 44% de cuota no es media vida*—, resuelta para el
cuello:

```
vida = 100 · (1 − (cuello/100)²)      →      cuello = 100 · √(1 − vida/100)

vida 95 → 22.36 → cap 22        vida 40 → 77.46 → cap 78
vida 80 → 44.72 → cap 45        vida 20 → 89.44 → cap 89
vida 60 → 63.25 → cap 63
```

La curva sobrevivió intacta a la reescritura; lo único que se había desviado era
su entrada. Hay un test que lo fija
(`TestTheThresholdsAreTheFirstVersionsComfortCurve`).

## La barra de la banda 1 usa esta misma escalera

La barra de contexto se pintaba con una escala propia de tres pasos —verde bajo
60, ámbar bajo 85, rojo por encima— y una paleta propia. El mismo pie pintaba la
misma sesión de dos colores que no coincidían en nada: con el contexto al 100 y
las cuotas bajas, barra roja al lado de un bicho turquesa «a gusto».

Colorearla por el contexto solo se intentó, y **seguía estando mal**: con el
contexto al 22 y la cuota de 5h al 46, la barra salía verde *fresca* al lado de un
bicho turquesa *a gusto*. Una diferencia que nadie ha pedido se lee como un fallo,
no como una lectura.

Así que la barra dice **dos cosas a la vez, a propósito**:

| | qué dice |
| --- | --- |
| su **largo** y su número | el contexto — lo único que puedes gestionar |
| su **color** | el cuello — cómo va la sesión entera, igual que el bicho |

Barra y bicho no pueden discrepar nunca. Y una barra **corta pero oscura** es
justo la lectura que faltaba: la ventana está vacía, te frena otra cosa. Los dos
números de cuota de la banda 1 se pintan con esta misma tabla, así que esa otra
cosa se lee ahí mismo.

## Dónde cae cada estado

| Uso | Estado | Ojos | Cabeza | Patas |
| --- | --- | --- | --- | --- |
| ≤22% | fresh ✦ | `>` `<` | sí | anda |
| ≤45% | lively | `>` `<` | sí | anda |
| ≤63% | easy | `o` `o` | sí | anda |
| ≤78% | sluggish | `▬` `▬` | sí | quieto |
| ≤89% | tired | `_` `_` | hundida | quieto |
| <100% | drowning | `x` `x` | hundida | quieto |
| **cualquiera de los tres al 100%** | k.o. | `x` `x` | hundida | tumbado, patas al aire |

Con **hambre ≥7** los ojos no cambian de forma: se apagan de color. El hambre es
de la otra capa y no toca el estado.

Son **cuatro señales independientes** que se van cayendo en orden: primero los
ojos, luego el paso de las patas, luego la cabeza se hunde y al final la silueta
se tumba. A un vistazo se distingue *tired* de *drowning* sin leer la etiqueta.

**El k.o. tiene su propia puerta.** Al ser una media, exigir el 100% de la media
significaba exigir los tres consumos al 100% *a la vez*: con ctx, 5h y 7d al 100,
90 y 90 la media sale 95, o sea *drowning*. Ese sprite no se veía nunca.

Así que el k.o. salta en cuanto **el contexto llega al 100%**, sin mirar la media.
Es coherente con por qué la media pondera como pondera: el contexto es lo único
que te para de verdad, y cuando se llena da igual cómo estén los otros dos.

## Cuando falta algún dato

`rate_limits` solo llega a las cuentas Pro y Max, y solo después de la primera
respuesta de la API en la sesión; Claude Code además retira cada ventana en
cuanto pasa su `resets_at`.

Si falta alguno de los tres, **su peso se reparte entre los que sí llegan**
(se normaliza sobre los pesos presentes). Con solo el contexto, `uso = ctx`. Si
no llega ninguno, el uso es 0 y el bicho sale fresco — no se inventa un estado.

## Qué NO mide

Ni el **coste en dólares**, ni el **tiempo de sesión**, ni las **líneas tocadas**,
ni el estado de git, ni la caché. Todo eso sale en las bandas de la izquierda,
pero no le afecta al bicho. El uso es solo consumo de cuota y de contexto: lo que
puede pararte.

Tampoco mide el **progreso**. Que el bicho esté *k.o.* no lo devuelve a larva:
la silueta la elige la XP, y la XP no la toca el uso. Un `surgeon` reventado
sigue siendo un surgeon, con cara de haber visto cosas.

## Honestidad

El bicho **no finge emociones**. No se pone contento porque el código compile ni
triste porque falle un test: refleja un número real y comprobable. Si está
cansado, es que la media va por el 85%.

---

Ver también el [README](../../README.md) para las bandas, la paleta y el resto de la
statusline, y [evolution.md](evolution.md) para la otra capa: XP, hambre,
comida y las 41 evoluciones.
