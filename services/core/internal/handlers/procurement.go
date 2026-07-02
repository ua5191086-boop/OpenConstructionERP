package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/openconstructionerp/oce/services/core/internal/models"
)

// ProcurementHandler handles Procurement module endpoints
type ProcurementHandler struct {
	db *sql.DB
}

func NewProcurementHandler(db *sql.DB) *ProcurementHandler {
	return &ProcurementHandler{db: db}
}

func (h *ProcurementHandler) RegisterRoutes(r chi.Router) {
	r.Route("/procurement", func(r chi.Router) {
		// Procurement Requests
		r.Get("/requests", h.ListRequests)
		r.Post("/requests", h.CreateRequest)
		r.Get("/requests/{id}", h.GetRequest)
		r.Put("/requests/{id}", h.UpdateRequest)
		r.Delete("/requests/{id}", h.DeleteRequest)
		r.Get("/requests/{id}/items", h.ListRequestItems)
		r.Post("/requests/{id}/items", h.CreateRequestItem)

		// Purchase Orders
		r.Get("/purchase-orders", h.ListPurchaseOrders)
		r.Post("/purchase-orders", h.CreatePurchaseOrder)
		r.Get("/purchase-orders/{id}", h.GetPurchaseOrder)
		r.Put("/purchase-orders/{id}", h.UpdatePurchaseOrder)
		r.Delete("/purchase-orders/{id}", h.DeletePurchaseOrder)
		r.Get("/purchase-orders/{id}/items", h.ListPurchaseOrderItems)
		r.Post("/purchase-orders/{id}/items", h.CreatePurchaseOrderItem)

		// Goods Receipts
		r.Get("/goods-receipts", h.ListGoodsReceipts)
		r.Post("/goods-receipts", h.CreateGoodsReceipt)
		r.Get("/goods-receipts/{id}", h.GetGoodsReceipt)
		r.Put("/goods-receipts/{id}", h.UpdateGoodsReceipt)
		r.Delete("/goods-receipts/{id}", h.DeleteGoodsReceipt)
		r.Get("/goods-receipts/{id}/items", h.ListGoodsReceiptItems)
		r.Post("/goods-receipts/{id}/items", h.CreateGoodsReceiptItem)

		// Inventory
		r.Get("/inventory", h.ListInventory)
		r.Post("/inventory", h.CreateInventoryItem)
		r.Get("/inventory/{id}", h.GetInventoryItem)
		r.Put("/inventory/{id}", h.UpdateInventoryItem)
		r.Delete("/inventory/{id}", h.DeleteInventoryItem)

		// Inventory Movements
		r.Get("/inventory-movements", h.ListInventoryMovements)
		r.Post("/inventory-movements", h.CreateInventoryMovement)

		// Stock Alerts
		r.Get("/stock-alerts", h.ListStockAlerts)

		// Inventory Summary
		r.Get("/inventory-summary", h.InventorySummary)

		// Vendor Evaluations
		r.Get("/vendor-evaluations", h.ListVendorEvaluations)
		r.Post("/vendor-evaluations", h.CreateVendorEvaluation)
		r.Get("/vendor-evaluations/{id}", h.GetVendorEvaluation)
	})
}

// --- Procurement Requests ---

