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

		// Purchase Orders
		r.Get("/purchase-orders", h.ListPurchaseOrders)
		r.Post("/purchase-orders", h.CreatePurchaseOrder)
		r.Get("/purchase-orders/{id}", h.GetPurchaseOrder)
		r.Put("/purchase-orders/{id}", h.UpdatePurchaseOrder)
		r.Delete("/purchase-orders/{id}", h.DeletePurchaseOrder)

		// Inventory
		r.Get("/inventory", h.ListInventory)
		r.Post("/inventory", h.CreateInventoryItem)
		r.Get("/inventory/{id}", h.GetInventoryItem)
		r.Put("/inventory/{id}", h.UpdateInventoryItem)
		r.Delete("/inventory/{id}", h.DeleteInventoryItem)
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
