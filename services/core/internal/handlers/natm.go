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

// NATMHandler handles NATM & Microtunnelling module endpoints
type NATMHandler struct {
	db *sql.DB
}

func NewNATMHandler(db *sql.DB) *NATMHandler {
	return &NATMHandler{db: db}
}

func (h *NATMHandler) RegisterRoutes(r chi.Router) {
	r.Route("/natm", func(r chi.Router) {
		// NATM Excavation
		r.Get("/excavation", h.ListExcavation)
		r.Post("/excavation", h.CreateExcavation)
		r.Get("/excavation/{id}", h.GetExcavation)
		r.Put("/excavation/{id}", h.UpdateExcavation)

		// Shotcrete
		r.Get("/shotcrete", h.ListShotcrete)
		r.Post("/shotcrete", h.CreateShotcrete)

		// Rock Bolts
		r.Get("/rock-bolts", h.ListRockBolts)
		r.Post("/rock-bolts", h.CreateRockBolt)

		// Steel Sets
		r.Get("/steel-sets", h.ListSteelSets)
		r.Post("/steel-sets", h.CreateSteelSet)

		// Convergence
		r.Get("/convergence", h.ListConvergence)
		r.Post("/convergence", h.CreateConvergence)

		// Face Mapping
		r.Get("/face-mapping", h.ListFaceMapping)
		r.Post("/face-mapping", h.CreateFaceMapping)

		// MTBM Drives
		r.Get("/mtbm-drives", h.ListMTBMDrives)
		r.Post("/mtbm-drives", h.CreateMTBMDrive)
		r.Get("/mtbm-drives/{id}", h.GetMTBMDrive)

		// MTBM Thrust
		r.Get("/mtbm-thrust", h.ListMTBMThrust)
		r.Post("/mtbm-thrust", h.CreateMTBMThrust)

		// MTBM Lubrication
		r.Get("/mtbm-lubrication", h.ListMTBMLubrication)
		r.Post("/mtbm-lubrication", h.CreateMTBMLubrication)

		// MTBM Survey
		r.Get("/mtbm-survey", h.ListMTBMSurvey)
		r.Post("/mtbm-survey", h.CreateMTBMSurvey)

		// Shafts
		r.Get("/shafts", h.ListShafts)
		r.Post("/shafts", h.CreateShaft)
		r.Get("/shafts/{id}", h.GetShaft)

		// Shaft Equipment
		r.Get("/shaft-equipment", h.ListShaftEquipment)
		r.Post("/shaft-equipment", h.CreateShaftEquipment)

		// Cross Passages
		r.Get("/cross-passages", h.ListCrossPassages)
		r.Post("/cross-passages", h.CreateCrossPassage)

		// Grouting
		r.Get("/grouting", h.ListGrouting)
		r.Post("/grouting", h.CreateGrouting)

		// Settlement
		r.Get("/settlement", h.ListSettlement)
		r.Post("/settlement", h.CreateSettlement)

		r.Get("/summary", h.GetSummary)
	})
}

