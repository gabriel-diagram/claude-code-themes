#!/bin/bash
# Statusline — estética cyberpunk anónima (synthwave apagado).
# PALETA ADAPTATIVA por entorno: AZUL en local (Mac), ROJO en el docker claude-work.
#   Detección: /.dockerenv o HOME del contenedor. Override manual: STATUSLINE_PROFILE=docker|local
# Líneas ADAPTATIVAS: truncan al ancho real de la terminal y sueltan elementos de menor
# prioridad cuando no caben, para que NUNCA haga wrap (que descuadraba la caja del prompt).
# Lee JSON de sesión por stdin (python3, sin jq). git opcional, degrada con elegancia.
#  L1: ▍modelo · estilo · dir (rama✷) · VIM     L2: [barra ctx] % · $coste · +/- líneas · effort · 5h/7d
#  L3: el bicho — cara + barra de vida del estado real de sesión (vida = peor cuello: ctx / 5h / 7d)
python3 -c '
import sys, json, os, re, subprocess, time
try:
    d = json.load(sys.stdin)
except Exception:
    d = {}

R="\033[0m"; B="\033[1m"
# --- PERFIL segun entorno: LOCAL (azulado) vs DOCKER claude-work (rojizo) ------------
# Deteccion robusta; STATUSLINE_PROFILE=docker|local fuerza el perfil a mano.
_prof=os.environ.get("STATUSLINE_PROFILE")
if _prof in ("docker","local"):
    DOCKER = (_prof=="docker")
else:
    DOCKER = os.path.exists("/.dockerenv") or os.environ.get("container") is not None

if DOCKER:
    # ROJIZO — synthwave calido (coral/salmon/rojo). Marca "estoy en el docker work".
    CYAN  ="\033[38;5;209m"  # coral calido       (primario / a salvo)
    VIOLET="\033[38;5;203m"  # salmon             (meta / secundario / warn)
    PINK  ="\033[38;5;196m"  # rojo puro          (alerta / negativo)
    GREEN ="\033[38;5;78m"   # verde-teal         (positivo, contraste +lineas)
    STEEL ="\033[38;5;138m"  # malva calido       (dir)
    GRAY  ="\033[38;5;240m"  # gris               (coste, tenue)
    SC    ="\033[38;5;237m"  # separador muy tenue
else:
    # AZULADO — todo azul (cian -> azure -> azul profundo). Local.
    CYAN  ="\033[38;5;51m"   # cian brillante     (primario / a salvo)
    VIOLET="\033[38;5;39m"   # azure              (meta / secundario / warn)
    PINK  ="\033[38;5;27m"   # azul profundo      (alerta / negativo)
    GREEN ="\033[38;5;78m"   # verde-teal         (positivo)
    STEEL ="\033[38;5;109m"  # azul acero         (dir)
    GRAY  ="\033[38;5;240m"  # gris               (coste, tenue)
    SC    ="\033[38;5;237m"  # separador muy tenue

SEP=SC+"│"+R             # separador (se rodea de espacios al unir)
ANSI=re.compile("\033\\[[0-9;]*m")

def g(o,*ks):
    for k in ks:
        o = (o or {}).get(k) if isinstance(o,dict) else None
    return o

def vis(s):                       # ancho visible (sin secuencias ANSI)
    return len(ANSI.sub("", s))

# --- ancho de la terminal (cascada robusta; -1 de margen anti auto-wrap) -----------
def term_width():
    w=g(d,"terminal","width") or g(d,"width")
    try:
        if w: return int(w)
    except Exception: pass
    c=os.environ.get("COLUMNS")
    if c and c.isdigit(): return int(c)
    try:
        return os.get_terminal_size(os.open("/dev/tty", os.O_RDONLY)).columns
    except Exception: pass
    return 80
WIDTH=max(20, term_width()-1)

# --- MASCOTA: reserva de hueco a la derecha de L2/L3 --------------------------------
# La L1 NO se toca: ahi Claude Code pinta sus badges alineados a la derecha.
MASCOT_W=8; MASCOT_GAP=2
MASCOT=(os.environ.get("STATUSLINE_MASCOT","1").lower() not in ("0","off","no")
        and WIDTH>=62)
CW=WIDTH-(MASCOT_W+MASCOT_GAP) if MASCOT else WIDTH

