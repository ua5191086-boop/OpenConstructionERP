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

		// Variation Orders (Addendums)
		r.Get("/{contractId}/vo", h.ListVOs)
		r.Post("/{contractId}/vo", h.CreateVO)
		r.Get("/{contractId}/vo/{voId}", h.GetVO)
		r.Put("/{contractId}/vo/{voId}", h.UpdateVO)
		r.Delete("/{contractId}/vo/{voId}", h.DeleteVO)

		// Work Acceptances (КС-2 / КС-3)
		r.Get("/{contractId}/acceptances", h.ListAcceptances)
		r.Post("/{contractId}/acceptances", h.CreateAcceptance)
		r.Get("/{contractId}/acceptances/{accId}", h.GetAcceptance)
		r.Put("/{contractId}/acceptances/{accId}", h.UpdateAcceptance)
		r.Delete("/{contractId}/acceptances/{accId}", h.DeleteAcceptance)
		r.Get("/{contractId}/acceptances/{accId}/items", h.ListAcceptanceItems)
		r.Post("/{contractId}/acceptances/{accId}/items", h.CreateAcceptanceItem)

		// Claims
		r.Get("/{contractId}/claims", h.ListClaims)
		r.Post("/{contractId}/claims", h.CreateClaim)
		r.Get("/{contractId}/claims/{claimId}", h.GetClaim)
		r.Put("/{contractId}/claims/{claimId}", h.UpdateClaim)
		r.Delete("/{contractId}/claims/{claimId}", h.DeleteClaim)

		// Contract Summary
		r.Get("/{contractId}/summary", h.GetContractSummary)

		// IPC Integration — returns IPC amounts per acceptance for finance
		r.Get("/{contractId}/ipc", h.ListIPC)
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

// =============================================================================
// Variation Orders (Addendums / Дополнительные соглашения — VO)
// =============================================================================

