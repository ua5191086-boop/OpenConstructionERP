-- ============================================================================
-- V009__Project_Management.sql
-- Модуль управления проектами (Project Management)
-- Часть OpenConstructionERP — Project Operating System
-- ============================================================================

-- ============================================================================
-- 1. Проекты (расширение базовой таблицы)
-- ============================================================================
CREATE TABLE projects (
    id              BIGSERIAL PRIMARY KEY,
    code            VARCHAR(50) NOT NULL UNIQUE,          -- P-2026-001
    name            VARCHAR(500) NOT NULL,
    description     TEXT,
    project_type    VARCHAR(100) NOT NULL,                 -- metro, tunnel, bridge, road, building, industrial, water
    status          VARCHAR(50) NOT NULL DEFAULT 'lead',   -- lead, tender, planning, design, construction, commissioning, operation, closed
    phase           VARCHAR(50) DEFAULT 'feasibility',     -- feasibility, design, tender, construction, handover
    
    -- Заказчик / владелец
    client_id       BIGINT REFERENCES contractors(id),
    owner_id        BIGINT REFERENCES contractors(id),
    
    -- Локация
    country         VARCHAR(100),
    city            VARCHAR(200),
    region          VARCHAR(200),
    address         TEXT,
    coordinates     POINT,                                  -- широта, долгота
    
    -- Сроки
    start_date      DATE,
    end_date        DATE,
    duration_days   INTEGER,
    
    -- Бюджет
    budget_total    NUMERIC(18,2),
    budget_currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    contingency     NUMERIC(18,2),
    contingency_pct NUMERIC(5,2),
    
    -- Метрики
    total_length_km NUMERIC(10,3),                         -- для линейных объектов
    total_area_m2   NUMERIC(14,2),
    total_volume_m3 NUMERIC(14,2),
    
    -- Команда
    project_manager_id BIGINT REFERENCES employees(id),
    sponsor_id      BIGINT REFERENCES employees(id),
    
    -- Классификация
    risk_class      VARCHAR(20),                            -- A, B, C, D
    complexity      VARCHAR(20),                            -- low, medium, high, mega
    confidentiality VARCHAR(50) DEFAULT 'internal',         -- public, internal, confidential, secret
    
    -- Документы
    logo_path       VARCHAR(500),
    contract_path   VARCHAR(500),
    
    notes           TEXT,
    created_by      VARCHAR(100),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_projects_status ON projects(status);
CREATE INDEX idx_projects_type ON projects(project_type);
CREATE INDEX idx_projects_client ON projects(client_id);
CREATE INDEX idx_projects_dates ON projects(start_date, end_date);
CREATE INDEX idx_projects_pm ON projects(project_manager_id);

-- ============================================================================
-- 2. WBS (Work Breakdown Structure)
-- ============================================================================
CREATE TABLE wbs_items (
    id              BIGSERIAL PRIMARY KEY,
    project_id      BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    parent_id       BIGINT REFERENCES wbs_items(id),
    wbs_code        VARCHAR(100) NOT NULL,                  -- 1.1.1, 1.1.2, etc
    name            VARCHAR(500) NOT NULL,
    description     TEXT,
    wbs_level       INTEGER NOT NULL,                       -- 1, 2, 3, 4, 5
    sort_order      INTEGER DEFAULT 0,
    is_leaf         BOOLEAN DEFAULT TRUE,
    
    -- Плановые показатели
    planned_start   DATE,
    planned_end     DATE,
    planned_duration INTEGER,
    planned_cost    NUMERIC(18,2),
    planned_hours   NUMERIC(12,2),
    
    -- Фактические
    actual_start    DATE,
    actual_end      DATE,
    actual_cost     NUMERIC(18,2),
    actual_hours    NUMERIC(12,2),
    progress_pct    NUMERIC(5,2) DEFAULT 0,                 -- % выполнения
    
    -- Ответственный
    responsible_id  BIGINT REFERENCES employees(id),
    
    -- Статус
    status          VARCHAR(50) NOT NULL DEFAULT 'planned',  -- planned, in_progress, completed, delayed, cancelled
    
    -- Привязки
    boq_section_id  BIGINT REFERENCES sections(id),
    contract_id     BIGINT REFERENCES contracts(id),
    
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(project_id, wbs_code)
);

CREATE INDEX idx_wbs_project ON wbs_items(project_id);
CREATE INDEX idx_wbs_parent ON wbs_items(parent_id);
CREATE INDEX idx_wbs_status ON wbs_items(status);
CREATE INDEX idx_wbs_dates ON wbs_items(planned_start, planned_end);

-- ============================================================================
-- 3. Milestones / вехи
-- ============================================================================
CREATE TABLE project_milestones (
    id              BIGSERIAL PRIMARY KEY,
    project_id      BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    wbs_item_id     BIGINT REFERENCES wbs_items(id),
    milestone_code  VARCHAR(50) NOT NULL,
    name            VARCHAR(500) NOT NULL,
    description     TEXT,
    milestone_type  VARCHAR(50) NOT NULL,                   -- start, finish, payment, approval, delivery, permit, review
    category        VARCHAR(50) DEFAULT 'technical',        -- technical, financial, contractual, regulatory
    
    -- Даты
    planned_date    DATE NOT NULL,
    forecast_date   DATE,
    actual_date     DATE,
    
    -- Статус
    status          VARCHAR(50) NOT NULL DEFAULT 'planned', -- planned, achieved, delayed, cancelled, at_risk
    delay_days      INTEGER DEFAULT 0,
    
    -- Вес / важность
    weight_pct      NUMERIC(5,2),                           -- вес milestone в проекте
    is_gate         BOOLEAN DEFAULT FALSE,                   -- gate review (go/no-go)
    
    -- Сумма
    amount          NUMERIC(18,2),                          -- сумма привязанная к milestone
    amount_currency VARCHAR(3) DEFAULT 'USD',
    
    -- Ответственный
    responsible_id  BIGINT REFERENCES employees(id),
    
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(project_id, milestone_code)
);

CREATE INDEX idx_milestones_project ON project_milestones(project_id);
CREATE INDEX idx_milestones_date ON project_milestones(planned_date);
CREATE INDEX idx_milestones_status ON project_milestones(status);

-- ============================================================================
-- 4. Фазы проекта
-- ============================================================================
CREATE TABLE project_phases (
    id              BIGSERIAL PRIMARY KEY,
    project_id      BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    phase_code      VARCHAR(50) NOT NULL,
    name            VARCHAR(500) NOT NULL,
    description     TEXT,
    sort_order      INTEGER DEFAULT 0,
    
    -- Даты
    planned_start   DATE,
    planned_end     DATE,
    actual_start    DATE,
    actual_end      DATE,
    
    -- Бюджет фазы
    budget_amount   NUMERIC(18,2),
    actual_amount   NUMERIC(18,2),
    
    -- Статус
    status          VARCHAR(50) NOT NULL DEFAULT 'pending', -- pending, active, completed, delayed
    
    -- Документы
    deliverables    TEXT,                                   -- JSON список результатов
    completion_pct  NUMERIC(5,2) DEFAULT 0,
    
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(project_id, phase_code)
);

-- ============================================================================
-- 5. Команда проекта
-- ============================================================================
CREATE TABLE project_team (
    id              BIGSERIAL PRIMARY KEY,
    project_id      BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    employee_id     BIGINT NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    role            VARCHAR(200) NOT NULL,
    role_category   VARCHAR(50) NOT NULL,                   -- management, engineering, supervision, admin, support
    start_date      DATE NOT NULL,
    end_date        DATE,
    allocation_pct  NUMERIC(5,2) DEFAULT 100,               -- % загрузки
    is_key          BOOLEAN DEFAULT FALSE,                   -- key personnel
    hourly_rate     NUMERIC(10,2),
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(project_id, employee_id, role)
);

CREATE INDEX idx_team_project ON project_team(project_id);
CREATE INDEX idx_team_employee ON project_team(employee_id);

-- ============================================================================
-- 6. Портфель проектов
-- ============================================================================
CREATE TABLE project_portfolio (
    id              BIGSERIAL PRIMARY KEY,
    code            VARCHAR(50) NOT NULL UNIQUE,
    name            VARCHAR(500) NOT NULL,
    description     TEXT,
    portfolio_type  VARCHAR(50) NOT NULL DEFAULT 'program',  -- program, portfolio, framework
    parent_id       BIGINT REFERENCES project_portfolio(id),
    owner_id        BIGINT REFERENCES employees(id),
    budget_total    NUMERIC(18,2),
    budget_currency VARCHAR(3) DEFAULT 'USD',
    status          VARCHAR(50) NOT NULL DEFAULT 'active',
    start_date      DATE,
    end_date        DATE,
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================================
-- 7. Проекты в портфеле
-- ============================================================================
CREATE TABLE portfolio_projects (
    id              BIGSERIAL PRIMARY KEY,
    portfolio_id    BIGINT NOT NULL REFERENCES project_portfolio(id) ON DELETE CASCADE,
    project_id      BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    sort_order      INTEGER DEFAULT 0,
    notes           TEXT,
    UNIQUE(portfolio_id, project_id)
);

-- ============================================================================
-- 8. Риски проекта (базовый risk register)
-- ============================================================================
CREATE TABLE project_risks (
    id              BIGSERIAL PRIMARY KEY,
    project_id      BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    wbs_item_id     BIGINT REFERENCES wbs_items(id),
    risk_code       VARCHAR(50) NOT NULL,
    name            VARCHAR(500) NOT NULL,
    description     TEXT,
    risk_category   VARCHAR(100) NOT NULL,                  -- technical, financial, schedule, legal, environmental, HSE, political
    risk_type       VARCHAR(50) NOT NULL,                   -- threat, opportunity
    
    -- Оценка
    probability     VARCHAR(20) NOT NULL,                   -- very_low, low, medium, high, very_high
    impact          VARCHAR(20) NOT NULL,                   -- very_low, low, medium, high, very_high
    probability_score INTEGER DEFAULT 1,                    -- 1-5
    impact_score    INTEGER DEFAULT 1,                       -- 1-5
    risk_score      INTEGER GENERATED ALWAYS AS (probability_score * impact_score) STORED,
    
    -- Стоимость
    potential_cost  NUMERIC(18,2),
    mitigation_cost NUMERIC(18,2),
    residual_cost   NUMERIC(18,2),
    
    -- Митигация
    mitigation_strategy VARCHAR(50),                        -- avoid, transfer, mitigate, accept
    mitigation_plan TEXT,
    contingency_plan TEXT,
    
    -- Статус
    status          VARCHAR(50) NOT NULL DEFAULT 'identified', -- identified, assessed, mitigation_planned, mitigation_in_progress, closed, realized
    owner_id        BIGINT REFERENCES employees(id),
    target_date     DATE,
    closed_date     DATE,
    
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(project_id, risk_code)
);

CREATE INDEX idx_risks_project ON project_risks(project_id);
CREATE INDEX idx_risks_score ON project_risks(risk_score);
CREATE INDEX idx_risks_status ON project_risks(status);

-- ============================================================================
-- 9. Изменения / Variations
-- ============================================================================
CREATE TABLE project_changes (
    id              BIGSERIAL PRIMARY KEY,
    project_id      BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    change_number   VARCHAR(50) NOT NULL,
    change_type     VARCHAR(50) NOT NULL,                   -- variation, change_order, scope_change, design_change
    source          VARCHAR(50) NOT NULL,                   -- client, contractor, design, regulatory, unforeseen
    description     TEXT NOT NULL,
    justification   TEXT,
    
    -- Влияние
    cost_impact     NUMERIC(18,2),
    schedule_impact INTEGER,                                -- дней
    scope_change    TEXT,
    
    -- Статус
    status          VARCHAR(50) NOT NULL DEFAULT 'submitted', -- submitted, review, approved, rejected, implemented
    submitted_by    BIGINT REFERENCES employees(id),
    submitted_at    TIMESTAMPTZ,
    approved_by     BIGINT REFERENCES employees(id),
    approved_at     TIMESTAMPTZ,
    
    -- Документы
    document_path   VARCHAR(1000),
    
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(project_id, change_number)
);

CREATE INDEX idx_changes_project ON project_changes(project_id);
CREATE INDEX idx_changes_status ON project_changes(status);

-- ============================================================================
-- 10. Уроки / Lessons Learned
-- ============================================================================
CREATE TABLE project_lessons (
    id              BIGSERIAL PRIMARY KEY,
    project_id      BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    wbs_item_id     BIGINT REFERENCES wbs_items(id),
    category        VARCHAR(100) NOT NULL,                  -- technical, management, financial, HSE, quality
    title           VARCHAR(500) NOT NULL,
    description     TEXT NOT NULL,
    root_cause      TEXT,
    impact          TEXT,
    recommendation  TEXT,
    is_positive     BOOLEAN DEFAULT FALSE,                  -- TRUE = success story, FALSE = lesson from problem
    severity        VARCHAR(20) DEFAULT 'medium',
    status          VARCHAR(50) DEFAULT 'draft',            -- draft, reviewed, published
    author_id       BIGINT REFERENCES employees(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================================
-- 11. Триггеры
-- ============================================================================
CREATE OR REPLACE FUNCTION update_project_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_project_updated
    BEFORE UPDATE ON projects
    FOR EACH ROW
    EXECUTE FUNCTION update_project_timestamp();

CREATE TRIGGER trg_wbs_updated
    BEFORE UPDATE ON wbs_items
    FOR EACH ROW
    EXECUTE FUNCTION update_project_timestamp();

CREATE TRIGGER trg_risk_updated
    BEFORE UPDATE ON project_risks
    FOR EACH ROW
    EXECUTE FUNCTION update_project_timestamp();

-- ============================================================================
-- Комментарии
-- ============================================================================
COMMENT ON TABLE projects IS 'Проекты (расширенная модель)';
COMMENT ON TABLE wbs_items IS 'Иерархическая структура работ (WBS)';
COMMENT ON TABLE project_milestones IS 'Вехи проекта';
COMMENT ON TABLE project_phases IS 'Фазы проекта';
COMMENT ON TABLE project_team IS 'Команда проекта';
COMMENT ON TABLE project_portfolio IS 'Портфели / программы проектов';
COMMENT ON TABLE portfolio_projects IS 'Проекты в портфеле';
COMMENT ON TABLE project_risks IS 'Реестр рисков';
COMMENT ON TABLE project_changes IS 'Изменения / вариации';
COMMENT ON TABLE project_lessons IS 'Уроки / Lessons Learned';
