# Session Summary: 2026-02-10 — Economic Model & Phase 2/3 Plan

## Status: Complete

## Goals
1. Design and document the city's economic model (no-ownership, licensed occupation, retirement fund)
2. Plan Phase 2 (bike & shuttle paths) and Phase 3 (sports fields) implementation

## Results

### Economic Model Document
Created `docs/specification/economic-model.md` covering the financial sustainability of the no-ownership charter city model.

**Key design decisions:**

- **Licensed occupation**: Residents hold occupation licenses, not property. City retains full planning authority, eliminates speculation, enables frictionless mobility between unit sizes.
- **License fee structure**: Single bundled monthly fee internally allocated across four funds — operating (40%), maintenance reserve (25%), debt service (20%), retirement fund (15%).
- **Maintenance inflation solution**: License fees alone can't sustain rising maintenance costs. External revenue streams compensate: energy export (solar surplus), commercial licensing (revenue-scaled fees), city equity partnerships, sovereign wealth fund, compute/data services, tourism.
- **Retirement fund**: Replaces property equity as the resident wealth-building mechanism. Pooled actuarial fund providing lifetime elder care (independent living → assisted living → nursing → memory care → hospice). Vesting schedule: 0-4yr none, 5-9yr 25%, 10-14yr 50%, 15-19yr 75%, 20+ yr 100% coverage. Acts as retention mechanism — leaving forfeits unvested portion.
- **Actuarial advantages**: City owns care facilities (no third-party markup), controls built environment (reduces care needs through accessibility), no adverse selection (universal participation), preventive care incentive, demographic management via admission policy, 20-30 year runway before first major claims.
- **Commercial model**: Revenue-scaled fees (not fixed rent), curated commercial mix per pod, equity participation for larger ventures.

### Phase 2/3 Implementation Plan
Created `docs/work/phase-2-refactor.md` with detailed 8-step implementation plan for:

**Phase 2 — Bike & Shuttle Paths:**
- Catmull-Rom spline primitives in `geo/spline.go` (Polyline type with Length/PointAt/NearestPoint/Offset)
- Ring corridor bike paths (closed loops through inter-pod green space with organic waypoint placement)
- Radial bike paths (center-to-edge S-curves through green corridors + countryside extensions)
- Shuttle routes co-located with bike paths via polyline offset (shared infrastructure ribbon)
- One station per pod as combined shuttle + bike mobility hub

**Phase 3 — Sports Fields:**
- Buffer zone identification between adjacent pods
- Stadium placement near ring3/4 boundary (110x75m)
- 10 soccer/cricket fields distributed to outer rings (105x68m)
- Small courts (basketball, tennis, pickleball) greedy-packed in remaining buffers

**Integration:**
- 4 new entity types + 1 system type in scene graph
- Extended `Assemble` signature and 4 new assembly functions
- 3 new generation calls in `runSolve` pipeline

## Files Created
| File | Purpose |
|------|---------|
| `docs/specification/economic-model.md` | Licensed occupation, revenue streams, retirement fund, commercial model |
| `docs/work/phase-2-refactor.md` | Implementation plan for bike paths, shuttle routes, sports fields |

## Next Steps
1. Implement Step 1: Catmull-Rom splines + Polyline type in `geo/spline.go`
2. Implement Step 2: New entity/system type constants in `scene/scene.go`
3. Implement Steps 3-8: Bike paths, shuttle routes, sports fields, assembly, pipeline, tests
