package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
)

// AIHandler — HTTP handler for AI assistant framework
type AIHandler struct {
	db *sqlx.DB
}

func NewAIHandler(db *sqlx.DB) *AIHandler {
	return &AIHandler{db: db}
}

func (h *AIHandler) RegisterRoutes(r chi.Router) {
	r.Route("/ai/providers", func(r chi.Router) {
		r.Get("/", h.ListProviders)
		r.Post("/", h.CreateProvider)
		r.Get("/{id}", h.GetProvider)
		r.Put("/{id}", h.UpdateProvider)
		r.Delete("/{id}", h.DeleteProvider)
		r.Get("/default", h.GetDefaultProvider)
	})
	r.Route("/ai/knowledge-base", func(r chi.Router) {
		r.Get("/", h.ListKB)
		r.Post("/", h.CreateKB)
		r.Get("/{id}", h.GetKB)
		r.Put("/{id}", h.UpdateKB)
		r.Delete("/{id}", h.DeleteKB)
		r.Get("/project/{projectId}", h.ListKBByProject)
		r.Get("/type/{kbType}", h.ListKBByType)
		r.Post("/{id}/embed", h.TriggerEmbedding)
	})
	r.Route("/ai/embeddings", func(r chi.Router) {
		r.Get("/", h.ListEmbeddings)
		r.Post("/search", h.SearchEmbeddings)
		r.Get("/kb/{kbId}", h.ListEmbeddingsByKB)
	})
	r.Route("/ai/sessions", func(r chi.Router) {
		r.Get("/", h.ListSessions)
		r.Post("/", h.CreateSession)
		r.Get("/{id}", h.GetSession)
		r.Put("/{id}", h.UpdateSession)
		r.Delete("/{id}", h.DeleteSession)
		r.Get("/project/{projectId}", h.ListSessionsByProject)
		r.Get("/pinned", h.ListPinnedSessions)
	})
	r.Route("/ai/messages", func(r chi.Router) {
		r.Get("/session/{sessionId}", h.ListMessages)
		r.Post("/", h.CreateMessage)
		r.Post("/session/{sessionId}/chat", h.Chat)
	})
	r.Route("/ai/functions", func(r chi.Router) {
		r.Get("/", h.ListFunctions)
		r.Post("/", h.CreateFunction)
		r.Get("/{id}", h.GetFunction)
		r.Put("/{id}", h.UpdateFunction)
		r.Delete("/{id}", h.DeleteFunction)
		r.Get("/category/{category}", h.ListFunctionsByCategory)
	})
	r.Route("/ai/classifications", func(r chi.Router) {
		r.Get("/", h.ListClassifications)
	})
	r.Route("/ai/predictions", func(r chi.Router) {
		r.Get("/", h.ListPredictions)
	})
	r.Route("/ai/cost-estimates", func(r chi.Router) {
		r.Get("/", h.ListCostEstimates)
	})
}

// === Providers ===

func (h *AIHandler) ListProviders(w http.ResponseWriter, r *http.Request) {
	var items []map[string]interface{}
	h.db.Select(&items, "SELECT * FROM ai_providers ORDER BY name")
	respondJSON(w, http.StatusOK, items)
}

func (h *AIHandler) CreateProvider(w http.ResponseWriter, r *http.Request) {
	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}
	_, err := h.db.NamedExec(`INSERT INTO ai_providers
		(name, provider_type, model, base_url, api_key_encrypted, max_tokens, temperature,
		 embedding_model, embedding_dimensions, cost_per_1k_input, cost_per_1k_output,
		 is_active, is_default, notes)
		VALUES (:name, :provider_type, :model, :base_url, :api_key_encrypted, :max_tokens, :temperature,
		 :embedding_model, :embedding_dimensions, :cost_per_1k_input, :cost_per_1k_output,
		 :is_active, :is_default, :notes)`, input)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("insert failed: %v", err))
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(input)
}

