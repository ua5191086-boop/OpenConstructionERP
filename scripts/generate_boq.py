#!/usr/bin/env python3
"""
OpenConstructionERP — BOQ Generator for Railway Infrastructure Projects
Generates test data: sections, complexes, objects, BOQ items with CBS mapping
"""

import json
import random
import uuid
from datetime import datetime
from pathlib import Path

# ============================================================
# Configuration
# ============================================================
PROJECT_NAME = "Railway Line Alpha-Beta"
PROJECT_CODE = "RL-AB-001"
TOTAL_LENGTH_KM = 120
STATIONS = ["Station A", "Station B", "Station C", "Station D", "Station E"]
CURRENCY = "USD"
CONTRACTORS = [
    {"id": str(uuid.uuid4()), "name": "China Railway Construction Corp (CRCC)"},
    {"id": str(uuid.uuid4()), "name": "Strabag SE"},
    {"id": str(uuid.uuid4()), "name": "Vinci Construction"},
    {"id": str(uuid.uuid4()), "name": "Bechtel Infrastructure"},
    {"id": str(uuid.uuid4()), "name": "Local Subcontractor JV"},
]

# ============================================================
# CBS Chapters (from migration V001)
# ============================================================
CBS_CHAPTERS = {
    "01": "Site Preparation",
    "01.01": "Land Acquisition",
    "01.02": "Utility Relocation",
    "01.03": "Demolition",
    "01.04": "Preliminary & Auxiliary Works",
    "02": "Earthworks",
    "02.01": "Excavation",
    "02.02": "Embankment",
    "02.03": "Drainage",
    "02.04": "Slope Stabilization",
    "02.05": "Geotechnical Structures",
    "03": "Civil Structures",
    "03.01": "Bridges",
    "03.02": "Elevated Structures",
    "03.03": "Viaducts",
    "03.04": "Tunnels",
    "03.05": "Culverts",
    "03.06": "Retaining Walls",
    "04": "Track Works",
    "04.01": "Ballast",
    "04.02": "Track Structure (Rails & Sleepers)",
    "04.03": "Turnouts & Crossovers",
    "04.04": "Continuous Welded Rail",
    "04.05": "Level Crossings",
    "05": "Signalling, Control & Communication",
    "05.01": "Signalling Systems",
    "05.02": "Interlocking",
    "05.03": "Block Systems",
    "05.04": "Telecommunications",
    "05.05": "Fibre Optic Lines",
    "05.06": "Radio Communication",
    "05.07": "Data Transmission Systems",
    "05.08": "Control Centres",
    "05.09": "Equipment Rooms",
    "05.10": "Dispatch Centres",
    "06": "Buildings & Structures",
    "06.01": "Railway Stations",
    "06.02": "Station Buildings",
    "06.03": "Locomotive Depots",
    "06.04": "Rolling Stock Depots",
    "06.05": "Service Facilities",
    "06.06": "Administrative Buildings",
    "06.07": "Warehouse Complexes",
    "06.08": "Industrial Buildings",
    "07": "Power Supply System",
    "07.01": "Traction Substations",
    "07.02": "Transformer Substations",
    "07.03": "Overhead Contact Lines",
    "07.04": "External Power Supply",
    "07.05": "Cable Networks",
    "07.06": "Lighting Systems",
    "07.07": "Power Distribution & Management",
    "07.08": "Distribution Substations",
    "07.09": "Energy Dispatch Facilities",
    "08": "Utilities",
    "08.01": "Water Supply",
    "08.02": "Sewerage",
    "08.03": "Heating",
    "08.04": "Gas Supply",
    "08.05": "Drainage",
    "08.06": "Stormwater",
    "08.07": "External Utilities",
    "09": "Rolling Stock, Equipment, Furniture & Inventory",
    "09.01": "Locomotives",
    "09.02": "Electric Multiple Units (EMU)",
    "09.03": "Passenger Cars",
    "09.04": "Freight Cars",
    "09.05": "Process & Operational Equipment",
    "09.06": "Furniture",
    "09.07": "Inventory",
    "09.08": "Operational Equipment",
    "10": "Temporary Buildings & Facilities",
    "10.01": "Construction Camps",
    "10.02": "Temporary Warehouses",
    "10.03": "Temporary Utilities",
    "10.04": "Temporary Roads",
    "10.05": "Temporary Power Supply",
    "11": "Training & Employer Organization Costs",
    "11.01": "Personnel Training",
    "11.02": "Staff Development",
    "11.03": "Employer Organization Costs",
    "11.04": "Construction Supervision",
    "11.05": "Design Supervision",
    "11.06": "Consultancy Services",
    "12": "Other Costs",
    "12.01": "Design & Engineering Services",
    "12.02": "Expertise & Approvals",
    "12.03": "Studies & Surveys",
    "12.04": "Permits & Initial Documentation",
    "12.05": "Insurance",
    "12.06": "Regulatory Approvals",
    "12.07": "Other Related Costs",
}


