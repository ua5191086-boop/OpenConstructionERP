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

// ContractsHandler handles Contract module endpoints
type ContractsHandler struct {
	db *sql.DB
}

func NewContractsHandler(db *sql.DB) *ContractsHandler {
	return &ContractsHandler{db: db}
}

func (h *ContractsHandler) RegisterRoutes(r chi.Router) {
	r.Route("/contracts", func(r chi.Router) {
		r.Get("/", h.ListContracts)
		r.Post("/", h.CreateContract)
		r.Get("/{id}", h.GetContract)
		r.Put("/{id}", h.UpdateContract)
		r.Delete("/{id}", h.DeleteContract)

		// Milestones
		r.Get("/{contractId}/milestones", h.ListMilestones)
		r.Post("/{contractId}/milestones", h.CreateMilestone)
		r.Get("/{contractId}/milestones/{milestoneId}", h.GetMilestone)
		r.Put("/{contractId}/milestones/{milestoneId}", h.UpdateMilestone)
		r.Delete("/{contractId}/milestones/{milestoneId}", h.DeleteMilestone)

		// Payments
		r.Get("/{contractId}/payments", h.ListPayments)
		r.Post("/{contractId}/payments", h.CreatePayment)
		r.Get("/{contractId}/payments/{paymentId}", h.GetPayment)
		r.Put("/{contractId}/payments/{paymentId}", h.UpdatePayment)
		r.Delete("/{contractId}/payments/{paymentId}", h.DeletePayment)
	})
}

// --- Contracts ---

