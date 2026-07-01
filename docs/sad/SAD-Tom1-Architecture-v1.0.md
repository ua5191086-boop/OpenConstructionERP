# OpenConstructionERP — Software Architecture Document
## Том 1: Системная архитектура и генеральные проектные решения

**Версия:** 1.0 | **Статус:** Baseline for Development | **Дата:** 02.07.2026
**Классификация:** Open Source (Apache 2.0 core / AGPL modules)
**Аудитория:** команда разработки 50 человек, техлиды, DevOps, продуктовая команда

---

## 0. Генеральные архитектурные решения (позиция главного архитектора)

Все решения приняты. Альтернативы рассмотрены и отклонены. Это baseline — изменения только через ADR-процесс.

| № | Решение | Выбор | Отклонено | Обоснование |
|---|---------|-------|-----------|-------------|
| GD-01 | Парадигма ядра | **Онтологическое ядро** (Foundry-подход): единая семантическая модель объектов, модули — приложения над онтологией | Классическая модульная ERP (Odoo-подход) | Инфраструктурный проект = единый граф объектов (кольцо → захватка → тоннель → контракт → платёж). Модульные ERP теряют связность на стыках — главная боль Primavera+SAP+Aconex-зоопарков |
| GD-02 | Архитектурный стиль | **Модульный монолит ядра + микросервисы периферии** | Чистые микросервисы со старта | 50 разработчиков не вытянут 80 микросервисов с первого дня. Ядро (онтология, ACL, документы) — монолит с модульными границами; тяжёлые домены (BIM-конвертация, AI, телеметрия TBM, аналитика) — отдельные сервисы. Расщепление ядра — по мере роста нагрузки |
| GD-03 | Основной язык backend | **Go** (ядро, сервисы) + **Python** (AI/аналитика/BIM-обработка) + **TypeScript** (BFF/gateway) | Java/Spring, C#/.NET | Go: производительность + простота найма + низкий porog вхождения для open-source контрибьюторов. Python безальтернативен для AI и IfcOpenShell |
| GD-04 | Основная СУБД | **PostgreSQL 16** + расширения: PostGIS (GIS), TimescaleDB (телеметрия), pgvector (AI-эмбеддинги), Apache AGE (граф) | Oracle, MS SQL, MongoDB как primary | Одна СУБД закрывает реляционку, гео, time-series, векторы и граф. Радикально упрощает эксплуатацию on-premise у заказчиков в СНГ/ЦА, где нет облаков |
| GD-05 | Аналитическая СУБД | **ClickHouse** | Druid, BigQuery | EVM/KPI/телеметрия по проекту $500M = миллиарды строк. ClickHouse — стандарт де-факто, open source |
| GD-06 | Шина событий | **Apache Kafka** (межсервисная) + **NATS** (лёгкие уведомления, mobile sync) | RabbitMQ | Event sourcing доменных событий обязателен (аудит, AI-обучение на истории проекта, replay). Kafka = журнал истины событий |
| GD-07 | Frontend | **React 18 + TypeScript**, Module Federation, единая дизайн-система | Angular, Vue | Экосистема, найм, module federation для независимой поставки модулей |
| GD-08 | Mobile | **Flutter**, offline-first: локальная SQLite + event-log синхронизация (CRDT для конфликтов) | React Native, PWA | Тоннель = нет связи. Прораб, маркшейдер, HSE-инспектор работают офлайн сутками. Это не фича, это условие выживания продукта на стройке |
| GD-09 | Аутентификация | **Keycloak** (OIDC/SAML) + **OPA** (Open Policy Agent) для ABAC-политик | Собственная auth | Банки и аудиторы требуют enterprise SSO/LDAP день один |
| GD-10 | Хранилище файлов | **MinIO** (S3-совместимое), версионирование, WORM-режим для контрактных документов | Файловая система, GridFS | On-premise S3 + юридическая неизменяемость Claims/Correspondence |
| GD-11 | Поиск | **OpenSearch** | Elasticsearch (лицензия), Meilisearch | Полнотекст по 10⁶ документов + переписке, лицензионная чистота |
| GD-12 | BIM-ядро | **IfcOpenShell** (парсинг) + **xeokit** (web-viewer) + собственный BCF 3.0 сервер + фрагментация геометрии в тайлы | Autodesk Forge (vendor lock) | Полный OpenBIM: IFC 4.3 (включая инфраструктурные сущности — рельсы, тоннели), никакой зависимости от Autodesk/Bentley API |
| GD-13 | AI-слой | **LLM-agnostic gateway** (LiteLLM): Claude/GPT/локальные Llama-модели взаимозаменяемы; агенты — LangGraph; RAG — pgvector + OpenSearch hybrid | Жёсткая привязка к одному вендору | Заказчики (госструктуры ЦА, банки) потребуют on-premise inference. Архитектура обязана работать и на Claude API, и на локальной Llama без переписывания |
| GD-14 | Мультиарендность | **Гибрид:** RLS (row-level security) внутри инсталляции + отдельные инсталляции per-holding | Schema-per-tenant | EPC-холдинг = одна инсталляция, проекты изолированы RLS. Конкурирующие подрядчики никогда не делят инсталляцию |
| GD-15 | Развёртывание | **Kubernetes-first** (Helm), но обязательный **single-node Docker Compose профиль** | Только k8s | Стройплощадка в Туркменистане ≠ AWS. Полный стек обязан подниматься на одном сервере 64GB RAM |

---

## 1. Принципы проектирования

1. **Ontology-first.** Любая сущность (кольцо обделки, RFI, платёж, экскаватор, скважина) — объект онтологии с типом, свойствами, связями, действиями и полной историей версий. Модули не владеют данными — они являются представлениями и рабочими процессами над общим графом объектов.
2. **Event-sourced домены.** Каждое изменение состояния — доменное событие в Kafka. Текущее состояние — проекция. Это даёт: полный аудит для банков/аудиторов, replay для AI-обучения, delay analysis "как было на дату X" из коробки (то, что в Primavera требует ручных baseline-снимков).
3. **Offline-first полевой контур.** Мобильные и планшетные клиенты полностью функциональны без сети; синхронизация — журнал событий с детерминированным разрешением конфликтов.
4. **Открытые стандарты как контракт.** IFC 4.3, BCF 3.0, ISO 19650 (CDE-статусы WIP/Shared/Published/Archived), ISO 21500, FIDIC-процессы (Variation, Claim, EOT), COBie, LandXML, buildingSMART Data Dictionary.
5. **API-first.** Ни одной функции UI, недоступной через API. UI — клиент собственного API.
6. **AI как сотрудник, не как чат.** AI-агенты имеют роли в ролевой модели, права доступа, журнал действий и принцип human-confirmation: агент готовит — человек утверждает (для операций записи в контрактные/финансовые домены).
7. **Безопасность по умолчанию.** Zero-trust между сервисами (mTLS), шифрование at-rest, ABAC до уровня поля, полный immutable audit-log.

