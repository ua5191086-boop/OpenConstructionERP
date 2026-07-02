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

type ShaftHandler struct{ db *sql.DB }

func NewShaftHandler(db *sql.DB) *ShaftHandler { return &ShaftHandler{db: db} }

func (h *ShaftHandler) RegisterRoutes(r chi.Router) {
	r.Route("/shafts", func(r chi.Router) {
		r.Get("/", h.ListProjects)
		r.Post("/", h.CreateProject)
		r.Get("/{id}", h.GetProject)
		r.Put("/{id}", h.UpdateProject)
		r.Delete("/{id}", h.DeleteProject)

		r.Get("/{shaftId}/sequences", h.ListSequences)
		r.Post("/{shaftId}/sequences", h.CreateSequence)
		r.Get("/sequences/{id}", h.GetSequence)
		r.Put("/sequences/{id}", h.UpdateSequence)

		r.Get("/{shaftId}/instruments", h.ListInstruments)
		r.Post("/{shaftId}/instruments", h.CreateInstrument)
		r.Get("/instruments/{id}", h.GetInstrument)
		r.Put("/instruments/{id}", h.UpdateInstrument)

		r.Get("/instruments/{instrumentId}/readings", h.ListReadings)
		r.Post("/instruments/{instrumentId}/readings", h.CreateReading)

		r.Get("/summary", h.GetSummary)
	})
}

