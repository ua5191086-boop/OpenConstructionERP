#!/usr/bin/env python3
"""Generate sample data for Permits module (V030)."""
import uuid, random
from datetime import datetime, timedelta

permit_types = ["construction", "environmental", "occupancy", "zoning"]
statuses = ["draft", "submitted", "under_review", "approved"]

print("-- Permits Module Seed Data (V030)")
print("BEGIN;")
# Regulatory bodies
bodies = [
    ("City Planning Dept", "CPD-01"),
    ("Environmental Agency", "EA-01"),
    ("Building Control", "BC-01"),
]
for name, code in bodies:
    bid = str(uuid.uuid4())
    print(f"INSERT INTO regulatory_bodies (id,body_name,body_code) VALUES ('{bid}','{name}','{code}');")

for i in range(8):
    aid = str(uuid.uuid4())
    pt = random.choice(permit_types)
    st = random.choice(statuses)
    pnum = f"PERM-{2024}-{1000+i}"
    sd = (datetime.now() - timedelta(days=random.randint(1,60))).date()
    print(f"INSERT INTO permit_applications (id,project_id,permit_number,permit_type,application_date,status) VALUES ('{aid}','proj-001','{pnum}','{pt}','{sd}','{st}');")
    # Conditions
    for j in range(2):
        cid = str(uuid.uuid4())
        print(f"INSERT INTO permit_conditions (id,permit_application_id,description,condition_type,status,due_date) VALUES ('{cid}','{aid}','Condition {j+1}','{random.choice(['prerequisite','ongoing'])}','{random.choice(['pending','satisfied'])}','{(datetime.now()+timedelta(days=30)).date()}');")
print("COMMIT;")