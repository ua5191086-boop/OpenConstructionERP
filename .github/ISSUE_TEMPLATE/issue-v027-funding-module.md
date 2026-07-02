---
title: "[MODULE] Funding — Funding Sources, Multi-Currency, Guarantees V027"
labels: enhancement, module, funding
assignees: ""
---

## Описание модуля Funding (V027)

Полный модуль финансирования проектов для OpenConstructionERP: источники финансирования, транши, выборки, ковенанты, мультивалютные курсы, хеджи, банковские гарантии и требования по ним.

### Миграция БД
- ✅ `database/migrations/V027__Funding_Module.sql` — 9 таблиц:
  - `funding_sources` — источники (банки, ECA, инвесторы, гранты)
  - `funding_tranches` — транши с датами и статусами
  - `funding_drawdowns` — выборки средств
  - `funding_covenants` — ковенанты (DSCR, LLCR, ICR)
  - `multi_currency_rates` — курсы валют
  - `currency_hedges` — хеджи (forward, swap, option)
  - `guarantees` — банковские гарантии (bid, performance, advance)
  - `guarantee_claims` — требования по гарантиям
  - `guarantee_amendments` — изменения гарантий

### Go API хендлеры
- ✅ `services/core/internal/handlers/funding.go` — полный CRUD:
  - `GET/POST /funding/sources`, `GET/PUT/DELETE /funding/sources/{id}`
  - `GET/POST /funding/tranches`, `GET /funding/tranches/{id}`
  - `GET/POST /funding/drawdowns`, `GET /funding/drawdowns/{id}`
  - `GET/POST /funding/covenants`, `GET/PUT /funding/covenants/{id}`
  - `GET/POST /funding/rates`
  - `GET/POST /funding/hedges`, `GET /funding/hedges/{id}`
  - `GET/POST /funding/guarantees`, `GET/PUT /funding/guarantees/{id}`
  - `GET/POST /funding/guarantees/{id}/claims`
  - `GET/POST /funding/guarantees/{id}/amendments`
- ✅ Зарегистрирован в `main.go`

### Модели
- ✅ `services/core/internal/models/models.go` — 14 типов (FundingSource, FundingTranche, FundingDrawdown, FundingCovenant, MultiCurrencyRate, CurrencyHedge, Guarantee, GuaranteeClaim, GuaranteeAmendment)

### Генератор тестовых данных
- ✅ `scripts/generate_funding.py` — 8 источников, 24 транша, 32 выборки, 16 ковенантов, 6 курсов, 4 хеджа, 5 гарантий, 10 требований, 10 изменений

### HTML-дашборд
- ✅ `apps/web/funding-dashboard.html` — тёмная тема, Chart.js:
  - Funding by Source Type (doughnut)
  - Covenant Compliance (pie)
  - Guarantee Status (doughnut)
  - Drawdown Timeline (bar)
  - Funding Sources table
  - Guarantees table

### React-страница
- ✅ `apps/frontend/src/pages/FundingPage.tsx` — тёмная тема, KPI cards, таблицы

### API-клиент (frontend)
- ✅ `apps/frontend/src/api.ts` — `fundingApi` с полным набором методов

### Типы TypeScript
- ✅ `apps/frontend/src/types.ts` — 10 интерфейсов (FundingSource, FundingTranche, FundingDrawdown, FundingCovenant, MultiCurrencyRate, CurrencyHedge, Guarantee, GuaranteeClaim)