package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
)

// AuditComplianceHandler — HTTP handler for internal audit & compliance
type AuditComplianceHandler struct {
	db *sqlx.DB
}

func NewAuditComplianceHandler(db *sqlx.DB) *AuditComplianceHandler {
	return &AuditComplianceHandler{db: db}
}

func (h *AuditComplianceHandler) RegisterRoutes(r chi.Router) {
	r.Route("/audits", func(r chi.Router) {
		r.Get("/", h.ListAudits)
		r.Post("/", h.CreateAudit)
		r.Get("/{id}", h.GetAudit)
		r.Put("/{id}", h.UpdateAudit)
		r.Delete("/{id}", h.DeleteAudit)
		r.Get("/project/{projectId}", h.ListAuditsByProject)
		r.Get("/status/{status}", h.ListAuditsByStatus)
	})
	r.Route("/audit-findings", func(r chi.Router) {
		r.Get("/", h.ListFindings)
		r.Post("/", h.CreateFinding)
		r.Get("/{id}", h.GetFinding)
		r.Put("/{id}", h.UpdateFinding)
		r.Get("/audit/{auditId}", h.ListFindingsByAudit)
		r.Get("/open", h.ListOpenFindings)
		r.Get("/overdue", h.ListOverdueFindings)
	})
	r.Route("/compliance-requirements", func(r chi.Router) {
		r.Get("/", h.ListRequirements)
		r.Post("/", h.CreateRequirement)
		r.Get("/{id}", h.GetRequirement)
		r.Put("/{id}", h.UpdateRequirement)
		r.Delete("/{id}", h.DeleteRequirement)
		r.Get("/project/{projectId}", h.ListRequirementsByProject)
		r.Get("/status/{status}", h.ListRequirementsByStatus)
	})
	r.Route("/compliance-checks", func(r chi.Router) {
		r.Get("/", h.ListChecks)
		r.Post("/", h.CreateCheck)
		r.Get("/{id}", h.GetCheck)
		r.Get("/requirement/{requirementId}", h.ListChecksByReq)
	})
}

func (h *AuditComplianceHandler) ListAudits(w http.ResponseWriter, r *http.Request) {
	var items []map[string]interface{}
	h.db.Select(&items, "SELECT * FROM internal_audits ORDER BY scheduled_date DESC NULLS LAST")
	respondJSON(w, http.StatusOK, items)
}

func (h *AuditComplianceHandler) CreateAudit(w http.ResponseWriter, r *http.Request) {
	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}
	_, err := h.db.NamedExec(`INSERT INTO internal_audits
		(project_id, audit_type, audit_scope, audit_period_from, audit_period_to, lead_auditor,
		 audit_team, standards, scheduled_date, actual_date, duration_days, summary, created_by)
		VALUES (:project_id, :audit_type, :audit_scope, :audit_period_from, :audit_period_to, :lead_auditor,
		 :audit_team, :standards, :scheduled_date, :actual_date, :duration_days, :summary, :created_by)`, input)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("insert failed: %v", err))
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(input)
}

func (h *AuditComplianceHandler) GetAudit(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var item map[string]interface{}
	if err := h.db.Get(&item, "SELECT * FROM internal_audits WHERE id=$1", id); err != nil {
		respondError(w, http.StatusNotFound, "audit not found")
		return
	}
	respondJSON(w, http.StatusOK, item)
}

func (h *AuditComplianceHandler) UpdateAudit(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}
	input["id"] = id
	h.db.NamedExec(`UPDATE internal_audits SET
		audit_type=:audit_type, audit_scope=:audit_scope, lead_auditor=:lead_auditor,
		status=:status, overall_rating=:overall_rating, summary=:summary,
		conclusion=:conclusion, report_document=:report_document, updated_at=NOW()
		WHERE id=:id`, input)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(input)
}

func (h *AuditComplianceHandler) DeleteAudit(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	h.db.Exec("DELETE FROM internal_audits WHERE id=$1", id)
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuditComplianceHandler) ListAuditsByProject(w http.ResponseWriter, r *http.Request) {
	pid := chi.URLParam(r, "projectId")
	var items []map[string]interface{}
	h.db.Select(&items, "SELECT * FROM internal_audits WHERE project_id=$1 ORDER BY scheduled_date DESC", pid)
	respondJSON(w, http.StatusOK, items)
}

