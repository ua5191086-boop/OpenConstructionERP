#!/usr/bin/env python3
"""
OpenConstructionERP — Risk Management Data Generator
Generates test data: Categories, Registers, Matrices, Monte Carlo, Scenarios, Mitigation, Escalation, Dashboard
"""
import json, random
from datetime import date, timedelta
from pathlib import Path

random.seed(42)
OUTPUT = Path(__file__).parent.parent / "apps" / "web" / "risk_data.json"

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

def pick(*c): return random.choice(c)

def generate_categories():
    cats = [
        ("TECH", "Technical", "Design, engineering, technology risks"),
        ("SCH", "Schedule", "Time and delay risks"),
        ("COS", "Cost", "Budget and financial risks"),
        ("RES", "Resource", "Resource availability risks"),
        ("EXT", "External", "Regulatory, environmental, force majeure"),
        ("CON", "Contractual", "Contract and legal risks"),
        ("SFT", "Safety", "HSE risks"),
        ("QTY", "Quality", "Quality and workmanship risks"),
    ]
    items = []
    for i, (code, name, desc) in enumerate(cats):
        items.append({
            "id": f"cat-{i:04d}", "category_code": code,
            "category_name": name, "category_type": "threat",
            "description": desc, "sort_order": i, "is_active": True,
        })
    return items

def generate_registers():
    items = []
    ratings = ["very_low","low","medium","high","extreme"]
    for proj in PROJECTS:
        n = random.randint(10, 20)
        for i in range(1, n+1):
            prob = round(random.uniform(1, 5), 1)
            impact = round(random.uniform(1, 5), 1)
            rf = pick(*ratings)
            items.append({
                "id": f"rsk-{proj['id']}-{i:04d}",
                "project_id": proj["id"], "project_name": proj["name"],
                "risk_number": i, "risk_code": f"RSK-{i:04d}",
                "risk_name": f"{pick(['Design delay','Cost overrun','Subcontractor default','Material shortage','Weather delay','Geotechnical issue','Permit delay','Labor strike','Equipment failure','Quality defect','Scope creep','Interface issue'])} — {proj['name']}",
                "risk_type": pick(["threat","threat","threat","opportunity"]),
                "description": f"Risk description for risk item {i}",
                "probability_score": prob,
                "impact_score": impact,
                "risk_score": round(prob * impact, 1),
                "risk_rating": rf,
                "cost_impact": round(random.uniform(10000, 2000000), 2),
                "schedule_impact_days": random.randint(5, 120),
                "risk_owner": pick(["Project Manager","Construction Manager","Design Manager","Procurement Manager","HSE Manager"]),
                "risk_response": pick(["avoid","transfer","mitigate","accept","exploit"]),
                "status": pick(["identified","analyzed","response_planned","monitoring","closed"]),
            })
    return items

def generate_scenarios():
    items = []
    for proj in PROJECTS:
        n = random.randint(3, 7)
        for i in range(1, n+1):
            items.append({
                "id": f"scn-{proj['id']}-{i:04d}",
                "project_id": proj["id"], "project_name": proj["name"],
                "scenario_number": i, "scenario_code": f"SCN-{i:04d}",
                "scenario_name": f"{pick(['Delayed Permitting','Material Price Spike','Labor Shortage','Weather Delay','Design Rework','Ground Conditions Change'])} Scenario",
                "scenario_type": pick(["what_if","best_case","worst_case","most_likely","stress_test"]),
                "description": f"Analysis of {pick(['schedule','cost','combined'])} impact",
                "cost_impact_min": round(random.uniform(50000, 500000), 2),
                "cost_impact_max": round(random.uniform(500000, 5000000), 2),
                "cost_impact_ml": round(random.uniform(100000, 1000000), 2),
                "schedule_impact_min": random.randint(10, 30),
                "schedule_impact_max": random.randint(60, 180),
                "schedule_impact_ml": random.randint(30, 90),
                "probability_pct": round(random.uniform(10, 90), 2),
                "severity": pick(["low","medium","high","extreme"]),
                "status": pick(["draft","analyzed","reviewed","approved"]),
            })
    return items

