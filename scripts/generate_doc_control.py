#!/usr/bin/env python3
"""
OpenConstructionERP — Document Control Data Generator
Generates test data: RFI, NCR, Submittals, Method Statements,
Shop Drawings, Correspondence, Minutes of Meeting, Daily Reports,
Document Transmittals, Document Revisions
"""
import json
import random
from datetime import datetime, date, timedelta
from pathlib import Path

random.seed(42)

OUTPUT = Path(__file__).parent.parent / "apps" / "web" / "doc_control_data.json"

STATUSES_RFI = ["open", "open", "answered", "closed", "closed", "overdue"]
STATUSES_NCR = ["open", "investigating", "action_planned", "action_taken", "verified", "closed", "closed"]
STATUSES_SUB = ["draft", "submitted", "under_review", "reviewed", "approved", "approved_with_comments", "rejected", "resubmit", "closed"]
STATUSES_MS = ["draft", "submitted", "under_review", "approved", "rejected", "revised", "closed"]
STATUSES_SD = ["draft", "submitted", "under_review", "approved", "approved_with_comments", "rejected"]
STATUSES_CORR = ["draft", "sent", "received", "acknowledged", "replied", "archived"]
STATUSES_MOM = ["draft", "distributed", "approved", "closed"]
STATUSES_DR = ["draft", "submitted", "approved", "approved"]
STATUSES_DT = ["prepared", "sent", "received", "acknowledged", "closed"]

DISCIPLINES = ["civil", "structural", "MEP", "geotech", "arch", "landscape", "track", "electrical", "mechanical"]
SEVERITIES = ["minor", "minor", "major", "critical"]
SUB_TYPES = ["material", "equipment", "drawing", "sample", "document", "method", "product_data", "other"]
NCR_TYPES = ["material", "workmanship", "design", "dimensional", "documentation", "safety", "other"]
MEETING_TYPES = ["progress", "technical", "coordination", "hse", "design_review", "kickoff", "closeout"]
CORR_TYPES = ["letter", "email", "memo", "fax", "transmittal", "notice", "instruction"]
PURPOSE_TYPES = ["for_review", "for_approval", "for_construction", "for_record", "for_comment", "for_information"]

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

def generate_rfis():
    items = []
    for proj in PROJECTS:
        n = random.randint(8, 20)
        for i in range(1, n + 1):
            status = random.choice(STATUSES_RFI)
            raised_at = rand_date(2024, 2025)
            due = raised_at + timedelta(days=random.randint(5, 30))
            answered_at = raised_at + timedelta(days=random.randint(1, 20)) if status in ["answered", "closed"] else None
            items.append({
                "id": f"rfi-{proj['id']}-{i:04d}",
                "project_id": proj["id"],
                "project_name": proj["name"],
                "rfi_number": i,
                "rfi_code": f"RFI-{i:04d}",
                "subject": random.choice([
                    "Clarify reinforcement detailing at pier P12",
                    "Confirm concrete grade for tunnel lining",
                    "Request structural load data for foundation",
                    "Clarify drainage pipe routing at chainage 2+500",
                    "Confirm cable tray installation elevation",
                    "Request approved shop drawing for steelwork",
                    "Clarify waterproofing membrane specification",
                    "Confirm anchor bolt locations for equipment",
                ]),
                "question": "Please provide clarification per contract specification section 5.3.",
                "answer": "Refer to drawing SK-2024-012. Reinforcement as per detail A." if answered_at else None,
                "discipline": random.choice(DISCIPLINES),
                "priority": random.choice(["low", "normal", "high", "urgent"]),
                "raised_by": random.choice(["A. Ahmed", "B. Smith", "C. Wang", "D. Kim", "E. Hassan"]),
                "assigned_to": random.choice(["M. Johnson", "P. Garcia", "L. Chen", "R. Patel", "S. Al-Rashid"]),
                "status": status,
                "due_date": due.isoformat(),
                "raised_at": raised_at.isoformat(),
                "answered_at": answered_at.isoformat() if answered_at else None,
                "closed_at": answered_at.isoformat() if answered_at else None,
            })
    return items

