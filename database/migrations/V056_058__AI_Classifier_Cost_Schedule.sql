-- ============================================================================
-- V056__AI_Document_Classifier.sql
-- Автоклассификация входящих документов по типу, проекту, контракту
-- ============================================================================
CREATE TABLE ai_document_classifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID NOT NULL,
    document_type VARCHAR(100) NOT NULL, -- contract, invoice, change_order, report, photo, drawing
    predicted_type VARCHAR(100),
    confidence NUMERIC(5,4),
    predicted_project UUID,
    predicted_contract UUID,
    predicted_category VARCHAR(100),
    extracted_fields JSONB,
    model_version VARCHAR(50),
    processing_time_ms INTEGER,
    reviewed BOOLEAN DEFAULT FALSE,
    corrected_type VARCHAR(100),
    reviewed_by VARCHAR(200),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- V057__AI_Cost_Estimator.sql
-- Оценка стоимости по аналогам, параметрическая
-- ============================================================================
CREATE TABLE ai_cost_models (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    model_name VARCHAR(300) NOT NULL,
    model_type VARCHAR(50) NOT NULL, -- parametric, analog, ml_regression, neural_network
    target_variable VARCHAR(100), -- cost_per_m, total_cost, duration
    features JSONB, -- ["tunnel_length","diameter","geology","depth"]
    coefficients JSONB,
    r2_score NUMERIC(8,6),
    mae NUMERIC(18,4),
    training_data_count INTEGER,
    training_date TIMESTAMPTZ,
    last_used_at TIMESTAMPTZ,
    status VARCHAR(50) DEFAULT 'active',
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE ai_cost_estimates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID REFERENCES projects(id),
    model_id UUID REFERENCES ai_cost_models(id),
    estimate_type VARCHAR(100) NOT NULL, -- boq_item, work_package, contract, total_project
    input_params JSONB NOT NULL,
    estimated_value NUMERIC(18,2) NOT NULL,
    confidence_pct NUMERIC(5,2),
    p10_value NUMERIC(18,2),
    p50_value NUMERIC(18,2),
    p90_value NUMERIC(18,2),
    currency VARCHAR(3) DEFAULT 'USD',
    comparable_projects INTEGER,
    created_by VARCHAR(200),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================================
-- V058__AI_Schedule_Optimizer.sql
-- Оптимизация расписания ML — предложения по duration, resources, critical path
-- ============================================================================
CREATE TABLE ai_schedule_optimizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID REFERENCES projects(id),
    schedule_version VARCHAR(50),
    optimization_type VARCHAR(100) NOT NULL, -- duration_compression, resource_leveling, cost_tradeoff, risk_adjusted
    original_duration INTEGER,
    optimized_duration INTEGER,
    original_cost NUMERIC(18,2),
    optimized_cost NUMERIC(18,2),
    compression_activities JSONB, -- activities recommended for crashing/fast-tracking
    resource_recommendations JSONB,
    risk_adjustments JSONB,
    confidence_score NUMERIC(5,2),
    status VARCHAR(50) DEFAULT 'draft',
    applied BOOLEAN DEFAULT FALSE,
    applied_at TIMESTAMPTZ,
    created_by VARCHAR(200),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_adc_doc ON ai_document_classifications(document_id);
CREATE INDEX idx_ace_project ON ai_cost_estimates(project_id);
CREATE INDEX idx_aso_project ON ai_schedule_optimizations(project_id);