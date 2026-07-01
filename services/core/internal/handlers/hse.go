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

// HSEHandler handles Health, Safety & Environment module endpoints
type HSEHandler struct {
	db *sql.DB
}

func NewHSEHandler(db *sql.DB) *HSEHandler {
	return &HSEHandler{db: db}
}

func (h *HSEHandler) RegisterRoutes(r chi.Router) {
	r.Route("/hse", func(r chi.Router) {
		// Incidents
		r.Get("/incidents", h.ListIncidents)
		r.Post("/incidents", h.CreateIncident)
		r.Get("/incidents/{id}", h.GetIncident)
		r.Put("/incidents/{id}", h.UpdateIncident)
		r.Delete("/incidents/{id}", h.DeleteIncident)

		// Permits
		r.Get("/permits", h.ListPermits)
		r.Post("/permits", h.CreatePermit)
		r.Get("/permits/{id}", h.GetPermit)
		r.Put("/permits/{id}", h.UpdatePermit)
		r.Delete("/permits/{id}", h.DeletePermit)

		// Audits
		r.Get("/audits", h.ListAudits)
		r.Post("/audits", h.CreateAudit)
		r.Get("/audits/{id}", h.GetAudit)
		r.Put("/audits/{id}", h.UpdateAudit)
		r.Delete("/audits/{id}", h.DeleteAudit)

		// Inspections
		r.Get("/inspections", h.ListInspections)
		r.Post("/inspections", h.CreateInspection)
		r.Get("/inspections/{id}", h.GetInspection)
		r.Put("/inspections/{id}", h.UpdateInspection)
		r.Delete("/inspections/{id}", h.DeleteInspection)

		// Training
		r.Get("/training", h.ListTraining)
		r.Post("/training", h.CreateTraining)
		r.Get("/training/{id}", h.GetTraining)
		r.Put("/training/{id}", h.UpdateTraining)
		r.Delete("/training/{id}", h.DeleteTraining)

		// PPE
		r.Get("/ppe", h.ListPPE)
		r.Post("/ppe", h.CreatePPE)
		r.Get("/ppe/{id}", h.GetPPE)
		r.Put("/ppe/{id}", h.UpdatePPE)
		r.Delete("/ppe/{id}", h.DeletePPE)

		// Drills
		r.Get("/drills", h.ListDrills)
		r.Post("/drills", h.CreateDrill)
		r.Get("/drills/{id}", h.GetDrill)
		r.Put("/drills/{id}", h.UpdateDrill)
		r.Delete("/drills/{id}", h.DeleteDrill)

		// Statistics
		r.Get("/statistics", h.ListStatistics)
		r.Post("/statistics", h.CreateStatistics)
		r.Get("/statistics/{id}", h.GetStatistics)
		r.Put("/statistics/{id}", h.UpdateStatistics)
		r.Delete("/statistics/{id}", h.DeleteStatistics)

		// Emergency Plans
		r.Get("/emergency-plans", h.ListEmergencyPlans)
		r.Post("/emergency-plans", h.CreateEmergencyPlan)
		r.Get("/emergency-plans/{id}", h.GetEmergencyPlan)
		r.Put("/emergency-plans/{id}", h.UpdateEmergencyPlan)
		r.Delete("/emergency-plans/{id}", h.DeleteEmergencyPlan)

		// Chemicals
		r.Get("/chemicals", h.ListChemicals)
		r.Post("/chemicals", h.CreateChemical)
		r.Get("/chemicals/{id}", h.GetChemical)
		r.Put("/chemicals/{id}", h.UpdateChemical)
		r.Delete("/chemicals/{id}", h.DeleteChemical)

		// Summary
		r.Get("/summary", h.GetSummary)
	})
}

// =============================================================================
// Incidents
// =============================================================================

