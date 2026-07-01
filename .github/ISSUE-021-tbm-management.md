# Module V021: TBM Management — Real-Time Telemetry, Alarms, & Performance

## Description
Implement the TBM Management module (V021) for real-time tunnel boring machine monitoring. This extends the existing V004 Tunnel Module with detailed operational tracking.

## Tables (SQL)
- `tbm_telemetry` — EPB/slurry parameters, thrust, torque, advance rate, face pressure, cutterhead, tail skin
- `tbm_alarms` — alarm codes, severity, acknowledgment, clearing
- `tbm_operators` — certified operators with qualifications
- `tbm_shifts` — shift logs with rings built, downtime
- `tbm_consumables` — cutterhead, seals, foam, bentonite, grease tracking
- `tbm_performance_metrics` — daily/shift aggregates: utilization, advance rate, torque, availability

## Topics
- Real-time telemetry ingestion (PLC/data source)
- Alarm management and notification
- Operator logs and shift handover
- Consumables usage tracking
- Performance analysis (OPE, utilization, advance rate trends)

## Files
- `database/migrations/V021__TBM_Management.sql`
- `scripts/generate_tbm.py`
- `apps/web/tbm-dashboard.html`
- `services/core/internal/handlers/tbm.go`
- `apps/frontend/src/pages/TBMPage.tsx`
- `apps/web/tbm_data.json`

## Acceptance Criteria
- [ ] SQL migration creates all 6 tables with indexes and ontology entries
- [ ] Python generator produces realistic test data (200+ telemetry points, 20+ alarms, operators, shifts)
- [ ] HTML dashboard (dark theme, Chart.js) shows: real-time gauges, alarm list, performance charts, consumables
- [ ] Go API handler with CRUD endpoints for all tables + summary
- [ ] React page with Recharts: telemetry chart, alarm pie, shift performance
- [ ] Registered routes in `main.go`
- [ ] Frontend integration: `App.tsx`, `Layout.tsx`, `api.ts`, `types.ts`

## Labels
`module`, `tunnel`, `tbm`, `V021`