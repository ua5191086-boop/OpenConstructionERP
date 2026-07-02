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
	"github.com/openconstructionerp/oce/services/core/internal/models"
)

// ScheduleHandler handles Schedule Management module endpoints
type ScheduleHandler struct {
	db *sql.DB
}

func NewScheduleHandler(db *sql.DB) *ScheduleHandler {
	return &ScheduleHandler{db: db}
}

func (h *ScheduleHandler) RegisterRoutes(r chi.Router) {
	r.Route("/schedule", func(r chi.Router) {
		// Schedules
		r.Get("/schedules", h.ListSchedules)
		r.Post("/schedules", h.CreateSchedule)
		r.Get("/schedules/{id}", h.GetSchedule)
		r.Put("/schedules/{id}", h.UpdateSchedule)
		r.Delete("/schedules/{id}", h.DeleteSchedule)

		// Activities
		r.Get("/activities", h.ListActivities)
		r.Post("/activities", h.CreateActivity)
		r.Get("/activities/{id}", h.GetActivity)
		r.Put("/activities/{id}", h.UpdateActivity)
		r.Delete("/activities/{id}", h.DeleteActivity)

		// Relationships
		r.Get("/relationships", h.ListRelationships)
		r.Post("/relationships", h.CreateRelationship)
		r.Get("/relationships/{id}", h.GetRelationship)
		r.Delete("/relationships/{id}", h.DeleteRelationship)

		// Resources
		r.Get("/resources", h.ListResources)
		r.Post("/resources", h.CreateResource)
		r.Get("/resources/{id}", h.GetResource)
		r.Put("/resources/{id}", h.UpdateResource)
		r.Delete("/resources/{id}", h.DeleteResource)

		// Baselines
		r.Get("/baselines", h.ListBaselines)
		r.Post("/baselines", h.CreateBaseline)
		r.Get("/baselines/{id}", h.GetBaseline)
		r.Delete("/baselines/{id}", h.DeleteBaseline)

		// Changes
		r.Get("/changes", h.ListChanges)
		r.Post("/changes", h.CreateChange)
		r.Get("/changes/{id}", h.GetChange)
		r.Put("/changes/{id}", h.UpdateChange)
		r.Delete("/changes/{id}", h.DeleteChange)

		// Critical Path Log
		r.Get("/critical-path-logs", h.ListCriticalPathLogs)
		r.Post("/critical-path-logs", h.CreateCriticalPathLog)
		r.Get("/critical-path-logs/{id}", h.GetCriticalPathLog)

		// CPM Calculation Engine
		r.Post("/cpm-calculate/{scheduleId}", h.CPMCalculate)

		// Gantt Chart Data
		r.Get("/gantt/{scheduleId}", h.GanttData)

		// Baseline Comparison
		r.Get("/baseline-compare/{scheduleId}", h.BaselineCompare)

		// Summary
		r.Get("/summary", h.GetSummary)
	})
}

// =============================================================================
// Schedules
// =============================================================================

func (h *ScheduleHandler) ListSchedules(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	status := r.URL.Query().Get("status")

	query := `SELECT id, project_id, schedule_code, schedule_name, schedule_type, calendar, data_date, status, total_float_pct, created_by, created_at, updated_at FROM schedules WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if projectID != "" { query += fmt.Sprintf(" AND project_id = $%d", argIdx); argIdx++; args = append(args, projectID) }
	if status != "" { query += fmt.Sprintf(" AND status = $%d", argIdx); argIdx++; args = append(args, status) }
	query += " ORDER BY created_at DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, pid, code, name, stype, cal, status, createdBy string
		var dataDate, createdAt, updatedAt time.Time
		var tfpct sql.NullFloat64
		err := rows.Scan(&id, &pid, &code, &name, &stype, &cal, &dataDate, &status, &tfpct, &createdBy, &createdAt, &updatedAt)
		if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
		item := map[string]interface{}{
			"id": id, "project_id": pid, "schedule_code": code, "schedule_name": name,
			"schedule_type": stype, "calendar": cal, "data_date": dataDate, "status": status,
			"created_by": createdBy, "created_at": createdAt, "updated_at": updatedAt,
		}
		if tfpct.Valid { item["total_float_pct"] = tfpct.Float64 }
		items = append(items, item)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *ScheduleHandler) CreateSchedule(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID    string  `json:"project_id"`
		ScheduleCode string  `json:"schedule_code"`
		ScheduleName string  `json:"schedule_name"`
		ScheduleType *string `json:"schedule_type"`
		Calendar     *string `json:"calendar"`
		Status       *string `json:"status"`
		CreatedBy    *string `json:"created_by"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO schedules (id, project_id, schedule_code, schedule_name, schedule_type, calendar, status, created_by, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		id, input.ProjectID, input.ScheduleCode, input.ScheduleName, input.ScheduleType, input.Calendar, input.Status, input.CreatedBy, now, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *ScheduleHandler) GetSchedule(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var pid, code, name, stype, cal, status, createdBy string
	var dataDate, createdAt, updatedAt time.Time
	var tfpct sql.NullFloat64
	err := h.db.QueryRow(`SELECT id, project_id, schedule_code, schedule_name, schedule_type, calendar, data_date, status, total_float_pct, created_by, created_at, updated_at FROM schedules WHERE id = $1`, id).
		Scan(&id, &pid, &code, &name, &stype, &cal, &dataDate, &status, &tfpct, &createdBy, &createdAt, &updatedAt)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "schedule not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	resp := map[string]interface{}{
		"id": id, "project_id": pid, "schedule_code": code, "schedule_name": name,
		"schedule_type": stype, "calendar": cal, "data_date": dataDate, "status": status,
		"created_by": createdBy, "created_at": createdAt, "updated_at": updatedAt,
	}
	if tfpct.Valid { resp["total_float_pct"] = tfpct.Float64 }
	respondJSON(w, http.StatusOK, resp)
}

