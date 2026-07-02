package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
)

// TunnelLogisticsHandler — tunnel logistics, ventilation, fire safety
type TunnelLogisticsHandler struct {
	db *sqlx.DB
}

func NewTunnelLogisticsHandler(db *sqlx.DB) *TunnelLogisticsHandler {
	return &TunnelLogisticsHandler{db: db}
}

func (h *TunnelLogisticsHandler) RegisterRoutes(r chi.Router) {
	// V051 — Logistics
	r.Route("/tunnel/logistics/routes", func(r chi.Router) {
		r.Get("/", h.list(h.db, "tunnel_logistics_routes"))
		r.Post("/", h.create("tunnel_logistics_routes"))
		r.Get("/{id}", h.get("tunnel_logistics_routes"))
		r.Put("/{id}", h.update("tunnel_logistics_routes"))
		r.Delete("/{id}", h.del("tunnel_logistics_routes"))
		r.Get("/project/{projectId}", h.listWhere("tunnel_logistics_routes", "project_id=$1"))
		r.Get("/type/{routeType}", h.listWhere("tunnel_logistics_routes", "route_type=$1"))
	})
	r.Route("/tunnel/logistics/schedules", func(r chi.Router) {
		r.Get("/", h.list(h.db, "tunnel_delivery_schedules"))
		r.Post("/", h.create("tunnel_delivery_schedules"))
		r.Get("/{id}", h.get("tunnel_delivery_schedules"))
		r.Put("/{id}", h.update("tunnel_delivery_schedules"))
		r.Get("/date/{date}", h.listWhere("tunnel_delivery_schedules", "delivery_date=$1"))
	})
	r.Route("/tunnel/logistics/events", func(r chi.Router) {
		r.Get("/", h.list(h.db, "tunnel_logistics_events"))
		r.Post("/", h.create("tunnel_logistics_events"))
		r.Get("/{id}", h.get("tunnel_logistics_events"))
		r.Get("/critical", h.listWhere("tunnel_logistics_events", "severity='critical'"))
	})
	r.Route("/tunnel/logistics/inventory", func(r chi.Router) {
		r.Get("/face", h.list(h.db, "tunnel_face_inventory"))
		r.Post("/", h.create("tunnel_face_inventory"))
		r.Get("/tbm/{tbmId}", h.listWhere("tunnel_face_inventory", "tbm_id=$1"))
	})
	// V052 — Ventilation
	r.Route("/tunnel/ventilation/zones", func(r chi.Router) {
		r.Get("/", h.list(h.db, "tunnel_ventilation_zones"))
		r.Post("/", h.create("tunnel_ventilation_zones"))
		r.Get("/{id}", h.get("tunnel_ventilation_zones"))
		r.Put("/{id}", h.update("tunnel_ventilation_zones"))
		r.Get("/project/{projectId}", h.listWhere("tunnel_ventilation_zones", "project_id=$1"))
	})
	r.Route("/tunnel/ventilation/equipment", func(r chi.Router) {
		r.Get("/", h.list(h.db, "tunnel_ventilation_equipment"))
		r.Post("/", h.create("tunnel_ventilation_equipment"))
		r.Get("/{id}", h.get("tunnel_ventilation_equipment"))
		r.Put("/{id}", h.update("tunnel_ventilation_equipment"))
		r.Get("/zone/{zoneId}", h.listWhere("tunnel_ventilation_equipment", "zone_id=$1"))
	})
	r.Route("/tunnel/air-quality", func(r chi.Router) {
		r.Get("/readings", h.list(h.db, "tunnel_air_quality_readings"))
		r.Post("/", h.create("tunnel_air_quality_readings"))
		r.Get("/alarms", h.listWhere("tunnel_air_quality_readings", "is_alarm=TRUE"))
		r.Get("/zone/{zoneId}", h.listWhere("tunnel_air_quality_readings", "zone_id=$1 ORDER BY reading_time DESC LIMIT 50"))
	})
	r.Route("/tunnel/ventilation/emergency", func(r chi.Router) {
		r.Get("/scenarios", h.list(h.db, "tunnel_ventilation_emergency_scenarios"))
		r.Post("/", h.create("tunnel_ventilation_emergency_scenarios"))
		r.Get("/{id}", h.get("tunnel_ventilation_emergency_scenarios"))
	})
	// V053 — Fire Safety
	r.Route("/tunnel/fire/zones", func(r chi.Router) {
		r.Get("/", h.list(h.db, "tunnel_fire_zones"))
		r.Post("/", h.create("tunnel_fire_zones"))
		r.Get("/{id}", h.get("tunnel_fire_zones"))
		r.Put("/{id}", h.update("tunnel_fire_zones"))
		r.Get("/project/{projectId}", h.listWhere("tunnel_fire_zones", "project_id=$1"))
	})
	r.Route("/tunnel/fire/equipment", func(r chi.Router) {
		r.Get("/", h.list(h.db, "tunnel_fire_equipment"))
		r.Post("/", h.create("tunnel_fire_equipment"))
		r.Get("/{id}", h.get("tunnel_fire_equipment"))
		r.Put("/{id}", h.update("tunnel_fire_equipment"))
		r.Get("/overdue", h.listWhere("tunnel_fire_equipment", "next_inspection < CURRENT_DATE AND status='operational'"))
	})
	r.Route("/tunnel/fire/evacuation", func(r chi.Router) {
		r.Get("/routes", h.list(h.db, "tunnel_evacuation_routes"))
		r.Post("/", h.create("tunnel_evacuation_routes"))
		r.Get("/{id}", h.get("tunnel_evacuation_routes"))
	})
	r.Route("/tunnel/fire/drills", func(r chi.Router) {
		r.Get("/", h.list(h.db, "tunnel_fire_drills"))
		r.Post("/", h.create("tunnel_fire_drills"))
		r.Get("/{id}", h.get("tunnel_fire_drills"))
	})
	r.Route("/tunnel/fire/protection-lining", func(r chi.Router) {
		r.Get("/", h.list(h.db, "tunnel_fire_protection_lining"))
		r.Post("/", h.create("tunnel_fire_protection_lining"))
		r.Get("/{id}", h.get("tunnel_fire_protection_lining"))
		r.Get("/project/{projectId}", h.listWhere("tunnel_fire_protection_lining", "project_id=$1"))
	})
}

