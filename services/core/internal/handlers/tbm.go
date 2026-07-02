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

// TBMHandler handles TBM Management module endpoints
type TBMHandler struct {
	db *sql.DB
}

func NewTBMHandler(db *sql.DB) *TBMHandler {
	return &TBMHandler{db: db}
}

func (h *TBMHandler) RegisterRoutes(r chi.Router) {
	r.Route("/tbm", func(r chi.Router) {
		// Telemetry
		r.Get("/telemetry", h.ListTelemetry)
		r.Post("/telemetry", h.CreateTelemetry)
		r.Get("/telemetry/{id}", h.GetTelemetry)

		// Alarms
		r.Get("/alarms", h.ListAlarms)
		r.Post("/alarms", h.CreateAlarm)
		r.Get("/alarms/{id}", h.GetAlarm)
		r.Put("/alarms/{id}", h.UpdateAlarm)

		// Operators
		r.Get("/operators", h.ListOperators)
		r.Post("/operators", h.CreateOperator)
		r.Get("/operators/{id}", h.GetOperator)
		r.Put("/operators/{id}", h.UpdateOperator)
		r.Delete("/operators/{id}", h.DeleteOperator)

		// Shifts
		r.Get("/shifts", h.ListShifts)
		r.Post("/shifts", h.CreateShift)
		r.Get("/shifts/{id}", h.GetShift)

		// Consumables
		r.Get("/consumables", h.ListConsumables)
		r.Post("/consumables", h.CreateConsumable)
		r.Get("/consumables/{id}", h.GetConsumable)

		// Performance
		r.Get("/performance", h.ListPerformance)
		r.Post("/performance", h.CreatePerformance)
		r.Get("/performance/{id}", h.GetPerformance)
		r.Get("/summary", h.GetSummary)
	})
}

// =============================================================================
// Telemetry
// =============================================================================
func (h *TBMHandler) ListTelemetry(w http.ResponseWriter, r *http.Request) {
	tbmID := r.URL.Query().Get("tbm_id")
	limit := r.URL.Query().Get("limit")
	query := `SELECT id, tbm_id, recorded_at, epb_face_pressure_bar, epb_screw_speed_rpm, epb_chamber_pressure_bar,
		slurry_density_kgm3, slurry_flow_in_m3h, slurry_flow_out_m3h, slurry_pressure_bar,
		thrust_force_kN, thrust_speed_mmmin, torque_kNm, torque_pct,
		advance_rate_mmmin, advance_mm, face_pressure_bar,
		cutterhead_rpm, cutterhead_torque_kNm, cutterhead_wear_mm,
		tail_skin_grease_bar, articulation_angle_deg,
		belt_weight_kg, total_power_kw, data_source, is_valid
		FROM tbm_telemetry WHERE 1=1`
	args := []interface{}{}
	argIdx := 1
	if tbmID != "" { query += fmt.Sprintf(" AND tbm_id = $%d", argIdx); argIdx++; args = append(args, tbmID) }
	query += " ORDER BY recorded_at DESC"
	if limit != "" { query += fmt.Sprintf(" LIMIT %s", limit) } else { query += " LIMIT 100" }

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, tbmID, ds string
		var recordedAt time.Time
		var epbFP, epbSS, epbCP, slurryDens, slurryIn, slurryOut, slurryP sql.NullFloat64
		var thrust, thrustSpeed, torq, torqPct, advRate, adv, fp sql.NullFloat64
		var cutRPM, cutTorq, cutWear, grease, artic, belt, pwr sql.NullFloat64
		var valid bool
		err := rows.Scan(&id, &tbmID, &recordedAt, &epbFP, &epbSS, &epbCP,
			&slurryDens, &slurryIn, &slurryOut, &slurryP,
			&thrust, &thrustSpeed, &torq, &torqPct,
			&advRate, &adv, &fp,
			&cutRPM, &cutTorq, &cutWear,
			&grease, &artic,
			&belt, &pwr, &ds, &valid)
		if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
		item := map[string]interface{}{"id": id, "tbm_id": tbmID, "recorded_at": recordedAt, "data_source": ds, "is_valid": valid}
		if epbFP.Valid { item["epb_face_pressure_bar"] = epbFP.Float64 }
		if epbSS.Valid { item["epb_screw_speed_rpm"] = epbSS.Float64 }
		if thrust.Valid { item["thrust_force_kN"] = thrust.Float64 }
		if advRate.Valid { item["advance_rate_mmmin"] = advRate.Float64 }
		if torq.Valid { item["torque_kNm"] = torq.Float64 }
		if fp.Valid { item["face_pressure_bar"] = fp.Float64 }
		items = append(items, item)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *TBMHandler) CreateTelemetry(w http.ResponseWriter, r *http.Request) {
	var input struct {
		TbmID string `json:"tbm_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	_, err := h.db.Exec(`INSERT INTO tbm_telemetry (id, tbm_id) VALUES ($1,$2)`, id, input.TbmID)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *TBMHandler) GetTelemetry(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var tbmID string
	err := h.db.QueryRow(`SELECT tbm_id FROM tbm_telemetry WHERE id = $1`, id).Scan(&tbmID)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "telemetry not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"id": id, "tbm_id": tbmID})
}

// =============================================================================
// Alarms
// =============================================================================
func (h *TBMHandler) ListAlarms(w http.ResponseWriter, r *http.Request) {
	tbmID := r.URL.Query().Get("tbm_id")
	active := r.URL.Query().Get("active")
	query := `SELECT id, tbm_id, alarm_code, alarm_severity, alarm_name, description, triggered_at, acknowledged_at, acknowledged_by, cleared_at, cleared_by, is_active FROM tbm_alarms WHERE 1=1`
	args := []interface{}{}
	argIdx := 1
	if tbmID != "" { query += fmt.Sprintf(" AND tbm_id = $%d", argIdx); argIdx++; args = append(args, tbmID) }
	if active == "true" { query += " AND is_active = TRUE" }
	query += " ORDER BY triggered_at DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, tbmID, code, sev, name, desc string
		var trigAt, ackAt, clearedAt sql.NullString
		var ackBy, clearedBy sql.NullString
		var isActive bool
		if err := rows.Scan(&id, &tbmID, &code, &sev, &name, &desc, &trigAt, &ackAt, &ackBy, &clearedAt, &clearedBy, &isActive); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		item := map[string]interface{}{"id": id, "tbm_id": tbmID, "alarm_code": code, "alarm_severity": sev, "alarm_name": name, "is_active": isActive}
		if trigAt.Valid { item["triggered_at"] = trigAt.String }
		items = append(items, item)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *TBMHandler) CreateAlarm(w http.ResponseWriter, r *http.Request) {
	var input struct {
		TbmID     string `json:"tbm_id"`
		AlarmCode string `json:"alarm_code"`
		Severity  string `json:"alarm_severity"`
		AlarmName string `json:"alarm_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	_, err := h.db.Exec(`INSERT INTO tbm_alarms (id, tbm_id, alarm_code, alarm_severity, alarm_name) VALUES ($1,$2,$3,$4,$5)`,
		id, input.TbmID, input.AlarmCode, input.Severity, input.AlarmName)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *TBMHandler) GetAlarm(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var code, name string
	err := h.db.QueryRow(`SELECT alarm_code, alarm_name FROM tbm_alarms WHERE id = $1`, id).Scan(&code, &name)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "alarm not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "alarm_code": code, "alarm_name": name})
}