func (h *AuditComplianceHandler) ListAuditsByStatus(w http.ResponseWriter, r *http.Request) {
	status := chi.URLParam(r, "status")
	var items []map[string]interface{}
	h.db.Select(&items, "SELECT * FROM internal_audits WHERE status=$1 ORDER BY scheduled_date DESC", status)
	respondJSON(w, http.StatusOK, items)
}

func (h *AuditComplianceHandler) ListFindings(w http.ResponseWriter, r *http.Request) {
	var items []map[string]interface{}
	h.db.Select(&items, "SELECT * FROM audit_findings ORDER BY created_at DESC")
	respondJSON(w, http.StatusOK, items)
}

func (h *AuditComplianceHandler) CreateFinding(w http.ResponseWriter, r *http.Request) {
	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}
	_, err := h.db.NamedExec(`INSERT INTO audit_findings
		(audit_id, finding_type, finding_code, title, description, root_cause, impact,
		 corrective_action, preventive_action, deadline_date, assigned_to)
		VALUES (:audit_id, :finding_type, :finding_code, :title, :description, :root_cause, :impact,
		 :corrective_action, :preventive_action, :deadline_date, :assigned_to)`, input)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("insert failed: %v", err))
		return
	}
	// Update audit finding counters
	h.db.Exec(`UPDATE internal_audits SET
		findings_count = findings_count + 1,
		critical_findings = critical_findings + CASE WHEN $1='critical' THEN 1 ELSE 0 END,
		major_findings = major_findings + CASE WHEN $1='major' THEN 1 ELSE 0 END,
		minor_findings = minor_findings + CASE WHEN $1='minor' THEN 1 ELSE 0 END
		WHERE id=$2`, input["finding_type"], input["audit_id"])
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(input)
}

func (h *AuditComplianceHandler) GetFinding(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var item map[string]interface{}
	if err := h.db.Get(&item, "SELECT * FROM audit_findings WHERE id=$1", id); err != nil {
		respondError(w, http.StatusNotFound, "finding not found")
		return
	}
	respondJSON(w, http.StatusOK, item)
}

func (h *AuditComplianceHandler) UpdateFinding(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}
	input["id"] = id
	h.db.NamedExec(`UPDATE audit_findings SET
		finding_type=:finding_type, title=:title, description=:description,
		corrective_action=:corrective_action, status=:status,
		resolution_notes=:resolution_notes, updated_at=NOW()
		WHERE id=:id`, input)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(input)
}

func (h *AuditComplianceHandler) ListFindingsByAudit(w http.ResponseWriter, r *http.Request) {
	aid := chi.URLParam(r, "auditId")
	var items []map[string]interface{}
	h.db.Select(&items, "SELECT * FROM audit_findings WHERE audit_id=$1 ORDER BY finding_type, finding_code", aid)
	respondJSON(w, http.StatusOK, items)
}

func (h *AuditComplianceHandler) ListOpenFindings(w http.ResponseWriter, r *http.Request) {
	var items []map[string]interface{}
	h.db.Select(&items, "SELECT * FROM audit_findings WHERE status NOT IN ('closed','resolved') ORDER BY deadline_date ASC")
	respondJSON(w, http.StatusOK, items)
}

func (h *AuditComplianceHandler) ListOverdueFindings(w http.ResponseWriter, r *http.Request) {
	var items []map[string]interface{}
	h.db.Select(&items, "SELECT * FROM audit_findings WHERE deadline_date < CURRENT_DATE AND status NOT IN ('closed','resolved') ORDER BY deadline_date ASC")
	respondJSON(w, http.StatusOK, items)
}

func (h *AuditComplianceHandler) ListRequirements(w http.ResponseWriter, r *http.Request) {
	var items []map[string]interface{}
	h.db.Select(&items, "SELECT * FROM compliance_requirements ORDER BY requirement_code")
	respondJSON(w, http.StatusOK, items)
}

