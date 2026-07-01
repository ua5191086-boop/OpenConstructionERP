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

// ChangeHandler handles Change Management module endpoints
type ChangeHandler struct {
	db *sql.DB
}

func NewChangeHandler(db *sql.DB) *ChangeHandler {
	return &ChangeHandler{db: db}
}

func (h *ChangeHandler) RegisterRoutes(r chi.Router) {
	r.Route("/change", func(r chi.Router) {
		r.Get("/requests", h.ListRequests)
		r.Post("/requests", h.CreateRequest)
		r.Get("/requests/{id}", h.GetRequest)
		r.Put("/requests/{id}", h.UpdateRequest)
		r.Delete("/requests/{id}", h.DeleteRequest)

		r.Get("/orders", h.ListOrders)
		r.Post("/orders", h.CreateOrder)
		r.Get("/orders/{id}", h.GetOrder)
		r.Put("/orders/{id}", h.UpdateOrder)
		r.Delete("/orders/{id}", h.DeleteOrder)

		r.Get("/impact-analysis", h.ListImpactAnalysis)
		r.Post("/impact-analysis", h.CreateImpactAnalysis)
		r.Get("/impact-analysis/{id}", h.GetImpactAnalysis)
		r.Put("/impact-analysis/{id}", h.UpdateImpactAnalysis)
		r.Delete("/impact-analysis/{id}", h.DeleteImpactAnalysis)

		r.Get("/approval-workflow", h.ListApprovalWorkflow)
		r.Post("/approval-workflow", h.CreateApprovalWorkflow)
		r.Get("/approval-workflow/{id}", h.GetApprovalWorkflow)
		r.Put("/approval-workflow/{id}", h.UpdateApprovalWorkflow)
		r.Delete("/approval-workflow/{id}", h.DeleteApprovalWorkflow)

		r.Get("/change-log", h.ListChangeLog)
		r.Post("/change-log", h.CreateChangeLog)
		r.Get("/change-log/{id}", h.GetChangeLog)
		r.Delete("/change-log/{id}", h.DeleteChangeLog)

		r.Get("/summary", h.GetSummary)
	})
}

// =============================================================================
// Change Requests
// =============================================================================
func (h *ChangeHandler) ListRequests(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	status := r.URL.Query().Get("status")
	crType := r.URL.Query().Get("cr_type")

	query := `SELECT id, project_id, cr_number, cr_code, cr_name, cr_type, source, priority, description, reason, proposed_by, proposed_date, required_by_date, status, created_at FROM change_requests WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if projectID != "" { query += fmt.Sprintf(" AND project_id = $%d", argIdx); argIdx++; args = append(args, projectID) }
	if status != "" { query += fmt.Sprintf(" AND status = $%d", argIdx); argIdx++; args = append(args, status) }
	if crType != "" { query += fmt.Sprintf(" AND cr_type = $%d", argIdx); argIdx++; args = append(args, crType) }
	query += " ORDER BY cr_number DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, pid, code, cname, ctype, src, prio, desc, reason, proposedBy, st string
		var num int
		var proposedDate, requiredDate, createdAt sql.NullString
		if err := rows.Scan(&id, &pid, &num, &code, &cname, &ctype, &src, &prio, &desc, &reason, &proposedBy, &proposedDate, &requiredDate, &st, &createdAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		item := map[string]interface{}{
			"id": id, "project_id": pid, "cr_number": num, "cr_code": code,
			"cr_name": cname, "cr_type": ctype, "source": src, "priority": prio,
			"description": desc, "reason": reason, "proposed_by": proposedBy, "status": st,
		}
		if proposedDate.Valid { item["proposed_date"] = proposedDate.String }
		if requiredDate.Valid { item["required_by_date"] = requiredDate.String }
		if createdAt.Valid { item["created_at"] = createdAt.String }
		items = append(items, item)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *ChangeHandler) CreateRequest(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID    string `json:"project_id"`
		CRCode       string `json:"cr_code"`
		CRName       string `json:"cr_name"`
		CRType       string `json:"cr_type"`
		Source       string `json:"source"`
		Priority     string `json:"priority"`
		Description  string `json:"description"`
		Reason       string `json:"reason"`
		ProposedBy   string `json:"proposed_by"`
		ProposedDate string `json:"proposed_date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO change_requests (id, project_id, cr_number, cr_code, cr_name, cr_type, source, priority, description, reason, proposed_by, proposed_date, status, created_at, updated_at) VALUES ($1,$2,(SELECT COALESCE(MAX(cr_number),0)+1 FROM change_requests WHERE project_id=$2),$3,$4,$5,$6,$7,$8,$9,$10,$11,'draft',$12,$13)`,
		id, input.ProjectID, input.CRCode, input.CRName, input.CRType, input.Source, input.Priority, input.Description, input.Reason, input.ProposedBy, input.ProposedDate, now, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *ChangeHandler) GetRequest(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var code, name string
	err := h.db.QueryRow(`SELECT cr_code, cr_name FROM change_requests WHERE id = $1`, id).Scan(&code, &name)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "change request not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "cr_code": code, "cr_name": name})
}

