#!/usr/bin/env python3
"""
OpenConstructionERP — Project Management Data Generator
Generates test data: projects, WBS, milestones, phases, team, portfolios, risks, changes, lessons
"""
import json
import random
from datetime import datetime, date, timedelta
from pathlib import Path

random.seed(42)

OUTPUT = Path(__file__).parent.parent / "apps" / "web" / "pm_data.json"

PROJECT_TYPES = ["metro", "tunnel", "bridge", "road", "building", "industrial", "water"]
STATUSES = ["lead", "tender", "planning", "design", "construction", "commissioning", "operation", "closed"]
PHASES = ["feasibility", "design", "tender", "construction", "handover"]
RISK_CATEGORIES = ["technical", "financial", "schedule", "legal", "environmental", "HSE", "political"]
RISK_PROB = ["very_low", "low", "medium", "high", "very_high"]
RISK_IMPACT = ["very_low", "low", "medium", "high", "very_high"]
MILESTONE_TYPES = ["start", "finish", "payment", "approval", "delivery", "permit", "review"]
CHANGE_TYPES = ["variation", "change_order", "scope_change", "design_change"]
LESSON_CATEGORIES = ["technical", "management", "financial", "HSE", "quality"]

PROJECTS_DATA = [
    {
        "code": "P-2026-001",
        "name": "Metro Line 3 — Central Corridor",
        "project_type": "metro",
        "status": "construction",
        "phase": "construction",
        "country": "Egypt",
        "city": "Cairo",
        "region": "Greater Cairo",
        "start_date": "2024-01-15",
        "end_date": "2028-06-30",
        "duration_days": 1627,
        "budget_total": 2_850_000_000,
        "budget_currency": "USD",
        "contingency": 285_000_000,
        "contingency_pct": 10.0,
        "total_length_km": 42.5,
        "risk_class": "A",
        "complexity": "mega",
        "notes": "Major metro expansion connecting 3 districts",
    },
    {
        "code": "P-2026-002",
        "name": "Highway Bypass — Northern Ring Road",
        "project_type": "road",
        "status": "design",
        "phase": "design",
        "country": "UAE",
        "city": "Dubai",
        "region": "Dubai North",
        "start_date": "2025-03-01",
        "end_date": "2027-09-30",
        "duration_days": 943,
        "budget_total": 420_000_000,
        "budget_currency": "USD",
        "contingency": 42_000_000,
        "contingency_pct": 10.0,
        "total_length_km": 28.0,
        "risk_class": "B",
        "complexity": "high",
        "notes": "6-lane highway bypass with 12 interchanges",
    },
    {
        "code": "P-2026-003",
        "name": "Water Treatment Plant — Phase II",
        "project_type": "water",
        "status": "tender",
        "phase": "tender",
        "country": "Saudi Arabia",
        "city": "Jeddah",
        "region": "Western Region",
        "start_date": "2025-06-01",
        "end_date": "2027-12-31",
        "duration_days": 943,
        "budget_total": 180_000_000,
        "budget_currency": "USD",
        "contingency": 18_000_000,
        "contingency_pct": 10.0,
        "total_area_m2": 45_000,
        "risk_class": "B",
        "complexity": "medium",
        "notes": "Expansion to 200,000 m³/day capacity",
    },
    {
        "code": "P-2026-004",
        "name": "Railway Bridge — River Delta Crossing",
        "project_type": "bridge",
        "status": "planning",
        "phase": "feasibility",
        "country": "Bangladesh",
        "city": "Dhaka",
        "region": "Delta Region",
        "start_date": "2025-09-01",
        "end_date": "2029-03-31",
        "duration_days": 1307,
        "budget_total": 890_000_000,
        "budget_currency": "USD",
        "contingency": 133_500_000,
        "contingency_pct": 15.0,
        "total_length_km": 4.8,
        "risk_class": "A",
        "complexity": "high",
        "notes": "Combined rail-road bridge, 4.8 km main span",
    },
    {
        "code": "P-2026-005",
        "name": "Industrial Park — Logistics Hub",
        "project_type": "industrial",
        "status": "lead",
        "phase": "feasibility",
        "country": "Kenya",
        "city": "Nairobi",
        "region": "Eastern Region",
        "start_date": "2026-01-01",
        "end_date": "2028-06-30",
        "duration_days": 911,
        "budget_total": 350_000_000,
        "budget_currency": "USD",
        "contingency": 35_000_000,
        "contingency_pct": 10.0,
        "total_area_m2": 120_000,
        "risk_class": "C",
        "complexity": "medium",
        "notes": "Integrated logistics hub with warehousing and rail siding",
    },
]

