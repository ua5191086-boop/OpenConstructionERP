package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
)

// --- Audit Trail ---

// AuditEntry — иммутабельная запись аудита
type AuditEntry struct {
	ID               string   `json:"id" db:"id"`
	ProjectID        *string  `json:"project_id" db:"project_id"`
	EntityType       string   `json:"entity_type" db:"entity_type"`
	EntityID         string   `json:"entity_id" db:"entity_id"`
	Action           string   `json:"action" db:"action"`
	FieldName        *string  `json:"field_name" db:"field_name"`
	OldValue         *string  `json:"old_value" db:"old_value"`
	NewValue         *string  `json:"new_value" db:"new_value"`
	ChangedBy        string   `json:"changed_by" db:"changed_by"`
	ChangedByRole    *string  `json:"changed_by_role" db:"changed_by_role"`
	ChangeReason     *string  `json:"change_reason" db:"change_reason"`
	FinancialImpact  *float64 `json:"financial_impact" db:"financial_impact"`
	Currency         *string  `json:"currency" db:"currency"`
	IsFinancial      bool     `json:"is_financial" db:"is_financial"`
	Checksum         string   `json:"checksum" db:"checksum"`
	PrevChecksum     *string  `json:"previous_checksum" db:"previous_checksum"`
	IPAddress        *string  `json:"ip_address" db:"ip_address"`
	UserAgent        *string  `json:"user_agent" db:"user_agent"`
	CreatedAt        string   `json:"created_at" db:"created_at"`
}

// TaxInvoice — налоговый счёт-фактура
type TaxInvoice struct {
	ID                string   `json:"id" db:"id"`
	ProjectID         string   `json:"project_id" db:"project_id"`
	ContractID        *string  `json:"contract_id" db:"contract_id"`
	InvoiceType       string   `json:"invoice_type" db:"invoice_type"`
	InvoiceNumber     string   `json:"invoice_number" db:"invoice_number"`
	InvoiceDate       string   `json:"invoice_date" db:"invoice_date"`
	Counterparty      string   `json:"counterparty" db:"counterparty"`
	CounterpartyTaxID *string  `json:"counterparty_tax_id" db:"counterparty_tax_id"`
	GrossAmount       float64  `json:"gross_amount" db:"gross_amount"`
	NetAmount         float64  `json:"net_amount" db:"net_amount"`
	TaxAmount         float64  `json:"tax_amount" db:"tax_amount"`
	TaxRate           float64  `json:"tax_rate" db:"tax_rate"`
	TaxCode           *string  `json:"tax_code" db:"tax_code"`
	Currency          string   `json:"currency" db:"currency"`
	Status            string   `json:"status" db:"status"`
	DueDate           *string  `json:"due_date" db:"due_date"`
	PaidDate          *string  `json:"paid_date" db:"paid_date"`
	IsReverseCharge   bool     `json:"is_reverse_charge" db:"is_reverse_charge"`
	FiscalPeriod      *string  `json:"fiscal_period" db:"fiscal_period"`
	DocumentRef       *string  `json:"document_ref" db:"document_ref"`
	Notes             *string  `json:"notes" db:"notes"`
	CreatedAt         string   `json:"created_at" db:"created_at"`
	UpdatedAt         string   `json:"updated_at" db:"updated_at"`
}

// AuditHandler — HTTP handler for audit trail, tax, stakeholders, ESG
type AuditHandler struct {
	db *sqlx.DB
}

func NewAuditHandler(db *sqlx.DB) *AuditHandler {
	return &AuditHandler{db: db}
}

