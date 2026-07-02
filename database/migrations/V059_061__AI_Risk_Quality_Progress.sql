-- ============================================================================
-- V059_061__AI_Risk_Quality_Progress.sql
-- Risk predictor, Quality inspector (photo defects), Progress monitor (photo)
-- ============================================================================
CREATE TABLE ai_risk_predictions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID REFERENCES projects(id),
    prediction_type VARCHAR(100) NOT NULL, -- cost_overrun, schedule_delay, safety_incident, quality_defect
    probability NUMERIC(5,4) NOT NULL,
    expected_impact NUMERIC(18,2),
    impact_currency VARCHAR(3) DEFAULT 'USD',
    affected_area VARCHAR(200),
    time_horizon VARCHAR(50), -- short_term, medium_term, long_term
    contributing_factors JSONB,
    mitigation_suggestions TEXT,
    model_version VARCHAR(50),
    prediction_date TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    accuracy_verified BOOLEAN DEFAULT FALSE,
    actual_outcome TEXT,
    notes TEXT
);

CREATE TABLE ai_quality_photo_analysis (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID REFERENCES projects(id),
    photo_url VARCHAR(500) NOT NULL,
    analysis_type VARCHAR(100) NOT NULL, -- segment_crack, weld_defect, spalling, honeycomb, geometry
    defects_found JSONB, -- [{"type":"crack","severity":"medium","location":"x:0.5,y:0.3","area_mm2":120}]
    defect_count INTEGER DEFAULT 0,
    overall_quality_score NUMERIC(5,2),
    confidence NUMERIC(5,4),
    model_version VARCHAR(50),
    analyzed_at TIMESTAMPTZ DEFAULT NOW(),
    reviewed BOOLEAN DEFAULT FALSE,
    reviewed_by VARCHAR(200),
    corrective_action TEXT,
    notes TEXT
);

CREATE TABLE ai_progress_photo_analysis (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID REFERENCES projects(id),
    photo_url VARCHAR(500) NOT NULL,
    photo_type VARCHAR(100), -- tunnel_face, segment_ring, portal, shaft, general
    analysis_type VARCHAR(100) NOT NULL, -- progress_pct, activity_detection, material_count, worker_count
    detected_activities JSONB, -- [{"activity":"segments_installed","count":3,"confidence":0.92}]
    completion_estimate NUMERIC(5,2),
    comparison_to_schedule NUMERIC(8,2), -- variance %
    worker_count INTEGER,
    equipment_count INTEGER,
    safety_violations INTEGER,
    confidence NUMERIC(5,4),
    model_version VARCHAR(50),
    analyzed_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_arp_project ON ai_risk_predictions(project_id);
CREATE INDEX idx_aqpa_project ON ai_quality_photo_analysis(project_id);
CREATE INDEX idx_appra_project ON ai_progress_photo_analysis(project_id);