def generate_ncrs():
    items = []
    for proj in PROJECTS:
        n = random.randint(4, 12)
        for i in range(1, n + 1):
            status = random.choice(STATUSES_NCR)
            severity = random.choice(SEVERITIES)
            reported_at = rand_date(2024, 2025)
            due = reported_at + timedelta(days=random.randint(7, 45))
            items.append({
                "id": f"ncr-{proj['id']}-{i:04d}",
                "project_id": proj["id"],
                "project_name": proj["name"],
                "ncr_number": i,
                "ncr_code": f"NCR-{i:04d}",
                "title": random.choice([
                    "Concrete compressive strength below spec at pier P05",
                    "Welding defects in steel beam connection",
                    "Reinforcement cover insufficient in wall panel W-12",
                    "Waterproofing membrane damaged during installation",
                    "Cable tray support spacing exceeds specification",
                    "Bolt torque values not achieved in steel frame",
                ]),
                "description": "Non-conformance identified during inspection. Details per attached report.",
                "location": random.choice(["Pier P05", "Section A-B", "Chainage 3+200", "Panel W-12", "Bay 4"]),
                "ncr_type": random.choice(NCR_TYPES),
                "severity": severity,
                "source": random.choice(["inspection", "test", "audit", "complaint"]),
                "reported_by": random.choice(["Inspector A", "QC Team", "Consultant", "Client Rep"]),
                "assigned_to": random.choice(["Contractor QC", "Project Engineer", "Construction Manager"]),
                "root_cause": "Inadequate quality control during placement." if status != "open" else None,
                "corrective_action": "Remove and replace defective work per method statement." if status != "open" else None,
                "preventive_action": "Enhanced inspection protocol implemented." if status != "open" else None,
                "status": status,
                "due_date": due.isoformat(),
                "reported_at": reported_at.isoformat(),
                "closed_at": (reported_at + timedelta(days=random.randint(14, 60))).isoformat() if status == "closed" else None,
            })
    return items

def generate_submittals():
    items = []
    for proj in PROJECTS:
        n = random.randint(10, 25)
        for i in range(1, n + 1):
            status = random.choice(STATUSES_SUB)
            submitted_at = rand_date(2024, 2025)
            approved_at = submitted_at + timedelta(days=random.randint(5, 30)) if status in ["approved", "approved_with_comments", "closed"] else None
            items.append({
                "id": f"sub-{proj['id']}-{i:04d}",
                "project_id": proj["id"],
                "project_name": proj["name"],
                "submittal_number": i,
                "submittal_code": f"SUB-{i:04d}",
                "title": random.choice([
                    "Structural Steel Specification",
                    "HVAC Equipment Data Sheets",
                    "Cable Tray & Ladder Installation Details",
                    "Waterproofing Membrane Product Data",
                    "Fire Protection System Shop Drawings",
                    "Plumbing Fixture Schedule",
                    "Lighting Control System Submittal",
                    "Elevator Equipment Specifications",
                ]),
                "description": "Submitted for review and approval per contract requirements.",
                "submittal_type": random.choice(SUB_TYPES),
                "specification_ref": f"SEC-{random.randint(100, 999)}",
                "submitted_by": random.choice(["Contractor A", "Subcontractor B", "Supplier C"]),
                "submitted_to": random.choice(["Consultant", "Project Manager", "Client"]),
                "status": status,
                "review_notes": random.choice(["Approved as submitted.", "Revise per comments.", "Rejected - resubmit."]) if status in ["reviewed", "approved", "approved_with_comments", "rejected", "resubmit"] else None,
                "resubmit_count": random.randint(0, 3),
                "submitted_at": submitted_at.isoformat(),
                "reviewed_at": (submitted_at + timedelta(days=random.randint(3, 15))).isoformat() if status not in ["draft", "submitted"] else None,
                "approved_at": approved_at.isoformat() if approved_at else None,
            })
    return items

