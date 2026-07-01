#!/usr/bin/env python3
"""
OpenConstructionERP — canonical reference project seeder.

Creates ALM-L3-REF: an idealized metro section from world practice —
one cut-and-cover station + twin single-track TBM running tunnels
(EPB Ø5.85 m, ring 1.4 m) — and populates EVERY implemented module
through the public API: BOQ, tunnel drives with 60 days of ring history,
daily reports with quantities, RFIs, CDE documents + transmittal,
budget baseline, cost transactions.

Purpose: a deterministic test bed and demo. Run after `docker compose up`:
    python3 scripts/seed_reference_project.py [--api http://localhost:8000]

Quantities are engineering-realistic (not contractual): station 140x22 m,
d-walls 800 mm / 28 m deep, top-down; tunnels 2 x 2750 m bored.
"""
import argparse
import json
import random
import urllib.request
from datetime import date, datetime, timedelta

random.seed(42)  # deterministic


def api(base, path, payload=None, method=None):
    req = urllib.request.Request(base + path,
                                 method=method or ("POST" if payload else "GET"))
    data = None
    if payload is not None:
        req.add_header("Content-Type", "application/json")
        data = json.dumps(payload, default=str).encode()
    with urllib.request.urlopen(req, data=data) as r:
        return json.loads(r.read())


# --------------------------------------------------------------- BOQ -------
# (cbs, code, name, unit, qty, rate) — idealized world-practice rates, USD
BOQ = [
    # Station box — ST-1 section
    ("01.04", "ST-PRE-001", "Site establishment, hoarding, traffic management", "ls", 1, 1_800_000),
    ("02.01", "ST-EXC-001", "Bulk excavation station box (top-down), incl. disposal", "m3", 86_000, 22),
    ("03.06", "ST-DW-001", "Diaphragm wall 800mm, depth 28m, incl. guide walls", "m2", 9_100, 340),
    ("03.06", "ST-DW-002", "Temporary steel struts and walers, supply & install", "t", 1_450, 2_100),
    ("06.02", "ST-STR-001", "Base slab C40/50, reinforced, waterproofed", "m3", 6_800, 310),
    ("06.02", "ST-STR-002", "Concourse and platform slabs C40/50", "m3", 5_200, 335),
    ("06.02", "ST-STR-003", "Internal walls and columns C40/50", "m3", 3_900, 360),
    ("06.02", "ST-STR-004", "Reinforcement B500C for station structures", "t", 2_650, 1_150),
    ("03.06", "ST-WP-001", "Waterproofing membrane system, base and walls", "m2", 14_500, 38),
    ("06.02", "ST-ARC-001", "Architectural finishes platform level", "m2", 8_400, 210),
    ("07.02", "ST-MEP-001", "Station MEP first fix (provisional)", "ls", 1, 6_500_000),
    ("08.05", "ST-DRN-001", "Station drainage and sump pumping stations", "ls", 1, 940_000),
    # Running tunnels — TUN-L / TUN-R sections
    ("03.04", "TN-TBM-001", "TBM boring single-track tunnel Ø5.85m EPB, left drive", "m", 2_750, 4_950),
    ("03.04", "TN-TBM-002", "TBM boring single-track tunnel Ø5.85m EPB, right drive", "m", 2_750, 4_950),
    ("03.04", "TN-RNG-001", "Precast segmental lining rings 1.4m, supply & erect, left", "ring", 1_964, 3_850),
    ("03.04", "TN-RNG-002", "Precast segmental lining rings 1.4m, supply & erect, right", "ring", 1_964, 3_850),
    ("03.04", "TN-GRT-001", "Annulus grouting, both drives", "m3", 12_400, 92),
    ("03.04", "TN-XP-001", "Cross passages 2 no., ground treatment + excavation + lining", "ea", 2, 1_650_000),
    ("02.03", "TN-DRN-001", "Tunnel invert drainage and walkway", "m", 5_500, 260),
    ("03.04", "TN-MOB-001", "TBM mobilisation, launch shaft fit-out, conveyor", "ls", 1, 7_800_000),
    ("03.04", "TN-DEM-001", "TBM reception, demobilisation and refurbishment allowance", "ls", 1, 2_900_000),
    # Trackwork & systems (provisional)
    ("04.02", "TR-TRK-001", "Slab track on rigid base, both bores", "m", 5_500, 690),
    ("05.01", "SY-SIG-001", "Signalling ATP/ATO section equipment (provisional)", "km", 5.5, 890_000),
    ("07.03", "SY-OCS-001", "Rigid overhead conductor rail, both bores", "m", 5_500, 305),
    # Prelims & monitoring
    ("12.03", "GN-MON-001", "Geotechnical instrumentation and monitoring, 30 months", "mo", 30, 46_000),
    ("10.01", "GN-CMP-001", "Site camp and offices, 30 months", "mo", 30, 58_000),
    ("11.04", "GN-SUP-001", "Construction supervision support (Employer's requirements)", "mo", 30, 41_000),
]

