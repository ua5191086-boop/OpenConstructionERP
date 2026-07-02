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

// LaboratoryHandler handles Laboratory module endpoints (V029)
type LaboratoryHandler struct {
	db *sql.DB
}

func NewLaboratoryHandler(db *sql.DB) *LaboratoryHandler {
	return &LaboratoryHandler{db: db}
}

func (h *LaboratoryHandler) RegisterRoutes(r chi.Router) {
	r.Route("/lab", func(r chi.Router) {
		// Material Tests
		r.Get("/tests", h.ListMaterialTests)
		r.Post("/tests", h.CreateMaterialTest)
		r.Get("/tests/{id}", h.GetMaterialTest)
		r.Put("/tests/{id}", h.UpdateMaterialTest)

		// Concrete Tests
		r.Get("/concrete-tests", h.ListConcreteTests)
		r.Post("/concrete-tests", h.CreateConcreteTest)
		r.Get("/concrete-tests/{id}", h.GetConcreteTest)

		// Soil Tests
		r.Get("/soil-tests", h.ListSoilTests)
		r.Post("/soil-tests", h.CreateSoilTest)
		r.Get("/soil-tests/{id}", h.GetSoilTest)

		// Steel Tests
		r.Get("/steel-tests", h.ListSteelTests)
		r.Post("/steel-tests", h.CreateSteelTest)
		r.Get("/steel-tests/{id}", h.GetSteelTest)

		// Lab Certificates
		r.Get("/certificates", h.ListCertificates)
		r.Post("/certificates", h.CreateCertificate)

		// Lab Equipment
		r.Get("/equipment", h.ListLabEquipment)
		r.Post("/equipment", h.CreateLabEquipment)
		r.Put("/equipment/{id}", h.UpdateLabEquipment)

		// Sampling Log
		r.Get("/samples", h.ListSamples)
		r.Post("/samples", h.CreateSample)
	})
}

// --- Material Tests ---

