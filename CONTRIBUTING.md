# Contributing to OpenConstructionERP

## Ground rules

1. **Architecture is governed by ADRs.** Read [SAD Tom 1](docs/sad/SAD-Tom1-Architecture-v1.0.md) and
   [docs/adr/](docs/adr/README.md) before proposing structural changes. To change an accepted decision,
   submit a superseding ADR as a PR — do not edit accepted ADRs.
2. **Domain accuracy over code elegance.** This platform is built by construction practitioners.
   If you change tunneling, scheduling, contract (FIDIC) or cost logic, cite the engineering/contractual
   basis in the PR description.
3. **Migrations are append-only.** Never edit an applied `V###__*.sql`; add a new one. CI installs the
   full chain from scratch on every PR — it must always pass.
4. **Conventions:** UUID v7 PKs; every project-scoped table carries `project_id` first in composite
   indexes; money = `NUMERIC(18,2)` + `currency CHAR(3)`; no soft deletes in legal domains (doc/contract/finance).

## Workflow

- Fork -> feature branch (`feat/...`, `fix/...`, `adr/...`) -> PR to `main`
- PR template: what/why/how tested; link the module code from SAD §4 (e.g. `L-03 Ring Register`)
- CI must be green: Python compile check + full migration chain install

## Development environment

```bash
docker compose -f infrastructure/docker/docker-compose.dev.yml up -d
python3 scripts/seed_reference_project.py   # seeds ALM-L3-REF: full reference project (BOQ $101.5M, 1000+ rings, reports, docs, costs)
```

## Licensing

The project is AGPL-3.0 today; an open-core split is proposed (ADR-016). Until that decision is signed
off, external contributors may be asked to sign a CLA so relicensing remains legally possible.
