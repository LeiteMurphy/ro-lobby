"""Artes de linha de classe do RO Lobby em pixel art 64x64 (PNG).

Arte original: cada linha é um emblema do arquétipo (arma, cor e tema) sobre um fundo
radial em degraus. Nada aqui copia arte da Gravity.
Uso, da raiz do repositório:
    python -I scripts/class-art/class_art.py web/static/portraits/classes <prévia.html>
Precisa do Pillow. Os ids das artes seguem a seção 6 de .specs/retrato-por-classe/tasks.md.
"""
import base64
import io
import math
import os
import sys

from PIL import Image

N = 64
OUTLINE = (18, 12, 10)


def hexc(h):
    return tuple(int(h[i:i + 2], 16) for i in (1, 3, 5))


def mix(a, b, t):
    return tuple(round(x + (y - x) * t) for x, y in zip(a, b))


def light(c, t=0.28):
    return mix(c, (255, 250, 235), t)


def dark(c, t=0.32):
    return mix(c, (10, 6, 12), t)


class Art:
    def __init__(self):
        self.px = [[None] * N for _ in range(N)]
        self.glow = []  # (cx, cy, r, cor) brilho atrás do emblema

    def set(self, x, y, c):
        x, y = int(round(x)), int(round(y))
        if 0 <= x < N and 0 <= y < N:
            self.px[y][x] = hexc(c) if isinstance(c, str) else c

    def poly(self, pts, c):
        ys = [p[1] for p in pts]
        for y in range(max(0, int(min(ys))), min(N, int(max(ys)) + 1)):
            yc = y + 0.5
            xs = []
            for (x0, y0), (x1, y1) in zip(pts, pts[1:] + pts[:1]):
                if (y0 <= yc < y1) or (y1 <= yc < y0):
                    xs.append(x0 + (yc - y0) * (x1 - x0) / (y1 - y0))
            xs.sort()
            for a, b in zip(xs[::2], xs[1::2]):
                for x in range(int(math.ceil(a - 0.5)), int(math.floor(b - 0.5)) + 1):
                    self.set(x, y, c)

    def ellipse(self, cx, cy, rx, ry, c):
        for y in range(N):
            for x in range(N):
                if ((x + 0.5 - cx) / rx) ** 2 + ((y + 0.5 - cy) / ry) ** 2 <= 1:
                    self.set(x, y, c)

    def disc(self, cx, cy, r, c):
        self.ellipse(cx, cy, r, r, c)

    def ring(self, cx, cy, r, w, c):
        for y in range(N):
            for x in range(N):
                d = math.hypot(x + 0.5 - cx, y + 0.5 - cy)
                if r - w / 2 <= d <= r + w / 2:
                    self.set(x, y, c)

    def thick(self, x0, y0, x1, y1, r, c):
        steps = int(max(abs(x1 - x0), abs(y1 - y0)) * 2) + 1
        for i in range(steps + 1):
            t = i / steps
            cx, cy = x0 + (x1 - x0) * t, y0 + (y1 - y0) * t
            for y in range(int(cy - r - 1), int(cy + r + 2)):
                for x in range(int(cx - r - 1), int(cx + r + 2)):
                    if math.hypot(x + 0.5 - cx, y + 0.5 - cy) <= r:
                        self.set(x, y, c)

    def curve(self, pts, r, c):
        """Curva de Bézier quadrática por trechos: lista de (x, y)."""
        out = []
        for i in range(0, len(pts) - 2, 2):
            (ax, ay), (bx, by), (cx, cy) = pts[i], pts[i + 1], pts[i + 2]
            for k in range(25):
                t = k / 24
                out.append(((1 - t) ** 2 * ax + 2 * (1 - t) * t * bx + t * t * cx,
                            (1 - t) ** 2 * ay + 2 * (1 - t) * t * by + t * t * cy))
        for (x0, y0), (x1, y1) in zip(out, out[1:]):
            self.thick(x0, y0, x1, y1, r, c)

    def star(self, cx, cy, ro, ri, n, c, rot=-90):
        pts = []
        for k in range(n * 2):
            a = math.radians(rot + k * 180 / n)
            r = ro if k % 2 == 0 else ri
            pts.append((cx + r * math.cos(a), cy + r * math.sin(a)))
        self.poly(pts, c)

    def sparkle(self, x, y, c, s=2):
        for k in range(-s, s + 1):
            self.set(x + k, y, c)
            self.set(x, y + k, c)
        self.set(x, y, (255, 255, 255))

    def render(self, theme):
        base = hexc(theme)
        img = Image.new('RGB', (N, N))
        # Fundo radial em 6 degraus, mais claro atrás do emblema.
        for y in range(N):
            for x in range(N):
                d = math.hypot(x + 0.5 - 32, y + 0.5 - 28) / 45
                level = min(5, int(d * 6))
                img.putpixel((x, y), mix(base, (6, 4, 10), 0.28 + level * 0.11))
        # Pontilhado de estrelas fracas no fundo.
        for i in range(14):
            x, y = (i * 37 + 11) % N, (i * 23 + 5) % N
            img.putpixel((x, y), mix(img.getpixel((x, y)), (255, 255, 255), 0.18))
        for cx, cy, r, c in self.glow:
            for y in range(N):
                for x in range(N):
                    d = math.hypot(x + 0.5 - cx, y + 0.5 - cy)
                    if d < r:
                        t = 0.45 * (1 - d / r)
                        t = round(t * 4) / 4
                        img.putpixel((x, y), mix(img.getpixel((x, y)), hexc(c), t))
        px = self.px
        # Sombra projetada para baixo e à direita.
        for y in range(N):
            for x in range(N):
                if px[y][x] is not None and x + 3 < N and y + 3 < N and px[y + 3][x + 3] is None:
                    img.putpixel((x + 3, y + 3), mix(img.getpixel((x + 3, y + 3)), (0, 0, 0), 0.45))
        # Contorno.
        for y in range(N):
            for x in range(N):
                if px[y][x] is None and any(
                    0 <= x + dx < N and 0 <= y + dy < N and px[y + dy][x + dx] is not None
                    for dx, dy in ((1, 0), (-1, 0), (0, 1), (0, -1))
                ):
                    img.putpixel((x, y), OUTLINE)
        # Emblema com luz de cima à esquerda: borda clara em cima/esquerda, escura embaixo/direita.
        for y in range(N):
            for x in range(N):
                c = px[y][x]
                if c is None:
                    continue
                up = px[y - 1][x] if y > 0 else None
                left = px[y][x - 1] if x > 0 else None
                down = px[y + 1][x] if y < N - 1 else None
                right = px[y][x + 1] if x < N - 1 else None
                if up != c or left != c:
                    c = light(c)
                elif down != c or right != c:
                    c = dark(c, 0.25)
                img.putpixel((x, y), c)
        return img