func (h *AuditHandler) RegisterRoutes(r chi.Router) {
	// Audit Trail
	r.Route("/audit", func(r chi.Router) {
		r.Get("/", h.ListAudit)
		r.Post("/", h.CreateAudit)
		r.Get("/{id}", h.GetAudit)
		r.Get("/entity/{entityType}/{entityId}", h.ListAuditByEntity)
		r.Get("/project/{projectId}", h.ListAuditByProject)
		r.Get("/financial", h.ListFinancialAudit)
		r.Get("/verify/{entityType}/{entityId}", h.VerifyAuditChain)
	})
	// Tax Invoices
	r.Route("/tax-invoices", func(r chi.Router) {
		r.Get("/", h.ListTaxInvoices)
		r.Post("/", h.CreateTaxInvoice)
		r.Get("/{id}", h.GetTaxInvoice)
		r.Put("/{id}", h.UpdateTaxInvoice)
		r.Delete("/{id}", h.DeleteTaxInvoice)
		r.Get("/project/{projectId}", h.ListTaxInvoicesByProject)
		r.Get("/fiscal/{period}", h.ListTaxInvoicesByFiscal)
		r.Get("/status/{status}", h.ListTaxInvoicesByStatus)
	})
	// Tax Returns
	r.Route("/tax-returns", func(r chi.Router) {
		r.Get("/", h.ListTaxReturns)
		r.Post("/", h.CreateTaxReturn)
		r.Get("/{id}", h.GetTaxReturn)
		r.Put("/{id}", h.UpdateTaxReturn)
		r.Get("/project/{projectId}", h.ListTaxReturnsByProject)
	})
	// Stakeholders
	r.Route("/stakeholders", func(r chi.Router) {
		r.Get("/", h.ListStakeholders)
		r.Post("/", h.CreateStakeholder)
		r.Get("/{id}", h.GetStakeholder)
		r.Put("/{id}", h.UpdateStakeholder)
		r.Delete("/{id}", h.DeleteStakeholder)
		r.Get("/project/{projectId}", h.ListStakeholdersByProject)
		r.Get("/type/{stakeholderType}", h.ListStakeholdersByType)
	})
	// Stakeholder Communications
	r.Route("/stakeholder-comms", func(r chi.Router) {
		r.Get("/", h.ListStakeholderComms)
		r.Post("/", h.CreateStakeholderComm)
		r.Get("/{id}", h.GetStakeholderComm)
		r.Get("/stakeholder/{stakeholderId}", h.ListCommsByStakeholder)
		r.Get("/pending-followup", h.ListPendingFollowUps)
	})
	// ESG Metrics
	r.Route("/esg", func(r chi.Router) {
		r.Get("/", h.ListESGMetrics)
		r.Post("/", h.CreateESGMetric)
		r.Get("/{id}", h.GetESGMetric)
		r.Put("/{id}", h.UpdateESGMetric)
		r.Get("/project/{projectId}", h.ListESGMetricsByProject)
		r.Get("/category/{category}", h.ListESGMetricsByCategory)
	})
	// Carbon Footprint
	r.Route("/carbon", func(r chi.Router) {
		r.Get("/", h.ListCarbon)
		r.Post("/", h.CreateCarbon)
		r.Get("/{id}", h.GetCarbon)
		r.Put("/{id}", h.UpdateCarbon)
		r.Get("/project/{projectId}", h.ListCarbonByProject)
		r.Get("/scope/{scope}", h.ListCarbonByScope)
	})
}

// ---- Audit Trail ----

func (h *AuditHandler) ListAudit(w http.ResponseWriter, r *http.Request) {
	var items []AuditEntry
	if err := h.db.Select(&items, "SELECT * FROM audit_trail ORDER BY created_at DESC LIMIT 100"); err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("query failed: %v", err))
		return
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *AuditHandler) CreateAudit(w http.ResponseWriter, r *http.Request) {
	var input AuditEntry
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}
	var item AuditEntry
	err := h.db.Get(&item, `INSERT INTO audit_trail
		(project_id, entity_type, entity_id, action, field_name, old_value, new_value,
		 changed_by, changed_by_role, change_reason, financial_impact, currency, is_financial,
		 ip_address, user_agent)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
		RETURNING *`,
		nullableString(input.ProjectID), input.EntityType, input.EntityID, input.Action,
		nullableString(input.FieldName), nullableString(input.OldValue), nullableString(input.NewValue),
		input.ChangedBy, nullableString(input.ChangedByRole), nullableString(input.ChangeReason),
		nullableFloat(input.FinancialImpact), nullableString(input.Currency), input.IsFinancial,
		nullableString(input.IPAddress), nullableString(input.UserAgent))
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("insert failed: %v", err))
		return
	}
	respondJSON(w, http.StatusCreated, item)
}

func (h *AuditHandler) GetAudit(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var item AuditEntry
	if err := h.db.Get(&item, "SELECT * FROM audit_trail WHERE id=$1", id); err != nil {
		respondError(w, http.StatusNotFound, "audit entry not found")
		return
	}
	respondJSON(w, http.StatusOK, item)
}

