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

// DocControlHandler handles Document Control module endpoints
type DocControlHandler struct {
	db *sql.DB
}

func NewDocControlHandler(db *sql.DB) *DocControlHandler {
	return &DocControlHandler{db: db}
}

func (h *DocControlHandler) RegisterRoutes(r chi.Router) {
	r.Route("/doc-control", func(r chi.Router) {
		// RFI
		r.Get("/rfis", h.ListRFIs)
		r.Post("/rfis", h.CreateRFI)
		r.Get("/rfis/{id}", h.GetRFI)
		r.Put("/rfis/{id}", h.UpdateRFI)
		r.Delete("/rfis/{id}", h.DeleteRFI)

		// NCR
		r.Get("/ncrs", h.ListNCRs)
		r.Post("/ncrs", h.CreateNCR)
		r.Get("/ncrs/{id}", h.GetNCR)
		r.Put("/ncrs/{id}", h.UpdateNCR)
		r.Delete("/ncrs/{id}", h.DeleteNCR)

		// Submittals
		r.Get("/submittals", h.ListSubmittals)
		r.Post("/submittals", h.CreateSubmittal)
		r.Get("/submittals/{id}", h.GetSubmittal)
		r.Put("/submittals/{id}", h.UpdateSubmittal)
		r.Delete("/submittals/{id}", h.DeleteSubmittal)

		// Method Statements
		r.Get("/method-statements", h.ListMethodStatements)
		r.Post("/method-statements", h.CreateMethodStatement)
		r.Get("/method-statements/{id}", h.GetMethodStatement)
		r.Put("/method-statements/{id}", h.UpdateMethodStatement)
		r.Delete("/method-statements/{id}", h.DeleteMethodStatement)

		// Shop Drawings
		r.Get("/shop-drawings", h.ListShopDrawings)
		r.Post("/shop-drawings", h.CreateShopDrawing)
		r.Get("/shop-drawings/{id}", h.GetShopDrawing)
		r.Put("/shop-drawings/{id}", h.UpdateShopDrawing)
		r.Delete("/shop-drawings/{id}", h.DeleteShopDrawing)

		// Correspondence
		r.Get("/correspondence", h.ListCorrespondence)
		r.Post("/correspondence", h.CreateCorrespondence)
		r.Get("/correspondence/{id}", h.GetCorrespondence)
		r.Put("/correspondence/{id}", h.UpdateCorrespondence)
		r.Delete("/correspondence/{id}", h.DeleteCorrespondence)

		// Minutes of Meeting
		r.Get("/minutes-of-meeting", h.ListMinutesOfMeeting)
		r.Post("/minutes-of-meeting", h.CreateMinutesOfMeeting)
		r.Get("/minutes-of-meeting/{id}", h.GetMinutesOfMeeting)
		r.Put("/minutes-of-meeting/{id}", h.UpdateMinutesOfMeeting)
		r.Delete("/minutes-of-meeting/{id}", h.DeleteMinutesOfMeeting)

		// Daily Reports (doc_daily_reports)
		r.Get("/daily-reports", h.ListDailyReports)
		r.Post("/daily-reports", h.CreateDailyReport)
		r.Get("/daily-reports/{id}", h.GetDailyReport)
		r.Put("/daily-reports/{id}", h.UpdateDailyReport)
		r.Delete("/daily-reports/{id}", h.DeleteDailyReport)

		// Document Transmittals
		r.Get("/transmittals", h.ListTransmittals)
		r.Post("/transmittals", h.CreateTransmittal)
		r.Get("/transmittals/{id}", h.GetTransmittal)
		r.Put("/transmittals/{id}", h.UpdateTransmittal)
		r.Delete("/transmittals/{id}", h.DeleteTransmittal)

		// Document Revisions
		r.Get("/revisions", h.ListRevisions)
		r.Post("/revisions", h.CreateRevision)
		r.Get("/revisions/{id}", h.GetRevision)
		r.Delete("/revisions/{id}", h.DeleteRevision)

		// Summary
		r.Get("/summary", h.GetSummary)
	})
}

// =============================================================================
// RFI
// =============================================================================