// Generic CRUD helpers
func (h *TunnelLogisticsHandler) list(db *sqlx.DB, table string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var items []map[string]interface{}
		if err := db.Select(&items, fmt.Sprintf("SELECT * FROM %s ORDER BY created_at DESC LIMIT 100", table)); err != nil {
			respondError(w, http.StatusInternalServerError, fmt.Sprintf("query failed: %v", err))
			return
		}
		respondJSON(w, http.StatusOK, items)
	}
}

func (h *TunnelLogisticsHandler) create(table string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			respondError(w, http.StatusBadRequest, "invalid json")
			return
		}
		cols, vals, args := buildInsert(table, input)
		_, err := h.db.Exec(fmt.Sprintf("INSERT INTO %s %s VALUES %s", table, cols, vals), args...)
		if err != nil {
			respondError(w, http.StatusInternalServerError, fmt.Sprintf("insert failed: %v", err))
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(input)
	}
}

func (h *TunnelLogisticsHandler) get(table string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		var item map[string]interface{}
		if err := h.db.Get(&item, fmt.Sprintf("SELECT * FROM %s WHERE id=$1", table), id); err != nil {
			respondError(w, http.StatusNotFound, "not found")
			return
		}
		respondJSON(w, http.StatusOK, item)
	}
}

func (h *TunnelLogisticsHandler) update(table string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		var input map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			respondError(w, http.StatusBadRequest, "invalid json")
			return
		}
		setClause, args := buildUpdate(table, input, id)
		_, err := h.db.Exec(fmt.Sprintf("UPDATE %s SET %s, updated_at=NOW() WHERE id=$%d", table, setClause, len(args)), args...)
		if err != nil {
			respondError(w, http.StatusInternalServerError, fmt.Sprintf("update failed: %v", err))
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(input)
	}
}

func (h *TunnelLogisticsHandler) del(table string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		h.db.Exec(fmt.Sprintf("DELETE FROM %s WHERE id=$1", table), id)
		w.WriteHeader(http.StatusNoContent)
	}
}

func (h *TunnelLogisticsHandler) listWhere(table, where string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var items []map[string]interface{}
		q := fmt.Sprintf("SELECT * FROM %s WHERE %s", table, where)
		h.db.Select(&items, q, chi.URLParam(r, "projectId"), chi.URLParam(r, "routeType"),
			chi.URLParam(r, "date"), chi.URLParam(r, "tbmId"), chi.URLParam(r, "zoneId"),
			chi.URLParam(r, "entityType"), chi.URLParam(r, "entityId"))
		respondJSON(w, http.StatusOK, items)
	}
}

func buildInsert(table string, input map[string]interface{}) (string, string, []interface{}) {
	cols := "("
	vals := "("
	var args []interface{}
	i := 1
	for k, v := range input {
		if k == "id" || k == "created_at" || k == "updated_at" {
			continue
		}
		cols += k + ","
		vals += fmt.Sprintf("$%d,", i)
		args = append(args, v)
		i++
	}
	cols = cols[:len(cols)-1] + ")"
	vals = vals[:len(vals)-1] + ")"
	return cols, vals, args
}

func buildUpdate(table string, input map[string]interface{}, id string) (string, []interface{}) {
	setClause := ""
	var args []interface{}
	i := 1
	for k, v := range input {
		if k == "id" || k == "created_at" || k == "updated_at" {
			continue
		}
		setClause += fmt.Sprintf("%s=$%d,", k, i)
		args = append(args, v)
		i++
	}
	setClause = setClause[:len(setClause)-1]
	args = append(args, id)
	return setClause, args
}

func init() { log.SetFlags(log.LstdFlags) }