WBS_TEMPLATES = {
    "metro": [
        {"code": "1", "name": "Preliminary Works", "level": 1, "children": [
            {"code": "1.1", "name": "Site Investigation", "level": 2, "children": [
                {"code": "1.1.1", "name": "Geotechnical Survey", "level": 3},
                {"code": "1.1.2", "name": "Topographic Survey", "level": 3},
            ]},
            {"code": "1.2", "name": "Utility Relocation", "level": 2, "children": [
                {"code": "1.2.1", "name": "Water Mains", "level": 3},
                {"code": "1.2.2", "name": "Power Cables", "level": 3},
            ]},
        ]},
        {"code": "2", "name": "Civil Works", "level": 1, "children": [
            {"code": "2.1", "name": "Tunnel Excavation", "level": 2, "children": [
                {"code": "2.1.1", "name": "TBM Drive — Eastbound", "level": 3},
                {"code": "2.1.2", "name": "TBM Drive — Westbound", "level": 3},
            ]},
            {"code": "2.2", "name": "Station Construction", "level": 2, "children": [
                {"code": "2.2.1", "name": "Station A — Central", "level": 3},
                {"code": "2.2.2", "name": "Station B — North", "level": 3},
                {"code": "2.2.3", "name": "Station C — South", "level": 3},
            ]},
        ]},
        {"code": "3", "name": "Track Works", "level": 1, "children": [
            {"code": "3.1", "name": "Ballast & Sleepers", "level": 2},
            {"code": "3.2", "name": "Rail Welding", "level": 2},
        ]},
        {"code": "4", "name": "Systems", "level": 1, "children": [
            {"code": "4.1", "name": "Signalling", "level": 2},
            {"code": "4.2", "name": "Power Supply", "level": 2},
            {"code": "4.3", "name": "Telecoms", "level": 2},
        ]},
        {"code": "5", "name": "Rolling Stock", "level": 1, "children": [
            {"code": "5.1", "name": "Train Procurement", "level": 2},
            {"code": "5.2", "name": "Depot Equipment", "level": 2},
        ]},
    ],
    "road": [
        {"code": "1", "name": "Preliminary", "level": 1, "children": [
            {"code": "1.1", "name": "Survey & Design", "level": 2},
            {"code": "1.2", "name": "Land Acquisition", "level": 2},
        ]},
        {"code": "2", "name": "Earthworks", "level": 1, "children": [
            {"code": "2.1", "name": "Cut & Fill", "level": 2},
            {"code": "2.2", "name": "Subgrade Preparation", "level": 2},
        ]},
        {"code": "3", "name": "Pavement", "level": 1, "children": [
            {"code": "3.1", "name": "Base Course", "level": 2},
            {"code": "3.2", "name": "Asphalt Laying", "level": 2},
        ]},
        {"code": "4", "name": "Structures", "level": 1, "children": [
            {"code": "4.1", "name": "Interchanges", "level": 2},
            {"code": "4.2", "name": "Bridges", "level": 2},
        ]},
        {"code": "5", "name": "Ancillary", "level": 1, "children": [
            {"code": "5.1", "name": "Lighting", "level": 2},
            {"code": "5.2", "name": "Signage", "level": 2},
        ]},
    ],
    "water": [
        {"code": "1", "name": "Design & Engineering", "level": 1},
        {"code": "2", "name": "Civil Works", "level": 1, "children": [
            {"code": "2.1", "name": "Excavation", "level": 2},
            {"code": "2.2", "name": "Concrete Structures", "level": 2},
        ]},
        {"code": "3", "name": "Process Equipment", "level": 1, "children": [
            {"code": "3.1", "name": "Filtration System", "level": 2},
            {"code": "3.2", "name": "Chemical Dosing", "level": 2},
        ]},
        {"code": "4", "name": "Piping & Valves", "level": 1},
        {"code": "5", "name": "Electrical & SCADA", "level": 1},
        {"code": "6", "name": "Commissioning", "level": 1},
    ],
    "bridge": [
        {"code": "1", "name": "Design", "level": 1},
        {"code": "2", "name": "Foundations", "level": 1, "children": [
            {"code": "2.1", "name": "Pile Driving", "level": 2},
            {"code": "2.2", "name": "Pile Caps", "level": 2},
        ]},
        {"code": "3", "name": "Substructure", "level": 1, "children": [
            {"code": "3.1", "name": "Pier Construction", "level": 2},
            {"code": "3.2", "name": "Abutments", "level": 2},
        ]},
        {"code": "4", "name": "Superstructure", "level": 1, "children": [
            {"code": "4.1", "name": "Girder Erection", "level": 2},
            {"code": "4.2", "name": "Deck Slab", "level": 2},
        ]},
        {"code": "5", "name": "Finishing", "level": 1},
    ],
    "industrial": [
        {"code": "1", "name": "Site Preparation", "level": 1},
        {"code": "2", "name": "Foundations & Slabs", "level": 1},
        {"code": "3", "name": "Structural Steel", "level": 1},
        {"code": "4", "name": "Building Envelope", "level": 1},
        {"code": "5", "name": "MEP Services", "level": 1},
        {"code": "6", "name": "Interior Fit-out", "level": 1},
        {"code": "7", "name": "External Works", "level": 1},
    ],
}

