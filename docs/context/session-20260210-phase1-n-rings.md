# Session Summary: 2026-02-10 — Phase 1: Generalize Solver for N Rings

## Status: Complete

## Goal
Implement Phase 1 of the ground-level 2D renderer plan: generalize the solver from hardcoded 3 rings (center/middle/edge) to support N rings, with a 5-ring, 64K-population default configuration.

## Results
- **All tests pass**: `go test ./...` — all packages OK
- **Validation**: `cityplanner validate` reports VALID (12 warnings about service thresholds)
- **Solver output**: 5,274 entities, 32 pods across 5 rings, spec version 0.2.0
- **Pod distribution**: center=1, ring4=2, ring3=3, ring2=7, ring1=19 (total 32, geometry-driven)
- **Population gradient**: center=3,091/pod → ring1=1,227/pod (capacity-weighted)

## Completed

### 1. Refactored `solver/pkg/spec/types.go`
- Replaced fixed `CityZones` struct (`Center`, `Middle`, `Edge` fields) with dynamic `Rings []RingDef`
- Added `RingDef` type with `Name`, `Character`, `RadiusFrom`, `RadiusTo`, `MaxStories`
- Added helper methods: `OuterRadius()`, `RingByName()`
- Kept `Perimeter` and `SolarRing` as separate fields

### 2. Updated `examples/default-city/city.yaml`
- Population: 50,000 → 64,000
- Spec version: 0.1.0 → 0.2.0
- 5 rings: center (0-250m, 32 stories), ring4 (250-500m, 16), ring3 (500-850m, 8), ring2 (850-1350m, 4), ring1 (1350-2200m, 2)
- Perimeter: 2200-2700m, Solar: 2700-3200m
- Updated ring assignments with services for all 5 rings
- Scaled infrastructure numbers for 64K (battery 3840 MWh, grid 200 MW, etc.)

### 3. Updated `shared/schema/city-spec.schema.json`
- Replaced fixed `center/middle/edge` properties with `rings` array
- Added `ring` definition with required fields: `name`, `radius_from`, `radius_to`, `max_stories`

### 4. Generalized `solver/pkg/analytics/geometry.go`
- `resolveRings()` now iterates over `s.CityZones.Rings` (any count)
- Two-pass algorithm: first computes pod counts per ring (geometry-based), then assigns uniform per-pod population
- Population model changed from area-proportional to **uniform per pod** (2,000 each)
- `resolveAreas()` uses `s.CityZones.OuterRadius()` instead of hardcoded `Edge.RadiusTo`

### 5. Updated `solver/pkg/layout/envelope.go`
- New `MaxStoriesFromRings(dist, []RingDef)` function — ring-based height lookup (flat per ring, interpolation only in gaps)
- Legacy `MaxStories(dist, center, middle, edge)` preserved for backward compatibility, delegates to `MaxStoriesFromRings`

### 6. Updated `solver/pkg/layout/buildings.go`
- Replaced hardcoded `ringRadii` map with dynamic lookup from `s.CityZones.Rings`
- All building placement functions now accept `[]spec.RingDef` instead of `maxCenter/maxMiddle/maxEdge`
- Uses `MaxStoriesFromRings()` for all height computations

### 7. Updated `solver/pkg/layout/zones.go`
- Added zone proportion entries for new ring characters: `high_density`, `urban_midrise`, `mixed_residential`, `low_density`

### 8. Updated `solver/pkg/layout/green.go`
- Replaced hardcoded 3-ring `ringRadii` map with dynamic lookup from `s.CityZones.Rings`

### 9. Updated `solver/pkg/validation/schema.go`
- `validateZones()` now iterates over `s.CityZones.Rings` for radii and max_stories validation
- Continuity checks work for any adjacent ring pair
- Empty rings list produces a schema error

### 10. Updated `solver/pkg/analytics/validate.go`
- Spec path references updated: `city_zones.center.max_stories` → `city_zones.rings[0].max_stories`

### 11. Updated `solver/pkg/routing/routing.go`
- Replaced `s.CityZones.Edge.RadiusTo` with `s.CityZones.OuterRadius()`
- Replaced hardcoded `Center.RadiusTo/Middle.RadiusTo` ring radii with dynamic lookup from `s.CityZones.Rings`
- Fixed `max()` builtin (Go 1.21+) for Go 1.19 compatibility

### 12. Updated `solver/pkg/cost/cost.go`
- Replaced `s.CityZones.Edge.RadiusTo` with `s.CityZones.OuterRadius()`

### 13. Fixed `solver/go.mod`
- Go version `1.25.7` (invalid) → `1.19` (matches installed Go)

### 14. Updated all test files
- `analytics/analytics_test.go` — Updated `fullDefaultSpec()` to use `Rings` list; updated density error spec paths
- `analytics/geometry_test.go` — Updated `defaultSpec()` to use `Rings` list; added `TestResolveRingsUniformPodPopulation`
- `layout/pods_test.go` — Updated `defaultSpec()` and `defaultParams()` for new format
- `layout/buildings_test.go` — Updated `TestEnvelopeMaxStories` for flat per-ring behavior (no interpolation within contiguous rings)
- `scene/assemble_test.go` — Updated `testSpec()` and `testParams()` for new format
- `scene/bench_test.go` — Updated `specForPopulation()` for Rings list; replaced `b.Loop()` with `b.N` (Go 1.19)
- `spec/spec_test.go` — Updated for 5-ring, 64K city.yaml (spec version 0.2.0, ring count 5, battery 3840 MWh)
- `validation/schema_test.go` — Updated `validSpec()` for Rings list; updated zone gap/inverted/max_stories tests for indexed ring access
- `routing/routing_test.go` — Updated `defaultSpec()` for Rings list
- `cost/cost_test.go` — Updated `defaultCostSpec()` for Rings list

