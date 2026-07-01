#!/usr/bin/env python3
"""
OpenConstructionERP — NATM & Microtunnelling Data Generator (V023)
Generates test data: excavation logs, shotcrete, rock bolts, steel sets,
convergence, face mapping, MTBM drives, thrust, lubrication, survey,
shafts, equipment, cross passages, grouting, settlement
"""
import json, random
from datetime import datetime, timedelta, date
from pathlib import Path

random.seed(42)
OUTPUT = Path(__file__).parent.parent / "apps" / "web" / "natm_data.json"

PROJECTS = [{"id": "p001", "name": "Metro Line 3 - NATM Section"}]
DRIVES = [
    {"id": "d-001", "code": "DRV-NATM-01", "name": "Tunnel Section A (NATM)"},
    {"id": "d-002", "code": "DRV-MTBM-01", "name": "Pipe Jacking Section B"},
]
MTBM_DRIVES = [
    {"id": "mtb-001", "code": "MTBM-DRV-01", "name": "MTBM Drive 1 - Sewer Line"},
    {"id": "mtb-002", "code": "MTBM-DRV-02", "name": "MTBM Drive 2 - Crossing"},
]

def rand_float(lo, hi, dec=2):
    return round(random.uniform(lo, hi), dec)

def rand_date(start=date(2025, 1, 1), end=date(2025, 6, 30)):
    s = datetime(start.year, start.month, start.day)
    e = datetime(end.year, end.month, end.day)
    return s + timedelta(days=random.randint(0, (e-s).days))

def generate_excavation():
    items = []
    for drive in DRIVES:
        if "NATM" not in drive["code"]:
            continue
        for rnd in range(1, 81):
            ch_from = rand_float(0 + (rnd-1)*1.5, 0 + (rnd-1)*1.5, 2)
            ch_to = ch_from + rand_float(1.2, 2.0)
            items.append({
                "id": f"exc-{drive['id']}-{rnd:03d}",
                "drive_id": drive["id"],
                "drive_code": drive["code"],
                "round_no": rnd,
                "chainage_from": ch_from,
                "chainage_to": ch_to,
                "excavation_date": rand_date(date(2025, 2, 1), date(2025, 6, 30)).isoformat(),
                "shift": random.choice(["A", "B", "C"]),
                "method": random.choice(["drill_blast", "mechanical", "top_heading", "bench"]),
                "round_length_m": round(ch_to - ch_from, 2),
                "excavated_volume_m3": rand_float(20, 80),
                "geotech_class": random.choice(["I", "II", "III", "IV", "V"]),
                "water_inflow_lmin": rand_float(0, 50),
                "support_class": random.choice(["S1", "S2", "S3", "S4"]),
                "standup_time_hours": rand_float(2, 48),
                "delay_minutes": random.randint(0, 90),
                "delay_reason": random.choice(["", "Survey", "Ventilation", "Water inflow", "Rock condition"]),
                "notes": random.choice(["", "Good progress", "Moderate overbreak"]),
            })
    return items

def generate_shotcrete():
    items = []
    for drive in DRIVES:
        if "NATM" not in drive["code"]:
            continue
        for _ in range(120):
            items.append({
                "id": f"shcr-{random.randint(1,9999):04d}",
                "drive_id": drive["id"],
                "drive_code": drive["code"],
                "round_id": f"exc-{drive['id']}-{random.randint(1,80):03d}",
                "application_date": rand_date(date(2025, 2, 1), date(2025, 6, 30)).isoformat(),
                "location_type": random.choice(["invert", "arch", "wall", "bench", "face", "complete"]),
                "shotcrete_type": random.choice(["dry_mix", "wet_mix", "fiber_reinforced", "steel_fiber"]),
                "design_class": "C25/30",
                "thickness_mm": rand_float(50, 350),
                "area_m2": rand_float(10, 80),
                "volume_m3": rand_float(1, 20),
                "compressive_strength_mpa": rand_float(20, 40),
                "fiber_content_kgm3": rand_float(0, 40),
                "accelerator_type": random.choice(["Sika", "BASF", "MAPEI"]),
                "accelerator_dosage_pct": rand_float(2, 8),
                "rebound_pct": rand_float(3, 20),
                "application_temp_c": rand_float(5, 35),
                "nozzleman": random.choice(["Nozzleman A", "Nozzleman B"]),
                "qc_status": random.choice(["pending", "passed", "passed", "passed", "failed"]),
                "test_core_28d_mpa": rand_float(25, 50),
            })
    return items

