-- V082: Telegram Bot — Messages & Subscribers tables
-- Extends integration_telegram_bot with message queue and subscriber management

BEGIN;

-- Telegram bot messages (outgoing queue + incoming log)
CREATE TABLE IF NOT EXISTS integration_telegram_messages (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bot_id          UUID NOT NULL REFERENCES integration_telegram_bot(id) ON DELETE CASCADE,
    chat_id         VARCHAR(100) NOT NULL,
    text            TEXT NOT NULL,
    parse_mode      VARCHAR(20) DEFAULT 'HTML',
    direction       VARCHAR(20) NOT NULL DEFAULT 'outgoing'
                        CHECK (direction IN ('outgoing', 'incoming')),
    status          VARCHAR(30) NOT NULL DEFAULT 'queued'
                        CHECK (status IN ('queued', 'sent', 'delivered', 'failed', 'read')),
    telegram_msg_id BIGINT,
    error_message   TEXT,
    sent_at         TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_tgm_bot ON integration_telegram_messages(bot_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_tgm_status ON integration_telegram_messages(status) WHERE status = 'queued';

-- Telegram bot subscribers
CREATE TABLE IF NOT EXISTS integration_telegram_subscribers (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bot_id          UUID NOT NULL REFERENCES integration_telegram_bot(id) ON DELETE CASCADE,
    chat_id         VARCHAR(100) NOT NULL,
    username        VARCHAR(255),
    first_name      VARCHAR(255),
    last_name       VARCHAR(255),
    language_code   VARCHAR(10) DEFAULT 'ru',
    is_active       BOOLEAN NOT NULL DEFAULT true,
    subscribed_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (bot_id, chat_id)
);

CREATE INDEX IF NOT EXISTS idx_tgs_bot ON integration_telegram_subscribers(bot_id, is_active);

COMMIT;
