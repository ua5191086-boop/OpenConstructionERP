package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
)

// ReportTemplate — шаблон отчёта
type ReportTemplate struct {
	ID           string          `json:"id" db:"id"`
	Name         string          `json:"name" db:"name"`
	Description  *string         `json:"description" db:"description"`
	Category     string          `json:"category" db:"category"`
	ReportType   string          `json:"report_type" db:"report_type"`
	DataSource   string          `json:"data_source" db:"data_source"`
	QueryText    *string         `json:"query_text" db:"query_text"`
	Parameters   json.RawMessage `json:"parameters" db:"parameters"`
	ColumnsConfig json.RawMessage `json:"columns_config" db:"columns_config"`
	ChartConfig  json.RawMessage `json:"chart_config" db:"chart_config"`
	Aggregation  json.RawMessage `json:"aggregation" db:"aggregation"`
	Filters      json.RawMessage `json:"filters" db:"filters"`
	SortConfig   json.RawMessage `json:"sort_config" db:"sort_config"`
	ExportFormats json.RawMessage `json:"export_formats" db:"export_formats"`
	IsSystem     bool            `json:"is_system" db:"is_system"`
	IsPublic     bool            `json:"is_public" db:"is_public"`
	OwnerID      *string         `json:"owner_id" db:"owner_id"`
	Version      int             `json:"version" db:"version"`
	Status       string          `json:"status" db:"status"`
	Notes        *string         `json:"notes" db:"notes"`
	CreatedAt    string          `json:"created_at" db:"created_at"`
	UpdatedAt    string          `json:"updated_at" db:"updated_at"`
}

// ReportHandler — HTTP handler for reporting builder
type ReportHandler struct {
	db *sqlx.DB
}

func NewReportHandler(db *sqlx.DB) *ReportHandler {
	return &ReportHandler{db: db}
}

func (h *ReportHandler) RegisterRoutes(r chi.Router) {
	// Report Templates
	r.Route("/reports/templates", func(r chi.Router) {
		r.Get("/", h.ListTemplates)
		r.Post("/", h.CreateTemplate)
		r.Get("/{id}", h.GetTemplate)
		r.Put("/{id}", h.UpdateTemplate)
		r.Delete("/{id}", h.DeleteTemplate)
		r.Get("/category/{category}", h.ListTemplatesByCategory)
		r.Post("/{id}/execute", h.ExecuteTemplate)
	})
	// Saved Reports
	r.Route("/reports/saved", func(r chi.Router) {
		r.Get("/", h.ListSaved)
		r.Post("/", h.CreateSaved)
		r.Get("/{id}", h.GetSaved)
		r.Put("/{id}", h.UpdateSaved)
		r.Delete("/{id}", h.DeleteSaved)
	})
	// Custom Dashboards
	r.Route("/reports/dashboards", func(r chi.Router) {
		r.Get("/", h.ListDashboards)
		r.Post("/", h.CreateDashboard)
		r.Get("/{id}", h.GetDashboard)
		r.Put("/{id}", h.UpdateDashboard)
		r.Delete("/{id}", h.DeleteDashboard)
		r.Get("/project/{projectId}", h.ListDashboardsByProject)
	})
}

// === Report Templates ===

func (h *ReportHandler) ListTemplates(w http.ResponseWriter, r *http.Request) {
	var items []ReportTemplate
	if err := h.db.Select(&items, "SELECT * FROM report_templates WHERE status='active' ORDER BY name"); err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("query failed: %v", err))
		return
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *ReportHandler) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	var input ReportTemplate
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}
	var item ReportTemplate
	err := h.db.Get(&item, `INSERT INTO report_templates
		(name, description, category, report_type, data_source, query_text, parameters,
		 columns_config, chart_config, aggregation, filters, sort_config, export_formats,
		 is_system, is_public, owner_id, status, notes)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18)
		RETURNING *`,
		input.Name, nullableString(input.Description), input.Category, input.ReportType,
		input.DataSource, nullableString(input.QueryText),
		nilifyJSON(input.Parameters), nilifyJSON(input.ColumnsConfig),
		nilifyJSON(input.ChartConfig), nilifyJSON(input.Aggregation),
		nilifyJSON(input.Filters), nilifyJSON(input.SortConfig),
		nilifyJSON(input.ExportFormats), input.IsSystem, input.IsPublic,
		nullableString(input.OwnerID), input.Status, nullableString(input.Notes))
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("insert failed: %v", err))
		return
	}
	respondJSON(w, http.StatusCreated, item)
}

