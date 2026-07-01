#!/usr/bin/env python3
"""
OpenConstructionERP — Ring Builder & Segment Tracking Data Generator (V022)
Generates test data: ring designs, segment production, curing, transport, installation, QC, inventory, measurements
"""
import json, random
from datetime import datetime, timedelta, date
from pathlib import Path

random.seed(42)
OUTPUT = Path(__file__).parent.parent / "apps" / "web" / "ringbuilder_data.json"

PROJECTS = [{"id": "p001", "name": "Metro Line 3"}]
DESIGN_TYPES = ["universal", "tapered", "straight", "bolted", "key", "special"]
SEGMENT_TYPES = ["A1", "A2", "A3", "A4", "A5", "A6", "B", "K"]
QC_RESULTS = ["pending", "pass", "conditional_pass", "fail", "rework"]
TRANSPORT_MODES = ["truck", "train", "flatbed", "crane", "gantry", "other"]
INSTRUMENTS = ["total_station", "laser_scanner", "photogrammetry", "convergence_tape", "distometer"]

def rand_float(lo, hi, dec=2):
    return round(random.uniform(lo, hi), dec)

def rand_date(start=date(2025, 1, 1), end=date(2025, 6, 30)):
    s = datetime(start.year, start.month, start.day)
    e = datetime(end.year, end.month, end.day)
    return s + timedelta(days=random.randint(0, (e-s).days))

def generate_designs():
    items = []
    for i, dtype in enumerate(DESIGN_TYPES):
        idx = i + 1
        seg_count = random.choice([6, 7, 8, 9])
        items.append({
            "id": f"dsg-{idx:04d}",
            "project_id": "p001",
            "project_name": "Metro Line 3",
            "design_code": f"RD-{idx:03d}",
            "design_name": f"Ring Design {dtype.title()} v{idx}",
            "ring_type": dtype,
            "inner_diameter_mm": 5600,
            "outer_diameter_mm": 6200,
            "ring_width_mm": 1500,
            "taper_mm": random.choice([0, 12, 18, 24, 36]),
            "segment_count": seg_count,
            "key_position": "K",
            "concrete_grade": "C45/55",
            "reinforcement_type": "steel",
            "weight_kg": rand_float(15000, 25000, 0),
            "status": "active",
        })
    return items

def generate_production():
    items = []
    for i in range(1, 301):
            seg_type = random.choice(SEGMENT_TYPES)
            cast = rand_date(date(2025, 1, 15), date(2025, 6, 15))
            curing_start = cast + timedelta(hours=random.randint(1, 3))
            curing_end = curing_start + timedelta(hours=random.randint(8, 16))
            demold = curing_end + timedelta(hours=random.randint(1, 4))
            transport = demold + timedelta(days=random.randint(1, 7))
            install = transport + timedelta(days=random.randint(1, 14))
            status_roll = random.random()
            if status_roll < 0.2:
                st = "cast"
                install = None
            elif status_roll < 0.4:
                st = "curing"
                demold = None; transport = None; install = None
            elif status_roll < 0.55:
                st = "demolded"
                transport = None; install = None
            elif status_roll < 0.7:
                st = "transport"
                install = None
            elif status_roll < 0.95:
                st = "in_stock"
                install = None
            else:
                st = "installed"
            items.append({
                "id": f"seg-{i:06d}",
                "project_id": "p001",
                "project_name": "Metro Line 3",
                "design_id": f"dsg-{random.randint(1, len(DESIGN_TYPES)):04d}",
                "segment_code": f"SEG-{i:04d}-{seg_type}",
                "segment_type": seg_type,
                "ring_designation": f"RNG-{i // len(SEGMENT_TYPES) + 1:04d}",
                "mold_id": f"MOLD-{random.randint(1,8):02d}",
                "cast_batch": f"BATCH-{random.randint(1,30):03d}",
                "concrete_grade": "C45/55",
                "concrete_volume_m3": rand_float(2.5, 4.5),
                "steel_weight_kg": rand_float(180, 350),
                "cast_at": cast.isoformat(),
                "cast_by": random.choice(["Ivanov", "Petrov", "Sidorov", "Kuznetsov"]),
                "curing_start_at": curing_start.isoformat(),
                "curing_end_at": curing_end.isoformat() if demold else None,
                "curing_method": random.choice(["steam", "water", "air", "accelerated"]),
                "demold_at": demold.isoformat() if demold else None,
                "transport_at": transport.isoformat() if transport else None,
                "install_at": install.isoformat() if install else None,
                "status": st,
                "qc_status": random.choice(["pending", "passed", "passed", "passed", "conditional", "failed"]),
                "qr_code": f"QR-SEG-{i:06d}",
                "location": random.choice(["Yard A", "Yard B", "Stockpile 1", "Tunnel face"]),
            })
    return items