// =============================================================================
// Excavation
// =============================================================================
func (h *NATMHandler) ListExcavation(w http.ResponseWriter, r *http.Request) {
	driveID := r.URL.Query().Get("drive_id")
	query := `SELECT id, drive_id, round_no, chainage_from, chainage_to, excavation_date, shift, method, round_length_m, excavated_volume_m3, geotech_class, water_inflow_lmin, support_class, standup_time_hours, delay_minutes, delay_reason FROM natm_excavation_log WHERE 1=1`
	args := []interface{}{}
	argIdx := 1
	if driveID != "" { query += fmt.Sprintf(" AND drive_id = $%d", argIdx); argIdx++; args = append(args, driveID) }
	query += " ORDER BY round_no"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, did, shift, method, geoClass, supClass, delayReason sql.NullString
		var rnd int
		var chFrom, chTo, rLen, vol, water, standup float64
		var delayMin int
		var excDate time.Time
		if err := rows.Scan(&id, &did, &rnd, &chFrom, &chTo, &excDate, &shift, &method, &rLen, &vol, &geoClass, &water, &supClass, &standup, &delayMin, &delayReason); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		items = append(items, map[string]interface{}{
			"id": id, "drive_id": did, "round_no": rnd,
			"chainage_from": chFrom, "chainage_to": chTo,
			"excavation_date": excDate, "method": method,
			"round_length_m": rLen, "excavated_volume_m3": vol,
			"geotech_class": geoClass, "support_class": supClass,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *NATMHandler) CreateExcavation(w http.ResponseWriter, r *http.Request) {
	var input struct {
		DriveID string `json:"drive_id"`
		RoundNo int    `json:"round_no"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	_, err := h.db.Exec(`INSERT INTO natm_excavation_log (id, drive_id, round_no) VALUES ($1,$2,$3)`, id, input.DriveID, input.RoundNo)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *NATMHandler) GetExcavation(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var rnd int
	err := h.db.QueryRow(`SELECT round_no FROM natm_excavation_log WHERE id = $1`, id).Scan(&rnd)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "excavation round not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "round_no": rnd})
}

func (h *NATMHandler) UpdateExcavation(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		SupportClass *string `json:"support_class"`
		GeotechClass *string `json:"geotech_class"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	_, err := h.db.Exec(`UPDATE natm_excavation_log SET support_class=COALESCE($1,support_class), geotech_class=COALESCE($2,geotech_class) WHERE id=$3`,
		input.SupportClass, input.GeotechClass, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// =============================================================================
// Shotcrete
// =============================================================================
func (h *NATMHandler) ListShotcrete(w http.ResponseWriter, r *http.Request) {
	driveID := r.URL.Query().Get("drive_id")
	query := `SELECT id, drive_id, application_date, location_type, shotcrete_type, thickness_mm, area_m2, volume_m3, compressive_strength_mpa, fiber_content_kgm3, rebound_pct, qc_status FROM natm_shotcrete WHERE 1=1`
	args := []interface{}{}
	argIdx := 1
	if driveID != "" { query += fmt.Sprintf(" AND drive_id = $%d", argIdx); argIdx++; args = append(args, driveID) }
	query += " ORDER BY application_date DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, did, locType, stype, qcStatus sql.NullString
		var thick, area, vol, strength, fiber, rebound float64
		var appDate time.Time
		if err := rows.Scan(&id, &did, &appDate, &locType, &stype, &thick, &area, &vol, &strength, &fiber, &rebound, &qcStatus); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		items = append(items, map[string]interface{}{
			"id": id, "drive_id": did, "application_date": appDate,
			"location_type": locType, "shotcrete_type": stype,
			"thickness_mm": thick, "area_m2": area, "volume_m3": vol,
			"compressive_strength_mpa": strength, "rebound_pct": rebound,
			"qc_status": qcStatus,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *NATMHandler) CreateShotcrete(w http.ResponseWriter, r *http.Request) {
	var input struct {
		DriveID    string `json:"drive_id"`
		ShotcreteType string `json:"shotcrete_type"`
		Thickness  float64 `json:"thickness_mm"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	_, err := h.db.Exec(`INSERT INTO natm_shotcrete (id, drive_id, shotcrete_type, thickness_mm) VALUES ($1,$2,$3,$4)`,
		id, input.DriveID, input.ShotcreteType, input.Thickness)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

// =============================================================================
// Rock Bolts
// =============================================================================
func (h *NATMHandler) ListRockBolts(w http.ResponseWriter, r *http.Request) {
	driveID := r.URL.Query().Get("drive_id")
	query := `SELECT id, drive_id, bolt_type, bolt_diameter_mm, bolt_length_mm, spacing_longitudinal_m, spacing_transverse_m, quantity_installed, pretension_kN, pullout_test_kN, pattern_type, qc_status FROM natm_rock_bolts WHERE 1=1`
	args := []interface{}{}
	argIdx := 1
	if driveID != "" { query += fmt.Sprintf(" AND drive_id = $%d", argIdx); argIdx++; args = append(args, driveID) }
	query += " ORDER BY installed_at DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, did, btype, pattern, qcStatus sql.NullString
		var diam, length int
		var spacingL, spacingT, pretension, pullout float64
		var qty int
		if err := rows.Scan(&id, &did, &btype, &diam, &length, &spacingL, &spacingT, &qty, &pretension, &pullout, &pattern, &qcStatus); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		items = append(items, map[string]interface{}{
			"id": id, "drive_id": did, "bolt_type": btype,
			"bolt_diameter_mm": diam, "bolt_length_mm": length,
			"spacing_longitudinal_m": spacingL, "spacing_transverse_m": spacingT,
			"quantity_installed": qty, "pretension_kN": pretension,
			"pullout_test_kN": pullout, "pattern_type": pattern,
			"qc_status": qcStatus,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *NATMHandler) CreateRockBolt(w http.ResponseWriter, r *http.Request) {
	var input struct {
		DriveID string `json:"drive_id"`
		BoltType string `json:"bolt_type"`
		Quantity int    `json:"quantity_installed"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	_, err := h.db.Exec(`INSERT INTO natm_rock_bolts (id, drive_id, bolt_type, quantity_installed) VALUES ($1,$2,$3,$4)`,
		id, input.DriveID, input.BoltType, input.Quantity)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

// =============================================================================
// Steel Sets
// =============================================================================
func (h *NATMHandler) ListSteelSets(w http.ResponseWriter, r *http.Request) {
	driveID := r.URL.Query().Get("drive_id")
	query := `SELECT id, drive_id, set_number, chainage, set_type, spacing_m, steel_grade, quantity_arches, qc_status FROM natm_steel_sets WHERE 1=1`
	args := []interface{}{}
	argIdx := 1
	if driveID != "" { query += fmt.Sprintf(" AND drive_id = $%d", argIdx); argIdx++; args = append(args, driveID) }
	query += " ORDER BY set_number"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, did, stype, grade, qcStatus sql.NullString
		var sn, arches int
		var chainage, spacing float64
		if err := rows.Scan(&id, &did, &sn, &chainage, &stype, &spacing, &grade, &arches, &qcStatus); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		items = append(items, map[string]interface{}{
			"id": id, "drive_id": did, "set_number": sn, "chainage": chainage,
			"set_type": stype, "spacing_m": spacing, "steel_grade": grade,
			"quantity_arches": arches, "qc_status": qcStatus,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *NATMHandler) CreateSteelSet(w http.ResponseWriter, r *http.Request) {
	var input struct {
		DriveID string `json:"drive_id"`
		SetNo   int    `json:"set_number"`
		SetType string `json:"set_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	_, err := h.db.Exec(`INSERT INTO natm_steel_sets (id, drive_id, set_number, set_type) VALUES ($1,$2,$3,$4)`,
		id, input.DriveID, input.SetNo, input.SetType)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

// =============================================================================
// Convergence
// =============================================================================
func (h *NATMHandler) ListConvergence(w http.ResponseWriter, r *http.Request) {
	driveID := r.URL.Query().Get("drive_id")
	query := `SELECT id, drive_id, measurement_point, chainage, measured_at, displacement_vertical_mm, displacement_horizontal_mm, convergence_rate_mmday, cumulative_displacement_mm, instrument_type, alarm_triggered FROM natm_convergence WHERE 1=1`
	args := []interface{}{}
	argIdx := 1
	if driveID != "" { query += fmt.Sprintf(" AND drive_id = $%d", argIdx); argIdx++; args = append(args, driveID) }
	query += " ORDER BY measured_at DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, did, pt, instr sql.NullString
		var chainage, dv, dh, rate, cumul float64
		var measuredAt time.Time
		var alarm bool
		if err := rows.Scan(&id, &did, &pt, &chainage, &measuredAt, &dv, &dh, &rate, &cumul, &instr, &alarm); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		items = append(items, map[string]interface{}{
			"id": id, "drive_id": did, "measurement_point": pt,
			"chainage": chainage, "measured_at": measuredAt,
			"displacement_vertical_mm": dv, "displacement_horizontal_mm": dh,
			"convergence_rate_mmday": rate, "cumulative_displacement_mm": cumul,
			"alarm_triggered": alarm,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *NATMHandler) CreateConvergence(w http.ResponseWriter, r *http.Request) {
	var input struct {
		DriveID string `json:"drive_id"`
		Point   string `json:"measurement_point"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	_, err := h.db.Exec(`INSERT INTO natm_convergence (id, drive_id, measurement_point) VALUES ($1,$2,$3)`, id, input.DriveID, input.Point)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

// =============================================================================
// Face Mapping
// =============================================================================
func (h *NATMHandler) ListFaceMapping(w http.ResponseWriter, r *http.Request) {
	driveID := r.URL.Query().Get("drive_id")
	query := `SELECT id, drive_id, chainage, mapped_at, rock_type, weathering_grade, rmr_score, q_score, gsi_value, joint_count, fault_zone, mapped_by FROM natm_face_mapping WHERE 1=1`
	args := []interface{}{}
	argIdx := 1
	if driveID != "" { query += fmt.Sprintf(" AND drive_id = $%d", argIdx); argIdx++; args = append(args, driveID) }
	query += " ORDER BY chainage"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, did, rock, weather, mapper sql.NullString
		var chainage, rmr, qscore, gsi float64
		var joints int
		var fault bool
		var mappedAt time.Time
		if err := rows.Scan(&id, &did, &chainage, &mappedAt, &rock, &weather, &rmr, &qscore, &gsi, &joints, &fault, &mapper); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		items = append(items, map[string]interface{}{
			"id": id, "drive_id": did, "chainage": chainage,
			"rock_type": rock, "weathering_grade": weather,
			"rmr_score": rmr, "q_score": qscore, "gsi_value": gsi,
			"joint_count": joints, "fault_zone": fault, "mapped_by": mapper,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *NATMHandler) CreateFaceMapping(w http.ResponseWriter, r *http.Request) {
	var input struct {
		DriveID string `json:"drive_id"`
		RockType string `json:"rock_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	_, err := h.db.Exec(`INSERT INTO natm_face_mapping (id, drive_id, rock_type) VALUES ($1,$2,$3)`, id, input.DriveID, input.RockType)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

// =============================================================================
// MTBM Drives
// =============================================================================
func (h *NATMHandler) ListMTBMDrives(w http.ResponseWriter, r *http.Request) {
	query := `SELECT id, project_id, drive_code, drive_name, pipe_type, pipe_diameter_mm, pipe_length_mm, design_length_m, max_jacking_force_kN, intermediate_jack_stations, lubrication_type, status, start_date FROM mtbm_drives ORDER BY drive_code`
	rows, err := h.db.Query(query)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, pid, code, name, pipeType, lubType, status sql.NullString
		var diam, pLen int
		var designLen, maxJack float64
		var jackStations int
		var startDate sql.NullString
		if err := rows.Scan(&id, &pid, &code, &name, &pipeType, &diam, &pLen, &designLen, &maxJack, &jackStations, &lubType, &status, &startDate); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		items = append(items, map[string]interface{}{
			"id": id, "project_id": pid, "drive_code": code, "drive_name": name,
			"pipe_type": pipeType, "pipe_diameter_mm": diam,
			"design_length_m": designLen, "max_jacking_force_kN": maxJack,
			"status": status,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *NATMHandler) CreateMTBMDrive(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID string `json:"project_id"`
		DriveCode string `json:"drive_code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	_, err := h.db.Exec(`INSERT INTO mtbm_drives (id, project_id, drive_code) VALUES ($1,$2,$3)`, id, input.ProjectID, input.DriveCode)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *NATMHandler) GetMTBMDrive(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var code string
	err := h.db.QueryRow(`SELECT drive_code FROM mtbm_drives WHERE id = $1`, id).Scan(&code)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "mtbm drive not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"id": id, "drive_code": code})
}

// =============================================================================
// MTBM Thrust
// =============================================================================
func (h *NATMHandler) ListMTBMThrust(w http.ResponseWriter, r *http.Request) {
	driveID := r.URL.Query().Get("mtbm_drive_id")
	query := `SELECT id, mtbm_drive_id, pipe_no, recorded_at, thrust_force_kN, thrust_pressure_bar, advance_speed_mmmin, torque_kNm, slurry_pressure_bar, slurry_flow_m3h, face_pressure_bar, alignment_vertical_mm, alignment_horizontal_mm FROM mtbm_thrust_log WHERE 1=1`
	args := []interface{}{}
	argIdx := 1
	if driveID != "" { query += fmt.Sprintf(" AND mtbm_drive_id = $%d", argIdx); argIdx++; args = append(args, driveID) }
	query += " ORDER BY pipe_no"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, did sql.NullString
		var pipeNo int
		var recordedAt time.Time
		var thrust, press, speed, torque, slurryP, slurryF, faceP, alignV, alignH float64
		if err := rows.Scan(&id, &did, &pipeNo, &recordedAt, &thrust, &press, &speed, &torque, &slurryP, &slurryF, &faceP, &alignV, &alignH); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		items = append(items, map[string]interface{}{
			"id": id, "mtbm_drive_id": did, "pipe_no": pipeNo,
			"thrust_force_kN": thrust, "thrust_pressure_bar": press,
			"advance_speed_mmmin": speed, "torque_kNm": torque,
			"slurry_pressure_bar": slurryP, "face_pressure_bar": faceP,
			"alignment_vertical_mm": alignV, "alignment_horizontal_mm": alignH,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *NATMHandler) CreateMTBMThrust(w http.ResponseWriter, r *http.Request) {
	var input struct {
		MtbmDriveID string `json:"mtbm_drive_id"`
		PipeNo      int    `json:"pipe_no"`
		ThrustForce float64 `json:"thrust_force_kN"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	_, err := h.db.Exec(`INSERT INTO mtbm_thrust_log (id, mtbm_drive_id, pipe_no, thrust_force_kN) VALUES ($1,$2,$3,$4)`,
		id, input.MtbmDriveID, input.PipeNo, input.ThrustForce)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

// =============================================================================
// MTBM Lubrication
// =============================================================================
func (h *NATMHandler) ListMTBMLubrication(w http.ResponseWriter, r *http.Request) {
	driveID := r.URL.Query().Get("mtbm_drive_id")
	query := `SELECT id, mtbm_drive_id, pipe_no, recorded_at, lubricant_type, injection_pressure_bar, flow_rate_lmin, total_volume_m3, density_kgm3, marsh_viscosity_sec FROM mtbm_lubrication WHERE 1=1`
	args := []interface{}{}
	argIdx := 1
	if driveID != "" { query += fmt.Sprintf(" AND mtbm_drive_id = $%d", argIdx); argIdx++; args = append(args, driveID) }
	query += " ORDER BY recorded_at DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, did, lubType sql.NullString
		var pipeNo int
		var recordedAt time.Time
		var press, flow, vol, dens, marsh float64
		if err := rows.Scan(&id, &did, &pipeNo, &recordedAt, &lubType, &press, &flow, &vol, &dens, &marsh); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		items = append(items, map[string]interface{}{
			"id": id, "mtbm_drive_id": did, "pipe_no": pipeNo,
			"lubricant_type": lubType, "injection_pressure_bar": press,
			"flow_rate_lmin": flow, "total_volume_m3": vol,
			"density_kgm3": dens, "marsh_viscosity_sec": marsh,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *NATMHandler) CreateMTBMLubrication(w http.ResponseWriter, r *http.Request) {
	var input struct {
		MtbmDriveID  string  `json:"mtbm_drive_id"`
		LubricantType string `json:"lubricant_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	_, err := h.db.Exec(`INSERT INTO mtbm_lubrication (id, mtbm_drive_id, lubricant_type) VALUES ($1,$2,$3)`, id, input.MtbmDriveID, input.LubricantType)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

// =============================================================================
// MTBM Survey
// =============================================================================
func (h *NATMHandler) ListMTBMSurvey(w http.ResponseWriter, r *http.Request) {
	driveID := r.URL.Query().Get("mtbm_drive_id")
	query := `SELECT id, mtbm_drive_id, pipe_no, surveyed_at, deviation_vertical_mm, deviation_horizontal_mm, deviation_roll_deg, instrument_type, survey_by FROM mtbm_survey WHERE 1=1`
	args := []interface{}{}
	argIdx := 1
	if driveID != "" { query += fmt.Sprintf(" AND mtbm_drive_id = $%d", argIdx); argIdx++; args = append(args, driveID) }
	query += " ORDER BY pipe_no"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, did, instr, surveyor sql.NullString
		var pipeNo int
		var surveyedAt time.Time
		var devV, devH, devRoll float64
		if err := rows.Scan(&id, &did, &pipeNo, &surveyedAt, &devV, &devH, &devRoll, &instr, &surveyor); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		items = append(items, map[string]interface{}{
			"id": id, "mtbm_drive_id": did, "pipe_no": pipeNo,
			"deviation_vertical_mm": devV, "deviation_horizontal_mm": devH,
			"instrument_type": instr, "survey_by": surveyor,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *NATMHandler) CreateMTBMSurvey(w http.ResponseWriter, r *http.Request) {
	var input struct {
		MtbmDriveID string `json:"mtbm_drive_id"`
		PipeNo      int    `json:"pipe_no"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	_, err := h.db.Exec(`INSERT INTO mtbm_survey (id, mtbm_drive_id, pipe_no) VALUES ($1,$2,$3)`, id, input.MtbmDriveID, input.PipeNo)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

// =============================================================================
// Shafts
// =============================================================================
func (h *NATMHandler) ListShafts(w http.ResponseWriter, r *http.Request) {
	query := `SELECT id, project_id, shaft_code, shaft_name, shaft_type, construction_method, diameter_m, depth_m, status, start_date, completion_date FROM shaft_construction ORDER BY shaft_code`
	rows, err := h.db.Query(query)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, pid, code, name, stype, cmethod, status sql.NullString
		var diam, depth float64
		var startDate, endDate sql.NullString
		if err := rows.Scan(&id, &pid, &code, &name, &stype, &cmethod, &diam, &depth, &status, &startDate, &endDate); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		items = append(items, map[string]interface{}{
			"id": id, "project_id": pid, "shaft_code": code, "shaft_name": name,
			"shaft_type": stype, "construction_method": cmethod,
			"diameter_m": diam, "depth_m": depth, "status": status,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *NATMHandler) CreateShaft(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID string `json:"project_id"`
		ShaftCode string `json:"shaft_code"`
		ShaftName string `json:"shaft_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	_, err := h.db.Exec(`INSERT INTO shaft_construction (id, project_id, shaft_code, shaft_name) VALUES ($1,$2,$3,$4)`,
		id, input.ProjectID, input.ShaftCode, input.ShaftName)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *NATMHandler) GetShaft(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var code, name string
	err := h.db.QueryRow(`SELECT shaft_code, shaft_name FROM shaft_construction WHERE id = $1`, id).Scan(&code, &name)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "shaft not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "shaft_code": code, "shaft_name": name})
}

// =============================================================================
// Shaft Equipment
// =============================================================================
func (h *NATMHandler) ListShaftEquipment(w http.ResponseWriter, r *http.Request) {
	shaftID := r.URL.Query().Get("shaft_id")
	query := `SELECT id, shaft_id, equipment_type, equipment_name, manufacturer, model, status FROM shaft_equipment WHERE 1=1`
	args := []interface{}{}
	argIdx := 1
	if shaftID != "" { query += fmt.Sprintf(" AND shaft_id = $%d", argIdx); argIdx++; args = append(args, shaftID) }
	query += " ORDER BY equipment_type"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, sid, etype, ename, mfr, model, status sql.NullString
		if err := rows.Scan(&id, &sid, &etype, &ename, &mfr, &model, &status); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		items = append(items, map[string]interface{}{
			"id": id, "shaft_id": sid, "equipment_type": etype,
			"equipment_name": ename, "manufacturer": mfr, "model": model,
			"status": status,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *NATMHandler) CreateShaftEquipment(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ShaftID       string `json:"shaft_id"`
		EquipmentType string `json:"equipment_type"`
		EquipmentName string `json:"equipment_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	_, err := h.db.Exec(`INSERT INTO shaft_equipment (id, shaft_id, equipment_type, equipment_name) VALUES ($1,$2,$3,$4)`,
		id, input.ShaftID, input.EquipmentType, input.EquipmentName)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

// =============================================================================
// Cross Passages
// =============================================================================
func (h *NATMHandler) ListCrossPassages(w http.ResponseWriter, r *http.Request) {
	query := `SELECT id, project_id, passage_code, passage_name, chainage, construction_method, length_m, width_m, height_m, lining_type, status FROM cross_passages ORDER BY passage_code`
	rows, err := h.db.Query(query)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, pid, code, name, cmethod, lining, status sql.NullString
		var chainage, length, width, height float64
		if err := rows.Scan(&id, &pid, &code, &name, &chainage, &cmethod, &length, &width, &height, &lining, &status); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		items = append(items, map[string]interface{}{
			"id": id, "project_id": pid, "passage_code": code, "passage_name": name,
			"chainage": chainage, "construction_method": cmethod,
			"length_m": length, "width_m": width, "height_m": height,
			"lining_type": lining, "status": status,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *NATMHandler) CreateCrossPassage(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID    string `json:"project_id"`
		PassageCode  string `json:"passage_code"`
		PassageName  string `json:"passage_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	_, err := h.db.Exec(`INSERT INTO cross_passages (id, project_id, passage_code, passage_name) VALUES ($1,$2,$3,$4)`,
		id, input.ProjectID, input.PassageCode, input.PassageName)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

// =============================================================================
// Grouting
// =============================================================================
func (h *NATMHandler) ListGrouting(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT id, project_id, grouting_type, location_type, chainage, grout_date, grout_mix_type, pressure_bar, flow_rate_lmin, volume_planned_m3, volume_actual_m3, take_kgm, supervisor FROM grouting_records WHERE 1=1`
	args := []interface{}{}
	argIdx := 1
	if projectID != "" { query += fmt.Sprintf(" AND project_id = $%d", argIdx); argIdx++; args = append(args, projectID) }
	query += " ORDER BY grout_date DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, pid, gtype, locType, mix, supervisor sql.NullString
		var chainage, press, flow, volPlan, volAct, take float64
		var groutDate time.Time
		if err := rows.Scan(&id, &pid, &gtype, &locType, &chainage, &groutDate, &mix, &press, &flow, &volPlan, &volAct, &take, &supervisor); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		items = append(items, map[string]interface{}{
			"id": id, "project_id": pid, "grouting_type": gtype,
			"location_type": locType, "chainage": chainage,
			"grout_date": groutDate, "grout_mix_type": mix,
			"pressure_bar": press, "flow_rate_lmin": flow,
			"volume_planned_m3": volPlan, "volume_actual_m3": volAct,
			"take_kgm": take,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *NATMHandler) CreateGrouting(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID    string `json:"project_id"`
		GroutingType string `json:"grouting_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	_, err := h.db.Exec(`INSERT INTO grouting_records (id, project_id, grouting_type) VALUES ($1,$2,$3)`, id, input.ProjectID, input.GroutingType)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

// =============================================================================
// Settlement
// =============================================================================
func (h *NATMHandler) ListSettlement(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT id, project_id, point_id, point_type, chainage, offset_m, monitored_at, settlement_mm, cumulative_settlement_mm, settlement_rate_mmday, alarm_triggered, instrument_type, reading_by FROM settlement_monitoring WHERE 1=1`
	args := []interface{}{}
	argIdx := 1
	if projectID != "" { query += fmt.Sprintf(" AND project_id = $%d", argIdx); argIdx++; args = append(args, projectID) }
	query += " ORDER BY monitored_at DESC LIMIT 200"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, pid, ptID, ptType, instr, reader sql.NullString
		var chainage, offset, settle, cumul, rate float64
		var monitoredAt time.Time
		var alarm bool
		if err := rows.Scan(&id, &pid, &ptID, &ptType, &chainage, &offset, &monitoredAt, &settle, &cumul, &rate, &alarm, &instr, &reader); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		items = append(items, map[string]interface{}{
			"id": id, "project_id": pid, "point_id": ptID, "point_type": ptType,
			"chainage": chainage, "offset_m": offset, "monitored_at": monitoredAt,
			"settlement_mm": settle, "cumulative_settlement_mm": cumul,
			"settlement_rate_mmday": rate, "alarm_triggered": alarm,
			"instrument_type": instr, "reading_by": reader,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *NATMHandler) CreateSettlement(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID string `json:"project_id"`
		PointID   string `json:"point_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	_, err := h.db.Exec(`INSERT INTO settlement_monitoring (id, project_id, point_id) VALUES ($1,$2,$3)`, id, input.ProjectID, input.PointID)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

// =============================================================================
// Summary
// =============================================================================
func (h *NATMHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	summary := map[string]interface{}{
		"total_excavation_rounds": 0, "total_shotcrete": 0, "total_rock_bolts": 0,
		"total_steel_sets": 0, "total_convergence_points": 0, "total_face_maps": 0,
		"total_mtbm_drives": 0, "total_shafts": 0, "total_settlement_points": 0,
		"active_settlement_alarms": 0,
	}

	cond := ""
	if projectID != "" {
		cond = " WHERE project_id = '" + projectID + "'"
	}

	rows, err := h.db.Query(`SELECT COUNT(*) FROM natm_excavation_log`)
	if err == nil && rows.Next() { rows.Scan(&summary["total_excavation_rounds"]); rows.Close() }

	rows2, err := h.db.Query(`SELECT COUNT(*) FROM natm_shotcrete`)
	if err == nil && rows2.Next() { rows2.Scan(&summary["total_shotcrete"]); rows2.Close() }

	rows3, err := h.db.Query(`SELECT COUNT(*) FROM natm_rock_bolts`)
	if err == nil && rows3.Next() { rows3.Scan(&summary["total_rock_bolts"]); rows3.Close() }

	rows4, err := h.db.Query(`SELECT COUNT(*) FROM natm_steel_sets`)
	if err == nil && rows4.Next() { rows4.Scan(&summary["total_steel_sets"]); rows4.Close() }

	rows5, err := h.db.Query(`SELECT COUNT(*) FROM settlement_monitoring WHERE alarm_triggered=TRUE` + cond)
	if err == nil && rows5.Next() { rows5.Scan(&summary["active_settlement_alarms"]); rows5.Close() }

	respondJSON(w, http.StatusOK, summary)
}