func (h *TBMHandler) UpdateAlarm(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		AcknowledgedAt *string `json:"acknowledged_at"`
		AcknowledgedBy *string `json:"acknowledged_by"`
		ClearedAt      *string `json:"cleared_at"`
		ClearedBy      *string `json:"cleared_by"`
		IsActive       *bool   `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	_, err := h.db.Exec(`UPDATE tbm_alarms SET acknowledged_at=COALESCE($1,acknowledged_at), acknowledged_by=COALESCE($2,acknowledged_by), cleared_at=COALESCE($3,cleared_at), cleared_by=COALESCE($4,cleared_by), is_active=COALESCE($5,is_active) WHERE id=$6`,
		input.AcknowledgedAt, input.AcknowledgedBy, input.ClearedAt, input.ClearedBy, input.IsActive, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// =============================================================================
// Operators
// =============================================================================
func (h *TBMHandler) ListOperators(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(`SELECT id, employee_id, full_name, qualification, certification_number, certification_expiry, tbm_types, phone, email, is_active FROM tbm_operators ORDER BY full_name`)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, empID, name, qual, cert, tbmTypes, phone, email string
		var certExp sql.NullString
		var active bool
		if err := rows.Scan(&id, &empID, &name, &qual, &cert, &certExp, &tbmTypes, &phone, &email, &active); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		item := map[string]interface{}{"id": id, "employee_id": empID, "full_name": name, "qualification": qual, "certification_number": cert, "tbm_types": tbmTypes, "phone": phone, "email": email, "is_active": active}
		if certExp.Valid { item["certification_expiry"] = certExp.String }
		items = append(items, item)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *TBMHandler) CreateOperator(w http.ResponseWriter, r *http.Request) {
	var input struct {
		EmployeeID string `json:"employee_id"`
		FullName   string `json:"full_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	_, err := h.db.Exec(`INSERT INTO tbm_operators (id, employee_id, full_name) VALUES ($1,$2,$3)`, id, input.EmployeeID, input.FullName)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *TBMHandler) GetOperator(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var name, empID string
	err := h.db.QueryRow(`SELECT full_name, employee_id FROM tbm_operators WHERE id = $1`, id).Scan(&name, &empID)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "operator not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "full_name": name, "employee_id": empID})
}

