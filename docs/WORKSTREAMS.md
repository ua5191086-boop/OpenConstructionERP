# Правила параллельной работы (WORKSTREAMS)

Несколько сессий/разработчиков пушат в этот репозиторий параллельно.
Два инцидента уже случились: гонка push в main и коллизия номера миграции
(две V002; тендерная V002 к тому же не устанавливалась — FK на несуществующие
`contractors`/`sections`, BIGINT против UUID). Чтобы это не повторялось:

## 1. Реестр номеров миграций (единственный источник истины)

Пересобран по факту 02.07 после массовой заливки V014–V026 мимо реестра.
**Проверка реальности:** полная цепочка V000–V028 установлена на чистый PostgreSQL — 207 таблиц, зелёная.

| № | Модуль | Статус | Владелец |
|---|--------|--------|----------|
| V000 | Core Foundation | applied | — |
| V001 | BOQ Module | applied | — |
| V002 | Ontology Core | applied | — |
| V003 | Tender Module | applied | — |
| V004 | Tunnel Module | applied | — |
| V005 | Site Documents | applied | — |
| V006 | CDE Core | applied | — |
| V007 | Contract Module | applied | — |
| V008 | HR Module | applied | — |
| V009 | Finance Module | applied | — |
| V010 | Procurement Module | applied | — |
| V011 | BIM Module | applied | — |
| V012 | AI Module | applied | — |
| V013 | Document Control | applied | — |
| V014 | Schedule Management | applied | — |
| V015 | Equipment Management | applied | — |
| V016 | HSE Module | applied | — |
| V017 | Quality Module | applied | — |
| V018 | GIS Survey Module | applied | — |
| V019 | Risk Management Module | applied | — |
| V020 | Change Management Module | applied | — |
| V021 | TBM Management | applied | — |
| V022 | Ring Builder Segment Tracking | applied | — |
| V023 | NATM Microtunnelling | applied | — |
| V024 | Auth Audit | applied | — |
| V025 | EVM Module | applied | — |
| V026 | Primavera P6 Connector | applied | — |
| V027 | Funding Module | applied | — |
| V028 | Quality HSE Minimum | applied | — |
| V029 | Laboratory Module | applied | — |
| V030 | Permits Module | applied | — |
| V031 | Insurance Module | applied | — |
| V032 | Fleet Module | applied | — |
| V033 | Neo4j Kafka Module | applied | — |
| V034 | Time Attendance Module | applied | — |
| V035 | Training Certifications Module | applied | — |
| V036 | Physical Progress IPC | applied | — |
| V037 | Shaft Management Module | applied | — |
| V038 | Cross Passage Geology Module | applied | — |
| V039 | Settlement Grouting Ventilation | applied | — |
| V040 | Instrumentation Dewatering TBM Maintenance | applied | — |
| V041 | Retention Guarantees MultiCurrency | applied | — |
| V042 | Audit Trail Tax Management | applied | — |
| V043 | Stakeholder ESG Sustainability | applied | — |
| V044 | Reporting Builder | applied | — |
| V045 | Asset Management | applied | — |
| V046 | Performance Benchmarking | applied | — |
| V047 | Integration Framework | applied | — |
| V048 | AI Assistant Framework | applied | — |
| V049 | Audit Compliance | applied | — |
| V050 | Mobile API Notifications | applied | — |
| V051 | Tunnel Logistics | applied | — |
| V052 | Tunnel Ventilation Design | applied | — |
| V053 | Tunnel Fire Safety Systems | applied | — |
| V054 | Financial Consolidation | applied | — |
| V055 | Project Financing Loans | applied | — |
| V056_058 | AI Classifier Cost Schedule | applied | — |
| V059_061 | AI Risk Quality Progress | applied | — |
| V062_065 | AI Contract Safety Report Predictive | applied | — |
| V066_068 | AI Chatbot PM ESG | applied | — |
| V069_074 | Integrations SAP BIM360 iTwin SharePoint TG PowerBI | applied | — |
| V075 | Lessons Learned | applied | — |
| V077 | Project Management | applied | — |
| V079 | Variation Orders | applied | — |
| V080 | Cost Tx Chapter Level | applied | — |
| V081 | Tunnel Fire Safety | applied | — |
| V082 | Seed Data TTZ V3 | applied | — |
| V082 | Telegram Bot Messages Subscribers | applied | — |
| V083 | Segment Factory Module | applied | — |
| V084 | — СЛЕДУЮЩИЙ СВОБОДНЫЙ — бронировать строкой в том же коммите | | |

**Инцидент 02.07:** на main лежали 6 неустанавливаемых миграций (дубль V009, дубль таблиц projects/document_revisions/hse_incidents, BIGINT-ключи против UUID, phantom-таблицы sections/hr_employees/pm_projects, expression в UNIQUE, не-IMMUTABLE индекс). Всё исправлено. CI теперь дополнительно запрещает BIGSERIAL/BIGINT в миграциях.

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
