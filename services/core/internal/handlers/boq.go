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

// BOQHandler handles BOQ module endpoints
type BOQHandler struct {
	db *sql.DB
}

func NewBOQHandler(db *sql.DB) *BOQHandler {
	return &BOQHandler{db: db}
}

// RegisterRoutes registers BOQ routes under the given router
func (h *BOQHandler) RegisterRoutes(r chi.Router) {
	r.Route("/boq", func(r chi.Router) {
		// CBS Chapters
		r.Get("/cbs-chapters", h.ListCBSChapters)
		r.Post("/cbs-chapters", h.CreateCBSChapter)
		r.Get("/cbs-chapters/{id}", h.GetCBSChapter)
		r.Put("/cbs-chapters/{id}", h.UpdateCBSChapter)
		r.Delete("/cbs-chapters/{id}", h.DeleteCBSChapter)

		// BOQ Sections
		r.Get("/sections", h.ListSections)
		r.Post("/sections", h.CreateSection)
		r.Get("/sections/{id}", h.GetSection)
		r.Put("/sections/{id}", h.UpdateSection)
		r.Delete("/sections/{id}", h.DeleteSection)

		// BOQ Complexes
		r.Get("/complexes", h.ListComplexes)
		r.Post("/complexes", h.CreateComplex)
		r.Get("/complexes/{id}", h.GetComplex)
		r.Put("/complexes/{id}", h.UpdateComplex)
		r.Delete("/complexes/{id}", h.DeleteComplex)

		// BOQ Objects
		r.Get("/objects", h.ListObjects)
		r.Post("/objects", h.CreateObject)
		r.Get("/objects/{id}", h.GetObject)
		r.Put("/objects/{id}", h.UpdateObject)
		r.Delete("/objects/{id}", h.DeleteObject)

		// BOQ Items
		r.Get("/items", h.ListItems)
		r.Post("/items", h.CreateItem)
		r.Get("/items/{id}", h.GetItem)
		r.Put("/items/{id}", h.UpdateItem)
		r.Delete("/items/{id}", h.DeleteItem)

		// Cost Transactions
		r.Get("/cost-transactions", h.ListCostTransactions)
		r.Post("/cost-transactions", h.CreateCostTransaction)
		r.Get("/cost-transactions/{id}", h.GetCostTransaction)
		r.Put("/cost-transactions/{id}", h.UpdateCostTransaction)
		r.Delete("/cost-transactions/{id}", h.DeleteCostTransaction)

		// Budget Versions
		r.Get("/budget-versions", h.ListBudgetVersions)
		r.Post("/budget-versions", h.CreateBudgetVersion)
		r.Get("/budget-versions/{id}", h.GetBudgetVersion)
		r.Put("/budget-versions/{id}", h.UpdateBudgetVersion)
		r.Delete("/budget-versions/{id}", h.DeleteBudgetVersion)
	})
}

// --- CBS Chapters ---

