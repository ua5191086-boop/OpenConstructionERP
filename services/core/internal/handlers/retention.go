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

// RetentionRelease — удержание по контракту
type RetentionRelease struct {
	ID              string   `json:"id" db:"id"`
	ContractID      string   `json:"contract_id" db:"contract_id"`
	ProjectID       string   `json:"project_id" db:"project_id"`
	RetentionPct    float64  `json:"retention_pct" db:"retention_pct"`
	RetentionAmount float64  `json:"retention_amount" db:"retention_amount"`
	ReleasedAmount  *float64 `json:"released_amount" db:"released_amount"`
	ReleaseCond     string   `json:"release_condition" db:"release_condition"`
	ReleaseDate     *string  `json:"release_date" db:"release_date"`
	ReleaseStatus   string   `json:"release_status" db:"release_status"`
	Currency        string   `json:"currency" db:"currency"`
	Notes           *string  `json:"notes" db:"notes"`
	CreatedAt       string   `json:"created_at" db:"created_at"`
	UpdatedAt       string   `json:"updated_at" db:"updated_at"`
}

// Guarantee — банковская гарантия
type Guarantee struct {
	ID              string  `json:"id" db:"id"`
	ContractID      string  `json:"contract_id" db:"contract_id"`
	ProjectID       string  `json:"project_id" db:"project_id"`
	GuaranteeType   string  `json:"guarantee_type" db:"guarantee_type"`
	GuaranteeNumber string  `json:"guarantee_number" db:"guarantee_number"`
	IssuerBank      string  `json:"issuer_bank" db:"issuer_bank"`
	Beneficiary     string  `json:"beneficiary" db:"beneficiary"`
	Amount          float64 `json:"amount" db:"amount"`
	Currency        string  `json:"currency" db:"currency"`
	IssueDate       string  `json:"issue_date" db:"issue_date"`
	ExpiryDate      string  `json:"expiry_date" db:"expiry_date"`
	ClaimDeadline   *string `json:"claim_deadline" db:"claim_deadline"`
	Status          string  `json:"status" db:"status"`
	ExtendedTo      *string `json:"extended_to" db:"extended_to"`
	ExtensionCount  *int    `json:"extension_count" db:"extension_count"`
	DocumentRef     *string `json:"document_ref" db:"document_ref"`
	Notes           *string `json:"notes" db:"notes"`
	CreatedAt       string  `json:"created_at" db:"created_at"`
	UpdatedAt       string  `json:"updated_at" db:"updated_at"`
}

// CurrencyRate — кросс-курс валют
type CurrencyRate struct {
	ID              string  `json:"id" db:"id"`
	BaseCurrency    string  `json:"base_currency" db:"base_currency"`
	TargetCurrency  string  `json:"target_currency" db:"target_currency"`
	Rate            float64 `json:"rate" db:"rate"`
	RateDate        string  `json:"rate_date" db:"rate_date"`
	Source          *string `json:"source" db:"source"`
	IsHistorical    bool    `json:"is_historical" db:"is_historical"`
	Notes           *string `json:"notes" db:"notes"`
	CreatedAt       string  `json:"created_at" db:"created_at"`
}

// MCTx — мультивалютная транзакция
type MCTx struct {
	ID               string   `json:"id" db:"id"`
	ProjectID        string   `json:"project_id" db:"project_id"`
	TxType           string   `json:"transaction_type" db:"transaction_type"`
	SourceCurrency   string   `json:"source_currency" db:"source_currency"`
	TargetCurrency   string   `json:"target_currency" db:"target_currency"`
	SourceAmount     float64  `json:"source_amount" db:"source_amount"`
	TargetAmount     float64  `json:"target_amount" db:"target_amount"`
	ExchangeRate     float64  `json:"exchange_rate" db:"exchange_rate"`
	TxDate           string   `json:"transaction_date" db:"transaction_date"`
	RefType          *string  `json:"reference_type" db:"reference_type"`
	RefID            *string  `json:"reference_id" db:"reference_id"`
	RealizedGL       *float64 `json:"realized_gain_loss" db:"realized_gain_loss"`
	Status           string   `json:"status" db:"status"`
	CreatedBy        *string  `json:"created_by" db:"created_by"`
	Notes            *string  `json:"notes" db:"notes"`
	CreatedAt        string   `json:"created_at" db:"created_at"`
}

// RetentionHandler — HTTP handler for retention/guarantees/multicurrency
type RetentionHandler struct {
	db *sqlx.DB
}

func NewRetentionHandler(db *sqlx.DB) *RetentionHandler {
	return &RetentionHandler{db: db}
}

