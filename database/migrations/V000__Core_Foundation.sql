-- OpenConstructionERP
-- V000: Core foundation — extensions and the projects table that V001 depends on.
-- FIX: V001__BOQ_Module.sql references projects(id) which was never created,
--      and uses LTREE / gen_random_uuid() without enabling extensions.

CREATE EXTENSION IF NOT EXISTS ltree;
CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    org_type VARCHAR(30) NOT NULL DEFAULT 'contractor'
        CHECK (org_type IN ('holding','contractor','client','consultant','subcontractor','supplier','bank')),
    country CHAR(2),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    name_ru VARCHAR(255),
    organization_id UUID REFERENCES organizations(id),
    client_id UUID REFERENCES organizations(id),
    project_type VARCHAR(30) NOT NULL DEFAULT 'metro'
        CHECK (project_type IN ('metro','tunnel','railway','hydro','microtunnel','road','other')),
    status VARCHAR(20) NOT NULL DEFAULT 'tender'
        CHECK (status IN ('lead','tender','mobilization','execution','commissioning','dlp','closed','cancelled')),
    country CHAR(2),
    currency CHAR(3) NOT NULL DEFAULT 'USD',
    start_date DATE,
    finish_date DATE,
    contract_value NUMERIC(18,2),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    created_by UUID,
    version INTEGER NOT NULL DEFAULT 1
);

CREATE INDEX IF NOT EXISTS idx_projects_org ON projects(organization_id);
CREATE INDEX IF NOT EXISTS idx_projects_status ON projects(status);

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    -- Identity lives in Keycloak (ADR-009); this table mirrors the principal for FK integrity.
    keycloak_id UUID UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    full_name VARCHAR(255) NOT NULL,
    organization_id UUID REFERENCES organizations(id),
    principal_type VARCHAR(10) NOT NULL DEFAULT 'human' CHECK (principal_type IN ('human','agent','service')),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS contracts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    code VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    contract_type VARCHAR(30) NOT NULL DEFAULT 'main'
        CHECK (contract_type IN ('main','subcontract','supply','services','framework')),
    contract_form VARCHAR(30)
        CHECK (contract_form IN ('fidic_red','fidic_yellow','fidic_silver','epc','epcm','bespoke')),
    party_id UUID REFERENCES organizations(id),
    counterparty_id UUID REFERENCES organizations(id),
    currency CHAR(3) NOT NULL DEFAULT 'USD',
    contract_value NUMERIC(18,2),
    signed_date DATE,
    status VARCHAR(20) NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft','negotiation','signed','active','completed','terminated')),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    version INTEGER NOT NULL DEFAULT 1,
    UNIQUE(project_id, code)
);

CREATE INDEX IF NOT EXISTS idx_contracts_project ON contracts(project_id);
CREATE INDEX IF NOT EXISTS idx_users_org ON users(organization_id);