def generate_curing():
    items = []
    for seg_id in [f"seg-{i:06d}" for i in range(1, 301) if random.random() > 0.3]:
        for stage in ["initial_set", "steam", "cooling", "final_set", "demold"]:
            if random.random() < 0.2:
                continue
            start = rand_date(date(2025, 1, 15), date(2025, 6, 15))
            items.append({
                "id": f"cure-{seg_id}-{stage}",
                "segment_id": seg_id,
                "curing_stage": stage,
                "start_time": start.isoformat(),
                "end_time": (start + timedelta(hours=random.randint(1, 8))).isoformat(),
                "temp_target_c": rand_float(40, 75),
                "temp_actual_c": rand_float(38, 78),
                "humidity_target_pct": rand_float(80, 100),
                "humidity_actual_pct": rand_float(75, 100),
                "gradient_rate_cph": rand_float(5, 20),
            })
    return items

def generate_transport():
    items = []
    for seg_id in [f"seg-{i:06d}" for i in range(1, 301) if random.random() > 0.4]:
        tdate = rand_date(date(2025, 2, 1), date(2025, 6, 20))
        items.append({
            "id": f"trp-{seg_id}",
            "segment_id": seg_id,
            "transport_date": tdate.isoformat(),
            "transport_mode": random.choice(TRANSPORT_MODES),
            "vehicle_number": f"TRK-{random.randint(100,999)}",
            "driver_name": random.choice(["Nikolaev", "Fedorov", "Mikhailov"]),
            "from_location": random.choice(["Casting Yard A", "Casting Yard B"]),
            "to_location": random.choice(["Tunnel Portal 1", "Shaft 3", "Stockpile 2"]),
            "departure_time": tdate.isoformat(),
            "arrival_time": (tdate + timedelta(hours=random.randint(2, 8))).isoformat(),
            "distance_km": rand_float(5, 80),
            "damage_reported": random.random() < 0.05,
            "temperature_c": rand_float(-5, 40),
            "transport_cost": rand_float(200, 2000),
            "created_by": "System",
        })
    return items

def generate_installation():
    items = []
    for i in range(1, 201):
        items.append({
            "id": f"inst-{i:06d}",
            "segment_id": f"seg-{i:06d}",
            "ring_id": f"ring-{i // len(SEGMENT_TYPES) + 1:04d}",
            "erector_cycle_time_sec": random.randint(120, 600),
            "bolt_count": random.randint(8, 16),
            "bolt_torque_nm": rand_float(100, 500),
            "packer_type": random.choice(["EPDM", "Neoprene", "Hydrophilic"]),
            "gap_mm": rand_float(0, 5),
            "offset_radial_mm": rand_float(0, 10),
            "offset_longitudinal_mm": rand_float(0, 15),
            "installed_by": random.choice(["Erector Team A", "Erector Team B"]),
            "installed_at": rand_date(date(2025, 2, 1), date(2025, 6, 25)).isoformat(),
        })
    return items

