#!/usr/bin/env python3
"""Generate original vector food artwork and local PNG exports. Never opens a DB.
Artwork: CC0-1.0. Standard-library Python; PNG export uses frontend Sharp development tooling.
"""
import html
import json
from pathlib import Path
import random
import subprocess

HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[1]
PUBLIC = ROOT / 'frontend/public/demo-images'
SOURCES = HERE / 'image-sources'
CATALOG = json.loads((HERE / 'catalog.json').read_text())


def ellipse(x, y, rx, ry, fill, extra=''):
    return f'<ellipse cx="{x}" cy="{y}" rx="{rx}" ry="{ry}" fill="{fill}" {extra}/>'


def rect(x, y, w, h, fill, radius=8, extra=''):
    return f'<rect x="{x}" y="{y}" width="{w}" height="{h}" rx="{radius}" fill="{fill}" {extra}/>'


def path(d, fill, extra=''):
    return f'<path d="{d}" fill="{fill}" {extra}/>'


def stroke(d, color, width=7):
    return path(d, 'none', f'stroke="{color}" stroke-width="{width}" stroke-linecap="round" stroke-linejoin="round"')


def art(item):
    key = item['key']
    category = item['category']
    rng = random.Random(key)
    s = ''
    if category == 'Burgers':
        s += ellipse(400, 446, 183, 24, '#9e6b3e')
        s += rect(220, 389, 360, 52, '#c47b30', 25)
        s += ellipse(400, 391, 180, 28, '#f3c274')
        s += rect(208, 350, 384, 42, '#60352a' if key != 'plant-burger' else '#766236', 21)
        s += path('M211 348Q235 322 259 343T308 340T357 340T405 340T453 340T501 340T550 340L586 350L570 365L230 365Z', '#6d9941')
        s += path('M224 329L286 313L390 319L456 309L579 327L496 348L448 331L360 349L301 330Z', '#f5b332')
        s += rect(214, 288, 372, 42, '#7a3e28' if key != 'crispy-chicken' else '#c27b30', 21)
        s += ellipse(400, 286, 175, 18, '#d35335')
        if key == 'smoky-bbq':
            s += stroke('M242 281Q280 251 321 279T398 274T478 274T552 275', '#a94330', 15)
        s += path('M215 260C215 145 580 145 585 260Q565 284 400 284T215 260Z', 'url(#bread)')
        for _ in range(24):
            x, y = rng.randint(260, 540), rng.randint(204, 252)
            s += ellipse(x,y,6,2,'#fff0bf',f'transform="rotate(-25 {x} {y})"')
    elif key in ('skin-on-fries','loaded-fries','garlic-bread'):
        if key == 'garlic-bread':
            for x,y in [(290,285),(410,320),(500,262)]:
                s += ellipse(x,y,78,110,'#b2763c',f'transform="rotate(25 {x} {y})"')
                s += ellipse(x,y-8,69,98,'#f7df98',f'transform="rotate(25 {x} {y})"')
                for _ in range(15): s += ellipse(x+rng.randint(-40,40),y+rng.randint(-60,60),4,3,'#698541')
        else:
            for _ in range(35):
                x,y = rng.randint(235,530),rng.randint(220,365)
                s += rect(x,y,22,rng.randint(85,145),'#edb64e',5,f'transform="rotate({rng.randint(-40,40)} {x} {y})"')
                s += rect(x+3,y+5,5,70,'#f9d581',2)
            if key == 'loaded-fries':
                s += stroke('M260 310Q370 245 536 315M270 360Q420 290 555 353', '#f9d048', 23)
                for _ in range(22): s += ellipse(rng.randint(280,530),rng.randint(280,420),5,4,'#4e7d40')
    elif category == 'Wings':
        for x,y,angle in [(285,285,-15),(395,255,15),(505,295,35),(305,390,-30),(420,370,20),(520,395,40)]:
            s += path('M-40-22Q-38-60 0-45Q40-68 62-22Q75 15 32 32L-33 40Q-75 20-40-22Z', '#b84d26' if key=='buffalo-wings' else '#813b25',f'transform="translate({x} {y}) rotate({angle})"')
            s += stroke(f'M{x-23} {y-13}q25-18 48 0', '#e69b50',5)
        s += ellipse(560,230,45,35,'#f7edcd')
        s += ellipse(560,230,32,24,'#fff8e6')
    elif category == 'Pizzas':
        s += ellipse(400,320,196,169,'#ba773b')+ellipse(400,311,185,159,'#edbf76')
        s += ellipse(400,307,165,139,'#be5137')+ellipse(400,305,155,130,'#f3d68e')
        for _ in range(28):
            x,y=rng.randint(270,530),rng.randint(210,400)
            if ((x-400)/145)**2+((y-305)/120)**2>1: continue
            if key=='pepperoni': s+=ellipse(x,y,19,16,'#a94030')+ellipse(x-3,y-3,12,10,'#bc5540')
            elif key=='roast-veg': s+=ellipse(x,y,19,11,rng.choice(['#785845','#738e49','#d16438']))
            else: s+=ellipse(x,y,19,11,'#fff0cc')
        for x,y in [(330,265),(445,360),(490,265),(355,365)]: s+=path('M0 0Q-30-38 10-44Q35-15 0 0','#568644',f'transform="translate({x} {y}) rotate({x%90})"')
        s+=stroke('M400 305L542 412M400 305L243 397M400 305L416 160','#b88141',2)
    elif category == 'Pasta' or key=='lasagne':
        if key=='lasagne':
            s+=path('M260 245L470 204L565 310L347 362Z','#f4d296')
            s+=path('M260 245L347 362L347 430L260 317Z','#d0874e')+path('M347 362L565 310L565 380L347 430Z','#ac4a32')
            s+=stroke('M273 288L347 388L551 340M279 310L347 411L550 364','#f0cb87',11)
        else:
            s+=ellipse(400,330,180,120,'#fbf2d8')
            for _ in range(70):
                x,y=rng.randint(260,540),rng.randint(235,405)
                if ((x-400)/160)**2+((y-320)/100)**2>1:continue
                color='#eec875' if key=='alfredo' else '#dd8146'
                if key=='arrabbiata':s+=rect(x,y,43,15,color,5,f'transform="rotate({rng.randint(-70,70)} {x} {y})"')
                else:s+=stroke(f'M{x} {y}q-35-24-45 6t50 9t-10 25',color,8)
            for _ in range(16):
                x,y=rng.randint(300,500),rng.randint(265,380)
                s+=ellipse(x,y,13,8,'#9c6143' if key=='bolognese' else '#d8b68b')
    elif category in ('Biryani','Curries') or key=='pilau-rice':
        s+=ellipse(400,330,193,147,'#a3a294')+ellipse(400,321,187,141,'#fff8e4')
        fill='#6e8442' if key=='palak-paneer' else ('#bc6437' if category=='Curries' else '#e8c785')
        s+=ellipse(400,318,166,119,fill)
        if category!='Curries':
            for _ in range(200):
                x,y=rng.randint(240,560),rng.randint(205,430)
                if ((x-400)/157)**2+((y-318)/110)**2<1:s+=ellipse(x,y,7,2,rng.choice(['#fff0bf','#d7954d','#f8dd9e']),f'transform="rotate({rng.randint(-80,80)} {x} {y})"')
        if key!='pilau-rice':
            for x,y in [(330,275),(425,260),(485,320),(400,365),(300,350)]:
                if key=='chana-masala':
                    for dx,dy in [(-10,0),(15,10),(0,-15)]:s+=ellipse(x+dx,y+dy,14,12,'#dfb26a')
                else:s+=rect(x-20,y-17,42,32,'#f1d79c' if key=='palak-paneer' else '#ad6c40',9,f'transform="rotate(15 {x} {y})"')
        for x,y in [(370,265),(480,350),(335,370),(450,290)]:s+=path('M0 0q-20-25-27-4q10 22 27 4q22-24 25-4q-8 20-25 4','#477343',f'transform="translate({x} {y})"')
    elif category=='Breads':
        for x,y,angle in [(340,290,-20),(455,345,20)]:
            s+=path('M-95 90Q-120-30-45-105Q40-140 102-25Q130 65 10 112Z','url(#bread)',f'transform="translate({x} {y}) rotate({angle})"')
            for _ in range(28):s+=ellipse(x+rng.randint(-55,55),y+rng.randint(-50,60),rng.randint(3,12),3,rng.choice(['#9d6c3c','#66894c']))
    elif category=='Wraps':
        for x,y,angle in [(340,320,-30),(475,335,25)]:
            s+=rect(x-58,y-118,116,228,'#edd1a0',30,f'transform="rotate({angle} {x} {y})"')
            s+=ellipse(x,y-112,56,31,'#b58a57',f'transform="rotate({angle} {x} {y})"')
            for _ in range(20):s+=ellipse(x+rng.randint(-38,38),y-110+rng.randint(-16,16),8,5,rng.choice(['#c66440','#5c8641','#eee1b4','#9c6c3b']))
            s+=stroke(f'M{x-35} {y-55}l70 17m-70 35l70 17m-70 35l70 17','#c29c62',3)
    elif category=='Grills':
        s+=ellipse(400,330,180,120,'#e7c98a')
        for i in range(3 if key=='mixed-grill' else 2):
            x=310+i*85
            s+=stroke(f'M{x-25} 215l60 237','#b28a53',5)
            for j in range(4):
                y=245+j*43
                s+=rect(x-28,y-19,58,40,'#945631' if key=='kofta' else '#bf8546',12,f'transform="rotate(-15 {x} {y})"')
                s+=stroke(f'M{x-20} {y-8}l35 12','#5f3c28',3)
        s+=ellipse(535,350,24,60,'#639442', 'transform="rotate(30 535 350)"')
    elif key=='falafel-box':
        for x,y in [(300,270),(405,255),(505,290),(340,380),(455,390)]:
            s+=ellipse(x,y,43,39,'#a17b35')+ellipse(x-8,y-9,30,22,'#bd9b4f')
            for _ in range(14):s+=ellipse(x+rng.randint(-24,24),y+rng.randint(-20,20),2,2,'#668344')
    elif key=='hummus':
        s+=ellipse(375,325,155,117,'#b6b1a1')+ellipse(375,315,147,107,'#fff4d7')+ellipse(375,315,130,94,'#e4c58d')
        for i in range(3):s+=stroke(f'M{290+i*20} 315q0 {-60+i*16} {80-i*10} {-60+i*16}t{80-i*10} {60-i*16}t{-80+i*10} {60-i*16}', '#c4a46b',6)
        s+=ellipse(380,315,30,20,'#c79b3f')
        s+=path('M505 260L590 400L470 410Z','url(#bread)')
    elif category=='Waffles':
        s+=rect(255,205,290,230,'#bc7b38',24)+rect(262,209,276,216,'#ebbb66',18)
        for x in range(277,530,50):
            for y in range(224,420,47):s+=rect(x,y,32,28,'#c78d43',4)+rect(x+4,y+3,23,18,'#db9f4c',3)
        s+=stroke('M280 235Q510 270 290 315T520 392','#6c3d29' if key=='strawberry-waffle' else '#a86b34',12)
        if key=='strawberry-waffle':
            for x,y in [(300,245),(395,310),(490,365)]:
                s+=path('M0-23Q-30-40-26-2Q-20 25 0 32Q30 6 26-14Q22-35 0-23Z','#cb5141',f'transform="translate({x} {y})"')
                s+=stroke(f'M{x} {y-18}l-13-12m13 12l13-12','#4d8043',4)
        else:
            for x,y in [(295,255),(350,300),(425,340),(480,382)]:s+=ellipse(x,y,27,18,'#fff0bf')+ellipse(x,y,7,4,'#d8bd7b')
        s+=ellipse(545,290,55,49,'#fff3d4')+ellipse(530,274,30,16,'#fff8e8')
    elif category=='Cakes':
        cream=key=='cheesecake'
        red=key=='red-velvet'
        base='#f0d8a5' if cream else '#a4453d' if red else '#643e31'
        s+=path('M265 265L445 200L560 340L365 403Z',base)
        s+=path('M265 265L365 403L365 445L265 310Z',base)+path('M365 403L560 340L560 390L365 445Z',base)
        s+=stroke('M272 286L365 422L551 361','#fff1d2' if cream or red else '#9d6a4b',13)
        s+=path('M265 265L445 200L560 340L365 403Z','#f7e8c2' if cream else '#914f3e' if red else '#4d3028')
        if cream:s+=path('M342 271Q385 240 430 242L470 290L400 332Z','#c34e43')
        if key=='brownie':s+=ellipse(440,275,64,47,'#fff1cd')
        for _ in range(18):s+=ellipse(rng.randint(320,480),rng.randint(260,325),3,2,'#f1cfa1')
    elif category=='Milkshakes':
        cookie=key=='cookie-shake'
        s+=ellipse(400,448,83,15,'#b4a190')
        s+=path('M315 215L485 215L465 435Q400 460 335 435Z','#d8e4df')
        s+=path('M325 228L475 228L455 427Q400 446 345 427Z','#dac6ae' if cookie else '#f8e5b7')
        s+=stroke('M448 245L485 127','#9b5844',12)
        for x,y,rx in [(400,220,88),(395,195,70),(405,170,49),(401,145,25)]:s+=ellipse(x,y,rx,23,'#fff9e8')
        s+=stroke('M350 255L365 413','#ffffff',7)
        if cookie:
            for _ in range(38):s+=ellipse(rng.randint(350,450),rng.randint(240,410),3,2,'#654535')
            s+=ellipse(355,190,30,24,'#654535')+ellipse(355,190,19,14,'#bba589')
    else:
        raise ValueError(f'No illustration for {key}')
    return s


