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

// PermitsHandler handles Permits module endpoints (V030)
type PermitsHandler struct {
	db *sql.DB
}

func NewPermitsHandler(db *sql.DB) *PermitsHandler {
	return &PermitsHandler{db: db}
}

func (h *PermitsHandler) RegisterRoutes(r chi.Router) {
	r.Route("/permits", func(r chi.Router) {
		// Regulatory Bodies
		r.Get("/bodies", h.ListBodies)
		r.Post("/bodies", h.CreateBody)
		r.Get("/bodies/{id}", h.GetBody)

		// Applications
		r.Get("/applications", h.ListApplications)
		r.Post("/applications", h.CreateApplication)
		r.Get("/applications/{id}", h.GetApplication)
		r.Put("/applications/{id}", h.UpdateApplication)

		// Documents
		r.Get("/applications/{id}/documents", h.ListDocuments)
		r.Post("/applications/{id}/documents", h.CreateDocument)

		// Inspections
		r.Get("/applications/{id}/inspections", h.ListInspections)
		r.Post("/applications/{id}/inspections", h.CreateInspection)
		r.Put("/inspections/{id}", h.UpdateInspection)

		// Renewals
		r.Get("/applications/{id}/renewals", h.ListRenewals)
		r.Post("/applications/{id}/renewals", h.CreateRenewal)

		// Conditions
		r.Get("/applications/{id}/conditions", h.ListConditions)
		r.Post("/applications/{id}/conditions", h.CreateCondition)
		r.Put("/conditions/{id}", h.UpdateCondition)
	})
}

// --- Regulatory Bodies ---

