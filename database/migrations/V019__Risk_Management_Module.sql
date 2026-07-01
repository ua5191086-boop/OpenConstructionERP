-- ============================================================================
-- V019__Risk_Management_Module.sql
-- Модуль Risk Management (RM) — Monte Carlo, Scenario Analysis, Escalation
-- Часть OpenConstructionERP — Project Operating System
-- ============================================================================

-- ============================================================================
-- 1. Risk Categories (категории рисков)
-- ============================================================================
CREATE TABLE risk_categories (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    category_code   VARCHAR(30) NOT NULL,
    category_name   VARCHAR(200) NOT NULL,
    category_type   VARCHAR(30) NOT NULL DEFAULT 'threat'
        CHECK (category_type IN ('threat','opportunity')),
    parent_id       UUID REFERENCES risk_categories(id) ON DELETE SET NULL,
    description     TEXT,
    sort_order      INTEGER DEFAULT 0,
    is_active       BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (category_code)
);

CREATE INDEX idx_rm_cat_parent ON risk_categories(parent_id);

COMMENT ON TABLE risk_categories IS 'Risk Categories — иерархическая структура категорий рисков';

-- ============================================================================
-- 2. Risk Registers (реестр рисков, расширение project_risks)
-- ============================================================================
CREATE TABLE risk_registers (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    risk_number     INTEGER NOT NULL,
    risk_code       VARCHAR(30) NOT NULL,                    -- 'RSK-0001'
    risk_name       VARCHAR(500) NOT NULL,
    risk_type       VARCHAR(20) NOT NULL DEFAULT 'threat'
        CHECK (risk_type IN ('threat','opportunity')),
    category_id     UUID REFERENCES risk_categories(id) ON DELETE SET NULL,
    wbs_code        VARCHAR(50),
    description     TEXT NOT NULL,
    root_cause      TEXT,
    consequence     TEXT,
    probability_score NUMERIC(3,1) DEFAULT 1,
    impact_score    NUMERIC(3,1) DEFAULT 1,
    risk_score      NUMERIC(5,1) GENERATED ALWAYS AS (probability_score * impact_score) STORED,
    probability_level VARCHAR(20)
        CHECK (probability_level IN ('very_low','low','medium','high','very_high')),
    impact_level    VARCHAR(20)
        CHECK (impact_level IN ('very_low','low','medium','high','very_high')),
    risk_rating     VARCHAR(20)
        CHECK (risk_rating IN ('very_low','low','medium','high','extreme')),
    cost_impact     NUMERIC(14,2),
    schedule_impact_days INTEGER,
    risk_owner      VARCHAR(200),
    risk_response   VARCHAR(30) DEFAULT 'accept'
        CHECK (risk_response IN ('avoid','transfer','mitigate','accept','exploit','share','enhance')),
    mitigation_strategy TEXT,
    contingency_plan TEXT,
    trigger_conditions TEXT,
    secondary_risks TEXT,
    status          VARCHAR(20) NOT NULL DEFAULT 'identified'
        CHECK (status IN ('identified','analyzed','response_planned','monitoring','closed','void')),
    reviewed_at     TIMESTAMPTZ,
    reviewed_by     VARCHAR(200),
    closed_at       TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, risk_number),
    UNIQUE (project_id, risk_code)
);

CREATE INDEX idx_rm_reg_project ON risk_registers(project_id);
CREATE INDEX idx_rm_reg_category ON risk_registers(category_id);
CREATE INDEX idx_rm_reg_rating ON risk_registers(project_id, risk_rating);
CREATE INDEX idx_rm_reg_status ON risk_registers(project_id, status);

COMMENT ON TABLE risk_registers IS 'Risk Register — расширенный реестр рисков проекта';

