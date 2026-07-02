#!/usr/bin/env python3
"""
Generator: Funding Module (V027)
Generates: funding_sources, tranches, drawdowns, covenants,
           multi_currency_rates, currency_hedges, guarantees, guarantee_claims
"""
import json, random, sys
from datetime import datetime, timedelta

SOURCE_COUNT = 8
TRANCHES_PER_SOURCE = 3
DRAWDOWNS_PER_SOURCE = 4
COVENANTS_PER_SOURCE = 2
RATE_COUNT = 6
HEDGE_COUNT = 4
GUARANTEE_COUNT = 5
CLAIMS_PER_GUARANTEE = 2

SOURCE_TYPES = ['bank', 'eca', 'investor', 'grant', 'internal']
SOURCE_NAMES = [
    'ЕБРР — Европейский банк реконструкции и развития',
    'АБР — Азиатский банк развития',
    'Госбюджет Республики Казахстан',
    'China Exim Bank',
    'ВТБ Капитал',
    'Halcrow Group Ltd',
    'Местный бюджет г. Алматы',
    'Фонд \"Самрук-Казына\"',
    'Japan International Cooperation Agency',
    'Deutsche Bank AG',
]
STATUSES = ['planned', 'active', 'disbursed', 'pending', 'cancelled']
CURRENCIES = ['USD', 'EUR', 'KZT', 'CNY', 'JPY', 'GBP']
COVENANT_TYPES = ['debt_service_coverage', 'loan_life_coverage', 'interest_coverage', 'leverage_ratio', 'minimum_equity']
HEDGE_TYPES = ['forward', 'swap', 'option', 'collar']
GUARANTEE_TYPES = ['bid_bond', 'performance', 'advance_payment', 'retention', 'warranty']

def rand(min_v, max_v):
    return round(random.uniform(min_v, max_v), 2)

def pick(values):
    return random.choice(values)