def generate_rock_bolts():
    items = []
    for drive in DRIVES:
        if "NATM" not in drive["code"]:
            continue
        for _ in range(80):
            items.append({
                "id": f"bolt-{random.randint(1,9999):04d}",
                "drive_id": drive["id"],
                "drive_code": drive["code"],
                "round_id": f"exc-{drive['id']}-{random.randint(1,80):03d}",
                "bolt_type": random.choice(["expansion", "resin", "mechanical", "swellex", "self_drilling", "friction", "tensioned", "grouted"]),
                "bolt_diameter_mm": random.choice([20, 22, 25]),
                "bolt_length_mm": rand_float(2000, 6000, 0),
                "bolt_grade": "500/550",
                "spacing_longitudinal_m": rand_float(1.0, 2.5),
                "spacing_transverse_m": rand_float(1.0, 2.0),
                "quantity_installed": random.randint(4, 20),
                "installed_at": rand_date().isoformat() + "Z",
                "pretension_kN": rand_float(50, 250),
                "pullout_test_kN": rand_float(80, 300),
                "grout_volume_l": rand_float(0.5, 10),
                "pattern_type": random.choice(["systematic", "spot", "pattern", "random"]),
                "installed_by": random.choice(["Bolting Team A", "Bolting Team B"]),
                "qc_status": random.choice(["pending", "passed", "passed", "failed"]),
            })
    return items

def generate_steel_sets():
    items = []
    for drive in DRIVES:
        if "NATM" not in drive["code"]:
            continue
        for sn in range(1, 41):
            items.append({
                "id": f"steel-{drive['id']}-{sn:03d}",
                "drive_id": drive["id"],
                "drive_code": drive["code"],
                "round_id": f"exc-{drive['id']}-{random.randint(1,80):03d}",
                "set_number": sn,
                "chainage": rand_float(0, 120),
                "set_type": random.choice(["TH", "TH-44", "TH-58", "GRI", "HEB", "IPE", "Lattice_girder", "UMC"]),
                "section_name": f"Profile {sn}",
                "spacing_m": rand_float(0.8, 2.0),
                "steel_grade": "S355",
                "weight_kg_m": rand_float(15, 60),
                "quantity_arches": 1,
                "installed_at": rand_date().isoformat() + "Z",
                "connection_type": "bolted",
                "lagging_type": random.choice(["Steel plate", "Mesh", "Timber", "None"]),
                "qc_status": random.choice(["pending", "passed", "passed", "passed"]),
            })
    return items

def generate_convergence():
    items = []
    for drive in DRIVES:
        if "NATM" not in drive["code"]:
            continue
        for pt in range(1, 31):
            for _ in range(random.randint(3, 12)):
                cumul = rand_float(-25, 25)
                items.append({
                    "id": f"conv-{drive['id']}-pt{pt:03d}-{random.randint(1,99):02d}",
                    "drive_id": drive["id"],
                    "drive_code": drive["code"],
                    "round_id": f"exc-{drive['id']}-{random.randint(1,80):03d}",
                    "measurement_point": f"CP-{pt:03d}",
                    "chainage": rand_float(0, 120),
                    "measured_at": rand_date().isoformat() + "Z",
                    "displacement_vertical_mm": rand_float(-15, 15),
                    "displacement_horizontal_mm": rand_float(-12, 12),
                    "displacement_longitudinal_mm": rand_float(-5, 5),
                    "convergence_rate_mmday": rand_float(-2, 2, 4),
                    "cumulative_displacement_mm": cumul,
                    "instrument_type": random.choice(["total_station", "extensometer", "convergence_tape", "inclinometer", "laser"]),
                    "distance_from_face_m": rand_float(0, 50),
                    "temperature_c": rand_float(10, 30),
                    "alarm_triggered": abs(cumul) > 20 and random.random() > 0.5,
                    "reading_by": random.choice(["Survey Team", "Geotech"]),
                })
    return items