# ============================================================
# BOQ Template — Typical railway work items
# ============================================================
TRACK_WORKS_TEMPLATE = [
    {"name": "Topsoil Stripping & Site Clearance", "unit": "m²", "qty_per_km": 12000, "price": 2.50, "cbs": "01.04"},
    {"name": "Excavation to Formation Level", "unit": "m³", "qty_per_km": 15000, "price": 8.00, "cbs": "02.01"},
    {"name": "Embankment Fill Material", "unit": "m³", "qty_per_km": 18000, "price": 12.00, "cbs": "02.02"},
    {"name": "Sub-ballast Layer (300mm)", "unit": "m³", "qty_per_km": 4500, "price": 18.00, "cbs": "04.01"},
    {"name": "Ballast Layer (350mm)", "unit": "m³", "qty_per_km": 5250, "price": 25.00, "cbs": "04.01"},
    {"name": "Rail Installation (60E1)", "unit": "m", "qty_per_km": 2000, "price": 150.00, "cbs": "04.02"},
    {"name": "Concrete Sleepers Installation", "unit": "pc", "qty_per_km": 1667, "price": 45.00, "cbs": "04.02"},
    {"name": "Rail Fastening System", "unit": "set", "qty_per_km": 3334, "price": 12.00, "cbs": "04.02"},
    {"name": "Continuous Welded Rail (CWR)", "unit": "m", "qty_per_km": 2000, "price": 85.00, "cbs": "04.04"},
    {"name": "Drainage System (Side Drains)", "unit": "m", "qty_per_km": 2000, "price": 35.00, "cbs": "02.03"},
    {"name": "Cable Ducts & Trenches", "unit": "m", "qty_per_km": 4000, "price": 20.00, "cbs": "07.05"},
    {"name": "Signalling Cable Laying", "unit": "m", "qty_per_km": 3000, "price": 28.00, "cbs": "05.01"},
    {"name": "Fibre Optic Cable Installation", "unit": "m", "qty_per_km": 2000, "price": 35.00, "cbs": "05.05"},
    {"name": "Overhead Contact Line (OCL) Mast", "unit": "pc", "qty_per_km": 33, "price": 2500.00, "cbs": "07.03"},
    {"name": "OCL Catenary Wire Installation", "unit": "m", "qty_per_km": 2000, "price": 95.00, "cbs": "07.03"},
    {"name": "Fencing (Both Sides)", "unit": "m", "qty_per_km": 2000, "price": 18.00, "cbs": "01.04"},
    {"name": "Slope Protection & Erosion Control", "unit": "m²", "qty_per_km": 8000, "price": 6.00, "cbs": "02.04"},
]