func (h *AuditComplianceHandler) CreateRequirement(w http.ResponseWriter, r *http.Request) {
	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}
	_, err := h.db.NamedExec(`INSERT INTO compliance_requirements
		(project_id, requirement_code, requirement_type, title, description, authority,
		 effective_date, expiry_date, applies_to, risk_if_noncompliant, control_measure,
		 review_frequency, responsible_party, notes)
		VALUES (:project_id, :requirement_code, :requirement_type, :title, :description, :authority,
		 :effective_date, :expiry_date, :applies_to, :risk_if_noncompliant, :control_measure,
		 :review_frequency, :responsible_party, :notes)`, input)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("insert failed: %v", err))
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(input)
}

func (h *AuditComplianceHandler) GetRequirement(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var item map[string]interface{}
	if err := h.db.Get(&item, "SELECT * FROM compliance_requirements WHERE id=$1", id); err != nil {
		respondError(w, http.StatusNotFound, "requirement not found")
		return
	}
	respondJSON(w, http.StatusOK, item)
}

func (h *AuditComplianceHandler) UpdateRequirement(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}
	input["id"] = id
	h.db.NamedExec(`UPDATE compliance_requirements SET
		title=:title, description=:description, status=:status,
		review_frequency=:review_frequency, last_review_date=:last_review_date,
		next_review_date=:next_review_date, responsible_party=:responsible_party,
		notes=:notes, updated_at=NOW()
		WHERE id=:id`, input)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(input)
}

func (h *AuditComplianceHandler) DeleteRequirement(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	h.db.Exec("DELETE FROM compliance_requirements WHERE id=$1", id)
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuditComplianceHandler) ListRequirementsByProject(w http.ResponseWriter, r *http.Request) {
	pid := chi.URLParam(r, "projectId")
	var items []map[string]interface{}
	h.db.Select(&items, "SELECT * FROM compliance_requirements WHERE project_id=$1 ORDER BY requirement_code", pid)
	respondJSON(w, http.StatusOK, items)
}

func (h *AuditComplianceHandler) ListRequirementsByStatus(w http.ResponseWriter, r *http.Request) {
	status := chi.URLParam(r, "status")
	var items []map[string]interface{}
	h.db.Select(&items, "SELECT * FROM compliance_requirements WHERE status=$1 ORDER BY requirement_code", status)
	respondJSON(w, http.StatusOK, items)
}

func (h *AuditComplianceHandler) ListChecks(w http.ResponseWriter, r *http.Request) {
	var items []map[string]interface{}
	h.db.Select(&items, "SELECT * FROM compliance_checks ORDER BY check_date DESC")
	respondJSON(w, http.StatusOK, items)
}

func (h *AuditComplianceHandler) CreateCheck(w http.ResponseWriter, r *http.Request) {
	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}
	_, err := h.db.NamedExec(`INSERT INTO compliance_checks
		(requirement_id, project_id, check_type, check_date, checked_by, result,
		 evidence, non_compliance_detail, corrective_action, deadline_date)
		VALUES (:requirement_id, :project_id, :check_type, :check_date, :checked_by, :result,
		 :evidence, :non_compliance_detail, :corrective_action, :deadline_date)`, input)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("insert failed: %v", err))
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(input)
}

func (h *AuditComplianceHandler) GetCheck(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var item map[string]interface{}
	if err := h.db.Get(&item, "SELECT * FROM compliance_checks WHERE id=$1", id); err != nil {
		respondError(w, http.StatusNotFound, "check not found")
		return
	}
	respondJSON(w, http.StatusOK, item)
}

func (h *AuditComplianceHandler) ListChecksByReq(w http.ResponseWriter, r *http.Request) {
	rid := chi.URLParam(r, "requirementId")
	var items []map[string]interface{}
	h.db.Select(&items, "SELECT * FROM compliance_checks WHERE requirement_id=$1 ORDER BY check_date DESC", rid)
	respondJSON(w, http.StatusOK, items)
}

func init() { log.SetFlags(log.LstdFlags) }