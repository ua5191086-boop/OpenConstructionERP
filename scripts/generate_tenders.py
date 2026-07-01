#!/usr/bin/env python3
"""
Генератор тестовых тендеров для OpenConstructionERP
Генерирует: тендеры, лоты, участников, ценовые предложения, оценки

Использование:
  python3 generate_tenders.py              # генерация в БД (если доступна)
  python3 generate_tenders.py --json       # вывод в JSON
  python3 generate_tenders.py --sql        # вывод SQL-вставок
"""

import json
import random
import sys
import os
from datetime import datetime, timedelta
from decimal import Decimal

# ============================================================================
# Конфигурация
# ============================================================================
TENDER_COUNT = 5
BIDDERS_PER_TENDER = 4
ITEMS_PER_LOT = 8

# ============================================================================
# Данные
# ============================================================================
TENDER_TYPES = ['open', 'limited', 'single_source', 'request_quote']
PROCUREMENT_METHODS = ['e-auction', 'competitive', 'negotiated']
FUNDING_SOURCES = ['Госбюджет РК', 'Местный бюджет', 'ГЧП', 'Частные инвестиции', 'Грант ЕБРР', 'Грант АБР']

SECTIONS = [
    'Участок 1 — ПК 0+000 — ПК 30+000',
    'Участок 2 — ПК 30+000 — ПК 60+000',
    'Участок 3 — ПК 60+000 — ПК 90+000',
    'Участок 4 — ПК 90+000 — ПК 120+000',
    'Станция А — Центральная',
    'Станция Б — Восточная',
    'Станция В — Западная',
    'Станция Г — Северная',
    'Станция Д — Южная',
    'Депо и инфраструктура',
]

CONTRACTORS = [
    ('CAI Interbudmontazh GmbH', 'Австрия'),
    ('ТОО "Алматыметрострой"', 'Казахстан'),
    ('АО "НК "Казахстан темір жолы"', 'Казахстан'),
    ('ТОО "СтройИнвест Групп"', 'Казахстан'),
    ('China Railway Construction Corp.', 'Китай'),
    ('Vinci Construction Grands Projets', 'Франция'),
    ('Strabag SE', 'Австрия'),
    ('Porr AG', 'Австрия'),
    ('Astaldi S.p.A.', 'Италия'),
    ('Gülermak A.Ş.', 'Турция'),
    ('ТОО "Базис-А"', 'Казахстан'),
    ('ТОО "Курылыс-Монтаж"', 'Казахстан'),
    ('ТОО "Астана-Инжиниринг"', 'Казахстан'),
    ('ТОО "BI-Group"', 'Казахстан'),
    ('ТОО "СтройТехСервис"', 'Казахстан'),
]

CBS_CHAPTERS = [
    ('01', 'Подготовительные работы'),
    ('02', 'Земляные работы'),
    ('03', 'Бетонные и железобетонные работы'),
    ('04', 'Металлоконструкции'),
    ('05', 'Тоннельные работы'),
    ('06', 'Путевые работы'),
    ('07', 'Электроснабжение и СЦБ'),
    ('08', 'Вентиляция и сантехника'),
    ('09', 'Отделочные работы'),
    ('10', 'Оборудование'),
    ('11', 'Пусконаладочные работы'),
    ('12', 'Прочие работы и затраты'),
]

TENDER_NAMES = [
    'Строительство участка метро — {section}',
    'Поставка ТПМК для проходки — {section}',
    'Электроснабжение — {section}',
    'Вентиляция и дымоудаление — {section}',
    'Отделочные работы — {section}',
    'Путевые работы — {section}',
    'СЦБ и связь — {section}',
    'Поставка оборудования — {section}',
    'Мониторинг деформаций — {section}',
    'Геотехнические изыскания — {section}',
]

