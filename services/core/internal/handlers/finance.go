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

// FinanceHandler handles Finance module endpoints
type FinanceHandler struct {
	db *sql.DB
}

func NewFinanceHandler(db *sql.DB) *FinanceHandler {
	return &FinanceHandler{db: db}
}

func (h *FinanceHandler) RegisterRoutes(r chi.Router) {
	r.Route("/finance", func(r chi.Router) {
		// Budgets
		r.Get("/budgets", h.ListBudgets)
		r.Post("/budgets", h.CreateBudget)
		r.Get("/budgets/{id}", h.GetBudget)
		r.Put("/budgets/{id}", h.UpdateBudget)
		r.Delete("/budgets/{id}", h.DeleteBudget)

		// Budget Items
		r.Get("/budget-items", h.ListBudgetItems)
		r.Post("/budget-items", h.CreateBudgetItem)
		r.Get("/budget-items/{id}", h.GetBudgetItem)
		r.Put("/budget-items/{id}", h.UpdateBudgetItem)
		r.Delete("/budget-items/{id}", h.DeleteBudgetItem)

		// Cash Flow
		r.Get("/cash-flow", h.ListCashFlow)
		r.Post("/cash-flow", h.CreateCashFlow)
		r.Get("/cash-flow/{id}", h.GetCashFlow)
		r.Put("/cash-flow/{id}", h.UpdateCashFlow)
		r.Delete("/cash-flow/{id}", h.DeleteCashFlow)

		// Invoices
		r.Get("/invoices", h.ListInvoices)
		r.Post("/invoices", h.CreateInvoice)
		r.Get("/invoices/{id}", h.GetInvoice)
		r.Put("/invoices/{id}", h.UpdateInvoice)
		r.Delete("/invoices/{id}", h.DeleteInvoice)
	})
}

// --- Budgets ---

