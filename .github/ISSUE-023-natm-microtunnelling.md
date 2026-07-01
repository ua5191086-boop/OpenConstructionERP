# Module V023: NATM & Microtunnelling — Sequential Excavation, Shotcrete, MTBM, Shafts, Settlement

## Description
Implement the NATM & Microtunnelling module (V023). Covers New Austrian Tunnelling Method (NATM) with sequential excavation, shotcrete, rock bolts, steel sets, face mapping, convergence; and microtunnelling (MTBM) with pipe jacking, thrust, lubrication, survey; plus shafts, cross passages, grouting, and settlement monitoring.

## Tables (SQL)
### NATM
- `natm_excavation_log` — round-based excavation (chainage, method, geotech class, standup time)
- `natm_shotcrete` — sprayed concrete (dry/wet/steel fiber, thickness, strength, rebound)
- `natm_rock_bolts` — expansion/resin/swellex/grouted bolts with pullout tests
- `natm_steel_sets` — TH/HEB/lattice girders with spacing and grade
- `natm_convergence` — displacement monitoring with rates and alarms
- `natm_face_mapping` — geological logging (RMR, Q-value, GSI, joints, water)

### Microtunnelling
- `mtbm_drives` — pipe jacking drive configuration
- `mtbm_thrust_log` — per-pipe thrust, torque, slurry, alignment
- `mtbm_lubrication` — bentonite/polymer injection parameters
- `mtbm_survey` — as-built position, deviation from design

### Shafts & Ancillary
- `shaft_construction` — launch/reception/ventilation shafts
- `shaft_equipment` — crane, fans, pumps, control panels
- `cross_passages` — connecting passages between tunnels
- `grouting_records` — contact/void/consolidation/curtain grouting
- `settlement_monitoring` — surface/subsurface/building monitoring with alarms

## Topics
- NATM sequential excavation cycle management
- Shotcrete application and quality control
- Rock reinforcement (systematic/spot bolting)
- Geological face mapping and ground classification
- MTBM pipe jacking thrust and lubrication optimization
- Shaft construction stage tracking
- Grouting program management
- Settlement monitoring with alarm thresholds

## Files
- `database/migrations/V023__NATM_Microtunnelling.sql`
- `scripts/generate_natm.py`
- `apps/web/natm-dashboard.html`
- `services/core/internal/handlers/natm.go`
- `apps/frontend/src/pages/NATMPage.tsx`
- `apps/web/natm_data.json`

## Acceptance Criteria
- [ ] SQL migration creates all 15 tables with indexes and ontology entries
- [ ] Python generator produces 80+ excavation rounds, shotcrete records, bolts, MTBM pipes
- [ ] HTML dashboard (dark theme, Chart.js) shows: excavation progress, shotcrete/bolt stats, convergence, MTBM thrust, settlement
- [ ] Go API handler with CRUD endpoints + summary for all entity types
- [ ] React page with Recharts: excavation timeline, face mapping, MTBM thrust, settlement map
- [ ] Registered routes in `main.go`
- [ ] Frontend integration: `App.tsx`, `Layout.tsx`, `api.ts`, `types.ts`

## Labels
`module`, `tunnel`, `natm`, `microtunnelling`, `V023`