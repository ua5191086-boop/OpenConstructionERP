#!/usr/bin/env python3
"""
OpenConstructionERP — HSE Data Generator
Generates test data: Incidents, Permits, Audits, Inspections, Training,
PPE, Drills, Statistics, Emergency Plans, Chemicals
"""
import json
import random
from datetime import datetime, date, timedelta
from pathlib import Path

random.seed(42)

OUTPUT = Path(__file__).parent.parent / "apps" / "web" / "hse_data.json"

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

def generate_incidents():
    items = []
    for proj in PROJECTS:
        n = random.randint(8, 18)
        for i in range(1, n + 1):
            severity = random.choice(["minor", "minor", "moderate", "major", "critical"])
            itype = random.choice(["near_miss", "near_miss", "first_aid", "medical_treatment", "lost_time_injury", "property_damage", "environmental", "fire"])
            inv_status = random.choice(["open", "investigating", "report_draft", "report_approved", "closed", "closed"])
            incident_date = rand_date(2024, 2026)
            lost_days = random.randint(1, 30) if itype == "lost_time_injury" else 0
            items.append({
                "id": f"inc-{proj['id']}-{i:04d}",
                "project_id": proj["id"],
                "project_name": proj["name"],
                "incident_number": i,
                "incident_code": f"INC-{i:04d}",
                "title": random.choice([
                    "Worker slipped on wet surface",
                    "Minor burn from hot pipe",
                    "Near miss — falling object from crane",
                    "Chemical splash to eye",
                    "Back injury during manual lifting",
                    "Fire in electrical panel room",
                    "Excavator struck underground utility",
                    "Pinch point injury on conveyor",
                    "Fall from scaffold — harness arrested",
                    "Welding flash burn",
                    "Gas leak detected in confined space",
                    "Vehicle collision in laydown area",
                ]),
                "description": "Detailed incident description per HSE reporting procedure.",
                "incident_type": itype,
                "severity": severity,
                "incident_date": incident_date.isoformat(),
                "incident_time": f"{random.randint(6,18):02d}:{random.randint(0,59):02d}",
                "location": random.choice(["Section A", "Tunnel Portal", "Workshop", "Pier P05", "Laydown Yard", "Admin Building"]),
                "area": random.choice(["Excavation Zone", "Concrete Works", "Steel Erection", "MEP Area", "Storage"]),
                "reported_by": random.choice(["Foreman A", "Operator B", "Safety Officer", "Worker"]),
                "reported_at": incident_date.isoformat(),
                "affected_person": random.choice(["Worker 1", "Technician C", "Operator D"]) if itype not in ["near_miss", "property_damage", "environmental"] else None,
                "lost_days": lost_days,
                "medical_cost": round(random.uniform(100, 50000), 2) if itype in ["first_aid", "medical_treatment", "lost_time_injury"] else 0,
                "property_cost": round(random.uniform(500, 100000), 2) if itype in ["property_damage", "fire"] else 0,
                "total_cost": round(random.uniform(500, 150000), 2),
                "root_cause": "Inadequate hazard identification and risk assessment." if inv_status != "open" else None,
                "investigation_status": inv_status,
                "investigation_lead": random.choice(["HSE Manager", "Safety Officer", "Project Engineer"]),
                "is_reportable": itype in ["lost_time_injury", "fatality", "environmental"],
                "status": "closed" if inv_status in ["closed", "report_approved"] else "open",
            })
    return items