def flatten_wbs(items, project_id, parent_id=None):
    """Flatten WBS tree into list of dicts"""
    result = []
    for item in items:
        wbs_id = f"wbs-{project_id}-{item['code'].replace('.', '-')}"
        entry = {
            "id": wbs_id,
            "project_id": project_id,
            "parent_id": parent_id,
            "wbs_code": item["code"],
            "name": item["name"],
            "wbs_level": item["level"],
            "sort_order": int(item["code"].split(".")[-1]) if "." in item["code"] else int(item["code"]),
            "is_leaf": "children" not in item,
            "planned_start": None,
            "planned_end": None,
            "planned_cost": round(random.uniform(500_000, 50_000_000), 2),
            "planned_hours": round(random.uniform(1000, 100_000), 2),
            "progress_pct": random.randint(0, 100) if random.random() > 0.3 else 0,
            "status": random.choice(["planned", "in_progress", "completed", "delayed"]),
        }
        result.append(entry)
        if "children" in item:
            result.extend(flatten_wbs(item["children"], project_id, wbs_id))
    return result

def generate_milestones(project_id, start_date, end_date, count=8):
    milestones = []
    start = datetime.strptime(start_date, "%Y-%m-%d")
    end = datetime.strptime(end_date, "%Y-%m-%d")
    total_days = (end - start).days

    milestone_names = [
        ("M-001", "Project Kick-off", "start"),
        ("M-002", "Design Complete", "finish"),
        ("M-003", "Permit Approval", "approval"),
        ("M-004", "Foundation Complete", "finish"),
        ("M-005", "Structural Topping Out", "finish"),
        ("M-006", "Systems Commissioning", "start"),
        ("M-007", "Substantial Completion", "finish"),
        ("M-008", "Final Handover", "delivery"),
    ]

    for i in range(min(count, len(milestone_names))):
        code, name, mtype = milestone_names[i]
        offset = int(total_days * (i + 1) / (count + 1))
        planned = start + timedelta(days=offset)
        actual = planned + timedelta(days=random.randint(-15, 30)) if random.random() > 0.4 else None
        status = "achieved" if actual and actual <= planned + timedelta(days=15) else random.choice(["planned", "achieved", "delayed", "at_risk"])

        milestones.append({
            "id": f"ms-{project_id}-{code.lower()}",
            "project_id": project_id,
            "milestone_code": code,
            "name": name,
            "milestone_type": mtype,
            "planned_date": planned.strftime("%Y-%m-%d"),
            "actual_date": actual.strftime("%Y-%m-%d") if actual else None,
            "status": status,
            "weight_pct": round(100.0 / count, 2),
            "is_gate": mtype in ("approval", "delivery"),
        })
    return milestones

