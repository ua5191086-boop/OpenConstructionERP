# Module V022: Ring Builder & Segment Tracking — Design, Production, QC, Convergence

## Description
Implement the Ring Builder & Segment Tracking module (V022). Manages the full lifecycle of tunnel lining segments from ring design through production, curing, transport, installation, QC, inventory, and convergence monitoring.

## Tables (SQL)
- `ring_designs` — ring geometry, segment mapping, concrete grade, reinforcement
- `segment_production` — full lifecycle tracking: cast → curing → demolded → transport → installed
- `segment_curing` — steam/water/air curing stages with temperature/humidity
- `segment_transport` — logistics: mode, vehicle, driver, damage reporting
- `segment_installation` — erector cycle time, bolt torque, gaps, offsets
- `segment_qc` — dimensional checks, surface defects, compressive strength, cover
- `segment_inventory` — planned/produced/passed/installed/defective tracking
- `ring_measurements` — convergence, ovality, deformation, settlement

## Topics
- Ring design catalog and revision control
- Segment production tracking (casting yard to tunnel face)
- Curing process monitoring with temperature gradients
- QC inspection workflow (pass/conditional/fail/rework)
- Inventory management with auto-calculated stock
- Convergence and ovality monitoring for long-term deformation

## Files
- `database/migrations/V022__Ring_Builder_Segment_Tracking.sql`
- `scripts/generate_ringbuilder.py`
- `apps/web/ringbuilder-dashboard.html`
- `services/core/internal/handlers/ringbuilder.go`
- `apps/frontend/src/pages/RingBuilderPage.tsx`
- `apps/web/ringbuilder_data.json`

## Acceptance Criteria
- [ ] SQL migration creates all 8 tables with indexes and ontology entries
- [ ] Python generator produces 300+ segments with realistic production timeline
- [ ] HTML dashboard (dark theme, Chart.js) shows: segment flow, QC pass rate, inventory, convergence charts
- [ ] Go API handler with CRUD endpoints + summary
- [ ] React page with Recharts: production funnel, QC pie, inventory, convergence
- [ ] Registered routes in `main.go`
- [ ] Frontend integration: `App.tsx`, `Layout.tsx`, `api.ts`, `types.ts`

## Labels
`module`, `tunnel`, `ring`, `segment`, `V022`