#!/usr/bin/env python3
"""
OpenConstructionERP — TBM Management Data Generator (V021)
Generates test data: telemetry, alarms, operators, shifts, consumables, performance metrics
"""
import json, random
from datetime import datetime, timedelta, date
from pathlib import Path

random.seed(42)
OUTPUT = Path(__file__).parent.parent / "apps" / "web" / "tbm_data.json"

TBMS = [
    {"id": "tbm-001", "code": "TBM-01", "name": "Herrenknecht S-1234 EPB", "type": "EPB"},
    {"id": "tbm-002", "code": "TBM-02", "name": "Robbins C-567 Slurry", "type": "SLURRY"},
]

OPERATORS = [
    {"id": "op-001", "name": "Ivan Petrov", "qual": "Senior TBM Operator"},
    {"id": "op-002", "name": "Alexei Smirnov", "qual": "TBM Operator"},
    {"id": "op-003", "name": "Dmitri Volkov", "qual": "Assistant Operator"},
    {"id": "op-004", "name": "Sergei Kuznetsov", "qual": "Shift Engineer"},
]

SHIFT_LABELS = ["A", "B", "C"]
CONSUMABLE_TYPES = ["cutterhead", "seals", "foam", "bentonite", "grease", "grout", "wear_parts", "hydraulic_oil"]
ALARM_CODES = [
    ("FACE_PRESS_HIGH", "Face Pressure High", "critical"),
    ("FACE_PRESS_LOW", "Face Pressure Low", "warning"),
    ("THRUST_OVERLOAD", "Thrust Overload", "critical"),
    ("TORQUE_SPIKE", "Torque Spike", "warning"),
    ("CUTTER_STALL", "Cutterhead Stall", "emergency"),
    ("TEMP_HIGH", "Hydraulic Temp High", "warning"),
    ("EPB_CHAMBER", "EPB Chamber Blockage", "critical"),
    ("SLURRY_DENSITY", "Slurry Density Abnormal", "warning"),
    ("BELT_STALL", "Belt Conveyor Stall", "warning"),
    ("GREASE_LOW", "Tail Skin Grease Low", "info"),
]

def rand_float(lo, hi, dec=2):
    return round(random.uniform(lo, hi), dec)

def rand_date(start=date(2025, 1, 1), end=date(2025, 6, 30)):
    s = datetime(start.year, start.month, start.day)
    e = datetime(end.year, end.month, end.day)
    return s + timedelta(days=random.randint(0, (e-s).days))

def generate_telemetry():
    items = []
    for tbm in TBMS:
        base = rand_date(date(2025, 1, 1), date(2025, 6, 1))
        for i in range(200):
            ts = base + timedelta(minutes=random.randint(0, 43200))
            epb_params = {
                "epb_face_pressure_bar": rand_float(0.5, 3.5),
                "epb_screw_speed_rpm": rand_float(0, 12),
                "epb_screw_torque_kNm": rand_float(10, 300),
                "epb_chamber_pressure_bar": rand_float(0.3, 3.0),
            } if tbm["type"] == "EPB" else {}
            slurry_params = {
                "slurry_density_kgm3": rand_float(1050, 1400),
                "slurry_flow_in_m3h": rand_float(200, 600),
                "slurry_flow_out_m3h": rand_float(180, 580),
                "slurry_pressure_bar": rand_float(0.5, 4.0),
            } if tbm["type"] == "SLURRY" else {}
            items.append({
                "id": f"tel-{tbm['id']}-{i:04d}",
                "tbm_id": tbm["id"],
                "tbm_code": tbm["code"],
                "recorded_at": ts.isoformat(),
                **epb_params, **slurry_params,
                "thrust_force_kN": rand_float(1000, 35000),
                "thrust_speed_mmmin": rand_float(0, 80),
                "torque_kNm": rand_float(100, 4500),
                "torque_pct": rand_float(10, 95),
                "advance_rate_mmmin": rand_float(0, 60),
                "advance_mm": rand_float(0, 12000),
                "face_pressure_bar": rand_float(0.5, 4.5),
                "cutterhead_rpm": rand_float(0, 3.5),
                "cutterhead_torque_kNm": rand_float(50, 3000),
                "cutterhead_wear_mm": rand_float(0, 15),
                "tail_skin_grease_bar": rand_float(0, 5),
                "articulation_angle_deg": rand_float(-2, 2),
                "belt_weight_kg": rand_float(0, 5000),
                "total_power_kw": rand_float(100, 2000),
            })
    return items

def generate_alarms():
    items = []
    for tbm in TBMS:
        for i in range(random.randint(8, 20)):
            code, name, sev = random.choice(ALARM_CODES)
            trig = rand_date()
            cleared = trig + timedelta(minutes=random.randint(5, 240)) if random.random() > 0.3 else None
            items.append({
                "id": f"alm-{tbm['id']}-{i:04d}",
                "tbm_id": tbm["id"],
                "tbm_code": tbm["code"],
                "alarm_code": code,
                "alarm_name": name,
                "alarm_severity": sev,
                "triggered_at": trig.isoformat(),
                "acknowledged_at": (trig + timedelta(minutes=random.randint(1, 30))).isoformat(),
                "acknowledged_by": random.choice([o["name"] for o in OPERATORS]),
                "cleared_at": cleared.isoformat() if cleared else None,
                "param_value": rand_float(0, 100),
                "threshold_value": rand_float(10, 80),
                "is_active": cleared is None,
            })
    return items