-- ============================================================================
-- 3. Risk Matrices (матрицы риск-менеджмента)
-- ============================================================================
CREATE TABLE risk_matrices (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    matrix_name     VARCHAR(300) NOT NULL,
    matrix_type     VARCHAR(30) NOT NULL DEFAULT 'probability_impact'
        CHECK (matrix_type IN ('probability_impact','cost_schedule','qualitative')),
    grid_data       JSONB NOT NULL,                          -- probability x impact grid
    levels          JSONB DEFAULT '[]'::jsonb,               -- [{name, min, max, color}]
    is_active       BOOLEAN DEFAULT TRUE,
    version         VARCHAR(10) DEFAULT '1.0',
    description     TEXT,
    created_by      VARCHAR(200),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, matrix_name)
);

CREATE INDEX idx_rm_mat_project ON risk_matrices(project_id);

COMMENT ON TABLE risk_matrices IS 'Risk Matrices — матрицы вероятности/влияния и оценки рисков';

-- ============================================================================
-- 4. Risk Monte Carlo Runs (симуляция Монте-Карло)
-- ============================================================================
CREATE TABLE risk_monte_carlo_runs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    run_label       VARCHAR(300) NOT NULL,
    run_type        VARCHAR(30) NOT NULL DEFAULT 'cost'
        CHECK (run_type IN ('cost','schedule','combined')),
    iterations      INTEGER NOT NULL DEFAULT 10000,
    random_seed     INTEGER,
    variables       JSONB DEFAULT '[]'::jsonb,               -- [{name, distribution, params}]
    correlations    JSONB DEFAULT '[]'::jsonb,
    results         JSONB DEFAULT '{}'::jsonb,               -- {p10, p50, p90, mean, stddev, histogram}
    p10_value       NUMERIC(14,2),
    p50_value       NUMERIC(14,2),
    p90_value       NUMERIC(14,2),
    mean_value      NUMERIC(14,2),
    std_dev         NUMERIC(14,2),
    confidence_level NUMERIC(5,2),
    execution_time_ms INTEGER,
    status          VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending','running','completed','failed','cancelled')),
    started_at      TIMESTAMPTZ,
    completed_at    TIMESTAMPTZ,
    error_message   TEXT,
    notes           TEXT,
    created_by      VARCHAR(200),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_rm_mc_project ON risk_monte_carlo_runs(project_id);
CREATE INDEX idx_rm_mc_type ON risk_monte_carlo_runs(project_id, run_type);
CREATE INDEX idx_rm_mc_status ON risk_monte_carlo_runs(project_id, status);

COMMENT ON TABLE risk_monte_carlo_runs IS 'Risk Monte Carlo — симуляции Монте-Карло для стоимости и расписания';

-- ============================================================================
-- 5. Risk Scenarios (сценарный анализ)
-- ============================================================================
CREATE TABLE risk_scenarios (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    scenario_number INTEGER NOT NULL,
    scenario_code   VARCHAR(30) NOT NULL,                    -- 'SCN-0001'
    scenario_name   VARCHAR(500) NOT NULL,
    scenario_type   VARCHAR(30) NOT NULL DEFAULT 'what_if'
        CHECK (scenario_type IN ('what_if','best_case','worst_case','most_likely','sensitivity','stress_test','combined')),
    description     TEXT NOT NULL,
    assumptions     TEXT,
    trigger_events  JSONB DEFAULT '[]'::jsonb,
    affected_risks  UUID[] DEFAULT '{}',
    cost_impact_min NUMERIC(14,2),
    cost_impact_max NUMERIC(14,2),
    cost_impact_ml  NUMERIC(14,2),                          -- most likely
    schedule_impact_min INTEGER,
    schedule_impact_max INTEGER,
    schedule_impact_ml  INTEGER,
    probability_pct NUMERIC(5,2),
    severity        VARCHAR(20)
        CHECK (severity IN ('very_low','low','medium','high','extreme')),
    recommendations TEXT,
    status          VARCHAR(20) NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft','analyzed','reviewed','approved','rejected')),
    approved_by     VARCHAR(200),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, scenario_number),
    UNIQUE (project_id, scenario_code)
);

