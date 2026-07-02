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

// EVMHandler handles Earned Value Management endpoints
type EVMHandler struct {
	db *sql.DB
}

func NewEVMHandler(db *sql.DB) *EVMHandler {
	return &EVMHandler{db: db}
}

func (h *EVMHandler) RegisterRoutes(r chi.Router) {
	r.Route("/evm", func(r chi.Router) {
		// EVM Projects
		r.Get("/projects/{id}/summary", h.GetProjectSummary)
		r.Get("/projects/{id}/metrics", h.GetProjectMetrics)
		r.Get("/projects/{id}/curve", h.GetSCurve)
		r.Get("/projects/{id}/forecast", h.GetForecast)

		// Control Accounts
		r.Get("/control-accounts", h.ListControlAccounts)
		r.Post("/control-accounts", h.CreateControlAccount)
		r.Get("/control-accounts/{id}", h.GetControlAccount)
		r.Put("/control-accounts/{id}", h.UpdateControlAccount)
		r.Delete("/control-accounts/{id}", h.DeleteControlAccount)

		// Baselines
		r.Get("/baselines", h.ListBaselines)
		r.Post("/baselines", h.CreateBaseline)
		r.Get("/baselines/{id}", h.GetBaseline)
		r.Put("/baselines/{id}", h.UpdateBaseline)
		r.Delete("/baselines/{id}", h.DeleteBaseline)
		r.Post("/baselines/{id}/activate", h.ActivateBaseline)

		// Periods (PV)
		r.Get("/periods", h.ListPeriods)
		r.Post("/periods", h.CreatePeriod)
		r.Get("/periods/{id}", h.GetPeriod)
		r.Put("/periods/{id}", h.UpdatePeriod)
		r.Delete("/periods/{id}", h.DeletePeriod)

		// Actuals (AC/EV)
		r.Get("/actuals", h.ListActuals)
		r.Post("/actuals", h.CreateActual)
		r.Get("/actuals/{id}", h.GetActual)
		r.Put("/actuals/{id}", h.UpdateActual)
		r.Delete("/actuals/{id}", h.DeleteActual)

		// Metrics
		r.Get("/metrics", h.ListMetrics)
		r.Post("/metrics/calculate", h.CalculateMetrics)

		// Forecasts
		r.Get("/forecasts", h.ListForecasts)
		r.Post("/forecasts", h.CreateForecast)

		// Earned Rules
		r.Get("/rules", h.ListRules)
		r.Post("/rules", h.CreateRule)
		r.Get("/rules/{id}", h.GetRule)
		r.Put("/rules/{id}", h.UpdateRule)
	})
}

// --- EVM Project Summary ---

func (h *EVMHandler) GetProjectSummary(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")

	// Get latest metrics
	var m models.EVMMetric
	err := h.db.QueryRow(`
		SELECT id, project_id, period_date, pv, ev, ac, bac, sv, cv, sv_pct, cv_pct, 
		       spi, cpi, eac, etc, vac, tcpi, calculated_at
		FROM evm_metrics 
		WHERE project_id = $1 AND metric_scope = 'project' AND is_cumulative = TRUE
		ORDER BY period_date DESC LIMIT 1`, projectID).Scan(
		&m.ID, &m.ProjectID, &m.PeriodDate, &m.PV, &m.EV, &m.AC, &m.BAC,
		&m.SV, &m.CV, &m.SVPct, &m.CVPct, &m.SPI, &m.CPI,
		&m.EAC, &m.ETC, &m.VAC, &m.TCPI, &m.CalculatedAt,
	)
	if err != nil && err != sql.ErrNoRows {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Get EVM project config
	var evmProj models.EVMProject
	err = h.db.QueryRow(`
		SELECT id, project_id, evm_enabled, default_baseline_id, reporting_freq, currency,
		       threshold_spi, threshold_cpi, threshold_sv_pct, threshold_cv_pct, created_at, updated_at
		FROM evm_projects WHERE project_id = $1`, projectID).Scan(
		&evmProj.ID, &evmProj.ProjectID, &evmProj.EVMEnabled, &evmProj.DefaultBaselineID,
		&evmProj.ReportingFreq, &evmProj.Currency,
		&evmProj.ThresholdSPI, &evmProj.ThresholdCPI, &evmProj.ThresholdSVPct, &evmProj.ThresholdCVPct,
		&evmProj.CreatedAt, &evmProj.UpdatedAt,
	)
	if err != nil && err != sql.ErrNoRows {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Get total periods count
	var periodCount int
	h.db.QueryRow(`SELECT COUNT(*) FROM evm_periods WHERE project_id = $1`, projectID).Scan(&periodCount)

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"metrics":      m,
		"evm_project":  evmProj,
		"period_count": periodCount,
	})
}