func (h *ContractsHandler) ListContracts(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	projectID := r.URL.Query().Get("project_id")
	contractorID := r.URL.Query().Get("contractor_id")

	query := `SELECT id, code, name, description, contract_type, status, tender_id, lot_id, client_id, contractor_id, project_id, contract_amount, currency, advance_amount, advance_pct, signed_at, start_date, end_date, duration_days, performance_bond_amount, performance_bond_pct, warranty_period_days, retention_pct, retention_release_days, penalty_rate_daily, penalty_max_pct, liquidated_damages, funding_source, payment_terms, payment_terms_type, document_path, notes, created_by, created_at, updated_at FROM contracts WHERE 1=1`
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
	if contractorID != "" {
		query += ` AND contractor_id = $` + itoa(argIdx)
		args = append(args, contractorID)
		argIdx++
	}
	query += ` ORDER BY created_at DESC`

	rows, err := h.db.Query(query, args...)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	contracts := make([]models.Contract, 0)
	for rows.Next() {
		var c models.Contract
		if err := rows.Scan(&c.ID, &c.Code, &c.Name, &c.Description, &c.ContractType, &c.Status, &c.TenderID, &c.LotID, &c.ClientID, &c.ContractorID, &c.ProjectID, &c.ContractAmount, &c.Currency, &c.AdvanceAmount, &c.AdvancePct, &c.SignedAt, &c.StartDate, &c.EndDate, &c.DurationDays, &c.PerformanceBondAmount, &c.PerformanceBondPct, &c.WarrantyPeriodDays, &c.RetentionPct, &c.RetentionReleaseDays, &c.PenaltyRateDaily, &c.PenaltyMaxPct, &c.LiquidatedDamages, &c.FundingSource, &c.PaymentTerms, &c.PaymentTermsType, &c.DocumentPath, &c.Notes, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		contracts = append(contracts, c)
	}
	respondJSON(w, http.StatusOK, contracts)
}

func (h *ContractsHandler) CreateContract(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Code                string   `json:"code"`
		Name                string   `json:"name"`
		Description         *string  `json:"description"`
		ContractType        string   `json:"contract_type"`
		Status              string   `json:"status"`
		TenderID            *string  `json:"tender_id"`
		LotID               *string  `json:"lot_id"`
		ClientID            string   `json:"client_id"`
		ContractorID        string   `json:"contractor_id"`
		ProjectID           *string  `json:"project_id"`
		ContractAmount      float64  `json:"contract_amount"`
		Currency            string   `json:"currency"`
		AdvanceAmount       *float64 `json:"advance_amount"`
		AdvancePct          *float64 `json:"advance_pct"`
		SignedAt            *string  `json:"signed_at"`
		StartDate           *string  `json:"start_date"`
		EndDate             *string  `json:"end_date"`
		DurationDays        *int     `json:"duration_days"`
		PerformanceBondAmount *float64 `json:"performance_bond_amount"`
		PerformanceBondPct  *float64 `json:"performance_bond_pct"`
		WarrantyPeriodDays  *int     `json:"warranty_period_days"`
		RetentionPct         *float64 `json:"retention_pct"`
		RetentionReleaseDays *int    `json:"retention_release_days"`
		PenaltyRateDaily     *float64 `json:"penalty_rate_daily"`
		PenaltyMaxPct        *float64 `json:"penalty_max_pct"`
		LiquidatedDamages    *float64 `json:"liquidated_damages"`
		FundingSource        *string  `json:"funding_source"`
		PaymentTerms         *string  `json:"payment_terms"`
		PaymentTermsType     *string  `json:"payment_terms_type"`
		DocumentPath         *string  `json:"document_path"`
		Notes                *string  `json:"notes"`
		CreatedBy            *string  `json:"created_by"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO contracts (id, code, name, description, contract_type, status, tender_id, lot_id, client_id, contractor_id, project_id, contract_amount, currency, advance_amount, advance_pct, signed_at, start_date, end_date, duration_days, performance_bond_amount, performance_bond_pct, warranty_period_days, retention_pct, retention_release_days, penalty_rate_daily, penalty_max_pct, liquidated_damages, funding_source, payment_terms, payment_terms_type, document_path, notes, created_by, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26,$27,$28,$29,$30,$31,$32,$33,$34,$35)`,
		id, input.Code, input.Name, input.Description, input.ContractType, input.Status, input.TenderID, input.LotID, input.ClientID, input.ContractorID, input.ProjectID, input.ContractAmount, input.Currency, input.AdvanceAmount, input.AdvancePct, input.SignedAt, input.StartDate, input.EndDate, input.DurationDays, input.PerformanceBondAmount, input.PerformanceBondPct, input.WarrantyPeriodDays, input.RetentionPct, input.RetentionReleaseDays, input.PenaltyRateDaily, input.PenaltyMaxPct, input.LiquidatedDamages, input.FundingSource, input.PaymentTerms, input.PaymentTermsType, input.DocumentPath, input.Notes, input.CreatedBy, now, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *ContractsHandler) GetContract(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var c models.Contract
	err := h.db.QueryRow(`SELECT id, code, name, description, contract_type, status, tender_id, lot_id, client_id, contractor_id, project_id, contract_amount, currency, advance_amount, advance_pct, signed_at, start_date, end_date, duration_days, performance_bond_amount, performance_bond_pct, warranty_period_days, retention_pct, retention_release_days, penalty_rate_daily, penalty_max_pct, liquidated_damages, funding_source, payment_terms, payment_terms_type, document_path, notes, created_by, created_at, updated_at FROM contracts WHERE id = $1`, id).
		Scan(&c.ID, &c.Code, &c.Name, &c.Description, &c.ContractType, &c.Status, &c.TenderID, &c.LotID, &c.ClientID, &c.ContractorID, &c.ProjectID, &c.ContractAmount, &c.Currency, &c.AdvanceAmount, &c.AdvancePct, &c.SignedAt, &c.StartDate, &c.EndDate, &c.DurationDays, &c.PerformanceBondAmount, &c.PerformanceBondPct, &c.WarrantyPeriodDays, &c.RetentionPct, &c.RetentionReleaseDays, &c.PenaltyRateDaily, &c.PenaltyMaxPct, &c.LiquidatedDamages, &c.FundingSource, &c.PaymentTerms, &c.PaymentTermsType, &c.DocumentPath, &c.Notes, &c.CreatedBy, &c.CreatedAt, &c.UpdatedAt)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "contract not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, c)
}

func (h *ContractsHandler) UpdateContract(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Name          *string  `json:"name"`
		Description   *string  `json:"description"`
		Status        *string  `json:"status"`
		ContractAmount *float64 `json:"contract_amount"`
		EndDate       *string  `json:"end_date"`
		Notes         *string  `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	now := time.Now()
	_, err := h.db.Exec(`UPDATE contracts SET name=COALESCE($1,name), description=COALESCE($2,description), status=COALESCE($3,status), contract_amount=COALESCE($4,contract_amount), end_date=COALESCE($5,end_date), notes=COALESCE($6,notes), updated_at=$7 WHERE id=$8`,
		input.Name, input.Description, input.Status, input.ContractAmount, input.EndDate, input.Notes, now, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *ContractsHandler) DeleteContract(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM contracts WHERE id = $1`, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// --- Milestones ---

func (h *ContractsHandler) ListMilestones(w http.ResponseWriter, r *http.Request) {
	contractID := chi.URLParam(r, "contractId")
	rows, err := h.db.Query(`SELECT id, contract_id, milestone_number, name, description, milestone_type, planned_date, actual_date, amount, amount_pct, status, completion_pct, notes, created_at FROM contract_milestones WHERE contract_id = $1 ORDER BY milestone_number`, contractID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	milestones := make([]models.ContractMilestone, 0)
	for rows.Next() {
		var m models.ContractMilestone
		if err := rows.Scan(&m.ID, &m.ContractID, &m.MilestoneNumber, &m.Name, &m.Description, &m.MilestoneType, &m.PlannedDate, &m.ActualDate, &m.Amount, &m.AmountPct, &m.Status, &m.CompletionPct, &m.Notes, &m.CreatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		milestones = append(milestones, m)
	}
	respondJSON(w, http.StatusOK, milestones)
}

func (h *ContractsHandler) CreateMilestone(w http.ResponseWriter, r *http.Request) {
	contractID := chi.URLParam(r, "contractId")
	var input struct {
		MilestoneNumber int      `json:"milestone_number"`
		Name            string   `json:"name"`
		Description     *string  `json:"description"`
		MilestoneType   string   `json:"milestone_type"`
		PlannedDate     *string  `json:"planned_date"`
		ActualDate      *string  `json:"actual_date"`
		Amount          *float64 `json:"amount"`
		AmountPct       *float64 `json:"amount_pct"`
		Status          string   `json:"status"`
		CompletionPct   *float64 `json:"completion_pct"`
		Notes           *string  `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO contract_milestones (id, contract_id, milestone_number, name, description, milestone_type, planned_date, actual_date, amount, amount_pct, status, completion_pct, notes, created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		id, contractID, input.MilestoneNumber, input.Name, input.Description, input.MilestoneType, input.PlannedDate, input.ActualDate, input.Amount, input.AmountPct, input.Status, input.CompletionPct, input.Notes, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *ContractsHandler) GetMilestone(w http.ResponseWriter, r *http.Request) {
	milestoneID := chi.URLParam(r, "milestoneId")
	var m models.ContractMilestone
	err := h.db.QueryRow(`SELECT id, contract_id, milestone_number, name, description, milestone_type, planned_date, actual_date, amount, amount_pct, status, completion_pct, notes, created_at FROM contract_milestones WHERE id = $1`, milestoneID).
		Scan(&m.ID, &m.ContractID, &m.MilestoneNumber, &m.Name, &m.Description, &m.MilestoneType, &m.PlannedDate, &m.ActualDate, &m.Amount, &m.AmountPct, &m.Status, &m.CompletionPct, &m.Notes, &m.CreatedAt)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "milestone not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, m)
}

func (h *ContractsHandler) UpdateMilestone(w http.ResponseWriter, r *http.Request) {
	milestoneID := chi.URLParam(r, "milestoneId")
	var input struct {
		Name          *string  `json:"name"`
		ActualDate    *string  `json:"actual_date"`
		Status        *string  `json:"status"`
		CompletionPct *float64 `json:"completion_pct"`
		Notes         *string  `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	_, err := h.db.Exec(`UPDATE contract_milestones SET name=COALESCE($1,name), actual_date=COALESCE($2,actual_date), status=COALESCE($3,status), completion_pct=COALESCE($4,completion_pct), notes=COALESCE($5,notes) WHERE id=$6`,
		input.Name, input.ActualDate, input.Status, input.CompletionPct, input.Notes, milestoneID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *ContractsHandler) DeleteMilestone(w http.ResponseWriter, r *http.Request) {
	milestoneID := chi.URLParam(r, "milestoneId")
	_, err := h.db.Exec(`DELETE FROM contract_milestones WHERE id = $1`, milestoneID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// --- Payments ---

func (h *ContractsHandler) ListPayments(w http.ResponseWriter, r *http.Request) {
	contractID := chi.URLParam(r, "contractId")
	rows, err := h.db.Query(`SELECT id, contract_id, acceptance_id, milestone_id, payment_number, payment_date, amount, currency, payment_type, payment_method, status, bank_ref, notes, created_at FROM contract_payments WHERE contract_id = $1 ORDER BY payment_date DESC`, contractID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	payments := make([]models.ContractPayment, 0)
	for rows.Next() {
		var p models.ContractPayment
		if err := rows.Scan(&p.ID, &p.ContractID, &p.AcceptanceID, &p.MilestoneID, &p.PaymentNumber, &p.PaymentDate, &p.Amount, &p.Currency, &p.PaymentType, &p.PaymentMethod, &p.Status, &p.BankRef, &p.Notes, &p.CreatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		payments = append(payments, p)
	}
	respondJSON(w, http.StatusOK, payments)
}

func (h *ContractsHandler) CreatePayment(w http.ResponseWriter, r *http.Request) {
	contractID := chi.URLParam(r, "contractId")
	var input struct {
		AcceptanceID  *string `json:"acceptance_id"`
		MilestoneID   *string `json:"milestone_id"`
		PaymentNumber string  `json:"payment_number"`
		PaymentDate   string  `json:"payment_date"`
		Amount        float64 `json:"amount"`
		Currency      string  `json:"currency"`
		PaymentType   string  `json:"payment_type"`
		PaymentMethod *string `json:"payment_method"`
		Status        string  `json:"status"`
		BankRef       *string `json:"bank_ref"`
		Notes         *string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO contract_payments (id, contract_id, acceptance_id, milestone_id, payment_number, payment_date, amount, currency, payment_type, payment_method, status, bank_ref, notes, created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		id, contractID, input.AcceptanceID, input.MilestoneID, input.PaymentNumber, input.PaymentDate, input.Amount, input.Currency, input.PaymentType, input.PaymentMethod, input.Status, input.BankRef, input.Notes, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *ContractsHandler) GetPayment(w http.ResponseWriter, r *http.Request) {
	paymentID := chi.URLParam(r, "paymentId")
	var p models.ContractPayment
	err := h.db.QueryRow(`SELECT id, contract_id, acceptance_id, milestone_id, payment_number, payment_date, amount, currency, payment_type, payment_method, status, bank_ref, notes, created_at FROM contract_payments WHERE id = $1`, paymentID).
		Scan(&p.ID, &p.ContractID, &p.AcceptanceID, &p.MilestoneID, &p.PaymentNumber, &p.PaymentDate, &p.Amount, &p.Currency, &p.PaymentType, &p.PaymentMethod, &p.Status, &p.BankRef, &p.Notes, &p.CreatedAt)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "payment not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, p)
}

func (h *ContractsHandler) UpdatePayment(w http.ResponseWriter, r *http.Request) {
	paymentID := chi.URLParam(r, "paymentId")
	var input struct {
		Status  *string `json:"status"`
		BankRef *string `json:"bank_ref"`
		Notes   *string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	_, err := h.db.Exec(`UPDATE contract_payments SET status=COALESCE($1,status), bank_ref=COALESCE($2,bank_ref), notes=COALESCE($3,notes) WHERE id=$4`,
		input.Status, input.BankRef, input.Notes, paymentID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *ContractsHandler) DeletePayment(w http.ResponseWriter, r *http.Request) {
	paymentID := chi.URLParam(r, "paymentId")
	_, err := h.db.Exec(`DELETE FROM contract_payments WHERE id = $1`, paymentID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
