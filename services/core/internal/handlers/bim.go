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

// BIMHandler handles BIM module endpoints
type BIMHandler struct {
	db *sql.DB
}

func NewBIMHandler(db *sql.DB) *BIMHandler {
	return &BIMHandler{db: db}
}

func (h *BIMHandler) RegisterRoutes(r chi.Router) {
	r.Route("/bim", func(r chi.Router) {
		// Models
		r.Get("/models", h.ListModels)
		r.Post("/models", h.CreateModel)
		r.Get("/models/{id}", h.GetModel)
		r.Put("/models/{id}", h.UpdateModel)
		r.Delete("/models/{id}", h.DeleteModel)

		// Elements
		r.Get("/elements", h.ListElements)
		r.Post("/elements", h.CreateElement)
		r.Get("/elements/{id}", h.GetElement)
		r.Put("/elements/{id}", h.UpdateElement)
		r.Delete("/elements/{id}", h.DeleteElement)

		// Clashes
		r.Get("/clashes", h.ListClashes)
		r.Post("/clashes", h.CreateClash)
		r.Get("/clashes/{id}", h.GetClash)
		r.Put("/clashes/{id}", h.UpdateClash)
		r.Delete("/clashes/{id}", h.DeleteClash)
	})
}

// --- BIM Models ---

