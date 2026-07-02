#!/usr/bin/env python3
"""Generate sample data for Fleet module (V032)."""
import uuid, random
from datetime import datetime, timedelta

vehicle_types = ["truck", "crane", "excavator", "dozer", "loader", "pickup"]
makes = ["CAT", "Komatsu", "Hitachi", "Volvo", "Scania"]
statuses = ["operational", "under_maintenance", "out_of_service"]

print("-- Fleet Module Seed Data (V032)")
print("BEGIN;")
# Drivers
drivers = []
for name in ["Ivan Petrov", "Sergei Ivanov", "Alexei Smirnov", "Dmitri Volkov"]:
    did = str(uuid.uuid4())
    drivers.append(did)
    print(f"INSERT INTO vehicle_drivers (id,project_id,driver_name,license_number,license_type,status) VALUES ('{did}','proj-001','{name}','LIC-{random.randint(10000,99999)}','{random.choice(['A','B','C'])}','active');")

# Vehicles
for i in range(8):
    vid = str(uuid.uuid4())
    vt = random.choice(vehicle_types)
    mk = random.choice(makes)
    plate = f"{random.choice(['AB','CD','EF'])}-{random.randint(100,999)}"
    st = random.choice(statuses)
    driver = random.choice(drivers) if random.random() > 0.3 else None
    driver_sql = f"'{driver}'" if driver else "NULL"
    print(f"INSERT INTO fleet_vehicles (id,project_id,vehicle_type,make,model,year,license_plate,fuel_type,status,assigned_driver,mileage_km) VALUES ('{vid}','proj-001','{vt}','{mk}','{mk}-{vt}-{i}',{2020+i},'{plate}','diesel','{st}',{driver_sql},{random.randint(5000,80000)});")
    # Fuel records
    for j in range(5):
        fid = str(uuid.uuid4())
        fd = (datetime.now() - timedelta(days=random.randint(1,30))).date()
        qty = round(random.uniform(20, 200), 1)
        price = round(random.uniform(1.2, 1.8), 3)
        print(f"INSERT INTO vehicle_fuel (id,project_id,vehicle_id,fuel_date,fuel_type,quantity_liters,unit_price,total_cost) VALUES ('{fid}','proj-001','{vid}','{fd}','diesel',{qty},{price},{round(qty*price,2)});")
print("COMMIT;")