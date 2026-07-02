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

// FundingHandler handles Funding module endpoints (V027)
type FundingHandler struct {
	db *sql.DB
}

func NewFundingHandler(db *sql.DB) *FundingHandler {
	return &FundingHandler{db: db}
}

func (h *FundingHandler) RegisterRoutes(r chi.Router) {
	r.Route("/funding", func(r chi.Router) {
		// Funding Sources
		r.Get("/sources", h.ListSources)
		r.Post("/sources", h.CreateSource)
		r.Get("/sources/{id}", h.GetSource)
		r.Put("/sources/{id}", h.UpdateSource)
		r.Delete("/sources/{id}", h.DeleteSource)

		// Tranches
		r.Get("/tranches", h.ListTranches)
		r.Post("/tranches", h.CreateTranche)
		r.Get("/tranches/{id}", h.GetTranche)
		r.Put("/tranches/{id}", h.UpdateTranche)

		// Drawdowns
		r.Get("/drawdowns", h.ListDrawdowns)
		r.Post("/drawdowns", h.CreateDrawdown)
		r.Get("/drawdowns/{id}", h.GetDrawdown)

		// Covenants
		r.Get("/covenants", h.ListCovenants)
		r.Post("/covenants", h.CreateCovenant)
		r.Get("/covenants/{id}", h.GetCovenant)
		r.Put("/covenants/{id}", h.UpdateCovenant)

		// Multi-Currency Rates
		r.Get("/rates", h.ListRates)
		r.Post("/rates", h.CreateRate)

		// Hedges
		r.Get("/hedges", h.ListHedges)
		r.Post("/hedges", h.CreateHedge)
		r.Get("/hedges/{id}", h.GetHedge)

		// Guarantees
		r.Get("/guarantees", h.ListGuarantees)
		r.Post("/guarantees", h.CreateGuarantee)
		r.Get("/guarantees/{id}", h.GetGuarantee)
		r.Put("/guarantees/{id}", h.UpdateGuarantee)

		// Guarantee Claims
		r.Get("/guarantees/{id}/claims", h.ListGuaranteeClaims)
		r.Post("/guarantees/{id}/claims", h.CreateGuaranteeClaim)

		// Guarantee Amendments
		r.Get("/guarantees/{id}/amendments", h.ListGuaranteeAmendments)
		r.Post("/guarantees/{id}/amendments", h.CreateGuaranteeAmendment)
	})
}

// --- Funding Sources ---

