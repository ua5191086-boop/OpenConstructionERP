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

// HRHandler handles HR module endpoints
type HRHandler struct {
	db *sql.DB
}

func NewHRHandler(db *sql.DB) *HRHandler {
	return &HRHandler{db: db}
}

func (h *HRHandler) RegisterRoutes(r chi.Router) {
	r.Route("/hr", func(r chi.Router) {
		// Employees
		r.Get("/employees", h.ListEmployees)
		r.Post("/employees", h.CreateEmployee)
		r.Get("/employees/{id}", h.GetEmployee)
		r.Put("/employees/{id}", h.UpdateEmployee)
		r.Delete("/employees/{id}", h.DeleteEmployee)

		// Departments
		r.Get("/departments", h.ListDepartments)
		r.Post("/departments", h.CreateDepartment)
		r.Get("/departments/{id}", h.GetDepartment)
		r.Put("/departments/{id}", h.UpdateDepartment)
		r.Delete("/departments/{id}", h.DeleteDepartment)

		// Time Attendance
		r.Get("/attendance", h.ListAttendance)
		r.Post("/attendance", h.CreateAttendance)
		r.Get("/attendance/{id}", h.GetAttendance)
		r.Put("/attendance/{id}", h.UpdateAttendance)
		r.Delete("/attendance/{id}", h.DeleteAttendance)

		// Leave
		r.Get("/leaves", h.ListLeaves)
		r.Post("/leaves", h.CreateLeave)
		r.Get("/leaves/{id}", h.GetLeave)
		r.Put("/leaves/{id}", h.UpdateLeave)
		r.Delete("/leaves/{id}", h.DeleteLeave)
	})
}

// --- Employees ---