STATION_TEMPLATE = [
    {"name": "Station Building Construction", "unit": "m²", "qty": 5000, "price": 1200.00, "cbs": "06.02"},
    {"name": "Platform Construction", "unit": "m", "qty": 400, "price": 3500.00, "cbs": "06.01"},
    {"name": "Canopy Roof Structure", "unit": "m²", "qty": 2000, "price": 450.00, "cbs": "06.01"},
    {"name": "Passenger Information System", "unit": "lump", "qty": 1, "price": 500000.00, "cbs": "05.04"},
    {"name": "Ticketing & Fare Collection System", "unit": "lump", "qty": 1, "price": 800000.00, "cbs": "05.04"},
    {"name": "Elevators & Escalators", "unit": "pc", "qty": 6, "price": 250000.00, "cbs": "06.01"},
    {"name": "HVAC System", "unit": "lump", "qty": 1, "price": 1200000.00, "cbs": "08.03"},
    {"name": "Fire Protection System", "unit": "lump", "qty": 1, "price": 600000.00, "cbs": "08.02"},
    {"name": "Lighting System (Station)", "unit": "lump", "qty": 1, "price": 400000.00, "cbs": "07.06"},
    {"name": "Water Supply & Sanitary", "unit": "lump", "qty": 1, "price": 350000.00, "cbs": "08.01"},
    {"name": "Signage & Wayfinding", "unit": "lump", "qty": 1, "price": 200000.00, "cbs": "06.01"},
    {"name": "Landscaping & External Works", "unit": "m²", "qty": 3000, "price": 85.00, "cbs": "01.04"},
    {"name": "Parking Facilities", "unit": "m²", "qty": 5000, "price": 150.00, "cbs": "06.01"},
    {"name": "Security Systems (CCTV, Access)", "unit": "lump", "qty": 1, "price": 300000.00, "cbs": "05.04"},
    {"name": "Interlocking Equipment Room", "unit": "lump", "qty": 1, "price": 450000.00, "cbs": "05.09"},
]

BRIDGE_TEMPLATE = [
    {"name": "Pile Foundation (Bored Piles)", "unit": "m", "qty": 200, "price": 450.00, "cbs": "03.01"},
    {"name": "Pier Construction (Reinforced Concrete)", "unit": "m³", "qty": 500, "price": 350.00, "cbs": "03.01"},
    {"name": "Pier Cap Construction", "unit": "pc", "qty": 10, "price": 25000.00, "cbs": "03.01"},
    {"name": "Girder Fabrication & Erection (Steel)", "unit": "ton", "qty": 300, "price": 3200.00, "cbs": "03.01"},
    {"name": "Deck Slab Construction", "unit": "m²", "qty": 1200, "price": 280.00, "cbs": "03.01"},
    {"name": "Bridge Bearings", "unit": "pc", "qty": 20, "price": 8500.00, "cbs": "03.01"},
    {"name": "Expansion Joints", "unit": "m", "qty": 40, "price": 1200.00, "cbs": "03.01"},
    {"name": "Waterproofing Membrane", "unit": "m²", "qty": 1200, "price": 45.00, "cbs": "03.01"},
    {"name": "Rail Fastening on Bridge", "unit": "m", "qty": 200, "price": 180.00, "cbs": "04.02"},
    {"name": "Bridge Parapet & Safety Barriers", "unit": "m", "qty": 200, "price": 250.00, "cbs": "03.01"},
]

TUNNEL_TEMPLATE = [
    {"name": "TBM Procurement & Mobilization", "unit": "lump", "qty": 1, "price": 15000000.00, "cbs": "03.04"},
    {"name": "TBM Tunnelling (EPB)", "unit": "m", "qty": 1000, "price": 8500.00, "cbs": "03.04"},
    {"name": "Segment Lining (Precast)", "unit": "ring", "qty": 667, "price": 3500.00, "cbs": "03.04"},
    {"name": "Backfill Grouting", "unit": "m³", "qty": 3000, "price": 85.00, "cbs": "03.04"},
    {"name": "Launch Shaft Construction", "unit": "lump", "qty": 1, "price": 2500000.00, "cbs": "03.04"},
    {"name": "Reception Shaft Construction", "unit": "lump", "qty": 1, "price": 1800000.00, "cbs": "03.04"},
    {"name": "Cross Passage Construction", "unit": "pc", "qty": 4, "price": 1200000.00, "cbs": "03.04"},
    {"name": "Track Slab (Invert)", "unit": "m", "qty": 1000, "price": 2500.00, "cbs": "04.02"},
    {"name": "Tunnel Drainage System", "unit": "m", "qty": 1000, "price": 350.00, "cbs": "02.03"},
    {"name": "Tunnel Lighting", "unit": "m", "qty": 1000, "price": 280.00, "cbs": "07.06"},
    {"name": "Tunnel Ventilation System", "unit": "lump", "qty": 1, "price": 3000000.00, "cbs": "03.04"},
    {"name": "Fire Protection (Tunnel)", "unit": "m", "qty": 1000, "price": 450.00, "cbs": "08.02"},
    {"name": "Settlement Monitoring System", "unit": "lump", "qty": 1, "price": 500000.00, "cbs": "03.04"},
    {"name": "Geotechnical Instrumentation", "unit": "lump", "qty": 1, "price": 350000.00, "cbs": "02.05"},
]


