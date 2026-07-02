-- ============================================================================
-- V066_068__AI_Chatbot_Predictive_Maintenance_ESG.sql
-- RAG-чат, Predictive Maintenance TBM, ESG Reporter
-- ============================================================================
CREATE TABLE ai_chatbot_interactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID REFERENCES ai_sessions(id),
    user_query TEXT NOT NULL,
    retrieved_chunks JSONB,
    used_functions JSONB,
    generated_response TEXT,
    token_count_input INTEGER,
    token_count_output INTEGER,
    latency_ms INTEGER,
    user_feedback VARCHAR(50), -- helpful, not_helpful, incorrect
    user_rating INTEGER, -- 1-5
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE ai_predictive_maintenance_models (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    equipment_type VARCHAR(100) NOT NULL, -- tbm, conveyor, crane, locomotive, vent_fan, pump
    model_type VARCHAR(100) NOT NULL, -- remaining_useful_life, failure_probability, anomaly_detection
    features_used JSONB,
    model_accuracy NUMERIC(8,6),
    training_period_from DATE,
    training_period_to DATE,
    last_retrained_at TIMESTAMPTZ,
    status VARCHAR(50) DEFAULT 'active',
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE ai_predictive_maintenance_predictions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    model_id UUID REFERENCES ai_predictive_maintenance_models(id),
    equipment_id UUID REFERENCES equipment(id),
    prediction_type VARCHAR(100) NOT NULL,
    predicted_remaining_hours NUMERIC(10,2),
    failure_probability NUMERIC(5,4),
    recommended_action TEXT,
    recommended_window_from TIMESTAMPTZ,
    recommended_window_to TIMESTAMPTZ,
    confidence NUMERIC(5,2),
    actual_failure BOOLEAN,
    actual_failure_at TIMESTAMPTZ,
    prediction_made_at TIMESTAMPTZ DEFAULT NOW(),
    status VARCHAR(50) DEFAULT 'active'
);

CREATE TABLE ai_esg_report_analyses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID REFERENCES projects(id),
    report_period VARCHAR(50) NOT NULL,
    environmental_data JSONB,
    social_data JSONB,
    governance_data JSONB,
    co2_total NUMERIC(18,4),
    co2_vs_target NUMERIC(8,2),
    esg_score NUMERIC(5,2),
    recommendations TEXT,
    report_generated TEXT,
    model_version VARCHAR(50),
    generated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_aci_session ON ai_chatbot_interactions(session_id);
CREATE INDEX idx_apmp_eq ON ai_predictive_maintenance_predictions(equipment_id);
CREATE INDEX idx_aera_project ON ai_esg_report_analyses(project_id);