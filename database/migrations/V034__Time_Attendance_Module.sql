-- ============================================================================
-- V034__Time_Attendance_Module.sql
-- Модуль Time & Attendance — учёт рабочего времени и посещаемости
-- Часть OpenConstructionERP — Project Operating System
-- ============================================================================

-- ============================================================================
-- 1. Календари / смены
-- ============================================================================
CREATE TABLE attendance_calendars (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID REFERENCES projects(id) ON DELETE CASCADE,
    calendar_code   VARCHAR(30) NOT NULL,
    calendar_name   VARCHAR(300) NOT NULL,
    calendar_type   VARCHAR(30) NOT NULL DEFAULT 'standard'
        CHECK (calendar_type IN ('standard','shift','flexible','compressed')),
    work_days       INTEGER[] DEFAULT '{1,2,3,4,5}',        -- 1=Mon..7=Sun
    work_hours_per_day NUMERIC(4,2) DEFAULT 8.0,
    start_time      TIME DEFAULT '08:00',
    end_time        TIME DEFAULT '17:00',
    break_minutes   INTEGER DEFAULT 60,
    is_active       BOOLEAN DEFAULT TRUE,
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, calendar_code)
);

CREATE INDEX idx_att_cal_project ON attendance_calendars(project_id);

COMMENT ON TABLE attendance_calendars IS 'Календари / смены — графики работы';

-- ============================================================================
-- 2. Сотрудники / участники (расширение employees)
-- ============================================================================
CREATE TABLE attendance_employees (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id     UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    project_id      UUID REFERENCES projects(id) ON DELETE CASCADE,
    calendar_id     UUID REFERENCES attendance_calendars(id),
    badge_number    VARCHAR(50),
    biometric_id    VARCHAR(100),                             -- Fingerprint / Face ID
    rfid_card       VARCHAR(50),
    hire_date       DATE,
    termination_date DATE,
    employment_type VARCHAR(30) DEFAULT 'full_time'
        CHECK (employment_type IN ('full_time','part_time','contractor','temporary','intern')),
    hourly_rate     NUMERIC(10,2),
    overtime_rate   NUMERIC(5,2) DEFAULT 1.5,
    is_active       BOOLEAN DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (employee_id, project_id)
);

CREATE INDEX idx_att_emp_project ON attendance_employees(project_id);
CREATE INDEX idx_att_emp_badge ON attendance_employees(badge_number);

COMMENT ON TABLE attendance_employees IS 'Сотрудники для учёта рабочего времени';

-- ============================================================================
-- 3. Табели / Timesheets
-- ============================================================================
CREATE TABLE attendance_timesheets (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    employee_id     UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    timesheet_date  DATE NOT NULL,
    day_of_week     INTEGER NOT NULL,                         -- 1=Mon..7=Sun
    clock_in        TIMESTAMPTZ,
    clock_out       TIMESTAMPTZ,
    break_start     TIMESTAMPTZ,
    break_end       TIMESTAMPTZ,
    hours_worked    NUMERIC(5,2) DEFAULT 0,
    hours_regular   NUMERIC(5,2) DEFAULT 0,
    hours_overtime  NUMERIC(5,2) DEFAULT 0,
    hours_absence   NUMERIC(5,2) DEFAULT 0,                  -- absenteeism
    absence_type    VARCHAR(50),                               -- sick, vacation, personal, unpaid, holiday, training
    status          VARCHAR(30) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending','approved','rejected','correction')),
    approved_by     UUID REFERENCES employees(id),
    approved_at     TIMESTAMPTZ,
    source          VARCHAR(50) DEFAULT 'manual'              -- manual, biometric, rfid, mobile, api
        CHECK (source IN ('manual','biometric','rfid','mobile','api','import')),
    location        POINT,                                    -- GPS при clock-in/out
    device_id       VARCHAR(100),
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (employee_id, timesheet_date)
);

CREATE INDEX idx_att_ts_project ON attendance_timesheets(project_id);
CREATE INDEX idx_att_ts_employee ON attendance_timesheets(employee_id);
CREATE INDEX idx_att_ts_date ON attendance_timesheets(timesheet_date);
CREATE INDEX idx_att_ts_status ON attendance_timesheets(project_id, status);

COMMENT ON TABLE attendance_timesheets IS 'Табели учёта рабочего времени';