# --- ensamblado adaptativo ----------------------------------------------------------
# seg = {p:prioridad_caida(mayor cae antes), txt:con_ansi, vis:ancho, sep:" "|SEP_espaciado,
#        color:prefijo_para_truncar|None, plain:texto_plano|None}
SEPX=" "+SEP+" "                  # separador normal con espacios
def assemble(segs, width):
    items=[s for s in segs if s]
    def total(its):
        return sum(it["vis"] + (0 if i==0 else it["sepw"]) for i,it in enumerate(its))
    while len(items)>1 and total(items)>width:
        v=max(range(len(items)), key=lambda i:(items[i]["p"], i))
        items.pop(v)
    if len(items)==1 and items[0]["vis"]>width:      # último recurso: truncar
        it=items[0]
        if it.get("color") is not None and it.get("plain") is not None:
            keep=max(1, width-1); t=it["plain"][:keep]+"…"
            it=dict(it, txt=it["color"]+t+R, vis=vis(it["color"]+t+R))
        else:                                         # compuesto: corte crudo sin color
            t=ANSI.sub("", it["txt"])[:max(1,width-1)]+"…"
            it=dict(it, txt=t, vis=len(t))
        items=[it]
    out=""
    for i,it in enumerate(items):
        out += ("" if i==0 else it["sep"]) + it["txt"]
    return out

def seg(p, txt, color=None, plain=None, sep=SEPX):
    return {"p":p, "txt":txt, "vis":vis(txt), "sep":sep, "sepw":vis(sep), "color":color, "plain":plain}

# --- datos --------------------------------------------------------------------------
model=g(d,"model","display_name") or "?"
cwd=g(d,"workspace","current_dir") or d.get("cwd") or os.getcwd()
home=os.path.expanduser("~")
ddir="~"+cwd[len(home):] if cwd.startswith(home) else cwd
parts=ddir.rstrip("/").split("/")
short="/".join(parts[-2:]) if len(parts)>2 else ddir
style=g(d,"output_style","name")
vim=g(d,"vim","mode")
effort=g(d,"effort","level")
cost_usd=g(d,"cost","total_cost_usd")
added=g(d,"cost","total_lines_added")
removed=g(d,"cost","total_lines_removed")
pct=g(d,"context_window","used_percentage")
rl5=g(d,"rate_limits","five_hour","used_percentage")
rl7=g(d,"rate_limits","seven_day","used_percentage")

branch=None; dirty=""
try:
    b=subprocess.run(["git","-C",cwd,"branch","--show-current"],capture_output=True,text=True,timeout=1).stdout.strip()
    if b:
        branch=b
        st=subprocess.run(["git","-C",cwd,"status","--porcelain"],capture_output=True,text=True,timeout=1).stdout.strip()
        dirty=(PINK+" ✷"+R) if st else ""
except Exception:
    pass

# --- MASCOTA: el bicho de Claude ----------------------------------------------------
# 2 filas de texto = 3 de pixel (medios bloques). El cuerpo se pinta con FONDO de color
# y los ojos negros encima, para que salga la silueta solida del logo y no bloques
# sueltos. Color = acento del tema (azul electrico en local, coral del logo en docker);
# STATUSLINE_MASCOT_COLOR=<0-255> lo fuerza, STATUSLINE_MASCOT=0 lo apaga.
# Color: se clava el acento del tema (#2e8bff Electric Blue) si hay truecolor; si no,
# cae al 256 mas cercano. STATUSLINE_MASCOT_COLOR acepta "#rrggbb" o un indice 0-255.
_TRUE=os.environ.get("COLORTERM","").lower() in ("truecolor","24bit")
def _c(spec, fg=True):
    lead=38 if fg else 48
    if isinstance(spec,tuple):
        if _TRUE: return "\033[%d;2;%d;%d;%dm"%((lead,)+spec[0])
        return "\033[%d;5;%dm"%(lead,spec[1])
    return "\033[%d;5;%dm"%(lead,spec)
