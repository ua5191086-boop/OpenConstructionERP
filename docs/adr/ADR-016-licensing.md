# ADR-016: Licensing model — open-core (Apache 2.0 core / AGPL-3.0 modules)

**Status:** PROPOSED — requires explicit Product Owner sign-off before first external contribution
**Current state:** entire repository is AGPL-3.0

## Context
AGPL protects against cloud vendors repackaging the platform, but blocks enterprise adoption:
banks and government customers (Central Asia, EU) routinely prohibit AGPL dependencies.
The SAD targets an open-core commercial model (managed SaaS + enterprise modules + support).

## Proposed decision
- Core platform (`services/core`, `packages/*`, base ontology): **Apache 2.0**
- Domain modules and enterprise features: **AGPL-3.0** or commercial license
- CLA (Contributor License Agreement) required from external contributors to keep dual-licensing possible

## Why decide now
Relicensing after external contributions requires consent of every contributor.
This decision is cheap today and legally expensive in a year.