DEFS = '<defs><linearGradient id="bread" x2="0" y2="1"><stop stop-color="#f3cc86"/><stop offset="1" stop-color="#cf8a3a"/></linearGradient><radialGradient id="bg"><stop stop-color="#fff8e8"/><stop offset="1" stop-color="#e3d9c3"/></radialGradient></defs>'


def document(name, body, width, height, background):
    return f'<svg xmlns="http://www.w3.org/2000/svg" width="{width}" height="{height}" viewBox="0 0 800 600"><title>{html.escape(name)}</title><desc>Original FoodHub demo food illustration. CC0-1.0. Illustrative serving.</desc>{DEFS}<rect width="800" height="600" fill="{background}"/>{body}</svg>'


def write_asset(name, relative, body, cover=False, background='#eee6d6'):
    width,height=(960,720) if cover else (480,360)
    source=SOURCES/relative.with_suffix('.svg')
    target=PUBLIC/relative
    source.parent.mkdir(parents=True,exist_ok=True)
    target.parent.mkdir(parents=True,exist_ok=True)
    source.write_text(document(name,body,width,height,background))
    return {'path':'/demo-images/'+relative.as_posix(),'source':str(source.relative_to(ROOT)),
            'label':name,'license':'CC0-1.0'}


def main():
    manifest=[]
    themes=['#e7d4b8','#e9d9c6','#dfd9bb','#d7decb','#eed9dc']
    for index,r in enumerate(CATALOG):
        for d in r['items']:
            # Plates and linen add a consistent editorial background; dish details vary.
            body=path('M0 45L145 0L250 600L0 600Z','#fff9ec','opacity=".48"')
            body+=ellipse(405,368,238,151,'#9c856c','opacity=".16"')
            body+=ellipse(400,329,245,181,'#fdf9ec')+ellipse(400,329,219,157,'#eee7d4')+ellipse(400,326,207,146,'#fffaf0')
            body+=art(d)
            manifest.append(write_asset(d['name'],Path('menu')/r['key']/(d['key']+'.png'),body,background=themes[index]))
        # A different local spread for each fictional restaurant, using its own dishes.
        first=r['items'][0]
        second=r['items'][6]
        body=path('M530 0L800 0L800 600L690 600Z','#fff7e5','opacity=".5"')
        body+=ellipse(400,325,285,222,'#fff8eb')+ellipse(400,325,259,197,'#e9e0cc')
        body+=f'<g transform="translate(-30 -20) scale(1.05)">{art(first)}</g>'
        body+=f'<g transform="translate(385 220) scale(.5)">{art(second)}</g>'
        body+=ellipse(200,220,39,34,'#d9e7ba')
        body+=path('M180 230Q195 175 225 205Q220 240 180 230','#70914c')
        manifest.append(write_asset(r['name']+' cover',Path('covers')/(r['key']+'.png'),body,cover=True,background=themes[index]))
    result=subprocess.run(['node',str(ROOT/'frontend/scripts/export-demo-images.mjs')],input=json.dumps(manifest),capture_output=True,text=True)
    if result.returncode: raise RuntimeError(result.stderr)
    print(result.stdout,end='')


if __name__=='__main__':
    main()