func (h *FinanceHandler) ListBudgets(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT id, project_id, version, name, description, budget_type, total_amount, currency, contingency_pct, contingency_amount, status, approved_by, approved_at, is_active, notes, created_at FROM project_budgets`
	var rows *sql.Rows
	var err error
	if projectID != "" {
		rows, err = h.db.Query(query+` WHERE project_id = $1 ORDER BY created_at DESC`, projectID)
	} else {
		rows, err = h.db.Query(query + ` ORDER BY created_at DESC`)
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	budgets := make([]models.ProjectBudget, 0)
	for rows.Next() {
		var b models.ProjectBudget
		if err := rows.Scan(&b.ID, &b.ProjectID, &b.Version, &b.Name, &b.Description, &b.BudgetType, &b.TotalAmount, &b.Currency, &b.ContingencyPct, &b.ContingencyAmount, &b.Status, &b.ApprovedBy, &b.ApprovedAt, &b.IsActive, &b.Notes, &b.CreatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		budgets = append(budgets, b)
	}
	respondJSON(w, http.StatusOK, budgets)
}

func (h *FinanceHandler) CreateBudget(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID   string   `json:"project_id"`
		Version     string   `json:"version"`
		Name        string   `json:"name"`
		Description *string  `json:"description"`
		BudgetType  string   `json:"budget_type"`
		TotalAmount float64  `json:"total_amount"`
		Currency    string   `json:"currency"`
		ContingencyPct *float64 `json:"contingency_pct"`
		Status      string   `json:"status"`
		Notes       *string  `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO project_budgets (id, project_id, version, name, description, budget_type, total_amount, currency, contingency_pct, status, notes, created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		id, input.ProjectID, input.Version, input.Name, input.Description, input.BudgetType, input.TotalAmount, input.Currency, input.ContingencyPct, input.Status, input.Notes, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *FinanceHandler) GetBudget(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var b models.ProjectBudget
	err := h.db.QueryRow(`SELECT id, project_id, version, name, description, budget_type, total_amount, currency, contingency_pct, contingency_amount, status, approved_by, approved_at, is_active, notes, created_at FROM project_budgets WHERE id = $1`, id).
		Scan(&b.ID, &b.ProjectID, &b.Version, &b.Name, &b.Description, &b.BudgetType, &b.TotalAmount, &b.Currency, &b.ContingencyPct, &b.ContingencyAmount, &b.Status, &b.ApprovedBy, &b.ApprovedAt, &b.IsActive, &b.Notes, &b.CreatedAt)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "budget not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, b)
}

func (h *FinanceHandler) UpdateBudget(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Name        *string  `json:"name"`
		TotalAmount *float64 `json:"total_amount"`
		Status      *string  `json:"status"`
		IsActive    *bool    `json:"is_active"`
		Notes       *string  `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	_, err := h.db.Exec(`UPDATE project_budgets SET name=COALESCE($1,name), total_amount=COALESCE($2,total_amount), status=COALESCE($3,status), is_active=COALESCE($4,is_active), notes=COALESCE($5,notes) WHERE id=$6`,
		input.Name, input.TotalAmount, input.Status, input.IsActive, input.Notes, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *FinanceHandler) DeleteBudget(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM project_budgets WHERE id = $1`, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// --- Budget Items ---

func (h *FinanceHandler) ListBudgetItems(w http.ResponseWriter, r *http.Request) {
	budgetID := r.URL.Query().Get("budget_id")
	query := `SELECT id, budget_id, parent_id, item_code, name, description, item_type, cbs_code, planned_amount, actual_amount, committed_amount, remaining_amount, currency, sort_order, is_leaf, notes, created_at FROM budget_items`
	var rows *sql.Rows
	var err error
	if budgetID != "" {
		rows, err = h.db.Query(query+` WHERE budget_id = $1 ORDER BY sort_order`, budgetID)
	} else {
		rows, err = h.db.Query(query + ` ORDER BY sort_order`)
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	items := make([]models.BudgetItem, 0)
	for rows.Next() {
		var bi models.BudgetItem
		if err := rows.Scan(&bi.ID, &bi.BudgetID, &bi.ParentID, &bi.ItemCode, &bi.Name, &bi.Description, &bi.ItemType, &bi.CBSCode, &bi.PlannedAmount, &bi.ActualAmount, &bi.CommittedAmount, &bi.RemainingAmount, &bi.Currency, &bi.SortOrder, &bi.IsLeaf, &bi.Notes, &bi.CreatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, bi)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *FinanceHandler) CreateBudgetItem(w http.ResponseWriter, r *http.Request) {
	var input struct {
		BudgetID        string   `json:"budget_id"`
		ParentID        *string  `json:"parent_id"`
		ItemCode        string   `json:"item_code"`
		Name            string   `json:"name"`
		Description     *string  `json:"description"`
		ItemType        string   `json:"item_type"`
		CBSCode         *string  `json:"cbs_code"`
		PlannedAmount   float64  `json:"planned_amount"`
		Currency        string   `json:"currency"`
		SortOrder       int      `json:"sort_order"`
		IsLeaf          bool     `json:"is_leaf"`
		Notes           *string  `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO budget_items (id, budget_id, parent_id, item_code, name, description, item_type, cbs_code, planned_amount, currency, sort_order, is_leaf, notes, created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		id, input.BudgetID, input.ParentID, input.ItemCode, input.Name, input.Description, input.ItemType, input.CBSCode, input.PlannedAmount, input.Currency, input.SortOrder, input.IsLeaf, input.Notes, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *FinanceHandler) GetBudgetItem(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var bi models.BudgetItem
	err := h.db.QueryRow(`SELECT id, budget_id, parent_id, item_code, name, description, item_type, cbs_code, planned_amount, actual_amount, committed_amount, remaining_amount, currency, sort_order, is_leaf, notes, created_at FROM budget_items WHERE id = $1`, id).
		Scan(&bi.ID, &bi.BudgetID, &bi.ParentID, &bi.ItemCode, &bi.Name, &bi.Description, &bi.ItemType, &bi.CBSCode, &bi.PlannedAmount, &bi.ActualAmount, &bi.CommittedAmount, &bi.RemainingAmount, &bi.Currency, &bi.SortOrder, &bi.IsLeaf, &bi.Notes, &bi.CreatedAt)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "budget item not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, bi)
}

func (h *FinanceHandler) UpdateBudgetItem(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		PlannedAmount  *float64 `json:"planned_amount"`
		ActualAmount   *float64 `json:"actual_amount"`
		CommittedAmount *float64 `json:"committed_amount"`
		Notes          *string  `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	_, err := h.db.Exec(`UPDATE budget_items SET planned_amount=COALESCE($1,planned_amount), actual_amount=COALESCE($2,actual_amount), committed_amount=COALESCE($3,committed_amount), notes=COALESCE($4,notes) WHERE id=$5`,
		input.PlannedAmount, input.ActualAmount, input.CommittedAmount, input.Notes, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *FinanceHandler) DeleteBudgetItem(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM budget_items WHERE id = $1`, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// --- Cash Flow ---

func (h *FinanceHandler) ListCashFlow(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	entryType := r.URL.Query().Get("entry_type")

	query := `SELECT id, project_id, contract_id, entry_date, entry_type, category, amount, currency, is_planned, description, reference_type, reference_id, status, reconciled_at, created_at FROM cash_flow WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if projectID != "" {
		query += ` AND project_id = $` + itoa(argIdx)
		args = append(args, projectID)
		argIdx++
	}
	if entryType != "" {
		query += ` AND entry_type = $` + itoa(argIdx)
		args = append(args, entryType)
		argIdx++
	}
	query += ` ORDER BY entry_date DESC`

	rows, err := h.db.Query(query, args...)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	entries := make([]models.CashFlow, 0)
	for rows.Next() {
		var cf models.CashFlow
		if err := rows.Scan(&cf.ID, &cf.ProjectID, &cf.ContractID, &cf.EntryDate, &cf.EntryType, &cf.Category, &cf.Amount, &cf.Currency, &cf.IsPlanned, &cf.Description, &cf.ReferenceType, &cf.ReferenceID, &cf.Status, &cf.ReconciledAt, &cf.CreatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		entries = append(entries, cf)
	}
	respondJSON(w, http.StatusOK, entries)
}

func (h *FinanceHandler) CreateCashFlow(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID     *string `json:"project_id"`
		ContractID    *string `json:"contract_id"`
		EntryDate     string  `json:"entry_date"`
		EntryType     string  `json:"entry_type"`
		Category      string  `json:"category"`
		Amount        float64 `json:"amount"`
		Currency      string  `json:"currency"`
		IsPlanned     bool    `json:"is_planned"`
		Description   *string `json:"description"`
		ReferenceType *string `json:"reference_type"`
		ReferenceID   *string `json:"reference_id"`
		Status        string  `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO cash_flow (id, project_id, contract_id, entry_date, entry_type, category, amount, currency, is_planned, description, reference_type, reference_id, status, created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		id, input.ProjectID, input.ContractID, input.EntryDate, input.EntryType, input.Category, input.Amount, input.Currency, input.IsPlanned, input.Description, input.ReferenceType, input.ReferenceID, input.Status, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *FinanceHandler) GetCashFlow(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var cf models.CashFlow
	err := h.db.QueryRow(`SELECT id, project_id, contract_id, entry_date, entry_type, category, amount, currency, is_planned, description, reference_type, reference_id, status, reconciled_at, created_at FROM cash_flow WHERE id = $1`, id).
		Scan(&cf.ID, &cf.ProjectID, &cf.ContractID, &cf.EntryDate, &cf.EntryType, &cf.Category, &cf.Amount, &cf.Currency, &cf.IsPlanned, &cf.Description, &cf.ReferenceType, &cf.ReferenceID, &cf.Status, &cf.ReconciledAt, &cf.CreatedAt)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "cash flow entry not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, cf)
}

func (h *FinanceHandler) UpdateCashFlow(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Amount   *float64 `json:"amount"`
		Status   *string  `json:"status"`
		Description *string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	_, err := h.db.Exec(`UPDATE cash_flow SET amount=COALESCE($1,amount), status=COALESCE($2,status), description=COALESCE($3,description) WHERE id=$4`,
		input.Amount, input.Status, input.Description, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *FinanceHandler) DeleteCashFlow(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM cash_flow WHERE id = $1`, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// --- Invoices ---

func (h *FinanceHandler) ListInvoices(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	invoiceType := r.URL.Query().Get("invoice_type")
	contractID := r.URL.Query().Get("contract_id")

	query := `SELECT id, invoice_number, invoice_type, contract_id, acceptance_id, issuer_id, recipient_id, invoice_date, due_date, amount, tax_amount, tax_rate, total_amount, currency, status, paid_at, payment_ref, notes, created_at FROM invoices WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if status != "" {
		query += ` AND status = $` + itoa(argIdx)
		args = append(args, status)
		argIdx++
	}
	if invoiceType != "" {
		query += ` AND invoice_type = $` + itoa(argIdx)
		args = append(args, invoiceType)
		argIdx++
	}
	if contractID != "" {
		query += ` AND contract_id = $` + itoa(argIdx)
		args = append(args, contractID)
		argIdx++
	}
	query += ` ORDER BY invoice_date DESC`

	rows, err := h.db.Query(query, args...)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	invoices := make([]models.Invoice, 0)
	for rows.Next() {
		var inv models.Invoice
		if err := rows.Scan(&inv.ID, &inv.InvoiceNumber, &inv.InvoiceType, &inv.ContractID, &inv.AcceptanceID, &inv.IssuerID, &inv.RecipientID, &inv.InvoiceDate, &inv.DueDate, &inv.Amount, &inv.TaxAmount, &inv.TaxRate, &inv.TotalAmount, &inv.Currency, &inv.Status, &inv.PaidAt, &inv.PaymentRef, &inv.Notes, &inv.CreatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		invoices = append(invoices, inv)
	}
	respondJSON(w, http.StatusOK, invoices)
}

func (h *FinanceHandler) CreateInvoice(w http.ResponseWriter, r *http.Request) {
	var input struct {
		InvoiceNumber string  `json:"invoice_number"`
		InvoiceType   string  `json:"invoice_type"`
		ContractID    *string `json:"contract_id"`
		AcceptanceID  *string `json:"acceptance_id"`
		IssuerID      *string `json:"issuer_id"`
		RecipientID   *string `json:"recipient_id"`
		InvoiceDate   string  `json:"invoice_date"`
		DueDate       *string `json:"due_date"`
		Amount        float64 `json:"amount"`
		TaxAmount     float64 `json:"tax_amount"`
		TaxRate       float64 `json:"tax_rate"`
		TotalAmount   float64 `json:"total_amount"`
		Currency      string  `json:"currency"`
		Status        string  `json:"status"`
		Notes         *string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO invoices (id, invoice_number, invoice_type, contract_id, acceptance_id, issuer_id, recipient_id, invoice_date, due_date, amount, tax_amount, tax_rate, total_amount, currency, status, notes, created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)`,
		id, input.InvoiceNumber, input.InvoiceType, input.ContractID, input.AcceptanceID, input.IssuerID, input.RecipientID, input.InvoiceDate, input.DueDate, input.Amount, input.TaxAmount, input.TaxRate, input.TotalAmount, input.Currency, input.Status, input.Notes, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *FinanceHandler) GetInvoice(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var inv models.Invoice
	err := h.db.QueryRow(`SELECT id, invoice_number, invoice_type, contract_id, acceptance_id, issuer_id, recipient_id, invoice_date, due_date, amount, tax_amount, tax_rate, total_amount, currency, status, paid_at, payment_ref, notes, created_at FROM invoices WHERE id = $1`, id).
		Scan(&inv.ID, &inv.InvoiceNumber, &inv.InvoiceType, &inv.ContractID, &inv.AcceptanceID, &inv.IssuerID, &inv.RecipientID, &inv.InvoiceDate, &inv.DueDate, &inv.Amount, &inv.TaxAmount, &inv.TaxRate, &inv.TotalAmount, &inv.Currency, &inv.Status, &inv.PaidAt, &inv.PaymentRef, &inv.Notes, &inv.CreatedAt)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "invoice not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, inv)
}

func (h *FinanceHandler) UpdateInvoice(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Status   *string `json:"status"`
		PaidAt   *string `json:"paid_at"`
		PaymentRef *string `json:"payment_ref"`
		Notes    *string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	_, err := h.db.Exec(`UPDATE invoices SET status=COALESCE($1,status), paid_at=COALESCE($2,paid_at), payment_ref=COALESCE($3,payment_ref), notes=COALESCE($4,notes) WHERE id=$5`,
		input.Status, input.PaidAt, input.PaymentRef, input.Notes, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *FinanceHandler) DeleteInvoice(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM invoices WHERE id = $1`, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
