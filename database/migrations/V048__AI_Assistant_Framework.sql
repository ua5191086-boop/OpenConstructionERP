-- ============================================================================
-- V048__AI_Assistant_Framework.sql
-- AI Assistant — knowledge base, prompts, sessions, embeddings
-- Часть OpenConstructionERP — Project Operating System
-- ============================================================================

-- Enable pgvector if available
CREATE EXTENSION IF NOT EXISTS vector;

-- ============================================================================
-- 1. AI-провайдеры
-- ============================================================================
CREATE TABLE ai_providers (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name                VARCHAR(200) NOT NULL,
    provider_type       VARCHAR(50) NOT NULL,                  -- openai, anthropic, local, custom
    model               VARCHAR(200) NOT NULL,
    base_url            VARCHAR(500),
    api_key_encrypted   TEXT,                                   -- зашифрованный ключ
    max_tokens          INTEGER DEFAULT 4096,
    temperature         NUMERIC(3,2) DEFAULT 0.7,
    embedding_model     VARCHAR(200),                           -- модель эмбеддингов
    embedding_dimensions INTEGER DEFAULT 1536,
    cost_per_1k_input   NUMERIC(10,6),                         -- цена за 1K токенов input
    cost_per_1k_output  NUMERIC(10,6),                         -- цена за 1K токенов output
    is_active           BOOLEAN DEFAULT TRUE,
    is_default          BOOLEAN DEFAULT FALSE,
    notes               TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ai_providers_active ON ai_providers(is_active) WHERE is_active = TRUE;

COMMENT ON TABLE ai_providers IS 'Провайдеры AI-моделей (LLM, embedding)';

-- ============================================================================
-- 2. Knowledge Base — документы для RAG
-- ============================================================================
CREATE TABLE ai_knowledge_base (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id          UUID REFERENCES projects(id) ON DELETE CASCADE,
    kb_type             VARCHAR(50) NOT NULL,                  -- document, specification, regulation, manual, contract, report
    title               VARCHAR(500) NOT NULL,
    source              VARCHAR(200),                           -- upload, integration, manual
    source_url          VARCHAR(500),
    original_filename   VARCHAR(500),
    content_type        VARCHAR(100),                           -- pdf, docx, txt, md, html, csv
    content_text        TEXT,                                   -- извлечённый текст
    chunk_count         INTEGER DEFAULT 0,
    total_tokens        INTEGER DEFAULT 0,
    embedding_status    VARCHAR(50) DEFAULT 'pending',         -- pending, processing, indexed, failed
    embedding_provider  UUID REFERENCES ai_providers(id),
    version             INTEGER DEFAULT 1,
    is_active           BOOLEAN DEFAULT TRUE,
    tags                JSONB DEFAULT '[]'::JSONB,
    metadata            JSONB,
    uploaded_by         VARCHAR(200),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ai_kb_project ON ai_knowledge_base(project_id);
CREATE INDEX idx_ai_kb_type ON ai_knowledge_base(kb_type);
CREATE INDEX idx_ai_kb_status ON ai_knowledge_base(embedding_status);
CREATE INDEX idx_ai_kb_tags ON ai_knowledge_base USING gin(tags);

COMMENT ON TABLE ai_knowledge_base IS 'База знаний для RAG — документы, спецификации, регламенты';

-- ============================================================================
-- 3. Чанки документов с эмбеддингами
-- ============================================================================
CREATE TABLE ai_embeddings (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    kb_id               UUID NOT NULL REFERENCES ai_knowledge_base(id) ON DELETE CASCADE,
    chunk_index         INTEGER NOT NULL,                       -- порядковый номер чанка
    chunk_text          TEXT NOT NULL,
    chunk_tokens        INTEGER,
    embedding           JSONB,                                  -- массив float: [0.001, 0.002, ...] (или vector если pgvector)
    embedding_model     VARCHAR(200),
    metadata            JSONB,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(kb_id, chunk_index)
);

CREATE INDEX idx_ai_embeddings_kb ON ai_embeddings(kb_id);

COMMENT ON TABLE ai_embeddings IS 'Чанки документов с векторными эмбеддингами для семантического поиска';

-- ============================================================================
-- 4. AI-сессии (чат с AI-ассистентом)
-- ============================================================================
CREATE TABLE ai_sessions (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id          UUID REFERENCES projects(id) ON DELETE CASCADE,
    session_type        VARCHAR(50) NOT NULL DEFAULT 'chat',    -- chat, analysis, report_generation, code_generation, data_query
    title               VARCHAR(500),
    provider_id         UUID REFERENCES ai_providers(id),
    system_prompt       TEXT,
    temperature         NUMERIC(3,2) DEFAULT 0.7,
    max_tokens          INTEGER DEFAULT 4096,
    context_kb_ids      JSONB DEFAULT '[]'::JSONB,             -- привязанные документы из KB
    message_count       INTEGER DEFAULT 0,
    total_tokens_input  INTEGER DEFAULT 0,
    total_tokens_output INTEGER DEFAULT 0,
    total_cost          NUMERIC(12,6) DEFAULT 0,
    is_pinned           BOOLEAN DEFAULT FALSE,
    user_id             VARCHAR(200),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ai_sessions_project ON ai_sessions(project_id);
CREATE INDEX idx_ai_sessions_type ON ai_sessions(session_type);
CREATE INDEX idx_ai_sessions_pinned ON ai_sessions(is_pinned) WHERE is_pinned = TRUE;

COMMENT ON TABLE ai_sessions IS 'Сессии AI-ассистента';

-- ============================================================================
-- 5. Сообщения AI-сессии
-- ============================================================================
CREATE TABLE ai_messages (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id          UUID NOT NULL REFERENCES ai_sessions(id) ON DELETE CASCADE,
    role                VARCHAR(50) NOT NULL,                   -- user, assistant, system, tool
    content             TEXT NOT NULL,
    tool_calls          JSONB,
    tool_results        JSONB,
    tokens_input        INTEGER DEFAULT 0,
    tokens_output       INTEGER DEFAULT 0,
    cost                NUMERIC(12,6) DEFAULT 0,
    latency_ms          INTEGER,
    metadata            JSONB,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ai_messages_session ON ai_messages(session_id);
CREATE INDEX idx_ai_messages_role ON ai_messages(role);
CREATE INDEX idx_ai_messages_created ON ai_messages(created_at);

COMMENT ON TABLE ai_messages IS 'Сообщения внутри AI-сессии';

-- ============================================================================
-- 6. AI-функции (function calling / tools для AI)
-- ============================================================================
CREATE TABLE ai_functions (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name                VARCHAR(200) NOT NULL UNIQUE,
    description         TEXT NOT NULL,
    function_schema     JSONB NOT NULL,                         -- OpenAPI/JSON schema function definition
    implementation_type VARCHAR(50) NOT NULL DEFAULT 'sql',     -- sql, api, python, internal
    implementation      TEXT,                                    -- SQL запрос, URL API, или код
    parameters          JSONB DEFAULT '[]'::JSONB,              -- описание параметров
    return_type         VARCHAR(100),
    is_system           BOOLEAN DEFAULT FALSE,
    is_active           BOOLEAN DEFAULT TRUE,
    category            VARCHAR(100),                           -- project, finance, schedule, tunnel, contract
    required_permission VARCHAR(100),                            -- какое право нужно для вызова
    usage_count         INTEGER DEFAULT 0,
    notes               TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ai_functions_category ON ai_functions(category);
CREATE INDEX idx_ai_functions_active ON ai_functions(is_active) WHERE is_active = TRUE;

COMMENT ON TABLE ai_functions IS 'Реестр AI-функций для function calling (tools)';