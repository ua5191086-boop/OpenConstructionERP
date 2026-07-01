---
title: "Модуль Equipment Management (V015/E-12) — Управление оборудованием"
labels: ["epic", "module", "equipment"]
assignees: []
---

## 🏗️ Equipment Management Module (V015)

### Описание
Создать модуль управления оборудованием для OpenConstructionERP. Модуль охватывает TBM, cranes, fleet, predictive maintenance, telemetry, fuel management, операторов и запчасти.

### Требования

#### 1. SQL схема (database/migrations/V015__Equipment_Management.sql)
- [x] Таблица `equipment_categories` — категории оборудования
- [x] Таблица `equipment` — основная таблица (TBM, cranes, fleet, heavy machinery)
- [x] Таблица `equipment_maintenance` — журнал обслуживания
- [x] Таблица `maintenance_schedules` — регламентные графики ТО
- [x] Таблица `equipment_telemetry` — телеметрия (IoT sensors)
- [x] Таблица `equipment_fuel` — учёт топлива
- [x] Таблица `equipment_operators` — назначение операторов
- [x] Таблица `equipment_downtime` — учёт простоев
- [x] Таблица `equipment_spare_parts` — запчасти
- [x] Представление `equipment_summary` для агрегированной статистики
- [x] Регистрация типов объектов в `object_types`

#### 2. Генератор данных (scripts/generate_equipment.py)
- [x] Генерация тестовых данных для всех 9 таблиц
- [x] Вывод в `apps/web/equipment_data.json`
- [x] TBM с параметрами (диаметр, длина, тип грунта)
- [x] Краны различных типов (мобильные, башенные, гусеничные)
- [x] Fleet vehicles с телеметрией
- [x] Реалистичные графики обслуживания

#### 3. HTML-дашборд (apps/web/equipment-dashboard.html)
- [x] Тёмная тема #0f172a
- [x] Chart.js для графиков
- [x] Табы: Overview, Fleet, TBM, Cranes, Maintenance, Telemetry
- [x] Фильтры по проекту, типу, статусу
- [x] Статистические карточки (total, available, in use, maintenance)
- [x] Fuel consumption charts
- [x] Telemetry dashboards (IoT)

#### 4. Go API хендлеры (services/core/internal/handlers/equipment.go)
- [x] CRUD для всех 9 таблиц
- [x] Фильтрация по статусу, типу, project_id
- [x] Summary endpoint
- [x] Регистрация в основном роутере

#### 5. React страница (apps/frontend/src/pages/EquipmentPage.tsx)
- [x] Интеграция с `/api/v1/equipment/` API
- [x] Recharts для графиков
- [x] Табы для разных типов оборудования
- [x] Фильтры по статусу, типу, категории
- [x] Статистические карточки

#### 6. Интеграция
- [ ] Добавить маршрут `/equipment` в `App.tsx`
- [ ] Добавить навигацию в `Layout.tsx`
- [ ] Зарегистрировать хендлер в `main.go`

### Детали реализации
- Миграция: V015
- Owner: E-12
- Unified API prefix: `/api/v1/equipment/`
- Dashboard URL: `/apps/web/equipment-dashboard.html`

### Связанные модули
- V008 — HR (operators, employees)
- V009 — Project Management
- V013 — Document Control (maintenance records)