ITEM_DESCRIPTIONS = [
    ('Разработка грунта', 'м³', 50000, 150000, 5.50, 12.00),
    ('Бетонирование стен', 'м³', 5000, 30000, 85.00, 180.00),
    ('Армирование', 'т', 500, 5000, 450.00, 1200.00),
    ('Опалубка стен', 'м²', 10000, 60000, 25.00, 65.00),
    ('Гидроизоляция', 'м²', 5000, 40000, 15.00, 45.00),
    ('Монтаж ТПМК', 'компл', 1, 3, 5000000, 15000000),
    ('Проходка тоннеля', 'м', 500, 5000, 15000, 35000),
    ('Монтаж ж/б блоков обделки', 'шт', 5000, 50000, 120.00, 350.00),
    ('Устройство путей', 'км', 1, 30, 500000, 1500000),
    ('Монтаж кабельных линий', 'км', 5, 100, 25000, 80000),
    ('Установка трансформаторов', 'шт', 2, 20, 50000, 200000),
    ('Монтаж вентиляционного оборудования', 'компл', 5, 50, 30000, 120000),
    ('Отделка станции', 'м²', 2000, 20000, 80.00, 300.00),
    ('Установка эскалаторов', 'шт', 4, 20, 200000, 600000),
    ('Монтаж лифтов', 'шт', 2, 10, 150000, 450000),
    ('ПНР электрооборудования', 'компл', 1, 10, 100000, 500000),
    ('Геодезический мониторинг', 'мес', 6, 36, 15000, 50000),
    ('Инженерно-геологические изыскания', 'компл', 1, 5, 50000, 300000),
    ('Разработка ПСД', 'компл', 1, 3, 200000, 1000000),
    ('Авторский надзор', 'мес', 6, 24, 20000, 80000),
]

# ============================================================================
# Генерация
# ============================================================================
def random_date(start_days=30, end_days=180):
    """Случайная дата в диапазоне от сегодня + start_days до + end_days"""
    today = datetime.now()
    delta = random.randint(start_days, end_days)
    return today + timedelta(days=delta)

def generate_tender(tender_num):
    """Генерация одного тендера"""
    section = random.choice(SECTIONS)
    name_template = random.choice(TENDER_NAMES)
    name = name_template.format(section=section)
    
    client = random.choice(CONTRACTORS)
    tender_type = random.choice(TENDER_TYPES)
    
    # Статус: первые 2 draft, остальные published/in_progress
    if tender_num <= 2:
        status = 'draft'
    elif tender_num <= 4:
        status = 'published'
    else:
        status = random.choice(['in_progress', 'evaluation'])
    
    published_at = None
    submission_deadline = None
    bid_open_date = None
    
    if status != 'draft':
        published_at = random_date(-60, -10)
        submission_deadline = published_at + timedelta(days=random.randint(30, 60))
        bid_open_date = submission_deadline + timedelta(days=1)
    
    budget = round(random.uniform(500000, 50000000), 2)
    
    return {
        'code': f'T-2026-{tender_num:03d}',
        'name': name,
        'description': f'Тендер на {name.lower()}. Заказчик: {client[0]}.',
        'tender_type': tender_type,
        'status': status,
        'client_name': client[0],
        'client_country': client[1],
        'budget_amount': budget,
        'currency': 'USD',
        'published_at': published_at.isoformat() if published_at else None,
        'submission_deadline': submission_deadline.isoformat() if submission_deadline else None,
        'bid_open_date': bid_open_date.isoformat() if bid_open_date else None,
        'bid_bond_pct': round(random.uniform(1.0, 5.0), 2),
        'performance_bond_pct': round(random.uniform(5.0, 15.0), 2),
        'advance_payment_pct': round(random.uniform(0, 30), 2),
        'procurement_method': random.choice(PROCUREMENT_METHODS),
        'funding_source': random.choice(FUNDING_SOURCES),
        'lots': [],
        'bidders': [],
    }

