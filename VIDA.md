# El uso del bicho

Qué mide exactamente el bicho de la statusline, y qué hace que pase de *fresca* a
*k.o.*

Esta es **una de las dos capas**. La vida es del momento: sube y baja con el uso
y se recupera al compactar. La otra capa, el progreso —la XP que elige en qué
evoluciona—, está en [EVOLUCIONES.md](EVOLUCIONES.md) y no baja nunca.

## Un solo número

Todo —estado, ojos, patas, cabeza, color— sale de **un número entre 0 y 100**: el
uso. Y el uso es una **media ponderada** de tres consumos.

| Peso | Cuello | Campo del JSON de la statusline |
| --- | --- | --- |
| **0.5** | contexto | `context_window.used_percentage` |
| **0.3** | límite de 5 horas | `rate_limits.five_hour.used_percentage` |
| **0.2** | límite de 7 días | `rate_limits.seven_day.used_percentage` |

```
uso = 0.5 · ctx  +  0.3 · 5h  +  0.2 · 7d
```

Ejemplo real: contexto al 36%, cinco horas al 41%, siete días al 13%.

```
0.5·36 + 0.3·41 + 0.2·13  =  18 + 12,3 + 2,6  =  32,9  →  33%  →  vibrante
```

## Por qué ponderada y no el peor de los tres

Porque **el contexto es lo único que puedes gestionar en el momento**. Si se te
llena, compactas, cierras la sesión o abres otra. Los límites de 5 horas y 7 días
solo avisan: no hay nada que hacer con ellos salvo esperar.

Coger el máximo de los tres castigaría igual una cosa accionable que una que no
lo es. Que el límite de 7 días vaya por el 80% no debería poner al bicho al borde
de la muerte si acabas de abrir la sesión con el contexto vacío.

Por eso el reparto 50 / 30 / 20: manda lo que puedes tocar, y lo demás pondera.

## Dónde cae cada estado

| Uso | Estado | Ojos | Cabeza | Patas |
| --- | --- | --- | --- | --- |
| ≤22% | fresca ✦ | `>` `<` | sí | anda |
| ≤45% | vibrante | `>` `<` | sí | anda |
| ≤63% | a gusto | `o` `o` | sí | anda |
| ≤78% | espesa | `▬` `▬` | sí | quieto |
| ≤89% | cansada | `_` `_` | hundida | quieto |
| <100% | ahogada | `x` `x` | hundida | quieto |
| **contexto al 100%** | k.o. | `x` `x` | hundida | tumbado, patas al aire |

Con **hambre ≥7** los ojos no cambian de forma: se apagan de color. El hambre es
de la otra capa y no toca el estado.

Son **cuatro señales independientes** que se van cayendo en orden: primero los
ojos, luego el paso de las patas, luego la cabeza se hunde y al final la silueta
se tumba. A un vistazo se distingue *cansada* de *ahogada* sin leer la etiqueta.

**El k.o. tiene su propia puerta.** Al ser una media, exigir el 100% de la media
significaba exigir los tres consumos al 100% *a la vez*: con ctx, 5h y 7d al 100,
90 y 90 la media sale 95, o sea *ahogada*. Ese sprite no se veía nunca.

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
la silueta la elige la XP, y la XP no la toca el uso. Un `cirujano` reventado
sigue siendo un cirujano, con cara de haber visto cosas.

## Honestidad

El bicho **no finge emociones**. No se pone contento porque el código compile ni
triste porque falle un test: refleja un número real y comprobable. Si está
cansado, es que la media va por el 85%.

---

Ver también el [README](README.md) para las bandas, la paleta y el resto de la
statusline, y [EVOLUCIONES.md](EVOLUCIONES.md) para la otra capa: XP, hambre,
comida y las 27 evoluciones.
