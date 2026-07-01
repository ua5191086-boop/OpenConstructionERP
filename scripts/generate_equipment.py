#!/usr/bin/env python3
"""
OpenConstructionERP — Equipment Management Data Generator
Generates test data: Equipment, Categories, Maintenance, Telemetry,
Fuel, Operators, Downtime, Spare Parts
"""
import json
import random
from datetime import datetime, date, timedelta
from pathlib import Path

random.seed(42)

OUTPUT = Path(__file__).parent.parent / "apps" / "web" / "equipment_data.json"

PROJECTS = [
    {"id": 1, "name": "Metro Line 3 — Central Corridor"},
    {"id": 2, "name": "Highway Bypass — Northern Ring Road"},
    {"id": 3, "name": "Water Treatment Plant — Phase II"},
    {"id": 4, "name": "Railway Bridge — River Delta Crossing"},
    {"id": 5, "name": "Port Expansion — Container Terminal"},
]

def rand_date(start_year=2024, end_year=2026):
    start = date(start_year, 1, 1)
    end = date(end_year, 6, 30)
    return start + timedelta(days=random.randint(0, (end - start).days))

def generate_categories():
    cats = [
        {"code": "TBM", "name": "Tunnel Boring Machines", "type": "tbm", "icon": "circle"},
        {"code": "CRANE_MOBILE", "name": "Mobile Cranes", "type": "crane", "icon": "arrow-up"},
        {"code": "CRANE_TOWER", "name": "Tower Cranes", "type": "crane", "icon": "arrow-up"},
        {"code": "CRANE_CRAWLER", "name": "Crawler Cranes", "type": "crane", "icon": "arrow-up"},
        {"code": "DUMP_TRUCK", "name": "Dump Trucks", "type": "fleet_vehicle", "icon": "truck"},
        {"code": "CONCRETE", "name": "Concrete Equipment", "type": "heavy_machine", "icon": "package"},
        {"code": "EXCAVATOR", "name": "Excavators", "type": "heavy_machine", "icon": "package"},
        {"code": "LOADER", "name": "Loaders & Dozers", "type": "heavy_machine", "icon": "package"},
        {"code": "PICKUP", "name": "Light Vehicles", "type": "light_equipment", "icon": "truck"},
        {"code": "GENERATOR", "name": "Generators & Compressors", "type": "general", "icon": "zap"},
    ]
    return [{
        "id": f"cat-{i+1:04d}",
        "category_code": c["code"],
        "category_name": c["name"],
        "description": f"Category for {c['name'].lower()}",
        "parent_id": None,
        "equipment_type": c["type"],
        "icon": c["icon"],
        "sort_order": (i + 1) * 10,
    } for i, c in enumerate(cats)]

