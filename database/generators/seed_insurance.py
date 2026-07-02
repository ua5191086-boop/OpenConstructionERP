#!/usr/bin/env python3
"""Generate sample data for Insurance module (V031)."""
import uuid, random
from datetime import datetime, timedelta

policy_types = ["constructor_all_risk", "third_party_liability", "professional_indemnity", "workers_comp", "motor"]
insurers = ["AXA XL", "Zurich Insurance", "Allianz Global", "Munich Re", "Chubb"]

print("-- Insurance Module Seed Data (V031)")
print("BEGIN;")
# Brokers
for name in ["Marsh & McLennan", "Aon Risk Solutions", "Willis Towers Watson"]:
    bid = str(uuid.uuid4())
    print(f"INSERT INTO insurance_brokers (id,broker_name) VALUES ('{bid}','{name}');")

# Policies
for i in range(6):
    pid = str(uuid.uuid4())
    pt = random.choice(policy_types)
    ins = random.choice(insurers)
    num = f"POL-{2024}-{1000+i}"
    start = (datetime.now() - timedelta(days=random.randint(1,365))).date()
    end = start + timedelta(days=365)
    si = round(random.uniform(500000, 50000000), 2)
    print(f"INSERT INTO insurance_policies (id,project_id,policy_number,policy_type,insurer,sum_insured,currency,start_date,end_date,status) VALUES ('{pid}','proj-001','{num}','{pt}','{ins}',{si},'USD','{start}','{end}','active');")
    # Coverage
    for j in range(2):
        cid = str(uuid.uuid4())
        cl = round(si * random.uniform(0.3, 0.8), 2)
        print(f"INSERT INTO insurance_coverage (id,policy_id,coverage_type,coverage_limit,currency) VALUES ('{cid}','{pid}','Coverage-{j+1}',{cl},'USD');")
print("COMMIT;")