func (h *HSEHandler) ListIncidents(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	status := r.URL.Query().Get("status")
	severity := r.URL.Query().Get("severity")

	query := `SELECT id, project_id, incident_number, incident_code, title, incident_type, severity, incident_date, incident_time, location, area, reported_by, reported_at, lost_days, medical_cost, property_cost, total_cost, investigation_status, is_reportable, status, closed_at, created_at, updated_at FROM hse_incidents WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if projectID != "" { query += fmt.Sprintf(" AND project_id = $%d", argIdx); argIdx++; args = append(args, projectID) }
	if status != "" { query += fmt.Sprintf(" AND status = $%d", argIdx); argIdx++; args = append(args, status) }
	if severity != "" { query += fmt.Sprintf(" AND severity = $%d", argIdx); argIdx++; args = append(args, severity) }
	query += " ORDER BY incident_date DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, pid, incCode, title, itype, severity, loc, area, repBy, invStatus, status string
		var incNum, lostDays int
		var medCost, propCost, totalCost float64
		var incDate, reportedAt, createdAt, updatedAt time.Time
		var incTime, closedAt sql.NullString
		var isReportable bool
		err := rows.Scan(&id, &pid, &incNum, &incCode, &title, &itype, &severity, &incDate, &incTime, &loc, &area, &repBy, &reportedAt, &lostDays, &medCost, &propCost, &totalCost, &invStatus, &isReportable, &status, &closedAt, &createdAt, &updatedAt)
		if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
		item := map[string]interface{}{
			"id": id, "project_id": pid, "incident_number": incNum, "incident_code": incCode,
			"title": title, "incident_type": itype, "severity": severity, "incident_date": incDate,
			"location": loc, "area": area, "reported_by": repBy, "reported_at": reportedAt,
			"lost_days": lostDays, "medical_cost": medCost, "property_cost": propCost,
			"total_cost": totalCost, "investigation_status": invStatus, "is_reportable": isReportable,
			"status": status, "created_at": createdAt, "updated_at": updatedAt,
		}
		if incTime.Valid { item["incident_time"] = incTime.String }
		if closedAt.Valid { item["closed_at"] = closedAt.String }
		items = append(items, item)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *HSEHandler) CreateIncident(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID    string  `json:"project_id"`
		IncidentCode string  `json:"incident_code"`
		Title        string  `json:"title"`
		IncidentType string  `json:"incident_type"`
		Severity     *string `json:"severity"`
		IncidentDate string  `json:"incident_date"`
		Location     *string `json:"location"`
		ReportedBy   *string `json:"reported_by"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO hse_incidents (id, project_id, incident_number, incident_code, title, incident_type, severity, incident_date, location, reported_by, reported_at, created_at, updated_at) VALUES ($1,$2,(SELECT COALESCE(MAX(incident_number),0)+1 FROM hse_incidents WHERE project_id=$2),$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		id, input.ProjectID, input.IncidentCode, input.Title, input.IncidentType, input.Severity, input.IncidentDate, input.Location, input.ReportedBy, now, now, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *HSEHandler) GetIncident(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var incCode, title string
	err := h.db.QueryRow(`SELECT incident_code, title FROM hse_incidents WHERE id = $1`, id).Scan(&incCode, &title)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "incident not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "incident_code": incCode, "title": title})
}

func (h *HSEHandler) UpdateIncident(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Status             *string `json:"status"`
		InvestigationStatus *string `json:"investigation_status"`
		RootCause           *string `json:"root_cause"`
		CorrectiveAction    *string `json:"corrective_action"`
		PreventiveAction    *string `json:"preventive_action"`
		InvestigationLead   *string `json:"investigation_lead"`
		ClosedAt            *string `json:"closed_at"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	now := time.Now()
	_, err := h.db.Exec(`UPDATE hse_incidents SET status=COALESCE($1,status), investigation_status=COALESCE($2,investigation_status), root_cause=COALESCE($3,root_cause), corrective_action=COALESCE($4,corrective_action), preventive_action=COALESCE($5,preventive_action), investigation_lead=COALESCE($6,investigation_lead), closed_at=COALESCE($7,closed_at), updated_at=$8 WHERE id=$9`,
		input.Status, input.InvestigationStatus, input.RootCause, input.CorrectiveAction, input.PreventiveAction, input.InvestigationLead, input.ClosedAt, now, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *HSEHandler) DeleteIncident(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM hse_incidents WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// Permits
// =============================================================================

func (h *HSEHandler) ListPermits(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	status := r.URL.Query().Get("status")
	permitType := r.URL.Query().Get("permit_type")

	query := `SELECT id, project_id, permit_number, permit_code, permit_type, title, location, work_description, issuing_authority, permit_holder, responsible_person, control_measures, ppe_required, valid_from, valid_to, status, issued_at, issued_by, closed_at, created_at, updated_at FROM hse_permits WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if projectID != "" { query += fmt.Sprintf(" AND project_id = $%d", argIdx); argIdx++; args = append(args, projectID) }
	if status != "" { query += fmt.Sprintf(" AND status = $%d", argIdx); argIdx++; args = append(args, status) }
	if permitType != "" { query += fmt.Sprintf(" AND permit_type = $%d", argIdx); argIdx++; args = append(args, permitType) }
	query += " ORDER BY valid_from DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, pid, pcode, ptype, title, loc, workDesc, issAuth, permitHolder, respPerson, ctrlMeasures, ppe, status, issuedBy string
		var pnum int
		var validFrom, validTo, createdAt, updatedAt time.Time
		var issuedAt, closedAt sql.NullString
		err := rows.Scan(&id, &pid, &pnum, &pcode, &ptype, &title, &loc, &workDesc, &issAuth, &permitHolder, &respPerson, &ctrlMeasures, &ppe, &validFrom, &validTo, &status, &issuedAt, &issuedBy, &closedAt, &createdAt, &updatedAt)
		if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
		item := map[string]interface{}{
			"id": id, "project_id": pid, "permit_number": pnum, "permit_code": pcode,
			"permit_type": ptype, "title": title, "location": loc, "work_description": workDesc,
			"issuing_authority": issAuth, "permit_holder": permitHolder, "responsible_person": respPerson,
			"control_measures": ctrlMeasures, "ppe_required": ppe, "status": status,
			"valid_from": validFrom, "valid_to": validTo, "issued_by": issuedBy,
			"created_at": createdAt, "updated_at": updatedAt,
		}
		if issuedAt.Valid { item["issued_at"] = issuedAt.String }
		if closedAt.Valid { item["closed_at"] = closedAt.String }
		items = append(items, item)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *HSEHandler) CreatePermit(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID      string  `json:"project_id"`
		PermitCode     string  `json:"permit_code"`
		PermitType     string  `json:"permit_type"`
		Title          string  `json:"title"`
		WorkDescription string `json:"work_description"`
		Location       *string `json:"location"`
		ValidFrom      string  `json:"valid_from"`
		ValidTo        string  `json:"valid_to"`
		PermitHolder   *string `json:"permit_holder"`
		IssuingAuthority *string `json:"issuing_authority"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO hse_permits (id, project_id, permit_number, permit_code, permit_type, title, work_description, location, valid_from, valid_to, permit_holder, issuing_authority, created_at, updated_at) VALUES ($1,$2,(SELECT COALESCE(MAX(permit_number),0)+1 FROM hse_permits WHERE project_id=$2),$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		id, input.ProjectID, input.PermitCode, input.PermitType, input.Title, input.WorkDescription, input.Location, input.ValidFrom, input.ValidTo, input.PermitHolder, input.IssuingAuthority, now, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *HSEHandler) GetPermit(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var pcode, ptype string
	err := h.db.QueryRow(`SELECT permit_code, permit_type FROM hse_permits WHERE id = $1`, id).Scan(&pcode, &ptype)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "permit not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "permit_code": pcode, "permit_type": ptype})
}

func (h *HSEHandler) UpdatePermit(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Status       *string `json:"status"`
		ValidTo      *string `json:"valid_to"`
		ExtendedTo   *string `json:"extended_to"`
		Remarks      *string `json:"remarks"`
		ClosedAt     *string `json:"closed_at"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	now := time.Now()
	_, err := h.db.Exec(`UPDATE hse_permits SET status=COALESCE($1,status), valid_to=COALESCE($2,valid_to), extended_to=COALESCE($3,extended_to), remarks=COALESCE($4,remarks), closed_at=COALESCE($5,closed_at), updated_at=$6 WHERE id=$7`,
		input.Status, input.ValidTo, input.ExtendedTo, input.Remarks, input.ClosedAt, now, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *HSEHandler) DeletePermit(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM hse_permits WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// Audits
// =============================================================================

func (h *HSEHandler) ListAudits(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	status := r.URL.Query().Get("status")

	query := `SELECT id, project_id, audit_number, audit_code, audit_type, title, scope, lead_auditor, audit_date, location, non_conformities, observations, score_pct, status, findings_summary, follow_up_date, completed_at, created_at, updated_at FROM hse_audits WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if projectID != "" { query += fmt.Sprintf(" AND project_id = $%d", argIdx); argIdx++; args = append(args, projectID) }
	if status != "" { query += fmt.Sprintf(" AND status = $%d", argIdx); argIdx++; args = append(args, status) }
	query += " ORDER BY audit_date DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, pid, acode, atype, title, scope, leadAuditor, loc, status, findings string
		var anum, nc, obs int
		var score sql.NullFloat64
		var auditDate, completedAt, createdAt, updatedAt time.Time
		var followUpDate sql.NullString
		err := rows.Scan(&id, &pid, &anum, &acode, &atype, &title, &scope, &leadAuditor, &auditDate, &loc, &nc, &obs, &score, &status, &findings, &followUpDate, &completedAt, &createdAt, &updatedAt)
		if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
		item := map[string]interface{}{
			"id": id, "project_id": pid, "audit_number": anum, "audit_code": acode,
			"audit_type": atype, "title": title, "scope": scope, "lead_auditor": leadAuditor,
			"audit_date": auditDate, "location": loc, "non_conformities": nc, "observations": obs,
			"status": status, "findings_summary": findings, "completed_at": completedAt,
			"created_at": createdAt, "updated_at": updatedAt,
		}
		if score.Valid { item["score_pct"] = score.Float64 }
		if followUpDate.Valid { item["follow_up_date"] = followUpDate.String }
		items = append(items, item)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *HSEHandler) CreateAudit(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID    string  `json:"project_id"`
		AuditCode    string  `json:"audit_code"`
		AuditType    string  `json:"audit_type"`
		Title        string  `json:"title"`
		Scope        string  `json:"scope"`
		LeadAuditor  *string `json:"lead_auditor"`
		AuditDate    string  `json:"audit_date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO hse_audits (id, project_id, audit_number, audit_code, audit_type, title, scope, lead_auditor, audit_date, created_at, updated_at) VALUES ($1,$2,(SELECT COALESCE(MAX(audit_number),0)+1 FROM hse_audits WHERE project_id=$2),$3,$4,$5,$6,$7,$8,$9,$10)`,
		id, input.ProjectID, input.AuditCode, input.AuditType, input.Title, input.Scope, input.LeadAuditor, input.AuditDate, now, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *HSEHandler) GetAudit(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var acode, atype string
	err := h.db.QueryRow(`SELECT audit_code, audit_type FROM hse_audits WHERE id = $1`, id).Scan(&acode, &atype)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "audit not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "audit_code": acode, "audit_type": atype})
}

func (h *HSEHandler) UpdateAudit(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Status          *string  `json:"status"`
		ScorePct        *float64 `json:"score_pct"`
		FindingsSummary *string  `json:"findings_summary"`
		CompletedAt     *string  `json:"completed_at"`
		FollowUpDate    *string  `json:"follow_up_date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	now := time.Now()
	_, err := h.db.Exec(`UPDATE hse_audits SET status=COALESCE($1,status), score_pct=COALESCE($2,score_pct), findings_summary=COALESCE($3,findings_summary), completed_at=COALESCE($4,completed_at), follow_up_date=COALESCE($5,follow_up_date), updated_at=$6 WHERE id=$7`,
		input.Status, input.ScorePct, input.FindingsSummary, input.CompletedAt, input.FollowUpDate, now, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *HSEHandler) DeleteAudit(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM hse_audits WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// Inspections
// =============================================================================

func (h *HSEHandler) ListInspections(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	status := r.URL.Query().Get("status")

	query := `SELECT id, project_id, inspection_number, inspection_code, inspection_type, title, location, inspector, inspection_date, violations_found, violations_resolved, severity, status, action_items, follow_up_date, closed_at, created_at, updated_at FROM hse_inspections WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if projectID != "" { query += fmt.Sprintf(" AND project_id = $%d", argIdx); argIdx++; args = append(args, projectID) }
	if status != "" { query += fmt.Sprintf(" AND status = $%d", argIdx); argIdx++; args = append(args, status) }
	query += " ORDER BY inspection_date DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, pid, icode, itype, title, loc, inspector, severity, status, actionItems string
		var inum, vf, vr int
		var inspDate, createdAt, updatedAt time.Time
		var followUpDate, closedAt sql.NullString
		err := rows.Scan(&id, &pid, &inum, &icode, &itype, &title, &loc, &inspector, &inspDate, &vf, &vr, &severity, &status, &actionItems, &followUpDate, &closedAt, &createdAt, &updatedAt)
		if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
		item := map[string]interface{}{
			"id": id, "project_id": pid, "inspection_number": inum, "inspection_code": icode,
			"inspection_type": itype, "title": title, "location": loc, "inspector": inspector,
			"inspection_date": inspDate, "violations_found": vf, "violations_resolved": vr,
			"severity": severity, "status": status, "action_items": actionItems,
			"created_at": createdAt, "updated_at": updatedAt,
		}
		if followUpDate.Valid { item["follow_up_date"] = followUpDate.String }
		if closedAt.Valid { item["closed_at"] = closedAt.String }
		items = append(items, item)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *HSEHandler) CreateInspection(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID      string  `json:"project_id"`
		InspectionCode string  `json:"inspection_code"`
		InspectionType string  `json:"inspection_type"`
		Title          string  `json:"title"`
		Inspector      *string `json:"inspector"`
		InspectionDate string  `json:"inspection_date"`
		Location       *string `json:"location"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO hse_inspections (id, project_id, inspection_number, inspection_code, inspection_type, title, inspector, inspection_date, location, created_at, updated_at) VALUES ($1,$2,(SELECT COALESCE(MAX(inspection_number),0)+1 FROM hse_inspections WHERE project_id=$2),$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		id, input.ProjectID, input.InspectionCode, input.InspectionType, input.Title, input.Inspector, input.InspectionDate, input.Location, now, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *HSEHandler) GetInspection(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var icode, itype string
	err := h.db.QueryRow(`SELECT inspection_code, inspection_type FROM hse_inspections WHERE id = $1`, id).Scan(&icode, &itype)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "inspection not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "inspection_code": icode, "inspection_type": itype})
}

func (h *HSEHandler) UpdateInspection(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Status            *string `json:"status"`
		ViolationsFound   *int    `json:"violations_found"`
		ViolationsResolved *int   `json:"violations_resolved"`
		ActionItems       *string `json:"action_items"`
		ClosedAt          *string `json:"closed_at"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	now := time.Now()
	_, err := h.db.Exec(`UPDATE hse_inspections SET status=COALESCE($1,status), violations_found=COALESCE($2,violations_found), violations_resolved=COALESCE($3,violations_resolved), action_items=COALESCE($4,action_items), closed_at=COALESCE($5,closed_at), updated_at=$6 WHERE id=$7`,
		input.Status, input.ViolationsFound, input.ViolationsResolved, input.ActionItems, input.ClosedAt, now, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *HSEHandler) DeleteInspection(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM hse_inspections WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// Training
// =============================================================================

func (h *HSEHandler) ListTraining(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	status := r.URL.Query().Get("status")

	query := `SELECT id, project_id, training_number, training_code, training_name, training_type, trainer, training_date, duration_hours, location, attendees, max_attendees, status, certificate_type, certificate_validity_days, total_cost, notes, created_at, updated_at FROM hse_training WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if projectID != "" { query += fmt.Sprintf(" AND project_id = $%d", argIdx); argIdx++; args = append(args, projectID) }
	if status != "" { query += fmt.Sprintf(" AND status = $%d", argIdx); argIdx++; args = append(args, status) }
	query += " ORDER BY training_date DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, pid, tcode, tname, ttype, trainer, loc, status, certType, notes string
		var tnum, attendees, maxAtt, certVal int
		var durHours, totalCost float64
		var trainingDate, createdAt, updatedAt time.Time
		var err2 error
		err2 = rows.Scan(&id, &pid, &tnum, &tcode, &tname, &ttype, &trainer, &trainingDate, &durHours, &loc, &attendees, &maxAtt, &status, &certType, &certVal, &totalCost, &notes, &createdAt, &updatedAt)
		if err2 != nil { respondError(w, http.StatusInternalServerError, err2.Error()); return }
		items = append(items, map[string]interface{}{
			"id": id, "project_id": pid, "training_number": tnum, "training_code": tcode,
			"training_name": tname, "training_type": ttype, "trainer": trainer,
			"training_date": trainingDate, "duration_hours": durHours, "location": loc,
			"attendees": attendees, "max_attendees": maxAtt, "status": status,
			"certificate_type": certType, "certificate_validity_days": certVal,
			"total_cost": totalCost, "notes": notes, "created_at": createdAt, "updated_at": updatedAt,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *HSEHandler) CreateTraining(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID    string  `json:"project_id"`
		TrainingCode string  `json:"training_code"`
		TrainingName string  `json:"training_name"`
		TrainingType string  `json:"training_type"`
		Trainer      *string `json:"trainer"`
		TrainingDate string  `json:"training_date"`
		DurationHours *float64 `json:"duration_hours"`
		Attendees    *int    `json:"attendees"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO hse_training (id, project_id, training_number, training_code, training_name, training_type, trainer, training_date, duration_hours, attendees, created_at, updated_at) VALUES ($1,$2,(SELECT COALESCE(MAX(training_number),0)+1 FROM hse_training WHERE project_id=$2),$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		id, input.ProjectID, input.TrainingCode, input.TrainingName, input.TrainingType, input.Trainer, input.TrainingDate, input.DurationHours, input.Attendees, now, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *HSEHandler) GetTraining(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var tcode, tname string
	err := h.db.QueryRow(`SELECT training_code, training_name FROM hse_training WHERE id = $1`, id).Scan(&tcode, &tname)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "training record not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "training_code": tcode, "training_name": tname})
}

func (h *HSEHandler) UpdateTraining(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Status       *string `json:"status"`
		Attendees    *int    `json:"attendees"`
		Notes        *string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	now := time.Now()
	_, err := h.db.Exec(`UPDATE hse_training SET status=COALESCE($1,status), attendees=COALESCE($2,attendees), notes=COALESCE($3,notes), updated_at=$4 WHERE id=$5`,
		input.Status, input.Attendees, input.Notes, now, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *HSEHandler) DeleteTraining(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM hse_training WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// PPE
// =============================================================================

func (h *HSEHandler) ListPPE(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	category := r.URL.Query().Get("category")

	query := `SELECT id, project_id, ppe_code, ppe_name, ppe_category, manufacturer, model, size, quantity_issued, quantity_stock, reorder_level, unit_cost, expiry_date, storage_location, is_active, created_at, updated_at FROM hse_ppe WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if projectID != "" { query += fmt.Sprintf(" AND project_id = $%d", argIdx); argIdx++; args = append(args, projectID) }
	if category != "" { query += fmt.Sprintf(" AND ppe_category = $%d", argIdx); argIdx++; args = append(args, category) }
	query += " ORDER BY ppe_category, ppe_code"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, pid, pcode, pname, pcat, manuf, model, size, loc string
		var qtyIssued, qtyStock, reorder int
		var unitCost float64
		var expiryDate, createdAt, updatedAt time.Time
		var isActive bool
		err := rows.Scan(&id, &pid, &pcode, &pname, &pcat, &manuf, &model, &size, &qtyIssued, &qtyStock, &reorder, &unitCost, &expiryDate, &loc, &isActive, &createdAt, &updatedAt)
		if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
		items = append(items, map[string]interface{}{
			"id": id, "project_id": pid, "ppe_code": pcode, "ppe_name": pname,
			"ppe_category": pcat, "manufacturer": manuf, "model": model, "size": size,
			"quantity_issued": qtyIssued, "quantity_stock": qtyStock, "reorder_level": reorder,
			"unit_cost": unitCost, "expiry_date": expiryDate, "storage_location": loc,
			"is_active": isActive, "created_at": createdAt, "updated_at": updatedAt,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *HSEHandler) CreatePPE(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID  string  `json:"project_id"`
		PPECode    string  `json:"ppe_code"`
		PPEName    string  `json:"ppe_name"`
		PPECategory string `json:"ppe_category"`
		Manufacturer *string `json:"manufacturer"`
		UnitCost   *float64 `json:"unit_cost"`
		QuantityStock *int `json:"quantity_stock"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO hse_ppe (id, project_id, ppe_code, ppe_name, ppe_category, manufacturer, unit_cost, quantity_stock, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		id, input.ProjectID, input.PPECode, input.PPEName, input.PPECategory, input.Manufacturer, input.UnitCost, input.QuantityStock, now, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *HSEHandler) GetPPE(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var pcode, pname string
	err := h.db.QueryRow(`SELECT ppe_code, ppe_name FROM hse_ppe WHERE id = $1`, id).Scan(&pcode, &pname)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "PPE item not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "ppe_code": pcode, "ppe_name": pname})
}

func (h *HSEHandler) UpdatePPE(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		QuantityStock *int `json:"quantity_stock"`
		ReorderLevel  *int `json:"reorder_level"`
		IsActive      *bool `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	now := time.Now()
	_, err := h.db.Exec(`UPDATE hse_ppe SET quantity_stock=COALESCE($1,quantity_stock), reorder_level=COALESCE($2,reorder_level), is_active=COALESCE($3,is_active), updated_at=$4 WHERE id=$5`,
		input.QuantityStock, input.ReorderLevel, input.IsActive, now, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *HSEHandler) DeletePPE(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM hse_ppe WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// Drills
// =============================================================================

func (h *HSEHandler) ListDrills(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")

	query := `SELECT id, project_id, drill_number, drill_code, drill_name, drill_type, location, drill_date, participants, duration_minutes, evaluator, score_pct, observations, improvements, status, created_at, updated_at FROM hse_drill WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if projectID != "" { query += fmt.Sprintf(" AND project_id = $%d", argIdx); argIdx++; args = append(args, projectID) }
	query += " ORDER BY drill_date DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, pid, dcode, dname, dtype, loc, evaluator, obs, improvements, status string
		var dnum, participants, durMin int
		var score sql.NullFloat64
		var drillDate, createdAt, updatedAt time.Time
		err := rows.Scan(&id, &pid, &dnum, &dcode, &dname, &dtype, &loc, &drillDate, &participants, &durMin, &evaluator, &score, &obs, &improvements, &status, &createdAt, &updatedAt)
		if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
		item := map[string]interface{}{
			"id": id, "project_id": pid, "drill_number": dnum, "drill_code": dcode,
			"drill_name": dname, "drill_type": dtype, "location": loc, "drill_date": drillDate,
			"participants": participants, "duration_minutes": durMin, "evaluator": evaluator,
			"observations": obs, "improvements": improvements, "status": status,
			"created_at": createdAt, "updated_at": updatedAt,
		}
		if score.Valid { item["score_pct"] = score.Float64 }
		items = append(items, item)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *HSEHandler) CreateDrill(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID    string  `json:"project_id"`
		DrillCode    string  `json:"drill_code"`
		DrillName    string  `json:"drill_name"`
		DrillType    string  `json:"drill_type"`
		DrillDate    string  `json:"drill_date"`
		Participants *int    `json:"participants"`
		Evaluator    *string `json:"evaluator"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO hse_drill (id, project_id, drill_number, drill_code, drill_name, drill_type, drill_date, participants, evaluator, created_at, updated_at) VALUES ($1,$2,(SELECT COALESCE(MAX(drill_number),0)+1 FROM hse_drill WHERE project_id=$2),$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		id, input.ProjectID, input.DrillCode, input.DrillName, input.DrillType, input.DrillDate, input.Participants, input.Evaluator, now, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *HSEHandler) GetDrill(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var dcode, dname string
	err := h.db.QueryRow(`SELECT drill_code, drill_name FROM hse_drill WHERE id = $1`, id).Scan(&dcode, &dname)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "drill not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "drill_code": dcode, "drill_name": dname})
}

func (h *HSEHandler) UpdateDrill(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Status       *string  `json:"status"`
		ScorePct     *float64 `json:"score_pct"`
		Observations *string  `json:"observations"`
		Improvements *string  `json:"improvements"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	now := time.Now()
	_, err := h.db.Exec(`UPDATE hse_drill SET status=COALESCE($1,status), score_pct=COALESCE($2,score_pct), observations=COALESCE($3,observations), improvements=COALESCE($4,improvements), updated_at=$5 WHERE id=$6`,
		input.Status, input.ScorePct, input.Observations, input.Improvements, now, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *HSEHandler) DeleteDrill(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM hse_drill WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// Statistics
// =============================================================================

func (h *HSEHandler) ListStatistics(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")

	query := `SELECT id, project_id, report_month, manhours, lost_time_injuries, recordable_injuries, fatalities, near_misses, first_aid_cases, property_damage, environmental_incidents, fire_incidents, lti_frequency, total_recordable_rate, days_since_last_lti, safety_training_hours, inspections_conducted, audits_conducted, permits_issued, notes, created_at, updated_at FROM hse_statistics WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if projectID != "" { query += fmt.Sprintf(" AND project_id = $%d", argIdx); argIdx++; args = append(args, projectID) }
	query += " ORDER BY report_month DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, pid, notes string
		var reportMonth time.Time
		var manhours, lti, ri, fatal, nm, fac, pd, env, fire, ltiFreq, trir, dsl, trainingHrs, inspCount, audCount, permCount int
		var createdAt, updatedAt time.Time
		err := rows.Scan(&id, &pid, &reportMonth, &manhours, &lti, &ri, &fatal, &nm, &fac, &pd, &env, &fire, &ltiFreq, &trir, &dsl, &trainingHrs, &inspCount, &audCount, &permCount, &notes, &createdAt, &updatedAt)
		if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
		items = append(items, map[string]interface{}{
			"id": id, "project_id": pid, "report_month": reportMonth,
			"manhours": manhours, "lost_time_injuries": lti, "recordable_injuries": ri,
			"fatalities": fatal, "near_misses": nm, "first_aid_cases": fac,
			"property_damage": pd, "environmental_incidents": env, "fire_incidents": fire,
			"lti_frequency": ltiFreq, "total_recordable_rate": trir,
			"days_since_last_lti": dsl, "safety_training_hours": trainingHrs,
			"inspections_conducted": inspCount, "audits_conducted": audCount,
			"permits_issued": permCount, "notes": notes,
			"created_at": createdAt, "updated_at": updatedAt,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *HSEHandler) CreateStatistics(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID  string `json:"project_id"`
		ReportMonth string `json:"report_month"`
		Manhours   *int   `json:"manhours"`
		LostTimeInjuries *int `json:"lost_time_injuries"`
		NearMisses *int   `json:"near_misses"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO hse_statistics (id, project_id, report_month, manhours, lost_time_injuries, near_misses, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		id, input.ProjectID, input.ReportMonth, input.Manhours, input.LostTimeInjuries, input.NearMisses, now, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *HSEHandler) GetStatistics(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var reportMonth time.Time
	err := h.db.QueryRow(`SELECT report_month FROM hse_statistics WHERE id = $1`, id).Scan(&reportMonth)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "statistics not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "report_month": reportMonth})
}

func (h *HSEHandler) UpdateStatistics(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Manhours            *int `json:"manhours"`
		LostTimeInjuries    *int `json:"lost_time_injuries"`
		NearMisses          *int `json:"near_misses"`
		SafetyTrainingHours *int `json:"safety_training_hours"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	now := time.Now()
	_, err := h.db.Exec(`UPDATE hse_statistics SET manhours=COALESCE($1,manhours), lost_time_injuries=COALESCE($2,lost_time_injuries), near_misses=COALESCE($3,near_misses), safety_training_hours=COALESCE($4,safety_training_hours), updated_at=$5 WHERE id=$6`,
		input.Manhours, input.LostTimeInjuries, input.NearMisses, input.SafetyTrainingHours, now, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *HSEHandler) DeleteStatistics(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM hse_statistics WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// Emergency Plans
// =============================================================================

func (h *HSEHandler) ListEmergencyPlans(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	status := r.URL.Query().Get("status")

	query := `SELECT id, project_id, plan_number, plan_code, plan_name, plan_type, description, response_procedure, evacuation_routes, assembly_points, emergency_contacts, responsible_person, deputy_person, drill_frequency, last_reviewed, next_review, status, version, approval_date, approved_by, created_at, updated_at FROM hse_emergency_plans WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if projectID != "" { query += fmt.Sprintf(" AND project_id = $%d", argIdx); argIdx++; args = append(args, projectID) }
	if status != "" { query += fmt.Sprintf(" AND status = $%d", argIdx); argIdx++; args = append(args, status) }
	query += " ORDER BY plan_number"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, pid, pcode, pname, ptype, desc, respProc, evacRoutes, assembly, contacts, respPerson, deputy, drillFreq, status, version, appBy string
		var pnum int
		var lastReviewed, nextReview, approvalDate, createdAt, updatedAt time.Time
		err := rows.Scan(&id, &pid, &pnum, &pcode, &pname, &ptype, &desc, &respProc, &evacRoutes, &assembly, &contacts, &respPerson, &deputy, &drillFreq, &lastReviewed, &nextReview, &status, &version, &approvalDate, &appBy, &createdAt, &updatedAt)
		if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
		items = append(items, map[string]interface{}{
			"id": id, "project_id": pid, "plan_number": pnum, "plan_code": pcode,
			"plan_name": pname, "plan_type": ptype, "description": desc,
			"response_procedure": respProc, "evacuation_routes": evacRoutes,
			"assembly_points": assembly, "emergency_contacts": contacts,
			"responsible_person": respPerson, "deputy_person": deputy,
			"drill_frequency": drillFreq, "last_reviewed": lastReviewed, "next_review": nextReview,
			"status": status, "version": version, "approval_date": approvalDate, "approved_by": appBy,
			"created_at": createdAt, "updated_at": updatedAt,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *HSEHandler) CreateEmergencyPlan(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID    string `json:"project_id"`
		PlanCode     string `json:"plan_code"`
		PlanName     string `json:"plan_name"`
		PlanType     string `json:"plan_type"`
		Description  string `json:"description"`
		ResponseProcedure string `json:"response_procedure"`
		ResponsiblePerson *string `json:"responsible_person"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO hse_emergency_plans (id, project_id, plan_number, plan_code, plan_name, plan_type, description, response_procedure, responsible_person, created_at, updated_at) VALUES ($1,$2,(SELECT COALESCE(MAX(plan_number),0)+1 FROM hse_emergency_plans WHERE project_id=$2),$3,$4,$5,$6,$7,$8,$9,$10)`,
		id, input.ProjectID, input.PlanCode, input.PlanName, input.PlanType, input.Description, input.ResponseProcedure, input.ResponsiblePerson, now, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *HSEHandler) GetEmergencyPlan(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var pcode, pname string
	err := h.db.QueryRow(`SELECT plan_code, plan_name FROM hse_emergency_plans WHERE id = $1`, id).Scan(&pcode, &pname)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "emergency plan not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "plan_code": pcode, "plan_name": pname})
}

func (h *HSEHandler) UpdateEmergencyPlan(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Status       *string `json:"status"`
		Version      *string `json:"version"`
		LastReviewed *string `json:"last_reviewed"`
		NextReview   *string `json:"next_review"`
		ApprovedBy   *string `json:"approved_by"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	now := time.Now()
	_, err := h.db.Exec(`UPDATE hse_emergency_plans SET status=COALESCE($1,status), version=COALESCE($2,version), last_reviewed=COALESCE($3,last_reviewed), next_review=COALESCE($4,next_review), approved_by=COALESCE($5,approved_by), updated_at=$6 WHERE id=$7`,
		input.Status, input.Version, input.LastReviewed, input.NextReview, input.ApprovedBy, now, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *HSEHandler) DeleteEmergencyPlan(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM hse_emergency_plans WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// Chemicals
// =============================================================================

func (h *HSEHandler) ListChemicals(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")

	query := `SELECT id, project_id, chemical_code, chemical_name, cas_number, hazard_class, manufacturer, supplier, storage_location, max_quantity, unit, is_hazardous, is_flammable, is_toxic, is_corrosive, is_environmentally_hazardous, sds_revision_date, expiry_date, is_active, created_at, updated_at FROM hse_chemicals WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if projectID != "" { query += fmt.Sprintf(" AND project_id = $%d", argIdx); argIdx++; args = append(args, projectID) }
	query += " ORDER BY chemical_code"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, pid, ccode, cname, cas, hazard, manuf, supp, loc, unit string
		var maxQty float64
		var isHazardous, isFlammable, isToxic, isCorrosive, isEnvHazard, isActive bool
		var sdsRevDate, expiryDate, createdAt, updatedAt time.Time
		err := rows.Scan(&id, &pid, &ccode, &cname, &cas, &hazard, &manuf, &supp, &loc, &maxQty, &unit, &isHazardous, &isFlammable, &isToxic, &isCorrosive, &isEnvHazard, &sdsRevDate, &expiryDate, &isActive, &createdAt, &updatedAt)
		if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
		items = append(items, map[string]interface{}{
			"id": id, "project_id": pid, "chemical_code": ccode, "chemical_name": cname,
			"cas_number": cas, "hazard_class": hazard, "manufacturer": manuf, "supplier": supp,
			"storage_location": loc, "max_quantity": maxQty, "unit": unit,
			"is_hazardous": isHazardous, "is_flammable": isFlammable, "is_toxic": isToxic,
			"is_corrosive": isCorrosive, "is_environmentally_hazardous": isEnvHazard,
			"sds_revision_date": sdsRevDate, "expiry_date": expiryDate, "is_active": isActive,
			"created_at": createdAt, "updated_at": updatedAt,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *HSEHandler) CreateChemical(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID    string  `json:"project_id"`
		ChemicalCode string  `json:"chemical_code"`
		ChemicalName string  `json:"chemical_name"`
		CasNumber    *string `json:"cas_number"`
		HazardClass  *string `json:"hazard_class"`
		Manufacturer *string `json:"manufacturer"`
		IsHazardous  *bool   `json:"is_hazardous"`
		MaxQuantity  *float64 `json:"max_quantity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO hse_chemicals (id, project_id, chemical_code, chemical_name, cas_number, hazard_class, manufacturer, is_hazardous, max_quantity, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		id, input.ProjectID, input.ChemicalCode, input.ChemicalName, input.CasNumber, input.HazardClass, input.Manufacturer, input.IsHazardous, input.MaxQuantity, now, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *HSEHandler) GetChemical(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var ccode, cname string
	err := h.db.QueryRow(`SELECT chemical_code, chemical_name FROM hse_chemicals WHERE id = $1`, id).Scan(&ccode, &cname)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "chemical not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "chemical_code": ccode, "chemical_name": cname})
}

func (h *HSEHandler) UpdateChemical(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		IsActive        *bool   `json:"is_active"`
		MaxQuantity     *float64 `json:"max_quantity"`
		ExpiryDate      *string `json:"expiry_date"`
		SdsRevisionDate *string `json:"sds_revision_date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	now := time.Now()
	_, err := h.db.Exec(`UPDATE hse_chemicals SET is_active=COALESCE($1,is_active), max_quantity=COALESCE($2,max_quantity), expiry_date=COALESCE($3,expiry_date), sds_revision_date=COALESCE($4,sds_revision_date), updated_at=$5 WHERE id=$6`,
		input.IsActive, input.MaxQuantity, input.ExpiryDate, input.SdsRevisionDate, now, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *HSEHandler) DeleteChemical(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM hse_chemicals WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// Summary
// =============================================================================

func (h *HSEHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")

	query := `SELECT project_id, open_incidents, high_severity_incidents, total_incidents, near_misses, lti, fatalities, active_permits, issued_permits, pending_audits, total_audits, inspections_7d, trainings_90d, total_ppe_items, low_stock_ppe, planned_drills, active_plans, hazardous_chemicals FROM hse_summary`
	args := []interface{}{}
	argIdx := 1
	if projectID != "" { query += fmt.Sprintf(" WHERE project_id = $%d", argIdx); args = append(args, projectID) }

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var pid string
		var openInc, highSev, totalInc, nearMiss, lti, fatal, activePerm, issuedPerm, pendingAudit, totalAudit, insp7d, train90d, totalPPE, lowStockPPE, plannedDrill, activePlan, hazChem int
		err := rows.Scan(&pid, &openInc, &highSev, &totalInc, &nearMiss, &lti, &fatal, &activePerm, &issuedPerm, &pendingAudit, &totalAudit, &insp7d, &train90d, &totalPPE, &lowStockPPE, &plannedDrill, &activePlan, &hazChem)
		if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
		items = append(items, map[string]interface{}{
			"project_id": pid, "open_incidents": openInc, "high_severity_incidents": highSev,
			"total_incidents": totalInc, "near_misses": nearMiss, "lti": lti, "fatalities": fatal,
			"active_permits": activePerm, "issued_permits": issuedPerm, "pending_audits": pendingAudit,
			"total_audits": totalAudit, "inspections_7d": insp7d, "trainings_90d": train90d,
			"total_ppe_items": totalPPE, "low_stock_ppe": lowStockPPE, "planned_drills": plannedDrill,
			"active_plans": activePlan, "hazardous_chemicals": hazChem,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func init() {
	log.Println("HSE handler initialized")
}