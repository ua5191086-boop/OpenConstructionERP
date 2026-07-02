-- ============================================================================
-- V035__Training_Certifications_Module.sql
-- Модуль Training & Certifications — обучение и компетенции
-- Часть OpenConstructionERP — Project Operating System
-- ============================================================================

-- ============================================================================
-- 1. Курсы / программы обучения
-- ============================================================================
CREATE TABLE training_courses (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID REFERENCES projects(id) ON DELETE CASCADE,
    course_code     VARCHAR(30) NOT NULL,
    course_name     VARCHAR(500) NOT NULL,
    course_type     VARCHAR(50) NOT NULL DEFAULT 'safety'
        CHECK (course_type IN ('safety','technical','management','quality','hse','language','leadership','compliance','trade','other')),
    description     TEXT,
    provider        VARCHAR(300),                              -- external training company
    duration_hours  NUMERIC(6,2),
    duration_days   INTEGER,
    max_participants INTEGER,
    cost_per_person NUMERIC(12,2),
    currency        VARCHAR(3) DEFAULT 'USD',
    is_mandatory    BOOLEAN DEFAULT FALSE,
    validity_days   INTEGER DEFAULT 365,                       -- срок действия сертификата
    recertification_required BOOLEAN DEFAULT FALSE,
    status          VARCHAR(30) NOT NULL DEFAULT 'active'
        CHECK (status IN ('active','inactive','draft','archived')),
    syllabus        TEXT,
    learning_objectives TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, course_code)
);

CREATE INDEX idx_tr_course_project ON training_courses(project_id);
CREATE INDEX idx_tr_course_type ON training_courses(course_type);

COMMENT ON TABLE training_courses IS 'Курсы и программы обучения';

-- ============================================================================
-- 2. Сессии / проведения курсов
-- ============================================================================
CREATE TABLE training_sessions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    course_id       UUID NOT NULL REFERENCES training_courses(id) ON DELETE CASCADE,
    project_id      UUID REFERENCES projects(id) ON DELETE CASCADE,
    session_code    VARCHAR(30) NOT NULL,
    session_date    DATE NOT NULL,
    end_date        DATE,
    instructor      VARCHAR(300),
    location        VARCHAR(300),
    room            VARCHAR(100),
    max_participants INTEGER,
    actual_participants INTEGER DEFAULT 0,
    status          VARCHAR(30) NOT NULL DEFAULT 'planned'
        CHECK (status IN ('planned','in_progress','completed','cancelled','postponed')),
    completion_rate NUMERIC(5,2),
    feedback_score  NUMERIC(3,1),                               -- 1.0 - 5.0
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (course_id, session_code)
);

CREATE INDEX idx_tr_sess_course ON training_sessions(course_id);
CREATE INDEX idx_tr_sess_date ON training_sessions(session_date);

COMMENT ON TABLE training_sessions IS 'Сессии / проведения курсов';

