package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"
)

// AssetHandler — HTTP handler for asset management
type AssetHandler struct {
	db *sqlx.DB
}

func NewAssetHandler(db *sqlx.DB) *AssetHandler {
	return &AssetHandler{db: db}
}

func (h *AssetHandler) RegisterRoutes(r chi.Router) {
	r.Route("/assets", func(r chi.Router) {
		r.Get("/", h.ListAssets)
		r.Post("/", h.CreateAsset)
		r.Get("/{id}", h.GetAsset)
		r.Put("/{id}", h.UpdateAsset)
		r.Delete("/{id}", h.DeleteAsset)
		r.Get("/project/{projectId}", h.ListAssetsByProject)
		r.Get("/type/{assetType}", h.ListAssetsByType)
		r.Get("/status/{status}", h.ListAssetsByStatus)
	})
	r.Route("/asset-movements", func(r chi.Router) {
		r.Get("/", h.ListMovements)
		r.Post("/", h.CreateMovement)
		r.Get("/asset/{assetId}", h.ListMovementsByAsset)
	})
	r.Route("/asset-inspections", func(r chi.Router) {
		r.Get("/", h.ListInspections)
		r.Post("/", h.CreateInspection)
		r.Get("/{id}", h.GetInspection)
		r.Get("/asset/{assetId}", h.ListInspectionsByAsset)
		r.Get("/overdue", h.ListOverdueInspections)
	})
	r.Route("/asset-depreciation", func(r chi.Router) {
		r.Get("/", h.ListDepreciation)
		r.Post("/", h.CreateDepreciation)
		r.Get("/asset/{assetId}", h.ListDepreciationByAsset)
		r.Get("/pending", h.ListUnpostedDepreciation)
	})
}

func (h *AssetHandler) ListAssets(w http.ResponseWriter, r *http.Request) {
	var items []map[string]interface{}
	h.db.Select(&items, "SELECT * FROM asset_registry ORDER BY created_at DESC")
	respondJSON(w, http.StatusOK, items)
}

func (h *AssetHandler) CreateAsset(w http.ResponseWriter, r *http.Request) {
	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}
	_, err := h.db.NamedExec(`INSERT INTO asset_registry
		(project_id, asset_type, asset_code, asset_name, serial_number, manufacturer, model,
		 year_manufactured, purchase_date, purchase_cost, currency, current_value,
		 depreciation_method, useful_life_years, salvage_value, depreciation_rate,
		 location, gps_coordinates, assigned_to, department, status, condition,
		 warranty_expiry, insurance_policy, insurance_value, qr_code, documents, notes)
		VALUES (:project_id, :asset_type, :asset_code, :asset_name, :serial_number, :manufacturer, :model,
		 :year_manufactured, :purchase_date, :purchase_cost, :currency, :current_value,
		 :depreciation_method, :useful_life_years, :salvage_value, :depreciation_rate,
		 :location, :gps_coordinates, :assigned_to, :department, :status, :condition,
		 :warranty_expiry, :insurance_policy, :insurance_value, :qr_code, :documents, :notes)`, input)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("insert failed: %v", err))
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(input)
}

func (h *AssetHandler) GetAsset(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var item map[string]interface{}
	if err := h.db.Get(&item, "SELECT * FROM asset_registry WHERE id=$1", id); err != nil {
		respondError(w, http.StatusNotFound, "asset not found")
		return
	}
	respondJSON(w, http.StatusOK, item)
}

func (h *AssetHandler) UpdateAsset(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}
	input["id"] = id
	_, err := h.db.NamedExec(`UPDATE asset_registry SET
		asset_type=:asset_type, asset_code=:asset_code, asset_name=:asset_name,
		serial_number=:serial_number, manufacturer=:manufacturer, model=:model,
		year_manufactured=:year_manufactured, purchase_date=:purchase_date,
		purchase_cost=:purchase_cost, currency=:currency, current_value=:current_value,
		depreciation_method=:depreciation_method, useful_life_years=:useful_life_years,
		salvage_value=:salvage_value, depreciation_rate=:depreciation_rate,
		location=:location, gps_coordinates=:gps_coordinates, assigned_to=:assigned_to,
		department=:department, status=:status, condition=:condition,
		warranty_expiry=:warranty_expiry, insurance_policy=:insurance_policy,
		insurance_value=:insurance_value, qr_code=:qr_code, documents=:documents,
		notes=:notes, updated_at=NOW()
		WHERE id=:id`, input)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("update failed: %v", err))
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(input)
}

func (h *AssetHandler) DeleteAsset(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	h.db.Exec("DELETE FROM asset_registry WHERE id=$1", id)
	w.WriteHeader(http.StatusNoContent)
}

func (h *AssetHandler) ListAssetsByProject(w http.ResponseWriter, r *http.Request) {
	pid := chi.URLParam(r, "projectId")
	var items []map[string]interface{}
	h.db.Select(&items, "SELECT * FROM asset_registry WHERE project_id=$1 ORDER BY asset_code", pid)
	respondJSON(w, http.StatusOK, items)
}

func (h *AssetHandler) ListAssetsByType(w http.ResponseWriter, r *http.Request) {
	at := chi.URLParam(r, "assetType")
	var items []map[string]interface{}
	h.db.Select(&items, "SELECT * FROM asset_registry WHERE asset_type=$1 ORDER BY asset_name", at)
	respondJSON(w, http.StatusOK, items)
}