STEEL, STEEL_D, WHITE = '#cfd7e3', '#7f8a9c', '#fff6dc'
GOLD, GOLD_D = '#f2b33d', '#a8701a'
WOOD, WOOD_D = '#9c5c28', '#5e3414'
RED, RED_D, ORANGE, YELLOW = '#d8402f', '#8e2420', '#f0832a', '#ffd84a'
PURPLE, PURPLE_L, PURPLE_D = '#8a4ad2', '#c49aff', '#4e2484'
GREEN, GREEN_L, GREEN_D = '#3aa856', '#a4e48e', '#22683a'
BLUE, BLUE_L, CYAN = '#3a78dc', '#a2ccff', '#64e2ea'
PINK, PINK_L = '#de4f97', '#ffb0d6'
BLACK, PARCH, SKIN = '#2b2533', '#efe0b6', '#e7b088'


def greatsword(a):  # Cavaleiro: espada de duas mãos e asa de dragão
    a.poly([(4, 10), (30, 12), (34, 18), (26, 20), (24, 28), (16, 24), (12, 34), (6, 26)], RED_D)
    a.poly([(8, 13), (28, 14), (24, 18), (18, 18), (14, 26), (9, 22)], RED)
    for x1, y1 in ((24, 20), (16, 24), (8, 28)):
        a.thick(5, 10, x1, y1, 0.8, '#5a1612')
    a.poly([(55, 7), (58, 10), (25, 43), (21, 39)], STEEL)
    a.thick(54, 9, 25, 39, 0.6, WHITE)
    a.thick(13, 33, 31, 51, 2.2, GOLD)
    a.thick(20, 44, 9, 55, 2.6, WOOD)
    for k in range(3):
        a.thick(18 - k * 3, 46 + k * 3, 16 - k * 3, 48 + k * 3, 0.7, WOOD_D)
    a.disc(7.5, 56.5, 4, GOLD)
    a.glow.append((40, 24, 22, '#ffb07a'))


