# Правила параллельной работы (WORKSTREAMS)

Несколько сессий/разработчиков пушат в этот репозиторий параллельно.
Два инцидента уже случились: гонка push в main и коллизия номера миграции
(две V002; тендерная V002 к тому же не устанавливалась — FK на несуществующие
`contractors`/`sections`, BIGINT против UUID). Чтобы это не повторялось:

## 1. Реестр номеров миграций (единственный источник истины)

| № | Модуль | Статус | Владелец |
|---|--------|--------|----------|
| V000 | Core Foundation (projects, organizations, users, contracts) | applied | core |
| V001 | BOQ Module | applied | core |
| V002 | Ontology Core + regional coefficients | applied | core |
| V003 | Tender Module (перенумерована из V002, FK/UUID исправлены) | applied | tenders |
| V004 | Tunnel Module (tbm, drives, rings, segments) | applied | tunnel |
| V005 | Site Documents (RFI, daily reports, work entries) | applied | core-py |
| V006 | CDE Core (documents, numbering rules, revisions, transmittals) | applied | core-py |
| V007 | Contract Module (перенумерована из V003-дубля; contracts из V000 расширена ALTER'ами) | applied | contracts |
| V008 | HR Module (перенумерована из V004-дубля) | applied | hr |
| V009 | Finance Module (перенумерована из V005-дубля) | applied | finance |
| V010 | Procurement Module (перенумерована из V006-дубля) | applied | procurement |
| V011 | BIM Module (перенумерована из V007) | applied | bim |
| V012 | AI Module (перенумерована из V008) | applied | ai |
| V013 | — СЛЕДУЮЩИЙ СВОБОДНЫЙ — перед использованием добавь строку сюда в ТОМ ЖЕ коммите | | |

> ⚠️ **2026-07-02, второй инцидент за день:** четыре новых дубля номеров (V003/V004/V005/V006
> залиты повторно параллельными сессиями) + BIGSERIAL/BIGINT против UUID + FK на contractors/sections
> + повторное создание таблицы contracts, существующей с V000. Всё исправлено, но теперь
> **CI автоматически падает при дублях номеров** — договариваться больше не нужно, машина не пустит.

**Правило:** новый номер бронируется строкой в этой таблице в том же коммите,
что и сама миграция. Перед началом работы — `git pull`. Нашёл занятый номер — перенумеруй свою.

## 2. Конвенции схемы (нарушение = красный CI)

- PK: **UUID** `DEFAULT gen_random_uuid()`. Никаких BIGSERIAL.
- FK только на реально существующие таблицы: `organizations` (не contractors), `boq_sections` (не sections).
- Деньги: `NUMERIC(18,2)` + `currency CHAR(3)`.
- CI ставит всю цепочку V000..Vnnn на чистый PostgreSQL при каждом PR — миграция обязана устанавливаться с нуля.

## 3. Полосы (lanes)

| Полоса | Область | Где живёт |
|--------|---------|-----------|
| core-py | Python reference implementation: онтология, BOQ, тоннель, финансы, API + дашборд | `services/core-py`, `database/migrations` |
| core-go | Перенос стабилизированных vertical'ей на Go (ADR-003) | `services/core` |
| tenders | Тендерный модуль | `database/migrations/V003`, `apps/web/tender-dashboard.html`, `scripts/generate_tenders.py` |
| infra | Compose/K8s/nginx | `infrastructure/` |

**Compose:** источник истины — `docker-compose.dev.yml` (разработка).
`docker-compose.single-node.yml` — прод-профиль; при изменении dev-файла — синхронизировать.

## 4. Git-дисциплина

- Длинные работы — в feature-ветке, merge в main мелкими порциями.
- Перед push всегда `git pull --rebase`.
- Не коммитить артефакты (`output/`, `__pycache__` — уже в .gitignore).
