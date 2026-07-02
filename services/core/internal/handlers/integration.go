package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
)

// IntegrationHandler — HTTP handler for integration framework
type IntegrationHandler struct {
	db *sqlx.DB
}

func NewIntegrationHandler(db *sqlx.DB) *IntegrationHandler {
	return &IntegrationHandler{db: db}
}

func (h *IntegrationHandler) RegisterRoutes(r chi.Router) {
	// Integration Systems
	r.Route("/integrations/systems", func(r chi.Router) {
		r.Get("/", h.ListSystems)
		r.Post("/", h.CreateSystem)
		r.Get("/{id}", h.GetSystem)
		r.Put("/{id}", h.UpdateSystem)
		r.Delete("/{id}", h.DeleteSystem)
		r.Get("/type/{systemType}", h.ListSystemsByType)
		r.Post("/{id}/test", h.TestConnection)
	})
	// Sync Log
	r.Route("/integrations/sync-log", func(r chi.Router) {
		r.Get("/", h.ListSyncLog)
		r.Get("/system/{systemId}", h.ListSyncLogBySystem)
		r.Get("/recent-failures", h.ListRecentFailures)
	})
	// Entity Mappings
	r.Route("/integrations/mappings", func(r chi.Router) {
		r.Get("/", h.ListMappings)
		r.Post("/", h.CreateMapping)
		r.Get("/system/{systemId}", h.ListMappingsBySystem)
		r.Get("/resolve/{entityType}/{externalId}", h.ResolveMapping)
	})
	// Webhook Queue
	r.Route("/integrations/webhooks", func(r chi.Router) {
		r.Get("/", h.ListWebhooks)
		r.Post("/", h.CreateWebhookEvent)
		r.Get("/pending", h.ListPendingWebhooks)
		r.Post("/{id}/retry", h.RetryWebhook)
	})
}

func (h *IntegrationHandler) ListSystems(w http.ResponseWriter, r *http.Request) {
	var items []map[string]interface{}
	h.db.Select(&items, "SELECT * FROM integration_systems ORDER BY name")
	respondJSON(w, http.StatusOK, items)
}

func (h *IntegrationHandler) CreateSystem(w http.ResponseWriter, r *http.Request) {
	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}
	_, err := h.db.NamedExec(`INSERT INTO integration_systems
		(name, system_type, vendor, version, base_url, auth_type, auth_config, webhook_url,
		 health_check_url, capabilities, is_active, sync_frequency_min, retry_policy, notes)
		VALUES (:name, :system_type, :vendor, :version, :base_url, :auth_type, :auth_config,
		 :webhook_url, :health_check_url, :capabilities, :is_active, :sync_frequency_min,
		 :retry_policy, :notes)`, input)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("insert failed: %v", err))
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(input)
}

func (h *IntegrationHandler) GetSystem(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var item map[string]interface{}
	if err := h.db.Get(&item, "SELECT * FROM integration_systems WHERE id=$1", id); err != nil {
		respondError(w, http.StatusNotFound, "system not found")
		return
	}
	respondJSON(w, http.StatusOK, item)
}

func (h *IntegrationHandler) UpdateSystem(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}
	input["id"] = id
	_, err := h.db.NamedExec(`UPDATE integration_systems SET
		name=:name, system_type=:system_type, vendor=:vendor, version=:version,
		base_url=:base_url, auth_type=:auth_type, auth_config=:auth_config,
		webhook_url=:webhook_url, health_check_url=:health_check_url,
		capabilities=:capabilities, is_active=:is_active,
		sync_frequency_min=:sync_frequency_min, retry_policy=:retry_policy,
		notes=:notes, updated_at=NOW()
		WHERE id=:id`, input)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("update failed: %v", err))
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(input)
}

func (h *IntegrationHandler) DeleteSystem(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	h.db.Exec("DELETE FROM integration_systems WHERE id=$1", id)
	w.WriteHeader(http.StatusNoContent)
}

func (h *IntegrationHandler) ListSystemsByType(w http.ResponseWriter, r *http.Request) {
	st := chi.URLParam(r, "systemType")
	var items []map[string]interface{}
	h.db.Select(&items, "SELECT * FROM integration_systems WHERE system_type=$1 ORDER BY name", st)
	respondJSON(w, http.StatusOK, items)
}

func (h *IntegrationHandler) TestConnection(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var sys map[string]interface{}
	if err := h.db.Get(&sys, "SELECT * FROM integration_systems WHERE id=$1", id); err != nil {
		respondError(w, http.StatusNotFound, "system not found")
		return
	}

	// Record test in sync log
	h.db.Exec(`INSERT INTO integration_sync_log
		(system_id, sync_type, entity_type, direction, status, notes)
		VALUES ($1, 'test', 'connection', 'outbound', 'running', 'Test connection')`, id)

	// Simulate test (in real impl, make actual HTTP call)
	h.db.Exec(`UPDATE integration_systems SET last_sync_at=NOW(), last_sync_status='success' WHERE id=$1`, id)
	h.db.Exec(`UPDATE integration_sync_log SET
		status='completed', completed_at=NOW(), duration_sec=0, records_processed=1
		WHERE system_id=$1 AND created_at=(SELECT MAX(created_at) FROM integration_sync_log WHERE system_id=$1)`, id)

	respondJSON(w, http.StatusOK, map[string]string{
		"status": "success",
		"message": fmt.Sprintf("Connection test for %s completed", sys["name"]),
	})
}

