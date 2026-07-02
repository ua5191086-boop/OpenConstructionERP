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

type TBMServiceHandler struct{ db *sql.DB }
func NewTBMServiceHandler(db *sql.DB) *TBMServiceHandler { return &TBMServiceHandler{db: db} }

func (h *TBMServiceHandler) RegisterRoutes(r chi.Router) {
	r.Route("/tunnel-services", func(r chi.Router) {
		r.Get("/sensors", h.ListSensors); r.Post("/sensors", h.CreateSensor)
		r.Get("/sensors/{id}", h.GetSensor); r.Put("/sensors/{id}", h.UpdateSensor)
		r.Get("/sensors/{sid}/readings", h.ListSensorReadings)
		r.Post("/sensors/{sid}/readings", h.CreateSensorReading)

		r.Get("/dewatering-wells", h.ListWells); r.Post("/dewatering-wells", h.CreateWell)
		r.Get("/dewatering-wells/{id}", h.GetWell); r.Put("/dewatering-wells/{id}", h.UpdateWell)
		r.Get("/dewatering-wells/{wid}/readings", h.ListWellReadings)
		r.Post("/dewatering-wells/{wid}/readings", h.CreateWellReading)

		r.Get("/tbm-maintenance", h.ListMaintTasks); r.Post("/tbm-maintenance", h.CreateMaintTask)
		r.Get("/tbm-maintenance/{id}", h.GetMaintTask); r.Put("/tbm-maintenance/{id}", h.UpdateMaintTask)
		r.Get("/tbm-maintenance/{tid}/logs", h.ListMaintLogs)
		r.Post("/tbm-maintenance/{tid}/logs", h.CreateMaintLog)

		r.Get("/summary", h.GetSummary)
	})
}