def generate_face_mapping():
    items = []
    for drive in DRIVES:
        if "NATM" not in drive["code"]:
            continue
        for rnd in range(1, 81):
            rmr = rand_float(20, 85)
            items.append({
                "id": f"face-{drive['id']}-{rnd:03d}",
                "drive_id": drive["id"],
                "drive_code": drive["code"],
                "round_id": f"exc-{drive['id']}-{rnd:03d}",
                "chainage": rand_float(0 + (rnd-1)*1.5, 0 + rnd*1.5),
                "mapped_at": rand_date().isoformat() + "Z",
                "rock_type": random.choice(["Limestone", "Sandstone", "Shale", "Granite", "Marl", "Claystone"]),
                "weathering_grade": random.choice(["I", "II", "III", "IV", "V", "VI"]),
                "rmr_score": rmr,
                "q_score": rand_float(0.1, 40, 2),
                "gsi_value": rand_float(20, 85),
                "joint_count": random.randint(0, 12),
                "joint_spacing_m": rand_float(0.1, 2.0),
                "joint_condition": random.choice(["Very rough", "Rough", "Smooth", "Polished"]),
                "groundwater_condition": random.choice(["Dry", "Damp", "Wet", "Dripping", "Flowing"]),
                "fault_zone": random.random() < 0.1,
                "water_inflow_estimated_lmin": rand_float(0, 60),
                "standup_time_est_hours": rand_float(1, 72),
                "support_recommendation": random.choice(["S2 shotcrete + bolts", "S3 shotcrete + bolts + lattice", "S4 heavy support"]),
                "mapped_by": random.choice(["Geologist A", "Geologist B", "Chief Geotech"]),
            })
    return items

def generate_mtbm_drives():
    items = []
    for mtbm in MTBM_DRIVES:
        items.append({
            "id": mtbm["id"],
            "project_id": "p001",
            "drive_code": mtbm["code"],
            "drive_name": mtbm["name"],
            "mtbm_id": f"MTBM-{random.randint(100,999)}",
            "pipe_type": random.choice(["concrete", "steel", "ductile_iron", "HDPE"]),
            "pipe_diameter_mm": random.choice([800, 1000, 1200, 1500, 1800]),
            "pipe_length_mm": 3000,
            "wall_thickness_mm": random.choice([80, 100, 120, 150]),
            "design_length_m": rand_float(50, 500),
            "max_jacking_force_kN": rand_float(2000, 12000),
            "intermediate_jack_stations": random.randint(0, 3),
            "lubrication_type": "bentonite",
            "chainage_from": rand_float(0, 10),
            "chainage_to": rand_float(50, 500),
            "status": random.choice(["planned", "jacking", "jacking", "breakthrough", "completed"]),
            "start_date": rand_date(date(2025, 1, 1), date(2025, 3, 1)).isoformat(),
        })
    return items

def generate_mtbm_thrust():
    items = []
    for mtbm in MTBM_DRIVES:
        for pipe in range(1, random.randint(10, 40)):
            items.append({
                "id": f"thr-{mtbm['id']}-{pipe:04d}",
                "mtbm_drive_id": mtbm["id"],
                "pipe_no": pipe,
                "recorded_at": rand_date(date(2025, 2, 1), date(2025, 6, 30)).isoformat(),
                "thrust_force_kN": rand_float(200, 8000),
                "thrust_pressure_bar": rand_float(20, 350),
                "push_ram_extent_mm": random.randint(500, 3500),
                "advance_speed_mmmin": rand_float(5, 80),
                "torque_kNm": rand_float(50, 2000),
                "torque_pct": rand_float(10, 90),
                "slurry_pressure_bar": rand_float(0.5, 4.0),
                "slurry_flow_m3h": rand_float(50, 300),
                "face_pressure_bar": rand_float(0.5, 3.5),
                "penetration_rate_mm_min": rand_float(5, 60),
                "alignment_vertical_mm": rand_float(-20, 20),
                "alignment_horizontal_mm": rand_float(-15, 15),
                "rod_count": random.randint(1, 20),
                "water_inflow_lmin": rand_float(0, 30),
            })
    return items

def generate_mtbm_lubrication():
    items = []
    for mtbm in MTBM_DRIVES:
        for _ in range(random.randint(15, 30)):
            items.append({
                "id": f"lub-{random.randint(1,9999):04d}",
                "mtbm_drive_id": mtbm["id"],
                "pipe_no": random.randint(1, 40),
                "recorded_at": rand_date(date(2025, 2, 1), date(2025, 6, 30)).isoformat(),
                "lubricant_type": random.choice(["bentonite", "polymer", "foam", "combined"]),
                "injection_pressure_bar": rand_float(0.5, 5.0),
                "flow_rate_lmin": rand_float(10, 150),
                "total_volume_m3": rand_float(0.5, 50),
                "density_kgm3": rand_float(1030, 1200),
                "viscosity_cp": rand_float(15, 60),
                "marsh_viscosity_sec": rand_float(30, 80),
                "filtrate_loss_ml": rand_float(5, 25),
                "ph_level": rand_float(7, 11),
            })
    return items

