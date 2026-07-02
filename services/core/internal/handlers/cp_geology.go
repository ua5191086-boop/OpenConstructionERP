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

type CrossPassageGeologyHandler struct{ db *sql.DB }
func NewCPGeoHandler(db *sql.DB) *CrossPassageGeologyHandler { return &CrossPassageGeologyHandler{db: db} }

func (h *CrossPassageGeologyHandler) RegisterRoutes(r chi.Router) {
	r.Route("/geology", func(r chi.Router) {
		// Cross Passages
		r.Get("/cross-passages", h.ListCP)
		r.Post("/cross-passages", h.CreateCP)
		r.Get("/cross-passages/{id}", h.GetCP)
		r.Put("/cross-passages/{id}", h.UpdateCP)
		r.Delete("/cross-passages/{id}", h.DeleteCP)
		r.Get("/cross-passages/{cpId}/stages", h.ListCPStages)
		r.Post("/cross-passages/{cpId}/stages", h.CreateCPStage)
		r.Get("/stages/{id}", h.GetCPStage)
		r.Put("/stages/{id}", h.UpdateCPStage)

		// Geology Units
		r.Get("/units", h.ListUnits)
		r.Post("/units", h.CreateUnit)
		r.Get("/units/{id}", h.GetUnit)
		r.Put("/units/{id}", h.UpdateUnit)
		r.Delete("/units/{id}", h.DeleteUnit)

		// Boreholes
		r.Get("/boreholes", h.ListBoreholes)
		r.Post("/boreholes", h.CreateBorehole)
		r.Get("/boreholes/{id}", h.GetBorehole)
		r.Put("/boreholes/{id}", h.UpdateBorehole)
		r.Delete("/boreholes/{id}", h.DeleteBorehole)
		r.Get("/boreholes/{bhId}/stratigraphy", h.ListStratigraphy)
		r.Post("/boreholes/{bhId}/stratigraphy", h.CreateStratigraphy)

		// Face Mapping
		r.Get("/face-maps", h.ListFaceMaps)
		r.Post("/face-maps", h.CreateFaceMap)
		r.Get("/face-maps/{id}", h.GetFaceMap)

		r.Get("/summary", h.GetSummary)
	})
}