def generate_method_statements():
    items = []
    for proj in PROJECTS:
        n = random.randint(5, 12)
        for i in range(1, n + 1):
            status = random.choice(STATUSES_MS)
            submitted_at = rand_date(2024, 2025)
            items.append({
                "id": f"ms-{proj['id']}-{i:04d}",
                "project_id": proj["id"],
                "project_name": proj["name"],
                "ms_number": i,
                "ms_code": f"MS-{i:04d}",
                "title": random.choice([
                    "Method Statement for Bored Pile Installation",
                    "Method Statement for Steel Erection",
                    "Method Statement for Waterproofing",
                    "Method Statement for Concrete Placement",
                    "Method Statement for Excavation & Shoring",
                    "Method Statement for MEP Installation",
                ]),
                "description": "Detailed work methodology for the specified activity.",
                "work_area": random.choice(["Section A", "Pier Area", "Bridge Deck", "Tunnel Section", "Building Core"]),
                "activity": random.choice(["Piling", "Steelwork", "Concrete", "Excavation", "Waterproofing"]),
                "method": "Sequential methodology per approved construction plan.",
                "resources": "Crane, excavator, concrete pump, crew of 8.",
                "hse_aspects": "Risk assessment, PPE, permit to work required.",
                "submitted_by": random.choice(["Contractor", "Subcontractor"]),
                "reviewed_by": random.choice(["Project Engineer", "HSE Manager", "Consultant"]) if status not in ["draft", "submitted"] else None,
                "status": status,
                "submitted_at": submitted_at.isoformat(),
                "approved_at": (submitted_at + timedelta(days=random.randint(3, 14))).isoformat() if status == "approved" else None,
            })
    return items

def generate_shop_drawings():
    items = []
    for proj in PROJECTS:
        n = random.randint(8, 20)
        for i in range(1, n + 1):
            status = random.choice(STATUSES_SD)
            submitted_at = rand_date(2024, 2025)
            items.append({
                "id": f"sd-{proj['id']}-{i:04d}",
                "project_id": proj["id"],
                "project_name": proj["name"],
                "drawing_number": i,
                "drawing_code": f"SD-{i:04d}",
                "title": random.choice([
                    "Steel Beam Connection Details",
                    "Reinforcement Layout Pier P08",
                    "Cable Tray Routing Plan",
                    "Ductwork Layout Level 2",
                    "Plumbing Riser Diagram",
                    "Electrical Panel Schedule",
                    "Foundation Plan Detail",
                    "Column Reinforcement Detail",
                ]),
                "description": f"Detailed shop drawing for {random.choice(['steelwork', 'concrete', 'MEP', 'civil'])}.",
                "discipline": random.choice(DISCIPLINES),
                "drawing_format": random.choice(["pdf", "dwg", "ifc"]),
                "revision": random.choice(["A", "B", "C", "0", "1", "2"]),
                "file_path": f"drawings/{proj['id']}/SD-{i:04d}_Rev{random.choice(['A','B','C'])}.pdf",
                "submitted_by": random.choice(["Drafting Team", "Subcontractor"]),
                "checked_by": random.choice(["Senior Engineer", "Design Manager"]),
                "status": status,
                "review_notes": "Check dimensions against architectural layout." if status not in ["draft", "submitted"] else None,
                "submitted_at": submitted_at.isoformat(),
                "approved_at": (submitted_at + timedelta(days=random.randint(5, 20))).isoformat() if status in ["approved", "approved_with_comments", "closed"] else None,
            })
    return items

def generate_correspondence():
    items = []
    for proj in PROJECTS:
        n = random.randint(15, 35)
        for i in range(1, n + 1):
            status = random.choice(STATUSES_CORR)
            sent_at = rand_date(2024, 2025)
            direction = random.choice(["incoming", "outgoing", "internal"])
            items.append({
                "id": f"corr-{proj['id']}-{i:04d}",
                "project_id": proj["id"],
                "project_name": proj["name"],
                "corr_number": i,
                "corr_code": f"LTR-{i:04d}",
                "subject": random.choice([
                    "Request for revised foundation drawings",
                    "Notification of site access restrictions",
                    "Confirmation of meeting schedule",
                    "Response to RFI-0042",
                    "Monthly progress report submission",
                    "Instruction to proceed with steelwork",
                    "Notice of non-compliance observation",
                    "Variation order proposal review",
                ]),
                "body": "Reference is made to the contract agreement. We hereby request...",
                "corr_type": random.choice(CORR_TYPES),
                "direction": direction,
                "from_entity": random.choice(["Contractor", "Consultant", "Client", "Subcontractor", "Supplier"]),
                "to_entity": random.choice(["Consultant", "Contractor", "Client", "PMT"]),
                "cc_entity": random.choice(["", "Project Director", "Quality Manager", "Client Rep"]),
                "priority": random.choice(["low", "normal", "high", "urgent"]),
                "status": status,
                "sent_at": sent_at.isoformat(),
                "received_at": (sent_at + timedelta(days=random.randint(0, 3))).isoformat() if direction == "incoming" else None,
            })
    return items