def generate_permits():
    items = []
    for proj in PROJECTS:
        n = random.randint(12, 25)
        for i in range(1, n + 1):
            ptype = random.choice(["hot_work", "confined_space", "work_at_height", "excavation", "electrical", "lifting", "chemical"])
            valid_from = rand_date(2025, 2026)
            valid_to = valid_from + timedelta(days=random.randint(1, 7))
            status = random.choice(["issued", "active", "active", "expired", "closed", "cancelled", "suspended"])
            items.append({
                "id": f"ptw-{proj['id']}-{i:04d}",
                "project_id": proj["id"],
                "project_name": proj["name"],
                "permit_number": i,
                "permit_code": f"PTW-{i:04d}",
                "permit_type": ptype,
                "title": f"Permit for {ptype.replace('_', ' ').title()} — {random.choice(['Welding at Pier P05', 'Entry to Manhole M-03', 'Scaffold Erection at B2', 'Excavation near Utility', 'Crane Lift of Steel Beam', 'Application of Epoxy Coating'])}",
                "description": f"Scope of work for {ptype.replace('_', ' ')}.",
                "location": random.choice(["Pier P05", "Tunnel Section 2", "Building Core", "Laydown Yard", "Electrical Room"]),
                "work_description": "Detailed work scope per method statement and risk assessment.",
                "issuing_authority": "HSE Department",
                "permit_holder": random.choice(["Contractor A", "Subcontractor B", "Main Contractor"]),
                "responsible_person": random.choice(["Site Engineer", "Foreman", "Supervisor"]),
                "control_measures": "1. Fire extinguisher at site\\n2. Gas monitor in confined space\\n3. Full body harness required",
                "ppe_required": "Hard hat, safety glasses, gloves, steel-toe boots, full body harness",
                "valid_from": valid_from.isoformat(),
                "valid_to": valid_to.isoformat(),
                "status": status,
                "issued_at": valid_from.isoformat() if status != "draft" else None,
                "issued_by": "HSE Manager",
                "remarks": "All control measures verified prior to work start.",
            })
    return items

def generate_audits():
    items = []
    for proj in PROJECTS:
        n = random.randint(3, 8)
        for i in range(1, n + 1):
            audit_date = rand_date(2025, 2026)
            types = ["internal", "external", "regulatory", "certification"]
            items.append({
                "id": f"aud-{proj['id']}-{i:04d}",
                "project_id": proj["id"],
                "project_name": proj["name"],
                "audit_number": i,
                "audit_code": f"AUD-{i:04d}",
                "audit_type": random.choice(types),
                "title": f"{random.choice(['ISO 45001', 'Internal HSE', 'Regulatory Compliance', 'Safety Management', 'Environmental'])} Audit #{i}",
                "scope": "Audit of all project HSE management system elements.",
                "criteria": "ISO 45001:2018, local regulations, company HSE policy",
                "lead_auditor": random.choice(["Auditor A", "Auditor B", "External Consultant"]),
                "audit_date": audit_date.isoformat(),
                "location": "Project Site Office",
                "non_conformities": random.randint(0, 8),
                "observations": random.randint(2, 15),
                "score_pct": round(random.uniform(60, 98), 2),
                "status": random.choice(["planned", "in_progress", "completed", "reviewed", "closed"]),
                "findings_summary": "The project demonstrates good HSE practices. Areas for improvement identified.",
                "follow_up_date": (audit_date + timedelta(days=30)).isoformat(),
            })
    return items

def generate_inspections():
    items = []
    for proj in PROJECTS:
        n = random.randint(15, 30)
        for i in range(1, n + 1):
            insp_date = rand_date(2025, 2026)
            items.append({
                "id": f"insp-{proj['id']}-{i:04d}",
                "project_id": proj["id"],
                "project_name": proj["name"],
                "inspection_number": i,
                "inspection_code": f"INSP-{i:04d}",
                "inspection_type": random.choice(["routine", "focused", "toolbox_talk", "safety_walk", "daily"]),
                "title": f"{random.choice(['Weekly Safety Walk', 'Daily Site Inspection', 'Focused Electrical Safety', 'Toolbox Talk Session', 'Scaffold Inspection'])} #{i}",
                "location": random.choice(["Site Wide", "Pier P05 Area", "Tunnel Section", "Workshop"]),
                "inspector": random.choice(["Safety Officer A", "Safety Officer B", "HSE Manager", "Supervisor"]),
                "inspection_date": insp_date.isoformat(),
                "findings": "General safety compliance observed. Few minor issues noted.",
                "violations_found": random.randint(0, 5),
                "violations_resolved": random.randint(0, 4),
                "severity": random.choice(["low", "low", "medium", "high"]),
                "status": random.choice(["completed", "completed", "reviewed", "closed"]),
                "action_items": "1. Correct housekeeping in Section A\\n2. Replace damaged safety sign",
                "follow_up_date": (insp_date + timedelta(days=7)).isoformat(),
            })
    return items

