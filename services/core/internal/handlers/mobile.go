package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
)

// MobileHandler — HTTP handler for mobile API, notifications, activity, comments
type MobileHandler struct {
	db *sqlx.DB
}

func NewMobileHandler(db *sqlx.DB) *MobileHandler {
	return &MobileHandler{db: db}
}

func (h *MobileHandler) RegisterRoutes(r chi.Router) {
	// Push Notifications
	r.Route("/notifications", func(r chi.Router) {
		r.Get("/", h.ListNotifications)
		r.Post("/", h.CreateNotification)
		r.Get("/{id}", h.GetNotification)
		r.Put("/{id}/read", h.MarkRead)
		r.Put("/{id}/action", h.MarkActioned)
		r.Get("/user/{userId}", h.ListNotificationsByUser)
		r.Get("/unread/{userId}", h.CountUnread)
	})
	// User Devices
	r.Route("/devices", func(r chi.Router) {
		r.Post("/register", h.RegisterDevice)
		r.Post("/unregister", h.UnregisterDevice)
		r.Get("/user/{userId}", h.ListUserDevices)
	})
	// Activity Feed
	r.Route("/activity", func(r chi.Router) {
		r.Get("/", h.ListActivity)
		r.Post("/", h.CreateActivity)
		r.Get("/project/{projectId}", h.ListActivityByProject)
		r.Get("/recent", h.ListRecentActivity)
	})
	// Universal Comments
	r.Route("/comments", func(r chi.Router) {
		r.Get("/entity/{entityType}/{entityId}", h.ListComments)
		r.Post("/", h.CreateComment)
		r.Put("/{id}", h.UpdateComment)
		r.Delete("/{id}", h.DeleteComment)
		r.Get("/pinned/{entityType}/{entityId}", h.ListPinnedComments)
	})
}

// === Notifications ===

func (h *MobileHandler) ListNotifications(w http.ResponseWriter, r *http.Request) {
	var items []map[string]interface{}
	h.db.Select(&items, "SELECT * FROM push_notifications ORDER BY created_at DESC LIMIT 100")
	respondJSON(w, http.StatusOK, items)
}

func (h *MobileHandler) CreateNotification(w http.ResponseWriter, r *http.Request) {
	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}
	_, err := h.db.NamedExec(`INSERT INTO push_notifications
		(project_id, notification_type, title, body, priority, deep_link, sender_id, recipient_id, recipient_role)
		VALUES (:project_id, :notification_type, :title, :body, :priority, :deep_link, :sender_id, :recipient_id, :recipient_role)`, input)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("insert failed: %v", err))
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(input)
}

func (h *MobileHandler) GetNotification(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var item map[string]interface{}
	if err := h.db.Get(&item, "SELECT * FROM push_notifications WHERE id=$1", id); err != nil {
		respondError(w, http.StatusNotFound, "notification not found")
		return
	}
	respondJSON(w, http.StatusOK, item)
}

func (h *MobileHandler) MarkRead(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	h.db.Exec("UPDATE push_notifications SET read_at=NOW(), status='read' WHERE id=$1", id)
	respondJSON(w, http.StatusOK, map[string]string{"status": "read"})
}

func (h *MobileHandler) MarkActioned(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		ActionType string `json:"action_type"`
	}
	json.NewDecoder(r.Body).Decode(&input)
	h.db.Exec("UPDATE push_notifications SET action_taken_at=NOW(), action_type=$1, status='actioned' WHERE id=$2", input.ActionType, id)
	respondJSON(w, http.StatusOK, map[string]string{"status": "actioned"})
}

func (h *MobileHandler) ListNotificationsByUser(w http.ResponseWriter, r *http.Request) {
	uid := chi.URLParam(r, "userId")
	var items []map[string]interface{}
	h.db.Select(&items, "SELECT * FROM push_notifications WHERE recipient_id=$1 ORDER BY created_at DESC", uid)
	respondJSON(w, http.StatusOK, items)
}

func (h *MobileHandler) CountUnread(w http.ResponseWriter, r *http.Request) {
	uid := chi.URLParam(r, "userId")
	var count int
	h.db.Get(&count, "SELECT COUNT(*) FROM push_notifications WHERE recipient_id=$1 AND read_at IS NULL", uid)
	respondJSON(w, http.StatusOK, map[string]int{"unread": count})
}

// === Devices ===