def generate_mom():
    items = []
    for proj in PROJECTS:
        n = random.randint(6, 15)
        for i in range(1, n + 1):
            status = random.choice(STATUSES_MOM)
            meeting_date = rand_date(2024, 2025)
            items.append({
                "id": f"mom-{proj['id']}-{i:04d}",
                "project_id": proj["id"],
                "project_name": proj["name"],
                "mom_number": i,
                "mom_code": f"MOM-{i:04d}",
                "meeting_title": f"Weekly Progress Meeting #{i}",
                "meeting_type": random.choice(MEETING_TYPES),
                "meeting_date": meeting_date.isoformat(),
                "location": random.choice(["Site Office", "Project Conference Room", "Online"]),
                "chairperson": random.choice(["Project Manager", "Construction Manager", "Client Rep"]),
                "attendees": "Contractor team, Consultant team, Client representatives",
                "minutes": "1. Review of previous action items\n2. Progress update\n3. Issues and risks\n4. New action items",
                "action_items": "1. Submit revised schedule by Friday\n2. Confirm material delivery dates\n3. Close out NCR-0038",
                "status": status,
                "distributed_at": (meeting_date + timedelta(days=2)).isoformat() if status != "draft" else None,
                "approved_at": (meeting_date + timedelta(days=5)).isoformat() if status in ["approved", "closed"] else None,
            })
    return items

def generate_daily_reports():
    items = []
    for proj in PROJECTS:
        n = random.randint(30, 60)
        base_date = date(2025, 1, 1)
        for i in range(n):
            report_date = base_date + timedelta(days=i * random.randint(1, 3))
            status = random.choice(STATUSES_DR)
            items.append({
                "id": f"dr-{proj['id']}-{i:04d}",
                "project_id": proj["id"],
                "project_name": proj["name"],
                "report_date": report_date.isoformat(),
                "shift": random.choice(["day", "night"]),
                "weather": random.choice(["Sunny", "Cloudy", "Rain", "Clear", "Windy", "Hot", "Mild"]),
                "temp_c": round(random.uniform(15, 42), 1),
                "manpower_total": random.randint(20, 150),
                "equipment_total": random.randint(5, 30),
                "narrative": f"Day shift operations focused on {random.choice(['excavation', 'concrete placement', 'steel erection', 'MEP installation', 'finishing works'])}.",
                "hse_notes": random.choice(["No incidents", "One near miss reported", "Safety induction completed", "Toolbox talk delivered"]),
                "delays": random.choice(["", "Equipment breakdown: 1 hour delay", "Weather delay: 30 min", "Material delivery delay"]),
                "work_completed": f"{random.choice(['Excavation', 'Concrete', 'Steelwork', 'Formwork', 'Rebar installation'])} at {random.choice(['Section A', 'Pier P05', 'Deck area', 'Building core'])}",
                "planned_tomorrow": f"Continue {random.choice(['excavation', 'concrete works', 'steel erection', 'MEP rough-in'])}",
                "author": random.choice(["Site Engineer 1", "Site Engineer 2", "General Foreman", "Shift Supervisor"]),
                "status": status,
            })
    return items

def generate_transmittals():
    items = []
    for proj in PROJECTS:
        n = random.randint(5, 15)
        for i in range(1, n + 1):
            status = random.choice(STATUSES_DT)
            sent_at = rand_date(2024, 2025)
            items.append({
                "id": f"dt-{proj['id']}-{i:04d}",
                "project_id": proj["id"],
                "project_name": proj["name"],
                "transmittal_number": i,
                "transmittal_code": f"DT-{i:04d}",
                "title": f"Document Transmittal #{i}",
                "purpose": random.choice(PURPOSE_TYPES),
                "from_entity": random.choice(["Contractor", "Consultant", "Design Office"]),
                "to_entity": random.choice(["Consultant", "Contractor", "Client"]),
                "document_list": f"1. Drawing SD-{random.randint(1,20):04d}_RevA.pdf\n2. Report RP-{random.randint(1,10):04d}.pdf\n3. Specification SEC-{random.randint(100,999)}.pdf",
                "notes": "For review and comment.",
                "status": status,
                "sent_at": sent_at.isoformat(),
                "received_at": (sent_at + timedelta(days=random.randint(0, 4))).isoformat() if status in ["received", "acknowledged", "rejected", "closed"] else None,
            })
    return items