def sword_shield(a):  # Templário: espada e escudo
    a.poly([(48, 3), (52, 8), (52, 40), (44, 40), (44, 8)], STEEL)
    a.thick(48, 6, 48, 39, 0.6, WHITE)
    a.thick(39, 42, 57, 42, 2.4, GOLD)
    a.thick(48, 44, 48, 56, 2.4, WOOD)
    a.disc(48, 58, 3, GOLD)
    a.poly([(8, 12), (40, 12), (40, 34), (24, 58), (8, 34)], GOLD)
    a.poly([(11, 15), (37, 15), (37, 33), (24, 53), (11, 33)], WHITE)
    a.poly([(21, 18), (27, 18), (27, 48), (21, 48)], RED)
    a.poly([(13, 26), (35, 26), (35, 32), (13, 32)], RED)
    a.glow.append((26, 30, 26, '#fff0b0'))


def fire_staff(a):  # Bruxo: cajado e fogo
    a.thick(12, 60, 38, 26, 2.4, WOOD)
    a.thick(14, 57, 36, 28, 0.7, '#c27a3c')
    a.curve([(36, 30), (34, 20), (40, 16)], 1.6, WOOD_D)
    a.curve([(40, 28), (50, 26), (50, 18)], 1.6, WOOD_D)
    a.poly([(30, 22), (36, 2), (40, 14), (46, 0), (50, 14), (56, 6), (56, 26), (44, 34), (32, 30)], RED)
    a.poly([(35, 22), (40, 8), (44, 18), (49, 10), (52, 24), (44, 30), (36, 28)], ORANGE)
    a.disc(44, 22, 6, YELLOW)
    a.disc(43, 21, 2.5, WHITE)
    for x, y in ((24, 10), (58, 34), (28, 4)):
        a.sparkle(x, y, '#ffb35c', 1)
    a.glow.append((44, 20, 24, '#ff7a2a'))


def book(a):  # Sábio: livro e elementos
    a.poly([(6, 26), (31, 22), (31, 54), (6, 56)], WHITE)
    a.poly([(33, 22), (58, 26), (58, 56), (33, 54)], PARCH)
    a.poly([(30, 21), (34, 21), (34, 56), (30, 56)], BLUE)
    a.poly([(4, 55), (31, 53), (33, 53), (60, 55), (60, 59), (4, 59)], BLUE)
    for y in range(31, 50, 5):
        a.thick(10, y, 27, y - 2, 0.6, STEEL_D)
        a.thick(37, y - 2, 54, y, 0.6, STEEL_D)
    a.poly([(12, 16), (16, 6), (20, 16), (16, 20)], ORANGE)
    a.disc(16, 15, 2.5, YELLOW)
    a.poly([(32, 4), (36, 12), (32, 16), (28, 12)], CYAN)
    a.poly([(44, 16), (52, 6), (54, 12), (48, 18)], GREEN)
    a.thick(46, 16, 52, 8, 0.5, GREEN_D)
    a.glow.append((32, 14, 22, '#7ac8ff'))


def katar(a):  # Mercenário: katar e sombra
    for cx, tilt in ((20, -3), (44, 3)):
        a.poly([(cx + tilt, 4), (cx + 6, 36), (cx - 6, 36)], STEEL)
        a.thick(cx + tilt * 0.8, 8, cx, 34, 0.6, WHITE)
        a.poly([(cx - 8, 36), (cx + 8, 36), (cx + 8, 40), (cx - 8, 40)], BLACK)
        a.poly([(cx - 8, 40), (cx - 5, 40), (cx - 5, 58), (cx - 8, 58)], BLACK)
        a.poly([(cx + 5, 40), (cx + 8, 40), (cx + 8, 58), (cx + 5, 58)], BLACK)
        for y in (46, 52):
            a.thick(cx - 5, y, cx + 5, y, 1.2, RED)
    a.curve([(4, 50), (10, 40), (6, 30)], 1.2, PURPLE_D)
    a.curve([(60, 48), (54, 38), (58, 28)], 1.2, PURPLE_D)
    a.glow.append((32, 22, 24, '#d04050'))


