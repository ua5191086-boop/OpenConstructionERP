#!/usr/bin/env python3
"""Генератор тестовых BIM-данных для OpenConstructionERP"""
import json, random, sys

IFC_TYPES = ['IfcWall', 'IfcSlab', 'IfcBeam', 'IfcColumn', 'IfcPipeSegment', 'IfcDuctSegment', 'IfcCableSegment', 'IfcStair', 'IfcRamp', 'IfcDoor', 'IfcWindow', 'IfcCovering']
LEVELS = ['B1', 'B2', 'L1', 'L2', 'L3', 'L4', 'L5', 'Roof']
MATERIALS = ['Бетон B25', 'Бетон B40', 'Сталь С245', 'Сталь С345', 'Кирпич', 'Гипсокартон', 'Стекло', 'Алюминий']

def gen_bim():
    elements = []
    for i in range(50):
        elements.append({
            'id': f'IFC-{i+1:06d}',
            'type': random.choice(IFC_TYPES),
            'name': f'{random.choice(IFC_TYPES).replace("Ifc","")} #{i+1}',
            'level': random.choice(LEVELS),
            'material': random.choice(MATERIALS),
            'volume': round(random.uniform(0.5, 50), 2),
            'area': round(random.uniform(5, 200), 2),
            'length': round(random.uniform(1, 30), 2),
            'weight': round(random.uniform(100, 50000), 2),
        })
    
    # Clashes
    clashes = []
    for i in range(10):
        a = random.randint(0, 49)
        b = random.randint(0, 49)
        while b == a:
            b = random.randint(0, 49)
        clashes.append({
            'type': random.choice(['hard', 'soft', 'clearance']),
            'severity': random.choice(['critical', 'high', 'medium', 'low']),
            'status': random.choice(['open', 'in_progress', 'resolved']),
            'element_a': elements[a]['id'],
            'element_b': elements[b]['id'],
            'distance': round(random.uniform(0, 0.5), 4),
        })
    
    return {'elements': elements, 'clashes': clashes}

def summary(data):
    lines = []
    lines.append("=" * 60)
    lines.append("🏗️ BIM-МОДЕЛЬ — СВОДКА")
    lines.append("=" * 60)
    lines.append(f"Элементов: {len(data['elements'])}")
    lines.append(f"Типов: {len(set(e['type'] for e in data['elements']))}")
    lines.append(f"Коллизий: {len(data['clashes'])}")
    open_c = sum(1 for c in data['clashes'] if c['status'] == 'open')
    lines.append(f"Открытых коллизий: {open_c}")
    lines.append("")
    by_type = {}
    for e in data['elements']:
        t = e['type'].replace('Ifc', '')
        by_type[t] = by_type.get(t, 0) + 1
    lines.append("По типам:")
    for t, c in sorted(by_type.items(), key=lambda x: -x[1]):
        lines.append(f"  {t}: {c}")
    return '\n'.join(lines)

if __name__ == '__main__':
    data = gen_bim()
    if '--json' in sys.argv:
        print(json.dumps(data, ensure_ascii=False, indent=2))
    else:
        print(summary(data))