func (h *ReportHandler) GetTemplate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var item ReportTemplate
	if err := h.db.Get(&item, "SELECT * FROM report_templates WHERE id=$1", id); err != nil {
		respondError(w, http.StatusNotFound, "template not found")
		return
	}
	respondJSON(w, http.StatusOK, item)
}

func (h *ReportHandler) UpdateTemplate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input ReportTemplate
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}
	var item ReportTemplate
	err := h.db.Get(&item, `UPDATE report_templates SET
		name=$1, description=$2, category=$3, report_type=$4, data_source=$5, query_text=$6,
		parameters=$7, columns_config=$8, chart_config=$9, aggregation=$10, filters=$11,
		sort_config=$12, export_formats=$13, is_public=$14, status=$15, notes=$16,
		version=version+1, updated_at=NOW()
		WHERE id=$17 RETURNING *`,
		input.Name, nullableString(input.Description), input.Category, input.ReportType,
		input.DataSource, nullableString(input.QueryText),
		nilifyJSON(input.Parameters), nilifyJSON(input.ColumnsConfig),
		nilifyJSON(input.ChartConfig), nilifyJSON(input.Aggregation),
		nilifyJSON(input.Filters), nilifyJSON(input.SortConfig),
		nilifyJSON(input.ExportFormats), input.IsPublic, input.Status,
		nullableString(input.Notes), id)
	if err != nil {
		respondError(w, http.StatusNotFound, "template not found")
		return
	}
	respondJSON(w, http.StatusOK, item)
}

func (h *ReportHandler) DeleteTemplate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	h.db.Exec("UPDATE report_templates SET status='archived' WHERE id=$1", id)
	w.WriteHeader(http.StatusNoContent)
}

func (h *ReportHandler) ListTemplatesByCategory(w http.ResponseWriter, r *http.Request) {
	cat := chi.URLParam(r, "category")
	var items []ReportTemplate
	h.db.Select(&items, "SELECT * FROM report_templates WHERE category=$1 AND status='active' ORDER BY name", cat)
	respondJSON(w, http.StatusOK, items)
}

func (h *ReportHandler) ExecuteTemplate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var params struct {
		Parameters json.RawMessage `json:"parameters"`
	}
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		// Use empty params
	}

	var tmpl ReportTemplate
	if err := h.db.Get(&tmpl, "SELECT * FROM report_templates WHERE id=$1", id); err != nil {
		respondError(w, http.StatusNotFound, "template not found")
		return
	}

	// Log execution
	h.db.Exec(`INSERT INTO report_execution_log
		(template_id, parameter_values, triggered_by, status)
		VALUES ($1, $2, 'api', 'running')`, id, nilifyJSON(params.Parameters))

	// For SQL data sources, execute the query
	var result []map[string]interface{}
	var execErr error
	if tmpl.DataSource == "sql" && tmpl.QueryText != nil {
		execErr = h.db.Select(&result, *tmpl.QueryText)
	} else {
		result = []map[string]interface{}{}
	}

	status := "completed"
	errMsg := ""
	if execErr != nil {
		status = "failed"
		errMsg = execErr.Error()
	}

	h.db.Exec(`UPDATE report_execution_log SET
		status=$1, error_message=$2, row_count=$3
		WHERE template_id=$4 AND created_at=(SELECT MAX(created_at) FROM report_execution_log WHERE template_id=$4)`,
		status, errMsg, len(result), id)

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"template_id": id,
		"name":       tmpl.Name,
		"status":     status,
		"row_count":  len(result),
		"data":       result,
	})
}

// === Saved Reports ===

