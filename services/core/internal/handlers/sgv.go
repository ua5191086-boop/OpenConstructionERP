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

type SettlementGroutingVentHandler struct{ db *sql.DB }
func NewSGVHandler(db *sql.DB) *SettlementGroutingVentHandler { return &SettlementGroutingVentHandler{db: db} }

func (h *SettlementGroutingVentHandler) RegisterRoutes(r chi.Router) {
	r.Route("/sgv", func(r chi.Router) {
		// Settlement
		r.Get("/settlement-points", h.ListSettlementPoints)
		r.Post("/settlement-points", h.CreateSettlementPoint)
		r.Get("/settlement-points/{id}", h.GetSettlementPoint)
		r.Put("/settlement-points/{id}", h.UpdateSettlementPoint)
		r.Get("/settlement-points/{pId}/readings", h.ListSettlementReadings)
		r.Post("/settlement-points/{pId}/readings", h.CreateSettlementReading)

		// Grouting
		r.Get("/grouting", h.ListGrouting)
		r.Post("/grouting", h.CreateGrouting)
		r.Get("/grouting/{id}", h.GetGrouting)
		r.Put("/grouting/{id}", h.UpdateGrouting)
		r.Get("/grouting/{gId}/records", h.ListGroutingRecords)
		r.Post("/grouting/{gId}/records", h.CreateGroutingRecord)

		// Ventilation
		r.Get("/ventilation", h.ListVentilation)
		r.Post("/ventilation", h.CreateVentilation)
		r.Get("/ventilation/{id}", h.GetVentilation)
		r.Put("/ventilation/{id}", h.UpdateVentilation)
		r.Get("/ventilation/{vId}/readings", h.ListVentReadings)
		r.Post("/ventilation/{vId}/readings", h.CreateVentReading)

		r.Get("/summary", h.GetSummary)
	})
}

// ========== SETTLEMENT ==========
func (h *SettlementGroutingVentHandler) ListSettlementPoints(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	q := `SELECT id, project_id, point_code, point_name, point_type, chainage_m, offset_m, initial_level_m, trigger_alert_mm, trigger_urgent_mm, status, notes, created_at FROM settlement_monitoring_points WHERE 1=1`
	var args []interface{}; ai := 1
	if projectID != "" { q += fmt.Sprintf(" AND project_id = $%d", ai); args = append(args, projectID); ai++ }
	q += " ORDER BY chainage_m"
	rows, err := h.db.Query(q, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()
	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, pid, code, name, ptype, status, notes string
		var chain, offset, init, alert, urgent float64
		var crAt time.Time
		if err := rows.Scan(&id, &pid, &code, &name, &ptype, &chain, &offset, &init, &alert, &urgent, &status, &notes, &crAt); err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
		items = append(items, map[string]interface{}{"id": id, "project_id": pid, "point_code": code, "point_name": name, "point_type": ptype, "chainage_m": chain, "offset_m": offset, "initial_level_m": init, "trigger_alert_mm": alert, "trigger_urgent_mm": urgent, "status": status, "notes": notes, "created_at": crAt})
	}
	respondJSON(w, http.StatusOK, items)
}
func (h *SettlementGroutingVentHandler) CreateSettlementPoint(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID   string  `json:"project_id"`
		PointCode   string  `json:"point_code"`
		PointName   string  `json:"point_name"`
		PointType   string  `json:"point_type"`
		ChainageM   float64 `json:"chainage_m"`
		OffsetM     float64 `json:"offset_m"`
		InitialLevelM float64 `json:"initial_level_m"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil { respondError(w, http.StatusBadRequest, "invalid body"); return }
	id := uuid.New().String()
	_, err := h.db.Exec(`INSERT INTO settlement_monitoring_points (id, project_id, point_code, point_name, point_type, chainage_m, offset_m, initial_level_m, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$9)`, id, input.ProjectID, input.PointCode, input.PointName, input.PointType, input.ChainageM, input.OffsetM, input.InitialLevelM, time.Now())
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}
func (h *SettlementGroutingVentHandler) GetSettlementPoint(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var code string
	err := h.db.QueryRow(`SELECT point_code FROM settlement_monitoring_points WHERE id = $1`, id).Scan(&code)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "point not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "point_code": code})
}
func (h *SettlementGroutingVentHandler) UpdateSettlementPoint(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct{ Status *string `json:"status"` }
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil { respondError(w, http.StatusBadRequest, "invalid body"); return }
	_, err := h.db.Exec(`UPDATE settlement_monitoring_points SET status=COALESCE($1,status), updated_at=$2 WHERE id=$3`, input.Status, time.Now(), id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}
func (h *SettlementGroutingVentHandler) ListSettlementReadings(w http.ResponseWriter, r *http.Request) {
	pID := chi.URLParam(r, "pId")
	rows, err := h.db.Query(`SELECT id, point_id, reading_time, level_m, settlement_mm, rate_mm_per_day, is_alert, is_urgent, notes, created_at FROM settlement_readings WHERE point_id = $1 ORDER BY reading_time DESC LIMIT 200`, pID)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()
	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, pid, notes string
		var rt time.Time; var level, settle, rate float64; var alert, urgent bool; var crAt time.Time
		if err := rows.Scan(&id, &pid, &rt, &level, &settle, &rate, &alert, &urgent, &notes, &crAt); err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
		items = append(items, map[string]interface{}{"id": id, "point_id": pid, "reading_time": rt, "level_m": level, "settlement_mm": settle, "rate_mm_per_day": rate, "is_alert": alert, "is_urgent": urgent, "notes": notes, "created_at": crAt})
	}
	respondJSON(w, http.StatusOK, items)
}
func (h *SettlementGroutingVentHandler) CreateSettlementReading(w http.ResponseWriter, r *http.Request) {
	pID := chi.URLParam(r, "pId")
	var input struct {
		LevelM    float64 `json:"level_m"`
		SettleMM  float64 `json:"settlement_mm"`
		RatePerDay float64 `json:"rate_mm_per_day"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil { respondError(w, http.StatusBadRequest, "invalid body"); return }
	// Check triggers
	var alert, urgent float64
	h.db.QueryRow(`SELECT trigger_alert_mm, trigger_urgent_mm FROM settlement_monitoring_points WHERE id = $1`, pID).Scan(&alert, &urgent)
	isAlert := input.SettleMM >= alert
	isUrgent := input.SettleMM >= urgent
	id := uuid.New().String(); now := time.Now()
	_, err := h.db.Exec(`INSERT INTO settlement_readings (id, point_id, reading_time, level_m, settlement_mm, rate_mm_per_day, is_alert, is_urgent, created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`, id, pID, now, input.LevelM, input.SettleMM, input.RatePerDay, isAlert, isUrgent, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id, "alert": fmt.Sprintf("%v", isAlert), "urgent": fmt.Sprintf("%v", isUrgent)})
}

