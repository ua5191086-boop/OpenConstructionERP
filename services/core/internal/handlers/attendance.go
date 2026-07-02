package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// AttendanceHandler handles Time & Attendance module endpoints
type AttendanceHandler struct {
	db *sql.DB
}

func NewAttendanceHandler(db *sql.DB) *AttendanceHandler {
	return &AttendanceHandler{db: db}
}

func (h *AttendanceHandler) RegisterRoutes(r chi.Router) {
	r.Route("/attendance", func(r chi.Router) {
		// Calendars
		r.Get("/calendars", h.ListCalendars)
		r.Post("/calendars", h.CreateCalendar)
		r.Get("/calendars/{id}", h.GetCalendar)
		r.Put("/calendars/{id}", h.UpdateCalendar)
		r.Delete("/calendars/{id}", h.DeleteCalendar)

		// Attendance Employees
		r.Get("/employees", h.ListAttEmployees)
		r.Post("/employees", h.CreateAttEmployee)
		r.Get("/employees/{id}", h.GetAttEmployee)
		r.Put("/employees/{id}", h.UpdateAttEmployee)
		r.Delete("/employees/{id}", h.DeleteAttEmployee)

		// Timesheets
		r.Get("/timesheets", h.ListTimesheets)
		r.Post("/timesheets", h.CreateTimesheet)
		r.Get("/timesheets/{id}", h.GetTimesheet)
		r.Put("/timesheets/{id}", h.UpdateTimesheet)
		r.Delete("/timesheets/{id}", h.DeleteTimesheet)
		r.Post("/timesheets/clock", h.ClockInOut)

		// Biometric Events
		r.Get("/biometric-events", h.ListBiometricEvents)
		r.Post("/biometric-events", h.CreateBiometricEvent)

		// Absences
		r.Get("/absences", h.ListAbsences)
		r.Post("/absences", h.CreateAbsence)
		r.Get("/absences/{id}", h.GetAbsence)
		r.Put("/absences/{id}", h.UpdateAbsence)
		r.Delete("/absences/{id}", h.DeleteAbsence)

		// Gate Log
		r.Get("/gate-log", h.ListGateLog)
		r.Post("/gate-log", h.CreateGateLog)

		// Summary
		r.Get("/summary", h.GetSummary)
	})
}

// =============================================================================
// Calendars
// =============================================================================
func (h *AttendanceHandler) ListCalendars(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT id, project_id, calendar_code, calendar_name, calendar_type, work_days, work_hours_per_day, start_time, end_time, break_minutes, is_active, notes, created_at FROM attendance_calendars WHERE 1=1`
	args := []interface{}{}
	argIdx := 1
	if projectID != "" { query += fmt.Sprintf(" AND project_id = $%d", argIdx); argIdx++; args = append(args, projectID) }
	query += " ORDER BY calendar_code"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, pid, code, name, ctype, notes string
		var hoursPerDay float64
		var startTime, endTime string
		var breakMin int
		var active bool
		var createdAt time.Time
		var workDays []byte
		if err := rows.Scan(&id, &pid, &code, &name, &ctype, &workDays, &hoursPerDay, &startTime, &endTime, &breakMin, &active, &notes, &createdAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		items = append(items, map[string]interface{}{
			"id": id, "project_id": pid, "calendar_code": code, "calendar_name": name,
			"calendar_type": ctype, "work_hours_per_day": hoursPerDay,
			"start_time": startTime, "end_time": endTime, "break_minutes": breakMin,
			"is_active": active, "notes": notes, "created_at": createdAt,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *AttendanceHandler) CreateCalendar(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID    *string  `json:"project_id"`
		CalendarCode string   `json:"calendar_code"`
		CalendarName string   `json:"calendar_name"`
		CalendarType *string  `json:"calendar_type"`
		HoursPerDay  *float64 `json:"work_hours_per_day"`
		StartTime    *string  `json:"start_time"`
		EndTime      *string  `json:"end_time"`
		BreakMin     *int     `json:"break_minutes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO attendance_calendars (id, project_id, calendar_code, calendar_name, calendar_type, work_hours_per_day, start_time, end_time, break_minutes, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$10)`,
		id, input.ProjectID, input.CalendarCode, input.CalendarName, input.CalendarType, input.HoursPerDay, input.StartTime, input.EndTime, input.BreakMin, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *AttendanceHandler) GetCalendar(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var code, name string
	err := h.db.QueryRow(`SELECT calendar_code, calendar_name FROM attendance_calendars WHERE id = $1`, id).Scan(&code, &name)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "calendar not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "calendar_code": code, "calendar_name": name})
}