def generate_mtbm_survey():
    items = []
    for mtbm in MTBM_DRIVES:
        for pipe in range(1, random.randint(10, 40)):
            dev_v = rand_float(-30, 30)
            dev_h = rand_float(-25, 25)
            items.append({
                "id": f"surv-{mtbm['id']}-{pipe:04d}",
                "mtbm_drive_id": mtbm["id"],
                "pipe_no": pipe,
                "surveyed_at": rand_date(date(2025, 2, 1), date(2025, 6, 30)).isoformat(),
                "northing_m": rand_float(500000, 500100),
                "easting_m": rand_float(200000, 200100),
                "elevation_m": rand_float(10, 50),
                "chainage_m": rand_float(0, 120),
                "deviation_vertical_mm": dev_v,
                "deviation_horizontal_mm": dev_h,
                "deviation_roll_deg": rand_float(-1, 1),
                "deviation_yaw_deg": rand_float(-1, 1),
                "deviation_pitch_deg": rand_float(-1, 1),
                "instrument_type": random.choice(["gyro", "total_station", "laser"]),
                "survey_by": random.choice(["Survey Team A", "Survey Team B"]),
            })
    return items

def generate_shafts():
    items = []
    for i, stype in enumerate(["launch", "reception", "intermediate", "ventilation"]):
        items.append({
            "id": f"shaft-{i+1:04d}",
            "project_id": "p001",
            "shaft_code": f"SH-{i+1:03d}",
            "shaft_name": f"{stype.title()} Shaft {i+1}",
            "shaft_type": stype,
            "construction_method": random.choice(["secant_pile", "sheet_pile", "diaphragm_wall", "caisson", "cut_and_cover"]),
            "diameter_m": rand_float(4, 15),
            "depth_m": rand_float(8, 35),
            "wall_thickness_mm": random.choice([800, 1000, 1200]),
            "excavation_method": "mechanical",
            "support_system": "RC lining",
            "dewatering_method": random.choice(["wellpoints", "deep_well", "sump_pumping"]),
            "chainage": rand_float(0, 200),
            "start_date": rand_date(date(2024, 8, 1), date(2025, 1, 1)).isoformat(),
            "completion_date": rand_date(date(2025, 3, 1), date(2025, 6, 30)).isoformat(),
            "status": random.choice(["completed", "completed", "equipment_install", "base_slab"]),
        })
    return items

def generate_shaft_equipment():
    items = []
    for shaft in [f"shaft-{i+1:04d}" for i in range(4)]:
        for etype in ["crane", "ventilation_fan", "pump", "transformer", "control_panel", "lighting"]:
            if random.random() < 0.3:
                continue
            items.append({
                "id": f"shfeq-{shaft}-{etype}",
                "shaft_id": shaft,
                "equipment_type": etype,
                "equipment_name": f"{etype.replace('_',' ').title()} Unit",
                "manufacturer": random.choice(["Siemens", "ABB", "Atlas Copco", "Grundfos"]),
                "model": f"{etype[:2].upper()}-{random.randint(100,999)}",
                "serial_number": f"SER-{random.randint(10000,99999)}",
                "installed_at": rand_date(date(2025, 2, 1), date(2025, 5, 1)).isoformat(),
                "commissioned_at": rand_date(date(2025, 3, 1), date(2025, 6, 1)).isoformat(),
                "rated_capacity": f"{rand_float(10, 500, 0)} kW",
                "power_kw": rand_float(10, 500),
                "weight_kg": rand_float(100, 5000, 0),
                "location_in_shaft": random.choice(["Top deck", "Bottom", "Mid-level"]),
                "status": random.choice(["installed", "commissioned", "operational"]),
            })
    return items

def generate_cross_passages():
    items = []
    for i in range(1, 6):
        items.append({
            "id": f"cp-{i:04d}",
            "project_id": "p001",
            "passage_code": f"CP-{i:03d}",
            "passage_name": f"Cross Passage {i}",
            "chainage": rand_float(50 + (i-1)*40, 50 + i*40),
            "construction_method": random.choice(["NATM", "SCL", "drill_blast", "pipe_jacking"]),
            "length_m": rand_float(3, 12),
            "width_m": rand_float(2.5, 4.5),
            "height_m": rand_float(2.5, 4.0),
            "lining_type": "shotcrete + RC",
            "water_proofing": random.choice(["PVC membrane", "sprayed membrane", "bentonite panels"]),
            "start_date": rand_date(date(2025, 3, 1), date(2025, 5, 1)).isoformat(),
            "completion_date": rand_date(date(2025, 5, 1), date(2025, 6, 30)).isoformat(),
            "status": random.choice(["planned", "excavation", "lining", "completed"]),
        })
    return items