def generate_phases(project_id, start_date, end_date):
    phases = []
    phase_names = [
        ("PH-01", "Feasibility & Planning", "feasibility"),
        ("PH-02", "Design & Engineering", "design"),
        ("PH-03", "Procurement & Tendering", "tender"),
        ("PH-04", "Construction", "construction"),
        ("PH-05", "Commissioning & Handover", "handover"),
    ]
    start = datetime.strptime(start_date, "%Y-%m-%d")
    end = datetime.strptime(end_date, "%Y-%m-%d")
    total_days = (end - start).days

    for i, (code, name, ptype) in enumerate(phase_names):
        p_start = start + timedelta(days=int(total_days * i / len(phase_names)))
        p_end = start + timedelta(days=int(total_days * (i + 1) / len(phase_names)))
        status = "completed" if i < 2 else random.choice(["pending", "active", "completed", "delayed"])
        phases.append({
            "id": f"ph-{project_id}-{code.lower()}",
            "project_id": project_id,
            "phase_code": code,
            "name": name,
            "sort_order": i + 1,
            "planned_start": p_start.strftime("%Y-%m-%d"),
            "planned_end": p_end.strftime("%Y-%m-%d"),
            "actual_start": p_start.strftime("%Y-%m-%d") if status == "completed" else None,
            "actual_end": p_end.strftime("%Y-%m-%d") if status == "completed" else None,
            "budget_amount": round(random.uniform(10_000_000, 500_000_000), 2),
            "status": status,
            "completion_pct": 100 if status == "completed" else random.randint(0, 80),
        })
    return phases

def generate_risks(project_id, count=6):
    risks = []
    risk_names = [
        ("R-001", "Geotechnical uncertainty", "technical"),
        ("R-002", "Budget overrun due to inflation", "financial"),
        ("R-003", "Permit delays from authorities", "legal"),
        ("R-004", "Supply chain disruption", "schedule"),
        ("R-005", "Environmental impact concerns", "environmental"),
        ("R-006", "Labour shortage", "HSE"),
        ("R-007", "Political instability", "political"),
        ("R-008", "Design changes from client", "technical"),
    ]
    for i in range(min(count, len(risk_names))):
        code, name, cat = risk_names[i]
        prob = random.choice(RISK_PROB)
        impact = random.choice(RISK_IMPACT)
        prob_score = RISK_PROB.index(prob) + 1
        impact_score = RISK_IMPACT.index(impact) + 1
        risks.append({
            "id": f"risk-{project_id}-{code.lower()}",
            "project_id": project_id,
            "risk_code": code,
            "name": name,
            "risk_category": cat,
            "risk_type": "threat",
            "probability": prob,
            "impact": impact,
            "probability_score": prob_score,
            "impact_score": impact_score,
            "risk_score": prob_score * impact_score,
            "potential_cost": round(random.uniform(100_000, 50_000_000), 2),
            "mitigation_strategy": random.choice(["avoid", "transfer", "mitigate", "accept"]),
            "status": random.choice(["identified", "assessed", "mitigation_planned", "mitigation_in_progress", "closed"]),
        })
    return risks

def generate_changes(project_id, count=3):
    changes = []
    for i in range(count):
        changes.append({
            "id": f"ch-{project_id}-{i+1:03d}",
            "project_id": project_id,
            "change_number": f"CO-{i+1:04d}",
            "change_type": random.choice(CHANGE_TYPES),
            "source": random.choice(["client", "contractor", "design", "regulatory", "unforeseen"]),
            "description": f"Change order #{i+1} — {random.choice(['additional scope', 'material substitution', 'schedule adjustment', 'design revision'])}",
            "cost_impact": round(random.uniform(-500_000, 2_000_000), 2),
            "schedule_impact": random.randint(-30, 90),
            "status": random.choice(["submitted", "review", "approved", "rejected", "implemented"]),
        })
    return changes

def generate_lessons(project_id, count=3):
    lessons = []
    for i in range(count):
        is_positive = random.random() > 0.5
        lessons.append({
            "id": f"ll-{project_id}-{i+1:03d}",
            "project_id": project_id,
            "category": random.choice(LESSON_CATEGORIES),
            "title": f"{'Success' if is_positive else 'Lesson'}: {random.choice(['Early engagement with stakeholders', 'BIM coordination process', 'Material testing protocol', 'Safety induction program', 'Quality control procedure'])}",
            "description": "Detailed description of the lesson learned from project execution.",
            "is_positive": is_positive,
            "severity": random.choice(["low", "medium", "high"]),
            "status": random.choice(["draft", "reviewed", "published"]),
        })
    return lessons

