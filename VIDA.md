# La vida del bicho

Qué mide exactamente la barra de vida de la statusline, y qué hace que el bicho
pase de *fresca* a *k.o.*

## Los tres cuellos

La vida sale de **un solo número**: el peor de tres cuellos de botella, todos en
tanto por ciento de consumo.

| Cuello | Campo del JSON de la statusline | Qué es |
| --- | --- | --- |
| `ctx` | `context_window.used_percentage` | cuánto llevas gastado de la ventana de contexto |
| `5h` | `rate_limits.five_hour.used_percentage` | cuánto llevas del límite de 5 horas |
| `7d` | `rate_limits.seven_day.used_percentage` | cuánto llevas del límite de 7 días |

Se coge el **máximo** de los tres. No una media: manda el que más aprieta, porque
es el que te va a parar.

## La fórmula

```
vida = 100 · (1 − (peor / 100)²)          acotada a 0–100
```

Curva cuadrática, no una resta.

## Por qué la curva y no `100 − peor`

Porque haber gastado el 44% de tu cuota **no es media vida**: te queda margen de
sobra y el bicho no debería estar ya medio muerto. La cuadrática lo mantiene
arriba mientras hay holgura y lo desploma solo cuando el cuello se acerca al tope.

| Peor cuello | Vida |
| --- | --- |
| 0% | 100 |
| 10% | 99 |
| 20% | 96 |
| 30% | 91 |
| 40% | 84 |
| 50% | 75 |
| 60% | 64 |
| 70% | 51 |
| 80% | 36 |
| 90% | 19 |
| 95% | 10 |
| 99% | 2 |
| 100% | 0 |

Fíjate en el tramo de arriba: del 0 al 40% de consumo la vida solo baja 16
puntos. Del 80 al 100%, baja 36. El aviso llega cuando importa.

## Dónde cae cada estado

| Peor cuello | Vida | Estado |
| --- | --- | --- |
| hasta 22,4% | ≥95 | fresca |
| hasta 44,7% | ≥80 | vibrante |
| hasta 63,2% | ≥60 | a gusto |
| hasta 77,5% | ≥40 | espesa |
| hasta 89,4% | ≥20 | cansada |
| más de 89,4% | <20 | ahogada |
| **exactamente 100%** | 0 | k.o. |

**El k.o. es literal.** Hace falta que un límite llegue al 100% de verdad. Al 99%
todavía le quedan 2 puntos de vida y el bicho sigue de pie (hundido, pero de pie).

## Qué NO mide

Ni el **coste en dólares**, ni el **tiempo de sesión**, ni las **líneas tocadas**,
ni el estado de git. Todo eso sale en la statusline, pero no le afecta al bicho.
La vida es solo consumo de cuota y de contexto: lo que puede pararte.

## Qué te dice cuando aprieta

- **Vida < 40** — aparece a la izquierda de la fila 4 qué cuello es el que aprieta:
  `cuello: ctx`.
- **Vida = 0** — aparece qué te mató: `✖ ctx al 100%`.

Así no tienes que adivinar cuál de los tres porcentajes es el problema.

## Cuando no hay datos

`rate_limits` solo llega a las cuentas Pro y Max, y solo después de la primera
respuesta de la API en la sesión. Claude Code además retira cada ventana en cuanto
pasa su `resets_at`. Si no llega ninguno de los tres campos, la vida se queda en
**100** y el bicho sale fresco — no se inventa un estado.

## Honestidad

El bicho **no finge emociones**. No se pone contento porque el código compile ni
triste porque falle un test: refleja un número real y comprobable. Si está
cansado, es que un límite va por el 85%. Si hace k.o., es que algo llegó al 100%
y te dice cuál.

---

Ver también el [README](README.md) para el resto de la statusline y las caras del
sprite.
