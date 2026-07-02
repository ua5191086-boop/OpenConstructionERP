---
title: "[MODULE] Permits — Applications, Inspections, Conditions V030"
labels: enhancement, module, permits
assignees: ""
---

## Описание модуля Permits (V030)

Полный модуль разрешительной документации для OpenConstructionERP: заявки, разрешения, инспекции, продления, условия, регулирующие органы.

### Миграция БД
- ✅ `database/migrations/V030__Permits_Module.sql` — 6 таблиц:
  - `regulatory_bodies` — регулирующие органы (GASK, СЭС, экология)
  - `permit_applications` — заявки на разрешения
  - `permit_documents` — прилагаемые документы
  - `permit_inspections` — инспекции
  - `permit_renewals` — продления разрешений
  - `permit_conditions` — условия и обременения

### Go API хендлеры
- ✅ `services/core/internal/handlers/permits.go` — полный CRUD:
  - `GET/POST /permits/bodies`, `GET /permits/bodies/{id}`
  - `GET/POST /permits/applications`, `GET/PUT /permits/applications/{id}`
  - `GET/POST /permits/applications/{id}/documents`
  - `GET/POST /permits/applications/{id}/inspections`, `PUT /permits/inspections/{id}`
  - `GET/POST /permits/applications/{id}/renewals`
  - `GET/POST /permits/applications/{id}/conditions`, `PUT /permits/conditions/{id}`
- ✅ Зарегистрирован в `main.go`

### Модели
- ✅ `services/core/internal/models/models.go` — 6 типов (RegulatoryBody, PermitApplication, PermitDocument, PermitInspection, PermitRenewal, PermitCondition)

### Генератор тестовых данных
- ✅ `scripts/generate_permits.py` — 6 органов, 15 заявок, 25 документов, 12 инспекций, 8 продлений, 20 условий

### HTML-дашборд
- ✅ `apps/web/permits-dashboard.html` — тёмная тема, Chart.js:
  - Applications by Status (doughnut), by Type (bar)
  - Inspection Results (pie), Condition Compliance (doughnut)
  - Applications table, Inspections table

### React-страница
- ✅ `apps/frontend/src/pages/PermitsPage.tsx`

### API-клиент (frontend)
- ✅ `apps/frontend/src/api.ts` — `permitsApi`

### Типы TypeScript
- ✅ `apps/frontend/src/types.ts` — 6 интерфейсов (RegulatoryBody, PermitApplication, PermitDocument, PermitInspection, PermitCondition)