CREATE INDEX idx_rm_scn_project ON risk_scenarios(project_id);
CREATE INDEX idx_rm_scn_type ON risk_scenarios(project_id, scenario_type);
CREATE INDEX idx_rm_scn_status ON risk_scenarios(project_id, status);

COMMENT ON TABLE risk_scenarios IS 'Risk Scenarios — сценарный анализ и what-if анализ';

-- ============================================================================
-- 6. Risk Mitigation Actions (меры по снижению рисков)
-- ============================================================================
CREATE TABLE risk_mitigation_actions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    risk_id         UUID NOT NULL REFERENCES risk_registers(id) ON DELETE CASCADE,
    action_number   INTEGER NOT NULL,
    action_code     VARCHAR(30) NOT NULL,                    -- 'MIT-0001'
    action_name     VARCHAR(500) NOT NULL,
    action_type     VARCHAR(30) NOT NULL DEFAULT 'preventive'
        CHECK (action_type IN ('preventive','contingency','corrective','fallback')),
    description     TEXT NOT NULL,
    assigned_to     VARCHAR(200),
    budget          NUMERIC(12,2),
    start_date      DATE,
    due_date        DATE,
    completed_at    TIMESTAMPTZ,
    effectiveness   VARCHAR(20)
        CHECK (effectiveness IN ('effective','partially_effective','not_effective','pending_review')),
    residual_probability NUMERIC(3,1),
    residual_impact NUMERIC(3,1),
    residual_score  NUMERIC(5,1) GENERATED ALWAYS AS (residual_probability * residual_impact) STORED,
    status          VARCHAR(20) NOT NULL DEFAULT 'planned'
        CHECK (status IN ('planned','in_progress','completed','cancelled','overdue')),
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, action_number),
    UNIQUE (project_id, action_code)
);

CREATE INDEX idx_rm_ma_project ON risk_mitigation_actions(project_id);
CREATE INDEX idx_rm_ma_risk ON risk_mitigation_actions(risk_id);
CREATE INDEX idx_rm_ma_status ON risk_mitigation_actions(project_id, status);

COMMENT ON TABLE risk_mitigation_actions IS 'Risk Mitigation — меры по снижению/обработке рисков';

-- ============================================================================
-- 7. Risk Escalation (эскалация рисков)
-- ============================================================================
CREATE TABLE risk_escalation (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    risk_id         UUID REFERENCES risk_registers(id) ON DELETE CASCADE,
    escalation_number INTEGER NOT NULL,
    escalation_code VARCHAR(30) NOT NULL,                    -- 'ESC-0001'
    title           VARCHAR(500) NOT NULL,
    reason          TEXT NOT NULL,
    current_status  TEXT,
    recommendation TEXT,
    escalated_to    VARCHAR(300) NOT NULL,
    escalated_by    VARCHAR(200) NOT NULL,
    escalated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    response        TEXT,
    responded_by    VARCHAR(200),
    responded_at    TIMESTAMPTZ,
    decision        VARCHAR(30)
        CHECK (decision IN ('approved','rejected','modified','deferred','noted')),
    outcome         TEXT,
    status          VARCHAR(20) NOT NULL DEFAULT 'escalated'
        CHECK (status IN ('escalated','acknowledged','responded','resolved','closed')),
    closed_at       TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, escalation_number),
    UNIQUE (project_id, escalation_code)
);

CREATE INDEX idx_rm_esc_project ON risk_escalation(project_id);
CREATE INDEX idx_rm_esc_risk ON risk_escalation(risk_id);
CREATE INDEX idx_rm_esc_status ON risk_escalation(project_id, status);

COMMENT ON TABLE risk_escalation IS 'Risk Escalation — эскалация рисков вышестоящему руководству';

