package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/openconstructionerp/oce/services/core/internal/models"
)

// TendersHandler handles Tender module endpoints
type TendersHandler struct {
	db *sql.DB
}

func NewTendersHandler(db *sql.DB) *TendersHandler {
	return &TendersHandler{db: db}
}

func (h *TendersHandler) RegisterRoutes(r chi.Router) {
	r.Route("/tenders", func(r chi.Router) {
		r.Get("/", h.ListTenders)
		r.Post("/", h.CreateTender)
		r.Get("/{id}", h.GetTender)
		r.Put("/{id}", h.UpdateTender)
		r.Delete("/{id}", h.DeleteTender)

		// Tender Lots
		r.Get("/{tenderId}/lots", h.ListLots)
		r.Post("/{tenderId}/lots", h.CreateLot)
		r.Get("/{tenderId}/lots/{lotId}", h.GetLot)
		r.Put("/{tenderId}/lots/{lotId}", h.UpdateLot)
		r.Delete("/{tenderId}/lots/{lotId}", h.DeleteLot)

		// Tender Bidders
		r.Get("/{tenderId}/bidders", h.ListBidders)
		r.Post("/{tenderId}/bidders", h.CreateBidder)
		r.Get("/{tenderId}/bidders/{bidderId}", h.GetBidder)
		r.Put("/{tenderId}/bidders/{bidderId}", h.UpdateBidder)
		r.Delete("/{tenderId}/bidders/{bidderId}", h.DeleteBidder)
	})
}

// --- Tenders ---

