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

// RingBuilderHandler handles Ring Builder & Segment Tracking module endpoints
type RingBuilderHandler struct {
	db *sql.DB
}

func NewRingBuilderHandler(db *sql.DB) *RingBuilderHandler {
	return &RingBuilderHandler{db: db}
}

func (h *RingBuilderHandler) RegisterRoutes(r chi.Router) {
	r.Route("/ringbuilder", func(r chi.Router) {
		// Ring Designs
		r.Get("/designs", h.ListDesigns)
		r.Post("/designs", h.CreateDesign)
		r.Get("/designs/{id}", h.GetDesign)
		r.Put("/designs/{id}", h.UpdateDesign)
		r.Delete("/designs/{id}", h.DeleteDesign)

		// Segment Production
		r.Get("/production", h.ListProduction)
		r.Post("/production", h.CreateProduction)
		r.Get("/production/{id}", h.GetProduction)
		r.Put("/production/{id}", h.UpdateProduction)

		// Segment Curing
		r.Get("/curing", h.ListCuring)
		r.Post("/curing", h.CreateCuring)

		// Segment Transport
		r.Get("/transport", h.ListTransport)
		r.Post("/transport", h.CreateTransport)

		// Segment Installation
		r.Get("/installation", h.ListInstallation)
		r.Post("/installation", h.CreateInstallation)

		// Segment QC
		r.Get("/qc", h.ListQC)
		r.Post("/qc", h.CreateQC)
		r.Get("/qc/{id}", h.GetQC)

		// Segment Inventory
		r.Get("/inventory", h.ListInventory)

		// Ring Measurements
		r.Get("/measurements", h.ListMeasurements)
		r.Post("/measurements", h.CreateMeasurement)

		r.Get("/summary", h.GetSummary)
	})
}

// =============================================================================
// Ring Designs
// =============================================================================
func (h *RingBuilderHandler) ListDesigns(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT id, project_id, design_code, design_name, ring_type, inner_diameter_mm, outer_diameter_mm, ring_width_mm, taper_mm, segment_count, key_position, concrete_grade, reinforcement_type, weight_kg, status, created_at FROM ring_designs WHERE 1=1`
	args := []interface{}{}
	argIdx := 1
	if projectID != "" { query += fmt.Sprintf(" AND project_id = $%d", argIdx); argIdx++; args = append(args, projectID) }
	query += " ORDER BY design_code"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, pid, code, name, rtype, kpos, concrete, rebar, status string
		var innerD, outerD, width, taper int
		var segCount int
		var weight sql.NullFloat64
		var createdAt time.Time
		if err := rows.Scan(&id, &pid, &code, &name, &rtype, &innerD, &outerD, &width, &taper, &segCount, &kpos, &concrete, &rebar, &weight, &status, &createdAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		items = append(items, map[string]interface{}{
			"id": id, "project_id": pid, "design_code": code, "design_name": name,
			"ring_type": rtype, "inner_diameter_mm": innerD, "outer_diameter_mm": outerD,
			"ring_width_mm": width, "taper_mm": taper, "segment_count": segCount,
			"key_position": kpos, "concrete_grade": concrete, "reinforcement_type": rebar,
			"weight_kg": weight.Float64, "status": status, "created_at": createdAt,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *RingBuilderHandler) CreateDesign(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID  string `json:"project_id"`
		DesignCode string `json:"design_code"`
		DesignName string `json:"design_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	_, err := h.db.Exec(`INSERT INTO ring_designs (id, project_id, design_code, design_name) VALUES ($1,$2,$3,$4)`,
		id, input.ProjectID, input.DesignCode, input.DesignName)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *RingBuilderHandler) GetDesign(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var code, name string
	err := h.db.QueryRow(`SELECT design_code, design_name FROM ring_designs WHERE id = $1`, id).Scan(&code, &name)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "design not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "design_code": code, "design_name": name})
}

