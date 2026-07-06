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
import sys, json, os, re, subprocess
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
if effort:
    s2.append(seg(2, VIOLET+str(effort)+R, color=VIOLET, plain=str(effort)))
rl=[]
if rl5 is not None: rl.append(GRAY+"5h "+R+lvl(rl5)+str(int(round(float(rl5))))+"%"+R)
if rl7 is not None: rl.append(GRAY+"7d "+R+lvl(rl7)+str(int(round(float(rl7))))+"%"+R)
if rl: s2.append(seg(4, " ".join(rl)))
line2=assemble(s2, WIDTH)

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
line3=glow+col+B+cara+R+" "+lifebar(vida,col)+" "+col+clbl+R
if vida<=0:
    line3+=SC+" │ "+R+GRAY+peor[1]+" 100%"+R     # qué me mató
elif vida<40:
    line3+=SC+" │ "+R+GRAY+"cuello "+peor[1]+R    # qué aprieta

print(line1); print(line2); print(line3)
'