func (h *AuditHandler) ListAuditByEntity(w http.ResponseWriter, r *http.Request) {
	entityType := chi.URLParam(r, "entityType")
	entityID := chi.URLParam(r, "entityId")
	var items []AuditEntry
	h.db.Select(&items, "SELECT * FROM audit_trail WHERE entity_type=$1 AND entity_id=$2 ORDER BY created_at DESC", entityType, entityID)
	respondJSON(w, http.StatusOK, items)
}

func (h *AuditHandler) ListAuditByProject(w http.ResponseWriter, r *http.Request) {
	pid := chi.URLParam(r, "projectId")
	var items []AuditEntry
	h.db.Select(&items, "SELECT * FROM audit_trail WHERE project_id=$1 ORDER BY created_at DESC", pid)
	respondJSON(w, http.StatusOK, items)
}

func (h *AuditHandler) ListFinancialAudit(w http.ResponseWriter, r *http.Request) {
	var items []AuditEntry
	h.db.Select(&items, "SELECT * FROM audit_trail WHERE is_financial=TRUE ORDER BY created_at DESC LIMIT 100")
	respondJSON(w, http.StatusOK, items)
}

func (h *AuditHandler) VerifyAuditChain(w http.ResponseWriter, r *http.Request) {
	entityType := chi.URLParam(r, "entityType")
	entityID := chi.URLParam(r, "entityId")
	var entries []AuditEntry
	h.db.Select(&entries, "SELECT * FROM audit_trail WHERE entity_type=$1 AND entity_id=$2 ORDER BY created_at ASC", entityType, entityID)

	if len(entries) == 0 {
		respondJSON(w, http.StatusOK, map[string]interface{}{
			"verified": true, "entries": 0, "message": "no entries to verify",
		})
		return
	}

	invalid := false
	for i := 1; i < len(entries); i++ {
		prev := entries[i-1]
		curr := entries[i]
		if curr.PrevChecksum == nil || *curr.PrevChecksum != prev.Checksum {
			invalid = true
			break
		}
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"verified": !invalid,
		"entries":  len(entries),
		"message":  fmt.Sprintf("audit chain %s — %d entries verified", map[bool]string{true: "INTACT", false: "BROKEN"}[!invalid], len(entries)),
	})
}

// ---- Tax Invoices ----

func (h *AuditHandler) ListTaxInvoices(w http.ResponseWriter, r *http.Request) {
	var items []TaxInvoice
	h.db.Select(&items, "SELECT * FROM tax_invoices ORDER BY created_at DESC")
	respondJSON(w, http.StatusOK, items)
}

func (h *AuditHandler) CreateTaxInvoice(w http.ResponseWriter, r *http.Request) {
	var input TaxInvoice
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}
	var item TaxInvoice
	err := h.db.Get(&item, `INSERT INTO tax_invoices
		(project_id, contract_id, invoice_type, invoice_number, invoice_date, counterparty,
		 counterparty_tax_id, gross_amount, net_amount, tax_amount, tax_rate, tax_code,
		 currency, status, due_date, paid_date, is_reverse_charge, fiscal_period, document_ref, notes)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20)
		RETURNING *`,
		input.ProjectID, nullableString(input.ContractID), input.InvoiceType, input.InvoiceNumber,
		input.InvoiceDate, input.Counterparty, nullableString(input.CounterpartyTaxID),
		input.GrossAmount, input.NetAmount, input.TaxAmount, input.TaxRate,
		nullableString(input.TaxCode), input.Currency, input.Status,
		nullableString(input.DueDate), nullableString(input.PaidDate), input.IsReverseCharge,
		nullableString(input.FiscalPeriod), nullableString(input.DocumentRef), nullableString(input.Notes))
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("insert failed: %v", err))
		return
	}
	respondJSON(w, http.StatusCreated, item)
}

func (h *AuditHandler) GetTaxInvoice(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var item TaxInvoice
	if err := h.db.Get(&item, "SELECT * FROM tax_invoices WHERE id=$1", id); err != nil {
		respondError(w, http.StatusNotFound, "invoice not found")
		return
	}
	respondJSON(w, http.StatusOK, item)
}

