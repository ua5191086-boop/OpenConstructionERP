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

// QualityHandler handles Quality Management module endpoints
type QualityHandler struct {
	db *sql.DB
}

func NewQualityHandler(db *sql.DB) *QualityHandler {
	return &QualityHandler{db: db}
}

func (h *QualityHandler) RegisterRoutes(r chi.Router) {
	r.Route("/quality", func(r chi.Router) {
		r.Get("/itps", h.ListITPs)
		r.Post("/itps", h.CreateITP)
		r.Get("/itps/{id}", h.GetITP)
		r.Put("/itps/{id}", h.UpdateITP)
		r.Delete("/itps/{id}", h.DeleteITP)

		r.Get("/inspections", h.ListInspections)
		r.Post("/inspections", h.CreateInspection)
		r.Get("/inspections/{id}", h.GetInspection)
		r.Put("/inspections/{id}", h.UpdateInspection)
		r.Delete("/inspections/{id}", h.DeleteInspection)

		r.Get("/test-results", h.ListTestResults)
		r.Post("/test-results", h.CreateTestResult)
		r.Get("/test-results/{id}", h.GetTestResult)
		r.Put("/test-results/{id}", h.UpdateTestResult)
		r.Delete("/test-results/{id}", h.DeleteTestResult)

		r.Get("/ncrs", h.ListNCRs)
		r.Post("/ncrs", h.CreateNCR)
		r.Get("/ncrs/{id}", h.GetNCR)
		r.Put("/ncrs/{id}", h.UpdateNCR)
		r.Delete("/ncrs/{id}", h.DeleteNCR)

		r.Get("/corrective-actions", h.ListCorrectiveActions)
		r.Post("/corrective-actions", h.CreateCorrectiveAction)
		r.Get("/corrective-actions/{id}", h.GetCorrectiveAction)
		r.Put("/corrective-actions/{id}", h.UpdateCorrectiveAction)
		r.Delete("/corrective-actions/{id}", h.DeleteCorrectiveAction)

		r.Get("/calibration", h.ListCalibration)
		r.Post("/calibration", h.CreateCalibration)
		r.Get("/calibration/{id}", h.GetCalibration)
		r.Put("/calibration/{id}", h.UpdateCalibration)
		r.Delete("/calibration/{id}", h.DeleteCalibration)

		r.Get("/quality-metrics", h.ListQualityMetrics)
		r.Post("/quality-metrics", h.CreateQualityMetrics)
		r.Get("/quality-metrics/{id}", h.GetQualityMetrics)
		r.Put("/quality-metrics/{id}", h.UpdateQualityMetrics)
		r.Delete("/quality-metrics/{id}", h.DeleteQualityMetrics)

		r.Get("/summary", h.GetSummary)
	})
}

// =============================================================================
// ITPs
// =============================================================================