def generate_training():
    items = []
    for proj in PROJECTS:
        n = random.randint(8, 15)
        for i in range(1, n + 1):
            training_date = rand_date(2025, 2026)
            ttype = random.choice(["safety_induction", "refresher", "first_aid", "fire_safety", "work_at_height", "confined_space", "chemical_handling", "lifting_ops", "emergency_response"])
            items.append({
                "id": f"trn-{proj['id']}-{i:04d}",
                "project_id": proj["id"],
                "project_name": proj["name"],
                "training_number": i,
                "training_code": f"TRN-{i:04d}",
                "training_name": f"{ttype.replace('_', ' ').title()} Training Session #{i}",
                "training_type": ttype,
                "description": f"Mandatory {ttype.replace('_', ' ')} training for all project personnel.",
                "trainer": random.choice(["Trainer A", "Trainer B", "External Consultant", "HSE Team"]),
                "training_date": training_date.isoformat(),
                "duration_hours": round(random.uniform(1, 8), 1),
                "location": random.choice(["Site Training Room", "Off-site Facility", "Online"]),
                "attendees": random.randint(5, 30),
                "max_attendees": 30,
                "status": random.choice(["completed", "completed", "completed", "scheduled", "planned"]),
                "certificate_type": "Safety Training Certificate",
                "certificate_validity_days": random.choice([365, 730]),
                "cost_per_person": round(random.uniform(0, 100), 2),
                "total_cost": round(random.uniform(0, 3000), 2),
            })
    return items

def generate_ppe():
    items = []
    ppe_items = [
        {"code": "PPE-HARD-HAT", "name": "Hard Hat JSP Mk8", "cat": "head", "cost": 15.50},
        {"code": "PPE-SAFETY-GLASS", "name": "Safety Glasses Clear", "cat": "eye", "cost": 5.00},
        {"code": "PPE-EAR-PLUG", "name": "Ear Plugs (box 200)", "cat": "hearing", "cost": 12.00},
        {"code": "PPE-RESP-N95", "name": "Respirator N95 Mask", "cat": "respiratory", "cost": 3.50},
        {"code": "PPE-GLOVE-CUT", "name": "Cut-Resistant Gloves L5", "cat": "hand", "cost": 22.00},
        {"code": "PPE-BOOT-STEEL", "name": "Safety Boots Steel Toe", "cat": "foot", "cost": 85.00},
        {"code": "PPE-HARNESS", "name": "Full Body Harness", "cat": "fall_protection", "cost": 180.00},
        {"code": "PPE-VEST-HIVIS", "name": "Hi-Vis Vest Class 3", "cat": "high_visibility", "cost": 12.00},
        {"code": "PPE-GOGGLE", "name": "Chemical Splash Goggles", "cat": "eye", "cost": 8.50},
        {"code": "PPE-WELD-SHIELD", "name": "Welding Face Shield", "cat": "face", "cost": 45.00},
    ]
    for proj in PROJECTS:
        for p in ppe_items:
            items.append({
                "id": f"ppe-{proj['id']}-{p['code'].lower()}",
                "project_id": proj["id"],
                "project_name": proj["name"],
                "ppe_code": p["code"],
                "ppe_name": p["name"],
                "ppe_category": p["cat"],
                "manufacturer": random.choice(["3M", "JSP", "MSA", "Honeywell"]),
                "model": f"Model-{random.randint(100,999)}",
                "size": random.choice(["One Size", "S", "M", "L", "XL"]),
                "quantity_issued": random.randint(20, 500),
                "quantity_stock": random.randint(0, 200),
                "reorder_level": 20,
                "unit_cost": p["cost"],
                "expiry_date": rand_date(2026, 2028).isoformat(),
                "storage_location": f"Store Room {chr(65 + random.randint(0,3))}",
                "is_active": True,
            })
    return items

