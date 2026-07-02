package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
)

// BenchmarkHandler — HTTP handler for performance benchmarking
type BenchmarkHandler struct {
	db *sqlx.DB
}

func NewBenchmarkHandler(db *sqlx.DB) *BenchmarkHandler {
	return &BenchmarkHandler{db: db}
}

func (h *BenchmarkHandler) RegisterRoutes(r chi.Router) {
	r.Route("/benchmarks/templates", func(r chi.Router) {
		r.Get("/", h.ListTemplates)
		r.Post("/", h.CreateTemplate)
		r.Get("/{id}", h.GetTemplate)
		r.Put("/{id}", h.UpdateTemplate)
		r.Delete("/{id}", h.DeleteTemplate)
		r.Get("/type/{projectType}", h.ListTemplatesByType)
	})
	r.Route("/benchmarks/kpis", func(r chi.Router) {
		r.Get("/", h.ListKPIs)
		r.Post("/", h.CreateKPI)
		r.Get("/{id}", h.GetKPI)
		r.Put("/{id}", h.UpdateKPI)
		r.Delete("/{id}", h.DeleteKPI)
		r.Get("/template/{templateId}", h.ListKPIsByTemplate)
		r.Get("/category/{category}", h.ListKPIsByCategory)
	})
	r.Route("/benchmarks/results", func(r chi.Router) {
		r.Get("/", h.ListResults)
		r.Post("/", h.CreateResult)
		r.Get("/project/{projectId}", h.ListResultsByProject)
		r.Get("/compare/{projectId1}/{projectId2}", h.CompareProjects)
	})
	r.Get("/benchmarks/summary/{projectId}", h.GetProjectSummary)
}

func (h *BenchmarkHandler) ListTemplates(w http.ResponseWriter, r *http.Request) {
	var items []map[string]interface{}
	h.db.Select(&items, "SELECT * FROM benchmark_templates ORDER BY name")
	respondJSON(w, http.StatusOK, items)
}

func (h *BenchmarkHandler) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}
	_, err := h.db.NamedExec(`INSERT INTO benchmark_templates
		(name, description, project_type, region, source, reliability_score, is_public, version, notes)
		VALUES (:name, :description, :project_type, :region, :source, :reliability_score, :is_public, :version, :notes)`, input)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("insert failed: %v", err))
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(input)
}

func (h *BenchmarkHandler) GetTemplate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var item map[string]interface{}
	if err := h.db.Get(&item, "SELECT * FROM benchmark_templates WHERE id=$1", id); err != nil {
		respondError(w, http.StatusNotFound, "template not found")
		return
	}
	respondJSON(w, http.StatusOK, item)
}

func (h *BenchmarkHandler) UpdateTemplate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}
	input["id"] = id
	h.db.NamedExec(`UPDATE benchmark_templates SET
		name=:name, description=:description, project_type=:project_type, region=:region,
		source=:source, reliability_score=:reliability_score, is_public=:is_public,
		version=:version, notes=:notes, updated_at=NOW()
		WHERE id=:id`, input)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(input)
}

func (h *BenchmarkHandler) DeleteTemplate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	h.db.Exec("DELETE FROM benchmark_templates WHERE id=$1", id)
	w.WriteHeader(http.StatusNoContent)
}

func (h *BenchmarkHandler) ListTemplatesByType(w http.ResponseWriter, r *http.Request) {
	pt := chi.URLParam(r, "projectType")
	var items []map[string]interface{}
	h.db.Select(&items, "SELECT * FROM benchmark_templates WHERE project_type=$1 ORDER BY name", pt)
	respondJSON(w, http.StatusOK, items)
}

// KPIs
func (h *BenchmarkHandler) ListKPIs(w http.ResponseWriter, r *http.Request) {
	var items []map[string]interface{}
	h.db.Select(&items, "SELECT * FROM benchmark_kpis ORDER BY category, kpi_code")
	respondJSON(w, http.StatusOK, items)
}

func (h *BenchmarkHandler) CreateKPI(w http.ResponseWriter, r *http.Request) {
	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}
	_, err := h.db.NamedExec(`INSERT INTO benchmark_kpis
		(template_id, kpi_code, kpi_name, category, unit, p10_value, p25_value, p50_value,
		 p75_value, p90_value, mean_value, std_dev, sample_size, period_from, period_to,
		 formula_desc, notes)
		VALUES (:template_id, :kpi_code, :kpi_name, :category, :unit, :p10_value, :p25_value,
		 :p50_value, :p75_value, :p90_value, :mean_value, :std_dev, :sample_size,
		 :period_from, :period_to, :formula_desc, :notes)`, input)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("insert failed: %v", err))
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(input)
}

func (h *BenchmarkHandler) GetKPI(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var item map[string]interface{}
	if err := h.db.Get(&item, "SELECT * FROM benchmark_kpis WHERE id=$1", id); err != nil {
		respondError(w, http.StatusNotFound, "kpi not found")
		return
	}
	respondJSON(w, http.StatusOK, item)
}

