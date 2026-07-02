---
title: "[MODULE] Laboratory — Material Testing, Equipment, Sampling V029"
labels: enhancement, module, laboratory
assignees: ""
---

## Описание модуля Laboratory (V029)

Полный модуль лабораторных испытаний для OpenConstructionERP: тестирование материалов, оборудование, сертификаты, журнал отбора проб.

### Миграция БД
- ✅ `database/migrations/V029__Laboratory_Module.sql` — 7 таблиц:
  - `material_testing` — испытания материалов (concrete, steel, soil, aggregate)
  - `concrete_tests` — специализированные тесты бетона (прочность, slump, air)
  - `soil_tests` — тесты грунта (Proctor, CBR, triaxial)
  - `steel_tests` — тесты стали (tensile, bend, weld)
  - `lab_certificates` — сертификаты (материалы, калибровка)
  - `lab_equipment` — лабораторное оборудование
  - `sampling_log` — журнал отбора проб

### Go API хендлеры
- ✅ `services/core/internal/handlers/laboratory.go` — полный CRUD:
  - `GET/POST /lab/tests`, `GET/PUT /lab/tests/{id}`
  - `GET/POST /lab/concrete-tests`, `GET /lab/concrete-tests/{id}`
  - `GET/POST /lab/soil-tests`, `GET /lab/soil-tests/{id}`
  - `GET/POST /lab/steel-tests`, `GET /lab/steel-tests/{id}`
  - `GET/POST /lab/certificates`
  - `GET/POST /lab/equipment`, `GET/PUT /lab/equipment/{id}`
  - `GET/POST /lab/samples`, `GET /lab/samples/{id}`
- ✅ Зарегистрирован в `main.go`

### Модели
- ✅ `services/core/internal/models/models.go` — 7 типов (MaterialTest, ConcreteTest, SoilTest, SteelTest, LabCertificate, LabEquipment, SamplingLog)

### Генератор тестовых данных
- ✅ `scripts/generate_lab.py` — 20 испытаний, 12 бетонных тестов, 10 грунтовых, 8 стальных, 6 сертификатов, 8 единиц оборудования, 15 проб

### HTML-дашборд
- ✅ `apps/web/lab-dashboard.html` — тёмная тема, Chart.js:
  - Tests by Status (doughnut), Concrete Strength (bar)
  - Soil Types (pie), Equipment Status (doughnut)
  - Recent Tests table, Equipment table

### React-страница
- ✅ `apps/frontend/src/pages/LabPage.tsx`

### API-клиент (frontend)
- ✅ `apps/frontend/src/api.ts` — `labApi`

### Типы TypeScript
- ✅ `apps/frontend/src/types.ts` — 7 интерфейсов (MaterialTest, ConcreteTest, SoilTest, SteelTest, LabEquipment, LabCertificate)