def generate_equipment(categories):
    items = []
    tbm_configs = [
        {"name": "Herrenknecht S-1100 EPB", "manuf": "Herrenknecht", "model": "S-1100", "cap": "11.0m", "fuel": "electric"},
        {"name": "Robbins Crossover XRE", "manuf": "Robbins", "model": "XRE-12.5", "cap": "12.5m", "fuel": "electric"},
        {"name": "NFM EPB Shield Ø9.5", "manuf": "NFM", "model": "EPB-95", "cap": "9.5m", "fuel": "electric"},
        {"name": "Mitsubishi Slurry Shield", "manuf": "Mitsubishi", "model": "Slurry-8.5", "cap": "8.5m", "fuel": "electric"},
    ]
    crane_configs = [
        {"name": "Liebherr LTM 1100", "manuf": "Liebherr", "model": "LTM 1100", "cap": "100t", "fuel": "diesel"},
        {"name": "Demag AC 350", "manuf": "Demag", "model": "AC 350", "cap": "350t", "fuel": "diesel"},
        {"name": "Terex CC 2800", "manuf": "Terex", "model": "CC 2800", "cap": "600t", "fuel": "diesel"},
        {"name": "Potain MDT 368", "manuf": "Potain", "model": "MDT 368", "cap": "12t", "fuel": "electric"},
    ]
    fleet_configs = [
        {"name": "CAT 740 Dump Truck", "manuf": "Caterpillar", "model": "740", "cap": "40t", "fuel": "diesel"},
        {"name": "Komatsu HD785", "manuf": "Komatsu", "model": "HD785", "cap": "92t", "fuel": "diesel"},
        {"name": "Volvo A40G", "manuf": "Volvo", "model": "A40G", "cap": "39t", "fuel": "diesel"},
        {"name": "Toyota Hilux 4x4", "manuf": "Toyota", "model": "Hilux", "cap": "1t", "fuel": "diesel"},
        {"name": "Isuzu NPR Crew Cab", "manuf": "Isuzu", "model": "NPR", "cap": "2t", "fuel": "diesel"},
    ]
    heavy_configs = [
        {"name": "CAT 336 Excavator", "manuf": "Caterpillar", "model": "336", "cap": "36t", "fuel": "diesel"},
        {"name": "Komatsu PC490", "manuf": "Komatsu", "model": "PC490", "cap": "49t", "fuel": "diesel"},
        {"name": "CAT D6 Dozer", "manuf": "Caterpillar", "model": "D6", "cap": "24t", "fuel": "diesel"},
        {"name": "WA500 Wheel Loader", "manuf": "Komatsu", "model": "WA500", "cap": "24t", "fuel": "diesel"},
        {"name": "CAT 980M Loader", "manuf": "Caterpillar", "model": "980M", "cap": "30t", "fuel": "diesel"},
    ]

    all_configs = tbm_configs + crane_configs + fleet_configs + heavy_configs
    eq_types = ["tbm"] * 4 + ["crane"] * 4 + ["fleet_vehicle"] * 5 + ["heavy_machine"] * 5

    for proj in PROJECTS:
        n = random.randint(6, 12)
        for i in range(1, n + 1):
            cfg = random.choice(all_configs)
            etype = eq_types[all_configs.index(cfg)]
            static = random.choice(["available", "in_use", "in_use", "in_use", "under_maintenance", "out_of_service"])
            # Find matching category
            cat_map = {"tbm": "TBM", "crane": "CRANE_MOBILE", "fleet_vehicle": "DUMP_TRUCK", "heavy_machine": "EXCAVATOR"}
            cat_id = f"cat-{list(cat_map.keys()).index(etype)+1:04d}" if etype in cat_map else "cat-0010"
            items.append({
                "id": f"eq-{proj['id']}-{i:04d}",
                "project_id": proj["id"],
                "project_name": proj["name"],
                "equipment_code": f"{cfg['manuf'][:3].upper()}-{proj['id']}-{i:04d}",
                "equipment_name": cfg["name"],
                "category_id": cat_id,
                "equipment_type": etype,
                "manufacturer": cfg["manuf"],
                "model": cfg["model"],
                "serial_number": f"SN-{cfg['manuf'][:3].upper()}-{random.randint(1000,9999)}",
                "year_manufactured": random.randint(2018, 2024),
                "capacity": cfg["cap"],
                "capacity_unit": "t" if "t" in cfg["cap"] else "m",
                "status": static,
                "location": random.choice(["Site A", "Site B", "Workshop", "Storage Yard", "Tunnel Portal"]),
                "purchase_date": rand_date(2019, 2023).isoformat(),
                "purchase_cost": round(random.uniform(50000, 5000000), 2),
                "current_value": round(random.uniform(10000, 3000000), 2),
                "fuel_type": cfg["fuel"],
                "fuel_capacity": round(random.uniform(100, 800), 2),
                "hourly_rate": round(random.uniform(50, 500), 2),
                "meter_type": random.choice(["hours", "hours", "km"]),
                "meter_reading": round(random.uniform(100, 15000), 2),
                "operator_required": etype != "fleet_vehicle" or random.random() < 0.5,
                "next_service_date": rand_date(2025, 2026).isoformat(),
                "is_active": True,
            })
    return items