-- ============================================================================
-- 4. Биометрические события
-- ============================================================================
CREATE TABLE attendance_biometric_events (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id     UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    project_id      UUID REFERENCES projects(id) ON DELETE CASCADE,
    event_time      TIMESTAMPTZ NOT NULL,
    event_type      VARCHAR(30) NOT NULL
        CHECK (event_type IN ('clock_in','clock_out','break_start','break_end','gate_entry','gate_exit')),
    device_id       VARCHAR(100),
    device_type     VARCHAR(50) DEFAULT 'fingerprint'         -- fingerprint, face, rfid, card, mobile
        CHECK (device_type IN ('fingerprint','face','rfid','card','mobile','manual')),
    location        POINT,
    temperature     NUMERIC(4,1),                             -- temperature screening
    verified        BOOLEAN DEFAULT TRUE,
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_att_be_employee ON attendance_biometric_events(employee_id);
CREATE INDEX idx_att_be_time ON attendance_biometric_events(event_time);
CREATE INDEX idx_att_be_type ON attendance_biometric_events(event_type);

COMMENT ON TABLE attendance_biometric_events IS 'Биометрические события — clock-in/out, gate access';

-- ============================================================================
-- 5. Отсутствия / Absences
-- ============================================================================
CREATE TABLE attendance_absences (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id     UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    project_id      UUID REFERENCES projects(id) ON DELETE CASCADE,
    absence_type    VARCHAR(50) NOT NULL
        CHECK (absence_type IN ('sick','vacation','personal','unpaid','holiday','training','maternity','military','bereavement','other')),
    start_date      DATE NOT NULL,
    end_date        DATE NOT NULL,
    total_days      INTEGER NOT NULL,
    reason          TEXT,
    document_path   VARCHAR(1000),
    status          VARCHAR(30) NOT NULL DEFAULT 'pending'
        CHECK (status IN ('pending','approved','rejected','cancelled')),
    approved_by     UUID REFERENCES employees(id),
    approved_at     TIMESTAMPTZ,
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_att_abs_employee ON attendance_absences(employee_id);
CREATE INDEX idx_att_abs_type ON attendance_absences(absence_type);
CREATE INDEX idx_att_abs_status ON attendance_absences(project_id, status);

COMMENT ON TABLE attendance_absences IS 'Отсутствия — отпуска, больничные, отгулы';

-- ============================================================================
-- 6. Gate Access / Пропускной режим
-- ============================================================================
CREATE TABLE attendance_gate_log (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id      UUID REFERENCES projects(id) ON DELETE CASCADE,
    employee_id     UUID REFERENCES employees(id) ON DELETE SET NULL,
    visitor_name    VARCHAR(300),
    visitor_company VARCHAR(300),
    id_document     VARCHAR(100),
    gate            VARCHAR(100),
    direction       VARCHAR(10) NOT NULL CHECK (direction IN ('entry','exit')),
    event_time      TIMESTAMPTZ NOT NULL,
    vehicle_plate   VARCHAR(50),
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_att_gl_project ON attendance_gate_log(project_id);
CREATE INDEX idx_att_gl_time ON attendance_gate_log(event_time);

COMMENT ON TABLE attendance_gate_log IS 'Журнал пропускного режима';

-- ============================================================================
-- 7. Сводка по посещаемости
-- ============================================================================
CREATE VIEW attendance_summary AS
SELECT
    p.id AS project_id,
    COUNT(DISTINCT ae.employee_id) AS registered_employees,
    COUNT(DISTINCT ae.employee_id) FILTER (WHERE ae.is_active = TRUE) AS active_employees,
    COUNT(DISTINCT at.employee_id) FILTER (WHERE at.timesheet_date >= CURRENT_DATE - INTERVAL '30 days') AS active_last_30d,
    COALESCE(SUM(at.hours_worked) FILTER (WHERE at.timesheet_date >= CURRENT_DATE - INTERVAL '30 days'), 0) AS total_hours_30d,
    COALESCE(SUM(at.hours_overtime) FILTER (WHERE at.timesheet_date >= CURRENT_DATE - INTERVAL '30 days'), 0) AS total_overtime_30d,
    COALESCE(SUM(at.hours_absence) FILTER (WHERE at.timesheet_date >= CURRENT_DATE - INTERVAL '30 days'), 0) AS total_absence_30d,
    COALESCE(AVG(at.hours_worked) FILTER (WHERE at.timesheet_date >= CURRENT_DATE - INTERVAL '30 days'), 0) AS avg_hours_per_day,
    COUNT(DISTINCT aa.id) FILTER (WHERE aa.start_date <= CURRENT_DATE AND aa.end_date >= CURRENT_DATE) AS active_absences,
    COUNT(DISTINCT agl.id) FILTER (WHERE agl.event_time >= CURRENT_DATE - INTERVAL '7 days') AS gate_events_7d
FROM projects p
LEFT JOIN attendance_employees ae ON ae.project_id = p.id
LEFT JOIN attendance_timesheets at ON at.project_id = p.id
LEFT JOIN attendance_absences aa ON aa.project_id = p.id
LEFT JOIN attendance_gate_log agl ON agl.project_id = p.id
GROUP BY p.id;

COMMENT ON VIEW attendance_summary IS 'Сводка по посещаемости — активность, часы, отсутствия';

-- ============================================================================
-- Register in object_types
-- ============================================================================
INSERT INTO object_types (code, name, icon, module_owner) VALUES
('attendance_calendar',     'Calendar',           'calendar',       'HR'),
('attendance_timesheet',    'Timesheet',          'clock',          'HR'),
('attendance_employee',     'Attendance Employee','user-check',     'HR'),
('attendance_absence',      'Absence',            'calendar-x',     'HR'),
('attendance_biometric',    'Biometric Event',    'fingerprint',    'HR'),
('attendance_gate',         'Gate Log',           'door-open',      'HR')
ON CONFLICT (code) DO NOTHING;

COMMENT ON TABLE attendance_calendars IS 'Графики работы / смены';
COMMENT ON TABLE attendance_employees IS 'Сотрудники с badge/биометрией';
COMMENT ON TABLE attendance_timesheets IS 'Табели рабочего времени';
COMMENT ON TABLE attendance_biometric_events IS 'Биометрические события доступа';
COMMENT ON TABLE attendance_absences IS 'Отсутствия и отпуска';
COMMENT ON TABLE attendance_gate_log IS 'Пропускной режим / Gate Access';