def generate_grouting():
    items = []
    for i in range(1, 61):
        gtype = random.choice(["contact", "void", "consolidation", "curtain", "compensation", "backfill", "annulus", "pre_excavation"])
        items.append({
            "id": f"grout-{i:04d}",
            "project_id": "p001",
            "grouting_type": gtype,
            "location_type": random.choice(["TBM", "NATM", "shaft", "cross_passage", "MTBM"]),
            "location_id": random.choice(["shaft-0001", "shaft-0002", "cp-0001", "d-001"]),
            "chainage": rand_float(0, 200),
            "grout_date": rand_date(date(2025, 1, 1), date(2025, 6, 30)).isoformat(),
            "grout_mix_type": random.choice(["cement_bentonite", "cement_silicate", "polyurethane", "epoxy"]),
            "grout_density_kgm3": rand_float(1200, 2200),
            "wc_ratio": rand_float(0.4, 1.5),
            "pressure_bar": rand_float(1, 20),
            "flow_rate_lmin": rand_float(10, 150),
            "volume_planned_m3": rand_float(0.5, 30),
            "volume_actual_m3": rand_float(0.3, 35),
            "injection_point": random.choice(["Hole 1", "Hole 2", "Port A", "Port B"]),
            "number_of_holes": random.randint(1, 8),
            "spacing_m": rand_float(0.5, 3),
            "take_kgm": rand_float(10, 200),
            "supervisor": random.choice(["Grouting Engineer A", "Grouting Engineer B"]),
        })
    return items

def generate_settlement():
    items = []
    for pt in range(1, 31):
        for _ in range(random.randint(5, 20)):
            s = rand_float(-30, 5)  # negative = heave
            items.append({
                "id": f"settle-pt{pt:04d}-{random.randint(1,99):02d}",
                "project_id": "p001",
                "point_id": f"SM-{pt:04d}",
                "point_type": random.choice(["surface", "subsurface", "building", "utility", "pavement", "track"]),
                "northing_m": rand_float(500010, 500100),
                "easting_m": rand_float(200010, 200100),
                "elevation_ref_m": rand_float(10, 45),
                "chainage": rand_float(0, 200),
                "offset_m": rand_float(-30, 30),
                "monitored_at": rand_date(date(2025, 1, 1), date(2025, 6, 30)).isoformat(),
                "settlement_mm": s,
                "cumulative_settlement_mm": s * random.uniform(0.8, 1.2),
                "settlement_rate_mmday": rand_float(-2, 2, 4),
                "horizontal_displacement_mm": rand_float(-5, 5),
                "tilt_ratio": rand_float(0, 0.005, 6),
                "strain_micron": rand_float(0, 500),
                "instrument_type": random.choice(["leveling", "total_station", "inclinometer", "piezometer", "extensometer", "tiltmeter"]),
                "reading_accuracy_mm": rand_float(0.1, 0.5, 3),
                "alarm_threshold_mm": 25,
                "alarm_triggered": s < -25 or s > 5,
                "reading_by": random.choice(["Survey Team", "Monitoring Team"]),
            })
    return items

def main():
    data = {
        "generated_at": datetime.utcnow().isoformat(),
        "excavation_logs": generate_excavation(),
        "shotcrete": generate_shotcrete(),
        "rock_bolts": generate_rock_bolts(),
        "steel_sets": generate_steel_sets(),
        "convergence": generate_convergence(),
        "face_mapping": generate_face_mapping(),
        "mtbm_drives": generate_mtbm_drives(),
        "mtbm_thrust": generate_mtbm_thrust(),
        "mtbm_lubrication": generate_mtbm_lubrication(),
        "mtbm_survey": generate_mtbm_survey(),
        "shafts": generate_shafts(),
        "shaft_equipment": generate_shaft_equipment(),
        "cross_passages": generate_cross_passages(),
        "grouting": generate_grouting(),
        "settlement": generate_settlement(),
    }
    OUTPUT.write_text(json.dumps(data, indent=2, ensure_ascii=False, default=str))
    print(f"[NATM Generator] Written {OUTPUT}")
    for k, v in data.items():
        print(f"  {k}: {len(v)}")

if __name__ == "__main__":
    main()