func (h *AuditHandler) UpdateTaxInvoice(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input TaxInvoice
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}
	var item TaxInvoice
	err := h.db.Get(&item, `UPDATE tax_invoices SET
		invoice_type=$1, invoice_number=$2, invoice_date=$3, counterparty=$4,
		counterparty_tax_id=$5, gross_amount=$6, net_amount=$7, tax_amount=$8, tax_rate=$9,
		tax_code=$10, currency=$11, status=$12, due_date=$13, paid_date=$14,
		is_reverse_charge=$15, fiscal_period=$16, document_ref=$17, notes=$18, updated_at=NOW()
		WHERE id=$19 RETURNING *`,
		input.InvoiceType, input.InvoiceNumber, input.InvoiceDate, input.Counterparty,
		nullableString(input.CounterpartyTaxID), input.GrossAmount, input.NetAmount,
		input.TaxAmount, input.TaxRate, nullableString(input.TaxCode), input.Currency,
		input.Status, nullableString(input.DueDate), nullableString(input.PaidDate),
		input.IsReverseCharge, nullableString(input.FiscalPeriod), nullableString(input.DocumentRef),
		nullableString(input.Notes), id)
	if err != nil {
		respondError(w, http.StatusNotFound, "invoice not found")
		return
	}
	respondJSON(w, http.StatusOK, item)
}

func (h *AuditHandler) DeleteTaxInvoice(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	h.db.Exec("DELETE FROM tax_invoices WHERE id=$1", id)
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuditHandler) ListTaxInvoicesByProject(w http.ResponseWriter, r *http.Request) {
	pid := chi.URLParam(r, "projectId")
	var items []TaxInvoice
	h.db.Select(&items, "SELECT * FROM tax_invoices WHERE project_id=$1 ORDER BY invoice_date DESC", pid)
	respondJSON(w, http.StatusOK, items)
}

func (h *AuditHandler) ListTaxInvoicesByFiscal(w http.ResponseWriter, r *http.Request) {
	period := chi.URLParam(r, "period")
	var items []TaxInvoice
	h.db.Select(&items, "SELECT * FROM tax_invoices WHERE fiscal_period=$1 ORDER BY invoice_date DESC", period)
	respondJSON(w, http.StatusOK, items)
}

func (h *AuditHandler) ListTaxInvoicesByStatus(w http.ResponseWriter, r *http.Request) {
	status := chi.URLParam(r, "status")
	var items []TaxInvoice
	h.db.Select(&items, "SELECT * FROM tax_invoices WHERE status=$1 ORDER BY invoice_date DESC", status)
	respondJSON(w, http.StatusOK, items)
}

// ---- Tax Returns ----

func (h *AuditHandler) ListTaxReturns(w http.ResponseWriter, r *http.Request) {
	type TaxReturn struct {
		ID                string   `json:"id" db:"id"`
		ProjectID         *string  `json:"project_id" db:"project_id"`
		TaxRegID          *string  `json:"tax_registration_id" db:"tax_registration_id"`
		ReturnType        string   `json:"return_type" db:"return_type"`
		FiscalPeriod      string   `json:"fiscal_period" db:"fiscal_period"`
		PeriodStart       string   `json:"period_start" db:"period_start"`
		PeriodEnd         string   `json:"period_end" db:"period_end"`
		TotalTaxable      float64  `json:"total_taxable_amount" db:"total_taxable_amount"`
		TotalTaxDue       float64  `json:"total_tax_due" db:"total_tax_due"`
		TotalTaxCredit    float64  `json:"total_tax_credit" db:"total_tax_credit"`
		NetTaxPayable     float64  `json:"net_tax_payable" db:"net_tax_payable"`
		Currency          string   `json:"currency" db:"currency"`
		FilingDate        *string  `json:"filing_date" db:"filing_date"`
		DueDate           *string  `json:"due_date" db:"due_date"`
		PaidDate          *string  `json:"paid_date" db:"paid_date"`
		Status            string   `json:"status" db:"status"`
		FiledBy           *string  `json:"filed_by" db:"filed_by"`
		AmendedReturnID   *string  `json:"amended_return_id" db:"amended_return_id"`
		Notes             *string  `json:"notes" db:"notes"`
		CreatedAt         string   `json:"created_at" db:"created_at"`
		UpdatedAt         string   `json:"updated_at" db:"updated_at"`
	}
	var items []TaxReturn
	h.db.Select(&items, "SELECT * FROM tax_returns ORDER BY created_at DESC")
	respondJSON(w, http.StatusOK, items)
}

