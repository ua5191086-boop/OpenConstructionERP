#!/usr/bin/env python3
"""
OpenConstructionERP — GIS & Survey Data Generator
Generates test data: Layers, Features, Survey Points, Survey Runs, Stations, Alignments, Cross Sections, Drone Flights
"""
import json, random
from datetime import date, timedelta
from pathlib import Path

random.seed(42)
OUTPUT = Path(__file__).parent.parent / "apps" / "web" / "gis_data.json"

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

def generate_layers():
    items = []
    layer_types = [
        ("Site Boundary", "polygon"), ("Existing Utilities", "linestring"), ("Topographic Contour", "linestring"),
        ("Boreholes", "point"), ("Traffic Control Zone", "polygon"), ("Environmental Buffer", "polygon"),
        ("Survey Control", "point"), ("Proposed Alignments", "linestring"), ("Drainage Network", "linestring"),
        ("Vegetation Index", "raster"), ("Orthophoto Mosaic", "raster"), ("Digital Terrain Model", "raster"),
    ]
    for proj in PROJECTS:
        for i, (ln, gt) in enumerate(layer_types):
            items.append({
                "id": f"lyr-{proj['id']}-{i:04d}",
                "project_id": proj["id"], "project_name": proj["name"],
                "layer_name": ln, "layer_type": "vector" if gt != "raster" else "raster",
                "geometry_type": gt, "source_type": pick(["upload","wms","api","generated"]),
                "is_visible": True, "status": "active",
            })
    return items

def generate_features():
    items = []
    for proj in PROJECTS:
        n = random.randint(30, 60)
        for i in range(1, n+1):
            ft = pick(["point","linestring","polygon"])
            lat = round(random.uniform(40.0, 42.0), 6)
            lng = round(random.uniform(-74.0, -72.0), 6)
            geom = {"type": ft, "coordinates": [[lng, lat], [lng+0.01, lat+0.01]] if ft == "linestring" else [lng, lat] if ft == "point" else [[[lng,lat],[lng+0.02,lat],[lng+0.02,lat+0.02],[lng,lat+0.02],[lng,lat]]]}
            items.append({
                "id": f"feat-{proj['id']}-{i:04d}",
                "project_id": proj["id"], "project_name": proj["name"],
                "feature_name": f"Feature {i}", "feature_type": ft,
                "geometry": geom,
                "properties": {"category": pick(["structure","utility","boundary","environment"])},
            })
    return items

def generate_survey_points():
    items = []
    for proj in PROJECTS:
        n = random.randint(20, 40)
        for i in range(1, n+1):
            items.append({
                "id": f"sp-{proj['id']}-{i:04d}",
                "project_id": proj["id"], "project_name": proj["name"],
                "point_number": i, "point_code": f"SP-{i:04d}",
                "point_name": f"Control Point {i}",
                "point_type": pick(["control","topographic","benchmark","boundary","monitoring"]),
                "latitude": round(random.uniform(40.0, 42.0), 7),
                "longitude": round(random.uniform(-74.0, -72.0), 7),
                "elevation": round(random.uniform(10, 200), 3),
                "northing": round(random.uniform(100000, 500000), 3),
                "easting": round(random.uniform(500000, 900000), 3),
                "zone": "18N",
                "accuracy_mm": round(random.uniform(2, 50), 2),
                "method": pick(["gps","total_station","level","drone"]),
                "survey_date": rand_date().isoformat(),
                "status": "active",
            })
    return items

def generate_survey_runs():
    items = []
    for proj in PROJECTS:
        n = random.randint(3, 8)
        for i in range(1, n+1):
            sd = rand_date()
            items.append({
                "id": f"sr-{proj['id']}-{i:04d}",
                "project_id": proj["id"], "project_name": proj["name"],
                "run_number": i, "run_code": f"SR-{i:04d}",
                "run_name": f"{pick(['Topographic','Control','As-Built','Staking','Monitoring'])} Survey Run #{i}",
                "survey_type": pick(["topographic","control","as_built","staking","monitoring"]),
                "start_date": sd.isoformat(),
                "end_date": (sd + timedelta(days=random.randint(1,14))).isoformat(),
                "instrument": pick(["Leica TS16","Trimble R12","Sokkia CX-105","Topcon GT"]),
                "crew_lead": pick(["Surveyor A","Surveyor B","Chief of Party"]),
                "point_count": random.randint(20, 200),
                "status": pick(["completed","completed","reviewed","approved","planned"]),
            })
    return items

