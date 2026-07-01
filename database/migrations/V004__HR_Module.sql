-- ============================================================================
-- V004__HR_Module.sql
-- Модуль управления персоналом (HR Management)
-- Часть OpenConstructionERP — Project Operating System
-- ============================================================================

-- ============================================================================
-- 1. Сотрудники
-- ============================================================================
CREATE TABLE employees (
    id              BIGSERIAL PRIMARY KEY,
    employee_code   VARCHAR(50) NOT NULL UNIQUE,          -- E-2026-001
    full_name       VARCHAR(300) NOT NULL,
    first_name      VARCHAR(100),
    last_name       VARCHAR(100),
    patronymic      VARCHAR(100),
    birth_date      DATE,
    gender          VARCHAR(10),
    nationality     VARCHAR(100),
    
    -- Контакты
    email           VARCHAR(200),
    phone           VARCHAR(50),
    phone_emergency VARCHAR(50),
    address         TEXT,
    
    -- Должность
    position        VARCHAR(200) NOT NULL,
    department      VARCHAR(200),
    position_type   VARCHAR(50) NOT NULL DEFAULT 'full_time', -- full_time, part_time, contract, seasonal
    position_category VARCHAR(50),                          -- worker, engineer, manager, executive, admin
    grade           VARCHAR(20),                             -- грейд/разряд
    
    -- Трудовые отношения
    status          VARCHAR(50) NOT NULL DEFAULT 'active',   -- active, suspended, maternity_leave, terminated
    hire_date       DATE NOT NULL,
    contract_end    DATE,
    termination_date DATE,
    termination_reason TEXT,
    
    -- Зарплата
    salary_base     NUMERIC(12,2),                          -- оклад
    salary_currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    hourly_rate     NUMERIC(10,2),                          -- часовая ставка
    bank_name       VARCHAR(200),
    bank_account    VARCHAR(100),
    tax_id          VARCHAR(50),
    social_security_id VARCHAR(50),
    
    -- Квалификация
    education       TEXT,
    certifications  TEXT,
    skills          TEXT,                                    -- JSON array
    experience_years INTEGER,
    
    -- Документы
    passport_number VARCHAR(50),
    passport_expiry DATE,
    work_permit     VARCHAR(50),
    work_permit_expiry DATE,
    medical_checkup_date DATE,
    medical_checkup_valid_until DATE,
    
    -- Медиа
    photo_path      VARCHAR(500),
    resume_path     VARCHAR(500),
    
    notes           TEXT,
    created_by      VARCHAR(100),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_employees_status ON employees(status);
CREATE INDEX idx_employees_department ON employees(department);
CREATE INDEX idx_employees_position ON employees(position);
CREATE INDEX idx_employees_hire_date ON employees(hire_date);

-- ============================================================================
-- 2. Отделы
-- ============================================================================
CREATE TABLE departments (
    id              BIGSERIAL PRIMARY KEY,
    code            VARCHAR(50) NOT NULL UNIQUE,
    name            VARCHAR(200) NOT NULL,
    description     TEXT,
    parent_id       BIGINT REFERENCES departments(id),
    head_employee_id BIGINT REFERENCES employees(id),
    cost_center     VARCHAR(50),
    location        VARCHAR(200),
    is_active       BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================================
-- 3. Табель рабочего времени
-- ============================================================================
CREATE TABLE time_attendance (
    id              BIGSERIAL PRIMARY KEY,
    employee_id     BIGINT NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    work_date       DATE NOT NULL,
    day_type        VARCHAR(20) NOT NULL DEFAULT 'workday', -- workday, weekend, holiday, vacation, sick_leave
    hours_worked    NUMERIC(4,1) DEFAULT 0,
    hours_overtime  NUMERIC(4,1) DEFAULT 0,
    status          VARCHAR(20) NOT NULL DEFAULT 'present',  -- present, absent, late, remote, business_trip
    reason          TEXT,
    approved_by     BIGINT REFERENCES employees(id),
    approved_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(employee_id, work_date)
);

CREATE INDEX idx_attendance_employee ON time_attendance(employee_id);
CREATE INDEX idx_attendance_date ON time_attendance(work_date);
CREATE INDEX idx_attendance_month ON time_attendance(employee_id, work_date);

-- ============================================================================
-- 4. Отпуска
-- ============================================================================
CREATE TABLE employee_leave (
    id              BIGSERIAL PRIMARY KEY,
    employee_id     BIGINT NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    leave_type      VARCHAR(50) NOT NULL,                   -- annual, sick, maternity, unpaid, study, special
    start_date      DATE NOT NULL,
    end_date        DATE NOT NULL,
    days_count      INTEGER NOT NULL,
    status          VARCHAR(50) NOT NULL DEFAULT 'pending', -- pending, approved, rejected, cancelled
    reason          TEXT,
    approved_by     BIGINT REFERENCES employees(id),
    approved_at     TIMESTAMPTZ,
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_leave_employee ON employee_leave(employee_id);
CREATE INDEX idx_leave_dates ON employee_leave(start_date, end_date);

-- ============================================================================
-- 5. Командировки
-- ============================================================================
CREATE TABLE employee_business_trips (
    id              BIGSERIAL PRIMARY KEY,
    employee_id     BIGINT NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    destination     VARCHAR(500) NOT NULL,
    purpose         TEXT NOT NULL,
    start_date      DATE NOT NULL,
    end_date        DATE NOT NULL,
    days_count      INTEGER NOT NULL,
    per_diem_amount NUMERIC(10,2),
    transport_cost  NUMERIC(12,2),
    accommodation_cost NUMERIC(12,2),
    total_cost      NUMERIC(12,2),
    status          VARCHAR(50) NOT NULL DEFAULT 'planned', -- planned, in_progress, completed, cancelled
    report_submitted BOOLEAN DEFAULT FALSE,
    approved_by     BIGINT REFERENCES employees(id),
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_trips_employee ON employee_business_trips(employee_id);

-- ============================================================================
-- 6. Начисления зарплаты
-- ============================================================================
CREATE TABLE payroll (
    id              BIGSERIAL PRIMARY KEY,
    employee_id     BIGINT NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    period_year     INTEGER NOT NULL,
    period_month    INTEGER NOT NULL,
    base_salary     NUMERIC(12,2),
    hours_worked    NUMERIC(4,1),
    overtime_pay    NUMERIC(12,2) DEFAULT 0,
    bonus           NUMERIC(12,2) DEFAULT 0,
    allowance       NUMERIC(12,2) DEFAULT 0,               -- надбавки
    per_diem        NUMERIC(12,2) DEFAULT 0,               -- суточные
    travel_allowance NUMERIC(12,2) DEFAULT 0,              -- командировочные
    gross_amount    NUMERIC(12,2),
    tax_amount      NUMERIC(12,2),
    insurance_amount NUMERIC(12,2),
    deductions      NUMERIC(12,2) DEFAULT 0,               -- удержания
    net_amount      NUMERIC(12,2),
    payment_date    DATE,
    status          VARCHAR(50) NOT NULL DEFAULT 'calculated', -- calculated, approved, paid, cancelled
    paid_at         TIMESTAMPTZ,
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(employee_id, period_year, period_month)
);

CREATE INDEX idx_payroll_employee ON payroll(employee_id);
CREATE INDEX idx_payroll_period ON payroll(period_year, period_month);

-- ============================================================================
-- 7. Охрана труда (HSE)
-- ============================================================================
CREATE TABLE hse_incidents (
    id              BIGSERIAL PRIMARY KEY,
    incident_number VARCHAR(50) NOT NULL UNIQUE,
    incident_type   VARCHAR(100) NOT NULL,                 -- accident, near_miss, fire, environmental, security
    severity        VARCHAR(50) NOT NULL,                   -- minor, moderate, serious, fatal
    incident_date   TIMESTAMPTZ NOT NULL,
    location        VARCHAR(500),
    description     TEXT NOT NULL,
    root_cause      TEXT,
    affected_employees INTEGER DEFAULT 0,
    lost_days       INTEGER DEFAULT 0,
    status          VARCHAR(50) NOT NULL DEFAULT 'reported', -- reported, investigating, resolved, closed
    reported_by     BIGINT REFERENCES employees(id),
    investigated_by BIGINT REFERENCES employees(id),
    resolution      TEXT,
    resolved_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_hse_date ON hse_incidents(incident_date);
CREATE INDEX idx_hse_type ON hse_incidents(incident_type);

-- ============================================================================
-- 8. Обучение
-- ============================================================================
CREATE TABLE employee_trainings (
    id              BIGSERIAL PRIMARY KEY,
    employee_id     BIGINT NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    training_name   VARCHAR(500) NOT NULL,
    training_type   VARCHAR(100) NOT NULL,                 -- safety, technical, management, certification
    provider        VARCHAR(200),
    start_date      DATE,
    end_date        DATE,
    hours           INTEGER,
    cost            NUMERIC(12,2),
    certificate_number VARCHAR(100),
    certificate_expiry DATE,
    status          VARCHAR(50) NOT NULL DEFAULT 'planned', -- planned, in_progress, completed, failed
    score           VARCHAR(20),
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_trainings_employee ON employee_trainings(employee_id);

-- ============================================================================
-- 9. Оценка персонала
-- ============================================================================
CREATE TABLE employee_performance (
    id              BIGSERIAL PRIMARY KEY,
    employee_id     BIGINT NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    review_period   VARCHAR(50) NOT NULL,                  -- Q1-2026, 2026-annual
    review_date     DATE NOT NULL,
    reviewer_id     BIGINT REFERENCES employees(id),
    overall_score   NUMERIC(3,1),                          -- 1.0 - 5.0
    criteria_scores TEXT,                                   -- JSON
    strengths       TEXT,
    improvements    TEXT,
    goals           TEXT,
    promotion_recommendation BOOLEAN DEFAULT FALSE,
    salary_review   NUMERIC(12,2),
    status          VARCHAR(50) NOT NULL DEFAULT 'draft',  -- draft, submitted, reviewed, completed
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(employee_id, review_period)
);

-- ============================================================================
-- 10. Триггеры
-- ============================================================================
CREATE OR REPLACE FUNCTION update_employee_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_employee_updated
    BEFORE UPDATE ON employees
    FOR EACH ROW
    EXECUTE FUNCTION update_employee_timestamp();

-- ============================================================================
-- Комментарии
-- ============================================================================
COMMENT ON TABLE employees IS 'Сотрудники / персонал';
COMMENT ON TABLE departments IS 'Отделы / подразделения';
COMMENT ON TABLE time_attendance IS 'Табель рабочего времени';
COMMENT ON TABLE employee_leave IS 'Отпуска';
COMMENT ON TABLE employee_business_trips IS 'Командировки';
COMMENT ON TABLE payroll IS 'Начисления зарплаты';
COMMENT ON TABLE hse_incidents IS 'Инциденты по охране труда';
COMMENT ON TABLE employee_trainings IS 'Обучение и повышение квалификации';
COMMENT ON TABLE employee_performance IS 'Оценка персонала';
