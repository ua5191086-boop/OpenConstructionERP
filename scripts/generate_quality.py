#!/usr/bin/env python3
"""
OpenConstructionERP — Quality Management Data Generator
Generates test data: ITPs, Inspections, Test Results, NCRs, Corrective Actions, Calibration, Quality Metrics
"""
import json, random
from datetime import date, timedelta
from pathlib import Path

random.seed(42)
OUTPUT = Path(__file__).parent.parent / "apps" / "web" / "quality_data.json"

PROJECTS = [
    {"id": 1, "name": "Metro Line 3 — Central Corridor"},
    {"id": 2, "name": "Highway Bypass — Northern Ring Road"},
    {"id": 3, "name": "Water Treatment Plant — Phase II"},
    {"id": 4, "name": "Railway Bridge — River Delta Crossing"},
    {"id": 5, "name": "Port Expansion — Container Terminal"},
]

def rand_date(start=2024, end=2026):
    s = date(start, 1, 1); e = date(end, 6, 30)
    return s + timedelta(days=random.randint(0, (e-s).days))

def pick(*choices): return random.choice(choices)

def generate_itps():
    items = []
    for proj in PROJECTS:
        n = random.randint(5, 12)
        for i in range(1, n+1):
            items.append({
                "id": f"itp-{proj['id']}-{i:04d}",
                "project_id": proj["id"], "project_name": proj["name"],
                "itp_number": i, "itp_code": f"ITP-{i:04d}",
                "itp_name": f"{pick(['Concrete','Steel','Piping','Electrical','Mechanical','Welding','Coating','Civil'])} Inspection & Test Plan #{i}",
                "itp_type": pick(["inspection","test","combined","witness","hold_point"]),
                "description": f"ITP for {pick(['concrete placement','steel erection','pipeline welding','cable pulling','mechanical alignment'])}",
                "status": pick(["active","active","completed","draft"]),
            })
    return items

def generate_inspections():
    items = []
    for proj in PROJECTS:
        n = random.randint(15, 30)
        for i in range(1, n+1):
            d = rand_date()
            items.append({
                "id": f"ins-{proj['id']}-{i:04d}",
                "project_id": proj["id"], "project_name": proj["name"],
                "record_number": i, "record_code": f"INS-{i:04d}",
                "title": f"{pick(['Visual','Dimensional','Concrete','Weld','Coating','Electrical','Mechanical'])} Inspection #{i}",
                "inspection_type": pick(["visual","dimensional","weld","concrete","electrical","mechanical","civil","structural"]),
                "inspector": pick(["Inspector A","Inspector B","QC Engineer","Third Party"]),
                "inspection_date": d.isoformat(),
                "result": pick(["pass","pass","pass","fail","conditional_pass","pending"]),
                "defects_found": random.randint(0, 5),
            })
    return items

def generate_test_results():
    items = []
    for proj in PROJECTS:
        n = random.randint(10, 20)
        for i in range(1, n+1):
            d = rand_date()
            items.append({
                "id": f"tst-{proj['id']}-{i:04d}",
                "project_id": proj["id"], "project_name": proj["name"],
                "test_number": i, "test_code": f"TST-{i:04d}",
                "test_name": f"{pick(['Concrete Cylinder','Tensile','Compression','Hardness','NDT','Pressure','Leak'])} Test #{i}",
                "test_type": pick(["concrete","steel","soil","weld","pressure","material"]),
                "test_date": d.isoformat(),
                "measured_value": round(random.uniform(10, 500), 2),
                "min_acceptable": round(random.uniform(5, 100), 2),
                "max_acceptable": round(random.uniform(100, 600), 2),
                "result": pick(["pass","pass","pass","pass","fail","conditional"]),
                "lab_name": pick(["Lab A","Lab B","Site Lab","External Lab"]),
            })
    return items

def generate_ncrs():
    items = []
    for proj in PROJECTS:
        n = random.randint(5, 12)
        for i in range(1, n+1):
            d = rand_date()
            sev = pick(["minor","minor","major","critical"])
            items.append({
                "id": f"ncr-{proj['id']}-{i:04d}",
                "project_id": proj["id"], "project_name": proj["name"],
                "ncr_number": i, "ncr_code": f"NCR-{i:04d}",
                "title": f"Non-Conformance: {pick(['Material','Workmanship','Dimension','Weld','Coating','Documentation'])} Issue #{i}",
                "ncr_category": pick(["material","workmanship","dimensional","weld","documentation"]),
                "severity": sev,
                "source": pick(["inspection","test","audit","customer"]),
                "description": f"Defect found during {pick(['visual inspection','dimensional check','weld inspection','material testing'])}",
                "discovered_date": d.isoformat(),
                "discovered_by": pick(["QC Inspector","Engineer","Supervisor","Third Party"]),
                "root_cause": pick(["human_error","material","procedure","equipment","training"]) if random.random() > 0.3 else None,
                "disposition_type": pick(["rework","repair","use_as_is","scrap"]) if random.random() > 0.2 else None,
                "rework_cost": round(random.uniform(100, 50000), 2),
                "schedule_impact": random.randint(0, 14),
                "status": pick(["open","open","investigating","disposition","closed","closed"]),
            })
    return items