func (h *RetentionHandler) RegisterRoutes(r chi.Router) {
	r.Route("/retention", func(r chi.Router) {
		r.Get("/", h.ListRetentions)
		r.Post("/", h.CreateRetention)
		r.Get("/{id}", h.GetRetention)
		r.Put("/{id}", h.UpdateRetention)
		r.Delete("/{id}", h.DeleteRetention)
		r.Get("/contract/{contractId}", h.ListRetentionsByContract)
		r.Get("/project/{projectId}", h.ListRetentionsByProject)
		r.Get("/active", h.ListActiveRetentions)
	})
	r.Route("/guarantees", func(r chi.Router) {
		r.Get("/", h.ListGuarantees)
		r.Post("/", h.CreateGuarantee)
		r.Get("/{id}", h.GetGuarantee)
		r.Put("/{id}", h.UpdateGuarantee)
		r.Delete("/{id}", h.DeleteGuarantee)
		r.Get("/contract/{contractId}", h.ListGuaranteesByContract)
		r.Get("/project/{projectId}", h.ListGuaranteesByProject)
		r.Get("/active", h.ListActiveGuarantees)
		r.Get("/expiring/{days}", h.ListExpiringGuarantees)
	})
	r.Route("/currency-rates", func(r chi.Router) {
		r.Get("/", h.ListCurrencyRates)
		r.Post("/", h.CreateCurrencyRate)
		r.Get("/{base}/{target}", h.GetLatestRate)
		r.Get("/{base}/{target}/history", h.GetRateHistory)
		r.Delete("/{id}", h.DeleteCurrencyRate)
	})
	r.Route("/mc-transactions", func(r chi.Router) {
		r.Get("/", h.ListMCTx)
		r.Post("/", h.CreateMCTx)
		r.Get("/{id}", h.GetMCTx)
		r.Get("/project/{projectId}", h.ListMCTxByProject)
	})
}

// ---- Retention ----

func (h *RetentionHandler) ListRetentions(w http.ResponseWriter, r *http.Request) {
	var items []RetentionRelease
	if err := h.db.Select(&items, "SELECT * FROM retention_releases ORDER BY created_at DESC"); err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("query failed: %v", err))
		return
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *RetentionHandler) CreateRetention(w http.ResponseWriter, r *http.Request) {
	var input RetentionRelease
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, fmt.Sprintf("invalid json: %v", err))
		return
	}
	var item RetentionRelease
	err := h.db.Get(&item, `INSERT INTO retention_releases 
		(contract_id, project_id, retention_pct, retention_amount, released_amount, release_condition, release_date, release_status, currency, notes)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		RETURNING *`,
		input.ContractID, input.ProjectID, input.RetentionPct, input.RetentionAmount,
		nullableFloat(input.ReleasedAmount), input.ReleaseCond, nullableString(input.ReleaseDate),
		input.ReleaseStatus, input.Currency, nullableString(input.Notes))
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("insert failed: %v", err))
		return
	}
	respondJSON(w, http.StatusCreated, item)
}

func (h *RetentionHandler) GetRetention(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var item RetentionRelease
	if err := h.db.Get(&item, "SELECT * FROM retention_releases WHERE id=$1", id); err != nil {
		respondError(w, http.StatusNotFound, "retention not found")
		return
	}
	respondJSON(w, http.StatusOK, item)
}

func (h *RetentionHandler) UpdateRetention(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input RetentionRelease
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}
	var item RetentionRelease
	err := h.db.Get(&item, `UPDATE retention_releases SET
		retention_pct=$1, retention_amount=$2, released_amount=$3, release_condition=$4,
		release_date=$5, release_status=$6, currency=$7, notes=$8, updated_at=NOW()
		WHERE id=$9 RETURNING *`,
		input.RetentionPct, input.RetentionAmount, nullableFloat(input.ReleasedAmount),
		input.ReleaseCond, nullableString(input.ReleaseDate), input.ReleaseStatus,
		input.Currency, nullableString(input.Notes), id)
	if err != nil {
		respondError(w, http.StatusNotFound, "retention not found")
		return
	}
	respondJSON(w, http.StatusOK, item)
}