func (h *DocControlHandler) ListRFIs(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	projectID := r.URL.Query().Get("project_id")
	discipline := r.URL.Query().Get("discipline")

	query := `SELECT id, project_id, rfi_number, rfi_code, subject, question, answer, discipline, priority, raised_by, assigned_to, status, due_date, raised_at, answered_at, closed_at, created_at, updated_at FROM rfi_documents WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIdx); argIdx++
		args = append(args, status)
	}
	if projectID != "" {
		query += fmt.Sprintf(" AND project_id = $%d", argIdx); argIdx++
		args = append(args, projectID)
	}
	if discipline != "" {
		query += fmt.Sprintf(" AND discipline = $%d", argIdx); argIdx++
		args = append(args, discipline)
	}
	query += " ORDER BY rfi_number DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, projectID, rfiCode, subject, question, discipline, raisedBy, assignedTo, status string
		var rfiNumber int
		var answer, priority sql.NullString
		var dueDate, answeredAt, closedAt sql.NullTime
		var raisedAt, createdAt, updatedAt time.Time

		err := rows.Scan(&id, &projectID, &rfiNumber, &rfiCode, &subject, &question, &answer, &discipline, &priority, &raisedBy, &assignedTo, &status, &dueDate, &raisedAt, &answeredAt, &closedAt, &createdAt, &updatedAt)
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		item := map[string]interface{}{
			"id": id, "project_id": projectID, "rfi_number": rfiNumber, "rfi_code": rfiCode,
			"subject": subject, "question": question, "discipline": discipline,
			"raised_by": raisedBy, "assigned_to": assignedTo, "status": status,
			"raised_at": raisedAt, "created_at": createdAt, "updated_at": updatedAt,
		}
		if answer.Valid { item["answer"] = answer.String }
		if priority.Valid { item["priority"] = priority.String }
		if dueDate.Valid { item["due_date"] = dueDate.Time }
		if answeredAt.Valid { item["answered_at"] = answeredAt.Time }
		if closedAt.Valid { item["closed_at"] = closedAt.Time }
		items = append(items, item)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *DocControlHandler) CreateRFI(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID  string  `json:"project_id"`
		RFINumber  *int    `json:"rfi_number"`
		RFICode    string  `json:"rfi_code"`
		Subject    string  `json:"subject"`
		Question   string  `json:"question"`
		Discipline *string `json:"discipline"`
		Priority   *string `json:"priority"`
		RaisedBy   *string `json:"raised_by"`
		AssignedTo *string `json:"assigned_to"`
		DueDate    *string `json:"due_date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO rfi_documents (id, project_id, rfi_number, rfi_code, subject, question, discipline, priority, raised_by, assigned_to, due_date, raised_at, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		id, input.ProjectID, input.RFINumber, input.RFICode, input.Subject, input.Question, input.Discipline, input.Priority, input.RaisedBy, input.AssignedTo, input.DueDate, now, now, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *DocControlHandler) GetRFI(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var item struct {
		ID, ProjectID, RFICode, Subject, Question, Discipline, RaisedBy, AssignedTo, Status string
		RFINumber int
		Answer, Priority sql.NullString
		DueDate, AnsweredAt, ClosedAt sql.NullTime
		RaisedAt, CreatedAt, UpdatedAt time.Time
	}
	err := h.db.QueryRow(`SELECT id, project_id, rfi_number, rfi_code, subject, question, answer, discipline, priority, raised_by, assigned_to, status, due_date, raised_at, answered_at, closed_at, created_at, updated_at FROM rfi_documents WHERE id = $1`, id).
		Scan(&item.ID, &item.ProjectID, &item.RFINumber, &item.RFICode, &item.Subject, &item.Question, &item.Answer, &item.Discipline, &item.Priority, &item.RaisedBy, &item.AssignedTo, &item.Status, &item.DueDate, &item.RaisedAt, &item.AnsweredAt, &item.ClosedAt, &item.CreatedAt, &item.UpdatedAt)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "RFI not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	resp := map[string]interface{}{
		"id": item.ID, "project_id": item.ProjectID, "rfi_number": item.RFINumber, "rfi_code": item.RFICode,
		"subject": item.Subject, "question": item.Question, "discipline": item.Discipline,
		"raised_by": item.RaisedBy, "assigned_to": item.AssignedTo, "status": item.Status,
		"raised_at": item.RaisedAt, "created_at": item.CreatedAt, "updated_at": item.UpdatedAt,
	}
	if item.Answer.Valid { resp["answer"] = item.Answer.String }
	if item.Priority.Valid { resp["priority"] = item.Priority.String }
	if item.DueDate.Valid { resp["due_date"] = item.DueDate.Time }
	if item.AnsweredAt.Valid { resp["answered_at"] = item.AnsweredAt.Time }
	if item.ClosedAt.Valid { resp["closed_at"] = item.ClosedAt.Time }
	respondJSON(w, http.StatusOK, resp)
}

