---
title: "[MODULE] EVM — Earned Value Management V025"
labels: enhancement, module, evm
assignees: ""
---

## Описание модуля Earned Value Management (EVM)

Полный модуль освоенного объёма (EVM) для OpenConstructionERP. Соответствует ANSI/EIA-748. Реализует контрольные счета, baseline, периоды отчётов, фактики, расчёт метрик (PV, EV, AC, SV, CV, SPI, CPI, EAC, ETC, TCPI), прогнозы, правила освоения и S-кривые.

### Миграция БД
- ✅ `database/migrations/V025__EVM_Module.sql` — 8 таблиц:
  - `evm_control_accounts` — контрольные счета
  - `evm_baselines` — базовые планы (P6-compatible)
  - `evm_periods` — плановые показатели по периодам
  - `evm_actuals` — фактические данные (ACWP, hours, progress)
  - `evm_metrics` — расчётные метрики (PV, EV, AC, SV, CV, SPI, CPI, EAC, ETC, TCPI)
  - `evm_forecasts` — прогнозы (EAC, ETC, VAC)
  - `evm_earned_rules` — правила освоения (0/100, 50/50, %complete, physical)
  - `evm_projects` — привязка EVM к проекту
- ✅ Функция `calculate_evm_metrics()` — авторасчёт метрик через SELECT

### Go API хендлеры
- ✅ `services/core/internal/handlers/evm.go` — полный CRUD + аналитические endpoints:
  - `GET /api/v1/evm/projects/{id}/summary` — сводка EVM
  - `GET /api/v1/evm/projects/{id}/metrics` — все метрики
  - `GET /api/v1/evm/projects/{id}/curve` — S-кривая (PV/EV/AC)
  - `GET /api/v1/evm/projects/{id}/forecast` — прогноз завершения
  - CRUD для control-accounts, baselines, periods, actuals, metrics, forecasts, rules
- ✅ Зарегистрирован в `main.go`

### Модели
- ✅ `services/core/internal/models/models.go` — 10 новых типов (EVMControlAccount, EVMBaseline, EVMPeriod, EVMActual, EVMMetric, EVMForecast, EVMRule, EVMProject)

### HTML-дашборд
- ✅ `apps/web/evm-dashboard.html` — тёмная тема, Chart.js:
  - S-кривая (PV, EV, AC) — line chart
  - SPI/CPI тренд — line chart
  - Ключевые метрики: BAC, PV, EV, AC, SV, CV, SPI, CPI, EAC, TCPI
  - Таблица контрольных счетов
  - Таблица периодов с цветовой индикацией
  - Выбор проекта, демо-данные

### Docker
- ✅ Docker-образ не требуется (работает через core API)