def generate_data():
    sources = []
    for i in range(SOURCE_COUNT):
        src = {
            'id': f'f{i:03d}',
            'project_id': f'p{i % 5 + 1:03d}',
            'source_type': pick(SOURCE_TYPES),
            'source_name': SOURCE_NAMES[i % len(SOURCE_NAMES)],
            'source_code': f'FS-{i+1:04d}',
            'description': f'{pick(SOURCE_TYPES).upper()} funding source #{i+1}',
            'commitment_amount': rand(500000, 50000000),
            'currency': pick(CURRENCIES),
            'status': pick(STATUSES),
            'is_active': True,
        }
        sources.append(src)

    tranches = []
    for s in sources:
        for t in range(TRANCHES_PER_SOURCE):
            tranches.append({
                'id': f'tr{s["id"]}-{t:02d}',
                'funding_source_id': s['id'],
                'project_id': s['project_id'],
                'tranche_name': f'Tranche {t+1} — {pick(["Initial","Interim","Final","Contingency","Performance"])}',
                'amount': rand(100000, s['commitment_amount']),
                'currency': s['currency'],
                'expected_date': (datetime.now() + timedelta(days=random.randint(30, 730))).strftime('%Y-%m-%d'),
                'actual_date': None if random.random() < 0.4 else (datetime.now() + timedelta(days=random.randint(-180, 30))).strftime('%Y-%m-%d'),
                'status': pick(STATUSES),
            })

    drawdowns = []
    for s in sources:
        for d in range(DRAWDOWNS_PER_SOURCE):
            drawdowns.append({
                'id': f'dw{s["id"]}-{d:02d}',
                'funding_source_id': s['id'],
                'project_id': s['project_id'],
                'drawdown_number': f'DD-{s["id"]}-{d+1:04d}',
                'amount': rand(50000, s['commitment_amount'] * 0.3),
                'currency': s['currency'],
                'drawdown_date': (datetime.now() + timedelta(days=random.randint(-365, 365))).strftime('%Y-%m-%d'),
                'purpose': pick(['Работы СМР', 'Оборудование', 'Проектирование', 'Зарплата', 'Материалы']),
                'status': pick(STATUSES),
            })

    covenants = []
    for s in sources:
        for _ in range(COVENANTS_PER_SOURCE):
            covenants.append({
                'id': f'cv{s["id"]}-{random.randint(0,99):02d}',
                'funding_source_id': s['id'],
                'project_id': s['project_id'],
                'covenant_type': pick(COVENANT_TYPES),
                'covenant_name': f'Covenant {pick(COVENANT_TYPES).replace("_"," ").title()}',
                'required_value': rand(1.0, 3.0),
                'current_value': rand(0.5, 4.0),
                'currency': s['currency'],
                'measurement_frequency': pick(['quarterly', 'semi_annual', 'annual']),
                'status': pick(['compliant', 'breached', 'waived', 'monitoring']),
            })

    rates = []
    for i in range(RATE_COUNT):
        base = pick(CURRENCIES)
        quote = pick([c for c in CURRENCIES if c != base])
        rates.append({
            'id': f'rate{i:03d}',
            'base_currency': base,
            'quote_currency': quote,
            'rate': rand(0.001, 500) if base == 'KZT' or quote == 'KZT' else rand(0.5, 2.0),
            'rate_date': (datetime.now() - timedelta(days=random.randint(0, 30))).strftime('%Y-%m-%d'),
            'source': pick(['ECB', 'NBK', 'Bloomberg', 'Reuters']),
            'is_active': True,
        })

    hedges = []
    for i in range(HEDGE_COUNT):
        hedges.append({
            'id': f'hdg{i:03d}',
            'project_id': f'p{random.randint(1,5):03d}',
            'hedge_type': pick(HEDGE_TYPES),
            'hedge_reference': f'HEDGE-{i+1:04d}',
            'notional_amount': rand(100000, 10000000),
            'currency_pair': f'{pick(CURRENCIES)}/{pick(CURRENCIES)}',
            'maturity_date': (datetime.now() + timedelta(days=random.randint(30, 730))).strftime('%Y-%m-%d'),
            'counterparty': SOURCE_NAMES[i % len(SOURCE_NAMES)],
            'status': pick(['active', 'matured', 'terminated']),
        })

    guarantees = []
    for i in range(GUARANTEE_COUNT):
        g = {
            'id': f'gt{i:03d}',
            'project_id': f'p{random.randint(1,5):03d}',
            'guarantee_number': f'BG-{i+1:04d}',
            'guarantee_type': pick(GUARANTEE_TYPES),
            'issuing_bank': SOURCE_NAMES[i % len(SOURCE_NAMES)],
            'beneficiary': f'Beneficiary Corp #{i+1}',
            'principal_amount': rand(50000, 5000000),
            'currency': pick(CURRENCIES),
            'issue_date': (datetime.now() - timedelta(days=random.randint(30, 365))).strftime('%Y-%m-%d'),
            'expiry_date': (datetime.now() + timedelta(days=random.randint(30, 730))).strftime('%Y-%m-%d'),
            'contract_reference': f'CON-{random.randint(1000,9999)}',
            'status': pick(['active', 'expired', 'claimed', 'cancelled']),
        }
        guarantees.append(g)

    claims = []
    for g in guarantees:
        for _ in range(CLAIMS_PER_GUARANTEE):
            claims.append({
                'id': f'cl{g["id"]}-{random.randint(0,99):02d}',
                'guarantee_id': g['id'],
                'claim_number': f'CLM-{random.randint(1000,9999)}',
                'claim_date': (datetime.now() - timedelta(days=random.randint(1, 180))).strftime('%Y-%m-%d'),
                'claim_amount': rand(10000, g['principal_amount'] * 0.5),
                'reason': pick(['Non-performance', 'Delay', 'Quality defect', 'Breach of contract', 'Advance payment recovery']),
                'status': pick(['submitted', 'under_review', 'approved', 'rejected', 'paid']),
                'settlement_date': None if random.random() < 0.5 else (datetime.now() - timedelta(days=random.randint(1, 30))).strftime('%Y-%m-%d'),
                'settlement_amount': None,
            })

    amendments = []
    for g in guarantees:
        for _ in range(random.randint(0, 2)):
            amendments.append({
                'id': f'am{g["id"]}-{random.randint(0,99):02d}',
                'guarantee_id': g['id'],
                'amendment_number': f'AMD-{random.randint(100,999)}',
                'amendment_date': (datetime.now() - timedelta(days=random.randint(1, 90))).strftime('%Y-%m-%d'),
                'description': pick(['Extension of expiry', 'Amount increase', 'Amount decrease', 'Beneficiary change', 'Text amendment']),
                'new_expiry_date': (datetime.now() + timedelta(days=random.randint(60, 800))).strftime('%Y-%m-%d'),
                'new_amount': rand(10000, 5000000),
            })

    return {
        'funding_sources': sources,
        'funding_tranches': tranches,
        'funding_drawdowns': drawdowns,
        'funding_covenants': covenants,
        'multi_currency_rates': rates,
        'currency_hedges': hedges,
        'guarantees': guarantees,
        'guarantee_claims': claims,
        'guarantee_amendments': amendments,
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