func (h *DocControlHandler) UpdateRFI(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Answer     *string `json:"answer"`
		Status     *string `json:"status"`
		AssignedTo *string `json:"assigned_to"`
		Priority   *string `json:"priority"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	now := time.Now()
	_, err := h.db.Exec(`UPDATE rfi_documents SET answer=COALESCE($1,answer), status=COALESCE($2,status), assigned_to=COALESCE($3,assigned_to), priority=COALESCE($4,priority), updated_at=$5 WHERE id=$6`,
		input.Answer, input.Status, input.AssignedTo, input.Priority, now, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *DocControlHandler) DeleteRFI(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM rfi_documents WHERE id = $1`, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// NCR
// =============================================================================

func (h *DocControlHandler) ListNCRs(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	projectID := r.URL.Query().Get("project_id")
	severity := r.URL.Query().Get("severity")

	query := `SELECT id, project_id, ncr_number, ncr_code, title, description, location, ncr_type, severity, source, reported_by, assigned_to, root_cause, corrective_action, preventive_action, status, due_date, reported_at, closed_at, created_at, updated_at FROM ncr_documents WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if status != "" { query += fmt.Sprintf(" AND status = $%d", argIdx); argIdx++; args = append(args, status) }
	if projectID != "" { query += fmt.Sprintf(" AND project_id = $%d", argIdx); argIdx++; args = append(args, projectID) }
	if severity != "" { query += fmt.Sprintf(" AND severity = $%d", argIdx); argIdx++; args = append(args, severity) }
	query += " ORDER BY ncr_number DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, projectID, ncrCode, title, description, ncrType, severity, source, reportedBy, assignedTo, status string
		var ncrNumber int
		var location, rootCause, correctiveAction, preventiveAction sql.NullString
		var dueDate, closedAt sql.NullTime
		var reportedAt, createdAt, updatedAt time.Time

		err := rows.Scan(&id, &projectID, &ncrNumber, &ncrCode, &title, &description, &location, &ncrType, &severity, &source, &reportedBy, &assignedTo, &rootCause, &correctiveAction, &preventiveAction, &status, &dueDate, &reportedAt, &closedAt, &createdAt, &updatedAt)
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		item := map[string]interface{}{
			"id": id, "project_id": projectID, "ncr_number": ncrNumber, "ncr_code": ncrCode,
			"title": title, "description": description, "ncr_type": ncrType,
			"severity": severity, "source": source, "reported_by": reportedBy,
			"assigned_to": assignedTo, "status": status,
			"reported_at": reportedAt, "created_at": createdAt, "updated_at": updatedAt,
		}
		if location.Valid { item["location"] = location.String }
		if rootCause.Valid { item["root_cause"] = rootCause.String }
		if correctiveAction.Valid { item["corrective_action"] = correctiveAction.String }
		if preventiveAction.Valid { item["preventive_action"] = preventiveAction.String }
		if dueDate.Valid { item["due_date"] = dueDate.Time }
		if closedAt.Valid { item["closed_at"] = closedAt.Time }
		items = append(items, item)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *DocControlHandler) CreateNCR(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID string `json:"project_id"`
		NCRNumber *int   `json:"ncr_number"`
		NCRCode   string `json:"ncr_code"`
		Title     string `json:"title"`
		Severity  *string `json:"severity"`
		NCRType   *string `json:"ncr_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO ncr_documents (id, project_id, ncr_number, ncr_code, title, severity, ncr_type, reported_at, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		id, input.ProjectID, input.NCRNumber, input.NCRCode, input.Title, input.Severity, input.NCRType, now, now, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *DocControlHandler) GetNCR(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	// Reuse the list query pattern for simplicity
	var title, ncrCode string
	var ncrNumber int
	err := h.db.QueryRow(`SELECT ncr_code, title, ncr_number FROM ncr_documents WHERE id = $1`, id).Scan(&ncrCode, &title, &ncrNumber)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "NCR not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "ncr_code": ncrCode, "title": title, "ncr_number": ncrNumber})
}

func (h *DocControlHandler) UpdateNCR(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Status     *string `json:"status"`
		RootCause  *string `json:"root_cause"`
		CorrectiveAction *string `json:"corrective_action"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	now := time.Now()
	_, err := h.db.Exec(`UPDATE ncr_documents SET status=COALESCE($1,status), root_cause=COALESCE($2,root_cause), corrective_action=COALESCE($3,corrective_action), updated_at=$4 WHERE id=$5`,
		input.Status, input.RootCause, input.CorrectiveAction, now, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *DocControlHandler) DeleteNCR(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM ncr_documents WHERE id = $1`, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// Submittals
// =============================================================================

func (h *DocControlHandler) ListSubmittals(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	projectID := r.URL.Query().Get("project_id")

	query := `SELECT id, project_id, submittal_number, submittal_code, title, submittal_type, specification_ref, submitted_by, submitted_to, status, resubmit_count, submitted_at, reviewed_at, approved_at, created_at, updated_at FROM submittals WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if status != "" { query += fmt.Sprintf(" AND status = $%d", argIdx); argIdx++; args = append(args, status) }
	if projectID != "" { query += fmt.Sprintf(" AND project_id = $%d", argIdx); argIdx++; args = append(args, projectID) }
	query += " ORDER BY submittal_number DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, projectID, subCode, title, subType, specRef, submittedBy, submittedTo, status string
		var subNumber, resubmitCount int
		var submittedAt, reviewedAt, approvedAt, createdAt, updatedAt time.Time

		err := rows.Scan(&id, &projectID, &subNumber, &subCode, &title, &subType, &specRef, &submittedBy, &submittedTo, &status, &resubmitCount, &submittedAt, &reviewedAt, &approvedAt, &createdAt, &updatedAt)
		if err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, map[string]interface{}{
			"id": id, "project_id": projectID, "submittal_number": subNumber, "submittal_code": subCode,
			"title": title, "submittal_type": subType, "specification_ref": specRef,
			"submitted_by": submittedBy, "submitted_to": submittedTo, "status": status,
			"resubmit_count": resubmitCount, "submitted_at": submittedAt,
			"reviewed_at": reviewedAt, "approved_at": approvedAt,
			"created_at": createdAt, "updated_at": updatedAt,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *DocControlHandler) CreateSubmittal(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID      string  `json:"project_id"`
		SubmittalCode  string  `json:"submittal_code"`
		Title          string  `json:"title"`
		SubmittalType  *string `json:"submittal_type"`
		SubmittedBy    *string `json:"submitted_by"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO submittals (id, project_id, submittal_code, title, submittal_type, submitted_by, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		id, input.ProjectID, input.SubmittalCode, input.Title, input.SubmittalType, input.SubmittedBy, now, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *DocControlHandler) GetSubmittal(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var title, subCode string
	err := h.db.QueryRow(`SELECT submittal_code, title FROM submittals WHERE id = $1`, id).Scan(&subCode, &title)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "submittal not found")
		return
	}
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"id": id, "submittal_code": subCode, "title": title})
}