func (h *RingBuilderHandler) UpdateDesign(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Status *string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	_, err := h.db.Exec(`UPDATE ring_designs SET status=COALESCE($1,status), updated_at=NOW() WHERE id=$2`, input.Status, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *RingBuilderHandler) DeleteDesign(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM ring_designs WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// Segment Production
// =============================================================================
func (h *RingBuilderHandler) ListProduction(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	status := r.URL.Query().Get("status")
	query := `SELECT p.id, p.project_id, p.segment_code, p.segment_type, p.cast_batch, p.concrete_grade, p.concrete_volume_m3, p.steel_weight_kg, p.cast_at, p.status, p.qc_status, p.qr_code, p.location, d.design_code
		FROM segment_production p LEFT JOIN ring_designs d ON p.design_id = d.id WHERE 1=1`
	args := []interface{}{}
	argIdx := 1
	if projectID != "" { query += fmt.Sprintf(" AND p.project_id = $%d", argIdx); argIdx++; args = append(args, projectID) }
	if status != "" { query += fmt.Sprintf(" AND p.status = $%d", argIdx); argIdx++; args = append(args, status) }
	query += " ORDER BY p.cast_at DESC LIMIT 200"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, pid, code, stype, batch, concrete, qr, loc string
		var designCode sql.NullString
		var vol, steel float64
		var castAt time.Time
		var status, qcStatus string
		if err := rows.Scan(&id, &pid, &code, &stype, &batch, &concrete, &vol, &steel, &castAt, &status, &qcStatus, &qr, &loc, &designCode); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		item := map[string]interface{}{
			"id": id, "project_id": pid, "segment_code": code, "segment_type": stype,
			"cast_batch": batch, "concrete_grade": concrete, "concrete_volume_m3": vol,
			"steel_weight_kg": steel, "cast_at": castAt, "status": status,
			"qc_status": qcStatus, "qr_code": qr, "location": loc,
		}
		if designCode.Valid { item["design_code"] = designCode.String }
		items = append(items, item)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *RingBuilderHandler) CreateProduction(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID   string `json:"project_id"`
		SegmentCode string `json:"segment_code"`
		SegmentType string `json:"segment_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	_, err := h.db.Exec(`INSERT INTO segment_production (id, project_id, segment_code, segment_type) VALUES ($1,$2,$3,$4)`,
		id, input.ProjectID, input.SegmentCode, input.SegmentType)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *RingBuilderHandler) GetProduction(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var code, stype string
	err := h.db.QueryRow(`SELECT segment_code, segment_type FROM segment_production WHERE id = $1`, id).Scan(&code, &stype)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "segment not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "segment_code": code, "segment_type": stype})
}

func (h *RingBuilderHandler) UpdateProduction(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Status *string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	_, err := h.db.Exec(`UPDATE segment_production SET status=COALESCE($1,status), updated_at=NOW() WHERE id=$2`, input.Status, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// =============================================================================
// Curing
// =============================================================================
func (h *RingBuilderHandler) ListCuring(w http.ResponseWriter, r *http.Request) {
	segmentID := r.URL.Query().Get("segment_id")
	query := `SELECT id, segment_id, curing_stage, start_time, end_time, temp_target_c, temp_actual_c, humidity_target_pct, humidity_actual_pct, gradient_rate_cph FROM segment_curing WHERE 1=1`
	args := []interface{}{}
	argIdx := 1
	if segmentID != "" { query += fmt.Sprintf(" AND segment_id = $%d", argIdx); argIdx++; args = append(args, segmentID) }
	query += " ORDER BY start_time"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, segID, stage string
		var startTime time.Time
		var endTime, tempT, tempA, humT, humA, grad sql.NullString
		if err := rows.Scan(&id, &segID, &stage, &startTime, &endTime, &tempT, &tempA, &humT, &humA, &grad); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		items = append(items, map[string]interface{}{"id": id, "segment_id": segID, "curing_stage": stage, "start_time": startTime})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *RingBuilderHandler) CreateCuring(w http.ResponseWriter, r *http.Request) {
	var input struct {
		SegmentID  string `json:"segment_id"`
		CuringStage string `json:"curing_stage"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	_, err := h.db.Exec(`INSERT INTO segment_curing (id, segment_id, curing_stage) VALUES ($1,$2,$3)`, id, input.SegmentID, input.CuringStage)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

// =============================================================================
// Transport
// =============================================================================
func (h *RingBuilderHandler) ListTransport(w http.ResponseWriter, r *http.Request) {
	segmentID := r.URL.Query().Get("segment_id")
	query := `SELECT id, segment_id, transport_date, transport_mode, vehicle_number, driver_name, from_location, to_location, departure_time, arrival_time, distance_km, damage_reported, temperature_c, transport_cost FROM segment_transport WHERE 1=1`
	args := []interface{}{}
	argIdx := 1
	if segmentID != "" { query += fmt.Sprintf(" AND segment_id = $%d", argIdx); argIdx++; args = append(args, segmentID) }
	query += " ORDER BY transport_date DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, segID, mode, veh, driver, fromLoc, toLoc string
		var depTime, arrTime sql.NullString
		var dist, cost float64
		var tdate time.Time
		var damaged bool
		var temp sql.NullFloat64
		if err := rows.Scan(&id, &segID, &tdate, &mode, &veh, &driver, &fromLoc, &toLoc, &depTime, &arrTime, &dist, &damaged, &temp, &cost); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		items = append(items, map[string]interface{}{
			"id": id, "segment_id": segID, "transport_date": tdate, "transport_mode": mode,
			"vehicle_number": veh, "driver_name": driver, "from_location": fromLoc,
			"to_location": toLoc, "damage_reported": damaged, "distance_km": dist,
			"transport_cost": cost,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *RingBuilderHandler) CreateTransport(w http.ResponseWriter, r *http.Request) {
	var input struct {
		SegmentID string `json:"segment_id"`
		TransMode string `json:"transport_mode"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	_, err := h.db.Exec(`INSERT INTO segment_transport (id, segment_id, transport_mode) VALUES ($1,$2,$3)`, id, input.SegmentID, input.TransMode)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

// =============================================================================
// Installation
// =============================================================================
func (h *RingBuilderHandler) ListInstallation(w http.ResponseWriter, r *http.Request) {
	ringID := r.URL.Query().Get("ring_id")
	query := `SELECT i.id, i.segment_id, i.ring_id, i.erector_cycle_time_sec, i.bolt_count, i.bolt_torque_nm, i.gap_mm, i.offset_radial_mm, i.offset_longitudinal_mm, i.installed_by, i.installed_at
		FROM segment_installation i WHERE 1=1`
	args := []interface{}{}
	argIdx := 1
	if ringID != "" { query += fmt.Sprintf(" AND i.ring_id = $%d", argIdx); argIdx++; args = append(args, ringID) }
	query += " ORDER BY i.installed_at DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, segID, ringID, installer string
		var cycleTime, boltCount int
		var boltTorque, gap, offRad, offLong float64
		var installedAt time.Time
		if err := rows.Scan(&id, &segID, &ringID, &cycleTime, &boltCount, &boltTorque, &gap, &offRad, &offLong, &installer, &installedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		items = append(items, map[string]interface{}{
			"id": id, "segment_id": segID, "ring_id": ringID,
			"erector_cycle_time_sec": cycleTime, "bolt_count": boltCount,
			"bolt_torque_nm": boltTorque, "gap_mm": gap,
			"offset_radial_mm": offRad, "offset_longitudinal_mm": offLong,
			"installed_by": installer, "installed_at": installedAt,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *RingBuilderHandler) CreateInstallation(w http.ResponseWriter, r *http.Request) {
	var input struct {
		SegmentID string `json:"segment_id"`
		RingID    string `json:"ring_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	_, err := h.db.Exec(`INSERT INTO segment_installation (id, segment_id, ring_id) VALUES ($1,$2,$3)`, id, input.SegmentID, input.RingID)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

// =============================================================================
// QC
// =============================================================================
func (h *RingBuilderHandler) ListQC(w http.ResponseWriter, r *http.Request) {
	segmentID := r.URL.Query().Get("segment_id")
	query := `SELECT id, segment_id, qc_result, qc_inspector, qc_date, compressive_strength_mpa, corrective_action FROM segment_qc WHERE 1=1`
	args := []interface{}{}
	argIdx := 1
	if segmentID != "" { query += fmt.Sprintf(" AND segment_id = $%d", argIdx); argIdx++; args = append(args, segmentID) }
	query += " ORDER BY qc_date DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, segID, result, inspector, corrAction string
		var strength sql.NullFloat64
		var qcDate time.Time
		if err := rows.Scan(&id, &segID, &result, &inspector, &qcDate, &strength, &corrAction); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		item := map[string]interface{}{"id": id, "segment_id": segID, "qc_result": result, "qc_inspector": inspector, "qc_date": qcDate}
		if strength.Valid { item["compressive_strength_mpa"] = strength.Float64 }
		items = append(items, item)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *RingBuilderHandler) CreateQC(w http.ResponseWriter, r *http.Request) {
	var input struct {
		SegmentID string `json:"segment_id"`
		QCResult  string `json:"qc_result"`
		Inspector string `json:"qc_inspector"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	_, err := h.db.Exec(`INSERT INTO segment_qc (id, segment_id, qc_result, qc_inspector, qc_date) VALUES ($1,$2,$3,$4,CURRENT_DATE)`,
		id, input.SegmentID, input.QCResult, input.Inspector)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *RingBuilderHandler) GetQC(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var result string
	err := h.db.QueryRow(`SELECT qc_result FROM segment_qc WHERE id = $1`, id).Scan(&result)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "qc record not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"id": id, "qc_result": result})
}

// =============================================================================
// Inventory
// =============================================================================
func (h *RingBuilderHandler) ListInventory(w http.ResponseWriter, r *http.Request) {
	query := `SELECT id, project_id, design_id, segment_type, quantity_planned, quantity_produced, quantity_passed_qc, quantity_installed, quantity_defective, quantity_in_transit, quantity_in_stock, stock_location FROM segment_inventory ORDER BY segment_type`

	rows, err := h.db.Query(query)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, pid, dsgID, stype, loc sql.NullString
		var planned, produced, passed, installed, defective, inTransit, inStock int
		if err := rows.Scan(&id, &pid, &dsgID, &stype, &planned, &produced, &passed, &installed, &defective, &inTransit, &inStock, &loc); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		item := map[string]interface{}{
			"id": id, "segment_type": stype, "quantity_planned": planned,
			"quantity_produced": produced, "quantity_passed_qc": passed,
			"quantity_installed": installed, "quantity_defective": defective,
			"quantity_in_transit": inTransit, "quantity_in_stock": inStock,
		}
		if loc.Valid { item["stock_location"] = loc.String }
		items = append(items, item)
	}
	respondJSON(w, http.StatusOK, items)
}

// =============================================================================
// Ring Measurements
// =============================================================================
func (h *RingBuilderHandler) ListMeasurements(w http.ResponseWriter, r *http.Request) {
	ringID := r.URL.Query().Get("ring_id")
	query := `SELECT id, ring_id, measured_at, horizontal_convergence_mm, vertical_convergence_mm, ovality_pct, ovality_mm, settlement_mm, instrument_type, measured_by FROM ring_measurements WHERE 1=1`
	args := []interface{}{}
	argIdx := 1
	if ringID != "" { query += fmt.Sprintf(" AND ring_id = $%d", argIdx); argIdx++; args = append(args, ringID) }
	query += " ORDER BY measured_at DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, ringID, instr, measBy string
		var measuredAt time.Time
		var hConv, vConv, ovalPct, ovalMm, settle sql.NullFloat64
		if err := rows.Scan(&id, &ringID, &measuredAt, &hConv, &vConv, &ovalPct, &ovalMm, &settle, &instr, &measBy); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		item := map[string]interface{}{"id": id, "ring_id": ringID, "measured_at": measuredAt, "instrument_type": instr, "measured_by": measBy}
		if hConv.Valid { item["horizontal_convergence_mm"] = hConv.Float64 }
		if ovalPct.Valid { item["ovality_pct"] = ovalPct.Float64 }
		items = append(items, item)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *RingBuilderHandler) CreateMeasurement(w http.ResponseWriter, r *http.Request) {
	var input struct {
		RingID string `json:"ring_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	_, err := h.db.Exec(`INSERT INTO ring_measurements (id, ring_id) VALUES ($1,$2)`, id, input.RingID)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

// =============================================================================
// Summary
// =============================================================================
func (h *RingBuilderHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")

	summary := map[string]interface{}{
		"total_designs": 0, "total_produced": 0, "total_passed_qc": 0,
		"total_installed": 0, "total_defective": 0,
	}

	rows, err := h.db.Query(`SELECT COUNT(*) FROM ring_designs WHERE status='active'`+mapCondRB(projectID, "project_id"))
	if err == nil && rows.Next() { rows.Scan(&summary["total_designs"]); rows.Close() }

	rows2, err := h.db.Query(`SELECT COUNT(*) FROM segment_production WHERE 1=1`+mapCondRB(projectID, "project_id"))
	if err == nil && rows2.Next() { rows2.Scan(&summary["total_produced"]); rows2.Close() }

	rows3, err := h.db.Query(`SELECT COUNT(*) FROM segment_production WHERE qc_status='passed'`+mapCondRB(projectID, "project_id"))
	if err == nil && rows3.Next() { rows3.Scan(&summary["total_passed_qc"]); rows3.Close() }

	rows4, err := h.db.Query(`SELECT COUNT(*) FROM segment_production WHERE status='installed'`+mapCondRB(projectID, "project_id"))
	if err == nil && rows4.Next() { rows4.Scan(&summary["total_installed"]); rows4.Close() }

	rows5, err := h.db.Query(`SELECT COALESCE(SUM(quantity_defective),0) FROM segment_inventory WHERE 1=1`+mapCondRB(projectID, "project_id"))
	if err == nil && rows5.Next() { rows5.Scan(&summary["total_defective"]); rows5.Close() }

	respondJSON(w, http.StatusOK, summary)
}

func mapCondRB(val, col string) string {
	if val != "" { return fmt.Sprintf(" AND %s = '%s'", col, val) }
	return ""
}