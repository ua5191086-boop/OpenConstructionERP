package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// GISHandler handles GIS & Survey module endpoints
type GISHandler struct {
	db *sql.DB
}

func NewGISHandler(db *sql.DB) *GISHandler {
	return &GISHandler{db: db}
}

func (h *GISHandler) RegisterRoutes(r chi.Router) {
	r.Route("/gis", func(r chi.Router) {
		r.Get("/layers", h.ListLayers)
		r.Post("/layers", h.CreateLayer)
		r.Get("/layers/{id}", h.GetLayer)
		r.Put("/layers/{id}", h.UpdateLayer)
		r.Delete("/layers/{id}", h.DeleteLayer)

		r.Get("/features", h.ListFeatures)
		r.Post("/features", h.CreateFeature)
		r.Get("/features/{id}", h.GetFeature)
		r.Put("/features/{id}", h.UpdateFeature)
		r.Delete("/features/{id}", h.DeleteFeature)

		r.Get("/survey-points", h.ListSurveyPoints)
		r.Post("/survey-points", h.CreateSurveyPoint)
		r.Get("/survey-points/{id}", h.GetSurveyPoint)
		r.Put("/survey-points/{id}", h.UpdateSurveyPoint)
		r.Delete("/survey-points/{id}", h.DeleteSurveyPoint)

		r.Get("/survey-runs", h.ListSurveyRuns)
		r.Post("/survey-runs", h.CreateSurveyRun)
		r.Get("/survey-runs/{id}", h.GetSurveyRun)
		r.Put("/survey-runs/{id}", h.UpdateSurveyRun)
		r.Delete("/survey-runs/{id}", h.DeleteSurveyRun)

		r.Get("/survey-stations", h.ListSurveyStations)
		r.Post("/survey-stations", h.CreateSurveyStation)
		r.Get("/survey-stations/{id}", h.GetSurveyStation)
		r.Put("/survey-stations/{id}", h.UpdateSurveyStation)
		r.Delete("/survey-stations/{id}", h.DeleteSurveyStation)

		r.Get("/alignments", h.ListAlignments)
		r.Post("/alignments", h.CreateAlignment)
		r.Get("/alignments/{id}", h.GetAlignment)
		r.Put("/alignments/{id}", h.UpdateAlignment)
		r.Delete("/alignments/{id}", h.DeleteAlignment)

		r.Get("/cross-sections", h.ListCrossSections)
		r.Post("/cross-sections", h.CreateCrossSection)
		r.Get("/cross-sections/{id}", h.GetCrossSection)
		r.Put("/cross-sections/{id}", h.UpdateCrossSection)
		r.Delete("/cross-sections/{id}", h.DeleteCrossSection)

		r.Get("/drone-flights", h.ListDroneFlights)
		r.Post("/drone-flights", h.CreateDroneFlight)
		r.Get("/drone-flights/{id}", h.GetDroneFlight)
		r.Put("/drone-flights/{id}", h.UpdateDroneFlight)
		r.Delete("/drone-flights/{id}", h.DeleteDroneFlight)

		r.Get("/summary", h.GetSummary)
	})
}