func (h *ProcurementHandler) ListRequests(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	projectID := r.URL.Query().Get("project_id")

	query := `SELECT id, request_number, project_id, section_id, requested_by, request_date, required_date, priority, status, description, justification, estimated_cost, currency, budget_item_id, approved_by, approved_at, notes, created_at, updated_at FROM procurement_requests WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if status != "" {
		query += ` AND status = $` + itoa(argIdx)
		args = append(args, status)
		argIdx++
	}
	if projectID != "" {
		query += ` AND project_id = $` + itoa(argIdx)
		args = append(args, projectID)
		argIdx++
	}
	query += ` ORDER BY request_date DESC`

	rows, err := h.db.Query(query, args...)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	requests := make([]models.ProcurementRequest, 0)
	for rows.Next() {
		var pr models.ProcurementRequest
		if err := rows.Scan(&pr.ID, &pr.RequestNumber, &pr.ProjectID, &pr.SectionID, &pr.RequestedBy, &pr.RequestDate, &pr.RequiredDate, &pr.Priority, &pr.Status, &pr.Description, &pr.Justification, &pr.EstimatedCost, &pr.Currency, &pr.BudgetItemID, &pr.ApprovedBy, &pr.ApprovedAt, &pr.Notes, &pr.CreatedAt, &pr.UpdatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		requests = append(requests, pr)
	}
	respondJSON(w, http.StatusOK, requests)
}

func (h *ProcurementHandler) CreateRequest(w http.ResponseWriter, r *http.Request) {
	var input struct {
		RequestNumber string  `json:"request_number"`
		ProjectID     *string `json:"project_id"`
		SectionID     *string `json:"section_id"`
		RequestedBy   *string `json:"requested_by"`
		RequestDate   string  `json:"request_date"`
		RequiredDate  *string `json:"required_date"`
		Priority      string  `json:"priority"`
		Status        string  `json:"status"`
		Description   *string `json:"description"`
		Justification *string `json:"justification"`
		EstimatedCost *float64 `json:"estimated_cost"`
		Currency      string  `json:"currency"`
		BudgetItemID  *string `json:"budget_item_id"`
		Notes         *string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO procurement_requests (id, request_number, project_id, section_id, requested_by, request_date, required_date, priority, status, description, justification, estimated_cost, currency, budget_item_id, notes, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)`,
		id, input.RequestNumber, input.ProjectID, input.SectionID, input.RequestedBy, input.RequestDate, input.RequiredDate, input.Priority, input.Status, input.Description, input.Justification, input.EstimatedCost, input.Currency, input.BudgetItemID, input.Notes, now, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *ProcurementHandler) GetRequest(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var pr models.ProcurementRequest
	err := h.db.QueryRow(`SELECT id, request_number, project_id, section_id, requested_by, request_date, required_date, priority, status, description, justification, estimated_cost, currency, budget_item_id, approved_by, approved_at, notes, created_at, updated_at FROM procurement_requests WHERE id = $1`, id).
		Scan(&pr.ID, &pr.RequestNumber, &pr.ProjectID, &pr.SectionID, &pr.RequestedBy, &pr.RequestDate, &pr.RequiredDate, &pr.Priority, &pr.Status, &pr.Description, &pr.Justification, &pr.EstimatedCost, &pr.Currency, &pr.BudgetItemID, &pr.ApprovedBy, &pr.ApprovedAt, &pr.Notes, &pr.CreatedAt, &pr.UpdatedAt)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "procurement request not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, pr)
}

