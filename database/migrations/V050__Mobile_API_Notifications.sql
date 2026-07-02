-- ============================================================================
-- V050__Mobile_API_Notifications.sql
-- Mobile API endpoints, push notifications, activity feed
-- Часть OpenConstructionERP — Project Operating System
-- ============================================================================

-- ============================================================================
-- 1. Push-уведомления
-- ============================================================================
CREATE TABLE push_notifications (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id          UUID REFERENCES projects(id) ON DELETE CASCADE,
    notification_type   VARCHAR(100) NOT NULL,                  -- task_assigned, approval_required, deadline_approaching, incident, change_order, document_upload
    title               VARCHAR(500) NOT NULL,
    body                TEXT NOT NULL,
    priority            VARCHAR(20) DEFAULT 'normal',           -- low, normal, high, urgent
    deep_link           VARCHAR(500),                            -- app://screen/id
    sender_id           VARCHAR(200),
    recipient_id        VARCHAR(200) NOT NULL,
    recipient_role      VARCHAR(100),
    read_at             TIMESTAMPTZ,
    delivered_at        TIMESTAMPTZ,
    action_taken_at     TIMESTAMPTZ,
    action_type         VARCHAR(50),                             -- approved, rejected, viewed, commented
    status              VARCHAR(50) NOT NULL DEFAULT 'pending', -- pending, sent, delivered, read, actioned, failed
    error_message       TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_push_notifications_recipient ON push_notifications(recipient_id);
CREATE INDEX idx_push_notifications_status ON push_notifications(status);
CREATE INDEX idx_push_notifications_type ON push_notifications(notification_type);
CREATE INDEX idx_push_notifications_created ON push_notifications(created_at);

COMMENT ON TABLE push_notifications IS 'Push-уведомления для мобильных устройств';

-- ============================================================================
-- 2. Устройства пользователей (FCM/APNS tokens)
-- ============================================================================
CREATE TABLE user_devices (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             VARCHAR(200) NOT NULL,
    device_name         VARCHAR(300),
    device_type         VARCHAR(50) NOT NULL,                   -- ios, android, web, desktop
    push_token          TEXT NOT NULL,
    platform            VARCHAR(50),                            -- fcm, apns, web_push
    app_version         VARCHAR(50),
    os_version          VARCHAR(50),
    is_active           BOOLEAN DEFAULT TRUE,
    last_seen_at        TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_user_devices_user ON user_devices(user_id);
CREATE INDEX idx_user_devices_active ON user_devices(is_active) WHERE is_active = TRUE;

COMMENT ON TABLE user_devices IS 'Зарегистрированные устройства пользователей';

-- ============================================================================
-- 3. Activity Feed (лента активности)
-- ============================================================================
CREATE TABLE activity_feed (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id          UUID REFERENCES projects(id) ON DELETE CASCADE,
    activity_type       VARCHAR(100) NOT NULL,                  -- document_upload, status_change, comment, approval, task_complete, milestone
    activity_title      VARCHAR(500) NOT NULL,
    activity_description TEXT,
    entity_type         VARCHAR(100),                            -- contract, document, task, boq, incident
    entity_id           UUID,
    actor_id            VARCHAR(200) NOT NULL,
    actor_name          VARCHAR(300),
    actor_avatar        VARCHAR(500),
    metadata            JSONB,
    importance          VARCHAR(20) DEFAULT 'normal',           -- low, normal, high, milestone
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_activity_feed_project ON activity_feed(project_id);
CREATE INDEX idx_activity_feed_type ON activity_feed(activity_type);
CREATE INDEX idx_activity_feed_actor ON activity_feed(actor_id);
CREATE INDEX idx_activity_feed_created ON activity_feed(created_at);

COMMENT ON TABLE activity_feed IS 'Лента активности проекта для мобильных и веб';

-- ============================================================================
-- 4. Комментарии (универсальные к любой сущности)
-- ============================================================================
CREATE TABLE universal_comments (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id          UUID REFERENCES projects(id) ON DELETE CASCADE,
    entity_type         VARCHAR(100) NOT NULL,
    entity_id           UUID NOT NULL,
    parent_id           UUID REFERENCES universal_comments(id) ON DELETE CASCADE,
    author_id           VARCHAR(200) NOT NULL,
    author_name         VARCHAR(300),
    content             TEXT NOT NULL,
    attachments         JSONB DEFAULT '[]'::JSONB,
    mentions            JSONB DEFAULT '[]'::JSONB,              -- упомянутые пользователи
    is_edited           BOOLEAN DEFAULT FALSE,
    is_pinned           BOOLEAN DEFAULT FALSE,
    is_archived         BOOLEAN DEFAULT FALSE,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_universal_comments_entity ON universal_comments(entity_type, entity_id);
CREATE INDEX idx_universal_comments_project ON universal_comments(project_id);
CREATE INDEX idx_universal_comments_author ON universal_comments(author_id);
CREATE INDEX idx_universal_comments_pinned ON universal_comments(is_pinned) WHERE is_pinned = TRUE;

COMMENT ON TABLE universal_comments IS 'Универсальные комментарии к любой сущности системы';