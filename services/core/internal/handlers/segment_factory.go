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

type SegmentFactoryHandler struct{ db *sql.DB }

func NewSegmentFactoryHandler(db *sql.DB) *SegmentFactoryHandler {
	return &SegmentFactoryHandler{db: db}
}

func (h *SegmentFactoryHandler) RegisterRoutes(r chi.Router) {
	r.Route("/segment-factory", func(r chi.Router) {
		r.Get("/lines", h.ListLines)
		r.Post("/lines", h.CreateLine)
		r.Get("/lines/{id}", h.GetLine)
		r.Put("/lines/{id}", h.UpdateLine)
		r.Delete("/lines/{id}", h.DeleteLine)

		r.Get("/plans", h.ListPlans)
		r.Post("/plans", h.CreatePlan)
		r.Get("/plans/{id}", h.GetPlan)
		r.Put("/plans/{id}", h.UpdatePlan)
		r.Delete("/plans/{id}", h.DeletePlan)

		r.Get("/batches", h.ListBatches)
		r.Post("/batches", h.CreateBatch)
		r.Get("/batches/{id}", h.GetBatch)
		r.Put("/batches/{id}", h.UpdateBatch)
		r.Delete("/batches/{id}", h.DeleteBatch)

		r.Get("/stock", h.ListStock)
		r.Post("/stock", h.CreateStockEntry)
		r.Get("/stock/{id}", h.GetStockEntry)
		r.Put("/stock/{id}", h.UpdateStockEntry)
		r.Post("/stock/ship", h.ShipStock)

		r.Get("/qc-records", h.ListQCRecords)
		r.Post("/qc-records", h.CreateQCRecord)
		r.Get("/qc-records/{id}", h.GetQCRecord)

		r.Get("/summary", h.GetSummary)
	})
}