func (h *ProcurementHandler) UpdateRequest(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Status      *string  `json:"status"`
		Priority    *string  `json:"priority"`
		Description *string  `json:"description"`
		Notes       *string  `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	now := time.Now()
	_, err := h.db.Exec(`UPDATE procurement_requests SET status=COALESCE($1,status), priority=COALESCE($2,priority), description=COALESCE($3,description), notes=COALESCE($4,notes), updated_at=$5 WHERE id=$6`,
		input.Status, input.Priority, input.Description, input.Notes, now, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *ProcurementHandler) DeleteRequest(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM procurement_requests WHERE id = $1`, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// --- Purchase Orders ---

func (h *ProcurementHandler) ListPurchaseOrders(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	vendorID := r.URL.Query().Get("vendor_id")

	query := `SELECT id, po_number, request_id, project_id, vendor_id, order_date, delivery_date, delivery_address, payment_terms, shipping_terms, subtotal, tax_amount, tax_rate, shipping_cost, total_amount, currency, status, approved_by, approved_at, notes, created_at, updated_at FROM purchase_orders WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if status != "" {
		query += ` AND status = $` + itoa(argIdx)
		args = append(args, status)
		argIdx++
	}
	if vendorID != "" {
		query += ` AND vendor_id = $` + itoa(argIdx)
		args = append(args, vendorID)
		argIdx++
	}
	query += ` ORDER BY order_date DESC`

	rows, err := h.db.Query(query, args...)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	orders := make([]models.PurchaseOrder, 0)
	for rows.Next() {
		var po models.PurchaseOrder
		if err := rows.Scan(&po.ID, &po.PONumber, &po.RequestID, &po.ProjectID, &po.VendorID, &po.OrderDate, &po.DeliveryDate, &po.DeliveryAddress, &po.PaymentTerms, &po.ShippingTerms, &po.Subtotal, &po.TaxAmount, &po.TaxRate, &po.ShippingCost, &po.TotalAmount, &po.Currency, &po.Status, &po.ApprovedBy, &po.ApprovedAt, &po.Notes, &po.CreatedAt, &po.UpdatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		orders = append(orders, po)
	}
	respondJSON(w, http.StatusOK, orders)
}

func (h *ProcurementHandler) CreatePurchaseOrder(w http.ResponseWriter, r *http.Request) {
	var input struct {
		PONumber        string   `json:"po_number"`
		RequestID       *string  `json:"request_id"`
		ProjectID       *string  `json:"project_id"`
		VendorID        string   `json:"vendor_id"`
		OrderDate       string   `json:"order_date"`
		DeliveryDate    *string  `json:"delivery_date"`
		DeliveryAddress *string  `json:"delivery_address"`
		PaymentTerms    *string  `json:"payment_terms"`
		ShippingTerms   *string  `json:"shipping_terms"`
		Subtotal        *float64 `json:"subtotal"`
		TaxAmount       float64  `json:"tax_amount"`
		TaxRate         float64  `json:"tax_rate"`
		ShippingCost    float64  `json:"shipping_cost"`
		TotalAmount     float64  `json:"total_amount"`
		Currency        string   `json:"currency"`
		Status          string   `json:"status"`
		Notes           *string  `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO purchase_orders (id, po_number, request_id, project_id, vendor_id, order_date, delivery_date, delivery_address, payment_terms, shipping_terms, subtotal, tax_amount, tax_rate, shipping_cost, total_amount, currency, status, notes, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20)`,
		id, input.PONumber, input.RequestID, input.ProjectID, input.VendorID, input.OrderDate, input.DeliveryDate, input.DeliveryAddress, input.PaymentTerms, input.ShippingTerms, input.Subtotal, input.TaxAmount, input.TaxRate, input.ShippingCost, input.TotalAmount, input.Currency, input.Status, input.Notes, now, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *ProcurementHandler) GetPurchaseOrder(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var po models.PurchaseOrder
	err := h.db.QueryRow(`SELECT id, po_number, request_id, project_id, vendor_id, order_date, delivery_date, delivery_address, payment_terms, shipping_terms, subtotal, tax_amount, tax_rate, shipping_cost, total_amount, currency, status, approved_by, approved_at, notes, created_at, updated_at FROM purchase_orders WHERE id = $1`, id).
		Scan(&po.ID, &po.PONumber, &po.RequestID, &po.ProjectID, &po.VendorID, &po.OrderDate, &po.DeliveryDate, &po.DeliveryAddress, &po.PaymentTerms, &po.ShippingTerms, &po.Subtotal, &po.TaxAmount, &po.TaxRate, &po.ShippingCost, &po.TotalAmount, &po.Currency, &po.Status, &po.ApprovedBy, &po.ApprovedAt, &po.Notes, &po.CreatedAt, &po.UpdatedAt)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "purchase order not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, po)
}

func (h *ProcurementHandler) UpdatePurchaseOrder(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Status      *string  `json:"status"`
		DeliveryDate *string `json:"delivery_date"`
		TotalAmount *float64 `json:"total_amount"`
		Notes       *string  `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	now := time.Now()
	_, err := h.db.Exec(`UPDATE purchase_orders SET status=COALESCE($1,status), delivery_date=COALESCE($2,delivery_date), total_amount=COALESCE($3,total_amount), notes=COALESCE($4,notes), updated_at=$5 WHERE id=$6`,
		input.Status, input.DeliveryDate, input.TotalAmount, input.Notes, now, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *ProcurementHandler) DeletePurchaseOrder(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM purchase_orders WHERE id = $1`, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// --- Inventory ---

func (h *ProcurementHandler) ListInventory(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	query := `SELECT id, item_code, name, description, category, unit, unit_price, currency, min_quantity, max_quantity, current_quantity, reserved_quantity, available_quantity, storage_location, warehouse, material_type, is_active, notes, created_at, updated_at FROM inventory_items WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if category != "" {
		query += ` AND category = $` + itoa(argIdx)
		args = append(args, category)
		argIdx++
	}
	query += ` ORDER BY name`

	rows, err := h.db.Query(query, args...)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	items := make([]models.InventoryItem, 0)
	for rows.Next() {
		var inv models.InventoryItem
		if err := rows.Scan(&inv.ID, &inv.ItemCode, &inv.Name, &inv.Description, &inv.Category, &inv.Unit, &inv.UnitPrice, &inv.Currency, &inv.MinQuantity, &inv.MaxQuantity, &inv.CurrentQuantity, &inv.ReservedQuantity, &inv.AvailableQuantity, &inv.StorageLocation, &inv.Warehouse, &inv.MaterialType, &inv.IsActive, &inv.Notes, &inv.CreatedAt, &inv.UpdatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, inv)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *ProcurementHandler) CreateInventoryItem(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ItemCode        string   `json:"item_code"`
		Name            string   `json:"name"`
		Description     *string  `json:"description"`
		Category        *string  `json:"category"`
		Unit            string   `json:"unit"`
		UnitPrice       *float64 `json:"unit_price"`
		Currency        string   `json:"currency"`
		MinQuantity     float64  `json:"min_quantity"`
		MaxQuantity     *float64 `json:"max_quantity"`
		CurrentQuantity float64  `json:"current_quantity"`
		StorageLocation *string  `json:"storage_location"`
		Warehouse       *string  `json:"warehouse"`
		MaterialType    *string  `json:"material_type"`
		Notes           *string  `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO inventory_items (id, item_code, name, description, category, unit, unit_price, currency, min_quantity, max_quantity, current_quantity, reserved_quantity, available_quantity, storage_location, warehouse, material_type, is_active, notes, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,0,$11,$12,$13,$14,true,$15,$16,$17)`,
		id, input.ItemCode, input.Name, input.Description, input.Category, input.Unit, input.UnitPrice, input.Currency, input.MinQuantity, input.MaxQuantity, input.CurrentQuantity, input.StorageLocation, input.Warehouse, input.MaterialType, input.Notes, now, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *ProcurementHandler) GetInventoryItem(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var inv models.InventoryItem
	err := h.db.QueryRow(`SELECT id, item_code, name, description, category, unit, unit_price, currency, min_quantity, max_quantity, current_quantity, reserved_quantity, available_quantity, storage_location, warehouse, material_type, is_active, notes, created_at, updated_at FROM inventory_items WHERE id = $1`, id).
		Scan(&inv.ID, &inv.ItemCode, &inv.Name, &inv.Description, &inv.Category, &inv.Unit, &inv.UnitPrice, &inv.Currency, &inv.MinQuantity, &inv.MaxQuantity, &inv.CurrentQuantity, &inv.ReservedQuantity, &inv.AvailableQuantity, &inv.StorageLocation, &inv.Warehouse, &inv.MaterialType, &inv.IsActive, &inv.Notes, &inv.CreatedAt, &inv.UpdatedAt)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "inventory item not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, inv)
}

func (h *ProcurementHandler) UpdateInventoryItem(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		CurrentQuantity  *float64 `json:"current_quantity"`
		AvailableQuantity *float64 `json:"available_quantity"`
		UnitPrice        *float64 `json:"unit_price"`
		MinQuantity      *float64 `json:"min_quantity"`
		MaxQuantity      *float64 `json:"max_quantity"`
		Notes            *string  `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	now := time.Now()
	_, err := h.db.Exec(`UPDATE inventory_items SET current_quantity=COALESCE($1,current_quantity), available_quantity=COALESCE($2,available_quantity), unit_price=COALESCE($3,unit_price), min_quantity=COALESCE($4,min_quantity), max_quantity=COALESCE($5,max_quantity), notes=COALESCE($6,notes), updated_at=$7 WHERE id=$8`,
		input.CurrentQuantity, input.AvailableQuantity, input.UnitPrice, input.MinQuantity, input.MaxQuantity, input.Notes, now, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *ProcurementHandler) DeleteInventoryItem(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM inventory_items WHERE id = $1`, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// Procurement Request Items
// =============================================================================

func (h *ProcurementHandler) ListRequestItems(w http.ResponseWriter, r *http.Request) {
	requestID := chi.URLParam(r, "id")
	rows, err := h.db.Query(`SELECT id, request_id, line_number, item_code, description, specification, unit, quantity, estimated_unit_price, estimated_total, currency, boq_item_id, material_code, catalog_number, preferred_vendor, sort_order, created_at FROM procurement_request_items WHERE request_id = $1 ORDER BY line_number`, requestID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, rid, ic, desc, spec, unit, currency, mc, catNum, pv string
		var line int
		var qty, eup, et float64
		var boqID sql.NullString
		var so sql.NullInt64
		var createdAt time.Time
		if err := rows.Scan(&id, &rid, &line, &ic, &desc, &spec, &unit, &qty, &eup, &et, &currency, &boqID, &mc, &catNum, &pv, &so, &createdAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, map[string]interface{}{
			"id": id, "request_id": rid, "line_number": line, "item_code": ic,
			"description": desc, "specification": spec, "unit": unit, "quantity": qty,
			"estimated_unit_price": eup, "estimated_total": et, "currency": currency,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *ProcurementHandler) CreateRequestItem(w http.ResponseWriter, r *http.Request) {
	requestID := chi.URLParam(r, "id")
	var input struct {
		ItemCode         string   `json:"item_code"`
		Description      string   `json:"description"`
		Specification    *string  `json:"specification"`
		Unit             string   `json:"unit"`
		Quantity         float64  `json:"quantity"`
		EstimatedUnitPrice *float64 `json:"estimated_unit_price"`
		Currency         string   `json:"currency"`
		BOQItemID        *string  `json:"boq_item_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	estTotal := 0.0
	if input.EstimatedUnitPrice != nil {
		estTotal = input.Quantity * *input.EstimatedUnitPrice
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO procurement_request_items (id, request_id, line_number, item_code, description, specification, unit, quantity, estimated_unit_price, estimated_total, currency, boq_item_id, created_at) VALUES ($1,$2,(SELECT COALESCE(MAX(line_number),0)+1 FROM procurement_request_items WHERE request_id=$2),$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		id, requestID, input.ItemCode, input.Description, input.Specification, input.Unit, input.Quantity, input.EstimatedUnitPrice, estTotal, input.Currency, input.BOQItemID, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

// =============================================================================
// Purchase Order Items
// =============================================================================

func (h *ProcurementHandler) ListPurchaseOrderItems(w http.ResponseWriter, r *http.Request) {
	poID := chi.URLParam(r, "id")
	rows, err := h.db.Query(`SELECT id, po_id, line_number, request_item_id, item_code, description, specification, unit, quantity_ordered, quantity_received, unit_price, total_price, currency, delivery_date, sort_order, created_at FROM purchase_order_items WHERE po_id = $1 ORDER BY line_number`, poID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, pid, ic, desc, spec, unit, currency string
		var line int
		var qtyOrd, qtyRec, up, tp float64
		var dd, createdAt time.Time
		if err := rows.Scan(&id, &pid, &line, &ic, &desc, &spec, &unit, &qtyOrd, &qtyRec, &up, &tp, &currency, &dd, &createdAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, map[string]interface{}{
			"id": id, "po_id": pid, "line_number": line, "item_code": ic,
			"description": desc, "unit": unit, "quantity_ordered": qtyOrd,
			"quantity_received": qtyRec, "unit_price": up, "total_price": tp,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *ProcurementHandler) CreatePurchaseOrderItem(w http.ResponseWriter, r *http.Request) {
	poID := chi.URLParam(r, "id")
	var input struct {
		ItemCode    string  `json:"item_code"`
		Description string  `json:"description"`
		Specification *string `json:"specification"`
		Unit        string  `json:"unit"`
		Quantity    float64 `json:"quantity"`
		UnitPrice   float64 `json:"unit_price"`
		Currency    string  `json:"currency"`
		DeliveryDate *string `json:"delivery_date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	totalPrice := input.Quantity * input.UnitPrice
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO purchase_order_items (id, po_id, line_number, item_code, description, specification, unit, quantity_ordered, quantity_received, unit_price, total_price, currency, delivery_date, created_at) VALUES ($1,$2,(SELECT COALESCE(MAX(line_number),0)+1 FROM purchase_order_items WHERE po_id=$2),$3,$4,$5,$6,$7,0,$8,$9,$10,$11,$12)`,
		id, poID, input.ItemCode, input.Description, input.Specification, input.Unit, input.Quantity, input.UnitPrice, totalPrice, input.Currency, input.DeliveryDate, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

// =============================================================================
// Goods Receipts
// =============================================================================

func (h *ProcurementHandler) ListGoodsReceipts(w http.ResponseWriter, r *http.Request) {
	poID := r.URL.Query().Get("po_id")
	status := r.URL.Query().Get("status")
	query := `SELECT id, receipt_number, po_id, receipt_date, received_by, status, notes, created_at FROM goods_receipts WHERE 1=1`
	args := []interface{}{}
	argIdx := 1
	if poID != "" { query += ` AND po_id = $` + itoa(argIdx); args = append(args, poID); argIdx++ }
	if status != "" { query += ` AND status = $` + itoa(argIdx); args = append(args, status); argIdx++ }
	query += ` ORDER BY receipt_date DESC`

	rows, err := h.db.Query(query, args...)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, rn, pid, rb, st, notes string
		var rd, createdAt time.Time
		if err := rows.Scan(&id, &rn, &pid, &rd, &rb, &st, &notes, &createdAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, map[string]interface{}{
			"id": id, "receipt_number": rn, "po_id": pid, "receipt_date": rd,
			"received_by": rb, "status": st, "notes": notes, "created_at": createdAt,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *ProcurementHandler) CreateGoodsReceipt(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ReceiptNumber string `json:"receipt_number"`
		POID          string `json:"po_id"`
		ReceiptDate   string `json:"receipt_date"`
		ReceivedBy    string `json:"received_by"`
		Notes         *string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO goods_receipts (id, receipt_number, po_id, receipt_date, received_by, status, notes, created_at) VALUES ($1,$2,$3,$4,$5,'pending',$6,$7)`,
		id, input.ReceiptNumber, input.POID, input.ReceiptDate, input.ReceivedBy, input.Notes, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *ProcurementHandler) GetGoodsReceipt(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var gid, rn, pid, rb, st, notes string
	var rd, createdAt time.Time
	err := h.db.QueryRow(`SELECT id, receipt_number, po_id, receipt_date, received_by, status, notes, created_at FROM goods_receipts WHERE id = $1`, id).Scan(&gid, &rn, &pid, &rd, &rb, &st, &notes, &createdAt)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "goods receipt not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"id": gid, "receipt_number": rn, "po_id": pid, "receipt_date": rd,
		"received_by": rb, "status": st, "notes": notes,
	})
}

func (h *ProcurementHandler) UpdateGoodsReceipt(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Status *string `json:"status"`
		Notes  *string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	_, err := h.db.Exec(`UPDATE goods_receipts SET status=COALESCE($1,status), notes=COALESCE($2,notes) WHERE id=$3`, input.Status, input.Notes, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *ProcurementHandler) DeleteGoodsReceipt(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM goods_receipts WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// --- Goods Receipt Items ---

func (h *ProcurementHandler) ListGoodsReceiptItems(w http.ResponseWriter, r *http.Request) {
	receiptID := chi.URLParam(r, "id")
	rows, err := h.db.Query(`SELECT id, receipt_id, po_item_id, item_code, description, unit, quantity_ordered, quantity_received, quantity_accepted, quantity_rejected, rejection_reason, unit_price, total_price, batch_number, serial_number, expiry_date, storage_location, created_at FROM goods_receipt_items WHERE receipt_id = $1`, receiptID)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, rid, ic, desc, unit, reason, batch, serial, loc string
		var qtyOrd, qtyRec, qtyAcc, qtyRej, up, tp float64
		var createdAt time.Time
		var expiry sql.NullTime
		var poiid sql.NullString
		if err := rows.Scan(&id, &rid, &poiid, &ic, &desc, &unit, &qtyOrd, &qtyRec, &qtyAcc, &qtyRej, &reason, &up, &tp, &batch, &serial, &expiry, &loc, &createdAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		items = append(items, map[string]interface{}{
			"id": id, "receipt_id": rid, "item_code": ic, "description": desc, "unit": unit,
			"quantity_ordered": qtyOrd, "quantity_received": qtyRec, "quantity_accepted": qtyAcc,
			"quantity_rejected": qtyRej, "rejection_reason": reason, "unit_price": up, "total_price": tp,
			"batch_number": batch, "serial_number": serial, "storage_location": loc,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *ProcurementHandler) CreateGoodsReceiptItem(w http.ResponseWriter, r *http.Request) {
	receiptID := chi.URLParam(r, "id")
	var input struct {
		POItemID          *string `json:"po_item_id"`
		ItemCode          string  `json:"item_code"`
		Description       *string `json:"description"`
		Unit              string  `json:"unit"`
		QuantityReceived  float64 `json:"quantity_received"`
		QuantityAccepted  float64 `json:"quantity_accepted"`
		QuantityRejected  float64 `json:"quantity_rejected"`
		RejectionReason   *string `json:"rejection_reason"`
		UnitPrice         float64 `json:"unit_price"`
		BatchNumber       *string `json:"batch_number"`
		SerialNumber      *string `json:"serial_number"`
		StorageLocation   *string `json:"storage_location"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	totalPrice := input.QuantityAccepted * input.UnitPrice
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO goods_receipt_items (id, receipt_id, po_item_id, item_code, description, unit, quantity_ordered, quantity_received, quantity_accepted, quantity_rejected, rejection_reason, unit_price, total_price, batch_number, serial_number, storage_location, created_at) VALUES ($1,$2,$3,$4,$5,$6,0,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)`,
		id, receiptID, input.POItemID, input.ItemCode, input.Description, input.Unit, input.QuantityReceived, input.QuantityAccepted, input.QuantityRejected, input.RejectionReason, input.UnitPrice, totalPrice, input.BatchNumber, input.SerialNumber, input.StorageLocation, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

// =============================================================================
// Inventory Movements
// =============================================================================

func (h *ProcurementHandler) ListInventoryMovements(w http.ResponseWriter, r *http.Request) {
	itemID := r.URL.Query().Get("item_id")
	movementType := r.URL.Query().Get("movement_type")
	query := `SELECT im.id, im.item_id, ii.item_code, ii.name, im.movement_date, im.movement_type, im.quantity, im.unit_price, im.total_price, im.reference_type, im.reference_id, im.from_location, im.to_location, im.performed_by, im.notes, im.created_at 
		FROM inventory_movements im JOIN inventory_items ii ON ii.id = im.item_id WHERE 1=1`
	args := []interface{}{}
	argIdx := 1
	if itemID != "" { query += ` AND im.item_id = $` + itoa(argIdx); args = append(args, itemID); argIdx++ }
	if movementType != "" { query += ` AND im.movement_type = $` + itoa(argIdx); args = append(args, movementType); argIdx++ }
	query += ` ORDER BY im.movement_date DESC LIMIT 200`

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, iid, ic, name, mtype, reftype, refid, fromLoc, toLoc, performedBy, notes string
		var mdate, createdAt time.Time
		var qty, up, tp float64
		if err := rows.Scan(&id, &iid, &ic, &name, &mdate, &mtype, &qty, &up, &tp, &reftype, &refid, &fromLoc, &toLoc, &performedBy, &notes, &createdAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		items = append(items, map[string]interface{}{
			"id": id, "item_id": iid, "item_code": ic, "item_name": name,
			"movement_date": mdate, "movement_type": mtype, "quantity": qty,
			"unit_price": up, "total_price": tp, "reference_type": reftype,
			"reference_id": refid, "from_location": fromLoc, "to_location": toLoc,
			"performed_by": performedBy, "notes": notes,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *ProcurementHandler) CreateInventoryMovement(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ItemID       string  `json:"item_id"`
		MovementType string  `json:"movement_type"`
		Quantity     float64 `json:"quantity"`
		UnitPrice    *float64 `json:"unit_price"`
		ReferenceType *string `json:"reference_type"`
		ReferenceID  *string `json:"reference_id"`
		FromLocation *string `json:"from_location"`
		ToLocation   *string `json:"to_location"`
		PerformedBy  string  `json:"performed_by"`
		Notes        *string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	id := uuid.New().String()
	now := time.Now()

	// Get current unit price if not provided
	if input.UnitPrice == nil {
		h.db.QueryRow(`SELECT unit_price FROM inventory_items WHERE id = $1`, input.ItemID).Scan(&input.UnitPrice)
	}
	totalPrice := input.Quantity * *input.UnitPrice

	tx, err := h.db.Begin()
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }

	// Record movement
	_, err = tx.Exec(`INSERT INTO inventory_movements (id, item_id, movement_date, movement_type, quantity, unit_price, total_price, reference_type, reference_id, from_location, to_location, performed_by, notes, created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		id, input.ItemID, now, input.MovementType, input.Quantity, input.UnitPrice, totalPrice, input.ReferenceType, input.ReferenceID, input.FromLocation, input.ToLocation, input.PerformedBy, input.Notes, now)
	if err != nil { tx.Rollback(); respondError(w, http.StatusInternalServerError, err.Error()); return }

	// Update inventory quantities
	sign := 1.0
	if input.MovementType == "issue" || input.MovementType == "transfer" {
		sign = -1.0
	}
	_, err = tx.Exec(`UPDATE inventory_items SET current_quantity = current_quantity + ($1 * $2), available_quantity = GREATEST(0, available_quantity + ($1 * $2)), updated_at = $3 WHERE id = $4`,
		sign, input.Quantity, now, input.ItemID)
	if err != nil { tx.Rollback(); respondError(w, http.StatusInternalServerError, err.Error()); return }

	if err := tx.Commit(); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id, "status": "movement_recorded"})
}

// =============================================================================
// Stock Alerts — items below minimum stock
// =============================================================================

func (h *ProcurementHandler) ListStockAlerts(w http.ResponseWriter, r *http.Request) {
	warehouse := r.URL.Query().Get("warehouse")
	query := `SELECT id, item_code, name, category, unit, current_quantity, reserved_quantity, available_quantity, min_quantity, max_quantity, storage_location, warehouse
		FROM inventory_items WHERE current_quantity <= min_quantity AND is_active = true`
	args := []interface{}{}
	argIdx := 1
	if warehouse != "" { query += ` AND warehouse = $` + itoa(argIdx); args = append(args, warehouse); argIdx++ }
	query += ` ORDER BY (min_quantity - current_quantity) DESC`

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, ic, name, cat, unit, loc, wh string
		var cur, res, avail, min, max float64
		if err := rows.Scan(&id, &ic, &name, &cat, &unit, &cur, &res, &avail, &min, &max, &loc, &wh); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		items = append(items, map[string]interface{}{
			"id": id, "item_code": ic, "name": name, "category": cat, "unit": unit,
			"current_quantity": cur, "reserved_quantity": res, "available_quantity": avail,
			"min_quantity": min, "max_quantity": max, "storage_location": loc, "warehouse": wh,
			"reorder_quantity": max - cur,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

// =============================================================================
// Inventory Summary
// =============================================================================

func (h *ProcurementHandler) InventorySummary(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")

	rows, err := h.db.Query(`
		SELECT
			COALESCE(COUNT(*), 0) AS total_items,
			COALESCE(SUM(current_quantity * unit_price), 0) AS total_stock_value,
			COALESCE(SUM(current_quantity), 0) AS total_units,
			COALESCE(COUNT(*) FILTER (WHERE current_quantity <= min_quantity), 0) AS low_stock_count,
			COALESCE(COUNT(*) FILTER (WHERE current_quantity <= 0), 0) AS out_of_stock_count,
			COUNT(DISTINCT category) AS category_count,
			COUNT(DISTINCT warehouse) AS warehouse_count
		FROM inventory_items WHERE is_active = true`)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	var items, lowStock, outOfStock, catCount, whCount int
	var totalValue, totalUnits float64
	if rows.Next() {
		rows.Scan(&items, &totalValue, &totalUnits, &lowStock, &outOfStock, &catCount, &whCount)
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"total_items": items, "total_stock_value": totalValue, "total_units": totalUnits,
		"low_stock_count": lowStock, "out_of_stock_count": outOfStock,
		"category_count": catCount, "warehouse_count": whCount,
	})
}

// =============================================================================
// Vendor Evaluations
// =============================================================================

func (h *ProcurementHandler) ListVendorEvaluations(w http.ResponseWriter, r *http.Request) {
	vendorID := r.URL.Query().Get("vendor_id")
	query := `SELECT ve.id, ve.vendor_id, o.name AS vendor_name, ve.evaluation_date, ve.overall_score, ve.comments, ve.is_approved, ve.valid_until, ve.created_at 
		FROM vendor_evaluations ve JOIN organizations o ON o.id = ve.vendor_id WHERE 1=1`
	args := []interface{}{}
	argIdx := 1
	if vendorID != "" { query += ` AND ve.vendor_id = $` + itoa(argIdx); args = append(args, vendorID); argIdx++ }
	query += ` ORDER BY ve.evaluation_date DESC`

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, vid, vname, comments string
		var ed, vu, createdAt time.Time
		var score float64
		var approved bool
		if err := rows.Scan(&id, &vid, &vname, &ed, &score, &comments, &approved, &vu, &createdAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		items = append(items, map[string]interface{}{
			"id": id, "vendor_id": vid, "vendor_name": vname, "evaluation_date": ed,
			"overall_score": score, "comments": comments, "is_approved": approved, "valid_until": vu,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *ProcurementHandler) CreateVendorEvaluation(w http.ResponseWriter, r *http.Request) {
	var input struct {
		VendorID      string  `json:"vendor_id"`
		EvaluationDate string `json:"evaluation_date"`
		OverallScore  float64 `json:"overall_score"`
		Comments      *string `json:"comments"`
		IsApproved    bool    `json:"is_approved"`
		ValidUntil    *string `json:"valid_until"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO vendor_evaluations (id, vendor_id, evaluation_date, overall_score, comments, is_approved, valid_until, evaluator, created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,'system',$8)`,
		id, input.VendorID, input.EvaluationDate, input.OverallScore, input.Comments, input.IsApproved, input.ValidUntil, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *ProcurementHandler) GetVendorEvaluation(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var vid, comments string
	var score float64
	var approved bool
	err := h.db.QueryRow(`SELECT vendor_id, overall_score, comments, is_approved FROM vendor_evaluations WHERE id = $1`, id).Scan(&vid, &score, &comments, &approved)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "vendor evaluation not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"id": id, "vendor_id": vid, "overall_score": score, "comments": comments, "is_approved": approved,
	})
}
