#!/usr/bin/env python3
"""Генератор тестовых финансовых данных для OpenConstructionERP"""
import json, random, sys
from datetime import datetime, timedelta

PROJECTS = ['Метро Алматы — Барлык', 'Ливнёвая канализация Алматы', 'Микротоннель Астана', 'Ж/Д ветка Туркестан']
MONTHS = ['Янв','Фев','Мар','Апр','Май','Июн','Июл','Авг','Сен','Окт','Ноя','Дек']

def gen_finance():
    data = []
    for p in PROJECTS:
        budget = round(random.uniform(10e6, 500e6), 2)
        # Cash flow по месяцам
        cf_plan = []
        cf_actual = []
        for m in range(12):
            plan = round(budget / 12 * random.uniform(0.5, 1.5), 2)
            actual = round(plan * random.uniform(0.7, 1.1), 2)
            cf_plan.append(plan)
            cf_actual.append(actual)
        
        total_plan = sum(cf_plan)
        total_actual = sum(cf_actual)
        variance = round(total_plan - total_actual, 2)
        variance_pct = round(variance / total_plan * 100, 2) if total_plan else 0
        
        data.append({
            'project': p,
            'budget': budget,
            'currency': 'USD',
            'cash_flow': {
                'months': MONTHS,
                'plan': cf_plan,
                'actual': cf_actual,
            },
            'total_plan': total_plan,
            'total_actual': total_actual,
            'variance': variance,
            'variance_pct': variance_pct,
            'status': 'green' if variance_pct < 5 else ('yellow' if variance_pct < 15 else 'red'),
        })
    return data

def summary(data):
    lines = []
    lines.append("=" * 60)
    lines.append("💰 ФИНАНСОВАЯ СВОДКА")
    lines.append("=" * 60)
    total_budget = sum(d['budget'] for d in data)
    total_plan = sum(d['total_plan'] for d in data)
    total_actual = sum(d['total_actual'] for d in data)
    lines.append(f"Общий бюджет: ${total_budget:,.2f}")
    lines.append(f"План: ${total_plan:,.2f}")
    lines.append(f"Факт: ${total_actual:,.2f}")
    lines.append(f"Отклонение: ${total_plan - total_actual:,.2f}")
    lines.append("")
    for d in data:
        lines.append(f"  {d['project']}")
        lines.append(f"    Бюджет: ${d['budget']:,.2f}")
        lines.append(f"    План: ${d['total_plan']:,.2f} | Факт: ${d['total_actual']:,.2f}")
        lines.append(f"    Отклонение: {d['variance_pct']}% — {d['status']}")
    return '\n'.join(lines)

if __name__ == '__main__':
    data = gen_finance()
    if '--json' in sys.argv:
        print(json.dumps(data, ensure_ascii=False, indent=2, default=str))
    else:
        print(summary(data))
