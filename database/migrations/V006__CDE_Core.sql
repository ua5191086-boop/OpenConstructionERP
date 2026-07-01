-- OpenConstructionERP
-- V006: CDE Core lite (D-01) — document register, numbering rules, revisions,
--       ISO 19650 status model, transmittals.
-- Owner: core-py lane. Registered in docs/WORKSTREAMS.md.

CREATE TABLE document_numbering_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    doc_type VARCHAR(10) NOT NULL,          -- DWG, SPC, RPT, MS, ITP, COR, CAL, MOM
    prefix VARCHAR(60) NOT NULL,            -- e.g. 'TTZ-CAI-DWG-'
    pad INTEGER NOT NULL DEFAULT 4,
    next_seq INTEGER NOT NULL DEFAULT 1,
    UNIQUE (project_id, doc_type)
);

CREATE TABLE documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    doc_number VARCHAR(80) NOT NULL,
    title VARCHAR(255) NOT NULL,
    doc_type VARCHAR(10) NOT NULL,
    discipline VARCHAR(30),
    originator VARCHAR(120),
    revision VARCHAR(10) NOT NULL DEFAULT 'P01',    -- P01.. preliminary, C01.. contractual
    -- ISO 19650 container state
    state VARCHAR(12) NOT NULL DEFAULT 'WIP'
        CHECK (state IN ('WIP','Shared','Published','Archived')),
    suitability VARCHAR(4),                          -- S1..S7, A, B (issue purpose code)
    file_key VARCHAR(255),                           -- MinIO object key (upload wired later)
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (project_id, doc_number)
);

CREATE INDEX idx_docs_project ON documents(project_id, doc_type, state);
CREATE INDEX idx_docs_fts ON documents
    USING GIN (to_tsvector('simple', doc_number || ' ' || title));

CREATE TABLE document_revisions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    revision VARCHAR(10) NOT NULL,
    state VARCHAR(12) NOT NULL,
    suitability VARCHAR(4),
    file_key VARCHAR(255),
    issued_by VARCHAR(120),
    issued_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    notes TEXT,
    UNIQUE (document_id, revision)
);

CREATE TABLE transmittals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    number INTEGER NOT NULL,
    code VARCHAR(20) NOT NULL,               -- 'TRN-0001'
    to_party VARCHAR(160) NOT NULL,          -- организация-получатель
    purpose VARCHAR(30) NOT NULL DEFAULT 'for_information'
        CHECK (purpose IN ('for_information','for_review','for_approval',
                           'for_construction','as_built')),
    cover_note TEXT,
    issued_by VARCHAR(120),
    issued_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, number),
    UNIQUE (project_id, code)
);

CREATE TABLE transmittal_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transmittal_id UUID NOT NULL REFERENCES transmittals(id) ON DELETE CASCADE,
    document_id UUID NOT NULL REFERENCES documents(id),
    revision VARCHAR(10) NOT NULL,            -- snapshot of the revision sent
    UNIQUE (transmittal_id, document_id)
);

INSERT INTO object_types (code, name, icon, module_owner) VALUES
('cde_document', 'Document',    'file-text', 'D-01'),
('transmittal',  'Transmittal', 'send',      'D-01')
ON CONFLICT (code) DO NOTHING;
