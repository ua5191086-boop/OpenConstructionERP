package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/openconstructionerp/oce/services/core/internal/models"
)

// EquipmentHandler handles Equipment Management module endpoints
type EquipmentHandler struct {
	db *sql.DB
}

func NewEquipmentHandler(db *sql.DB) *EquipmentHandler {
	return &EquipmentHandler{db: db}
}

func (h *EquipmentHandler) RegisterRoutes(r chi.Router) {
	r.Route("/equipment", func(r chi.Router) {
		// Categories
		r.Get("/categories", h.ListCategories)
		r.Post("/categories", h.CreateCategory)
		r.Get("/categories/{id}", h.GetCategory)
		r.Put("/categories/{id}", h.UpdateCategory)
		r.Delete("/categories/{id}", h.DeleteCategory)

		// Equipment
		r.Get("/items", h.ListEquipment)
		r.Post("/items", h.CreateEquipment)
		r.Get("/items/{id}", h.GetEquipment)
		r.Put("/items/{id}", h.UpdateEquipment)
		r.Delete("/items/{id}", h.DeleteEquipment)

		// Maintenance
		r.Get("/maintenance", h.ListMaintenance)
		r.Post("/maintenance", h.CreateMaintenance)
		r.Get("/maintenance/{id}", h.GetMaintenance)
		r.Put("/maintenance/{id}", h.UpdateMaintenance)
		r.Delete("/maintenance/{id}", h.DeleteMaintenance)

		// Maintenance Schedules
		r.Get("/maintenance-schedules", h.ListMaintenanceSchedules)
		r.Post("/maintenance-schedules", h.CreateMaintenanceSchedule)
		r.Get("/maintenance-schedules/{id}", h.GetMaintenanceSchedule)
		r.Put("/maintenance-schedules/{id}", h.UpdateMaintenanceSchedule)
		r.Delete("/maintenance-schedules/{id}", h.DeleteMaintenanceSchedule)

		// Telemetry
		r.Get("/telemetry", h.ListTelemetry)
		r.Post("/telemetry", h.CreateTelemetry)
		r.Get("/telemetry/{id}", h.GetTelemetry)

		// Fuel
		r.Get("/fuel", h.ListFuel)
		r.Post("/fuel", h.CreateFuel)
		r.Get("/fuel/{id}", h.GetFuel)

		// Operators
		r.Get("/operators", h.ListOperators)
		r.Post("/operators", h.CreateOperator)
		r.Get("/operators/{id}", h.GetOperator)
		r.Delete("/operators/{id}", h.DeleteOperator)

		// Downtime
		r.Get("/downtime", h.ListDowntime)
		r.Post("/downtime", h.CreateDowntime)
		r.Get("/downtime/{id}", h.GetDowntime)
		r.Put("/downtime/{id}", h.UpdateDowntime)
		r.Delete("/downtime/{id}", h.DeleteDowntime)

		// Spare Parts
		r.Get("/spare-parts", h.ListSpareParts)
		r.Post("/spare-parts", h.CreateSparePart)
		r.Get("/spare-parts/{id}", h.GetSparePart)
		r.Put("/spare-parts/{id}", h.UpdateSparePart)
		r.Delete("/spare-parts/{id}", h.DeleteSparePart)

		// Summary
		r.Get("/summary", h.GetSummary)
	})
}

// =============================================================================
// Categories
// =============================================================================