def generate_drills():
    items = []
    for proj in PROJECTS:
        n = random.randint(3, 8)
        for i in range(1, n + 1):
            drill_date = rand_date(2025, 2026)
            dtype = random.choice(["fire", "evacuation", "first_aid", "confined_space_rescue", "height_rescue", "chemical_spill"])
            items.append({
                "id": f"drl-{proj['id']}-{i:04d}",
                "project_id": proj["id"],
                "project_name": proj["name"],
                "drill_number": i,
                "drill_code": f"DRL-{i:04d}",
                "drill_name": f"{dtype.replace('_', ' ').title()} Drill — {drill_date.strftime('%b %Y')}",
                "drill_type": dtype,
                "description": f"Monthly {dtype.replace('_', ' ')} drill for project personnel.",
                "location": random.choice(["Main Site", "Tunnel Portal", "Workshop", "Admin Building"]),
                "drill_date": drill_date.isoformat(),
                "participants": random.randint(10, 50),
                "duration_minutes": random.randint(15, 90),
                "evaluator": random.choice(["HSE Manager", "Safety Officer", "Fire Marshall"]),
                "score_pct": round(random.uniform(60, 100), 2),
                "observations": "Team responded promptly. Evacuation route clear.",
                "improvements": "Brief radio communication protocol.",
                "status": random.choice(["completed", "completed", "debriefed", "planned"]),
            })
    return items

def generate_statistics():
    items = []
    for proj in PROJECTS:
        start = date(2024, 7, 1)
        for m in range(12):
            report_month = date(start.year, start.month + (m % 12), 1) if start.month + (m % 12) <= 12 else date(start.year + 1, start.month + (m % 12) - 12, 1)
            manhours = random.randint(20000, 80000)
            lti = random.randint(0, 2)
            trir = round(lti * 1000000 / manhours * 200000, 2) if manhours > 0 else 0
            items.append({
                "id": f"stats-{proj['id']}-{report_month.isoformat()}",
                "project_id": proj["id"],
                "project_name": proj["name"],
                "report_month": report_month.isoformat(),
                "manhours": manhours,
                "lost_time_injuries": lti,
                "recordable_injuries": random.randint(0, 4),
                "fatalities": 0,
                "near_misses": random.randint(2, 15),
                "first_aid_cases": random.randint(1, 8),
                "property_damage": random.randint(0, 3),
                "environmental_incidents": random.randint(0, 2),
                "fire_incidents": random.randint(0, 1),
                "lti_frequency": round(lti * 1000000 / manhours, 2) if manhours > 0 else 0,
                "total_recordable_rate": trir,
                "days_since_last_lti": random.randint(30, 365),
                "days_since_last_fatality": random.randint(365, 1000),
                "safety_training_hours": random.randint(100, 500),
                "inspections_conducted": random.randint(10, 40),
                "audits_conducted": random.randint(0, 3),
                "permits_issued": random.randint(15, 50),
            })
    return items

def generate_emergency_plans():
    items = []
    plan_types = [
        ("general_emergency", "General Emergency Response Plan"),
        ("fire", "Fire Emergency Plan"),
        ("spill", "Chemical Spill Response Plan"),
        ("collapse", "Structural Collapse Emergency Plan"),
        ("flood", "Flood Response Plan"),
        ("medical", "Medical Emergency Evacuation Plan"),
        ("confined_space", "Confined Space Rescue Plan"),
        ("height_rescue", "Working at Height Rescue Plan"),
    ]
    for proj in PROJECTS:
        for i, (ptype, pname) in enumerate(plan_types, 1):
            items.append({
                "id": f"ep-{proj['id']}-{i:04d}",
                "project_id": proj["id"],
                "project_name": proj["name"],
                "plan_number": i,
                "plan_code": f"EP-{i:04d}",
                "plan_name": pname,
                "plan_type": ptype,
                "description": f"Comprehensive emergency response plan for {ptype.replace('_', ' ')} scenarios.",
                "response_procedure": "1. Raise alarm\\n2. Evacuate affected area\\n3. Contact emergency services\\n4. Account for all personnel\\n5. Incident investigation",
                "evacuation_routes": "Primary: Main gate via East corridor. Secondary: Emergency exit West side.",
                "assembly_points": "Main Assembly: Site Car Park. Secondary: North Gate.",
                "emergency_contacts": "Site HSE: Ext 500\\nFire Dept: 997\\nAmbulance: 998\\nPolice: 999",
                "responsible_person": "Project Manager",
                "deputy_person": "Construction Manager",
                "drill_frequency": "Monthly",
                "last_reviewed": rand_date(2024, 2025).isoformat(),
                "next_review": rand_date(2025, 2026).isoformat(),
                "status": random.choice(["active", "active", "active", "approved", "active"]),
                "version": f"{random.randint(1,3)}.{random.randint(0,2)}",
                "approval_date": rand_date(2024, 2025).isoformat(),
                "approved_by": "Project Director",
            })
    return items