---

## 2. Архитектура системы: слои

### 2.1 Общая схема

```
┌─────────────────────────────────────────────────────────────────────┐
│  CLIENTS: Web SPA (React) │ Mobile (Flutter) │ Tablet │ TG-bot │ API │
├─────────────────────────────────────────────────────────────────────┤
│  EDGE: NGINX Ingress → Kong API Gateway (rate limit, auth, routing) │
├─────────────────────────────────────────────────────────────────────┤
│  BFF LAYER: GraphQL Gateway (Apollo Router) + REST Aggregators      │
├──────────────────────────────┬──────────────────────────────────────┤
│  CORE PLATFORM (Go,          │  DOMAIN SERVICES (микросервисы)      │
│  модульный монолит):         │  • bim-svc (Python: IFC/BCF/tiles)   │
│  • Ontology Engine           │  • ai-svc (Python: agents/RAG)       │
│  • Identity & Access (OPA)   │  • telemetry-svc (Go: TBM/monitoring)│
│  • Workflow Engine (BPMN)    │  • schedule-svc (Go: CPM/resource)   │
│  • Document Engine (CDE)     │  • analytics-svc (ClickHouse ETL)    │
│  • Forms & Checklists        │  • gis-svc (GeoServer + vector tiles)│
│  • Notification Hub          │  • integration-svc (коннекторы)      │
│  • Reporting Engine          │  • sync-svc (mobile offline sync)    │
├──────────────────────────────┴──────────────────────────────────────┤
│  MESSAGE BACKBONE: Kafka (domain events) │ NATS (realtime/push)     │
├─────────────────────────────────────────────────────────────────────┤
│  DATA: PostgreSQL16(+PostGIS+Timescale+pgvector+AGE) │ ClickHouse   │
│        MinIO (S3) │ OpenSearch │ Redis (cache/session)              │
├─────────────────────────────────────────────────────────────────────┤
│  PLATFORM OPS: K8s │ ArgoCD │ Prometheus/Grafana/Loki/Tempo │ Vault │
└─────────────────────────────────────────────────────────────────────┘
```

### 2.2 Слои — состав и решения

**Frontend Layer**
- React 18 + TypeScript, Vite, Module Federation: каждый функциональный домен — независимо поставляемый remote-модуль.
- Дизайн-система `@oce/ui` (собственная, на Radix primitives): таблицы виртуализированные (10⁵ строк BOQ), Gantt-компонент собственный (canvas-рендер, до 50k задач), формы — schema-driven (JSON Schema → UI).
- State: TanStack Query (server state) + Zustand (UI state). Realtime — подписки через NATS-WebSocket.
- Спец-компоненты: BIM-viewer (xeokit), GIS-карта (MapLibre GL), профиль тоннеля (D3, пикетаж/кольца), P6-подобная сетевая диаграмма.

**Backend / Core Platform (Go)**
- Модульный монолит `oce-core`: границы модулей — Go-пакеты с явными интерфейсами, общение внутри — только через интерфейсы + события (готовность к выносу в сервисы).
- Ontology Engine: реестр типов объектов (ObjectType), свойств (PropertyType), связей (LinkType), действий (ActionType); версионирование объектов (bitemporal: valid-time + transaction-time).
- Workflow Engine: встроенный BPMN 2.0-совместимый движок (адаптация Zeebe-модели) — все документные процессы (RFI, NCR, VO, Claim) — конфигурируемые схемы, не код.
- Document Engine: CDE по ISO 19650 — статусная модель, нумерация по конфигурируемым маскам, revision control, transmittals, распределение.

**API Layer**
- REST v1: `/api/v1/{domain}/{resource}` — полный CRUD + actions; OpenAPI 3.1 генерируется из кода; пагинация cursor-based; идемпотентность через `Idempotency-Key`.
- GraphQL: федеративная схема (Apollo Federation) — основной API для web-клиента; подписки для realtime.
- Webhooks исходящие: подписка на любое доменное событие.
- gRPC: только межсервисное общение.

**Database Layer** — см. GD-04/05 и раздел 6.

**Message Broker** — Kafka: топики по доменам (`oce.project.events`, `oce.tunnel.telemetry`, `oce.finance.events`...), схемы — Avro + Schema Registry, retention доменных событий — бессрочный (event store). NATS — эфемерные realtime-каналы.

**Authentication / Security Layer**
- Keycloak: OIDC, SAML, LDAP/AD-федерация, MFA (TOTP/WebAuthn).
- Авторизация: RBAC (роли) × ABAC (атрибуты: проект, организация, дисциплина, гриф) через OPA; политики — Rego, версионируются в Git.
- mTLS между сервисами (Istio ambient или Linkerd), секреты — HashiCorp Vault, шифрование БД — at-rest (LUKS/pgcrypto для полей ПДн).
- Audit: каждый запрос записи → immutable audit-log (append-only таблица + хэш-цепочка).

**Storage Layer** — MinIO: бакеты per-project, версионирование, Object Lock (WORM) для юридически значимых документов, lifecycle-политики (горячее → холодное), антивирус-скан (ClamAV) на загрузке, presigned URLs.

**Search Engine** — OpenSearch: индексы документов (текст извлекается Apache Tika + OCR Tesseract для сканов), объектов онтологии, переписки; hybrid search (BM25 + kNN-векторы) для AI-RAG.

**AI Layer** — отдельный контур `ai-svc` (Python/FastAPI): LiteLLM-шлюз (Claude API / OpenAI / vLLM+Llama on-prem), LangGraph-агенты, RAG-пайплайн (chunking → embedding → pgvector + OpenSearch), Knowledge Graph на Apache AGE, model registry (MLflow) для прогнозных моделей (задержки, cost overrun, осадки). Подробно — Том 6.

**GIS Layer** — PostGIS (истина геоданных: трасса, пикетаж, землеотвод, коммуникации, скважины) + GeoServer (WMS/WFS/vector tiles) + MapLibre на клиенте; линейная система координат (LRS) — пикетаж как первоклассная ось: любой объект адресуется ПК+смещение.