func (h *ContractsHandler) ListVOs(w http.ResponseWriter, r *http.Request) {
	contractID := chi.URLParam(r, "contractId")
	rows, err := h.db.Query(`SELECT id, contract_id, addendum_number, name, description, addendum_type, amount_change, days_change, new_end_date, status, signed_at, document_path, notes, created_at FROM contract_addendums WHERE contract_id = $1 ORDER BY addendum_number`, contractID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, cid, name, desc, atype, st, docPath, notes string
		var num int
		var amountChange float64
		var daysChange int
		var newEndDate, signedAt, createdAt sql.NullString
		if err := rows.Scan(&id, &cid, &num, &name, &desc, &atype, &amountChange, &daysChange, &newEndDate, &st, &signedAt, &docPath, &notes, &createdAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		item := map[string]interface{}{
			"id": id, "contract_id": cid, "addendum_number": num, "name": name,
			"description": desc, "addendum_type": atype, "amount_change": amountChange,
			"days_change": daysChange, "status": st, "document_path": docPath, "notes": notes,
		}
		if newEndDate.Valid { item["new_end_date"] = newEndDate.String }
		if signedAt.Valid { item["signed_at"] = signedAt.String }
		if createdAt.Valid { item["created_at"] = createdAt.String }
		items = append(items, item)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *ContractsHandler) CreateVO(w http.ResponseWriter, r *http.Request) {
	contractID := chi.URLParam(r, "contractId")
	var input struct {
		Name        string   `json:"name"`
		Description *string  `json:"description"`
		AddendumType string  `json:"addendum_type"`
		AmountChange float64 `json:"amount_change"`
		DaysChange   int     `json:"days_change"`
		NewEndDate   *string `json:"new_end_date"`
		DocumentPath *string `json:"document_path"`
		Notes        *string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO contract_addendums (id, contract_id, addendum_number, name, description, addendum_type, amount_change, days_change, new_end_date, status, document_path, notes, created_at) VALUES ($1,$2,(SELECT COALESCE(MAX(addendum_number),0)+1 FROM contract_addendums WHERE contract_id=$2),$3,$4,$5,$6,$7,$8,'draft',$9,$10,$11)`,
		id, contractID, input.Name, input.Description, input.AddendumType, input.AmountChange, input.DaysChange, input.NewEndDate, input.DocumentPath, input.Notes, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *ContractsHandler) GetVO(w http.ResponseWriter, r *http.Request) {
	voID := chi.URLParam(r, "voId")
	var id, cid, name, st string
	var num int
	err := h.db.QueryRow(`SELECT id, contract_id, addendum_number, name, status FROM contract_addendums WHERE id = $1`, voID).Scan(&id, &cid, &num, &name, &st)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "VO not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "contract_id": cid, "addendum_number": num, "name": name, "status": st})
}

func (h *ContractsHandler) UpdateVO(w http.ResponseWriter, r *http.Request) {
	voID := chi.URLParam(r, "voId")
	var input struct {
		Status *string `json:"status"`
		Notes  *string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	_, err := h.db.Exec(`UPDATE contract_addendums SET status=COALESCE($1,status), notes=COALESCE($2,notes), signed_at=CASE WHEN $1='signed' THEN NOW() ELSE signed_at END WHERE id=$3`,
		input.Status, input.Notes, voID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *ContractsHandler) DeleteVO(w http.ResponseWriter, r *http.Request) {
	voID := chi.URLParam(r, "voId")
	_, err := h.db.Exec(`DELETE FROM contract_addendums WHERE id = $1`, voID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// Work Acceptances (КС-2 / КС-3 / IPC)
// =============================================================================

func (h *ContractsHandler) ListAcceptances(w http.ResponseWriter, r *http.Request) {
	contractID := chi.URLParam(r, "contractId")
	rows, err := h.db.Query(`SELECT id, contract_id, milestone_id, acceptance_number, acceptance_date, period_from, period_to, amount, currency, status, approved_by, approved_at, paid_at, payment_ref, notes, created_at FROM contract_work_acceptances WHERE contract_id = $1 ORDER BY acceptance_date DESC`, contractID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, cid, mid, accNum, currency, st, approvedBy, paymentRef, notes string
		var accDate, periodFrom, periodTo sql.NullString
		var amount float64
		var approvedAt, paidAt, createdAt sql.NullString
		if err := rows.Scan(&id, &cid, &mid, &accNum, &accDate, &periodFrom, &periodTo, &amount, &currency, &st, &approvedBy, &approvedAt, &paidAt, &paymentRef, &notes, &createdAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		item := map[string]interface{}{
			"id": id, "contract_id": cid, "milestone_id": mid, "acceptance_number": accNum,
			"acceptance_date": accDate, "amount": amount, "currency": currency,
			"status": st, "approved_by": approvedBy, "payment_ref": paymentRef, "notes": notes,
		}
		if periodFrom.Valid { item["period_from"] = periodFrom.String }
		if periodTo.Valid { item["period_to"] = periodTo.String }
		if approvedAt.Valid { item["approved_at"] = approvedAt.String }
		if paidAt.Valid { item["paid_at"] = paidAt.String }
		if createdAt.Valid { item["created_at"] = createdAt.String }
		items = append(items, item)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *ContractsHandler) CreateAcceptance(w http.ResponseWriter, r *http.Request) {
	contractID := chi.URLParam(r, "contractId")
	var input struct {
		AcceptanceNumber string  `json:"acceptance_number"`
		AcceptanceDate   string  `json:"acceptance_date"`
		PeriodFrom       *string `json:"period_from"`
		PeriodTo         *string `json:"period_to"`
		Amount           float64 `json:"amount"`
		Currency         string  `json:"currency"`
		Notes            *string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO contract_work_acceptances (id, contract_id, acceptance_number, acceptance_date, period_from, period_to, amount, currency, status, notes, created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,'draft',$9,$10)`,
		id, contractID, input.AcceptanceNumber, input.AcceptanceDate, input.PeriodFrom, input.PeriodTo, input.Amount, input.Currency, input.Notes, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *ContractsHandler) GetAcceptance(w http.ResponseWriter, r *http.Request) {
	accID := chi.URLParam(r, "accId")
	var id, cid, accNum, st string
	var amount float64
	err := h.db.QueryRow(`SELECT id, contract_id, acceptance_number, amount, status FROM contract_work_acceptances WHERE id = $1`, accID).Scan(&id, &cid, &accNum, &amount, &st)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "acceptance not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "contract_id": cid, "acceptance_number": accNum, "amount": amount, "status": st})
}

func (h *ContractsHandler) UpdateAcceptance(w http.ResponseWriter, r *http.Request) {
	accID := chi.URLParam(r, "accId")
	var input struct {
		Status      *string  `json:"status"`
		ApprovedBy  *string  `json:"approved_by"`
		PaymentRef  *string  `json:"payment_ref"`
		Notes       *string  `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	_, err := h.db.Exec(`UPDATE contract_work_acceptances SET status=COALESCE($1,status), approved_by=COALESCE($2,approved_by), payment_ref=COALESCE($3,payment_ref), notes=COALESCE($4,notes), approved_at=CASE WHEN $1='approved' THEN NOW() ELSE approved_at END, paid_at=CASE WHEN $1='paid' THEN NOW() ELSE paid_at END WHERE id=$5`,
		input.Status, input.ApprovedBy, input.PaymentRef, input.Notes, accID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *ContractsHandler) DeleteAcceptance(w http.ResponseWriter, r *http.Request) {
	accID := chi.URLParam(r, "accId")
	_, err := h.db.Exec(`DELETE FROM contract_work_acceptances WHERE id = $1`, accID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// --- Acceptance Items (привязка к BOQ) ---

func (h *ContractsHandler) ListAcceptanceItems(w http.ResponseWriter, r *http.Request) {
	accID := chi.URLParam(r, "accId")
	rows, err := h.db.Query(`SELECT id, acceptance_id, boq_item_id, item_code, description, unit, contract_quantity, prev_quantity, current_quantity, total_quantity, unit_price, current_amount, total_amount, sort_order, created_at FROM contract_acceptance_items WHERE acceptance_id = $1 ORDER BY sort_order`, accID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, accID2, boqID, itemCode, desc, unit string
		var cq, pq, curq, tq, up, ca, ta float64
		var so int
		var createdAt time.Time
		if err := rows.Scan(&id, &accID2, &boqID, &itemCode, &desc, &unit, &cq, &pq, &curq, &tq, &up, &ca, &ta, &so, &createdAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, map[string]interface{}{
			"id": id, "acceptance_id": accID2, "boq_item_id": boqID, "item_code": itemCode,
			"description": desc, "unit": unit, "contract_quantity": cq, "prev_quantity": pq,
			"current_quantity": curq, "total_quantity": tq, "unit_price": up,
			"current_amount": ca, "total_amount": ta, "sort_order": so,
			"created_at": createdAt,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *ContractsHandler) CreateAcceptanceItem(w http.ResponseWriter, r *http.Request) {
	accID := chi.URLParam(r, "accId")
	var input struct {
		BOQItemID        *string  `json:"boq_item_id"`
		ItemCode         string   `json:"item_code"`
		Description      *string  `json:"description"`
		Unit             string   `json:"unit"`
		ContractQuantity float64  `json:"contract_quantity"`
		PrevQuantity     float64  `json:"prev_quantity"`
		CurrentQuantity  float64  `json:"current_quantity"`
		UnitPrice        float64  `json:"unit_price"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	totalQuantity := input.PrevQuantity + input.CurrentQuantity
	currentAmount := input.CurrentQuantity * input.UnitPrice
	totalAmount := totalQuantity * input.UnitPrice
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO contract_acceptance_items (id, acceptance_id, boq_item_id, item_code, description, unit, contract_quantity, prev_quantity, current_quantity, total_quantity, unit_price, current_amount, total_amount, created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		id, accID, input.BOQItemID, input.ItemCode, input.Description, input.Unit, input.ContractQuantity, input.PrevQuantity, input.CurrentQuantity, totalQuantity, input.UnitPrice, currentAmount, totalAmount, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

// =============================================================================
// Claims (Претензии)
// =============================================================================

func (h *ContractsHandler) ListClaims(w http.ResponseWriter, r *http.Request) {
	contractID := chi.URLParam(r, "contractId")
	rows, err := h.db.Query(`SELECT id, contract_id, claim_number, claim_type, description, amount_claimed, amount_approved, currency, status, submitted_by, submitted_at, resolved_at, resolution, document_path, created_at FROM contract_claims WHERE contract_id = $1 ORDER BY claim_number`, contractID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, cid, cnum, ctype, desc, currency, st, submittedBy, resolution, docPath string
		var amountClaimed, amountApproved float64
		var submittedAt, resolvedAt, createdAt sql.NullString
		if err := rows.Scan(&id, &cid, &cnum, &ctype, &desc, &amountClaimed, &amountApproved, &currency, &st, &submittedBy, &submittedAt, &resolvedAt, &resolution, &docPath, &createdAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		item := map[string]interface{}{
			"id": id, "contract_id": cid, "claim_number": cnum, "claim_type": ctype,
			"description": desc, "amount_claimed": amountClaimed, "amount_approved": amountApproved,
			"currency": currency, "status": st, "submitted_by": submittedBy,
			"resolution": resolution, "document_path": docPath,
		}
		if submittedAt.Valid { item["submitted_at"] = submittedAt.String }
		if resolvedAt.Valid { item["resolved_at"] = resolvedAt.String }
		if createdAt.Valid { item["created_at"] = createdAt.String }
		items = append(items, item)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *ContractsHandler) CreateClaim(w http.ResponseWriter, r *http.Request) {
	contractID := chi.URLParam(r, "contractId")
	var input struct {
		ClaimType     string   `json:"claim_type"`
		Description   string   `json:"description"`
		AmountClaimed float64  `json:"amount_claimed"`
		Currency      string   `json:"currency"`
		SubmittedBy   *string  `json:"submitted_by"`
		DocumentPath  *string  `json:"document_path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO contract_claims (id, contract_id, claim_number, claim_type, description, amount_claimed, currency, status, submitted_by, submitted_at, document_path, created_at) VALUES ($1,$2,(SELECT COALESCE(MAX(claim_number),0)+1 FROM contract_claims WHERE contract_id=$2),$3,$4,$5,$6,'submitted',$7,$8,$9,$10)`,
		id, contractID, input.ClaimType, input.Description, input.AmountClaimed, input.Currency, input.SubmittedBy, now, input.DocumentPath, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *ContractsHandler) GetClaim(w http.ResponseWriter, r *http.Request) {
	claimID := chi.URLParam(r, "claimId")
	var id, cid, ctype, st string
	var amount float64
	err := h.db.QueryRow(`SELECT id, contract_id, claim_type, amount_claimed, status FROM contract_claims WHERE id = $1`, claimID).Scan(&id, &cid, &ctype, &amount, &st)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "claim not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "contract_id": cid, "claim_type": ctype, "amount_claimed": amount, "status": st})
}

func (h *ContractsHandler) UpdateClaim(w http.ResponseWriter, r *http.Request) {
	claimID := chi.URLParam(r, "claimId")
	var input struct {
		Status          *string  `json:"status"`
		AmountApproved  *float64 `json:"amount_approved"`
		Resolution      *string  `json:"resolution"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	_, err := h.db.Exec(`UPDATE contract_claims SET status=COALESCE($1,status), amount_approved=COALESCE($2,amount_approved), resolution=COALESCE($3,resolution), resolved_at=CASE WHEN $1 IN ('approved','rejected','withdrawn') THEN NOW() ELSE resolved_at END WHERE id=$4`,
		input.Status, input.AmountApproved, input.Resolution, claimID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *ContractsHandler) DeleteClaim(w http.ResponseWriter, r *http.Request) {
	claimID := chi.URLParam(r, "claimId")
	_, err := h.db.Exec(`DELETE FROM contract_claims WHERE id = $1`, claimID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// Contract Summary
// =============================================================================

func (h *ContractsHandler) GetContractSummary(w http.ResponseWriter, r *http.Request) {
	contractID := chi.URLParam(r, "contractId")

	// Contract details
	contractInfo := struct {
		ID, Code, Name, Status string
		Amount                  float64
	}{}
	err := h.db.QueryRow(`SELECT id, code, name, status, contract_amount FROM contracts WHERE id = $1`, contractID).Scan(&contractInfo.ID, &contractInfo.Code, &contractInfo.Name, &contractInfo.Status, &contractInfo.Amount)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "contract not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var voCount, voTotal float64
	h.db.QueryRow(`SELECT COUNT(*), COALESCE(SUM(amount_change),0) FROM contract_addendums WHERE contract_id = $1 AND status='signed'`, contractID).Scan(&voCount, &voTotal)

	var accCount, accApproved, accAmount sql.NullFloat64
	h.db.QueryRow(`SELECT COUNT(*), COUNT(*) FILTER (WHERE status='approved' OR status='paid'), COALESCE(SUM(amount),0) FROM contract_work_acceptances WHERE contract_id = $1`, contractID).Scan(&accCount, &accApproved, &accAmount)

	var claimCount, claimApproved float64
	h.db.QueryRow(`SELECT COUNT(*), COALESCE(SUM(amount_approved),0) FROM contract_claims WHERE contract_id = $1`, contractID).Scan(&claimCount, &claimApproved)

	totalVO := contractInfo.Amount + voTotal

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"contract_id":           contractInfo.ID,
		"code":                  contractInfo.Code,
		"name":                  contractInfo.Name,
		"status":                contractInfo.Status,
		"original_amount":       contractInfo.Amount,
		"vo_count":              voCount,
		"vo_total_change":       voTotal,
		"current_contract_value": totalVO,
		"acceptance_count":      accCount,
		"acceptance_approved":   accApproved,
		"accepted_total":        accAmount,
		"claim_count":           claimCount,
		"claim_approved_total":  claimApproved,
	})
}

// =============================================================================
// IPC — amounts per acceptance for Finance integration
// =============================================================================

func (h *ContractsHandler) ListIPC(w http.ResponseWriter, r *http.Request) {
	contractID := chi.URLParam(r, "contractId")
	rows, err := h.db.Query(`
		SELECT
			a.id AS acceptance_id,
			a.acceptance_number,
			a.acceptance_date,
			a.period_from,
			a.period_to,
			a.amount AS acceptance_amount,
			a.status,
			COALESCE(ai.item_count, 0) AS item_count,
			COALESCE(ai.total_quantity, 0) AS total_quantity
		FROM contract_work_acceptances a
		LEFT JOIN (
			SELECT acceptance_id,
				COUNT(*) AS item_count,
				SUM(current_quantity) AS total_quantity
			FROM contract_acceptance_items
			GROUP BY acceptance_id
		) ai ON ai.acceptance_id = a.id
		WHERE a.contract_id = $1
		ORDER BY a.acceptance_date DESC`, contractID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var accID, accNum, st string
		var accDate, periodFrom, periodTo sql.NullString
		var amount, itemCount, totalQty float64
		if err := rows.Scan(&accID, &accNum, &accDate, &periodFrom, &periodTo, &amount, &st, &itemCount, &totalQty); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		item := map[string]interface{}{
			"acceptance_id": accID, "acceptance_number": accNum, "acceptance_amount": amount,
			"status": st, "item_count": itemCount, "total_quantity": totalQty,
		}
		if accDate.Valid { item["acceptance_date"] = accDate.String }
		if periodFrom.Valid { item["period_from"] = periodFrom.String }
		if periodTo.Valid { item["period_to"] = periodTo.String }
		items = append(items, item)
	}
	respondJSON(w, http.StatusOK, items)
}