SECTION = {"ST": "ST-1", "TN": "TUN", "TR": "TUN", "SY": "SYS", "GN": "GEN"}


def build_boq_xlsx():
    import io
    import openpyxl
    wb = openpyxl.Workbook(); ws = wb.active
    ws.append(["code", "name", "unit", "quantity", "unit_price", "section", "cbs"])
    for cbs, code, name, unit, qty, rate in BOQ:
        ws.append([code, name, unit, qty, rate, SECTION[code[:2]], cbs])
    buf = io.BytesIO(); wb.save(buf)
    return buf.getvalue()


def upload(base, path, content, filename="seed.xlsx"):
    boundary = "----oceseed"
    body = (f"--{boundary}\r\nContent-Disposition: form-data; name=\"file\"; "
            f"filename=\"{filename}\"\r\nContent-Type: application/octet-stream"
            f"\r\n\r\n").encode() + content + f"\r\n--{boundary}--\r\n".encode()
    req = urllib.request.Request(base + path, data=body, method="POST")
    req.add_header("Content-Type", f"multipart/form-data; boundary={boundary}")
    with urllib.request.urlopen(req) as r:
        return json.loads(r.read())


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--api", default="http://localhost:8000")
    ap.add_argument("--code", default="ALM-L3-REF")
    a = ap.parse_args()
    base = a.api

    # 1) Project ---------------------------------------------------------
    projects = api(base, "/api/v1/projects")
    proj = next((p for p in projects if p["code"] == a.code), None)
    if proj:
        print(f"Project {a.code} already exists — aborting to keep seed deterministic.")
        return
    proj = api(base, "/api/v1/projects", {
        "code": a.code,
        "name": "Almaty Metro Line 3 — Reference Section (station + twin bores)",
        "project_type": "metro", "status": "execution",
        "country": "KZ", "currency": "USD"})
    pid = proj["id"]
    print(f"[1/7] project {a.code}")

    # 2) BOQ --------------------------------------------------------------
    res = upload(base, f"/api/v1/projects/{pid}/boq/import", build_boq_xlsx())
    summary = api(base, f"/api/v1/projects/{pid}/boq/summary?region=KZ")
    print(f"[2/7] BOQ: {res['imported']} items, base ${summary['base_total']:,.0f}, "
          f"KZ-adjusted ${summary['adjusted_total']:,.0f}")

    # 3) Budget baseline --------------------------------------------------
    api(base, f"/api/v1/projects/{pid}/budget/versions",
        {"version_name": "Baseline B0", "notes": "Reference seed baseline"})
    print("[3/7] budget baseline B0 frozen")

    # 4) Tunnel: two drives + 60 days of rings ---------------------------
    drives = {}
    for code, name, frm, to in (("DR-L", "Left running tunnel", 0, 2750),
                                ("DR-R", "Right running tunnel", 0, 2750)):
        d = api(base, f"/api/v1/projects/{pid}/tunnel/drives",
                {"code": code, "name": name, "chainage_from": frm,
                 "chainage_to": to, "ring_width_mm": 1400, "tbm_code": "S-880"})
        drives[code] = d["id"]
    start = date.today() - timedelta(days=60)
    ring_no = {"DR-L": 0, "DR-R": 0}
    lag = {"DR-L": 0, "DR-R": 14}  # right drive launches 2 weeks later
    for day in range(60):
        d0 = start + timedelta(days=day)
        for code in ("DR-L", "DR-R"):
            if day < lag[code]:
                continue
            # learning curve: 6 -> 14 rings/day with noise, one maintenance day/12
            if (day - lag[code]) % 12 == 11:
                continue
            base_rate = min(6 + (day - lag[code]) * 0.18, 14)
            rings = []
            for shift, share in (("day", 0.55), ("night", 0.45)):
                n = max(1, round(base_rate * share + random.uniform(-1, 1)))
                for _ in range(n):
                    ring_no[code] += 1
                    rings.append({
                        "ring_no": ring_no[code],
                        "built_at": datetime.combine(
                            d0, datetime.min.time()).replace(
                            hour=8 if shift == "day" else 20).isoformat(),
                        "shift": shift, "advance_mm": 1400,
                        "grout_volume_m3": round(random.uniform(2.9, 3.4), 2),
                        "grout_pressure_bar": round(random.uniform(2.0, 2.8), 2)})
            api(base, f"/api/v1/projects/{pid}/tunnel/drives/{drives[code]}/rings",
                {"rings": rings})
    for code in ("DR-L", "DR-R"):
        p = api(base, f"/api/v1/projects/{pid}/tunnel/drives/{drives[code]}/progress")
        print(f"[4/7] {code}: {p['rings_built']} rings ({p['percent']}%), "
              f"{p['avg_rings_per_day']}/day, ETA {p['eta_working_days']} wd")

    # 5) Daily reports (last 10 days, with station quantities) -----------
    for day in range(10):
        d0 = date.today() - timedelta(days=10 - day)
        api(base, f"/api/v1/projects/{pid}/daily-reports", {
            "report_date": str(d0), "shift": "day",
            "weather": random.choice(["ясно", "облачно", "дождь"]),
            "temp_c": round(random.uniform(18, 31), 1),
            "manpower_total": random.randint(240, 290),
            "equipment_total": random.randint(28, 36),
            "narrative": "TBM boring both drives; station: d-wall panels, excavation L3",
            "author": "Shift Engineer",
            "entries": [
                {"boq_item_code": "ST-DW-001",
                 "qty_done": round(random.uniform(60, 95), 1), "location": "panels"},
                {"boq_item_code": "ST-EXC-001",
                 "qty_done": round(random.uniform(700, 1100), 0), "location": "L3"},
                {"boq_item_code": "TN-GRT-001",
                 "qty_done": round(random.uniform(55, 80), 1), "location": "both drives"},
            ]})
    prog = api(base, f"/api/v1/projects/{pid}/progress/physical")
    print(f"[5/7] daily reports: physical {prog['physical_percent']}% "
          f"(earned ${prog['earned_value']:,.0f})")

    # 6) Cost: commitments + monthly actuals ------------------------------
    txs = [
        {"boq_item_code": "TN-TBM-001", "transaction_type": "Commitment",
         "amount": 13_612_500, "period": str(start), "description": "TBM boring left, subcontract"},
        {"boq_item_code": "TN-TBM-002", "transaction_type": "Commitment",
         "amount": 13_612_500, "period": str(start), "description": "TBM boring right, subcontract"},
        {"boq_item_code": "ST-DW-001", "transaction_type": "Commitment",
         "amount": 3_094_000, "period": str(start), "description": "D-wall specialist"},
    ]
    for m in range(2):
        p0 = (start + timedelta(days=30 * (m + 1)))
        txs += [
            {"boq_item_code": "TN-TBM-001", "transaction_type": "Actual",
             "amount": 2_050_000 + m * 480_000, "period": str(p0), "description": f"IPC-{m+1} left"},
            {"boq_item_code": "ST-EXC-001", "transaction_type": "Actual",
             "amount": 340_000 + m * 110_000, "period": str(p0), "description": f"IPC-{m+1} excavation"},
        ]
    api(base, f"/api/v1/projects/{pid}/cost/transactions", {"transactions": txs})
    cs = api(base, f"/api/v1/projects/{pid}/cost/summary")
    print(f"[6/7] cost: committed ${cs['total']['committed']:,.0f}, "
          f"actual ${cs['total']['actual']:,.0f}")

    # 7) CDE + RFI --------------------------------------------------------
    docs = []
    for t, ty, disc in (
            ("Method Statement — TBM launch and initial drive", "MS", "tunnelling"),
            ("Method Statement — diaphragm wall construction", "MS", "geotech"),
            ("Drawing — universal ring Ø5.85m segment details", "DWG", "structural"),
            ("ITP — segmental lining erection", "ITP", "quality")):
        d = api(base, f"/api/v1/projects/{pid}/documents",
                {"title": t, "doc_type": ty, "discipline": disc})
        docs.append(d)
    for d in docs[:3]:
        api(base, f"/api/v1/projects/{pid}/documents/{d['id']}/state",
            {"state": "Shared", "suitability": "S2"})
    api(base, f"/api/v1/projects/{pid}/transmittals", {
        "to_party": "Engineer (Almaty Metro Directorate)",
        "purpose": "for_review", "issued_by": "Document Controller",
        "document_numbers": [d["doc_number"] for d in docs[:3]]})
    api(base, f"/api/v1/projects/{pid}/rfis", {
        "subject": "EPB face pressure range through water-bearing gravels PK14-PK19",
        "question": "GBR indicates 2.1-2.4 bar hydrostatic; confirm design face "
                    "pressure envelope and permissible deviation for the fleet.",
        "discipline": "geotech", "raised_by": "TBM Manager",
        "assigned_to": "Designer", "due_date": str(date.today() + timedelta(days=7))})
    api(base, f"/api/v1/projects/{pid}/rfis", {
        "subject": "Cross passage CP-1 ground freezing vs jet grouting",
        "question": "Request confirmation of ground treatment method for CP-1 "
                    "given utilities congestion above.",
        "discipline": "geotech", "raised_by": "Chief Engineer",
        "due_date": str(date.today() - timedelta(days=3))})  # deliberately overdue
    print(f"[7/7] CDE: {len(docs)} docs, TRN-0001 issued; 2 RFIs (1 overdue)")
    print(f"\nDone. Dashboards: {base}/  {base}/tunnel.html   "
          f"Executive: {base}/api/v1/projects/{pid}/reports/executive.xlsx")


if __name__ == "__main__":
    main()