// --- Metrics ---

func (h *EVMHandler) GetProjectMetrics(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")

	rows, err := h.db.Query(`
		SELECT id, project_id, control_account_id, period_date, pv, ev, ac, bac,
		       sv, cv, sv_pct, cv_pct, spi, cpi, eac, etc, vac, tcpi,
		       metric_scope, is_cumulative, calculated_at, created_at
		FROM evm_metrics 
		WHERE project_id = $1
		ORDER BY period_date DESC, metric_scope`, projectID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	metrics := make([]models.EVMMetric, 0)
	for rows.Next() {
		var m models.EVMMetric
		if err := rows.Scan(&m.ID, &m.ProjectID, &m.ControlAccountID, &m.PeriodDate,
			&m.PV, &m.EV, &m.AC, &m.BAC, &m.SV, &m.CV, &m.SVPct, &m.CVPct,
			&m.SPI, &m.CPI, &m.EAC, &m.ETC, &m.VAC, &m.TCPI,
			&m.MetricScope, &m.IsCumulative, &m.CalculatedAt, &m.CreatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		metrics = append(metrics, m)
	}
	respondJSON(w, http.StatusOK, metrics)
}

// --- S-Curve ---

func (h *EVMHandler) GetSCurve(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")

	// Get PV over periods
	rows, err := h.db.Query(`
		SELECT period_date, planned_value, planned_hours 
		FROM evm_periods 
		WHERE project_id = $1 AND is_cumulative = FALSE
		ORDER BY period_date`, projectID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	periods := make([]map[string]interface{}, 0)
	for rows.Next() {
		var date time.Time
		var pv, hours float64
		rows.Scan(&date, &pv, &hours)
		periods = append(periods, map[string]interface{}{
			"period_date":  date.Format("2006-01-02"),
			"planned_value": pv,
			"planned_hours": hours,
		})
	}

	// Get metrics curve (cumulative PV, EV, AC)
	rows2, err := h.db.Query(`
		SELECT period_date, pv, ev, ac
		FROM evm_metrics 
		WHERE project_id = $1 AND metric_scope = 'project' AND is_cumulative = TRUE
		ORDER BY period_date`, projectID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows2.Close()

	curve := make([]map[string]interface{}, 0)
	for rows2.Next() {
		var date time.Time
		var pv, ev, ac float64
		rows2.Scan(&date, &pv, &ev, &ac)
		curve = append(curve, map[string]interface{}{
			"period_date": date.Format("2006-01-02"),
			"pv": pv,
			"ev": ev,
			"ac": ac,
		})
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"periods": periods,
		"curve":   curve,
	})
}

// --- Forecasts ---

func (h *EVMHandler) GetForecast(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")

	rows, err := h.db.Query(`
		SELECT id, project_id, control_account_id, forecast_date, forecast_type, method,
		       eac_value, etc_value, vac_value, completion_date, confidence_pct, notes, created_by, created_at
		FROM evm_forecasts 
		WHERE project_id = $1
		ORDER BY forecast_date DESC`, projectID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	forecasts := make([]models.EVMForecast, 0)
	for rows.Next() {
		var f models.EVMForecast
		if err := rows.Scan(&f.ID, &f.ProjectID, &f.ControlAccountID, &f.ForecastDate,
			&f.ForecastType, &f.Method, &f.EACValue, &f.ETCValue, &f.VACValue,
			&f.CompletionDate, &f.ConfidencePct, &f.Notes, &f.CreatedBy, &f.CreatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		forecasts = append(forecasts, f)
	}
	respondJSON(w, http.StatusOK, forecasts)
}

// --- Control Accounts ---

func (h *EVMHandler) ListControlAccounts(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT id, project_id, ca_code, ca_name, description, wbs_code, responsible, sort_order, is_active, created_at, updated_at FROM evm_control_accounts WHERE 1=1`
	var rows *sql.Rows
	var err error
	if projectID != "" {
		rows, err = h.db.Query(query+` AND project_id = $1 ORDER BY sort_order`, projectID)
	} else {
		rows, err = h.db.Query(query + ` ORDER BY project_id, sort_order`)
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	items := make([]models.EVMControlAccount, 0)
	for rows.Next() {
		var c models.EVMControlAccount
		if err := rows.Scan(&c.ID, &c.ProjectID, &c.CACode, &c.CAName, &c.Description, &c.WBSCode, &c.Responsible, &c.SortOrder, &c.IsActive, &c.CreatedAt, &c.UpdatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, c)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *EVMHandler) CreateControlAccount(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID   string  `json:"project_id"`
		CACode      string  `json:"ca_code"`
		CAName      string  `json:"ca_name"`
		Description *string `json:"description"`
		WBSCode     *string `json:"wbs_code"`
		Responsible *string `json:"responsible"`
		SortOrder   int     `json:"sort_order"`
		IsActive    bool    `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	_, err := h.db.Exec(`
		INSERT INTO evm_control_accounts (id, project_id, ca_code, ca_name, description, wbs_code, responsible, sort_order, is_active)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		id, input.ProjectID, input.CACode, input.CAName, input.Description, input.WBSCode, input.Responsible, input.SortOrder, input.IsActive)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *EVMHandler) GetControlAccount(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var c models.EVMControlAccount
	err := h.db.QueryRow(`
		SELECT id, project_id, ca_code, ca_name, description, wbs_code, responsible, sort_order, is_active, created_at, updated_at
		FROM evm_control_accounts WHERE id = $1`, id).Scan(
		&c.ID, &c.ProjectID, &c.CACode, &c.CAName, &c.Description, &c.WBSCode, &c.Responsible, &c.SortOrder, &c.IsActive, &c.CreatedAt, &c.UpdatedAt)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, c)
}

func (h *EVMHandler) UpdateControlAccount(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		CACode      string  `json:"ca_code"`
		CAName      string  `json:"ca_name"`
		Description *string `json:"description"`
		WBSCode     *string `json:"wbs_code"`
		Responsible *string `json:"responsible"`
		SortOrder   int     `json:"sort_order"`
		IsActive    bool    `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	_, err := h.db.Exec(`
		UPDATE evm_control_accounts SET ca_code=$1, ca_name=$2, description=$3, wbs_code=$4, responsible=$5, sort_order=$6, is_active=$7, updated_at=NOW()
		WHERE id=$8`,
		input.CACode, input.CAName, input.Description, input.WBSCode, input.Responsible, input.SortOrder, input.IsActive, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *EVMHandler) DeleteControlAccount(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM evm_control_accounts WHERE id = $1`, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// --- Baselines ---

func (h *EVMHandler) ListBaselines(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT id, project_id, baseline_name, baseline_type, version, description, is_approved, approved_by, approved_at, is_active, created_at FROM evm_baselines WHERE 1=1`
	var rows *sql.Rows
	var err error
	if projectID != "" {
		rows, err = h.db.Query(query+` AND project_id = $1 ORDER BY created_at DESC`, projectID)
	} else {
		rows, err = h.db.Query(query + ` ORDER BY created_at DESC`)
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	items := make([]models.EVMBaseline, 0)
	for rows.Next() {
		var b models.EVMBaseline
		if err := rows.Scan(&b.ID, &b.ProjectID, &b.BaselineName, &b.BaselineType, &b.Version, &b.Description, &b.IsApproved, &b.ApprovedBy, &b.ApprovedAt, &b.IsActive, &b.CreatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, b)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *EVMHandler) CreateBaseline(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID    string  `json:"project_id"`
		BaselineName string  `json:"baseline_name"`
		BaselineType string  `json:"baseline_type"`
		Version      string  `json:"version"`
		Description  *string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	_, err := h.db.Exec(`
		INSERT INTO evm_baselines (id, project_id, baseline_name, baseline_type, version, description)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		id, input.ProjectID, input.BaselineName, input.BaselineType, input.Version, input.Description)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *EVMHandler) GetBaseline(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var b models.EVMBaseline
	err := h.db.QueryRow(`
		SELECT id, project_id, baseline_name, baseline_type, version, description, is_approved, approved_by, approved_at, is_active, created_at
		FROM evm_baselines WHERE id = $1`, id).Scan(
		&b.ID, &b.ProjectID, &b.BaselineName, &b.BaselineType, &b.Version, &b.Description, &b.IsApproved, &b.ApprovedBy, &b.ApprovedAt, &b.IsActive, &b.CreatedAt)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, b)
}

func (h *EVMHandler) UpdateBaseline(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		BaselineName string  `json:"baseline_name"`
		BaselineType string  `json:"baseline_type"`
		Version      string  `json:"version"`
		Description  *string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	_, err := h.db.Exec(`
		UPDATE evm_baselines SET baseline_name=$1, baseline_type=$2, version=$3, description=$4 WHERE id=$5`,
		input.BaselineName, input.BaselineType, input.Version, input.Description, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *EVMHandler) DeleteBaseline(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM evm_baselines WHERE id = $1`, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (h *EVMHandler) ActivateBaseline(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// Get project_id from baseline
	var projectID string
	err := h.db.QueryRow(`SELECT project_id FROM evm_baselines WHERE id = $1`, id).Scan(&projectID)
	if err != nil {
		respondError(w, http.StatusNotFound, "baseline not found")
		return
	}

	// Deactivate all other baselines for this project, activate this one
	tx, err := h.db.Begin()
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec(`UPDATE evm_baselines SET is_active = FALSE WHERE project_id = $1`, projectID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	_, err = tx.Exec(`UPDATE evm_baselines SET is_active = TRUE WHERE id = $1`, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	tx.Commit()
	respondJSON(w, http.StatusOK, map[string]string{"status": "activated"})
}

// --- Periods ---

func (h *EVMHandler) ListPeriods(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT id, project_id, control_account_id, baseline_id, period_date, period_type, planned_value, planned_hours, planned_progress, is_cumulative, notes, created_at FROM evm_periods WHERE 1=1`
	var rows *sql.Rows
	var err error
	if projectID != "" {
		rows, err = h.db.Query(query+` AND project_id = $1 ORDER BY period_date`, projectID)
	} else {
		rows, err = h.db.Query(query + ` ORDER BY period_date`)
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	items := make([]models.EVMPeriod, 0)
	for rows.Next() {
		var p models.EVMPeriod
		if err := rows.Scan(&p.ID, &p.ProjectID, &p.ControlAccountID, &p.BaselineID, &p.PeriodDate, &p.PeriodType, &p.PlannedValue, &p.PlannedHours, &p.PlannedProgress, &p.IsCumulative, &p.Notes, &p.CreatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, p)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *EVMHandler) CreatePeriod(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID        string  `json:"project_id"`
		ControlAccountID *string `json:"control_account_id"`
		BaselineID       *string `json:"baseline_id"`
		PeriodDate       string  `json:"period_date"`
		PeriodType       string  `json:"period_type"`
		PlannedValue     float64 `json:"planned_value"`
		PlannedHours     float64 `json:"planned_hours"`
		PlannedProgress  float64 `json:"planned_progress"`
		IsCumulative     bool    `json:"is_cumulative"`
		Notes            *string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	_, err := h.db.Exec(`
		INSERT INTO evm_periods (id, project_id, control_account_id, baseline_id, period_date, period_type, planned_value, planned_hours, planned_progress, is_cumulative, notes)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		id, input.ProjectID, input.ControlAccountID, input.BaselineID, input.PeriodDate, input.PeriodType,
		input.PlannedValue, input.PlannedHours, input.PlannedProgress, input.IsCumulative, input.Notes)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *EVMHandler) GetPeriod(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var p models.EVMPeriod
	err := h.db.QueryRow(`
		SELECT id, project_id, control_account_id, baseline_id, period_date, period_type, planned_value, planned_hours, planned_progress, is_cumulative, notes, created_at
		FROM evm_periods WHERE id = $1`, id).Scan(
		&p.ID, &p.ProjectID, &p.ControlAccountID, &p.BaselineID, &p.PeriodDate, &p.PeriodType, &p.PlannedValue, &p.PlannedHours, &p.PlannedProgress, &p.IsCumulative, &p.Notes, &p.CreatedAt)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, p)
}

func (h *EVMHandler) UpdatePeriod(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		PlannedValue    float64 `json:"planned_value"`
		PlannedHours    float64 `json:"planned_hours"`
		PlannedProgress float64 `json:"planned_progress"`
		Notes           *string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	_, err := h.db.Exec(`
		UPDATE evm_periods SET planned_value=$1, planned_hours=$2, planned_progress=$3, notes=$4 WHERE id=$5`,
		input.PlannedValue, input.PlannedHours, input.PlannedProgress, input.Notes, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *EVMHandler) DeletePeriod(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM evm_periods WHERE id = $1`, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// --- Actuals ---

func (h *EVMHandler) ListActuals(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT id, project_id, control_account_id, period_date, actual_cost, actual_hours, earned_value, progress_pct, physical_pct, data_source, source_id, recorded_at, created_at FROM evm_actuals WHERE 1=1`
	var rows *sql.Rows
	var err error
	if projectID != "" {
		rows, err = h.db.Query(query+` AND project_id = $1 ORDER BY period_date`, projectID)
	} else {
		rows, err = h.db.Query(query + ` ORDER BY period_date`)
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	items := make([]models.EVMActual, 0)
	for rows.Next() {
		var a models.EVMActual
		if err := rows.Scan(&a.ID, &a.ProjectID, &a.ControlAccountID, &a.PeriodDate, &a.ActualCost, &a.ActualHours, &a.EarnedValue, &a.ProgressPct, &a.PhysicalPct, &a.DataSource, &a.SourceID, &a.RecordedAt, &a.CreatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, a)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *EVMHandler) CreateActual(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID        string   `json:"project_id"`
		ControlAccountID *string  `json:"control_account_id"`
		PeriodDate       string   `json:"period_date"`
		ActualCost       float64  `json:"actual_cost"`
		ActualHours      float64  `json:"actual_hours"`
		EarnedValue      float64  `json:"earned_value"`
		ProgressPct      float64  `json:"progress_pct"`
		PhysicalPct      *float64 `json:"physical_pct"`
		DataSource       string   `json:"data_source"`
		SourceID         *string  `json:"source_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	_, err := h.db.Exec(`
		INSERT INTO evm_actuals (id, project_id, control_account_id, period_date, actual_cost, actual_hours, earned_value, progress_pct, physical_pct, data_source, source_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		id, input.ProjectID, input.ControlAccountID, input.PeriodDate,
		input.ActualCost, input.ActualHours, input.EarnedValue, input.ProgressPct, input.PhysicalPct, input.DataSource, input.SourceID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *EVMHandler) GetActual(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var a models.EVMActual
	err := h.db.QueryRow(`
		SELECT id, project_id, control_account_id, period_date, actual_cost, actual_hours, earned_value, progress_pct, physical_pct, data_source, source_id, recorded_at, created_at
		FROM evm_actuals WHERE id = $1`, id).Scan(
		&a.ID, &a.ProjectID, &a.ControlAccountID, &a.PeriodDate, &a.ActualCost, &a.ActualHours, &a.EarnedValue, &a.ProgressPct, &a.PhysicalPct, &a.DataSource, &a.SourceID, &a.RecordedAt, &a.CreatedAt)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, a)
}

func (h *EVMHandler) UpdateActual(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		ActualCost   float64  `json:"actual_cost"`
		ActualHours  float64  `json:"actual_hours"`
		EarnedValue  float64  `json:"earned_value"`
		ProgressPct  float64  `json:"progress_pct"`
		PhysicalPct  *float64 `json:"physical_pct"`
		DataSource   string   `json:"data_source"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	_, err := h.db.Exec(`
		UPDATE evm_actuals SET actual_cost=$1, actual_hours=$2, earned_value=$3, progress_pct=$4, physical_pct=$5, data_source=$6 WHERE id=$7`,
		input.ActualCost, input.ActualHours, input.EarnedValue, input.ProgressPct, input.PhysicalPct, input.DataSource, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *EVMHandler) DeleteActual(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM evm_actuals WHERE id = $1`, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// --- Metrics ---

func (h *EVMHandler) ListMetrics(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT id, project_id, control_account_id, period_date, pv, ev, ac, bac, sv, cv, sv_pct, cv_pct, spi, cpi, eac, etc, vac, tcpi, metric_scope, is_cumulative, calculated_at, created_at FROM evm_metrics WHERE 1=1`
	var rows *sql.Rows
	var err error
	if projectID != "" {
		rows, err = h.db.Query(query+` AND project_id = $1 ORDER BY period_date DESC`, projectID)
	} else {
		rows, err = h.db.Query(query + ` ORDER BY period_date DESC`)
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	items := make([]models.EVMMetric, 0)
	for rows.Next() {
		var m models.EVMMetric
		if err := rows.Scan(&m.ID, &m.ProjectID, &m.ControlAccountID, &m.PeriodDate,
			&m.PV, &m.EV, &m.AC, &m.BAC, &m.SV, &m.CV, &m.SVPct, &m.CVPct,
			&m.SPI, &m.CPI, &m.EAC, &m.ETC, &m.VAC, &m.TCPI,
			&m.MetricScope, &m.IsCumulative, &m.CalculatedAt, &m.CreatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, m)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *EVMHandler) CalculateMetrics(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID  string `json:"project_id"`
		PeriodDate string `json:"period_date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var metricID string
	err := h.db.QueryRow(`SELECT calculate_evm_metrics($1, $2)`, input.ProjectID, input.PeriodDate).Scan(&metricID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"metric_id": metricID})
}

// --- Forecasts ---

func (h *EVMHandler) ListForecasts(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT id, project_id, control_account_id, forecast_date, forecast_type, method, eac_value, etc_value, vac_value, completion_date, confidence_pct, notes, created_by, created_at FROM evm_forecasts WHERE 1=1`
	var rows *sql.Rows
	var err error
	if projectID != "" {
		rows, err = h.db.Query(query+` AND project_id = $1 ORDER BY forecast_date DESC`, projectID)
	} else {
		rows, err = h.db.Query(query + ` ORDER BY forecast_date DESC`)
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	items := make([]models.EVMForecast, 0)
	for rows.Next() {
		var f models.EVMForecast
		if err := rows.Scan(&f.ID, &f.ProjectID, &f.ControlAccountID, &f.ForecastDate,
			&f.ForecastType, &f.Method, &f.EACValue, &f.ETCValue, &f.VACValue,
			&f.CompletionDate, &f.ConfidencePct, &f.Notes, &f.CreatedBy, &f.CreatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, f)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *EVMHandler) CreateForecast(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID        string   `json:"project_id"`
		ControlAccountID *string  `json:"control_account_id"`
		ForecastDate     string   `json:"forecast_date"`
		ForecastType     string   `json:"forecast_type"`
		Method           string   `json:"method"`
		EACValue         *float64 `json:"eac_value"`
		ETCValue         *float64 `json:"etc_value"`
		VACValue         *float64 `json:"vac_value"`
		CompletionDate   *string  `json:"completion_date"`
		ConfidencePct    *float64 `json:"confidence_pct"`
		Notes            *string  `json:"notes"`
		CreatedBy        *string  `json:"created_by"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	_, err := h.db.Exec(`
		INSERT INTO evm_forecasts (id, project_id, control_account_id, forecast_date, forecast_type, method, eac_value, etc_value, vac_value, completion_date, confidence_pct, notes, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		id, input.ProjectID, input.ControlAccountID, input.ForecastDate, input.ForecastType, input.Method,
		input.EACValue, input.ETCValue, input.VACValue, input.CompletionDate, input.ConfidencePct, input.Notes, input.CreatedBy)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

// --- Earned Rules ---

func (h *EVMHandler) ListRules(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT id, project_id, rule_name, rule_type, description, weight_pct, config, is_active, created_at FROM evm_earned_rules WHERE 1=1`
	var rows *sql.Rows
	var err error
	if projectID != "" {
		rows, err = h.db.Query(query+` AND project_id = $1 ORDER BY rule_name`, projectID)
	} else {
		rows, err = h.db.Query(query + ` ORDER BY rule_name`)
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	items := make([]models.EVMRule, 0)
	for rows.Next() {
		var rl models.EVMRule
		if err := rows.Scan(&rl.ID, &rl.ProjectID, &rl.RuleName, &rl.RuleType, &rl.Description, &rl.WeightPct, &rl.Config, &rl.IsActive, &rl.CreatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, rl)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *EVMHandler) CreateRule(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID   string  `json:"project_id"`
		RuleName    string  `json:"rule_name"`
		RuleType    string  `json:"rule_type"`
		Description *string `json:"description"`
		WeightPct   float64 `json:"weight_pct"`
		Config      *string `json:"config"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	_, err := h.db.Exec(`
		INSERT INTO evm_earned_rules (id, project_id, rule_name, rule_type, description, weight_pct, config)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		id, input.ProjectID, input.RuleName, input.RuleType, input.Description, input.WeightPct, input.Config)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *EVMHandler) GetRule(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var rl models.EVMRule
	err := h.db.QueryRow(`
		SELECT id, project_id, rule_name, rule_type, description, weight_pct, config, is_active, created_at
		FROM evm_earned_rules WHERE id = $1`, id).Scan(
		&rl.ID, &rl.ProjectID, &rl.RuleName, &rl.RuleType, &rl.Description, &rl.WeightPct, &rl.Config, &rl.IsActive, &rl.CreatedAt)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, rl)
}

func (h *EVMHandler) UpdateRule(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		RuleName    string  `json:"rule_name"`
		RuleType    string  `json:"rule_type"`
		Description *string `json:"description"`
		WeightPct   float64 `json:"weight_pct"`
		Config      *string `json:"config"`
		IsActive    bool    `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	_, err := h.db.Exec(`
		UPDATE evm_earned_rules SET rule_name=$1, rule_type=$2, description=$3, weight_pct=$4, config=$5, is_active=$6 WHERE id=$7`,
		input.RuleName, input.RuleType, input.Description, input.WeightPct, input.Config, input.IsActive, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}