def generate_corrective_actions():
    items = []
    for proj in PROJECTS:
        n = random.randint(4, 10)
        for i in range(1, n+1):
            items.append({
                "id": f"ca-{proj['id']}-{i:04d}",
                "project_id": proj["id"], "project_name": proj["name"],
                "ca_number": i, "ca_code": f"CA-{i:04d}",
                "title": f"Corrective Action: {pick(['Rework','Procedure Update','Training','Material Replacement','Design Change'])} #{i}",
                "action_type": pick(["corrective","preventive","improvement"]),
                "assigned_to": pick(["QC Team","Engineer A","Supervisor B","Construction Manager"]),
                "priority": pick(["low","medium","high","critical"]),
                "due_date": rand_date(2025, 2026).isoformat(),
                "effectiveness": pick(["effective","partially_effective","not_verified"]) if random.random() > 0.3 else None,
                "status": pick(["open","in_progress","implemented","verified","closed","closed"]),
            })
    return items

def generate_calibration():
    items = []
    equipments = [
        {"name": "Digital Pressure Gauge", "model": "DPG-200", "freq": 180},
        {"name": "Universal Testing Machine", "model": "UTM-500", "freq": 365},
        {"name": "Concrete Compression Machine", "model": "CCM-3000", "freq": 365},
        {"name": "Theodolite", "model": "THEO-5", "freq": 365},
        {"name": "Total Station", "model": "TS-2000", "freq": 365},
        {"name": "Temperature Probe", "model": "TP-100", "freq": 90},
        {"name": "Caliper Digital", "model": "CD-200", "freq": 180},
        {"name": "Torque Wrench", "model": "TW-500", "freq": 180},
        {"name": "Multimeter", "model": "MM-1000", "freq": 365},
        {"name": "Flow Meter", "model": "FM-200", "freq": 365},
    ]
    for proj in PROJECTS:
        for eq in equipments:
            last = rand_date(2024, 2025)
            items.append({
                "id": f"cal-{proj['id']}-{eq['name'].lower().replace(' ','_')}",
                "project_id": proj["id"], "project_name": proj["name"],
                "equipment_name": eq["name"], "equipment_model": eq["model"],
                "serial_number": f"SN-{random.randint(1000,9999)}",
                "calibration_frequency_days": eq["freq"],
                "last_calibration_date": last.isoformat(),
                "next_calibration_date": (last + timedelta(days=eq["freq"])).isoformat(),
                "calibration_result": pick(["pass","pass","pass","conditional","expired"]),
                "status": pick(["active","active","active","expired"]),
            })
    return items

def generate_quality_metrics():
    items = []
    for proj in PROJECTS:
        for m in range(12):
            from datetime import date as dt
            report_month = dt(2025, 1, 1) if m == 0 else dt(2025, (m % 12) + 1, 1)
            total_ins = random.randint(20, 80)
            failed_ins = random.randint(0, int(total_ins * 0.15))
            total_tst = random.randint(15, 50)
            failed_tst = random.randint(0, int(total_tst * 0.1))
            ncrs = random.randint(0, 8)
            items.append({
                "id": f"qm-{proj['id']}-{report_month.isoformat()}",
                "project_id": proj["id"], "project_name": proj["name"],
                "report_month": report_month.isoformat(),
                "total_inspections": total_ins,
                "inspections_passed": total_ins - failed_ins,
                "inspections_failed": failed_ins,
                "total_tests": total_tst,
                "tests_passed": total_tst - failed_tst,
                "tests_failed": failed_tst,
                "ncr_opened": ncrs, "ncr_closed": random.randint(0, ncrs),
                "ncr_critical": random.randint(0, 2),
                "first_pass_yield": round(random.uniform(85, 99.5), 2),
                "rework_cost": round(random.uniform(0, 25000), 2),
            })
    return items

def main():
    data = {
        "generated_at": date.today().isoformat(),
        "summary": {}, "itps": generate_itps(),
        "inspections": generate_inspections(),
        "test_results": generate_test_results(),
        "ncrs": generate_ncrs(),
        "corrective_actions": generate_corrective_actions(),
        "calibration": generate_calibration(),
        "quality_metrics": generate_quality_metrics(),
    }
    for k in ["itps","inspections","test_results","ncrs","corrective_actions","calibration","quality_metrics"]:
        data["summary"][k] = len(data[k])
        print(f"[QM] Generated {len(data[k])} {k}")

    OUTPUT.parent.mkdir(parents=True, exist_ok=True)
    with open(OUTPUT, "w", encoding="utf-8") as f:
        json.dump(data, f, indent=2, ensure_ascii=False)
    print(f"[QM] Output: {OUTPUT}")

if __name__ == "__main__":
    main()