-- ============================================================================
-- 8. Risk Dashboard (materialized dashboard data)
-- ============================================================================
CREATE TABLE risk_dashboard (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    snapshot_date   DATE NOT NULL DEFAULT CURRENT_DATE,
    total_risks     INTEGER DEFAULT 0,
    open_risks      INTEGER DEFAULT 0,
    extreme_risks   INTEGER DEFAULT 0,
    high_risks      INTEGER DEFAULT 0,
    medium_risks    INTEGER DEFAULT 0,
    low_risks       INTEGER DEFAULT 0,
    threats         INTEGER DEFAULT 0,
    opportunities   INTEGER DEFAULT 0,
    risk_exposure   NUMERIC(14,2) DEFAULT 0,
    contingency_required NUMERIC(14,2) DEFAULT 0,
    risks_by_category JSONB DEFAULT '{}'::jsonb,
    risks_by_status JSONB DEFAULT '{}'::jsonb,
    risks_by_owner  JSONB DEFAULT '{}'::jsonb,
    mitigation_progress_pct NUMERIC(5,2),
    monte_carlo_p10 NUMERIC(14,2),
    monte_carlo_p50 NUMERIC(14,2),
    monte_carlo_p90 NUMERIC(14,2),
    top_risks       JSONB DEFAULT '[]'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, snapshot_date)
);

CREATE INDEX idx_rm_dash_project ON risk_dashboard(project_id);
CREATE INDEX idx_rm_dash_date ON risk_dashboard(project_id, snapshot_date DESC);

COMMENT ON TABLE risk_dashboard IS 'Risk Dashboard — снимки показателей риск-менеджмента';

-- ============================================================================
-- Register module in object_types
-- ============================================================================
INSERT INTO object_types (code, name, icon, module_owner) VALUES
('risk_category',       'Risk Category',      'folder-tree',     'RM'),
('risk_register',       'Risk Register',      'alert-triangle',  'RM'),
('risk_matrix',         'Risk Matrix',        'grid',            'RM'),
('risk_monte_carlo',    'Monte Carlo Run',    'sigma',           'RM'),
('risk_scenario',       'Risk Scenario',      'git-branch',      'RM'),
('risk_mitigation',     'Mitigation Action',  'shield',          'RM'),
('risk_escalation',     'Risk Escalation',    'arrow-up-circle', 'RM'),
('risk_dashboard',      'Risk Dashboard',     'bar-chart-3',     'RM')
ON CONFLICT (code) DO NOTHING;

-- ============================================================================
-- Module summary view
-- ============================================================================
CREATE VIEW risk_summary AS
SELECT
    p.id AS project_id,
    (SELECT COUNT(*) FROM risk_registers WHERE project_id = p.id AND status NOT IN ('closed','void')) AS open_risks,
    (SELECT COUNT(*) FROM risk_registers WHERE project_id = p.id) AS total_risks,
    (SELECT COUNT(*) FROM risk_registers WHERE project_id = p.id AND risk_rating = 'extreme') AS extreme_risks,
    (SELECT COUNT(*) FROM risk_registers WHERE project_id = p.id AND risk_rating = 'high') AS high_risks,
    (SELECT COUNT(*) FROM risk_registers WHERE project_id = p.id AND risk_type = 'threat') AS threats,
    (SELECT COUNT(*) FROM risk_registers WHERE project_id = p.id AND risk_type = 'opportunity') AS opportunities,
    (SELECT COALESCE(SUM(cost_impact),0) FROM risk_registers WHERE project_id = p.id AND risk_type = 'threat') AS total_threat_cost,
    (SELECT COUNT(*) FROM risk_mitigation_actions WHERE project_id = p.id AND status IN ('planned','in_progress')) AS pending_mitigations,
    (SELECT COUNT(*) FROM risk_mitigation_actions WHERE project_id = p.id AND status = 'completed') AS completed_mitigations,
    (SELECT COUNT(*) FROM risk_escalation WHERE project_id = p.id AND status = 'escalated') AS active_escalations,
    (SELECT COUNT(*) FROM risk_monte_carlo_runs WHERE project_id = p.id AND status = 'completed') AS mc_runs_completed,
    (SELECT COUNT(*) FROM risk_scenarios WHERE project_id = p.id AND status = 'analyzed') AS analyzed_scenarios
FROM projects p;