func (h *ShaftHandler) ListProjects(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	q := `SELECT id, project_id, shaft_code, shaft_name, shaft_type, construction_method, diameter_m, depth_m, status, start_date, end_date, notes, created_at FROM shaft_projects WHERE 1=1`
	var args []interface{}
	argIdx := 1
	if projectID != "" { q += fmt.Sprintf(" AND project_id = $%d", argIdx); args = append(args, projectID) }
	q += " ORDER BY shaft_code"
	rows, err := h.db.Query(q, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()
	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, pid, code, name, stype, method, status, notes string
		var diam, depth float64
		var sd, ed, crAt sql.NullString
		if err := rows.Scan(&id, &pid, &code, &name, &stype, &method, &diam, &depth, &status, &sd, &ed, &notes, &crAt); err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
		item := map[string]interface{}{"id": id, "project_id": pid, "shaft_code": code, "shaft_name": name, "shaft_type": stype, "construction_method": method, "diameter_m": diam, "depth_m": depth, "status": status, "notes": notes}
		if sd.Valid { item["start_date"] = sd.String }
		if ed.Valid { item["end_date"] = ed.String }
		if crAt.Valid { item["created_at"] = crAt.String }
		items = append(items, item)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *ShaftHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID string  `json:"project_id"`
		ShaftCode string  `json:"shaft_code"`
		ShaftName string  `json:"shaft_name"`
		ShaftType string  `json:"shaft_type"`
		DiameterM *float64 `json:"diameter_m"`
		DepthM   *float64 `json:"depth_m"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil { respondError(w, http.StatusBadRequest, "invalid body"); return }
	id := uuid.New().String()
	_, err := h.db.Exec(`INSERT INTO shaft_projects (id, project_id, shaft_code, shaft_name, shaft_type, diameter_m, depth_m, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$8)`, id, input.ProjectID, input.ShaftCode, input.ShaftName, input.ShaftType, input.DiameterM, input.DepthM, time.Now())
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *ShaftHandler) GetProject(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var code, name string
	err := h.db.QueryRow(`SELECT shaft_code, shaft_name FROM shaft_projects WHERE id = $1`, id).Scan(&code, &name)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "shaft not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "shaft_code": code, "shaft_name": name})
}

func (h *ShaftHandler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct{ Status *string `json:"status"` }
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil { respondError(w, http.StatusBadRequest, "invalid body"); return }
	_, err := h.db.Exec(`UPDATE shaft_projects SET status=COALESCE($1,status), updated_at=$2 WHERE id=$3`, input.Status, time.Now(), id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *ShaftHandler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM shaft_projects WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (h *ShaftHandler) ListSequences(w http.ResponseWriter, r *http.Request) {
	shaftID := chi.URLParam(r, "shaftId")
	rows, err := h.db.Query(`SELECT id, shaft_id, sequence_number, sequence_name, sequence_type, volume_m3, concrete_m3, reinforcement_kg, status, start_date, end_date, notes, created_at FROM shaft_construction_sequences WHERE shaft_id = $1 ORDER BY sequence_number`, shaftID)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()
	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, sid, sname, stype, status, notes string
		var seqNum int
		var vol, conc, reinf float64
		var sd, ed, crAt sql.NullString
		if err := rows.Scan(&id, &sid, &seqNum, &sname, &stype, &vol, &conc, &reinf, &status, &sd, &ed, &notes, &crAt); err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
		item := map[string]interface{}{"id": id, "shaft_id": sid, "sequence_number": seqNum, "sequence_name": sname, "sequence_type": stype, "volume_m3": vol, "concrete_m3": conc, "reinforcement_kg": reinf, "status": status, "notes": notes}
		if sd.Valid { item["start_date"] = sd.String }
		if ed.Valid { item["end_date"] = ed.String }
		items = append(items, item)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *ShaftHandler) CreateSequence(w http.ResponseWriter, r *http.Request) {
	shaftID := chi.URLParam(r, "shaftId")
	var input struct {
		SequenceName string  `json:"sequence_name"`
		SequenceType string  `json:"sequence_type"`
		VolumeM3     *float64 `json:"volume_m3"`
		ConcreteM3   *float64 `json:"concrete_m3"`
		ReinfKg      *float64 `json:"reinforcement_kg"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil { respondError(w, http.StatusBadRequest, "invalid body"); return }
	id := uuid.New().String()
	_, err := h.db.Exec(`INSERT INTO shaft_construction_sequences (id, shaft_id, sequence_number, sequence_name, sequence_type, volume_m3, concrete_m3, reinforcement_kg, created_at, updated_at) VALUES ($1,$2,(SELECT COALESCE(MAX(sequence_number),0)+1 FROM shaft_construction_sequences WHERE shaft_id=$2),$3,$4,$5,$6,$7,$8,$8)`, id, shaftID, input.SequenceName, input.SequenceType, input.VolumeM3, input.ConcreteM3, input.ReinfKg, time.Now())
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}
func (h *ShaftHandler) GetSequence(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var name string
	err := h.db.QueryRow(`SELECT sequence_name FROM shaft_construction_sequences WHERE id = $1`, id).Scan(&name)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "sequence not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "sequence_name": name})
}
func (h *ShaftHandler) UpdateSequence(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct{ Status *string `json:"status"` }
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil { respondError(w, http.StatusBadRequest, "invalid body"); return }
	_, err := h.db.Exec(`UPDATE shaft_construction_sequences SET status=COALESCE($1,status), updated_at=$2 WHERE id=$3`, input.Status, time.Now(), id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *ShaftHandler) ListInstruments(w http.ResponseWriter, r *http.Request) {
	shaftID := chi.URLParam(r, "shaftId")
	rows, err := h.db.Query(`SELECT id, shaft_id, instrument_code, instrument_type, elevation_m, depth_m, install_date, reading_interval_hours, alarm_threshold, status, notes, created_at FROM shaft_instrumentation WHERE shaft_id = $1 ORDER BY instrument_code`, shaftID)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()
	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, sid, code, itype, status, notes string
		var elev, depth, interval, alarm float64
		var installDate sql.NullString
		var crAt time.Time
		if err := rows.Scan(&id, &sid, &code, &itype, &elev, &depth, &installDate, &interval, &alarm, &status, &notes, &crAt); err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
		item := map[string]interface{}{"id": id, "shaft_id": sid, "instrument_code": code, "instrument_type": itype, "elevation_m": elev, "depth_m": depth, "reading_interval_hours": interval, "alarm_threshold": alarm, "status": status, "notes": notes, "created_at": crAt}
		if installDate.Valid { item["install_date"] = installDate.String }
		items = append(items, item)
	}
	respondJSON(w, http.StatusOK, items)
}
func (h *ShaftHandler) CreateInstrument(w http.ResponseWriter, r *http.Request) {
	shaftID := chi.URLParam(r, "shaftId")
	var input struct {
		InstrumentCode string  `json:"instrument_code"`
		InstrumentType string  `json:"instrument_type"`
		ElevationM     float64 `json:"elevation_m"`
		DepthM         float64 `json:"depth_m"`
		AlarmThreshold float64 `json:"alarm_threshold"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil { respondError(w, http.StatusBadRequest, "invalid body"); return }
	id := uuid.New().String()
	_, err := h.db.Exec(`INSERT INTO shaft_instrumentation (id, shaft_id, instrument_code, instrument_type, elevation_m, depth_m, alarm_threshold, created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`, id, shaftID, input.InstrumentCode, input.InstrumentType, input.ElevationM, input.DepthM, input.AlarmThreshold, time.Now())
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}
func (h *ShaftHandler) GetInstrument(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var code, itype string
	err := h.db.QueryRow(`SELECT instrument_code, instrument_type FROM shaft_instrumentation WHERE id = $1`, id).Scan(&code, &itype)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "instrument not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "instrument_code": code, "instrument_type": itype})
}
func (h *ShaftHandler) UpdateInstrument(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct{ Status *string `json:"status"` }
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil { respondError(w, http.StatusBadRequest, "invalid body"); return }
	_, err := h.db.Exec(`UPDATE shaft_instrumentation SET status=COALESCE($1,status) WHERE id=$2`, input.Status, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *ShaftHandler) ListReadings(w http.ResponseWriter, r *http.Request) {
	instrumentID := chi.URLParam(r, "instrumentId")
	rows, err := h.db.Query(`SELECT id, instrument_id, reading_time, value, unit, temperature, is_alarm, notes, created_at FROM shaft_monitoring_readings WHERE instrument_id = $1 ORDER BY reading_time DESC LIMIT 500`, instrumentID)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()
	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, iid, unit, notes string
		var rt time.Time
		var val, temp float64
		var alarm bool
		var crAt time.Time
		if err := rows.Scan(&id, &iid, &rt, &val, &unit, &temp, &alarm, &notes, &crAt); err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
		items = append(items, map[string]interface{}{"id": id, "instrument_id": iid, "reading_time": rt, "value": val, "unit": unit, "temperature": temp, "is_alarm": alarm, "notes": notes, "created_at": crAt})
	}
	respondJSON(w, http.StatusOK, items)
}
func (h *ShaftHandler) CreateReading(w http.ResponseWriter, r *http.Request) {
	instrumentID := chi.URLParam(r, "instrumentId")
	var input struct {
		Value       float64 `json:"value"`
		Unit        string  `json:"unit"`
		Temperature *float64 `json:"temperature"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil { respondError(w, http.StatusBadRequest, "invalid body"); return }
	id := uuid.New().String()
	now := time.Now()
	// Check alarm threshold
	var threshold float64
	h.db.QueryRow(`SELECT alarm_threshold FROM shaft_instrumentation WHERE id = $1`, instrumentID).Scan(&threshold)
	alarm := threshold > 0 && input.Value > threshold
	_, err := h.db.Exec(`INSERT INTO shaft_monitoring_readings (id, instrument_id, reading_time, value, unit, temperature, is_alarm, created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`, id, instrumentID, now, input.Value, input.Unit, input.Temperature, alarm, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id, "alarm": fmt.Sprintf("%v", alarm)})
}

func (h *ShaftHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	q := `SELECT project_id, total_shafts, active_shafts, completed_shafts, total_excavated_m3, active_instruments, active_alarms FROM shaft_summary`
	var args []interface{}
	argIdx := 1
	if projectID != "" { q += fmt.Sprintf(" WHERE project_id = $%d", argIdx); args = append(args, projectID) }
	rows, err := h.db.Query(q, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()
	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var pid string
		var total, active, completed, instruments, alarms int
		var excav float64
		if err := rows.Scan(&pid, &total, &active, &completed, &excav, &instruments, &alarms); err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
		items = append(items, map[string]interface{}{"project_id": pid, "total_shafts": total, "active_shafts": active, "completed_shafts": completed, "total_excavated_m3": excav, "active_instruments": instruments, "active_alarms": alarms})
	}
	respondJSON(w, http.StatusOK, items)
}