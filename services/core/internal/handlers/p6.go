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

// P6Handler handles Primavera P6 Connector endpoints
type P6Handler struct {
	db *sql.DB
}

func NewP6Handler(db *sql.DB) *P6Handler {
	return &P6Handler{db: db}
}

func (h *P6Handler) RegisterRoutes(r chi.Router) {
	r.Route("/p6", func(r chi.Router) {
		// Import
		r.Post("/import", h.ImportXER)

		// Projects
		r.Get("/projects", h.ListProjects)
		r.Post("/projects", h.CreateProjectMapping)
		r.Get("/projects/{id}", h.GetProjectMapping)
		r.Put("/projects/{id}", h.UpdateProjectMapping)
		r.Delete("/projects/{id}", h.DeleteProjectMapping)

		// Sync
		r.Get("/sync/status", h.SyncStatus)
		r.Post("/sync/{id}", h.TriggerSync)

		// WBS
		r.Get("/wbs", h.ListWBS)
		r.Post("/wbs", h.CreateWBS)
		r.Get("/wbs/{id}", h.GetWBS)

		// Activities
		r.Get("/activities", h.ListActivities)
		r.Post("/activities", h.CreateActivity)
		r.Get("/activities/{id}", h.GetActivity)
		r.Put("/activities/{id}", h.UpdateActivity)

		// Relationships
		r.Get("/relationships", h.ListRelationships)
		r.Post("/relationships", h.CreateRelationship)

		// Resources
		r.Get("/resources", h.ListResources)
		r.Post("/resources", h.CreateResource)

		// Sync Log
		r.Get("/sync-log", h.ListSyncLog)
		r.Get("/sync-log/{id}", h.GetSyncLog)
	})
}

// --- Import XER ---

