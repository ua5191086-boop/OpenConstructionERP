# Issue #9: Project Management Module — V009 Full Module

## Описание
Полный модуль управления проектами (Project Management) для OpenConstructionERP. Включает генератор тестовых данных, HTML-дашборд, Go API хендлеры и React страницу.

## Состав

### 1. Миграция БД
- `database/migrations/V009__Project_Management.sql` — 10 таблиц (projects, wbs_items, project_milestones, project_phases, project_team, project_portfolio, portfolio_projects, project_risks, project_changes, project_lessons)

### 2. Генератор данных
- `scripts/generate_pm.py` — генерирует 5 проектов с WBS (68 элементов), 40 milestones, 25 фаз, 30 рисков, 15 изменений, 15 уроков, 3 портфеля
- Вывод: `apps/web/pm_data.json`

### 3. HTML-дашборд
- `apps/web/pm-dashboard.html` — тёмная тема (#0f172a), Chart.js, фильтры (портфель, статус, тип, поиск), аккордеоны для WBS-дерева, Gantt-like расписания, матрицы рисков, таблицы изменений

### 4. Go API хендлеры
- `services/core/internal/handlers/pm.go` — полный CRUD для всех 10 таблиц
- Роуты: `/api/v1/pm/projects`, `/pm/wbs-items`, `/pm/milestones`, `/pm/phases`, `/pm/team`, `/pm/portfolios`, `/pm/portfolio-projects`, `/pm/risks`, `/pm/changes`, `/pm/lessons`, `/pm/dashboard`
- Зарегистрирован в `main.go`

### 5. React страница
- `apps/frontend/src/pages/PMProjectPage.tsx` — тёмная тема, Recharts, фильтры, карточки проектов с раскрывающимся WBS-деревом, таблицы изменений и уроков
- Маршрут `/pm` в `App.tsx`
- Пункт навигации в `Layout.tsx`
- API-клиент в `api.ts`
- TypeScript типы в `types.ts`

## Проверка
```bash
# 1. Сгенерировать данные
python3 scripts/generate_pm.py

# 2. Открыть дашборд
# Открой apps/web/pm-dashboard.html в браузере

# 3. Запустить Go API
cd services/core && go run cmd/api/main.go

# 4. Запустить React frontend
cd apps/frontend && npm run dev
# Открой http://localhost:5173/pm
```
