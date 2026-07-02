-- ============================================================================
-- V075__Lessons_Learned.sql
-- База знаний, post-mortem, replay
-- ============================================================================
CREATE TABLE lessons_learned (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id),
    lesson_category VARCHAR(100) NOT NULL, -- design, construction, procurement, management, safety, quality, financial
    lesson_type VARCHAR(50) NOT NULL, -- positive, negative, improvement, innovation
    title VARCHAR(500) NOT NULL,
    description TEXT NOT NULL,
    root_cause TEXT,
    impact TEXT,
    recommendation TEXT,
    phase VARCHAR(100), -- planning, design, procurement, construction, commissioning
    related_entity_type VARCHAR(100),
    related_entity_id UUID,
    tags JSONB DEFAULT '[]'::JSONB,
    severity VARCHAR(50) DEFAULT 'medium',
    recurrence_risk VARCHAR(50) DEFAULT 'medium',
    shared_with_team BOOLEAN DEFAULT FALSE,
    shared_at TIMESTAMPTZ,
    submitted_by VARCHAR(200),
    reviewed_by VARCHAR(200),
    reviewed_at TIMESTAMPTZ,
    status VARCHAR(50) DEFAULT 'draft',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE lessons_learned_reviews (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    lesson_id UUID NOT NULL REFERENCES lessons_learned(id),
    review_type VARCHAR(100) NOT NULL, -- peer_review, management_review, expert_review
    reviewer VARCHAR(300) NOT NULL,
    comments TEXT,
    recommendation VARCHAR(50), -- approve, reject, modify
    action_items TEXT,
    reviewed_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_ll_project ON lessons_learned(project_id);
CREATE INDEX idx_ll_category ON lessons_learned(lesson_category);
CREATE INDEX idx_ll_type ON lessons_learned(lesson_type);
CREATE INDEX idx_ll_tags ON lessons_learned USING gin(tags);
CREATE INDEX idx_llr_lesson ON lessons_learned_reviews(lesson_id);