**BIM Layer** — bim-svc: приём IFC 4.3 → IfcOpenShell → извлечение структуры (spatial tree, элементы, свойства) в онтологию + геометрия → фрагментированные тайлы (xeokit XKT) в MinIO; BCF 3.0 сервер (issues ↔ модель); связь элемент модели ↔ объект онтологии ↔ задача графика (4D) ↔ позиция BOQ (5D). Подробно — Том 5.

**Mobile Layer** — Flutter-приложения: «Прораб» (наряды, объёмы, daily report, фото), «QA/QC» (ITP, инспекции, NCR), «HSE» (обходы, permit-to-work, инциденты), «Маркшейдер» (съёмка, реперы), «Склад» (приёмка/выдача, штрих-коды). Общее offline-ядро: SQLite + event-log, фоновая синхронизация, конфликты — last-writer-wins с ручной эскалацией для критичных полей.

**Notification Layer** — Notification Hub: правила подписки (объект/событие/роль), каналы: in-app, email (SMTP), Telegram-bot, push (FCM/APNs), SMS-шлюз, WhatsApp Business API; дайджесты и эскалации (нет реакции на RFI за N дней → руководителю).

**Analytics Layer** — ClickHouse: CDC из PostgreSQL (Debezium → Kafka → ClickHouse); семантический слой метрик (metrics-as-code, YAML-определения KPI); встроенные дашборды (собственный BI-конструктор на ECharts) + коннектор Power BI/Superset.

**Document Layer** — см. Document Engine + раздел модулей D-группы.

**Integration Layer** — integration-svc: коннекторная платформа (декларативные коннекторы, retry, mapping DSL): Primavera P6 (XER/XML + P6 EPPM API), MS Project (XML), SAP (IDoc/OData), Excel (шаблонный импорт-экспорт), Power BI (DirectQuery к ClickHouse), Autodesk (Forge/ACC API), Bentley (ProjectWise API), SharePoint/OneDrive (Graph API), Nextcloud (WebDAV), Google Drive, Telegram, Email (IMAP-ingest переписки в CDE).

**Deployment Layer** — Helm-чарты per-service + umbrella chart; профили: `single-node` (Docker Compose, 1 сервер), `standard` (k8s, 3+ узла), `enterprise` (мульти-кластер, DR). GitOps: ArgoCD.

**Monitoring / Logging** — OpenTelemetry во всех сервисах → Prometheus (метрики), Loki (логи), Tempo (трейсы), Grafana (единая панель); SLO-алерты (Alertmanager → Telegram/PagerDuty); Sentry (ошибки клиентов).

**Backup Layer** — PostgreSQL: pgBackRest (PITR, WAL-архив в MinIO), ClickHouse: clickhouse-backup, MinIO: репликация в offsite-бакет; DR-план: RPO 15 мин, RTO 4 часа (standard-профиль); автоматическое тестовое восстановление еженедельно.

---

## 3. Онтологическое ядро (ключевой дифференциатор)

### 3.1 Модель
```
ObjectType(id, code, name, icon, schema JSONB, lifecycle, module_owner)
PropertyType(id, object_type_id, code, datatype, unit, required, indexed)
LinkType(id, code, from_type, to_type, cardinality, semantics)
ActionType(id, object_type_id, code, input_schema, workflow_ref, permissions)
Object(id UUID, type_id, project_id, props JSONB, geom GEOMETRY?, 
       lrs_from, lrs_to,            -- пикетажная адресация
       valid_from, valid_to,        -- bitemporal
       tx_from, tx_to,
       created_by, version)
Link(id, link_type_id, from_object, to_object, props JSONB, valid_from, valid_to)
```

### 3.2 Базовая онтология инфраструктурного проекта (фрагмент)
```
Holding → Company → Project → Contract → Section(участок) → 
  Structure(тоннель/станция/шахта/портал) → WorkPackage → Activity → Assignment
Tunnel → Ring(кольцо) → Segment(блок обделки)
Tunnel → Chainage(пикет) ← BoreHole, Instrument, SettlementPoint
BOQItem ↔ Activity ↔ BIMElement ↔ CostAccount (4D/5D-связка)
Document ↔ {любой объект} (полиморфная связь)
Event(домен) → затрагивает Objects[]
```

### 3.3 Что это даёт
- Вопрос «покажи все NCR по кольцам 340–420 перегона, их влияние на график и стоимость, и связанную переписку» — один запрос по графу, а не выгрузка из четырёх систем.
- AI-агенты работают с одним графом знаний, а не с зоопарком API.
- Новый модуль = новые типы объектов + связи + представления. Ядро не меняется.

---

## 4. Реестр модулей (112 модулей, 16 доменов)

Формат: код — название — назначение. Полные спецификации каждого модуля (функции, роли, входы/выходы, процессы) — **Том 2**.

### A. Платформа (10)
| Код | Модуль | Назначение |
|-----|--------|-----------|
| A-01 | Ontology Manager | Управление типами объектов, свойствами, связями |
| A-02 | Identity & Access | Пользователи, роли, политики OPA, SSO |
| A-03 | Organization Registry | Холдинг, компании, оргструктура, контрагенты |
| A-04 | Workflow Designer | Визуальный конструктор BPMN-процессов |
| A-05 | Forms Builder | Конструктор форм/чек-листов (JSON Schema) |
| A-06 | Notification Hub | Правила уведомлений, каналы, эскалации |
| A-07 | Audit & Compliance | Immutable-аудит, отчёты для аудиторов |
| A-08 | Report Designer | Конструктор печатных форм (шаблоны docx/pdf) |
| A-09 | Data Import/Export | Массовый импорт (Excel/CSV/XER/IFC), валидация |
| A-10 | Admin Console | Настройки инсталляции, лицензии, тенанты |

### B. CRM, тендеры и преддоговорная работа (8)
| B-01 | Lead & Opportunity | Лиды, воронка, скоринг проектов |
| B-02 | Tender Management | Реестр тендеров, дедлайны, go/no-go |
| B-03 | Prequalification | PQ-досье компании, справки, сертификаты |
| B-04 | Estimating & Bid Pricing | Тендерная смета, ресурсные нормы, коэффициенты (интеграция логики CAI Cost Intelligence) |
| B-05 | Bid Document Assembly | Сборка тендерного пакета, формы, версии |
| B-06 | Competitor Intelligence | Досье конкурентов, история цен, анализ BOQ конкурентов |
| B-07 | JV & Partnering | Консорциумы, доли, pre-bid соглашения |
| B-08 | Handover to Execution | Передача тендер → проект (бюджет-паспорт) |