def generate_lots(tender, lot_count=2):
    """Генерация лотов для тендера"""
    for i in range(lot_count):
        lot = {
            'lot_number': i + 1,
            'name': f'Лот {i+1}: {random.choice(["Основные работы", "Дополнительные работы", "Оборудование", "Материалы", "Услуги"])}',
            'description': f'Лот {i+1} тендера {tender["code"]}',
            'estimated_amount': round(tender['budget_amount'] / lot_count * random.uniform(0.8, 1.2), 2),
            'currency': 'USD',
            'section': random.choice(SECTIONS),
            'status': 'active',
            'items': [],
        }
        
        # Позиции лота
        for j in range(ITEMS_PER_LOT):
            desc, unit, qty_min, qty_max, price_min, price_max = random.choice(ITEM_DESCRIPTIONS)
            qty = round(random.uniform(qty_min, qty_max), 2)
            unit_price = round(random.uniform(price_min, price_max), 2)
            total = round(qty * unit_price, 2)
            
            item = {
                'item_code': f'{tender["code"]}-L{lot["lot_number"]}-{j+1:03d}',
                'description': desc,
                'unit': unit,
                'quantity': qty,
                'estimated_unit_price': unit_price,
                'estimated_total': total,
                'sort_order': j + 1,
            }
            lot['items'].append(item)
        
        tender['lots'].append(lot)

def generate_bidders(tender):
    """Генерация участников для каждого лота"""
    for lot in tender['lots']:
        # Выбираем участников (не включая заказчика)
        available = [c for c in CONTRACTORS if c[0] != tender['client_name']]
        selected = random.sample(available, min(BIDDERS_PER_TENDER, len(available)))
        
        for idx, contractor in enumerate(selected):
            bid_amount = round(lot['estimated_amount'] * random.uniform(0.85, 1.15), 2)
            
            bidder = {
                'contractor_name': contractor[0],
                'contractor_country': contractor[1],
                'bid_number': f'B-{tender["code"]}-L{lot["lot_number"]}-{idx+1:03d}',
                'status': 'submitted',
                'bid_amount': bid_amount,
                'currency': 'USD',
                'bid_bond_amount': round(bid_amount * tender['bid_bond_pct'] / 100, 2),
                'validity_days': random.choice([60, 90, 120]),
                'submission_date': tender.get('submission_deadline'),
                'is_winner': False,
                'items': [],
            }
            
            # Ценовые предложения по позициям
            for item in lot['items']:
                # Участник может дать цену выше или ниже сметы
                price_factor = random.uniform(0.80, 1.20)
                unit_price = round(item['estimated_unit_price'] * price_factor, 2)
                total_price = round(item['quantity'] * unit_price, 2)
                
                bid_item = {
                    'item_code': item['item_code'],
                    'description': item['description'],
                    'unit': item['unit'],
                    'quantity': item['quantity'],
                    'unit_price': unit_price,
                    'total_price': total_price,
                }
                bidder['items'].append(bid_item)
            
            lot.setdefault('bidders', []).append(bidder)
        
        # Определяем победителя (самая низкая цена)
        if lot['bidders']:
            winner = min(lot['bidders'], key=lambda b: b['bid_amount'])
            winner['is_winner'] = True
            winner['status'] = 'winner'
            lot['award_amount'] = winner['bid_amount']

def generate_evaluations(tender):
    """Генерация оценок для тендеров в статусе evaluation"""
    if tender['status'] != 'evaluation':
        return
    
    for lot in tender['lots']:
        for bidder in lot.get('bidders', []):
            # Техническая оценка
            tech_score = round(random.uniform(60, 100), 2)
            # Финансовая оценка (чем ниже цена, тем выше балл)
            min_bid = min(b['bid_amount'] for b in lot['bidders'])
            fin_score = round(100 * min_bid / bidder['bid_amount'], 2)
            # Общая оценка (техника 40% + финансы 60%)
            combined = round(tech_score * 0.4 + fin_score * 0.6, 2)
            
            bidder['evaluations'] = [
                {'type': 'technical', 'score': tech_score, 'max_score': 100, 'weight': 40.0},
                {'type': 'financial', 'score': fin_score, 'max_score': 100, 'weight': 60.0},
                {'type': 'combined', 'score': combined, 'max_score': 100, 'weight': 100.0},
            ]

