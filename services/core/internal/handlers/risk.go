package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// RiskHandler handles Risk Management module endpoints
type RiskHandler struct {
	db *sql.DB
}

func NewRiskHandler(db *sql.DB) *RiskHandler {
	return &RiskHandler{db: db}
}

func (h *RiskHandler) RegisterRoutes(r chi.Router) {
	r.Route("/risk", func(r chi.Router) {
		r.Get("/categories", h.ListCategories)
		r.Post("/categories", h.CreateCategory)
		r.Get("/categories/{id}", h.GetCategory)
		r.Put("/categories/{id}", h.UpdateCategory)
		r.Delete("/categories/{id}", h.DeleteCategory)

		r.Get("/registers", h.ListRegisters)
		r.Post("/registers", h.CreateRegister)
		r.Get("/registers/{id}", h.GetRegister)
		r.Put("/registers/{id}", h.UpdateRegister)
		r.Delete("/registers/{id}", h.DeleteRegister)

		r.Get("/matrices", h.ListMatrices)
		r.Post("/matrices", h.CreateMatrix)
		r.Get("/matrices/{id}", h.GetMatrix)
		r.Put("/matrices/{id}", h.UpdateMatrix)
		r.Delete("/matrices/{id}", h.DeleteMatrix)

		r.Get("/monte-carlo", h.ListMonteCarlo)
		r.Post("/monte-carlo", h.CreateMonteCarlo)
		r.Get("/monte-carlo/{id}", h.GetMonteCarlo)
		r.Put("/monte-carlo/{id}", h.UpdateMonteCarlo)
		r.Delete("/monte-carlo/{id}", h.DeleteMonteCarlo)

		r.Get("/scenarios", h.ListScenarios)
		r.Post("/scenarios", h.CreateScenario)
		r.Get("/scenarios/{id}", h.GetScenario)
		r.Put("/scenarios/{id}", h.UpdateScenario)
		r.Delete("/scenarios/{id}", h.DeleteScenario)

		r.Get("/mitigations", h.ListMitigations)
		r.Post("/mitigations", h.CreateMitigation)
		r.Get("/mitigations/{id}", h.GetMitigation)
		r.Put("/mitigations/{id}", h.UpdateMitigation)
		r.Delete("/mitigations/{id}", h.DeleteMitigation)

		r.Get("/escalations", h.ListEscalations)
		r.Post("/escalations", h.CreateEscalation)
		r.Get("/escalations/{id}", h.GetEscalation)
		r.Put("/escalations/{id}", h.UpdateEscalation)
		r.Delete("/escalations/{id}", h.DeleteEscalation)

		r.Get("/dashboard", h.ListDashboard)
		r.Post("/dashboard", h.CreateDashboard)
		r.Get("/dashboard/{id}", h.GetDashboard)
		r.Put("/dashboard/{id}", h.UpdateDashboard)
		r.Delete("/dashboard/{id}", h.DeleteDashboard)

		r.Get("/summary", h.GetSummary)
	})
}