func (h *ScheduleHandler) UpdateSchedule(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Status       *string `json:"status"`
		ScheduleName *string `json:"schedule_name"`
		Calendar     *string `json:"calendar"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	now := time.Now()
	_, err := h.db.Exec(`UPDATE schedules SET status=COALESCE($1,status), schedule_name=COALESCE($2,schedule_name), calendar=COALESCE($3,calendar), updated_at=$4 WHERE id=$5`,
		input.Status, input.ScheduleName, input.Calendar, now, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *ScheduleHandler) DeleteSchedule(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM schedules WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// Activities
// =============================================================================

func (h *ScheduleHandler) ListActivities(w http.ResponseWriter, r *http.Request) {
	scheduleID := r.URL.Query().Get("schedule_id")
	status := r.URL.Query().Get("status")
	critical := r.URL.Query().Get("critical")

	query := `SELECT id, schedule_id, activity_id, wbs_code, activity_name, activity_type, status, original_duration, remaining_duration, actual_duration, percent_complete, early_start, early_finish, late_start, late_finish, actual_start, actual_finish, start_date, finish_date, float_free, float_total, is_critical, is_driving, constraint_type, constraint_date, notes, created_at, updated_at FROM schedule_activities WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if scheduleID != "" { query += fmt.Sprintf(" AND schedule_id = $%d", argIdx); argIdx++; args = append(args, scheduleID) }
	if status != "" { query += fmt.Sprintf(" AND status = $%d", argIdx); argIdx++; args = append(args, status) }
	if critical == "true" { query += " AND is_critical = TRUE" }
	query += " ORDER BY activity_id"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, sid, actID, wbs, name, atype, status, constraintType string
		var origDur, remDur, actDur, floatFree, floatTotal int
		var pct sql.NullFloat64
		var es, ef, ls, lf, ast, af, sd, fd, constraintDate, notes sql.NullString
		var isCritical, isDriving bool
		var createdAt, updatedAt time.Time

		err := rows.Scan(&id, &sid, &actID, &wbs, &name, &atype, &status, &origDur, &remDur, &actDur, &pct, &es, &ef, &ls, &lf, &ast, &af, &sd, &fd, &floatFree, &floatTotal, &isCritical, &isDriving, &constraintType, &constraintDate, &notes, &createdAt, &updatedAt)
		if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
		item := map[string]interface{}{
			"id": id, "schedule_id": sid, "activity_id": actID, "wbs_code": wbs,
			"activity_name": name, "activity_type": atype, "status": status,
			"original_duration": origDur, "remaining_duration": remDur, "actual_duration": actDur,
			"float_free": floatFree, "float_total": floatTotal,
			"is_critical": isCritical, "is_driving": isDriving, "constraint_type": constraintType,
			"created_at": createdAt, "updated_at": updatedAt,
		}
		if pct.Valid { item["percent_complete"] = pct.Float64 }
		if es.Valid { item["early_start"] = es.String }
		if ef.Valid { item["early_finish"] = ef.String }
		if ls.Valid { item["late_start"] = ls.String }
		if lf.Valid { item["late_finish"] = lf.String }
		if ast.Valid { item["actual_start"] = ast.String }
		if af.Valid { item["actual_finish"] = af.String }
		if sd.Valid { item["start_date"] = sd.String }
		if fd.Valid { item["finish_date"] = fd.String }
		if constraintDate.Valid { item["constraint_date"] = constraintDate.String }
		if notes.Valid { item["notes"] = notes.String }
		items = append(items, item)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *ScheduleHandler) CreateActivity(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ScheduleID       string  `json:"schedule_id"`
		ActivityID       string  `json:"activity_id"`
		ActivityName     string  `json:"activity_name"`
		ActivityType     *string `json:"activity_type"`
		OriginalDuration *int    `json:"original_duration"`
		StartDate        *string `json:"start_date"`
		FinishDate       *string `json:"finish_date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO schedule_activities (id, schedule_id, activity_id, activity_name, activity_type, original_duration, start_date, finish_date, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		id, input.ScheduleID, input.ActivityID, input.ActivityName, input.ActivityType, input.OriginalDuration, input.StartDate, input.FinishDate, now, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *ScheduleHandler) GetActivity(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var actID, name string
	var origDur int
	err := h.db.QueryRow(`SELECT activity_id, activity_name, original_duration FROM schedule_activities WHERE id = $1`, id).Scan(&actID, &name, &origDur)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "activity not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "activity_id": actID, "activity_name": name, "original_duration": origDur})
}

func (h *ScheduleHandler) UpdateActivity(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Status            *string  `json:"status"`
		PercentComplete   *float64 `json:"percent_complete"`
		RemainingDuration *int     `json:"remaining_duration"`
		ActualStart       *string  `json:"actual_start"`
		ActualFinish      *string  `json:"actual_finish"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	now := time.Now()
	_, err := h.db.Exec(`UPDATE schedule_activities SET status=COALESCE($1,status), percent_complete=COALESCE($2,percent_complete), remaining_duration=COALESCE($3,remaining_duration), actual_start=COALESCE($4,actual_start), actual_finish=COALESCE($5,actual_finish), updated_at=$6 WHERE id=$7`,
		input.Status, input.PercentComplete, input.RemainingDuration, input.ActualStart, input.ActualFinish, now, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *ScheduleHandler) DeleteActivity(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM schedule_activities WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// Relationships
// =============================================================================

func (h *ScheduleHandler) ListRelationships(w http.ResponseWriter, r *http.Request) {
	scheduleID := r.URL.Query().Get("schedule_id")

	query := `SELECT id, schedule_id, predecessor_id, successor_id, relation_type, lag_days, created_at FROM schedule_relationships WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if scheduleID != "" { query += fmt.Sprintf(" AND schedule_id = $%d", argIdx); argIdx++; args = append(args, scheduleID) }
	query += " ORDER BY predecessor_id"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, sid, predID, succID, relType string
		var lagDays int
		var createdAt time.Time
		err := rows.Scan(&id, &sid, &predID, &succID, &relType, &lagDays, &createdAt)
		if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
		items = append(items, map[string]interface{}{
			"id": id, "schedule_id": sid, "predecessor_id": predID, "successor_id": succID,
			"relation_type": relType, "lag_days": lagDays, "created_at": createdAt,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *ScheduleHandler) CreateRelationship(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ScheduleID    string `json:"schedule_id"`
		PredecessorID string `json:"predecessor_id"`
		SuccessorID   string `json:"successor_id"`
		RelationType  string `json:"relation_type"`
		LagDays       *int   `json:"lag_days"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	_, err := h.db.Exec(`INSERT INTO schedule_relationships (id, schedule_id, predecessor_id, successor_id, relation_type, lag_days) VALUES ($1,$2,$3,$4,$5,$6)`,
		id, input.ScheduleID, input.PredecessorID, input.SuccessorID, input.RelationType, input.LagDays)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *ScheduleHandler) GetRelationship(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var relType string
	err := h.db.QueryRow(`SELECT relation_type FROM schedule_relationships WHERE id = $1`, id).Scan(&relType)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "relationship not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "relation_type": relType})
}

func (h *ScheduleHandler) DeleteRelationship(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM schedule_relationships WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// Resources
// =============================================================================

func (h *ScheduleHandler) ListResources(w http.ResponseWriter, r *http.Request) {
	scheduleID := r.URL.Query().Get("schedule_id")
	resourceType := r.URL.Query().Get("resource_type")

	query := `SELECT id, schedule_id, activity_id, resource_type, resource_code, resource_name, units_per_day, total_units, unit_cost, total_cost, bid_price, actual_units, actual_cost, created_at, updated_at FROM schedule_resources WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if scheduleID != "" { query += fmt.Sprintf(" AND schedule_id = $%d", argIdx); argIdx++; args = append(args, scheduleID) }
	if resourceType != "" { query += fmt.Sprintf(" AND resource_type = $%d", argIdx); argIdx++; args = append(args, resourceType) }
	query += " ORDER BY resource_code"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, sid, actID, rtype, rcode, rname string
		var unitsPerDay, totalUnits, unitCost, totalCost, actUnits, actCost float64
		var bidPrice sql.NullFloat64
		var createdAt, updatedAt time.Time
		err := rows.Scan(&id, &sid, &actID, &rtype, &rcode, &rname, &unitsPerDay, &totalUnits, &unitCost, &totalCost, &bidPrice, &actUnits, &actCost, &createdAt, &updatedAt)
		if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
		item := map[string]interface{}{
			"id": id, "schedule_id": sid, "activity_id": actID, "resource_type": rtype,
			"resource_code": rcode, "resource_name": rname, "units_per_day": unitsPerDay,
			"total_units": totalUnits, "unit_cost": unitCost, "total_cost": totalCost,
			"actual_units": actUnits, "actual_cost": actCost,
			"created_at": createdAt, "updated_at": updatedAt,
		}
		if bidPrice.Valid { item["bid_price"] = bidPrice.Float64 }
		items = append(items, item)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *ScheduleHandler) CreateResource(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ScheduleID   string  `json:"schedule_id"`
		ActivityID   string  `json:"activity_id"`
		ResourceType string  `json:"resource_type"`
		ResourceCode string  `json:"resource_code"`
		ResourceName string  `json:"resource_name"`
		TotalUnits   *float64 `json:"total_units"`
		UnitCost     *float64 `json:"unit_cost"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO schedule_resources (id, schedule_id, activity_id, resource_type, resource_code, resource_name, total_units, unit_cost, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		id, input.ScheduleID, input.ActivityID, input.ResourceType, input.ResourceCode, input.ResourceName, input.TotalUnits, input.UnitCost, now, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *ScheduleHandler) GetResource(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var rcode, rname string
	err := h.db.QueryRow(`SELECT resource_code, resource_name FROM schedule_resources WHERE id = $1`, id).Scan(&rcode, &rname)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "resource not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "resource_code": rcode, "resource_name": rname})
}

func (h *ScheduleHandler) UpdateResource(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		ActualUnits *float64 `json:"actual_units"`
		ActualCost  *float64 `json:"actual_cost"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	now := time.Now()
	_, err := h.db.Exec(`UPDATE schedule_resources SET actual_units=COALESCE($1,actual_units), actual_cost=COALESCE($2,actual_cost), updated_at=$3 WHERE id=$4`,
		input.ActualUnits, input.ActualCost, now, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *ScheduleHandler) DeleteResource(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM schedule_resources WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// Baselines
// =============================================================================

func (h *ScheduleHandler) ListBaselines(w http.ResponseWriter, r *http.Request) {
	scheduleID := r.URL.Query().Get("schedule_id")

	query := `SELECT id, schedule_id, baseline_number, baseline_name, baseline_date, is_current, total_float_pct, created_by, created_at FROM schedule_baselines WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if scheduleID != "" { query += fmt.Sprintf(" AND schedule_id = $%d", argIdx); argIdx++; args = append(args, scheduleID) }
	query += " ORDER BY baseline_number"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, sid, blName, createdBy string
		var blNum int
		var blDate time.Time
		var isCurrent bool
		var tfpct sql.NullFloat64
		var createdAt time.Time
		err := rows.Scan(&id, &sid, &blNum, &blName, &blDate, &isCurrent, &tfpct, &createdBy, &createdAt)
		if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
		item := map[string]interface{}{
			"id": id, "schedule_id": sid, "baseline_number": blNum, "baseline_name": blName,
			"baseline_date": blDate, "is_current": isCurrent, "created_by": createdBy, "created_at": createdAt,
		}
		if tfpct.Valid { item["total_float_pct"] = tfpct.Float64 }
		items = append(items, item)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *ScheduleHandler) CreateBaseline(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ScheduleID    string `json:"schedule_id"`
		BaselineName  string `json:"baseline_name"`
		BaselineDate  string `json:"baseline_date"`
		IsCurrent     *bool  `json:"is_current"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	_, err := h.db.Exec(`INSERT INTO schedule_baselines (id, schedule_id, baseline_number, baseline_name, baseline_date, is_current) VALUES ($1, $2, (SELECT COALESCE(MAX(baseline_number),0)+1 FROM schedule_baselines WHERE schedule_id=$2), $3, $4, $5)`,
		id, input.ScheduleID, input.BaselineName, input.BaselineDate, input.IsCurrent)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *ScheduleHandler) GetBaseline(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var blName string
	err := h.db.QueryRow(`SELECT baseline_name FROM schedule_baselines WHERE id = $1`, id).Scan(&blName)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "baseline not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "baseline_name": blName})
}

func (h *ScheduleHandler) DeleteBaseline(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM schedule_baselines WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// Changes
// =============================================================================

func (h *ScheduleHandler) ListChanges(w http.ResponseWriter, r *http.Request) {
	scheduleID := r.URL.Query().Get("schedule_id")
	status := r.URL.Query().Get("status")

	query := `SELECT id, schedule_id, change_number, change_code, change_type, description, reason, impact_days, impact_cost, activity_id, baseline_id, approved_by, status, proposed_at, approved_at, created_at, updated_at FROM schedule_changes WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if scheduleID != "" { query += fmt.Sprintf(" AND schedule_id = $%d", argIdx); argIdx++; args = append(args, scheduleID) }
	if status != "" { query += fmt.Sprintf(" AND status = $%d", argIdx); argIdx++; args = append(args, status) }
	query += " ORDER BY change_number"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, sid, chCode, chType, desc, reason, apBy, status string
		var chNum int
		var impactDays, impactCost float64
		var actID, blID, proposedAt, approvedAt, createdAt, updatedAt sql.NullString
		err := rows.Scan(&id, &sid, &chNum, &chCode, &chType, &desc, &reason, &impactDays, &impactCost, &actID, &blID, &apBy, &status, &proposedAt, &approvedAt, &createdAt, &updatedAt)
		if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
		item := map[string]interface{}{
			"id": id, "schedule_id": sid, "change_number": chNum, "change_code": chCode,
			"change_type": chType, "description": desc, "reason": reason,
			"impact_days": impactDays, "impact_cost": impactCost,
			"approved_by": apBy, "status": status,
		}
		if actID.Valid { item["activity_id"] = actID.String }
		if blID.Valid { item["baseline_id"] = blID.String }
		if proposedAt.Valid { item["proposed_at"] = proposedAt.String }
		if approvedAt.Valid { item["approved_at"] = approvedAt.String }
		items = append(items, item)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *ScheduleHandler) CreateChange(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ScheduleID  string  `json:"schedule_id"`
		ChangeCode  string  `json:"change_code"`
		ChangeType  string  `json:"change_type"`
		Description string  `json:"description"`
		Reason      *string `json:"reason"`
		ImpactDays  *int    `json:"impact_days"`
		ImpactCost  *float64 `json:"impact_cost"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO schedule_changes (id, schedule_id, change_number, change_code, change_type, description, reason, impact_days, impact_cost, proposed_at, created_at, updated_at) VALUES ($1, $2, (SELECT COALESCE(MAX(change_number),0)+1 FROM schedule_changes WHERE schedule_id=$2), $3, $4, $5, $6, $7, $8, $9, $9, $9)`,
		id, input.ScheduleID, input.ChangeCode, input.ChangeType, input.Description, input.Reason, input.ImpactDays, input.ImpactCost, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *ScheduleHandler) GetChange(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var chCode, chType, desc string
	err := h.db.QueryRow(`SELECT change_code, change_type, description FROM schedule_changes WHERE id = $1`, id).Scan(&chCode, &chType, &desc)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "change not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "change_code": chCode, "change_type": chType, "description": desc})
}

func (h *ScheduleHandler) UpdateChange(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Status      *string `json:"status"`
		ApprovedBy  *string `json:"approved_by"`
		ImpactDays  *int    `json:"impact_days"`
		ImpactCost  *float64 `json:"impact_cost"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	now := time.Now()
	_, err := h.db.Exec(`UPDATE schedule_changes SET status=COALESCE($1,status), approved_by=COALESCE($2,approved_by), impact_days=COALESCE($3,impact_days), impact_cost=COALESCE($4,impact_cost), updated_at=$5 WHERE id=$6`,
		input.Status, input.ApprovedBy, input.ImpactDays, input.ImpactCost, now, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *ScheduleHandler) DeleteChange(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM schedule_changes WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// Critical Path Log
// =============================================================================

func (h *ScheduleHandler) ListCriticalPathLogs(w http.ResponseWriter, r *http.Request) {
	scheduleID := r.URL.Query().Get("schedule_id")

	query := `SELECT id, schedule_id, run_number, run_at, total_activities, critical_count, longest_path, total_float_min, total_float_max, total_float_avg, critical_path, duration, status, error_message, created_at FROM critical_path_log WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if scheduleID != "" { query += fmt.Sprintf(" AND schedule_id = $%d", argIdx); argIdx++; args = append(args, scheduleID) }
	query += " ORDER BY run_number DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, sid, cp, status, errMsg string
		var runNum, totalAct, critCount, longestPath, floatMin, floatMax, dur int
		var floatAvg float64
		var runAt, createdAt time.Time
		err := rows.Scan(&id, &sid, &runNum, &runAt, &totalAct, &critCount, &longestPath, &floatMin, &floatMax, &floatAvg, &cp, &dur, &status, &errMsg, &createdAt)
		if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
		items = append(items, map[string]interface{}{
			"id": id, "schedule_id": sid, "run_number": runNum, "run_at": runAt,
			"total_activities": totalAct, "critical_count": critCount, "longest_path": longestPath,
			"total_float_min": floatMin, "total_float_max": floatMax, "total_float_avg": floatAvg,
			"critical_path": cp, "duration": dur, "status": status, "error_message": errMsg,
			"created_at": createdAt,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *ScheduleHandler) CreateCriticalPathLog(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ScheduleID   string  `json:"schedule_id"`
		TotalActivities *int `json:"total_activities"`
		CriticalCount  *int  `json:"critical_count"`
		LongestPath    *int  `json:"longest_path"`
		TotalFloatAvg  *float64 `json:"total_float_avg"`
		CriticalPath   *string `json:"critical_path"`
		Duration       *int  `json:"duration"`
		Status         *string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO critical_path_log (id, schedule_id, run_number, run_at, total_activities, critical_count, longest_path, total_float_avg, critical_path, duration, status, created_at) VALUES ($1,$2,(SELECT COALESCE(MAX(run_number),0)+1 FROM critical_path_log WHERE schedule_id=$2),$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		id, input.ScheduleID, now, input.TotalActivities, input.CriticalCount, input.LongestPath, input.TotalFloatAvg, input.CriticalPath, input.Duration, input.Status, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *ScheduleHandler) GetCriticalPathLog(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var runNum int
	err := h.db.QueryRow(`SELECT run_number FROM critical_path_log WHERE id = $1`, id).Scan(&runNum)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "CPM log not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "run_number": runNum})
}

// =============================================================================
// Summary
// =============================================================================

func (h *ScheduleHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")

	query := `SELECT project_id, total_schedules, active_schedules, total_activities, not_started, in_progress, completed, critical_activities, total_relationships, total_resources, total_baselines, pending_changes, cpm_runs FROM schedule_summary`
	args := []interface{}{}
	argIdx := 1
	if projectID != "" { query += fmt.Sprintf(" WHERE project_id = $%d", argIdx); args = append(args, projectID) }

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var pid string
		var totalSched, activeSched, totalAct, notStarted, inProgress, completed, critical, totalRel, totalRes, totalBl, pendingCh, cpmRuns int
		err := rows.Scan(&pid, &totalSched, &activeSched, &totalAct, &notStarted, &inProgress, &completed, &critical, &totalRel, &totalRes, &totalBl, &pendingCh, &cpmRuns)
		if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
		items = append(items, map[string]interface{}{
			"project_id": pid, "total_schedules": totalSched, "active_schedules": activeSched,
			"total_activities": totalAct, "not_started": notStarted, "in_progress": inProgress,
			"completed": completed, "critical_activities": critical,
			"total_relationships": totalRel, "total_resources": totalRes,
			"total_baselines": totalBl, "pending_changes": pendingCh, "cpm_runs": cpmRuns,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func init() {
	log.Println("Schedule Management handler initialized")
}

// =============================================================================
// CPM Calculation Engine (Forward/Backward Pass)
// =============================================================================

func (h *ScheduleHandler) CPMCalculate(w http.ResponseWriter, r *http.Request) {
	scheduleID := chi.URLParam(r, "scheduleId")

	// 1. Get all activities
	actRows, err := h.db.Query(`SELECT id, activity_id, original_duration, start_date, finish_date FROM schedule_activities WHERE schedule_id = $1 ORDER BY activity_id`, scheduleID)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer actRows.Close()

	type activity struct {
		ID         string
		ActID      string
		Duration   int
		Predecessors []string
		Successors   []string
		ES, EF, LS, LF int // days from project start
		FloatTotal int
	}
	activities := make(map[string]*activity)
	actOrder := []string{}

	for actRows.Next() {
		var id, actID string
		var dur int
		var sd, fd sql.NullString
		if err := actRows.Scan(&id, &actID, &dur, &sd, &fd); err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
		activities[actID] = &activity{
			ID:       id,
			ActID:    actID,
			Duration: dur,
			Predecessors: []string{},
			Successors:   []string{},
		}
		actOrder = append(actOrder, actID)
	}

	// 2. Get relationships
	relRows, err := h.db.Query(`SELECT predecessor_id, successor_id, relation_type, lag_days FROM schedule_relationships WHERE schedule_id = $1`, scheduleID)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer relRows.Close()

	for relRows.Next() {
		var predID, succID, relType string
		var lag int
		if err := relRows.Scan(&predID, &succID, &relType, &lag); err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
		if a, ok := activities[succID]; ok {
			a.Predecessors = append(a.Predecessors, predID)
		}
		if a, ok := activities[predID]; ok {
			a.Successors = append(a.Successors, succID)
		}
	}

	// 3. Topological sort (Kahn's algorithm)
	inDegree := make(map[string]int)
	for _, id := range actOrder {
		inDegree[id] = len(activities[id].Predecessors)
	}
	queue := []string{}
	for _, id := range actOrder {
		if inDegree[id] == 0 {
			queue = append(queue, id)
		}
	}
	sorted := []string{}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		sorted = append(sorted, cur)
		for _, s := range activities[cur].Successors {
			inDegree[s]--
			if inDegree[s] == 0 {
				queue = append(queue, s)
			}
		}
	}

	// 4. Forward Pass (ES, EF)
	for _, id := range sorted {
		a := activities[id]
		a.ES = 0
		for _, p := range a.Predecessors {
			if pred, ok := activities[p]; ok && pred.EF > a.ES {
				a.ES = pred.EF
			}
		}
		a.EF = a.ES + a.Duration
	}

	// 5. Backward Pass (LS, LF) — project duration = max EF
	projectDuration := 0
	for _, a := range activities {
		if a.EF > projectDuration {
			projectDuration = a.EF
		}
	}
	for i := len(sorted) - 1; i >= 0; i-- {
		id := sorted[i]
		a := activities[id]
		a.LF = projectDuration
		for _, s := range a.Successors {
			if succ, ok := activities[s]; ok && succ.LS < a.LF {
				a.LF = succ.LS
			}
		}
		a.LS = a.LF - a.Duration
		a.FloatTotal = a.LS - a.ES
	}

	// 6. Update DB with CPM results & mark critical
	criticalCount := 0
	criticalPath := []string{}
	for _, a := range activities {
		isCritical := a.FloatTotal == 0
		if isCritical {
			criticalCount++
			criticalPath = append(criticalPath, a.ActID)
		}
		h.db.Exec(`UPDATE schedule_activities SET early_start=$1, early_finish=$2, late_start=$3, late_finish=$4, float_total=$5, is_critical=$6, updated_at=NOW() WHERE id=$7`,
			fmt.Sprintf("DAY_%d", a.ES), fmt.Sprintf("DAY_%d", a.EF),
			fmt.Sprintf("DAY_%d", a.LS), fmt.Sprintf("DAY_%d", a.LF),
			a.FloatTotal, isCritical, a.ID)
	}

	// 7. Log CPM run
	now := time.Now()
	cpJSON, _ := json.Marshal(criticalPath)
	h.db.Exec(`INSERT INTO critical_path_log (id, schedule_id, run_number, run_at, total_activities, critical_count, longest_path, total_float_min, total_float_max, critical_path, duration, status, created_at) VALUES ($1,$2,(SELECT COALESCE(MAX(run_number),0)+1 FROM critical_path_log WHERE schedule_id=$2),$3,$4,$5,$6,$7,$8,$9,0,'completed',$10)`,
		uuid.New().String(), scheduleID, now, len(activities), criticalCount, projectDuration, 0, projectDuration, string(cpJSON), now)

	// 8. Return result
	actSummary := make([]map[string]interface{}, 0)
	for _, a := range activities {
		actSummary = append(actSummary, map[string]interface{}{
			"activity_id": a.ActID, "duration": a.Duration,
			"es": a.ES, "ef": a.EF, "ls": a.LS, "lf": a.LF,
			"float_total": a.FloatTotal, "is_critical": a.FloatTotal == 0,
		})
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"schedule_id":      scheduleID,
		"total_activities": len(activities),
		"project_duration": projectDuration,
		"critical_count":   criticalCount,
		"critical_path":    criticalPath,
		"activities":       actSummary,
	})
}

// =============================================================================
// Gantt Chart Data
// =============================================================================

func (h *ScheduleHandler) GanttData(w http.ResponseWriter, r *http.Request) {
	scheduleID := chi.URLParam(r, "scheduleId")

	rows, err := h.db.Query(`
		SELECT a.id, a.activity_id, a.activity_name, a.activity_type, a.status,
			a.original_duration, a.percent_complete, a.early_start, a.early_finish,
			a.late_start, a.late_finish, a.actual_start, a.actual_finish,
			a.is_critical, a.float_total, a.wbs_code
		FROM schedule_activities a WHERE a.schedule_id = $1
		ORDER BY COALESCE(a.early_start, a.start_date), a.activity_id`, scheduleID)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, actID, name, atype, status string
		var dur int
		var pct sql.NullFloat64
		var es, ef, ls, lf, ast, af sql.NullString
		var isCritical bool
		var floatTotal int
		var wbs sql.NullString
		if err := rows.Scan(&id, &actID, &name, &atype, &status, &dur, &pct, &es, &ef, &ls, &lf, &ast, &af, &isCritical, &floatTotal, &wbs); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		item := map[string]interface{}{
			"id": id, "activity_id": actID, "activity_name": name, "activity_type": atype,
			"status": status, "original_duration": dur, "is_critical": isCritical,
			"float_total": floatTotal,
		}
		if pct.Valid { item["percent_complete"] = pct.Float64 }
		if es.Valid { item["early_start"] = es.String }
		if ef.Valid { item["early_finish"] = ef.String }
		if ls.Valid { item["late_start"] = ls.String }
		if lf.Valid { item["late_finish"] = lf.String }
		if ast.Valid { item["actual_start"] = ast.String }
		if af.Valid { item["actual_finish"] = af.String }
		if wbs.Valid { item["wbs_code"] = wbs.String }
		items = append(items, item)
	}
	respondJSON(w, http.StatusOK, items)
}

// =============================================================================
// Baseline Comparison
// =============================================================================

func (h *ScheduleHandler) BaselineCompare(w http.ResponseWriter, r *http.Request) {
	scheduleID := chi.URLParam(r, "scheduleId")
	baselineID := r.URL.Query().Get("baseline_id")

	if baselineID == "" {
		// Get current baseline
		h.db.QueryRow(`SELECT id FROM schedule_baselines WHERE schedule_id = $1 AND is_current = TRUE`, scheduleID).Scan(&baselineID)
		if baselineID == "" {
			respondError(w, http.StatusNotFound, "no baseline found for this schedule")
			return
		}
	}

	// Get baseline activities (stored in a separate copy or the log)
	// For now, compare current vs baseline plan data
	rows, err := h.db.Query(`
		SELECT a.activity_id, a.activity_name, a.original_duration AS baseline_dur,
			a.remaining_duration, a.actual_duration, a.percent_complete,
			a.start_date AS baseline_start, a.finish_date AS baseline_finish,
			a.early_start, a.early_finish, a.float_total, a.is_critical,
			a.status
		FROM schedule_activities a
		WHERE a.schedule_id = $1
		ORDER BY a.activity_id`, scheduleID)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	var totalVariance, lateCount, criticalCount int
	for rows.Next() {
		var actID, name, status string
		var baseDur, remDur, actDur, floatTotal int
		var pct float64
		var baseStart, baseFinish, earlyStart, earlyFinish sql.NullString
		var isCritical bool
		if err := rows.Scan(&actID, &name, &baseDur, &remDur, &actDur, &pct, &baseStart, &baseFinish, &earlyStart, &earlyFinish, &floatTotal, &isCritical, &status); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		variance := actDur - baseDur
		totalVariance += variance
		if variance > 0 { lateCount++ }
		if isCritical { criticalCount++ }

		items = append(items, map[string]interface{}{
			"activity_id": actID, "activity_name": name, "status": status,
			"baseline_duration": baseDur, "remaining_duration": remDur,
			"actual_duration": actDur, "percent_complete": pct,
			"variance_days": variance, "is_critical": isCritical,
		})
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"schedule_id":        scheduleID,
		"baseline_id":        baselineID,
		"total_activities":   len(items),
		"total_variance_days": totalVariance,
		"late_activities":    lateCount,
		"critical_activities": criticalCount,
		"activities":         items,
	})
}