### C. Управление проектом (9)
| C-01 | Project Registry | Паспорт проекта, фазы, статусы |
| C-02 | WBS/CBS/OBS Manager | Структуры декомпозиции, маппинг между ними |
| C-03 | Scope Management | Реестр объёмов, scope-изменения |
| C-04 | Progress Measurement | Правила измерения прогресса, физобъёмы |
| C-05 | Daily/Weekly/Monthly Reporting | Автосборка отчётности из данных системы |
| C-06 | Meeting & Minutes | Совещания, протоколы, поручения с контролем |
| C-07 | Action & Issue Tracker | Поручения, проблемы, эскалации (светофор) |
| C-08 | Stakeholder Management | Реестр стейкхолдеров, матрица коммуникаций |
| C-09 | Project Closeout | Чек-листы закрытия, as-built пакет, lessons learned |

### D. Документооборот и CDE (10)
| D-01 | CDE Core (ISO 19650) | Статусы, нумерация, ревизии, transmittals |
| D-02 | Correspondence | Входящая/исходящая переписка, email-ingest |
| D-03 | RFI Management | Запросы информации, SLA, связь с моделью/графиком |
| D-04 | Submittals & Shop Drawings | Передача на согласование, статусные коды A/B/C/D |
| D-05 | Method Statements & ITP | ППР/технологические карты, планы контроля |
| D-06 | Drawing Register | Реестр чертежей, superseded-контроль |
| D-07 | Contract Documents | Контракты, приложения, гарантии — WORM-хранение |
| D-08 | Minutes & Instructions | Site Instructions, протоколы, директивы инженера |
| D-09 | Records & Archive | Архивирование, retention-политики |
| D-10 | e-Signature | ЭЦП/квалифицированная подпись, маршруты подписания |

### E. График и планирование (7)
| E-01 | Master Schedule (CPM) | Сетевой график, критический путь, календари |
| E-02 | Resource Planning | Ресурсная загрузка, выравнивание |
| E-03 | Lookahead Planning | 2–6-недельные планы, Last Planner, PPC% |
| E-04 | Daily Work Assignment | Наряд-задания, факт за смену |
| E-05 | Baseline & Snapshot | Базовые планы, снимки на дату (из event store) |
| E-06 | Delay & Forensic Analysis | TIA, window analysis, as-planned vs as-built |
| E-07 | Schedule Risk Analysis | Monte-Carlo (P50/P80), tornado-диаграммы |

### F. Финансы и стоимость (12)
| F-01 | Project Budget (CBS) | Бюджет по CBS, версии, утверждение |
| F-02 | Cost Control & EVM | PV/EV/AC, CPI/SPI, EAC/ETC, отчёты EVM |
| F-03 | Cash Flow | Прогноз ДДС помесячно/понедельно, S-curves |
| F-04 | Forecasting | Прогноз к завершению, тренды, сценарии |
| F-05 | Invoicing & Payment Certificates | Акты (IPC), КС-2/КС-3-аналоги, ретеншн |
| F-06 | Payments & Treasury | Платёжный календарь, банковские счета |
| F-07 | Guarantees & Bonds | Банковские гарантии, аккредитивы, сроки |
| F-08 | Funding & Loans | Кредитные линии, ECA/экспортное финансирование, ковенанты, drawdown-графики |
| F-09 | Currency & Escalation | Мультивалютность, индексация, формулы эскалации |
| F-10 | Cost Coding & Allocation | Разноска затрат, коды затрат, драйверы |
| F-11 | Accounting Bridge | Мост в бухгалтерию (SAP/1С/локальные), сверка |
| F-12 | Financial Consolidation | Консолидация портфеля, отчётность холдинга |

### G. Контракты и претензии (7)
| G-01 | Contract Administration | Обязательства, вехи, notices по FIDIC-срокам |
| G-02 | Variation Orders | VO/CO: инициация → оценка → согласование → включение в бюджет/график |
| G-03 | Claims Management | Претензии: события, notices, quantum, EOT |
| G-04 | Subcontract Management | Субподряды, back-to-back условия, акты субчиков |
| G-05 | Insurance & Liability | Страхование CAR/EAR, инциденты, покрытия |
| G-06 | Obligations & Compliance Register | Реестр обязательств из контрактов (AI-извлечение) |
| G-07 | Dispute & Arbitration | DAB/арбитраж, доказательная база из event store |

### H. Закупки и логистика (8)
| H-01 | Procurement Planning | План закупок из графика (даты «нужно на площадке») |
| H-02 | RFQ & Bid Evaluation | Запросы КП, сравнительные таблицы, TCO-оценка |
| H-03 | Purchase Orders | Заказы, согласование, статусы поставки |
| H-04 | Expediting & Shipment Tracking | Экспедирование, инспекции у изготовителя |
| H-05 | Customs & Import | ВЭД, таможня, сертификация (критично для ЦА) |
| H-06 | Vendor Management | Реестр и рейтинг поставщиков, квалификация |
| H-07 | Long-Lead Items | Контроль длинноцикловых позиций (TBM, эскалаторы) |
| H-08 | Framework Agreements | Рамочные договоры, каталоги цен |

### I. Склад и материалы (6)
| I-01 | Warehouse Operations | Приёмка, хранение, выдача, штрих/QR-коды |
| I-02 | Inventory & Stock Control | Остатки, min/max, резервирование под работы |
| I-03 | Material Requests | Заявки с площадки → склад/закупка |
| I-04 | Bulk Materials Tracking | Бетон, арматура, ГСМ: баланс план/факт по захваткам |
| I-05 | Material Certificates | Сертификаты, паспорта качества, привязка к конструкциям |
| I-06 | Logistics & Site Deliveries | Графики завоза, окна разгрузки, транспорт |

### J. Персонал и трудовые ресурсы (7)
| J-01 | Workforce Registry | Персонал, квалификации, допуски, визы/разрешения |
| J-02 | Timesheets & Attendance | Табели, турникеты/биометрия, смены |
| J-03 | Crew & Shift Management | Бригады, вахты, ротации |
| J-04 | Training & Certification | Обучение, аттестации, сроки действия допусков |
| J-05 | Camp & Accommodation | Вахтовые городки, размещение, питание |
| J-06 | Payroll Bridge | Мост в расчёт ЗП, стоимость часа по видам работ |
| J-07 | Manpower Histograms | Гистограммы численности план/факт |