func (h *QualityHandler) ListITPs(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	status := r.URL.Query().Get("status")

	query := `SELECT id, project_id, itp_number, itp_code, itp_name, itp_type, description, status, created_at, updated_at FROM qm_itp WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if projectID != "" { query += fmt.Sprintf(" AND project_id = $%d", argIdx); argIdx++; args = append(args, projectID) }
	if status != "" { query += fmt.Sprintf(" AND status = $%d", argIdx); argIdx++; args = append(args, status) }
	query += " ORDER BY itp_number DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, pid, code, name, itype, desc, status string
		var num int
		var createdAt, updatedAt time.Time
		if err := rows.Scan(&id, &pid, &num, &code, &name, &itype, &desc, &status, &createdAt, &updatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		items = append(items, map[string]interface{}{
			"id": id, "project_id": pid, "itp_number": num, "itp_code": code,
			"itp_name": name, "itp_type": itype, "description": desc,
			"status": status, "created_at": createdAt, "updated_at": updatedAt,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *QualityHandler) CreateITP(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID   string `json:"project_id"`
		ITPCode     string `json:"itp_code"`
		ITPName     string `json:"itp_name"`
		ITPType     string `json:"itp_type"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO qm_itp (id, project_id, itp_number, itp_code, itp_name, itp_type, description, status, created_at, updated_at) VALUES ($1,$2,(SELECT COALESCE(MAX(itp_number),0)+1 FROM qm_itp WHERE project_id=$2),$3,$4,$5,$6,'draft',$7,$8)`,
		id, input.ProjectID, input.ITPCode, input.ITPName, input.ITPType, input.Description, now, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *QualityHandler) GetITP(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var code, name string
	err := h.db.QueryRow(`SELECT itp_code, itp_name FROM qm_itp WHERE id = $1`, id).Scan(&code, &name)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "ITP not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "itp_code": code, "itp_name": name})
}

func (h *QualityHandler) UpdateITP(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Status *string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	_, err := h.db.Exec(`UPDATE qm_itp SET status=COALESCE($1,status), updated_at=$2 WHERE id=$3`, input.Status, time.Now(), id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *QualityHandler) DeleteITP(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM qm_itp WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// Inspections
// =============================================================================

func (h *QualityHandler) ListInspections(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	result := r.URL.Query().Get("result")

	query := `SELECT id, project_id, record_number, record_code, title, inspection_type, inspector, inspection_date, result, defects_found, status, created_at FROM qm_inspection_records WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if projectID != "" { query += fmt.Sprintf(" AND project_id = $%d", argIdx); argIdx++; args = append(args, projectID) }
	if result != "" { query += fmt.Sprintf(" AND result = $%d", argIdx); argIdx++; args = append(args, result) }
	query += " ORDER BY inspection_date DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, pid, code, title, itype, inspector, result, status string
		var num, defects int
		var inspDate, createdAt time.Time
		if err := rows.Scan(&id, &pid, &num, &code, &title, &itype, &inspector, &inspDate, &result, &defects, &status, &createdAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		items = append(items, map[string]interface{}{
			"id": id, "project_id": pid, "record_number": num, "record_code": code,
			"title": title, "inspection_type": itype, "inspector": inspector,
			"inspection_date": inspDate, "result": result, "defects_found": defects,
			"status": status, "created_at": createdAt,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *QualityHandler) CreateInspection(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID      string `json:"project_id"`
		RecordCode     string `json:"record_code"`
		Title          string `json:"title"`
		InspectionType string `json:"inspection_type"`
		Inspector      string `json:"inspector"`
		InspectionDate string `json:"inspection_date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO qm_inspection_records (id, project_id, record_number, record_code, title, inspection_type, inspector, inspection_date, created_at, updated_at) VALUES ($1,$2,(SELECT COALESCE(MAX(record_number),0)+1 FROM qm_inspection_records WHERE project_id=$2),$3,$4,$5,$6,$7,$8,$9)`,
		id, input.ProjectID, input.RecordCode, input.Title, input.InspectionType, input.Inspector, input.InspectionDate, now, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *QualityHandler) GetInspection(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var code, title string
	err := h.db.QueryRow(`SELECT record_code, title FROM qm_inspection_records WHERE id = $1`, id).Scan(&code, &title)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "inspection not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "record_code": code, "title": title})
}

func (h *QualityHandler) UpdateInspection(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Result *string `json:"result"`
		Status *string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	_, err := h.db.Exec(`UPDATE qm_inspection_records SET result=COALESCE($1,result), status=COALESCE($2,status), updated_at=$3 WHERE id=$4`, input.Result, input.Status, time.Now(), id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *QualityHandler) DeleteInspection(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM qm_inspection_records WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// Test Results
// =============================================================================

func (h *QualityHandler) ListTestResults(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	result := r.URL.Query().Get("result")

	query := `SELECT id, project_id, test_number, test_code, test_name, test_type, test_date, result, measured_value, min_acceptable, max_acceptable, lab_name, status FROM qm_test_results WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if projectID != "" { query += fmt.Sprintf(" AND project_id = $%d", argIdx); argIdx++; args = append(args, projectID) }
	if result != "" { query += fmt.Sprintf(" AND result = $%d", argIdx); argIdx++; args = append(args, result) }
	query += " ORDER BY test_date DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, pid, code, tname, ttype, result, lab, status string
		var num int
		var tdate time.Time
		var measured, minAcc, maxAcc float64
		if err := rows.Scan(&id, &pid, &num, &code, &tname, &ttype, &tdate, &result, &measured, &minAcc, &maxAcc, &lab, &status); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		items = append(items, map[string]interface{}{
			"id": id, "project_id": pid, "test_number": num, "test_code": code,
			"test_name": tname, "test_type": ttype, "test_date": tdate,
			"result": result, "measured_value": measured, "min_acceptable": minAcc,
			"max_acceptable": maxAcc, "lab_name": lab, "status": status,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *QualityHandler) CreateTestResult(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID string  `json:"project_id"`
		TestCode  string  `json:"test_code"`
		TestName  string  `json:"test_name"`
		TestType  string  `json:"test_type"`
		TestDate  string  `json:"test_date"`
		LabName   *string `json:"lab_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO qm_test_results (id, project_id, test_number, test_code, test_name, test_type, test_date, lab_name, created_at, updated_at) VALUES ($1,$2,(SELECT COALESCE(MAX(test_number),0)+1 FROM qm_test_results WHERE project_id=$2),$3,$4,$5,$6,$7,$8,$9)`,
		id, input.ProjectID, input.TestCode, input.TestName, input.TestType, input.TestDate, input.LabName, now, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *QualityHandler) GetTestResult(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var code, name string
	err := h.db.QueryRow(`SELECT test_code, test_name FROM qm_test_results WHERE id = $1`, id).Scan(&code, &name)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "test result not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "test_code": code, "test_name": name})
}

func (h *QualityHandler) UpdateTestResult(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Result *string `json:"result"`
		Status *string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	_, err := h.db.Exec(`UPDATE qm_test_results SET result=COALESCE($1,result), status=COALESCE($2,status), updated_at=$3 WHERE id=$4`, input.Result, input.Status, time.Now(), id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *QualityHandler) DeleteTestResult(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM qm_test_results WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// NCRs
// =============================================================================

func (h *QualityHandler) ListNCRs(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	status := r.URL.Query().Get("status")
	severity := r.URL.Query().Get("severity")

	query := `SELECT id, project_id, ncr_number, ncr_code, title, ncr_category, severity, source, description, discovered_date, root_cause, disposition_type, rework_cost, schedule_impact, status FROM qm_ncr WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if projectID != "" { query += fmt.Sprintf(" AND project_id = $%d", argIdx); argIdx++; args = append(args, projectID) }
	if status != "" { query += fmt.Sprintf(" AND status = $%d", argIdx); argIdx++; args = append(args, status) }
	if severity != "" { query += fmt.Sprintf(" AND severity = $%d", argIdx); argIdx++; args = append(args, severity) }
	query += " ORDER BY ncr_number DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, pid, code, title, cat, sev, src, desc, rc, disp, st string
		var num, schedImp int
		var discDate time.Time
		var rwCost float64
		if err := rows.Scan(&id, &pid, &num, &code, &title, &cat, &sev, &src, &desc, &discDate, &rc, &disp, &rwCost, &schedImp, &st); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		items = append(items, map[string]interface{}{
			"id": id, "project_id": pid, "ncr_number": num, "ncr_code": code,
			"title": title, "ncr_category": cat, "severity": sev, "source": src,
			"description": desc, "discovered_date": discDate, "root_cause": rc,
			"disposition_type": disp, "rework_cost": rwCost,
			"schedule_impact": schedImp, "status": st,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *QualityHandler) CreateNCR(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID     string  `json:"project_id"`
		NCRCode       string  `json:"ncr_code"`
		Title         string  `json:"title"`
		NCRCategory   string  `json:"ncr_category"`
		Severity      string  `json:"severity"`
		Description   string  `json:"description"`
		DiscoveredDate string `json:"discovered_date"`
		Source        string  `json:"source"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO qm_ncr (id, project_id, ncr_number, ncr_code, title, ncr_category, severity, description, discovered_date, source, created_at, updated_at) VALUES ($1,$2,(SELECT COALESCE(MAX(ncr_number),0)+1 FROM qm_ncr WHERE project_id=$2),$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		id, input.ProjectID, input.NCRCode, input.Title, input.NCRCategory, input.Severity, input.Description, input.DiscoveredDate, input.Source, now, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *QualityHandler) GetNCR(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var code, title string
	err := h.db.QueryRow(`SELECT ncr_code, title FROM qm_ncr WHERE id = $1`, id).Scan(&code, &title)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "NCR not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "ncr_code": code, "title": title})
}

func (h *QualityHandler) UpdateNCR(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Status            *string `json:"status"`
		RootCause         *string `json:"root_cause"`
		DispositionType   *string `json:"disposition_type"`
		ApprovedDisposition *string `json:"approved_disposition"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	_, err := h.db.Exec(`UPDATE qm_ncr SET status=COALESCE($1,status), root_cause=COALESCE($2,root_cause), disposition_type=COALESCE($3,disposition_type), approved_disposition=COALESCE($4,approved_disposition), updated_at=$5 WHERE id=$6`,
		input.Status, input.RootCause, input.DispositionType, input.ApprovedDisposition, time.Now(), id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *QualityHandler) DeleteNCR(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM qm_ncr WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// Corrective Actions
// =============================================================================

func (h *QualityHandler) ListCorrectiveActions(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	status := r.URL.Query().Get("status")

	query := `SELECT id, project_id, ca_number, ca_code, title, action_type, assigned_to, priority, due_date, effectiveness, status FROM qm_corrective_actions WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if projectID != "" { query += fmt.Sprintf(" AND project_id = $%d", argIdx); argIdx++; args = append(args, projectID) }
	if status != "" { query += fmt.Sprintf(" AND status = $%d", argIdx); argIdx++; args = append(args, status) }
	query += " ORDER BY ca_number DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, pid, code, title, atype, assignee, prio, eff, st string
		var num int
		var dueDate sql.NullString
		if err := rows.Scan(&id, &pid, &num, &code, &title, &atype, &assignee, &prio, &dueDate, &eff, &st); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		item := map[string]interface{}{
			"id": id, "project_id": pid, "ca_number": num, "ca_code": code,
			"title": title, "action_type": atype, "assigned_to": assignee,
			"priority": prio, "effectiveness": eff, "status": st,
		}
		if dueDate.Valid { item["due_date"] = dueDate.String }
		items = append(items, item)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *QualityHandler) CreateCorrectiveAction(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID  string `json:"project_id"`
		CACode     string `json:"ca_code"`
		Title      string `json:"title"`
		ActionType string `json:"action_type"`
		AssignedTo string `json:"assigned_to"`
		Priority   string `json:"priority"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO qm_corrective_actions (id, project_id, ca_number, ca_code, title, action_type, assigned_to, priority, created_at, updated_at) VALUES ($1,$2,(SELECT COALESCE(MAX(ca_number),0)+1 FROM qm_corrective_actions WHERE project_id=$2),$3,$4,$5,$6,$7,$8,$9)`,
		id, input.ProjectID, input.CACode, input.Title, input.ActionType, input.AssignedTo, input.Priority, now, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *QualityHandler) GetCorrectiveAction(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var code, title string
	err := h.db.QueryRow(`SELECT ca_code, title FROM qm_corrective_actions WHERE id = $1`, id).Scan(&code, &title)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "corrective action not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "ca_code": code, "title": title})
}

func (h *QualityHandler) UpdateCorrectiveAction(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Status       *string `json:"status"`
		Effectiveness *string `json:"effectiveness"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	_, err := h.db.Exec(`UPDATE qm_corrective_actions SET status=COALESCE($1,status), effectiveness=COALESCE($2,effectiveness), updated_at=$3 WHERE id=$4`, input.Status, input.Effectiveness, time.Now(), id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *QualityHandler) DeleteCorrectiveAction(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM qm_corrective_actions WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// Calibration
// =============================================================================

func (h *QualityHandler) ListCalibration(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	status := r.URL.Query().Get("status")

	query := `SELECT id, project_id, equipment_name, equipment_model, serial_number, calibration_type, last_calibration_date, next_calibration_date, calibration_result, status FROM qm_calibration WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if projectID != "" { query += fmt.Sprintf(" AND project_id = $%d", argIdx); argIdx++; args = append(args, projectID) }
	if status != "" { query += fmt.Sprintf(" AND status = $%d", argIdx); argIdx++; args = append(args, status) }
	query += " ORDER BY next_calibration_date ASC"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, pid, eqName, eqModel, sn, calType, result, st string
		var lastCal, nextCal sql.NullString
		if err := rows.Scan(&id, &pid, &eqName, &eqModel, &sn, &calType, &lastCal, &nextCal, &result, &st); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		item := map[string]interface{}{
			"id": id, "project_id": pid, "equipment_name": eqName, "equipment_model": eqModel,
			"serial_number": sn, "calibration_type": calType, "calibration_result": result, "status": st,
		}
		if lastCal.Valid { item["last_calibration_date"] = lastCal.String }
		if nextCal.Valid { item["next_calibration_date"] = nextCal.String }
		items = append(items, item)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *QualityHandler) CreateCalibration(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID      string  `json:"project_id"`
		EquipmentName  string  `json:"equipment_name"`
		SerialNumber   string  `json:"serial_number"`
		CalibrationType string `json:"calibration_type"`
		NextCalDate    string  `json:"next_calibration_date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO qm_calibration (id, project_id, equipment_name, serial_number, calibration_type, next_calibration_date, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		id, input.ProjectID, input.EquipmentName, input.SerialNumber, input.CalibrationType, input.NextCalDate, now, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *QualityHandler) GetCalibration(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var eqName, sn string
	err := h.db.QueryRow(`SELECT equipment_name, serial_number FROM qm_calibration WHERE id = $1`, id).Scan(&eqName, &sn)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "calibration record not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "equipment_name": eqName, "serial_number": sn})
}

func (h *QualityHandler) UpdateCalibration(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Result *string `json:"calibration_result"`
		Status *string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	_, err := h.db.Exec(`UPDATE qm_calibration SET calibration_result=COALESCE($1,calibration_result), status=COALESCE($2,status), updated_at=$3 WHERE id=$4`, input.Result, input.Status, time.Now(), id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *QualityHandler) DeleteCalibration(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM qm_calibration WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// Quality Metrics
// =============================================================================

func (h *QualityHandler) ListQualityMetrics(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")

	query := `SELECT id, project_id, report_month, total_inspections, inspections_passed, inspections_failed, total_tests, tests_passed, tests_failed, ncr_opened, ncr_closed, first_pass_yield, rework_cost FROM qm_quality_metrics WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if projectID != "" { query += fmt.Sprintf(" AND project_id = $%d", argIdx); argIdx++; args = append(args, projectID) }
	query += " ORDER BY report_month DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, pid string
		var rm time.Time
		var totalIns, passIns, failIns, totalTst, passTst, failTst, ncrOp, ncrCl int
		var fpy, rwCost float64
		if err := rows.Scan(&id, &pid, &rm, &totalIns, &passIns, &failIns, &totalTst, &passTst, &failTst, &ncrOp, &ncrCl, &fpy, &rwCost); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		items = append(items, map[string]interface{}{
			"id": id, "project_id": pid, "report_month": rm,
			"total_inspections": totalIns, "inspections_passed": passIns, "inspections_failed": failIns,
			"total_tests": totalTst, "tests_passed": passTst, "tests_failed": failTst,
			"ncr_opened": ncrOp, "ncr_closed": ncrCl,
			"first_pass_yield": fpy, "rework_cost": rwCost,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *QualityHandler) CreateQualityMetrics(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID string `json:"project_id"`
		ReportMonth string `json:"report_month"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO qm_quality_metrics (id, project_id, report_month, created_at, updated_at) VALUES ($1,$2,$3,$4,$5)`,
		id, input.ProjectID, input.ReportMonth, now, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *QualityHandler) GetQualityMetrics(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var rm time.Time
	err := h.db.QueryRow(`SELECT report_month FROM qm_quality_metrics WHERE id = $1`, id).Scan(&rm)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "quality metrics not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "report_month": rm})
}

func (h *QualityHandler) UpdateQualityMetrics(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		FirstPassYield *float64 `json:"first_pass_yield"`
		ReworkCost     *float64 `json:"rework_cost"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	_, err := h.db.Exec(`UPDATE qm_quality_metrics SET first_pass_yield=COALESCE($1,first_pass_yield), rework_cost=COALESCE($2,rework_cost), updated_at=$3 WHERE id=$4`, input.FirstPassYield, input.ReworkCost, time.Now(), id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *QualityHandler) DeleteQualityMetrics(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM qm_quality_metrics WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// Summary
// =============================================================================

func (h *QualityHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT project_id, active_itps, total_itps, failed_inspections, inspections_30d, total_inspections, failed_tests, open_ncrs, critical_ncrs, total_ncrs, open_cas, overdue_calibrations, total_rework_cost FROM qm_summary`
	args := []interface{}{}
	argIdx := 1
	if projectID != "" { query += fmt.Sprintf(" WHERE project_id = $%d", argIdx); args = append(args, projectID) }

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var pid string
		var activeITPs, totalITPs, failIns, ins30d, totalIns, failTst, openNCRs, critNCRs, totalNCRs, openCAs, overdueCal int
		var rwCost float64
		if err := rows.Scan(&pid, &activeITPs, &totalITPs, &failIns, &ins30d, &totalIns, &failTst, &openNCRs, &critNCRs, &totalNCRs, &openCAs, &overdueCal, &rwCost); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		items = append(items, map[string]interface{}{
			"project_id": pid, "active_itps": activeITPs, "total_itps": totalITPs,
			"failed_inspections": failIns, "inspections_30d": ins30d, "total_inspections": totalIns,
			"failed_tests": failTst, "open_ncrs": openNCRs, "critical_ncrs": critNCRs,
			"total_ncrs": totalNCRs, "open_cas": openCAs, "overdue_calibrations": overdueCal,
			"total_rework_cost": rwCost,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

// Ensure respondError and respondJSON are available (they should be in the same package)
func init() {
	// Silence unused import warning
	var _ = log.Printf
}