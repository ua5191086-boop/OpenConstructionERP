-- ============================================================================
-- V008__AI_Module.sql
-- Модуль искусственного интеллекта (AI/ML Management)
-- Часть OpenConstructionERP — Project Operating System
-- ============================================================================

-- ============================================================================
-- 1. AI-агенты
-- ============================================================================
CREATE TABLE ai_agents (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_name      VARCHAR(200) NOT NULL UNIQUE,
    agent_type      VARCHAR(100) NOT NULL,                 -- classifier, extractor, predictor, recommender, chatbot
    description     TEXT,
    model_name      VARCHAR(200),                          -- deepseek, gpt-4, claude, llama
    model_provider  VARCHAR(100),                           -- openai, anthropic, deepseek, ollama
    system_prompt   TEXT,
    temperature     NUMERIC(3,2) DEFAULT 0.7,
    max_tokens      INTEGER DEFAULT 4096,
    is_active       BOOLEAN DEFAULT TRUE,
    version         VARCHAR(50) DEFAULT '1.0',
    config          JSONB,                                  -- дополнительные настройки
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================================
-- 2. AI-задачи
-- ============================================================================
CREATE TABLE ai_tasks (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id        UUID REFERENCES ai_agents(id),
    task_type       VARCHAR(100) NOT NULL,                 -- classification, extraction, generation, analysis, prediction
    input_data      TEXT NOT NULL,
    input_format    VARCHAR(50) DEFAULT 'text',             -- text, json, image, document
    output_data     TEXT,
    output_format   VARCHAR(50) DEFAULT 'text',
    confidence      NUMERIC(5,2),                          -- 0-100%
    tokens_used     INTEGER,
    cost            NUMERIC(10,6),                          -- стоимость запроса
    processing_time_ms INTEGER,                             -- время обработки
    status          VARCHAR(50) NOT NULL DEFAULT 'pending', -- pending, processing, completed, failed
    error_message   TEXT,
    source_type     VARCHAR(50),                            -- boq, tender, contract, hr, finance, procurement, bim
    source_id       UUID,
    created_by      VARCHAR(100),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at    TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_ai_tasks_agent ON ai_tasks(agent_id);
CREATE INDEX IF NOT EXISTS idx_ai_tasks_type ON ai_tasks(task_type);
CREATE INDEX IF NOT EXISTS idx_ai_tasks_status ON ai_tasks(status);
CREATE INDEX IF NOT EXISTS idx_ai_tasks_source ON ai_tasks(source_type, source_id);
CREATE INDEX IF NOT EXISTS idx_ai_tasks_created ON ai_tasks(created_at);

-- ============================================================================
-- 3. Классификация запросов (Ruslan OS dispatcher)
-- ============================================================================
CREATE TABLE ai_classifications (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id         UUID REFERENCES ai_tasks(id) ON DELETE CASCADE,
    raw_input       TEXT NOT NULL,
    intent          VARCHAR(200),                           -- намерение пользователя
    entities        JSONB,                                  -- извлечённые сущности
    confidence      NUMERIC(5,2),
    suggested_action VARCHAR(200),                          -- рекомендуемое действие
    suggested_module VARCHAR(100),                          -- boq, tender, contract, hr, finance, procurement, bim
    parameters      JSONB,                                  -- параметры для n8n
    is_confirmed    BOOLEAN DEFAULT FALSE,                  -- подтверждено пользователем
    confirmed_by    VARCHAR(100),
    confirmed_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================================
-- 4. Извлечение данных (OCR / парсинг)
-- ============================================================================
CREATE TABLE ai_extractions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id         UUID REFERENCES ai_tasks(id) ON DELETE CASCADE,
    document_type   VARCHAR(100),                          -- invoice, contract, boq, report, drawing
    file_path       VARCHAR(1000),
    extracted_data  JSONB,                                  -- извлечённые данные
    fields_count    INTEGER,
    accuracy        NUMERIC(5,2),                           -- точность извлечения
    validated       BOOLEAN DEFAULT FALSE,
    validated_by    UUID REFERENCES employees(id),
    validated_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================================
-- 5. Прогнозы / предсказания
-- ============================================================================
CREATE TABLE ai_predictions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id         UUID REFERENCES ai_tasks(id) ON DELETE CASCADE,
    prediction_type VARCHAR(100) NOT NULL,                  -- cost_forecast, delay_risk, quality_risk, budget_overrun
    target_entity   VARCHAR(100),                           -- project, contract, section
    target_id       UUID,
    predicted_value NUMERIC(18,2),
    confidence      NUMERIC(5,2),
    actual_value    NUMERIC(18,2),
    accuracy        NUMERIC(5,2),                           -- точность после верификации
    features_used   TEXT,                                   -- какие признаки использовались
    model_version   VARCHAR(50),
    prediction_date DATE NOT NULL,
    verified_date   DATE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ai_predictions_type ON ai_predictions(prediction_type);
CREATE INDEX IF NOT EXISTS idx_ai_predictions_target ON ai_predictions(target_entity, target_id);

-- ============================================================================
-- 6. Рекомендации
-- ============================================================================
CREATE TABLE ai_recommendations (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id         UUID REFERENCES ai_tasks(id) ON DELETE CASCADE,
    recommendation_type VARCHAR(100) NOT NULL,              -- vendor, material, method, schedule, risk_mitigation
    title           VARCHAR(500) NOT NULL,
    description     TEXT,
    reasoning       TEXT,                                   -- почему эта рекомендация
    expected_impact TEXT,                                   -- ожидаемый эффект
    confidence      NUMERIC(5,2),
    status          VARCHAR(50) NOT NULL DEFAULT 'pending', -- pending, accepted, rejected, implemented
    accepted_by     UUID REFERENCES employees(id),
    accepted_at     TIMESTAMPTZ,
    implemented_at  TIMESTAMPTZ,
    actual_impact   TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ============================================================================
-- 7. Чат-история (Ruslan OS conversations)
-- ============================================================================
CREATE TABLE ai_conversations (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id      VARCHAR(100) NOT NULL,
    user_message    TEXT NOT NULL,
    assistant_message TEXT NOT NULL,
    intent          VARCHAR(200),
    entities        JSONB,
    module_used     VARCHAR(100),
    action_taken    VARCHAR(200),
    tokens_used     INTEGER,
    processing_time_ms INTEGER,
    feedback_score  INTEGER,                                -- 1-5 оценка пользователя
    feedback_text   TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ai_conv_session ON ai_conversations(session_id);
CREATE INDEX IF NOT EXISTS idx_ai_conv_created ON ai_conversations(created_at);

-- ============================================================================
-- 8. AI-метрики
-- ============================================================================
CREATE TABLE ai_metrics (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    report_date     DATE NOT NULL DEFAULT CURRENT_DATE,
    total_requests  INTEGER DEFAULT 0,
    total_tokens    INTEGER DEFAULT 0,
    total_cost      NUMERIC(12,4) DEFAULT 0,
    avg_confidence  NUMERIC(5,2),
    avg_processing_time_ms INTEGER,
    success_rate    NUMERIC(5,2),                           -- % успешных
    by_type         JSONB,                                  -- статистика по типам задач
    by_module       JSONB,                                  -- статистика по модулям
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(report_date)
);

-- ============================================================================
-- Комментарии
-- ============================================================================
COMMENT ON TABLE ai_agents IS 'AI-агенты (классификаторы, экстракторы, предикторы)';
COMMENT ON TABLE ai_tasks IS 'AI-задачи (все запросы к моделям)';
COMMENT ON TABLE ai_classifications IS 'Классификация запросов (Ruslan OS dispatcher)';
COMMENT ON TABLE ai_extractions IS 'Извлечение данных (OCR, парсинг документов)';
COMMENT ON TABLE ai_predictions IS 'Прогнозы (стоимость, риски, сроки)';
COMMENT ON TABLE ai_recommendations IS 'Рекомендации (поставщики, материалы, методы)';
COMMENT ON TABLE ai_conversations IS 'История чатов (Ruslan OS)';
COMMENT ON TABLE ai_metrics IS 'Метрики использования AI';