def generate_stations():
    items = []
    for proj in PROJECTS:
        n = random.randint(15, 30)
        for i in range(1, n+1):
            items.append({
                "id": f"st-{proj['id']}-{i:04d}",
                "project_id": proj["id"], "project_name": proj["name"],
                "station_number": i, "station_code": f"STN-{i:04d}",
                "station_name": f"Station {i}",
                "station_type": pick(["traverse","benchmark","turning_point","control_point"]),
                "northing": round(random.uniform(100000, 500000), 3),
                "easting": round(random.uniform(500000, 900000), 3),
                "elevation": round(random.uniform(10, 200), 3),
            })
    return items

def generate_alignments():
    items = []
    for proj in PROJECTS:
        n = random.randint(1, 4)
        for i in range(1, n+1):
            items.append({
                "id": f"aln-{proj['id']}-{i:04d}",
                "project_id": proj["id"], "project_name": proj["name"],
                "alignment_code": f"ALN-{i:04d}",
                "alignment_name": f"{pick(['Main Road','Railway Track','Pipeline Route','Canal','Service Road'])} #{i}",
                "alignment_type": pick(["road","railway","pipeline","canal","bridge"]),
                "start_chainage": 0,
                "end_chainage": round(random.uniform(500, 5000), 2),
                "total_length": round(random.uniform(500, 5000), 2),
                "geometry": {"type": "LineString", "coordinates": [[-73.9, 40.7], [-73.8, 40.8], [-73.7, 40.75]]},
                "status": pick(["design","approved","as_built"]),
            })
    return items

def generate_cross_sections():
    items = []
    for proj in PROJECTS:
        n = random.randint(10, 25)
        for i in range(1, n+1):
            chainage = i * random.randint(20, 100)
            items.append({
                "id": f"cs-{proj['id']}-{i:04d}",
                "project_id": proj["id"], "project_name": proj["name"],
                "section_number": i, "chainage": chainage,
                "offset_left": round(random.uniform(10, 50), 2),
                "offset_right": round(random.uniform(10, 50), 2),
                "geometry": {"type": "LineString", "coordinates": [[-73.9, 40.7], [-73.89, 40.71]]},
                "points": [{"distance": -20, "elevation": round(random.uniform(50, 100), 2)},
                           {"distance": 0, "elevation": round(random.uniform(50, 100), 2)},
                           {"distance": 20, "elevation": round(random.uniform(50, 100), 2)}],
                "cut_area": round(random.uniform(0, 50), 3),
                "fill_area": round(random.uniform(0, 50), 3),
            })
    return items

def generate_drone_flights():
    items = []
    for proj in PROJECTS:
        n = random.randint(3, 8)
        for i in range(1, n+1):
            fd = rand_date()
            items.append({
                "id": f"uav-{proj['id']}-{i:04d}",
                "project_id": proj["id"], "project_name": proj["name"],
                "flight_number": i, "flight_code": f"UAV-{i:04d}",
                "flight_name": f"{pick(['Ortho','Thermal','Multispectral','Topo'])} Flight #{i} — {fd.strftime('%b %Y')}",
                "drone_model": pick(["DJI M300 RTK","DJI Mavic 3E","Autel EVO II","senseFly eBee"]),
                "pilot": pick(["Pilot A","Pilot B","Drone Operator"]),
                "flight_date": fd.isoformat(),
                "flight_duration_minutes": random.randint(10, 60),
                "altitude_m": round(random.uniform(50, 200), 2),
                "area_covered_ha": round(random.uniform(1, 50), 4),
                "gsd_cm": round(random.uniform(1, 10), 3),
                "overlap_pct": round(random.uniform(60, 90), 2),
                "images_count": random.randint(50, 500),
                "processing_status": pick(["completed","completed","processing","pending","failed"]),
                "sensor_type": pick(["rgb","multispectral","thermal","lidar"]),
                "output_type": pick(["orthophoto","point_cloud","dsm","dtm"]),
                "status": pick(["completed","completed","completed","planned","cancelled"]),
            })
    return items

def main():
    data = {
        "generated_at": date.today().isoformat(), "summary": {},
        "layers": generate_layers(), "features": generate_features(),
        "survey_points": generate_survey_points(), "survey_runs": generate_survey_runs(),
        "stations": generate_stations(), "alignments": generate_alignments(),
        "cross_sections": generate_cross_sections(), "drone_flights": generate_drone_flights(),
    }
    for k in ["layers","features","survey_points","survey_runs","stations","alignments","cross_sections","drone_flights"]:
        data["summary"][k] = len(data[k])
        print(f"[GS] Generated {len(data[k])} {k}")

    OUTPUT.parent.mkdir(parents=True, exist_ok=True)
    with open(OUTPUT, "w", encoding="utf-8") as f:
        json.dump(data, f, indent=2, ensure_ascii=False)
    print(f"[GS] Output: {OUTPUT}")

if __name__ == "__main__":
    main()