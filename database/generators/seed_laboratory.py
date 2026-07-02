#!/usr/bin/env python3
"""Generate sample data for Laboratory module (V029)."""
import uuid, random
from datetime import datetime, timedelta

materials = ["concrete", "steel", "soil"]
statuses = ["pending", "in_progress", "completed"]
results = ["pass", "fail", "pending"]

print("-- Laboratory Module Seed Data (V029)")
print("BEGIN;")
for i in range(10):
    tid = str(uuid.uuid4())
    mt = random.choice(materials)
    st = random.choice(statuses)
    num = f"MT-{2024}-{1000+i}"
    sd = (datetime.now() - timedelta(days=random.randint(1,30))).date()
    print(f"INSERT INTO material_testing (id,project_id,test_number,material_type,test_type,test_date,status,tested_by) VALUES ('{tid}','proj-001','{num}','{mt}','compression','{sd}','{st}','Tech-{i+1}');")
    if mt == "concrete":
        ctid = str(uuid.uuid4())
        grade = random.choice(["C25/30","C30/37","C35/45"])
        cs28 = round(random.uniform(25,45),1)
        print(f"INSERT INTO concrete_tests (id,project_id,material_test_id,concrete_grade,compressive_strength_28d,test_date,result) VALUES ('{ctid}','proj-001','{tid}','{grade}',{cs28},'{sd}','pass');")
print("COMMIT;")