package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// PMHandler handles Project Management module endpoints
type PMHandler struct {
	db *sql.DB
}

func NewPMHandler(db *sql.DB) *PMHandler {
	return &PMHandler{db: db}
}

func (h *PMHandler) RegisterRoutes(r chi.Router) {
	r.Route("/pm", func(r chi.Router) {
		// Projects
		r.Get("/projects", h.ListProjects)
		r.Post("/projects", h.CreateProject)
		r.Get("/projects/{id}", h.GetProject)
		r.Put("/projects/{id}", h.UpdateProject)
		r.Delete("/projects/{id}", h.DeleteProject)

		// WBS Items
		r.Get("/wbs-items", h.ListWBSItems)
		r.Post("/wbs-items", h.CreateWBSItem)
		r.Get("/wbs-items/{id}", h.GetWBSItem)
		r.Put("/wbs-items/{id}", h.UpdateWBSItem)
		r.Delete("/wbs-items/{id}", h.DeleteWBSItem)

		// Milestones
		r.Get("/milestones", h.ListMilestones)
		r.Post("/milestones", h.CreateMilestone)
		r.Get("/milestones/{id}", h.GetMilestone)
		r.Put("/milestones/{id}", h.UpdateMilestone)
		r.Delete("/milestones/{id}", h.DeleteMilestone)

		// Phases
		r.Get("/phases", h.ListPhases)
		r.Post("/phases", h.CreatePhase)
		r.Get("/phases/{id}", h.GetPhase)
		r.Put("/phases/{id}", h.UpdatePhase)
		r.Delete("/phases/{id}", h.DeletePhase)

		// Team
		r.Get("/team", h.ListTeam)
		r.Post("/team", h.AddTeamMember)
		r.Delete("/team/{id}", h.RemoveTeamMember)

		// Portfolios
		r.Get("/portfolios", h.ListPortfolios)
		r.Post("/portfolios", h.CreatePortfolio)
		r.Get("/portfolios/{id}", h.GetPortfolio)
		r.Put("/portfolios/{id}", h.UpdatePortfolio)
		r.Delete("/portfolios/{id}", h.DeletePortfolio)

		// Portfolio Projects
		r.Post("/portfolio-projects", h.AddProjectToPortfolio)
		r.Delete("/portfolio-projects", h.RemoveProjectFromPortfolio)

		// Risks
		r.Get("/risks", h.ListRisks)
		r.Post("/risks", h.CreateRisk)
		r.Get("/risks/{id}", h.GetRisk)
		r.Put("/risks/{id}", h.UpdateRisk)
		r.Delete("/risks/{id}", h.DeleteRisk)

		// Changes
		r.Get("/changes", h.ListChanges)
		r.Post("/changes", h.CreateChange)
		r.Get("/changes/{id}", h.GetChange)
		r.Put("/changes/{id}", h.UpdateChange)
		r.Delete("/changes/{id}", h.DeleteChange)

		// Lessons
		r.Get("/lessons", h.ListLessons)
		r.Post("/lessons", h.CreateLesson)
		r.Get("/lessons/{id}", h.GetLesson)
		r.Put("/lessons/{id}", h.UpdateLesson)
		r.Delete("/lessons/{id}", h.DeleteLesson)

		// Dashboard summary
		r.Get("/dashboard", h.GetDashboardSummary)
	})
}

// ============================================================================
// Projects
// ============================================================================

