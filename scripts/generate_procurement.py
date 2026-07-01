#!/usr/bin/env python3
"""Генератор тестовых закупок для OpenConstructionERP"""
import json, random, sys
from datetime import datetime, timedelta

VENDORS = [
    'ТОО "СнабСтрой"', 'ТОО "МеталлИнвест"', 'АО "Цементный завод"',
    'ТОО "ЭлектроСнаб"', 'ТОО "ТрубМеталл"', 'China Railway Materials',
    'Vöslauer Baustoffe', 'ТОО "Бетон-Норд"', 'ТОО "Арматура-Юг"',
    'ТОО "КабельСтрой"', 'ТОО "НасосСервис"', 'ТОО "Вентиляция-КЗ"',
]
MATERIALS = [
    ('Цемент М500', 'т', 100, 5000, 80, 150),
    ('Арматура А500С', 'т', 50, 2000, 350, 600),
    ('Бетон В25', 'м³', 500, 10000, 45, 90),
    ('Кабель силовой', 'м', 1000, 50000, 5, 25),
    ('Труба стальная', 'т', 20, 500, 500, 1200),
    ('Щебень фр.20-40', 'м³', 500, 5000, 12, 30),
    ('Песок строительный', 'м³', 500, 5000, 8, 20),
    ('Электроды сварочные', 'кг', 100, 5000, 2, 8),
    ('Краска антикоррозийная', 'л', 200, 3000, 5, 15),
    ('Болты М20', 'шт', 500, 20000, 0.5, 2),
    ('Насос дренажный', 'шт', 5, 50, 500, 3000),
    ('Вентилятор осевой', 'шт', 10, 100, 1000, 5000),
    ('Трансформатор ТМ-1000', 'шт', 1, 10, 10000, 50000),
    ('Кабель-канал', 'м', 500, 10000, 3, 10),
    ('Геотекстиль', 'м²', 1000, 20000, 1, 5),
]

def gen_procurement():
    items = []
    for i in range(15):
        mat = random.choice(MATERIALS)
        qty = round(random.uniform(mat[2], mat[3]), 2)
        price = round(random.uniform(mat[4], mat[5]), 2)
        total = round(qty * price, 2)
        items.append({
            'code': f'MAT-{i+1:03d}',
            'name': mat[0],
            'unit': mat[1],
            'quantity': qty,
            'unit_price': price,
            'total': total,
        })
    return items

def summary(items):
    lines = []
    lines.append("=" * 60)
    lines.append("📦 ЗАКУПКИ — СВОДКА")
    lines.append("=" * 60)
    total = sum(i['total'] for i in items)
    lines.append(f"Позиций: {len(items)} | Общая сумма: ${total:,.2f}")
    lines.append("")
    for i in items:
        lines.append(f"  {i['code']} — {i['name']} — {i['quantity']} {i['unit']} × ${i['unit_price']} = ${i['total']:,.2f}")
    return '\n'.join(lines)

if __name__ == '__main__':
    items = gen_procurement()
    if '--json' in sys.argv:
        print(json.dumps(items, ensure_ascii=False, indent=2))
    else:
        print(summary(items))
