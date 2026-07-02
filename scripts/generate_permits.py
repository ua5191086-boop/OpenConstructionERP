#!/usr/bin/env python3
"""
Generator: Permits Module (V030)
Generates: regulatory_bodies, permit_applications, permit_documents,
           permit_inspections, permit_renewals, permit_conditions
"""
import json, random, sys
from datetime import datetime, timedelta

BODY_COUNT = 6
APP_COUNT = 15
DOC_COUNT = 25
INSP_COUNT = 12
RENEWAL_COUNT = 8
CONDITION_COUNT = 20

BODIES = [
    ('Комитет по делам строительства МИИР РК', 'GASK', 'Республиканский'),
    ('Управление архитектуры г. Алматы', 'ALM-ARCH', 'Региональный'),
    ('Департамент экологии г. Алматы', 'ALM-ECO', 'Региональный'),
    ('Комитет промышленной безопасности МЧС РК', 'PROM-BEZ', 'Республиканский'),
    ('Департамент по ЧС г. Алматы', 'ALM-EMER', 'Региональный'),
    ('СЭС г. Алматы', 'ALM-SES', 'Региональный'),
    ('Управление транспорта г. Алматы', 'ALM-TRANS', 'Региональный'),
]
APP_TYPES = ['construction', 'demolition', 'excavation', 'environmental', 'safety', 'traffic_management', 'utility', 'special']
STATUSES = ['draft', 'submitted', 'under_review', 'approved', 'rejected', 'expired', 'suspended']
INSP_TYPES = ['initial', 'periodic', 'follow_up', 'complaint', 'final']
INSP_STATUSES = ['scheduled', 'completed', 'failed', 'deferred']
COND_TYPES = ['pre_requisite', 'ongoing', 'reporting', 'mitigation']

def rand(min_v, max_v):
    return round(random.uniform(min_v, max_v), 2)

def pick(values):
    return random.choice(values)

