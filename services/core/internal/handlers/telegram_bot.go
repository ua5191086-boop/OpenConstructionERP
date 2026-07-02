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
)

// TelegramBotHandler manages Telegram bot integration
type TelegramBotHandler struct {
	db *sql.DB
}

func NewTelegramBotHandler(db *sql.DB) *TelegramBotHandler {
	return &TelegramBotHandler{db: db}
}

func (h *TelegramBotHandler) RegisterRoutes(r chi.Router) {
	r.Route("/telegram-bot", func(r chi.Router) {
		r.Get("/", h.ListBots)
		r.Post("/", h.CreateBot)
		r.Get("/{id}", h.GetBot)
		r.Put("/{id}", h.UpdateBot)
		r.Delete("/{id}", h.DeleteBot)
		r.Post("/{id}/send", h.SendMessage)
		r.Get("/{id}/messages", h.ListMessages)
		r.Get("/{id}/subscribers", h.ListSubscribers)
		r.Post("/{id}/subscribers", h.AddSubscriber)
		r.Delete("/{id}/subscribers/{subId}", h.RemoveSubscriber)
	})
}

type telegramBotResponse struct {
	ID                string    `json:"id"`
	ProjectID         *string   `json:"project_id,omitempty"`
	ChatID            string    `json:"chat_id"`
	ChatTitle         *string   `json:"chat_title,omitempty"`
	BotTokenEncrypted *string   `json:"bot_token_encrypted,omitempty"`
	NotificationTypes []string  `json:"notification_types"`
	IsActive          bool      `json:"is_active"`
	LastMessageAt     *string   `json:"last_message_at,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
}

func (h *TelegramBotHandler) ListBots(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")

	query := `SELECT id, project_id, chat_id, chat_title, bot_token_encrypted,
		notification_types, is_active, last_message_at, created_at
		FROM integration_telegram_bot WHERE 1=1`
	var args []interface{}
	argIdx := 1

	if projectID != "" {
		query += fmt.Sprintf(" AND project_id = $%d", argIdx)
		args = append(args, projectID)
		argIdx++
	}
	query += " ORDER BY created_at DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	bots := make([]telegramBotResponse, 0)
	for rows.Next() {
		var b telegramBotResponse
		var pID, chatTitle, botToken, lastMsg sql.NullString
		var notifTypes []byte

		err := rows.Scan(&b.ID, &pID, &b.ChatID, &chatTitle, &botToken,
			&notifTypes, &b.IsActive, &lastMsg, &b.CreatedAt)
		if err != nil {
			log.Printf("ListBots scan error: %v", err)
			continue
		}
		if pID.Valid { b.ProjectID = &pID.String }
		if chatTitle.Valid { b.ChatTitle = &chatTitle.String }
		if botToken.Valid { t := maskToken(botToken.String); b.BotTokenEncrypted = &t }
		if lastMsg.Valid { b.LastMessageAt = &lastMsg.String }
		json.Unmarshal(notifTypes, &b.NotificationTypes)
		bots = append(bots, b)
	}

	respondJSON(w, http.StatusOK, bots)
}

func (h *TelegramBotHandler) GetBot(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		respondError(w, http.StatusBadRequest, "invalid UUID")
		return
	}

	var b telegramBotResponse
	var pID, chatTitle, botToken, lastMsg sql.NullString
	var notifTypes []byte

	err := h.db.QueryRow(`SELECT id, project_id, chat_id, chat_title, bot_token_encrypted,
		notification_types, is_active, last_message_at, created_at
		FROM integration_telegram_bot WHERE id = $1`, id).
		Scan(&b.ID, &pID, &b.ChatID, &chatTitle, &botToken,
			&notifTypes, &b.IsActive, &lastMsg, &b.CreatedAt)
	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "bot not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if pID.Valid { b.ProjectID = &pID.String }
	if chatTitle.Valid { b.ChatTitle = &chatTitle.String }
	if botToken.Valid { t := maskToken(botToken.String); b.BotTokenEncrypted = &t }
	if lastMsg.Valid { b.LastMessageAt = &lastMsg.String }
	json.Unmarshal(notifTypes, &b.NotificationTypes)

	respondJSON(w, http.StatusOK, b)
}

type telegramBotInput struct {
	ProjectID         *string  `json:"project_id,omitempty"`
	ChatID            string   `json:"chat_id"`
	ChatTitle         *string  `json:"chat_title,omitempty"`
	BotTokenEncrypted *string  `json:"bot_token_encrypted,omitempty"`
	NotificationTypes []string `json:"notification_types,omitempty"`
	IsActive          *bool    `json:"is_active,omitempty"`
}

func (h *TelegramBotHandler) CreateBot(w http.ResponseWriter, r *http.Request) {
	var input telegramBotInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if input.ChatID == "" {
		respondError(w, http.StatusBadRequest, "chat_id is required")
		return
	}

	notifTypes := `["alerts","daily_summary","approvals"]`
	if len(input.NotificationTypes) > 0 {
		b, _ := json.Marshal(input.NotificationTypes)
		notifTypes = string(b)
	}
	isActive := true
	if input.IsActive != nil {
		isActive = *input.IsActive
	}

	var id string
	err := h.db.QueryRow(`INSERT INTO integration_telegram_bot
		(project_id, chat_id, chat_title, bot_token_encrypted, notification_types, is_active)
		VALUES ($1,$2,$3,$4,$5::jsonb,$6) RETURNING id`,
		input.ProjectID, input.ChatID, input.ChatTitle, input.BotTokenEncrypted,
		notifTypes, isActive).Scan(&id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	r.URL.Path = "/api/v1/telegram-bot/" + id
	h.GetBot(w, r)
}

func (h *TelegramBotHandler) UpdateBot(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		respondError(w, http.StatusBadRequest, "invalid UUID")
		return
	}

	var input telegramBotInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	notifTypes := ""
	if len(input.NotificationTypes) > 0 {
		b, _ := json.Marshal(input.NotificationTypes)
		notifTypes = string(b)
	}

	_, err := h.db.Exec(`UPDATE integration_telegram_bot SET
		chat_id=COALESCE(NULLIF($1,''), chat_id),
		chat_title=COALESCE($2, chat_title),
		bot_token_encrypted=COALESCE(NULLIF($3,''), bot_token_encrypted),
		notification_types=COALESCE(NULLIF($4,'')::jsonb, notification_types),
		is_active=COALESCE($5, is_active)
		WHERE id=$6`,
		input.ChatID, input.ChatTitle, input.BotTokenEncrypted,
		notifTypes, input.IsActive, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.GetBot(w, r)
}

func (h *TelegramBotHandler) DeleteBot(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if _, err := uuid.Parse(id); err != nil {
		respondError(w, http.StatusBadRequest, "invalid UUID")
		return
	}

	_, err := h.db.Exec(`DELETE FROM integration_telegram_bot WHERE id = $1`, id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type sendMessageInput struct {
	ChatID    string `json:"chat_id"`
	Message   string `json:"message"`
	ParseMode string `json:"parse_mode,omitempty"`
}

func (h *TelegramBotHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	botID := chi.URLParam(r, "id")
	if _, err := uuid.Parse(botID); err != nil {
		respondError(w, http.StatusBadRequest, "invalid UUID")
		return
	}

	var input sendMessageInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if input.ChatID == "" || input.Message == "" {
		respondError(w, http.StatusBadRequest, "chat_id and message are required")
		return
	}

	parseMode := "HTML"
	if input.ParseMode != "" {
		parseMode = input.ParseMode
	}

	var msgID string
	err := h.db.QueryRow(`INSERT INTO integration_telegram_messages
		(bot_id, chat_id, text, parse_mode, direction, status)
		VALUES ($1, $2, $3, $4, 'outgoing', 'queued') RETURNING id`,
		botID, input.ChatID, input.Message, parseMode).Scan(&msgID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"success":  true,
		"message_id": msgID,
		"status":   "queued",
	})
}

func (h *TelegramBotHandler) ListMessages(w http.ResponseWriter, r *http.Request) {
	botID := chi.URLParam(r, "id")
	if _, err := uuid.Parse(botID); err != nil {
		respondError(w, http.StatusBadRequest, "invalid UUID")
		return
	}

	rows, err := h.db.Query(`SELECT id, bot_id, chat_id, text, parse_mode, direction, status,
		telegram_msg_id, error_message, sent_at, created_at
		FROM integration_telegram_messages WHERE bot_id = $1
		ORDER BY created_at DESC LIMIT 100`, botID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	messages := make([]map[string]interface{}, 0)
	for rows.Next() {
		var msgID, botID2, chatID, text, parseMode, direction, status string
		var tgMsgID, errMsg, sentAt sql.NullString
		var createdAt time.Time
		if err := rows.Scan(&msgID, &botID2, &chatID, &text, &parseMode, &direction, &status,
			&tgMsgID, &errMsg, &sentAt, &createdAt); err != nil {
			continue
		}
		msg := map[string]interface{}{
			"id": msgID, "bot_id": botID2, "chat_id": chatID,
			"text": text, "parse_mode": parseMode, "direction": direction,
			"status": status, "created_at": createdAt,
		}
		if tgMsgID.Valid { msg["telegram_msg_id"] = tgMsgID.String }
		if errMsg.Valid { msg["error_message"] = errMsg.String }
		if sentAt.Valid { msg["sent_at"] = sentAt.String }
		messages = append(messages, msg)
	}

	respondJSON(w, http.StatusOK, messages)
}

func (h *TelegramBotHandler) ListSubscribers(w http.ResponseWriter, r *http.Request) {
	botID := chi.URLParam(r, "id")
	if _, err := uuid.Parse(botID); err != nil {
		respondError(w, http.StatusBadRequest, "invalid UUID")
		return
	}

	rows, err := h.db.Query(`SELECT id, bot_id, chat_id, username, first_name, last_name,
		language_code, is_active, subscribed_at
		FROM integration_telegram_subscribers WHERE bot_id = $1
		ORDER BY subscribed_at DESC`, botID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	subs := make([]map[string]interface{}, 0)
	for rows.Next() {
		var subID, botID2, chatID string
		var username, firstName, lastName, langCode sql.NullString
		var isActive bool
		var subscribedAt time.Time
		if err := rows.Scan(&subID, &botID2, &chatID, &username, &firstName, &lastName,
			&langCode, &isActive, &subscribedAt); err != nil {
			continue
		}
		sub := map[string]interface{}{
			"id": subID, "bot_id": botID2, "chat_id": chatID,
			"is_active": isActive, "subscribed_at": subscribedAt,
		}
		if username.Valid { sub["username"] = username.String }
		if firstName.Valid { sub["first_name"] = firstName.String }
		if lastName.Valid { sub["last_name"] = lastName.String }
		if langCode.Valid { sub["language_code"] = langCode.String }
		subs = append(subs, sub)
	}

	respondJSON(w, http.StatusOK, subs)
}

type addSubscriberInput struct {
	ChatID       string  `json:"chat_id"`
	Username     *string `json:"username,omitempty"`
	FirstName    *string `json:"first_name,omitempty"`
	LastName     *string `json:"last_name,omitempty"`
	LanguageCode *string `json:"language_code,omitempty"`
}

func (h *TelegramBotHandler) AddSubscriber(w http.ResponseWriter, r *http.Request) {
	botID := chi.URLParam(r, "id")
	if _, err := uuid.Parse(botID); err != nil {
		respondError(w, http.StatusBadRequest, "invalid UUID")
		return
	}

	var input addSubscriberInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if input.ChatID == "" {
		respondError(w, http.StatusBadRequest, "chat_id is required")
		return
	}

	lang := "ru"
	if input.LanguageCode != nil {
		lang = *input.LanguageCode
	}

	var id string
	err := h.db.QueryRow(`INSERT INTO integration_telegram_subscribers
		(bot_id, chat_id, username, first_name, last_name, language_code, is_active)
		VALUES ($1,$2,$3,$4,$5,$6,true)
		ON CONFLICT (bot_id, chat_id) DO UPDATE SET
			is_active=true, username=COALESCE($3, integration_telegram_subscribers.username),
			first_name=COALESCE($4, integration_telegram_subscribers.first_name),
			last_name=COALESCE($5, integration_telegram_subscribers.last_name),
			updated_at=NOW()
		RETURNING id`,
		botID, input.ChatID, input.Username, input.FirstName, input.LastName, lang).Scan(&id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, map[string]string{"id": id, "status": "subscribed"})
}

func (h *TelegramBotHandler) RemoveSubscriber(w http.ResponseWriter, r *http.Request) {
	subID := chi.URLParam(r, "subId")
	if _, err := uuid.Parse(subID); err != nil {
		respondError(w, http.StatusBadRequest, "invalid UUID")
		return
	}

	_, err := h.db.Exec(`UPDATE integration_telegram_subscribers SET is_active=false, updated_at=NOW() WHERE id = $1`, subID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func maskToken(token string) string {
	if len(token) <= 8 {
		return "****"
	}
	return token[:4] + "..." + token[len(token)-4:]
}