// ===== SENSORS =====
func (h *TBMServiceHandler) ListSensors(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id"); q := `SELECT id,project_id,sensor_code,sensor_name,sensor_type,chainage_m,location,install_date,reading_unit,is_active,notes,created_at FROM instrumentation_sensors WHERE 1=1`
	var args []interface{}; ai := 1
	if projectID != "" { q += fmt.Sprintf(" AND project_id = $%d", ai); args = append(args, projectID) }
	q += " ORDER BY sensor_code"
	rows, err := h.db.Query(q, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()
	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, pid, code, name, stype, loc, unit, notes string
		var chain float64; var instDate, crAt sql.NullString; var active bool
		if err := rows.Scan(&id, &pid, &code, &name, &stype, &chain, &loc, &instDate, &unit, &active, &notes, &crAt); err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
		items = append(items, map[string]interface{}{"id": id, "project_id": pid, "sensor_code": code, "sensor_name": name, "sensor_type": stype, "chainage_m": chain, "location": loc, "reading_unit": unit, "is_active": active, "notes": notes})
	}
	respondJSON(w, http.StatusOK, items)
}
func (h *TBMServiceHandler) CreateSensor(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID  string `json:"project_id"`; SensorCode string `json:"sensor_code"`
		SensorName string `json:"sensor_name"`; SensorType string `json:"sensor_type"`
		ChainageM  float64 `json:"chainage_m"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil { respondError(w, http.StatusBadRequest, "invalid body"); return }
	id := uuid.New().String()
	_, err := h.db.Exec(`INSERT INTO instrumentation_sensors (id,project_id,sensor_code,sensor_name,sensor_type,chainage_m,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$7)`, id, input.ProjectID, input.SensorCode, input.SensorName, input.SensorType, input.ChainageM, time.Now())
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}
func (h *TBMServiceHandler) GetSensor(w http.ResponseWriter, r *http.Request) { id := chi.URLParam(r, "id")
	var code string; err := h.db.QueryRow(`SELECT sensor_code FROM instrumentation_sensors WHERE id=$1`, id).Scan(&code)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "sensor not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "sensor_code": code})
}
func (h *TBMServiceHandler) UpdateSensor(w http.ResponseWriter, r *http.Request) { id := chi.URLParam(r, "id")

	var input struct{ IsActive *bool `json:"is_active"` }
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil { respondError(w, http.StatusBadRequest, "invalid body"); return }
	_, err := h.db.Exec(`UPDATE instrumentation_sensors SET is_active=COALESCE($1,is_active),updated_at=$2 WHERE id=$3`, input.IsActive, time.Now(), id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}
func (h *TBMServiceHandler) ListSensorReadings(w http.ResponseWriter, r *http.Request) {
	sid := chi.URLParam(r, "sid")
	rows, err := h.db.Query(`SELECT id,sensor_id,reading_time,value,unit,temperature,is_alarm,battery_level,notes,created_at FROM instrumentation_readings WHERE sensor_id=$1 ORDER BY reading_time DESC LIMIT 500`, sid)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()
	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, sid2, unit, notes string; var rt, crAt time.Time; var val, temp, batt float64; var alarm bool
		if err := rows.Scan(&id, &sid2, &rt, &val, &unit, &temp, &alarm, &batt, &notes, &crAt); err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
		items = append(items, map[string]interface{}{"id": id, "sensor_id": sid2, "reading_time": rt, "value": val, "unit": unit, "temperature": temp, "is_alarm": alarm, "battery_level": batt, "notes": notes, "created_at": crAt})
	}
	respondJSON(w, http.StatusOK, items)
}
func (h *TBMServiceHandler) CreateSensorReading(w http.ResponseWriter, r *http.Request) {
	sid := chi.URLParam(r, "sid")
	var input struct{ Value float64 `json:"value"`; Temp float64 `json:"temperature"`; Batt float64 `json:"battery_level"`; Unit string `json:"unit"` }
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil { respondError(w, http.StatusBadRequest, "invalid body"); return }
	var alarmLow, alarmHigh float64; h.db.QueryRow(`SELECT alarm_low,alarm_high FROM instrumentation_sensors WHERE id=$1`, sid).Scan(&alarmLow, &alarmHigh)
	alarm := (alarmHigh > 0 && input.Value > alarmHigh) || (alarmLow > 0 && input.Value < alarmLow)
	id := uuid.New().String(); now := time.Now()
	_, err := h.db.Exec(`INSERT INTO instrumentation_readings (id,sensor_id,reading_time,value,unit,temperature,is_alarm,battery_level,created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`, id, sid, now, input.Value, input.Unit, input.Temp, alarm, input.Batt, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id, "alarm": fmt.Sprintf("%v", alarm)})
}

// ===== DEWATERING =====
func (h *TBMServiceHandler) ListWells(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id"); q := `SELECT id,project_id,well_code,well_name,well_type,chainage_m,depth_m,pump_capacity_m3h,static_water_level_m,status,installation_date,notes,created_at FROM dewatering_wells WHERE 1=1`
	var args []interface{}; ai := 1
	if projectID != "" { q += fmt.Sprintf(" AND project_id = $%d", ai); args = append(args, projectID) }
	q += " ORDER BY well_code"
	rows, err := h.db.Query(q, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()
	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, pid, code, name, wtype, status, notes string
		var chain, depth, cap, swl float64; var instDate, crAt sql.NullString
		if err := rows.Scan(&id, &pid, &code, &name, &wtype, &chain, &depth, &cap, &swl, &status, &instDate, &notes, &crAt); err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
		items = append(items, map[string]interface{}{"id": id, "project_id": pid, "well_code": code, "well_name": name, "well_type": wtype, "chainage_m": chain, "depth_m": depth, "pump_capacity_m3h": cap, "static_water_level_m": swl, "status": status, "notes": notes})
	}
	respondJSON(w, http.StatusOK, items)
}
func (h *TBMServiceHandler) CreateWell(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID string `json:"project_id"`; WellCode string `json:"well_code"`; WellName string `json:"well_name"`
		WellType  string `json:"well_type"`; ChainageM float64 `json:"chainage_m"`; DepthM float64 `json:"depth_m"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil { respondError(w, http.StatusBadRequest, "invalid body"); return }
	id := uuid.New().String()
	_, err := h.db.Exec(`INSERT INTO dewatering_wells (id,project_id,well_code,well_name,well_type,chainage_m,depth_m,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$8)`, id, input.ProjectID, input.WellCode, input.WellName, input.WellType, input.ChainageM, input.DepthM, time.Now())
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}
func (h *TBMServiceHandler) GetWell(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id"); var code string
	err := h.db.QueryRow(`SELECT well_code FROM dewatering_wells WHERE id=$1`, id).Scan(&code)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "well not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "well_code": code})
}
func (h *TBMServiceHandler) UpdateWell(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id"); var input struct{ Status *string `json:"status"` }
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil { respondError(w, http.StatusBadRequest, "invalid body"); return }
	_, err := h.db.Exec(`UPDATE dewatering_wells SET status=COALESCE($1,status),updated_at=$2 WHERE id=$3`, input.Status, time.Now(), id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}
func (h *TBMServiceHandler) ListWellReadings(w http.ResponseWriter, r *http.Request) {
	wid := chi.URLParam(r, "wid")
	rows, err := h.db.Query(`SELECT id,well_id,reading_time,water_level_m,flow_rate_m3h,pump_running,pump_pressure_bar,energy_kwh,sediment_pct,notes,created_at FROM dewatering_readings WHERE well_id=$1 ORDER BY reading_time DESC LIMIT 200`, wid)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()
	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, wid2, notes string; var rt, crAt time.Time; var wl, flow, press, energy, sed float64; var pump bool
		if err := rows.Scan(&id, &wid2, &rt, &wl, &flow, &pump, &press, &energy, &sed, &notes, &crAt); err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
		items = append(items, map[string]interface{}{"id": id, "well_id": wid2, "reading_time": rt, "water_level_m": wl, "flow_rate_m3h": flow, "pump_running": pump, "pump_pressure_bar": press, "energy_kwh": energy, "sediment_pct": sed, "notes": notes, "created_at": crAt})
	}
	respondJSON(w, http.StatusOK, items)
}
func (h *TBMServiceHandler) CreateWellReading(w http.ResponseWriter, r *http.Request) {
	wid := chi.URLParam(r, "wid")
	var input struct{ WaterLevel float64 `json:"water_level_m"`; FlowRate float64 `json:"flow_rate_m3h"`; PumpOn bool `json:"pump_running"` }
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil { respondError(w, http.StatusBadRequest, "invalid body"); return }
	id := uuid.New().String(); now := time.Now()
	_, err := h.db.Exec(`INSERT INTO dewatering_readings (id,well_id,reading_time,water_level_m,flow_rate_m3h,pump_running,created_at) VALUES ($1,$2,$3,$4,$5,$6,$7)`, id, wid, now, input.WaterLevel, input.FlowRate, input.PumpOn, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

// ===== TBM MAINTENANCE =====
func (h *TBMServiceHandler) ListMaintTasks(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id"); tbmID := r.URL.Query().Get("tbm_id")
	q := `SELECT id,project_id,tbm_id,task_code,task_name,task_type,component,priority,interval_ring,interval_days,estimated_hours,actual_hours,assigned_crew,status,notes,created_at FROM tbm_maintenance_tasks WHERE 1=1`
	var args []interface{}; ai := 1
	if projectID != "" { q += fmt.Sprintf(" AND project_id = $%d", ai); args = append(args, projectID); ai++ }
	if tbmID != "" { q += fmt.Sprintf(" AND tbm_id = $%d", ai); args = append(args, tbmID); ai++ }
	q += " ORDER BY task_code"
	rows, err := h.db.Query(q, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()
	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, pid, tid, code, name, ttype, comp, prio, crew, status, notes string
		var intRing, intDays int; var estHrs, actHrs float64
		var crAt time.Time
		if err := rows.Scan(&id, &pid, &tid, &code, &name, &ttype, &comp, &prio, &intRing, &intDays, &estHrs, &actHrs, &crew, &status, &notes, &crAt); err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
		items = append(items, map[string]interface{}{"id": id, "project_id": pid, "tbm_id": tid, "task_code": code, "task_name": name, "task_type": ttype, "component": comp, "priority": prio, "interval_ring": intRing, "interval_days": intDays, "estimated_hours": estHrs, "actual_hours": actHrs, "assigned_crew": crew, "status": status, "notes": notes, "created_at": crAt})
	}
	respondJSON(w, http.StatusOK, items)
}
func (h *TBMServiceHandler) CreateMaintTask(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID string `json:"project_id"`; TbmID string `json:"tbm_id"`; TaskCode string `json:"task_code"`
		TaskName  string `json:"task_name"`; TaskType string `json:"task_type"`; Component string `json:"component"`
		Priority  string `json:"priority"`; IntervalRing int `json:"interval_ring"`; IntervalDays int `json:"interval_days"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil { respondError(w, http.StatusBadRequest, "invalid body"); return }
	id := uuid.New().String()
	_, err := h.db.Exec(`INSERT INTO tbm_maintenance_tasks (id,project_id,tbm_id,task_code,task_name,task_type,component,priority,interval_ring,interval_days,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$11)`, id, input.ProjectID, input.TbmID, input.TaskCode, input.TaskName, input.TaskType, input.Component, input.Priority, input.IntervalRing, input.IntervalDays, time.Now())
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}
func (h *TBMServiceHandler) GetMaintTask(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id"); var code string
	err := h.db.QueryRow(`SELECT task_code FROM tbm_maintenance_tasks WHERE id=$1`, id).Scan(&code)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "task not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "task_code": code})
}
func (h *TBMServiceHandler) UpdateMaintTask(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct{ Status *string `json:"status"`; ActualHours *float64 `json:"actual_hours"` }
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil { respondError(w, http.StatusBadRequest, "invalid body"); return }
	_, err := h.db.Exec(`UPDATE tbm_maintenance_tasks SET status=COALESCE($1,status),actual_hours=COALESCE($2,actual_hours),completion_date=CASE WHEN $1='completed' THEN NOW() ELSE completion_date END,updated_at=$3 WHERE id=$4`, input.Status, input.ActualHours, time.Now(), id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}
func (h *TBMServiceHandler) ListMaintLogs(w http.ResponseWriter, r *http.Request) {
	tid := chi.URLParam(r, "tid")
	rows, err := h.db.Query(`SELECT id,task_id,log_time,ring_number,action,duration_hours,parts_used,downtime_hours,performed_by,notes,created_at FROM tbm_maintenance_logs WHERE task_id=$1 ORDER BY log_time DESC LIMIT 100`, tid)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()
	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, tid2, action, parts, crew, notes string; var rt, crAt time.Time; var ring int; var dur, down float64
		if err := rows.Scan(&id, &tid2, &rt, &ring, &action, &dur, &parts, &down, &crew, &notes, &crAt); err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
		items = append(items, map[string]interface{}{"id": id, "task_id": tid2, "log_time": rt, "ring_number": ring, "action": action, "duration_hours": dur, "parts_used": parts, "downtime_hours": down, "performed_by": crew, "notes": notes, "created_at": crAt})
	}
	respondJSON(w, http.StatusOK, items)
}
func (h *TBMServiceHandler) CreateMaintLog(w http.ResponseWriter, r *http.Request) {
	tid := chi.URLParam(r, "tid")
	var input struct{ Ring int `json:"ring_number"`; Action string `json:"action"`; Dur float64 `json:"duration_hours"`; Parts string `json:"parts_used"`; Down float64 `json:"downtime_hours"`; Crew string `json:"performed_by"` }
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil { respondError(w, http.StatusBadRequest, "invalid body"); return }
	id := uuid.New().String(); now := time.Now()
	_, err := h.db.Exec(`INSERT INTO tbm_maintenance_logs (id,task_id,log_time,ring_number,action,duration_hours,parts_used,downtime_hours,performed_by,created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`, id, tid, now, input.Ring, input.Action, input.Dur, input.Parts, input.Down, input.Crew, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *TBMServiceHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	q := `SELECT project_id,active_sensors,sensor_readings,sensor_alarms,active_dewatering_wells,dewatering_readings,tbm_maintenance_tasks,pending_maintenance FROM tunnel_services_summary`
	var args []interface{}; ai := 1
	if projectID != "" { q += fmt.Sprintf(" WHERE project_id = $%d", ai); args = append(args, projectID) }
	rows, err := h.db.Query(q, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()
	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var pid string; var sens, reads, alarms, wells, dwells, mtasks, pending int
		if err := rows.Scan(&pid, &sens, &reads, &alarms, &wells, &dwells, &mtasks, &pending); err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
		items = append(items, map[string]interface{}{"project_id": pid, "active_sensors": sens, "sensor_readings": reads, "sensor_alarms": alarms, "active_dewatering_wells": wells, "dewatering_readings": dwells, "tbm_maintenance_tasks": mtasks, "pending_maintenance": pending})
	}
	respondJSON(w, http.StatusOK, items)
}