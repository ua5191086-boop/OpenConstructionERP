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

// TrainingHandler handles Training & Certifications module
type TrainingHandler struct {
	db *sql.DB
}

func NewTrainingHandler(db *sql.DB) *TrainingHandler {
	return &TrainingHandler{db: db}
}

func (h *TrainingHandler) RegisterRoutes(r chi.Router) {
	r.Route("/training", func(r chi.Router) {
		r.Get("/courses", h.ListCourses)
		r.Post("/courses", h.CreateCourse)
		r.Get("/courses/{id}", h.GetCourse)
		r.Put("/courses/{id}", h.UpdateCourse)
		r.Delete("/courses/{id}", h.DeleteCourse)

		r.Get("/sessions", h.ListSessions)
		r.Post("/sessions", h.CreateSession)
		r.Get("/sessions/{id}", h.GetSession)
		r.Put("/sessions/{id}", h.UpdateSession)
		r.Delete("/sessions/{id}", h.DeleteSession)

		r.Get("/participants", h.ListParticipants)
		r.Post("/participants", h.CreateParticipant)
		r.Get("/participants/{id}", h.GetParticipant)
		r.Put("/participants/{id}", h.UpdateParticipant)

		r.Get("/certifications", h.ListCertifications)
		r.Post("/certifications", h.CreateCertification)
		r.Get("/certifications/{id}", h.GetCertification)
		r.Put("/certifications/{id}", h.UpdateCertification)
		r.Delete("/certifications/{id}", h.DeleteCertification)

		r.Get("/competencies", h.ListCompetencies)
		r.Post("/competencies", h.CreateCompetency)
		r.Get("/competencies/{id}", h.GetCompetency)
		r.Put("/competencies/{id}", h.UpdateCompetency)
		r.Delete("/competencies/{id}", h.DeleteCompetency)

		r.Get("/summary", h.GetSummary)
	})
}