func (h *AuditHandler) CreateTaxReturn(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID     *string `json:"project_id"`
		TaxRegID      *string `json:"tax_registration_id"`
		ReturnType    string  `json:"return_type"`
		FiscalPeriod  string  `json:"fiscal_period"`
		PeriodStart   string  `json:"period_start"`
		PeriodEnd     string  `json:"period_end"`
		TotalTaxable  float64 `json:"total_taxable_amount"`
		TotalTaxDue   float64 `json:"total_tax_due"`
		TotalTaxCredit float64 `json:"total_tax_credit"`
		NetTaxPayable float64 `json:"net_tax_payable"`
		Currency      string  `json:"currency"`
		DueDate       *string `json:"due_date"`
		Status        string  `json:"status"`
		FiledBy       *string `json:"filed_by"`
		Notes         *string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}
	_, err := h.db.Exec(`INSERT INTO tax_returns
		(project_id, tax_registration_id, return_type, fiscal_period, period_start, period_end,
		 total_taxable_amount, total_tax_due, total_tax_credit, net_tax_payable,
		 currency, due_date, status, filed_by, notes)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
		nullableString(input.ProjectID), nullableString(input.TaxRegID), input.ReturnType,
		input.FiscalPeriod, input.PeriodStart, input.PeriodEnd,
		input.TotalTaxable, input.TotalTaxDue, input.TotalTaxCredit, input.NetTaxPayable,
		input.Currency, nullableString(input.DueDate), input.Status,
		nullableString(input.FiledBy), nullableString(input.Notes))
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("insert failed: %v", err))
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "created"})
}

func (h *AuditHandler) GetTaxReturn(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var item map[string]interface{}
	if err := h.db.Get(&item, "SELECT * FROM tax_returns WHERE id=$1", id); err != nil {
		respondError(w, http.StatusNotFound, "tax return not found")
		return
	}
	respondJSON(w, http.StatusOK, item)
}

func (h *AuditHandler) UpdateTaxReturn(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Status    string  `json:"status"`
		FilingDate *string `json:"filing_date"`
		PaidDate  *string `json:"paid_date"`
		FiledBy   *string `json:"filed_by"`
		Notes     *string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}
	_, err := h.db.Exec(`UPDATE tax_returns SET
		status=$1, filing_date=$2, paid_date=$3, filed_by=$4, notes=$5, updated_at=NOW()
		WHERE id=$6`,
		input.Status, nullableString(input.FilingDate), nullableString(input.PaidDate),
		nullableString(input.FiledBy), nullableString(input.Notes), id)
	if err != nil {
		respondError(w, http.StatusNotFound, "tax return not found")
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

func (h *AuditHandler) ListTaxReturnsByProject(w http.ResponseWriter, r *http.Request) {
	pid := chi.URLParam(r, "projectId")
	var items []map[string]interface{}
	h.db.Select(&items, "SELECT * FROM tax_returns WHERE project_id=$1 ORDER BY fiscal_period DESC", pid)
	respondJSON(w, http.StatusOK, items)
}

// ---- Stakeholders ----

func (h *AuditHandler) ListStakeholders(w http.ResponseWriter, r *http.Request) {
	var items []map[string]interface{}
	h.db.Select(&items, "SELECT * FROM stakeholders ORDER BY created_at DESC")
	respondJSON(w, http.StatusOK, items)
}

func (h *AuditHandler) CreateStakeholder(w http.ResponseWriter, r *http.Request) {
	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}
	_, err := h.db.NamedExec(`INSERT INTO stakeholders 
		(project_id, stakeholder_type, name, organization, contact_person, email, phone,
		 interest_level, influence_level, engagement_strategy, expectations, concerns,
		 communication_freq, status, notes)
		VALUES (:project_id, :stakeholder_type, :name, :organization, :contact_person, :email, :phone,
		 :interest_level, :influence_level, :engagement_strategy, :expectations, :concerns,
		 :communication_freq, :status, :notes)`, input)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("insert failed: %v", err))
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(input)
}

func (h *AuditHandler) GetStakeholder(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var item map[string]interface{}
	if err := h.db.Get(&item, "SELECT * FROM stakeholders WHERE id=$1", id); err != nil {
		respondError(w, http.StatusNotFound, "stakeholder not found")
		return
	}
	respondJSON(w, http.StatusOK, item)
}

func (h *AuditHandler) UpdateStakeholder(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}
	input["id"] = id
	_, err := h.db.NamedExec(`UPDATE stakeholders SET
		stakeholder_type=:stakeholder_type, name=:name, organization=:organization,
		contact_person=:contact_person, email=:email, phone=:phone,
		interest_level=:interest_level, influence_level=:influence_level,
		engagement_strategy=:engagement_strategy, expectations=:expectations,
		concerns=:concerns, communication_freq=:communication_freq, status=:status,
		notes=:notes, updated_at=NOW()
		WHERE id=:id`, input)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("update failed: %v", err))
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(input)
}

func (h *AuditHandler) DeleteStakeholder(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	h.db.Exec("DELETE FROM stakeholders WHERE id=$1", id)
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuditHandler) ListStakeholdersByProject(w http.ResponseWriter, r *http.Request) {
	pid := chi.URLParam(r, "projectId")
	var items []map[string]interface{}
	h.db.Select(&items, "SELECT * FROM stakeholders WHERE project_id=$1 ORDER BY name", pid)
	respondJSON(w, http.StatusOK, items)
}

func (h *AuditHandler) ListStakeholdersByType(w http.ResponseWriter, r *http.Request) {
	st := chi.URLParam(r, "stakeholderType")
	var items []map[string]interface{}
	h.db.Select(&items, "SELECT * FROM stakeholders WHERE stakeholder_type=$1 ORDER BY name", st)
	respondJSON(w, http.StatusOK, items)
}

// ---- Stakeholder Communications ----

func (h *AuditHandler) ListStakeholderComms(w http.ResponseWriter, r *http.Request) {
	var items []map[string]interface{}
	h.db.Select(&items, "SELECT * FROM stakeholder_communications ORDER BY communication_date DESC LIMIT 100")
	respondJSON(w, http.StatusOK, items)
}

func (h *AuditHandler) CreateStakeholderComm(w http.ResponseWriter, r *http.Request) {
	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}
	_, err := h.db.NamedExec(`INSERT INTO stakeholder_communications
		(stakeholder_id, project_id, communication_type, subject, summary, outcome,
		 action_items, communication_date, duration_minutes, conducted_by, participants,
		 document_ref, follow_up_date, follow_up_status, satisfaction_score)
		VALUES (:stakeholder_id, :project_id, :communication_type, :subject, :summary, :outcome,
		 :action_items, :communication_date, :duration_minutes, :conducted_by, :participants,
		 :document_ref, :follow_up_date, :follow_up_status, :satisfaction_score)`, input)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("insert failed: %v", err))
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(input)
}

func (h *AuditHandler) GetStakeholderComm(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var item map[string]interface{}
	if err := h.db.Get(&item, "SELECT * FROM stakeholder_communications WHERE id=$1", id); err != nil {
		respondError(w, http.StatusNotFound, "communication not found")
		return
	}
	respondJSON(w, http.StatusOK, item)
}

func (h *AuditHandler) ListCommsByStakeholder(w http.ResponseWriter, r *http.Request) {
	sid := chi.URLParam(r, "stakeholderId")
	var items []map[string]interface{}
	h.db.Select(&items, "SELECT * FROM stakeholder_communications WHERE stakeholder_id=$1 ORDER BY communication_date DESC", sid)
	respondJSON(w, http.StatusOK, items)
}

func (h *AuditHandler) ListPendingFollowUps(w http.ResponseWriter, r *http.Request) {
	var items []map[string]interface{}
	h.db.Select(&items, `SELECT sc.*, s.name as stakeholder_name 
		FROM stakeholder_communications sc
		LEFT JOIN stakeholders s ON s.id = sc.stakeholder_id
		WHERE sc.follow_up_status IN ('pending','overdue')
		ORDER BY sc.follow_up_date ASC`)
	respondJSON(w, http.StatusOK, items)
}

// ---- ESG Metrics ----

func (h *AuditHandler) ListESGMetrics(w http.ResponseWriter, r *http.Request) {
	var items []map[string]interface{}
	h.db.Select(&items, "SELECT * FROM esg_metrics ORDER BY created_at DESC")
	respondJSON(w, http.StatusOK, items)
}

func (h *AuditHandler) CreateESGMetric(w http.ResponseWriter, r *http.Request) {
	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}
	_, err := h.db.NamedExec(`INSERT INTO esg_metrics
		(project_id, category, metric_code, metric_name, metric_description, unit,
		 target_value, current_value, measurement_date, reporting_period, data_source, status, notes)
		VALUES (:project_id, :category, :metric_code, :metric_name, :metric_description, :unit,
		 :target_value, :current_value, :measurement_date, :reporting_period, :data_source, :status, :notes)`, input)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("insert failed: %v", err))
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(input)
}

func (h *AuditHandler) GetESGMetric(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var item map[string]interface{}
	if err := h.db.Get(&item, "SELECT * FROM esg_metrics WHERE id=$1", id); err != nil {
		respondError(w, http.StatusNotFound, "esg metric not found")
		return
	}
	respondJSON(w, http.StatusOK, item)
}

func (h *AuditHandler) UpdateESGMetric(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}
	input["id"] = id
	_, err := h.db.NamedExec(`UPDATE esg_metrics SET
		category=:category, metric_code=:metric_code, metric_name=:metric_name,
		metric_description=:metric_description, unit=:unit, target_value=:target_value,
		current_value=:current_value, measurement_date=:measurement_date,
		reporting_period=:reporting_period, data_source=:data_source, status=:status,
		notes=:notes, updated_at=NOW()
		WHERE id=:id`, input)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("update failed: %v", err))
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(input)
}

func (h *AuditHandler) ListESGMetricsByProject(w http.ResponseWriter, r *http.Request) {
	pid := chi.URLParam(r, "projectId")
	var items []map[string]interface{}
	h.db.Select(&items, "SELECT * FROM esg_metrics WHERE project_id=$1 ORDER BY category, metric_code", pid)
	respondJSON(w, http.StatusOK, items)
}

func (h *AuditHandler) ListESGMetricsByCategory(w http.ResponseWriter, r *http.Request) {
	cat := chi.URLParam(r, "category")
	var items []map[string]interface{}
	h.db.Select(&items, "SELECT * FROM esg_metrics WHERE category=$1 ORDER BY metric_code", cat)
	respondJSON(w, http.StatusOK, items)
}

// ---- Carbon Footprint ----

func (h *AuditHandler) ListCarbon(w http.ResponseWriter, r *http.Request) {
	var items []map[string]interface{}
	h.db.Select(&items, "SELECT * FROM carbon_footprint ORDER BY created_at DESC")
	respondJSON(w, http.StatusOK, items)
}

func (h *AuditHandler) CreateCarbon(w http.ResponseWriter, r *http.Request) {
	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}
	_, err := h.db.NamedExec(`INSERT INTO carbon_footprint
		(project_id, scope, category, source_description, co2_amount, co2_unit,
		 activity_data, emission_factor, emission_factor_source, reporting_period, status, notes)
		VALUES (:project_id, :scope, :category, :source_description, :co2_amount, :co2_unit,
		 :activity_data, :emission_factor, :emission_factor_source, :reporting_period, :status, :notes)`, input)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("insert failed: %v", err))
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(input)
}

func (h *AuditHandler) GetCarbon(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var item map[string]interface{}
	if err := h.db.Get(&item, "SELECT * FROM carbon_footprint WHERE id=$1", id); err != nil {
		respondError(w, http.StatusNotFound, "carbon entry not found")
		return
	}
	respondJSON(w, http.StatusOK, item)
}

func (h *AuditHandler) UpdateCarbon(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}
	input["id"] = id
	_, err := h.db.NamedExec(`UPDATE carbon_footprint SET
		scope=:scope, category=:category, source_description=:source_description,
		co2_amount=:co2_amount, co2_unit=:co2_unit, activity_data=:activity_data,
		emission_factor=:emission_factor, emission_factor_source=:emission_factor_source,
		reporting_period=:reporting_period, status=:status, notes=:notes, updated_at=NOW()
		WHERE id=:id`, input)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("update failed: %v", err))
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(input)
}

func (h *AuditHandler) ListCarbonByProject(w http.ResponseWriter, r *http.Request) {
	pid := chi.URLParam(r, "projectId")
	var items []map[string]interface{}
	h.db.Select(&items, "SELECT * FROM carbon_footprint WHERE project_id=$1 ORDER BY scope, category", pid)
	respondJSON(w, http.StatusOK, items)
}

func (h *AuditHandler) ListCarbonByScope(w http.ResponseWriter, r *http.Request) {
	scope := chi.URLParam(r, "scope")
	var items []map[string]interface{}
	h.db.Select(&items, "SELECT * FROM carbon_footprint WHERE scope=$1::INTEGER ORDER BY co2_amount DESC", scope)
	respondJSON(w, http.StatusOK, items)
}

func init() {
	_ = time.Now
	_ = log.LstdFlags
}