-- ============================================================================
-- V062_065__AI_Contract_Safety_Report_Predictive.sql
-- Contract analyzer, Procurement optimizer, Safety monitor, Report generator
-- ============================================================================
CREATE TABLE ai_contract_analyses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    contract_id UUID REFERENCES contracts(id),
    analysis_type VARCHAR(100) NOT NULL, -- risk_clause, obligation_extraction, deviation_detection, price_analysis
    clauses_found INTEGER,
    risk_clauses INTEGER,
    obligations_extracted JSONB,
    deviations JSONB,
    risk_score NUMERIC(5,2),
    summary TEXT,
    model_version VARCHAR(50),
    analyzed_at TIMESTAMPTZ DEFAULT NOW(),
    reviewed_by VARCHAR(200),
    notes TEXT
);

CREATE TABLE ai_procurement_optimizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID REFERENCES projects(id),
    optimization_type VARCHAR(100) NOT NULL, -- supplier_selection, quantity_optimization, timing, consolidation
    purchase_items_analyzed INTEGER,
    recommendations JSONB, -- [{"item":"...","supplier":"...","saving":1000,"confidence":0.85}]
    total_potential_savings NUMERIC(18,2),
    currency VARCHAR(3) DEFAULT 'USD',
    confidence NUMERIC(5,2),
    model_version VARCHAR(50),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    applied_items INTEGER DEFAULT 0,
    notes TEXT
);

CREATE TABLE ai_safety_monitor_analyses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID REFERENCES projects(id),
    source_type VARCHAR(100) NOT NULL, -- camera_feed, photo_upload, wearable_sensor, report
    source_id VARCHAR(300),
    detection_type VARCHAR(100) NOT NULL, -- no_helmet, no_vest, restricted_zone, unsafe_behavior, equipment_misuse
    confidence NUMERIC(5,4),
    location VARCHAR(300),
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    image_url VARCHAR(500),
    reviewed BOOLEAN DEFAULT FALSE,
    reviewed_by VARCHAR(200),
    action_taken TEXT,
    severity VARCHAR(50), -- low, medium, high, critical
    notes TEXT
);

CREATE TABLE ai_report_generations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID REFERENCES projects(id),
    report_type VARCHAR(100) NOT NULL, -- daily_progress, weekly_summary, monthly_report, executive_dashboard
    parameters JSONB,
    generated_title VARCHAR(500),
    generated_summary TEXT,
    sections JSONB,
    charts JSONB,
    token_count INTEGER,
    model_used VARCHAR(200),
    generation_time_ms INTEGER,
    status VARCHAR(50) DEFAULT 'generated',
    feedback_score INTEGER, -- 1-5
    feedback_notes TEXT,
    created_by VARCHAR(200),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_aca_contract ON ai_contract_analyses(contract_id);
CREATE INDEX idx_apo_project ON ai_procurement_optimizations(project_id);
CREATE INDEX idx_asma_project ON ai_safety_monitor_analyses(project_id);
CREATE INDEX idx_arg_project ON ai_report_generations(project_id);