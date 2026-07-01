#!/usr/bin/env python3
"""
OpenConstructionERP — Schedule Management Data Generator
Generates test data: Schedules, Activities, Relationships, Resources,
Baselines, Changes, Critical Path Log
"""
import json
import random
from datetime import datetime, date, timedelta
from pathlib import Path

random.seed(42)

OUTPUT = Path(__file__).parent.parent / "apps" / "web" / "schedule_data.json"

PROJECTS = [
    {"id": 1, "name": "Metro Line 3 — Central Corridor"},
    {"id": 2, "name": "Highway Bypass — Northern Ring Road"},
    {"id": 3, "name": "Water Treatment Plant — Phase II"},
    {"id": 4, "name": "Railway Bridge — River Delta Crossing"},
    {"id": 5, "name": "Port Expansion — Container Terminal"},
]

SCHEDULE_TYPES = ["current", "baseline", "target", "what_if"]
ACTIVITY_TYPES = ["task", "task", "task", "milestone", "start_milestone", "finish_milestone", "level_of_effort", "wbs_summary"]
CONSTRAINT_TYPES = ["as_late_as_possible", "start_no_earlier", "start_no_later", "finish_no_earlier", "mandatory_start", "mandatory_finish"]
RELATION_TYPES = ["FS", "FS", "FS", "SS", "FF", "SF"]
RESOURCE_TYPES = ["labor", "labor", "material", "equipment", "cost", "role"]
CHANGE_TYPES = ["scope", "duration", "relationship", "resource", "constraint"]

def rand_date(start_year=2024, end_year=2026):
    start = date(start_year, 1, 1)
    end = date(end_year, 6, 30)
    return start + timedelta(days=random.randint(0, (end - start).days))

def generate_schedules():
    items = []
    for proj in PROJECTS:
        n = random.randint(2, 4)
        for i in range(1, n + 1):
            stype = random.choice(SCHEDULE_TYPES)
            created = rand_date(2024, 2025)
            items.append({
                "id": f"sched-{proj['id']}-{i:04d}",
                "project_id": proj["id"],
                "project_name": proj["name"],
                "schedule_code": f"SC-{proj['id']:04d}-{i:04d}",
                "schedule_name": f"{proj['name']} — {stype.title()} Schedule v{i}",
                "schedule_type": stype,
                "calendar": random.choice(["5_day_week", "6_day_week", "7_day_week", "4x10"]),
                "data_date": (created + timedelta(days=random.randint(1, 30))).isoformat(),
                "status": random.choice(["active", "active", "active", "baselined", "closed"]),
                "total_float_pct": round(random.uniform(0, 15), 2),
                "created_by": random.choice(["Scheduler_A", "Scheduler_B", "PMO_Team"]),
                "created_at": created.isoformat(),
            })
    return items

def generate_activities(schedules):
    items = []
    activity_names = [
        "Site Mobilization", "Survey & Setting Out", "Excavation", "Piling Works",
        "Foundation Concreting", "Formwork Installation", "Reinforcement Steel",
        "Concrete Pouring", "Curing & Stripping", "Backfilling & Compaction",
        "Steel Structure Erection", "Bolt Tightening & Torque Test", "MEP Rough-in",
        "Electrical Conduit Installation", "HVAC Ductwork", "Plumbing Installation",
        "Fire Protection System", "Cable Tray & Ladder", "Instrumentation & Control",
        "Floor Slab Concreting", "Wall Construction", "Waterproofing", "Roofing",
        "Ceiling Installation", "Painting & Finishing", "Joinery & Cabinets",
        "Doors & Windows Installation", "Tiling Works", "External Works", "Landscaping",
        "Road Works", "Fencing & Security", "Testing & Commissioning",
        "Handover Preparation", "Project Closeout"
    ]
    act_id = 1
    for sched in schedules:
        n = random.randint(20, 35)
        sched_start = date(2025, 1, 1) + timedelta(days=random.randint(0, 60))
        current_date = sched_start
        for i in range(n):
            name = random.choice(activity_names) + f" ({sched['project_name'][:10]}...)"
            orig_dur = random.randint(2, 25)
            pct = random.choice([0, 0, 0, 25, 50, 75, 100, 100])
            es = current_date
            ef = es + timedelta(days=orig_dur - 1)
            is_critical = random.random() < 0.2
            float_total = 0 if is_critical else random.randint(1, 15)
            items.append({
                "id": f"act-{act_id:06d}",
                "schedule_id": sched["id"],
                "project_id": sched["project_id"],
                "project_name": sched["project_name"],
                "activity_id": f"A{act_id:06d}",
                "wbs_code": f"WBS-{random.randint(1,5)}.{random.randint(1,15)}",
                "activity_name": name[:200],
                "activity_type": random.choice(ACTIVITY_TYPES),
                "status": "completed" if pct == 100 else ("in_progress" if pct > 0 else "not_started"),
                "original_duration": orig_dur,
                "remaining_duration": max(0, orig_dur - int(orig_dur * pct / 100)),
                "actual_duration": int(orig_dur * pct / 100),
                "percent_complete": float(pct),
                "early_start": es.isoformat(),
                "early_finish": ef.isoformat(),
                "late_start": (es + timedelta(days=float_total)).isoformat(),
                "late_finish": (ef + timedelta(days=float_total)).isoformat(),
                "actual_start": es.isoformat() if pct > 0 else None,
                "actual_finish": ef.isoformat() if pct == 100 else None,
                "start_date": es.isoformat(),
                "finish_date": ef.isoformat(),
                "float_free": random.randint(0, float_total),
                "float_total": float_total,
                "is_critical": is_critical,
                "is_driving": is_critical and random.random() < 0.3,
                "constraint_type": random.choice(CONSTRAINT_TYPES),
                "constraint_date": ef.isoformat() if random.random() < 0.2 else None,
            })
            current_date = ef + timedelta(days=random.randint(0, 3))
            act_id += 1
    return items

