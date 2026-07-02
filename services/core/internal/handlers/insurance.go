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

// InsuranceHandler handles Insurance module endpoints (V031)
type InsuranceHandler struct {
	db *sql.DB
}

func NewInsuranceHandler(db *sql.DB) *InsuranceHandler {
	return &InsuranceHandler{db: db}
}

func (h *InsuranceHandler) RegisterRoutes(r chi.Router) {
	r.Route("/insurance", func(r chi.Router) {
		// Brokers
		r.Get("/brokers", h.ListBrokers)
		r.Post("/brokers", h.CreateBroker)
		r.Get("/brokers/{id}", h.GetBroker)

		// Policies
		r.Get("/policies", h.ListPolicies)
		r.Post("/policies", h.CreatePolicy)
		r.Get("/policies/{id}", h.GetPolicy)
		r.Put("/policies/{id}", h.UpdatePolicy)

		// Coverage
		r.Get("/policies/{id}/coverage", h.ListCoverage)
		r.Post("/policies/{id}/coverage", h.CreateCoverage)

		// Premiums
		r.Get("/policies/{id}/premiums", h.ListPremiums)
		r.Post("/policies/{id}/premiums", h.CreatePremium)

		// Claims
		r.Get("/claims", h.ListClaims)
		r.Post("/claims", h.CreateClaim)
		r.Get("/claims/{id}", h.GetClaim)
		r.Put("/claims/{id}", h.UpdateClaim)

		// Certificates
		r.Get("/policies/{id}/certificates", h.ListCertificates)
		r.Post("/policies/{id}/certificates", h.CreateCertificate)
	})
}

// --- Brokers ---

