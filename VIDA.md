# El uso del bicho

Qué mide exactamente el bicho de la statusline, y qué hace que pase de *fresca* a
*k.o.*

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
| ≤63% | a gusto | `●` `●` | sí | anda |
| ≤78% | espesa | `▬` `▬` | sí | quieto |
| ≤89% | cansada | `◠` `◠` | hundida | quieto |
| <100% | ahogada | `✕` `✕` | hundida | quieto |
| **100% exacto** | k.o. | `✕` `✕` | hundida | tumbado, patas al aire |

Son **cuatro señales independientes** que se van cayendo en orden: primero los
ojos, luego el paso de las patas, luego la cabeza se hunde y al final la silueta
se tumba. A un vistazo se distingue *cansada* de *ahogada* sin leer la etiqueta.

**El k.o. es literal.** Al ser una media, hace falta el 100% clavado: con el
contexto al 100% pero los límites a cero, el uso sale 50 y el bicho está *a
gusto*. El k.o. de verdad es que se te haya acabado todo a la vez.

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

## Honestidad

El bicho **no finge emociones**. No se pone contento porque el código compile ni
triste porque falle un test: refleja un número real y comprobable. Si está
cansado, es que la media va por el 85%.

---

Ver también el [README](README.md) para las bandas, la paleta y el resto de la
statusline.
