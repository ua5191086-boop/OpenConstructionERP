package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
)

type FinancialHandler struct {
	db *sqlx.DB
}

func NewFinancialHandler(db *sqlx.DB) *FinancialHandler {
	return &FinancialHandler{db: db}
}

func (h *FinancialHandler) RegisterRoutes(r chi.Router) {
	// V054 — Consolidation
	r.Route("/finance/group-entities", crud(h.db, "group_legal_entities"))
	r.Route("/finance/intercompany", func(r chi.Router) {
		r.Get("/", h.list(h.db, "intercompany_transactions"))
		r.Post("/", h.genInsert("intercompany_transactions"))
		r.Get("/{id}", h.genGet("intercompany_transactions"))
		r.Get("/from/{projectId}", h.genList("intercompany_transactions", "from_project_id=$1"))
		r.Get("/to/{projectId}", h.genList("intercompany_transactions", "to_project_id=$1"))
	})
	r.Route("/finance/consolidation", func(r chi.Router) {
		r.Get("/reports", h.list(h.db, "consolidation_reports"))
		r.Post("/", h.genInsert("consolidation_reports"))
		r.Get("/{id}", h.genGet("consolidation_reports"))
	})
	// V055 — Loans
	r.Route("/finance/loans", func(r chi.Router) {
		r.Get("/facilities", h.list(h.db, "loan_facilities"))
		r.Post("/facilities", h.genInsert("loan_facilities"))
		r.Get("/facilities/{id}", h.genGet("loan_facilities"))
		r.Put("/facilities/{id}", h.genUpdate("loan_facilities"))
		r.Get("/facilities/project/{projectId}", h.genList("loan_facilities", "project_id=$1"))
	})
	r.Route("/finance/loans/drawdowns", func(r chi.Router) {
		r.Get("/", h.list(h.db, "loan_drawdowns"))
		r.Post("/", h.genInsert("loan_drawdowns"))
		r.Get("/{id}", h.genGet("loan_drawdowns"))
		r.Get("/facility/{facilityId}", h.genList("loan_drawdowns", "facility_id=$1"))
	})
	r.Route("/finance/loans/repayments", func(r chi.Router) {
		r.Get("/", h.list(h.db, "loan_repayment_schedule"))
		r.Post("/", h.genInsert("loan_repayment_schedule"))
		r.Get("/overdue", h.genList("loan_repayment_schedule", "status='pending' AND due_date < CURRENT_DATE"))
	})
	r.Route("/finance/loans/covenants", func(r chi.Router) {
		r.Get("/", h.list(h.db, "loan_covenant_monitoring"))
		r.Post("/", h.genInsert("loan_covenant_monitoring"))
		r.Get("/breaches", h.genList("loan_covenant_monitoring", "status='breach'"))
	})
}

func (h *FinancialHandler) list(db *sqlx.DB, table string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var items []map[string]interface{}
		db.Select(&items, fmt.Sprintf("SELECT * FROM %s ORDER BY created_at DESC LIMIT 100", table))
		respondJSON(w, http.StatusOK, items)
	}
}

func (h *FinancialHandler) genGet(table string) http.HandlerFunc {
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

func (h *FinancialHandler) genInsert(table string) http.HandlerFunc {
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

func (h *FinancialHandler) genUpdate(table string) http.HandlerFunc {
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

func (h *FinancialHandler) genList(table, where string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var items []map[string]interface{}
		q := fmt.Sprintf("SELECT * FROM %s WHERE %s", table, where)
		h.db.Select(&items, q, chi.URLParam(r, "projectId"), chi.URLParam(r, "facilityId"))
		respondJSON(w, http.StatusOK, items)
	}
}

func crud(db *sqlx.DB, table string) func(r chi.Router) {
	h := &FinancialHandler{db: db}
	return func(r chi.Router) {
		r.Get("/", h.list(db, table))
		r.Post("/", h.genInsert(table))
		r.Get("/{id}", h.genGet(table))
		r.Put("/{id}", h.genUpdate(table))
		r.Delete("/{id}", func(w http.ResponseWriter, req *http.Request) {
			h.db.Exec(fmt.Sprintf("DELETE FROM %s WHERE id=$1", table), chi.URLParam(req, "id"))
			w.WriteHeader(http.StatusNoContent)
		})
	}
}

func init() { log.SetFlags(log.LstdFlags) }