def generate_mitigations():
    items = []
    for proj in PROJECTS:
        n = random.randint(8, 15)
        for i in range(1, n+1):
            items.append({
                "id": f"mit-{proj['id']}-{i:04d}",
                "project_id": proj["id"], "project_name": proj["name"],
                "action_number": i, "action_code": f"MIT-{i:04d}",
                "action_name": f"{pick(['Hire additional staff','Procure backup equipment','Engage consultant','Revise schedule','Add contingency','Training program','Redesign element'])} — Action {i}",
                "action_type": pick(["preventive","contingency","corrective","fallback"]),
                "assigned_to": pick(["Team A","Engineer B","Manager C","Contractor D"]),
                "budget": round(random.uniform(5000, 200000), 2),
                "due_date": rand_date(2025, 2026).isoformat(),
                "effectiveness": pick(["effective","partially_effective","pending_review"]),
                "status": pick(["planned","in_progress","completed","overdue"]),
            })
    return items

def generate_escalations():
    items = []
    for proj in PROJECTS:
        n = random.randint(2, 5)
        for i in range(1, n+1):
            items.append({
                "id": f"esc-{proj['id']}-{i:04d}",
                "project_id": proj["id"], "project_name": proj["name"],
                "escalation_number": i, "escalation_code": f"ESC-{i:04d}",
                "title": f"Escalation: {pick(['Budget overrun risk','Schedule delay','Quality issue','Safety concern'])} #{i}",
                "reason": f"Risk exceeds threshold requiring senior management attention",
                "escalated_to": pick(["Project Director","VP Operations","Client Representative"]),
                "escalated_by": pick(["Project Manager","Risk Manager"]),
                "decision": pick(["approved","approved","noted","deferred",None]),
                "status": pick(["escalated","acknowledged","responded","resolved","closed"]),
            })
    return items

def generate_mc_runs():
    items = []
    for proj in PROJECTS:
        n = random.randint(2, 5)
        for i in range(1, n+1):
            items.append({
                "id": f"mc-{proj['id']}-{i:04d}",
                "project_id": proj["id"], "project_name": proj["name"],
                "run_label": f"Monte Carlo {pick(['Cost','Schedule','Combined'])} Run {i}",
                "run_type": pick(["cost","schedule","combined"]),
                "iterations": 10000,
                "p10_value": round(random.uniform(1000000, 5000000), 2),
                "p50_value": round(random.uniform(1500000, 7000000), 2),
                "p90_value": round(random.uniform(2000000, 10000000), 2),
                "mean_value": round(random.uniform(1500000, 7000000), 2),
                "confidence_level": round(random.uniform(80, 95), 2),
                "status": pick(["completed","completed","completed","failed"]),
            })
    return items

def generate_dashboard():
    items = []
    for proj in PROJECTS:
        items.append({
            "id": f"dash-{proj['id']}",
            "project_id": proj["id"], "project_name": proj["name"],
            "snapshot_date": date.today().isoformat(),
            "total_risks": random.randint(15, 25),
            "open_risks": random.randint(5, 15),
            "extreme_risks": random.randint(0, 3),
            "high_risks": random.randint(2, 8),
            "medium_risks": random.randint(3, 10),
            "low_risks": random.randint(2, 8),
            "threats": random.randint(10, 18),
            "opportunities": random.randint(2, 6),
            "risk_exposure": round(random.uniform(500000, 5000000), 2),
            "mitigation_progress_pct": round(random.uniform(20, 80), 2),
        })
    return items

def main():
    data = {
        "generated_at": date.today().isoformat(), "summary": {},
        "categories": generate_categories(), "registers": generate_registers(),
        "scenarios": generate_scenarios(), "mitigations": generate_mitigations(),
        "escalations": generate_escalations(), "monte_carlo_runs": generate_mc_runs(),
        "dashboard": generate_dashboard(),
    }
    for k in ["categories","registers","scenarios","mitigations","escalations","monte_carlo_runs","dashboard"]:
        data["summary"][k] = len(data[k])
        print(f"[RM] Generated {len(data[k])} {k}")

    OUTPUT.parent.mkdir(parents=True, exist_ok=True)
    with open(OUTPUT, "w", encoding="utf-8") as f:
        json.dump(data, f, indent=2, ensure_ascii=False)
    print(f"[RM] Output: {OUTPUT}")

if __name__ == "__main__":
    main()