# El runtime

Por qué esto es un binario de Go, a dónde se va el tiempo, y las dos cosas que
tuvieron que arreglarse para que los números de arriba fueran ciertos.

## Por qué Go

Estaba en Python y funcionaba. El problema no era el código —medido, hacía su
trabajo en 1,5 ms— sino lo que cuesta que Python se presente: 5,4 ms de intérprete
más 12,9 de imports, de los cuales 10 eran `subprocess` y `re` con todo lo que
arrastran. Ese peaje se pagaba **una vez por segundo** en la statusline y **en cada
llamada a herramienta** en el hook.

| | Python | Go |
| --- | --- | --- |
| statusline (1 vez/segundo) | 22,4 ms | **1,5 ms** |
| hook, camino lento (`Bash`, `Edit`, `TodoWrite`) | 21,3 ms | **1,7 ms** |
| hook, camino rápido (todo lo demás) | 2,6 ms | **1,5 ms** |
| panel `/pet` | 14,7 ms | **1,4 ms** |

El hook es el que importa: colgaba 21 ms de cada `Bash` y cada `Edit`.

> Esa tabla es la medición de la migración, con las dos columnas tomadas igual.
> Reproducirla hoy solo se puede a medias —la mitad Python ya no existe— y el orden
> de magnitud es lo que aguanta, no el decimal: 200 invocaciones seguidas de
> `ccpet statusline` en WSL2 dan **1,7–1,9 ms de reloj**, de los cuales unos
> **0,45 ms son el coste de arrancar un proceso cualquiera** — un `/bin/true`
> medido en el mismo bucle. WSL2 mide con ruido: una ronda de nueve devolvió un
> delta negativo. Mídelo en tu máquina:
>
> ```bash
> for i in $(seq 200); do COLUMNS=116 ./bin/ccpet-linux-amd64 statusline < payload.json >/dev/null; done
> ```

Con Go desaparecieron dos cosas que solo existían para abaratar Python: el
prefiltro de bash del hook (arrancar el intérprete costaba 15 ms, así que había que
evitarlo) y la purga manual de `sys.path` — un `python3 -c` mete el directorio
actual en la ruta de imports, y un `json.py` cualquiera del repo que tuvieras
abierto secuestraba la statusline. Comprobado: pasaba de verdad.

Queda un `bin/ccpet` de veinte líneas de bash que elige binario por plataforma,
porque `hooks.json` necesita una ruta fija. Usa `$OSTYPE` y `$MACHTYPE`, que bash
rellena solo: `uname` serían dos forks en algo que corre en cada llamada. Y ni
siquiera está en el camino caliente — `ccpet link` deja dos enlaces estables al
binario de tu máquina, y tanto el hook como la statusline van directos. El shim es
el plan B, y de paso repara los enlaces cuando un update del plugin los deja
colgando.

## La mitad de la statusline era `git`

La statusline marcaba 3,5 ms en la primera medición y ahora 1,5, y la diferencia no
es que Go corriera más:

| | |
| --- | --- |
| leer la rama de `.git/HEAD` | **0,9 µs** |
| `git status` para saber si el árbol está sucio (un fork) | **1,1 ms** |
| ese mismo dato, ya en caché | **4,2 µs** |
| el resto del refresco: parsear, medir, componer las cuatro bandas | **3,4 µs** |

Con `refreshInterval: 1` eso era un fork de `git` por segundo y por sesión abierta
para redibujar algo que casi nunca ha cambiado. Así que las dos mitades se separan:
**la rama sale de leer `.git/HEAD`** —sin fork, siempre exacta, y de paso acierta en
detached, en worktree y en un repo sin ningún commit— y solo el asterisco de «árbol
sucio» pasa por `git`, **con tres segundos de caché por repo y sesión**.

Eso es lo único del pie que puede ir con retraso: haces un commit y el ✳ tarda hasta
un refresco largo en apagarse. A cambio, dos de cada tres refrescos no forkean nada.

## Un `pet.json`, muchas ventanas

`~/.claude/pet.json` es **uno solo** para todas tus sesiones y todos tus repos, y el
hook lo toca en **cada llamada a herramienta**. Como Claude Code lanza herramientas
en paralelo, dos escrituras a la vez no son el caso raro: son el caso normal.

Se escribía con `rename`, que es atómico, y eso resuelve un problema distinto del
que había: garantiza que nadie lea un json a medias, y no impide que dos escritores
se pisen. Como cada escritura vuelca el estado **entero**, el que llega tarde
devuelve al fichero todo lo que leyó y deshace al otro.

```
100 comidas en serie      800 xp
100 comidas en paralelo    72 xp    ← 91 % perdido
```

Ahora toda modificación pasa por un candado (`pet.json.lock`, al lado, vacío): se
lee y se escribe dentro de él, así que las 100 comidas dejan los 800 puntos. Es un
`flock` del kernel —`LockFileEx` por kernel32 en Windows, sin dependencias—, no un
fichero centinela, de modo que un proceso que muera a media escritura lo suelta él
solo y no deja nada atascado.

Si el candado no se consigue en dos segundos se escribe igualmente sin él: perder un
punto de xp de vez en cuando es un arañazo, y un hook colgado bloquea la herramienta
que tiene detrás.

El mismo candado protege el registro de herramientas de la sesión, que se leía y se
borraba en dos pasos y perdía lo que llegara entre uno y otro — **17.508 nombres de
24.000 bajo carga**. Eso alimentaba `sniper`, que cuenta cuántas herramientas
distintas usas entre dos tareas cerradas, y le hacía ver tareas de una sola
herramienta que no lo eran.

Dos guardarraíles recorren el código y fallan si alguien vuelve al patrón viejo.

## Los binarios están en el repo, y el CI los comprueba

El plugin se instala clonando este repo, así que las cinco compilaciones viven en
`bin/`. Eso solo funciona si están al día, y por eso el CI recompila y compara.

Para que esa comparación signifique algo la compilación tiene que ser
**reproducible**: `scripts/build.sh` pasa `-trimpath -buildvcs=false`. Sin lo
segundo, Go sella en cada binario el commit y una marca `vcs.modified` con el estado
del árbol — y como `bin/` está versionado:

```
árbol limpio    -> build -> vcs.modified=false dentro del binario
bin/ escrito    -> el árbol ya está sucio
siguiente build -> vcs.modified=true -> otro binario, mismo código
```

Dos compilaciones del mismo código nunca daban el mismo binario, y el trabajo del CI
habría fallado siempre diciendo «bin/ is stale» con `bin/` al día. Nada de ese sello
hace falta aquí: la versión entra por `-X`.

`.github/workflows/ci.yml` corre gofmt, `vet`, `go test -race` en ubuntu y macos, la
sintaxis de los scripts y la validez de los json. Usa `go-version-file: go.mod` y no
`stable`: una compilación de Go es reproducible con el *mismo* toolchain, no con
cualquiera.