func (h *AIHandler) GetProvider(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var item map[string]interface{}
	if err := h.db.Get(&item, "SELECT * FROM ai_providers WHERE id=$1", id); err != nil {
		respondError(w, http.StatusNotFound, "provider not found")
		return
	}
	respondJSON(w, http.StatusOK, item)
}

func (h *AIHandler) UpdateProvider(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}
	input["id"] = id
	h.db.NamedExec(`UPDATE ai_providers SET
		name=:name, provider_type=:provider_type, model=:model, base_url=:base_url,
		api_key_encrypted=:api_key_encrypted, max_tokens=:max_tokens, temperature=:temperature,
		embedding_model=:embedding_model, embedding_dimensions=:embedding_dimensions,
		cost_per_1k_input=:cost_per_1k_input, cost_per_1k_output=:cost_per_1k_output,
		is_active=:is_active, is_default=:is_default, notes=:notes, updated_at=NOW()
		WHERE id=:id`, input)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(input)
}

func (h *AIHandler) DeleteProvider(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	h.db.Exec("DELETE FROM ai_providers WHERE id=$1", id)
	w.WriteHeader(http.StatusNoContent)
}

func (h *AIHandler) GetDefaultProvider(w http.ResponseWriter, r *http.Request) {
	var item map[string]interface{}
	if err := h.db.Get(&item, "SELECT * FROM ai_providers WHERE is_default=TRUE AND is_active=TRUE LIMIT 1"); err != nil {
		respondError(w, http.StatusNotFound, "no default provider configured")
		return
	}
	respondJSON(w, http.StatusOK, item)
}

// === Knowledge Base ===

func (h *AIHandler) ListKB(w http.ResponseWriter, r *http.Request) {
	var items []map[string]interface{}
	h.db.Select(&items, "SELECT * FROM ai_knowledge_base WHERE is_active=TRUE ORDER BY created_at DESC")
	respondJSON(w, http.StatusOK, items)
}

func (h *AIHandler) CreateKB(w http.ResponseWriter, r *http.Request) {
	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}
	_, err := h.db.NamedExec(`INSERT INTO ai_knowledge_base
		(project_id, kb_type, title, source, source_url, original_filename, content_type,
		 content_text, embedding_provider, tags, metadata, uploaded_by, notes)
		VALUES (:project_id, :kb_type, :title, :source, :source_url, :original_filename, :content_type,
		 :content_text, :embedding_provider, :tags, :metadata, :uploaded_by, :notes)`, input)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("insert failed: %v", err))
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(input)
}

func (h *AIHandler) GetKB(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var item map[string]interface{}
	if err := h.db.Get(&item, "SELECT * FROM ai_knowledge_base WHERE id=$1", id); err != nil {
		respondError(w, http.StatusNotFound, "kb entry not found")
		return
	}
	respondJSON(w, http.StatusOK, item)
}

func (h *AIHandler) UpdateKB(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}
	input["id"] = id
	h.db.NamedExec(`UPDATE ai_knowledge_base SET
		kb_type=:kb_type, title=:title, content_text=:content_text,
		tags=:tags, metadata=:metadata, notes=:notes, updated_at=NOW()
		WHERE id=:id`, input)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(input)
}

func (h *AIHandler) DeleteKB(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	h.db.Exec("UPDATE ai_knowledge_base SET is_active=FALSE WHERE id=$1", id)
	w.WriteHeader(http.StatusNoContent)
}

func (h *AIHandler) ListKBByProject(w http.ResponseWriter, r *http.Request) {
	pid := chi.URLParam(r, "projectId")
	var items []map[string]interface{}
	h.db.Select(&items, "SELECT * FROM ai_knowledge_base WHERE project_id=$1 AND is_active=TRUE ORDER BY created_at DESC", pid)
	respondJSON(w, http.StatusOK, items)
}