func (h *LaboratoryHandler) ListMaterialTests(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT id,project_id,test_number,material_type,test_type,specification,sample_id,sampling_date,test_date,result,status,tested_by,approved_by,notes,is_active,created_at,updated_at FROM material_testing`
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
	items := make([]models.MaterialTest, 0)
	for rows.Next() {
		var m models.MaterialTest
		if err := rows.Scan(&m.ID, &m.ProjectID, &m.TestNumber, &m.MaterialType, &m.TestType, &m.Specification, &m.SampleID, &m.SamplingDate, &m.TestDate, &m.Result, &m.Status, &m.TestedBy, &m.ApprovedBy, &m.Notes, &m.IsActive, &m.CreatedAt, &m.UpdatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, m)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *LaboratoryHandler) CreateMaterialTest(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID     string  `json:"project_id"`
		TestNumber    string  `json:"test_number"`
		MaterialType  string  `json:"material_type"`
		TestType      string  `json:"test_type"`
		Specification *string `json:"specification"`
		SampleID      *string `json:"sample_id"`
		SamplingDate  *string `json:"sampling_date"`
		TestDate      *string `json:"test_date"`
		Status        *string `json:"status"`
		TestedBy      *string `json:"tested_by"`
		Notes         *string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	now := time.Now()
	status := "pending"
	if input.Status != nil {
		status = *input.Status
	}
	_, err := h.db.Exec(`INSERT INTO material_testing (id,project_id,test_number,material_type,test_type,specification,sample_id,sampling_date,test_date,status,tested_by,notes,created_at,updated_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		id, input.ProjectID, input.TestNumber, input.MaterialType, input.TestType, input.Specification, input.SampleID, input.SamplingDate, input.TestDate, status, input.TestedBy, input.Notes, now, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *LaboratoryHandler) GetMaterialTest(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var m models.MaterialTest
	err := h.db.QueryRow(`SELECT id,project_id,test_number,material_type,test_type,specification,sample_id,sampling_date,test_date,result,status,tested_by,approved_by,notes,is_active,created_at,updated_at FROM material_testing WHERE id=$1`, id).Scan(
		&m.ID, &m.ProjectID, &m.TestNumber, &m.MaterialType, &m.TestType, &m.Specification, &m.SampleID, &m.SamplingDate, &m.TestDate, &m.Result, &m.Status, &m.TestedBy, &m.ApprovedBy, &m.Notes, &m.IsActive, &m.CreatedAt, &m.UpdatedAt)
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

func (h *LaboratoryHandler) UpdateMaterialTest(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Status   *string `json:"status"`
		Result   *string `json:"result"`
		TestedBy *string `json:"tested_by"`
		ApprovedBy *string `json:"approved_by"`
		Notes    *string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	_, err := h.db.Exec(`UPDATE material_testing SET status=COALESCE($1,status), result=COALESCE($2,result), tested_by=COALESCE($3,tested_by), approved_by=COALESCE($4,approved_by), notes=COALESCE($5,notes), updated_at=NOW() WHERE id=$6`,
		input.Status, input.Result, input.TestedBy, input.ApprovedBy, input.Notes, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// --- Concrete Tests ---

func (h *LaboratoryHandler) ListConcreteTests(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT id,project_id,material_test_id,sample_id,concrete_grade,slump,compressive_strength_7d,compressive_strength_14d,compressive_strength_28d,flexural_strength,air_content,temperature,unit_weight,curing_method,test_date,result,tested_by,notes,created_at FROM concrete_tests`
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
	items := make([]models.ConcreteTest, 0)
	for rows.Next() {
		var m models.ConcreteTest
		if err := rows.Scan(&m.ID, &m.ProjectID, &m.MaterialTestID, &m.SampleID, &m.ConcreteGrade, &m.Slump, &m.CompressiveStrength7d, &m.CompressiveStrength14d, &m.CompressiveStrength28d, &m.FlexuralStrength, &m.AirContent, &m.Temperature, &m.UnitWeight, &m.CuringMethod, &m.TestDate, &m.Result, &m.TestedBy, &m.Notes, &m.CreatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, m)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *LaboratoryHandler) CreateConcreteTest(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID             string   `json:"project_id"`
		MaterialTestID        *string  `json:"material_test_id"`
		SampleID              *string  `json:"sample_id"`
		ConcreteGrade         *string  `json:"concrete_grade"`
		Slump                 *float64 `json:"slump"`
		CompressiveStrength7d *float64 `json:"compressive_strength_7d"`
		TestDate              *string  `json:"test_date"`
		TestedBy              *string  `json:"tested_by"`
		Notes                 *string  `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	_, err := h.db.Exec(`INSERT INTO concrete_tests (id,project_id,material_test_id,sample_id,concrete_grade,slump,compressive_strength_7d,test_date,tested_by,notes,created_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,NOW())`,
		id, input.ProjectID, input.MaterialTestID, input.SampleID, input.ConcreteGrade, input.Slump, input.CompressiveStrength7d, input.TestDate, input.TestedBy, input.Notes)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *LaboratoryHandler) GetConcreteTest(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var m models.ConcreteTest
	err := h.db.QueryRow(`SELECT id,project_id,material_test_id,sample_id,concrete_grade,slump,compressive_strength_7d,compressive_strength_14d,compressive_strength_28d,flexural_strength,air_content,temperature,unit_weight,curing_method,test_date,result,tested_by,notes,created_at FROM concrete_tests WHERE id=$1`, id).Scan(
		&m.ID, &m.ProjectID, &m.MaterialTestID, &m.SampleID, &m.ConcreteGrade, &m.Slump, &m.CompressiveStrength7d, &m.CompressiveStrength14d, &m.CompressiveStrength28d, &m.FlexuralStrength, &m.AirContent, &m.Temperature, &m.UnitWeight, &m.CuringMethod, &m.TestDate, &m.Result, &m.TestedBy, &m.Notes, &m.CreatedAt)
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

// --- Soil Tests ---

func (h *LaboratoryHandler) ListSoilTests(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT id,project_id,material_test_id,sample_id,soil_type,moisture_content,dry_density,atterberg_limit_liquid,atterberg_limit_plastic,plasticity_index,compaction_pct,cbr_value,shear_strength,test_date,result,tested_by,notes,created_at FROM soil_tests`
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
	items := make([]models.SoilTest, 0)
	for rows.Next() {
		var m models.SoilTest
		if err := rows.Scan(&m.ID, &m.ProjectID, &m.MaterialTestID, &m.SampleID, &m.SoilType, &m.MoistureContent, &m.DryDensity, &m.AtterbergLimitLiquid, &m.AtterbergLimitPlastic, &m.PlasticityIndex, &m.CompactionPct, &m.CbrValue, &m.ShearStrength, &m.TestDate, &m.Result, &m.TestedBy, &m.Notes, &m.CreatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, m)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *LaboratoryHandler) CreateSoilTest(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID      string   `json:"project_id"`
		MaterialTestID *string  `json:"material_test_id"`
		SampleID       *string  `json:"sample_id"`
		SoilType       *string  `json:"soil_type"`
		MoistureContent *float64 `json:"moisture_content"`
		TestDate       *string  `json:"test_date"`
		TestedBy       *string  `json:"tested_by"`
		Notes          *string  `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	_, err := h.db.Exec(`INSERT INTO soil_tests (id,project_id,material_test_id,sample_id,soil_type,moisture_content,test_date,tested_by,notes,created_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,NOW())`,
		id, input.ProjectID, input.MaterialTestID, input.SampleID, input.SoilType, input.MoistureContent, input.TestDate, input.TestedBy, input.Notes)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *LaboratoryHandler) GetSoilTest(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var m models.SoilTest
	err := h.db.QueryRow(`SELECT id,project_id,material_test_id,sample_id,soil_type,moisture_content,dry_density,atterberg_limit_liquid,atterberg_limit_plastic,plasticity_index,compaction_pct,cbr_value,shear_strength,test_date,result,tested_by,notes,created_at FROM soil_tests WHERE id=$1`, id).Scan(
		&m.ID, &m.ProjectID, &m.MaterialTestID, &m.SampleID, &m.SoilType, &m.MoistureContent, &m.DryDensity, &m.AtterbergLimitLiquid, &m.AtterbergLimitPlastic, &m.PlasticityIndex, &m.CompactionPct, &m.CbrValue, &m.ShearStrength, &m.TestDate, &m.Result, &m.TestedBy, &m.Notes, &m.CreatedAt)
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

// --- Steel Tests ---

func (h *LaboratoryHandler) ListSteelTests(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT id,project_id,material_test_id,sample_id,steel_grade,diameter_mm,yield_strength,tensile_strength,elongation_pct,bend_test_result,weld_test_result,test_date,result,tested_by,notes,created_at FROM steel_tests`
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
	items := make([]models.SteelTest, 0)
	for rows.Next() {
		var m models.SteelTest
		if err := rows.Scan(&m.ID, &m.ProjectID, &m.MaterialTestID, &m.SampleID, &m.SteelGrade, &m.DiameterMm, &m.YieldStrength, &m.TensileStrength, &m.ElongationPct, &m.BendTestResult, &m.WeldTestResult, &m.TestDate, &m.Result, &m.TestedBy, &m.Notes, &m.CreatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, m)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *LaboratoryHandler) CreateSteelTest(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID      string   `json:"project_id"`
		MaterialTestID *string  `json:"material_test_id"`
		SampleID       *string  `json:"sample_id"`
		SteelGrade     *string  `json:"steel_grade"`
		DiameterMm     *float64 `json:"diameter_mm"`
		YieldStrength  *float64 `json:"yield_strength"`
		TestDate       *string  `json:"test_date"`
		TestedBy       *string  `json:"tested_by"`
		Notes          *string  `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	_, err := h.db.Exec(`INSERT INTO steel_tests (id,project_id,material_test_id,sample_id,steel_grade,diameter_mm,yield_strength,test_date,tested_by,notes,created_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,NOW())`,
		id, input.ProjectID, input.MaterialTestID, input.SampleID, input.SteelGrade, input.DiameterMm, input.YieldStrength, input.TestDate, input.TestedBy, input.Notes)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *LaboratoryHandler) GetSteelTest(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var m models.SteelTest
	err := h.db.QueryRow(`SELECT id,project_id,material_test_id,sample_id,steel_grade,diameter_mm,yield_strength,tensile_strength,elongation_pct,bend_test_result,weld_test_result,test_date,result,tested_by,notes,created_at FROM steel_tests WHERE id=$1`, id).Scan(
		&m.ID, &m.ProjectID, &m.MaterialTestID, &m.SampleID, &m.SteelGrade, &m.DiameterMm, &m.YieldStrength, &m.TensileStrength, &m.ElongationPct, &m.BendTestResult, &m.WeldTestResult, &m.TestDate, &m.Result, &m.TestedBy, &m.Notes, &m.CreatedAt)
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

// --- Lab Certificates ---

func (h *LaboratoryHandler) ListCertificates(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT id,project_id,certificate_number,certificate_type,issuing_body,issue_date,expiry_date,description,document_url,status,is_active,created_at,updated_at FROM lab_certificates`
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
	items := make([]models.LabCertificate, 0)
	for rows.Next() {
		var m models.LabCertificate
		if err := rows.Scan(&m.ID, &m.ProjectID, &m.CertificateNumber, &m.CertificateType, &m.IssuingBody, &m.IssueDate, &m.ExpiryDate, &m.Description, &m.DocumentURL, &m.Status, &m.IsActive, &m.CreatedAt, &m.UpdatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, m)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *LaboratoryHandler) CreateCertificate(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID         string  `json:"project_id"`
		CertificateNumber string  `json:"certificate_number"`
		CertificateType   string  `json:"certificate_type"`
		IssuingBody       *string `json:"issuing_body"`
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
	_, err := h.db.Exec(`INSERT INTO lab_certificates (id,project_id,certificate_number,certificate_type,issuing_body,issue_date,expiry_date,description,document_url,created_at,updated_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		id, input.ProjectID, input.CertificateNumber, input.CertificateType, input.IssuingBody, input.IssueDate, input.ExpiryDate, input.Description, input.DocumentURL, now, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

// --- Lab Equipment ---

func (h *LaboratoryHandler) ListLabEquipment(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT id,project_id,equipment_code,equipment_name,equipment_type,manufacturer,model,serial_number,calibration_due,status,notes,is_active,created_at,updated_at FROM lab_equipment`
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
	items := make([]models.LabEquipment, 0)
	for rows.Next() {
		var m models.LabEquipment
		if err := rows.Scan(&m.ID, &m.ProjectID, &m.EquipmentCode, &m.EquipmentName, &m.EquipmentType, &m.Manufacturer, &m.Model, &m.SerialNumber, &m.CalibrationDue, &m.Status, &m.Notes, &m.IsActive, &m.CreatedAt, &m.UpdatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, m)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *LaboratoryHandler) CreateLabEquipment(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID     string  `json:"project_id"`
		EquipmentCode string  `json:"equipment_code"`
		EquipmentName string  `json:"equipment_name"`
		EquipmentType *string `json:"equipment_type"`
		Manufacturer  *string `json:"manufacturer"`
		Model         *string `json:"model"`
		SerialNumber  *string `json:"serial_number"`
		Status        *string `json:"status"`
		Notes         *string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	now := time.Now()
	status := "operational"
	if input.Status != nil {
		status = *input.Status
	}
	_, err := h.db.Exec(`INSERT INTO lab_equipment (id,project_id,equipment_code,equipment_name,equipment_type,manufacturer,model,serial_number,status,notes,created_at,updated_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		id, input.ProjectID, input.EquipmentCode, input.EquipmentName, input.EquipmentType, input.Manufacturer, input.Model, input.SerialNumber, status, input.Notes, now, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *LaboratoryHandler) UpdateLabEquipment(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Status         *string `json:"status"`
		CalibrationDue *string `json:"calibration_due"`
		Notes          *string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	_, err := h.db.Exec(`UPDATE lab_equipment SET status=COALESCE($1,status), calibration_due=COALESCE($2,calibration_due), notes=COALESCE($3,notes), updated_at=NOW() WHERE id=$4`,
		input.Status, input.CalibrationDue, input.Notes, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// --- Sampling Log ---

func (h *LaboratoryHandler) ListSamples(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT id,project_id,sample_id,sample_type,location,sampling_date,sampled_by,material_test_id,notes,created_at FROM sampling_log`
	var rows *sql.Rows
	var err error
	if projectID != "" {
		rows, err = h.db.Query(query+` WHERE project_id=$1 ORDER BY sampling_date DESC`, projectID)
	} else {
		rows, err = h.db.Query(query + ` ORDER BY sampling_date DESC`)
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	items := make([]models.SamplingLog, 0)
	for rows.Next() {
		var m models.SamplingLog
		if err := rows.Scan(&m.ID, &m.ProjectID, &m.SampleID, &m.SampleType, &m.Location, &m.SamplingDate, &m.SampledBy, &m.MaterialTestID, &m.Notes, &m.CreatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, m)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *LaboratoryHandler) CreateSample(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID    string  `json:"project_id"`
		SampleID     string  `json:"sample_id"`
		SampleType   string  `json:"sample_type"`
		Location     *string `json:"location"`
		SamplingDate string  `json:"sampling_date"`
		SampledBy    *string `json:"sampled_by"`
		Notes        *string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	_, err := h.db.Exec(`INSERT INTO sampling_log (id,project_id,sample_id,sample_type,location,sampling_date,sampled_by,notes,created_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,NOW())`,
		id, input.ProjectID, input.SampleID, input.SampleType, input.Location, input.SamplingDate, input.SampledBy, input.Notes)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}