-- ============================================================================
-- 3. Участники обучения
-- ============================================================================
CREATE TABLE training_participants (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id      UUID NOT NULL REFERENCES training_sessions(id) ON DELETE CASCADE,
    employee_id     UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    registration_date DATE NOT NULL DEFAULT CURRENT_DATE,
    attended        BOOLEAN DEFAULT FALSE,
    hours_attended  NUMERIC(5,2),
    score           NUMERIC(5,2),                              -- тест/экзамен (0-100)
    passed          BOOLEAN DEFAULT FALSE,
    certificate_number VARCHAR(100),
    certificate_issued DATE,
    certificate_expiry DATE,
    feedback        TEXT,
    status          VARCHAR(30) NOT NULL DEFAULT 'registered'
        CHECK (status IN ('registered','attended','completed','failed','no_show','cancelled')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (session_id, employee_id)
);

CREATE INDEX idx_tr_part_session ON training_participants(session_id);
CREATE INDEX idx_tr_part_employee ON training_participants(employee_id);
CREATE INDEX idx_tr_part_status ON training_participants(status);

COMMENT ON TABLE training_participants IS 'Участники обучения';

-- ============================================================================
-- 4. Сертификаты / квалификации сотрудников
-- ============================================================================
CREATE TABLE employee_certifications (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id     UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    participant_id  UUID REFERENCES training_participants(id) ON DELETE SET NULL,
    cert_code       VARCHAR(50) NOT NULL,
    cert_name       VARCHAR(500) NOT NULL,
    cert_type       VARCHAR(50) NOT NULL DEFAULT 'training'
        CHECK (cert_type IN ('training','license','competency','qualification','membership')),
    issuing_body    VARCHAR(300),
    cert_number     VARCHAR(100),
    issue_date      DATE NOT NULL,
    expiry_date     DATE,
    status          VARCHAR(30) NOT NULL DEFAULT 'active'
        CHECK (status IN ('active','expired','revoked','suspended')),
    attachments     JSONB DEFAULT '[]'::jsonb,
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (employee_id, cert_code)
);

CREATE INDEX idx_tr_cert_employee ON employee_certifications(employee_id);
CREATE INDEX idx_tr_cert_expiry ON employee_certifications(expiry_date) WHERE expiry_date IS NOT NULL;
CREATE INDEX idx_tr_cert_type ON employee_certifications(cert_type);

COMMENT ON TABLE employee_certifications IS 'Сертификаты и квалификации сотрудников';

-- ============================================================================
-- 5. Компетенции / навыки
-- ============================================================================
CREATE TABLE employee_competencies (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id     UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    competency_code VARCHAR(30) NOT NULL,
    competency_name VARCHAR(300) NOT NULL,
    category        VARCHAR(50) DEFAULT 'technical'
        CHECK (category IN ('technical','soft_skill','safety','management','language','regulatory')),
    proficiency_level VARCHAR(20) NOT NULL DEFAULT 'intermediate'
        CHECK (proficiency_level IN ('beginner','intermediate','advanced','expert','master')),
    years_experience NUMERIC(4,1),
    last_assessed   DATE,
    assessed_by     VARCHAR(200),
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (employee_id, competency_code)
);

CREATE INDEX idx_tr_comp_employee ON employee_competencies(employee_id);
CREATE INDEX idx_tr_comp_category ON employee_competencies(category);

COMMENT ON TABLE employee_competencies IS 'Компетенции и навыки сотрудников';

-- ============================================================================
-- 6. Сводка по обучению
-- ============================================================================
CREATE OR REPLACE VIEW training_summary AS
SELECT
    p.id AS project_id,
    (SELECT COUNT(*) FROM training_courses WHERE project_id = p.id) AS total_courses,
    (SELECT COUNT(*) FROM training_sessions WHERE project_id = p.id) AS total_sessions,
    (SELECT COUNT(*) FROM training_sessions WHERE project_id = p.id AND status = 'completed') AS completed_sessions,
    (SELECT COUNT(*) FROM training_participants tp JOIN training_sessions ts ON tp.session_id = ts.id WHERE ts.project_id = p.id) AS total_participants,
    (SELECT COUNT(*) FROM training_participants tp JOIN training_sessions ts ON tp.session_id = ts.id WHERE ts.project_id = p.id AND tp.passed = TRUE) AS passed,
    (SELECT AVG(tp.score) FROM training_participants tp JOIN training_sessions ts ON tp.session_id = ts.id WHERE ts.project_id = p.id AND tp.score IS NOT NULL) AS avg_score,
    (SELECT COUNT(*) FROM employee_certifications ec WHERE ec.expiry_date IS NOT NULL AND ec.expiry_date < CURRENT_DATE) AS expired_certs,
    (SELECT COUNT(*) FROM employee_competencies) AS total_competencies
FROM projects p;

COMMENT ON VIEW training_summary IS 'Сводка по обучению и сертификации';

-- ============================================================================
-- Register in object_types
-- ============================================================================
INSERT INTO object_types (code, name, icon, module_owner) VALUES
('training_course',         'Training Course',       'book-open',       'HR'),
('training_session',        'Training Session',      'calendar-check',  'HR'),
('training_participant',    'Participant',           'user-plus',       'HR'),
('employee_certification',  'Certification',         'award',           'HR'),
('employee_competency',     'Competency',            'star',            'HR')
ON CONFLICT (code) DO NOTHING;

COMMENT ON TABLE training_courses IS 'Курсы обучения';
COMMENT ON TABLE training_sessions IS 'Сессии проведения курсов';
COMMENT ON TABLE training_participants IS 'Участники обучения';
COMMENT ON TABLE employee_certifications IS 'Сертификаты сотрудников';
COMMENT ON TABLE employee_competencies IS 'Навыки и компетенции';