func (h *AIHandler) ListKBByType(w http.ResponseWriter, r *http.Request) {
	kt := chi.URLParam(r, "kbType")
	var items []map[string]interface{}
	h.db.Select(&items, "SELECT * FROM ai_knowledge_base WHERE kb_type=$1 AND is_active=TRUE ORDER BY title", kt)
	respondJSON(w, http.StatusOK, items)
}

func (h *AIHandler) TriggerEmbedding(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	h.db.Exec(`UPDATE ai_knowledge_base SET embedding_status='processing' WHERE id=$1`, id)
	respondJSON(w, http.StatusOK, map[string]string{"status": "embedding_triggered", "kb_id": id})
}

// === Embeddings ===

func (h *AIHandler) ListEmbeddings(w http.ResponseWriter, r *http.Request) {
	var items []map[string]interface{}
	h.db.Select(&items, "SELECT id, kb_id, chunk_index, chunk_text, chunk_tokens, embedding_model, created_at FROM ai_embeddings ORDER BY kb_id, chunk_index LIMIT 200")
	respondJSON(w, http.StatusOK, items)
}

func (h *AIHandler) SearchEmbeddings(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Query      string   `json:"query"`
		KBIds      []string `json:"kb_ids"`
		Limit      int      `json:"limit"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}
	if input.Limit <= 0 {
		input.Limit = 10
	}

	// Simple text search fallback (no pgvector yet)
	var items []map[string]interface{}
	like := "%" + input.Query + "%"

	if len(input.KBIds) > 0 {
		query, args, _ := sqlx.In(`SELECT e.id, e.kb_id, e.chunk_index, e.chunk_text, e.embedding_model,
			k.title as kb_title, k.kb_type
			FROM ai_embeddings e
			JOIN ai_knowledge_base k ON k.id=e.kb_id
			WHERE e.kb_id IN (?) AND e.chunk_text ILIKE ?
			ORDER BY e.chunk_index
			LIMIT ?`, input.KBIds, like, input.Limit)
		h.db.Select(&items, query, args...)
	} else {
		h.db.Select(&items, `SELECT e.id, e.kb_id, e.chunk_index, e.chunk_text, e.embedding_model,
			k.title as kb_title, k.kb_type
			FROM ai_embeddings e
			JOIN ai_knowledge_base k ON k.id=e.kb_id
			WHERE e.chunk_text ILIKE $1
			ORDER BY e.chunk_index LIMIT $2`, like, input.Limit)
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"query":   input.Query,
		"results": items,
		"total":   len(items),
	})
}

func (h *AIHandler) ListEmbeddingsByKB(w http.ResponseWriter, r *http.Request) {
	kbID := chi.URLParam(r, "kbId")
	var items []map[string]interface{}
	h.db.Select(&items, "SELECT id, kb_id, chunk_index, chunk_text, chunk_tokens, embedding_model, created_at FROM ai_embeddings WHERE kb_id=$1 ORDER BY chunk_index", kbID)
	respondJSON(w, http.StatusOK, items)
}

// === Sessions ===

func (h *AIHandler) ListSessions(w http.ResponseWriter, r *http.Request) {
	var items []map[string]interface{}
	err := h.db.Select(&items, "SELECT * FROM ai_sessions ORDER BY updated_at DESC LIMIT 50")
	if err != nil {
		items = []map[string]interface{}{}
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *AIHandler) CreateSession(w http.ResponseWriter, r *http.Request) {
	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}
	_, err := h.db.NamedExec(`INSERT INTO ai_sessions
		(project_id, session_type, title, provider_id, system_prompt, temperature, max_tokens, context_kb_ids, user_id)
		VALUES (:project_id, :session_type, :title, :provider_id, :system_prompt, :temperature, :max_tokens, :context_kb_ids, :user_id)`, input)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("insert failed: %v", err))
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(input)
}

func (h *AIHandler) GetSession(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var item map[string]interface{}
	if err := h.db.Get(&item, "SELECT * FROM ai_sessions WHERE id=$1", id); err != nil {
		respondError(w, http.StatusNotFound, "session not found")
		return
	}
	respondJSON(w, http.StatusOK, item)
}

func (h *AIHandler) UpdateSession(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}
	input["id"] = id
	h.db.NamedExec(`UPDATE ai_sessions SET
		title=:title, system_prompt=:system_prompt, temperature=:temperature,
		max_tokens=:max_tokens, context_kb_ids=:context_kb_ids, is_pinned=:is_pinned,
		notes=:notes, updated_at=NOW() WHERE id=:id`, input)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(input)
}

func (h *AIHandler) DeleteSession(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	h.db.Exec("DELETE FROM ai_sessions WHERE id=$1", id)
	w.WriteHeader(http.StatusNoContent)
}

func (h *AIHandler) ListSessionsByProject(w http.ResponseWriter, r *http.Request) {
	pid := chi.URLParam(r, "projectId")
	var items []map[string]interface{}
	h.db.Select(&items, "SELECT * FROM ai_sessions WHERE project_id=$1 ORDER BY updated_at DESC", pid)
	respondJSON(w, http.StatusOK, items)
}

func (h *AIHandler) ListPinnedSessions(w http.ResponseWriter, r *http.Request) {
	var items []map[string]interface{}
	h.db.Select(&items, "SELECT * FROM ai_sessions WHERE is_pinned=TRUE ORDER BY updated_at DESC")
	respondJSON(w, http.StatusOK, items)
}

// === Messages & Chat ===

func (h *AIHandler) ListMessages(w http.ResponseWriter, r *http.Request) {
	sid := chi.URLParam(r, "sessionId")
	var items []map[string]interface{}
	h.db.Select(&items, "SELECT * FROM ai_messages WHERE session_id=$1 ORDER BY created_at ASC", sid)
	respondJSON(w, http.StatusOK, items)
}

func (h *AIHandler) CreateMessage(w http.ResponseWriter, r *http.Request) {
	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}
	_, err := h.db.NamedExec(`INSERT INTO ai_messages
		(session_id, role, content, tool_calls, tool_results, tokens_input, tokens_output, cost, latency_ms, metadata)
		VALUES (:session_id, :role, :content, :tool_calls, :tool_results, :tokens_input, :tokens_output, :cost, :latency_ms, :metadata)`, input)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("insert failed: %v", err))
		return
	}
	// Update session counters
	h.db.Exec(`UPDATE ai_sessions SET
		message_count = message_count + 1,
		total_tokens_input = total_tokens_input + COALESCE($1, 0),
		total_tokens_output = total_tokens_output + COALESCE($2, 0),
		total_cost = total_cost + COALESCE($3, 0),
		updated_at = NOW()
		WHERE id = $4`, input["tokens_input"], input["tokens_output"], input["cost"], input["session_id"])
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(input)
}

func (h *AIHandler) Chat(w http.ResponseWriter, r *http.Request) {
	sid := chi.URLParam(r, "sessionId")
	var input struct {
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}

	// Save user message
	h.db.Exec(`INSERT INTO ai_messages (session_id, role, content) VALUES ($1, 'user', $2)`, sid, input.Message)
	h.db.Exec(`UPDATE ai_sessions SET message_count=message_count+1, updated_at=NOW() WHERE id=$1`, sid)

	// Get session context
	var session map[string]interface{}
	if err := h.db.Get(&session, "SELECT * FROM ai_sessions WHERE id=$1", sid); err != nil {
		respondError(w, http.StatusNotFound, "session not found")
		return
	}

	// Get conversation history for context
	var history []map[string]interface{}
	h.db.Select(&history, "SELECT role, content FROM ai_messages WHERE session_id=$1 ORDER BY created_at ASC LIMIT 50", sid)

	// Simulate AI response (in real impl, call external LLM)
	response := fmt.Sprintf("AI Assistant: Received your message \"%s\". Session %s has %d messages. Full AI integration requires connecting an LLM provider.", input.Message, sid[:8], len(history))

	// Save assistant response
	h.db.Exec(`INSERT INTO ai_messages (session_id, role, content) VALUES ($1, 'assistant', $2)`, sid, response)
	h.db.Exec(`UPDATE ai_sessions SET message_count=message_count+1, updated_at=NOW() WHERE id=$1`, sid)

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"session_id": sid,
		"role":       "assistant",
		"content":    response,
	})
}

// === Functions ===

func (h *AIHandler) ListFunctions(w http.ResponseWriter, r *http.Request) {
	var items []map[string]interface{}
	h.db.Select(&items, "SELECT * FROM ai_functions WHERE is_active=TRUE ORDER BY category, name")
	respondJSON(w, http.StatusOK, items)
}

func (h *AIHandler) CreateFunction(w http.ResponseWriter, r *http.Request) {
	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}
	_, err := h.db.NamedExec(`INSERT INTO ai_functions
		(name, description, function_schema, implementation_type, implementation, parameters,
		 return_type, is_system, category, required_permission, notes)
		VALUES (:name, :description, :function_schema, :implementation_type, :implementation, :parameters,
		 :return_type, :is_system, :category, :required_permission, :notes)`, input)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("insert failed: %v", err))
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(input)
}

func (h *AIHandler) GetFunction(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var item map[string]interface{}
	if err := h.db.Get(&item, "SELECT * FROM ai_functions WHERE id=$1", id); err != nil {
		respondError(w, http.StatusNotFound, "function not found")
		return
	}
	respondJSON(w, http.StatusOK, item)
}

func (h *AIHandler) UpdateFunction(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}
	input["id"] = id
	h.db.NamedExec(`UPDATE ai_functions SET
		name=:name, description=:description, function_schema=:function_schema,
		implementation=:implementation, parameters=:parameters, return_type=:return_type,
		category=:category, required_permission=:required_permission, notes=:notes,
		is_active=:is_active, updated_at=NOW()
		WHERE id=:id`, input)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(input)
}

func (h *AIHandler) DeleteFunction(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	h.db.Exec("UPDATE ai_functions SET is_active=FALSE WHERE id=$1", id)
	w.WriteHeader(http.StatusNoContent)
}

func (h *AIHandler) ListFunctionsByCategory(w http.ResponseWriter, r *http.Request) {
	cat := chi.URLParam(r, "category")
	var items []map[string]interface{}
	h.db.Select(&items, "SELECT * FROM ai_functions WHERE category=$1 AND is_active=TRUE ORDER BY name", cat)
	respondJSON(w, http.StatusOK, items)
}

// === Classifications ===
func (h *AIHandler) ListClassifications(w http.ResponseWriter, r *http.Request) {
	var items []map[string]interface{}
	err := h.db.Select(&items, "SELECT * FROM ai_document_classifications ORDER BY created_at DESC LIMIT 50")
	if err != nil {
		items = []map[string]interface{}{}
	}
	respondJSON(w, http.StatusOK, items)
}

// === Predictions ===
func (h *AIHandler) ListPredictions(w http.ResponseWriter, r *http.Request) {
	var items []map[string]interface{}
	err := h.db.Select(&items, "SELECT * FROM ai_predictions ORDER BY created_at DESC LIMIT 50")
	if err != nil {
		items = []map[string]interface{}{}
	}
	respondJSON(w, http.StatusOK, items)
}

// === Cost Estimates ===
func (h *AIHandler) ListCostEstimates(w http.ResponseWriter, r *http.Request) {
	var items []map[string]interface{}
	err := h.db.Select(&items, "SELECT * FROM ai_cost_estimates ORDER BY created_at DESC LIMIT 50")
	if err != nil {
		items = []map[string]interface{}{}
	}
	respondJSON(w, http.StatusOK, items)
}

func init() { log.SetFlags(log.LstdFlags) }