def dagger(a):  # Arruaceiro: adaga, roxo
    a.poly([(56, 6), (40, 32), (34, 26)], STEEL)
    a.thick(54, 9, 38, 28, 0.7, WHITE)
    a.thick(26, 22, 46, 40, 2.4, PURPLE)
    a.disc(36, 31, 3, PURPLE_L)
    a.thick(33, 34, 14, 53, 3, PURPLE_D)
    for k in range(4):
        a.thick(30 - k * 4, 37 + k * 4, 27 - k * 4, 40 + k * 4, 0.7, PURPLE)
    a.disc(12, 55, 4, PURPLE_L)
    a.curve([(8, 34), (2, 22), (10, 14)], 1.3, PURPLE_L)
    a.curve([(14, 10), (22, 2), (30, 8)], 1.0, PURPLE)
    a.glow.append((40, 26, 26, '#a060ff'))


def hammer(a):  # Ferreiro: martelo e engrenagem
    for k in range(10):
        ang = k * 36
        a.disc(46 + 13 * math.cos(math.radians(ang)), 46 + 13 * math.sin(math.radians(ang)), 3, STEEL_D)
    a.disc(46, 46, 12, STEEL_D)
    a.disc(46, 46, 5, BLACK)
    # Cabo na diagonal e a cabeça perpendicular a ele, presa na ponta.
    ux, uy = 0.53, -0.85  # direção do cabo (para cima)
    vx, vy = 0.85, 0.53   # perpendicular
    a.thick(12, 60, 31, 29.5, 3, WOOD)
    a.thick(14, 57, 30, 31, 0.8, '#c27a3c')
    cx, cy = 34, 24
    corner = lambda su, sv: (cx + ux * 9 * su + vx * 12 * sv, cy + uy * 9 * su + vy * 12 * sv)
    a.poly([corner(1, -1), corner(1, 1), corner(-1, 1), corner(-1, -1)], STEEL)
    a.poly([corner(1, 0.7), corner(1, 1), corner(-1, 1), corner(-1, 0.7)], STEEL_D)  # faces de bater
    a.poly([corner(1, -1), corner(1, -0.7), corner(-1, -0.7), corner(-1, -1)], STEEL_D)
    a.thick(*corner(0.8, -0.6), *corner(0.8, 0.6), 0.6, WHITE)
    for x, y in ((8, 20), (56, 8), (10, 8)):
        a.sparkle(x, y, YELLOW, 2)
    a.glow.append((32, 24, 22, '#ffb040'))

def flask(a):  # Alquimista: frascos
    a.poly([(25, 4), (39, 4), (39, 11), (25, 11)], WOOD)
    a.poly([(27, 11), (37, 11), (37, 24), (27, 24)], BLUE_L)
    a.disc(32, 40, 18, BLUE_L)
    a.ellipse(32, 46, 16, 11, GREEN)
    a.poly([(16, 40), (48, 40), (48, 44), (16, 44)], GREEN)
    for x, y, r in ((26, 46, 2.5), (36, 50, 2), (39, 42, 1.5), (30, 36, 1.5)):
        a.disc(x, y, r, GREEN_L)
    a.thick(22, 30, 24, 26, 1, WHITE)
    a.poly([(52, 30), (58, 30), (58, 56), (52, 56)], BLUE_L)
    a.poly([(52, 44), (58, 44), (58, 56), (52, 56)], PINK)
    a.glow.append((32, 42, 26, '#7aff9a'))


def holy_staff(a):  # Sacerdote: cajado sagrado e luz
    a.thick(32, 28, 32, 62, 2.2, GOLD)
    a.thick(31, 30, 31, 60, 0.6, YELLOW)
    a.ring(32, 18, 11, 3, GOLD)
    a.poly([(30, 9), (34, 9), (34, 27), (30, 27)], WHITE)
    a.poly([(25, 15), (39, 15), (39, 19), (25, 19)], WHITE)
    for k in range(8):
        ang = math.radians(k * 45 + 22)
        a.thick(32 + 15 * math.cos(ang), 18 + 15 * math.sin(ang),
                32 + 20 * math.cos(ang), 18 + 20 * math.sin(ang), 0.8, BLUE_L)
    a.glow.append((32, 18, 26, '#bfe4ff'))


def fist(a):  # Monge: punho fechado e esferas de energia
    for x in (14, 22, 30, 38):
        a.poly([(x, 20), (x + 8, 20), (x + 8, 32), (x, 32)], SKIN)
        a.disc(x + 4, 21, 4, SKIN)
    a.poly([(14, 30), (46, 30), (46, 46), (18, 48), (14, 42)], SKIN)
    a.poly([(6, 32), (20, 32), (22, 40), (8, 40)], SKIN)
    for x in (22, 30, 38):
        a.thick(x, 22, x, 30, 0.6, '#b07850')
    a.poly([(18, 46), (46, 46), (46, 60), (18, 60)], ORANGE)
    a.poly([(18, 46), (46, 46), (46, 50), (18, 50)], YELLOW)
    for cx, cy, r in ((54, 14, 5), (8, 14, 4), (56, 40, 4)):
        a.ring(cx, cy, r + 2.5, 1, YELLOW)
        a.disc(cx, cy, r, YELLOW)
        a.disc(cx - 1, cy - 1, r / 2.5, WHITE)
    a.glow.append((30, 34, 26, '#ffc060'))


