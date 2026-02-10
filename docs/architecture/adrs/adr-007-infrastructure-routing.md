# ADR-007: Infrastructure Network Routing

**Status:** Proposed
**Date:** 2026-02-09
**Deciders:** David Cornelson

---

## Context

The underground level contains three vertically separated layers:

1. **Bottom:** Sewage and water mains (gravity-dependent)
2. **Middle:** Utility corridors (electrical, telecom)
3. **Top:** Vehicle lanes (maximum height clearance)

Each layer contains one or more networks that must be routed from perimeter entry points to every pod and building. The spec defines arterial widths (6m for vehicle lanes), service branch widths (4m), utility corridor widths (2.5m), and capacity requirements (e.g., water at 100 gpd per person, electrical peak at 2.5 kW per capita).

Infrastructure routing is the most computationally intensive part of the solver and the most constrained — networks must avoid each other within a layer, connect to the correct endpoints, and satisfy capacity constraints.

## Decision Drivers

- Multiple independent networks must coexist in three layers without conflicts
- Capacity constraints: pipes and lanes must be sized for the load they carry
- Sewage requires gravity flow — routing must respect grade toward perimeter
- Vehicle lanes need connected routes from perimeter freight staging to every pod
- Utility corridors run parallel to vehicle lanes (per spec) with maintenance access points
- The solver must detect and report infeasible routings
- Routing must be deterministic for a given spec

## Options

### Option A: Template-Based Radial Routing

Define routing templates based on the concentric ring model: radial arterials from perimeter to center, ring connectors at each ring boundary. All networks follow the same backbone with layer separation.

**Pros:**
- Simple, predictable, and fast to compute
- Natural match for the concentric ring city layout
- Easy to validate capacity (arterials carry ring-to-ring flow, branches carry pod-level flow)
- Deterministic by construction

**Cons:**
- Inflexible — doesn't adapt to irregular footprints or uneven pod distribution
- May over-provision some routes and under-provision others
- Doesn't optimize for shortest path or minimum material
- All networks share the same routing geometry — hard to deconflict where space is tight

### Option B: Graph-Based Network Flow

Model the underground as a graph. Nodes are pod centers, perimeter entry points, and building access points. Edges are potential corridor segments. Use minimum-cost network flow algorithms to route each system (water, sewage, electrical, vehicle) with capacity constraints.

**Pros:**
- Optimal or near-optimal routing for each network
- Capacity constraints are first-class — the algorithm sizes pipes and lanes
- Can handle irregular layouts and uneven demand
- Well-studied algorithmic foundation (network flow, min-cost max-flow)

**Cons:**
- Computationally expensive — multiple networks, each with capacity constraints
- Networks must be routed sequentially or jointly to avoid conflicts
- Graph construction itself is non-trivial (what are the candidate edges?)
- May produce non-intuitive routes that are hard to validate visually

### Option C: Hierarchical Trunk-and-Branch

Route each network in two phases. Phase 1: route trunk lines from perimeter to ring junctions using the radial/ring backbone (like Option A). Phase 2: route branch lines from ring junctions to individual pods and buildings using shortest-path algorithms within each ring sector.

**Pros:**
- Combines predictability of templates (trunk) with optimization (branches)
- Trunk routing is simple and fast; branch routing is localized and parallelizable
- Natural capacity hierarchy: trunks are sized for aggregate flow, branches for local
- Easy to reason about: "water enters at perimeter, flows radially inward, branches to pods"

**Cons:**
- Trunk template may not be optimal for all city shapes
- Junction points between trunk and branch are design decisions, not computed
- Still requires conflict detection between networks at the branch level

### Option D: Constraint Satisfaction with Spatial Grid

Discretize the underground into a 3D voxel grid (layer × X × Y). Each network requests paths through the grid. A constraint solver (backtracking, SAT, or CP) assigns grid cells to networks ensuring no two networks occupy the same cell (within a layer) and all capacity/connectivity requirements are met.

**Pros:**
- Guaranteed conflict-free if a solution exists
- Handles all constraints simultaneously
- Can detect infeasibility definitively
- Maximum flexibility — no assumptions about routing topology

**Cons:**
- Computationally expensive — potentially NP-hard for complex constraints
- Grid resolution creates a space/accuracy tradeoff
- Produces grid-aligned routes (staircase artifacts unless post-smoothed)
- Difficult to scale to full city size at useful resolution

## Recommendation

**Option C: Hierarchical Trunk-and-Branch.**

This approach matches the city's physical structure. The concentric ring layout naturally defines a trunk network (radial arterials + ring connectors) that carries aggregate flow. Branch routing within each ring sector is a smaller, parallelizable problem. This is tractable, predictable, and produces infrastructure layouts that humans can understand and validate.

### Routing Order and Layer Assignment

Networks are routed in dependency order:

1. **Sewage (Layer 1, bottom):** Routed first because gravity flow is the hardest constraint. Trunk lines follow radial paths with calculated grade toward perimeter treatment plant. Branches connect pods with gravity-feasible slopes.
2. **Water (Layer 1, bottom):** Routed alongside sewage in the same layer. Pressurized, so routing is more flexible. Follows the same trunk backbone but on the opposite side of the corridor.
3. **Electrical (Layer 2, middle):** Trunk from perimeter substation radially inward. Branches to pod-level transformer rooms. Sized for peak demand.
4. **Telecom (Layer 2, middle):** Fiber backbone follows electrical trunk. Node breakouts every 75m per spec.
5. **Vehicle lanes (Layer 3, top):** Arterial lanes (6m) radially from perimeter freight staging. Service branches (4m) to each pod. Must accommodate two-way autonomous vehicle traffic.

### Capacity Validation

After routing, each network segment is validated:
- Water trunk must carry `downstream_population × 100 gpd`
- Sewage trunk must carry `downstream_population × 95 gpd` at gravity-feasible slope
- Electrical trunk must carry `downstream_population × 2.5 kW` peak
- Vehicle lanes must handle `estimated_daily_trips / hours / lanes` without congestion

## Consequences

- The routing algorithm depends on pod layout (ADR-005) for destination points
- Building placement (ADR-006) must provide underground access point locations
- Trunk routing geometry is shared across networks — a common "backbone" data structure
- Branch routing is per-pod and parallelizable across pods
- Infeasible routings (e.g., sewage grade impossible) must produce clear error messages pointing to the conflicting constraints
- The routing output feeds directly into the scene graph (ADR-008) as positioned, sized geometry
