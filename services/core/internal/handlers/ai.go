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

// AIHandler handles AI module endpoints
type AIHandler struct {
	db *sql.DB
}

func NewAIHandler(db *sql.DB) *AIHandler {
	return &AIHandler{db: db}
}

func (h *AIHandler) RegisterRoutes(r chi.Router) {
	r.Route("/ai", func(r chi.Router) {
		// Agents
		r.Get("/agents", h.ListAgents)
		r.Post("/agents", h.CreateAgent)
		r.Get("/agents/{id}", h.GetAgent)
		r.Put("/agents/{id}", h.UpdateAgent)
		r.Delete("/agents/{id}", h.DeleteAgent)

		// Tasks
		r.Get("/tasks", h.ListTasks)
		r.Post("/tasks", h.CreateTask)
		r.Get("/tasks/{id}", h.GetTask)
		r.Put("/tasks/{id}", h.UpdateTask)
		r.Delete("/tasks/{id}", h.DeleteTask)

		// Conversations
		r.Get("/conversations", h.ListConversations)
		r.Post("/conversations", h.CreateConversation)
		r.Get("/conversations/{id}", h.GetConversation)
	})
}

// --- AI Agents ---

func (h *AIHandler) ListAgents(w http.ResponseWriter, r *http.Request) {
	agentType := r.URL.Query().Get("agent_type")
	isActive := r.URL.Query().Get("is_active")

	query := `SELECT id, agent_name, agent_type, description, model_name, model_provider, system_prompt, temperature, max_tokens, is_active, version, config, created_at, updated_at FROM ai_agents WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if agentType != "" {
		query += ` AND agent_type = $` + itoa(argIdx)
		args = append(args, agentType)
		argIdx++
	}
	if isActive == "true" {
		query += ` AND is_active = true`
	}
	query += ` ORDER BY agent_name`

	rows, err := h.db.Query(query, args...)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	agents := make([]models.AIAgent, 0)
	for rows.Next() {
		var a models.AIAgent
		if err := rows.Scan(&a.ID, &a.AgentName, &a.AgentType, &a.Description, &a.ModelName, &a.ModelProvider, &a.SystemPrompt, &a.Temperature, &a.MaxTokens, &a.IsActive, &a.Version, &a.Config, &a.CreatedAt, &a.UpdatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		agents = append(agents, a)
	}
	respondJSON(w, http.StatusOK, agents)
}

func (h *AIHandler) CreateAgent(w http.ResponseWriter, r *http.Request) {
	var input struct {
		AgentName     string   `json:"agent_name"`
		AgentType     string   `json:"agent_type"`
		Description   *string  `json:"description"`
		ModelName     *string  `json:"model_name"`
		ModelProvider *string  `json:"model_provider"`
		SystemPrompt  *string  `json:"system_prompt"`
		Temperature   *float64 `json:"temperature"`
		MaxTokens     *int     `json:"max_tokens"`
		IsActive      bool     `json:"is_active"`
		Version       string   `json:"version"`
		Config        *string  `json:"config"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO ai_agents (id, agent_name, agent_type, description, model_name, model_provider, system_prompt, temperature, max_tokens, is_active, version, config, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		id, input.AgentName, input.AgentType, input.Description, input.ModelName, input.ModelProvider, input.SystemPrompt, input.Temperature, input.MaxTokens, input.IsActive, input.Version, input.Config, now, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *AIHandler) GetAgent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var a models.AIAgent
	err := h.db.QueryRow(`SELECT id, agent_name, agent_type, description, model_name, model_provider, system_prompt, temperature, max_tokens, is_active, version, config, created_at, updated_at FROM ai_agents WHERE id = $1`, id).
		Scan(&a.ID, &a.AgentName, &a.AgentType, &a.Description, &a.ModelName, &a.ModelProvider, &a.SystemPrompt, &a.Temperature, &a.MaxTokens, &a.IsActive, &a.Version, &a.Config, &a.CreatedAt, &a.UpdatedAt)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "AI agent not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, a)
}

