---
title: "[MODULE] Primavera P6 Connector V026"
labels: enhancement, module, p6, integration
assignees: ""
---

## Описание модуля Primavera P6 Connector

Модуль интеграции с Oracle Primavera P6 EPPM. Реализует маппинг проектов, WBS, активностей, ресурсов и связей между P6 и OpenConstructionERP. Поддержка импорта XER и отслеживание синхронизации.

### Миграция БД
- ✅ `database/migrations/V026__Primavera_P6_Connector.sql` — 6 таблиц:
  - `p6_projects` — маппинг проектов P6 ↔ OCE
  - `p6_wbs` — маппинг WBS
  - `p6_activities` — маппинг активностей
  - `p6_resources` — маппинг ресурсов
  - `p6_relationships` — связи predecessor/successor
  - `p6_sync_log` — лог синхронизации
- ✅ Функция `p6_find_local_entity()` — поиск OCE-сущности по P6 ID

### Go API хендлеры
- ✅ `services/core/internal/handlers/p6.go` — полный CRUD + импорт:
  - `POST /api/v1/p6/import` — импорт из P6 XER
  - `GET /api/v1/p6/projects` — список связанных проектов
  - `GET /api/v1/p6/sync/status` — статус синхронизации
  - `POST /api/v1/p6/sync/{id}` — запуск синхронизации
  - CRUD для WBS, Activities, Relationships, Resources
  - Лог синхронизации
- ✅ Зарегистрирован в `main.go`

### Модели
- ✅ `services/core/internal/models/models.go` — 7 новых типов (P6Project, P6WBS, P6Activity, P6Relationship, P6Resource, P6SyncLog)

### HTML-дашборд
- ✅ `apps/web/p6-dashboard.html` — тёмная тема, Chart.js:
  - Обзор: статус активностей (doughnut chart), прогресс по WBS (bar chart)
  - Таблица связанных проектов P6
  - Таблица последних синхронизаций
  - Таблица активностей с фильтром по статусу
  - Таблица WBS с иерархией
  - Полный лог синхронизации
  - Вкладки: Обзор, Активности, WBS, Синхронизация
  - Кнопка импорта XER (демо)

### Docker
- ✅ Docker-образ не требуется (работает через core API)