def generate_team(project_id, count=6):
    roles = [
        ("PM-001", "Project Manager", "management", True),
        ("PM-002", "Construction Manager", "management", True),
        ("PM-003", "Senior Engineer", "engineering", True),
        ("PM-004", "QA/QC Manager", "supervision", False),
        ("PM-005", "Safety Officer", "support", False),
        ("PM-006", "Contract Administrator", "admin", False),
        ("PM-007", "Site Supervisor", "supervision", False),
        ("PM-008", "Cost Controller", "admin", True),
    ]
    team = []
    for i in range(min(count, len(roles))):
        code, role, cat, is_key = roles[i]
        team.append({
            "id": f"team-{project_id}-{code.lower()}",
            "project_id": project_id,
            "employee_id": f"EMP-{random.randint(100, 999)}",
            "role": role,
            "role_category": cat,
            "allocation_pct": random.choice([50, 75, 100, 100]),
            "is_key": is_key,
        })
    return team

def main():
    output = {
        "generated_at": datetime.utcnow().isoformat() + "Z",
        "projects": [],
        "wbs_items": [],
        "milestones": [],
        "phases": [],
        "team": [],
        "risks": [],
        "changes": [],
        "lessons": [],
        "portfolios": [
            {
                "id": "pf-001",
                "code": "PF-INFRA-2026",
                "name": "Infrastructure Program 2026",
                "portfolio_type": "program",
                "status": "active",
                "budget_total": 4_690_000_000,
                "budget_currency": "USD",
            },
            {
                "id": "pf-002",
                "code": "PF-TRANSPORT",
                "name": "Transport & Mobility Portfolio",
                "portfolio_type": "portfolio",
                "parent_id": "pf-001",
                "status": "active",
                "budget_total": 4_160_000_000,
                "budget_currency": "USD",
            },
            {
                "id": "pf-003",
                "code": "PF-UTILITIES",
                "name": "Water & Utilities Portfolio",
                "portfolio_type": "portfolio",
                "parent_id": "pf-001",
                "status": "active",
                "budget_total": 530_000_000,
                "budget_currency": "USD",
            },
        ],
        "portfolio_projects": [
            {"portfolio_id": "pf-002", "project_id": "P-2026-001", "sort_order": 1},
            {"portfolio_id": "pf-002", "project_id": "P-2026-002", "sort_order": 2},
            {"portfolio_id": "pf-002", "project_id": "P-2026-004", "sort_order": 3},
            {"portfolio_id": "pf-003", "project_id": "P-2026-003", "sort_order": 1},
            {"portfolio_id": "pf-001", "project_id": "P-2026-005", "sort_order": 4},
        ],
    }

    for proj in PROJECTS_DATA:
        pid = proj["code"]
        output["projects"].append(proj)

        # WBS
        template = WBS_TEMPLATES.get(proj["project_type"], WBS_TEMPLATES["industrial"])
        wbs_items = flatten_wbs(template, pid)
        output["wbs_items"].extend(wbs_items)

        # Milestones
        output["milestones"].extend(generate_milestones(pid, proj["start_date"], proj["end_date"]))

        # Phases
        output["phases"].extend(generate_phases(pid, proj["start_date"], proj["end_date"]))

        # Team
        output["team"].extend(generate_team(pid))

        # Risks
        output["risks"].extend(generate_risks(pid))

        # Changes
        output["changes"].extend(generate_changes(pid))

        # Lessons
        output["lessons"].extend(generate_lessons(pid))

    OUTPUT.parent.mkdir(parents=True, exist_ok=True)
    with open(OUTPUT, "w", encoding="utf-8") as f:
        json.dump(output, f, indent=2, ensure_ascii=False)

    print(f"✅ PM data generated: {OUTPUT}")
    print(f"   Projects: {len(output['projects'])}")
    print(f"   WBS items: {len(output['wbs_items'])}")
    print(f"   Milestones: {len(output['milestones'])}")
    print(f"   Phases: {len(output['phases'])}")
    print(f"   Team members: {len(output['team'])}")
    print(f"   Risks: {len(output['risks'])}")
    print(f"   Changes: {len(output['changes'])}")
    print(f"   Lessons: {len(output['lessons'])}")
    print(f"   Portfolios: {len(output['portfolios'])}")

if __name__ == "__main__":
    main()