// ListCBSChapters godoc
// @Summary List CBS chapters
// @Tags BOQ
// @Produce json
// @Param project_id query string false "Filter by project ID"
// @Success 200 {object} models.APIResponse
// @Router /api/v1/boq/cbs-chapters [get]
func (h *BOQHandler) ListCBSChapters(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	var rows *sql.Rows
	var err error

	if projectID != "" {
		rows, err = h.db.Query(`SELECT id, project_id, code, name, name_ru, parent_id, level, sort_order, path, is_active, created_at, updated_at FROM cbs_chapters WHERE project_id = $1 OR project_id IS NULL ORDER BY sort_order`, projectID)
	} else {
		rows, err = h.db.Query(`SELECT id, project_id, code, name, name_ru, parent_id, level, sort_order, path, is_active, created_at, updated_at FROM cbs_chapters ORDER BY sort_order`)
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	chapters := make([]models.CBSChapter, 0)
	for rows.Next() {
		var c models.CBSChapter
		if err := rows.Scan(&c.ID, &c.ProjectID, &c.Code, &c.Name, &c.NameRU, &c.ParentID, &c.Level, &c.SortOrder, &c.Path, &c.IsActive, &c.CreatedAt, &c.UpdatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		chapters = append(chapters, c)
	}
	respondJSON(w, http.StatusOK, chapters)
}

// CreateCBSChapter godoc
// @Summary Create a CBS chapter
// @Tags BOQ
// @Accept json
// @Produce json
// @Param body body object true "CBS Chapter data"
// @Success 201 {object} models.APIResponse
// @Router /api/v1/boq/cbs-chapters [post]
func (h *BOQHandler) CreateCBSChapter(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID *string `json:"project_id"`
		Code      string  `json:"code"`
		Name      string  `json:"name"`
		NameRU    *string `json:"name_ru"`
		ParentID  *string `json:"parent_id"`
		Level     int     `json:"level"`
		SortOrder int     `json:"sort_order"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO cbs_chapters (id, project_id, code, name, name_ru, parent_id, level, sort_order, is_active, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,true,$9,$10)`,
		id, input.ProjectID, input.Code, input.Name, input.NameRU, input.ParentID, input.Level, input.SortOrder, now, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

// GetCBSChapter godoc
// @Summary Get a CBS chapter by ID
// @Tags BOQ
// @Produce json
// @Param id path string true "Chapter ID"
// @Success 200 {object} models.APIResponse
// @Router /api/v1/boq/cbs-chapters/{id} [get]
func (h *BOQHandler) GetCBSChapter(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var c models.CBSChapter
	err := h.db.QueryRow(`SELECT id, project_id, code, name, name_ru, parent_id, level, sort_order, path, is_active, created_at, updated_at FROM cbs_chapters WHERE id = $1`, id).
		Scan(&c.ID, &c.ProjectID, &c.Code, &c.Name, &c.NameRU, &c.ParentID, &c.Level, &c.SortOrder, &c.Path, &c.IsActive, &c.CreatedAt, &c.UpdatedAt)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "CBS chapter not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, c)
}

// UpdateCBSChapter godoc
// @Summary Update a CBS chapter
// @Tags BOQ
// @Accept json
// @Produce json
// @Param id path string true "Chapter ID"
// @Param body body object true "CBS Chapter data"
// @Success 200 {object} models.APIResponse
// @Router /api/v1/boq/cbs-chapters/{id} [put]
func (h *BOQHandler) UpdateCBSChapter(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Code      *string `json:"code"`
		Name      *string `json:"name"`
		NameRU    *string `json:"name_ru"`
		ParentID  *string `json:"parent_id"`
		Level     *int    `json:"level"`
		SortOrder *int    `json:"sort_order"`
		IsActive  *bool   `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	now := time.Now()
	_, err := h.db.Exec(`UPDATE cbs_chapters SET code=COALESCE($1,code), name=COALESCE($2,name), name_ru=COALESCE($3,name_ru), parent_id=COALESCE($4,parent_id), level=COALESCE($5,level), sort_order=COALESCE($6,sort_order), is_active=COALESCE($7,is_active), updated_at=$8 WHERE id=$9`,
		input.Code, input.Name, input.NameRU, input.ParentID, input.Level, input.SortOrder, input.IsActive, now, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// DeleteCBSChapter godoc
// @Summary Delete a CBS chapter
// @Tags BOQ
// @Produce json
// @Param id path string true "Chapter ID"
// @Success 200 {object} models.APIResponse
// @Router /api/v1/boq/cbs-chapters/{id} [delete]
func (h *BOQHandler) DeleteCBSChapter(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM cbs_chapters WHERE id = $1`, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// --- BOQ Sections ---

func (h *BOQHandler) ListSections(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT id, project_id, code, name, section_type, start_km, end_km, sort_order, created_at, updated_at FROM boq_sections`
	var rows *sql.Rows
	var err error
	if projectID != "" {
		rows, err = h.db.Query(query+` WHERE project_id = $1 ORDER BY sort_order`, projectID)
	} else {
		rows, err = h.db.Query(query + ` ORDER BY sort_order`)
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	sections := make([]models.BOQSection, 0)
	for rows.Next() {
		var s models.BOQSection
		if err := rows.Scan(&s.ID, &s.ProjectID, &s.Code, &s.Name, &s.SectionType, &s.StartKM, &s.EndKM, &s.SortOrder, &s.CreatedAt, &s.UpdatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		sections = append(sections, s)
	}
	respondJSON(w, http.StatusOK, sections)
}

func (h *BOQHandler) CreateSection(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID   string  `json:"project_id"`
		Code        string  `json:"code"`
		Name        string  `json:"name"`
		SectionType string  `json:"section_type"`
		StartKM     *float64 `json:"start_km"`
		EndKM       *float64 `json:"end_km"`
		SortOrder   int     `json:"sort_order"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO boq_sections (id, project_id, code, name, section_type, start_km, end_km, sort_order, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		id, input.ProjectID, input.Code, input.Name, input.SectionType, input.StartKM, input.EndKM, input.SortOrder, now, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *BOQHandler) GetSection(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var s models.BOQSection
	err := h.db.QueryRow(`SELECT id, project_id, code, name, section_type, start_km, end_km, sort_order, created_at, updated_at FROM boq_sections WHERE id = $1`, id).
		Scan(&s.ID, &s.ProjectID, &s.Code, &s.Name, &s.SectionType, &s.StartKM, &s.EndKM, &s.SortOrder, &s.CreatedAt, &s.UpdatedAt)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "section not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, s)
}

func (h *BOQHandler) UpdateSection(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Code        *string  `json:"code"`
		Name        *string  `json:"name"`
		SectionType *string  `json:"section_type"`
		StartKM     *float64 `json:"start_km"`
		EndKM       *float64 `json:"end_km"`
		SortOrder   *int     `json:"sort_order"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	now := time.Now()
	_, err := h.db.Exec(`UPDATE boq_sections SET code=COALESCE($1,code), name=COALESCE($2,name), section_type=COALESCE($3,section_type), start_km=COALESCE($4,start_km), end_km=COALESCE($5,end_km), sort_order=COALESCE($6,sort_order), updated_at=$7 WHERE id=$8`,
		input.Code, input.Name, input.SectionType, input.StartKM, input.EndKM, input.SortOrder, now, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *BOQHandler) DeleteSection(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM boq_sections WHERE id = $1`, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// --- BOQ Complexes ---

func (h *BOQHandler) ListComplexes(w http.ResponseWriter, r *http.Request) {
	sectionID := r.URL.Query().Get("section_id")
	query := `SELECT id, project_id, section_id, code, name, sort_order, created_at, updated_at FROM boq_complexes`
	var rows *sql.Rows
	var err error
	if sectionID != "" {
		rows, err = h.db.Query(query+` WHERE section_id = $1 ORDER BY sort_order`, sectionID)
	} else {
		rows, err = h.db.Query(query + ` ORDER BY sort_order`)
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	complexes := make([]models.BOQComplex, 0)
	for rows.Next() {
		var c models.BOQComplex
		if err := rows.Scan(&c.ID, &c.ProjectID, &c.SectionID, &c.Code, &c.Name, &c.SortOrder, &c.CreatedAt, &c.UpdatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		complexes = append(complexes, c)
	}
	respondJSON(w, http.StatusOK, complexes)
}

func (h *BOQHandler) CreateComplex(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID string `json:"project_id"`
		SectionID string `json:"section_id"`
		Code      string `json:"code"`
		Name      string `json:"name"`
		SortOrder int    `json:"sort_order"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO boq_complexes (id, project_id, section_id, code, name, sort_order, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		id, input.ProjectID, input.SectionID, input.Code, input.Name, input.SortOrder, now, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *BOQHandler) GetComplex(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var c models.BOQComplex
	err := h.db.QueryRow(`SELECT id, project_id, section_id, code, name, sort_order, created_at, updated_at FROM boq_complexes WHERE id = $1`, id).
		Scan(&c.ID, &c.ProjectID, &c.SectionID, &c.Code, &c.Name, &c.SortOrder, &c.CreatedAt, &c.UpdatedAt)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "complex not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, c)
}

func (h *BOQHandler) UpdateComplex(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Code      *string `json:"code"`
		Name      *string `json:"name"`
		SortOrder *int    `json:"sort_order"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	now := time.Now()
	_, err := h.db.Exec(`UPDATE boq_complexes SET code=COALESCE($1,code), name=COALESCE($2,name), sort_order=COALESCE($3,sort_order), updated_at=$4 WHERE id=$5`,
		input.Code, input.Name, input.SortOrder, now, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *BOQHandler) DeleteComplex(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM boq_complexes WHERE id = $1`, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// --- BOQ Objects ---

func (h *BOQHandler) ListObjects(w http.ResponseWriter, r *http.Request) {
	complexID := r.URL.Query().Get("complex_id")
	query := `SELECT id, project_id, complex_id, code, name, sort_order, created_at, updated_at FROM boq_objects`
	var rows *sql.Rows
	var err error
	if complexID != "" {
		rows, err = h.db.Query(query+` WHERE complex_id = $1 ORDER BY sort_order`, complexID)
	} else {
		rows, err = h.db.Query(query + ` ORDER BY sort_order`)
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	objects := make([]models.BOQObject, 0)
	for rows.Next() {
		var o models.BOQObject
		if err := rows.Scan(&o.ID, &o.ProjectID, &o.ComplexID, &o.Code, &o.Name, &o.SortOrder, &o.CreatedAt, &o.UpdatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		objects = append(objects, o)
	}
	respondJSON(w, http.StatusOK, objects)
}

func (h *BOQHandler) CreateObject(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID string `json:"project_id"`
		ComplexID string `json:"complex_id"`
		Code      string `json:"code"`
		Name      string `json:"name"`
		SortOrder int    `json:"sort_order"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO boq_objects (id, project_id, complex_id, code, name, sort_order, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		id, input.ProjectID, input.ComplexID, input.Code, input.Name, input.SortOrder, now, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *BOQHandler) GetObject(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var o models.BOQObject
	err := h.db.QueryRow(`SELECT id, project_id, complex_id, code, name, sort_order, created_at, updated_at FROM boq_objects WHERE id = $1`, id).
		Scan(&o.ID, &o.ProjectID, &o.ComplexID, &o.Code, &o.Name, &o.SortOrder, &o.CreatedAt, &o.UpdatedAt)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "object not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, o)
}

func (h *BOQHandler) UpdateObject(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Code      *string `json:"code"`
		Name      *string `json:"name"`
		SortOrder *int    `json:"sort_order"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	now := time.Now()
	_, err := h.db.Exec(`UPDATE boq_objects SET code=COALESCE($1,code), name=COALESCE($2,name), sort_order=COALESCE($3,sort_order), updated_at=$4 WHERE id=$5`,
		input.Code, input.Name, input.SortOrder, now, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *BOQHandler) DeleteObject(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM boq_objects WHERE id = $1`, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// --- BOQ Items ---

func (h *BOQHandler) ListItems(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	objectID := r.URL.Query().Get("object_id")
	status := r.URL.Query().Get("status")

	query := `SELECT id, project_id, object_id, cbs_chapter_id, code, name, description, unit, quantity, unit_price, total_cost, currency, contractor_id, contract_id, funding_source, phase, status, notes, created_at, updated_at FROM boq_items WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if projectID != "" {
		query += fmt.Sprintf(" AND project_id = $%d", argIdx)
		args = append(args, projectID)
		argIdx++
	}
	if objectID != "" {
		query += fmt.Sprintf(" AND object_id = $%d", argIdx)
		args = append(args, objectID)
		argIdx++
	}
	if status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	}
	query += " ORDER BY code"

	rows, err := h.db.Query(query, args...)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	items := make([]models.BOQItem, 0)
	for rows.Next() {
		var i models.BOQItem
		if err := rows.Scan(&i.ID, &i.ProjectID, &i.ObjectID, &i.CBSChapterID, &i.Code, &i.Name, &i.Description, &i.Unit, &i.Quantity, &i.UnitPrice, &i.TotalCost, &i.Currency, &i.ContractorID, &i.ContractID, &i.FundingSource, &i.Phase, &i.Status, &i.Notes, &i.CreatedAt, &i.UpdatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, i)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *BOQHandler) CreateItem(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID    string  `json:"project_id"`
		ObjectID     string  `json:"object_id"`
		CBSChapterID string  `json:"cbs_chapter_id"`
		Code         string  `json:"code"`
		Name         string  `json:"name"`
		Description  *string `json:"description"`
		Unit         string  `json:"unit"`
		Quantity     float64 `json:"quantity"`
		UnitPrice    float64 `json:"unit_price"`
		Currency     string  `json:"currency"`
		ContractorID *string `json:"contractor_id"`
		ContractID   *string `json:"contract_id"`
		FundingSource *string `json:"funding_source"`
		Phase        *string `json:"phase"`
		Status       string  `json:"status"`
		Notes        *string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO boq_items (id, project_id, object_id, cbs_chapter_id, code, name, description, unit, quantity, unit_price, currency, contractor_id, contract_id, funding_source, phase, status, notes, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)`,
		id, input.ProjectID, input.ObjectID, input.CBSChapterID, input.Code, input.Name, input.Description, input.Unit, input.Quantity, input.UnitPrice, input.Currency, input.ContractorID, input.ContractID, input.FundingSource, input.Phase, input.Status, input.Notes, now, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *BOQHandler) GetItem(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var i models.BOQItem
	err := h.db.QueryRow(`SELECT id, project_id, object_id, cbs_chapter_id, code, name, description, unit, quantity, unit_price, total_cost, currency, contractor_id, contract_id, funding_source, phase, status, notes, created_at, updated_at FROM boq_items WHERE id = $1`, id).
		Scan(&i.ID, &i.ProjectID, &i.ObjectID, &i.CBSChapterID, &i.Code, &i.Name, &i.Description, &i.Unit, &i.Quantity, &i.UnitPrice, &i.TotalCost, &i.Currency, &i.ContractorID, &i.ContractID, &i.FundingSource, &i.Phase, &i.Status, &i.Notes, &i.CreatedAt, &i.UpdatedAt)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "BOQ item not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, i)
}

func (h *BOQHandler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Code         *string  `json:"code"`
		Name         *string  `json:"name"`
		Description  *string  `json:"description"`
		Unit         *string  `json:"unit"`
		Quantity     *float64 `json:"quantity"`
		UnitPrice    *float64 `json:"unit_price"`
		Currency     *string  `json:"currency"`
		ContractorID *string  `json:"contractor_id"`
		ContractID   *string  `json:"contract_id"`
		FundingSource *string `json:"funding_source"`
		Phase        *string  `json:"phase"`
		Status       *string  `json:"status"`
		Notes        *string  `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	now := time.Now()
	_, err := h.db.Exec(`UPDATE boq_items SET code=COALESCE($1,code), name=COALESCE($2,name), description=COALESCE($3,description), unit=COALESCE($4,unit), quantity=COALESCE($5,quantity), unit_price=COALESCE($6,unit_price), currency=COALESCE($7,currency), contractor_id=COALESCE($8,contractor_id), contract_id=COALESCE($9,contract_id), funding_source=COALESCE($10,funding_source), phase=COALESCE($11,phase), status=COALESCE($12,status), notes=COALESCE($13,notes), updated_at=$14 WHERE id=$15`,
		input.Code, input.Name, input.Description, input.Unit, input.Quantity, input.UnitPrice, input.Currency, input.ContractorID, input.ContractID, input.FundingSource, input.Phase, input.Status, input.Notes, now, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *BOQHandler) DeleteItem(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM boq_items WHERE id = $1`, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// --- Cost Transactions ---

func (h *BOQHandler) ListCostTransactions(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT id, project_id, boq_item_id, cbs_chapter_id, contractor_id, contract_id, transaction_type, amount, currency, exchange_rate, period, funding_source, description, reference_type, reference_id, created_by, created_at FROM cost_transactions`
	var rows *sql.Rows
	var err error
	if projectID != "" {
		rows, err = h.db.Query(query+` WHERE project_id = $1 ORDER BY period DESC`, projectID)
	} else {
		rows, err = h.db.Query(query + ` ORDER BY period DESC`)
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	txs := make([]models.CostTransaction, 0)
	for rows.Next() {
		var t models.CostTransaction
		if err := rows.Scan(&t.ID, &t.ProjectID, &t.BOQItemID, &t.CBSChapterID, &t.ContractorID, &t.ContractID, &t.TransactionType, &t.Amount, &t.Currency, &t.ExchangeRate, &t.Period, &t.FundingSource, &t.Description, &t.ReferenceType, &t.ReferenceID, &t.CreatedBy, &t.CreatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		txs = append(txs, t)
	}
	respondJSON(w, http.StatusOK, txs)
}

func (h *BOQHandler) CreateCostTransaction(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID       string  `json:"project_id"`
		BOQItemID       string  `json:"boq_item_id"`
		CBSChapterID    string  `json:"cbs_chapter_id"`
		ContractorID    *string `json:"contractor_id"`
		ContractID      *string `json:"contract_id"`
		TransactionType string  `json:"transaction_type"`
		Amount          float64 `json:"amount"`
		Currency        string  `json:"currency"`
		ExchangeRate    float64 `json:"exchange_rate"`
		Period          string  `json:"period"`
		FundingSource   *string `json:"funding_source"`
		Description     *string `json:"description"`
		ReferenceType   *string `json:"reference_type"`
		ReferenceID     *string `json:"reference_id"`
		CreatedBy       *string `json:"created_by"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO cost_transactions (id, project_id, boq_item_id, cbs_chapter_id, contractor_id, contract_id, transaction_type, amount, currency, exchange_rate, period, funding_source, description, reference_type, reference_id, created_by, created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)`,
		id, input.ProjectID, input.BOQItemID, input.CBSChapterID, input.ContractorID, input.ContractID, input.TransactionType, input.Amount, input.Currency, input.ExchangeRate, input.Period, input.FundingSource, input.Description, input.ReferenceType, input.ReferenceID, input.CreatedBy, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *BOQHandler) GetCostTransaction(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var t models.CostTransaction
	err := h.db.QueryRow(`SELECT id, project_id, boq_item_id, cbs_chapter_id, contractor_id, contract_id, transaction_type, amount, currency, exchange_rate, period, funding_source, description, reference_type, reference_id, created_by, created_at FROM cost_transactions WHERE id = $1`, id).
		Scan(&t.ID, &t.ProjectID, &t.BOQItemID, &t.CBSChapterID, &t.ContractorID, &t.ContractID, &t.TransactionType, &t.Amount, &t.Currency, &t.ExchangeRate, &t.Period, &t.FundingSource, &t.Description, &t.ReferenceType, &t.ReferenceID, &t.CreatedBy, &t.CreatedAt)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "cost transaction not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, t)
}

func (h *BOQHandler) UpdateCostTransaction(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Amount          *float64 `json:"amount"`
		Currency        *string  `json:"currency"`
		ExchangeRate    *float64 `json:"exchange_rate"`
		Description     *string  `json:"description"`
		FundingSource   *string  `json:"funding_source"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	_, err := h.db.Exec(`UPDATE cost_transactions SET amount=COALESCE($1,amount), currency=COALESCE($2,currency), exchange_rate=COALESCE($3,exchange_rate), description=COALESCE($4,description), funding_source=COALESCE($5,funding_source) WHERE id=$6`,
		input.Amount, input.Currency, input.ExchangeRate, input.Description, input.FundingSource, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *BOQHandler) DeleteCostTransaction(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM cost_transactions WHERE id = $1`, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// --- Budget Versions ---

func (h *BOQHandler) ListBudgetVersions(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT id, project_id, version_number, version_name, status, total_amount, approved_by, approved_at, notes, created_by, created_at FROM budget_versions`
	var rows *sql.Rows
	var err error
	if projectID != "" {
		rows, err = h.db.Query(query+` WHERE project_id = $1 ORDER BY version_number DESC`, projectID)
	} else {
		rows, err = h.db.Query(query + ` ORDER BY version_number DESC`)
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	versions := make([]models.BudgetVersion, 0)
	for rows.Next() {
		var v models.BudgetVersion
		if err := rows.Scan(&v.ID, &v.ProjectID, &v.VersionNumber, &v.VersionName, &v.Status, &v.TotalAmount, &v.ApprovedBy, &v.ApprovedAt, &v.Notes, &v.CreatedBy, &v.CreatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		versions = append(versions, v)
	}
	respondJSON(w, http.StatusOK, versions)
}

func (h *BOQHandler) CreateBudgetVersion(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID     string  `json:"project_id"`
		VersionNumber int     `json:"version_number"`
		VersionName   *string `json:"version_name"`
		Status        string  `json:"status"`
		TotalAmount   *float64 `json:"total_amount"`
		Notes         *string `json:"notes"`
		CreatedBy     *string `json:"created_by"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO budget_versions (id, project_id, version_number, version_name, status, total_amount, notes, created_by, created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		id, input.ProjectID, input.VersionNumber, input.VersionName, input.Status, input.TotalAmount, input.Notes, input.CreatedBy, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *BOQHandler) GetBudgetVersion(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var v models.BudgetVersion
	err := h.db.QueryRow(`SELECT id, project_id, version_number, version_name, status, total_amount, approved_by, approved_at, notes, created_by, created_at FROM budget_versions WHERE id = $1`, id).
		Scan(&v.ID, &v.ProjectID, &v.VersionNumber, &v.VersionName, &v.Status, &v.TotalAmount, &v.ApprovedBy, &v.ApprovedAt, &v.Notes, &v.CreatedBy, &v.CreatedAt)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "budget version not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, v)
}

func (h *BOQHandler) UpdateBudgetVersion(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Status      *string  `json:"status"`
		TotalAmount *float64 `json:"total_amount"`
		Notes       *string  `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	_, err := h.db.Exec(`UPDATE budget_versions SET status=COALESCE($1,status), total_amount=COALESCE($2,total_amount), notes=COALESCE($3,notes) WHERE id=$4`,
		input.Status, input.TotalAmount, input.Notes, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *BOQHandler) DeleteBudgetVersion(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM budget_versions WHERE id = $1`, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// --- Helpers ---

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(models.APIResponse{Success: status >= 200 && status < 300, Data: data})
}

func respondError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(models.APIResponse{Success: false, Error: message})
}

func init() {
	log.Println("BOQ handler initialized")
}