func (h *TBMHandler) UpdateOperator(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		IsActive *bool `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	_, err := h.db.Exec(`UPDATE tbm_operators SET is_active=COALESCE($1,is_active), updated_at=NOW() WHERE id=$2`, input.IsActive, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *TBMHandler) DeleteOperator(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM tbm_operators WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// Shifts
// =============================================================================
func (h *TBMHandler) ListShifts(w http.ResponseWriter, r *http.Request) {
	tbmID := r.URL.Query().Get("tbm_id")
	query := `SELECT s.id, s.tbm_id, s.shift_date, s.shift_label, s.operator_id, o.full_name, s.rings_built, s.advance_mm, s.downtime_minutes, s.downtime_reason, s.notes, s.start_time, s.end_time
		FROM tbm_shifts s LEFT JOIN tbm_operators o ON s.operator_id = o.id WHERE 1=1`
	args := []interface{}{}
	argIdx := 1
	if tbmID != "" { query += fmt.Sprintf(" AND s.tbm_id = $%d", argIdx); argIdx++; args = append(args, tbmID) }
	query += " ORDER BY s.shift_date DESC, s.shift_label"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, tbmID, label string
		var opID, opName, reason, notes sql.NullString
		var rings, advance, downtime int
		var shiftDate time.Time
		var startTime, endTime sql.NullString
		if err := rows.Scan(&id, &tbmID, &shiftDate, &label, &opID, &opName, &rings, &advance, &downtime, &reason, &notes, &startTime, &endTime); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		item := map[string]interface{}{
			"id": id, "tbm_id": tbmID, "shift_date": shiftDate, "shift_label": label,
			"rings_built": rings, "advance_mm": advance, "downtime_minutes": downtime,
		}
		if opName.Valid { item["operator_name"] = opName.String }
		items = append(items, item)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *TBMHandler) CreateShift(w http.ResponseWriter, r *http.Request) {
	var input struct {
		TbmID      string `json:"tbm_id"`
		ShiftDate  string `json:"shift_date"`
		ShiftLabel string `json:"shift_label"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	_, err := h.db.Exec(`INSERT INTO tbm_shifts (id, tbm_id, shift_date, shift_label) VALUES ($1,$2,$3,$4)`,
		id, input.TbmID, input.ShiftDate, input.ShiftLabel)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *TBMHandler) GetShift(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var label string
	err := h.db.QueryRow(`SELECT shift_label FROM tbm_shifts WHERE id = $1`, id).Scan(&label)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "shift not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "shift_label": label})
}

