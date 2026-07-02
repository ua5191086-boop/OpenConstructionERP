#!/usr/bin/env python3
"""
Generator: Insurance Module (V031)
Generates: insurance_brokers, insurance_policies, insurance_coverage,
           insurance_premiums, insurance_claims, certificates_of_insurance
"""
import json, random, sys
from datetime import datetime, timedelta

BROKER_COUNT = 5
POLICY_COUNT = 12
COVERAGE_PER_POLICY = 3
PREMIUM_YEARS = 3
CLAIM_COUNT = 8
CERT_COUNT = 6

BROKERS = [
    ('AIG Kazakhstan', '+7 (727) 311-22-33', 'almaty@aig.kz'),
    ('Allianz Kazakhstan', '+7 (727) 244-55-66', 'info@allianz.kz'),
    ('СК \"Евразия\"', '+7 (727) 123-45-67', 'info@evrasia.kz'),
    ('СК \"Номад Иншуранс\"', '+7 (727) 987-65-43', 'info@nomad-ins.kz'),
    ('Marsh Kazakhstan', '+7 (727) 333-44-55', 'marsh@marsh.kz'),
    ('Zurich Insurance Kazakhstan', '+7 (727) 555-66-77', 'info@zurich.kz'),
]
POLICY_TYPES = ['construction_all_risk', 'third_party_liability', 'professional_indemnity', 'employer_liability', 'marine_cargo', 'motor_fleet', 'workers_compensation', 'environmental_liability']
COVERAGE_TYPES = ['property_damage', 'third_party_bodily', 'third_party_property', 'professional_fees', 'clean_up_costs', 'business_interruption', 'defence_costs', 'emergency_response']
STATUSES = ['active', 'expired', 'cancelled', 'pending_renewal']
CLAIM_STATUSES = ['submitted', 'adjusting', 'approved', 'rejected', 'settled', 'appealed']
PAYMENT_FREQUENCIES = ['annual', 'semi_annual', 'quarterly', 'monthly']

def rand(min_v, max_v):
    return round(random.uniform(min_v, max_v), 2)

def pick(values):
    return random.choice(values)

def generate_data():
    brokers = []
    for i in range(BROKER_COUNT):
        name, phone, email = BROKERS[i % len(BROKERS)]
        brokers.append({
            'id': f'b{i:03d}',
            'broker_name': name,
            'contact_person': pick(['Иванов И.И.', 'Петров П.П.', 'Сидоров С.С.', 'Козлов А.А.']),
            'email': email,
            'phone': phone,
            'address': f'г. Алматы, ул. {pick(["Абая","Достык","Саина","Жибек Жолы","Тимирязева"])}, д. {random.randint(1,200)}',
            'license_number': f'LIC-{random.randint(100,999)}-{random.randint(1000,9999)}',
            'notes': f'Insurance broker #{i+1}',
        })

    policies = []
    for i in range(POLICY_COUNT):
        b = pick(brokers)
        start = datetime.now() - timedelta(days=random.randint(30, 365))
        end = start + timedelta(days=365)
        policies.append({
            'id': f'p{i:03d}',
            'broker_id': b['id'],
            'project_id': f'p{random.randint(1,5):03d}',
            'policy_number': f'POL-{i+1:04d}',
            'policy_type': pick(POLICY_TYPES),
            'insurer': pick(['AIG Kazakhstan', 'Allianz Kazakhstan', 'СК \"Евразия\"', 'СК \"Номад Иншуранс\"', 'Zurich Insurance']),
            'insured_entity': f'Project Company #{random.randint(1,5)}',
            'sum_insured': rand(500000, 50000000),
            'currency': pick(['USD', 'EUR', 'KZT']),
            'deductible': rand(5000, 500000),
            'start_date': start.strftime('%Y-%m-%d'),
            'end_date': end.strftime('%Y-%m-%d'),
            'premium_amount': rand(10000, 500000),
            'premium_currency': 'USD',
            'status': pick(STATUSES),
            'notes': f'Policy #{i+1} through {b["broker_name"]}',
        })

    coverage = []
    for p in policies:
        for _ in range(COVERAGE_PER_POLICY):
            coverage.append({
                'id': f'cov{p["id"]}-{random.randint(0,99):02d}',
                'policy_id': p['id'],
                'coverage_type': pick(COVERAGE_TYPES),
                'description': f'Coverage for {pick(COVERAGE_TYPES).replace("_"," ")}',
                'limit_amount': rand(p['sum_insured'] * 0.3, p['sum_insured']),
                'deductible': rand(p['deductible'] * 0.5, p['deductible'] * 2),
                'is_active': True,
            })

    premiums = []
    for p in policies:
        for y in range(PREMIUM_YEARS):
            premiums.append({
                'id': f'prm{p["id"]}-y{p["start_date"][:4]}-{y}',
                'policy_id': p['id'],
                'installment_number': y + 1,
                'amount_due': rand(p['premium_amount'] * 0.3, p['premium_amount']),
                'amount_paid': None if random.random() < 0.3 else rand(10000, 500000),
                'due_date': (datetime.strptime(p['start_date'], '%Y-%m-%d') + timedelta(days=365 * y)).strftime('%Y-%m-%d'),
                'paid_date': None if random.random() < 0.3 else (datetime.now() - timedelta(days=random.randint(1, 60))).strftime('%Y-%m-%d'),
                'payment_frequency': pick(PAYMENT_FREQUENCIES),
                'status': pick(['due', 'paid', 'overdue']),
            })

    claims = []
    for i in range(CLAIM_COUNT):
        p = pick(policies)
        claims.append({
            'id': f'clm{i:03d}',
            'policy_id': p['id'],
            'project_id': p['project_id'],
            'claim_number': f'CLM-{i+1:04d}',
            'claim_date': (datetime.now() - timedelta(days=random.randint(1, 180))).strftime('%Y-%m-%d'),
            'incident_date': (datetime.now() - timedelta(days=random.randint(1, 200))).strftime('%Y-%m-%d'),
            'incident_type': pick(['accident', 'fire', 'flood', 'collapse', 'theft', 'vandalism', 'third_party_damage', 'professional_error']),
            'description': f'{pick(["Structural collapse","Fire in warehouse","Vehicle accident","Water damage","Equipment theft","Third party injury"])}',
            'estimated_amount': rand(10000, p['sum_insured'] * 0.5),
            'settled_amount': None if random.random() < 0.5 else rand(5000, 500000),
            'adjuster': pick(['Independent Adjuster Inc.', 'McLarens Kazakhstan', 'Crawford & Company', 'In-house adjuster']),
            'status': pick(CLAIM_STATUSES),
        })

    certs = []
    for i in range(CERT_COUNT):
        p = pick(policies)
        certs.append({
            'id': f'cert{i:03d}',
            'policy_id': p['id'],
            'certificate_number': f'COI-{i+1:04d}',
            'holder_name': f'Certificate Holder #{i+1}',
            'issue_date': (datetime.now() - timedelta(days=random.randint(1, 180))).strftime('%Y-%m-%d'),
            'valid_until': (datetime.now() + timedelta(days=random.randint(30, 365))).strftime('%Y-%m-%d'),
            'additional_insured': random.choice([True, False]),
            'status': pick(['valid', 'expired']),
        })

    return {
        'insurance_brokers': brokers,
        'insurance_policies': policies,
        'insurance_coverage': coverage,
        'insurance_premiums': premiums,
        'insurance_claims': claims,
        'certificates_of_insurance': certs,
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