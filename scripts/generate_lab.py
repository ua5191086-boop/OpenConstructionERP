#!/usr/bin/env python3
"""
Generator: Laboratory Module (V029)
Generates: material_testing, concrete_tests, soil_tests, steel_tests,
           lab_certificates, lab_equipment, sampling_log
"""
import json, random, sys
from datetime import datetime, timedelta

TEST_COUNT = 20
CONCRETE_COUNT = 12
SOIL_COUNT = 10
STEEL_COUNT = 8
CERT_COUNT = 6
EQ_COUNT = 8
SAMPLE_COUNT = 15

MATERIAL_TYPES = ['concrete', 'steel', 'soil', 'aggregate', 'asphalt']
TEST_TYPES_MAP = {
    'concrete': ['compression', 'slump', 'air_content', 'temperature', 'density', 'split_tensile'],
    'soil': ['proctor', 'sieve', 'atterberg', 'triaxial', 'direct_shear', 'cbr', 'permeability', 'consolidation'],
    'steel': ['tensile', 'bend', 'hardness', 'impact', 'chemical', 'ultrasonic', 'magnetic_particle', 'radiographic'],
    'aggregate': ['sieve', 'specific_gravity', 'water_absorption', 'abrasion', 'flakiness', 'elongation'],
    'asphalt': ['marshall', 'extraction', 'bulk_density', 'air_voids', 'flow', 'stability'],
}
STATUSES = ['pending', 'in_progress', 'completed', 'rejected']
CONCRETE_GRADES = ['C16/20', 'C20/25', 'C25/30', 'C30/37', 'C35/45', 'C40/50', 'C45/55', 'C50/60']
SOIL_TYPES = ['clay', 'silt', 'sand', 'gravel', 'till', 'loam', 'peat', 'fill']
STEEL_GRADES = ['S235', 'S275', 'S355', 'S420', 'S460', 'B500B', 'B500C', 'St52']

EQUIPMENT_NAMES = [
    'Compression Testing Machine 2000kN', 'Universal Testing Machine 600kN', 'Proctor Compaction Apparatus',
    'Sieve Shaker', 'Concrete Slump Cone', 'Marshall Stability Apparatus',
    'Triaxial Test System', 'CBR Test Apparatus',
]
EQ_STATUSES = ['operational', 'calibration_due', 'maintenance', 'out_of_service']

def rand(min_v, max_v):
    return round(random.uniform(min_v, max_v), 2)

def pick(values):
    return random.choice(values)