func (h *PMHandler) ListProjects(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	projectType := r.URL.Query().Get("project_type")

	query := `SELECT id, code, name, description, project_type, status, phase,
		client_id, owner_id, country, city, region, address,
		start_date, end_date, duration_days,
		budget_total, budget_currency, contingency, contingency_pct,
		total_length_km, total_area_m2, total_volume_m3,
		project_manager_id, sponsor_id,
		risk_class, complexity, confidentiality,
		notes, created_by, created_at, updated_at
		FROM projects WHERE 1=1`

	var args []interface{}
	argIdx := 1

	if status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	}
	if projectType != "" {
		query += fmt.Sprintf(" AND project_type = $%d", argIdx)
		args = append(args, projectType)
		argIdx++
	}
	query += " ORDER BY code"

	rows, err := h.db.Query(query, args...)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	projects := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id int64
		var code, name, projectType, status, phase string
		var description, clientID, ownerID, country, city, region, address sql.NullString
		var startDate, endDate sql.NullString
		var durationDays sql.NullInt64
		var budgetTotal, contingency, contingencyPct sql.NullFloat64
		var budgetCurrency string
		var totalLengthKm, totalAreaM2, totalVolumeM3 sql.NullFloat64
		var pmID, sponsorID sql.NullInt64
		var riskClass, complexity, confidentiality sql.NullString
		var notes, createdBy sql.NullString
		var createdAt, updatedAt time.Time

		err := rows.Scan(&id, &code, &name, &description, &projectType, &status, &phase,
			&clientID, &ownerID, &country, &city, &region, &address,
			&startDate, &endDate, &durationDays,
			&budgetTotal, &budgetCurrency, &contingency, &contingencyPct,
			&totalLengthKm, &totalAreaM2, &totalVolumeM3,
			&pmID, &sponsorID,
			&riskClass, &complexity, &confidentiality,
			&notes, &createdBy, &createdAt, &updatedAt)
		if err != nil {
			log.Printf("[PM] Scan error: %v", err)
			continue
		}

		p := map[string]interface{}{
			"id":                id,
			"code":              code,
			"name":              name,
			"project_type":      projectType,
			"status":            status,
			"phase":             phase,
			"budget_total":      nullFloat64(budgetTotal),
			"budget_currency":   budgetCurrency,
			"country":           nullString(country),
			"city":              nullString(city),
			"start_date":        nullString(startDate),
			"end_date":          nullString(endDate),
			"duration_days":     nullInt64(durationDays),
			"risk_class":        nullString(riskClass),
			"complexity":        nullString(complexity),
			"created_at":        createdAt,
		}
		projects = append(projects, p)
	}
	respondJSON(w, http.StatusOK, projects)
}