// ========== GROUTING ==========
func (h *SettlementGroutingVentHandler) ListGrouting(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	q := `SELECT id, project_id, grout_code, grout_name, grout_type, chainage_from_m, chainage_to_m, mix_design, target_pressure_bar, target_volume_m3, actual_volume_m3, status, start_date, end_date, supervisor, notes, created_at FROM grouting_activities WHERE 1=1`
	var args []interface{}; ai := 1
	if projectID != "" { q += fmt.Sprintf(" AND project_id = $%d", ai); args = append(args, projectID); ai++ }
	q += " ORDER BY chainage_from_m"
	rows, err := h.db.Query(q, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()
	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, pid, code, name, gtype, mix, status, sup, notes string
		var fromM, toM, tPress, tVol, aVol float64
		var sd, ed, crAt sql.NullString
		if err := rows.Scan(&id, &pid, &code, &name, &gtype, &fromM, &toM, &mix, &tPress, &tVol, &aVol, &status, &sd, &ed, &sup, &notes, &crAt); err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
		item := map[string]interface{}{"id": id, "project_id": pid, "grout_code": code, "grout_name": name, "grout_type": gtype, "chainage_from_m": fromM, "chainage_to_m": toM, "mix_design": mix, "target_pressure_bar": tPress, "target_volume_m3": tVol, "actual_volume_m3": aVol, "status": status, "supervisor": sup, "notes": notes}
		if sd.Valid { item["start_date"] = sd.String }; if ed.Valid { item["end_date"] = ed.String }
		items = append(items, item)
	}
	respondJSON(w, http.StatusOK, items)
}
func (h *SettlementGroutingVentHandler) CreateGrouting(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID  string `json:"project_id"`
		GroutCode  string `json:"grout_code"`
		GroutName  string `json:"grout_name"`
		GroutType  string `json:"grout_type"`
		FromM      float64 `json:"chainage_from_m"`
		ToM        float64 `json:"chainage_to_m"`
		TargetVol  float64 `json:"target_volume_m3"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil { respondError(w, http.StatusBadRequest, "invalid body"); return }
	id := uuid.New().String()
	_, err := h.db.Exec(`INSERT INTO grouting_activities (id, project_id, grout_code, grout_name, grout_type, chainage_from_m, chainage_to_m, target_volume_m3, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$9)`, id, input.ProjectID, input.GroutCode, input.GroutName, input.GroutType, input.FromM, input.ToM, input.TargetVol, time.Now())
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}
func (h *SettlementGroutingVentHandler) GetGrouting(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var code string
	err := h.db.QueryRow(`SELECT grout_code FROM grouting_activities WHERE id = $1`, id).Scan(&code)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "grouting not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "grout_code": code})
}
func (h *SettlementGroutingVentHandler) UpdateGrouting(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		ActualVolume *float64 `json:"actual_volume_m3"`
		Status       *string  `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil { respondError(w, http.StatusBadRequest, "invalid body"); return }
	_, err := h.db.Exec(`UPDATE grouting_activities SET actual_volume_m3=COALESCE($1,actual_volume_m3), status=COALESCE($2,status), updated_at=$3 WHERE id=$4`, input.ActualVolume, input.Status, time.Now(), id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}
func (h *SettlementGroutingVentHandler) ListGroutingRecords(w http.ResponseWriter, r *http.Request) {
	gID := chi.URLParam(r, "gId")
	rows, err := h.db.Query(`SELECT id, grout_id, record_time, pressure_bar, flow_rate_lpm, volume_m3, density_kg_m3, temperature, notes, created_at FROM grouting_records WHERE grout_id = $1 ORDER BY record_time`, gID)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()
	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, gid, notes string
		var rt time.Time; var press, flow, vol, dens, temp float64; var crAt time.Time
		if err := rows.Scan(&id, &gid, &rt, &press, &flow, &vol, &dens, &temp, &notes, &crAt); err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
		items = append(items, map[string]interface{}{"id": id, "grout_id": gid, "record_time": rt, "pressure_bar": press, "flow_rate_lpm": flow, "volume_m3": vol, "density_kg_m3": dens, "temperature": temp, "notes": notes, "created_at": crAt})
	}
	respondJSON(w, http.StatusOK, items)
}
func (h *SettlementGroutingVentHandler) CreateGroutingRecord(w http.ResponseWriter, r *http.Request) {
	gID := chi.URLParam(r, "gId")
	var input struct {
		Pressure float64 `json:"pressure_bar"`
		Flow     float64 `json:"flow_rate_lpm"`
		Volume   float64 `json:"volume_m3"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil { respondError(w, http.StatusBadRequest, "invalid body"); return }
	id := uuid.New().String(); now := time.Now()
	_, err := h.db.Exec(`INSERT INTO grouting_records (id, grout_id, record_time, pressure_bar, flow_rate_lpm, volume_m3, created_at) VALUES ($1,$2,$3,$4,$5,$6,$7)`, id, gID, now, input.Pressure, input.Flow, input.Volume, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

// ========== VENTILATION ==========
func (h *SettlementGroutingVentHandler) ListVentilation(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	q := `SELECT id, project_id, system_code, system_name, vent_type, fan_count, fan_power_kw, airflow_m3_s, duct_diameter_mm, duct_length_m, chainage_m, status, installed_date, last_maintenance, notes, created_at FROM ventilation_systems WHERE 1=1`
	var args []interface{}; ai := 1
	if projectID != "" { q += fmt.Sprintf(" AND project_id = $%d", ai); args = append(args, projectID); ai++ }
	q += " ORDER BY system_code"
	rows, err := h.db.Query(q, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()
	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, pid, code, name, vtype, status, notes string
		var fans int; var power, airflow float64; var diam, length, chain float64
		var instDate, lastMaint, crAt sql.NullString
		if err := rows.Scan(&id, &pid, &code, &name, &vtype, &fans, &power, &airflow, &diam, &length, &chain, &status, &instDate, &lastMaint, &notes, &crAt); err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
		item := map[string]interface{}{"id": id, "project_id": pid, "system_code": code, "system_name": name, "vent_type": vtype, "fan_count": fans, "fan_power_kw": power, "airflow_m3_s": airflow, "duct_diameter_mm": diam, "duct_length_m": length, "chainage_m": chain, "status": status, "notes": notes}
		if instDate.Valid { item["installed_date"] = instDate.String }; if lastMaint.Valid { item["last_maintenance"] = lastMaint.String }
		items = append(items, item)
	}
	respondJSON(w, http.StatusOK, items)
}
func (h *SettlementGroutingVentHandler) CreateVentilation(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID  string `json:"project_id"`
		SystemCode string `json:"system_code"`
		SystemName string `json:"system_name"`
		VentType   string `json:"vent_type"`
		Airflow    float64 `json:"airflow_m3_s"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil { respondError(w, http.StatusBadRequest, "invalid body"); return }
	id := uuid.New().String()
	_, err := h.db.Exec(`INSERT INTO ventilation_systems (id, project_id, system_code, system_name, vent_type, airflow_m3_s, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$7)`, id, input.ProjectID, input.SystemCode, input.SystemName, input.VentType, input.Airflow, time.Now())
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}
func (h *SettlementGroutingVentHandler) GetVentilation(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var code string
	err := h.db.QueryRow(`SELECT system_code FROM ventilation_systems WHERE id = $1`, id).Scan(&code)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "ventilation not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "system_code": code})
}
func (h *SettlementGroutingVentHandler) UpdateVentilation(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct{ Status *string `json:"status"` }
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil { respondError(w, http.StatusBadRequest, "invalid body"); return }
	_, err := h.db.Exec(`UPDATE ventilation_systems SET status=COALESCE($1,status), updated_at=$2 WHERE id=$3`, input.Status, time.Now(), id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}
func (h *SettlementGroutingVentHandler) ListVentReadings(w http.ResponseWriter, r *http.Request) {
	vID := chi.URLParam(r, "vId")
	rows, err := h.db.Query(`SELECT id, system_id, reading_time, airflow_m3_s, temperature_c, humidity_pct, co_ppm, co2_ppm, no2_ppm, dust_mg_m3, fan_speed_pct, power_kw, notes, created_at FROM ventilation_readings WHERE system_id = $1 ORDER BY reading_time DESC LIMIT 200`, vID)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()
	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, sid, notes string
		var rt, crAt time.Time; var airflow, temp, hum, co, co2, no2, dust, fanSpd, power float64
		if err := rows.Scan(&id, &sid, &rt, &airflow, &temp, &hum, &co, &co2, &no2, &dust, &fanSpd, &power, &notes, &crAt); err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
		items = append(items, map[string]interface{}{"id": id, "system_id": sid, "reading_time": rt, "airflow_m3_s": airflow, "temperature_c": temp, "humidity_pct": hum, "co_ppm": co, "co2_ppm": co2, "no2_ppm": no2, "dust_mg_m3": dust, "fan_speed_pct": fanSpd, "power_kw": power, "notes": notes, "created_at": crAt})
	}
	respondJSON(w, http.StatusOK, items)
}
func (h *SettlementGroutingVentHandler) CreateVentReading(w http.ResponseWriter, r *http.Request) {
	vID := chi.URLParam(r, "vId")
	var input struct {
		Airflow   float64 `json:"airflow_m3_s"`
		Temp      float64 `json:"temperature_c"`
		Humidity  float64 `json:"humidity_pct"`
		CO        float64 `json:"co_ppm"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil { respondError(w, http.StatusBadRequest, "invalid body"); return }
	id := uuid.New().String(); now := time.Now()
	_, err := h.db.Exec(`INSERT INTO ventilation_readings (id, system_id, reading_time, airflow_m3_s, temperature_c, humidity_pct, co_ppm, created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`, id, vID, now, input.Airflow, input.Temp, input.Humidity, input.CO, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *SettlementGroutingVentHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	q := `SELECT project_id, settlement_points, active_settlement_points, settlement_readings, settlement_alerts, settlement_urgent, grouting_activities, active_grouting, total_grout_volume_m3, ventilation_systems, ventilation_readings FROM settlement_grouting_summary`
	var args []interface{}; ai := 1
	if projectID != "" { q += fmt.Sprintf(" WHERE project_id = $%d", ai); args = append(args, projectID) }
	rows, err := h.db.Query(q, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()
	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var pid string
		var sp, asp, sr, sa, su, ga, ag, gv, vs, vr int
		if err := rows.Scan(&pid, &sp, &asp, &sr, &sa, &su, &ga, &ag, &gv, &vs, &vr); err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
		items = append(items, map[string]interface{}{"project_id": pid, "settlement_points": sp, "active_settlement_points": asp, "settlement_readings": sr, "settlement_alerts": sa, "settlement_urgent": su, "grouting_activities": ga, "active_grouting": ag, "total_grout_volume_m3": gv, "ventilation_systems": vs, "ventilation_readings": vr})
	}
	respondJSON(w, http.StatusOK, items)
}