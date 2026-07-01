#!/usr/bin/env python3
"""
Генератор тестовых договоров для OpenConstructionERP
"""
import json, random, sys
from datetime import datetime, timedelta

CONTRACT_COUNT = 6
CONTRACTORS = [
    'CAI Interbudmontazh GmbH', 'ТОО "Алматыметрострой"', 'АО "НК "КТЖ"',
    'ТОО "СтройИнвест Групп"', 'China Railway Construction Corp.', 'Vinci Construction',
    'Strabag SE', 'Porr AG', 'Astaldi S.p.A.', 'Gülermak A.Ş.',
    'ТОО "Базис-А"', 'ТОО "Курылыс-Монтаж"', 'ТОО "BI-Group"',
]

CONTRACT_TYPES = ['lump_sum', 'unit_price', 'cost_plus', 'design_build', 'epc']
STATUSES = ['draft', 'negotiation', 'signed', 'active', 'completed']
FUNDING = ['Госбюджет РК', 'Местный бюджет', 'ГЧП', 'Грант ЕБРР', 'Частные']

CONTRACT_NAMES = [
    'Строительство участка метро {section}',
    'Поставка и монтаж ТПМК {section}',
    'Электроснабжение и СЦБ {section}',
    'Вентиляция и сантехника {section}',
    'Отделочные работы {section}',
    'Путевые работы {section}',
    'Генеральный подряд {section}',
    'Инженерные изыскания {section}',
    'Разработка ПСД {section}',
    'Авторский надзор {section}',
]

SECTIONS = [
    'ПК 0+000 — ПК 30+000', 'ПК 30+000 — ПК 60+000', 'ПК 60+000 — ПК 90+000',
    'ПК 90+000 — ПК 120+000', 'Станция Центральная', 'Станция Восточная',
    'Станция Западная', 'Депо и инфраструктура',
]

def rand(min_v, max_v):
    return random.uniform(min_v, max_v)

def generate_contracts(count):
    contracts = []
    for i in range(count):
        section = random.choice(SECTIONS)
        name = random.choice(CONTRACT_NAMES).format(section=section)
        ctype = random.choice(CONTRACT_TYPES)
        
        if i < 2:
            status = 'draft'
        elif i < 4:
            status = 'signed'
        else:
            status = random.choice(['active', 'completed'])
        
        amount = round(rand(500000, 50000000), 2)
        start = datetime.now() - timedelta(days=random.randint(30, 365))
        duration = random.randint(180, 1095)
        end = start + timedelta(days=duration)
        
        client = random.choice(CONTRACTORS)
        contractor = random.choice([c for c in CONTRACTORS if c != client])
        
        milestones = []
        ms_count = random.randint(3, 8)
        for m in range(ms_count):
            ms_date = start + timedelta(days=int(duration * (m + 1) / ms_count))
            ms_amount = round(amount * rand(0.05, 0.25), 2)
            ms_status = 'completed' if ms_date < datetime.now() else ('pending' if status != 'completed' else 'completed')
            milestones.append({
                'number': m + 1,
                'name': f'Этап {m+1}: {random.choice(["Проектирование", "Закупка материалов", "Мобилизация", "Строительство", "Монтаж", "Пусконаладка", "Сдача"])}',
                'planned_date': ms_date.strftime('%Y-%m-%d'),
                'amount': ms_amount,
                'amount_pct': round(ms_amount / amount * 100, 2),
                'status': ms_status,
            })
        
        payments = []
        for m in milestones:
            if m['status'] == 'completed':
                payments.append({
                    'number': f'P-{i+1:03d}-{m["number"]:02d}',
                    'date': m['planned_date'],
                    'amount': m['amount'],
                    'type': 'progress',
                    'status': 'confirmed',
                })
        
        contracts.append({
            'code': f'C-2026-{i+1:03d}',
            'name': name,
            'contract_type': ctype,
            'status': status,
            'client': client,
            'contractor': contractor,
            'amount': amount,
            'currency': 'USD',
            'advance_pct': round(rand(0, 30), 2),
            'start_date': start.strftime('%Y-%m-%d'),
            'end_date': end.strftime('%Y-%m-%d'),
            'duration_days': duration,
            'funding_source': random.choice(FUNDING),
            'milestones': milestones,
            'payments': payments,
            'total_paid': round(sum(p['amount'] for p in payments), 2),
        })
    return contracts

def summary(contracts):
    lines = []
    lines.append("=" * 60)
    lines.append("📋 СГЕНЕРИРОВАННЫЕ ДОГОВОРЫ")
    lines.append("=" * 60)
    total = sum(c['amount'] for c in contracts)
    total_paid = sum(c['total_paid'] for c in contracts)
    lines.append(f"Всего: {len(contracts)} | Сумма: ${total:,.2f} | Оплачено: ${total_paid:,.2f}")
    lines.append("")
    for c in contracts:
        pct = round(c['total_paid'] / c['amount'] * 100, 1) if c['amount'] else 0
        lines.append(f"  {c['code']} — {c['name'][:55]}")
        lines.append(f"    {c['status']} | ${c['amount']:,.2f} | {c['contract_type']} | {c['contractor'][:30]}")
        lines.append(f"    Этапов: {len(c['milestones'])} | Оплачено: ${c['total_paid']:,.2f} ({pct}%)")
        lines.append("")
    return '\n'.join(lines)

if __name__ == '__main__':
    contracts = generate_contracts(CONTRACT_COUNT)
    if '--json' in sys.argv:
        print(json.dumps(contracts, ensure_ascii=False, indent=2, default=str))
    elif '--sql' in sys.argv:
        print("-- SQL generation not implemented for V003 yet")
    else:
        print(summary(contracts))