def bow(a):  # Caçador: arco e flecha
    a.curve([(20, 4), (44, 32), (20, 60)], 2.2, WOOD)
    a.curve([(21, 7), (41, 32), (21, 57)], 0.6, '#c27a3c')
    a.thick(19, 5, 19, 59, 0.5, WHITE)
    a.thick(6, 32, 54, 32, 1, STEEL_D)
    a.poly([(52, 27), (62, 32), (52, 37)], STEEL)
    a.poly([(4, 27), (12, 32), (4, 37), (8, 32)], GREEN)
    a.poly([(48, 6), (56, 4), (54, 12)], GREEN_L)
    a.thick(48, 6, 58, 16, 0.5, GREEN_D)
    a.glow.append((32, 32, 26, '#a0ff90'))


def violin(a):  # Bardo: violino
    a.ellipse(30, 44, 14, 12, WOOD)
    a.ellipse(30, 24, 11, 10, WOOD)
    a.poly([(22, 30), (38, 30), (38, 38), (22, 38)], WOOD)
    a.ellipse(26, 33, 2, 2.5, '#3a1e0c')
    a.ellipse(34, 33, 2, 2.5, '#3a1e0c')
    a.poly([(28, 4), (32, 4), (32, 30), (28, 30)], BLACK)
    a.disc(30, 4, 3, BLACK)
    a.poly([(26, 52), (34, 52), (34, 56), (26, 56)], BLACK)
    a.thick(6, 60, 56, 10, 0.8, WHITE)
    a.poly([(4, 58), (10, 58), (10, 62), (4, 62)], WOOD_D)
    for x, y in ((50, 30), (56, 44)):
        a.disc(x, y, 2.5, BLUE_L)
        a.thick(x + 2, y, x + 2, y - 8, 0.6, BLUE_L)
    a.glow.append((32, 30, 26, '#80a8ff'))


def whip(a):  # Odalisca: chicote
    a.thick(8, 58, 20, 44, 3, WOOD)
    a.thick(18, 46, 22, 42, 3.4, GOLD)
    a.curve([(22, 42), (30, 22), (46, 16), (60, 14), (58, 30), (56, 44), (42, 42), (34, 40), (38, 30),
             (42, 24), (48, 30)], 1.6, PINK)
    a.disc(48, 30, 2.5, PINK_L)
    for x, y in ((10, 12), (30, 6), (56, 54), (6, 30)):
        a.sparkle(x, y, PINK_L, 2)
    a.glow.append((40, 28, 26, '#ff7ab8'))


def sun_star(a):  # Mestre Taekwon: sol, lua e estrelas
    a.star(26, 34, 22, 14, 12, ORANGE)
    a.disc(26, 34, 13, YELLOW)
    a.disc(22, 30, 4, WHITE)
    a.disc(50, 14, 9, '#f4f0d0')
    a.disc(54, 11, 8, None) if False else None
    for y in range(N):
        for x in range(N):
            if math.hypot(x + 0.5 - 54, y + 0.5 - 11) < 8 and a.px[y][x] == hexc('#f4f0d0'):
                a.px[y][x] = None
    a.star(52, 46, 7, 3, 5, WHITE)
    a.star(8, 10, 4, 1.8, 5, WHITE)
    a.glow.append((26, 34, 28, '#ffd060'))


def talisman(a):  # Espiritualista: talismã e almas
    a.poly([(20, 6), (44, 6), (44, 58), (20, 58)], PARCH)
    a.poly([(20, 6), (44, 6), (44, 11), (20, 11)], RED)
    a.poly([(20, 53), (44, 53), (44, 58), (20, 58)], RED)
    a.thick(32, 16, 32, 46, 1, RED)
    a.thick(25, 22, 39, 22, 1, RED)
    a.thick(26, 32, 32, 38, 1, RED)
    a.thick(38, 32, 32, 38, 1, RED)
    a.ring(32, 44, 4, 1.4, RED)
    for cx, cy in ((9, 18), (55, 30), (11, 48)):
        a.poly([(cx, cy - 8), (cx + 4, cy), (cx, cy + 4), (cx - 4, cy)], PURPLE_L)
        a.disc(cx, cy + 1, 3, BLUE_L)
    a.glow.append((32, 32, 28, '#b080ff'))


