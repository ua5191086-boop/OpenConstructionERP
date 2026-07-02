package handlers

import (
	"database/sql"
	"encoding/json"
	"math"
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

		// Cost Control / Earned Value (cost_control table)
		r.Get("/cost-control", h.ListCostControl)
		r.Post("/cost-control", h.CreateCostControl)
		r.Get("/cost-control/{id}", h.GetCostControl)
		r.Put("/cost-control/{id}", h.UpdateCostControl)
		r.Delete("/cost-control/{id}", h.DeleteCostControl)

		// Physical Progress (progress measurements from site)
		r.Get("/physical-progress", h.ListPhysicalProgress)
		r.Post("/physical-progress", h.CreatePhysicalProgress)
		r.Get("/physical-progress/{id}", h.GetPhysicalProgress)
		r.Put("/physical-progress/{id}", h.UpdatePhysicalProgress)

		// IPC vs Physical — comparison endpoint
		r.Get("/ipc-vs-physical", h.IPCvsPhysical)
		r.Get("/ipc-vs-physical/contract/{contractId}", h.IPCvsPhysicalByContract)

		// Financial Reports
		r.Get("/reports", h.ListReports)
		r.Post("/reports", h.CreateReport)
		r.Get("/reports/{id}", h.GetReport)
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

// =============================================================================
// Cost Control / Earned Value (таблица cost_control — IPC vs Budget)
// =============================================================================

func (h *FinanceHandler) ListCostControl(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT id, project_id, report_date, report_type, total_budget, total_committed, total_actual, total_forecast, variance_amount, variance_pct, earned_value, planned_value, spi, cpi, status, notes, created_at FROM cost_control WHERE 1=1`
	args := []interface{}{}
	argIdx := 1
	if projectID != "" {
		query += ` AND project_id = $` + itoa(argIdx)
		args = append(args, projectID)
		argIdx++
	}
	query += ` ORDER BY report_date DESC`

	rows, err := h.db.Query(query, args...)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, pid, rtype, st, notes string
		var rdate string
		var budget, committed, actual, forecast, variance, variancePct, ev, pv, spi, cpi float64
		var createdAt time.Time
		if err := rows.Scan(&id, &pid, &rdate, &rtype, &budget, &committed, &actual, &forecast, &variance, &variancePct, &ev, &pv, &spi, &cpi, &st, &notes, &createdAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, map[string]interface{}{
			"id": id, "project_id": pid, "report_date": rdate, "report_type": rtype,
			"total_budget": budget, "total_committed": committed, "total_actual": actual,
			"total_forecast": forecast, "variance_amount": variance, "variance_pct": variancePct,
			"earned_value": ev, "planned_value": pv, "spi": spi, "cpi": cpi,
			"status": st, "notes": notes, "created_at": createdAt,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *FinanceHandler) CreateCostControl(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID    string   `json:"project_id"`
		ReportDate   string   `json:"report_date"`
		ReportType   string   `json:"report_type"`
		TotalBudget  float64  `json:"total_budget"`
		TotalCommitted float64 `json:"total_committed"`
		TotalActual  float64  `json:"total_actual"`
		TotalForecast float64 `json:"total_forecast"`
		EarnedValue  float64  `json:"earned_value"`
		PlannedValue float64  `json:"planned_value"`
		Notes        *string  `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	variance := input.TotalActual - input.TotalBudget
	variancePct := 0.0
	if input.TotalBudget != 0 {
		variancePct = (variance / input.TotalBudget) * 100
	}
	spi := 0.0
	if input.PlannedValue != 0 {
		spi = input.EarnedValue / input.PlannedValue
	}
	cpi := 0.0
	if input.TotalActual != 0 {
		cpi = input.EarnedValue / input.TotalActual
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO cost_control (id, project_id, report_date, report_type, total_budget, total_committed, total_actual, total_forecast, variance_amount, variance_pct, earned_value, planned_value, spi, cpi, status, notes, created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,CASE WHEN $9 > 0 THEN 'red' WHEN $9 < 0 THEN 'green' ELSE 'green' END,$15,$16)`,
		id, input.ProjectID, input.ReportDate, input.ReportType, input.TotalBudget, input.TotalCommitted, input.TotalActual, input.TotalForecast, variance, variancePct, input.EarnedValue, input.PlannedValue, spi, cpi, input.Notes, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *FinanceHandler) GetCostControl(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var pid, rdate, rtype, st string
	var ev, pv, spi, cpi float64
	err := h.db.QueryRow(`SELECT project_id, report_date, report_type, earned_value, planned_value, spi, cpi, status FROM cost_control WHERE id = $1`, id).Scan(&pid, &rdate, &rtype, &ev, &pv, &spi, &cpi, &st)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "cost control record not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"id": id, "project_id": pid, "report_date": rdate, "report_type": rtype,
		"earned_value": ev, "planned_value": pv, "spi": spi, "cpi": cpi, "status": st,
	})
}

func (h *FinanceHandler) UpdateCostControl(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		TotalActual  *float64 `json:"total_actual"`
		EarnedValue  *float64 `json:"earned_value"`
		TotalForecast *float64 `json:"total_forecast"`
		Notes        *string  `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	_, err := h.db.Exec(`UPDATE cost_control SET total_actual=COALESCE($1,total_actual), earned_value=COALESCE($2,earned_value), total_forecast=COALESCE($3,total_forecast), notes=COALESCE($4,notes) WHERE id=$5`,
		input.TotalActual, input.EarnedValue, input.TotalForecast, input.Notes, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *FinanceHandler) DeleteCostControl(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM cost_control WHERE id = $1`, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// Physical Progress — физический прогресс с площадки
// =============================================================================

func (h *FinanceHandler) ListPhysicalProgress(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	contractID := r.URL.Query().Get("contract_id")
	query := `SELECT id, project_id, contract_id, boq_item_id, measurement_date, item_code, description, unit, contract_quantity, prev_cumulative_qty, current_qty, total_cumulative_qty, completion_pct, unit_price, ipc_amount, source, verified_by, verified_at, notes, created_at FROM physical_progress WHERE 1=1`
	args := []interface{}{}
	argIdx := 1
	if projectID != "" { query += ` AND project_id = $` + itoa(argIdx); argIdx++; args = append(args, projectID) }
	if contractID != "" { query += ` AND contract_id = $` + itoa(argIdx); argIdx++; args = append(args, contractID) }
	query += ` ORDER BY measurement_date DESC`

	rows, err := h.db.Query(query, args...)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, pid, cid, biid, ic, desc, unit, src, verifiedBy string
		var mdate string
		var cq, prev, cur, total, pct, up, ipc float64
		var verifiedAt, notes, createdAt sql.NullString
		if err := rows.Scan(&id, &pid, &cid, &biid, &mdate, &ic, &desc, &unit, &cq, &prev, &cur, &total, &pct, &up, &ipc, &src, &verifiedBy, &verifiedAt, &notes, &createdAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		item := map[string]interface{}{
			"id": id, "project_id": pid, "contract_id": cid, "boq_item_id": biid,
			"measurement_date": mdate, "item_code": ic, "description": desc, "unit": unit,
			"contract_quantity": cq, "prev_cumulative_qty": prev, "current_qty": cur,
			"total_cumulative_qty": total, "completion_pct": pct, "unit_price": up,
			"ipc_amount": ipc, "source": src, "verified_by": verifiedBy,
		}
		if verifiedAt.Valid { item["verified_at"] = verifiedAt.String }
		if notes.Valid { item["notes"] = notes.String }
		if createdAt.Valid { item["created_at"] = createdAt.String }
		items = append(items, item)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *FinanceHandler) CreatePhysicalProgress(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID      string   `json:"project_id"`
		ContractID     *string  `json:"contract_id"`
		BOQItemID      *string  `json:"boq_item_id"`
		MeasurementDate string  `json:"measurement_date"`
		ItemCode       string   `json:"item_code"`
		Description    *string  `json:"description"`
		Unit           string   `json:"unit"`
		ContractQty    float64  `json:"contract_quantity"`
		PrevCumQty     float64  `json:"prev_cumulative_qty"`
		CurrentQty     float64  `json:"current_qty"`
		UnitPrice      float64  `json:"unit_price"`
		Source         string   `json:"source"`
		VerifiedBy     *string  `json:"verified_by"`
		Notes          *string  `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	totalCum := input.PrevCumQty + input.CurrentQty
	pct := 0.0
	if input.ContractQty != 0 {
		pct = (totalCum / input.ContractQty) * 100
	}
	ipcAmount := input.CurrentQty * input.UnitPrice
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO physical_progress (id, project_id, contract_id, boq_item_id, measurement_date, item_code, description, unit, contract_quantity, prev_cumulative_qty, current_qty, total_cumulative_qty, completion_pct, unit_price, ipc_amount, source, verified_by, notes, created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)`,
		id, input.ProjectID, input.ContractID, input.BOQItemID, input.MeasurementDate, input.ItemCode, input.Description, input.Unit, input.ContractQty, input.PrevCumQty, input.CurrentQty, totalCum, pct, input.UnitPrice, ipcAmount, input.Source, input.VerifiedBy, input.Notes, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *FinanceHandler) GetPhysicalProgress(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var pid, ic, st string
	var pct, ipc float64
	err := h.db.QueryRow(`SELECT project_id, item_code, completion_pct, ipc_amount, source FROM physical_progress WHERE id = $1`, id).Scan(&pid, &ic, &pct, &ipc, &st)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "physical progress record not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"id": id, "project_id": pid, "item_code": ic, "completion_pct": pct, "ipc_amount": ipc, "source": st,
	})
}

func (h *FinanceHandler) UpdatePhysicalProgress(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		CurrentQty *float64 `json:"current_qty"`
		VerifiedBy *string  `json:"verified_by"`
		Notes      *string  `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	_, err := h.db.Exec(`UPDATE physical_progress SET current_qty=COALESCE($1,current_qty), verified_by=COALESCE($2,verified_by), verified_at=CASE WHEN $2 IS NOT NULL THEN NOW() ELSE verified_at END, notes=COALESCE($3,notes) WHERE id=$4`,
		input.CurrentQty, input.VerifiedBy, input.Notes, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// =============================================================================
// IPC vs Physical Progress — сравнительный анализ
// =============================================================================

func (h *FinanceHandler) IPCvsPhysical(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")

	q := `
	SELECT
		pp.project_id,
		pp.contract_id,
		c.code AS contract_code,
		c.name AS contract_name,
		COUNT(DISTINCT pp.id) AS progress_entries,
		COALESCE(SUM(pp.ipc_amount), 0) AS total_ipc_from_progress,
		COALESCE(SUM(a.amount), 0) AS total_accepted_ipc,
		COALESCE(SUM(a.amount), 0) - COALESCE(SUM(pp.ipc_amount), 0) AS ipc_variance,
		COALESCE(SUM(pp.completion_pct * pp.contract_quantity) / NULLIF(SUM(pp.contract_quantity), 0), 0) AS avg_physical_pct,
		CASE
			WHEN c.contract_amount > 0
			THEN (COALESCE(SUM(a.amount), 0) / c.contract_amount) * 100
			ELSE 0
		END AS ipc_pct_of_contract
	FROM physical_progress pp
	LEFT JOIN contracts c ON c.id = pp.contract_id
	LEFT JOIN contract_work_acceptances a ON a.contract_id = pp.contract_id AND a.status IN ('approved','paid')
	WHERE 1=1`
	args := []interface{}{}
	argIdx := 1
	if projectID != "" {
		q += ` AND pp.project_id = $` + itoa(argIdx)
		args = append(args, projectID)
		argIdx++
	}
	q += ` GROUP BY pp.project_id, pp.contract_id, c.code, c.name, c.contract_amount ORDER BY contract_code`

	rows, err := h.db.Query(q, args...)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var pid, cid, ccode, cname string
		var entries int
		var ipcFromProgress, acceptedIPC, variance, avgPhysPct, ipcPct float64
		if err := rows.Scan(&pid, &cid, &ccode, &cname, &entries, &ipcFromProgress, &acceptedIPC, &variance, &avgPhysPct, &ipcPct); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, map[string]interface{}{
			"project_id": pid, "contract_id": cid, "contract_code": ccode, "contract_name": cname,
			"progress_entries": entries, "total_ipc_from_progress": ipcFromProgress,
			"total_accepted_ipc": acceptedIPC, "ipc_variance": variance,
			"avg_physical_completion_pct": math.Round(avgPhysPct*100) / 100,
			"ipc_pct_of_contract": math.Round(ipcPct*100) / 100,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *FinanceHandler) IPCvsPhysicalByContract(w http.ResponseWriter, r *http.Request) {
	contractID := chi.URLParam(r, "contractId")

	rows, err := h.db.Query(`
		SELECT
			pp.id,
			pp.item_code,
			pp.description,
			pp.unit,
			pp.contract_quantity,
			pp.total_cumulative_qty,
			pp.completion_pct,
			pp.unit_price,
			pp.ipc_amount,
			pp.measurement_date,
			pp.source,
			pp.verified_by
		FROM physical_progress pp
		WHERE pp.contract_id = $1
		ORDER BY pp.item_code, pp.measurement_date DESC`, contractID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, ic, desc, unit, src, verifiedBy string
		var cq, totalCum, pct, up, ipc float64
		var mdate string
		if err := rows.Scan(&id, &ic, &desc, &unit, &cq, &totalCum, &pct, &up, &ipc, &mdate, &src, &verifiedBy); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, map[string]interface{}{
			"id": id, "item_code": ic, "description": desc, "unit": unit,
			"contract_quantity": cq, "total_cumulative_qty": totalCum,
			"completion_pct": pct, "unit_price": up, "ipc_amount": ipc,
			"measurement_date": mdate, "source": src, "verified_by": verifiedBy,
		})
	}

	// Summary
	var totalContract, totalAccepted, totalProgress float64
	h.db.QueryRow(`SELECT contract_amount FROM contracts WHERE id = $1`, contractID).Scan(&totalContract)
	h.db.QueryRow(`SELECT COALESCE(SUM(amount),0) FROM contract_work_acceptances WHERE contract_id = $1 AND status IN ('approved','paid')`, contractID).Scan(&totalAccepted)
	h.db.QueryRow(`SELECT COALESCE(SUM(ipc_amount),0) FROM physical_progress WHERE contract_id = $1`, contractID).Scan(&totalProgress)

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"contract_id":            contractID,
		"contract_amount":        totalContract,
		"total_accepted_ipc":     totalAccepted,
		"total_ipc_from_progress": totalProgress,
		"ipc_variance":           totalAccepted - totalProgress,
		"items":                  items,
	})
}

// =============================================================================
// Financial Reports
// =============================================================================

func (h *FinanceHandler) ListReports(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT id, project_id, report_type, report_period, total_revenue, total_expense, net_profit, currency, status, generated_at, generated_by, notes FROM financial_reports WHERE 1=1`
	args := []interface{}{}
	argIdx := 1
	if projectID != "" {
		query += ` AND project_id = $` + itoa(argIdx)
		args = append(args, projectID)
		argIdx++
	}
	query += ` ORDER BY generated_at DESC`

	rows, err := h.db.Query(query, args...)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, pid, rtype, period, currency, st, genBy, notes string
		var rev, exp, profit float64
		var genAt time.Time
		if err := rows.Scan(&id, &pid, &rtype, &period, &rev, &exp, &profit, &currency, &st, &genAt, &genBy, &notes); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, map[string]interface{}{
			"id": id, "project_id": pid, "report_type": rtype, "report_period": period,
			"total_revenue": rev, "total_expense": exp, "net_profit": profit,
			"currency": currency, "status": st, "generated_at": genAt,
			"generated_by": genBy, "notes": notes,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *FinanceHandler) CreateReport(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID  string  `json:"project_id"`
		ReportType string  `json:"report_type"`
		Period     string  `json:"report_period"`
		Revenue    float64 `json:"total_revenue"`
		Expense    float64 `json:"total_expense"`
		Currency   string  `json:"currency"`
		Notes      *string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	now := time.Now()
	profit := input.Revenue - input.Expense
	_, err := h.db.Exec(`INSERT INTO financial_reports (id, project_id, report_type, report_period, total_revenue, total_expense, net_profit, currency, status, generated_at, notes) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,'draft',$9,$10)`,
		id, input.ProjectID, input.ReportType, input.Period, input.Revenue, input.Expense, profit, input.Currency, now, input.Notes)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *FinanceHandler) GetReport(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var pid, rtype, period, st string
	var rev, exp, profit float64
	err := h.db.QueryRow(`SELECT project_id, report_type, report_period, total_revenue, total_expense, net_profit, status FROM financial_reports WHERE id = $1`, id).Scan(&pid, &rtype, &period, &rev, &exp, &profit, &st)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "financial report not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"id": id, "project_id": pid, "report_type": rtype, "report_period": period,
		"total_revenue": rev, "total_expense": exp, "net_profit": profit, "status": st,
	})
}
