---
title: "Модуль HSE (V016/H-13) — Health, Safety & Environment"
labels: ["epic", "module", "hse"]
assignees: []
---

## 🛡️ HSE Module (V016)

### Описание
Создать модуль HSE (Health, Safety & Environment) для OpenConstructionERP. Модуль охватывает инциденты, наряды-допуски, аудиты, инспекции, обучение, СИЗ, тренировки, статистику, планы ЧС и учёт химикатов.

### Требования

#### 1. SQL схема (database/migrations/V016__HSE_Module.sql)
- [x] Таблица `hse_incidents` — расширенная регистрация происшествий (10+ типов, 5 уровней severity)
- [x] Таблица `hse_permits` — наряды-допуски (13 типов)
- [x] Таблица `hse_audits` — аудиты безопасности
- [x] Таблица `hse_inspections` — инспекции
- [x] Таблица `hse_training` — обучение
- [x] Таблица `hse_ppe` — СИЗ
- [x] Таблица `hse_drill` — учебные тревоги
- [x] Таблица `hse_statistics` — сводная статистика (LTIF, TRIR)
- [x] Таблица `hse_emergency_plans` — планы ЧС
- [x] Таблица `hse_chemicals` — опасные вещества
- [x] Представление `hse_summary` для агрегированной статистики
- [x] Регистрация типов объектов в `object_types`

#### 2. Генератор данных (scripts/generate_hse.py)
- [x] Генерация тестовых данных для всех 10 таблиц
- [x] Вывод в `apps/web/hse_data.json`
- [x] Реалистичные инциденты с расследованиями
- [x] PTW (Permit to Work) с разными статусами
- [x] Статистика LTIF, TRIR по месяцам

#### 3. HTML-дашборд (apps/web/hse-dashboard.html)
- [x] Тёмная тема #0f172a
- [x] Chart.js для графиков
- [x] Табы: Incidents, Permits, Audits & Inspections, Training, PPE, Stats, Emergency
- [x] Фильтры по проекту, типу, статусу, severity
- [x] Статистические карточки (open incidents, LTIF, active permits)
- [x] Incident severity matrix
- [x] Safety stats (TRIR, LTIF trends)

#### 4. Go API хендлеры (services/core/internal/handlers/hse.go)
- [x] CRUD для всех 10 таблиц
- [x] Фильтрация по статусу, типу, severity, project_id
- [x] Summary endpoint
- [x] Регистрация в основном роутере

#### 5. React страница (apps/frontend/src/pages/HSEPage.tsx)
- [x] Интеграция с `/api/v1/hse/` API
- [x] Recharts для графиков
- [x] Табы для разных типов
- [x] Фильтры
- [x] Статистические карточки

#### 6. Интеграция
- [ ] Добавить маршрут `/hse` в `App.tsx`
- [ ] Добавить навигацию в `Layout.tsx`
- [ ] Зарегистрировать хендлер в `main.go`

### Детали реализации
- Миграция: V016
- Owner: H-13
- Unified API prefix: `/api/v1/hse/`
- Dashboard URL: `/apps/web/hse-dashboard.html`

### Связанные модули
- V008 — HR (training, employees)
- V013 — Document Control (method statements with HSE aspects)
- V014 — Schedule (permit validity linked to schedule)