def generate_revisions():
    items = []
    doc_types = ["rfi", "ncr", "submittal", "ms", "sd", "correspondence", "mom", "transmittal"]
    for proj in PROJECTS:
        for _ in range(random.randint(10, 30)):
            doc_type = random.choice(doc_types)
            items.append({
                "id": f"rev-{proj['id']}-{_+1:04d}",
                "project_id": proj["id"],
                "project_name": proj["name"],
                "document_type": doc_type,
                "document_id": f"{doc_type}-{proj['id']}-{random.randint(1, 20):04d}",
                "revision": random.choice(["A", "B", "C", "0", "1", "2"]),
                "change_summary": random.choice([
                    "Initial submission",
                    "Revised per consultant comments",
                    "Updated quantities",
                    "Drawing corrections per site condition",
                    "Specification update",
                ]),
                "created_by": random.choice(["Drafting Team", "Engineer", "Document Controller"]),
                "status": random.choice(["current", "superseded", "current", "current", "archived"]),
                "created_at": rand_date(2024, 2025).isoformat(),
            })
    return items

def main():
    data = {
        "generated_at": datetime.utcnow().isoformat(),
        "summary": {
            "rfis": 0, "ncrs": 0, "submittals": 0, "method_statements": 0,
            "shop_drawings": 0, "correspondence": 0, "minutes_of_meeting": 0,
            "daily_reports": 0, "transmittals": 0, "revisions": 0,
        },
        "rfis": generate_rfis(),
        "ncrs": generate_ncrs(),
        "submittals": generate_submittals(),
        "method_statements": generate_method_statements(),
        "shop_drawings": generate_shop_drawings(),
        "correspondence": generate_correspondence(),
        "minutes_of_meeting": generate_mom(),
        "daily_reports": generate_daily_reports(),
        "transmittals": generate_transmittals(),
        "revisions": generate_revisions(),
    }
    data["summary"]["rfis"] = len(data["rfis"])
    data["summary"]["ncrs"] = len(data["ncrs"])
    data["summary"]["submittals"] = len(data["submittals"])
    data["summary"]["method_statements"] = len(data["method_statements"])
    data["summary"]["shop_drawings"] = len(data["shop_drawings"])
    data["summary"]["correspondence"] = len(data["correspondence"])
    data["summary"]["minutes_of_meeting"] = len(data["minutes_of_meeting"])
    data["summary"]["daily_reports"] = len(data["daily_reports"])
    data["summary"]["transmittals"] = len(data["transmittals"])
    data["summary"]["revisions"] = len(data["revisions"])

    OUTPUT.parent.mkdir(parents=True, exist_ok=True)
    with open(OUTPUT, "w", encoding="utf-8") as f:
        json.dump(data, f, indent=2, ensure_ascii=False)

    print(f"[DocControl] Generated {data['summary']['rfis']} RFIs")
    print(f"[DocControl] Generated {data['summary']['ncrs']} NCRs")
    print(f"[DocControl] Generated {data['summary']['submittals']} Submittals")
    print(f"[DocControl] Generated {data['summary']['method_statements']} Method Statements")
    print(f"[DocControl] Generated {data['summary']['shop_drawings']} Shop Drawings")
    print(f"[DocControl] Generated {data['summary']['correspondence']} Correspondence")
    print(f"[DocControl] Generated {data['summary']['minutes_of_meeting']} Minutes of Meeting")
    print(f"[DocControl] Generated {data['summary']['daily_reports']} Daily Reports")
    print(f"[DocControl] Generated {data['summary']['transmittals']} Transmittals")
    print(f"[DocControl] Generated {data['summary']['revisions']} Revisions")
    print(f"[DocControl] Output: {OUTPUT}")

if __name__ == "__main__":
    main()