func (h *EquipmentHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(`SELECT id, category_code, category_name, description, parent_id, equipment_type, icon, sort_order, created_at FROM equipment_categories ORDER BY sort_order`)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, code, name, etype, icon string
		var desc, parentID sql.NullString
		var sortOrder int
		var createdAt time.Time
		err := rows.Scan(&id, &code, &name, &desc, &parentID, &etype, &icon, &sortOrder, &createdAt)
		if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
		item := map[string]interface{}{
			"id": id, "category_code": code, "category_name": name, "equipment_type": etype,
			"icon": icon, "sort_order": sortOrder, "created_at": createdAt,
		}
		if desc.Valid { item["description"] = desc.String }
		if parentID.Valid { item["parent_id"] = parentID.String }
		items = append(items, item)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *EquipmentHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var input struct {
		CategoryCode string `json:"category_code"`
		CategoryName string `json:"category_name"`
		EquipmentType string `json:"equipment_type"`
		Description  *string `json:"description"`
		ParentID     *string `json:"parent_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	_, err := h.db.Exec(`INSERT INTO equipment_categories (id, category_code, category_name, description, parent_id, equipment_type) VALUES ($1,$2,$3,$4,$5,$6)`,
		id, input.CategoryCode, input.CategoryName, input.Description, input.ParentID, input.EquipmentType)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *EquipmentHandler) GetCategory(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var code, name string
	err := h.db.QueryRow(`SELECT category_code, category_name FROM equipment_categories WHERE id = $1`, id).Scan(&code, &name)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "category not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "category_code": code, "category_name": name})
}

func (h *EquipmentHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		CategoryName  *string `json:"category_name"`
		Description   *string `json:"description"`
		SortOrder     *int    `json:"sort_order"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	_, err := h.db.Exec(`UPDATE equipment_categories SET category_name=COALESCE($1,category_name), description=COALESCE($2,description), sort_order=COALESCE($3,sort_order) WHERE id=$4`,
		input.CategoryName, input.Description, input.SortOrder, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *EquipmentHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM equipment_categories WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// Equipment
// =============================================================================

func (h *EquipmentHandler) ListEquipment(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	status := r.URL.Query().Get("status")
	eqType := r.URL.Query().Get("equipment_type")

	query := `SELECT id, project_id, equipment_code, equipment_name, category_id, equipment_type, manufacturer, model, serial_number, year_manufactured, capacity, capacity_unit, status, location, purchase_date, purchase_cost, current_value, fuel_type, fuel_capacity, hourly_rate, meter_type, meter_reading, operator_required, next_service_date, is_active, created_at, updated_at FROM equipment WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if projectID != "" { query += fmt.Sprintf(" AND project_id = $%d", argIdx); argIdx++; args = append(args, projectID) }
	if status != "" { query += fmt.Sprintf(" AND status = $%d", argIdx); argIdx++; args = append(args, status) }
	if eqType != "" { query += fmt.Sprintf(" AND equipment_type = $%d", argIdx); argIdx++; args = append(args, eqType) }
	query += " ORDER BY equipment_code"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, pid, ecode, ename, catID, etype, manuf, model, serial, cap, capUnit, status, location, fuelType, meterType string
		var year, mRead int
		var purchaseCost, currentValue, fuelCap, hourlyRate float64
		var purchaseDate, nextServiceDate, createdAt, updatedAt time.Time
		var opRequired, isActive bool
		err := rows.Scan(&id, &pid, &ecode, &ename, &catID, &etype, &manuf, &model, &serial, &year, &cap, &capUnit, &status, &location, &purchaseDate, &purchaseCost, &currentValue, &fuelType, &fuelCap, &hourlyRate, &meterType, &mRead, &opRequired, &nextServiceDate, &isActive, &createdAt, &updatedAt)
		if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
		items = append(items, map[string]interface{}{
			"id": id, "project_id": pid, "equipment_code": ecode, "equipment_name": ename,
			"category_id": catID, "equipment_type": etype, "manufacturer": manuf, "model": model,
			"serial_number": serial, "year_manufactured": year, "capacity": cap, "capacity_unit": capUnit,
			"status": status, "location": location, "purchase_date": purchaseDate,
			"purchase_cost": purchaseCost, "current_value": currentValue, "fuel_type": fuelType,
			"fuel_capacity": fuelCap, "hourly_rate": hourlyRate, "meter_type": meterType,
			"meter_reading": mRead, "operator_required": opRequired, "next_service_date": nextServiceDate,
			"is_active": isActive, "created_at": createdAt, "updated_at": updatedAt,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *EquipmentHandler) CreateEquipment(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID     string  `json:"project_id"`
		EquipmentCode string  `json:"equipment_code"`
		EquipmentName string  `json:"equipment_name"`
		EquipmentType *string `json:"equipment_type"`
		CategoryID    *string `json:"category_id"`
		Manufacturer  *string `json:"manufacturer"`
		Model         *string `json:"model"`
		Status        *string `json:"status"`
		Capacity      *string `json:"capacity"`
		FuelType      *string `json:"fuel_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO equipment (id, project_id, equipment_code, equipment_name, equipment_type, category_id, manufacturer, model, status, capacity, fuel_type, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		id, input.ProjectID, input.EquipmentCode, input.EquipmentName, input.EquipmentType, input.CategoryID, input.Manufacturer, input.Model, input.Status, input.Capacity, input.FuelType, now, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *EquipmentHandler) GetEquipment(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var ecode, ename, etype string
	err := h.db.QueryRow(`SELECT equipment_code, equipment_name, equipment_type FROM equipment WHERE id = $1`, id).Scan(&ecode, &ename, &etype)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "equipment not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "equipment_code": ecode, "equipment_name": ename, "equipment_type": etype})
}

func (h *EquipmentHandler) UpdateEquipment(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Status        *string  `json:"status"`
		Location      *string  `json:"location"`
		HourlyRate    *float64 `json:"hourly_rate"`
		MeterReading  *float64 `json:"meter_reading"`
		CurrentValue  *float64 `json:"current_value"`
		NextServiceDate *string `json:"next_service_date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	now := time.Now()
	_, err := h.db.Exec(`UPDATE equipment SET status=COALESCE($1,status), location=COALESCE($2,location), hourly_rate=COALESCE($3,hourly_rate), meter_reading=COALESCE($4,meter_reading), current_value=COALESCE($5,current_value), next_service_date=COALESCE($6,next_service_date), updated_at=$7 WHERE id=$8`,
		input.Status, input.Location, input.HourlyRate, input.MeterReading, input.CurrentValue, input.NextServiceDate, now, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *EquipmentHandler) DeleteEquipment(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM equipment WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// Maintenance
// =============================================================================

func (h *EquipmentHandler) ListMaintenance(w http.ResponseWriter, r *http.Request) {
	equipmentID := r.URL.Query().Get("equipment_id")
	status := r.URL.Query().Get("status")

	query := `SELECT id, equipment_id, maintenance_code, maintenance_type, description, priority, status, meter_at_service, cost_estimated, cost_actual, downtime_hours, technician, findings, scheduled_date, started_at, completed_at, next_service_date, created_at, updated_at FROM equipment_maintenance WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if equipmentID != "" { query += fmt.Sprintf(" AND equipment_id = $%d", argIdx); argIdx++; args = append(args, equipmentID) }
	if status != "" { query += fmt.Sprintf(" AND status = $%d", argIdx); argIdx++; args = append(args, status) }
	query += " ORDER BY scheduled_date DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, eid, mcode, mtype, desc, priority, status, tech, findings string
		var costEst, costAct, downtime float64
		var meterAtService, schedDate, nextServiceDate, createdAt, updatedAt sql.NullString
		var startedAt, completedAt sql.NullString
		err := rows.Scan(&id, &eid, &mcode, &mtype, &desc, &priority, &status, &meterAtService, &costEst, &costAct, &downtime, &tech, &findings, &schedDate, &startedAt, &completedAt, &nextServiceDate, &createdAt, &updatedAt)
		if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
		item := map[string]interface{}{
			"id": id, "equipment_id": eid, "maintenance_code": mcode, "maintenance_type": mtype,
			"description": desc, "priority": priority, "status": status,
			"cost_estimated": costEst, "cost_actual": costAct, "downtime_hours": downtime,
			"technician": tech, "findings": findings,
		}
		if meterAtService.Valid { item["meter_at_service"] = meterAtService.String }
		if schedDate.Valid { item["scheduled_date"] = schedDate.String }
		if startedAt.Valid { item["started_at"] = startedAt.String }
		if completedAt.Valid { item["completed_at"] = completedAt.String }
		if nextServiceDate.Valid { item["next_service_date"] = nextServiceDate.String }
		items = append(items, item)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *EquipmentHandler) CreateMaintenance(w http.ResponseWriter, r *http.Request) {
	var input struct {
		EquipmentID    string  `json:"equipment_id"`
		MaintenanceCode string `json:"maintenance_code"`
		MaintenanceType string `json:"maintenance_type"`
		Description    string  `json:"description"`
		Priority       *string `json:"priority"`
		CostEstimated  *float64 `json:"cost_estimated"`
		ScheduledDate  *string `json:"scheduled_date"`
		Technician     *string `json:"technician"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO equipment_maintenance (id, equipment_id, maintenance_code, maintenance_type, description, priority, cost_estimated, scheduled_date, technician, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		id, input.EquipmentID, input.MaintenanceCode, input.MaintenanceType, input.Description, input.Priority, input.CostEstimated, input.ScheduledDate, input.Technician, now, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *EquipmentHandler) GetMaintenance(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var mcode, mtype string
	err := h.db.QueryRow(`SELECT maintenance_code, maintenance_type FROM equipment_maintenance WHERE id = $1`, id).Scan(&mcode, &mtype)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "maintenance record not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "maintenance_code": mcode, "maintenance_type": mtype})
}

func (h *EquipmentHandler) UpdateMaintenance(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Status       *string  `json:"status"`
		CostActual   *float64 `json:"cost_actual"`
		DowntimeHours *float64 `json:"downtime_hours"`
		Findings     *string  `json:"findings"`
		CompletedAt  *string  `json:"completed_at"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	now := time.Now()
	_, err := h.db.Exec(`UPDATE equipment_maintenance SET status=COALESCE($1,status), cost_actual=COALESCE($2,cost_actual), downtime_hours=COALESCE($3,downtime_hours), findings=COALESCE($4,findings), completed_at=COALESCE($5,completed_at), updated_at=$6 WHERE id=$7`,
		input.Status, input.CostActual, input.DowntimeHours, input.Findings, input.CompletedAt, now, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *EquipmentHandler) DeleteMaintenance(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM equipment_maintenance WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// Maintenance Schedules
// =============================================================================

func (h *EquipmentHandler) ListMaintenanceSchedules(w http.ResponseWriter, r *http.Request) {
	equipmentID := r.URL.Query().Get("equipment_id")

	query := `SELECT id, equipment_id, schedule_name, interval_type, interval_days, interval_meter, task_list, estimated_hours, required_skills, is_active, created_at, updated_at FROM maintenance_schedules WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if equipmentID != "" { query += fmt.Sprintf(" AND equipment_id = $%d", argIdx); argIdx++; args = append(args, equipmentID) }
	query += " ORDER BY schedule_name"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, eid, sname, intervalType, taskList, reqSkills string
		var intervalDays int
		var intervalMeter, estHours float64
		var isActive bool
		var createdAt, updatedAt time.Time
		err := rows.Scan(&id, &eid, &sname, &intervalType, &intervalDays, &intervalMeter, &taskList, &estHours, &reqSkills, &isActive, &createdAt, &updatedAt)
		if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
		items = append(items, map[string]interface{}{
			"id": id, "equipment_id": eid, "schedule_name": sname, "interval_type": intervalType,
			"interval_days": intervalDays, "interval_meter": intervalMeter, "task_list": taskList,
			"estimated_hours": estHours, "required_skills": reqSkills, "is_active": isActive,
			"created_at": createdAt, "updated_at": updatedAt,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *EquipmentHandler) CreateMaintenanceSchedule(w http.ResponseWriter, r *http.Request) {
	var input struct {
		EquipmentID    string  `json:"equipment_id"`
		ScheduleName   string  `json:"schedule_name"`
		IntervalType   string  `json:"interval_type"`
		IntervalDays   *int    `json:"interval_days"`
		IntervalMeter  *float64 `json:"interval_meter"`
		EstimatedHours *float64 `json:"estimated_hours"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO maintenance_schedules (id, equipment_id, schedule_name, interval_type, interval_days, interval_meter, estimated_hours, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		id, input.EquipmentID, input.ScheduleName, input.IntervalType, input.IntervalDays, input.IntervalMeter, input.EstimatedHours, now, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *EquipmentHandler) GetMaintenanceSchedule(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var sname string
	err := h.db.QueryRow(`SELECT schedule_name FROM maintenance_schedules WHERE id = $1`, id).Scan(&sname)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "maintenance schedule not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "schedule_name": sname})
}

func (h *EquipmentHandler) UpdateMaintenanceSchedule(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		IsActive       *bool   `json:"is_active"`
		IntervalDays   *int    `json:"interval_days"`
		IntervalMeter  *float64 `json:"interval_meter"`
		EstimatedHours *float64 `json:"estimated_hours"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	now := time.Now()
	_, err := h.db.Exec(`UPDATE maintenance_schedules SET is_active=COALESCE($1,is_active), interval_days=COALESCE($2,interval_days), interval_meter=COALESCE($3,interval_meter), estimated_hours=COALESCE($4,estimated_hours), updated_at=$5 WHERE id=$6`,
		input.IsActive, input.IntervalDays, input.IntervalMeter, input.EstimatedHours, now, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *EquipmentHandler) DeleteMaintenanceSchedule(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM maintenance_schedules WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// Telemetry
// =============================================================================

func (h *EquipmentHandler) ListTelemetry(w http.ResponseWriter, r *http.Request) {
	equipmentID := r.URL.Query().Get("equipment_id")
	limit := r.URL.Query().Get("limit")

	query := `SELECT id, equipment_id, recorded_at, meter_value, fuel_level_pct, engine_temp_c, oil_pressure_bar, rpm, speed_kph, gps_lat, gps_lon, battery_voltage, error_codes, is_operating, data_source FROM equipment_telemetry WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if equipmentID != "" { query += fmt.Sprintf(" AND equipment_id = $%d", argIdx); argIdx++; args = append(args, equipmentID) }
	query += " ORDER BY recorded_at DESC"
	if limit != "" { query += fmt.Sprintf(" LIMIT %s", limit) }

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, eid, ds string
		var recordedAt time.Time
		var meterVal, fuelPct, engineTemp, oilPress, speed, batVolt sql.NullFloat64
		var rpm, gpsLat, gpsLon sql.NullFloat64
		var errCodes sql.NullString
		var isOperating bool
		err := rows.Scan(&id, &eid, &recordedAt, &meterVal, &fuelPct, &engineTemp, &oilPress, &rpm, &speed, &gpsLat, &gpsLon, &batVolt, &errCodes, &isOperating, &ds)
		if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
		item := map[string]interface{}{
			"id": id, "equipment_id": eid, "recorded_at": recordedAt, "is_operating": isOperating, "data_source": ds,
		}
		if meterVal.Valid { item["meter_value"] = meterVal.Float64 }
		if fuelPct.Valid { item["fuel_level_pct"] = fuelPct.Float64 }
		if engineTemp.Valid { item["engine_temp_c"] = engineTemp.Float64 }
		if oilPress.Valid { item["oil_pressure_bar"] = oilPress.Float64 }
		if rpm.Valid { item["rpm"] = rpm.Float64 }
		if speed.Valid { item["speed_kph"] = speed.Float64 }
		if batVolt.Valid { item["battery_voltage"] = batVolt.Float64 }
		if errCodes.Valid { item["error_codes"] = errCodes.String }
		items = append(items, item)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *EquipmentHandler) CreateTelemetry(w http.ResponseWriter, r *http.Request) {
	var input struct {
		EquipmentID string  `json:"equipment_id"`
		MeterValue  *float64 `json:"meter_value"`
		FuelLevelPct *float64 `json:"fuel_level_pct"`
		EngineTemp  *float64 `json:"engine_temp_c"`
		Rpm         *int    `json:"rpm"`
		IsOperating *bool   `json:"is_operating"`
		DataSource  *string `json:"data_source"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO equipment_telemetry (id, equipment_id, recorded_at, meter_value, fuel_level_pct, engine_temp_c, rpm, is_operating, data_source) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		id, input.EquipmentID, now, input.MeterValue, input.FuelLevelPct, input.EngineTemp, input.Rpm, input.IsOperating, input.DataSource)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *EquipmentHandler) GetTelemetry(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var recordedAt time.Time
	err := h.db.QueryRow(`SELECT recorded_at FROM equipment_telemetry WHERE id = $1`, id).Scan(&recordedAt)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "telemetry record not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "recorded_at": recordedAt})
}

// =============================================================================
// Fuel
// =============================================================================

func (h *EquipmentHandler) ListFuel(w http.ResponseWriter, r *http.Request) {
	equipmentID := r.URL.Query().Get("equipment_id")

	query := `SELECT id, equipment_id, refuel_date, fuel_type, quantity_liters, cost_per_liter, total_cost, meter_reading, operator, vendor, created_at FROM equipment_fuel WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if equipmentID != "" { query += fmt.Sprintf(" AND equipment_id = $%d", argIdx); argIdx++; args = append(args, equipmentID) }
	query += " ORDER BY refuel_date DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, eid, ftype, op, vendor string
		var refuelDate, createdAt time.Time
		var qty, costPerL, totalCost, meterRead float64
		err := rows.Scan(&id, &eid, &refuelDate, &ftype, &qty, &costPerL, &totalCost, &meterRead, &op, &vendor, &createdAt)
		if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
		items = append(items, map[string]interface{}{
			"id": id, "equipment_id": eid, "refuel_date": refuelDate, "fuel_type": ftype,
			"quantity_liters": qty, "cost_per_liter": costPerL, "total_cost": totalCost,
			"meter_reading": meterRead, "operator": op, "vendor": vendor, "created_at": createdAt,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *EquipmentHandler) CreateFuel(w http.ResponseWriter, r *http.Request) {
	var input struct {
		EquipmentID  string  `json:"equipment_id"`
		FuelType     string  `json:"fuel_type"`
		QuantityLiters float64 `json:"quantity_liters"`
		CostPerLiter *float64 `json:"cost_per_liter"`
		MeterReading *float64 `json:"meter_reading"`
		Operator     *string `json:"operator"`
		Vendor       *string `json:"vendor"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	now := time.Now()
	totalCost := input.QuantityLiters
	if input.CostPerLiter != nil { totalCost = input.QuantityLiters * *input.CostPerLiter }
	_, err := h.db.Exec(`INSERT INTO equipment_fuel (id, equipment_id, refuel_date, fuel_type, quantity_liters, cost_per_liter, total_cost, meter_reading, operator, vendor, created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		id, input.EquipmentID, now, input.FuelType, input.QuantityLiters, input.CostPerLiter, totalCost, input.MeterReading, input.Operator, input.Vendor, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *EquipmentHandler) GetFuel(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var qty float64
	err := h.db.QueryRow(`SELECT quantity_liters FROM equipment_fuel WHERE id = $1`, id).Scan(&qty)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "fuel record not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "quantity_liters": qty})
}

// =============================================================================
// Operators
// =============================================================================

func (h *EquipmentHandler) ListOperators(w http.ResponseWriter, r *http.Request) {
	equipmentID := r.URL.Query().Get("equipment_id")

	query := `SELECT id, equipment_id, employee_id, full_name, certification, certification_expiry, assigned_date, end_date, shift, is_primary, notes, created_at FROM equipment_operators WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if equipmentID != "" { query += fmt.Sprintf(" AND equipment_id = $%d", argIdx); argIdx++; args = append(args, equipmentID) }
	query += " ORDER BY assigned_date DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, eid, empID, fullName, cert, shift, notes string
		var certExpiry, assignedDate, endDate, createdAt sql.NullString
		var isPrimary bool
		err := rows.Scan(&id, &eid, &empID, &fullName, &cert, &certExpiry, &assignedDate, &endDate, &shift, &isPrimary, &notes, &createdAt)
		if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
		item := map[string]interface{}{
			"id": id, "equipment_id": eid, "employee_id": empID, "full_name": fullName,
			"certification": cert, "shift": shift, "is_primary": isPrimary, "notes": notes,
		}
		if certExpiry.Valid { item["certification_expiry"] = certExpiry.String }
		if assignedDate.Valid { item["assigned_date"] = assignedDate.String }
		if endDate.Valid { item["end_date"] = endDate.String }
		items = append(items, item)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *EquipmentHandler) CreateOperator(w http.ResponseWriter, r *http.Request) {
	var input struct {
		EquipmentID    string  `json:"equipment_id"`
		EmployeeID     string  `json:"employee_id"`
		FullName       string  `json:"full_name"`
		Certification  *string `json:"certification"`
		Shift          *string `json:"shift"`
		IsPrimary      *bool   `json:"is_primary"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	_, err := h.db.Exec(`INSERT INTO equipment_operators (id, equipment_id, employee_id, full_name, certification, shift, is_primary) VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		id, input.EquipmentID, input.EmployeeID, input.FullName, input.Certification, input.Shift, input.IsPrimary)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *EquipmentHandler) GetOperator(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var fullName string
	err := h.db.QueryRow(`SELECT full_name FROM equipment_operators WHERE id = $1`, id).Scan(&fullName)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "operator not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "full_name": fullName})
}

func (h *EquipmentHandler) DeleteOperator(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM equipment_operators WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// Downtime
// =============================================================================

func (h *EquipmentHandler) ListDowntime(w http.ResponseWriter, r *http.Request) {
	equipmentID := r.URL.Query().Get("equipment_id")
	status := r.URL.Query().Get("status")

	query := `SELECT id, equipment_id, downtime_type, start_time, end_time, duration_hours, reason, impact, cost_impact, reported_by, status, resolution, created_at, updated_at FROM equipment_downtime WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if equipmentID != "" { query += fmt.Sprintf(" AND equipment_id = $%d", argIdx); argIdx++; args = append(args, equipmentID) }
	if status != "" { query += fmt.Sprintf(" AND status = $%d", argIdx); argIdx++; args = append(args, status) }
	query += " ORDER BY start_time DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, eid, dtype, reason, impact, reportedBy, status, res string
		var startTime, endTime, createdAt, updatedAt sql.NullString
		var durHours, costImpact sql.NullFloat64
		err := rows.Scan(&id, &eid, &dtype, &startTime, &endTime, &durHours, &reason, &impact, &costImpact, &reportedBy, &status, &res, &createdAt, &updatedAt)
		if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
		item := map[string]interface{}{
			"id": id, "equipment_id": eid, "downtime_type": dtype, "reason": reason,
			"impact": impact, "reported_by": reportedBy, "status": status, "resolution": res,
		}
		if startTime.Valid { item["start_time"] = startTime.String }
		if endTime.Valid { item["end_time"] = endTime.String }
		if durHours.Valid { item["duration_hours"] = durHours.Float64 }
		if costImpact.Valid { item["cost_impact"] = costImpact.Float64 }
		items = append(items, item)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *EquipmentHandler) CreateDowntime(w http.ResponseWriter, r *http.Request) {
	var input struct {
		EquipmentID  string  `json:"equipment_id"`
		DowntimeType string  `json:"downtime_type"`
		StartTime    string  `json:"start_time"`
		EndTime      *string `json:"end_time"`
		Reason       *string `json:"reason"`
		CostImpact   *float64 `json:"cost_impact"`
		ReportedBy   *string `json:"reported_by"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO equipment_downtime (id, equipment_id, downtime_type, start_time, end_time, reason, cost_impact, reported_by, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		id, input.EquipmentID, input.DowntimeType, input.StartTime, input.EndTime, input.Reason, input.CostImpact, input.ReportedBy, now, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *EquipmentHandler) GetDowntime(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var dtype string
	err := h.db.QueryRow(`SELECT downtime_type FROM equipment_downtime WHERE id = $1`, id).Scan(&dtype)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "downtime record not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "downtime_type": dtype})
}

func (h *EquipmentHandler) UpdateDowntime(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Status       *string  `json:"status"`
		EndTime      *string  `json:"end_time"`
		DurationHours *float64 `json:"duration_hours"`
		Resolution   *string  `json:"resolution"`
		CostImpact   *float64 `json:"cost_impact"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	now := time.Now()
	_, err := h.db.Exec(`UPDATE equipment_downtime SET status=COALESCE($1,status), end_time=COALESCE($2,end_time), duration_hours=COALESCE($3,duration_hours), resolution=COALESCE($4,resolution), cost_impact=COALESCE($5,cost_impact), updated_at=$6 WHERE id=$7`,
		input.Status, input.EndTime, input.DurationHours, input.Resolution, input.CostImpact, now, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *EquipmentHandler) DeleteDowntime(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM equipment_downtime WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// Spare Parts
// =============================================================================

func (h *EquipmentHandler) ListSpareParts(w http.ResponseWriter, r *http.Request) {
	equipmentID := r.URL.Query().Get("equipment_id")
	lowStock := r.URL.Query().Get("low_stock")

	query := `SELECT id, equipment_id, part_code, part_name, part_number, category, unit, quantity_on_hand, min_stock_level, unit_cost, supplier, lead_time_days, storage_location, last_restocked, created_at, updated_at FROM equipment_spare_parts WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if equipmentID != "" { query += fmt.Sprintf(" AND equipment_id = $%d", argIdx); argIdx++; args = append(args, equipmentID) }
	if lowStock == "true" { query += " AND quantity_on_hand <= min_stock_level" }
	query += " ORDER BY part_code"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, eid, pcode, pname, pnum, cat, unit, supp, loc string
		var qty, minStock, unitCost float64
		var leadTime int
		var lastRestocked, createdAt, updatedAt sql.NullString
		err := rows.Scan(&id, &eid, &pcode, &pname, &pnum, &cat, &unit, &qty, &minStock, &unitCost, &supp, &leadTime, &loc, &lastRestocked, &createdAt, &updatedAt)
		if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
		item := map[string]interface{}{
			"id": id, "equipment_id": eid, "part_code": pcode, "part_name": pname,
			"part_number": pnum, "category": cat, "unit": unit, "quantity_on_hand": qty,
			"min_stock_level": minStock, "unit_cost": unitCost, "supplier": supp,
			"lead_time_days": leadTime, "storage_location": loc,
		}
		if lastRestocked.Valid { item["last_restocked"] = lastRestocked.String }
		items = append(items, item)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *EquipmentHandler) CreateSparePart(w http.ResponseWriter, r *http.Request) {
	var input struct {
		EquipmentID  *string `json:"equipment_id"`
		PartCode     string  `json:"part_code"`
		PartName     string  `json:"part_name"`
		PartNumber   *string `json:"part_number"`
		Category     *string `json:"category"`
		QuantityOnHand *float64 `json:"quantity_on_hand"`
		MinStockLevel *float64 `json:"min_stock_level"`
		UnitCost     *float64 `json:"unit_cost"`
		Supplier     *string `json:"supplier"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO equipment_spare_parts (id, equipment_id, part_code, part_name, part_number, category, quantity_on_hand, min_stock_level, unit_cost, supplier, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		id, input.EquipmentID, input.PartCode, input.PartName, input.PartNumber, input.Category, input.QuantityOnHand, input.MinStockLevel, input.UnitCost, input.Supplier, now, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *EquipmentHandler) GetSparePart(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var pcode, pname string
	err := h.db.QueryRow(`SELECT part_code, part_name FROM equipment_spare_parts WHERE id = $1`, id).Scan(&pcode, &pname)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "spare part not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "part_code": pcode, "part_name": pname})
}

func (h *EquipmentHandler) UpdateSparePart(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		QuantityOnHand *float64 `json:"quantity_on_hand"`
		MinStockLevel  *float64 `json:"min_stock_level"`
		UnitCost       *float64 `json:"unit_cost"`
		Supplier       *string  `json:"supplier"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	now := time.Now()
	_, err := h.db.Exec(`UPDATE equipment_spare_parts SET quantity_on_hand=COALESCE($1,quantity_on_hand), min_stock_level=COALESCE($2,min_stock_level), unit_cost=COALESCE($3,unit_cost), supplier=COALESCE($4,supplier), updated_at=$5 WHERE id=$6`,
		input.QuantityOnHand, input.MinStockLevel, input.UnitCost, input.Supplier, now, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *EquipmentHandler) DeleteSparePart(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM equipment_spare_parts WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// Summary
// =============================================================================

func (h *EquipmentHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")

	query := `SELECT project_id, total_equipment, available, in_use, under_maintenance, out_of_service, tbms, cranes, fleet, pending_maintenance, active_downtime, downtime_30d, refuels_30d FROM equipment_summary`
	args := []interface{}{}
	argIdx := 1
	if projectID != "" { query += fmt.Sprintf(" WHERE project_id = $%d", argIdx); args = append(args, projectID) }

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var pid string
		var total, avail, inUse, underMaint, outOfService, tbms, cranes, fleet, pendingMaint, activeDowntime, downtime30d, refuels30d int
		err := rows.Scan(&pid, &total, &avail, &inUse, &underMaint, &outOfService, &tbms, &cranes, &fleet, &pendingMaint, &activeDowntime, &downtime30d, &refuels30d)
		if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
		items = append(items, map[string]interface{}{
			"project_id": pid, "total_equipment": total, "available": avail, "in_use": inUse,
			"under_maintenance": underMaint, "out_of_service": outOfService, "tbms": tbms,
			"cranes": cranes, "fleet": fleet, "pending_maintenance": pendingMaint,
			"active_downtime": activeDowntime, "downtime_30d": downtime30d, "refuels_30d": refuels30d,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func init() {
	log.Println("Equipment Management handler initialized")
}