func (h *PermitsHandler) ListBodies(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(`SELECT id,body_name,body_code,jurisdiction,contact_info,website,notes,is_active,created_at,updated_at FROM regulatory_bodies ORDER BY body_name`)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	items := make([]models.RegulatoryBody, 0)
	for rows.Next() {
		var m models.RegulatoryBody
		if err := rows.Scan(&m.ID, &m.BodyName, &m.BodyCode, &m.Jurisdiction, &m.ContactInfo, &m.Website, &m.Notes, &m.IsActive, &m.CreatedAt, &m.UpdatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, m)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *PermitsHandler) CreateBody(w http.ResponseWriter, r *http.Request) {
	var input struct {
		BodyName    string  `json:"body_name"`
		BodyCode    *string `json:"body_code"`
		Jurisdiction *string `json:"jurisdiction"`
		ContactInfo *string `json:"contact_info"`
		Website     *string `json:"website"`
		Notes       *string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO regulatory_bodies (id,body_name,body_code,jurisdiction,contact_info,website,notes,created_at,updated_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		id, input.BodyName, input.BodyCode, input.Jurisdiction, input.ContactInfo, input.Website, input.Notes, now, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *PermitsHandler) GetBody(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var m models.RegulatoryBody
	err := h.db.QueryRow(`SELECT id,body_name,body_code,jurisdiction,contact_info,website,notes,is_active,created_at,updated_at FROM regulatory_bodies WHERE id=$1`, id).Scan(
		&m.ID, &m.BodyName, &m.BodyCode, &m.Jurisdiction, &m.ContactInfo, &m.Website, &m.Notes, &m.IsActive, &m.CreatedAt, &m.UpdatedAt)
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

// --- Applications ---

func (h *PermitsHandler) ListApplications(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT id,project_id,regulatory_body_id,permit_number,permit_type,description,application_date,decision_date,status,approved_by,expiry_date,notes,is_active,created_at,updated_at FROM permit_applications`
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
	items := make([]models.PermitApplication, 0)
	for rows.Next() {
		var m models.PermitApplication
		if err := rows.Scan(&m.ID, &m.ProjectID, &m.RegulatoryBodyID, &m.PermitNumber, &m.PermitType, &m.Description, &m.ApplicationDate, &m.DecisionDate, &m.Status, &m.ApprovedBy, &m.ExpiryDate, &m.Notes, &m.IsActive, &m.CreatedAt, &m.UpdatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, m)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *PermitsHandler) CreateApplication(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID        string  `json:"project_id"`
		RegulatoryBodyID *string `json:"regulatory_body_id"`
		PermitNumber     *string `json:"permit_number"`
		PermitType       string  `json:"permit_type"`
		Description      *string `json:"description"`
		ApplicationDate  *string `json:"application_date"`
		Status           *string `json:"status"`
		Notes            *string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	now := time.Now()
	status := "draft"
	if input.Status != nil {
		status = *input.Status
	}
	_, err := h.db.Exec(`INSERT INTO permit_applications (id,project_id,regulatory_body_id,permit_number,permit_type,description,application_date,status,notes,created_at,updated_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		id, input.ProjectID, input.RegulatoryBodyID, input.PermitNumber, input.PermitType, input.Description, input.ApplicationDate, status, input.Notes, now, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *PermitsHandler) GetApplication(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var m models.PermitApplication
	err := h.db.QueryRow(`SELECT id,project_id,regulatory_body_id,permit_number,permit_type,description,application_date,decision_date,status,approved_by,expiry_date,notes,is_active,created_at,updated_at FROM permit_applications WHERE id=$1`, id).Scan(
		&m.ID, &m.ProjectID, &m.RegulatoryBodyID, &m.PermitNumber, &m.PermitType, &m.Description, &m.ApplicationDate, &m.DecisionDate, &m.Status, &m.ApprovedBy, &m.ExpiryDate, &m.Notes, &m.IsActive, &m.CreatedAt, &m.UpdatedAt)
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

func (h *PermitsHandler) UpdateApplication(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Status      *string `json:"status"`
		DecisionDate *string `json:"decision_date"`
		ApprovedBy  *string `json:"approved_by"`
		ExpiryDate  *string `json:"expiry_date"`
		Notes       *string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	_, err := h.db.Exec(`UPDATE permit_applications SET status=COALESCE($1,status), decision_date=COALESCE($2,decision_date), approved_by=COALESCE($3,approved_by), expiry_date=COALESCE($4,expiry_date), notes=COALESCE($5,notes), updated_at=NOW() WHERE id=$6`,
		input.Status, input.DecisionDate, input.ApprovedBy, input.ExpiryDate, input.Notes, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// --- Documents ---

func (h *PermitsHandler) ListDocuments(w http.ResponseWriter, r *http.Request) {
	appID := chi.URLParam(r, "id")
	rows, err := h.db.Query(`SELECT id,permit_application_id,document_type,document_name,document_url,version,submitted_date,status,notes,created_at,updated_at FROM permit_documents WHERE permit_application_id=$1 ORDER BY created_at`, appID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	items := make([]models.PermitDocument, 0)
	for rows.Next() {
		var m models.PermitDocument
		if err := rows.Scan(&m.ID, &m.PermitApplicationID, &m.DocumentType, &m.DocumentName, &m.DocumentURL, &m.Version, &m.SubmittedDate, &m.Status, &m.Notes, &m.CreatedAt, &m.UpdatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, m)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *PermitsHandler) CreateDocument(w http.ResponseWriter, r *http.Request) {
	appID := chi.URLParam(r, "id")
	var input struct {
		DocumentType  string  `json:"document_type"`
		DocumentName  *string `json:"document_name"`
		DocumentURL   *string `json:"document_url"`
		Version       *string `json:"version"`
		SubmittedDate *string `json:"submitted_date"`
		Notes         *string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO permit_documents (id,permit_application_id,document_type,document_name,document_url,version,submitted_date,notes,created_at,updated_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		id, appID, input.DocumentType, input.DocumentName, input.DocumentURL, input.Version, input.SubmittedDate, input.Notes, now, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

// --- Inspections ---

func (h *PermitsHandler) ListInspections(w http.ResponseWriter, r *http.Request) {
	appID := chi.URLParam(r, "id")
	rows, err := h.db.Query(`SELECT id,permit_application_id,inspection_type,inspection_date,inspector_name,inspector_agency,result,findings,corrective_actions,scheduled_date,completed_date,status,is_active,created_at,updated_at FROM permit_inspections WHERE permit_application_id=$1 ORDER BY scheduled_date`, appID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	items := make([]models.PermitInspection, 0)
	for rows.Next() {
		var m models.PermitInspection
		if err := rows.Scan(&m.ID, &m.PermitApplicationID, &m.InspectionType, &m.InspectionDate, &m.InspectorName, &m.InspectorAgency, &m.Result, &m.Findings, &m.CorrectiveActions, &m.ScheduledDate, &m.CompletedDate, &m.Status, &m.IsActive, &m.CreatedAt, &m.UpdatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, m)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *PermitsHandler) CreateInspection(w http.ResponseWriter, r *http.Request) {
	appID := chi.URLParam(r, "id")
	var input struct {
		InspectionType  string  `json:"inspection_type"`
		InspectorName   *string `json:"inspector_name"`
		InspectorAgency *string `json:"inspector_agency"`
		ScheduledDate   *string `json:"scheduled_date"`
		Notes           *string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO permit_inspections (id,permit_application_id,inspection_type,inspector_name,inspector_agency,scheduled_date,status,created_at,updated_at) VALUES($1,$2,$3,$4,$5,$6,'scheduled',$7,$8)`,
		id, appID, input.InspectionType, input.InspectorName, input.InspectorAgency, input.ScheduledDate, now, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *PermitsHandler) UpdateInspection(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Result            *string `json:"result"`
		Findings          *string `json:"findings"`
		CorrectiveActions *string `json:"corrective_actions"`
		CompletedDate     *string `json:"completed_date"`
		Status            *string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	_, err := h.db.Exec(`UPDATE permit_inspections SET result=COALESCE($1,result), findings=COALESCE($2,findings), corrective_actions=COALESCE($3,corrective_actions), completed_date=COALESCE($4,completed_date), status=COALESCE($5,status), updated_at=NOW() WHERE id=$6`,
		input.Result, input.Findings, input.CorrectiveActions, input.CompletedDate, input.Status, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// --- Renewals ---

func (h *PermitsHandler) ListRenewals(w http.ResponseWriter, r *http.Request) {
	appID := chi.URLParam(r, "id")
	rows, err := h.db.Query(`SELECT id,permit_application_id,renewal_number,renewal_date,expiry_date,fee_amount,fee_currency,status,notes,created_at,updated_at FROM permit_renewals WHERE permit_application_id=$1 ORDER BY created_at`, appID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	items := make([]models.PermitRenewal, 0)
	for rows.Next() {
		var m models.PermitRenewal
		if err := rows.Scan(&m.ID, &m.PermitApplicationID, &m.RenewalNumber, &m.RenewalDate, &m.ExpiryDate, &m.FeeAmount, &m.FeeCurrency, &m.Status, &m.Notes, &m.CreatedAt, &m.UpdatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, m)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *PermitsHandler) CreateRenewal(w http.ResponseWriter, r *http.Request) {
	appID := chi.URLParam(r, "id")
	var input struct {
		RenewalNumber *string  `json:"renewal_number"`
		RenewalDate   *string  `json:"renewal_date"`
		ExpiryDate    *string  `json:"expiry_date"`
		FeeAmount     *float64 `json:"fee_amount"`
		FeeCurrency   *string  `json:"fee_currency"`
		Notes         *string  `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	now := time.Now()
	currency := "USD"
	if input.FeeCurrency != nil {
		currency = *input.FeeCurrency
	}
	_, err := h.db.Exec(`INSERT INTO permit_renewals (id,permit_application_id,renewal_number,renewal_date,expiry_date,fee_amount,fee_currency,status,created_at,updated_at) VALUES($1,$2,$3,$4,$5,$6,$7,'pending',$8,$9)`,
		id, appID, input.RenewalNumber, input.RenewalDate, input.ExpiryDate, input.FeeAmount, currency, now, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

// --- Conditions ---

func (h *PermitsHandler) ListConditions(w http.ResponseWriter, r *http.Request) {
	appID := chi.URLParam(r, "id")
	rows, err := h.db.Query(`SELECT id,permit_application_id,condition_number,description,condition_type,due_date,status,satisfied_date,verified_by,created_at,updated_at FROM permit_conditions WHERE permit_application_id=$1 ORDER BY due_date`, appID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	items := make([]models.PermitCondition, 0)
	for rows.Next() {
		var m models.PermitCondition
		if err := rows.Scan(&m.ID, &m.PermitApplicationID, &m.ConditionNumber, &m.Description, &m.ConditionType, &m.DueDate, &m.Status, &m.SatisfiedDate, &m.VerifiedBy, &m.CreatedAt, &m.UpdatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, m)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *PermitsHandler) CreateCondition(w http.ResponseWriter, r *http.Request) {
	appID := chi.URLParam(r, "id")
	var input struct {
		ConditionNumber *string `json:"condition_number"`
		Description     string  `json:"description"`
		ConditionType   *string `json:"condition_type"`
		DueDate         *string `json:"due_date"`
		VerifiedBy      *string `json:"verified_by"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO permit_conditions (id,permit_application_id,condition_number,description,condition_type,due_date,status,verified_by,created_at,updated_at) VALUES($1,$2,$3,$4,$5,$6,'pending',$7,$8,$9)`,
		id, appID, input.ConditionNumber, input.Description, input.ConditionType, input.DueDate, input.VerifiedBy, now, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *PermitsHandler) UpdateCondition(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Status        *string `json:"status"`
		SatisfiedDate *string `json:"satisfied_date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	_, err := h.db.Exec(`UPDATE permit_conditions SET status=COALESCE($1,status), satisfied_date=COALESCE($2,satisfied_date), updated_at=NOW() WHERE id=$3`,
		input.Status, input.SatisfiedDate, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}