def generate_maintenance(equipment):
    items = []
    for eq in equipment:
        n = random.randint(2, 5)
        for i in range(1, n + 1):
            mtype = random.choice(["preventive", "corrective", "predictive", "inspection"])
            status = random.choice(["scheduled", "in_progress", "completed", "completed", "completed"])
            cost_est = round(random.uniform(200, 50000), 2)
            items.append({
                "id": f"maint-{eq['id']}-{i:04d}",
                "equipment_id": eq["id"],
                "project_id": eq["project_id"],
                "equipment_name": eq["equipment_name"],
                "maintenance_code": f"MT-{eq['project_id']}-{eq['equipment_code'][:5]}-{i:04d}",
                "maintenance_type": mtype,
                "description": f"{mtype.title()} maintenance — {random.choice(['Oil change', 'Filter replacement', 'Hydraulic check', 'Brake inspection', 'Engine tune-up', 'Tire rotation'])}",
                "priority": random.choice(["normal", "normal", "high", "critical"]),
                "status": status,
                "meter_at_service": round(random.uniform(500, 10000), 2),
                "cost_estimated": cost_est,
                "cost_actual": round(cost_est * random.uniform(0.8, 1.3), 2) if status == "completed" else None,
                "downtime_hours": round(random.uniform(1, 48), 2) if status == "completed" else None,
                "technician": random.choice(["Mech_A", "Mech_B", "Elec_Team", "Vendor_Service"]),
                "scheduled_date": rand_date(2025, 2026).isoformat(),
                "completed_at": rand_date(2025, 2026).isoformat() if status == "completed" else None,
                "next_service_meter": round(random.uniform(1000, 15000), 2),
                "next_service_date": rand_date(2025, 2026).isoformat(),
            })
    return items

def generate_maintenance_schedules(equipment):
    items = []
    for eq in equipment:
        types = [
            {"name": "Daily Pre-start Check", "days": 1, "meter": 8, "hours": 0.5},
            {"name": "Weekly Inspection", "days": 7, "meter": 40, "hours": 1},
            {"name": "Monthly Service", "days": 30, "meter": 200, "hours": 4},
            {"name": "Quarterly Overhaul", "days": 90, "meter": 500, "hours": 8},
            {"name": "Annual Major Service", "days": 365, "meter": 2000, "hours": 24},
        ]
        for i, t in enumerate(types):
            items.append({
                "id": f"msch-{eq['id']}-{i:04d}",
                "equipment_id": eq["id"],
                "project_id": eq["project_id"],
                "schedule_name": f"{eq['equipment_name']} — {t['name']}",
                "interval_type": "both",
                "interval_days": t["days"],
                "interval_meter": float(t["meter"]),
                "task_list": f"1. Check fluid levels\\n2. Inspect belts and hoses\\n3. Test safety systems",
                "estimated_hours": float(t["hours"]),
                "required_skills": random.choice(["Operator", "Mechanic", "Electrician"]),
                "is_active": random.random() < 0.8,
            })
    return items

def generate_telemetry(equipment):
    items = []
    for eq in equipment:
        n = random.randint(5, 15)
        base = date(2025, 6, 1)
        for i in range(n):
            dt = base + timedelta(hours=i * random.randint(1, 24))
            items.append({
                "id": f"tel-{eq['id']}-{i:04d}",
                "equipment_id": eq["id"],
                "project_id": eq["project_id"],
                "equipment_name": eq["equipment_name"],
                "recorded_at": dt.isoformat(),
                "meter_value": round(random.uniform(0, 15000), 2),
                "fuel_level_pct": round(random.uniform(20, 100), 2),
                "engine_temp_c": round(random.uniform(70, 110), 1),
                "oil_pressure_bar": round(random.uniform(2, 6), 2),
                "rpm": random.randint(800, 2500),
                "speed_kph": round(random.uniform(0, 60), 2) if eq["equipment_type"] == "fleet_vehicle" else 0,
                "battery_voltage": round(random.uniform(11.5, 14.5), 2),
                "is_operating": random.random() < 0.6,
                "data_source": random.choice(["manual", "iot_sensor", "iot_sensor", "gps_tracker"]),
            })
    return items

