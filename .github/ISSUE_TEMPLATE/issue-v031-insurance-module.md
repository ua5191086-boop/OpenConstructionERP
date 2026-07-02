---
title: "[MODULE] Insurance — Policies, Claims, Brokers V031"
labels: enhancement, module, insurance
assignees: ""
---

## Описание модуля Insurance (V031)

Полный модуль страхования для OpenConstructionERP: брокеры, полисы, покрытия, премии, страховые случаи, сертификаты.

### Миграция БД
- ✅ `database/migrations/V031__Insurance_Module.sql` — 6 таблиц:
  - `insurance_brokers` — страховые брокеры
  - `insurance_policies` — полисы (CAR, TPL, PI, EL, etc.)
  - `insurance_coverage` — виды покрытия
  - `insurance_premiums` — премии (платежи по полисам)
  - `insurance_claims` — страховые случаи
  - `certificates_of_insurance` — сертификаты

### Go API хендлеры
- ✅ `services/core/internal/handlers/insurance.go` — полный CRUD:
  - `GET/POST /insurance/brokers`, `GET /insurance/brokers/{id}`
  - `GET/POST /insurance/policies`, `GET/PUT /insurance/policies/{id}`
  - `GET/POST /insurance/policies/{id}/coverage`
  - `GET/POST /insurance/policies/{id}/premiums`
  - `GET/POST /insurance/claims`, `GET/PUT /insurance/claims/{id}`
  - `GET/POST /insurance/policies/{id}/certificates`
- ✅ Зарегистрирован в `main.go`

### Модели
- ✅ `services/core/internal/models/models.go` — 6 типов (InsuranceBroker, InsurancePolicy, InsuranceCoverage, InsurancePremium, InsuranceClaim, CertificateOfInsurance)

### Генератор тестовых данных
- ✅ `scripts/generate_insurance.py` — 5 брокеров, 12 полисов, 36 покрытий, 36 премий, 8 страховых случаев, 6 сертификатов

### HTML-дашборд
- ✅ `apps/web/insurance-dashboard.html` — тёмная тема, Chart.js:
  - Policies by Status (doughnut), Claims by Status (pie)
  - Premiums Paid vs Due (bar), Policies by Type (bar)
  - Active Policies table, Recent Claims table

### React-страница
- ✅ `apps/frontend/src/pages/InsurancePage.tsx`

### API-клиент (frontend)
- ✅ `apps/frontend/src/api.ts` — `insuranceApi`

### Типы TypeScript
- ✅ `apps/frontend/src/types.ts` — 7 интерфейсов (InsuranceBroker, InsurancePolicy, InsuranceCoverage, InsurancePremium, InsuranceClaim, CertificateOfInsurance)