func (h *TendersHandler) ListTenders(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	projectID := r.URL.Query().Get("project_id")

	query := `SELECT id, code, name, description, tender_type, status, client_id, project_id, budget_amount, currency, published_at, submission_deadline, bid_open_date, award_date, contract_start, contract_end, bid_bond_pct, performance_bond_pct, advance_payment_pct, retention_pct, retention_release_days, procurement_method, funding_source, notes, created_by, created_at, updated_at FROM tenders WHERE 1=1`
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
	query += ` ORDER BY created_at DESC`

	rows, err := h.db.Query(query, args...)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	tenders := make([]models.Tender, 0)
	for rows.Next() {
		var t models.Tender
		if err := rows.Scan(&t.ID, &t.Code, &t.Name, &t.Description, &t.TenderType, &t.Status, &t.ClientID, &t.ProjectID, &t.BudgetAmount, &t.Currency, &t.PublishedAt, &t.SubmissionDeadline, &t.BidOpenDate, &t.AwardDate, &t.ContractStart, &t.ContractEnd, &t.BidBondPct, &t.PerformanceBondPct, &t.AdvancePaymentPct, &t.RetentionPct, &t.RetentionReleaseDays, &t.ProcurementMethod, &t.FundingSource, &t.Notes, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		tenders = append(tenders, t)
	}
	respondJSON(w, http.StatusOK, tenders)
}

func (h *TendersHandler) CreateTender(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Code                string     `json:"code"`
		Name                string     `json:"name"`
		Description         *string    `json:"description"`
		TenderType          string     `json:"tender_type"`
		Status              string     `json:"status"`
		ClientID            *string    `json:"client_id"`
		ProjectID           *string    `json:"project_id"`
		BudgetAmount        *float64   `json:"budget_amount"`
		Currency            string     `json:"currency"`
		PublishedAt         *time.Time `json:"published_at"`
		SubmissionDeadline  *time.Time `json:"submission_deadline"`
		BidOpenDate         *time.Time `json:"bid_open_date"`
		AwardDate           *time.Time `json:"award_date"`
		ContractStart       *string    `json:"contract_start"`
		ContractEnd         *string    `json:"contract_end"`
		BidBondPct          *float64   `json:"bid_bond_pct"`
		PerformanceBondPct  *float64   `json:"performance_bond_pct"`
		AdvancePaymentPct   *float64   `json:"advance_payment_pct"`
		RetentionPct        *float64   `json:"retention_pct"`
		RetentionReleaseDays *int     `json:"retention_release_days"`
		ProcurementMethod   *string    `json:"procurement_method"`
		FundingSource       *string    `json:"funding_source"`
		Notes               *string    `json:"notes"`
		CreatedBy           *string    `json:"created_by"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO tenders (id, code, name, description, tender_type, status, client_id, project_id, budget_amount, currency, published_at, submission_deadline, bid_open_date, award_date, contract_start, contract_end, bid_bond_pct, performance_bond_pct, advance_payment_pct, retention_pct, retention_release_days, procurement_method, funding_source, notes, created_by, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26,$27)`,
		id, input.Code, input.Name, input.Description, input.TenderType, input.Status, input.ClientID, input.ProjectID, input.BudgetAmount, input.Currency, input.PublishedAt, input.SubmissionDeadline, input.BidOpenDate, input.AwardDate, input.ContractStart, input.ContractEnd, input.BidBondPct, input.PerformanceBondPct, input.AdvancePaymentPct, input.RetentionPct, input.RetentionReleaseDays, input.ProcurementMethod, input.FundingSource, input.Notes, input.CreatedBy, now, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *TendersHandler) GetTender(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var t models.Tender
	err := h.db.QueryRow(`SELECT id, code, name, description, tender_type, status, client_id, project_id, budget_amount, currency, published_at, submission_deadline, bid_open_date, award_date, contract_start, contract_end, bid_bond_pct, performance_bond_pct, advance_payment_pct, retention_pct, retention_release_days, procurement_method, funding_source, notes, created_by, created_at, updated_at FROM tenders WHERE id = $1`, id).
		Scan(&t.ID, &t.Code, &t.Name, &t.Description, &t.TenderType, &t.Status, &t.ClientID, &t.ProjectID, &t.BudgetAmount, &t.Currency, &t.PublishedAt, &t.SubmissionDeadline, &t.BidOpenDate, &t.AwardDate, &t.ContractStart, &t.ContractEnd, &t.BidBondPct, &t.PerformanceBondPct, &t.AdvancePaymentPct, &t.RetentionPct, &t.RetentionReleaseDays, &t.ProcurementMethod, &t.FundingSource, &t.Notes, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "tender not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, t)
}

func (h *TendersHandler) UpdateTender(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Name                *string    `json:"name"`
		Description         *string    `json:"description"`
		Status              *string    `json:"status"`
		BudgetAmount        *float64   `json:"budget_amount"`
		SubmissionDeadline  *time.Time `json:"submission_deadline"`
		BidOpenDate         *time.Time `json:"bid_open_date"`
		AwardDate           *time.Time `json:"award_date"`
		Notes               *string    `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	now := time.Now()
	_, err := h.db.Exec(`UPDATE tenders SET name=COALESCE($1,name), description=COALESCE($2,description), status=COALESCE($3,status), budget_amount=COALESCE($4,budget_amount), submission_deadline=COALESCE($5,submission_deadline), bid_open_date=COALESCE($6,bid_open_date), award_date=COALESCE($7,award_date), notes=COALESCE($8,notes), updated_at=$9 WHERE id=$10`,
		input.Name, input.Description, input.Status, input.BudgetAmount, input.SubmissionDeadline, input.BidOpenDate, input.AwardDate, input.Notes, now, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *TendersHandler) DeleteTender(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM tenders WHERE id = $1`, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// --- Tender Lots ---

func (h *TendersHandler) ListLots(w http.ResponseWriter, r *http.Request) {
	tenderID := chi.URLParam(r, "tenderId")
	rows, err := h.db.Query(`SELECT id, tender_id, lot_number, name, description, estimated_amount, currency, section_id, status, award_decision, created_at FROM tender_lots WHERE tender_id = $1 ORDER BY lot_number`, tenderID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	lots := make([]models.TenderLot, 0)
	for rows.Next() {
		var l models.TenderLot
		if err := rows.Scan(&l.ID, &l.TenderID, &l.LotNumber, &l.Name, &l.Description, &l.EstimatedAmount, &l.Currency, &l.SectionID, &l.Status, &l.AwardDecision, &l.CreatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		lots = append(lots, l)
	}
	respondJSON(w, http.StatusOK, lots)
}

func (h *TendersHandler) CreateLot(w http.ResponseWriter, r *http.Request) {
	tenderID := chi.URLParam(r, "tenderId")
	var input struct {
		LotNumber       int      `json:"lot_number"`
		Name            string   `json:"name"`
		Description     *string  `json:"description"`
		EstimatedAmount *float64 `json:"estimated_amount"`
		Currency        string   `json:"currency"`
		SectionID       *string  `json:"section_id"`
		Status          string   `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO tender_lots (id, tender_id, lot_number, name, description, estimated_amount, currency, section_id, status, created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		id, tenderID, input.LotNumber, input.Name, input.Description, input.EstimatedAmount, input.Currency, input.SectionID, input.Status, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *TendersHandler) GetLot(w http.ResponseWriter, r *http.Request) {
	lotID := chi.URLParam(r, "lotId")
	var l models.TenderLot
	err := h.db.QueryRow(`SELECT id, tender_id, lot_number, name, description, estimated_amount, currency, section_id, status, award_decision, created_at FROM tender_lots WHERE id = $1`, lotID).
		Scan(&l.ID, &l.TenderID, &l.LotNumber, &l.Name, &l.Description, &l.EstimatedAmount, &l.Currency, &l.SectionID, &l.Status, &l.AwardDecision, &l.CreatedAt)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "lot not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, l)
}

func (h *TendersHandler) UpdateLot(w http.ResponseWriter, r *http.Request) {
	lotID := chi.URLParam(r, "lotId")
	var input struct {
		Name            *string  `json:"name"`
		Description     *string  `json:"description"`
		EstimatedAmount *float64 `json:"estimated_amount"`
		Status          *string  `json:"status"`
		AwardDecision   *string  `json:"award_decision"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	_, err := h.db.Exec(`UPDATE tender_lots SET name=COALESCE($1,name), description=COALESCE($2,description), estimated_amount=COALESCE($3,estimated_amount), status=COALESCE($4,status), award_decision=COALESCE($5,award_decision) WHERE id=$6`,
		input.Name, input.Description, input.EstimatedAmount, input.Status, input.AwardDecision, lotID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *TendersHandler) DeleteLot(w http.ResponseWriter, r *http.Request) {
	lotID := chi.URLParam(r, "lotId")
	_, err := h.db.Exec(`DELETE FROM tender_lots WHERE id = $1`, lotID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// --- Tender Bidders ---

func (h *TendersHandler) ListBidders(w http.ResponseWriter, r *http.Request) {
	tenderID := chi.URLParam(r, "tenderId")
	rows, err := h.db.Query(`SELECT id, tender_id, lot_id, contractor_id, bid_number, status, bid_amount, currency, bid_bond_amount, validity_days, submission_date, is_winner, award_amount, award_reason, notes, created_at FROM tender_bidders WHERE tender_id = $1 ORDER BY created_at`, tenderID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	bidders := make([]models.TenderBidder, 0)
	for rows.Next() {
		var b models.TenderBidder
		if err := rows.Scan(&b.ID, &b.TenderID, &b.LotID, &b.ContractorID, &b.BidNumber, &b.Status, &b.BidAmount, &b.Currency, &b.BidBondAmount, &b.ValidityDays, &b.SubmissionDate, &b.IsWinner, &b.AwardAmount, &b.AwardReason, &b.Notes, &b.CreatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		bidders = append(bidders, b)
	}
	respondJSON(w, http.StatusOK, bidders)
}

func (h *TendersHandler) CreateBidder(w http.ResponseWriter, r *http.Request) {
	tenderID := chi.URLParam(r, "tenderId")
	var input struct {
		LotID        *string  `json:"lot_id"`
		ContractorID string   `json:"contractor_id"`
		BidNumber    *string  `json:"bid_number"`
		Status       string   `json:"status"`
		BidAmount    *float64 `json:"bid_amount"`
		Currency     string   `json:"currency"`
		BidBondAmount *float64 `json:"bid_bond_amount"`
		ValidityDays *int     `json:"validity_days"`
		Notes        *string  `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO tender_bidders (id, tender_id, lot_id, contractor_id, bid_number, status, bid_amount, currency, bid_bond_amount, validity_days, notes, created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		id, tenderID, input.LotID, input.ContractorID, input.BidNumber, input.Status, input.BidAmount, input.Currency, input.BidBondAmount, input.ValidityDays, input.Notes, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *TendersHandler) GetBidder(w http.ResponseWriter, r *http.Request) {
	bidderID := chi.URLParam(r, "bidderId")
	var b models.TenderBidder
	err := h.db.QueryRow(`SELECT id, tender_id, lot_id, contractor_id, bid_number, status, bid_amount, currency, bid_bond_amount, validity_days, submission_date, is_winner, award_amount, award_reason, notes, created_at FROM tender_bidders WHERE id = $1`, bidderID).
		Scan(&b.ID, &b.TenderID, &b.LotID, &b.ContractorID, &b.BidNumber, &b.Status, &b.BidAmount, &b.Currency, &b.BidBondAmount, &b.ValidityDays, &b.SubmissionDate, &b.IsWinner, &b.AwardAmount, &b.AwardReason, &b.Notes, &b.CreatedAt)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "bidder not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, b)
}

func (h *TendersHandler) UpdateBidder(w http.ResponseWriter, r *http.Request) {
	bidderID := chi.URLParam(r, "bidderId")
	var input struct {
		Status      *string  `json:"status"`
		BidAmount   *float64 `json:"bid_amount"`
		IsWinner    *bool    `json:"is_winner"`
		AwardAmount *float64 `json:"award_amount"`
		AwardReason *string  `json:"award_reason"`
		Notes       *string  `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	_, err := h.db.Exec(`UPDATE tender_bidders SET status=COALESCE($1,status), bid_amount=COALESCE($2,bid_amount), is_winner=COALESCE($3,is_winner), award_amount=COALESCE($4,award_amount), award_reason=COALESCE($5,award_reason), notes=COALESCE($6,notes) WHERE id=$7`,
		input.Status, input.BidAmount, input.IsWinner, input.AwardAmount, input.AwardReason, input.Notes, bidderID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *TendersHandler) DeleteBidder(w http.ResponseWriter, r *http.Request) {
	bidderID := chi.URLParam(r, "bidderId")
	_, err := h.db.Exec(`DELETE FROM tender_bidders WHERE id = $1`, bidderID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func itoa(i int) string {
	return strconv.Itoa(i)
}