func (h *BenchmarkHandler) UpdateKPI(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}
	input["id"] = id
	h.db.NamedExec(`UPDATE benchmark_kpis SET
		kpi_code=:kpi_code, kpi_name=:kpi_name, category=:category, unit=:unit,
		p10_value=:p10_value, p25_value=:p25_value, p50_value=:p50_value,
		p75_value=:p75_value, p90_value=:p90_value, mean_value=:mean_value,
		std_dev=:std_dev, sample_size=:sample_size, period_from=:period_from,
		period_to=:period_to, formula_desc=:formula_desc, notes=:notes
		WHERE id=:id`, input)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(input)
}

func (h *BenchmarkHandler) DeleteKPI(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	h.db.Exec("DELETE FROM benchmark_kpis WHERE id=$1", id)
	w.WriteHeader(http.StatusNoContent)
}

func (h *BenchmarkHandler) ListKPIsByTemplate(w http.ResponseWriter, r *http.Request) {
	tid := chi.URLParam(r, "templateId")
	var items []map[string]interface{}
	h.db.Select(&items, "SELECT * FROM benchmark_kpis WHERE template_id=$1 ORDER BY category, kpi_code", tid)
	respondJSON(w, http.StatusOK, items)
}

func (h *BenchmarkHandler) ListKPIsByCategory(w http.ResponseWriter, r *http.Request) {
	cat := chi.URLParam(r, "category")
	var items []map[string]interface{}
	h.db.Select(&items, "SELECT * FROM benchmark_kpis WHERE category=$1 ORDER BY kpi_code", cat)
	respondJSON(w, http.StatusOK, items)
}

// Results
func (h *BenchmarkHandler) ListResults(w http.ResponseWriter, r *http.Request) {
	var items []map[string]interface{}
	h.db.Select(&items, `SELECT br.*, bk.kpi_name, bk.kpi_code, bk.unit, bt.name as benchmark_name
		FROM benchmark_results br
		JOIN benchmark_kpis bk ON bk.id=br.benchmark_kpi_id
		JOIN benchmark_templates bt ON bt.id=br.template_id
		ORDER BY br.assessment_date DESC`)
	respondJSON(w, http.StatusOK, items)
}

func (h *BenchmarkHandler) CreateResult(w http.ResponseWriter, r *http.Request) {
	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}
	_, err := h.db.NamedExec(`INSERT INTO benchmark_results
		(project_id, benchmark_kpi_id, template_id, project_value, benchmark_value,
		 variance_pct, percentile_rank, rating, assessment_date, assessed_by, notes)
		VALUES (:project_id, :benchmark_kpi_id, :template_id, :project_value, :benchmark_value,
		 :variance_pct, :percentile_rank, :rating, :assessment_date, :assessed_by, :notes)`, input)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("insert failed: %v", err))
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(input)
}

func (h *BenchmarkHandler) ListResultsByProject(w http.ResponseWriter, r *http.Request) {
	pid := chi.URLParam(r, "projectId")
	var items []map[string]interface{}
	h.db.Select(&items, `SELECT br.*, bk.kpi_name, bk.kpi_code, bk.unit, bt.name as benchmark_name
		FROM benchmark_results br
		JOIN benchmark_kpis bk ON bk.id=br.benchmark_kpi_id
		JOIN benchmark_templates bt ON bt.id=br.template_id
		WHERE br.project_id=$1
		ORDER BY br.assessment_date DESC`, pid)
	respondJSON(w, http.StatusOK, items)
}

func (h *BenchmarkHandler) CompareProjects(w http.ResponseWriter, r *http.Request) {
	p1 := chi.URLParam(r, "projectId1")
	p2 := chi.URLParam(r, "projectId2")
	var items []map[string]interface{}
	h.db.Select(&items, `SELECT bk.kpi_code, bk.kpi_name, bk.unit,
		MAX(CASE WHEN br.project_id=$1 THEN br.project_value END) as project1_value,
		MAX(CASE WHEN br.project_id=$2 THEN br.project_value END) as project2_value,
		MAX(CASE WHEN br.project_id=$1 THEN br.rating END) as project1_rating,
		MAX(CASE WHEN br.project_id=$2 THEN br.rating END) as project2_rating,
		bk.p50_value as industry_benchmark
		FROM benchmark_results br
		JOIN benchmark_kpis bk ON bk.id=br.benchmark_kpi_id
		WHERE br.project_id IN ($1, $2)
		GROUP BY bk.kpi_code, bk.kpi_name, bk.unit, bk.p50_value
		ORDER BY bk.category, bk.kpi_code`, p1, p2)
	respondJSON(w, http.StatusOK, items)
}

func (h *BenchmarkHandler) GetProjectSummary(w http.ResponseWriter, r *http.Request) {
	pid := chi.URLParam(r, "projectId")
	var items []map[string]interface{}
	h.db.Select(&items, "SELECT * FROM project_benchmark_summary WHERE project_id=$1 ORDER BY category, kpi_code", pid)
	respondJSON(w, http.StatusOK, items)
}

func init() { log.SetFlags(log.LstdFlags) }