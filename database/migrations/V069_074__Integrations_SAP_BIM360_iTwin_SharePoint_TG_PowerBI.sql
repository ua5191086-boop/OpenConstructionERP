-- ============================================================================
-- V069_074__Integrations_SAP_Autodesk_Bentley_SharePoint_Telegram_PowerBI.sql
-- ============================================================================
CREATE TABLE integration_sap_mappings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID REFERENCES projects(id),
    sap_entity_type VARCHAR(100) NOT NULL, -- project, order, material, vendor, invoice, cost_center
    sap_key VARCHAR(200) NOT NULL,
    oce_entity_type VARCHAR(100) NOT NULL,
    oce_entity_id UUID NOT NULL,
    last_sync_at TIMESTAMPTZ DEFAULT NOW(),
    sync_status VARCHAR(50) DEFAULT 'synced',
    sap_data JSONB,
    UNIQUE(sap_entity_type, sap_key)
);

CREATE TABLE integration_bim360_issues (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID REFERENCES projects(id),
    bim360_id VARCHAR(200) NOT NULL,
    issue_type VARCHAR(100),
    title VARCHAR(500),
    description TEXT,
    status VARCHAR(50),
    assigned_to VARCHAR(300),
    due_date DATE,
    linked_documents JSONB,
    last_sync_at TIMESTAMPTZ DEFAULT NOW(),
    oce_issue_id UUID,
    UNIQUE(bim360_id)
);

CREATE TABLE integration_itwin_models (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID REFERENCES projects(id),
    itwin_id VARCHAR(200) NOT NULL,
    model_name VARCHAR(500),
    model_type VARCHAR(100),
    version VARCHAR(50),
    last_sync_at TIMESTAMPTZ DEFAULT NOW(),
    sync_status VARCHAR(50) DEFAULT 'synced',
    UNIQUE(itwin_id)
);

CREATE TABLE integration_sharepoint_libraries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID REFERENCES projects(id),
    site_url VARCHAR(500) NOT NULL,
    library_name VARCHAR(300) NOT NULL,
    folder_path VARCHAR(500),
    sync_direction VARCHAR(20) DEFAULT 'bidirectional',
    last_sync_at TIMESTAMPTZ DEFAULT NOW(),
    sync_status VARCHAR(50) DEFAULT 'active'
);

CREATE TABLE integration_telegram_bot (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID REFERENCES projects(id),
    chat_id VARCHAR(100) NOT NULL,
    chat_title VARCHAR(500),
    bot_token_encrypted TEXT,
    notification_types JSONB DEFAULT '["alerts","daily_summary","approvals"]',
    is_active BOOLEAN DEFAULT TRUE,
    last_message_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE integration_powerbi_datasets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID REFERENCES projects(id),
    dataset_name VARCHAR(300) NOT NULL,
    tables JSONB,
    refresh_frequency VARCHAR(50) DEFAULT 'daily',
    last_refresh_at TIMESTAMPTZ,
    status VARCHAR(50) DEFAULT 'active',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_ism_project ON integration_sap_mappings(project_id);
CREATE INDEX idx_ib6_project ON integration_bim360_issues(project_id);
CREATE INDEX idx_iim_project ON integration_itwin_models(project_id);
CREATE INDEX idx_isl_project ON integration_sharepoint_libraries(project_id);
CREATE INDEX idx_itb_chat ON integration_telegram_bot(chat_id);