def generate_relationships(schedules, activities):
    items = []
    rel_id = 1
    for sched in schedules:
        sched_acts = [a for a in activities if a["schedule_id"] == sched["id"]]
        for i in range(len(sched_acts) - 1):
            if random.random() < 0.7:
                items.append({
                    "id": f"rel-{rel_id:06d}",
                    "schedule_id": sched["id"],
                    "project_id": sched["project_id"],
                    "predecessor_id": sched_acts[i]["id"],
                    "successor_id": sched_acts[i + 1]["id"],
                    "predecessor_name": sched_acts[i]["activity_name"][:50],
                    "successor_name": sched_acts[i+1]["activity_name"][:50],
                    "relation_type": random.choice(RELATION_TYPES),
                    "lag_days": random.randint(0, 3),
                })
                rel_id += 1
        # Extra cross-path relationships
        for _ in range(random.randint(1, 5)):
            a1, a2 = random.sample(sched_acts, 2)
            items.append({
                "id": f"rel-{rel_id:06d}",
                "schedule_id": sched["id"],
                "project_id": sched["project_id"],
                "predecessor_id": a1["id"],
                "successor_id": a2["id"],
                "predecessor_name": a1["activity_name"][:50],
                "successor_name": a2["activity_name"][:50],
                "relation_type": random.choice(["FS", "SS"]),
                "lag_days": random.randint(-2, 5),
            })
            rel_id += 1
    return items

material_names = ["Cement", "Steel Rebar 16mm", "Aggregate 20mm", "Sand", "Timber", "Plywood", "PVC Pipe 110mm", "Cable 4x16mm2", "Paint", "Tile 60x60"]

def generate_resources(schedules, activities):
    items = []
    res_id = 1
    labor_names = ["Excavator Operator", "Crane Operator", "Concrete Worker", "Steel Fixer", "Electrician", "Plumber", "Welder", "Rigger", "General Laborer", "Foreman"]
    equip_names = ["Excavator CAT 320", "Crane Liebherr 100t", "Concrete Pump", "Bulldozer D6", "Roller", "Dump Truck", "Forklift", "Generator 50kVA", "Welding Machine", "Compressor"]
    for sched in schedules:
        sched_acts = [a for a in activities if a["schedule_id"] == sched["id"]]
        for act in sched_acts[:random.randint(10, len(sched_acts))]:
            rtype = random.choice(RESOURCE_TYPES)
            if rtype == "labor":
                name = random.choice(labor_names)
            elif rtype == "equipment":
                name = random.choice(equip_names)
            elif rtype == "material":
                name = random.choice(material_names)
            else:
                name = random.choice(["Project Engineer", "Supervisor"])
            units = round(random.uniform(1, 10), 2)
            cost = round(random.uniform(20, 150), 2)
            items.append({
                "id": f"res-{res_id:06d}",
                "schedule_id": sched["id"],
                "project_id": sched["project_id"],
                "activity_id": act["id"],
                "activity_name": act["activity_name"][:50],
                "resource_type": rtype,
                "resource_code": f"R{rtype[0].upper()}-{res_id:04d}",
                "resource_name": name,
                "units_per_day": units,
                "total_units": round(units * act["original_duration"], 2),
                "unit_cost": cost,
                "total_cost": round(cost * units * act["original_duration"], 2),
                "bid_price": round(cost * units * act["original_duration"] * random.uniform(0.9, 1.1), 2),
                "actual_units": round(units * act["actual_duration"], 2),
                "actual_cost": round(cost * units * act["actual_duration"], 2),
            })
            res_id += 1
    return items