def generate_data():
    bodies = []
    for i in range(BODY_COUNT):
        name, code, juris = BODIES[i % len(BODIES)]
        bodies.append({
            'id': f'rb{i:03d}',
            'body_name': name,
            'body_code': code,
            'jurisdiction': juris,
            'contact_info': json.dumps({'phone': f'+7 (727) {random.randint(100,999)}-{random.randint(10,99)}-{random.randint(10,99)}', 'email': f'{code.lower()}@gov.kz'}),
            'website': f'https://{code.lower()}.gov.kz',
            'notes': f'Regulatory body #{i+1}',
        })

    apps = []
    for i in range(APP_COUNT):
        apps.append({
            'id': f'app{i:03d}',
            'project_id': f'p{random.randint(1,5):03d}',
            'body_id': pick(bodies)['id'],
            'application_number': f'PERM-{i+1:04d}',
            'application_type': pick(APP_TYPES),
            'title': f'{pick(APP_TYPES).replace("_"," ").title()} Permit Application #{i+1}',
            'description': f'Description for {pick(APP_TYPES).replace("_"," ")} works',
            'submission_date': (datetime.now() - timedelta(days=random.randint(1, 365))).strftime('%Y-%m-%d'),
            'decision_date': None if random.random() < 0.3 else (datetime.now() - timedelta(days=random.randint(0, 60))).strftime('%Y-%m-%d'),
            'valid_from': None,
            'valid_until': None,
            'fee_amount': rand(1000, 50000),
            'fee_currency': 'KZT',
            'status': pick(STATUSES),
            'assigned_to': pick(['Иванов И.И.', 'Петров П.П.', 'Сидоров С.С.']),
        })
        if apps[-1]['status'] in ['approved', 'under_review']:
            apps[-1]['valid_from'] = (datetime.now() - timedelta(days=random.randint(1, 180))).strftime('%Y-%m-%d')
            apps[-1]['valid_until'] = (datetime.now() + timedelta(days=random.randint(30, 730))).strftime('%Y-%m-%d')

    docs = []
    doc_types = ['application_form', 'site_plan', 'structural_calculations', 'environmental_impact', 'safety_plan', 'insurance_certificate', 'land_title', 'approval_letter']
    for i in range(DOC_COUNT):
        a = pick(apps)
        docs.append({
            'id': f'doc{i:03d}',
            'application_id': a['id'],
            'document_type': pick(doc_types),
            'document_name': f'{pick(doc_types).replace("_"," ").title()} - {i+1}',
            'file_path': f'/documents/permits/{a["id"]}/{pick(doc_types)}_{i}.pdf',
            'status': pick(['uploaded', 'verified', 'rejected', 'resubmitted']),
            'uploaded_by': pick(['Иванов И.И.', 'Петров П.П.', 'Сидоров С.С.']),
        })

    inspections = []
    for i in range(INSP_COUNT):
        a = pick(apps)
        inspections.append({
            'id': f'ins{i:03d}',
            'application_id': a['id'],
            'inspection_type': pick(INSP_TYPES),
            'scheduled_date': (datetime.now() + timedelta(days=random.randint(-30, 60))).strftime('%Y-%m-%d'),
            'actual_date': None if random.random() < 0.3 else (datetime.now() + timedelta(days=random.randint(-30, 30))).strftime('%Y-%m-%d'),
            'inspector_name': pick(['Главный инспектор А.', 'Инспектор Б.', 'Старший инспектор В.', 'Инспектор Г.']),
            'findings': pick(['No violations found', 'Minor issues noted', 'Major non-compliance identified', 'Requires follow-up']),
            'result': pick(['passed', 'conditional_pass', 'failed']),
            'status': pick(INSP_STATUSES),
        })

    renewals = []
    for i in range(RENEWAL_COUNT):
        a = pick(apps)
        renewals.append({
            'id': f'ren{i:03d}',
            'application_id': a['id'],
            'renewal_number': f'REN-{i+1:04d}',
            'submitted_date': (datetime.now() - timedelta(days=random.randint(1, 90))).strftime('%Y-%m-%d'),
            'new_valid_until': (datetime.now() + timedelta(days=random.randint(30, 365))).strftime('%Y-%m-%d'),
            'fee_amount': rand(1000, 20000),
            'status': pick(['submitted', 'approved', 'rejected', 'pending_payment']),
        })

    conditions = []
    for i in range(CONDITION_COUNT):
        a = pick(apps)
        conditions.append({
            'id': f'cnd{i:03d}',
            'application_id': a['id'],
            'condition_type': pick(COND_TYPES),
            'description': pick([
                'Submit monthly progress reports', 'Maintain public liability insurance',
                'Install noise barriers', 'Implement dust control measures',
                'Archaeological monitoring required', 'Traffic management plan required',
                'Environmental monitoring quarterly', 'Submit as-built drawings',
                'Maintain site access for inspectors', 'Provide financial guarantee',
                'Restrict working hours', 'Protect existing utilities',
                'Submit waste management plan', 'Install security fencing',
            ]),
            'due_date': (datetime.now() + timedelta(days=random.randint(7, 365))).strftime('%Y-%m-%d'),
            'status': pick(['pending', 'compliant', 'overdue', 'waived', 'breached']),
        })

    return {
        'regulatory_bodies': bodies,
        'permit_applications': apps,
        'permit_documents': docs,
        'permit_inspections': inspections,
        'permit_renewals': renewals,
        'permit_conditions': conditions,
    }

def main():
    data = generate_data()
    if '--json' in sys.argv:
        print(json.dumps(data, ensure_ascii=False, indent=2))
    elif '--sql' in sys.argv:
        for table, rows in data.items():
            for row in rows:
                cols = ', '.join(row.keys())
                vals = ', '.join(f"'{v}'" if isinstance(v, str) else str(v) for v in row.values())
                print(f"INSERT INTO {table} ({cols}) VALUES ({vals});")
    else:
        print(json.dumps({k: len(v) for k, v in data.items()}, ensure_ascii=False))

if __name__ == '__main__':
    main()