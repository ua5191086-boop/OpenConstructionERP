#!/usr/bin/env python3
"""Генератор тестовых HR-данных для OpenConstructionERP"""
import json, random, sys
from datetime import datetime, timedelta

EMPLOYEE_COUNT = 20
DEPARTMENTS = [
    ('DIR', 'Дирекция'), ('PMO', 'Проектный офис'), ('ENG', 'Инженерный отдел'),
    ('PROD', 'Производственный отдел'), ('PUR', 'Отдел закупок'), ('FIN', 'Финансовый отдел'),
    ('HR', 'Отдел кадров'), ('HSE', 'Охрана труда'), ('QA', 'Контроль качества'),
    ('LOG', 'Логистика'), ('LEG', 'Юридический отдел'), ('IT', 'IT-отдел'),
]
POSITIONS = {
    'DIR': ['Генеральный директор', 'Технический директор', 'Финансовый директор'],
    'PMO': ['Руководитель проектов', 'Менеджер проектов', 'Планировщик'],
    'ENG': ['Главный инженер', 'Инженер-строитель', 'Инженер-геотехник', 'Инженер-геодезист', 'Инженер-конструктор'],
    'PROD': ['Начальник участка', 'Производитель работ', 'Мастер СМР', 'Бригадир'],
    'PUR': ['Начальник отдела закупок', 'Специалист по закупкам', 'Кладовщик'],
    'FIN': ['Главный бухгалтер', 'Бухгалтер', 'Экономист', 'Финансовый аналитик'],
    'HR': ['Начальник отдела кадров', 'HR-специалист', 'Инспектор по кадрам'],
    'HSE': ['Инженер по охране труда', 'Специалист по промбезопасности'],
    'QA': ['Инженер ОТК', 'Лаборант'],
    'LOG': ['Начальник логистики', 'Логист', 'Водитель'],
    'LEG': ['Юрисконсульт'],
    'IT': ['Системный администратор', 'Разработчик'],
}
FIRST_NAMES_M = ['Алексей','Дмитрий','Сергей','Андрей','Владимир','Иван','Михаил','Николай','Павел','Александр','Олег','Евгений','Виктор','Константин','Максим']
FIRST_NAMES_F = ['Елена','Ольга','Наталья','Ирина','Анна','Татьяна','Светлана','Мария','Екатерина','Юлия']
LAST_NAMES = ['Иванов','Петров','Сидоров','Кузнецов','Смирнов','Попов','Васильев','Зайцев','Павлов','Соколов','Михайлов','Федоров','Морозов','Волков','Алексеев','Лебедев','Семенов','Егоров','Козлов','Новиков']

def gen_employees(count):
    employees = []
    for i in range(count):
        is_male = i < count * 0.7
        first = random.choice(FIRST_NAMES_M if is_male else FIRST_NAMES_F)
        last = random.choice(LAST_NAMES)
        full = f'{last} {first}'
        
        dept_code = random.choice([d[0] for d in DEPARTMENTS])
        dept_name = next(d[1] for d in DEPARTMENTS if d[0] == dept_code)
        position = random.choice(POSITIONS[dept_code])
        
        hire = datetime.now() - timedelta(days=random.randint(30, 2000))
        salary = round(random.uniform(1500, 15000), 2)
        
        status = 'active' if i < count - 2 else 'terminated'
        
        employees.append({
            'code': f'E-2026-{i+1:03d}',
            'name': full,
            'position': position,
            'department': dept_name,
            'dept_code': dept_code,
            'status': status,
            'hire_date': hire.strftime('%Y-%m-%d'),
            'salary': salary,
            'type': random.choice(['full_time', 'full_time', 'full_time', 'contract']),
        })
    return employees

def summary(emps):
    lines = []
    lines.append("=" * 60)
    lines.append("👥 СГЕНЕРИРОВАННЫЕ СОТРУДНИКИ")
    lines.append("=" * 60)
    lines.append(f"Всего: {len(emps)}")
    lines.append(f"Активных: {sum(1 for e in emps if e['status']=='active')}")
    lines.append(f"Уволенных: {sum(1 for e in emps if e['status']=='terminated')}")
    total_salary = sum(e['salary'] for e in emps if e['status']=='active')
    lines.append(f"ФОТ: ${total_salary:,.2f}/мес")
    lines.append("")
    depts = {}
    for e in emps:
        if e['status'] == 'active':
            depts[e['department']] = depts.get(e['department'], 0) + 1
    lines.append("По отделам:")
    for d, c in sorted(depts.items(), key=lambda x: -x[1]):
        lines.append(f"  {d}: {c}")
    lines.append("")
    for e in emps:
        lines.append(f"  {e['code']} — {e['name']} — {e['position']} — {e['department']} — ${e['salary']:,.2f} — {e['status']}")
    return '\n'.join(lines)

if __name__ == '__main__':
    emps = gen_employees(EMPLOYEE_COUNT)
    if '--json' in sys.argv:
        print(json.dumps(emps, ensure_ascii=False, indent=2))
    else:
        print(summary(emps))