func (h *AssetHandler) ListAssetsByStatus(w http.ResponseWriter, r *http.Request) {
	st := chi.URLParam(r, "status")
	var items []map[string]interface{}
	h.db.Select(&items, "SELECT * FROM asset_registry WHERE status=$1 ORDER BY asset_name", st)
	respondJSON(w, http.StatusOK, items)
}

// Movements
func (h *AssetHandler) ListMovements(w http.ResponseWriter, r *http.Request) {
	var items []map[string]interface{}
	h.db.Select(&items, "SELECT * FROM asset_movements ORDER BY movement_date DESC")
	respondJSON(w, http.StatusOK, items)
}

func (h *AssetHandler) CreateMovement(w http.ResponseWriter, r *http.Request) {
	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}
	_, err := h.db.NamedExec(`INSERT INTO asset_movements
		(asset_id, project_id, movement_type, from_location, to_location, from_assignee,
		 to_assignee, movement_date, reference_doc, authorized_by, reason, notes)
		VALUES (:asset_id, :project_id, :movement_type, :from_location, :to_location, :from_assignee,
		 :to_assignee, :movement_date, :reference_doc, :authorized_by, :reason, :notes)`, input)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("insert failed: %v", err))
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(input)
}

func (h *AssetHandler) ListMovementsByAsset(w http.ResponseWriter, r *http.Request) {
	aid := chi.URLParam(r, "assetId")
	var items []map[string]interface{}
	h.db.Select(&items, "SELECT * FROM asset_movements WHERE asset_id=$1 ORDER BY movement_date DESC", aid)
	respondJSON(w, http.StatusOK, items)
}

// Inspections
func (h *AssetHandler) ListInspections(w http.ResponseWriter, r *http.Request) {
	var items []map[string]interface{}
	h.db.Select(&items, "SELECT * FROM asset_inspections ORDER BY inspection_date DESC")
	respondJSON(w, http.StatusOK, items)
}

func (h *AssetHandler) CreateInspection(w http.ResponseWriter, r *http.Request) {
	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}
	_, err := h.db.NamedExec(`INSERT INTO asset_inspections
		(asset_id, project_id, inspection_type, inspection_date, next_inspection_date,
		 inspector, inspector_company, result, findings, recommendations, action_taken,
		 cost, document_ref, status, notes)
		VALUES (:asset_id, :project_id, :inspection_type, :inspection_date, :next_inspection_date,
		 :inspector, :inspector_company, :result, :findings, :recommendations, :action_taken,
		 :cost, :document_ref, :status, :notes)`, input)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("insert failed: %v", err))
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(input)
}

func (h *AssetHandler) GetInspection(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var item map[string]interface{}
	if err := h.db.Get(&item, "SELECT * FROM asset_inspections WHERE id=$1", id); err != nil {
		respondError(w, http.StatusNotFound, "inspection not found")
		return
	}
	respondJSON(w, http.StatusOK, item)
}

func (h *AssetHandler) ListInspectionsByAsset(w http.ResponseWriter, r *http.Request) {
	aid := chi.URLParam(r, "assetId")
	var items []map[string]interface{}
	h.db.Select(&items, "SELECT * FROM asset_inspections WHERE asset_id=$1 ORDER BY inspection_date DESC", aid)
	respondJSON(w, http.StatusOK, items)
}

func (h *AssetHandler) ListOverdueInspections(w http.ResponseWriter, r *http.Request) {
	var items []map[string]interface{}
	h.db.Select(&items, `SELECT ai.*, ar.asset_code, ar.asset_name
		FROM asset_inspections ai
		JOIN asset_registry ar ON ar.id=ai.asset_id
		WHERE ai.next_inspection_date < CURRENT_DATE AND ai.status != 'overdue'`)
	respondJSON(w, http.StatusOK, items)
}

// Depreciation
func (h *AssetHandler) ListDepreciation(w http.ResponseWriter, r *http.Request) {
	var items []map[string]interface{}
	h.db.Select(&items, "SELECT * FROM asset_depreciation ORDER BY period_start DESC")
	respondJSON(w, http.StatusOK, items)
}

func (h *AssetHandler) CreateDepreciation(w http.ResponseWriter, r *http.Request) {
	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}
	_, err := h.db.NamedExec(`INSERT INTO asset_depreciation
		(asset_id, project_id, period_start, period_end, depreciation_amount,
		 accumulated_depr, book_value, method, posted, notes)
		VALUES (:asset_id, :project_id, :period_start, :period_end, :depreciation_amount,
		 :accumulated_depr, :book_value, :method, :posted, :notes)`, input)
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("insert failed: %v", err))
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(input)
}

func (h *AssetHandler) ListDepreciationByAsset(w http.ResponseWriter, r *http.Request) {
	aid := chi.URLParam(r, "assetId")
	var items []map[string]interface{}
	h.db.Select(&items, "SELECT * FROM asset_depreciation WHERE asset_id=$1 ORDER BY period_start DESC", aid)
	respondJSON(w, http.StatusOK, items)
}

func (h *AssetHandler) ListUnpostedDepreciation(w http.ResponseWriter, r *http.Request) {
	var items []map[string]interface{}
	h.db.Select(&items, "SELECT * FROM asset_depreciation WHERE posted=FALSE ORDER BY period_start DESC")
	respondJSON(w, http.StatusOK, items)
}

func init() { log.SetFlags(log.LstdFlags) }