func (h *AIHandler) UpdateAgent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Description   *string  `json:"description"`
		SystemPrompt  *string  `json:"system_prompt"`
		Temperature   *float64 `json:"temperature"`
		MaxTokens     *int     `json:"max_tokens"`
		IsActive      *bool    `json:"is_active"`
		Config        *string  `json:"config"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	now := time.Now()
	_, err := h.db.Exec(`UPDATE ai_agents SET description=COALESCE($1,description), system_prompt=COALESCE($2,system_prompt), temperature=COALESCE($3,temperature), max_tokens=COALESCE($4,max_tokens), is_active=COALESCE($5,is_active), config=COALESCE($6,config), updated_at=$7 WHERE id=$8`,
		input.Description, input.SystemPrompt, input.Temperature, input.MaxTokens, input.IsActive, input.Config, now, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *AIHandler) DeleteAgent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM ai_agents WHERE id = $1`, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// --- AI Tasks ---

func (h *AIHandler) ListTasks(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	taskType := r.URL.Query().Get("task_type")
	sourceType := r.URL.Query().Get("source_type")

	query := `SELECT id, agent_id, task_type, input_data, input_format, output_data, output_format, confidence, tokens_used, cost, processing_time_ms, status, error_message, source_type, source_id, created_by, created_at, completed_at FROM ai_tasks WHERE 1=1`
	args := []interface{}{}
	argIdx := 1

	if status != "" {
		query += ` AND status = $` + itoa(argIdx)
		args = append(args, status)
		argIdx++
	}
	if taskType != "" {
		query += ` AND task_type = $` + itoa(argIdx)
		args = append(args, taskType)
		argIdx++
	}
	if sourceType != "" {
		query += ` AND source_type = $` + itoa(argIdx)
		args = append(args, sourceType)
		argIdx++
	}
	query += ` ORDER BY created_at DESC`

	rows, err := h.db.Query(query, args...)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	tasks := make([]models.AITask, 0)
	for rows.Next() {
		var t models.AITask
		if err := rows.Scan(&t.ID, &t.AgentID, &t.TaskType, &t.InputData, &t.InputFormat, &t.OutputData, &t.OutputFormat, &t.Confidence, &t.TokensUsed, &t.Cost, &t.ProcessingTimeMs, &t.Status, &t.ErrorMessage, &t.SourceType, &t.SourceID, &t.CreatedBy, &t.CreatedAt, &t.CompletedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		tasks = append(tasks, t)
	}
	respondJSON(w, http.StatusOK, tasks)
}

func (h *AIHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var input struct {
		AgentID     *string `json:"agent_id"`
		TaskType    string  `json:"task_type"`
		InputData   string  `json:"input_data"`
		InputFormat string  `json:"input_format"`
		OutputFormat string `json:"output_format"`
		Status      string  `json:"status"`
		SourceType  *string `json:"source_type"`
		SourceID    *string `json:"source_id"`
		CreatedBy   *string `json:"created_by"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO ai_tasks (id, agent_id, task_type, input_data, input_format, output_format, status, source_type, source_id, created_by, created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		id, input.AgentID, input.TaskType, input.InputData, input.InputFormat, input.OutputFormat, input.Status, input.SourceType, input.SourceID, input.CreatedBy, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *AIHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var t models.AITask
	err := h.db.QueryRow(`SELECT id, agent_id, task_type, input_data, input_format, output_data, output_format, confidence, tokens_used, cost, processing_time_ms, status, error_message, source_type, source_id, created_by, created_at, completed_at FROM ai_tasks WHERE id = $1`, id).
		Scan(&t.ID, &t.AgentID, &t.TaskType, &t.InputData, &t.InputFormat, &t.OutputData, &t.OutputFormat, &t.Confidence, &t.TokensUsed, &t.Cost, &t.ProcessingTimeMs, &t.Status, &t.ErrorMessage, &t.SourceType, &t.SourceID, &t.CreatedBy, &t.CreatedAt, &t.CompletedAt)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "AI task not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, t)
}

func (h *AIHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Status          *string  `json:"status"`
		OutputData      *string  `json:"output_data"`
		Confidence      *float64 `json:"confidence"`
		TokensUsed      *int     `json:"tokens_used"`
		Cost            *float64 `json:"cost"`
		ProcessingTimeMs *int    `json:"processing_time_ms"`
		ErrorMessage    *string  `json:"error_message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	now := time.Now()
	_, err := h.db.Exec(`UPDATE ai_tasks SET status=COALESCE($1,status), output_data=COALESCE($2,output_data), confidence=COALESCE($3,confidence), tokens_used=COALESCE($4,tokens_used), cost=COALESCE($5,cost), processing_time_ms=COALESCE($6,processing_time_ms), error_message=COALESCE($7,error_message), completed_at=CASE WHEN $1 IN ('completed','failed') THEN $8 ELSE completed_at END WHERE id=$9`,
		input.Status, input.OutputData, input.Confidence, input.TokensUsed, input.Cost, input.ProcessingTimeMs, input.ErrorMessage, now, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *AIHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM ai_tasks WHERE id = $1`, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// --- AI Conversations ---

func (h *AIHandler) ListConversations(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session_id")
	query := `SELECT id, session_id, user_message, assistant_message, intent, entities, module_used, action_taken, tokens_used, processing_time_ms, feedback_score, feedback_text, created_at FROM ai_conversations`
	args := []interface{}{}
	argIdx := 1

	if sessionID != "" {
		query += ` WHERE session_id = $` + itoa(argIdx)
		args = append(args, sessionID)
		argIdx++
	}
	query += ` ORDER BY created_at DESC`

	rows, err := h.db.Query(query, args...)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	convs := make([]models.AIConversation, 0)
	for rows.Next() {
		var c models.AIConversation
		if err := rows.Scan(&c.ID, &c.SessionID, &c.UserMessage, &c.AssistantMessage, &c.Intent, &c.Entities, &c.ModuleUsed, &c.ActionTaken, &c.TokensUsed, &c.ProcessingTimeMs, &c.FeedbackScore, &c.FeedbackText, &c.CreatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
		convs = append(convs, c)
	}
	respondJSON(w, http.StatusOK, convs)
}

func (h *AIHandler) CreateConversation(w http.ResponseWriter, r *http.Request) {
	var input struct {
		SessionID        string  `json:"session_id"`
		UserMessage      string  `json:"user_message"`
		AssistantMessage string  `json:"assistant_message"`
		Intent           *string `json:"intent"`
		Entities         *string `json:"entities"`
		ModuleUsed       *string `json:"module_used"`
		ActionTaken      *string `json:"action_taken"`
		TokensUsed       *int    `json:"tokens_used"`
		ProcessingTimeMs *int    `json:"processing_time_ms"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO ai_conversations (id, session_id, user_message, assistant_message, intent, entities, module_used, action_taken, tokens_used, processing_time_ms, created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`,
		id, input.SessionID, input.UserMessage, input.AssistantMessage, input.Intent, input.Entities, input.ModuleUsed, input.ActionTaken, input.TokensUsed, input.ProcessingTimeMs, now)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *AIHandler) GetConversation(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var c models.AIConversation
	err := h.db.QueryRow(`SELECT id, session_id, user_message, assistant_message, intent, entities, module_used, action_taken, tokens_used, processing_time_ms, feedback_score, feedback_text, created_at FROM ai_conversations WHERE id = $1`, id).
		Scan(&c.ID, &c.SessionID, &c.UserMessage, &c.AssistantMessage, &c.Intent, &c.Entities, &c.ModuleUsed, &c.ActionTaken, &c.TokensUsed, &c.ProcessingTimeMs, &c.FeedbackScore, &c.FeedbackText, &c.CreatedAt)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "conversation not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, c)
}