def generate_chemicals():
    items = []
    chemicals = [
        {"code": "CHEM-DIESEL", "name": "Diesel Fuel", "cas": "68334-30-5", "hazard": "Flammable", "flammable": True},
        {"code": "CHEM-CONC-ADD", "name": "Concrete Plasticizer Additive", "cas": "9003-01-4", "hazard": "Irritant", "flammable": False},
        {"code": "CHEM-PAINT", "name": "Epoxy Paint (Solvent-based)", "cas": "108-88-3", "hazard": "Flammable, Toxic", "flammable": True, "toxic": True},
        {"code": "CHEM-ACID", "name": "Hydrochloric Acid 32%", "cas": "7647-01-0", "hazard": "Corrosive", "corrosive": True},
        {"code": "CHEM-GROUT", "name": "Cement Grout Chemical", "cas": "65997-15-1", "hazard": "Irritant", "flammable": False},
        {"code": "CHEM-SOLVENT", "name": "Industrial Degreaser Solvent", "cas": "127-18-4", "hazard": "Toxic", "toxic": True},
        {"code": "CHEM-LUBE", "name": "Multi-purpose Lubricant", "cas": "8002-05-9", "hazard": "Combustible", "flammable": True},
        {"code": "CHEM-WELD-GAS", "name": "Acetylene (Dissolved)", "cas": "74-86-2", "hazard": "Extremely Flammable", "flammable": True},
    ]
    for proj in PROJECTS:
        for c in chemicals:
            items.append({
                "id": f"chem-{proj['id']}-{c['code'].lower()}",
                "project_id": proj["id"],
                "project_name": proj["name"],
                "chemical_code": c["code"],
                "chemical_name": c["name"],
                "cas_number": c["cas"],
                "hazard_class": c["hazard"],
                "manufacturer": random.choice(["BASF", "Sika", "Shell", "Dow", "AkzoNobel"]),
                "supplier": random.choice(["Local Distributor", "National Supply Co", "Industrial Chem"]),
                "storage_location": f"Chem Store {chr(65 + random.randint(0,3))}",
                "max_quantity": round(random.uniform(50, 1000), 2),
                "unit": random.choice(["L", "kg", "m3"]),
                "is_hazardous": True,
                "is_flammable": c.get("flammable", False),
                "is_toxic": c.get("toxic", False),
                "is_corrosive": c.get("corrosive", False),
                "is_environmentally_hazardous": random.random() < 0.3,
                "sds_revision_date": rand_date(2023, 2025).isoformat(),
                "expiry_date": rand_date(2025, 2027).isoformat(),
                "is_active": True,
            })
    return items

def main():
    data = {
        "generated_at": datetime.utcnow().isoformat(),
        "summary": {
            "incidents": 0, "permits": 0, "audits": 0, "inspections": 0,
            "training": 0, "ppe": 0, "drills": 0, "statistics": 0,
            "emergency_plans": 0, "chemicals": 0,
        },
        "incidents": generate_incidents(),
        "permits": generate_permits(),
        "audits": generate_audits(),
        "inspections": generate_inspections(),
        "training": generate_training(),
        "ppe": generate_ppe(),
        "drills": generate_drills(),
        "statistics": generate_statistics(),
        "emergency_plans": generate_emergency_plans(),
        "chemicals": generate_chemicals(),
    }
    data["summary"]["incidents"] = len(data["incidents"])
    data["summary"]["permits"] = len(data["permits"])
    data["summary"]["audits"] = len(data["audits"])
    data["summary"]["inspections"] = len(data["inspections"])
    data["summary"]["training"] = len(data["training"])
    data["summary"]["ppe"] = len(data["ppe"])
    data["summary"]["drills"] = len(data["drills"])
    data["summary"]["statistics"] = len(data["statistics"])
    data["summary"]["emergency_plans"] = len(data["emergency_plans"])
    data["summary"]["chemicals"] = len(data["chemicals"])

    OUTPUT.parent.mkdir(parents=True, exist_ok=True)
    with open(OUTPUT, "w", encoding="utf-8") as f:
        json.dump(data, f, indent=2, ensure_ascii=False)

    for k, v in data["summary"].items():
        print(f"[HSE] Generated {v} {k}")
    print(f"[HSE] Output: {OUTPUT}")

if __name__ == "__main__":
    main()