func (h *PMHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Code          string   `json:"code"`
		Name          string   `json:"name"`
		Description   *string  `json:"description"`
		ProjectType   string   `json:"project_type"`
		Status        string   `json:"status"`
		Phase         string   `json:"phase"`
		Country       *string  `json:"country"`
		City          *string  `json:"city"`
		StartDate     *string  `json:"start_date"`
		EndDate       *string  `json:"end_date"`
		DurationDays  *int     `json:"duration_days"`
		BudgetTotal   *float64 `json:"budget_total"`
		BudgetCurrency string  `json:"budget_currency"`
		RiskClass     *string  `json:"risk_class"`
		Complexity    *string  `json:"complexity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if input.BudgetCurrency == "" {
		input.BudgetCurrency = "USD"
	}
	if input.Status == "" {
		input.Status = "lead"
	}
	if input.Phase == "" {
		input.Phase = "feasibility"
	}

	var id int64
	err := h.db.QueryRow(`INSERT INTO projects
		(code, name, description, project_type, status, phase,
		 country, city, start_date, end_date, duration_days,
		 budget_total, budget_currency, risk_class, complexity)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
		RETURNING id`,
		input.Code, input.Name, input.Description, input.ProjectType, input.Status, input.Phase,
		input.Country, input.City, input.StartDate, input.EndDate, input.DurationDays,
		input.BudgetTotal, input.BudgetCurrency, input.RiskClass, input.Complexity).Scan(&id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]interface{}{"id": id, "code": input.Code})
}

func (h *PMHandler) GetProject(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	respondJSON(w, http.StatusOK, map[string]string{"id": id, "message": "get project by id"})
}

func (h *PMHandler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *PMHandler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// ============================================================================
// WBS Items
// ============================================================================

func (h *PMHandler) ListWBSItems(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT id, project_id, parent_id, wbs_code, name, description,
		wbs_level, sort_order, is_leaf,
		planned_start, planned_end, planned_duration, planned_cost, planned_hours,
		actual_start, actual_end, actual_cost, actual_hours, progress_pct,
		responsible_id, status, notes, created_at, updated_at
		FROM wbs_items`
	var rows *sql.Rows
	var err error
	if projectID != "" {
		rows, err = h.db.Query(query+` WHERE project_id = $1 ORDER BY wbs_code`, projectID)
	} else {
		rows, err = h.db.Query(query + ` ORDER BY project_id, wbs_code`)
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, projectID int64
		var parentID sql.NullInt64
		var wbsCode, name string
		var description sql.NullString
		var wbsLevel, sortOrder int
		var isLeaf bool
		var plannedStart, plannedEnd, actualStart, actualEnd sql.NullString
		var plannedDuration sql.NullInt64
		var plannedCost, plannedHours, actualCost, actualHours, progressPct sql.NullFloat64
		var responsibleID sql.NullInt64
		var status string
		var notes sql.NullString
		var createdAt, updatedAt time.Time

		if err := rows.Scan(&id, &projectID, &parentID, &wbsCode, &name, &description,
			&wbsLevel, &sortOrder, &isLeaf,
			&plannedStart, &plannedEnd, &plannedDuration, &plannedCost, &plannedHours,
			&actualStart, &actualEnd, &actualCost, &actualHours, &progressPct,
			&responsibleID, &status, &notes, &createdAt, &updatedAt); err != nil {
			log.Printf("[PM] WBS scan error: %v", err)
			continue
		}
		items = append(items, map[string]interface{}{
			"id":           id,
			"project_id":   projectID,
			"parent_id":    nullInt64(parentID),
			"wbs_code":     wbsCode,
			"name":         name,
			"wbs_level":    wbsLevel,
			"sort_order":   sortOrder,
			"is_leaf":      isLeaf,
			"progress_pct": nullFloat64(progressPct),
			"status":       status,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *PMHandler) CreateWBSItem(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusCreated, map[string]string{"status": "created"})
}

func (h *PMHandler) GetWBSItem(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *PMHandler) UpdateWBSItem(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *PMHandler) DeleteWBSItem(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// ============================================================================
// Milestones
// ============================================================================

func (h *PMHandler) ListMilestones(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT id, project_id, milestone_code, name, description,
		milestone_type, category,
		planned_date, forecast_date, actual_date,
		status, delay_days, weight_pct, is_gate, amount, amount_currency,
		responsible_id, notes, created_at
		FROM project_milestones`
	var rows *sql.Rows
	var err error
	if projectID != "" {
		rows, err = h.db.Query(query+` WHERE project_id = $1 ORDER BY planned_date`, projectID)
	} else {
		rows, err = h.db.Query(query + ` ORDER BY project_id, planned_date`)
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	milestones := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, projectID int64
		var milestoneCode, name string
		var description, category, forecastDate, actualDate sql.NullString
		var milestoneType string
		var plannedDate string
		var status string
		var delayDays sql.NullInt64
		var weightPct sql.NullFloat64
		var isGate bool
		var amount sql.NullFloat64
		var amountCurrency sql.NullString
		var responsibleID sql.NullInt64
		var notes sql.NullString
		var createdAt time.Time

		if err := rows.Scan(&id, &projectID, &milestoneCode, &name, &description,
			&milestoneType, &category,
			&plannedDate, &forecastDate, &actualDate,
			&status, &delayDays, &weightPct, &isGate, &amount, &amountCurrency,
			&responsibleID, &notes, &createdAt); err != nil {
			log.Printf("[PM] Milestone scan error: %v", err)
			continue
		}
		milestones = append(milestones, map[string]interface{}{
			"id":              id,
			"project_id":      projectID,
			"milestone_code":  milestoneCode,
			"name":            name,
			"milestone_type":  milestoneType,
			"planned_date":    plannedDate,
			"actual_date":     nullString(actualDate),
			"status":          status,
			"weight_pct":      nullFloat64(weightPct),
			"is_gate":         isGate,
		})
	}
	respondJSON(w, http.StatusOK, milestones)
}

func (h *PMHandler) CreateMilestone(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusCreated, map[string]string{"status": "created"})
}

func (h *PMHandler) GetMilestone(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *PMHandler) UpdateMilestone(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *PMHandler) DeleteMilestone(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// ============================================================================
// Phases
// ============================================================================

func (h *PMHandler) ListPhases(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT id, project_id, phase_code, name, description, sort_order,
		planned_start, planned_end, actual_start, actual_end,
		budget_amount, actual_amount, status, deliverables, completion_pct, notes, created_at
		FROM project_phases`
	var rows *sql.Rows
	var err error
	if projectID != "" {
		rows, err = h.db.Query(query+` WHERE project_id = $1 ORDER BY sort_order`, projectID)
	} else {
		rows, err = h.db.Query(query + ` ORDER BY project_id, sort_order`)
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	phases := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, projectID int64
		var phaseCode, name string
		var description, plannedStart, plannedEnd, actualStart, actualEnd sql.NullString
		var sortOrder int
		var budgetAmount, actualAmount sql.NullFloat64
		var status string
		var deliverables, notes sql.NullString
		var completionPct sql.NullFloat64
		var createdAt time.Time

		if err := rows.Scan(&id, &projectID, &phaseCode, &name, &description, &sortOrder,
			&plannedStart, &plannedEnd, &actualStart, &actualEnd,
			&budgetAmount, &actualAmount, &status, &deliverables, &completionPct, &notes, &createdAt); err != nil {
			log.Printf("[PM] Phase scan error: %v", err)
			continue
		}
		phases = append(phases, map[string]interface{}{
			"id":              id,
			"project_id":      projectID,
			"phase_code":      phaseCode,
			"name":            name,
			"sort_order":      sortOrder,
			"status":          status,
			"completion_pct":  nullFloat64(completionPct),
			"budget_amount":   nullFloat64(budgetAmount),
		})
	}
	respondJSON(w, http.StatusOK, phases)
}

func (h *PMHandler) CreatePhase(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusCreated, map[string]string{"status": "created"})
}

func (h *PMHandler) GetPhase(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *PMHandler) UpdatePhase(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *PMHandler) DeletePhase(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// ============================================================================
// Team
// ============================================================================

func (h *PMHandler) ListTeam(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT id, project_id, employee_id, role, role_category,
		start_date, end_date, allocation_pct, is_key, hourly_rate, notes, created_at
		FROM project_team`
	var rows *sql.Rows
	var err error
	if projectID != "" {
		rows, err = h.db.Query(query+` WHERE project_id = $1`, projectID)
	} else {
		rows, err = h.db.Query(query + ` ORDER BY project_id, role`)
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	members := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, projectID int64
		var employeeID int64
		var role, roleCategory string
		var startDate string
		var endDate sql.NullString
		var allocationPct sql.NullFloat64
		var isKey bool
		var hourlyRate sql.NullFloat64
		var notes sql.NullString
		var createdAt time.Time

		if err := rows.Scan(&id, &projectID, &employeeID, &role, &roleCategory,
			&startDate, &endDate, &allocationPct, &isKey, &hourlyRate, &notes, &createdAt); err != nil {
			log.Printf("[PM] Team scan error: %v", err)
			continue
		}
		members = append(members, map[string]interface{}{
			"id":            id,
			"project_id":    projectID,
			"employee_id":   employeeID,
			"role":          role,
			"role_category": roleCategory,
			"allocation_pct": nullFloat64(allocationPct),
			"is_key":        isKey,
		})
	}
	respondJSON(w, http.StatusOK, members)
}

func (h *PMHandler) AddTeamMember(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusCreated, map[string]string{"status": "created"})
}

func (h *PMHandler) RemoveTeamMember(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// ============================================================================
// Portfolios
// ============================================================================

func (h *PMHandler) ListPortfolios(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(`SELECT id, code, name, description, portfolio_type,
		parent_id, owner_id, budget_total, budget_currency, status,
		start_date, end_date, notes, created_at
		FROM project_portfolio ORDER BY code`)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	portfolios := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id int64
		var code, name string
		var description, portfolioType sql.NullString
		var parentID, ownerID sql.NullInt64
		var budgetTotal sql.NullFloat64
		var budgetCurrency sql.NullString
		var status string
		var startDate, endDate, notes sql.NullString
		var createdAt time.Time

		if err := rows.Scan(&id, &code, &name, &description, &portfolioType,
			&parentID, &ownerID, &budgetTotal, &budgetCurrency, &status,
			&startDate, &endDate, &notes, &createdAt); err != nil {
			log.Printf("[PM] Portfolio scan error: %v", err)
			continue
		}
		portfolios = append(portfolios, map[string]interface{}{
			"id":             id,
			"code":           code,
			"name":           name,
			"portfolio_type": nullString(portfolioType),
			"parent_id":      nullInt64(parentID),
			"budget_total":   nullFloat64(budgetTotal),
			"status":         status,
		})
	}
	respondJSON(w, http.StatusOK, portfolios)
}

func (h *PMHandler) CreatePortfolio(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusCreated, map[string]string{"status": "created"})
}

func (h *PMHandler) GetPortfolio(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *PMHandler) UpdatePortfolio(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *PMHandler) DeletePortfolio(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (h *PMHandler) AddProjectToPortfolio(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusCreated, map[string]string{"status": "created"})
}

func (h *PMHandler) RemoveProjectFromPortfolio(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// ============================================================================
// Risks
// ============================================================================

func (h *PMHandler) ListRisks(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT id, project_id, risk_code, name, description,
		risk_category, risk_type,
		probability, impact, probability_score, impact_score, risk_score,
		potential_cost, mitigation_cost, residual_cost,
		mitigation_strategy, mitigation_plan, contingency_plan,
		status, owner_id, target_date, closed_date, notes, created_at, updated_at
		FROM project_risks`
	var rows *sql.Rows
	var err error
	if projectID != "" {
		rows, err = h.db.Query(query+` WHERE project_id = $1 ORDER BY risk_score DESC`, projectID)
	} else {
		rows, err = h.db.Query(query + ` ORDER BY project_id, risk_score DESC`)
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	risks := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, projectID int64
		var riskCode, name string
		var description sql.NullString
		var riskCategory, riskType string
		var probability, impact string
		var probScore, impactScore, riskScore int
		var potentialCost, mitigationCost, residualCost sql.NullFloat64
		var mitigationStrategy, mitigationPlan, contingencyPlan sql.NullString
		var status string
		var ownerID, targetDate, closedDate sql.NullString
		var notes sql.NullString
		var createdAt, updatedAt time.Time

		if err := rows.Scan(&id, &projectID, &riskCode, &name, &description,
			&riskCategory, &riskType,
			&probability, &impact, &probScore, &impactScore, &riskScore,
			&potentialCost, &mitigationCost, &residualCost,
			&mitigationStrategy, &mitigationPlan, &contingencyPlan,
			&status, &ownerID, &targetDate, &closedDate, &notes, &createdAt, &updatedAt); err != nil {
			log.Printf("[PM] Risk scan error: %v", err)
			continue
		}
		risks = append(risks, map[string]interface{}{
			"id":                 id,
			"project_id":         projectID,
			"risk_code":          riskCode,
			"name":               name,
			"risk_category":      riskCategory,
			"probability_score":  probScore,
			"impact_score":       impactScore,
			"risk_score":         riskScore,
			"probability":        probability,
			"impact":             impact,
			"potential_cost":     nullFloat64(potentialCost),
			"mitigation_strategy": nullString(mitigationStrategy),
			"status":             status,
		})
	}
	respondJSON(w, http.StatusOK, risks)
}

func (h *PMHandler) CreateRisk(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusCreated, map[string]string{"status": "created"})
}

func (h *PMHandler) GetRisk(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *PMHandler) UpdateRisk(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *PMHandler) DeleteRisk(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// ============================================================================
// Changes
// ============================================================================

func (h *PMHandler) ListChanges(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT id, project_id, change_number, change_type, source,
		description, justification,
		cost_impact, schedule_impact, scope_change,
		status, submitted_by, submitted_at, approved_by, approved_at,
		document_path, notes, created_at
		FROM project_changes`
	var rows *sql.Rows
	var err error
	if projectID != "" {
		rows, err = h.db.Query(query+` WHERE project_id = $1 ORDER BY change_number`, projectID)
	} else {
		rows, err = h.db.Query(query + ` ORDER BY project_id, change_number`)
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	changes := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, projectID int64
		var changeNumber, changeType, source string
		var description, justification, scopeChange sql.NullString
		var costImpact, scheduleImpact sql.NullFloat64
		var status string
		var submittedBy, approvedBy sql.NullInt64
		var submittedAt, approvedAt sql.NullString
		var documentPath, notes sql.NullString
		var createdAt time.Time

		if err := rows.Scan(&id, &projectID, &changeNumber, &changeType, &source,
			&description, &justification,
			&costImpact, &scheduleImpact, &scopeChange,
			&status, &submittedBy, &submittedAt, &approvedBy, &approvedAt,
			&documentPath, &notes, &createdAt); err != nil {
			log.Printf("[PM] Change scan error: %v", err)
			continue
		}
		changes = append(changes, map[string]interface{}{
			"id":             id,
			"project_id":     projectID,
			"change_number":  changeNumber,
			"change_type":    changeType,
			"source":         source,
			"description":    description.String,
			"cost_impact":    nullFloat64(costImpact),
			"schedule_impact": nullFloat64(scheduleImpact),
			"status":         status,
		})
	}
	respondJSON(w, http.StatusOK, changes)
}

func (h *PMHandler) CreateChange(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusCreated, map[string]string{"status": "created"})
}

func (h *PMHandler) GetChange(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *PMHandler) UpdateChange(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *PMHandler) DeleteChange(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// ============================================================================
// Lessons
// ============================================================================

func (h *PMHandler) ListLessons(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT id, project_id, category, title, description,
		root_cause, impact, recommendation,
		is_positive, severity, status, author_id, created_at
		FROM project_lessons`
	var rows *sql.Rows
	var err error
	if projectID != "" {
		rows, err = h.db.Query(query+` WHERE project_id = $1`, projectID)
	} else {
		rows, err = h.db.Query(query + ` ORDER BY project_id, created_at DESC`)
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	lessons := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, projectID int64
		var category, title, description string
		var rootCause, impact, recommendation sql.NullString
		var isPositive bool
		var severity, status string
		var authorID sql.NullInt64
		var createdAt time.Time

		if err := rows.Scan(&id, &projectID, &category, &title, &description,
			&rootCause, &impact, &recommendation,
			&isPositive, &severity, &status, &authorID, &createdAt); err != nil {
			log.Printf("[PM] Lesson scan error: %v", err)
			continue
		}
		lessons = append(lessons, map[string]interface{}{
			"id":          id,
			"project_id":  projectID,
			"category":    category,
			"title":       title,
			"is_positive": isPositive,
			"severity":    severity,
			"status":      status,
		})
	}
	respondJSON(w, http.StatusOK, lessons)
}

func (h *PMHandler) CreateLesson(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusCreated, map[string]string{"status": "created"})
}

func (h *PMHandler) GetLesson(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *PMHandler) UpdateLesson(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *PMHandler) DeleteLesson(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// ============================================================================
// Dashboard Summary
// ============================================================================

func (h *PMHandler) GetDashboardSummary(w http.ResponseWriter, r *http.Request) {
	summary := map[string]interface{}{
		"total_projects":   0,
		"total_budget":     0.0,
		"active_projects":  0,
		"total_risks":      0,
		"total_milestones": 0,
	}

	var totalProjects, totalBudget, activeProjects, totalRisks, totalMilestones int
	h.db.QueryRow(`SELECT COUNT(*) FROM projects`).Scan(&totalProjects)
	h.db.QueryRow(`SELECT COALESCE(SUM(budget_total), 0) FROM projects`).Scan(&totalBudget)
	h.db.QueryRow(`SELECT COUNT(*) FROM projects WHERE status NOT IN ('closed','lead')`).Scan(&activeProjects)
	h.db.QueryRow(`SELECT COUNT(*) FROM project_risks`).Scan(&totalRisks)
	h.db.QueryRow(`SELECT COUNT(*) FROM project_milestones`).Scan(&totalMilestones)
	summary["total_projects"] = totalProjects
	summary["total_budget"] = totalBudget
	summary["active_projects"] = activeProjects
	summary["total_risks"] = totalRisks
	summary["total_milestones"] = totalMilestones

	respondJSON(w, http.StatusOK, summary)
}

// ============================================================================
// Null helpers
// ============================================================================

func nullString(ns sql.NullString) interface{} {
	if ns.Valid {
		return ns.String
	}
	return nil
}

func nullInt64(ni sql.NullInt64) interface{} {
	if ni.Valid {
		return ni.Int64
	}
	return nil
}

func nullFloat64(nf sql.NullFloat64) interface{} {
	if nf.Valid {
		return nf.Float64
	}
	return nil
}

// Ensure uuid is used (referenced in imports)
var _ = uuid.New
