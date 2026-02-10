# ADR-009: Constraint Solver Strategy

**Status:** Proposed
**Date:** 2026-02-09
**Deciders:** David Cornelson

---

## Context

The technical spec describes a central feedback loop:

```
P_total → service thresholds → min pod size → density requirement
→ housing mix → land area → excavation cost → per-capita cost
→ economic viability → back to P_total
```

The solver must take a declarative spec and produce a fully resolved city layout. This requires satisfying many interrelated constraints simultaneously: population must support services, density must fit within land area, pods must meet service thresholds, infrastructure must have capacity, and the bowl height envelope must be respected. Some specs may be infeasible (e.g., 50,000 people in 100 hectares at 4 stories max).

This ADR decides the overall solving strategy — how the engine resolves interdependent constraints and what happens when they conflict.

## Decision Drivers

- Constraints are interdependent (population affects density affects land area affects cost)
- Some constraints are hard (walk radius, service thresholds) and some are soft (cost targets, green space ratios)
- The solver must terminate in seconds for interactive use
- Infeasible specs must be detected and reported clearly
- The solver should suggest adjustments when constraints conflict
- Deterministic: same spec = same output

## Options

### Option A: Fixed-Order Sequential Pipeline

Solve each concern in a predetermined order: demographics → pod layout → building placement → infrastructure routing → cost calculation. Each stage consumes outputs from prior stages. No backtracking.

**Pros:**
- Simplest to implement and debug
- Predictable execution — each stage runs once
- Fast — no iteration or search
- Easy to pinpoint which stage failed

**Cons:**
- No backtracking — if infrastructure routing fails because building placement left no room, the solver cannot adjust buildings
- Cannot optimize across stages (e.g., adjust pod sizes to minimize cost)
- The feedback loop in the spec is explicitly iterative — this approach ignores it
- May produce valid but suboptimal solutions

### Option B: Iterative Relaxation

Run the sequential pipeline, then validate the result against all constraints. If constraints are violated, adjust parameters and re-run. Repeat until all constraints are satisfied or a maximum iteration count is reached.

**Pros:**
- Handles the feedback loop explicitly
- Each iteration is a full sequential pipeline (reuses Option A's simplicity)
- Can converge on a valid solution when initial parameters are close
- Adjustment heuristics can be domain-specific (e.g., "if pod population too low, increase density")

**Cons:**
- May not converge — oscillation between competing constraints is possible
- Iteration count is unpredictable
- Adjustment heuristics must be hand-tuned per constraint type
- No guarantee of optimality — stops at first feasible solution

### Option C: Mathematical Optimization (LP/MIP)

Formulate the city design as a linear or mixed-integer program. Objective function: minimize cost (or maximize population, or optimize a composite metric). Constraints encode all spec requirements. Solve with an off-the-shelf solver (e.g., HiGHS, GLPK, or CBC via JavaScript bindings).

**Pros:**
- Guaranteed optimal solution (within the formulation)
- Handles constraint interdependencies natively
- Infeasibility is detected and reported by the solver
- Can optimize for user-specified objectives

**Cons:**
- Formulating spatial layout problems as LP/MIP is extremely difficult
- Building placement and infrastructure routing don't reduce to linear constraints
- LP/MIP solvers are powerful for numerical optimization but poor for geometric problems
- Adds a heavy dependency (optimization solver library)
- The problem has both continuous (density, areas) and discrete (building count, pod count) variables — MIP is NP-hard in general

### Option D: Two-Phase Solver — Analytical Resolution + Sequential Generation

**Phase 1 (Analytical):** Resolve the numerical feedback loop algebraically. Given population, compute demographics, service counts, pod count, density requirement, land area, and cost. Detect numerical infeasibility here (e.g., required density exceeds max height). This phase runs in microseconds.

**Phase 2 (Generative):** Using the resolved numbers from Phase 1, run the spatial generation pipeline sequentially: pod layout → building placement → infrastructure routing. Each stage uses the validated parameters from Phase 1 as hard targets. If a spatial stage fails (e.g., buildings can't fit), report the failure with the specific conflicting constraints.

**Pros:**
- Phase 1 catches most infeasibilities instantly — no wasted spatial computation
- Phase 1 is pure math — testable, fast, deterministic
- Phase 2 is a sequential pipeline operating on validated parameters — simpler than backtracking
- Clear separation: "are the numbers possible?" vs. "can we build it spatially?"
- Reports infeasibility at the right level: numerical conflicts from Phase 1, spatial conflicts from Phase 2

**Cons:**
- Phase 1 cannot catch all spatial infeasibilities (e.g., pods that won't pack into an irregular footprint)
- No cross-phase optimization (Phase 2 can't ask Phase 1 to adjust numbers)
- Phase 2 failures require manual spec adjustment rather than automatic resolution

## Recommendation

**Option D: Two-Phase Solver — Analytical Resolution + Sequential Generation.**

This matches the problem's natural structure. Most constraint violations are numerical (density too high, population too low for a hospital, cost exceeds target) and can be caught analytically without running expensive spatial algorithms. Phase 1 acts as a fast "pre-flight check" that resolves the feedback loop the spec describes.

Phase 2 only runs when the numbers check out, and its stages have validated targets to hit. This makes each spatial stage simpler because it knows its parameters are feasible in the aggregate.

### Phase 1: Analytical Resolution

Inputs: population, demographics, site area, height limits, walk radius, service thresholds.

Computed:
- Household count and cohort breakdown
- Dependency ratio (validate within target range)
- Service counts per type
- Pod count and target population per pod
- Required density (du/ha) per ring
- Total residential, commercial, civic, and green space area
- Excavation volume and cost
- Per-capita cost and average rent

Validation:
- Required density ≤ achievable density at max height per ring
- Pod population ≥ max service threshold (or service sharing is feasible)
- Total area ≤ site area
- Cost per capita within spec-defined viability range (if specified)

If any validation fails, Phase 1 reports the conflict with the specific values and which spec parameters would need to change to resolve it.

### Phase 2: Sequential Generation

Stages (each consuming prior stage output):
1. Pod layout (ADR-005)
2. Building placement (ADR-006)
3. Infrastructure routing (ADR-007)
4. Scene graph assembly (ADR-008)
5. Cost aggregation (ADR-010)

Each stage validates its own output. Failures are reported with spatial context (e.g., "Pod 7 in the edge ring cannot fit 3,200 residents at 4 stories within its 48-hectare boundary").

## Consequences

- Phase 1 is a pure function: spec in, resolved parameters + validation report out
- Phase 1 can run independently as a "spec check" CLI command (useful for rapid iteration on numbers)
- Phase 2 stages are independently testable with fixed inputs
- No automatic backtracking — the user adjusts the spec when constraints conflict
- This is simpler than optimization-based approaches but requires the user to be the optimizer
- A future "suggest fix" feature could use Phase 1's math to recommend parameter adjustments