func (h *ChangeHandler) UpdateRequest(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Status *string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	now := time.Now()
	_, err := h.db.Exec(`UPDATE change_requests SET status=COALESCE($1,status), updated_at=$2 WHERE id=$3`, input.Status, now, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *ChangeHandler) DeleteRequest(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM change_requests WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// Change Orders
// =============================================================================
func (h *ChangeHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT id, project_id, co_number, co_code, co_name, co_type, scope_change, cost_change, schedule_change_days, contractor_name, approved_by, status FROM change_orders WHERE 1=1`
	args := []interface{}{}
	argIdx := 1
	if projectID != "" { query += fmt.Sprintf(" AND project_id = $%d", argIdx); args = append(args, projectID) }
	query += " ORDER BY co_number DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, pid, code, coname, cotype, scope, contName, appBy, st string
		var num, schedDays int
		var cost float64
		if err := rows.Scan(&id, &pid, &num, &code, &coname, &cotype, &scope, &cost, &schedDays, &contName, &appBy, &st); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		items = append(items, map[string]interface{}{
			"id": id, "project_id": pid, "co_number": num, "co_code": code,
			"co_name": coname, "co_type": cotype, "scope_change": scope,
			"cost_change": cost, "schedule_change_days": schedDays,
			"contractor_name": contName, "approved_by": appBy, "status": st,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *ChangeHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID        string  `json:"project_id"`
		COCode           string  `json:"co_code"`
		COName           string  `json:"co_name"`
		COSource         string  `json:"co_type"`
		ScopeChange      string  `json:"scope_change"`
		CostChange       float64 `json:"cost_change"`
		ScheduleChange   int     `json:"schedule_change_days"`
		ContractorName   string  `json:"contractor_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO change_orders (id, project_id, co_number, co_code, co_name, co_type, scope_change, cost_change, schedule_change_days, contractor_name, status, created_at, updated_at) VALUES ($1,$2,(SELECT COALESCE(MAX(co_number),0)+1 FROM change_orders WHERE project_id=$2),$3,$4,$5,$6,$7,$8,$9,'draft',$10,$11)`,
		id, input.ProjectID, input.COCode, input.COName, input.COSource, input.ScopeChange, input.CostChange, input.ScheduleChange, input.ContractorName, now, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *ChangeHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var code, name string
	err := h.db.QueryRow(`SELECT co_code, co_name FROM change_orders WHERE id = $1`, id).Scan(&code, &name)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "change order not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "co_code": code, "co_name": name})
}

func (h *ChangeHandler) UpdateOrder(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Status *string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	_, err := h.db.Exec(`UPDATE change_orders SET status=COALESCE($1,status), updated_at=$2 WHERE id=$3`, input.Status, time.Now(), id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *ChangeHandler) DeleteOrder(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM change_orders WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// Impact Analysis
// =============================================================================
func (h *ChangeHandler) ListImpactAnalysis(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT id, project_id, impact_type, description, impact_level, cost_impact, schedule_impact_days, analyzed_by, analysis_date, status FROM change_impact_analysis WHERE 1=1`
	args := []interface{}{}
	argIdx := 1
	if projectID != "" { query += fmt.Sprintf(" AND project_id = $%d", argIdx); args = append(args, projectID) }
	query += " ORDER BY analysis_date DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, pid, itype, desc, ilevel, analyzedBy, st string
		var cost float64
		var schedDays int
		var analysisDate sql.NullString
		if err := rows.Scan(&id, &pid, &itype, &desc, &ilevel, &cost, &schedDays, &analyzedBy, &analysisDate, &st); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		item := map[string]interface{}{
			"id": id, "project_id": pid, "impact_type": itype, "description": desc,
			"impact_level": ilevel, "cost_impact": cost, "schedule_impact_days": schedDays,
			"analyzed_by": analyzedBy, "status": st,
		}
		if analysisDate.Valid { item["analysis_date"] = analysisDate.String }
		items = append(items, item)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *ChangeHandler) CreateImpactAnalysis(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID    string  `json:"project_id"`
		ImpactType   string  `json:"impact_type"`
		Description  string  `json:"description"`
		ImpactLevel  string  `json:"impact_level"`
		CostImpact   float64 `json:"cost_impact"`
		ScheduleDays int     `json:"schedule_impact_days"`
		AnalyzedBy   string  `json:"analyzed_by"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO change_impact_analysis (id, project_id, impact_type, description, impact_level, cost_impact, schedule_impact_days, analyzed_by, analysis_date, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		id, input.ProjectID, input.ImpactType, input.Description, input.ImpactLevel, input.CostImpact, input.ScheduleDays, input.AnalyzedBy, now, now, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *ChangeHandler) GetImpactAnalysis(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var itype, desc string
	err := h.db.QueryRow(`SELECT impact_type, description FROM change_impact_analysis WHERE id = $1`, id).Scan(&itype, &desc)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "impact analysis not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "impact_type": itype, "description": desc})
}

func (h *ChangeHandler) UpdateImpactAnalysis(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Status *string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	_, err := h.db.Exec(`UPDATE change_impact_analysis SET status=COALESCE($1,status), updated_at=$2 WHERE id=$3`, input.Status, time.Now(), id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *ChangeHandler) DeleteImpactAnalysis(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM change_impact_analysis WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// Approval Workflow
// =============================================================================
func (h *ChangeHandler) ListApprovalWorkflow(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT id, project_id, step_order, step_name, approver_role, approver_name, status, decision, comments, decided_at FROM change_approval_workflow WHERE 1=1`
	args := []interface{}{}
	argIdx := 1
	if projectID != "" { query += fmt.Sprintf(" AND project_id = $%d", argIdx); args = append(args, projectID) }
	query += " ORDER BY step_order"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, pid, stepName, appRole, appName, st, decision, comments string
		var stepOrder int
		var decidedAt sql.NullString
		if err := rows.Scan(&id, &pid, &stepOrder, &stepName, &appRole, &appName, &st, &decision, &comments, &decidedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		item := map[string]interface{}{
			"id": id, "project_id": pid, "step_order": stepOrder, "step_name": stepName,
			"approver_role": appRole, "approver_name": appName,
			"status": st, "decision": decision, "comments": comments,
		}
		if decidedAt.Valid { item["decided_at"] = decidedAt.String }
		items = append(items, item)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *ChangeHandler) CreateApprovalWorkflow(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID    string `json:"project_id"`
		StepOrder    int    `json:"step_order"`
		StepName     string `json:"step_name"`
		ApproverRole string `json:"approver_role"`
		ApproverName string `json:"approver_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO change_approval_workflow (id, project_id, step_order, step_name, approver_role, approver_name, status, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,'pending',$7,$8)`,
		id, input.ProjectID, input.StepOrder, input.StepName, input.ApproverRole, input.ApproverName, now, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *ChangeHandler) GetApprovalWorkflow(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var stepName, status string
	err := h.db.QueryRow(`SELECT step_name, status FROM change_approval_workflow WHERE id = $1`, id).Scan(&stepName, &status)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "approval step not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "step_name": stepName, "status": status})
}

func (h *ChangeHandler) UpdateApprovalWorkflow(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Status   *string `json:"status"`
		Decision *string `json:"decision"`
		Comments *string `json:"comments"`
		DecidedBy *string `json:"decided_by"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	now := time.Now()
	_, err := h.db.Exec(`UPDATE change_approval_workflow SET status=COALESCE($1,status), decision=COALESCE($2,decision), comments=COALESCE($3,comments), decided_by=COALESCE($4,decided_by), decided_at=CASE WHEN $1 IS NOT NULL AND $1 IN ('approved','rejected') THEN $5 ELSE decided_at END, updated_at=$5 WHERE id=$6`,
		input.Status, input.Decision, input.Comments, input.DecidedBy, now, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *ChangeHandler) DeleteApprovalWorkflow(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM change_approval_workflow WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// Change Log
// =============================================================================
func (h *ChangeHandler) ListChangeLog(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT id, project_id, log_type, description, changed_by, changed_at FROM change_log WHERE 1=1`
	args := []interface{}{}
	argIdx := 1
	if projectID != "" { query += fmt.Sprintf(" AND project_id = $%d", argIdx); args = append(args, projectID) }
	query += " ORDER BY changed_at DESC LIMIT 200"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, pid, ltype, desc, changedBy string
		var changedAt time.Time
		if err := rows.Scan(&id, &pid, &ltype, &desc, &changedBy, &changedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		items = append(items, map[string]interface{}{
			"id": id, "project_id": pid, "log_type": ltype, "description": desc,
			"changed_by": changedBy, "changed_at": changedAt,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *ChangeHandler) CreateChangeLog(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID      string `json:"project_id"`
		LogType        string `json:"log_type"`
		Description    string `json:"description"`
		ChangedBy      string `json:"changed_by"`
		PreviousStatus string `json:"previous_status"`
		NewStatus      string `json:"new_status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO change_log (id, project_id, log_type, description, changed_by, changed_at, previous_status, new_status) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		id, input.ProjectID, input.LogType, input.Description, input.ChangedBy, now, input.PreviousStatus, input.NewStatus)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *ChangeHandler) GetChangeLog(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var ltype, desc string
	err := h.db.QueryRow(`SELECT log_type, description FROM change_log WHERE id = $1`, id).Scan(&ltype, &desc)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "change log entry not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "log_type": ltype, "description": desc})
}

func (h *ChangeHandler) DeleteChangeLog(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM change_log WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// Summary
// =============================================================================
func (h *ChangeHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT project_id, open_crs, total_crs, approved_crs, implemented_crs, rejected_crs, high_priority_open, total_cost_change, total_schedule_impact, pending_approvals, high_impact_analyses FROM change_summary`
	args := []interface{}{}
	argIdx := 1
	if projectID != "" { query += fmt.Sprintf(" WHERE project_id = $%d", argIdx); args = append(args, projectID) }

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var pid string
		var open, total, approved, implemented, rejected, highPrio, pendApprovals, highImpact int
		var costChange, schedImpact float64
		if err := rows.Scan(&pid, &open, &total, &approved, &implemented, &rejected, &highPrio, &costChange, &schedImpact, &pendApprovals, &highImpact); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		items = append(items, map[string]interface{}{
			"project_id": pid, "open_crs": open, "total_crs": total,
			"approved_crs": approved, "implemented_crs": implemented, "rejected_crs": rejected,
			"high_priority_open": highPrio, "total_cost_change": costChange,
			"total_schedule_impact": schedImpact, "pending_approvals": pendApprovals,
			"high_impact_analyses": highImpact,
		})
	}
	respondJSON(w, http.StatusOK, items)
}