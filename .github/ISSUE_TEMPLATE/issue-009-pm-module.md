---
title: "[MODULE] Project Management — V009 full module (PM)"
labels: enhancement, module
assignees: ""
---

## Описание модуля Project Management

Полный модуль управления проектами для OpenConstructionERP. Реализует портфельное управление, WBS, вехи, риски, изменения и уроки.

### Миграция БД
- ✅ `V009__Project_Management.sql` — 10 таблиц (projects, wbs_items, project_milestones, project_phases, project_team, project_portfolio, portfolio_projects, project_risks, project_changes, project_lessons)

### Генератор данных
- ✅ `scripts/generate_pm.py` — генерирует 5 проектов с WBS (68 элементов), 40 milestones, 25 фаз, 30 рисков, 15 изменений, 15 уроков, 3 портфеля
- ✅ Вывод: `apps/web/pm_data.json`

### HTML-дашборд
- ✅ `apps/web/pm-dashboard.html` — тёмная тема (#0f172a), Chart.js, фильтры (портфель, статус, тип, поиск), аккордеоны для WBS-дерева, Gantt-like расписания, матрицы рисков, таблицы изменений
- ✅ Портфельный обзор с карточками проектов и прогресс-барами
- ✅ WBS-дерево с коллапсом/экспандом
- ✅ Матрица рисков 5×5 (Probability × Impact)
- ✅ Gantt-like отображение фаз проекта

### Go API хендлеры
- ✅ `services/core/internal/handlers/pm.go` — полный CRUD для всех 10 таблиц
- ✅ Роуты: `/api/v1/pm/projects`, `/pm/wbs-items`, `/pm/milestones`, `/pm/phases`, `/pm/team`, `/pm/portfolios`, `/pm/portfolio-projects`, `/pm/risks`, `/pm/changes`, `/pm/lessons`, `/pm/dashboard`
- ✅ Зарегистрирован в `main.go`

### React страница
- ✅ `apps/frontend/src/pages/PMProjectPage.tsx` — тёмная тема, Recharts, фильтры, карточки проектов с раскрывающимся WBS-деревом, таблицы изменений и уроков
- ✅ Добавлен маршрут `/pm` в `App.tsx`
- ✅ Добавлен пункт навигации в `Layout.tsx`
- ✅ API-клиент в `api.ts`
- ✅ TypeScript типы в `types.ts`

### Файлы для коммита
```
scripts/generate_pm.py
apps/web/pm-dashboard.html
apps/web/pm_data.json
services/core/internal/handlers/pm.go
services/core/cmd/api/main.go (изменён)
apps/frontend/src/pages/PMProjectPage.tsx
apps/frontend/src/components/Layout.tsx (изменён)
apps/frontend/src/App.tsx (изменён)
apps/frontend/src/api.ts (изменён)
apps/frontend/src/types.ts (изменён)
```