func (h *P6Handler) ImportXER(w http.ResponseWriter, r *http.Request) {
	// In a real implementation, this would parse XER file from multipart form
	// For now, accept metadata and create sync log entry
	var input struct {
		ProjectID   string `json:"project_id"`
		P6ProjectID string `json:"p6_project_id"`
		SyncType    string `json:"sync_type"`
		SyncFile    string `json:"sync_file"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	logID := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`
		INSERT INTO p6_sync_log (id, project_id, p6_project_id, sync_type, status, started_at)
		VALUES ($1, $2, $3, $4, 'running', $5)`,
		logID, input.ProjectID, input.P6ProjectID, input.SyncType, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusAccepted, map[string]interface{}{
		"sync_log_id": logID,
		"status":      "running",
		"message":     "Import started. Use GET /api/v1/p6/sync-log/" + logID + " to check status",
	})
}

// --- Project Mapping ---

func (h *P6Handler) ListProjects(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(`
		SELECT id, project_id, p6_project_id, p6_uid, p6_project_code, p6_project_name, last_sync_at, sync_status, sync_error, config, is_active, created_at, updated_at
		FROM p6_projects ORDER BY created_at DESC`)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	items := make([]models.P6Project, 0)
	for rows.Next() {
		var p models.P6Project
		if err := rows.Scan(&p.ID, &p.ProjectID, &p.P6ProjectID, &p.P6UID, &p.P6ProjectCode, &p.P6ProjectName, &p.LastSyncAt, &p.SyncStatus, &p.SyncError, &p.Config, &p.IsActive, &p.CreatedAt, &p.UpdatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, p)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *P6Handler) CreateProjectMapping(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID     string  `json:"project_id"`
		P6ProjectID   string  `json:"p6_project_id"`
		P6UID         *string `json:"p6_uid"`
		P6ProjectCode *string `json:"p6_project_code"`
		P6ProjectName *string `json:"p6_project_name"`
		Config        *string `json:"config"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	_, err := h.db.Exec(`
		INSERT INTO p6_projects (id, project_id, p6_project_id, p6_uid, p6_project_code, p6_project_name, config)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		id, input.ProjectID, input.P6ProjectID, input.P6UID, input.P6ProjectCode, input.P6ProjectName, input.Config)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *P6Handler) GetProjectMapping(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var p models.P6Project
	err := h.db.QueryRow(`
		SELECT id, project_id, p6_project_id, p6_uid, p6_project_code, p6_project_name, last_sync_at, sync_status, sync_error, config, is_active, created_at, updated_at
		FROM p6_projects WHERE id = $1`, id).Scan(
		&p.ID, &p.ProjectID, &p.P6ProjectID, &p.P6UID, &p.P6ProjectCode, &p.P6ProjectName, &p.LastSyncAt, &p.SyncStatus, &p.SyncError, &p.Config, &p.IsActive, &p.CreatedAt, &p.UpdatedAt)
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

func (h *P6Handler) UpdateProjectMapping(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		P6ProjectID   string  `json:"p6_project_id"`
		P6UID         *string `json:"p6_uid"`
		P6ProjectCode *string `json:"p6_project_code"`
		P6ProjectName *string `json:"p6_project_name"`
		IsActive      bool    `json:"is_active"`
		Config        *string `json:"config"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	_, err := h.db.Exec(`
		UPDATE p6_projects SET p6_project_id=$1, p6_uid=$2, p6_project_code=$3, p6_project_name=$4, is_active=$5, config=$6, updated_at=NOW()
		WHERE id=$7`,
		input.P6ProjectID, input.P6UID, input.P6ProjectCode, input.P6ProjectName, input.IsActive, input.Config, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *P6Handler) DeleteProjectMapping(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM p6_projects WHERE id = $1`, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// --- Sync ---

func (h *P6Handler) SyncStatus(w http.ResponseWriter, r *http.Request) {
	// Get latest sync log entry for each project
	rows, err := h.db.Query(`
		SELECT DISTINCT ON (project_id) id, project_id, p6_project_id, sync_type, status, started_at, completed_at, duration_sec, records_processed, records_created, records_updated, records_deleted, sync_file, error_message
		FROM p6_sync_log
		WHERE project_id IS NOT NULL
		ORDER BY project_id, started_at DESC`)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var s models.P6SyncLog
		if err := rows.Scan(&s.ID, &s.ProjectID, &s.P6ProjectID, &s.SyncType, &s.Status, &s.StartedAt, &s.CompletedAt, &s.DurationSec, &s.RecordsProcessed, &s.RecordsCreated, &s.RecordsUpdated, &s.RecordsDeleted, &s.SyncFile, &s.ErrorMessage); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, map[string]interface{}{
			"id":                s.ID,
			"project_id":        s.ProjectID,
			"p6_project_id":     s.P6ProjectID,
			"sync_type":         s.SyncType,
			"status":            s.Status,
			"started_at":        s.StartedAt,
			"completed_at":      s.CompletedAt,
			"duration_sec":      s.DurationSec,
			"records_processed": s.RecordsProcessed,
			"sync_file":         s.SyncFile,
			"error_message":     s.ErrorMessage,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *P6Handler) TriggerSync(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		SyncType string `json:"sync_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	logID := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`
		INSERT INTO p6_sync_log (id, project_id, p6_project_id, sync_type, status, started_at)
		SELECT $1, project_id, p6_project_id, $2, 'running', $3
		FROM p6_projects WHERE id = $4`,
		logID, input.SyncType, now, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusAccepted, map[string]string{"sync_log_id": logID, "status": "running"})
}

// --- WBS ---

func (h *P6Handler) ListWBS(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("p6_project_id")
	query := `SELECT id, p6_project_id, p6_wbs_id, p6_wbs_code, p6_wbs_name, p6_parent_wbs_id, level, wbs_path, mapped_element_type, mapped_element_id, is_active, created_at FROM p6_wbs WHERE 1=1`
	var rows *sql.Rows
	var err error
	if projectID != "" {
		rows, err = h.db.Query(query+` AND p6_project_id = $1 ORDER BY level, p6_wbs_code`, projectID)
	} else {
		rows, err = h.db.Query(query + ` ORDER BY p6_project_id, level`)
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	items := make([]models.P6WBS, 0)
	for rows.Next() {
		var p6wbs models.P6WBS
		if err := rows.Scan(&p6wbs.ID, &p6wbs.P6ProjectID, &p6wbs.P6WBSID, &p6wbs.P6WBSCode, &p6wbs.P6WBSName, &p6wbs.P6ParentWBSID, &p6wbs.Level, &p6wbs.WBSPath, &p6wbs.MappedElementType, &p6wbs.MappedElementID, &p6wbs.IsActive, &p6wbs.CreatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, p6wbs)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *P6Handler) CreateWBS(w http.ResponseWriter, r *http.Request) {
	var input struct {
		P6ProjectID string  `json:"p6_project_id"`
		P6WBSID     string  `json:"p6_wbs_id"`
		P6WBSCode   *string `json:"p6_wbs_code"`
		P6WBSName   *string `json:"p6_wbs_name"`
		ParentWBSID *string `json:"p6_parent_wbs_id"`
		Level       int     `json:"level"`
		WBSPath     *string `json:"wbs_path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	_, err := h.db.Exec(`
		INSERT INTO p6_wbs (id, p6_project_id, p6_wbs_id, p6_wbs_code, p6_wbs_name, p6_parent_wbs_id, level, wbs_path)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		id, input.P6ProjectID, input.P6WBSID, input.P6WBSCode, input.P6WBSName, input.ParentWBSID, input.Level, input.WBSPath)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *P6Handler) GetWBS(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var p6wbs_item models.P6WBS
	err := h.db.QueryRow(`
		SELECT id, p6_project_id, p6_wbs_id, p6_wbs_code, p6_wbs_name, p6_parent_wbs_id, level, wbs_path, mapped_element_type, mapped_element_id, is_active, created_at
		FROM p6_wbs WHERE id = $1`, id).Scan(
		&p6wbs_item.ID, &p6wbs_item.P6ProjectID, &p6wbs_item.P6WBSID, &p6wbs_item.P6WBSCode, &p6wbs_item.P6WBSName, &p6wbs_item.P6ParentWBSID, &p6wbs_item.Level, &p6wbs_item.WBSPath, &p6wbs_item.MappedElementType, &p6wbs_item.MappedElementID, &p6wbs_item.IsActive, &p6wbs_item.CreatedAt)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, p6wbs_item)
}

// --- Activities ---

func (h *P6Handler) ListActivities(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("p6_project_id")
	query := `SELECT id, p6_project_id, p6_wbs_id, p6_activity_id, p6_activity_code, p6_activity_name, activity_type, status, planned_start, planned_finish, actual_start, actual_finish, remaining_duration, at_completion_duration, percent_complete, physical_complete, duration_type, mapped_to_type, mapped_element_id, is_active, last_sync_at, created_at FROM p6_activities WHERE 1=1`
	var rows *sql.Rows
	var err error
	if projectID != "" {
		rows, err = h.db.Query(query+` AND p6_project_id = $1 ORDER BY p6_activity_code`, projectID)
	} else {
		rows, err = h.db.Query(query + ` ORDER BY p6_project_id, p6_activity_code`)
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	items := make([]models.P6Activity, 0)
	for rows.Next() {
		var a models.P6Activity
		if err := rows.Scan(&a.ID, &a.P6ProjectID, &a.P6WBSID, &a.P6ActivityID, &a.P6ActivityCode, &a.P6ActivityName,
			&a.ActivityType, &a.Status, &a.PlannedStart, &a.PlannedFinish, &a.ActualStart, &a.ActualFinish,
			&a.RemainingDuration, &a.AtCompletionDuration, &a.PercentComplete, &a.PhysicalComplete,
			&a.DurationType, &a.MappedToType, &a.MappedElementID, &a.IsActive, &a.LastSyncAt, &a.CreatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, a)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *P6Handler) CreateActivity(w http.ResponseWriter, r *http.Request) {
	var input struct {
		P6ProjectID      string     `json:"p6_project_id"`
		P6WBSID          *string    `json:"p6_wbs_id"`
		P6ActivityID     string     `json:"p6_activity_id"`
		P6ActivityCode   *string    `json:"p6_activity_code"`
		P6ActivityName   *string    `json:"p6_activity_name"`
		ActivityType     *string    `json:"activity_type"`
		Status           *string    `json:"status"`
		PlannedStart     *time.Time `json:"planned_start"`
		PlannedFinish    *time.Time `json:"planned_finish"`
		ActualStart      *time.Time `json:"actual_start"`
		ActualFinish     *time.Time `json:"actual_finish"`
		RemainingDuration *int      `json:"remaining_duration"`
		PercentComplete  *float64   `json:"percent_complete"`
		PhysicalComplete *float64   `json:"physical_complete"`
		DurationType     *string    `json:"duration_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	_, err := h.db.Exec(`
		INSERT INTO p6_activities (id, p6_project_id, p6_wbs_id, p6_activity_id, p6_activity_code, p6_activity_name, activity_type, status, planned_start, planned_finish, actual_start, actual_finish, remaining_duration, percent_complete, physical_complete, duration_type)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)`,
		id, input.P6ProjectID, input.P6WBSID, input.P6ActivityID, input.P6ActivityCode, input.P6ActivityName,
		input.ActivityType, input.Status, input.PlannedStart, input.PlannedFinish, input.ActualStart, input.ActualFinish,
		input.RemainingDuration, input.PercentComplete, input.PhysicalComplete, input.DurationType)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *P6Handler) GetActivity(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var a models.P6Activity
	err := h.db.QueryRow(`
		SELECT id, p6_project_id, p6_wbs_id, p6_activity_id, p6_activity_code, p6_activity_name, activity_type, status, planned_start, planned_finish, actual_start, actual_finish, remaining_duration, at_completion_duration, percent_complete, physical_complete, duration_type, mapped_to_type, mapped_element_id, is_active, last_sync_at, created_at
		FROM p6_activities WHERE id = $1`, id).Scan(
		&a.ID, &a.P6ProjectID, &a.P6WBSID, &a.P6ActivityID, &a.P6ActivityCode, &a.P6ActivityName,
		&a.ActivityType, &a.Status, &a.PlannedStart, &a.PlannedFinish, &a.ActualStart, &a.ActualFinish,
		&a.RemainingDuration, &a.AtCompletionDuration, &a.PercentComplete, &a.PhysicalComplete,
		&a.DurationType, &a.MappedToType, &a.MappedElementID, &a.IsActive, &a.LastSyncAt, &a.CreatedAt)
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

func (h *P6Handler) UpdateActivity(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Status           *string   `json:"status"`
		PlannedStart     *string   `json:"planned_start"`
		PlannedFinish    *string   `json:"planned_finish"`
		ActualStart      *string   `json:"actual_start"`
		ActualFinish     *string   `json:"actual_finish"`
		RemainingDuration *int     `json:"remaining_duration"`
		PercentComplete  *float64  `json:"percent_complete"`
		PhysicalComplete *float64  `json:"physical_complete"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	_, err := h.db.Exec(`
		UPDATE p6_activities SET status=$1, planned_start=$2, planned_finish=$3, actual_start=$4, actual_finish=$5, remaining_duration=$6, percent_complete=$7, physical_complete=$8, last_sync_at=NOW()
		WHERE id=$9`,
		input.Status, input.PlannedStart, input.PlannedFinish, input.ActualStart, input.ActualFinish,
		input.RemainingDuration, input.PercentComplete, input.PhysicalComplete, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// --- Relationships ---

func (h *P6Handler) ListRelationships(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("p6_project_id")
	query := `SELECT id, p6_project_id, predecessor_id, successor_id, relationship_type, lag_days, lag_type, is_active, created_at FROM p6_relationships WHERE 1=1`
	var rows *sql.Rows
	var err error
	if projectID != "" {
		rows, err = h.db.Query(query+` AND p6_project_id = $1 ORDER BY predecessor_id`, projectID)
	} else {
		rows, err = h.db.Query(query + ` ORDER BY p6_project_id`)
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	items := make([]models.P6Relationship, 0)
	for rows.Next() {
		var rl models.P6Relationship
		if err := rows.Scan(&rl.ID, &rl.P6ProjectID, &rl.PredecessorID, &rl.SuccessorID, &rl.RelationshipType, &rl.LagDays, &rl.LagType, &rl.IsActive, &rl.CreatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, rl)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *P6Handler) CreateRelationship(w http.ResponseWriter, r *http.Request) {
	var input struct {
		P6ProjectID      string `json:"p6_project_id"`
		PredecessorID    string `json:"predecessor_id"`
		SuccessorID      string `json:"successor_id"`
		RelationshipType string `json:"relationship_type"`
		LagDays          int    `json:"lag_days"`
		LagType          string `json:"lag_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	_, err := h.db.Exec(`
		INSERT INTO p6_relationships (id, p6_project_id, predecessor_id, successor_id, relationship_type, lag_days, lag_type)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		id, input.P6ProjectID, input.PredecessorID, input.SuccessorID, input.RelationshipType, input.LagDays, input.LagType)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

// --- Resources ---

func (h *P6Handler) ListResources(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("p6_project_id")
	query := `SELECT id, p6_project_id, p6_resource_id, p6_resource_name, resource_type, unit_of_measure, unit_price, currency, mapped_to_type, mapped_element_id, is_active, created_at FROM p6_resources WHERE 1=1`
	var rows *sql.Rows
	var err error
	if projectID != "" {
		rows, err = h.db.Query(query+` AND p6_project_id = $1 ORDER BY p6_resource_name`, projectID)
	} else {
		rows, err = h.db.Query(query + ` ORDER BY p6_project_id`)
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	items := make([]models.P6Resource, 0)
	for rows.Next() {
		var rl models.P6Resource
		if err := rows.Scan(&rl.ID, &rl.P6ProjectID, &rl.P6ResourceID, &rl.P6ResourceName, &rl.ResourceType, &rl.UnitOfMeasure, &rl.UnitPrice, &rl.Currency, &rl.MappedToType, &rl.MappedElementID, &rl.IsActive, &rl.CreatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, rl)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *P6Handler) CreateResource(w http.ResponseWriter, r *http.Request) {
	var input struct {
		P6ProjectID    string   `json:"p6_project_id"`
		P6ResourceID   string   `json:"p6_resource_id"`
		P6ResourceName *string  `json:"p6_resource_name"`
		ResourceType   *string  `json:"resource_type"`
		UnitOfMeasure  *string  `json:"unit_of_measure"`
		UnitPrice      *float64 `json:"unit_price"`
		Currency       string   `json:"currency"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	_, err := h.db.Exec(`
		INSERT INTO p6_resources (id, p6_project_id, p6_resource_id, p6_resource_name, resource_type, unit_of_measure, unit_price, currency)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		id, input.P6ProjectID, input.P6ResourceID, input.P6ResourceName, input.ResourceType, input.UnitOfMeasure, input.UnitPrice, input.Currency)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

// --- Sync Log ---

func (h *P6Handler) ListSyncLog(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT id, project_id, p6_project_id, sync_type, status, started_at, completed_at, duration_sec, records_processed, records_created, records_updated, records_deleted, sync_file, error_message, details FROM p6_sync_log WHERE 1=1`
	var rows *sql.Rows
	var err error
	if projectID != "" {
		rows, err = h.db.Query(query+` AND project_id = $1 ORDER BY started_at DESC LIMIT 50`, projectID)
	} else {
		rows, err = h.db.Query(query + ` ORDER BY started_at DESC LIMIT 50`)
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	items := make([]models.P6SyncLog, 0)
	for rows.Next() {
		var s models.P6SyncLog
		if err := rows.Scan(&s.ID, &s.ProjectID, &s.P6ProjectID, &s.SyncType, &s.Status, &s.StartedAt,
			&s.CompletedAt, &s.DurationSec, &s.RecordsProcessed, &s.RecordsCreated, &s.RecordsUpdated,
			&s.RecordsDeleted, &s.SyncFile, &s.ErrorMessage, &s.Details); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, s)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *P6Handler) GetSyncLog(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var s models.P6SyncLog
	err := h.db.QueryRow(`
		SELECT id, project_id, p6_project_id, sync_type, status, started_at, completed_at, duration_sec, records_processed, records_created, records_updated, records_deleted, sync_file, error_message, details
		FROM p6_sync_log WHERE id = $1`, id).Scan(
		&s.ID, &s.ProjectID, &s.P6ProjectID, &s.SyncType, &s.Status, &s.StartedAt,
		&s.CompletedAt, &s.DurationSec, &s.RecordsProcessed, &s.RecordsCreated, &s.RecordsUpdated,
		&s.RecordsDeleted, &s.SyncFile, &s.ErrorMessage, &s.Details)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, s)
}