// =============================================================================
// Categories
// =============================================================================
func (h *RiskHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(`SELECT id, category_code, category_name, category_type, description, sort_order, is_active FROM risk_categories ORDER BY sort_order`)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, code, name, ctype, desc string
		var sortOrder int
		var active bool
		if err := rows.Scan(&id, &code, &name, &ctype, &desc, &sortOrder, &active); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		items = append(items, map[string]interface{}{
			"id": id, "category_code": code, "category_name": name,
			"category_type": ctype, "description": desc, "sort_order": sortOrder, "is_active": active,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *RiskHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var input struct {
		CategoryCode string `json:"category_code"`
		CategoryName string `json:"category_name"`
		CategoryType string `json:"category_type"`
		Description  string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO risk_categories (id, category_code, category_name, category_type, description, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		id, input.CategoryCode, input.CategoryName, input.CategoryType, input.Description, now, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *RiskHandler) GetCategory(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var code, name string
	err := h.db.QueryRow(`SELECT category_code, category_name FROM risk_categories WHERE id = $1`, id).Scan(&code, &name)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "category not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "category_code": code, "category_name": name})
}

func (h *RiskHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		IsActive *bool `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	_, err := h.db.Exec(`UPDATE risk_categories SET is_active=COALESCE($1,is_active), updated_at=$2 WHERE id=$3`, input.IsActive, time.Now(), id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *RiskHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM risk_categories WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// Risk Registers
// =============================================================================
func (h *RiskHandler) ListRegisters(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	status := r.URL.Query().Get("status")
	rating := r.URL.Query().Get("risk_rating")

	query := `SELECT id, project_id, risk_number, risk_code, risk_name, risk_type, description, probability_score, impact_score, risk_score, risk_rating, cost_impact, schedule_impact_days, risk_owner, risk_response, status, created_at FROM risk_registers WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if projectID != "" { query += fmt.Sprintf(" AND project_id = $%d", argIdx); argIdx++; args = append(args, projectID) }
	if status != "" { query += fmt.Sprintf(" AND status = $%d", argIdx); argIdx++; args = append(args, status) }
	if rating != "" { query += fmt.Sprintf(" AND risk_rating = $%d", argIdx); argIdx++; args = append(args, rating) }
	query += " ORDER BY risk_score DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, pid, code, rname, rtype, desc, rrating, owner, resp, st string
		var num, schedDays int
		var prob, impact, score, cost float64
		var createdAt time.Time
		if err := rows.Scan(&id, &pid, &num, &code, &rname, &rtype, &desc, &prob, &impact, &score, &rrating, &cost, &schedDays, &owner, &resp, &st, &createdAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		items = append(items, map[string]interface{}{
			"id": id, "project_id": pid, "risk_number": num, "risk_code": code,
			"risk_name": rname, "risk_type": rtype, "description": desc,
			"probability_score": prob, "impact_score": impact, "risk_score": score,
			"risk_rating": rrating, "cost_impact": cost, "schedule_impact_days": schedDays,
			"risk_owner": owner, "risk_response": resp, "status": st, "created_at": createdAt,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *RiskHandler) CreateRegister(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID   string  `json:"project_id"`
		RiskCode    string  `json:"risk_code"`
		RiskName    string  `json:"risk_name"`
		RiskType    string  `json:"risk_type"`
		Description string  `json:"description"`
		Probability float64 `json:"probability_score"`
		Impact      float64 `json:"impact_score"`
		CostImpact  float64 `json:"cost_impact"`
		RiskOwner   string  `json:"risk_owner"`
		RiskResponse string `json:"risk_response"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO risk_registers (id, project_id, risk_number, risk_code, risk_name, risk_type, description, probability_score, impact_score, cost_impact, risk_owner, risk_response, status, created_at, updated_at) VALUES ($1,$2,(SELECT COALESCE(MAX(risk_number),0)+1 FROM risk_registers WHERE project_id=$2),$3,$4,$5,$6,$7,$8,$9,$10,$11,'identified',$12,$13)`,
		id, input.ProjectID, input.RiskCode, input.RiskName, input.RiskType, input.Description, input.Probability, input.Impact, input.CostImpact, input.RiskOwner, input.RiskResponse, now, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *RiskHandler) GetRegister(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var code, name string
	err := h.db.QueryRow(`SELECT risk_code, risk_name FROM risk_registers WHERE id = $1`, id).Scan(&code, &name)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "risk not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "risk_code": code, "risk_name": name})
}

func (h *RiskHandler) UpdateRegister(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Status          *string  `json:"status"`
		MitigationStrategy *string `json:"mitigation_strategy"`
		ContingencyPlan *string  `json:"contingency_plan"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	_, err := h.db.Exec(`UPDATE risk_registers SET status=COALESCE($1,status), mitigation_strategy=COALESCE($2,mitigation_strategy), contingency_plan=COALESCE($3,contingency_plan), updated_at=$4 WHERE id=$5`,
		input.Status, input.MitigationStrategy, input.ContingencyPlan, time.Now(), id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *RiskHandler) DeleteRegister(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM risk_registers WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// Matrices
// =============================================================================
func (h *RiskHandler) ListMatrices(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT id, project_id, matrix_name, matrix_type, is_active, version, description, created_at FROM risk_matrices WHERE 1=1`
	args := []interface{}{}
	argIdx := 1
	if projectID != "" { query += fmt.Sprintf(" AND project_id = $%d", argIdx); args = append(args, projectID) }
	query += " ORDER BY matrix_name"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, pid, mname, mtype, ver, desc string
		var active bool
		var createdAt time.Time
		if err := rows.Scan(&id, &pid, &mname, &mtype, &active, &ver, &desc, &createdAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		items = append(items, map[string]interface{}{
			"id": id, "project_id": pid, "matrix_name": mname, "matrix_type": mtype,
			"is_active": active, "version": ver, "description": desc, "created_at": createdAt,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *RiskHandler) CreateMatrix(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID   string `json:"project_id"`
		MatrixName  string `json:"matrix_name"`
		MatrixType  string `json:"matrix_type"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO risk_matrices (id, project_id, matrix_name, matrix_type, description, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		id, input.ProjectID, input.MatrixName, input.MatrixType, input.Description, now, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *RiskHandler) GetMatrix(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var name string
	err := h.db.QueryRow(`SELECT matrix_name FROM risk_matrices WHERE id = $1`, id).Scan(&name)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "matrix not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "matrix_name": name})
}

func (h *RiskHandler) UpdateMatrix(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		IsActive *bool `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	_, err := h.db.Exec(`UPDATE risk_matrices SET is_active=COALESCE($1,is_active), updated_at=$2 WHERE id=$3`, input.IsActive, time.Now(), id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *RiskHandler) DeleteMatrix(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM risk_matrices WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// Monte Carlo
// =============================================================================
func (h *RiskHandler) ListMonteCarlo(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT id, project_id, run_label, run_type, iterations, p10_value, p50_value, p90_value, mean_value, confidence_level, status, created_at FROM risk_monte_carlo_runs WHERE 1=1`
	args := []interface{}{}
	argIdx := 1
	if projectID != "" { query += fmt.Sprintf(" AND project_id = $%d", argIdx); args = append(args, projectID) }
	query += " ORDER BY created_at DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, pid, label, rtype, st string
		var iters int
		var p10, p50, p90, mean, conf float64
		var createdAt time.Time
		if err := rows.Scan(&id, &pid, &label, &rtype, &iters, &p10, &p50, &p90, &mean, &conf, &st, &createdAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		items = append(items, map[string]interface{}{
			"id": id, "project_id": pid, "run_label": label, "run_type": rtype,
			"iterations": iters, "p10_value": p10, "p50_value": p50, "p90_value": p90,
			"mean_value": mean, "confidence_level": conf, "status": st, "created_at": createdAt,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *RiskHandler) CreateMonteCarlo(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID  string `json:"project_id"`
		RunLabel   string `json:"run_label"`
		RunType    string `json:"run_type"`
		Iterations int    `json:"iterations"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO risk_monte_carlo_runs (id, project_id, run_label, run_type, iterations, status, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,'pending',$6,$7)`,
		id, input.ProjectID, input.RunLabel, input.RunType, input.Iterations, now, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *RiskHandler) GetMonteCarlo(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var label, rtype string
	err := h.db.QueryRow(`SELECT run_label, run_type FROM risk_monte_carlo_runs WHERE id = $1`, id).Scan(&label, &rtype)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "monte carlo run not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "run_label": label, "run_type": rtype})
}

func (h *RiskHandler) UpdateMonteCarlo(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Status    *string  `json:"status"`
		P10Value  *float64 `json:"p10_value"`
		P50Value  *float64 `json:"p50_value"`
		P90Value  *float64 `json:"p90_value"`
		MeanValue *float64 `json:"mean_value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	_, err := h.db.Exec(`UPDATE risk_monte_carlo_runs SET status=COALESCE($1,status), p10_value=COALESCE($2,p10_value), p50_value=COALESCE($3,p50_value), p90_value=COALESCE($4,p90_value), mean_value=COALESCE($5,mean_value), updated_at=$6 WHERE id=$7`,
		input.Status, input.P10Value, input.P50Value, input.P90Value, input.MeanValue, time.Now(), id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *RiskHandler) DeleteMonteCarlo(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM risk_monte_carlo_runs WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// Scenarios
// =============================================================================
func (h *RiskHandler) ListScenarios(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT id, project_id, scenario_number, scenario_code, scenario_name, scenario_type, description, cost_impact_min, cost_impact_max, cost_impact_ml, schedule_impact_min, schedule_impact_max, schedule_impact_ml, probability_pct, severity, status FROM risk_scenarios WHERE 1=1`
	args := []interface{}{}
	argIdx := 1
	if projectID != "" { query += fmt.Sprintf(" AND project_id = $%d", argIdx); args = append(args, projectID) }
	query += " ORDER BY scenario_number"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, pid, code, sname, stype, desc, sev, st string
		var num, schedMin, schedMax, schedML int
		var costMin, costMax, costML, prob float64
		if err := rows.Scan(&id, &pid, &num, &code, &sname, &stype, &desc, &costMin, &costMax, &costML, &schedMin, &schedMax, &schedML, &prob, &sev, &st); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		items = append(items, map[string]interface{}{
			"id": id, "project_id": pid, "scenario_number": num, "scenario_code": code,
			"scenario_name": sname, "scenario_type": stype, "description": desc,
			"cost_impact_min": costMin, "cost_impact_max": costMax, "cost_impact_ml": costML,
			"schedule_impact_min": schedMin, "schedule_impact_max": schedMax, "schedule_impact_ml": schedML,
			"probability_pct": prob, "severity": sev, "status": st,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *RiskHandler) CreateScenario(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID    string  `json:"project_id"`
		ScenarioCode string  `json:"scenario_code"`
		ScenarioName string  `json:"scenario_name"`
		ScenarioType string  `json:"scenario_type"`
		Description  string  `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO risk_scenarios (id, project_id, scenario_number, scenario_code, scenario_name, scenario_type, description, status, created_at, updated_at) VALUES ($1,$2,(SELECT COALESCE(MAX(scenario_number),0)+1 FROM risk_scenarios WHERE project_id=$2),$3,$4,$5,$6,'draft',$7,$8)`,
		id, input.ProjectID, input.ScenarioCode, input.ScenarioName, input.ScenarioType, input.Description, now, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *RiskHandler) GetScenario(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var code, name string
	err := h.db.QueryRow(`SELECT scenario_code, scenario_name FROM risk_scenarios WHERE id = $1`, id).Scan(&code, &name)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "scenario not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "scenario_code": code, "scenario_name": name})
}

func (h *RiskHandler) UpdateScenario(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Status *string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	_, err := h.db.Exec(`UPDATE risk_scenarios SET status=COALESCE($1,status), updated_at=$2 WHERE id=$3`, input.Status, time.Now(), id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *RiskHandler) DeleteScenario(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM risk_scenarios WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// Mitigations
// =============================================================================
func (h *RiskHandler) ListMitigations(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT id, project_id, action_number, action_code, action_name, action_type, assigned_to, budget, due_date, effectiveness, status FROM risk_mitigation_actions WHERE 1=1`
	args := []interface{}{}
	argIdx := 1
	if projectID != "" { query += fmt.Sprintf(" AND project_id = $%d", argIdx); args = append(args, projectID) }
	query += " ORDER BY action_number"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, pid, code, aname, atype, assignee, eff, st string
		var num int
		var budget float64
		var dueDate sql.NullString
		if err := rows.Scan(&id, &pid, &num, &code, &aname, &atype, &assignee, &budget, &dueDate, &eff, &st); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		item := map[string]interface{}{
			"id": id, "project_id": pid, "action_number": num, "action_code": code,
			"action_name": aname, "action_type": atype, "assigned_to": assignee,
			"budget": budget, "effectiveness": eff, "status": st,
		}
		if dueDate.Valid { item["due_date"] = dueDate.String }
		items = append(items, item)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *RiskHandler) CreateMitigation(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID  string  `json:"project_id"`
		ActionCode string  `json:"action_code"`
		ActionName string  `json:"action_name"`
		ActionType string  `json:"action_type"`
		AssignedTo string  `json:"assigned_to"`
		Budget     float64 `json:"budget"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO risk_mitigation_actions (id, project_id, action_number, action_code, action_name, action_type, assigned_to, budget, created_at, updated_at) VALUES ($1,$2,(SELECT COALESCE(MAX(action_number),0)+1 FROM risk_mitigation_actions WHERE project_id=$2),$3,$4,$5,$6,$7,$8,$9)`,
		id, input.ProjectID, input.ActionCode, input.ActionName, input.ActionType, input.AssignedTo, input.Budget, now, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *RiskHandler) GetMitigation(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var code, name string
	err := h.db.QueryRow(`SELECT action_code, action_name FROM risk_mitigation_actions WHERE id = $1`, id).Scan(&code, &name)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "mitigation action not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "action_code": code, "action_name": name})
}

func (h *RiskHandler) UpdateMitigation(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Status         *string  `json:"status"`
		Effectiveness  *string  `json:"effectiveness"`
		ResidualProb   *float64 `json:"residual_probability"`
		ResidualImpact *float64 `json:"residual_impact"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	_, err := h.db.Exec(`UPDATE risk_mitigation_actions SET status=COALESCE($1,status), effectiveness=COALESCE($2,effectiveness), residual_probability=COALESCE($3,residual_probability), residual_impact=COALESCE($4,residual_impact), updated_at=$5 WHERE id=$6`,
		input.Status, input.Effectiveness, input.ResidualProb, input.ResidualImpact, time.Now(), id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *RiskHandler) DeleteMitigation(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM risk_mitigation_actions WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// Escalations
// =============================================================================
func (h *RiskHandler) ListEscalations(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT id, project_id, escalation_number, escalation_code, title, reason, escalated_to, escalated_by, escalated_at, decision, status FROM risk_escalation WHERE 1=1`
	args := []interface{}{}
	argIdx := 1
	if projectID != "" { query += fmt.Sprintf(" AND project_id = $%d", argIdx); args = append(args, projectID) }
	query += " ORDER BY escalated_at DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, pid, code, title, reason, escTo, escBy, decision, st string
		var num int
		var escAt time.Time
		if err := rows.Scan(&id, &pid, &num, &code, &title, &reason, &escTo, &escBy, &escAt, &decision, &st); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		items = append(items, map[string]interface{}{
			"id": id, "project_id": pid, "escalation_number": num, "escalation_code": code,
			"title": title, "reason": reason, "escalated_to": escTo, "escalated_by": escBy,
			"escalated_at": escAt, "decision": decision, "status": st,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *RiskHandler) CreateEscalation(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID     string `json:"project_id"`
		EscalationCode string `json:"escalation_code"`
		Title         string `json:"title"`
		Reason        string `json:"reason"`
		EscalatedTo   string `json:"escalated_to"`
		EscalatedBy   string `json:"escalated_by"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO risk_escalation (id, project_id, escalation_number, escalation_code, title, reason, escalated_to, escalated_by, escalated_at, created_at, updated_at) VALUES ($1,$2,(SELECT COALESCE(MAX(escalation_number),0)+1 FROM risk_escalation WHERE project_id=$2),$3,$4,$5,$6,$7,$8,$9,$10)`,
		id, input.ProjectID, input.EscalationCode, input.Title, input.Reason, input.EscalatedTo, input.EscalatedBy, now, now, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *RiskHandler) GetEscalation(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var code, title string
	err := h.db.QueryRow(`SELECT escalation_code, title FROM risk_escalation WHERE id = $1`, id).Scan(&code, &title)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "escalation not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "escalation_code": code, "title": title})
}

func (h *RiskHandler) UpdateEscalation(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Status   *string `json:"status"`
		Decision *string `json:"decision"`
		Response *string `json:"response"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	_, err := h.db.Exec(`UPDATE risk_escalation SET status=COALESCE($1,status), decision=COALESCE($2,decision), response=COALESCE($3,response), responded_at=CASE WHEN $1 IS NOT NULL AND $1 = 'responded' THEN $4 ELSE responded_at END, updated_at=$4 WHERE id=$5`,
		input.Status, input.Decision, input.Response, time.Now(), id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *RiskHandler) DeleteEscalation(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM risk_escalation WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// Dashboard
// =============================================================================
func (h *RiskHandler) ListDashboard(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT id, project_id, snapshot_date, total_risks, open_risks, extreme_risks, high_risks, medium_risks, low_risks, threats, opportunities, risk_exposure, mitigation_progress_pct FROM risk_dashboard WHERE 1=1`
	args := []interface{}{}
	argIdx := 1
	if projectID != "" { query += fmt.Sprintf(" AND project_id = $%d", argIdx); args = append(args, projectID) }
	query += " ORDER BY snapshot_date DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, pid string
		var snapDate time.Time
		var total, open, extreme, high, med, low, threats, opps int
		var exposure, mitPct float64
		if err := rows.Scan(&id, &pid, &snapDate, &total, &open, &extreme, &high, &med, &low, &threats, &opps, &exposure, &mitPct); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		items = append(items, map[string]interface{}{
			"id": id, "project_id": pid, "snapshot_date": snapDate,
			"total_risks": total, "open_risks": open, "extreme_risks": extreme,
			"high_risks": high, "medium_risks": med, "low_risks": low,
			"threats": threats, "opportunities": opps,
			"risk_exposure": exposure, "mitigation_progress_pct": mitPct,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *RiskHandler) CreateDashboard(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID string `json:"project_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO risk_dashboard (id, project_id, snapshot_date, created_at) VALUES ($1,$2,$3,$4)`,
		id, input.ProjectID, now, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *RiskHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var pid string
	var snapDate time.Time
	err := h.db.QueryRow(`SELECT project_id, snapshot_date FROM risk_dashboard WHERE id = $1`, id).Scan(&pid, &snapDate)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "dashboard entry not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "project_id": pid, "snapshot_date": snapDate})
}

func (h *RiskHandler) UpdateDashboard(w http.ResponseWriter, r *http.Request) {
	// Dashboard is typically auto-generated, but allow updates
	id := chi.URLParam(r, "id")
	var input struct {
		RiskExposure *float64 `json:"risk_exposure"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	_, err := h.db.Exec(`UPDATE risk_dashboard SET risk_exposure=COALESCE($1,risk_exposure), updated_at=$2 WHERE id=$3`, input.RiskExposure, time.Now(), id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *RiskHandler) DeleteDashboard(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM risk_dashboard WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// Summary
// =============================================================================
func (h *RiskHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT project_id, open_risks, total_risks, extreme_risks, high_risks, threats, opportunities, total_threat_cost, pending_mitigations, completed_mitigations, active_escalations, mc_runs_completed, analyzed_scenarios FROM risk_summary`
	args := []interface{}{}
	argIdx := 1
	if projectID != "" { query += fmt.Sprintf(" WHERE project_id = $%d", argIdx); args = append(args, projectID) }

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var pid string
		var open, total, extreme, high, threats, opps, pendMit, compMit, activeEsc, mcRuns, scenarios int
		var threatCost float64
		if err := rows.Scan(&pid, &open, &total, &extreme, &high, &threats, &opps, &threatCost, &pendMit, &compMit, &activeEsc, &mcRuns, &scenarios); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		items = append(items, map[string]interface{}{
			"project_id": pid, "open_risks": open, "total_risks": total,
			"extreme_risks": extreme, "high_risks": high, "threats": threats,
			"opportunities": opps, "total_threat_cost": threatCost,
			"pending_mitigations": pendMit, "completed_mitigations": compMit,
			"active_escalations": activeEsc, "mc_runs_completed": mcRuns,
			"analyzed_scenarios": scenarios,
		})
	}
	respondJSON(w, http.StatusOK, items)
}