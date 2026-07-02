package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/openconstructionerp/oce/services/core/internal/models"
)

// FleetHandler handles Fleet module endpoints (V032)
type FleetHandler struct {
	db *sql.DB
}

func NewFleetHandler(db *sql.DB) *FleetHandler {
	return &FleetHandler{db: db}
}

func (h *FleetHandler) RegisterRoutes(r chi.Router) {
	r.Route("/fleet", func(r chi.Router) {
		// Vehicles
		r.Get("/vehicles", h.ListVehicles)
		r.Post("/vehicles", h.CreateVehicle)
		r.Get("/vehicles/{id}", h.GetVehicle)
		r.Put("/vehicles/{id}", h.UpdateVehicle)

		// Drivers
		r.Get("/drivers", h.ListDrivers)
		r.Post("/drivers", h.CreateDriver)
		r.Get("/drivers/{id}", h.GetDriver)

		// Fuel
		r.Get("/fuel", h.ListFuel)
		r.Post("/fuel", h.CreateFuel)

		// Maintenance
		r.Get("/maintenance", h.ListMaintenance)
		r.Post("/maintenance", h.CreateMaintenance)
		r.Put("/maintenance/{id}", h.UpdateMaintenance)

		// Accidents
		r.Get("/accidents", h.ListAccidents)
		r.Post("/accidents", h.CreateAccident)
		r.Put("/accidents/{id}", h.UpdateAccident)

		// Tracking
		r.Get("/tracking", h.ListTracking)
		r.Post("/tracking", h.CreateTracking)

		// Telematics
		r.Get("/telematics/{vehicleId}", h.ListTelematics)
		r.Post("/telematics/{vehicleId}", h.CreateTelematics)
	})
}

// --- Vehicles ---

func (h *FleetHandler) ListVehicles(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT id,project_id,equipment_id,vehicle_type,make,model,year,vin,license_plate,registration_number,fuel_type,engine_capacity,horsepower,weight_kg,load_capacity_kg,status,assigned_driver,location,mileage_km,is_active,notes,created_at,updated_at FROM fleet_vehicles`
	var rows *sql.Rows
	var err error
	if projectID != "" {
		rows, err = h.db.Query(query+` WHERE project_id=$1 ORDER BY created_at DESC`, projectID)
	} else {
		rows, err = h.db.Query(query + ` ORDER BY created_at DESC`)
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	items := make([]models.FleetVehicle, 0)
	for rows.Next() {
		var m models.FleetVehicle
		if err := rows.Scan(&m.ID, &m.ProjectID, &m.EquipmentID, &m.VehicleType, &m.Make, &m.Model, &m.Year, &m.VIN, &m.LicensePlate, &m.RegistrationNum, &m.FuelType, &m.EngineCapacity, &m.Horsepower, &m.WeightKg, &m.LoadCapacityKg, &m.Status, &m.AssignedDriver, &m.Location, &m.MileageKm, &m.IsActive, &m.Notes, &m.CreatedAt, &m.UpdatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, m)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *FleetHandler) CreateVehicle(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID    string  `json:"project_id"`
		VehicleType  string  `json:"vehicle_type"`
		Make         *string `json:"make"`
		Model        *string `json:"model"`
		Year         *int    `json:"year"`
		VIN          *string `json:"vin"`
		LicensePlate *string `json:"license_plate"`
		FuelType     *string `json:"fuel_type"`
		Status       *string `json:"status"`
		Notes        *string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	now := time.Now()
	status := "operational"
	if input.Status != nil {
		status = *input.Status
	}
	_, err := h.db.Exec(`INSERT INTO fleet_vehicles (id,project_id,vehicle_type,make,model,year,vin,license_plate,fuel_type,status,notes,created_at,updated_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		id, input.ProjectID, input.VehicleType, input.Make, input.Model, input.Year, input.VIN, input.LicensePlate, input.FuelType, status, input.Notes, now, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *FleetHandler) GetVehicle(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var m models.FleetVehicle
	err := h.db.QueryRow(`SELECT id,project_id,equipment_id,vehicle_type,make,model,year,vin,license_plate,registration_number,fuel_type,engine_capacity,horsepower,weight_kg,load_capacity_kg,status,assigned_driver,location,mileage_km,is_active,notes,created_at,updated_at FROM fleet_vehicles WHERE id=$1`, id).Scan(
		&m.ID, &m.ProjectID, &m.EquipmentID, &m.VehicleType, &m.Make, &m.Model, &m.Year, &m.VIN, &m.LicensePlate, &m.RegistrationNum, &m.FuelType, &m.EngineCapacity, &m.Horsepower, &m.WeightKg, &m.LoadCapacityKg, &m.Status, &m.AssignedDriver, &m.Location, &m.MileageKm, &m.IsActive, &m.Notes, &m.CreatedAt, &m.UpdatedAt)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, m)
}