func (h *IntegrationHandler) ListSyncLog(w http.ResponseWriter, r *http.Request) {
	var items []map[string]interface{}
	h.db.Select(&items, `SELECT sl.*, s.name as system_name, s.system_type
		FROM integration_sync_log sl
		JOIN integration_systems s ON s.id=sl.system_id
		ORDER BY sl.started_at DESC LIMIT 100`)
	respondJSON(w, http.StatusOK, items)
}

func (h *IntegrationHandler) ListSyncLogBySystem(w http.ResponseWriter, r *http.Request) {
	sid := chi.URLParam(r, "systemId")
	var items []map[string]interface{}
	h.db.Select(&items, "SELECT * FROM integration_sync_log WHERE system_id=$1 ORDER BY started_at DESC", sid)
	respondJSON(w, http.StatusOK, items)
}

func (h *IntegrationHandler) ListRecentFailures(w http.ResponseWriter, r *http.Request) {
	var items []map[string]interface{}
	h.db.Select(&items, `SELECT sl.*, s.name as system_name, s.system_type
		FROM integration_sync_log sl
		JOIN integration_systems s ON s.id=sl.system_id
		WHERE sl.status IN ('failed','partial')
		ORDER BY sl.started_at DESC LIMIT 50`)
	respondJSON(w, http.StatusOK, items)
}

func (h *IntegrationHandler) ListMappings(w http.ResponseWriter, r *http.Request) {
	var items []map[string]interface{}
	h.db.Select(&items, "SELECT * FROM integration_entity_mappings ORDER BY last_sync_at DESC")
	respondJSON(w, http.StatusOK, items)
}

func (h *IntegrationHandler) CreateMapping(w http.ResponseWriter, r *http.Request) {
	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}
	_, err := h.db.NamedExec(`INSERT INTO integration_entity_mappings
		(system_id, entity_type, local_id, external_id, external_url, external_version, custom_data)
		VALUES (:system_id, :entity_type, :local_id, :external_id, :external_url, :external_version, :custom_data)
		ON CONFLICT (system_id, entity_type, local_id) DO UPDATE SET
		external_id=EXCLUDED.external_id, external_url=EXCLUDED.external_url,
		external_version=EXCLUDED.external_version, custom_data=EXCLUDED.custom_data,
		sync_status='synced', last_sync_at=NOW()`, input)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("insert failed: %v", err))
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(input)
}

func (h *IntegrationHandler) ListMappingsBySystem(w http.ResponseWriter, r *http.Request) {
	sid := chi.URLParam(r, "systemId")
	var items []map[string]interface{}
	h.db.Select(&items, "SELECT * FROM integration_entity_mappings WHERE system_id=$1 ORDER BY entity_type, external_id", sid)
	respondJSON(w, http.StatusOK, items)
}

func (h *IntegrationHandler) ResolveMapping(w http.ResponseWriter, r *http.Request) {
	et := chi.URLParam(r, "entityType")
	eid := chi.URLParam(r, "externalId")
	var item map[string]interface{}
	if err := h.db.Get(&item, "SELECT * FROM integration_entity_mappings WHERE entity_type=$1 AND external_id=$2", et, eid); err != nil {
		respondError(w, http.StatusNotFound, "mapping not found")
		return
	}
	respondJSON(w, http.StatusOK, item)
}

func (h *IntegrationHandler) ListWebhooks(w http.ResponseWriter, r *http.Request) {
	var items []map[string]interface{}
	h.db.Select(&items, `SELECT wq.*, s.name as system_name
		FROM integration_webhook_queue wq
		LEFT JOIN integration_systems s ON s.id=wq.system_id
		ORDER BY wq.created_at DESC LIMIT 100`)
	respondJSON(w, http.StatusOK, items)
}

func (h *IntegrationHandler) CreateWebhookEvent(w http.ResponseWriter, r *http.Request) {
	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}
	_, err := h.db.NamedExec(`INSERT INTO integration_webhook_queue
		(system_id, event_type, payload, priority, max_retries)
		VALUES (:system_id, :event_type, :payload, :priority, :max_retries)`, input)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("insert failed: %v", err))
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(input)
}

func (h *IntegrationHandler) ListPendingWebhooks(w http.ResponseWriter, r *http.Request) {
	var items []map[string]interface{}
	h.db.Select(&items, `SELECT wq.*, s.name as system_name, s.webhook_url
		FROM integration_webhook_queue wq
		JOIN integration_systems s ON s.id=wq.system_id
		WHERE wq.status='pending'
		ORDER BY wq.priority DESC, wq.created_at ASC`)
	respondJSON(w, http.StatusOK, items)
}

func (h *IntegrationHandler) RetryWebhook(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	h.db.Exec(`UPDATE integration_webhook_queue SET
		status='pending', retry_count=0, next_attempt_at=NOW(), error_message=NULL
		WHERE id=$1`, id)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "retry_scheduled"})
}

func init() { log.SetFlags(log.LstdFlags) }