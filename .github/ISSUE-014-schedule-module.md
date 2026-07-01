---
title: "Модуль Schedule Management (V014/S-11) — Управление расписанием"
labels: ["epic", "module", "schedule"]
assignees: []
---

## 📅 Schedule Management Module (V014)

### Описание
Создать модуль управления расписанием для OpenConstructionERP. Модуль охватывает P6-совместимое управление расписаниями, CPM analysis, critical path, resource loading, Gantt и baseline management.

### Требования

#### 1. SQL схема (database/migrations/V014__Schedule_Management.sql)
- [x] Таблица `schedules` — основные расписания проектов (baseline/target/current/what_if)
- [x] Таблица `schedule_activities` — активности с CPM-атрибутами (ES/EF/LS/LF, floats, critical)
- [x] Таблица `schedule_relationships` — связи предшествования (FS/SS/FF/SF + lag)
- [x] Таблица `schedule_resources` — ресурсная загрузка (labor/material/equipment/cost)
- [x] Таблица `schedule_baselines` — управление базовыми планами
- [x] Таблица `schedule_changes` — изменения расписания
- [x] Таблица `critical_path_log` — история расчётов CPM
- [x] Представление `schedule_summary` для агрегированной статистики
- [x] Регистрация типов объектов в `object_types`

#### 2. Генератор данных (scripts/generate_schedule.py)
- [x] Генерация тестовых данных для всех 7 таблиц
- [x] Вывод в `apps/web/schedule_data.json`
- [x] P6-совместимые activity IDs, WBS коды
- [x] Реалистичные CPM-расчёты (ES/EF/LS/LF, floats)
- [x] Критический путь с несколькими параллельными треками

#### 3. HTML-дашборд (apps/web/schedule-dashboard.html)
- [x] Тёмная тема #0f172a
- [x] Chart.js для графиков (Gantt-like, status, resource histogram)
- [x] Табы: Overview, Gantt, Critical Path, Resources, Baselines, Changes
- [x] Фильтры по проекту, статусу, типу расписания
- [x] Статистические карточки (total activities, critical, float)
- [x] Resource loading histogram

#### 4. Go API хендлеры (services/core/internal/handlers/schedule.go)
- [x] CRUD для всех 7 таблиц
- [x] Фильтрация по статусу, schedule_id, project_id
- [x] Summary endpoint для агрегированной статистики
- [x] Регистрация в основном роутере

#### 5. React страница (apps/frontend/src/pages/SchedulePage.tsx)
- [x] Интеграция с `/api/v1/schedule/` API
- [x] Recharts для графиков
- [x] Табы: Activities, Gantt, Critical Path, Resources
- [x] Фильтры по статусу, типу, поиску
- [x] Статистические карточки

#### 6. Интеграция
- [ ] Добавить маршрут `/schedule` в `App.tsx`
- [ ] Добавить навигацию в `Layout.tsx`
- [ ] Зарегистрировать хендлер в `main.go`

### Детали реализации
- Миграция: V014
- Owner: S-11
- Unified API prefix: `/api/v1/schedule/`
- Dashboard URL: `/apps/web/schedule-dashboard.html`

### Связанные модули
- V009 — Project Management (проекты, WBS)
- V013 — Document Control (method statements)