// =============================================================================
// Layers
// =============================================================================
func (h *GISHandler) ListLayers(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT id, project_id, layer_name, layer_type, geometry_type, source_type, is_visible, status, created_at FROM gis_layers WHERE 1=1`
	args := []interface{}{}
	argIdx := 1
	if projectID != "" { query += fmt.Sprintf(" AND project_id = $%d", argIdx); argIdx++; args = append(args, projectID) }
	query += " ORDER BY layer_name"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, pid, name, ltype, gtype, src, st string
		var visible bool
		var createdAt time.Time
		if err := rows.Scan(&id, &pid, &name, &ltype, &gtype, &src, &visible, &st, &createdAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		items = append(items, map[string]interface{}{
			"id": id, "project_id": pid, "layer_name": name, "layer_type": ltype,
			"geometry_type": gtype, "source_type": src, "is_visible": visible, "status": st, "created_at": createdAt,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *GISHandler) CreateLayer(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID    string `json:"project_id"`
		LayerName    string `json:"layer_name"`
		LayerType    string `json:"layer_type"`
		GeometryType string `json:"geometry_type"`
		SourceType   string `json:"source_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO gis_layers (id, project_id, layer_name, layer_type, geometry_type, source_type, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
		id, input.ProjectID, input.LayerName, input.LayerType, input.GeometryType, input.SourceType, now, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *GISHandler) GetLayer(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var name string
	err := h.db.QueryRow(`SELECT layer_name FROM gis_layers WHERE id = $1`, id).Scan(&name)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "layer not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "layer_name": name})
}