def generate_operators():
    items = []
    for op in OPERATORS:
        items.append({
            "id": op["id"],
            "employee_id": f"EMP-{random.randint(100,999)}",
            "full_name": op["name"],
            "qualification": op["qual"],
            "certification_number": f"CERT-{random.randint(1000,9999)}",
            "certification_expiry": date(2026, random.randint(1,12), random.randint(1,28)).isoformat(),
            "tbm_types": "EPB,SLURRY",
            "phone": f"+7-{random.randint(900,999)}-{random.randint(100,999)}-{random.randint(10,99)}-{random.randint(10,99)}",
            "email": f"{op['name'].lower().replace(' ','.')}@contractor.com",
            "is_active": True,
        })
    return items

def generate_shifts():
    items = []
    for tbm in TBMS:
        for d in range(90):
            shift_date = date(2025, 3, 1) + timedelta(days=d)
            for lbl in random.sample(SHIFT_LABELS, k=random.randint(1, 3)):
                op = random.choice(OPERATORS)
                asst = random.choice(OPERATORS)
                rings = random.randint(0, 6)
                items.append({
                    "id": f"shf-{tbm['id']}-{d:03d}-{lbl}",
                    "tbm_id": tbm["id"],
                    "tbm_code": tbm["code"],
                    "shift_date": shift_date.isoformat(),
                    "shift_label": lbl,
                    "operator_id": op["id"],
                    "operator_name": op["name"],
                    "assistant_id": asst["id"],
                    "assistant_name": asst["name"],
                    "start_time": datetime(shift_date.year, shift_date.month, shift_date.day, random.randint(6,23), 0).isoformat(),
                    "end_time": datetime(shift_date.year, shift_date.month, shift_date.day, min(random.randint(6,23)+8, 23), 0).isoformat() if random.random() > 0.1 else None,
                    "rings_built": rings,
                    "advance_mm": rings * random.randint(1200, 1800),
                    "downtime_minutes": random.randint(0, 120),
                    "downtime_reason": random.choice(["", "Maintenance", "Survey", "Segment delivery delay", "Mechanical issue"]),
                    "notes": random.choice(["", "Standard shift", "Good progress"]),
                })
    return items

def generate_consumables():
    items = []
    for tbm in TBMS:
        for ctype in CONSUMABLE_TYPES:
            for _ in range(random.randint(5, 15)):
                shift = random.choice(items) if items else None
                items.append({
                    "id": f"con-{tbm['id']}-{ctype}-{random.randint(1,999):04d}",
                    "tbm_id": tbm["id"],
                    "tbm_code": tbm["code"],
                    "consumable_type": ctype,
                    "item_name": f"{ctype.replace('_',' ').title()} Item",
                    "item_code": f"{ctype[:3].upper()}-{random.randint(100,999)}",
                    "unit": random.choice(["kg", "L", "pcs", "m"]),
                    "quantity_used": rand_float(0.5, 500),
                    "quantity_remaining": rand_float(50, 5000),
                    "unit_price": rand_float(0.5, 50),
                    "used_at": rand_date().isoformat(),
                    "shift_id": shift["id"] if shift else None,
                    "recorded_by": random.choice([o["name"] for o in OPERATORS]),
                })
    return items

def generate_performance():
    items = []
    for tbm in TBMS:
        for d in range(90):
            sd = date(2025, 3, 1) + timedelta(days=d)
            for lbl in SHIFT_LABELS:
                if random.random() < 0.3:
                    continue
                rings = random.randint(0, 5)
                advance = rings * random.randint(1400, 1700)
                downtime = random.randint(0, 180)
                total_min = 480
                items.append({
                    "id": f"perf-{tbm['id']}-{sd.isoformat()}-{lbl}",
                    "tbm_id": tbm["id"],
                    "tbm_code": tbm["code"],
                    "metric_date": sd.isoformat(),
                    "shift_label": lbl,
                    "rings_built": rings,
                    "advance_mm": advance,
                    "avg_advance_rate_mmmin": rand_float(10, 50),
                    "max_advance_rate_mmmin": rand_float(30, 65),
                    "avg_thrust_force_kN": rand_float(5000, 25000),
                    "max_thrust_force_kN": rand_float(10000, 35000),
                    "avg_torque_kNm": rand_float(500, 3500),
                    "avg_face_pressure_bar": rand_float(0.8, 3.5),
                    "total_downtime_minutes": downtime,
                    "utilisation_pct": round((total_min - downtime) / total_min * 100, 1),
                    "tbm_availability_pct": round(random.uniform(85, 99), 1),
                    "performance_factor": round(random.uniform(0.6, 1.3), 2),
                    "grout_volume_m3": rand_float(0, 8),
                    "grout_pressure_avg_bar": rand_float(0.5, 3.0),
                    "cutterhead_wear_avg_mm": rand_float(0, 10),
                    "foam_consumption_kg": rand_float(0, 200),
                    "bentonite_consumption_kg": rand_float(0, 500),
                    "data_points": random.randint(50, 960),
                })
    return items

def main():
    data = {
        "generated_at": datetime.utcnow().isoformat(),
        "tbms": TBMS,
        "telemetry": generate_telemetry(),
        "alarms": generate_alarms(),
        "operators": generate_operators(),
        "shifts": generate_shifts(),
        "consumables": generate_consumables(),
        "performance_metrics": generate_performance(),
    }
    OUTPUT.write_text(json.dumps(data, indent=2, ensure_ascii=False, default=str))
    print(f"[TBM Generator] Written {OUTPUT}")
    print(f"  Telemetry: {len(data['telemetry'])}")
    print(f"  Alarms: {len(data['alarms'])}")
    print(f"  Operators: {len(data['operators'])}")
    print(f"  Shifts: {len(data['shifts'])}")
    print(f"  Consumables: {len(data['consumables'])}")
    print(f"  Performance: {len(data['performance_metrics'])}")

if __name__ == "__main__":
    main()