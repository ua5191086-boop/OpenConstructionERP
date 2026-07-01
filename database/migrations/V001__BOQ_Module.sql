-- OpenConstructionERP
-- BOQ Module — Bill of Quantities for Linear Infrastructure Projects
-- Version: MVP-1

-- ============================================================
-- 1. CBS (Cost Breakdown Structure) — Главы сметы
-- ============================================================
CREATE TABLE cbs_chapters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    -- NULL project_id = global catalog chapter (template shared across projects)
    project_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    code VARCHAR(10) NOT NULL,
    name VARCHAR(255) NOT NULL,
    name_ru VARCHAR(255),
    parent_id UUID REFERENCES cbs_chapters(id),
    level INTEGER NOT NULL DEFAULT 1,
    sort_order INTEGER DEFAULT 0,
    path LTREE,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_cbs_project ON cbs_chapters(project_id);
CREATE INDEX idx_cbs_path ON cbs_chapters USING GIST(path);
CREATE UNIQUE INDEX uq_cbs_global_code ON cbs_chapters(code) WHERE project_id IS NULL;

-- ============================================================
-- 2. BOQ Hierarchy
-- ============================================================
CREATE TABLE boq_sections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    code VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    section_type VARCHAR(20) NOT NULL CHECK (section_type IN ('Track', 'Station', 'Depot', 'Yard')),
    start_km DECIMAL(10,3),
    end_km DECIMAL(10,3),
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(project_id, code)
);

CREATE INDEX idx_boq_sections_project ON boq_sections(project_id);

CREATE TABLE boq_complexes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    section_id UUID NOT NULL REFERENCES boq_sections(id) ON DELETE CASCADE,
    code VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(project_id, code)
);

CREATE INDEX idx_boq_complexes_section ON boq_complexes(section_id);

CREATE TABLE boq_objects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    complex_id UUID NOT NULL REFERENCES boq_complexes(id) ON DELETE CASCADE,
    code VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(project_id, code)
);

CREATE INDEX idx_boq_objects_complex ON boq_objects(complex_id);

CREATE TABLE boq_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    object_id UUID NOT NULL REFERENCES boq_objects(id) ON DELETE CASCADE,
    cbs_chapter_id UUID NOT NULL REFERENCES cbs_chapters(id),
    code VARCHAR(100) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    unit VARCHAR(20) NOT NULL,
    quantity DECIMAL(20,4) NOT NULL DEFAULT 0,
    unit_price DECIMAL(20,2) NOT NULL DEFAULT 0,
    total_cost DECIMAL(20,2) GENERATED ALWAYS AS (quantity * unit_price) STORED,
    currency VARCHAR(3) DEFAULT 'USD',
    contractor_id UUID REFERENCES organizations(id),
    contract_id UUID REFERENCES contracts(id),
    funding_source VARCHAR(50),
    phase VARCHAR(50),
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'completed', 'cancelled', 'pending')),
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_boq_items_object ON boq_items(object_id);
CREATE INDEX idx_boq_items_cbs ON boq_items(cbs_chapter_id);
CREATE INDEX idx_boq_items_contractor ON boq_items(contractor_id);
CREATE INDEX idx_boq_items_contract ON boq_items(contract_id);
CREATE INDEX idx_boq_items_status ON boq_items(project_id, status);
CREATE INDEX idx_boq_items_fts ON boq_items USING GIN(to_tsvector('english', name || ' ' || COALESCE(description, '')));

-- ============================================================
-- 3. Cost Transactions (Plan / Actual / Forecast / Variance)
-- ============================================================
CREATE TABLE cost_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    boq_item_id UUID NOT NULL REFERENCES boq_items(id) ON DELETE CASCADE,
    cbs_chapter_id UUID NOT NULL REFERENCES cbs_chapters(id),
    contractor_id UUID REFERENCES organizations(id),
    contract_id UUID REFERENCES contracts(id),
    transaction_type VARCHAR(20) NOT NULL CHECK (transaction_type IN ('Plan', 'Actual', 'Forecast', 'Variance', 'Commitment')),
    amount DECIMAL(20,2) NOT NULL DEFAULT 0,
    currency VARCHAR(3) DEFAULT 'USD',
    exchange_rate DECIMAL(10,6) DEFAULT 1,
    amount_base_currency DECIMAL(20,2) GENERATED ALWAYS AS (amount * COALESCE(exchange_rate, 1)) STORED,
    period DATE NOT NULL,
    funding_source VARCHAR(50),
    description TEXT,
    reference_type VARCHAR(50),
    reference_id UUID,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_cost_tx_project ON cost_transactions(project_id, period);