def shuriken(a):  # Ninja: shuriken
    a.star(32, 32, 27, 8, 4, STEEL, rot=-90)
    a.star(32, 32, 20, 6, 4, STEEL_D, rot=-45)
    a.disc(32, 32, 5, BLACK)
    a.disc(32, 32, 2, RED)
    for k in range(4):
        ang = math.radians(-90 + k * 90)
        a.thick(32 + 9 * math.cos(ang), 32 + 9 * math.sin(ang),
                32 + 24 * math.cos(ang), 32 + 24 * math.sin(ang), 0.5, WHITE)
    a.glow.append((32, 32, 28, '#d04040'))


def revolver(a):  # Justiceiro: revólver
    a.poly([(20, 16), (60, 16), (60, 24), (20, 24)], STEEL)
    a.thick(22, 17, 58, 17, 0.5, WHITE)
    a.poly([(16, 14), (34, 14), (34, 32), (16, 32)], STEEL_D)
    for y in (18, 23, 28):
        a.thick(17, y, 33, y, 0.5, BLACK)
    a.poly([(12, 10), (18, 14), (14, 16)], STEEL_D)
    a.poly([(12, 30), (26, 30), (22, 56), (6, 54)], WOOD)
    a.thick(10, 40, 20, 42, 0.6, WOOD_D)
    a.ring(30, 36, 6, 2, STEEL_D)
    a.thick(30, 32, 28, 38, 1, BLACK)
    a.disc(58, 20, 1.5, BLACK)
    a.glow.append((36, 26, 26, '#ffc080'))


def wings(a):  # Superaprendiz: mochila de aventureiro iniciante
    bag, bag_d, flap = '#f08a3a', '#b85a1c', '#ffb85c'
    a.thick(20, 14, 20, 30, 2, bag_d)  # alças
    a.thick(44, 14, 44, 30, 2, bag_d)
    a.ring(32, 14, 8, 2.4, bag_d)  # alça de mão
    a.poly([(12, 22), (52, 22), (54, 58), (10, 58)], bag)
    a.ellipse(32, 24, 20, 6, bag)
    a.poly([(12, 22), (52, 22), (52, 36), (12, 36)], flap)
    a.ellipse(32, 36, 20, 5, flap)
    a.disc(32, 38, 3, YELLOW)  # fecho
    a.poly([(20, 44), (44, 44), (44, 56), (20, 56)], '#7ac8ff')  # bolso
    a.thick(20, 44, 44, 44, 0.8, '#4a8ad8')
    a.star(32, 50, 5, 2.2, 5, YELLOW)  # remendo de estrela
    a.poly([(50, 10), (58, 4), (60, 12)], '#8ae07a')  # folha presa
    a.disc(8, 46, 3, '#ffd84a')  # pirulito
    a.thick(8, 49, 8, 58, 0.7, WHITE)
    a.glow.append((32, 36, 28, '#ffd07a'))

def paw(a):  # Invocador (Doram): pata
    a.ellipse(32, 42, 15, 13, ORANGE)
    a.ellipse(32, 44, 8, 7, '#ffd3a8')
    for cx, cy in ((13, 26), (24, 16), (40, 16), (51, 26)):
        a.ellipse(cx, cy, 6, 7, ORANGE)
        a.ellipse(cx, cy + 1, 3, 3.5, '#ffd3a8')
    a.poly([(52, 50), (60, 44), (58, 54)], GREEN)
    a.glow.append((32, 34, 26, '#ffb060'))


def leaf(a):  # Druida: Árvore do Mundo, folha
    pts = []
    for k in range(41):
        t = k / 40
        y = 6 + t * 50
        w = 20 * math.sin(math.pi * t) ** 0.8
        pts.append((32 + w, y))
    pts += [(64 - x, y) for x, y in reversed(pts)]
    a.poly(pts, GREEN)
    a.thick(32, 8, 32, 60, 1, GREEN_D)
    for y in (18, 28, 38, 48):
        a.thick(32, y + 6, 22, y, 0.6, GREEN_D)
        a.thick(32, y + 6, 42, y, 0.6, GREEN_D)
    a.curve([(32, 58), (24, 62), (14, 58)], 1.2, WOOD)
    for x, y in ((10, 14), (54, 20), (52, 52)):
        a.sparkle(x, y, GREEN_L, 2)
    a.glow.append((32, 30, 28, '#80ff90'))


