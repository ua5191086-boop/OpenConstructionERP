#!/usr/bin/env python3
"""
Generator: Fleet Module (V032)
Generates: fleet_vehicles, fleet_drivers, fleet_fuel, fleet_maintenance,
           fleet_tracking, fleet_accidents, fleet_telematics
"""
import json, random, sys
from datetime import datetime, timedelta

VEHICLE_COUNT = 15
DRIVER_COUNT = 10
FUEL_COUNT = 40
MAINT_COUNT = 20
TRACKING_COUNT = 50
ACCIDENT_COUNT = 6
TELEMATICS_COUNT = 50

VEHICLE_TYPES = ['dump_truck', 'concrete_mixer', 'excavator', 'bulldozer', 'crane', 'pickup', 'van', 'bus', 'water_truck', 'fuel_truck']
MAKE_MODEL = {
    'dump_truck': [('Scania', 'G460'), ('Volvo', 'FH16'), ('MAN', 'TGX'), ('KamAZ', '6520')],
    'concrete_mixer': [('Scania', 'P360'), ('DAF', 'CF'), ('Силач', '584130')],
    'excavator': [('Caterpillar', '336'), ('Komatsu', 'PC300'), ('Hitachi', 'ZX350')],
    'bulldozer': [('Caterpillar', 'D6'), ('Komatsu', 'D155'), ('Liebherr', 'PR 736')],
    'crane': [('Liebherr', 'LTM 1050'), ('Tadano', 'ATF 70G'), ('Grove', 'GMK3050')],
    'pickup': [('Toyota', 'Hilux'), ('Mitsubishi', 'L200'), ('Ford', 'Ranger')],
    'van': [('Mercedes-Benz', 'Sprinter'), ('Ford', 'Transit'), ('Volkswagen', 'Crafter')],
    'bus': [('Yutong', 'ZK6129H'), ('ПАЗ', '3205')],
    'water_truck': [('КрАЗ', '65055'), ('MAZ', '6317')],
    'fuel_truck': [('Scania', 'R460'), ('Volvo', 'FE')],
}
FUEL_TYPES = ['diesel', 'gasoline', 'adblue']
MAINT_TYPES = ['routine', 'repair', 'emergency', 'tire', 'brake', 'engine', 'transmission']
ACCIDENT_TYPES = ['collision', 'rollover', 'fire', 'pedestrian', 'property_damage', 'single_vehicle']
SEVERITIES = ['minor', 'moderate', 'serious', 'fatal']
REGIONS = ['Алматы', 'Талдыкорган', 'Конаев', 'Капшагай', 'Узынагаш', 'Кеген', 'Жаркент', 'Чилик']

def rand(min_v, max_v):
    return round(random.uniform(min_v, max_v), 2)

def pick(values):
    return random.choice(values)