func (h *AttendanceHandler) UpdateCalendar(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		CalendarName *string  `json:"calendar_name"`
		IsActive     *bool    `json:"is_active"`
		HoursPerDay  *float64 `json:"work_hours_per_day"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	_, err := h.db.Exec(`UPDATE attendance_calendars SET calendar_name=COALESCE($1,calendar_name), is_active=COALESCE($2,is_active), work_hours_per_day=COALESCE($3,work_hours_per_day), updated_at=$4 WHERE id=$5`,
		input.CalendarName, input.IsActive, input.HoursPerDay, time.Now(), id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *AttendanceHandler) DeleteCalendar(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM attendance_calendars WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// Attendance Employees
// =============================================================================
func (h *AttendanceHandler) ListAttEmployees(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT ae.id, ae.employee_id, e.first_name, e.last_name, ae.project_id, ae.calendar_id, ae.badge_number, ae.biometric_id, ae.rfid_card, ae.employment_type, ae.hourly_rate, ae.is_active, ae.created_at
		FROM attendance_employees ae JOIN employees e ON e.id = ae.employee_id WHERE 1=1`
	args := []interface{}{}
	argIdx := 1
	if projectID != "" { query += fmt.Sprintf(" AND ae.project_id = $%d", argIdx); argIdx++; args = append(args, projectID) }
	query += " ORDER BY e.last_name, e.first_name"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, eid, fn, ln, pid, calID, badge, bio, rfid, etype string
		var rate float64
		var active bool
		var createdAt time.Time
		if err := rows.Scan(&id, &eid, &fn, &ln, &pid, &calID, &badge, &bio, &rfid, &etype, &rate, &active, &createdAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		items = append(items, map[string]interface{}{
			"id": id, "employee_id": eid, "first_name": fn, "last_name": ln,
			"project_id": pid, "calendar_id": calID, "badge_number": badge,
			"biometric_id": bio, "rfid_card": rfid, "employment_type": etype,
			"hourly_rate": rate, "is_active": active, "created_at": createdAt,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *AttendanceHandler) CreateAttEmployee(w http.ResponseWriter, r *http.Request) {
	var input struct {
		EmployeeID    string  `json:"employee_id"`
		ProjectID     *string `json:"project_id"`
		CalendarID    *string `json:"calendar_id"`
		BadgeNumber   *string `json:"badge_number"`
		BiometricID   *string `json:"biometric_id"`
		RFIDCard      *string `json:"rfid_card"`
		EmploymentType *string `json:"employment_type"`
		HourlyRate    *float64 `json:"hourly_rate"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO attendance_employees (id, employee_id, project_id, calendar_id, badge_number, biometric_id, rfid_card, employment_type, hourly_rate, created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		id, input.EmployeeID, input.ProjectID, input.CalendarID, input.BadgeNumber, input.BiometricID, input.RFIDCard, input.EmploymentType, input.HourlyRate, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *AttendanceHandler) GetAttEmployee(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var badge, etype string
	err := h.db.QueryRow(`SELECT badge_number, employment_type FROM attendance_employees WHERE id = $1`, id).Scan(&badge, &etype)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "attendance employee not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "badge_number": badge, "employment_type": etype})
}

func (h *AttendanceHandler) UpdateAttEmployee(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		CalendarID  *string `json:"calendar_id"`
		IsActive    *bool   `json:"is_active"`
		BadgeNumber *string `json:"badge_number"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	_, err := h.db.Exec(`UPDATE attendance_employees SET calendar_id=COALESCE($1,calendar_id), is_active=COALESCE($2,is_active), badge_number=COALESCE($3,badge_number) WHERE id=$4`,
		input.CalendarID, input.IsActive, input.BadgeNumber, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *AttendanceHandler) DeleteAttEmployee(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM attendance_employees WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// Timesheets
// =============================================================================
func (h *AttendanceHandler) ListTimesheets(w http.ResponseWriter, r *http.Request) {
	employeeID := r.URL.Query().Get("employee_id")
	projectID := r.URL.Query().Get("project_id")
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")

	query := `SELECT at.id, at.project_id, at.employee_id, e.first_name, e.last_name, at.timesheet_date, at.clock_in, at.clock_out, at.hours_worked, at.hours_regular, at.hours_overtime, at.hours_absence, at.absence_type, at.status, at.source, at.notes, at.created_at
		FROM attendance_timesheets at JOIN employees e ON e.id = at.employee_id WHERE 1=1`
	args := []interface{}{}
	argIdx := 1
	if employeeID != "" { query += fmt.Sprintf(" AND at.employee_id = $%d", argIdx); argIdx++; args = append(args, employeeID) }
	if projectID != "" { query += fmt.Sprintf(" AND at.project_id = $%d", argIdx); argIdx++; args = append(args, projectID) }
	if from != "" { query += fmt.Sprintf(" AND at.timesheet_date >= $%d", argIdx); argIdx++; args = append(args, from) }
	if to != "" { query += fmt.Sprintf(" AND at.timesheet_date <= $%d", argIdx); argIdx++; args = append(args, to) }
	query += " ORDER BY at.timesheet_date DESC LIMIT 200"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, pid, eid, fn, ln, status, src, notes string
		var tsDate time.Time
		var clockIn, clockOut sql.NullTime
		var hrsWorked, hrsReg, hrsOT, hrsAbs sql.NullFloat64
		var absType sql.NullString
		var createdAt time.Time
		if err := rows.Scan(&id, &pid, &eid, &fn, &ln, &tsDate, &clockIn, &clockOut, &hrsWorked, &hrsReg, &hrsOT, &hrsAbs, &absType, &status, &src, &notes, &createdAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		item := map[string]interface{}{
			"id": id, "project_id": pid, "employee_id": eid, "employee_name": fn + " " + ln,
			"timesheet_date": tsDate, "status": status, "source": src, "notes": notes, "created_at": createdAt,
		}
		if clockIn.Valid { item["clock_in"] = clockIn.Time }
		if clockOut.Valid { item["clock_out"] = clockOut.Time }
		if hrsWorked.Valid { item["hours_worked"] = hrsWorked.Float64 }
		if hrsReg.Valid { item["hours_regular"] = hrsReg.Float64 }
		if hrsOT.Valid { item["hours_overtime"] = hrsOT.Float64 }
		if hrsAbs.Valid { item["hours_absence"] = hrsAbs.Float64 }
		if absType.Valid { item["absence_type"] = absType.String }
		items = append(items, item)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *AttendanceHandler) CreateTimesheet(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID   string  `json:"project_id"`
		EmployeeID  string  `json:"employee_id"`
		TimesheetDate string `json:"timesheet_date"`
		ClockIn     *string `json:"clock_in"`
		ClockOut    *string `json:"clock_out"`
		HoursWorked *float64 `json:"hours_worked"`
		HoursReg    *float64 `json:"hours_regular"`
		HoursOT     *float64 `json:"hours_overtime"`
		HoursAbs    *float64 `json:"hours_absence"`
		AbsenceType *string `json:"absence_type"`
		Source      *string `json:"source"`
		Notes       *string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	now := time.Now()
	// Attempt to parse day of week
	dow := 0
	if parsed, err := time.Parse("2006-01-02", input.TimesheetDate); err == nil {
		dow = int(parsed.Weekday())
		if dow == 0 { dow = 7 }
	}
	_, err := h.db.Exec(`INSERT INTO attendance_timesheets (id, project_id, employee_id, timesheet_date, day_of_week, clock_in, clock_out, hours_worked, hours_regular, hours_overtime, hours_absence, absence_type, source, notes, status, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,'pending',$15,$15)`,
		id, input.ProjectID, input.EmployeeID, input.TimesheetDate, dow, input.ClockIn, input.ClockOut, input.HoursWorked, input.HoursReg, input.HoursOT, input.HoursAbs, input.AbsenceType, input.Source, input.Notes, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *AttendanceHandler) GetTimesheet(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var eid string
	var tsDate time.Time
	var hrs float64
	err := h.db.QueryRow(`SELECT employee_id, timesheet_date, hours_worked FROM attendance_timesheets WHERE id = $1`, id).Scan(&eid, &tsDate, &hrs)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "timesheet not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "employee_id": eid, "timesheet_date": tsDate, "hours_worked": hrs})
}

func (h *AttendanceHandler) UpdateTimesheet(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Status      *string  `json:"status"`
		HoursWorked *float64 `json:"hours_worked"`
		HoursOT     *float64 `json:"hours_overtime"`
		ApprovedBy  *string  `json:"approved_by"`
		Notes       *string  `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	_, err := h.db.Exec(`UPDATE attendance_timesheets SET status=COALESCE($1,status), hours_worked=COALESCE($2,hours_worked), hours_overtime=COALESCE($3,hours_overtime), notes=COALESCE($4,notes), approved_by=COALESCE($5,approved_by), approved_at=CASE WHEN $1='approved' THEN NOW() ELSE approved_at END, updated_at=NOW() WHERE id=$6`,
		input.Status, input.HoursWorked, input.HoursOT, input.Notes, input.ApprovedBy, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *AttendanceHandler) DeleteTimesheet(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM attendance_timesheets WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// Clock In/Out — creates or updates timesheet for today
func (h *AttendanceHandler) ClockInOut(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID  string `json:"project_id"`
		EmployeeID string `json:"employee_id"`
		Action     string `json:"action"` // clock_in, clock_out, break_start, break_end
		Source     string `json:"source"`
		DeviceID   *string `json:"device_id"`
		Lat        *float64 `json:"lat"`
		Lng        *float64 `json:"lng"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}

	now := time.Now()
	today := now.Format("2006-01-02")
	dow := int(now.Weekday())
	if dow == 0 { dow = 7 }

	// Check if timesheet exists for today
	var existingID string
	err := h.db.QueryRow(`SELECT id FROM attendance_timesheets WHERE employee_id = $1 AND timesheet_date = $2`, input.EmployeeID, today).Scan(&existingID)

	if err == sql.ErrNoRows {
		// Create new
		id := uuid.New().String()
		_, err = h.db.Exec(`INSERT INTO attendance_timesheets (id, project_id, employee_id, timesheet_date, day_of_week, clock_in, clock_out, source, device_id, location, status, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,point($10,$11),'pending',$12,$12)`,
			id, input.ProjectID, input.EmployeeID, today, dow, now, nil, input.Source, input.DeviceID, input.Lat, input.Lng, now)
		if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
		respondJSON(w, http.StatusCreated, map[string]string{"id": id, "action": "clocked_in"})
	} else if err == nil {
		// Update existing
		if input.Action == "clock_out" {
			// Calculate hours worked
			var clockIn sql.NullTime
			h.db.QueryRow(`SELECT clock_in FROM attendance_timesheets WHERE id = $1`, existingID).Scan(&clockIn)
			hoursWorked := 0.0
			hoursReg := 0.0
			hoursOT := 0.0
			if clockIn.Valid {
				hoursWorked = math.Round(now.Sub(clockIn.Time).Hours()*100) / 100
				if hoursWorked > 8 {
					hoursReg = 8
					hoursOT = hoursWorked - 8
				} else {
					hoursReg = hoursWorked
				}
			}
			_, err = h.db.Exec(`UPDATE attendance_timesheets SET clock_out=$1, hours_worked=$2, hours_regular=$3, hours_overtime=$4, updated_at=$5 WHERE id=$6`,
				now, hoursWorked, hoursReg, hoursOT, now, existingID)
		} else if input.Action == "break_start" {
			_, err = h.db.Exec(`UPDATE attendance_timesheets SET break_start=$1, updated_at=$2 WHERE id=$3`, now, now, existingID)
		} else if input.Action == "break_end" {
			_, err = h.db.Exec(`UPDATE attendance_timesheets SET break_end=$1, updated_at=$2 WHERE id=$3`, now, now, existingID)
		}
		if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
		respondJSON(w, http.StatusOK, map[string]string{"id": existingID, "action": input.Action})
	} else {
		respondError(w, http.StatusInternalServerError, err.Error())
	}
}

// =============================================================================
// Biometric Events
// =============================================================================
func (h *AttendanceHandler) ListBiometricEvents(w http.ResponseWriter, r *http.Request) {
	employeeID := r.URL.Query().Get("employee_id")
	eventType := r.URL.Query().Get("event_type")
	query := `SELECT be.id, be.employee_id, e.first_name, e.last_name, be.event_time, be.event_type, be.device_id, be.device_type, be.verified, be.notes, be.created_at
		FROM attendance_biometric_events be JOIN employees e ON e.id = be.employee_id WHERE 1=1`
	args := []interface{}{}
	argIdx := 1
	if employeeID != "" { query += fmt.Sprintf(" AND be.employee_id = $%d", argIdx); argIdx++; args = append(args, employeeID) }
	if eventType != "" { query += fmt.Sprintf(" AND be.event_type = $%d", argIdx); argIdx++; args = append(args, eventType) }
	query += " ORDER BY be.event_time DESC LIMIT 200"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, eid, fn, ln, etype, devID, devType, notes string
		var eventTime, createdAt time.Time
		var verified bool
		if err := rows.Scan(&id, &eid, &fn, &ln, &eventTime, &etype, &devID, &devType, &verified, &notes, &createdAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		items = append(items, map[string]interface{}{
			"id": id, "employee_id": eid, "employee_name": fn + " " + ln,
			"event_time": eventTime, "event_type": etype, "device_id": devID,
			"device_type": devType, "verified": verified, "notes": notes,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *AttendanceHandler) CreateBiometricEvent(w http.ResponseWriter, r *http.Request) {
	var input struct {
		EmployeeID string  `json:"employee_id"`
		ProjectID  *string `json:"project_id"`
		EventType  string  `json:"event_type"`
		DeviceID   *string `json:"device_id"`
		DeviceType *string `json:"device_type"`
		Verified   *bool   `json:"verified"`
		Temperature *float64 `json:"temperature"`
		Notes      *string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO attendance_biometric_events (id, employee_id, project_id, event_time, event_type, device_id, device_type, verified, temperature, notes, created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		id, input.EmployeeID, input.ProjectID, now, input.EventType, input.DeviceID, input.DeviceType, input.Verified, input.Temperature, input.Notes, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id, "status": "recorded"})
}

// =============================================================================
// Absences
// =============================================================================
func (h *AttendanceHandler) ListAbsences(w http.ResponseWriter, r *http.Request) {
	employeeID := r.URL.Query().Get("employee_id")
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT aa.id, aa.employee_id, e.first_name, e.last_name, aa.absence_type, aa.start_date, aa.end_date, aa.total_days, aa.reason, aa.status, aa.created_at
		FROM attendance_absences aa JOIN employees e ON e.id = aa.employee_id WHERE 1=1`
	args := []interface{}{}
	argIdx := 1
	if employeeID != "" { query += fmt.Sprintf(" AND aa.employee_id = $%d", argIdx); argIdx++; args = append(args, employeeID) }
	if projectID != "" { query += fmt.Sprintf(" AND aa.project_id = $%d", argIdx); argIdx++; args = append(args, projectID) }
	query += " ORDER BY aa.start_date DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, eid, fn, ln, atype, reason, status string
		var sd, ed, createdAt time.Time
		var days int
		if err := rows.Scan(&id, &eid, &fn, &ln, &atype, &sd, &ed, &days, &reason, &status, &createdAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		items = append(items, map[string]interface{}{
			"id": id, "employee_id": eid, "employee_name": fn + " " + ln,
			"absence_type": atype, "start_date": sd, "end_date": ed,
			"total_days": days, "reason": reason, "status": status, "created_at": createdAt,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *AttendanceHandler) CreateAbsence(w http.ResponseWriter, r *http.Request) {
	var input struct {
		EmployeeID  string `json:"employee_id"`
		ProjectID   *string `json:"project_id"`
		AbsenceType string `json:"absence_type"`
		StartDate   string `json:"start_date"`
		EndDate     string `json:"end_date"`
		Reason      *string `json:"reason"`
		DocumentPath *string `json:"document_path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	now := time.Now()
	// Calculate days
	var sd, ed time.Time
	sd, _ = time.Parse("2006-01-02", input.StartDate)
	ed, _ = time.Parse("2006-01-02", input.EndDate)
	days := int(ed.Sub(sd).Hours()/24) + 1
	if days < 1 { days = 1 }

	_, err := h.db.Exec(`INSERT INTO attendance_absences (id, employee_id, project_id, absence_type, start_date, end_date, total_days, reason, document_path, status, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,'pending',$10,$10)`,
		id, input.EmployeeID, input.ProjectID, input.AbsenceType, input.StartDate, input.EndDate, days, input.Reason, input.DocumentPath, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *AttendanceHandler) GetAbsence(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var atype, status string
	err := h.db.QueryRow(`SELECT absence_type, status FROM attendance_absences WHERE id = $1`, id).Scan(&atype, &status)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "absence not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "absence_type": atype, "status": status})
}

func (h *AttendanceHandler) UpdateAbsence(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Status   *string `json:"status"`
		ApprovedBy *string `json:"approved_by"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	_, err := h.db.Exec(`UPDATE attendance_absences SET status=COALESCE($1,status), approved_by=COALESCE($2,approved_by), approved_at=CASE WHEN $1='approved' THEN NOW() ELSE approved_at END, updated_at=NOW() WHERE id=$3`,
		input.Status, input.ApprovedBy, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *AttendanceHandler) DeleteAbsence(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM attendance_absences WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// Gate Log
// =============================================================================
func (h *AttendanceHandler) ListGateLog(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT gl.id, gl.project_id, gl.employee_id, e.first_name, e.last_name, gl.visitor_name, gl.visitor_company, gl.gate, gl.direction, gl.event_time, gl.vehicle_plate, gl.notes, gl.created_at
		FROM attendance_gate_log gl LEFT JOIN employees e ON e.id = gl.employee_id WHERE 1=1`
	args := []interface{}{}
	argIdx := 1
	if projectID != "" { query += fmt.Sprintf(" AND gl.project_id = $%d", argIdx); argIdx++; args = append(args, projectID) }
	query += " ORDER BY gl.event_time DESC LIMIT 200"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, pid, eid, fn, ln, vname, vcomp, gate, dir, plate, notes string
		var eventTime, createdAt time.Time
		if err := rows.Scan(&id, &pid, &eid, &fn, &ln, &vname, &vcomp, &gate, &dir, &eventTime, &plate, &notes, &createdAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		item := map[string]interface{}{
			"id": id, "project_id": pid, "gate": gate, "direction": dir,
			"event_time": eventTime, "vehicle_plate": plate, "notes": notes, "created_at": createdAt,
		}
		if eid != "" {
			item["employee_id"] = eid
			item["employee_name"] = fn + " " + ln
		} else {
			item["visitor_name"] = vname
			item["visitor_company"] = vcomp
		}
		items = append(items, item)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *AttendanceHandler) CreateGateLog(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID     *string `json:"project_id"`
		EmployeeID    *string `json:"employee_id"`
		VisitorName   *string `json:"visitor_name"`
		VisitorCompany *string `json:"visitor_company"`
		Gate          string  `json:"gate"`
		Direction     string  `json:"direction"`
		VehiclePlate  *string `json:"vehicle_plate"`
		Notes         *string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO attendance_gate_log (id, project_id, employee_id, visitor_name, visitor_company, gate, direction, event_time, vehicle_plate, notes, created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		id, input.ProjectID, input.EmployeeID, input.VisitorName, input.VisitorCompany, input.Gate, input.Direction, now, input.VehiclePlate, input.Notes, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

// =============================================================================
// Summary
// =============================================================================
func (h *AttendanceHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT project_id, registered_employees, active_employees, active_last_30d, total_hours_30d, total_overtime_30d, total_absence_30d, avg_hours_per_day, active_absences, gate_events_7d FROM attendance_summary`
	args := []interface{}{}
	argIdx := 1
	if projectID != "" { query += fmt.Sprintf(" WHERE project_id = $%d", argIdx); args = append(args, projectID) }

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var pid string
		var reg, active, active30, absences, gateEvents int
		var hrs30, ot30, abs30, avgHrs float64
		if err := rows.Scan(&pid, &reg, &active, &active30, &hrs30, &ot30, &abs30, &avgHrs, &absences, &gateEvents); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		items = append(items, map[string]interface{}{
			"project_id": pid, "registered_employees": reg, "active_employees": active,
			"active_last_30d": active30, "total_hours_30d": hrs30, "total_overtime_30d": ot30,
			"total_absence_30d": abs30, "avg_hours_per_day": avgHrs,
			"active_absences": absences, "gate_events_7d": gateEvents,
		})
	}
	respondJSON(w, http.StatusOK, items)
}