func (h *GISHandler) UpdateLayer(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Status    *string `json:"status"`
		Visible   *bool   `json:"is_visible"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	_, err := h.db.Exec(`UPDATE gis_layers SET status=COALESCE($1,status), is_visible=COALESCE($2,is_visible), updated_at=$3 WHERE id=$4`, input.Status, input.Visible, time.Now(), id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *GISHandler) DeleteLayer(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM gis_layers WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// Features
// =============================================================================
func (h *GISHandler) ListFeatures(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT id, project_id, feature_name, feature_type, source, status, created_at FROM gis_features WHERE 1=1`
	args := []interface{}{}
	argIdx := 1
	if projectID != "" { query += fmt.Sprintf(" AND project_id = $%d", argIdx); args = append(args, projectID) }
	query += " ORDER BY feature_name"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, pid, fname, ftype, src, st string
		var createdAt time.Time
		if err := rows.Scan(&id, &pid, &fname, &ftype, &src, &st, &createdAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		items = append(items, map[string]interface{}{
			"id": id, "project_id": pid, "feature_name": fname, "feature_type": ftype,
			"source": src, "status": st, "created_at": createdAt,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *GISHandler) CreateFeature(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID   string `json:"project_id"`
		FeatureName string `json:"feature_name"`
		FeatureType string `json:"feature_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO gis_features (id, project_id, feature_name, feature_type, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6)`,
		id, input.ProjectID, input.FeatureName, input.FeatureType, now, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *GISHandler) GetFeature(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var name string
	err := h.db.QueryRow(`SELECT feature_name FROM gis_features WHERE id = $1`, id).Scan(&name)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "feature not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "feature_name": name})
}

func (h *GISHandler) UpdateFeature(w http.ResponseWriter, r *http.Request) {
	// Simple status update
	id := chi.URLParam(r, "id")
	var input struct {
		Status *string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	_, err := h.db.Exec(`UPDATE gis_features SET status=COALESCE($1,status), updated_at=$2 WHERE id=$3`, input.Status, time.Now(), id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *GISHandler) DeleteFeature(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM gis_features WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// Survey Points
// =============================================================================
func (h *GISHandler) ListSurveyPoints(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT id, project_id, point_number, point_code, point_name, point_type, latitude, longitude, elevation, northing, easting, zone, accuracy_mm, method, survey_date, status FROM gis_survey_points WHERE 1=1`
	args := []interface{}{}
	argIdx := 1
	if projectID != "" { query += fmt.Sprintf(" AND project_id = $%d", argIdx); args = append(args, projectID) }
	query += " ORDER BY point_number"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, pid, code, pname, ptype, zone, method, st string
		var num int
		var lat, lng, elev, north, east, accMM float64
		var surveyDate sql.NullString
		if err := rows.Scan(&id, &pid, &num, &code, &pname, &ptype, &lat, &lng, &elev, &north, &east, &zone, &accMM, &method, &surveyDate, &st); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		item := map[string]interface{}{
			"id": id, "project_id": pid, "point_number": num, "point_code": code,
			"point_name": pname, "point_type": ptype, "latitude": lat, "longitude": lng,
			"elevation": elev, "northing": north, "easting": east, "zone": zone,
			"accuracy_mm": accMM, "method": method, "status": st,
		}
		if surveyDate.Valid { item["survey_date"] = surveyDate.String }
		items = append(items, item)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *GISHandler) CreateSurveyPoint(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID string  `json:"project_id"`
		PointCode string  `json:"point_code"`
		PointName string  `json:"point_name"`
		PointType string  `json:"point_type"`
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
		Elevation float64 `json:"elevation"`
		Method    string  `json:"method"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO gis_survey_points (id, project_id, point_number, point_code, point_name, point_type, latitude, longitude, elevation, method, created_at, updated_at) VALUES ($1,$2,(SELECT COALESCE(MAX(point_number),0)+1 FROM gis_survey_points WHERE project_id=$2),$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		id, input.ProjectID, input.PointCode, input.PointName, input.PointType, input.Latitude, input.Longitude, input.Elevation, input.Method, now, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *GISHandler) GetSurveyPoint(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var code, pname string
	err := h.db.QueryRow(`SELECT point_code, point_name FROM gis_survey_points WHERE id = $1`, id).Scan(&code, &pname)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "survey point not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "point_code": code, "point_name": pname})
}

func (h *GISHandler) UpdateSurveyPoint(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Status *string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	_, err := h.db.Exec(`UPDATE gis_survey_points SET status=COALESCE($1,status), updated_at=$2 WHERE id=$3`, input.Status, time.Now(), id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *GISHandler) DeleteSurveyPoint(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM gis_survey_points WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// Survey Runs
// =============================================================================
func (h *GISHandler) ListSurveyRuns(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT id, project_id, run_number, run_code, run_name, survey_type, start_date, end_date, instrument, crew_lead, point_count, status FROM gis_survey_runs WHERE 1=1`
	args := []interface{}{}
	argIdx := 1
	if projectID != "" { query += fmt.Sprintf(" AND project_id = $%d", argIdx); args = append(args, projectID) }
	query += " ORDER BY run_number DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, pid, code, rname, stype, instr, crew, st string
		var num, ptCount int
		var startDate, endDate sql.NullString
		if err := rows.Scan(&id, &pid, &num, &code, &rname, &stype, &startDate, &endDate, &instr, &crew, &ptCount, &st); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		item := map[string]interface{}{
			"id": id, "project_id": pid, "run_number": num, "run_code": code,
			"run_name": rname, "survey_type": stype, "instrument": instr,
			"crew_lead": crew, "point_count": ptCount, "status": st,
		}
		if startDate.Valid { item["start_date"] = startDate.String }
		if endDate.Valid { item["end_date"] = endDate.String }
		items = append(items, item)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *GISHandler) CreateSurveyRun(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID  string `json:"project_id"`
		RunCode    string `json:"run_code"`
		RunName    string `json:"run_name"`
		SurveyType string `json:"survey_type"`
		StartDate  string `json:"start_date"`
		Instrument string `json:"instrument"`
		CrewLead   string `json:"crew_lead"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO gis_survey_runs (id, project_id, run_number, run_code, run_name, survey_type, start_date, instrument, crew_lead, created_at, updated_at) VALUES ($1,$2,(SELECT COALESCE(MAX(run_number),0)+1 FROM gis_survey_runs WHERE project_id=$2),$3,$4,$5,$6,$7,$8,$9,$10)`,
		id, input.ProjectID, input.RunCode, input.RunName, input.SurveyType, input.StartDate, input.Instrument, input.CrewLead, now, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *GISHandler) GetSurveyRun(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var code, name string
	err := h.db.QueryRow(`SELECT run_code, run_name FROM gis_survey_runs WHERE id = $1`, id).Scan(&code, &name)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "survey run not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "run_code": code, "run_name": name})
}

func (h *GISHandler) UpdateSurveyRun(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Status *string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	_, err := h.db.Exec(`UPDATE gis_survey_runs SET status=COALESCE($1,status), updated_at=$2 WHERE id=$3`, input.Status, time.Now(), id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *GISHandler) DeleteSurveyRun(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM gis_survey_runs WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// Survey Stations
// =============================================================================
func (h *GISHandler) ListSurveyStations(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT id, project_id, station_number, station_code, station_name, station_type, northing, easting, elevation FROM gis_survey_stations WHERE 1=1`
	args := []interface{}{}
	argIdx := 1
	if projectID != "" { query += fmt.Sprintf(" AND project_id = $%d", argIdx); args = append(args, projectID) }
	query += " ORDER BY station_number"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, pid, code, sname, stype string
		var num int
		var north, east, elev float64
		if err := rows.Scan(&id, &pid, &num, &code, &sname, &stype, &north, &east, &elev); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		items = append(items, map[string]interface{}{
			"id": id, "project_id": pid, "station_number": num, "station_code": code,
			"station_name": sname, "station_type": stype, "northing": north,
			"easting": east, "elevation": elev,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *GISHandler) CreateSurveyStation(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID    string  `json:"project_id"`
		StationCode  string  `json:"station_code"`
		StationName  string  `json:"station_name"`
		StationType  string  `json:"station_type"`
		Northing     float64 `json:"northing"`
		Easting      float64 `json:"easting"`
		Elevation    float64 `json:"elevation"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO gis_survey_stations (id, project_id, station_number, station_code, station_name, station_type, northing, easting, elevation, created_at, updated_at) VALUES ($1,$2,(SELECT COALESCE(MAX(station_number),0)+1 FROM gis_survey_stations WHERE project_id=$2),$3,$4,$5,$6,$7,$8,$9,$10)`,
		id, input.ProjectID, input.StationCode, input.StationName, input.StationType, input.Northing, input.Easting, input.Elevation, now, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *GISHandler) GetSurveyStation(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var code, name string
	err := h.db.QueryRow(`SELECT station_code, station_name FROM gis_survey_stations WHERE id = $1`, id).Scan(&code, &name)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "station not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "station_code": code, "station_name": name})
}

func (h *GISHandler) UpdateSurveyStation(w http.ResponseWriter, r *http.Request) {
	// Update coordinates or notes
	id := chi.URLParam(r, "id")
	var input struct {
		Northing  *float64 `json:"northing"`
		Easting   *float64 `json:"easting"`
		Elevation *float64 `json:"elevation"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	_, err := h.db.Exec(`UPDATE gis_survey_stations SET northing=COALESCE($1,northing), easting=COALESCE($2,easting), elevation=COALESCE($3,elevation), updated_at=$4 WHERE id=$5`,
		input.Northing, input.Easting, input.Elevation, time.Now(), id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *GISHandler) DeleteSurveyStation(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM gis_survey_stations WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// Alignments
// =============================================================================
func (h *GISHandler) ListAlignments(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT id, project_id, alignment_code, alignment_name, alignment_type, start_chainage, end_chainage, total_length, status FROM gis_alignments WHERE 1=1`
	args := []interface{}{}
	argIdx := 1
	if projectID != "" { query += fmt.Sprintf(" AND project_id = $%d", argIdx); args = append(args, projectID) }
	query += " ORDER BY alignment_name"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, pid, code, aname, atype, st string
		var startCh, endCh, length float64
		if err := rows.Scan(&id, &pid, &code, &aname, &atype, &startCh, &endCh, &length, &st); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		items = append(items, map[string]interface{}{
			"id": id, "project_id": pid, "alignment_code": code, "alignment_name": aname,
			"alignment_type": atype, "start_chainage": startCh, "end_chainage": endCh,
			"total_length": length, "status": st,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *GISHandler) CreateAlignment(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID     string  `json:"project_id"`
		AlignmentCode string  `json:"alignment_code"`
		AlignmentName string  `json:"alignment_name"`
		AlignmentType string  `json:"alignment_type"`
		TotalLength   float64 `json:"total_length"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO gis_alignments (id, project_id, alignment_code, alignment_name, alignment_type, total_length, end_chainage, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		id, input.ProjectID, input.AlignmentCode, input.AlignmentName, input.AlignmentType, input.TotalLength, input.TotalLength, now, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *GISHandler) GetAlignment(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var code, name string
	err := h.db.QueryRow(`SELECT alignment_code, alignment_name FROM gis_alignments WHERE id = $1`, id).Scan(&code, &name)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "alignment not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "alignment_code": code, "alignment_name": name})
}

func (h *GISHandler) UpdateAlignment(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Status *string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	_, err := h.db.Exec(`UPDATE gis_alignments SET status=COALESCE($1,status), updated_at=$2 WHERE id=$3`, input.Status, time.Now(), id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *GISHandler) DeleteAlignment(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM gis_alignments WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// Cross Sections
// =============================================================================
func (h *GISHandler) ListCrossSections(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT id, project_id, section_number, chainage, offset_left, offset_right, cut_area, fill_area, total_area, source FROM gis_cross_sections WHERE 1=1`
	args := []interface{}{}
	argIdx := 1
	if projectID != "" { query += fmt.Sprintf(" AND project_id = $%d", argIdx); args = append(args, projectID) }
	query += " ORDER BY chainage"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, pid, src string
		var secNum int
		var ch, offL, offR, cut, fill, total float64
		if err := rows.Scan(&id, &pid, &secNum, &ch, &offL, &offR, &cut, &fill, &total, &src); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		items = append(items, map[string]interface{}{
			"id": id, "project_id": pid, "section_number": secNum, "chainage": ch,
			"offset_left": offL, "offset_right": offR, "cut_area": cut,
			"fill_area": fill, "total_area": total, "source": src,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *GISHandler) CreateCrossSection(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID   string  `json:"project_id"`
		Chainage    float64 `json:"chainage"`
		CutArea     float64 `json:"cut_area"`
		FillArea    float64 `json:"fill_area"`
		Source      string  `json:"source"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO gis_cross_sections (id, project_id, section_number, chainage, cut_area, fill_area, source, created_at, updated_at) VALUES ($1,$2,(SELECT COALESCE(MAX(section_number),0)+1 FROM gis_cross_sections WHERE project_id=$2),$3,$4,$5,$6,$7,$8)`,
		id, input.ProjectID, input.Chainage, input.CutArea, input.FillArea, input.Source, now, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *GISHandler) GetCrossSection(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var secNum int
	var chainage float64
	err := h.db.QueryRow(`SELECT section_number, chainage FROM gis_cross_sections WHERE id = $1`, id).Scan(&secNum, &chainage)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "cross section not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "section_number": secNum, "chainage": chainage})
}

func (h *GISHandler) UpdateCrossSection(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		CutArea  *float64 `json:"cut_area"`
		FillArea *float64 `json:"fill_area"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	_, err := h.db.Exec(`UPDATE gis_cross_sections SET cut_area=COALESCE($1,cut_area), fill_area=COALESCE($2,fill_area), updated_at=$3 WHERE id=$4`, input.CutArea, input.FillArea, time.Now(), id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *GISHandler) DeleteCrossSection(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM gis_cross_sections WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// Drone Flights
// =============================================================================
func (h *GISHandler) ListDroneFlights(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT id, project_id, flight_number, flight_code, flight_name, drone_model, pilot, flight_date, flight_duration_minutes, altitude_m, area_covered_ha, gsd_cm, images_count, processing_status, sensor_type, output_type, status FROM gis_drone_flights WHERE 1=1`
	args := []interface{}{}
	argIdx := 1
	if projectID != "" { query += fmt.Sprintf(" AND project_id = $%d", argIdx); args = append(args, projectID) }
	query += " ORDER BY flight_date DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, pid, code, fname, drone, pilot, procStatus, sensorType, outputType, st string
		var num, durMin, imgCount int
		var alt, areaHa, gsd float64
		var flightDate sql.NullString
		if err := rows.Scan(&id, &pid, &num, &code, &fname, &drone, &pilot, &flightDate, &durMin, &alt, &areaHa, &gsd, &imgCount, &procStatus, &sensorType, &outputType, &st); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		item := map[string]interface{}{
			"id": id, "project_id": pid, "flight_number": num, "flight_code": code,
			"flight_name": fname, "drone_model": drone, "pilot": pilot,
			"flight_duration_minutes": durMin, "altitude_m": alt, "area_covered_ha": areaHa,
			"gsd_cm": gsd, "images_count": imgCount, "processing_status": procStatus,
			"sensor_type": sensorType, "output_type": outputType, "status": st,
		}
		if flightDate.Valid { item["flight_date"] = flightDate.String }
		items = append(items, item)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *GISHandler) CreateDroneFlight(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID     string `json:"project_id"`
		FlightCode    string `json:"flight_code"`
		FlightName    string `json:"flight_name"`
		DroneModel    string `json:"drone_model"`
		Pilot         string `json:"pilot"`
		FlightDate    string `json:"flight_date"`
		SensorType    string `json:"sensor_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO gis_drone_flights (id, project_id, flight_number, flight_code, flight_name, drone_model, pilot, flight_date, sensor_type, created_at, updated_at) VALUES ($1,$2,(SELECT COALESCE(MAX(flight_number),0)+1 FROM gis_drone_flights WHERE project_id=$2),$3,$4,$5,$6,$7,$8,$9,$10)`,
		id, input.ProjectID, input.FlightCode, input.FlightName, input.DroneModel, input.Pilot, input.FlightDate, input.SensorType, now, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *GISHandler) GetDroneFlight(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var code, name string
	err := h.db.QueryRow(`SELECT flight_code, flight_name FROM gis_drone_flights WHERE id = $1`, id).Scan(&code, &name)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "drone flight not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "flight_code": code, "flight_name": name})
}

func (h *GISHandler) UpdateDroneFlight(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		ProcessingStatus *string `json:"processing_status"`
		Status           *string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	_, err := h.db.Exec(`UPDATE gis_drone_flights SET processing_status=COALESCE($1,processing_status), status=COALESCE($2,status), updated_at=$3 WHERE id=$4`, input.ProcessingStatus, input.Status, time.Now(), id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *GISHandler) DeleteDroneFlight(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM gis_drone_flights WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// Summary
// =============================================================================
func (h *GISHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT project_id, active_layers, active_features, survey_points, completed_surveys, pending_surveys, active_alignments, cross_sections, completed_flights, processing_flights FROM gis_summary`
	args := []interface{}{}
	argIdx := 1
	if projectID != "" { query += fmt.Sprintf(" WHERE project_id = $%d", argIdx); args = append(args, projectID) }

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var pid string
		var actLayers, actFeatures, survPts, compSurveys, pendSurveys, actAlign, crossSec, compFlights, procFlights int
		if err := rows.Scan(&pid, &actLayers, &actFeatures, &survPts, &compSurveys, &pendSurveys, &actAlign, &crossSec, &compFlights, &procFlights); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		items = append(items, map[string]interface{}{
			"project_id": pid, "active_layers": actLayers, "active_features": actFeatures,
			"survey_points": survPts, "completed_surveys": compSurveys, "pending_surveys": pendSurveys,
			"active_alignments": actAlign, "cross_sections": crossSec,
			"completed_flights": compFlights, "processing_flights": procFlights,
		})
	}
	respondJSON(w, http.StatusOK, items)
}