-- ============================================================================
-- V006__Procurement_Module.sql
-- Модуль управления закупками (Procurement Management)
-- Часть OpenConstructionERP — Project Operating System
-- ============================================================================

-- ============================================================================
-- 1. Заявки на закупку
-- ============================================================================
CREATE TABLE procurement_requests (
    id              BIGSERIAL PRIMARY KEY,
    request_number  VARCHAR(50) NOT NULL UNIQUE,          -- PR-2026-001
    project_id      BIGINT REFERENCES projects(id),
    section_id      BIGINT REFERENCES sections(id),
    requested_by    BIGINT REFERENCES employees(id),
    request_date    DATE NOT NULL,
    required_date   DATE,                                  -- срок поставки
    priority        VARCHAR(20) NOT NULL DEFAULT 'normal', -- low, normal, high, urgent
    status          VARCHAR(50) NOT NULL DEFAULT 'draft',  -- draft, submitted, approved, rejected, ordered, partially_received, received, closed
    description     TEXT,
    justification   TEXT,
    estimated_cost  NUMERIC(18,2),
    currency        VARCHAR(3) NOT NULL DEFAULT 'USD',
    budget_item_id  BIGINT REFERENCES budget_items(id),    -- статья бюджета
    approved_by     BIGINT REFERENCES employees(id),
    approved_at     TIMESTAMPTZ,
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_pr_project ON procurement_requests(project_id);
CREATE INDEX idx_pr_status ON procurement_requests(status);
CREATE INDEX idx_pr_date ON procurement_requests(request_date);
CREATE INDEX idx_pr_priority ON procurement_requests(priority);

-- ============================================================================
-- 2. Позиции заявки
-- ============================================================================
CREATE TABLE procurement_request_items (
    id              BIGSERIAL PRIMARY KEY,
    request_id      BIGINT NOT NULL REFERENCES procurement_requests(id) ON DELETE CASCADE,
    line_number     INTEGER NOT NULL,
    item_code       VARCHAR(100),
    description     TEXT NOT NULL,
    specification   TEXT,
    unit            VARCHAR(20) NOT NULL,
    quantity        NUMERIC(18,4) NOT NULL,
    estimated_unit_price NUMERIC(18,2),
    estimated_total NUMERIC(18,2),
    currency        VARCHAR(3) NOT NULL DEFAULT 'USD',
    boq_item_id     BIGINT REFERENCES boq_items(id),       -- привязка к BOQ
    material_code   VARCHAR(100),                           -- код материала/оборудования
    catalog_number  VARCHAR(200),                           -- номер по каталогу
    preferred_vendor BIGINT REFERENCES contractors(id),
    sort_order      INTEGER DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(request_id, line_number)
);

CREATE INDEX idx_pri_request ON procurement_request_items(request_id);

-- ============================================================================
-- 3. Заказы на поставку (PO)
-- ============================================================================
CREATE TABLE purchase_orders (
    id              BIGSERIAL PRIMARY KEY,
    po_number       VARCHAR(50) NOT NULL UNIQUE,          -- PO-2026-001
    request_id      BIGINT REFERENCES procurement_requests(id),
    project_id      BIGINT REFERENCES projects(id),
    vendor_id       BIGINT NOT NULL REFERENCES contractors(id),
    order_date      DATE NOT NULL,
    delivery_date   DATE,
    delivery_address TEXT,
    payment_terms   VARCHAR(200),
    shipping_terms  VARCHAR(200),                          -- INCOTERMS
    subtotal        NUMERIC(18,2),
    tax_amount      NUMERIC(18,2) DEFAULT 0,
    tax_rate        NUMERIC(5,2) DEFAULT 0,
    shipping_cost   NUMERIC(18,2) DEFAULT 0,
    total_amount    NUMERIC(18,2) NOT NULL,
    currency        VARCHAR(3) NOT NULL DEFAULT 'USD',
    status          VARCHAR(50) NOT NULL DEFAULT 'draft',  -- draft, sent, confirmed, shipped, partially_received, received, cancelled
    approved_by     BIGINT REFERENCES employees(id),
    approved_at     TIMESTAMPTZ,
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_po_vendor ON purchase_orders(vendor_id);
CREATE INDEX idx_po_project ON purchase_orders(project_id);
CREATE INDEX idx_po_status ON purchase_orders(status);
CREATE INDEX idx_po_date ON purchase_orders(order_date);

-- ============================================================================
-- 4. Позиции заказа
-- ============================================================================
CREATE TABLE purchase_order_items (
    id              BIGSERIAL PRIMARY KEY,
    po_id           BIGINT NOT NULL REFERENCES purchase_orders(id) ON DELETE CASCADE,
    line_number     INTEGER NOT NULL,
    request_item_id BIGINT REFERENCES procurement_request_items(id),
    item_code       VARCHAR(100),
    description     TEXT NOT NULL,
    specification   TEXT,
    unit            VARCHAR(20) NOT NULL,
    quantity_ordered NUMERIC(18,4) NOT NULL,
    quantity_received NUMERIC(18,4) DEFAULT 0,
    unit_price      NUMERIC(18,2) NOT NULL,
    total_price     NUMERIC(18,2) NOT NULL,
    currency        VARCHAR(3) NOT NULL DEFAULT 'USD',
    delivery_date   DATE,
    sort_order      INTEGER DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(po_id, line_number)
);

CREATE INDEX idx_poi_po ON purchase_order_items(po_id);

-- ============================================================================
-- 5. Поставки / приход
-- ============================================================================
CREATE TABLE goods_receipts (
    id              BIGSERIAL PRIMARY KEY,
    receipt_number  VARCHAR(50) NOT NULL UNIQUE,          -- GR-2026-001
    po_id           BIGINT NOT NULL REFERENCES purchase_orders(id),
    receipt_date    DATE NOT NULL,
    received_by     BIGINT REFERENCES employees(id),
    status          VARCHAR(50) NOT NULL DEFAULT 'pending', -- pending, inspected, accepted, rejected, partially_accepted
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_gr_po ON goods_receipts(po_id);
CREATE INDEX idx_gr_date ON goods_receipts(receipt_date);

-- ============================================================================
-- 6. Позиции поставки
-- ============================================================================
CREATE TABLE goods_receipt_items (
    id              BIGSERIAL PRIMARY KEY,
    receipt_id      BIGINT NOT NULL REFERENCES goods_receipts(id) ON DELETE CASCADE,
    po_item_id      BIGINT REFERENCES purchase_order_items(id),
    item_code       VARCHAR(100),
    description     TEXT,
    unit            VARCHAR(20),
    quantity_ordered NUMERIC(18,4),
    quantity_received NUMERIC(18,4) NOT NULL,
    quantity_accepted NUMERIC(18,4),
    quantity_rejected NUMERIC(18,4) DEFAULT 0,
    rejection_reason TEXT,
    unit_price      NUMERIC(18,2),
    total_price     NUMERIC(18,2),
    batch_number    VARCHAR(100),
    serial_number   VARCHAR(100),
    expiry_date     DATE,
    storage_location VARCHAR(200),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_gri_receipt ON goods_receipt_items(receipt_id);

-- ============================================================================
-- 7. Склад / инвентаризация
-- ============================================================================
CREATE TABLE inventory_items (
    id              BIGSERIAL PRIMARY KEY,
    item_code       VARCHAR(100) NOT NULL UNIQUE,
    name            VARCHAR(500) NOT NULL,
    description     TEXT,
    category        VARCHAR(200),
    unit            VARCHAR(20) NOT NULL,
    unit_price      NUMERIC(18,2),
    currency        VARCHAR(3) NOT NULL DEFAULT 'USD',
    min_quantity    NUMERIC(18,4) DEFAULT 0,               -- минимальный запас
    max_quantity    NUMERIC(18,4),                          -- максимальный запас
    current_quantity NUMERIC(18,4) DEFAULT 0,
    reserved_quantity NUMERIC(18,4) DEFAULT 0,              -- зарезервировано
    available_quantity NUMERIC(18,4) DEFAULT 0,
    storage_location VARCHAR(200),
    warehouse       VARCHAR(200),
    material_type   VARCHAR(50),                            -- raw_material, consumable, spare_part, equipment, tool
    is_active       BOOLEAN DEFAULT TRUE,
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_inv_category ON inventory_items(category);
CREATE INDEX idx_inv_location ON inventory_items(storage_location);

-- ============================================================================
-- 8. Движение склада
-- ============================================================================
CREATE TABLE inventory_movements (
    id              BIGSERIAL PRIMARY KEY,
    item_id         BIGINT NOT NULL REFERENCES inventory_items(id) ON DELETE CASCADE,
    movement_date   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    movement_type   VARCHAR(50) NOT NULL,                   -- receipt, issue, transfer, adjustment, return
    quantity        NUMERIC(18,4) NOT NULL,
    unit_price      NUMERIC(18,2),
    total_price     NUMERIC(18,2),
    reference_type  VARCHAR(50),                           -- goods_receipt, purchase_order, work_order, adjustment
    reference_id    BIGINT,
    from_location   VARCHAR(200),
    to_location     VARCHAR(200),
    performed_by    BIGINT REFERENCES employees(id),
    notes           TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_inv_mov_item ON inventory_movements(item_id);
CREATE INDEX idx_inv_mov_date ON inventory_movements(movement_date);
CREATE INDEX idx_inv_mov_type ON inventory_movements(movement_type);

-- ============================================================================
-- 9. Вендоры / поставщики (расширение contractors)
-- ============================================================================
CREATE TABLE vendor_evaluations (
    id              BIGSERIAL PRIMARY KEY,
    vendor_id       BIGINT NOT NULL REFERENCES contractors(id) ON DELETE CASCADE,
    evaluation_date DATE NOT NULL,
    evaluator       BIGINT REFERENCES employees(id),
    criteria_scores TEXT,                                   -- JSON: quality, delivery, price, service
    overall_score   NUMERIC(3,1),                          -- 1.0 - 5.0
    comments        TEXT,
    is_approved     BOOLEAN DEFAULT FALSE,
    valid_until     DATE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_vendor_eval_vendor ON vendor_evaluations(vendor_id);

-- ============================================================================
-- 10. Триггеры
-- ============================================================================
CREATE OR REPLACE FUNCTION update_procurement_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_pr_updated
    BEFORE UPDATE ON procurement_requests
    FOR EACH ROW
    EXECUTE FUNCTION update_procurement_timestamp();

CREATE TRIGGER trg_po_updated
    BEFORE UPDATE ON purchase_orders
    FOR EACH ROW
    EXECUTE FUNCTION update_procurement_timestamp();

-- ============================================================================
-- Комментарии
-- ============================================================================
COMMENT ON TABLE procurement_requests IS 'Заявки на закупку';
COMMENT ON TABLE procurement_request_items IS 'Позиции заявки';
COMMENT ON TABLE purchase_orders IS 'Заказы на поставку (PO)';
COMMENT ON TABLE purchase_order_items IS 'Позиции заказа';
COMMENT ON TABLE goods_receipts IS 'Поставки / приход';
COMMENT ON TABLE goods_receipt_items IS 'Позиции поставки';
COMMENT ON TABLE inventory_items IS 'Складские позиции';
COMMENT ON TABLE inventory_movements IS 'Движение склада';
COMMENT ON TABLE vendor_evaluations IS 'Оценка поставщиков';