func (h *RetentionHandler) DeleteRetention(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec("DELETE FROM retention_releases WHERE id=$1", id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("delete failed: %v", err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *RetentionHandler) ListRetentionsByContract(w http.ResponseWriter, r *http.Request) {
	cid := chi.URLParam(r, "contractId")
	var items []RetentionRelease
	if err := h.db.Select(&items, "SELECT * FROM retention_releases WHERE contract_id=$1 ORDER BY created_at DESC", cid); err != nil {
		respondError(w, http.StatusInternalServerError, "query failed")
		return
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *RetentionHandler) ListRetentionsByProject(w http.ResponseWriter, r *http.Request) {
	pid := chi.URLParam(r, "projectId")
	var items []RetentionRelease
	if err := h.db.Select(&items, "SELECT * FROM retention_releases WHERE project_id=$1 ORDER BY created_at DESC", pid); err != nil {
		respondError(w, http.StatusInternalServerError, "query failed")
		return
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *RetentionHandler) ListActiveRetentions(w http.ResponseWriter, r *http.Request) {
	var items []RetentionRelease
	if err := h.db.Select(&items, "SELECT * FROM retention_releases WHERE release_status IN ('held','partially_released') ORDER BY created_at DESC"); err != nil {
		respondError(w, http.StatusInternalServerError, "query failed")
		return
	}
	respondJSON(w, http.StatusOK, items)
}

// ---- Guarantees ----

func (h *RetentionHandler) ListGuarantees(w http.ResponseWriter, r *http.Request) {
	var items []Guarantee
	if err := h.db.Select(&items, "SELECT * FROM guarantees ORDER BY created_at DESC"); err != nil {
		respondError(w, http.StatusInternalServerError, "query failed")
		return
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *RetentionHandler) CreateGuarantee(w http.ResponseWriter, r *http.Request) {
	var input Guarantee
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}
	var item Guarantee
	err := h.db.Get(&item, `INSERT INTO guarantees
		(contract_id, project_id, guarantee_type, guarantee_number, issuer_bank, beneficiary,
		 amount, currency, issue_date, expiry_date, claim_deadline, status, document_ref, notes)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
		RETURNING *`,
		input.ContractID, input.ProjectID, input.GuaranteeType, input.GuaranteeNumber,
		input.IssuerBank, input.Beneficiary, input.Amount, input.Currency,
		input.IssueDate, input.ExpiryDate, nullableString(input.ClaimDeadline),
		input.Status, nullableString(input.DocumentRef), nullableString(input.Notes))
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("insert failed: %v", err))
		return
	}
	respondJSON(w, http.StatusCreated, item)
}

func (h *RetentionHandler) GetGuarantee(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var item Guarantee
	if err := h.db.Get(&item, "SELECT * FROM guarantees WHERE id=$1", id); err != nil {
		respondError(w, http.StatusNotFound, "guarantee not found")
		return
	}
	respondJSON(w, http.StatusOK, item)
}

func (h *RetentionHandler) UpdateGuarantee(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input Guarantee
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}
	var item Guarantee
	err := h.db.Get(&item, `UPDATE guarantees SET
		guarantee_type=$1, guarantee_number=$2, issuer_bank=$3, beneficiary=$4,
		amount=$5, currency=$6, issue_date=$7, expiry_date=$8, claim_deadline=$9,
		status=$10, extended_to=$11, extension_count=$12, document_ref=$13, notes=$14,
		updated_at=NOW()
		WHERE id=$15 RETURNING *`,
		input.GuaranteeType, input.GuaranteeNumber, input.IssuerBank, input.Beneficiary,
		input.Amount, input.Currency, input.IssueDate, input.ExpiryDate,
		nullableString(input.ClaimDeadline), input.Status, nullableString(input.ExtendedTo),
		nullableInt(input.ExtensionCount), nullableString(input.DocumentRef),
		nullableString(input.Notes), id)
	if err != nil {
		respondError(w, http.StatusNotFound, "guarantee not found")
		return
	}
	respondJSON(w, http.StatusOK, item)
}

func (h *RetentionHandler) DeleteGuarantee(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec("DELETE FROM guarantees WHERE id=$1", id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "delete failed")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *RetentionHandler) ListGuaranteesByContract(w http.ResponseWriter, r *http.Request) {
	cid := chi.URLParam(r, "contractId")
	var items []Guarantee
	h.db.Select(&items, "SELECT * FROM guarantees WHERE contract_id=$1 ORDER BY created_at DESC", cid)
	respondJSON(w, http.StatusOK, items)
}

func (h *RetentionHandler) ListGuaranteesByProject(w http.ResponseWriter, r *http.Request) {
	pid := chi.URLParam(r, "projectId")
	var items []Guarantee
	h.db.Select(&items, "SELECT * FROM guarantees WHERE project_id=$1 ORDER BY created_at DESC", pid)
	respondJSON(w, http.StatusOK, items)
}

func (h *RetentionHandler) ListActiveGuarantees(w http.ResponseWriter, r *http.Request) {
	var items []Guarantee
	h.db.Select(&items, "SELECT * FROM guarantees WHERE status='active' ORDER BY expiry_date ASC")
	respondJSON(w, http.StatusOK, items)
}

func (h *RetentionHandler) ListExpiringGuarantees(w http.ResponseWriter, r *http.Request) {
	days := chi.URLParam(r, "days")
	if days == "" {
		days = "30"
	}
	var items []Guarantee
	h.db.Select(&items, `SELECT * FROM guarantees 
		WHERE status='active' AND expiry_date <= CURRENT_DATE + ($1||' days')::INTERVAL
		ORDER BY expiry_date ASC`, days)
	respondJSON(w, http.StatusOK, items)
}

// ---- Currency Rates ----

func (h *RetentionHandler) ListCurrencyRates(w http.ResponseWriter, r *http.Request) {
	var items []CurrencyRate
	h.db.Select(&items, "SELECT * FROM currency_rates ORDER BY rate_date DESC")
	respondJSON(w, http.StatusOK, items)
}

func (h *RetentionHandler) CreateCurrencyRate(w http.ResponseWriter, r *http.Request) {
	var input CurrencyRate
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}
	var item CurrencyRate
	err := h.db.Get(&item, `INSERT INTO currency_rates (base_currency, target_currency, rate, rate_date, source, is_historical, notes)
		VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING *`,
		input.BaseCurrency, input.TargetCurrency, input.Rate, input.RateDate,
		nullableString(input.Source), input.IsHistorical, nullableString(input.Notes))
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("insert failed: %v", err))
		return
	}
	respondJSON(w, http.StatusCreated, item)
}