def generate_data():
    brands_models = []
    for _ in range(VEHICLE_COUNT):
        vt = pick(VEHICLE_TYPES)
        make, model = random.choice(MAKE_MODEL.get(vt, [('Unknown', 'X')]))
        brands_models.append((vt, make, model))

    vehicles = []
    for i in range(VEHICLE_COUNT):
        vt, make, model = brands_models[i]
        vehicles.append({
            'id': f'v{i:03d}',
            'project_id': f'p{random.randint(1,5):03d}',
            'vehicle_number': f'V-{i+1:04d}',
            'license_plate': f'{random.randint(100,999)}{random.choice(["AAA","BBB","ZZZ","ABC","XYZ"])}{random.randint(1,99):02d}',
            'vehicle_type': vt,
            'make': make,
            'model': model,
            'year': random.randint(2015, 2024),
            'vin': f'VIN{random.randint(10000000000000000, 99999999999999999)}',
            'fuel_type': pick(FUEL_TYPES),
            'fuel_capacity_l': random.choice([100, 200, 300, 400, 500, 600]),
            'registration_date': (datetime.now() - timedelta(days=random.randint(30, 1800))).strftime('%Y-%m-%d'),
            'status': pick(['active', 'in_maintenance', 'out_of_service', 'decommissioned']),
            'location': pick(REGIONS),
            'current_meter_km': random.randint(5000, 200000),
            'notes': f'Fleet vehicle #{i+1}',
        })

    drivers = []
    for i in range(DRIVER_COUNT):
        drivers.append({
            'id': f'drv{i:03d}',
            'full_name': pick(['Алиев А.А.', 'Бакиров Б.Б.', 'Давлетов Д.Д.', 'Ермеков Е.Е.', 'Жумабаев Ж.Ж.',
                              'Касымов К.К.', 'Нургалиев Н.Н.', 'Омаров О.О.', 'Сагындыков С.С.', 'Темиров Т.Т.',
                              'Утепов У.У.', 'Хасанов Х.Х.']),
            'license_number': f'LIC-DRV-{random.randint(1000,9999)}',
            'license_category': pick(['B', 'C', 'D', 'E', 'CE']),
            'phone': f'+7 (7{random.randint(0,9)}{random.randint(0,9)}) {random.randint(100,999)}-{random.randint(10,99)}-{random.randint(10,99)}',
            'email': f'driver{i+1}@project.kz',
            'assigned_vehicle_id': pick(vehicles)['id'] if random.random() < 0.7 else None,
            'hire_date': (datetime.now() - timedelta(days=random.randint(30, 1000))).strftime('%Y-%m-%d'),
            'medical_expiry': (datetime.now() + timedelta(days=random.randint(30, 365))).strftime('%Y-%m-%d'),
            'status': pick(['active', 'on_leave', 'suspended', 'terminated']),
        })

    fuel = []
    for i in range(FUEL_COUNT):
        v = pick(vehicles)
        fuel.append({
            'id': f'fl{i:03d}',
            'vehicle_id': v['id'],
            'refuel_date': (datetime.now() - timedelta(days=random.randint(1, 90))).strftime('%Y-%m-%d'),
            'quantity_l': random.randint(20, v['fuel_capacity_l']),
            'cost_per_unit': round(random.uniform(180, 280), 2),
            'total_cost': 0,
            'fuel_type': v['fuel_type'],
            'operator_name': pick(drivers)['full_name'] if drivers else 'Unknown',
            'odometer_km': random.randint(v['current_meter_km'] - 20000, v['current_meter_km'] + 5000),
            'station_name': pick(['Helios', 'Ala Too', 'PetroKazakhstan', 'Sinooil', 'Royal Petroleum']),
        })
        fuel[-1]['total_cost'] = round(fuel[-1]['quantity_l'] * fuel[-1]['cost_per_unit'], 2)

    maintenance = []
    for i in range(MAINT_COUNT):
        v = pick(vehicles)
        maintenance.append({
            'id': f'mnt{i:03d}',
            'vehicle_id': v['id'],
            'maintenance_type': pick(MAINT_TYPES),
            'description': f'{pick(MAINT_TYPES).title()} service for {v["make"]} {v["model"]}',
            'start_date': (datetime.now() - timedelta(days=random.randint(1, 120))).strftime('%Y-%m-%d'),
            'end_date': None if random.random() < 0.3 else (datetime.now() - timedelta(days=random.randint(0, 30))).strftime('%Y-%m-%d'),
            'cost': rand(5000, 500000),
            'currency': 'KZT',
            'vendor': pick(['ТОО \"Автосервис А\"', 'ИП \"Механик\"', 'Сервисный центр Б', 'Официальный дилер']),
            'odometer_at_service_km': random.randint(v['current_meter_km'] - 30000, v['current_meter_km']),
            'status': pick(['planned', 'in_progress', 'completed', 'cancelled']),
        })

    tracking = []
    for i in range(TRACKING_COUNT):
        v = pick(vehicles)
        tracking.append({
            'id': f'gps{i:03d}',
            'vehicle_id': v['id'],
            'recorded_at': (datetime.now() - timedelta(hours=random.randint(1, 168))).strftime('%Y-%m-%dT%H:%M:%SZ'),
            'latitude': round(random.uniform(43.20, 43.35), 6),
            'longitude': round(random.uniform(76.80, 77.00), 6),
            'speed_kmh': random.randint(0, 80),
            'heading_deg': random.randint(0, 360),
            'altitude_m': random.randint(600, 1200),
            'is_ignition_on': random.random() < 0.7,
        })

    accidents = []
    for i in range(ACCIDENT_COUNT):
        v = pick(vehicles)
        d = pick(drivers) if drivers else None
        accidents.append({
            'id': f'acc{i:03d}',
            'vehicle_id': v['id'],
            'driver_id': d['id'] if d else None,
            'accident_date': (datetime.now() - timedelta(days=random.randint(1, 365))).strftime('%Y-%m-%d'),
            'accident_type': pick(ACCIDENT_TYPES),
            'severity': pick(SEVERITIES),
            'description': f'{pick(["Collision with","Struck by","Hit","Overturned","Damaged by"])} {pick(["another vehicle","fixed object","pedestrian","guardrail","debris"])}',
            'location': pick(REGIONS),
            'property_damage_cost': rand(10000, 1000000),
            'injury_count': random.randint(0, 3) if random.random() < 0.3 else 0,
            'fatality_count': 1 if random.random() < 0.05 else 0,
            'police_report_ref': f'POL-{random.randint(1000,9999)}',
            'is_reportable': True,
            'status': pick(['under_investigation', 'closed', 'pending_insurance']),
        })

    telematics = []
    for i in range(TELEMATICS_COUNT):
        v = pick(vehicles)
        telematics.append({
            'id': f'tlm{i:03d}',
            'vehicle_id': v['id'],
            'recorded_at': (datetime.now() - timedelta(hours=random.randint(1, 168))).strftime('%Y-%m-%dT%H:%M:%SZ'),
            'engine_temp_c': random.randint(70, 115),
            'oil_pressure_bar': round(random.uniform(1.0, 6.0), 1),
            'battery_voltage': round(random.uniform(11.5, 14.5), 1),
            'fuel_level_pct': random.randint(5, 100),
            'coolant_temp_c': random.randint(60, 105),
            'rpm': random.randint(600, 3500),
            'engine_hours': random.randint(100, 15000),
            'diagnostic_code': None if random.random() < 0.8 else f'DTC-{random.randint(100,999)}',
        })

    return {
        'fleet_vehicles': vehicles,
        'fleet_drivers': drivers,
        'fleet_fuel': fuel,
        'fleet_maintenance': maintenance,
        'fleet_tracking': tracking,
        'fleet_accidents': accidents,
        'fleet_telematics': telematics,
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