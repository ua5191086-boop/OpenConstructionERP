---
title: "Модуль Document Control (V013/D-10) — Управление документацией"
labels: ["epic", "module", "document-control"]
assignees: []
---

## 📄 Document Control Module (V013)

### Описание
Создать модуль управления документацией для OpenConstructionERP. Модуль охватывает все аспекты документооборота строительного проекта: RFI, NCR, Submittals, Method Statements, Shop Drawings, Correspondence, Minutes of Meeting, Daily Reports, Document Transmittals и Document Revisions.

### Требования

#### 1. SQL схема (database/migrations/V013__Document_Control.sql)
- [x] Таблица `rfi_documents` — Requests for Information
- [x] Таблица `ncr_documents` — Non-Conformance Reports
- [x] Таблица `submittals` — подача материалов на утверждение
- [x] Таблица `method_statements` — методы производства работ
- [x] Таблица `shop_drawings` — деталировочные чертежи
- [x] Таблица `correspondence` — переписка (входящая/исходящая)
- [x] Таблица `minutes_of_meeting` — протоколы совещаний
- [x] Таблица `doc_daily_reports` — ежедневные отчёты
- [x] Таблица `document_transmittals` — сопроводительные письма
- [x] Таблица `document_revisions` — история версий
- [x] Представление `doc_control_summary` для агрегированной статистики
- [x] Регистрация типов объектов в `object_types`

#### 2. Генератор данных (scripts/generate_doc_control.py)
- [x] Генерация тестовых данных для всех 10 таблиц
- [x] Вывод в `apps/web/doc_control_data.json`
- [x] Распределение по 5 проектам
- [x] Реалистичные статусы, приоритеты, дисциплины

#### 3. HTML-дашборд (apps/web/doc-control-dashboard.html)
- [x] Тёмная тема #0f172a
- [x] Chart.js для графиков (статус, тренды, типы)
- [x] Табы для переключения между типами документов
- [x] Фильтры по статусу, типу, проекту
- [x] Поиск по коду/названию
- [x] Статистические карточки

#### 4. Go API хендлеры (services/core/internal/handlers/doc_control.go)
- [x] CRUD для всех 10 таблиц (50+ endpoint'ов)
- [x] Фильтрация по статусу, project_id, типу
- [x] Summary endpoint для агрегированной статистики
- [x] Регистрация в основном роутере

#### 5. React страница (apps/frontend/src/pages/DocControlPage.tsx)
- [x] Интеграция с `/api/v1/doc-control/` API
- [x] Recharts для графиков (BarChart, PieChart)
- [x] Табы для переключения типов документов
- [x] Фильтры по статусу, типу, поиску
- [x] Статистические карточки
- [x] Таблица с данными

#### 6. Интеграция
- [ ] Добавить маршрут `/doc-control` в `App.tsx`
- [ ] Добавить навигацию в `Layout.tsx`
- [ ] Зарегистрировать хендлер в `main.go`

### Детали реализации
- Миграция: V013 (V010 занят Procurement)
- Owner: D-10
- Unified API prefix: `/api/v1/doc-control/`
- Dashboard URL: `/apps/web/doc-control-dashboard.html`

### Связанные модули
- V005 — Site Documents (частично пересекается с daily_reports/rfi)
- V006 — CDE Core (документооборот)