// =============================================================================
// Lines
// =============================================================================
func (h *SegmentFactoryHandler) ListLines(w http.ResponseWriter, r *http.Request) {
	query := `SELECT id, project_id, line_code, line_name, line_type, capacity_per_day, mould_count, curing_method, curing_hours, status, location, notes, created_at FROM segment_factory_lines ORDER BY line_code`
	rows, err := h.db.Query(query)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()
	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, pid, code, name, ltype, curing, status, loc, notes string
		var capPerDay, moulds int
		var curingHrs float64
		var crAt time.Time
		if err := rows.Scan(&id, &pid, &code, &name, &ltype, &capPerDay, &moulds, &curing, &curingHrs, &status, &loc, &notes, &crAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		items = append(items, map[string]interface{}{
			"id": id, "project_id": pid, "line_code": code, "line_name": name,
			"line_type": ltype, "capacity_per_day": capPerDay, "mould_count": moulds,
			"curing_method": curing, "curing_hours": curingHrs, "status": status,
			"location": loc, "notes": notes, "created_at": crAt,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *SegmentFactoryHandler) CreateLine(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID     string `json:"project_id"`
		LineCode      string `json:"line_code"`
		LineName      string `json:"line_name"`
		LineType      string `json:"line_type"`
		CapacityDay   int    `json:"capacity_per_day"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	_, err := h.db.Exec(`INSERT INTO segment_factory_lines (id, project_id, line_code, line_name, line_type, capacity_per_day, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$7)`,
		id, input.ProjectID, input.LineCode, input.LineName, input.LineType, input.CapacityDay, time.Now())
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *SegmentFactoryHandler) GetLine(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var code, name string
	err := h.db.QueryRow(`SELECT line_code, line_name FROM segment_factory_lines WHERE id = $1`, id).Scan(&code, &name)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "line not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "line_code": code, "line_name": name})
}

func (h *SegmentFactoryHandler) UpdateLine(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Status *string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	_, err := h.db.Exec(`UPDATE segment_factory_lines SET status=COALESCE($1,status), updated_at=$2 WHERE id=$3`, input.Status, time.Now(), id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *SegmentFactoryHandler) DeleteLine(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM segment_factory_lines WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// Plans
// =============================================================================
func (h *SegmentFactoryHandler) ListPlans(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT p.id, p.project_id, p.plan_code, p.plan_name, p.plan_date, p.line_id, l.line_name, p.ring_type, p.concrete_grade, p.segment_count, p.planned_rings, p.produced_rings, p.rejected_rings, p.status, p.notes, p.created_at
		FROM segment_production_plans p LEFT JOIN segment_factory_lines l ON l.id = p.line_id WHERE 1=1`
	args := []interface{}{}
	argIdx := 1
	if projectID != "" { query += fmt.Sprintf(" AND p.project_id = $%d", argIdx); argIdx++; args = append(args, projectID) }
	query += " ORDER BY p.plan_date DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()
	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, pid, code, name, ringType, concrete, lineName, notes, status string
		var planDate time.Time
		var lineID sql.NullString
		var segCount, planned, produced, rejected int
		var crAt time.Time
		if err := rows.Scan(&id, &pid, &code, &name, &planDate, &lineID, &lineName, &ringType, &concrete, &segCount, &planned, &produced, &rejected, &status, &notes, &crAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		items = append(items, map[string]interface{}{
			"id": id, "project_id": pid, "plan_code": code, "plan_name": name,
			"plan_date": planDate, "line_id": lineID, "line_name": lineName,
			"ring_type": ringType, "concrete_grade": concrete, "segment_count": segCount,
			"planned_rings": planned, "produced_rings": produced, "rejected_rings": rejected,
			"status": status, "notes": notes, "created_at": crAt,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *SegmentFactoryHandler) CreatePlan(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID    string `json:"project_id"`
		PlanCode     string `json:"plan_code"`
		PlanName     string `json:"plan_name"`
		PlanDate     string `json:"plan_date"`
		RingType     string `json:"ring_type"`
		SegmentCount int    `json:"segment_count"`
		PlannedRings int    `json:"planned_rings"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	_, err := h.db.Exec(`INSERT INTO segment_production_plans (id, project_id, plan_code, plan_name, plan_date, ring_type, segment_count, planned_rings, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$9)`,
		id, input.ProjectID, input.PlanCode, input.PlanName, input.PlanDate, input.RingType, input.SegmentCount, input.PlannedRings, time.Now())
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *SegmentFactoryHandler) GetPlan(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var code, name string
	err := h.db.QueryRow(`SELECT plan_code, plan_name FROM segment_production_plans WHERE id = $1`, id).Scan(&code, &name)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "plan not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "plan_code": code, "plan_name": name})
}

func (h *SegmentFactoryHandler) UpdatePlan(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		ProducedRings *int `json:"produced_rings"`
		Status        *string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	_, err := h.db.Exec(`UPDATE segment_production_plans SET produced_rings=COALESCE($1,produced_rings), status=COALESCE($2,status), updated_at=$3 WHERE id=$4`, input.ProducedRings, input.Status, time.Now(), id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *SegmentFactoryHandler) DeletePlan(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM segment_production_plans WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// Production Batches
// =============================================================================
func (h *SegmentFactoryHandler) ListBatches(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	status := r.URL.Query().Get("status")
	query := `SELECT b.id, b.project_id, b.plan_id, p.plan_code, b.line_id, l.line_name, b.batch_number, b.ring_number, b.segment_number, b.segment_type, b.status, b.pour_time, b.qc_passed, b.stocked_at, b.notes, b.created_at
		FROM segment_production_batches b
		LEFT JOIN segment_production_plans p ON p.id = b.plan_id
		LEFT JOIN segment_factory_lines l ON l.id = b.line_id WHERE 1=1`
	args := []interface{}{}
	argIdx := 1
	if projectID != "" { query += fmt.Sprintf(" AND b.project_id = $%d", argIdx); argIdx++; args = append(args, projectID) }
	if status != "" { query += fmt.Sprintf(" AND b.status = $%d", argIdx); argIdx++; args = append(args, status) }
	query += " ORDER BY b.ring_number, b.segment_number"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()
	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, pid, lineName, batchNum, segType, notes, status string
		var planID, lineID sql.NullString
		var planCode sql.NullString
		var ringNum, segNum int
		var pourTime, stockedAt sql.NullTime
		var qcPassed sql.NullBool
		var crAt time.Time
		if err := rows.Scan(&id, &pid, &planID, &planCode, &lineID, &lineName, &batchNum, &ringNum, &segNum, &segType, &status, &pourTime, &qcPassed, &stockedAt, &notes, &crAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		item := map[string]interface{}{
			"id": id, "project_id": pid, "batch_number": batchNum,
			"ring_number": ringNum, "segment_number": segNum, "segment_type": segType,
			"status": status, "notes": notes, "created_at": crAt,
			"plan_code": planCode, "line_name": lineName,
		}
		if planID.Valid { item["plan_id"] = planID.String }
		if lineID.Valid { item["line_id"] = lineID.String }
		if pourTime.Valid { item["pour_time"] = pourTime.Time }
		if stockedAt.Valid { item["stocked_at"] = stockedAt.Time }
		if qcPassed.Valid { item["qc_passed"] = qcPassed.Bool }
		items = append(items, item)
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *SegmentFactoryHandler) CreateBatch(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID    string `json:"project_id"`
		BatchNumber  string `json:"batch_number"`
		RingNumber   int    `json:"ring_number"`
		SegmentNumber int   `json:"segment_number"`
		SegmentType  string `json:"segment_type"`
		ConcreteGrade string `json:"concrete_grade"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	_, err := h.db.Exec(`INSERT INTO segment_production_batches (id, project_id, batch_number, ring_number, segment_number, segment_type, concrete_grade, status, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,'planned',$8,$8)`,
		id, input.ProjectID, input.BatchNumber, input.RingNumber, input.SegmentNumber, input.SegmentType, input.ConcreteGrade, time.Now())
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *SegmentFactoryHandler) GetBatch(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var batchNum string
	var ringNum, segNum int
	err := h.db.QueryRow(`SELECT batch_number, ring_number, segment_number FROM segment_production_batches WHERE id = $1`, id).Scan(&batchNum, &ringNum, &segNum)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "batch not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "batch_number": batchNum, "ring_number": ringNum, "segment_number": segNum})
}

func (h *SegmentFactoryHandler) UpdateBatch(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Status    *string `json:"status"`
		QCPassed  *bool   `json:"qc_passed"`
		QCBy      *string `json:"qc_checked_by"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	now := time.Now()
	qs := "UPDATE segment_production_batches SET status=COALESCE($1,status), updated_at=$2"
	args := []interface{}{input.Status, now}

	if input.QCPassed != nil {
		qs += ", qc_passed=$3, qc_checked_by=$4, qc_checked_at=$5"
		args = append(args, *input.QCPassed, input.QCBy, now)
		statusVal := "qc_passed"
		if !*input.QCPassed { statusVal = "qc_failed" }
		qs += ", status='" + statusVal + "'"
	}
	qs += " WHERE id=$" + itoa(len(args)+1)
	args = append(args, id)
	_, err := h.db.Exec(qs, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *SegmentFactoryHandler) DeleteBatch(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	_, err := h.db.Exec(`DELETE FROM segment_production_batches WHERE id = $1`, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// =============================================================================
// Stock
// =============================================================================
func (h *SegmentFactoryHandler) ListStock(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	status := r.URL.Query().Get("status")
	query := `SELECT s.id, s.project_id, s.batch_id, b.batch_number, s.ring_number, s.segment_number, s.segment_type, s.production_date, s.stock_date, s.location, s.status, s.destination, s.notes, s.created_at
		FROM segment_stock s LEFT JOIN segment_production_batches b ON b.id = s.batch_id WHERE 1=1`
	args := []interface{}{}
	argIdx := 1
	if projectID != "" { query += fmt.Sprintf(" AND s.project_id = $%d", argIdx); argIdx++; args = append(args, projectID) }
	if status != "" { query += fmt.Sprintf(" AND s.status = $%d", argIdx); argIdx++; args = append(args, status) }
	query += " ORDER BY s.ring_number, s.segment_number"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()
	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, pid, ringType, loc, st, dest, notes string
		var batchID, batchNum sql.NullString
		var ringNum, segNum int
		var prodDate, stockDate, crAt time.Time
		if err := rows.Scan(&id, &pid, &batchID, &batchNum, &ringNum, &segNum, &ringType, &prodDate, &stockDate, &loc, &st, &dest, &notes, &crAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		items = append(items, map[string]interface{}{
			"id": id, "project_id": pid, "ring_number": ringNum, "segment_number": segNum,
			"segment_type": ringType, "production_date": prodDate, "stock_date": stockDate,
			"location": loc, "status": st, "destination": dest, "notes": notes,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *SegmentFactoryHandler) CreateStockEntry(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID      string `json:"project_id"`
		BatchID        string `json:"batch_id"`
		RingNumber     int    `json:"ring_number"`
		SegmentNumber  int    `json:"segment_number"`
		SegmentType    string `json:"segment_type"`
		ProductionDate string `json:"production_date"`
		Location       string `json:"location"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO segment_stock (id, project_id, batch_id, ring_number, segment_number, segment_type, production_date, stock_date, location, status, created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,'in_stock',$10)`,
		id, input.ProjectID, input.BatchID, input.RingNumber, input.SegmentNumber, input.SegmentType, input.ProductionDate, now, input.Location, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	// Update batch status
	h.db.Exec(`UPDATE segment_production_batches SET status='stocked', stocked_at=$1 WHERE id=$2`, now, input.BatchID)
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *SegmentFactoryHandler) GetStockEntry(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var ringNum, segNum int
	err := h.db.QueryRow(`SELECT ring_number, segment_number FROM segment_stock WHERE id = $1`, id).Scan(&ringNum, &segNum)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "stock entry not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "ring_number": ringNum, "segment_number": segNum})
}

func (h *SegmentFactoryHandler) UpdateStockEntry(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var input struct {
		Location *string `json:"location"`
		Notes    *string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	_, err := h.db.Exec(`UPDATE segment_stock SET location=COALESCE($1,location), notes=COALESCE($2,notes) WHERE id=$3`, input.Location, input.Notes, id)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *SegmentFactoryHandler) ShipStock(w http.ResponseWriter, r *http.Request) {
	var input struct {
		IDs         []string `json:"ids"`
		Destination string   `json:"destination"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	now := time.Now()
	for _, sid := range input.IDs {
		h.db.Exec(`UPDATE segment_stock SET status='shipped', destination=$1, shipped_date=$2 WHERE id=$3`, input.Destination, now, sid)
	}
	respondJSON(w, http.StatusOK, map[string]string{"shipped": fmt.Sprintf("%d segments", len(input.IDs))})
}

// =============================================================================
// QC Records
// =============================================================================
func (h *SegmentFactoryHandler) ListQCRecords(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT id, project_id, batch_id, ring_number, segment_number, check_type, result, deviation_pct, defect_type, defect_severity, checked_by, checked_at, created_at FROM segment_qc_records WHERE 1=1`
	args := []interface{}{}
	argIdx := 1
	if projectID != "" { query += fmt.Sprintf(" AND project_id = $%d", argIdx); argIdx++; args = append(args, projectID) }
	query += " ORDER BY checked_at DESC"

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()
	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, pid, checkType, result, defectType, defectSev, checkedBy string
		var batchID sql.NullString
		var ringNum, segNum int
		var devPct float64
		var checkedAt, crAt time.Time
		if err := rows.Scan(&id, &pid, &batchID, &ringNum, &segNum, &checkType, &result, &devPct, &defectType, &defectSev, &checkedBy, &checkedAt, &crAt); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		items = append(items, map[string]interface{}{
			"id": id, "project_id": pid, "batch_id": batchID, "ring_number": ringNum,
			"segment_number": segNum, "check_type": checkType, "result": result,
			"deviation_pct": devPct, "defect_type": defectType, "defect_severity": defectSev,
			"checked_by": checkedBy, "checked_at": checkedAt, "created_at": crAt,
		})
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *SegmentFactoryHandler) CreateQCRecord(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProjectID     string  `json:"project_id"`
		BatchID       string  `json:"batch_id"`
		RingNumber    int     `json:"ring_number"`
		SegmentNumber int     `json:"segment_number"`
		CheckType     string  `json:"check_type"`
		Result        string  `json:"result"`
		DeviationPct  *float64 `json:"deviation_pct"`
		DefectType    *string `json:"defect_type"`
		DefectSev     *string `json:"defect_severity"`
		CheckedBy     string  `json:"checked_by"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body"); return
	}
	id := uuid.New().String()
	now := time.Now()
	_, err := h.db.Exec(`INSERT INTO segment_qc_records (id, project_id, batch_id, ring_number, segment_number, check_type, result, deviation_pct, defect_type, defect_severity, checked_by, checked_at, created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$12)`,
		id, input.ProjectID, input.BatchID, input.RingNumber, input.SegmentNumber, input.CheckType, input.Result, input.DeviationPct, input.DefectType, input.DefectSev, input.CheckedBy, now)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *SegmentFactoryHandler) GetQCRecord(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var checkType, result string
	err := h.db.QueryRow(`SELECT check_type, result FROM segment_qc_records WHERE id = $1`, id).Scan(&checkType, &result)
	if err == sql.ErrNoRows { respondError(w, http.StatusNotFound, "QC record not found"); return }
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	respondJSON(w, http.StatusOK, map[string]interface{}{"id": id, "check_type": checkType, "result": result})
}

// =============================================================================
// Summary
// =============================================================================
func (h *SegmentFactoryHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	projectID := r.URL.Query().Get("project_id")
	query := `SELECT project_id, active_lines, active_plans, total_planned_rings, total_produced_rings, qc_passed, qc_failed, scrapped, in_stock, shipped FROM segment_factory_summary`
	args := []interface{}{}
	argIdx := 1
	if projectID != "" { query += fmt.Sprintf(" WHERE project_id = $%d", argIdx); args = append(args, projectID) }

	rows, err := h.db.Query(query, args...)
	if err != nil { respondError(w, http.StatusInternalServerError, err.Error()); return }
	defer rows.Close()
	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var pid string
		var lines, plans, planned, produced, passed, failed, scrapped, stock, shipped int
		if err := rows.Scan(&pid, &lines, &plans, &planned, &produced, &passed, &failed, &scrapped, &stock, &shipped); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error()); return
		}
		items = append(items, map[string]interface{}{
			"project_id": pid, "active_lines": lines, "active_plans": plans,
			"total_planned_rings": planned, "total_produced_rings": produced,
			"qc_passed": passed, "qc_failed": failed, "scrapped": scrapped,
			"in_stock": stock, "shipped": shipped,
		})
	}
	respondJSON(w, http.StatusOK, items)
}