def generate_data():
    tests = []
    for i in range(TEST_COUNT):
        mt = pick(MATERIAL_TYPES)
        tt = pick(TEST_TYPES_MAP.get(mt, ['general']))
        tests.append({
            'id': f't{i:03d}',
            'project_id': f'p{random.randint(1,5):03d}',
            'test_number': f'TST-{i+1:04d}',
            'material_type': mt,
            'test_type': tt,
            'specification': f'{pick(["ASTM","EN","GOST","BS","ISO"])} {random.randint(100,9999)}',
            'sample_id': f'SMP-{random.randint(1000,9999)}',
            'sampling_date': (datetime.now() - timedelta(days=random.randint(1, 90))).strftime('%Y-%m-%d'),
            'test_date': (datetime.now() - timedelta(days=random.randint(0, 30))).strftime('%Y-%m-%d'),
            'result': f'{rand(10, 500)} {pick(["MPa","kN","mm","%","kg/m³","kJ/m²"])}',
            'status': pick(STATUSES),
            'tested_by': pick(['Иванов И.И.', 'Петров П.П.', 'Сидоров С.С.', 'Козлов А.А.', 'Смирнова Е.В.']),
            'approved_by': pick(['Главный инженер', 'Начальник лаборатории', 'Технический директор']),
        })

    concrete_tests = []
    for i in range(CONCRETE_COUNT):
        concrete_tests.append({
            'id': f'ct{i:03d}',
            'project_id': f'p{random.randint(1,5):03d}',
            'test_number': f'CT-{i+1:04d}',
            'concrete_grade': pick(CONCRETE_GRADES),
            'design_strength_mpa': rand(20, 60),
            'actual_strength_mpa': rand(15, 70),
            'slump_mm': random.randint(20, 220),
            'air_content_pct': rand(1.0, 8.0),
            'temperature_c': rand(5, 35),
            'density_kgm3': random.randint(2200, 2600),
            'water_cement_ratio': round(random.uniform(0.35, 0.65), 2),
            'curing_days': random.choice([3, 7, 14, 28, 56]),
            'sample_location': f'ПК {random.randint(0, 120)}+{random.randint(0, 99):02d}',
            'status': pick(STATUSES),
        })

    soil_tests = []
    for i in range(SOIL_COUNT):
        soil_tests.append({
            'id': f'st{i:03d}',
            'project_id': f'p{random.randint(1,5):03d}',
            'test_number': f'ST-{i+1:04d}',
            'soil_type': pick(SOIL_TYPES),
            'moisture_content_pct': rand(5, 40),
            'dry_density_kgm3': random.randint(1400, 2200),
            'liquid_limit_pct': rand(20, 80),
            'plastic_limit_pct': rand(10, 40),
            'plasticity_index': rand(5, 50),
            'cbr_value': rand(2, 60),
            'internal_friction_deg': random.randint(20, 45),
            'cohesion_kpa': rand(0, 200),
            'permeability_m': f'{rand(1e-10, 1e-3):.2e}',
            'depth_m': rand(0.5, 30),
            'location': f'BH-{random.randint(1,20):02d}',
            'status': pick(STATUSES),
        })

    steel_tests = []
    for i in range(STEEL_COUNT):
        steel_tests.append({
            'id': f'stl{i:03d}',
            'project_id': f'p{random.randint(1,5):03d}',
            'test_number': f'STL-{i+1:04d}',
            'steel_grade': pick(STEEL_GRADES),
            'diameter_mm': random.choice([6, 8, 10, 12, 16, 20, 25, 32, 40]),
            'yield_strength_mpa': random.randint(235, 500),
            'ultimate_strength_mpa': random.randint(400, 700),
            'elongation_pct': rand(10, 30),
            'bend_test': pick(['passed', 'passed', 'passed', 'failed']),
            'weld_test': pick(['passed', 'passed', 'passed', 'passed', 'failed']),
            'heat_number': f'H{random.randint(10000,99999)}',
            'supplier': pick(['ArcelorMittal', 'Evraz', 'MMK', 'NLMK', 'Severstal']),
            'status': pick(STATUSES),
        })

    certs = []
    for i in range(CERT_COUNT):
        certs.append({
            'id': f'crt{i:03d}',
            'project_id': f'p{random.randint(1,5):03d}',
            'certificate_number': f'CERT-{i+1:04d}',
            'certificate_type': pick(['material', 'concrete_mix', 'weld_procedure', 'calibration', 'personnel']),
            'issued_by': pick(['Казахстанский центр сертификации', 'SGS Kazakhstan', 'TÜV Rheinland', 'Bureau Veritas', 'ICQC']),
            'issue_date': (datetime.now() - timedelta(days=random.randint(30, 365))).strftime('%Y-%m-%d'),
            'expiry_date': (datetime.now() + timedelta(days=random.randint(30, 730))).strftime('%Y-%m-%d'),
            'scope': f'{pick(["All","Structural","Welding","Electrical","Geotechnical"])} works',
            'status': pick(['valid', 'expired', 'suspended']),
        })

    equipment = []
    for i in range(EQ_COUNT):
        equipment.append({
            'id': f'eq{i:03d}',
            'equipment_name': EQUIPMENT_NAMES[i % len(EQUIPMENT_NAMES)],
            'equipment_code': f'LAB-EQ-{i+1:04d}',
            'manufacturer': pick(['Matest S.p.A.', 'Controls Group', 'ELE International', 'UTEST', 'Humboldt Mfg.']),
            'model': f'M-{random.randint(100,999)}',
            'serial_number': f'SN-{random.randint(10000,99999)}',
            'last_calibration_date': (datetime.now() - timedelta(days=random.randint(30, 365))).strftime('%Y-%m-%d'),
            'next_calibration_date': (datetime.now() + timedelta(days=random.randint(30, 365))).strftime('%Y-%m-%d'),
            'location': f'Lab Room {random.randint(1,5)}',
            'status': pick(EQ_STATUSES),
            'is_active': True,
        })

    samples = []
    for i in range(SAMPLE_COUNT):
        samples.append({
            'id': f'smp{i:03d}',
            'project_id': f'p{random.randint(1,5):03d}',
            'sample_number': f'SMP-{i+1:04d}',
            'material_type': pick(MATERIAL_TYPES),
            'sample_type': pick(['cube', 'cylinder', 'beam', 'bulk', 'core', 'disturbed', 'undisturbed']),
            'sampling_location': f'ПК {random.randint(0,120)}+{random.randint(0,99):02d}',
            'sampled_by': pick(['Лаборант А.', 'Лаборант Б.', 'Инженер В.', 'Техник Г.']),
            'sampling_date': (datetime.now() - timedelta(days=random.randint(1, 60))).strftime('%Y-%m-%d'),
            'sample_condition': pick(['good', 'fair', 'poor', 'damaged']),
            'test_id': pick(tests)['id'] if tests else None,
            'status': pick(['collected', 'in_transit', 'received', 'tested', 'discarded']),
        })

    return {
        'material_testing': tests,
        'concrete_tests': concrete_tests,
        'soil_tests': soil_tests,
        'steel_tests': steel_tests,
        'lab_certificates': certs,
        'lab_equipment': equipment,
        'sampling_log': samples,
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