### K. Оборудование и механизация (8)
| K-01 | Plant & Equipment Registry | Реестр техники, паспорта, наработка |
| K-02 | Equipment Deployment | Дислокация по объектам, загрузка, простои |
| K-03 | Maintenance (CMMS) | ППР/ТО, дефекты, заявки, история ремонтов |
| K-04 | Fuel Management | ГСМ: заправки, нормы, контроль расхода |
| K-05 | Spare Parts | Склад запчастей, критичные позиции, заявки |
| K-06 | Batch/Concrete Plants | БСУ: рецептуры, замесы, паспорта бетона |
| K-07 | Segment Factory | Завод обделки: формы, циклы, склад колец, QC блоков |
| K-08 | Rental & Owned Cost | Стоимость машино-часа, аренда vs владение |

### L. Тоннельный домен (14) — ядро специализации
| L-01 | TBM Operations | Проходка: кольца/сутки, параметры режимов, сменные рапорты |
| L-02 | TBM Telemetry | Приём телеметрии ЩПК (Timescale): давление в забое, крутящий момент, скорость, объём пригруза — realtime + история |
| L-03 | Ring Builder & Ring Register | Реестр колец: положение, ориентация, зазоры, тампонаж |
| L-04 | Segment Tracking | Прослеживаемость блока: отливка → склад → кольцо (QR) |
| L-05 | TBM Maintenance | ТО щита: ротор, резцы (учёт износа/замен), гидравлика |
| L-06 | NATM/Drill&Blast | Заходки, паспорта БВР, крепь, классы выработки |
| L-07 | Microtunnelling & Pipe Jacking | МТПК: интервалы, домкратные усилия, бентонит |
| L-08 | Shafts & Portals | Стволы: проходка, крепление, армирование |
| L-09 | Cross Passages | Сбойки: заморозка/цементация, последовательность |
| L-10 | Geology & GBR | Геология: скважины, разрезы, факт vs GBR, геориски |
| L-11 | Grouting & Ground Improvement | Цементация, jet-grouting, заморозка: карты, объёмы |
| L-12 | Instrumentation & Monitoring | КИА: марки, инклинометры, пьезометры — пороги, тренды |
| L-13 | Settlement Management | Осадки: мульды, здания в зоне влияния, alert-уровни |
| L-14 | Slurry/EPB Management | Сепарация, баланс выемки (объём факт vs теория — контроль перебора) |

### M. Качество (6)
| M-01 | ITP Execution | Исполнение планов контроля, hold/witness points |
| M-02 | Inspections & Test Requests | Заявки на инспекции (WIR/MIR), лаборатория |
| M-03 | NCR & CAPA | Несоответствия, корректирующие действия |
| M-04 | Material Testing Lab | Лаборатория: пробы бетона, грунтов, сварки |
| M-05 | Weld & Rebar Control | Сварные соединения, входной контроль арматуры |
| M-06 | Punch Lists & Handover QC | Дефектные ведомости, приёмка |

### N. HSE (6)
| N-01 | Permit to Work | Наряды-допуски (огневые, замкнутые пространства, газоопасные) |
| N-02 | Incident Management | Происшествия, near-miss, расследования, LTIFR |
| N-03 | HSE Inspections & Observations | Обходы, предписания, поведенческий аудит |
| N-04 | Gas & Ventilation Monitoring | Газовый контроль в выработках, вентиляция |
| N-05 | Emergency Response | Планы эвакуации, учёт людей под землёй (tally) |
| N-06 | Environmental Monitoring | Шум, вибрация, вода, отходы, разрешения |

### O. BIM и инженерия (8)
| O-01 | IFC Model Hub | Приём/версии моделей, federated model |
| O-02 | Model Viewer & Sections | Web-просмотр, сечения, измерения, аннотации |
| O-03 | BCF Issue Management | Коллизии и замечания в контексте модели |
| O-04 | Clash Detection | Автопроверка коллизий между дисциплинами |
| O-05 | 4D Simulation | Модель ↔ график: визуализация последовательности |
| O-06 | 5D Quantities | Модель ↔ BOQ: объёмы из модели, факт на модели |
| O-07 | Digital Twin & 6D/7D | As-built двойник, данные эксплуатации, паспортизация |
| O-08 | Survey & Geodesy | Маркшейдерия: реперы, съёмки, исполнительные схемы, увязка с LRS |

### P. Аналитика, KPI и AI (13)
| P-01 | Executive Dashboard | «3 вопроса за 30 секунд»: статус, деньги, риски |
| P-02 | KPI Engine | Metrics-as-code, пороги, светофоры (зелёный <1 нед., жёлтый — 1 нед. на критическом пути, красный — 2 недели подряд) |
| P-03 | Portfolio Analytics | Портфель проектов холдинга, сравнение |
| P-04 | Risk Register & Analytics | Реестр рисков, вероятность/влияние, владельцы |
| P-05 | AI Copilot (Chat) | Диалоговый доступ ко всему графу проекта |
| P-06 | AI Document Analyzer | Извлечение обязательств/сроков/сумм из контрактов и переписки |
| P-07 | AI Scheduler | Генерация/проверка графиков, поиск логических ошибок |
| P-08 | AI Cost Engineer | Проверка смет, бенчмаркинг расценок, аномалии BOQ |
| P-09 | AI Claims & Delay Analyst | Сборка доказательной базы претензий из event store |
| P-10 | AI Risk & Forecast | ML-прогнозы: срыв сроков, перерасход, осадки, износ резцов |
| P-11 | AI Procurement Advisor | Рекомендации поставщиков, прогноз цен |
| P-12 | Knowledge Graph & Lessons | Граф знаний холдинга, поиск прецедентов |
| P-13 | AI Report Writer | Автогенерация нарративов отчётов (draft → human confirm) |

### Q. Интеграции и портал (6)
| Q-01 | Integration Hub | Коннекторы, мониторинг потоков, mapping |
| Q-02 | Client Portal | Портал заказчика: прогресс, документы, платежи |
| Q-03 | Subcontractor Portal | Портал субподрядчика: наряды, акты, документы |
| Q-04 | Bank/Lender Portal | Портал банка: drawdown, ковенанты, отчёты IE |
| Q-05 | Public API & Developer Hub | API-ключи, документация, песочница |
| Q-06 | Telegram/WhatsApp Bots | Отчёты и подтверждения в мессенджерах |

**Итого: 112 модулей.**

---

## 5. Ролевая модель