// =============================================================================
// Courses
// =============================================================================
func (h *TrainingHandler) ListCourses(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT id, project_id, course_code, course_name, course_type, description, provider, duration_hours, duration_days, max_participants, cost_per_person, currency, is_mandatory, validity_days, status, created_at FROM training_courses WHERE 1=1`
	args := []interface{}{}
	argIdx := 1
	if projectID != "" { query += fmt.Sprintf(" AND project_id = $%d", argIdx); argIdx++; args = append(args, projectID) }
	query += " ORDER BY course_code"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, pid, code, name, ctype, desc, provider, currency, status string
		var durHours float64
		var durDays, maxPart, validDays int
		var cost float64
		var mandatory bool
		var createdAt time.Time
		if err := rows.Scan(&id, &pid, &code, &name, &ctype, &desc, &provider, &durHours, &durDays, &maxPart, &cost, &currency, &mandatory, &validDays, &status, &createdAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		items = append(items, map[string]interface{}{
			"id": id, "project_id": pid, "course_code": code, "course_name": name,
			"course_type": ctype, "description": desc, "provider": provider,
			"duration_hours": durHours, "duration_days": durDays, "max_participants": maxPart,
			"cost_per_person": cost, "currency": currency, "is_mandatory": mandatory,
			"validity_days": validDays, "status": status, "created_at": createdAt,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *TrainingHandler) CreateCourse(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID  *string `json:"project_id"`
		CourseCode string  `json:"course_code"`
		CourseName string  `json:"course_name"`
		CourseType *string `json:"course_type"`
		Description *string `json:"description"`
		Provider   *string `json:"provider"`
		DurationHours *float64 `json:"duration_hours"`
		DurationDays *int    `json:"duration_days"`
		MaxParticipants *int `json:"max_participants"`
		CostPerPerson *float64 `json:"cost_per_person"`
		IsMandatory *bool   `json:"is_mandatory"`
		ValidityDays *int   `json:"validity_days"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO training_courses (id, project_id, course_code, course_name, course_type, description, provider, duration_hours, duration_days, max_participants, cost_per_person, is_mandatory, validity_days, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$14)`,
		id, input.ProjectID, input.CourseCode, input.CourseName, input.CourseType, input.Description, input.Provider, input.DurationHours, input.DurationDays, input.MaxParticipants, input.CostPerPerson, input.IsMandatory, input.ValidityDays, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *TrainingHandler) GetCourse(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var code, name string
	err := h.db.QueryRow(`SELECT course_code, course_name FROM training_courses WHERE id = $1`, id).Scan(&code, &name)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "course not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "course_code": code, "course_name": name})
}

func (h *TrainingHandler) UpdateCourse(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		CourseName *string `json:"course_name"`
		Status     *string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	_, err := h.db.Exec(`UPDATE training_courses SET course_name=COALESCE($1,course_name), status=COALESCE($2,status), updated_at=$3 WHERE id=$4`, input.CourseName, input.Status, time.Now(), id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *TrainingHandler) DeleteCourse(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM training_courses WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// Sessions (simplified — full CRUD with participants, dates, scores)
// =============================================================================
func (h *TrainingHandler) ListSessions(w http.ResponseWriter, r *http.Request) {
	courseID := r.URL.Query().Get("course_id")
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT s.id, s.course_id, c.course_name, s.session_code, s.session_date, s.end_date, s.instructor, s.location, s.max_participants, s.actual_participants, s.status, s.completion_rate, s.feedback_score, s.created_at
		FROM training_sessions s JOIN training_courses c ON c.id = s.course_id WHERE 1=1`
	args := []interface{}{}
	argIdx := 1
	if courseID != "" { query += fmt.Sprintf(" AND s.course_id = $%d", argIdx); argIdx++; args = append(args, courseID) }
	if projectID != "" { query += fmt.Sprintf(" AND s.project_id = $%d", argIdx); argIdx++; args = append(args, projectID) }
	query += " ORDER BY s.session_date DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, cid, cname, code, instructor, loc, status string
		var sd, ed time.Time
		var maxP, actP int
		var rate, score float64
		var createdAt time.Time
		if err := rows.Scan(&id, &cid, &cname, &code, &sd, &ed, &instructor, &loc, &maxP, &actP, &status, &rate, &score, &createdAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		items = append(items, map[string]interface{}{
			"id": id, "course_id": cid, "course_name": cname, "session_code": code,
			"session_date": sd, "end_date": ed, "instructor": instructor, "location": loc,
			"max_participants": maxP, "actual_participants": actP, "status": status,
			"completion_rate": rate, "feedback_score": score, "created_at": createdAt,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *TrainingHandler) CreateSession(w http.ResponseWriter, r *http.Request) {
	var input struct {
		CourseID  string  `json:"course_id"`
		ProjectID *string `json:"project_id"`
		SessionCode string `json:"session_code"`
		SessionDate string `json:"session_date"`
		EndDate    *string `json:"end_date"`
		Instructor *string `json:"instructor"`
		Location   *string `json:"location"`
		MaxParticipants *int `json:"max_participants"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO training_sessions (id, course_id, project_id, session_code, session_date, end_date, instructor, location, max_participants, status, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,'planned',$10,$10)`,
		id, input.CourseID, input.ProjectID, input.SessionCode, input.SessionDate, input.EndDate, input.Instructor, input.Location, input.MaxParticipants, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *TrainingHandler) GetSession(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var code, status string
	err := h.db.QueryRow(`SELECT session_code, status FROM training_sessions WHERE id = $1`, id).Scan(&code, &status)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "session not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "session_code": code, "status": status})
}

func (h *TrainingHandler) UpdateSession(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Status       *string  `json:"status"`
		ActualP      *int     `json:"actual_participants"`
		FeedbackScore *float64 `json:"feedback_score"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	_, err := h.db.Exec(`UPDATE training_sessions SET status=COALESCE($1,status), actual_participants=COALESCE($2,actual_participants), feedback_score=COALESCE($3,feedback_score), updated_at=$4 WHERE id=$5`,
		input.Status, input.ActualP, input.FeedbackScore, time.Now(), id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *TrainingHandler) DeleteSession(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM training_sessions WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// Participants (simplified)
// =============================================================================
func (h *TrainingHandler) ListParticipants(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session_id")
	employeeID := r.URL.Query().Get("employee_id")
	query := `SELECT tp.id, tp.session_id, tp.employee_id, e.first_name, e.last_name, tp.attended, tp.hours_attended, tp.score, tp.passed, tp.certificate_number, tp.status, tp.created_at
		FROM training_participants tp JOIN employees e ON e.id = tp.employee_id WHERE 1=1`
	args := []interface{}{}
	argIdx := 1
	if sessionID != "" { query += fmt.Sprintf(" AND tp.session_id = $%d", argIdx); argIdx++; args = append(args, sessionID) }
	if employeeID != "" { query += fmt.Sprintf(" AND tp.employee_id = $%d", argIdx); argIdx++; args = append(args, employeeID) }
	query += " ORDER BY e.last_name"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, sid, eid, fn, ln, certNum, status string
		var attended bool
		var hrs, score float64
		var passed bool
		var createdAt time.Time
		if err := rows.Scan(&id, &sid, &eid, &fn, &ln, &attended, &hrs, &score, &passed, &certNum, &status, &createdAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		items = append(items, map[string]interface{}{
			"id": id, "session_id": sid, "employee_id": eid, "employee_name": fn + " " + ln,
			"attended": attended, "hours_attended": hrs, "score": score, "passed": passed,
			"certificate_number": certNum, "status": status, "created_at": createdAt,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *TrainingHandler) CreateParticipant(w http.ResponseWriter, r *http.Request) {
	var input struct {
		SessionID  string  `json:"session_id"`
		EmployeeID string  `json:"employee_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO training_participants (id, session_id, employee_id, registration_date, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$5)`,
		id, input.SessionID, input.EmployeeID, now, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *TrainingHandler) GetParticipant(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var status string
	err := h.db.QueryRow(`SELECT status FROM training_participants WHERE id = $1`, id).Scan(&status)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "participant not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "status": status})
}

func (h *TrainingHandler) UpdateParticipant(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Status    *string  `json:"status"`
		Attended  *bool    `json:"attended"`
		Score     *float64 `json:"score"`
		Passed    *bool    `json:"passed"`
		HoursAttended *float64 `json:"hours_attended"`
		CertNumber *string `json:"certificate_number"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	_, err := h.db.Exec(`UPDATE training_participants SET status=COALESCE($1,status), attended=COALESCE($2,attended), score=COALESCE($3,score), passed=COALESCE($4,passed), hours_attended=COALESCE($5,hours_attended), certificate_number=COALESCE($6,certificate_number), certificate_issued=CASE WHEN $6 IS NOT NULL THEN NOW() ELSE certificate_issued END, updated_at=$7 WHERE id=$8`,
		input.Status, input.Attended, input.Score, input.Passed, input.HoursAttended, input.CertNumber, time.Now(), id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// =============================================================================
// Certifications
// =============================================================================
func (h *TrainingHandler) ListCertifications(w http.ResponseWriter, r *http.Request) {
	employeeID := r.URL.Query().Get("employee_id")
	query := `SELECT ec.id, ec.employee_id, e.first_name, e.last_name, ec.cert_code, ec.cert_name, ec.cert_type, ec.issuing_body, ec.cert_number, ec.issue_date, ec.expiry_date, ec.status, ec.created_at
		FROM employee_certifications ec JOIN employees e ON e.id = ec.employee_id WHERE 1=1`
	args := []interface{}{}
	argIdx := 1
	if employeeID != "" { query += fmt.Sprintf(" AND ec.employee_id = $%d", argIdx); argIdx++; args = append(args, employeeID) }
	query += " ORDER BY ec.expiry_date NULLS LAST"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, eid, fn, ln, code, name, ctype, body, certNum, status string
		var issueDate, expiryDate, createdAt time.Time
		if err := rows.Scan(&id, &eid, &fn, &ln, &code, &name, &ctype, &body, &certNum, &issueDate, &expiryDate, &status, &createdAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		items = append(items, map[string]interface{}{
			"id": id, "employee_id": eid, "employee_name": fn + " " + ln,
			"cert_code": code, "cert_name": name, "cert_type": ctype,
			"issuing_body": body, "cert_number": certNum, "issue_date": issueDate,
			"expiry_date": expiryDate, "status": status, "created_at": createdAt,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *TrainingHandler) CreateCertification(w http.ResponseWriter, r *http.Request) {
	var input struct {
		EmployeeID  string  `json:"employee_id"`
		CertCode    string  `json:"cert_code"`
		CertName    string  `json:"cert_name"`
		CertType    *string `json:"cert_type"`
		IssuingBody *string `json:"issuing_body"`
		CertNumber  *string `json:"cert_number"`
		IssueDate   string  `json:"issue_date"`
		ExpiryDate  *string `json:"expiry_date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO employee_certifications (id, employee_id, cert_code, cert_name, cert_type, issuing_body, cert_number, issue_date, expiry_date, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$10)`,
		id, input.EmployeeID, input.CertCode, input.CertName, input.CertType, input.IssuingBody, input.CertNumber, input.IssueDate, input.ExpiryDate, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *TrainingHandler) GetCertification(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var code, name string
	err := h.db.QueryRow(`SELECT cert_code, cert_name FROM employee_certifications WHERE id = $1`, id).Scan(&code, &name)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "certification not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "cert_code": code, "cert_name": name})
}

func (h *TrainingHandler) UpdateCertification(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Status     *string `json:"status"`
		ExpiryDate *string `json:"expiry_date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	_, err := h.db.Exec(`UPDATE employee_certifications SET status=COALESCE($1,status), expiry_date=COALESCE($2,expiry_date), updated_at=$3 WHERE id=$4`, input.Status, input.ExpiryDate, time.Now(), id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *TrainingHandler) DeleteCertification(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM employee_certifications WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// Competencies (simplified CRUD)
// =============================================================================
func (h *TrainingHandler) ListCompetencies(w http.ResponseWriter, r *http.Request) {
	employeeID := r.URL.Query().Get("employee_id")
	query := `SELECT c.id, c.employee_id, e.first_name, e.last_name, c.competency_code, c.competency_name, c.category, c.proficiency_level, c.years_experience, c.last_assessed, c.notes, c.created_at
		FROM employee_competencies c JOIN employees e ON e.id = c.employee_id WHERE 1=1`
	args := []interface{}{}
	argIdx := 1
	if employeeID != "" { query += fmt.Sprintf(" AND c.employee_id = $%d", argIdx); argIdx++; args = append(args, employeeID) }
	query += " ORDER BY c.category, c.competency_code"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, eid, fn, ln, code, name, cat, level, notes string
		var years float64
		var assessed, createdAt time.Time
		if err := rows.Scan(&id, &eid, &fn, &ln, &code, &name, &cat, &level, &years, &assessed, &notes, &createdAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		items = append(items, map[string]interface{}{
			"id": id, "employee_id": eid, "employee_name": fn + " " + ln,
			"competency_code": code, "competency_name": name, "category": cat,
			"proficiency_level": level, "years_experience": years,
			"last_assessed": assessed, "notes": notes, "created_at": createdAt,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *TrainingHandler) CreateCompetency(w http.ResponseWriter, r *http.Request) {
	var input struct {
		EmployeeID string `json:"employee_id"`
		CompCode   string `json:"competency_code"`
		CompName   string `json:"competency_name"`
		Category   *string `json:"category"`
		Level      *string `json:"proficiency_level"`
		YearsExp   *float64 `json:"years_experience"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO employee_competencies (id, employee_id, competency_code, competency_name, category, proficiency_level, years_experience, last_assessed, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$9)`,
		id, input.EmployeeID, input.CompCode, input.CompName, input.Category, input.Level, input.YearsExp, now, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *TrainingHandler) GetCompetency(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var code, name string
	err := h.db.QueryRow(`SELECT competency_code, competency_name FROM employee_competencies WHERE id = $1`, id).Scan(&code, &name)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "competency not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "competency_code": code, "competency_name": name})
}

func (h *TrainingHandler) UpdateCompetency(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		ProficiencyLevel *string `json:"proficiency_level"`
		YearsExp         *float64 `json:"years_experience"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	_, err := h.db.Exec(`UPDATE employee_competencies SET proficiency_level=COALESCE($1,proficiency_level), years_experience=COALESCE($2,years_experience), last_assessed=NOW(), updated_at=$3 WHERE id=$4`, input.ProficiencyLevel, input.YearsExp, time.Now(), id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *TrainingHandler) DeleteCompetency(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM employee_competencies WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// Summary
// =============================================================================
func (h *TrainingHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT project_id, total_courses, total_sessions, completed_sessions, total_participants, passed, avg_score, expired_certs, total_competencies FROM training_summary`
	args := []interface{}{}
	argIdx := 1
	if projectID != "" { query += fmt.Sprintf(" WHERE project_id = $%d", argIdx); args = append(args, projectID) }

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var pid string
		var courses, sessions, completed, participants, passed, expiredCerts, competencies int
		var avgScore float64
		if err := rows.Scan(&pid, &courses, &sessions, &completed, &participants, &passed, &avgScore, &expiredCerts, &competencies); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		items = append(items, map[string]interface{}{
			"project_id": pid, "total_courses": courses, "total_sessions": sessions,
			"completed_sessions": completed, "total_participants": participants,
			"passed": passed, "avg_score": avgScore, "expired_certs": expiredCerts,
			"total_competencies": competencies,
		})
	}
	respondJSON(w, http.StatusOK, items)
}