func (h *DocControlHandler) UpdateSubmittal(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Status *string `json:"status"`
		ReviewNotes *string `json:"review_notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	now := time.Now()
	_, err := h.db.Exec(`UPDATE submittals SET status=COALESCE($1,status), review_notes=COALESCE($2,review_notes), updated_at=$3 WHERE id=$4`,
		input.Status, input.ReviewNotes, now, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *DocControlHandler) DeleteSubmittal(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM submittals WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// Method Statements
// =============================================================================

func (h *DocControlHandler) ListMethodStatements(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	projectID := r.URL.Query().Get("project_id")

	query := `SELECT id, project_id, ms_number, ms_code, title, work_area, activity, status, submitted_at, approved_at, created_at, updated_at FROM method_statements WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if status != "" { query += fmt.Sprintf(" AND status = $%d", argIdx); argIdx++; args = append(args, status) }
	if projectID != "" { query += fmt.Sprintf(" AND project_id = $%d", argIdx); argIdx++; args = append(args, projectID) }
	query += " ORDER BY ms_number DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, projectID, msCode, title, workArea, activity, status string
		var msNumber int
		var submittedAt, approvedAt, createdAt, updatedAt time.Time

		err := rows.Scan(&id, &projectID, &msNumber, &msCode, &title, &workArea, &activity, &status, &submittedAt, &approvedAt, &createdAt, &updatedAt)
		if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
		items = append(items, map[string]interface{}{
			"id": id, "project_id": projectID, "ms_number": msNumber, "ms_code": msCode,
			"title": title, "work_area": workArea, "activity": activity, "status": status,
			"submitted_at": submittedAt, "approved_at": approvedAt,
			"created_at": createdAt, "updated_at": updatedAt,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *DocControlHandler) CreateMethodStatement(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID string `json:"project_id"`
		MSCode    string `json:"ms_code"`
		Title     string `json:"title"`
		Activity  *string `json:"activity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO method_statements (id, project_id, ms_code, title, activity, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		id, input.ProjectID, input.MSCode, input.Title, input.Activity, now, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *DocControlHandler) GetMethodStatement(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var title, msCode string
	err := h.db.QueryRow(`SELECT ms_code, title FROM method_statements WHERE id = $1`, id).Scan(&msCode, &title)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "method statement not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"id": id, "ms_code": msCode, "title": title})
}

func (h *DocControlHandler) UpdateMethodStatement(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Status *string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	now := time.Now()
	_, err := h.db.Exec(`UPDATE method_statements SET status=COALESCE($1,status), updated_at=$2 WHERE id=$3`, input.Status, now, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *DocControlHandler) DeleteMethodStatement(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM method_statements WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// Shop Drawings
// =============================================================================

func (h *DocControlHandler) ListShopDrawings(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	projectID := r.URL.Query().Get("project_id")
	discipline := r.URL.Query().Get("discipline")

	query := `SELECT id, project_id, drawing_number, drawing_code, title, discipline, drawing_format, revision, file_path, submitted_by, checked_by, status, resubmit_count, submitted_at, approved_at, created_at, updated_at FROM shop_drawings WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if status != "" { query += fmt.Sprintf(" AND status = $%d", argIdx); argIdx++; args = append(args, status) }
	if projectID != "" { query += fmt.Sprintf(" AND project_id = $%d", argIdx); argIdx++; args = append(args, projectID) }
	if discipline != "" { query += fmt.Sprintf(" AND discipline = $%d", argIdx); argIdx++; args = append(args, discipline) }
	query += " ORDER BY drawing_number DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, projectID, drCode, title, discipline, drFormat, revision, filePath, submittedBy, checkedBy, status string
		var drNumber, resubmitCount int
		var submittedAt, approvedAt, createdAt, updatedAt time.Time

		err := rows.Scan(&id, &projectID, &drNumber, &drCode, &title, &discipline, &drFormat, &revision, &filePath, &submittedBy, &checkedBy, &status, &resubmitCount, &submittedAt, &approvedAt, &createdAt, &updatedAt)
		if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
		items = append(items, map[string]interface{}{
			"id": id, "project_id": projectID, "drawing_number": drNumber, "drawing_code": drCode,
			"title": title, "discipline": discipline, "drawing_format": drFormat,
			"revision": revision, "file_path": filePath, "submitted_by": submittedBy,
			"checked_by": checkedBy, "status": status, "resubmit_count": resubmitCount,
			"submitted_at": submittedAt, "approved_at": approvedAt,
			"created_at": createdAt, "updated_at": updatedAt,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *DocControlHandler) CreateShopDrawing(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID   string  `json:"project_id"`
		DrawingCode string  `json:"drawing_code"`
		Title       string  `json:"title"`
		Discipline  *string `json:"discipline"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO shop_drawings (id, project_id, drawing_code, title, discipline, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		id, input.ProjectID, input.DrawingCode, input.Title, input.Discipline, now, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *DocControlHandler) GetShopDrawing(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var title, drCode string
	err := h.db.QueryRow(`SELECT drawing_code, title FROM shop_drawings WHERE id = $1`, id).Scan(&drCode, &title)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "shop drawing not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"id": id, "drawing_code": drCode, "title": title})
}

func (h *DocControlHandler) UpdateShopDrawing(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Status *string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	now := time.Now()
	_, err := h.db.Exec(`UPDATE shop_drawings SET status=COALESCE($1,status), updated_at=$2 WHERE id=$3`, input.Status, now, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *DocControlHandler) DeleteShopDrawing(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM shop_drawings WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// Correspondence
// =============================================================================

func (h *DocControlHandler) ListCorrespondence(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	projectID := r.URL.Query().Get("project_id")
	corrType := r.URL.Query().Get("corr_type")
	direction := r.URL.Query().Get("direction")

	query := `SELECT id, project_id, corr_number, corr_code, subject, corr_type, direction, from_entity, to_entity, priority, status, sent_at, received_at, created_at, updated_at FROM correspondence WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if status != "" { query += fmt.Sprintf(" AND status = $%d", argIdx); argIdx++; args = append(args, status) }
	if projectID != "" { query += fmt.Sprintf(" AND project_id = $%d", argIdx); argIdx++; args = append(args, projectID) }
	if corrType != "" { query += fmt.Sprintf(" AND corr_type = $%d", argIdx); argIdx++; args = append(args, corrType) }
	if direction != "" { query += fmt.Sprintf(" AND direction = $%d", argIdx); argIdx++; args = append(args, direction) }
	query += " ORDER BY corr_number DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, projectID, corrCode, subject, corrType, direction, fromEntity, toEntity, status string
		var corrNumber int
		var priority sql.NullString
		var sentAt, receivedAt, createdAt, updatedAt time.Time

		err := rows.Scan(&id, &projectID, &corrNumber, &corrCode, &subject, &corrType, &direction, &fromEntity, &toEntity, &priority, &status, &sentAt, &receivedAt, &createdAt, &updatedAt)
		if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
		item := map[string]interface{}{
			"id": id, "project_id": projectID, "corr_number": corrNumber, "corr_code": corrCode,
			"subject": subject, "corr_type": corrType, "direction": direction,
			"from_entity": fromEntity, "to_entity": toEntity, "status": status,
			"sent_at": sentAt, "received_at": receivedAt,
			"created_at": createdAt, "updated_at": updatedAt,
		}
		if priority.Valid { item["priority"] = priority.String }
		items = append(items, item)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *DocControlHandler) CreateCorrespondence(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID string `json:"project_id"`
		CorrCode  string `json:"corr_code"`
		Subject   string `json:"subject"`
		Body      string `json:"body"`
		Direction *string `json:"direction"`
		CorrType  *string `json:"corr_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO correspondence (id, project_id, corr_code, subject, body, direction, corr_type, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		id, input.ProjectID, input.CorrCode, input.Subject, input.Body, input.Direction, input.CorrType, now, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *DocControlHandler) GetCorrespondence(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var subject, corrCode string
	err := h.db.QueryRow(`SELECT corr_code, subject FROM correspondence WHERE id = $1`, id).Scan(&corrCode, &subject)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "correspondence not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"id": id, "corr_code": corrCode, "subject": subject})
}

func (h *DocControlHandler) UpdateCorrespondence(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Status *string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	now := time.Now()
	_, err := h.db.Exec(`UPDATE correspondence SET status=COALESCE($1,status), updated_at=$2 WHERE id=$3`, input.Status, now, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *DocControlHandler) DeleteCorrespondence(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM correspondence WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// Minutes of Meeting
// =============================================================================

func (h *DocControlHandler) ListMinutesOfMeeting(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	meetingType := r.URL.Query().Get("meeting_type")

	query := `SELECT id, project_id, mom_number, mom_code, meeting_title, meeting_type, meeting_date, location, chairperson, status, distributed_at, approved_at, created_at, updated_at FROM minutes_of_meeting WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if projectID != "" { query += fmt.Sprintf(" AND project_id = $%d", argIdx); argIdx++; args = append(args, projectID) }
	if meetingType != "" { query += fmt.Sprintf(" AND meeting_type = $%d", argIdx); argIdx++; args = append(args, meetingType) }
	query += " ORDER BY meeting_date DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, projectID, momCode, meetingTitle, meetingType, location, chairperson, status string
		var momNumber int
		var meetingDate time.Time
		var distributedAt, approvedAt, createdAt, updatedAt time.Time

		err := rows.Scan(&id, &projectID, &momNumber, &momCode, &meetingTitle, &meetingType, &meetingDate, &location, &chairperson, &status, &distributedAt, &approvedAt, &createdAt, &updatedAt)
		if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
		items = append(items, map[string]interface{}{
			"id": id, "project_id": projectID, "mom_number": momNumber, "mom_code": momCode,
			"meeting_title": meetingTitle, "meeting_type": meetingType, "meeting_date": meetingDate,
			"location": location, "chairperson": chairperson, "status": status,
			"distributed_at": distributedAt, "approved_at": approvedAt,
			"created_at": createdAt, "updated_at": updatedAt,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *DocControlHandler) CreateMinutesOfMeeting(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID    string `json:"project_id"`
		MOMCode      string `json:"mom_code"`
		MeetingTitle string `json:"meeting_title"`
		MeetingType  *string `json:"meeting_type"`
		MeetingDate  string `json:"meeting_date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO minutes_of_meeting (id, project_id, mom_code, meeting_title, meeting_type, meeting_date, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		id, input.ProjectID, input.MOMCode, input.MeetingTitle, input.MeetingType, input.MeetingDate, now, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *DocControlHandler) GetMinutesOfMeeting(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var title, momCode string
	err := h.db.QueryRow(`SELECT mom_code, meeting_title FROM minutes_of_meeting WHERE id = $1`, id).Scan(&momCode, &title)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "minutes not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"id": id, "mom_code": momCode, "meeting_title": title})
}

func (h *DocControlHandler) UpdateMinutesOfMeeting(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Status *string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	now := time.Now()
	_, err := h.db.Exec(`UPDATE minutes_of_meeting SET status=COALESCE($1,status), updated_at=$2 WHERE id=$3`, input.Status, now, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *DocControlHandler) DeleteMinutesOfMeeting(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM minutes_of_meeting WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// Daily Reports
// =============================================================================

func (h *DocControlHandler) ListDailyReports(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	status := r.URL.Query().Get("status")
	fromDate := r.URL.Query().Get("from_date")
	toDate := r.URL.Query().Get("to_date")

	query := `SELECT id, project_id, report_date, shift, weather, temp_c, manpower_total, equipment_total, narrative, hse_notes, delays, work_completed, planned_tomorrow, author, status, created_at, updated_at FROM doc_daily_reports WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if projectID != "" { query += fmt.Sprintf(" AND project_id = $%d", argIdx); argIdx++; args = append(args, projectID) }
	if status != "" { query += fmt.Sprintf(" AND status = $%d", argIdx); argIdx++; args = append(args, status) }
	if fromDate != "" { query += fmt.Sprintf(" AND report_date >= $%d", argIdx); argIdx++; args = append(args, fromDate) }
	if toDate != "" { query += fmt.Sprintf(" AND report_date <= $%d", argIdx); argIdx++; args = append(args, toDate) }
	query += " ORDER BY report_date DESC, shift"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, projectID, shift, weather, narrative, hseNotes, delays, workCompleted, plannedTomorrow, author, status string
		var reportDate time.Time
		var tempC, manpowerTotal, equipmentTotal sql.NullFloat64
		var createdAt, updatedAt time.Time

		err := rows.Scan(&id, &projectID, &reportDate, &shift, &weather, &tempC, &manpowerTotal, &equipmentTotal, &narrative, &hseNotes, &delays, &workCompleted, &plannedTomorrow, &author, &status, &createdAt, &updatedAt)
		if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
		item := map[string]interface{}{
			"id": id, "project_id": projectID, "report_date": reportDate, "shift": shift,
			"weather": weather, "narrative": narrative, "hse_notes": hseNotes,
			"delays": delays, "work_completed": workCompleted, "planned_tomorrow": plannedTomorrow,
			"author": author, "status": status, "created_at": createdAt, "updated_at": updatedAt,
		}
		if tempC.Valid { item["temp_c"] = tempC.Float64 }
		if manpowerTotal.Valid { item["manpower_total"] = int(manpowerTotal.Float64) }
		if equipmentTotal.Valid { item["equipment_total"] = int(equipmentTotal.Float64) }
		items = append(items, item)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *DocControlHandler) CreateDailyReport(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID  string `json:"project_id"`
		ReportDate string `json:"report_date"`
		Shift      string `json:"shift"`
		Narrative  *string `json:"narrative"`
		Author     *string `json:"author"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO doc_daily_reports (id, project_id, report_date, shift, narrative, author, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		id, input.ProjectID, input.ReportDate, input.Shift, input.Narrative, input.Author, now, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *DocControlHandler) GetDailyReport(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var reportDate time.Time
	var shift, author string
	err := h.db.QueryRow(`SELECT report_date, shift, author FROM doc_daily_reports WHERE id = $1`, id).Scan(&reportDate, &shift, &author)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "daily report not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "report_date": reportDate, "shift": shift, "author": author})
}

func (h *DocControlHandler) UpdateDailyReport(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Status *string `json:"status"`
		Narrative *string `json:"narrative"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	now := time.Now()
	_, err := h.db.Exec(`UPDATE doc_daily_reports SET status=COALESCE($1,status), narrative=COALESCE($2,narrative), updated_at=$3 WHERE id=$4`,
		input.Status, input.Narrative, now, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *DocControlHandler) DeleteDailyReport(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM doc_daily_reports WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// Document Transmittals
// =============================================================================

func (h *DocControlHandler) ListTransmittals(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	projectID := r.URL.Query().Get("project_id")
	purpose := r.URL.Query().Get("purpose")

	query := `SELECT id, project_id, transmittal_number, transmittal_code, title, purpose, from_entity, to_entity, status, sent_at, received_at, acknowledged_at, created_at, updated_at FROM document_transmittals WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if status != "" { query += fmt.Sprintf(" AND status = $%d", argIdx); argIdx++; args = append(args, status) }
	if projectID != "" { query += fmt.Sprintf(" AND project_id = $%d", argIdx); argIdx++; args = append(args, projectID) }
	if purpose != "" { query += fmt.Sprintf(" AND purpose = $%d", argIdx); argIdx++; args = append(args, purpose) }
	query += " ORDER BY transmittal_number DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, projectID, dtCode, title, purpose, fromEntity, toEntity, status string
		var dtNumber int
		var sentAt, receivedAt, acknowledgedAt, createdAt, updatedAt time.Time

		err := rows.Scan(&id, &projectID, &dtNumber, &dtCode, &title, &purpose, &fromEntity, &toEntity, &status, &sentAt, &receivedAt, &acknowledgedAt, &createdAt, &updatedAt)
		if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
		items = append(items, map[string]interface{}{
			"id": id, "project_id": projectID, "transmittal_number": dtNumber, "transmittal_code": dtCode,
			"title": title, "purpose": purpose, "from_entity": fromEntity, "to_entity": toEntity,
			"status": status, "sent_at": sentAt, "received_at": receivedAt,
			"acknowledged_at": acknowledgedAt, "created_at": createdAt, "updated_at": updatedAt,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *DocControlHandler) CreateTransmittal(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID string `json:"project_id"`
		DTCode    string `json:"transmittal_code"`
		Title     string `json:"title"`
		Purpose   *string `json:"purpose"`
		FromEntity *string `json:"from_entity"`
		ToEntity   *string `json:"to_entity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO document_transmittals (id, project_id, transmittal_code, title, purpose, from_entity, to_entity, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		id, input.ProjectID, input.DTCode, input.Title, input.Purpose, input.FromEntity, input.ToEntity, now, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *DocControlHandler) GetTransmittal(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var title, dtCode string
	err := h.db.QueryRow(`SELECT transmittal_code, title FROM document_transmittals WHERE id = $1`, id).Scan(&dtCode, &title)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "transmittal not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"id": id, "transmittal_code": dtCode, "title": title})
}

func (h *DocControlHandler) UpdateTransmittal(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Status *string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	now := time.Now()
	_, err := h.db.Exec(`UPDATE document_transmittals SET status=COALESCE($1,status), updated_at=$2 WHERE id=$3`, input.Status, now, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *DocControlHandler) DeleteTransmittal(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM document_transmittals WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// Document Revisions
// =============================================================================

func (h *DocControlHandler) ListRevisions(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	docType := r.URL.Query().Get("document_type")
	docID := r.URL.Query().Get("document_id")

	query := `SELECT id, project_id, document_type, document_id, revision, change_summary, file_path, file_size, created_by, status, created_at FROM document_revisions WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if projectID != "" { query += fmt.Sprintf(" AND project_id = $%d", argIdx); argIdx++; args = append(args, projectID) }
	if docType != "" { query += fmt.Sprintf(" AND document_type = $%d", argIdx); argIdx++; args = append(args, docType) }
	if docID != "" { query += fmt.Sprintf(" AND document_id = $%d", argIdx); argIdx++; args = append(args, docID) }
	query += " ORDER BY created_at DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, projectID, docType, docID, revision, changeSummary, filePath, createdBy, status string
		var fileSize sql.NullInt64
		var createdAt time.Time

		err := rows.Scan(&id, &projectID, &docType, &docID, &revision, &changeSummary, &filePath, &fileSize, &createdBy, &status, &createdAt)
		if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
		item := map[string]interface{}{
			"id": id, "project_id": projectID, "document_type": docType, "document_id": docID,
			"revision": revision, "change_summary": changeSummary, "file_path": filePath,
			"created_by": createdBy, "status": status, "created_at": createdAt,
		}
		if fileSize.Valid { item["file_size"] = fileSize.Int64 }
		items = append(items, item)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *DocControlHandler) CreateRevision(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID    string  `json:"project_id"`
		DocumentType string  `json:"document_type"`
		DocumentID   string  `json:"document_id"`
		Revision     string  `json:"revision"`
		ChangeSummary *string `json:"change_summary"`
		CreatedBy    *string `json:"created_by"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO document_revisions (id, project_id, document_type, document_id, revision, change_summary, created_by, created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		id, input.ProjectID, input.DocumentType, input.DocumentID, input.Revision, input.ChangeSummary, input.CreatedBy, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *DocControlHandler) GetRevision(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var docType, revision, docID string
	err := h.db.QueryRow(`SELECT document_type, document_id, revision FROM document_revisions WHERE id = $1`, id).Scan(&docType, &docID, &revision)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "revision not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "document_type": docType, "document_id": docID, "revision": revision})
}

func (h *DocControlHandler) DeleteRevision(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM document_revisions WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// Summary
// =============================================================================

func (h *DocControlHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")

	query := `SELECT project_id, total_rfi, open_rfi, total_ncr, open_ncr, critical_ncr, total_submittals, pending_submittals, rejected_submittals, total_ms, pending_ms, total_sd, pending_sd, total_corr, active_corr, total_mom, recent_reports, total_transmittals, pending_transmittals FROM doc_control_summary`
	args := []interface{}{}
	argIdx := 1
	if projectID != "" {
		query += fmt.Sprintf(" WHERE project_id = $%d", argIdx)
		args = append(args, projectID)
	}

	rows, err := h.db.Query(query, args...)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var pid string
		var totalRfi, openRfi, totalNcr, openNcr, criticalNcr, totalSub, pendingSub, rejectedSub, totalMs, pendingMs, totalSd, pendingSd, totalCorr, activeCorr, totalMom, recentReports, totalDt, pendingDt int

		err := rows.Scan(&pid, &totalRfi, &openRfi, &totalNcr, &openNcr, &criticalNcr, &totalSub, &pendingSub, &rejectedSub, &totalMs, &pendingMs, &totalSd, &pendingSd, &totalCorr, &activeCorr, &totalMom, &recentReports, &totalDt, &pendingDt)
		if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
		items = append(items, map[string]interface{}{
			"project_id": pid, "total_rfi": totalRfi, "open_rfi": openRfi,
			"total_ncr": totalNcr, "open_ncr": openNcr, "critical_ncr": criticalNcr,
			"total_submittals": totalSub, "pending_submittals": pendingSub, "rejected_submittals": rejectedSub,
			"total_method_statements": totalMs, "pending_method_statements": pendingMs,
			"total_shop_drawings": totalSd, "pending_shop_drawings": pendingSd,
			"total_correspondence": totalCorr, "active_correspondence": activeCorr,
			"total_minutes_of_meeting": totalMom,
			"recent_daily_reports": recentReports,
			"total_transmittals": totalDt, "pending_transmittals": pendingDt,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

// =============================================================================
// Helpers (reused from other handlers via go-chi patterns — these are defined
// in boq.go and accessible because handlers is a single package)
// =============================================================================

func init() {
	log.Println("Document Control handler initialized")
}