### 5.1 Принцип
`Доступ = Роль (что можно делать) × Скоуп (где: холдинг/компания/проект/участок) × Атрибуты (дисциплина, гриф, организация) × Статус объекта`. Политики — OPA/Rego. AI-агенты — субъекты той же модели с отдельным типом принципала `agent` и обязательным журналом действий.

### 5.2 Роли (27 базовых, расширяемые)

| Роль | Скоуп по умолчанию | Ключевые права (сводно) |
|------|--------------------|--------------------------|
| CEO / Board | Холдинг | Read-all, утверждение бюджетов/контрактов > порога, портфельные дашборды |
| Project Director | Проект | Полное управление проектом, утверждение VO/платежей до лимита |
| Construction Manager | Проект/участки | Графики, наряды, ресурсы, приём daily reports |
| Chief Engineer | Проект | Техрешения, ППР/ITP-утверждение, RFI-ответы, модель |
| Planning Engineer | Проект | Ведение WBS, lookahead, прогресс |
| Scheduler | Проект | CPM-график, baseline (создание — да, утверждение — PD) |
| Cost Engineer | Проект | Бюджет, EVM, прогнозы; платежи — только подготовка |
| Contract Manager | Проект | Контракты, VO, Claims, notices; подписание — по доверенности |
| Procurement Manager | Компания/проект | RFQ, PO до лимита, вендоры |
| Warehouse Keeper | Склад | Приёмка/выдача, инвентаризация; цены — read |
| HSE Manager | Проект | Permit to work, инциденты, стоп-карты (право остановки работ — фиксируется системой) |
| QA/QC Manager | Проект | ITP, NCR, лаборатория; закрытие NCR — только QA |
| Surveyor (маркшейдер) | Участок | Съёмки, реперы, исполнительные; редактирование чужих съёмок — нет |
| TBM Manager | Тоннель | L-домен полностью, телеметрия, ТО щита |
| Shift Engineer | Смена | Сменные рапорты, наряды смены |
| Foreman (прораб) | Захватка | Мобильный контур: наряды, факт, фото, заявки на материалы |
| Design Manager | Проект | Модели, чертежи, drawing register |
| Document Controller | Проект | CDE: нумерация, transmittals, распределение; контент — не редактирует |
| HR/Camp Manager | Проект | J-домен; персональные данные — гриф ПДн |
| Equipment Manager | Компания | K-домен, дислокация, CMMS |
| Subcontractor | Свой контракт | Портал: свои наряды/акты/документы; чужие данные — невидимы |
| Client / Engineer (заказчик) | Портал | Согласования, инспекции, прогресс; внутренние затраты — невидимы |
| Consultant | Назначенные пакеты | Read + комментарии в своей дисциплине |
| Bank / Lender | Портал | Drawdown-отчёты, ковенанты, IE-отчёты; операционные данные — нет |
| Auditor | Read-only + audit log | Всё на чтение в скоупе аудита, выгрузки с водяными знаками |
| System Admin | Инсталляция | Платформа; бизнес-данные — по отдельному явному гранту |
| AI Agent (класс) | По назначению | Read по скоупу задачи; write — только draft-статусы, финализация человеком |

Полная матрица (роль × модуль × операция CRUD+Actions, ~112×27) — **Том 2, приложение**.

---

## 6. Архитектура данных

### 6.1 Домены данных и владение
Каждый домен владеет своими таблицами; междоменные ссылки — через ID объектов онтологии. Схемы: `core` (онтология, ACL, аудит), `doc`, `sched`, `fin`, `contract`, `proc`, `inv`, `hr`, `equip`, `tunnel`, `qa`, `hse`, `bim`, `gis`, `ai`. Полная структура (~420 таблиц) — **Том 3**.

### 6.2 Конвенции
- PK: `id UUID v7` (сортируемые). FK: `{entity}_id`. Обязательные поля: `project_id`, `created_at/by`, `updated_at/by`, `version int`.
- Мягкое удаление запрещено в юридических доменах (doc, contract, fin) — только статусы + event store.
- Все money-поля: `NUMERIC(18,2)` + `currency CHAR(3)` + курс на дату события.
- Индексация: `project_id` в каждом составном индексе первым; JSONB-props — GIN; геометрия — GIST; телеметрия — Timescale hypertables (партиции по времени + tbm_id).

### 6.3 Эталонный фрагмент схемы — тоннельный домен (стандарт качества для Тома 3)

```sql
CREATE TABLE tunnel.tbm (
  id UUID PRIMARY KEY,
  project_id UUID NOT NULL REFERENCES core.project(id),
  code TEXT NOT NULL,                    -- 'TBM-01'
  manufacturer TEXT, model TEXT,
  type TEXT CHECK (type IN ('EPB','SLURRY','OPEN','MIXSHIELD','GRIPPER')),
  diameter_mm INT NOT NULL,
  commissioning_date DATE,
  status TEXT NOT NULL DEFAULT 'assembly',
  UNIQUE (project_id, code)
);

CREATE TABLE tunnel.drive (                -- проходка (перегон/интервал)
  id UUID PRIMARY KEY,
  project_id UUID NOT NULL,
  tbm_id UUID REFERENCES tunnel.tbm(id),
  name TEXT NOT NULL,
  method TEXT CHECK (method IN ('TBM','NATM','DRILL_BLAST','MTBM','PIPE_JACKING')),
  chainage_from NUMERIC(10,2) NOT NULL,   -- пикетаж, м
  chainage_to   NUMERIC(10,2) NOT NULL,
  design_rings INT,
  alignment GEOMETRY(LINESTRINGZ, 4326)
);

CREATE TABLE tunnel.ring (
  id UUID PRIMARY KEY,
  drive_id UUID NOT NULL REFERENCES tunnel.drive(id),
  ring_no INT NOT NULL,
  chainage NUMERIC(10,2),
  built_at TIMESTAMPTZ,
  shift_id UUID REFERENCES hr.shift(id),
  ring_type TEXT,                          -- универсальное/левое/правое
  key_position SMALLINT,                   -- позиция замкового блока
  advance_mm INT,                          -- ход за кольцо
  grout_volume_m3 NUMERIC(8,2),
  grout_pressure_bar NUMERIC(6,2),
  attitude JSONB,                          -- крен/тангаж/отклонения от оси
  UNIQUE (drive_id, ring_no)
);

CREATE TABLE tunnel.segment (
  id UUID PRIMARY KEY,
  project_id UUID NOT NULL,
  mould_id UUID REFERENCES equip.mould(id),
  cast_batch_id UUID REFERENCES equip.concrete_batch(id),
  segment_type TEXT NOT NULL,              -- A1..K
  qr_code TEXT UNIQUE,
  cast_at TIMESTAMPTZ,
  qc_status TEXT DEFAULT 'pending',
  ring_id UUID REFERENCES tunnel.ring(id), -- NULL пока на складе
  position_in_ring SMALLINT
);

-- Телеметрия: Timescale hypertable
CREATE TABLE tunnel.tbm_telemetry (
  time TIMESTAMPTZ NOT NULL,
  tbm_id UUID NOT NULL,
  face_pressure_bar NUMERIC(6,2),
  torque_mnm NUMERIC(8,2),
  thrust_kn NUMERIC(10,1),
  advance_speed_mm_min NUMERIC(6,1),
  cutterhead_rpm NUMERIC(5,2),
  muck_volume_m3 NUMERIC(8,2),
  foam_flow_l_min NUMERIC(8,2),
  raw JSONB
);
SELECT create_hypertable('tunnel.tbm_telemetry','time');

CREATE TABLE tunnel.instrument (
  id UUID PRIMARY KEY,
  project_id UUID NOT NULL,
  type TEXT CHECK (type IN ('SETTLEMENT_PIN','INCLINOMETER','PIEZOMETER',
                            'EXTENSOMETER','CRACK_GAUGE','TILT_METER')),
  code TEXT NOT NULL,
  location GEOMETRY(POINTZ,4326),
  chainage NUMERIC(10,2),
  alert_level_1 NUMERIC, alert_level_2 NUMERIC, alert_level_3 NUMERIC,
  structure_ref UUID                       -- здание/сооружение в зоне влияния
);

CREATE TABLE tunnel.instrument_reading (
  time TIMESTAMPTZ NOT NULL,
  instrument_id UUID NOT NULL REFERENCES tunnel.instrument(id),
  value NUMERIC(12,4) NOT NULL,
  source TEXT DEFAULT 'manual'             -- manual/auto
);
SELECT create_hypertable('tunnel.instrument_reading','time');
```