# (rgb, fallback256)
CORAL =((0xda,0x77,0x56),173)   # coral del logo  -> perfil docker
AZUL  =((0x2e,0x8b,0xff), 33)   # claude / Electric Blue -> perfil local
ROJO  =((0xff,0x4b,0x3e),203)   # error del tema  -> agonizando
GRIS  =((0x5a,0x5a,0x5a),240)   # k.o.
_mc=os.environ.get("STATUSLINE_MASCOT_COLOR","").strip()
if _mc.startswith("#") and len(_mc)==7:
    try: ACC=((int(_mc[1:3],16),int(_mc[3:5],16),int(_mc[5:7],16)),33)
    except ValueError: ACC=CORAL if DOCKER else AZUL
elif _mc.isdigit(): ACC=int(_mc)
else: ACC=CORAL if DOCKER else AZUL

def mascot(v):
    t=int(time.time())                      # 1 frame/s: es el techo real de la statusline
    col=GRIS if v<=0 else (ROJO if v<20 else ACC)
    if   v<=0:  e1,e2="\u2716","\u2716"           # k.o.
    elif v>=80: e1,e2=">","<"               # los ojos del logo
    elif v>=60: e1,e2="\u203a","\u2039"           # mas relajados
    elif v>=40: e1,e2="-","-"               # entrecerrados
    elif v>=20: e1,e2="~","~"               # fundido
    else:       e1,e2="\u00d7","\u00d7"           # agonizando
    if v>=40 and t%6==0: e1,e2="-","-"      # parpadeo: 1 de cada 6 frames
    FGC=_c(col); BGC=_c(col,False); EYE="\033[38;5;16m"
    top=FGC+"\u2590"+R+BGC+EYE+" "+e1+"  "+e2+" "+R+FGC+"\u258c"+R   # orejas + cuerpo + ojos
    # Anda a ratos, no sin parar: 3 pasos cada 12 s. Un baile continuo en la esquina
    # del ojo distrae mas que decora. STATUSLINE_MASCOT_WALK=1 lo pone a andar siempre.
    _q=t%12; _paso=(_q<3) or os.environ.get("STATUSLINE_MASCOT_WALK","")=="1"
    if   v<=0:  legs="  \u2580\u2580\u2580\u2580  "          # tumbado, sin patas
    elif _paso: legs=" \u2588\u2588  \u2580\u2580 " if t%2 else " \u2580\u2580  \u2588\u2588 "
    else:       legs=" \u2580\u2580  \u2580\u2580 "          # quieto, de pie
    return top, FGC+legs+R                  # patas alternas = anda; planas = muerto

def padr(s,w):
    n=w-vis(s)
    return s+(" "*n) if n>0 else s

# --- L1: modelo · estilo · dir (rama) · vim -----------------------------------------
s1=[seg(0, CYAN+B+"▍"+str(model)+R, color=CYAN+B, plain="▍"+str(model))]
if style: s1.append(seg(3, VIOLET+str(style)+R, color=VIOLET, plain=str(style)))
s1.append(seg(1, STEEL+short+R, color=STEEL, plain=short))   # dir (truncable)
if branch:                                                    # rama pegada al dir (sep " ")
    rt=VIOLET+"("+branch+(dirty or "")+VIOLET+")"+R
    s1.append(seg(2, rt, sep=" "))
if vim:
    vc=CYAN if vim=="INSERT" else VIOLET
    s1.append(seg(4, vc+B+str(vim)+R, color=vc+B, plain=str(vim)))
line1=assemble(s1, WIDTH)

# --- L2: barra% · coste · +/-líneas · effort · 5h/7d --------------------------------
def bar(p,w=16):
    p=max(0,min(100,int(p)))
    f=round(w*p/100.0)
    c=CYAN if p<60 else (VIOLET if p<85 else PINK)
    return c+("█"*f)+SC+("░"*(w-f))+R+" "+c+str(p)+"%"+R
def lvl(p): return CYAN if p<60 else (VIOLET if p<85 else PINK)

s2=[seg(0, bar(pct) if pct is not None else SC+("░"*16)+R+" --%")]
if cost_usd is not None:
    s2.append(seg(3, GRAY+"$"+format(float(cost_usd),".2f")+R, color=GRAY, plain="$"+format(float(cost_usd),".2f")))
if added is not None or removed is not None:
    s2.append(seg(1, GREEN+"+"+str(added or 0)+R+SC+"/"+R+PINK+"-"+str(removed or 0)+R))
s2b=[]
if effort:
    s2b.append(seg(2, VIOLET+str(effort)+R, color=VIOLET, plain=str(effort)))