func (h *ReportHandler) ListSaved(w http.ResponseWriter, r *http.Request) {
	var items []map[string]interface{}
	h.db.Select(&items, "SELECT sr.*, rt.name as template_name FROM saved_reports sr JOIN report_templates rt ON rt.id=sr.template_id ORDER BY sr.created_at DESC")
	respondJSON(w, http.StatusOK, items)
}

func (h *ReportHandler) CreateSaved(w http.ResponseWriter, r *http.Request) {
	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}
	_, err := h.db.NamedExec(`INSERT INTO saved_reports
		(template_id, name, description, parameter_values, schedule, recipients, output_format, notes)
		VALUES (:template_id, :name, :description, :parameter_values, :schedule, :recipients, :output_format, :notes)`, input)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("insert failed: %v", err))
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(input)
}

func (h *ReportHandler) GetSaved(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var item map[string]interface{}
	if err := h.db.Get(&item, "SELECT * FROM saved_reports WHERE id=$1", id); err != nil {
		respondError(w, http.StatusNotFound, "saved report not found")
		return
	}
	respondJSON(w, http.StatusOK, item)
}

func (h *ReportHandler) UpdateSaved(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}
	input["id"] = id
	h.db.NamedExec(`UPDATE saved_reports SET
		name=:name, description=:description, parameter_values=:parameter_values,
		schedule=:schedule, recipients=:recipients, output_format=:output_format,
		notes=:notes, updated_at=NOW() WHERE id=:id`, input)
	respondJSON(w, http.StatusOK, input)
}

func (h *ReportHandler) DeleteSaved(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	h.db.Exec("DELETE FROM saved_reports WHERE id=$1", id)
	w.WriteHeader(http.StatusNoContent)
}

// === Custom Dashboards ===

func (h *ReportHandler) ListDashboards(w http.ResponseWriter, r *http.Request) {
	var items []map[string]interface{}
	h.db.Select(&items, "SELECT * FROM custom_dashboards ORDER BY name")
	respondJSON(w, http.StatusOK, items)
}

func (h *ReportHandler) CreateDashboard(w http.ResponseWriter, r *http.Request) {
	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}
	_, err := h.db.NamedExec(`INSERT INTO custom_dashboards
		(project_id, name, description, layout_config, is_default, is_public, owner_id, auto_refresh_sec, notes)
		VALUES (:project_id, :name, :description, :layout_config, :is_default, :is_public, :owner_id, :auto_refresh_sec, :notes)`, input)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("insert failed: %v", err))
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(input)
}

func (h *ReportHandler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var item map[string]interface{}
	if err := h.db.Get(&item, "SELECT * FROM custom_dashboards WHERE id=$1", id); err != nil {
		respondError(w, http.StatusNotFound, "dashboard not found")
		return
	}
	respondJSON(w, http.StatusOK, item)
}

func (h *ReportHandler) UpdateDashboard(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}
	input["id"] = id
	h.db.NamedExec(`UPDATE custom_dashboards SET
		name=:name, description=:description, layout_config=:layout_config,
		is_default=:is_default, is_public=:is_public, auto_refresh_sec=:auto_refresh_sec,
		notes=:notes, updated_at=NOW() WHERE id=:id`, input)
	respondJSON(w, http.StatusOK, input)
}

func (h *ReportHandler) DeleteDashboard(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	h.db.Exec("DELETE FROM custom_dashboards WHERE id=$1", id)
	w.WriteHeader(http.StatusNoContent)
}

func (h *ReportHandler) ListDashboardsByProject(w http.ResponseWriter, r *http.Request) {
	pid := chi.URLParam(r, "projectId")
	var items []map[string]interface{}
	h.db.Select(&items, "SELECT * FROM custom_dashboards WHERE project_id=$1 ORDER BY name", pid)
	respondJSON(w, http.StatusOK, items)
}

// nilifyJSON converts json.RawMessage to nil if empty
func nilifyJSON(j json.RawMessage) interface{} {
	if len(j) == 0 || string(j) == "null" {
		return nil
	}
	return j
}

func init() {
	log.SetFlags(log.LstdFlags)
}