Связи с другими доменами: `ring ↔ sched.activity` (прогресс = кольца), `ring ↔ fin.boq_item` (5D: оплата за кольцо), `segment ↔ qa.ncr`, `drive ↔ bim.element`.

---

## 7. Карта процессов (L1, end-to-end)

Полные BPMN-схемы каждого процесса — **Том 4**. Здесь — реестр без пропусков по жизненному циклу:

**Фаза 1. Развитие бизнеса:** P-101 Захват и квалификация лида → P-102 Go/No-Go → P-103 Prequalification → P-104 Тендерная смета → P-105 Сборка и подача предложения → P-106 Переговоры/BAFO → P-107 Подписание контракта → P-108 Передача в исполнение (бюджет-паспорт, тендерные допущения → риски проекта).

**Фаза 2. Мобилизация:** P-201 Открытие проекта (WBS/CBS/OBS, кодировки, CDE) → P-202 Baseline-график и бюджет → P-203 Гарантии/страхование/авансы → P-204 Мобилизация персонала (визы, допуски, кэмп) → P-205 Мобилизация техники → P-206 План закупок и long-lead (TBM!) → P-207 Субподрядные пакеты → P-208 Разрешительная документация.

**Фаза 3. Исполнение (циклические):** P-301 Lookahead → наряды → факт → прогресс → P-302 Проходческий цикл (смена: рапорт → кольца → телеметрия → отклонения) → P-303 Материальный цикл (заявка → закупка/склад → выдача → списание на захватку) → P-304 Качество (ITP → WIR → инспекция → NCR/закрытие) → P-305 HSE-цикл (permit → работа → обход → инцидент/расследование) → P-306 Мониторинг (показания КИА → пороги → alert → mitigation) → P-307 Документооборот (RFI/Submittal/Correspondence по SLA) → P-308 Ежемесячный цикл: физобъёмы → акт (IPC) → счёт → платёж → EVM → отчёт → прогноз.

**Фаза 4. Изменения и споры:** P-401 Variation Order (событие → notice в контрактный срок → оценка cost/time → согласование → ревизия baseline) → P-402 Claim (событие → доказательная база из event store → quantum → подача → урегулирование) → P-403 EOT/Delay analysis → P-404 Споры/DAB.

**Фаза 5. Завершение:** P-501 Testing & Commissioning → P-502 Punch list → P-503 Substantial Completion / Taking Over → P-504 As-built + Digital Twin передача → P-505 Финансовое закрытие (final account, возврат ретеншна/гарантий) → P-506 DLP (гарантийный период, дефекты) → P-507 Demobilization → P-508 Lessons learned → Knowledge Graph → P-509 Архивация (retention).

---

## 8. API

### 8.1 REST — карта неймспейсов (полный перечень endpoints — Том 7)
```
/api/v1/core/{object-types,objects,links,actions,search}
/api/v1/projects/{id}/...          # всё проектное — под проектом
/api/v1/docs/{documents,rfis,submittals,transmittals,correspondence}
/api/v1/schedule/{schedules,activities,baselines,lookaheads,assignments}
/api/v1/finance/{budgets,cost-items,evm,cashflow,invoices,payments,guarantees}
/api/v1/contracts/{contracts,variations,claims,subcontracts,obligations}
/api/v1/procurement/{rfqs,pos,vendors,expediting}
/api/v1/inventory/{warehouses,items,movements,requests}
/api/v1/hr/{people,timesheets,crews,trainings}
/api/v1/equipment/{units,deployments,maintenance,fuel}
/api/v1/tunnel/{tbms,drives,rings,segments,telemetry,instruments,readings}
/api/v1/quality/{itps,inspections,ncrs,tests}
/api/v1/hse/{permits,incidents,observations}
/api/v1/bim/{models,elements,bcf,clashes,links4d,links5d}
/api/v1/gis/{layers,features,alignment}
/api/v1/ai/{chat,agents,analyses,forecasts}
/api/v1/analytics/{kpis,dashboards,reports}
/api/v1/integrations/{connectors,jobs,mappings}
```
Стандарты: cursor-пагинация, `?fields=`-проекции, `?filter=`-RSQL, ETag/If-Match для optimistic locking, `POST .../{id}/actions/{action}` для доменных действий, `Idempotency-Key` обязателен для мутаций.

### 8.2 GraphQL
Федеративная схема: subgraph per domain-service. Корневые типы: `Project`, `Object(type, id)` (универсальный доступ к онтологии), доменные типы поверх. Subscriptions: `objectChanged`, `kpiBreached`, `telemetryStream(tbmId)`. Persisted queries only в production (безопасность и кэш).