// --- Cross Passages ---
func (h *CrossPassageGeologyHandler) ListCP(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	q := `SELECT id, project_id, cp_code, cp_name, chainage_m, span_m, height_m, length_m, excavation_method, lining_type, status, start_date, end_date, notes, created_at FROM cross_passages WHERE 1=1`
	var args []interface{}; argIdx := 1
	if projectID != "" { q += fmt.Sprintf(" AND project_id = $%d", argIdx); args = append(args, projectID) }
	q += " ORDER BY chainage_m"
	rows, err := h.db.Query(q, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()
	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, pid, code, name, method, lining, status, notes string
		var chain, span, hgt, length float64
		var sd, ed, crAt sql.NullString
		if err := rows.Scan(&id, &pid, &code, &name, &chain, &span, &hgt, &length, &method, &lining, &status, &sd, &ed, &notes, &crAt); err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
		item := map[string]interface{}{"id": id, "project_id": pid, "cp_code": code, "cp_name": name, "chainage_m": chain, "span_m": span, "height_m": hgt, "length_m": length, "excavation_method": method, "lining_type": lining, "status": status, "notes": notes}
		if sd.Valid { item["start_date"] = sd.String }; if ed.Valid { item["end_date"] = ed.String }; if crAt.Valid { item["created_at"] = crAt.String }
		items = append(items, item)
	}
	respondJSON(w, http.StatusOK, items)
}
func (h *CrossPassageGeologyHandler) CreateCP(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID   string  `json:"project_id"`
		CPCode      string  `json:"cp_code"`
		CPName      string  `json:"cp_name"`
		ChainageM   float64 `json:"chainage_m"`
		SpanM       *float64 `json:"span_m"`
		HeightM     *float64 `json:"height_m"`
		LengthM     *float64 `json:"length_m"`
		ExcMethod   string  `json:"excavation_method"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil { respondError(w, http.StatusBadRequest, "invalid body"); return }
	id := uuid.New().String()
	_, err := h.db.Exec(`INSERT INTO cross_passages (id, project_id, cp_code, cp_name, chainage_m, span_m, height_m, length_m, excavation_method, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$10)`, id, input.ProjectID, input.CPCode, input.CPName, input.ChainageM, input.SpanM, input.HeightM, input.LengthM, input.ExcMethod, time.Now())
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}
func (h *CrossPassageGeologyHandler) GetCP(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var code, name string
	err := h.db.QueryRow(`SELECT cp_code, cp_name FROM cross_passages WHERE id = $1`, id).Scan(&code, &name)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "CP not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "cp_code": code, "cp_name": name})
}
func (h *CrossPassageGeologyHandler) UpdateCP(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct{ Status *string `json:"status"` }
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil { respondError(w, http.StatusBadRequest, "invalid body"); return }
	_, err := h.db.Exec(`UPDATE cross_passages SET status=COALESCE($1,status), updated_at=$2 WHERE id=$3`, input.Status, time.Now(), id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}
func (h *CrossPassageGeologyHandler) DeleteCP(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM cross_passages WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
func (h *CrossPassageGeologyHandler) ListCPStages(w http.ResponseWriter, r *http.Request) {
	cpID := chi.URLParam(r, "cpId")
	rows, err := h.db.Query(`SELECT id, cp_id, stage_number, stage_name, stage_type, volume_m3, concrete_m3, reinforcement_kg, status, start_date, end_date, notes, created_at FROM cp_construction_stages WHERE cp_id = $1 ORDER BY stage_number`, cpID)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()
	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, cpid, name, stype, status, notes string
		var num int; var vol, conc, reinf float64; var sd, ed, crAt sql.NullString
		if err := rows.Scan(&id, &cpid, &num, &name, &stype, &vol, &conc, &reinf, &status, &sd, &ed, &notes, &crAt); err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
		item := map[string]interface{}{"id": id, "cp_id": cpid, "stage_number": num, "stage_name": name, "stage_type": stype, "volume_m3": vol, "concrete_m3": conc, "reinforcement_kg": reinf, "status": status, "notes": notes}
		if sd.Valid { item["start_date"] = sd.String }; if ed.Valid { item["end_date"] = ed.String }
		items = append(items, item)
	}
	respondJSON(w, http.StatusOK, items)
}
func (h *CrossPassageGeologyHandler) CreateCPStage(w http.ResponseWriter, r *http.Request) {
	cpID := chi.URLParam(r, "cpId")
	var input struct {
		StageName   string `json:"stage_name"`
		StageType   string `json:"stage_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil { respondError(w, http.StatusBadRequest, "invalid body"); return }
	id := uuid.New().String()
	_, err := h.db.Exec(`INSERT INTO cp_construction_stages (id, cp_id, stage_number, stage_name, stage_type, created_at, updated_at) VALUES ($1,$2,(SELECT COALESCE(MAX(stage_number),0)+1 FROM cp_construction_stages WHERE cp_id=$2),$3,$4,$5,$5)`, id, cpID, input.StageName, input.StageType, time.Now())
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}
func (h *CrossPassageGeologyHandler) GetCPStage(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var name string
	err := h.db.QueryRow(`SELECT stage_name FROM cp_construction_stages WHERE id = $1`, id).Scan(&name)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "stage not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "stage_name": name})
}
func (h *CrossPassageGeologyHandler) UpdateCPStage(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct{ Status *string `json:"status"` }
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil { respondError(w, http.StatusBadRequest, "invalid body"); return }
	_, err := h.db.Exec(`UPDATE cp_construction_stages SET status=COALESCE($1,status), updated_at=$2 WHERE id=$3`, input.Status, time.Now(), id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// --- Geology Units ---
func (h *CrossPassageGeologyHandler) ListUnits(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	q := `SELECT id, project_id, unit_code, unit_name, geology_type, soil_class, rock_class, description, density_kg_m3, cohesion_kPa, friction_angle, modulus_mpa, created_at FROM geology_units WHERE 1=1`
	var args []interface{}; argIdx := 1
	if projectID != "" { q += fmt.Sprintf(" AND project_id = $%d", argIdx); args = append(args, projectID) }
	q += " ORDER BY unit_code"
	rows, err := h.db.Query(q, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()
	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, pid, code, name, gtype, sclass, rclass, desc string
		var dens, coh, fric, mod float64
		var crAt time.Time
		if err := rows.Scan(&id, &pid, &code, &name, &gtype, &sclass, &rclass, &desc, &dens, &coh, &fric, &mod, &crAt); err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
		items = append(items, map[string]interface{}{"id": id, "project_id": pid, "unit_code": code, "unit_name": name, "geology_type": gtype, "soil_class": sclass, "rock_class": rclass, "description": desc, "density_kg_m3": dens, "cohesion_kPa": coh, "friction_angle": fric, "modulus_mpa": mod, "created_at": crAt})
	}
	respondJSON(w, http.StatusOK, items)
}
func (h *CrossPassageGeologyHandler) CreateUnit(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID string `json:"project_id"`
		UnitCode  string `json:"unit_code"`
		UnitName  string `json:"unit_name"`
		GeoType   string `json:"geology_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil { respondError(w, http.StatusBadRequest, "invalid body"); return }
	id := uuid.New().String()
	_, err := h.db.Exec(`INSERT INTO geology_units (id, project_id, unit_code, unit_name, geology_type, created_at) VALUES ($1,$2,$3,$4,$5,$6)`, id, input.ProjectID, input.UnitCode, input.UnitName, input.GeoType, time.Now())
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}
func (h *CrossPassageGeologyHandler) GetUnit(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var code, name string
	err := h.db.QueryRow(`SELECT unit_code, unit_name FROM geology_units WHERE id = $1`, id).Scan(&code, &name)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "unit not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "unit_code": code, "unit_name": name})
}
func (h *CrossPassageGeologyHandler) UpdateUnit(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct{ RockClass *string `json:"rock_class"` }
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil { respondError(w, http.StatusBadRequest, "invalid body"); return }
	_, err := h.db.Exec(`UPDATE geology_units SET rock_class=COALESCE($1,rock_class) WHERE id=$2`, input.RockClass, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}
func (h *CrossPassageGeologyHandler) DeleteUnit(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM geology_units WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// --- Boreholes ---
func (h *CrossPassageGeologyHandler) ListBoreholes(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	q := `SELECT id, project_id, borehole_code, location_name, chainage_m, ground_level_m, total_depth_m, water_table_m, drilling_date, status, notes, created_at FROM geology_boreholes WHERE 1=1`
	var args []interface{}; argIdx := 1
	if projectID != "" { q += fmt.Sprintf(" AND project_id = $%d", argIdx); args = append(args, projectID) }
	q += " ORDER BY chainage_m"
	rows, err := h.db.Query(q, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()
	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, pid, code, loc, status, notes string
		var chain, gl, depth, wt float64
		var dd, crAt sql.NullString
		if err := rows.Scan(&id, &pid, &code, &loc, &chain, &gl, &depth, &wt, &dd, &status, &notes, &crAt); err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
		item := map[string]interface{}{"id": id, "project_id": pid, "borehole_code": code, "location_name": loc, "chainage_m": chain, "ground_level_m": gl, "total_depth_m": depth, "water_table_m": wt, "status": status, "notes": notes}
		if dd.Valid { item["drilling_date"] = dd.String }; if crAt.Valid { item["created_at"] = crAt.String }
		items = append(items, item)
	}
	respondJSON(w, http.StatusOK, items)
}
func (h *CrossPassageGeologyHandler) CreateBorehole(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID   string `json:"project_id"`
		BoreholeCode string `json:"borehole_code"`
		ChainageM   float64 `json:"chainage_m"`
		GroundLevelM float64 `json:"ground_level_m"`
		TotalDepthM float64 `json:"total_depth_m"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil { respondError(w, http.StatusBadRequest, "invalid body"); return }
	id := uuid.New().String()
	_, err := h.db.Exec(`INSERT INTO geology_boreholes (id, project_id, borehole_code, chainage_m, ground_level_m, total_depth_m, status, created_at) VALUES ($1,$2,$3,$4,$5,$6,'planned',$7)`, id, input.ProjectID, input.BoreholeCode, input.ChainageM, input.GroundLevelM, input.TotalDepthM, time.Now())
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}
func (h *CrossPassageGeologyHandler) GetBorehole(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var code string
	err := h.db.QueryRow(`SELECT borehole_code FROM geology_boreholes WHERE id = $1`, id).Scan(&code)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "borehole not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "borehole_code": code})
}
func (h *CrossPassageGeologyHandler) UpdateBorehole(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct{ Status *string `json:"status"` }
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil { respondError(w, http.StatusBadRequest, "invalid body"); return }
	_, err := h.db.Exec(`UPDATE geology_boreholes SET status=COALESCE($1,status) WHERE id=$2`, input.Status, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}
func (h *CrossPassageGeologyHandler) DeleteBorehole(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM geology_boreholes WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
func (h *CrossPassageGeologyHandler) ListStratigraphy(w http.ResponseWriter, r *http.Request) {
	bhID := chi.URLParam(r, "bhId")
	rows, err := h.db.Query(`SELECT s.id, s.borehole_id, s.unit_id, u.unit_code, u.unit_name, s.depth_from_m, s.depth_to_m, s.thickness_m, s.description, s.spt_value, s.rqd_pct, s.created_at FROM geology_stratigraphy s LEFT JOIN geology_units u ON u.id = s.unit_id WHERE s.borehole_id = $1 ORDER BY s.depth_from_m`, bhID)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()
	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, bid, desc string
		var uid, ucode, uname sql.NullString
		var fromM, toM, thick float64
		var spt, rqd float64
		var crAt time.Time
		if err := rows.Scan(&id, &bid, &uid, &ucode, &uname, &fromM, &toM, &thick, &desc, &spt, &rqd, &crAt); err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
		items = append(items, map[string]interface{}{"id": id, "borehole_id": bid, "unit_id": uid, "unit_code": ucode, "unit_name": uname, "depth_from_m": fromM, "depth_to_m": toM, "thickness_m": thick, "description": desc, "spt_value": spt, "rqd_pct": rqd, "created_at": crAt})
	}
	respondJSON(w, http.StatusOK, items)
}
func (h *CrossPassageGeologyHandler) CreateStratigraphy(w http.ResponseWriter, r *http.Request) {
	bhID := chi.URLParam(r, "bhId")
	var input struct {
		UnitID    string  `json:"unit_id"`
		FromM     float64 `json:"depth_from_m"`
		ToM       float64 `json:"depth_to_m"`
		Desc      string  `json:"description"`
		SPT       float64 `json:"spt_value"`
		RQD       float64 `json:"rqd_pct"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil { respondError(w, http.StatusBadRequest, "invalid body"); return }
	id := uuid.New().String()
	_, err := h.db.Exec(`INSERT INTO geology_stratigraphy (id, borehole_id, unit_id, depth_from_m, depth_to_m, description, spt_value, rqd_pct, created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`, id, bhID, input.UnitID, input.FromM, input.ToM, input.Desc, input.SPT, input.RQD, time.Now())
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

// --- Face Mapping ---
func (h *CrossPassageGeologyHandler) ListFaceMaps(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	q := `SELECT id, project_id, mapping_date, chainage_from_m, chainage_to_m, unit_id, rock_class, weathering, fracture_count, water_inflow, mapping_by, notes, created_at FROM geology_face_mapping WHERE 1=1`
	var args []interface{}; argIdx := 1
	if projectID != "" { q += fmt.Sprintf(" AND project_id = $%d", argIdx); args = append(args, projectID) }
	q += " ORDER BY chainage_from_m"
	rows, err := h.db.Query(q, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()
	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, pid, rclass, weather, inflow, mapper, notes string
		var md time.Time
		var fromM, toM float64
		var unitID, fractures sql.NullString
		var crAt time.Time
		if err := rows.Scan(&id, &pid, &md, &fromM, &toM, &unitID, &rclass, &weather, &fractures, &inflow, &mapper, &notes, &crAt); err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
		item := map[string]interface{}{"id": id, "project_id": pid, "mapping_date": md, "chainage_from_m": fromM, "chainage_to_m": toM, "rock_class": rclass, "weathering": weather, "water_inflow": inflow, "mapping_by": mapper, "notes": notes, "created_at": crAt}
		if unitID.Valid { item["unit_id"] = unitID.String }
		items = append(items, item)
	}
	respondJSON(w, http.StatusOK, items)
}
func (h *CrossPassageGeologyHandler) CreateFaceMap(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID   string  `json:"project_id"`
		MappingDate string  `json:"mapping_date"`
		FromM       float64 `json:"chainage_from_m"`
		ToM         *float64 `json:"chainage_to_m"`
		RockClass   string  `json:"rock_class"`
		Mapper      string  `json:"mapping_by"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil { respondError(w, http.StatusBadRequest, "invalid body"); return }
	id := uuid.New().String()
	_, err := h.db.Exec(`INSERT INTO geology_face_mapping (id, project_id, mapping_date, chainage_from_m, chainage_to_m, rock_class, mapping_by, created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`, id, input.ProjectID, input.MappingDate, input.FromM, input.ToM, input.RockClass, input.Mapper, time.Now())
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}
func (h *CrossPassageGeologyHandler) GetFaceMap(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var rclass string
	err := h.db.QueryRow(`SELECT rock_class FROM geology_face_mapping WHERE id = $1`, id).Scan(&rclass)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "face map not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "rock_class": rclass})
}

func (h *CrossPassageGeologyHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	q := `SELECT project_id, boreholes, geology_units, face_mappings, cross_passages, active_cp_stages FROM geology_summary`
	var args []interface{}; argIdx := 1
	if projectID != "" { q += fmt.Sprintf(" WHERE project_id = $%d", argIdx); args = append(args, projectID) }
	rows, err := h.db.Query(q, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()
	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var pid string
		var bh, gu, fm, cp, cpStage int
		if err := rows.Scan(&pid, &bh, &gu, &fm, &cp, &cpStage); err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
		items = append(items, map[string]interface{}{"project_id": pid, "boreholes": bh, "geology_units": gu, "face_mappings": fm, "cross_passages": cp, "active_cp_stages": cpStage})
	}
	respondJSON(w, http.StatusOK, items)
}