func (h *FundingHandler) ListSources(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT id, project_id, source_type, source_name, source_code, description, contact_info, commitment_amount, currency, status, is_active, notes, created_at, updated_at FROM funding_sources`
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

	items := make([]models.FundingSource, 0)
	for rows.Next() {
		var m models.FundingSource
		if err := rows.Scan(&m.ID, &m.ProjectID, &m.SourceType, &m.SourceName, &m.SourceCode, &m.Description, &m.ContactInfo, &m.CommitmentAmount, &m.Currency, &m.Status, &m.IsActive, &m.Notes, &m.CreatedAt, &m.UpdatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, m)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *FundingHandler) CreateSource(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID       string   `json:"project_id"`
		SourceType      string   `json:"source_type"`
		SourceName      string   `json:"source_name"`
		SourceCode      *string  `json:"source_code"`
		Description     *string  `json:"description"`
		ContactInfo     *string  `json:"contact_info"`
		CommitmentAmount float64 `json:"commitment_amount"`
		Currency        string   `json:"currency"`
		Status          *string  `json:"status"`
		Notes           *string  `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	now := time.Now()
	status := "active"
	if input.Status != nil {
		status = *input.Status
	}
	_, err := h.db.Exec(`INSERT INTO funding_sources (id, project_id, source_type, source_name, source_code, description, contact_info, commitment_amount, currency, status, notes, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		id, input.ProjectID, input.SourceType, input.SourceName, input.SourceCode, input.Description, input.ContactInfo, input.CommitmentAmount, input.Currency, status, input.Notes, now, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *FundingHandler) GetSource(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var m models.FundingSource
	err := h.db.QueryRow(`SELECT id, project_id, source_type, source_name, source_code, description, contact_info, commitment_amount, currency, status, is_active, notes, created_at, updated_at FROM funding_sources WHERE id = $1`, id).Scan(
		&m.ID, &m.ProjectID, &m.SourceType, &m.SourceName, &m.SourceCode, &m.Description, &m.ContactInfo, &m.CommitmentAmount, &m.Currency, &m.Status, &m.IsActive, &m.Notes, &m.CreatedAt, &m.UpdatedAt)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, m)
}

func (h *FundingHandler) UpdateSource(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		SourceType      *string  `json:"source_type"`
		SourceName      *string  `json:"source_name"`
		SourceCode      *string  `json:"source_code"`
		Description     *string  `json:"description"`
		ContactInfo     *string  `json:"contact_info"`
		CommitmentAmount *float64 `json:"commitment_amount"`
		Currency        *string  `json:"currency"`
		Status          *string  `json:"status"`
		Notes           *string  `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	now := time.Now()
	_, err := h.db.Exec(`UPDATE funding_sources SET source_type=COALESCE($1,source_type), source_name=COALESCE($2,source_name), source_code=COALESCE($3,source_code), description=COALESCE($4,description), contact_info=COALESCE($5,contact_info), commitment_amount=COALESCE($6,commitment_amount), currency=COALESCE($7,currency), status=COALESCE($8,status), notes=COALESCE($9,notes), updated_at=$10 WHERE id=$11`,
		input.SourceType, input.SourceName, input.SourceCode, input.Description, input.ContactInfo, input.CommitmentAmount, input.Currency, input.Status, input.Notes, now, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *FundingHandler) DeleteSource(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`UPDATE funding_sources SET is_active=FALSE, updated_at=NOW() WHERE id=$1`, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// --- Tranches ---

func (h *FundingHandler) ListTranches(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT id, project_id, funding_source_id, tranche_name, amount, currency, expected_date, actual_date, status, terms, is_active, created_at, updated_at FROM funding_tranches`
	var rows *sql.Rows
	var err error
	if projectID != "" {
		rows, err = h.db.Query(query+` WHERE project_id = $1 ORDER BY expected_date`, projectID)
	} else {
		rows, err = h.db.Query(query + ` ORDER BY expected_date`)
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	items := make([]models.FundingTranche, 0)
	for rows.Next() {
		var m models.FundingTranche
		if err := rows.Scan(&m.ID, &m.ProjectID, &m.FundingSourceID, &m.TrancheName, &m.Amount, &m.Currency, &m.ExpectedDate, &m.ActualDate, &m.Status, &m.Terms, &m.IsActive, &m.CreatedAt, &m.UpdatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, m)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *FundingHandler) CreateTranche(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID       string  `json:"project_id"`
		FundingSourceID string  `json:"funding_source_id"`
		TrancheName     string  `json:"tranche_name"`
		Amount          float64 `json:"amount"`
		Currency        string  `json:"currency"`
		ExpectedDate    *string `json:"expected_date"`
		ActualDate      *string `json:"actual_date"`
		Status          *string `json:"status"`
		Terms           *string `json:"terms"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	now := time.Now()
	status := "planned"
	if input.Status != nil {
		status = *input.Status
	}
	_, err := h.db.Exec(`INSERT INTO funding_tranches (id,project_id,funding_source_id,tranche_name,amount,currency,expected_date,actual_date,status,terms,created_at,updated_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		id, input.ProjectID, input.FundingSourceID, input.TrancheName, input.Amount, input.Currency, input.ExpectedDate, input.ActualDate, status, input.Terms, now, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *FundingHandler) GetTranche(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var m models.FundingTranche
	err := h.db.QueryRow(`SELECT id,project_id,funding_source_id,tranche_name,amount,currency,expected_date,actual_date,status,terms,is_active,created_at,updated_at FROM funding_tranches WHERE id=$1`, id).Scan(
		&m.ID, &m.ProjectID, &m.FundingSourceID, &m.TrancheName, &m.Amount, &m.Currency, &m.ExpectedDate, &m.ActualDate, &m.Status, &m.Terms, &m.IsActive, &m.CreatedAt, &m.UpdatedAt)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, m)
}

func (h *FundingHandler) UpdateTranche(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		TrancheName *string  `json:"tranche_name"`
		Amount      *float64 `json:"amount"`
		Status      *string  `json:"status"`
		ActualDate  *string  `json:"actual_date"`
		Terms       *string  `json:"terms"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	_, err := h.db.Exec(`UPDATE funding_tranches SET tranche_name=COALESCE($1,tranche_name), amount=COALESCE($2,amount), status=COALESCE($3,status), actual_date=COALESCE($4,actual_date), terms=COALESCE($5,terms), updated_at=NOW() WHERE id=$6`,
		input.TrancheName, input.Amount, input.Status, input.ActualDate, input.Terms, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// --- Drawdowns ---

func (h *FundingHandler) ListDrawdowns(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT id,project_id,funding_source_id,tranche_id,drawdown_date,amount,currency,exchange_rate,reference,status,notes,created_at FROM funding_drawdowns`
	var rows *sql.Rows
	var err error
	if projectID != "" {
		rows, err = h.db.Query(query+` WHERE project_id=$1 ORDER BY drawdown_date DESC`, projectID)
	} else {
		rows, err = h.db.Query(query + ` ORDER BY drawdown_date DESC`)
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	items := make([]models.FundingDrawdown, 0)
	for rows.Next() {
		var m models.FundingDrawdown
		if err := rows.Scan(&m.ID, &m.ProjectID, &m.FundingSourceID, &m.TrancheID, &m.DrawdownDate, &m.Amount, &m.Currency, &m.ExchangeRate, &m.Reference, &m.Status, &m.Notes, &m.CreatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, m)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *FundingHandler) CreateDrawdown(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID       string  `json:"project_id"`
		FundingSourceID string  `json:"funding_source_id"`
		TrancheID       *string `json:"tranche_id"`
		DrawdownDate    string  `json:"drawdown_date"`
		Amount          float64 `json:"amount"`
		Currency        string  `json:"currency"`
		ExchangeRate    float64 `json:"exchange_rate"`
		Reference       *string `json:"reference"`
		Status          *string `json:"status"`
		Notes           *string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	status := "completed"
	if input.Status != nil {
		status = *input.Status
	}
	_, err := h.db.Exec(`INSERT INTO funding_drawdowns (id,project_id,funding_source_id,tranche_id,drawdown_date,amount,currency,exchange_rate,reference,status,notes,created_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,NOW())`,
		id, input.ProjectID, input.FundingSourceID, input.TrancheID, input.DrawdownDate, input.Amount, input.Currency, input.ExchangeRate, input.Reference, status, input.Notes)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *FundingHandler) GetDrawdown(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var m models.FundingDrawdown
	err := h.db.QueryRow(`SELECT id,project_id,funding_source_id,tranche_id,drawdown_date,amount,currency,exchange_rate,reference,status,notes,created_at FROM funding_drawdowns WHERE id=$1`, id).Scan(
		&m.ID, &m.ProjectID, &m.FundingSourceID, &m.TrancheID, &m.DrawdownDate, &m.Amount, &m.Currency, &m.ExchangeRate, &m.Reference, &m.Status, &m.Notes, &m.CreatedAt)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, m)
}

// --- Covenants ---

func (h *FundingHandler) ListCovenants(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT id,project_id,funding_source_id,covenant_type,covenant_name,description,metric,threshold,status,breach_date,breach_notes,is_active,created_at,updated_at FROM funding_covenants`
	var rows *sql.Rows
	var err error
	if projectID != "" {
		rows, err = h.db.Query(query+` WHERE project_id=$1 ORDER BY created_at`, projectID)
	} else {
		rows, err = h.db.Query(query + ` ORDER BY created_at`)
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	items := make([]models.FundingCovenant, 0)
	for rows.Next() {
		var m models.FundingCovenant
		if err := rows.Scan(&m.ID, &m.ProjectID, &m.FundingSourceID, &m.CovenantType, &m.CovenantName, &m.Description, &m.Metric, &m.Threshold, &m.Status, &m.BreachDate, &m.BreachNotes, &m.IsActive, &m.CreatedAt, &m.UpdatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, m)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *FundingHandler) CreateCovenant(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID       string  `json:"project_id"`
		FundingSourceID string  `json:"funding_source_id"`
		CovenantType    string  `json:"covenant_type"`
		CovenantName    string  `json:"covenant_name"`
		Description     *string `json:"description"`
		Metric          *string `json:"metric"`
		Threshold       *string `json:"threshold"`
		Status          *string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	now := time.Now()
	status := "active"
	if input.Status != nil {
		status = *input.Status
	}
	_, err := h.db.Exec(`INSERT INTO funding_covenants (id,project_id,funding_source_id,covenant_type,covenant_name,description,metric,threshold,status,created_at,updated_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		id, input.ProjectID, input.FundingSourceID, input.CovenantType, input.CovenantName, input.Description, input.Metric, input.Threshold, status, now, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *FundingHandler) GetCovenant(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var m models.FundingCovenant
	err := h.db.QueryRow(`SELECT id,project_id,funding_source_id,covenant_type,covenant_name,description,metric,threshold,status,breach_date,breach_notes,is_active,created_at,updated_at FROM funding_covenants WHERE id=$1`, id).Scan(
		&m.ID, &m.ProjectID, &m.FundingSourceID, &m.CovenantType, &m.CovenantName, &m.Description, &m.Metric, &m.Threshold, &m.Status, &m.BreachDate, &m.BreachNotes, &m.IsActive, &m.CreatedAt, &m.UpdatedAt)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, m)
}

func (h *FundingHandler) UpdateCovenant(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Status      *string `json:"status"`
		BreachDate  *string `json:"breach_date"`
		BreachNotes *string `json:"breach_notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	_, err := h.db.Exec(`UPDATE funding_covenants SET status=COALESCE($1,status), breach_date=COALESCE($2,breach_date), breach_notes=COALESCE($3,breach_notes), updated_at=NOW() WHERE id=$4`,
		input.Status, input.BreachDate, input.BreachNotes, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// --- Rates ---

func (h *FundingHandler) ListRates(w http.ResponseWriter, r *http.Request) {
	query := `SELECT id,base_currency,target_currency,rate,rate_date,source,is_historical,notes,created_at FROM multi_currency_rates ORDER BY rate_date DESC`
	rows, err := h.db.Query(query)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	items := make([]models.MultiCurrencyRate, 0)
	for rows.Next() {
		var m models.MultiCurrencyRate
		if err := rows.Scan(&m.ID, &m.BaseCurrency, &m.TargetCurrency, &m.Rate, &m.RateDate, &m.Source, &m.IsHistorical, &m.Notes, &m.CreatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, m)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *FundingHandler) CreateRate(w http.ResponseWriter, r *http.Request) {
	var input struct {
		BaseCurrency   string  `json:"base_currency"`
		TargetCurrency string  `json:"target_currency"`
		Rate           float64 `json:"rate"`
		RateDate       string  `json:"rate_date"`
		Source         *string `json:"source"`
		IsHistorical   bool    `json:"is_historical"`
		Notes          *string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	_, err := h.db.Exec(`INSERT INTO multi_currency_rates (id,base_currency,target_currency,rate,rate_date,source,is_historical,notes,created_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,NOW())`,
		id, input.BaseCurrency, input.TargetCurrency, input.Rate, input.RateDate, input.Source, input.IsHistorical, input.Notes)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

// --- Hedges ---

func (h *FundingHandler) ListHedges(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT id,project_id,hedge_type,base_currency,hedge_currency,notional_amount,strike_rate,maturity_date,counterparty,status,is_active,notes,created_at,updated_at FROM currency_hedges`
	var rows *sql.Rows
	var err error
	if projectID != "" {
		rows, err = h.db.Query(query+` WHERE project_id=$1 ORDER BY created_at`, projectID)
	} else {
		rows, err = h.db.Query(query + ` ORDER BY created_at`)
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	items := make([]models.CurrencyHedge, 0)
	for rows.Next() {
		var m models.CurrencyHedge
		if err := rows.Scan(&m.ID, &m.ProjectID, &m.HedgeType, &m.BaseCurrency, &m.HedgeCurrency, &m.NotionalAmount, &m.StrikeRate, &m.MaturityDate, &m.Counterparty, &m.Status, &m.IsActive, &m.Notes, &m.CreatedAt, &m.UpdatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, m)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *FundingHandler) CreateHedge(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID      string   `json:"project_id"`
		HedgeType      string   `json:"hedge_type"`
		BaseCurrency   string   `json:"base_currency"`
		HedgeCurrency  string   `json:"hedge_currency"`
		NotionalAmount float64  `json:"notional_amount"`
		StrikeRate     *float64 `json:"strike_rate"`
		MaturityDate   *string  `json:"maturity_date"`
		Counterparty   *string  `json:"counterparty"`
		Status         *string  `json:"status"`
		Notes          *string  `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	now := time.Now()
	status := "active"
	if input.Status != nil {
		status = *input.Status
	}
	_, err := h.db.Exec(`INSERT INTO currency_hedges (id,project_id,hedge_type,base_currency,hedge_currency,notional_amount,strike_rate,maturity_date,counterparty,status,notes,created_at,updated_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		id, input.ProjectID, input.HedgeType, input.BaseCurrency, input.HedgeCurrency, input.NotionalAmount, input.StrikeRate, input.MaturityDate, input.Counterparty, status, input.Notes, now, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *FundingHandler) GetHedge(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var m models.CurrencyHedge
	err := h.db.QueryRow(`SELECT id,project_id,hedge_type,base_currency,hedge_currency,notional_amount,strike_rate,maturity_date,counterparty,status,is_active,notes,created_at,updated_at FROM currency_hedges WHERE id=$1`, id).Scan(
		&m.ID, &m.ProjectID, &m.HedgeType, &m.BaseCurrency, &m.HedgeCurrency, &m.NotionalAmount, &m.StrikeRate, &m.MaturityDate, &m.Counterparty, &m.Status, &m.IsActive, &m.Notes, &m.CreatedAt, &m.UpdatedAt)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, m)
}

// --- Guarantees ---

func (h *FundingHandler) ListGuarantees(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT id,project_id,contract_id,guarantee_type,guarantee_number,issuing_ban,beneficiary,applicant,amount,currency,issue_date,expiry_date,claim_expiry_date,status,is_active,notes,created_at,updated_at FROM guarantees`
	// Fix column name
	query = `SELECT id,project_id,contract_id,guarantee_type,guarantee_number,issuing_bank,beneficiary,applicant,amount,currency,issue_date,expiry_date,claim_expiry_date,status,is_active,notes,created_at,updated_at FROM guarantees`
	var rows *sql.Rows
	var err error
	if projectID != "" {
		rows, err = h.db.Query(query+` WHERE project_id=$1 ORDER BY created_at`, projectID)
	} else {
		rows, err = h.db.Query(query + ` ORDER BY created_at`)
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	items := make([]models.Guarantee, 0)
	for rows.Next() {
		var m models.Guarantee
		if err := rows.Scan(&m.ID, &m.ProjectID, &m.ContractID, &m.GuaranteeType, &m.GuaranteeNumber, &m.IssuingBank, &m.Beneficiary, &m.Applicant, &m.Amount, &m.Currency, &m.IssueDate, &m.ExpiryDate, &m.ClaimExpiryDate, &m.Status, &m.IsActive, &m.Notes, &m.CreatedAt, &m.UpdatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, m)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *FundingHandler) CreateGuarantee(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID       string   `json:"project_id"`
		ContractID      *string  `json:"contract_id"`
		GuaranteeType   string   `json:"guarantee_type"`
		GuaranteeNumber *string  `json:"guarantee_number"`
		IssuingBank     *string  `json:"issuing_bank"`
		Beneficiary     *string  `json:"beneficiary"`
		Applicant       *string  `json:"applicant"`
		Amount          float64  `json:"amount"`
		Currency        string   `json:"currency"`
		IssueDate       *string  `json:"issue_date"`
		ExpiryDate      *string  `json:"expiry_date"`
		Status          *string  `json:"status"`
		Notes           *string  `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	now := time.Now()
	status := "active"
	if input.Status != nil {
		status = *input.Status
	}
	_, err := h.db.Exec(`INSERT INTO guarantees (id,project_id,contract_id,guarantee_type,guarantee_number,issuing_bank,beneficiary,applicant,amount,currency,issue_date,expiry_date,status,notes,created_at,updated_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)`,
		id, input.ProjectID, input.ContractID, input.GuaranteeType, input.GuaranteeNumber, input.IssuingBank, input.Beneficiary, input.Applicant, input.Amount, input.Currency, input.IssueDate, input.ExpiryDate, status, input.Notes, now, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *FundingHandler) GetGuarantee(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var m models.Guarantee
	err := h.db.QueryRow(`SELECT id,project_id,contract_id,guarantee_type,guarantee_number,issuing_bank,beneficiary,applicant,amount,currency,issue_date,expiry_date,claim_expiry_date,status,is_active,notes,created_at,updated_at FROM guarantees WHERE id=$1`, id).Scan(
		&m.ID, &m.ProjectID, &m.ContractID, &m.GuaranteeType, &m.GuaranteeNumber, &m.IssuingBank, &m.Beneficiary, &m.Applicant, &m.Amount, &m.Currency, &m.IssueDate, &m.ExpiryDate, &m.ClaimExpiryDate, &m.Status, &m.IsActive, &m.Notes, &m.CreatedAt, &m.UpdatedAt)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, m)
}

func (h *FundingHandler) UpdateGuarantee(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Status     *string `json:"status"`
		ExpiryDate *string `json:"expiry_date"`
		Notes      *string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	_, err := h.db.Exec(`UPDATE guarantees SET status=COALESCE($1,status), expiry_date=COALESCE($2,expiry_date), notes=COALESCE($3,notes), updated_at=NOW() WHERE id=$4`,
		input.Status, input.ExpiryDate, input.Notes, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// --- Guarantee Claims ---

func (h *FundingHandler) ListGuaranteeClaims(w http.ResponseWriter, r *http.Request) {
	guaranteeID := chi.URLParam(r, "id")
	rows, err := h.db.Query(`SELECT id,project_id,guarantee_id,claim_date,claim_amount,claim_reason,claim_status,response_date,settlement_amount,notes,created_at,updated_at FROM guarantee_claims WHERE guarantee_id=$1 ORDER BY created_at`, guaranteeID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	items := make([]models.GuaranteeClaim, 0)
	for rows.Next() {
		var m models.GuaranteeClaim
		if err := rows.Scan(&m.ID, &m.ProjectID, &m.GuaranteeID, &m.ClaimDate, &m.ClaimAmount, &m.ClaimReason, &m.ClaimStatus, &m.ResponseDate, &m.SettlementAmount, &m.Notes, &m.CreatedAt, &m.UpdatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, m)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *FundingHandler) CreateGuaranteeClaim(w http.ResponseWriter, r *http.Request) {
	guaranteeID := chi.URLParam(r, "id")
	var input struct {
		ProjectID   string   `json:"project_id"`
		ClaimDate   string   `json:"claim_date"`
		ClaimAmount float64  `json:"claim_amount"`
		ClaimReason *string  `json:"claim_reason"`
		Notes       *string  `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO guarantee_claims (id,project_id,guarantee_id,claim_date,claim_amount,claim_reason,claim_status,notes,created_at,updated_at) VALUES($1,$2,$3,$4,$5,$6,'submitted',$7,$8,$9)`,
		id, input.ProjectID, guaranteeID, input.ClaimDate, input.ClaimAmount, input.ClaimReason, input.Notes, now, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

// --- Guarantee Amendments ---

func (h *FundingHandler) ListGuaranteeAmendments(w http.ResponseWriter, r *http.Request) {
	guaranteeID := chi.URLParam(r, "id")
	rows, err := h.db.Query(`SELECT id,guarantee_id,amendment_number,amendment_date,description,new_amount,new_expiry_date,notes,created_at FROM guarantee_amendments WHERE guarantee_id=$1 ORDER BY created_at`, guaranteeID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	items := make([]models.GuaranteeAmendment, 0)
	for rows.Next() {
		var m models.GuaranteeAmendment
		if err := rows.Scan(&m.ID, &m.GuaranteeID, &m.AmendmentNumber, &m.AmendmentDate, &m.Description, &m.NewAmount, &m.NewExpiryDate, &m.Notes, &m.CreatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, m)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *FundingHandler) CreateGuaranteeAmendment(w http.ResponseWriter, r *http.Request) {
	guaranteeID := chi.URLParam(r, "id")
	var input struct {
		AmendmentNumber *string  `json:"amendment_number"`
		AmendmentDate   *string  `json:"amendment_date"`
		Description     *string  `json:"description"`
		NewAmount       *float64 `json:"new_amount"`
		NewExpiryDate   *string  `json:"new_expiry_date"`
		Notes           *string  `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	_, err := h.db.Exec(`INSERT INTO guarantee_amendments (id,guarantee_id,amendment_number,amendment_date,description,new_amount,new_expiry_date,notes,created_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,NOW())`,
		id, guaranteeID, input.AmendmentNumber, input.AmendmentDate, input.Description, input.NewAmount, input.NewExpiryDate, input.Notes)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}