def generate_qc():
    items = []
    for seg_id in [f"seg-{i:06d}" for i in range(1, 301) if random.random() > 0.1]:
        result = random.choice(QC_RESULTS)
        items.append({
            "id": f"qc-{seg_id}",
            "segment_id": seg_id,
            "length_mm": rand_float(2980, 3020),
            "width_mm": rand_float(1180, 1220),
            "thickness_mm": rand_float(295, 305),
            "diagonal_diff_mm": rand_float(0, 3),
            "surface_defects": random.choice(["None", "Minor pitting", "Edge spalling < 5mm"]),
            "honeycomb_pct": rand_float(0, 2),
            "spalling_pct": rand_float(0, 1.5),
            "cracking": random.choice(["None", "Hairline cracks < 0.2mm"]),
            "cover_min_mm": rand_float(30, 45),
            "cover_max_mm": rand_float(45, 60),
            "compressive_strength_mpa": rand_float(50, 75),
            "water_absorption_pct": rand_float(1.5, 4.0),
            "bolt_socket_present": True,
            "lifting_anchor_present": True,
            "qc_result": result,
            "qc_inspector": random.choice(["QC Ivanov", "QC Petrova", "QC Smirnov"]),
            "qc_date": rand_date(date(2025, 1, 15), date(2025, 6, 15)).isoformat(),
            "corrective_action": "None required" if result in ("pass", "conditional_pass") else "Grind spalling",
        })
    return items

def generate_inventory():
    items = []
    for dsg_id in [f"dsg-{i:04d}" for i in range(1, 7)]:
        for stype in SEGMENT_TYPES:
            produced = random.randint(20, 80)
            passed = int(produced * random.uniform(0.85, 0.98))
            installed = int(passed * random.uniform(0.3, 0.7))
            defective = produced - passed
            in_transit = int(passed * random.uniform(0.05, 0.15))
            items.append({
                "id": f"inv-{dsg_id}-{stype}",
                "project_id": "p001",
                "design_id": dsg_id,
                "segment_type": stype,
                "quantity_planned": produced + random.randint(10, 30),
                "quantity_produced": produced,
                "quantity_passed_qc": passed,
                "quantity_installed": installed,
                "quantity_defective": defective,
                "quantity_in_transit": in_transit,
                "stock_location": random.choice(["Yard A", "Yard B"]),
            })
    return items

def generate_measurements():
    items = []
    for ring_no in range(1, 101):
        n_meas = random.randint(1, 5)
        for m in range(n_meas):
            ov_pct = rand_float(0.05, 1.5, 3)
            items.append({
                "id": f"meas-ring-{ring_no:04d}-{m:02d}",
                "ring_id": f"ring-{ring_no:04d}",
                "measured_at": rand_date(date(2025, 3, 1), date(2025, 6, 30)).isoformat(),
                "horizontal_convergence_mm": rand_float(-8, 8),
                "vertical_convergence_mm": rand_float(-10, 10),
                "diagonal_1_mm": rand_float(-6, 6),
                "diagonal_2_mm": rand_float(-6, 6),
                "ovality_pct": ov_pct,
                "ovality_mm": round(ov_pct * 6200 / 100, 2),
                "deformation_vertical_mm": rand_float(-5, 5),
                "deformation_horizontal_mm": rand_float(-4, 4),
                "settlement_mm": rand_float(-3, 3),
                "profile_chainage": rand_float(100, 2500),
                "section_area_loss_pct": rand_float(0, 0.5, 3),
                "instrument_type": random.choice(INSTRUMENTS),
                "measured_by": random.choice(["Survey Team A", "Survey Team B", "Geotech Lab"]),
                "weather_conditions": random.choice(["Dry", "Humid", "Wet", "Normal"]),
            })
    return items

def main():
    data = {
        "generated_at": datetime.utcnow().isoformat(),
        "ring_designs": generate_designs(),
        "segment_production": generate_production(),
        "segment_curing": generate_curing(),
        "segment_transport": generate_transport(),
        "segment_installation": generate_installation(),
        "segment_qc": generate_qc(),
        "segment_inventory": generate_inventory(),
        "ring_measurements": generate_measurements(),
    }
    OUTPUT.write_text(json.dumps(data, indent=2, ensure_ascii=False, default=str))
    print(f"[Ring Builder Generator] Written {OUTPUT}")
    for k, v in data.items():
        print(f"  {k}: {len(v)}")

if __name__ == "__main__":
    main()