def generate_boq():
    """Generate complete BOQ for a railway project"""
    project_id = str(uuid.uuid4())
    output = {
        "project": {
            "id": project_id,
            "code": PROJECT_CODE,
            "name": PROJECT_NAME,
            "total_length_km": TOTAL_LENGTH_KM,
            "stations": STATIONS,
            "currency": CURRENCY,
        },
        "cbs_chapters": [],
        "sections": [],
        "complexes": [],
        "objects": [],
        "boq_items": [],
        "summary": {},
    }

    # Add CBS chapters
    for code, name in CBS_CHAPTERS.items():
        output["cbs_chapters"].append({
            "id": str(uuid.uuid4()),
            "code": code,
            "name": name,
            "level": 1 if len(code) == 2 else 2,
        })

    # Generate track sections (every 30 km)
    section_lengths = [30, 30, 30, 30]  # 4 sections × 30 km = 120 km
    current_km = 0
    section_id_map = {}

    for i, length in enumerate(section_lengths):
        start_km = current_km
        end_km = current_km + length
        section_id = str(uuid.uuid4())
        section = {
            "id": section_id,
            "code": f"TR-{i+1:02d}",
            "name": f"Track Section KM {start_km}-{end_km}",
            "section_type": "Track",
            "start_km": start_km,
            "end_km": end_km,
            "length_km": length,
        }
        output["sections"].append(section)
        section_id_map[f"TR-{i+1:02d}"] = section_id
        current_km = end_km

    # Add station sections
    for i, station in enumerate(STATIONS):
        section_id = str(uuid.uuid4())
        section = {
            "id": section_id,
            "code": f"ST-{i+1:02d}",
            "name": station,
            "section_type": "Station",
            "start_km": None,
            "end_km": None,
            "length_km": None,
        }
        output["sections"].append(section)
        section_id_map[f"ST-{i+1:02d}"] = section_id

    # Generate complexes and objects for each track section
    for sec in output["sections"]:
        if sec["section_type"] == "Track":
            # Earthworks Complex
            complex_id = str(uuid.uuid4())
            output["complexes"].append({
                "id": complex_id,
                "section_id": sec["id"],
                "code": f"{sec['code']}-EW",
                "name": f"Earthworks — {sec['name']}",
            })
            # Earthworks Object
            obj_id = str(uuid.uuid4())
            output["objects"].append({
                "id": obj_id,
                "complex_id": complex_id,
                "code": f"{sec['code']}-EW-01",
                "name": f"Main Earthworks — {sec['name']}",
            })
            # Add BOQ items for earthworks
            for item in TRACK_WORKS_TEMPLATE[:6]:
                qty = round(item["qty_per_km"] * sec["length_km"], 2)
                output["boq_items"].append({
                    "id": str(uuid.uuid4()),
                    "object_id": obj_id,
                    "cbs_code": item["cbs"],
                    "code": f"{sec['code']}-{item['cbs'].replace('.', '')}-{random.randint(100,999)}",
                    "name": item["name"],
                    "unit": item["unit"],
                    "quantity": qty,
                    "unit_price": item["price"],
                    "total_cost": round(qty * item["price"], 2),
                    "contractor": random.choice(CONTRACTORS),
                })

            # Track Works Complex
            complex_id2 = str(uuid.uuid4())
            output["complexes"].append({
                "id": complex_id2,
                "section_id": sec["id"],
                "code": f"{sec['code']}-TW",
                "name": f"Track Works — {sec['name']}",
            })
            obj_id2 = str(uuid.uuid4())
            output["objects"].append({
                "id": obj_id2,
                "complex_id": complex_id2,
                "code": f"{sec['code']}-TW-01",
                "name": f"Main Track — {sec['name']}",
            })
            for item in TRACK_WORKS_TEMPLATE[6:12]:
                qty = round(item["qty_per_km"] * sec["length_km"], 2)
                output["boq_items"].append({
                    "id": str(uuid.uuid4()),
                    "object_id": obj_id2,
                    "cbs_code": item["cbs"],
                    "code": f"{sec['code']}-{item['cbs'].replace('.', '')}-{random.randint(100,999)}",
                    "name": item["name"],
                    "unit": item["unit"],
                    "quantity": qty,
                    "unit_price": item["price"],
                    "total_cost": round(qty * item["price"], 2),
                    "contractor": random.choice(CONTRACTORS),
                })

            # OCL & Signalling Complex
            complex_id3 = str(uuid.uuid4())
            output["complexes"].append({
                "id": complex_id3,
                "section_id": sec["id"],
                "code": f"{sec['code']}-OS",
                "name": f"OCL & Signalling — {sec['name']}",
            })
            obj_id3 = str(uuid.uuid4())
            output["objects"].append({
                "id": obj_id3,
                "complex_id": complex_id3,
                "code": f"{sec['code']}-OS-01",
                "name": f"Overhead Line & Signalling — {sec['name']}",
            })
            for item in TRACK_WORKS_TEMPLATE[12:]:
                qty = round(item["qty_per_km"] * sec["length_km"], 2)
                output["boq_items"].append({
                    "id": str(uuid.uuid4()),
                    "object_id": obj_id3,
                    "cbs_code": item["cbs"],
                    "code": f"{sec['code']}-{item['cbs'].replace('.', '')}-{random.randint(100,999)}",
                    "name": item["name"],
                    "unit": item["unit"],
                    "quantity": qty,
                    "unit_price": item["price"],
                    "total_cost": round(qty * item["price"], 2),
                    "contractor": random.choice(CONTRACTORS),
                })

            # Bridge (one per section)
            complex_id4 = str(uuid.uuid4())
            output["complexes"].append({
                "id": complex_id4,
                "section_id": sec["id"],
                "code": f"{sec['code']}-BR",
                "name": f"Bridge Structures — {sec['name']}",
            })
            obj_id4 = str(uuid.uuid4())
            output["objects"].append({
                "id": obj_id4,
                "complex_id": complex_id4,
                "code": f"{sec['code']}-BR-01",
                "name": f"Bridge at KM {sec['start_km'] + sec['length_km']/2}",
            })
            for item in BRIDGE_TEMPLATE:
                output["boq_items"].append({
                    "id": str(uuid.uuid4()),
                    "object_id": obj_id4,
                    "cbs_code": item["cbs"],
                    "code": f"{sec['code']}-{item['cbs'].replace('.', '')}-{random.randint(100,999)}",
                    "name": item["name"],
                    "unit": item["unit"],
                    "quantity": item["qty"],
                    "unit_price": item["price"],
                    "total_cost": round(item["qty"] * item["price"], 2),
                    "contractor": random.choice(CONTRACTORS),
                })

        elif sec["section_type"] == "Station":
            # Station Building Complex
            complex_id = str(uuid.uuid4())
            output["complexes"].append({
                "id": complex_id,
                "section_id": sec["id"],
                "code": f"{sec['code']}-BLD",
                "name": f"Buildings — {sec['name']}",
            })
            obj_id = str(uuid.uuid4())
            output["objects"].append({
                "id": obj_id,
                "complex_id": complex_id,
                "code": f"{sec['code']}-BLD-01",
                "name": f"Main Station Building — {sec['name']}",
            })
            for item in STATION_TEMPLATE[:8]:
                output["boq_items"].append({
                    "id": str(uuid.uuid4()),
                    "object_id": obj_id,
                    "cbs_code": item["cbs"],
                    "code": f"{sec['code']}-{item['cbs'].replace('.', '')}-{random.randint(100,999)}",
                    "name": item["name"],
                    "unit": item["unit"],
                    "quantity": item["qty"],
                    "unit_price": item["price"],
                    "total_cost": round(item["qty"] * item["price"], 2),
                    "contractor": random.choice(CONTRACTORS),
                })

            # Station Systems Complex
            complex_id2 = str(uuid.uuid4())
            output["complexes"].append({
                "id": complex_id2,
                "section_id": sec["id"],
                "code": f"{sec['code']}-SYS",
                "name": f"Station Systems — {sec['name']}",
            })
            obj_id2 = str(uuid.uuid4())
            output["objects"].append({
                "id": obj_id2,
                "complex_id": complex_id2,
                "code": f"{sec['code']}-SYS-01",
                "name": f"Systems & Equipment — {sec['name']}",
            })
            for item in STATION_TEMPLATE[8:]:
                output["boq_items"].append({
                    "id": str(uuid.uuid4()),
                    "object_id": obj_id2,
                    "cbs_code": item["cbs"],
                    "code": f"{sec['code']}-{item['cbs'].replace('.', '')}-{random.randint(100,999)}",
                    "name": item["name"],
                    "unit": item["unit"],
                    "quantity": item["qty"],
                    "unit_price": item["price"],
                    "total_cost": round(item["qty"] * item["price"], 2),
                    "contractor": random.choice(CONTRACTORS),
                })

    # Calculate summary
    total_boq = sum(item["total_cost"] for item in output["boq_items"])
    output["summary"] = {
        "total_boq_cost": round(total_boq, 2),
        "total_sections": len(output["sections"]),
        "total_complexes": len(output["complexes"]),
        "total_objects": len(output["objects"]),
        "total_boq_items": len(output["boq_items"]),
        "currency": CURRENCY,
        "generated_at": datetime.now().isoformat(),
    }

    return output