// =============================================================================
// Consumables
// =============================================================================
func (h *TBMHandler) ListConsumables(w http.ResponseWriter, r *http.Request) {
	tbmID := r.URL.Query().Get("tbm_id")
	ctype := r.URL.Query().Get("consumable_type")
	query := `SELECT id, tbm_id, consumable_type, item_name, item_code, unit, quantity_used, quantity_remaining, unit_price, used_at, recorded_by FROM tbm_consumables WHERE 1=1`
	args := []interface{}{}
	argIdx := 1
	if tbmID != "" { query += fmt.Sprintf(" AND tbm_id = $%d", argIdx); argIdx++; args = append(args, tbmID) }
	if ctype != "" { query += fmt.Sprintf(" AND consumable_type = $%d", argIdx); argIdx++; args = append(args, ctype) }
	query += " ORDER BY used_at DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, tbmID, ctype, name, code, unit, recBy string
		var qtyUsed, qtyRem, price float64
		var usedAt time.Time
		if err := rows.Scan(&id, &tbmID, &ctype, &name, &code, &unit, &qtyUsed, &qtyRem, &price, &usedAt, &recBy); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		items = append(items, map[string]interface{}{
			"id": id, "tbm_id": tbmID, "consumable_type": ctype, "item_name": name,
			"item_code": code, "unit": unit, "quantity_used": qtyUsed,
			"quantity_remaining": qtyRem, "unit_price": price, "used_at": usedAt,
			"recorded_by": recBy,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *TBMHandler) CreateConsumable(w http.ResponseWriter, r *http.Request) {
	var input struct {
		TbmID          string  `json:"tbm_id"`
		ConsumableType string  `json:"consumable_type"`
		ItemName       string  `json:"item_name"`
		QuantityUsed   float64 `json:"quantity_used"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	_, err := h.db.Exec(`INSERT INTO tbm_consumables (id, tbm_id, consumable_type, item_name, quantity_used) VALUES ($1,$2,$3,$4,$5)`,
		id, input.TbmID, input.ConsumableType, input.ItemName, input.QuantityUsed)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *TBMHandler) GetConsumable(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var name string
	err := h.db.QueryRow(`SELECT item_name FROM tbm_consumables WHERE id = $1`, id).Scan(&name)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "consumable not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "item_name": name})
}

// =============================================================================
// Performance Metrics
// =============================================================================
func (h *TBMHandler) ListPerformance(w http.ResponseWriter, r *http.Request) {
	tbmID := r.URL.Query().Get("tbm_id")
	query := `SELECT id, tbm_id, metric_date, shift_label, rings_built, advance_mm, avg_advance_rate_mmmin, max_advance_rate_mmmin, avg_thrust_force_kN, avg_torque_kNm, avg_face_pressure_bar, total_downtime_minutes, utilisation_pct, tbm_availability_pct, performance_factor, grout_volume_m3, foam_consumption_kg, bentonite_consumption_kg FROM tbm_performance_metrics WHERE 1=1`
	args := []interface{}{}
	argIdx := 1
	if tbmID != "" { query += fmt.Sprintf(" AND tbm_id = $%d", argIdx); argIdx++; args = append(args, tbmID) }
	query += " ORDER BY metric_date DESC, shift_label"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, tbmID, label string
		var metricDate time.Time
		var rings, advance, downtime int
		var avgAdv, maxAdv, avgThrust, avgTorq, avgFP, util, avail, perfFactor, grout, foam, bent float64
		if err := rows.Scan(&id, &tbmID, &metricDate, &label, &rings, &advance,
			&avgAdv, &maxAdv, &avgThrust, &avgTorq, &avgFP,
			&downtime, &util, &avail, &perfFactor, &grout, &foam, &bent); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		items = append(items, map[string]interface{}{
			"id": id, "tbm_id": tbmID, "metric_date": metricDate, "shift_label": label,
			"rings_built": rings, "advance_mm": advance, "avg_advance_rate_mmmin": avgAdv,
			"max_advance_rate_mmmin": maxAdv, "avg_thrust_force_kN": avgThrust,
			"avg_torque_kNm": avgTorq, "avg_face_pressure_bar": avgFP,
			"total_downtime_minutes": downtime, "utilisation_pct": util,
			"tbm_availability_pct": avail, "performance_factor": perfFactor,
			"grout_volume_m3": grout, "foam_consumption_kg": foam, "bentonite_consumption_kg": bent,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *TBMHandler) CreatePerformance(w http.ResponseWriter, r *http.Request) {
	var input struct {
		TbmID string `json:"tbm_id"`
		Date  string `json:"metric_date"`
		Label string `json:"shift_label"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	_, err := h.db.Exec(`INSERT INTO tbm_performance_metrics (id, tbm_id, metric_date, shift_label) VALUES ($1,$2,$3,$4)`,
		id, input.TbmID, input.Date, input.Label)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *TBMHandler) GetPerformance(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var tbmID string
	err := h.db.QueryRow(`SELECT tbm_id FROM tbm_performance_metrics WHERE id = $1`, id).Scan(&tbmID)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "performance metric not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"id": id, "tbm_id": tbmID})
}

// =============================================================================
// Summary
// =============================================================================
func (h *TBMHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	tbmID := r.URL.Query().Get("tbm_id")
	summary := map[string]interface{}{
		"total_alarms": 0, "active_alarms": 0, "total_shifts": 0,
		"total_rings": 0, "total_advance_mm": 0, "avg_utilisation_pct": 0,
	}

	rows, err := h.db.Query(`SELECT COUNT(*) FROM tbm_telemetry WHERE 1=1` + mapCond(tbmID, "tbm_id"))
	if err == nil && rows.Next() { var v int; rows.Scan(&v); summary["total_telemetry"] = v; rows.Close() }

	rows2, err := h.db.Query(`SELECT COUNT(*) FROM tbm_alarms WHERE is_active=TRUE` + mapCond(tbmID, "tbm_id"))
	if err == nil && rows2.Next() { var v int; rows2.Scan(&v); summary["active_alarms"] = v; rows2.Close() }

	rows3, err := h.db.Query(`SELECT COALESCE(SUM(rings_built),0), COALESCE(SUM(advance_mm),0), COALESCE(AVG(utilisation_pct),0) FROM tbm_performance_metrics WHERE 1=1` + mapCond(tbmID, "tbm_id"))
	if err == nil && rows3.Next() { var r, a int; var u float64; rows3.Scan(&r, &a, &u); summary["total_rings"] = r; summary["total_advance_mm"] = a; summary["avg_utilisation_pct"] = u; rows3.Close() }

	respondJSON(w, http.StatusOK, summary)
}

func mapCond(val, col string) string {
	if val != "" { return fmt.Sprintf(" AND %s = '%s'", col, val) }
	return ""
}