func (h *RetentionHandler) GetLatestRate(w http.ResponseWriter, r *http.Request) {
	base := chi.URLParam(r, "base")
	target := chi.URLParam(r, "target")
	var item CurrencyRate
	err := h.db.Get(&item, `SELECT * FROM currency_rates 
		WHERE base_currency=$1 AND target_currency=$2 
		ORDER BY rate_date DESC LIMIT 1`, base, target)
	if err != nil {
		respondError(w, http.StatusNotFound, "rate not found")
		return
	}
	respondJSON(w, http.StatusOK, item)
}

func (h *RetentionHandler) GetRateHistory(w http.ResponseWriter, r *http.Request) {
	base := chi.URLParam(r, "base")
	target := chi.URLParam(r, "target")
	var items []CurrencyRate
	h.db.Select(&items, `SELECT * FROM currency_rates 
		WHERE base_currency=$1 AND target_currency=$2 
		ORDER BY rate_date DESC`, base, target)
	respondJSON(w, http.StatusOK, items)
}

func (h *RetentionHandler) DeleteCurrencyRate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	h.db.Exec("DELETE FROM currency_rates WHERE id=$1", id)
	w.WriteHeader(http.StatusNoContent)
}

// ---- Multi-Currency Transactions ----

func (h *RetentionHandler) ListMCTx(w http.ResponseWriter, r *http.Request) {
	var items []MCTx
	h.db.Select(&items, "SELECT * FROM multi_currency_transactions ORDER BY created_at DESC")
	respondJSON(w, http.StatusOK, items)
}

func (h *RetentionHandler) CreateMCTx(w http.ResponseWriter, r *http.Request) {
	var input MCTx
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}
	var item MCTx
	err := h.db.Get(&item, `INSERT INTO multi_currency_transactions
		(project_id, transaction_type, source_currency, target_currency, source_amount, target_amount,
		 exchange_rate, transaction_date, reference_type, reference_id, realized_gain_loss, status, created_by, notes)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14) RETURNING *`,
		input.ProjectID, input.TxType, input.SourceCurrency, input.TargetCurrency,
		input.SourceAmount, input.TargetAmount, input.ExchangeRate, input.TxDate,
		nullableString(input.RefType), nullableString(input.RefID),
		nullableFloat(input.RealizedGL), input.Status, nullableString(input.CreatedBy),
		nullableString(input.Notes))
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("insert failed: %v", err))
		return
	}
	respondJSON(w, http.StatusCreated, item)
}

func (h *RetentionHandler) GetMCTx(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var item MCTx
	if err := h.db.Get(&item, "SELECT * FROM multi_currency_transactions WHERE id=$1", id); err != nil {
		respondError(w, http.StatusNotFound, "transaction not found")
		return
	}
	respondJSON(w, http.StatusOK, item)
}

func (h *RetentionHandler) ListMCTxByProject(w http.ResponseWriter, r *http.Request) {
	pid := chi.URLParam(r, "projectId")
	var items []MCTx
	h.db.Select(&items, "SELECT * FROM multi_currency_transactions WHERE project_id=$1 ORDER BY created_at DESC", pid)
	respondJSON(w, http.StatusOK, items)
}

// ---- Helpers ----

func nullableFloat(f *float64) interface{} {
	if f == nil {
		return nil
	}
	return *f
}

func nullableInt(i *int) interface{} {
	if i == nil {
		return nil
	}
	return *i
}

func nullableString(s *string) interface{} {
	if s == nil {
		return nil
	}
	return *s
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, msg string) {
	log.Printf("[RetentionHandler] Error %d: %s", status, msg)
	respondJSON(w, status, map[string]string{"error": msg})
}

func init() {
	// Ensure time import is used
	_ = time.Now
}