def main():
    print("=" * 70)
    print(f"OpenConstructionERP — BOQ Generator")
    print(f"Project: {PROJECT_NAME} ({PROJECT_CODE})")
    print(f"Length: {TOTAL_LENGTH_KM} km, Stations: {len(STATIONS)}")
    print("=" * 70)

    boq = generate_boq()

    # Save as JSON
    output_dir = Path(__file__).parent.parent / "output"
    output_dir.mkdir(exist_ok=True)
    output_file = output_dir / f"boq_{PROJECT_CODE.lower()}.json"

    with open(output_file, "w", encoding="utf-8") as f:
        json.dump(boq, f, indent=2, ensure_ascii=False)

    print(f"\n✅ BOQ generated successfully!")
    print(f"📄 File: {output_file}")
    print(f"\n📊 Summary:")
    print(f"   Sections:     {boq['summary']['total_sections']}")
    print(f"   Complexes:    {boq['summary']['total_complexes']}")
    print(f"   Objects:      {boq['summary']['total_objects']}")
    print(f"   BOQ Items:    {boq['summary']['total_boq_items']}")
    print(f"   Total Cost:    ${boq['summary']['total_boq_cost']:,.2f}")

    # Print CBS breakdown
    print(f"\n📊 Cost by CBS Chapter:")
    cbs_totals = {}
    for item in boq["boq_items"]:
        cbs = item["cbs_code"]
        cbs_totals[cbs] = cbs_totals.get(cbs, 0) + item["total_cost"]

    for cbs_code in sorted(cbs_totals.keys()):
        cbs_name = CBS_CHAPTERS.get(cbs_code, "Unknown")
        total = cbs_totals[cbs_code]
        print(f"   {cbs_code} {cbs_name}: ${total:>12,.2f}")

    print(f"\n{'=' * 70}")
    print(f"TOTAL BOQ: ${boq['summary']['total_boq_cost']:>14,.2f}")
    print(f"{'=' * 70}")


if __name__ == "__main__":
    main()