def generate_baselines(schedules):
    items = []
    for sched in schedules:
        n = random.randint(1, 3)
        for i in range(1, n + 1):
            items.append({
                "id": f"bl-{sched['id']}-{i}",
                "schedule_id": sched["id"],
                "project_id": sched["project_id"],
                "baseline_number": i,
                "baseline_name": f"Baseline {i}",
                "baseline_date": rand_date(2024, 2025).isoformat(),
                "is_current": i == n,
                "total_float_pct": round(random.uniform(0, 10), 2),
                "created_by": random.choice(["Planner_A", "Planner_B"]),
                "created_at": datetime.utcnow().isoformat(),
            })
    return items

def generate_changes(schedules, activities):
    items = []
    for sched in schedules:
        n = random.randint(2, 6)
        sched_acts = [a for a in activities if a["schedule_id"] == sched["id"]]
        for i in range(1, n + 1):
            related_act = random.choice(sched_acts) if sched_acts else None
            items.append({
                "id": f"ch-{sched['id']}-{i:04d}",
                "schedule_id": sched["id"],
                "project_id": sched["project_id"],
                "change_number": i,
                "change_code": f"SCHC-{sched['project_id']:04d}-{i:04d}",
                "change_type": random.choice(CHANGE_TYPES),
                "description": random.choice([
                    "Additional scope for foundation works",
                    "Duration extension due to weather",
                    "Revised logic between excavation and piling",
                    "Resource reallocation from Tunneling to Structures",
                    "Constraint added for milestone delivery",
                    "Calendar change from 5-day to 6-day week",
                ]),
                "reason": random.choice(["Client request", "Site condition", "Design change", "Supplier delay", "Weather impact"]),
                "impact_days": random.randint(-10, 30),
                "impact_cost": round(random.uniform(0, 500000), 2),
                "activity_id": related_act["id"] if related_act else None,
                "baseline_id": None,
                "approved_by": random.choice(["PM", "Client Rep", "Construction Manager"]),
                "status": random.choice(["proposed", "reviewed", "approved", "rejected", "implemented"]),
                "proposed_at": rand_date(2024, 2025).isoformat(),
                "approved_at": rand_date(2024, 2026).isoformat(),
            })
    return items

def generate_critical_path_log(schedules, activities):
    items = []
    run_id = 1
    for sched in schedules:
        sched_acts = [a for a in activities if a["schedule_id"] == sched["id"]]
        for run in range(1, random.randint(2, 5)):
            critical_acts = [a for a in sched_acts if a["is_critical"]]
            items.append({
                "id": f"cpl-{run_id:06d}",
                "schedule_id": sched["id"],
                "project_id": sched["project_id"],
                "run_number": run,
                "run_at": rand_date(2025, 2026).isoformat(),
                "total_activities": len(sched_acts),
                "critical_count": len(critical_acts),
                "longest_path": max((a["original_duration"] for a in sched_acts), default=0),
                "total_float_min": min((a["float_total"] for a in sched_acts), default=0),
                "total_float_max": max((a["float_total"] for a in sched_acts), default=0),
                "total_float_avg": round(sum(a["float_total"] for a in sched_acts) / max(len(sched_acts), 1), 2),
                "critical_path": json.dumps([a["id"] for a in critical_acts[:5]]),
                "duration": random.randint(50, 500),
                "status": "completed",
            })
            run_id += 1
    return items

def main():
    schedules = generate_schedules()
    activities = generate_activities(schedules)
    relationships = generate_relationships(schedules, activities)
    resources = generate_resources(schedules, activities)
    baselines = generate_baselines(schedules)
    changes = generate_changes(schedules, activities)
    critical_path_log = generate_critical_path_log(schedules, activities)

    data = {
        "generated_at": datetime.utcnow().isoformat(),
        "summary": {
            "schedules": len(schedules),
            "activities": len(activities),
            "relationships": len(relationships),
            "resources": len(resources),
            "baselines": len(baselines),
            "changes": len(changes),
            "critical_path_logs": len(critical_path_log),
        },
        "schedules": schedules,
        "activities": activities,
        "relationships": relationships,
        "resources": resources,
        "baselines": baselines,
        "changes": changes,
        "critical_path_logs": critical_path_log,
    }

    OUTPUT.parent.mkdir(parents=True, exist_ok=True)
    with open(OUTPUT, "w", encoding="utf-8") as f:
        json.dump(data, f, indent=2, ensure_ascii=False)

    print(f"[Schedule] Generated {data['summary']['schedules']} schedules")
    print(f"[Schedule] Generated {data['summary']['activities']} activities")
    print(f"[Schedule] Generated {data['summary']['relationships']} relationships")
    print(f"[Schedule] Generated {data['summary']['resources']} resources")
    print(f"[Schedule] Generated {data['summary']['baselines']} baselines")
    print(f"[Schedule] Generated {data['summary']['changes']} changes")
    print(f"[Schedule] Generated {data['summary']['critical_path_logs']} CPM runs")
    print(f"[Schedule] Output: {OUTPUT}")

if __name__ == "__main__":
    main()