def short_sword(a):  # Espadachim: espada curta
    a.poly([(32, 4), (36, 10), (36, 40), (28, 40), (28, 10)], STEEL)
    a.thick(32, 8, 32, 39, 0.6, WHITE)
    a.thick(20, 42, 44, 42, 2.4, GOLD)
    a.thick(32, 44, 32, 56, 2.4, WOOD)
    a.disc(32, 58, 3.4, GOLD)
    a.glow.append((32, 26, 24, '#d0d8ff'))


def wand(a):  # Mago: varinha e estrela
    a.thick(10, 58, 36, 30, 2, WOOD)
    a.star(42, 22, 15, 6.5, 5, YELLOW)
    a.disc(42, 22, 4, WHITE)
    for x, y in ((14, 14), (58, 44), (24, 30)):
        a.sparkle(x, y, YELLOW, 2)
    a.glow.append((42, 22, 26, '#ffe070'))


def mask(a):  # Gatuno: máscara
    a.poly([(6, 24), (58, 24), (54, 40), (36, 40), (32, 34), (28, 40), (10, 40)], BLACK)
    a.ellipse(20, 31, 6, 3.5, WHITE)
    a.ellipse(44, 31, 6, 3.5, WHITE)
    a.curve([(6, 28), (0, 34), (4, 48)], 1.4, PURPLE)
    a.curve([(58, 28), (64, 34), (60, 48)], 1.4, PURPLE)
    a.glow.append((32, 32, 26, '#9a70d0'))


def coin_bag(a):  # Mercador: saco de moedas
    a.ellipse(30, 42, 18, 16, '#c79c62')
    a.poly([(22, 16), (38, 16), (36, 28), (24, 28)], '#c79c62')
    a.thick(22, 28, 38, 28, 1.6, WOOD_D)
    a.disc(30, 44, 6, GOLD)
    a.thick(30, 40, 30, 48, 0.8, GOLD_D)
    for cx, cy in ((52, 52), (56, 46), (48, 58)):
        a.ellipse(cx, cy, 5, 3, GOLD)
    a.glow.append((30, 40, 26, '#ffd070'))


def cross(a):  # Noviço: símbolo sagrado
    a.poly([(28, 6), (36, 6), (36, 58), (28, 58)], WHITE)
    a.poly([(14, 18), (50, 18), (50, 26), (14, 26)], WHITE)
    a.disc(32, 22, 5, GOLD)
    a.glow.append((32, 26, 28, '#cfe6ff'))


def arrow(a):  # Arqueiro: flecha
    a.thick(10, 54, 50, 14, 1.4, WOOD)
    a.poly([(46, 10), (58, 6), (54, 18)], STEEL)
    a.poly([(4, 52), (10, 48), (16, 54), (10, 60)], GREEN)
    a.poly([(8, 56), (14, 50), (18, 56)], GREEN_L)
    a.glow.append((32, 32, 24, '#a0ff90'))


def belt(a):  # Taekwon: faixa
    a.poly([(4, 26), (60, 26), (60, 36), (4, 36)], BLACK)
    a.poly([(26, 22), (38, 22), (38, 40), (26, 40)], '#3e3850')
    a.poly([(28, 40), (34, 40), (24, 60), (16, 58)], BLACK)
    a.poly([(30, 40), (36, 40), (46, 58), (38, 60)], BLACK)
    a.thick(6, 28, 58, 28, 0.5, '#5a5470')
    for x, y in ((12, 12), (52, 12)):
        a.star(x, y, 5, 2, 5, YELLOW)
    a.glow.append((32, 32, 26, '#ffd060'))