CREATE INDEX idx_cost_tx_boq ON cost_transactions(boq_item_id);
CREATE INDEX idx_cost_tx_cbs ON cost_transactions(cbs_chapter_id);
CREATE INDEX idx_cost_tx_type ON cost_transactions(project_id, transaction_type);
CREATE INDEX idx_cost_tx_contractor ON cost_transactions(contractor_id);
CREATE INDEX idx_cost_tx_contract ON cost_transactions(contract_id);
CREATE INDEX idx_cost_tx_period ON cost_transactions(project_id, period DESC);

-- ============================================================
-- 4. Budget Versions
-- ============================================================
CREATE TABLE budget_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    version_number INTEGER NOT NULL,
    version_name VARCHAR(100),
    status VARCHAR(20) DEFAULT 'draft' CHECK (status IN ('draft', 'approved', 'superseded')),
    total_amount DECIMAL(20,2),
    approved_by UUID REFERENCES users(id),
    approved_at TIMESTAMPTZ,
    notes TEXT,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_budget_versions_project ON budget_versions(project_id);

-- ============================================================
-- 5. Seed Data: CBS Chapters for Railway Projects
-- ============================================================
INSERT INTO cbs_chapters (code, name, name_ru, level) VALUES
('01', 'Site Preparation', 'Подготовка площадки', 1),
('01.01', 'Land Acquisition', 'Приобретение земли', 2),
('01.02', 'Utility Relocation', 'Перемещение коммуникаций', 2),
('01.03', 'Demolition', 'Снос существующих объектов', 2),
('01.04', 'Preliminary & Auxiliary Works', 'Предварительные и вспомогательные работы', 2),

('02', 'Earthworks', 'Земляные работы', 1),
('02.01', 'Excavation', 'Экскавация', 2),
('02.02', 'Embankment', 'Насыпь', 2),
('02.03', 'Drainage', 'Дренаж', 2),
('02.04', 'Slope Stabilization', 'Стабилизация склонов', 2),
('02.05', 'Geotechnical Structures', 'Геотехнические сооружения', 2),

('03', 'Civil Structures', 'Гражданские сооружения', 1),
('03.01', 'Bridges', 'Мосты', 2),
('03.02', 'Elevated Structures', 'Эстакады', 2),
('03.03', 'Viaducts', 'Виадуки', 2),
('03.04', 'Tunnels', 'Туннели', 2),
('03.05', 'Culverts', 'Водопропускные трубы', 2),
('03.06', 'Retaining Walls', 'Стены удержания', 2),

('04', 'Track Works', 'Путевые работы', 1),
('04.01', 'Ballast', 'Щебень', 2),
('04.02', 'Track Structure (Rails & Sleepers)', 'Конструкция пути (рельсы и шпалы)', 2),
('04.03', 'Turnouts & Crossovers', 'Стрелочные переводы и разъединители', 2),
('04.04', 'Continuous Welded Rail', 'Непрерывный сварной рельс', 2),
('04.05', 'Level Crossings', 'Переходы через железнодорожные пути', 2),

('05', 'Signalling, Control & Communication', 'Сигнализация, управление и связь', 1),
('05.01', 'Signalling Systems', 'Сигнальные системы', 2),
('05.02', 'Interlocking', 'Интерлоккинг', 2),
('05.03', 'Block Systems', 'Блочные системы', 2),
('05.04', 'Telecommunications', 'Телекоммуникации', 2),
('05.05', 'Fibre Optic Lines', 'Волоконно-оптические линии связи', 2),
('05.06', 'Radio Communication', 'Радиосвязь', 2),
('05.07', 'Data Transmission Systems', 'Системы передачи данных', 2),
('05.08', 'Control Centres', 'Центры управления интерлоккингом', 2),
('05.09', 'Equipment Rooms', 'Коммутационные помещения', 2),
('05.10', 'Dispatch Centres', 'Диспетчерские центры', 2),