func (h *BIMHandler) ListModels(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	discipline := r.URL.Query().Get("discipline")

	query := `SELECT id, project_id, model_name, model_version, description, discipline, author, software, file_format, file_path, file_size, ifc_schema, lod, status, checksum, is_latest, notes, uploaded_by, uploaded_at FROM bim_models WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if projectID != "" {
		query += ` AND project_id = $` + itoa(argIdx)
		args = append(args, projectID)
		argIdx++
	}
	if discipline != "" {
		query += ` AND discipline = $` + itoa(argIdx)
		args = append(args, discipline)
		argIdx++
	}
	query += ` ORDER BY uploaded_at DESC`

	rows, err := h.db.Query(query, args...)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	modelsList := make([]models.BIMModel, 0)
	for rows.Next() {
		var m models.BIMModel
		if err := rows.Scan(&m.ID, &m.ProjectID, &m.ModelName, &m.ModelVersion, &m.Description, &m.Discipline, &m.Author, &m.Software, &m.FileFormat, &m.FilePath, &m.FileSize, &m.IFCSchema, &m.LOD, &m.Status, &m.Checksum, &m.IsLatest, &m.Notes, &m.UploadedBy, &m.UploadedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		modelsList = append(modelsList, m)
	}
	respondJSON(w, http.StatusOK, modelsList)
}

func (h *BIMHandler) CreateModel(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID    string  `json:"project_id"`
		ModelName    string  `json:"model_name"`
		ModelVersion string  `json:"model_version"`
		Description  *string `json:"description"`
		Discipline   string  `json:"discipline"`
		Author       *string `json:"author"`
		Software     *string `json:"software"`
		FileFormat   *string `json:"file_format"`
		FilePath     *string `json:"file_path"`
		IFCSchema    *string `json:"ifc_schema"`
		LOD          *string `json:"lod"`
		Status       string  `json:"status"`
		Notes        *string `json:"notes"`
		UploadedBy   *string `json:"uploaded_by"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO bim_models (id, project_id, model_name, model_version, description, discipline, author, software, file_format, file_path, ifc_schema, lod, status, notes, uploaded_by, uploaded_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)`,
		id, input.ProjectID, input.ModelName, input.ModelVersion, input.Description, input.Discipline, input.Author, input.Software, input.FileFormat, input.FilePath, input.IFCSchema, input.LOD, input.Status, input.Notes, input.UploadedBy, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *BIMHandler) GetModel(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var m models.BIMModel
	err := h.db.QueryRow(`SELECT id, project_id, model_name, model_version, description, discipline, author, software, file_format, file_path, file_size, ifc_schema, lod, status, checksum, is_latest, notes, uploaded_by, uploaded_at FROM bim_models WHERE id = $1`, id).
		Scan(&m.ID, &m.ProjectID, &m.ModelName, &m.ModelVersion, &m.Description, &m.Discipline, &m.Author, &m.Software, &m.FileFormat, &m.FilePath, &m.FileSize, &m.IFCSchema, &m.LOD, &m.Status, &m.Checksum, &m.IsLatest, &m.Notes, &m.UploadedBy, &m.UploadedAt)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "BIM model not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, m)
}

func (h *BIMHandler) UpdateModel(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		ModelName    *string `json:"model_name"`
		ModelVersion *string `json:"model_version"`
		Status       *string `json:"status"`
		IsLatest     *bool   `json:"is_latest"`
		Notes        *string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	_, err := h.db.Exec(`UPDATE bim_models SET model_name=COALESCE($1,model_name), model_version=COALESCE($2,model_version), status=COALESCE($3,status), is_latest=COALESCE($4,is_latest), notes=COALESCE($5,notes) WHERE id=$6`,
		input.ModelName, input.ModelVersion, input.Status, input.IsLatest, input.Notes, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *BIMHandler) DeleteModel(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM bim_models WHERE id = $1`, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// --- BIM Elements ---

func (h *BIMHandler) ListElements(w http.ResponseWriter, r *http.Request) {
	modelID := r.URL.Query().Get("model_id")
	ifcType := r.URL.Query().Get("ifc_type")

	query := `SELECT id, model_id, ifc_global_id, ifc_type, ifc_class, name, description, level, material, volume, area, length, weight, elevation, x_position, y_position, z_position, properties, status, boq_item_id, created_at FROM bim_elements WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if modelID != "" {
		query += ` AND model_id = $` + itoa(argIdx)
		args = append(args, modelID)
		argIdx++
	}
	if ifcType != "" {
		query += ` AND ifc_type = $` + itoa(argIdx)
		args = append(args, ifcType)
		argIdx++
	}
	query += ` ORDER BY name`

	rows, err := h.db.Query(query, args...)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	elements := make([]models.BIMElement, 0)
	for rows.Next() {
		var e models.BIMElement
		if err := rows.Scan(&e.ID, &e.ModelID, &e.IFCGlobalID, &e.IFCType, &e.IFCClass, &e.Name, &e.Description, &e.Level, &e.Material, &e.Volume, &e.Area, &e.Length, &e.Weight, &e.Elevation, &e.XPosition, &e.YPosition, &e.ZPosition, &e.Properties, &e.Status, &e.BOQItemID, &e.CreatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		elements = append(elements, e)
	}
	respondJSON(w, http.StatusOK, elements)
}

func (h *BIMHandler) CreateElement(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ModelID     string   `json:"model_id"`
		IFCGlobalID *string  `json:"ifc_global_id"`
		IFCType     string   `json:"ifc_type"`
		IFCClass    *string  `json:"ifc_class"`
		Name        *string  `json:"name"`
		Description *string  `json:"description"`
		Level       *string  `json:"level"`
		Material    *string  `json:"material"`
		Volume      *float64 `json:"volume"`
		Area        *float64 `json:"area"`
		Length      *float64 `json:"length"`
		Weight      *float64 `json:"weight"`
		Elevation   *float64 `json:"elevation"`
		XPosition   *float64 `json:"x_position"`
		YPosition   *float64 `json:"y_position"`
		ZPosition   *float64 `json:"z_position"`
		Properties  *string  `json:"properties"`
		Status      string   `json:"status"`
		BOQItemID   *string  `json:"boq_item_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO bim_elements (id, model_id, ifc_global_id, ifc_type, ifc_class, name, description, level, material, volume, area, length, weight, elevation, x_position, y_position, z_position, properties, status, boq_item_id, created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21)`,
		id, input.ModelID, input.IFCGlobalID, input.IFCType, input.IFCClass, input.Name, input.Description, input.Level, input.Material, input.Volume, input.Area, input.Length, input.Weight, input.Elevation, input.XPosition, input.YPosition, input.ZPosition, input.Properties, input.Status, input.BOQItemID, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *BIMHandler) GetElement(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var e models.BIMElement
	err := h.db.QueryRow(`SELECT id, model_id, ifc_global_id, ifc_type, ifc_class, name, description, level, material, volume, area, length, weight, elevation, x_position, y_position, z_position, properties, status, boq_item_id, created_at FROM bim_elements WHERE id = $1`, id).
		Scan(&e.ID, &e.ModelID, &e.IFCGlobalID, &e.IFCType, &e.IFCClass, &e.Name, &e.Description, &e.Level, &e.Material, &e.Volume, &e.Area, &e.Length, &e.Weight, &e.Elevation, &e.XPosition, &e.YPosition, &e.ZPosition, &e.Properties, &e.Status, &e.BOQItemID, &e.CreatedAt)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "BIM element not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, e)
}

func (h *BIMHandler) UpdateElement(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Name       *string  `json:"name"`
		Material   *string  `json:"material"`
		Status     *string  `json:"status"`
		Properties *string  `json:"properties"`
		BOQItemID  *string  `json:"boq_item_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	_, err := h.db.Exec(`UPDATE bim_elements SET name=COALESCE($1,name), material=COALESCE($2,material), status=COALESCE($3,status), properties=COALESCE($4,properties), boq_item_id=COALESCE($5,boq_item_id) WHERE id=$6`,
		input.Name, input.Material, input.Status, input.Properties, input.BOQItemID, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *BIMHandler) DeleteElement(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM bim_elements WHERE id = $1`, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// --- BIM Clashes ---

func (h *BIMHandler) ListClashes(w http.ResponseWriter, r *http.Request) {
	modelID := r.URL.Query().Get("model_id")
	status := r.URL.Query().Get("status")
	severity := r.URL.Query().Get("severity")

	query := `SELECT id, model_id, clash_group, clash_type, severity, status, element_a_id, element_b_id, element_a_name, element_b_name, distance, tolerance, location_x, location_y, location_z, screenshot_path, assigned_to, resolution, resolved_by, resolved_at, created_at FROM bim_clashes WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if modelID != "" {
		query += ` AND model_id = $` + itoa(argIdx)
		args = append(args, modelID)
		argIdx++
	}
	if status != "" {
		query += ` AND status = $` + itoa(argIdx)
		args = append(args, status)
		argIdx++
	}
	if severity != "" {
		query += ` AND severity = $` + itoa(argIdx)
		args = append(args, severity)
		argIdx++
	}
	query += ` ORDER BY created_at DESC`

	rows, err := h.db.Query(query, args...)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	clashes := make([]models.BIMClash, 0)
	for rows.Next() {
		var c models.BIMClash
		if err := rows.Scan(&c.ID, &c.ModelID, &c.ClashGroup, &c.ClashType, &c.Severity, &c.Status, &c.ElementAID, &c.ElementBID, &c.ElementAName, &c.ElementBName, &c.Distance, &c.Tolerance, &c.LocationX, &c.LocationY, &c.LocationZ, &c.ScreenshotPath, &c.AssignedTo, &c.Resolution, &c.ResolvedBy, &c.ResolvedAt, &c.CreatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		clashes = append(clashes, c)
	}
	respondJSON(w, http.StatusOK, clashes)
}

func (h *BIMHandler) CreateClash(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ModelID       string   `json:"model_id"`
		ClashGroup    *string  `json:"clash_group"`
		ClashType     string   `json:"clash_type"`
		Severity      string   `json:"severity"`
		Status        string   `json:"status"`
		ElementAID    *string  `json:"element_a_id"`
		ElementBID    *string  `json:"element_b_id"`
		ElementAName  *string  `json:"element_a_name"`
		ElementBName  *string  `json:"element_b_name"`
		Distance      *float64 `json:"distance"`
		Tolerance     *float64 `json:"tolerance"`
		LocationX     *float64 `json:"location_x"`
		LocationY     *float64 `json:"location_y"`
		LocationZ     *float64 `json:"location_z"`
		ScreenshotPath *string `json:"screenshot_path"`
		AssignedTo    *string  `json:"assigned_to"`
		Resolution    *string  `json:"resolution"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO bim_clashes (id, model_id, clash_group, clash_type, severity, status, element_a_id, element_b_id, element_a_name, element_b_name, distance, tolerance, location_x, location_y, location_z, screenshot_path, assigned_to, resolution, created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)`,
		id, input.ModelID, input.ClashGroup, input.ClashType, input.Severity, input.Status, input.ElementAID, input.ElementBID, input.ElementAName, input.ElementBName, input.Distance, input.Tolerance, input.LocationX, input.LocationY, input.LocationZ, input.ScreenshotPath, input.AssignedTo, input.Resolution, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *BIMHandler) GetClash(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var c models.BIMClash
	err := h.db.QueryRow(`SELECT id, model_id, clash_group, clash_type, severity, status, element_a_id, element_b_id, element_a_name, element_b_name, distance, tolerance, location_x, location_y, location_z, screenshot_path, assigned_to, resolution, resolved_by, resolved_at, created_at FROM bim_clashes WHERE id = $1`, id).
		Scan(&c.ID, &c.ModelID, &c.ClashGroup, &c.ClashType, &c.Severity, &c.Status, &c.ElementAID, &c.ElementBID, &c.ElementAName, &c.ElementBName, &c.Distance, &c.Tolerance, &c.LocationX, &c.LocationY, &c.LocationZ, &c.ScreenshotPath, &c.AssignedTo, &c.Resolution, &c.ResolvedBy, &c.ResolvedAt, &c.CreatedAt)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "clash not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, c)
}

func (h *BIMHandler) UpdateClash(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Status     *string `json:"status"`
		AssignedTo *string `json:"assigned_to"`
		Resolution *string `json:"resolution"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	now := time.Now()
	_, err := h.db.Exec(`UPDATE bim_clashes SET status=COALESCE($1,status), assigned_to=COALESCE($2,assigned_to), resolution=COALESCE($3,resolution), resolved_at=CASE WHEN $1='resolved' THEN $4 ELSE resolved_at END WHERE id=$5`,
		input.Status, input.AssignedTo, input.Resolution, now, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *BIMHandler) DeleteClash(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM bim_clashes WHERE id = $1`, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
