-- ============================================================
-- V024: Auth & Audit Log
-- Creates auth_audit_log table and trigger for change logging
-- ============================================================

-- -----------------------------------------------------------
-- 1. AUDIT LOG TABLE
-- -----------------------------------------------------------
CREATE TABLE IF NOT EXISTS auth_audit_log (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         VARCHAR(255) NOT NULL,
    username        VARCHAR(255) NOT NULL DEFAULT '',
    action          VARCHAR(100) NOT NULL,
    resource        VARCHAR(500) NOT NULL,
    resource_id     VARCHAR(255),
    ip_address      VARCHAR(45),
    user_agent      TEXT,
    old_values      JSONB,
    new_values      JSONB,
    outcome         VARCHAR(20) NOT NULL DEFAULT 'success',
    error_message   TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_auth_audit_log_user_id ON auth_audit_log(user_id);
CREATE INDEX idx_auth_audit_log_action ON auth_audit_log(action);
CREATE INDEX idx_auth_audit_log_resource ON auth_audit_log(resource);
CREATE INDEX idx_auth_audit_log_created_at ON auth_audit_log(created_at DESC);
CREATE INDEX idx_auth_audit_log_outcome ON auth_audit_log(outcome);

-- -----------------------------------------------------------
-- 2. AUDIT FUNCTION (generic change logger)
-- -----------------------------------------------------------
CREATE OR REPLACE FUNCTION fn_audit_log_change()
RETURNS TRIGGER AS $$
DECLARE
    v_user_id   VARCHAR(255);
    v_username  VARCHAR(255);
    v_action    VARCHAR(100);
    v_resource  VARCHAR(500);
    v_resource_id VARCHAR(255);
    v_old_json  JSONB;
    v_new_json  JSONB;
BEGIN
    -- Attempt to read user context set by application
    v_user_id   := COALESCE(current_setting('app.current_user_id', TRUE), 'system');
    v_username  := COALESCE(current_setting('app.current_username', TRUE), 'system');

    v_resource := TG_TABLE_SCHEMA || '.' || TG_TABLE_NAME;

    IF TG_OP = 'INSERT' THEN
        v_action := 'CREATE';
        v_resource_id := NEW.id::VARCHAR;
        v_old_json := NULL;
        v_new_json := to_jsonb(NEW);
    ELSIF TG_OP = 'UPDATE' THEN
        v_action := 'UPDATE';
        v_resource_id := NEW.id::VARCHAR;
        v_old_json := to_jsonb(OLD);
        v_new_json := to_jsonb(NEW);
    ELSIF TG_OP = 'DELETE' THEN
        v_action := 'DELETE';
        v_resource_id := OLD.id::VARCHAR;
        v_old_json := to_jsonb(OLD);
        v_new_json := NULL;
    ELSE
        RETURN NULL;
    END IF;

    INSERT INTO auth_audit_log (user_id, username, action, resource, resource_id, old_values, new_values)
    VALUES (v_user_id, v_username, v_action, v_resource, v_resource_id, v_old_json, v_new_json);

    RETURN NULL;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- -----------------------------------------------------------
-- 3. SAMPLE TRIGGERS on core business tables
--    (Extend as needed for each module)
-- -----------------------------------------------------------

-- Example: BOQ items audit
DROP TRIGGER IF EXISTS trg_audit_boq_items ON boq_items;
CREATE TRIGGER trg_audit_boq_items
    AFTER INSERT OR UPDATE OR DELETE ON boq_items
    FOR EACH ROW EXECUTE FUNCTION fn_audit_log_change();

-- Example: Tenders audit
DROP TRIGGER IF EXISTS trg_audit_tenders ON tenders;
CREATE TRIGGER trg_audit_tenders
    AFTER INSERT OR UPDATE OR DELETE ON tenders
    FOR EACH ROW EXECUTE FUNCTION fn_audit_log_change();

-- Example: Contracts audit
DROP TRIGGER IF EXISTS trg_audit_contracts ON contracts;
CREATE TRIGGER trg_audit_contracts
    AFTER INSERT OR UPDATE OR DELETE ON contracts
    FOR EACH ROW EXECUTE FUNCTION fn_audit_log_change();

-- Example: HR employees audit
DROP TRIGGER IF EXISTS trg_audit_hr_employees ON employees;
CREATE TRIGGER trg_audit_hr_employees
    AFTER INSERT OR UPDATE OR DELETE ON employees
    FOR EACH ROW EXECUTE FUNCTION fn_audit_log_change();

-- Example: PM Projects audit
DROP TRIGGER IF EXISTS trg_audit_pm_projects ON projects;
CREATE TRIGGER trg_audit_pm_projects
    AFTER INSERT OR UPDATE OR DELETE ON projects
    FOR EACH ROW EXECUTE FUNCTION fn_audit_log_change();

-- -----------------------------------------------------------
-- 4. AUTH LOGIN ATTEMPT LOG (separate lightweight table)
-- -----------------------------------------------------------
CREATE TABLE IF NOT EXISTS auth_login_log (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username        VARCHAR(255) NOT NULL,
    ip_address      VARCHAR(45),
    user_agent      TEXT,
    success         BOOLEAN NOT NULL,
    failure_reason  VARCHAR(255),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_auth_login_log_username ON auth_login_log(username);
CREATE INDEX idx_auth_login_log_created_at ON auth_login_log(created_at DESC);