func (h *FleetHandler) UpdateVehicle(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Status         *string  `json:"status"`
		AssignedDriver *string  `json:"assigned_driver"`
		Location       *string  `json:"location"`
		MileageKm      *float64 `json:"mileage_km"`
		Notes          *string  `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	_, err := h.db.Exec(`UPDATE fleet_vehicles SET status=COALESCE($1,status), assigned_driver=COALESCE($2,assigned_driver), location=COALESCE($3,location), mileage_km=COALESCE($4,mileage_km), notes=COALESCE($5,notes), updated_at=NOW() WHERE id=$6`,
		input.Status, input.AssignedDriver, input.Location, input.MileageKm, input.Notes, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// --- Drivers ---

func (h *FleetHandler) ListDrivers(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT id,project_id,driver_name,license_number,license_type,license_expiry,contact_phone,email,certifications,status,is_active,notes,created_at,updated_at FROM vehicle_drivers`
	var rows *sql.Rows
	var err error
	if projectID != "" {
		rows, err = h.db.Query(query+` WHERE project_id=$1 ORDER BY driver_name`, projectID)
	} else {
		rows, err = h.db.Query(query + ` ORDER BY driver_name`)
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	items := make([]models.VehicleDriver, 0)
	for rows.Next() {
		var m models.VehicleDriver
		if err := rows.Scan(&m.ID, &m.ProjectID, &m.DriverName, &m.LicenseNumber, &m.LicenseType, &m.LicenseExpiry, &m.ContactPhone, &m.Email, &m.Certifications, &m.Status, &m.IsActive, &m.Notes, &m.CreatedAt, &m.UpdatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, m)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *FleetHandler) CreateDriver(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID    string  `json:"project_id"`
		DriverName   string  `json:"driver_name"`
		LicenseNumber *string `json:"license_number"`
		LicenseType  *string `json:"license_type"`
		LicenseExpiry *string `json:"license_expiry"`
		ContactPhone *string `json:"contact_phone"`
		Email        *string `json:"email"`
		Notes        *string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO vehicle_drivers (id,project_id,driver_name,license_number,license_type,license_expiry,contact_phone,email,notes,created_at,updated_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		id, input.ProjectID, input.DriverName, input.LicenseNumber, input.LicenseType, input.LicenseExpiry, input.ContactPhone, input.Email, input.Notes, now, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *FleetHandler) GetDriver(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var m models.VehicleDriver
	err := h.db.QueryRow(`SELECT id,project_id,driver_name,license_number,license_type,license_expiry,contact_phone,email,certifications,status,is_active,notes,created_at,updated_at FROM vehicle_drivers WHERE id=$1`, id).Scan(
		&m.ID, &m.ProjectID, &m.DriverName, &m.LicenseNumber, &m.LicenseType, &m.LicenseExpiry, &m.ContactPhone, &m.Email, &m.Certifications, &m.Status, &m.IsActive, &m.Notes, &m.CreatedAt, &m.UpdatedAt)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, m)
}

// --- Fuel ---

func (h *FleetHandler) ListFuel(w http.ResponseWriter, r *http.Request) {
	vehicleID := r.URL.Query().Get("vehicle_id")
	query := `SELECT id,project_id,vehicle_id,driver_id,fuel_date,fuel_type,quantity_liters,unit_price,total_cost,currency,odometer_km,station_name,receipt_number,notes,created_at FROM vehicle_fuel`
	var rows *sql.Rows
	var err error
	if vehicleID != "" {
		rows, err = h.db.Query(query+` WHERE vehicle_id=$1 ORDER BY fuel_date DESC`, vehicleID)
	} else {
		rows, err = h.db.Query(query + ` ORDER BY fuel_date DESC`)
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	items := make([]models.VehicleFuel, 0)
	for rows.Next() {
		var m models.VehicleFuel
		if err := rows.Scan(&m.ID, &m.ProjectID, &m.VehicleID, &m.DriverID, &m.FuelDate, &m.FuelType, &m.QuantityLiters, &m.UnitPrice, &m.TotalCost, &m.Currency, &m.OdometerKm, &m.StationName, &m.ReceiptNumber, &m.Notes, &m.CreatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, m)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *FleetHandler) CreateFuel(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID      string   `json:"project_id"`
		VehicleID      string   `json:"vehicle_id"`
		DriverID       *string  `json:"driver_id"`
		FuelDate       string   `json:"fuel_date"`
		FuelType       *string  `json:"fuel_type"`
		QuantityLiters float64  `json:"quantity_liters"`
		UnitPrice      float64  `json:"unit_price"`
		OdometerKm     *float64 `json:"odometer_km"`
		StationName    *string  `json:"station_name"`
		Notes          *string  `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	totalCost := input.QuantityLiters * input.UnitPrice
	_, err := h.db.Exec(`INSERT INTO vehicle_fuel (id,project_id,vehicle_id,driver_id,fuel_date,fuel_type,quantity_liters,unit_price,total_cost,odometer_km,station_name,notes,created_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,NOW())`,
		id, input.ProjectID, input.VehicleID, input.DriverID, input.FuelDate, input.FuelType, input.QuantityLiters, input.UnitPrice, totalCost, input.OdometerKm, input.StationName, input.Notes)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

// --- Maintenance ---

func (h *FleetHandler) ListMaintenance(w http.ResponseWriter, r *http.Request) {
	vehicleID := r.URL.Query().Get("vehicle_id")
	query := `SELECT id,project_id,vehicle_id,maintenance_type,description,scheduled_date,completed_date,odometer_km,cost_amount,currency,vendor,invoice_number,status,notes,created_at,updated_at FROM vehicle_maintenance`
	var rows *sql.Rows
	var err error
	if vehicleID != "" {
		rows, err = h.db.Query(query+` WHERE vehicle_id=$1 ORDER BY scheduled_date DESC`, vehicleID)
	} else {
		rows, err = h.db.Query(query + ` ORDER BY scheduled_date DESC`)
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	items := make([]models.VehicleMaintenance, 0)
	for rows.Next() {
		var m models.VehicleMaintenance
		if err := rows.Scan(&m.ID, &m.ProjectID, &m.VehicleID, &m.MaintenanceType, &m.Description, &m.ScheduledDate, &m.CompletedDate, &m.OdometerKm, &m.CostAmount, &m.Currency, &m.Vendor, &m.InvoiceNumber, &m.Status, &m.Notes, &m.CreatedAt, &m.UpdatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, m)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *FleetHandler) CreateMaintenance(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID       string   `json:"project_id"`
		VehicleID       string   `json:"vehicle_id"`
		MaintenanceType string   `json:"maintenance_type"`
		Description     *string  `json:"description"`
		ScheduledDate   *string  `json:"scheduled_date"`
		OdometerKm      *float64 `json:"odometer_km"`
		Vendor          *string  `json:"vendor"`
		Notes           *string  `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO vehicle_maintenance (id,project_id,vehicle_id,maintenance_type,description,scheduled_date,odometer_km,vendor,notes,status,created_at,updated_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,'scheduled',$10,$11)`,
		id, input.ProjectID, input.VehicleID, input.MaintenanceType, input.Description, input.ScheduledDate, input.OdometerKm, input.Vendor, input.Notes, now, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *FleetHandler) UpdateMaintenance(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Status        *string  `json:"status"`
		CompletedDate *string  `json:"completed_date"`
		CostAmount    *float64 `json:"cost_amount"`
		InvoiceNumber *string  `json:"invoice_number"`
		Notes         *string  `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	_, err := h.db.Exec(`UPDATE vehicle_maintenance SET status=COALESCE($1,status), completed_date=COALESCE($2,completed_date), cost_amount=COALESCE($3,cost_amount), invoice_number=COALESCE($4,invoice_number), notes=COALESCE($5,notes), updated_at=NOW() WHERE id=$6`,
		input.Status, input.CompletedDate, input.CostAmount, input.InvoiceNumber, input.Notes, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// --- Accidents ---

func (h *FleetHandler) ListAccidents(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT id,project_id,vehicle_id,driver_id,accident_date,location,description,severity,damages,injuries,fatalities,police_report,insurance_claim_id,cost_estimate,status,notes,created_at,updated_at FROM vehicle_accidents`
	var rows *sql.Rows
	var err error
	if projectID != "" {
		rows, err = h.db.Query(query+` WHERE project_id=$1 ORDER BY accident_date DESC`, projectID)
	} else {
		rows, err = h.db.Query(query + ` ORDER BY accident_date DESC`)
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	items := make([]models.VehicleAccident, 0)
	for rows.Next() {
		var m models.VehicleAccident
		if err := rows.Scan(&m.ID, &m.ProjectID, &m.VehicleID, &m.DriverID, &m.AccidentDate, &m.Location, &m.Description, &m.Severity, &m.Damages, &m.Injuries, &m.Fatalities, &m.PoliceReport, &m.InsuranceClaimID, &m.CostEstimate, &m.Status, &m.Notes, &m.CreatedAt, &m.UpdatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, m)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *FleetHandler) CreateAccident(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID   string  `json:"project_id"`
		VehicleID   string  `json:"vehicle_id"`
		DriverID    *string `json:"driver_id"`
		AccidentDate string  `json:"accident_date"`
		Location    *string `json:"location"`
		Description *string `json:"description"`
		Severity    *string `json:"severity"`
		Notes       *string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO vehicle_accidents (id,project_id,vehicle_id,driver_id,accident_date,location,description,severity,status,notes,created_at,updated_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,'reported',$9,$10,$11)`,
		id, input.ProjectID, input.VehicleID, input.DriverID, input.AccidentDate, input.Location, input.Description, input.Severity, input.Notes, now, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *FleetHandler) UpdateAccident(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Status          *string  `json:"status"`
		CostEstimate    *float64 `json:"cost_estimate"`
		InsuranceClaimID *string `json:"insurance_claim_id"`
		Notes           *string  `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	_, err := h.db.Exec(`UPDATE vehicle_accidents SET status=COALESCE($1,status), cost_estimate=COALESCE($2,cost_estimate), insurance_claim_id=COALESCE($3,insurance_claim_id), notes=COALESCE($4,notes), updated_at=NOW() WHERE id=$5`,
		input.Status, input.CostEstimate, input.InsuranceClaimID, input.Notes, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// --- Tracking ---

func (h *FleetHandler) ListTracking(w http.ResponseWriter, r *http.Request) {
	vehicleID := r.URL.Query().Get("vehicle_id")
	date := r.URL.Query().Get("date")
	query := `SELECT id,vehicle_id,driver_id,track_date,start_time,end_time,start_location,end_location,distance_km,duration_minutes,purpose,notes,created_at FROM vehicle_tracking`
	var rows *sql.Rows
	var err error
	if vehicleID != "" && date != "" {
		rows, err = h.db.Query(query+` WHERE vehicle_id=$1 AND track_date=$2 ORDER BY start_time`, vehicleID, date)
	} else if vehicleID != "" {
		rows, err = h.db.Query(query+` WHERE vehicle_id=$1 ORDER BY track_date DESC`, vehicleID)
	} else {
		rows, err = h.db.Query(query + ` ORDER BY track_date DESC LIMIT 50`)
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	items := make([]models.VehicleTracking, 0)
	for rows.Next() {
		var m models.VehicleTracking
		if err := rows.Scan(&m.ID, &m.VehicleID, &m.DriverID, &m.TrackDate, &m.StartTime, &m.EndTime, &m.StartLocation, &m.EndLocation, &m.DistanceKm, &m.DurationMinutes, &m.Purpose, &m.Notes, &m.CreatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, m)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *FleetHandler) CreateTracking(w http.ResponseWriter, r *http.Request) {
	var input struct {
		VehicleID       string   `json:"vehicle_id"`
		DriverID        *string  `json:"driver_id"`
		TrackDate       string   `json:"track_date"`
		StartTime       *string  `json:"start_time"`
		EndTime         *string  `json:"end_time"`
		StartLocation   *string  `json:"start_location"`
		EndLocation     *string  `json:"end_location"`
		DistanceKm      *float64 `json:"distance_km"`
		DurationMinutes *int     `json:"duration_minutes"`
		Purpose         *string  `json:"purpose"`
		Notes           *string  `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	_, err := h.db.Exec(`INSERT INTO vehicle_tracking (id,vehicle_id,driver_id,track_date,start_time,end_time,start_location,end_location,distance_km,duration_minutes,purpose,notes,created_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,NOW())`,
		id, input.VehicleID, input.DriverID, input.TrackDate, input.StartTime, input.EndTime, input.StartLocation, input.EndLocation, input.DistanceKm, input.DurationMinutes, input.Purpose, input.Notes)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

// --- Telematics ---

func (h *FleetHandler) ListTelematics(w http.ResponseWriter, r *http.Request) {
	vehicleID := chi.URLParam(r, "vehicleId")
	limit := r.URL.Query().Get("limit")
	q := `SELECT id,vehicle_id,recorded_at,latitude,longitude,speed_kph,heading,altitude_m,engine_temp,fuel_level_pct,battery_voltage,tire_pressure,engine_rpm,odometer_km,diagnostics FROM vehicle_telematics WHERE vehicle_id=$1 ORDER BY recorded_at DESC`
	if limit != "" {
		q += ` LIMIT ` + limit
	} else {
		q += ` LIMIT 100`
	}
	rows, err := h.db.Query(q, vehicleID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	items := make([]models.VehicleTelematics, 0)
	for rows.Next() {
		var m models.VehicleTelematics
		if err := rows.Scan(&m.ID, &m.VehicleID, &m.RecordedAt, &m.Latitude, &m.Longitude, &m.SpeedKph, &m.Heading, &m.AltitudeM, &m.EngineTemp, &m.FuelLevelPct, &m.BatteryVoltage, &m.TirePressure, &m.EngineRPM, &m.OdometerKm, &m.Diagnostics); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		items = append(items, m)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *FleetHandler) CreateTelematics(w http.ResponseWriter, r *http.Request) {
	vehicleID := chi.URLParam(r, "vehicleId")
	var input struct {
		Latitude      *float64 `json:"latitude"`
		Longitude     *float64 `json:"longitude"`
		SpeedKph      *float64 `json:"speed_kph"`
		Heading       *float64 `json:"heading"`
		FuelLevelPct  *float64 `json:"fuel_level_pct"`
		OdometerKm    *float64 `json:"odometer_km"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	_, err := h.db.Exec(`INSERT INTO vehicle_telematics (id,vehicle_id,recorded_at,latitude,longitude,speed_kph,heading,fuel_level_pct,odometer_km) VALUES($1,$2,NOW(),$3,$4,$5,$6,$7,$8)`,
		id, vehicleID, input.Latitude, input.Longitude, input.SpeedKph, input.Heading, input.FuelLevelPct, input.OdometerKm)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}