## Architecture Decisions

### Capacity-Weighted Pod Population
Population distributed proportionally to each ring's residential capacity: `weight = area × max_stories × residential_fraction(character)`. Inner rings with tall buildings and civic/commercial character get more people per pod (singles/couples, avg HH 1.8). Outer rings with low-rise family housing get fewer (avg HH 3.5). Total households derived from per-ring averages, not city-wide.

### Flat Per-Ring Height Envelope
Changed from continuous interpolation across the city radius to flat max_stories per ring. Each ring defines a fixed building height; interpolation only applies in gaps between non-contiguous rings (which shouldn't normally exist).

### Ring List Format
YAML format changed from named map keys to ordered list:
```yaml
# Before (0.1.0):
city_zones:
  center: { radius_from: 0, radius_to: 300 }
  middle: { radius_from: 300, radius_to: 600 }

# After (0.2.0):
city_zones:
  rings:
    - name: center
      radius_from: 0
      radius_to: 250
    - name: ring4
      radius_from: 250
      radius_to: 500
```

## Files Modified
| File | Change |
|------|--------|
| `solver/go.mod` | Go version 1.25.7 → 1.19 |
| `solver/pkg/spec/types.go` | `CityZones` → dynamic `Rings []RingDef` |
| `examples/default-city/city.yaml` | 5-ring, 64K spec |
| `shared/schema/city-spec.schema.json` | Rings array schema |
| `solver/pkg/analytics/geometry.go` | N-ring resolution, uniform pop |
| `solver/pkg/analytics/validate.go` | Updated spec paths |
| `solver/pkg/layout/envelope.go` | `MaxStoriesFromRings()` |
| `solver/pkg/layout/buildings.go` | Dynamic ring radii |
| `solver/pkg/layout/zones.go` | New character proportions |
| `solver/pkg/layout/green.go` | Dynamic ring radii |
| `solver/pkg/validation/schema.go` | N-ring zone validation |
| `solver/pkg/routing/routing.go` | Dynamic ring radii, OuterRadius() |
| `solver/pkg/cost/cost.go` | OuterRadius() |
| `solver/pkg/analytics/analytics_test.go` | Updated for Rings list |
| `solver/pkg/analytics/geometry_test.go` | Updated for Rings list |
| `solver/pkg/layout/pods_test.go` | Updated for Rings list |
| `solver/pkg/layout/buildings_test.go` | Updated envelope test expectations |
| `solver/pkg/scene/assemble_test.go` | Updated for Rings list |
| `solver/pkg/scene/bench_test.go` | Updated for Rings list + Go 1.19 |
| `solver/pkg/spec/spec_test.go` | Updated for 5-ring 64K spec |
| `solver/pkg/validation/schema_test.go` | Updated for Rings list |
| `solver/pkg/routing/routing_test.go` | Updated for Rings list |
| `solver/pkg/cost/cost_test.go` | Updated for Rings list |

## Phase 1b: Capacity-Weighted Population (same session)

Replaced the uniform per-pod population model with a capacity-weighted model where inner ring pods have more people (dense, small apartments for singles/couples) and outer ring pods have fewer (family housing).

### Model
- `weight = ring_area × max_stories × residential_fraction(character)`
- Population distributed proportionally to weights
- Per-ring avg household size derived from character (1.8 for civic/dense, 3.5 for low-density family)

### Results
| Ring | Pods | Pop/Pod | Avg HH |
|------|------|---------|--------|
| center | 1 | 3,091 | 1.8 |
| ring4 | 2 | 5,022 | 1.8 |
| ring3 | 3 | 3,894 | 2.2 |
| ring2 | 7 | 2,266 | 3.0 |
| ring1 | 19 | 1,227 | 3.5 |

### Files Modified
| File | Change |
|------|--------|
| `solver/pkg/analytics/geometry.go` | `characterResidentialFraction()`, `characterAvgHouseholdSize()`, capacity-weighted `resolveRings()` |
| `solver/pkg/analytics/analytics.go` | Derive totalHH from ring sum, remove `math` import |
| `solver/pkg/analytics/types.go` | Add `AvgHouseholdSize` to `RingData` |
| `solver/pkg/analytics/geometry_test.go` | Capacity-weighted population test |
| `solver/pkg/analytics/analytics_test.go` | Widened household count assertion |
| `solver/pkg/layout/pods_test.go` | Updated `defaultParams()` |
| `solver/pkg/routing/routing_test.go` | Updated `defaultParams()` |
| `solver/pkg/scene/assemble_test.go` | Updated `testParams()` |

## Next Steps (Phase 2+)
1. Implement organic bike/shuttle paths (Phase 2 of ground-level plan)
2. Add sports fields and recreation areas (Phase 3)
3. Build 2D SVG renderer (Phase 4)
