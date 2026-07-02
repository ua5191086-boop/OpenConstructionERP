---
title: "[MODULE] Fleet — Vehicles, Drivers, Fuel, Accidents V032"
labels: enhancement, module, fleet
assignees: ""
---

## Описание модуля Fleet (V032)

Полный модуль управления автопарком для OpenConstructionERP: транспортные средства, водители, топливо, ТО, GPS-трекинг, аварии, телематика.

### Миграция БД
- ✅ `database/migrations/V032__Fleet_Module.sql` — 7 таблиц:
  - `fleet_vehicles` — транспортные средства
  - `fleet_drivers` — водители
  - `fleet_fuel` — заправки
  - `fleet_maintenance` — техническое обслуживание
  - `fleet_tracking` — GPS-трекинг
  - `fleet_accidents` — ДТП
  - `fleet_telematics` — телематика (CAN bus)

### Go API хендлеры
- ✅ `services/core/internal/handlers/fleet.go` — полный CRUD:
  - `GET/POST /fleet/vehicles`, `GET/PUT /fleet/vehicles/{id}`
  - `GET/POST /fleet/drivers`, `GET /fleet/drivers/{id}`
  - `GET/POST /fleet/fuel`
  - `GET/POST /fleet/maintenance`, `PUT /fleet/maintenance/{id}`
  - `GET/POST /fleet/accidents`, `PUT /fleet/accidents/{id}`
  - `GET/POST /fleet/tracking`
  - `GET/POST /fleet/telematics/{vehicleId}`
- ✅ Зарегистрирован в `main.go`

### Модели
- ✅ `services/core/internal/models/models.go` — 7 типов (FleetVehicle, FleetDriver, FleetFuel, FleetMaintenance, FleetTracking, FleetAccident, FleetTelematics)

### Генератор тестовых данных
- ✅ `scripts/generate_fleet.py` — 15 ТС, 10 водителей, 40 заправок, 20 ТО, 50 треков, 6 аварий, 50 телеметрий

### HTML-дашборд
- ✅ `apps/web/fleet-dashboard.html` — тёмная тема, Chart.js:
  - Vehicles by Type (doughnut), Fuel Cost by Month (bar)
  - Maintenance by Type (pie), Accident Severity (doughnut)
  - Fleet Vehicles table, Drivers table

### React-страница
- ✅ `apps/frontend/src/pages/FleetPage.tsx`

### API-клиент (frontend)
- ✅ `apps/frontend/src/api.ts` — `fleetApi`

### Типы TypeScript
- ✅ `apps/frontend/src/types.ts` — 7 интерфейсов (FleetVehicle, FleetDriver, FleetFuel, FleetMaintenance, FleetTracking, FleetAccident, FleetTelematics)