func (h *InsuranceHandler) ListBrokers(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(`SELECT id,broker_name,contact_person,email,phone,address,license_number,notes,is_active,created_at,updated_at FROM insurance_brokers ORDER BY broker_name`)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	items := make([]models.InsuranceBroker, 0)
	for rows.Next() {
		var m models.InsuranceBroker
		if err := rows.Scan(&m.ID, &m.BrokerName, &m.ContactPerson, &m.Email, &m.Phone, &m.Address, &m.LicenseNumber, &m.Notes, &m.IsActive, &m.CreatedAt, &m.UpdatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, m)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *InsuranceHandler) CreateBroker(w http.ResponseWriter, r *http.Request) {
	var input struct {
		BrokerName    string  `json:"broker_name"`
		ContactPerson *string `json:"contact_person"`
		Email         *string `json:"email"`
		Phone         *string `json:"phone"`
		Address       *string `json:"address"`
		LicenseNumber *string `json:"license_number"`
		Notes         *string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO insurance_brokers (id,broker_name,contact_person,email,phone,address,license_number,notes,created_at,updated_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		id, input.BrokerName, input.ContactPerson, input.Email, input.Phone, input.Address, input.LicenseNumber, input.Notes, now, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *InsuranceHandler) GetBroker(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var m models.InsuranceBroker
	err := h.db.QueryRow(`SELECT id,broker_name,contact_person,email,phone,address,license_number,notes,is_active,created_at,updated_at FROM insurance_brokers WHERE id=$1`, id).Scan(
		&m.ID, &m.BrokerName, &m.ContactPerson, &m.Email, &m.Phone, &m.Address, &m.LicenseNumber, &m.Notes, &m.IsActive, &m.CreatedAt, &m.UpdatedAt)
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

// --- Policies ---

func (h *InsuranceHandler) ListPolicies(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT id,project_id,policy_number,policy_type,insurer,broker_id,insured_party,sum_insured,currency,premium_amount,deductible,excess,start_date,end_date,renewal_date,territory,status,description,is_active,created_at,updated_at FROM insurance_policies`
	var rows *sql.Rows
	var err error
	if projectID != "" {
		rows, err = h.db.Query(query+` WHERE project_id=$1 ORDER BY created_at DESC`, projectID)
	} else {
		rows, err = h.db.Query(query + ` ORDER BY created_at DESC`)
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	items := make([]models.InsurancePolicy, 0)
	for rows.Next() {
		var m models.InsurancePolicy
		if err := rows.Scan(&m.ID, &m.ProjectID, &m.PolicyNumber, &m.PolicyType, &m.Insurer, &m.BrokerID, &m.InsuredParty, &m.SumInsured, &m.Currency, &m.PremiumAmount, &m.Deductible, &m.Excess, &m.StartDate, &m.EndDate, &m.RenewalDate, &m.Territory, &m.Status, &m.Description, &m.IsActive, &m.CreatedAt, &m.UpdatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, m)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *InsuranceHandler) CreatePolicy(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID    *string  `json:"project_id"`
		PolicyNumber string   `json:"policy_number"`
		PolicyType   string   `json:"policy_type"`
		Insurer      string   `json:"insurer"`
		BrokerID     *string  `json:"broker_id"`
		InsuredParty *string  `json:"insured_party"`
		SumInsured   float64  `json:"sum_insured"`
		Currency     string   `json:"currency"`
		PremiumAmount *float64 `json:"premium_amount"`
		Deductible   *float64 `json:"deductible"`
		StartDate    string   `json:"start_date"`
		EndDate      string   `json:"end_date"`
		Status       *string  `json:"status"`
		Description  *string  `json:"description"`
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
	_, err := h.db.Exec(`INSERT INTO insurance_policies (id,project_id,policy_number,policy_type,insurer,broker_id,insured_party,sum_insured,currency,premium_amount,deductible,start_date,end_date,status,description,created_at,updated_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)`,
		id, input.ProjectID, input.PolicyNumber, input.PolicyType, input.Insurer, input.BrokerID, input.InsuredParty, input.SumInsured, input.Currency, input.PremiumAmount, input.Deductible, input.StartDate, input.EndDate, status, input.Description, now, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *InsuranceHandler) GetPolicy(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var m models.InsurancePolicy
	err := h.db.QueryRow(`SELECT id,project_id,policy_number,policy_type,insurer,broker_id,insured_party,sum_insured,currency,premium_amount,deductible,excess,start_date,end_date,renewal_date,territory,status,description,is_active,created_at,updated_at FROM insurance_policies WHERE id=$1`, id).Scan(
		&m.ID, &m.ProjectID, &m.PolicyNumber, &m.PolicyType, &m.Insurer, &m.BrokerID, &m.InsuredParty, &m.SumInsured, &m.Currency, &m.PremiumAmount, &m.Deductible, &m.Excess, &m.StartDate, &m.EndDate, &m.RenewalDate, &m.Territory, &m.Status, &m.Description, &m.IsActive, &m.CreatedAt, &m.UpdatedAt)
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

func (h *InsuranceHandler) UpdatePolicy(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Status      *string `json:"status"`
		EndDate     *string `json:"end_date"`
		RenewalDate *string `json:"renewal_date"`
		Description *string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	_, err := h.db.Exec(`UPDATE insurance_policies SET status=COALESCE($1,status), end_date=COALESCE($2,end_date), renewal_date=COALESCE($3,renewal_date), description=COALESCE($4,description), updated_at=NOW() WHERE id=$5`,
		input.Status, input.EndDate, input.RenewalDate, input.Description, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// --- Coverage ---

func (h *InsuranceHandler) ListCoverage(w http.ResponseWriter, r *http.Request) {
	policyID := chi.URLParam(r, "id")
	rows, err := h.db.Query(`SELECT id,policy_id,coverage_type,coverage_limit,currency,deductible,sublimit,description,is_active,created_at FROM insurance_coverage WHERE policy_id=$1`, policyID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	items := make([]models.InsuranceCoverage, 0)
	for rows.Next() {
		var m models.InsuranceCoverage
		if err := rows.Scan(&m.ID, &m.PolicyID, &m.CoverageType, &m.CoverageLimit, &m.Currency, &m.Deductible, &m.Sublimit, &m.Description, &m.IsActive, &m.CreatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, m)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *InsuranceHandler) CreateCoverage(w http.ResponseWriter, r *http.Request) {
	policyID := chi.URLParam(r, "id")
	var input struct {
		CoverageType  string   `json:"coverage_type"`
		CoverageLimit *float64 `json:"coverage_limit"`
		Currency      *string  `json:"currency"`
		Deductible    *float64 `json:"deductible"`
		Description   *string  `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	currency := "USD"
	if input.Currency != nil {
		currency = *input.Currency
	}
	_, err := h.db.Exec(`INSERT INTO insurance_coverage (id,policy_id,coverage_type,coverage_limit,currency,deductible,description,created_at) VALUES($1,$2,$3,$4,$5,$6,$7,NOW())`,
		id, policyID, input.CoverageType, input.CoverageLimit, currency, input.Deductible, input.Description)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

// --- Premiums ---

func (h *InsuranceHandler) ListPremiums(w http.ResponseWriter, r *http.Request) {
	policyID := chi.URLParam(r, "id")
	rows, err := h.db.Query(`SELECT id,policy_id,premium_number,amount,currency,due_date,paid_date,payment_method,status,notes,created_at,updated_at FROM insurance_premiums WHERE policy_id=$1 ORDER BY due_date`, policyID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	items := make([]models.InsurancePremium, 0)
	for rows.Next() {
		var m models.InsurancePremium
		if err := rows.Scan(&m.ID, &m.PolicyID, &m.PremiumNumber, &m.Amount, &m.Currency, &m.DueDate, &m.PaidDate, &m.PaymentMethod, &m.Status, &m.Notes, &m.CreatedAt, &m.UpdatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, m)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *InsuranceHandler) CreatePremium(w http.ResponseWriter, r *http.Request) {
	policyID := chi.URLParam(r, "id")
	var input struct {
		PremiumNumber *string  `json:"premium_number"`
		Amount        float64  `json:"amount"`
		Currency      *string  `json:"currency"`
		DueDate       *string  `json:"due_date"`
		Notes         *string  `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	now := time.Now()
	currency := "USD"
	if input.Currency != nil {
		currency = *input.Currency
	}
	_, err := h.db.Exec(`INSERT INTO insurance_premiums (id,policy_id,premium_number,amount,currency,due_date,status,notes,created_at,updated_at) VALUES($1,$2,$3,$4,$5,$6,'pending',$7,$8,$9)`,
		id, policyID, input.PremiumNumber, input.Amount, currency, input.DueDate, input.Notes, now, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

// --- Claims ---

func (h *InsuranceHandler) ListClaims(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT id,project_id,policy_id,claim_number,claim_date,incident_date,incident_type,cause,description,claimed_amount,currency,settled_amount,status,adjuster_name,decision_date,notes,is_active,created_at,updated_at FROM insurance_claims`
	var rows *sql.Rows
	var err error
	if projectID != "" {
		rows, err = h.db.Query(query+` WHERE project_id=$1 ORDER BY created_at DESC`, projectID)
	} else {
		rows, err = h.db.Query(query + ` ORDER BY created_at DESC`)
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	items := make([]models.InsuranceClaim, 0)
	for rows.Next() {
		var m models.InsuranceClaim
		if err := rows.Scan(&m.ID, &m.ProjectID, &m.PolicyID, &m.ClaimNumber, &m.ClaimDate, &m.IncidentDate, &m.IncidentType, &m.Cause, &m.Description, &m.ClaimedAmount, &m.Currency, &m.SettledAmount, &m.Status, &m.AdjusterName, &m.DecisionDate, &m.Notes, &m.IsActive, &m.CreatedAt, &m.UpdatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, m)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *InsuranceHandler) CreateClaim(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID    *string  `json:"project_id"`
		PolicyID     string   `json:"policy_id"`
		ClaimNumber  string   `json:"claim_number"`
		ClaimDate    string   `json:"claim_date"`
		IncidentDate *string  `json:"incident_date"`
		IncidentType *string  `json:"incident_type"`
		Cause        *string  `json:"cause"`
		Description  *string  `json:"description"`
		ClaimedAmount *float64 `json:"claimed_amount"`
		Currency     *string  `json:"currency"`
		Notes        *string  `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	now := time.Now()
	currency := "USD"
	if input.Currency != nil {
		currency = *input.Currency
	}
	_, err := h.db.Exec(`INSERT INTO insurance_claims (id,project_id,policy_id,claim_number,claim_date,incident_date,incident_type,cause,description,claimed_amount,currency,status,notes,created_at,updated_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,'submitted',$12,$13,$14)`,
		id, input.ProjectID, input.PolicyID, input.ClaimNumber, input.ClaimDate, input.IncidentDate, input.IncidentType, input.Cause, input.Description, input.ClaimedAmount, currency, input.Notes, now, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *InsuranceHandler) GetClaim(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var m models.InsuranceClaim
	err := h.db.QueryRow(`SELECT id,project_id,policy_id,claim_number,claim_date,incident_date,incident_type,cause,description,claimed_amount,currency,settled_amount,status,adjuster_name,decision_date,notes,is_active,created_at,updated_at FROM insurance_claims WHERE id=$1`, id).Scan(
		&m.ID, &m.ProjectID, &m.PolicyID, &m.ClaimNumber, &m.ClaimDate, &m.IncidentDate, &m.IncidentType, &m.Cause, &m.Description, &m.ClaimedAmount, &m.Currency, &m.SettledAmount, &m.Status, &m.AdjusterName, &m.DecisionDate, &m.Notes, &m.IsActive, &m.CreatedAt, &m.UpdatedAt)
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

func (h *InsuranceHandler) UpdateClaim(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Status        *string  `json:"status"`
		SettledAmount *float64 `json:"settled_amount"`
		DecisionDate  *string  `json:"decision_date"`
		AdjusterName  *string  `json:"adjuster_name"`
		Notes         *string  `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	_, err := h.db.Exec(`UPDATE insurance_claims SET status=COALESCE($1,status), settled_amount=COALESCE($2,settled_amount), decision_date=COALESCE($3,decision_date), adjuster_name=COALESCE($4,adjuster_name), notes=COALESCE($5,notes), updated_at=NOW() WHERE id=$6`,
		input.Status, input.SettledAmount, input.DecisionDate, input.AdjusterName, input.Notes, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// --- Certificates ---

func (h *InsuranceHandler) ListCertificates(w http.ResponseWriter, r *http.Request) {
	policyID := chi.URLParam(r, "id")
	rows, err := h.db.Query(`SELECT id,policy_id,certificate_number,certificate_holder,issue_date,expiry_date,description,document_url,status,is_active,created_at,updated_at FROM certificates_of_insurance WHERE policy_id=$1 ORDER BY issue_date`, policyID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	items := make([]models.CertificateOfInsurance, 0)
	for rows.Next() {
		var m models.CertificateOfInsurance
		if err := rows.Scan(&m.ID, &m.PolicyID, &m.CertificateNumber, &m.CertificateHolder, &m.IssueDate, &m.ExpiryDate, &m.Description, &m.DocumentURL, &m.Status, &m.IsActive, &m.CreatedAt, &m.UpdatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, m)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *InsuranceHandler) CreateCertificate(w http.ResponseWriter, r *http.Request) {
	policyID := chi.URLParam(r, "id")
	var input struct {
		CertificateNumber string  `json:"certificate_number"`
		CertificateHolder *string `json:"certificate_holder"`
		IssueDate         *string `json:"issue_date"`
		ExpiryDate        *string `json:"expiry_date"`
		Description       *string `json:"description"`
		DocumentURL       *string `json:"document_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO certificates_of_insurance (id,policy_id,certificate_number,certificate_holder,issue_date,expiry_date,description,document_url,created_at,updated_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		id, policyID, input.CertificateNumber, input.CertificateHolder, input.IssueDate, input.ExpiryDate, input.Description, input.DocumentURL, now, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}