func (h *MobileHandler) RegisterDevice(w http.ResponseWriter, r *http.Request) {
	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}
	// Upsert
	_, err := h.db.NamedExec(`INSERT INTO user_devices
		(user_id, device_name, device_type, push_token, platform, app_version, os_version)
		VALUES (:user_id, :device_name, :device_type, :push_token, :platform, :app_version, :os_version)
		ON CONFLICT DO NOTHING`, input)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("register failed: %v", err))
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(input)
}

func (h *MobileHandler) UnregisterDevice(w http.ResponseWriter, r *http.Request) {
	var input struct {
		PushToken string `json:"push_token"`
		UserID    string `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}
	h.db.Exec("UPDATE user_devices SET is_active=FALSE WHERE push_token=$1 AND user_id=$2", input.PushToken, input.UserID)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "unregistered"})
}

func (h *MobileHandler) ListUserDevices(w http.ResponseWriter, r *http.Request) {
	uid := chi.URLParam(r, "userId")
	var items []map[string]interface{}
	h.db.Select(&items, "SELECT * FROM user_devices WHERE user_id=$1 AND is_active=TRUE", uid)
	respondJSON(w, http.StatusOK, items)
}

// === Activity Feed ===

func (h *MobileHandler) ListActivity(w http.ResponseWriter, r *http.Request) {
	var items []map[string]interface{}
	h.db.Select(&items, "SELECT * FROM activity_feed ORDER BY created_at DESC LIMIT 100")
	respondJSON(w, http.StatusOK, items)
}

func (h *MobileHandler) CreateActivity(w http.ResponseWriter, r *http.Request) {
	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}
	_, err := h.db.NamedExec(`INSERT INTO activity_feed
		(project_id, activity_type, activity_title, activity_description, entity_type, entity_id,
		 actor_id, actor_name, metadata, importance)
		VALUES (:project_id, :activity_type, :activity_title, :activity_description, :entity_type, :entity_id,
		 :actor_id, :actor_name, :metadata, :importance)`, input)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("insert failed: %v", err))
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(input)
}

func (h *MobileHandler) ListActivityByProject(w http.ResponseWriter, r *http.Request) {
	pid := chi.URLParam(r, "projectId")
	var items []map[string]interface{}
	h.db.Select(&items, "SELECT * FROM activity_feed WHERE project_id=$1 ORDER BY created_at DESC LIMIT 100", pid)
	respondJSON(w, http.StatusOK, items)
}

func (h *MobileHandler) ListRecentActivity(w http.ResponseWriter, r *http.Request) {
	var items []map[string]interface{}
	h.db.Select(&items, "SELECT * FROM activity_feed ORDER BY created_at DESC LIMIT 20")
	respondJSON(w, http.StatusOK, items)
}

// === Comments ===

func (h *MobileHandler) ListComments(w http.ResponseWriter, r *http.Request) {
	et := chi.URLParam(r, "entityType")
	eid := chi.URLParam(r, "entityId")
	var items []map[string]interface{}
	h.db.Select(&items, "SELECT * FROM universal_comments WHERE entity_type=$1 AND entity_id=$2 AND is_archived=FALSE ORDER BY created_at ASC", et, eid)
	respondJSON(w, http.StatusOK, items)
}

func (h *MobileHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}
	_, err := h.db.NamedExec(`INSERT INTO universal_comments
		(project_id, entity_type, entity_id, parent_id, author_id, author_name, content, attachments, mentions)
		VALUES (:project_id, :entity_type, :entity_id, :parent_id, :author_id, :author_name, :content, :attachments, :mentions)`, input)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("insert failed: %v", err))
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(input)
}

func (h *MobileHandler) UpdateComment(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}
	h.db.Exec("UPDATE universal_comments SET content=$1, is_edited=TRUE, updated_at=NOW() WHERE id=$2", input.Content, id)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
}

func (h *MobileHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	h.db.Exec("UPDATE universal_comments SET is_archived=TRUE WHERE id=$1", id)
	w.WriteHeader(http.StatusNoContent)
}

func (h *MobileHandler) ListPinnedComments(w http.ResponseWriter, r *http.Request) {
	et := chi.URLParam(r, "entityType")
	eid := chi.URLParam(r, "entityId")
	var items []map[string]interface{}
	h.db.Select(&items, "SELECT * FROM universal_comments WHERE entity_type=$1 AND entity_id=$2 AND is_pinned=TRUE ORDER BY created_at ASC", et, eid)
	respondJSON(w, http.StatusOK, items)
}

func init() { log.SetFlags(log.LstdFlags) }