def generate_fuel(equipment):
    items = []
    for eq in equipment:
        n = random.randint(3, 10)
        for i in range(1, n + 1):
            liters = round(random.uniform(20, 300), 2)
            cost_per = round(random.uniform(1.2, 1.8), 2)
            items.append({
                "id": f"fuel-{eq['id']}-{i:04d}",
                "equipment_id": eq["id"],
                "project_id": eq["project_id"],
                "equipment_name": eq["equipment_name"],
                "refuel_date": rand_date(2025, 2026).isoformat(),
                "fuel_type": eq["fuel_type"],
                "quantity_liters": liters,
                "cost_per_liter": cost_per,
                "total_cost": round(liters * cost_per, 2),
                "meter_reading": round(random.uniform(100, 10000), 2),
                "operator": random.choice(["Opr_A", "Opr_B", "Opr_C"]),
                "vendor": random.choice(["Shell", "BP", "Exxon", "Local_Distributor"]),
            })
    return items

def generate_operators(equipment):
    items = []
    op_names = ["Ahmed Hassan", "John Smith", "Carlos Garcia", "Wei Zhang", "Raj Patel", "Omar Al-Rashid", "Ivan Petrov", "Kim Min-jun"]
    for eq in equipment:
        if not eq["operator_required"]:
            continue
        n = random.randint(1, 2)
        for i in range(n):
            name = random.choice(op_names)
            items.append({
                "id": f"op-{eq['id']}-{i:04d}",
                "equipment_id": eq["id"],
                "project_id": eq["project_id"],
                "equipment_name": eq["equipment_name"],
                "employee_id": f"EMP-{random.randint(100, 999)}",
                "full_name": name,
                "certification": random.choice(["CAT Certified", "Liebherr Certified", "OSHA 30", "Crane License", "CDL A"]),
                "certification_expiry": rand_date(2025, 2027).isoformat(),
                "assigned_date": rand_date(2024, 2025).isoformat(),
                "shift": random.choice(["day", "night", "rotating"]),
                "is_primary": i == 0,
            })
    return items

def generate_downtime(equipment):
    items = []
    for eq in equipment:
        n = random.randint(1, 4)
        for i in range(1, n + 1):
            start = rand_date(2025, 2026)
            dur = round(random.uniform(0.5, 24), 2)
            end = datetime.combine(start, datetime.min.time()) + timedelta(hours=dur)
            items.append({
                "id": f"dt-{eq['id']}-{i:04d}",
                "equipment_id": eq["id"],
                "project_id": eq["project_id"],
                "equipment_name": eq["equipment_name"],
                "downtime_type": random.choice(["breakdown", "maintenance", "fueling", "operator_unavailable", "weather"]),
                "start_time": start.isoformat(),
                "end_time": end.isoformat(),
                "duration_hours": dur,
                "reason": random.choice(["Engine failure", "Hydraulic leak", "Flat tire", "No operator", "Heavy rain"]),
                "cost_impact": round(dur * random.uniform(50, 200), 2),
                "reported_by": random.choice(["Foreman", "Operator", "Site Engineer"]),
                "status": random.choice(["open", "resolved", "closed"]),
            })
    return items

