# ADR-010: Cost Model Computation

**Status:** Proposed
**Date:** 2026-02-09
**Deciders:** David Cornelson

---

## Context

The technical spec includes detailed cost formulas: excavation ($35/m³), per-capita cost ($100K-160K), debt service calculations, and break-even rent analysis. Cost is a critical output of the design tool — it determines whether a given city configuration is economically viable. The spec's feedback loop explicitly includes cost as a constraint that feeds back into population and density decisions.

We need to decide when and how cost is computed, whether it acts as a constraint or just a report, and how it integrates with the two-phase solver (ADR-009).

## Decision Drivers

- Cost depends on nearly everything: land area, excavation depth, building count, infrastructure length, solar farm size, battery capacity
- The spec's feedback loop uses cost as a viability check
- Phase 1 (ADR-009) needs at least aggregate cost estimates for pre-flight validation
- Phase 2 produces exact quantities that enable precise bottom-up cost
- Phased construction (3 phases) requires cost to be separable by phase
- Users need to see how parameter changes affect cost in real-time

## Options

### Option A: Post-Solve Report Only

Cost is computed after the full solver pipeline completes. It reads the scene graph and sums up material quantities × unit costs. No feedback into the solver.

**Pros:**
- Simplest — cost is a pure read-only analysis of solver output
- Exact quantities from the scene graph produce accurate costs
- No coupling between cost logic and solver logic

**Cons:**
- User must run the full pipeline, see the cost, manually adjust, and re-run
- No early warning if the spec is heading toward infeasible economics
- Phased construction analysis requires post-processing the scene graph by phase

### Option B: Dual Computation — Estimate in Phase 1, Precise in Phase 2

Phase 1 computes an aggregate cost estimate using formulas from the spec (excavation volume × unit cost, building area × cost/m², etc.). Phase 2 computes precise bottom-up cost from actual generated quantities. Both are reported; discrepancy between them is a useful diagnostic.

**Pros:**
- Early cost feedback in Phase 1 (microseconds, before spatial generation)
- Precise cost from Phase 2 validates the estimate
- Phase 1 can flag specs that are obviously uneconomic before spending time on spatial generation
- Estimate vs. actual comparison helps calibrate the cost model over time

**Cons:**
- Two cost computation paths to maintain
- Discrepancy between estimate and actual may confuse users
- Phase 1 estimates may be inaccurate for unusual configurations

### Option C: Cost as a Hard Constraint

Cost targets are part of the spec (e.g., `max_per_capita_cost: 150000`). The solver treats this as a constraint and rejects configurations that exceed it. This requires Phase 1 to check cost feasibility and Phase 2 to validate actuals.

**Pros:**
- Prevents generating cities that are economically infeasible
- Forces the user to think about cost as a design parameter, not an afterthought
- Enables the solver to report "your population is too small to achieve your cost target at this density"

**Cons:**
- Overly rigid — users may want to see what a design costs before deciding if it's feasible
- Cost parameters (unit costs, interest rates) are estimates themselves — hard constraints on estimates feel false precision
- Adds coupling between cost model and solver logic

## Recommendation

**Option B: Dual Computation — Estimate in Phase 1, Precise in Phase 2.**

Phase 1 produces a cost estimate from aggregate quantities (area × unit costs). This serves as an early warning system. Phase 2 produces precise costs from actual generated geometry. Both are reported side by side.

Cost is **not** a hard constraint by default — it's a reported metric. Users can optionally set cost warning thresholds in the spec that produce warnings (not errors) when exceeded. This keeps the solver simple while giving users the economic feedback they need.

### Phase 1 Cost Model (Aggregate Estimate)

```
excavation_cost = city_area × depth × cost_per_m³
slab_cost = city_area × slab_cost_per_m²
building_cost = total_building_area × cost_per_m² (varies by type)
infrastructure_cost = estimated_network_length × cost_per_m (by system)
solar_cost = solar_area × cost_per_m²
battery_cost = battery_capacity_mwh × cost_per_mwh
total_construction = sum of above
annual_debt_service = amortize(total_construction, rate, term)
annual_operations = population × operations_per_capita
break_even_rent = (debt_service + operations) / units / 12
per_capita_cost = total_construction / population
```

### Phase 2 Cost Model (Bottom-Up Precise)

After spatial generation, the precise model sums:
- Actual excavation volume (accounting for terrain if non-flat)
- Actual building footprint area × stories × construction cost per floor-m²
- Actual pipe/conduit/lane lengths × diameter-dependent cost per meter
- Actual solar panel area from placed panels
- Actual battery storage from pod-level placement
- Site-specific costs (waterproofing, soil remediation, etc.)

### Phased Construction Costs

Both Phase 1 and Phase 2 cost models separate costs by construction phase:
- Phase 1 (center + inner middle): entities within 0-450m of center
- Phase 2 (outer middle): entities within 450-600m
- Phase 3 (edge): entities within 600-900m
- Perimeter + solar ring: allocated proportionally or to Phase 1

### Cost Output Structure

```typescript
interface CostReport {
  estimate: PhasedCost;      // from Phase 1
  actual?: PhasedCost;       // from Phase 2 (absent if Phase 2 hasn't run)
  comparison?: CostDelta;    // estimate vs actual

  summary: {
    total_construction: number;
    per_capita: number;
    annual_debt_service: number;
    annual_operations: number;
    break_even_monthly_rent: number;
  };
}

interface PhasedCost {
  phase_1: CostBreakdown;
  phase_2: CostBreakdown;
  phase_3: CostBreakdown;
  perimeter_and_solar: CostBreakdown;
  total: CostBreakdown;
}

interface CostBreakdown {
  excavation: number;
  structural: number;
  buildings: number;
  infrastructure: number;
  solar: number;
  battery: number;
  other: number;
  total: number;
}
```

## Consequences

- The cost model requires a unit cost table as part of the spec or as a separate configuration file
- Unit costs should be adjustable (regional variation, inflation, material choices)
- Phase 1 cost estimates enable rapid "what-if" exploration without running the full solver
- The cost report is included in the scene graph metadata or as a sidecar file
- A future "cost optimizer" could use Phase 1's analytical model to search for population/density combinations that minimize per-capita cost