LINES = [
    ('cavaleiro', 'Linha do Cavaleiro', 'Cavaleiro, Lorde, Cavaleiro Rúnico, Cavaleiro Draconiano', '#a8322a', greatsword),
    ('templario', 'Linha do Templário', 'Templário, Paladino, Guardião Real, Guardião Imperial', '#b8923a', sword_shield),
    ('bruxo', 'Linha do Bruxo', 'Bruxo, Arquimago, Arcano, Magus', '#c0502a', fire_staff),
    ('sabio', 'Linha do Sábio', 'Sábio, Professor, Feiticeiro, Elementalista', '#2f6fb8', book),
    ('mercenario', 'Linha do Mercenário', 'Mercenário, Algoz, Sicário, Executor', '#5a2230', katar),
    ('arruaceiro', 'Linha do Arruaceiro', 'Arruaceiro, Desordeiro, Renegado, Mandraque', '#6b3aa8', dagger),
    ('ferreiro', 'Linha do Ferreiro', 'Ferreiro, Mestre-Ferreiro, Mecânico, Engenheiro', '#9a6a3a', hammer),
    ('alquimista', 'Linha do Alquimista', 'Alquimista, Criador, Bioquímico, Cientista', '#2f8a5a', flask),
    ('sacerdote', 'Linha do Sacerdote', 'Sacerdote, Sumo Sacerdote, Arcebispo, Cardeal', '#4a8ac8', holy_staff),
    ('monge', 'Linha do Monge', 'Monge, Mestre, Shura, Inquisidor', '#c0702a', fist),
    ('cacador', 'Linha do Caçador', 'Caçador, Atirador de Elite, Sentinela, Falcão do Vento', '#3a7a3a', bow),
    ('bardo', 'Linha do Bardo', 'Bardo, Menestrel, Trovador, Maestro', '#3a5ac8', violin),
    ('odalisca', 'Linha da Odalisca', 'Odalisca, Cigana, Musa, Diva', '#b8447e', whip),
    ('mestre-taekwon', 'Linha do Mestre Taekwon', 'Mestre Taekwon, Mestre Estelar, Mestre Celestial', '#b08a22', sun_star),
    ('espiritualista', 'Linha do Espiritualista', 'Espiritualista, Ceifador de Almas, Asceta das Almas', '#5a3a9a', talisman),
    ('ninja', 'Linha do Ninja', 'Ninja, Kagerou, Oboro, Shinkiro, Shiranui', '#3a3a4a', shuriken),
    ('justiceiro', 'Linha do Justiceiro', 'Justiceiro, Insurgente, Guerrilheiro', '#7a5a3a', revolver),
    ('superaprendiz', 'Linha do Superaprendiz', 'Aprendiz, Superaprendiz, Superaprendiz EX, Hiperaprendiz', '#4a8ac8', wings),
    ('invocador', 'Linha do Invocador (Doram)', 'Invocador, Animista', '#b87a32', paw),
    ('druida', 'Linha do Druida', 'Druida, Karnos, Alitea', '#2a7a44', leaf),
    ('espadachim', 'Família Espadachim', 'Espadachim, antes de escolher a linha', '#6a6a7a', short_sword),
    ('mago', 'Família Mago', 'Mago, antes de escolher a linha', '#4a4aa8', wand),
    ('gatuno', 'Família Gatuno', 'Gatuno, antes de escolher a linha', '#4a3a5a', mask),
    ('mercador', 'Família Mercador', 'Mercador, antes de escolher a linha', '#8a6a2a', coin_bag),
    ('novico', 'Família Noviço', 'Noviço, antes de escolher a linha', '#5a7aa8', cross),
    ('arqueiro', 'Família Arqueiro', 'Arqueiro, antes de escolher a linha', '#4a7a4a', arrow),
    ('taekwon', 'Família Taekwon', 'Taekwon, antes de escolher a linha', '#a87a2a', belt),
]

PREVIEW_HEAD = open(os.path.join(os.path.dirname(os.path.abspath(__file__)), 'preview_head.html'), encoding='utf-8').read()


def main():
    out_dir, preview = sys.argv[1], sys.argv[2]
    os.makedirs(out_dir, exist_ok=True)
    rows = []
    for id_, name, classes, theme, draw in LINES:
        a = Art()
        draw(a)
        img = a.render(theme)
        path = os.path.join(out_dir, f'{id_}.png')
        img.save(path, optimize=True)
        buf = io.BytesIO()
        img.save(buf, 'PNG', optimize=True)
        uri = 'data:image/png;base64,' + base64.b64encode(buf.getvalue()).decode()
        rows.append(f'<figure><img src="{uri}" alt="{name}" width="256" height="256"><figcaption><b>{name}</b><span>{classes}</span><code>{id_}.png · {len(buf.getvalue()) // 100 / 10} KB</code></figcaption></figure>')
    with open(preview, 'w', encoding='utf-8', newline='\n') as f:
        f.write(PREVIEW_HEAD + '\n'.join(rows) + '\n</div></main>')
    print(len(rows), 'artes')


if __name__ == '__main__':
    main()