rl=[]
if rl5 is not None: rl.append(GRAY+"5h "+R+lvl(rl5)+str(int(round(float(rl5))))+"%"+R)
if rl7 is not None: rl.append(GRAY+"7d "+R+lvl(rl7)+str(int(round(float(rl7))))+"%"+R)
if rl: s2b.append(seg(4, " ".join(rl)))
s2a=s2                      # con tarjeta: fila 2 = contexto/coste, fila 3 = effort/limites
s2=s2+s2b                   # sin tarjeta: todo junto en una sola fila, como antes

# --- L3: el bicho — cara del estado REAL de sesion (vida = 100 - peor cuello) --------
# Honesto: no finge emociones; refleja el cuello mas apretado (ctx / rate 5h / 7d).
# Rango con nacimiento (fresca) y muerte real (k.o. cuando un limite llega a 100%).
def face(v):                           # cara + etiqueta + COLOR, del mismo tier (neon vivo, sin apagados)
    if v<=0:   return "✖_✖","k.o.",PINK      # muerte real: un limite al 100%, no puedo seguir
    if v>=95:  return "◕▿◕","fresca",GREEN   # nacimiento: sesion recien empezada
    if v>=80:  return "◕‿◕","vibrante",GREEN
    if v>=60:  return "•‿•","a gusto",CYAN   # cian vivo (el que tenias y te gustaba)
    if v>=40:  return "·_·","espesa",VIOLET
    if v>=20:  return "¬_¬","cansada",VIOLET
    return "×_×","ahogada",PINK
def lifebar(v,col,w=6):
    f=round(w*v/100.0)
    if v>0: f=max(1,f)                 # min 1 bloque hasta que muere del todo
    return col+("▰"*f)+SC+("▱"*(w-f))+R
usos=[]
for x,n in ((pct,"ctx"),(rl5,"5h"),(rl7,"7d")):
    try:
        if x is not None: usos.append((float(x),n))
    except (TypeError,ValueError): pass
if usos:
    # vida NO lineal: el margen no duele hasta cerca del tope (un 44% de cuota no es media vida).
    # curva cuadratica — alta y plana abajo, se desploma solo cuando el cuello se acerca al 100%.
    peor=max(usos,key=lambda t:t[0]); vida=int(round(100.0*(1.0-(peor[0]/100.0)**2)))
else:
    vida,peor=100,(0.0,"—")
vida=max(0,min(100,vida))
cara,clbl,col=face(vida)
glow=(GRAY+"✦ "+R) if vida>=95 else ""    # destello solo al nacer
# qué aprieta / qué me mató: sale a la izquierda, no dentro de la tarjeta
if   vida<=0: cuello=GRAY+"\u2716 "+peor[1]+" al 100%"+R
elif vida<40: cuello=GRAY+"cuello: "+peor[1]+R
else:         cuello=""

# --- LA TARJETA: estado arriba, sprite en medio, barra de vida abajo ----------------
# 4 filas de 8 celdas ancladas a la derecha de las filas 2-5. La 1 no, que ahi Claude
# Code alinea sus badges. Se centra todo al ancho del sprite.
def ctr(plain, colored, w=MASCOT_W):
    n=w-len(plain)
    if n<0: return colored
    return " "*(n//2)+colored+" "*(n-n//2)

izq=[assemble(s2a, CW), assemble(s2b, CW), cuello, ""]
if MASCOT and all(vis(x)<=CW for x in izq):
    etq=("\u2726 "+clbl) if vida>=95 else clbl       # el destello cabe justo en 8
    _t,_b=mascot(vida)
    tarjeta=[ctr(etq, col+B+etq+R), _t, _b, ctr("\u25b0"*6, lifebar(vida,col))]
    print(line1)
    for _l,_c in zip(izq,tarjeta):
        print(padr(_l,CW)+" "*MASCOT_GAP+_c)
else:                                            # terminal estrecha: el diseño de 3 filas
    line2=assemble(s2, WIDTH)
    line3=glow+col+B+cara+R+" "+lifebar(vida,col)+" "+col+clbl+R
    if vida<=0:   line3+=SC+" \u2502 "+R+GRAY+peor[1]+" 100%"+R
    elif vida<40: line3+=SC+" \u2502 "+R+GRAY+"cuello "+peor[1]+R
    print(line1); print(line2); print(line3)
'