def generate_all():
    """Генерация всех данных"""
    tenders = []
    for i in range(1, TENDER_COUNT + 1):
        tender = generate_tender(i)
        generate_lots(tender, lot_count=random.randint(1, 3))
        generate_bidders(tender)
        generate_evaluations(tender)
        tenders.append(tender)
    return tenders

# ============================================================================
# Вывод
# ============================================================================
def to_json(tenders):
    return json.dumps(tenders, ensure_ascii=False, indent=2, default=str)

def to_sql(tenders):
    """Генерация SQL-вставок"""
    lines = []
    lines.append("-- ============================================================================")
    lines.append("-- Тестовые данные: Тендеры")
    lines.append(f"-- Сгенерировано: {datetime.now().isoformat()}")
    lines.append("-- ============================================================================")
    lines.append("")
    lines.append("BEGIN;")
    lines.append("")
    
    # Вставка тендеров
    lines.append("-- Тендеры")
    for t in tenders:
        published = f"'{t['published_at']}'" if t['published_at'] else 'NULL'
        deadline = f"'{t['submission_deadline']}'" if t['submission_deadline'] else 'NULL'
        bid_open = f"'{t['bid_open_date']}'" if t['bid_open_date'] else 'NULL'
        
        lines.append(f"""
INSERT INTO tenders (code, name, description, tender_type, status, budget_amount, currency,
                     published_at, submission_deadline, bid_open_date,
                     bid_bond_pct, performance_bond_pct, advance_payment_pct,
                     procurement_method, funding_source)
VALUES ('{t['code']}', '{t['name'].replace("'", "''")}', '{t['description'].replace("'", "''")}',
        '{t['tender_type']}', '{t['status']}', {t['budget_amount']}, '{t['currency']}',
        {published}, {deadline}, {bid_open},
        {t['bid_bond_pct']}, {t['performance_bond_pct']}, {t['advance_payment_pct']},
        '{t['procurement_method']}', '{t['funding_source']}');
""")
    
    lines.append("")
    lines.append("COMMIT;")
    
    return '\n'.join(lines)

def to_summary(tenders):
    """Краткая сводка"""
    total_budget = sum(t['budget_amount'] for t in tenders)
    by_status = {}
    for t in tenders:
        by_status[t['status']] = by_status.get(t['status'], 0) + 1
    
    lines = []
    lines.append("=" * 60)
    lines.append("📊 СГЕНЕРИРОВАННЫЕ ТЕНДЕРЫ")
    lines.append("=" * 60)
    lines.append(f"Всего тендеров: {len(tenders)}")
    lines.append(f"Общий бюджет: ${total_budget:,.2f}")
    lines.append(f"По статусам: {', '.join(f'{k}: {v}' for k, v in by_status.items())}")
    lines.append("")
    
    for t in tenders:
        lines.append(f"  📋 {t['code']} — {t['name'][:60]}")
        lines.append(f"     Статус: {t['status']} | Бюджет: ${t['budget_amount']:,.2f}")
        lines.append(f"     Заказчик: {t['client_name']} | Тип: {t['tender_type']}")
        lines.append(f"     Лотов: {len(t['lots'])}")
        
        for lot in t['lots']:
            bidders = len(lot.get('bidders', []))
            winner = next((b for b in lot.get('bidders', []) if b.get('is_winner')), None)
            winner_info = f" | Победитель: {winner['contractor_name']} (${winner['bid_amount']:,.2f})" if winner else ""
            lines.append(f"       📦 Лот {lot['lot_number']}: {lot['name']} — ${lot['estimated_amount']:,.2f} | {bidders} участников{winner_info}")
        
        lines.append("")
    
    return '\n'.join(lines)


# ============================================================================
# Main
# ============================================================================
if __name__ == '__main__':
    tenders = generate_all()
    
    if '--json' in sys.argv:
        print(to_json(tenders))
    elif '--sql' in sys.argv:
        print(to_sql(tenders))
    elif '--summary' in sys.argv:
        print(to_summary(tenders))
    else:
        # По умолчанию — сводка
        print(to_summary(tenders))
        print()
        print("Для JSON: python3 generate_tenders.py --json")
        print("Для SQL:  python3 generate_tenders.py --sql")