func (h *HRHandler) ListEmployees(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	department := r.URL.Query().Get("department")

	query := `SELECT id, employee_code, full_name, first_name, last_name, patronymic, birth_date, gender, nationality, email, phone, phone_emergency, address, position, department, position_type, position_category, grade, status, hire_date, contract_end, termination_date, termination_reason, salary_base, salary_currency, hourly_rate, bank_name, bank_account, tax_id, social_security_id, education, certifications, skills, experience_years, passport_number, passport_expiry, work_permit, work_permit_expiry, medical_checkup_date, medical_checkup_valid_until, photo_path, resume_path, notes, created_by, created_at, updated_at FROM employees WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if status != "" {
		query += ` AND status = $` + itoa(argIdx)
		args = append(args, status)
		argIdx++
	}
	if department != "" {
		query += ` AND department = $` + itoa(argIdx)
		args = append(args, department)
		argIdx++
	}
	query += ` ORDER BY full_name`

	rows, err := h.db.Query(query, args...)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	employees := make([]models.Employee, 0)
	for rows.Next() {
		var e models.Employee
		if err := rows.Scan(&e.ID, &e.EmployeeCode, &e.FullName, &e.FirstName, &e.LastName, &e.Patronymic, &e.BirthDate, &e.Gender, &e.Nationality, &e.Email, &e.Phone, &e.PhoneEmergency, &e.Address, &e.Position, &e.Department, &e.PositionType, &e.PositionCategory, &e.Grade, &e.Status, &e.HireDate, &e.ContractEnd, &e.TerminationDate, &e.TerminationReason, &e.SalaryBase, &e.SalaryCurrency, &e.HourlyRate, &e.BankName, &e.BankAccount, &e.TaxID, &e.SocialSecurityID, &e.Education, &e.Certifications, &e.Skills, &e.ExperienceYears, &e.PassportNumber, &e.PassportExpiry, &e.WorkPermit, &e.WorkPermitExpiry, &e.MedicalCheckupDate, &e.MedicalCheckupValidUntil, &e.PhotoPath, &e.ResumePath, &e.Notes, &e.CreatedBy, &e.CreatedAt, &e.UpdatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		employees = append(employees, e)
	}
	respondJSON(w, http.StatusOK, employees)
}

func (h *HRHandler) CreateEmployee(w http.ResponseWriter, r *http.Request) {
	var input struct {
		EmployeeCode string  `json:"employee_code"`
		FullName     string  `json:"full_name"`
		FirstName    *string `json:"first_name"`
		LastName     *string `json:"last_name"`
		Patronymic   *string `json:"patronymic"`
		BirthDate    *string `json:"birth_date"`
		Gender       *string `json:"gender"`
		Nationality  *string `json:"nationality"`
		Email        *string `json:"email"`
		Phone        *string `json:"phone"`
		Position     string  `json:"position"`
		Department   *string `json:"department"`
		PositionType string  `json:"position_type"`
		Status       string  `json:"status"`
		HireDate     string  `json:"hire_date"`
		SalaryBase   *float64 `json:"salary_base"`
		SalaryCurrency string `json:"salary_currency"`
		Notes        *string `json:"notes"`
		CreatedBy    *string `json:"created_by"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO employees (id, employee_code, full_name, first_name, last_name, patronymic, birth_date, gender, nationality, email, phone, position, department, position_type, status, hire_date, salary_base, salary_currency, notes, created_by, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22)`,
		id, input.EmployeeCode, input.FullName, input.FirstName, input.LastName, input.Patronymic, input.BirthDate, input.Gender, input.Nationality, input.Email, input.Phone, input.Position, input.Department, input.PositionType, input.Status, input.HireDate, input.SalaryBase, input.SalaryCurrency, input.Notes, input.CreatedBy, now, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *HRHandler) GetEmployee(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var e models.Employee
	err := h.db.QueryRow(`SELECT id, employee_code, full_name, first_name, last_name, patronymic, birth_date, gender, nationality, email, phone, phone_emergency, address, position, department, position_type, position_category, grade, status, hire_date, contract_end, termination_date, termination_reason, salary_base, salary_currency, hourly_rate, bank_name, bank_account, tax_id, social_security_id, education, certifications, skills, experience_years, passport_number, passport_expiry, work_permit, work_permit_expiry, medical_checkup_date, medical_checkup_valid_until, photo_path, resume_path, notes, created_by, created_at, updated_at FROM employees WHERE id = $1`, id).
		Scan(&e.ID, &e.EmployeeCode, &e.FullName, &e.FirstName, &e.LastName, &e.Patronymic, &e.BirthDate, &e.Gender, &e.Nationality, &e.Email, &e.Phone, &e.PhoneEmergency, &e.Address, &e.Position, &e.Department, &e.PositionType, &e.PositionCategory, &e.Grade, &e.Status, &e.HireDate, &e.ContractEnd, &e.TerminationDate, &e.TerminationReason, &e.SalaryBase, &e.SalaryCurrency, &e.HourlyRate, &e.BankName, &e.BankAccount, &e.TaxID, &e.SocialSecurityID, &e.Education, &e.Certifications, &e.Skills, &e.ExperienceYears, &e.PassportNumber, &e.PassportExpiry, &e.WorkPermit, &e.WorkPermitExpiry, &e.MedicalCheckupDate, &e.MedicalCheckupValidUntil, &e.PhotoPath, &e.ResumePath, &e.Notes, &e.CreatedBy, &e.CreatedAt, &e.UpdatedAt)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "employee not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, e)
}

func (h *HRHandler) UpdateEmployee(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		FullName   *string  `json:"full_name"`
		Email      *string  `json:"email"`
		Phone      *string  `json:"phone"`
		Position   *string  `json:"position"`
		Department *string  `json:"department"`
		Status     *string  `json:"status"`
		SalaryBase *float64 `json:"salary_base"`
		Notes      *string  `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	now := time.Now()
	_, err := h.db.Exec(`UPDATE employees SET full_name=COALESCE($1,full_name), email=COALESCE($2,email), phone=COALESCE($3,phone), position=COALESCE($4,position), department=COALESCE($5,department), status=COALESCE($6,status), salary_base=COALESCE($7,salary_base), notes=COALESCE($8,notes), updated_at=$9 WHERE id=$10`,
		input.FullName, input.Email, input.Phone, input.Position, input.Department, input.Status, input.SalaryBase, input.Notes, now, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *HRHandler) DeleteEmployee(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM employees WHERE id = $1`, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// --- Departments ---

func (h *HRHandler) ListDepartments(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(`SELECT id, code, name, description, parent_id, head_employee_id, cost_center, location, is_active, created_at FROM departments ORDER BY name`)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	depts := make([]models.Department, 0)
	for rows.Next() {
		var d models.Department
		if err := rows.Scan(&d.ID, &d.Code, &d.Name, &d.Description, &d.ParentID, &d.HeadEmployeeID, &d.CostCenter, &d.Location, &d.IsActive, &d.CreatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		depts = append(depts, d)
	}
	respondJSON(w, http.StatusOK, depts)
}

func (h *HRHandler) CreateDepartment(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Code          string  `json:"code"`
		Name          string  `json:"name"`
		Description   *string `json:"description"`
		ParentID      *string `json:"parent_id"`
		HeadEmployeeID *string `json:"head_employee_id"`
		CostCenter    *string `json:"cost_center"`
		Location      *string `json:"location"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO departments (id, code, name, description, parent_id, head_employee_id, cost_center, location, is_active, created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,true,$9)`,
		id, input.Code, input.Name, input.Description, input.ParentID, input.HeadEmployeeID, input.CostCenter, input.Location, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *HRHandler) GetDepartment(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var d models.Department
	err := h.db.QueryRow(`SELECT id, code, name, description, parent_id, head_employee_id, cost_center, location, is_active, created_at FROM departments WHERE id = $1`, id).
		Scan(&d.ID, &d.Code, &d.Name, &d.Description, &d.ParentID, &d.HeadEmployeeID, &d.CostCenter, &d.Location, &d.IsActive, &d.CreatedAt)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "department not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, d)
}

func (h *HRHandler) UpdateDepartment(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Name          *string `json:"name"`
		Description   *string `json:"description"`
		HeadEmployeeID *string `json:"head_employee_id"`
		IsActive      *bool   `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	_, err := h.db.Exec(`UPDATE departments SET name=COALESCE($1,name), description=COALESCE($2,description), head_employee_id=COALESCE($3,head_employee_id), is_active=COALESCE($4,is_active) WHERE id=$5`,
		input.Name, input.Description, input.HeadEmployeeID, input.IsActive, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *HRHandler) DeleteDepartment(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM departments WHERE id = $1`, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// --- Attendance ---

func (h *HRHandler) ListAttendance(w http.ResponseWriter, r *http.Request) {
	employeeID := r.URL.Query().Get("employee_id")
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")

	query := `SELECT id, employee_id, work_date, day_type, hours_worked, hours_overtime, status, reason, approved_by, approved_at, created_at FROM time_attendance WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if employeeID != "" {
		query += ` AND employee_id = $` + itoa(argIdx)
		args = append(args, employeeID)
		argIdx++
	}
	if from != "" {
		query += ` AND work_date >= $` + itoa(argIdx)
		args = append(args, from)
		argIdx++
	}
	if to != "" {
		query += ` AND work_date <= $` + itoa(argIdx)
		args = append(args, to)
		argIdx++
	}
	query += ` ORDER BY work_date DESC`

	rows, err := h.db.Query(query, args...)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	records := make([]models.TimeAttendance, 0)
	for rows.Next() {
		var a models.TimeAttendance
		if err := rows.Scan(&a.ID, &a.EmployeeID, &a.WorkDate, &a.DayType, &a.HoursWorked, &a.HoursOvertime, &a.Status, &a.Reason, &a.ApprovedBy, &a.ApprovedAt, &a.CreatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		records = append(records, a)
	}
	respondJSON(w, http.StatusOK, records)
}

func (h *HRHandler) CreateAttendance(w http.ResponseWriter, r *http.Request) {
	var input struct {
		EmployeeID    string  `json:"employee_id"`
		WorkDate      string  `json:"work_date"`
		DayType       string  `json:"day_type"`
		HoursWorked   float64 `json:"hours_worked"`
		HoursOvertime float64 `json:"hours_overtime"`
		Status        string  `json:"status"`
		Reason        *string `json:"reason"`
		ApprovedBy    *string `json:"approved_by"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO time_attendance (id, employee_id, work_date, day_type, hours_worked, hours_overtime, status, reason, approved_by, created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		id, input.EmployeeID, input.WorkDate, input.DayType, input.HoursWorked, input.HoursOvertime, input.Status, input.Reason, input.ApprovedBy, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *HRHandler) GetAttendance(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var a models.TimeAttendance
	err := h.db.QueryRow(`SELECT id, employee_id, work_date, day_type, hours_worked, hours_overtime, status, reason, approved_by, approved_at, created_at FROM time_attendance WHERE id = $1`, id).
		Scan(&a.ID, &a.EmployeeID, &a.WorkDate, &a.DayType, &a.HoursWorked, &a.HoursOvertime, &a.Status, &a.Reason, &a.ApprovedBy, &a.ApprovedAt, &a.CreatedAt)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "attendance record not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, a)
}

func (h *HRHandler) UpdateAttendance(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		HoursWorked   *float64 `json:"hours_worked"`
		HoursOvertime *float64 `json:"hours_overtime"`
		Status        *string  `json:"status"`
		Reason        *string  `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	_, err := h.db.Exec(`UPDATE time_attendance SET hours_worked=COALESCE($1,hours_worked), hours_overtime=COALESCE($2,hours_overtime), status=COALESCE($3,status), reason=COALESCE($4,reason) WHERE id=$5`,
		input.HoursWorked, input.HoursOvertime, input.Status, input.Reason, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *HRHandler) DeleteAttendance(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM time_attendance WHERE id = $1`, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// --- Leave ---

func (h *HRHandler) ListLeaves(w http.ResponseWriter, r *http.Request) {
	employeeID := r.URL.Query().Get("employee_id")
	status := r.URL.Query().Get("status")

	query := `SELECT id, employee_id, leave_type, start_date, end_date, days_count, status, reason, approved_by, approved_at, notes, created_at FROM employee_leave WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if employeeID != "" {
		query += ` AND employee_id = $` + itoa(argIdx)
		args = append(args, employeeID)
		argIdx++
	}
	if status != "" {
		query += ` AND status = $` + itoa(argIdx)
		args = append(args, status)
		argIdx++
	}
	query += ` ORDER BY start_date DESC`

	rows, err := h.db.Query(query, args...)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	leaves := make([]models.EmployeeLeave, 0)
	for rows.Next() {
		var l models.EmployeeLeave
		if err := rows.Scan(&l.ID, &l.EmployeeID, &l.LeaveType, &l.StartDate, &l.EndDate, &l.DaysCount, &l.Status, &l.Reason, &l.ApprovedBy, &l.ApprovedAt, &l.Notes, &l.CreatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		leaves = append(leaves, l)
	}
	respondJSON(w, http.StatusOK, leaves)
}

func (h *HRHandler) CreateLeave(w http.ResponseWriter, r *http.Request) {
	var input struct {
		EmployeeID string  `json:"employee_id"`
		LeaveType  string  `json:"leave_type"`
		StartDate  string  `json:"start_date"`
		EndDate    string  `json:"end_date"`
		DaysCount  int     `json:"days_count"`
		Status     string  `json:"status"`
		Reason     *string `json:"reason"`
		Notes      *string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO employee_leave (id, employee_id, leave_type, start_date, end_date, days_count, status, reason, notes, created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		id, input.EmployeeID, input.LeaveType, input.StartDate, input.EndDate, input.DaysCount, input.Status, input.Reason, input.Notes, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *HRHandler) GetLeave(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var l models.EmployeeLeave
	err := h.db.QueryRow(`SELECT id, employee_id, leave_type, start_date, end_date, days_count, status, reason, approved_by, approved_at, notes, created_at FROM employee_leave WHERE id = $1`, id).
		Scan(&l.ID, &l.EmployeeID, &l.LeaveType, &l.StartDate, &l.EndDate, &l.DaysCount, &l.Status, &l.Reason, &l.ApprovedBy, &l.ApprovedAt, &l.Notes, &l.CreatedAt)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "leave record not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, l)
}

func (h *HRHandler) UpdateLeave(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Status     *string `json:"status"`
		ApprovedBy *string `json:"approved_by"`
		Notes      *string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	now := time.Now()
	_, err := h.db.Exec(`UPDATE employee_leave SET status=COALESCE($1,status), approved_by=COALESCE($2,approved_by), approved_at=CASE WHEN $1='approved' THEN $3 ELSE approved_at END, notes=COALESCE($4,notes) WHERE id=$5`,
		input.Status, input.ApprovedBy, now, input.Notes, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *HRHandler) DeleteLeave(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM employee_leave WHERE id = $1`, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
