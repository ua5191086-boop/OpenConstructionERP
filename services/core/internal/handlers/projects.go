package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// ProjectHandler handles project CRUD with real DB schema
type ProjectHandler struct {
	db *sql.DB
}

func NewProjectHandler(db *sql.DB) *ProjectHandler {
	return &ProjectHandler{db: db}
}

func (h *ProjectHandler) RegisterRoutes(r chi.Router) {
	r.Route("/projects", func(r chi.Router) {
		r.Get("/", h.ListProjects)
		r.Post("/", h.CreateProject)
		r.Get("/{id}", h.GetProject)
		r.Put("/{id}", h.UpdateProject)
		r.Delete("/{id}", h.DeleteProject)
	})
}

type projectResponse struct {
	ID            string     `json:"id"`
	Code          string     `json:"code"`
	Name          string     `json:"name"`
	NameRu        *string    `json:"name_ru,omitempty"`
	ProjectType   string     `json:"project_type"`
	Status        string     `json:"status"`
	Country       *string    `json:"country,omitempty"`
	Currency      *string    `json:"currency,omitempty"`
	StartDate     *string    `json:"start_date,omitempty"`
	FinishDate    *string    `json:"finish_date,omitempty"`
	ContractValue *float64   `json:"contract_value,omitempty"`
	Priority      *string    `json:"priority,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

func (h *ProjectHandler) ListProjects(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")

	query := `SELECT id, code, name, name_ru, project_type, status, country, currency,
		start_date, finish_date, contract_value, priority, created_at, updated_at
		FROM projects WHERE 1=1`
	var args []interface{}
	argIdx := 1

	if status != "" {
		query += ` AND status = $` + string(rune('0'+argIdx))
		args = append(args, status)
		argIdx++
	}
	query += " ORDER BY code"

	rows, err := h.db.Query(query, args...)
	if err != nil {
		log.Printf("ListProjects query error: %v", err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	projects := make([]projectResponse, 0)
	for rows.Next() {
		var p projectResponse
		var nameRu, country, currency, startDate, finishDate, priority sql.NullString
		var contractValue sql.NullFloat64

		err := rows.Scan(&p.ID, &p.Code, &p.Name, &nameRu, &p.ProjectType, &p.Status,
			&country, &currency, &startDate, &finishDate, &contractValue, &priority,
			&p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			log.Printf("ListProjects scan error: %v", err)
			continue
		}
		if nameRu.Valid { p.NameRu = &nameRu.String }
		if country.Valid { p.Country = &country.String }
		if currency.Valid { p.Currency = &currency.String }
		if startDate.Valid { p.StartDate = &startDate.String }
		if finishDate.Valid { p.FinishDate = &finishDate.String }
		if contractValue.Valid { p.ContractValue = &contractValue.Float64 }
		if priority.Valid { p.Priority = &priority.String }
		projects = append(projects, p)
	}

	respondJSON(w, http.StatusOK, projects)
}

func (h *ProjectHandler) GetProject(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		respondError(w, http.StatusBadRequest, "invalid UUID")
		return
	}

	var p projectResponse
	var nameRu, country, currency, startDate, finishDate, priority sql.NullString
	var contractValue sql.NullFloat64

	err := h.db.QueryRow(`SELECT id, code, name, name_ru, project_type, status, country, currency,
		start_date, finish_date, contract_value, priority, created_at, updated_at
		FROM projects WHERE id = $1`, id).Scan(
		&p.ID, &p.Code, &p.Name, &nameRu, &p.ProjectType, &p.Status,
		&country, &currency, &startDate, &finishDate, &contractValue, &priority,
		&p.CreatedAt, &p.UpdatedAt)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "project not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if nameRu.Valid { p.NameRu = &nameRu.String }
	if country.Valid { p.Country = &country.String }
	if currency.Valid { p.Currency = &currency.String }
	if startDate.Valid { p.StartDate = &startDate.String }
	if finishDate.Valid { p.FinishDate = &finishDate.String }
	if contractValue.Valid { p.ContractValue = &contractValue.Float64 }
	if priority.Valid { p.Priority = &priority.String }

	respondJSON(w, http.StatusOK, p)
}

type projectInput struct {
	Code          string   `json:"code"`
	Name          string   `json:"name"`
	NameRu        *string  `json:"name_ru,omitempty"`
	ProjectType   string   `json:"project_type"`
	Status        *string  `json:"status,omitempty"`
	Country       *string  `json:"country,omitempty"`
	Currency      *string  `json:"currency,omitempty"`
	StartDate     *string  `json:"start_date,omitempty"`
	FinishDate    *string  `json:"finish_date,omitempty"`
	ContractValue *float64 `json:"contract_value,omitempty"`
	Priority      *string  `json:"priority,omitempty"`
}

func (h *ProjectHandler) CreateProject(w http.ResponseWriter, r *http.Request) {
	var input projectInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if input.Code == "" || input.Name == "" || input.ProjectType == "" {
		respondError(w, http.StatusBadRequest, "code, name, project_type are required")
		return
	}

	status := "tender"
	if input.Status != nil { status = *input.Status }

	var id string
	err := h.db.QueryRow(`INSERT INTO projects (code, name, name_ru, project_type, status, country, currency, start_date, finish_date, contract_value, priority)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8::date,$9::date,$10,$11) RETURNING id`,
		input.Code, input.Name, input.NameRu, input.ProjectType, status,
		input.Country, input.Currency, input.StartDate, input.FinishDate,
		input.ContractValue, input.Priority).Scan(&id)
	if err != nil {
		log.Printf("CreateProject error: %v", err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Fetch created project
	r.URL.Path = "/api/v1/projects/" + id
	h.GetProject(w, r)
}

func (h *ProjectHandler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		respondError(w, http.StatusBadRequest, "invalid UUID")
		return
	}

	var input projectInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	_, err := h.db.Exec(`UPDATE projects SET code=$1, name=$2, name_ru=$3, project_type=$4, status=$5,
		country=$6, currency=$7, start_date=$8::date, finish_date=$9::date, contract_value=$10, priority=$11
		WHERE id=$12`,
		input.Code, input.Name, input.NameRu, input.ProjectType, input.Status,
		input.Country, input.Currency, input.StartDate, input.FinishDate,
		input.ContractValue, input.Priority, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.GetProject(w, r)
}

func (h *ProjectHandler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		respondError(w, http.StatusBadRequest, "invalid UUID")
		return
	}

	_, err := h.db.Exec(`DELETE FROM projects WHERE id = $1`, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