('06', 'Buildings & Structures', 'Здания и сооружения', 1),
('06.01', 'Railway Stations', 'Железнодорожные станции', 2),
('06.02', 'Station Buildings', 'Станционные здания', 2),
('06.03', 'Locomotive Depots', 'Депо локомотивов', 2),
('06.04', 'Rolling Stock Depots', 'Депо подвижного состава', 2),
('06.05', 'Service Facilities', 'Сервисные объекты', 2),
('06.06', 'Administrative Buildings', 'Административные и бытовые здания', 2),
('06.07', 'Warehouse Complexes', 'Складские комплексы', 2),
('06.08', 'Industrial Buildings', 'Промышленные здания', 2),

('07', 'Power Supply System', 'Система электроснабжения', 1),
('07.01', 'Traction Substations', 'Тяговые подстанции', 2),
('07.02', 'Transformer Substations', 'Трансформаторные подстанции', 2),
('07.03', 'Overhead Contact Lines', 'Воздушные контактные линии', 2),
('07.04', 'External Power Supply', 'Внешнее электроснабжение', 2),
('07.05', 'Cable Networks', 'Кабельные сети', 2),
('07.06', 'Lighting Systems', 'Системы освещения', 2),
('07.07', 'Power Distribution & Management', 'Система распределения и управления электроэнергией', 2),
('07.08', 'Distribution Substations', 'Распределительные подстанции', 2),
('07.09', 'Energy Dispatch Facilities', 'Объекты диспетчеризации энергоресурсов', 2),

('08', 'Utilities', 'Инженерные сети', 1),
('08.01', 'Water Supply', 'Водоснабжение', 2),
('08.02', 'Sewerage', 'Канализация', 2),
('08.03', 'Heating', 'Отопление', 2),
('08.04', 'Gas Supply', 'Газоснабжение', 2),
('08.05', 'Drainage', 'Дренаж', 2),
('08.06', 'Stormwater', 'Ливневая канализация', 2),
('08.07', 'External Utilities', 'Внешние инженерные сети', 2),

('09', 'Rolling Stock, Equipment, Furniture & Inventory', 'Подвижной состав, оборудование, мебель и инвентарь', 1),
('09.01', 'Locomotives', 'Локомотивы', 2),
('09.02', 'Electric Multiple Units (EMU)', 'Электрические многоцелевые поезда (ЭМУ)', 2),
('09.03', 'Passenger Cars', 'Пассажирские вагоны', 2),
('09.04', 'Freight Cars', 'Грузовые вагоны', 2),
('09.05', 'Process & Operational Equipment', 'Процессное и операционное оборудование', 2),
('09.06', 'Furniture', 'Мебель', 2),
('09.07', 'Inventory', 'Инвентарь', 2),
('09.08', 'Operational Equipment', 'Эксплуатационное оборудование', 2),

('10', 'Temporary Buildings & Facilities', 'Временные здания и сооружения', 1),
('10.01', 'Construction Camps', 'Строительные лагеря', 2),
('10.02', 'Temporary Warehouses', 'Временные склады', 2),
('10.03', 'Temporary Utilities', 'Временные инженерные сети', 2),
('10.04', 'Temporary Roads', 'Временные дороги', 2),
('10.05', 'Temporary Power Supply', 'Временное электроснабжение', 2),

('11', 'Training & Employer Organization Costs', 'Расходы на обучение и организацию работодателя', 1),
('11.01', 'Personnel Training', 'Персональное обучение', 2),
('11.02', 'Staff Development', 'Развитие персонала', 2),
('11.03', 'Employer Organization Costs', 'Расходы на организацию работодателя', 2),
('11.04', 'Construction Supervision', 'Надзор за строительством', 2),
('11.05', 'Design Supervision', 'Надзор за проектированием', 2),
('11.06', 'Consultancy Services', 'Консультационные услуги', 2),

('12', 'Other Costs', 'Прочие расходы', 1),
('12.01', 'Design & Engineering Services', 'Проектные и инженерные услуги', 2),
('12.02', 'Expertise & Approvals', 'Экспертиза и согласования', 2),
('12.03', 'Studies & Surveys', 'Исследования и обследования', 2),
('12.04', 'Permits & Initial Documentation', 'Разрешения и исходная документация', 2),
('12.05', 'Insurance', 'Страхование', 2),
('12.06', 'Regulatory Approvals', 'Нормативные согласования', 2),
('12.07', 'Other Related Costs', 'Прочие сопутствующие расходы', 2);
