#!/usr/bin/env python3
"""
OpenConstructionERP — Change Management Data Generator
Generates test data: Change Requests, Change Orders, Impact Analysis, Approval Workflow, Change Log
"""
import json, random
from datetime import date, timedelta
from pathlib import Path

random.seed(42)
OUTPUT = Path(__file__).parent.parent / "apps" / "web" / "change_data.json"

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

def generate_change_requests():
    items = []
    statuses = ["draft","submitted","under_review","approved","rejected","deferred","implemented","closed","cancelled"]
    for proj in PROJECTS:
        n = random.randint(8, 15)
        for i in range(1, n+1):
            s = pick(*statuses)
            items.append({
                "id": f"cr-{proj['id']}-{i:04d}",
                "project_id": proj["id"], "project_name": proj["name"],
                "cr_number": i, "cr_code": f"CR-{i:04d}",
                "cr_name": f"{pick(['Scope change','Design revision','Material substitution','Schedule adjustment','Contract variation','Specification update'])} — CR#{i}",
                "cr_type": pick(["scope","design","specification","schedule","cost","contract","quality"]),
                "source": pick(["owner","contractor","designer","regulatory","internal"]),
                "priority": pick(["low","medium","high","emergency"]),
                "description": f"Detailed description of change request {i} for {proj['name']}",
                "reason": f"{pick(['Owner requirement change','Design error','Site condition','Value engineering','Regulatory compliance','Schedule recovery'])}",
                "proposed_by": pick(["Project Manager","Design Team","Contractor","Owner Representative"]),
                "proposed_date": rand_date().isoformat(),
                "required_by_date": rand_date(2025, 2026).isoformat(),
                "status": s,
            })
    return items

def generate_change_orders():
    items = []
    for proj in PROJECTS:
        n = random.randint(4, 10)
        for i in range(1, n+1):
            cost_change = round(random.uniform(10000, 500000), 2)
            sched_change = random.randint(-30, 60)
            items.append({
                "id": f"co-{proj['id']}-{i:04d}",
                "project_id": proj["id"], "project_name": proj["name"],
                "co_number": i, "co_code": f"CO-{i:04d}",
                "co_name": f"Variation Order #{i} — {pick(['Additional Works','Scope Change','Design Change','Site Condition'])}",
                "co_type": pick(["variation","change_directive","claim_settlement","compensation_event"]),
                "scope_change": f"Scope modification for {pick(['Section A','Pier P05','Tunnel B2','Building Core'])}",
                "cost_change": cost_change,
                "schedule_change_days": sched_change,
                "justification": f"Justification for variation order #{i}",
                "contractor_name": pick(["Contractor A","Contractor B","Subcontractor C","JV Partner"]),
                "approved_by": pick(["Project Director","Client Rep","Contract Manager"]),
                "status": pick(["draft","submitted","under_review","approved","executed","closed"]),
            })
    return items

def generate_impact_analyses():
    items = []
    for proj in PROJECTS:
        n = random.randint(10, 20)
        for i in range(1, n+1):
            items.append({
                "id": f"cia-{proj['id']}-{i:04d}",
                "project_id": proj["id"], "project_name": proj["name"],
                "impact_type": pick(["cost","schedule","scope","quality","safety","environment","stakeholder","resource","contract","risk"]),
                "description": f"Impact analysis for {pick(['cost increase','schedule delay','scope reduction','quality impact'])}",
                "impact_level": pick(["very_low","low","medium","high","very_high"]),
                "cost_impact": round(random.uniform(0, 500000), 2),
                "schedule_impact_days": random.randint(0, 90),
                "analyzed_by": pick(["Change Manager","Project Engineer","Quantity Surveyor"]),
                "analysis_date": rand_date().isoformat(),
                "status": pick(["draft","reviewed","approved","superseded"]),
            })
    return items

def generate_approval_workflow():
    items = []
    steps = ["Initial Review","Technical Assessment","Cost Review","Legal Review","Director Approval","Client Approval"]
    for proj in PROJECTS:
        for i in range(1, min(6, len(steps)+1)):
            items.append({
                "id": f"aw-{proj['id']}-{i:04d}",
                "project_id": proj["id"], "project_name": proj["name"],
                "step_order": i,
                "step_name": steps[i-1] if i <= len(steps) else f"Step {i}",
                "approver_role": pick(["Project Manager","Engineering Manager","Commercial Manager","Legal Advisor","Director","Client Rep"]),
                "status": pick(["pending","approved","rejected"]),
            })
    return items

def generate_change_log():
    items = []
    log_types = ["status_change","comment","document_added","approval","rejection","implementation"]
    for proj in PROJECTS:
        n = random.randint(10, 20)
        for i in range(1, n+1):
            items.append({
                "id": f"cl-{proj['id']}-{i:04d}",
                "project_id": proj["id"], "project_name": proj["name"],
                "log_type": pick(*log_types),
                "previous_status": pick(["draft","submitted","under_review"]),
                "new_status": pick(["submitted","under_review","approved","implemented","closed"]),
                "description": f"Change log entry #{i}: {pick(['Status changed','Document added','Comment added','Approval granted'])}",
                "changed_by": pick(["System","User A","User B","Approver C"]),
                "changed_at": rand_date(2025, 2026).isoformat(),
            })
    return items

def main():
    data = {
        "generated_at": date.today().isoformat(), "summary": {},
        "change_requests": generate_change_requests(),
        "change_orders": generate_change_orders(),
        "impact_analyses": generate_impact_analyses(),
        "approval_workflow": generate_approval_workflow(),
        "change_log": generate_change_log(),
    }
    for k in ["change_requests","change_orders","impact_analyses","approval_workflow","change_log"]:
        data["summary"][k] = len(data[k])
        print(f"[CM] Generated {len(data[k])} {k}")

    OUTPUT.parent.mkdir(parents=True, exist_ok=True)
    with open(OUTPUT, "w", encoding="utf-8") as f:
        json.dump(data, f, indent=2, ensure_ascii=False)
    print(f"[CM] Output: {OUTPUT}")

if __name__ == "__main__":
    main()