def generate_spare_parts(equipment):
    items = []
    parts_data = [
        {"code": "FIL-OIL-001", "name": "Oil Filter", "cat": "Filters"},
        {"code": "FIL-AIR-002", "name": "Air Filter", "cat": "Filters"},
        {"code": "BELT-001", "name": "Drive Belt", "cat": "Belts"},
        {"code": "HOSE-HYD-001", "name": "Hydraulic Hose 1m", "cat": "Hoses"},
        {"code": "BRK-PAD-001", "name": "Brake Pad Set", "cat": "Brakes"},
        {"code": "BAT-12V-100", "name": "12V Battery 100Ah", "cat": "Electrical"},
        {"code": "FUSE-30A", "name": "30A Fuse", "cat": "Electrical"},
        {"code": "TIR-17.5-25", "name": "Tire 17.5-25 E3", "cat": "Tires"},
        {"code": "SEAL-HYD-010", "name": "Hydraulic Seal Kit", "cat": "Seals"},
        {"code": "BEARING-6205", "name": "Ball Bearing 6205", "cat": "Bearings"},
    ]
    for eq in equipment[:3]:
        for part in parts_data[:random.randint(3, 7)]:
            items.append({
                "id": f"sp-{eq['id']}-{part['code'].lower()}",
                "equipment_id": eq["id"],
                "project_id": eq["project_id"],
                "part_code": part["code"],
                "part_name": part["name"],
                "part_number": f"OEM-{random.randint(10000,99999)}",
                "category": part["cat"],
                "unit": "pcs",
                "quantity_on_hand": round(random.uniform(0, 20), 2),
                "min_stock_level": 2,
                "unit_cost": round(random.uniform(5, 500), 2),
                "supplier": random.choice(["CAT Parts", "Komatsu Parts", "Volvo Parts", "Local Supplier"]),
                "lead_time_days": random.randint(1, 30),
                "storage_location": f"Aisle-{random.randint(1,5)} Shelf-{random.choice(['A','B','C'])}{random.randint(1,5)}",
            })
    return items

def main():
    categories = generate_categories()
    equipment = generate_equipment(categories)
    maintenance = generate_maintenance(equipment)
    maintenance_schedules = generate_maintenance_schedules(equipment)
    telemetry = generate_telemetry(equipment)
    fuel = generate_fuel(equipment)
    operators = generate_operators(equipment)
    downtime = generate_downtime(equipment)
    spare_parts = generate_spare_parts(equipment)

    data = {
        "generated_at": datetime.utcnow().isoformat(),
        "summary": {
            "categories": len(categories),
            "equipment": len(equipment),
            "maintenance": len(maintenance),
            "maintenance_schedules": len(maintenance_schedules),
            "telemetry": len(telemetry),
            "fuel": len(fuel),
            "operators": len(operators),
            "downtime": len(downtime),
            "spare_parts": len(spare_parts),
        },
        "categories": categories,
        "equipment": equipment,
        "maintenance": maintenance,
        "maintenance_schedules": maintenance_schedules,
        "telemetry": telemetry,
        "fuel": fuel,
        "operators": operators,
        "downtime": downtime,
        "spare_parts": spare_parts,
    }

    OUTPUT.parent.mkdir(parents=True, exist_ok=True)
    with open(OUTPUT, "w", encoding="utf-8") as f:
        json.dump(data, f, indent=2, ensure_ascii=False)

    print(f"[Equipment] Generated {data['summary']['categories']} categories")
    print(f"[Equipment] Generated {data['summary']['equipment']} equipment")
    print(f"[Equipment] Generated {data['summary']['maintenance']} maintenance records")
    print(f"[Equipment] Generated {data['summary']['maintenance_schedules']} maintenance schedules")
    print(f"[Equipment] Generated {data['summary']['telemetry']} telemetry records")
    print(f"[Equipment] Generated {data['summary']['fuel']} fuel records")
    print(f"[Equipment] Generated {data['summary']['operators']} operator assignments")
    print(f"[Equipment] Generated {data['summary']['downtime']} downtime records")
    print(f"[Equipment] Generated {data['summary']['spare_parts']} spare parts")
    print(f"[Equipment] Output: {OUTPUT}")

if __name__ == "__main__":
    main()