---

## 9. Структура monorepo (GitHub: `openconstructionerp/oce`)

```
oce/
├── apps/
│   ├── web/                    # React shell + module federation host
│   ├── web-modules/{docs,schedule,finance,tunnel,bim,...}/
│   ├── mobile/{foreman,qaqc,hse,survey,warehouse}/   # Flutter
│   └── bots/telegram/
├── services/
│   ├── core/                   # Go модульный монолит
│   │   └── internal/{ontology,iam,workflow,cde,forms,notify,report}/
│   ├── bim-svc/                # Python
│   ├── ai-svc/                 # Python
│   ├── telemetry-svc/          # Go
│   ├── schedule-svc/           # Go (CPM engine)
│   ├── analytics-svc/
│   ├── gis-svc/
│   ├── integration-svc/
│   └── sync-svc/
├── packages/
│   ├── ui/                     # дизайн-система
│   ├── api-client/             # генерируемые SDK (TS/Python/Go)
│   ├── ontology-schemas/       # YAML-описания базовой онтологии
│   └── metrics/                # KPI as code
├── infra/
│   ├── docker/ (compose profiles: single-node, dev)
│   ├── helm/ (charts per service + umbrella)
│   ├── terraform/ (референс-инфраструктура)
│   └── argocd/
├── db/
│   ├── migrations/{core,doc,sched,fin,tunnel,...}/   # Atlas/goose
│   └── seed/
├── docs/
│   ├── adr/                    # Architecture Decision Records
│   ├── sad/                    # этот документ и тома 2–8
│   └── api/
├── tools/{codegen,importers(xer,ifc,xlsx)}/
└── .github/workflows/
```

---

## 10. Docker / Kubernetes / CI-CD

**Docker:** distroless-образы (Go), slim (Python); multi-stage builds; SBOM (syft) + подпись (cosign) в pipeline; `docker-compose.single-node.yml` — полный стек: core, 3 domain-сервиса-минимум, postgres, clickhouse, kafka(kraft), minio, opensearch, keycloak, nginx — проверенный профиль на 1×64GB сервер.

**Kubernetes:** namespace per environment; StatefulSets — данные (или операторы: CloudNativePG, Strimzi, ClickHouse-operator); HPA по RPS/lag; PodDisruptionBudgets; NetworkPolicies default-deny; Istio ambient (mTLS); Vault-injector для секретов.

**CI/CD (GitHub Actions + ArgoCD):**
```
PR: lint → unit → contract-tests (API) → build → trivy scan → preview env
main: e2e (Playwright) → integration (testcontainers) → publish images →
      ArgoCD sync dev → auto-promote staging → manual gate → prod (canary 10%)
```
Релизный поезд: minor каждые 2 недели; миграции БД — только expand/contract (zero-downtime).

---

## 11. Дорожная карта (команда 50 человек)

**Состав команды:** 6 команд × 6–7 инженеров + платформенная группа: Core Platform (8), Field & Mobile (7), Tunnel & Engineering (7), Finance & Contracts (7), BIM & GIS (6), AI & Analytics (6), DevOps/SRE (4), QA (3), Design (2).

### MVP — месяцы 1–9 («один проект живёт в системе»)
Ядро: онтология, IAM, CDE (D-01..03), проектный паспорт, WBS/CBS, график (E-01, импорт XER), Daily reports + мобильный «Прораб» (offline), бюджет + IPC (F-01, F-05), тоннельный минимум (L-01, L-03, L-04, L-12/13 ручной ввод), NCR (M-03), Permit to Work (N-01), Executive Dashboard (P-01), single-node деплой.
**Критерий выхода:** реальный проект (пилот — один из действующих проектов CAI) ведёт daily-цикл, кольца, акты и RFI только в системе 60 дней подряд.

### Beta — месяцы 10–18 («замещение Primavera+Aconex на проекте»)
Полный CPM-движок + baseline/snapshots (E-05), EVM/CashFlow (F-02/03), VO+Claims (G-01..03), закупки+склад (H, I), CMMS (K-03), телеметрия TBM realtime (L-02), BIM viewer + BCF + 4D (O-01..05), AI Copilot + Document Analyzer (P-05/06), портал заказчика (Q-02), интеграции P6/Excel/Telegram.
**Критерий:** параллельный прогон с P6/Aconex на проекте — расхождения < 1%, после чего legacy отключается.

### Release 1.0 — месяцы 19–27 («продукт для рынка»)
Все 112 модулей в базовой функциональности, GraphQL public API, marketplace-механизм модулей, Delay Analysis (E-06), Schedule Risk MC (E-07), 5D (O-06), AI Scheduler/Cost/Claims (P-07..09), портал банка (Q-04), i18n (EN/RU/DE/TR/UZ), сертификация безопасности (pentest, ISO 27001-контур).

### Enterprise — месяцы 28–40
Мульти-кластер/DR, консолидация холдинга (F-12), Digital Twin 6D/7D (O-07), ML-прогнозы производственные (P-10), Knowledge Graph межпроектный (P-12), SAP/Bentley глубокие коннекторы, SOC2, управляемое SaaS-предложение + коммерческая поддержка (модель дохода open-core).

---

## 12. Состав документации (тома 2–8)

| Том | Содержание | Объём (оценка) |
|-----|-----------|----------------|
| Том 2 | Спецификации 112 модулей: функции, роли, входы/выходы, экраны, процессы + полная матрица прав 112×27 | ~350 стр. |
| Том 3 | Полная модель данных: ~420 таблиц, все поля/типы/ключи/индексы, ER-диаграммы по доменам | ~250 стр. |
| Том 4 | BPMN всех процессов P-101…P-509 + документные workflow (RFI, NCR, VO, Claim...) | ~180 стр. |
| Том 5 | BIM/GIS: IFC 4.3-маппинг, 4D/5D-связывание, Digital Twin, геодезия/LRS | ~120 стр. |
| Том 6 | AI: архитектура агентов, промпты, RAG, guardrails, ML-модели, датасеты | ~120 стр. |
| Том 7 | API: полный REST (все endpoints, схемы) + GraphQL SDL + webhooks | ~200 стр. |
| Том 8 | Эксплуатация: развёртывание, безопасность, DR, SRE-runbooks, миграция с P6/Aconex | ~100 стр. |

---
*OpenConstructionERP SAD v1.0 — Том 1. Изменения — через ADR в `docs/adr/`.*
