#!/usr/bin/env python3
"""Generate sample data for Funding module (V027)."""
import uuid, random, json
from datetime import datetime, timedelta

sources = [
    {"type": "bank", "name": "European Investment Bank", "amount": 500000000},
    {"type": "eca", "name": "Export-Import Bank USA", "amount": 200000000},
    {"type": "investor", "name": "Global Infrastructure Partners", "amount": 150000000},
    {"type": "grant", "name": "EU Horizon Grant", "amount": 50000000},
]
print("-- Funding Module Seed Data (V027)")
print("BEGIN;")
for s in sources:
    sid = str(uuid.uuid4())
    print(f"INSERT INTO funding_sources (id,project_id,source_type,source_name,commitment_amount,currency,status) VALUES ('{sid}','proj-001','{s['type']}','{s['name']}',{s['amount']},'USD','active');")
    for i in range(3):
        tid = str(uuid.uuid4())
        amt = s['amount'] // 3
        d = datetime.now() + timedelta(days=90*i)
        print(f"INSERT INTO funding_tranches (id,project_id,funding_source_id,tranche_name,amount,currency,expected_date,status) VALUES ('{tid}','proj-001','{sid}','Tranche {i+1}',{amt},'USD','{d.date()}','planned');")
print("COMMIT;")