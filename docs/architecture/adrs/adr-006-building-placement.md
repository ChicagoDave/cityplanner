# ADR-006: Building Placement and Procedural Generation

**Status:** Proposed
**Date:** 2026-02-09
**Deciders:** David Cornelson

---

## Context

Once pods are laid out (ADR-005), buildings must be placed within each pod's boundary. The technical spec calls for procedural generation within pod boundaries respecting the bowl height envelope (4 stories at edge, 10 in middle, 20 at center). Buildings must accommodate the required housing units, commercial space, and service facilities for each pod while maintaining walkways, plazas, and green space between them.

The city has no roads, so the constraints are different from conventional urban planning — there's no street grid to align to, but there must be a connected pedestrian/bicycle network.

## Decision Drivers

- Building heights must follow the bowl envelope: max height decreases with distance from city center
- Each pod must house its assigned population in the correct unit mix (studio through 4-bed)
- Commercial and service spaces (grocery, clinic, school, etc.) must be placed per pod requirements
- Pedestrian/bicycle paths must connect all buildings within a pod and between adjacent pods
- No roads — but freight elevators and emergency access from underground must reach buildings
- The surface should feel human-scale with adequate green space, plazas, and sightlines
- Generation must be deterministic for a given spec (same spec = same city)

## Options

### Option A: Grid Subdivision

Divide each pod's polygon into a regular grid of lots. Assign building types to lots based on pod requirements. Height is determined by lot position relative to city center.

**Pros:**
- Simple to implement — rectangular lots, predictable placement
- Easy to calculate floor area and unit counts
- Straightforward path routing along grid lines
- Deterministic by construction

**Cons:**
- Produces monotonous, repetitive layouts
- Grid doesn't adapt well to irregular pod polygons (Voronoi cells)
- Feels like a spreadsheet, not a city
- Wastes space where grid doesn't align with pod boundary

### Option B: Procedural Block-and-Lot Generation

Generate city blocks within each pod by creating a pedestrian street network first (organic or semi-regular), then subdividing blocks into building lots. Place buildings on lots according to type, height envelope, and density requirements.

**Pros:**
- Produces varied, organic-feeling layouts
- Street network is generated first, guaranteeing connectivity
- Block shapes adapt to pod polygon boundaries
- Different block patterns per ring create distinct neighborhood character

**Cons:**
- More complex to implement
- Must carefully control density to hit population targets
- Street network generation is itself a non-trivial algorithm
- Harder to guarantee exact unit counts without iteration

### Option C: Space-Filling with Placement Rules

Define building footprint templates (small residential, large residential, commercial, civic) with size and spacing rules. Place buildings one at a time using a greedy or constraint-based placement algorithm that maintains minimum spacing, path connectivity, and service distribution.

**Pros:**
- Flexible — handles any pod shape
- Can optimize for specific goals (maximize green space, minimize walking distance to services)
- Natural variation in layout
- Can enforce fine-grained rules (e.g., "grocery must be central in pod")

**Cons:**
- Greedy placement may produce poor layouts or fail to place all required buildings
- Order-dependent — placement sequence affects outcome
- Harder to guarantee connected path network until all buildings are placed
- Performance may be poor for large pods with many buildings

### Option D: Hierarchical Decomposition

Decompose each pod into functional zones (residential core, commercial spine, green space, civic anchor) using proportional area allocation. Within each zone, apply zone-appropriate placement logic (e.g., residential gets a courtyard block pattern, commercial gets a linear street pattern, civic gets a plaza-centered layout).

**Pros:**
- Produces legible, functional neighborhoods with clear structure
- Zone-level allocation ensures required areas are met before building placement
- Different zone patterns create variety within each pod
- Hierarchical approach is tractable: solve zones first, then fill zones
- Naturally places services where they belong (commercial on the spine, school at the edge of residential)

**Cons:**
- Zone boundary design is an additional algorithmic step
- Transitions between zones need careful handling
- More opinionated — less variation between pods of the same ring type

## Recommendation

**Option D: Hierarchical Decomposition.**

This approach matches how the spec thinks about pods — each has a character, required services, and population target. Decomposing into zones first (residential, commercial, civic, green) ensures the macro layout is functional before placing individual buildings. Within zones, simpler placement algorithms (grid or courtyard patterns for residential, linear for commercial) are tractable and produce legible results.

### Generation Pipeline

1. **Zone allocation:** Divide pod polygon into zones based on pod character and requirements. Residential gets the largest share; commercial and civic zones are positioned for accessibility (e.g., near pod center or along inter-pod paths).
2. **Path network:** Generate the pedestrian/bicycle network connecting zones within the pod and connecting to adjacent pods. This defines the circulation spine.
3. **Block subdivision:** Within each zone, generate blocks bounded by paths. Block shapes vary by zone type.
4. **Building placement:** Place buildings on blocks respecting height envelope, floor area requirements, and spacing rules.
5. **Validation:** Confirm total dwelling units, commercial square footage, and service facilities meet pod targets. Adjust if needed.

### Height Envelope

```
max_stories(distance_from_center) =
  if distance ≤ 300m:  20
  if distance ≤ 600m:  linear interpolation 20 → 10
  if distance ≤ 900m:  linear interpolation 10 → 4
```

Each building's max height is determined by the distance from its footprint centroid to the city center, interpolated within ring bands for a smooth bowl profile.

## Consequences

- The building placement system depends on pod layout (ADR-005) being complete
- Path network generation is a sub-problem that produces connectivity for the entire surface level
- Building templates/footprints must be defined as reusable geometry (parameterized by width, depth, height)
- The generation must be seeded-deterministic: same spec + same seed = same layout
- Underground access points (freight elevators, emergency access) must align with building positions — this creates a dependency with infrastructure routing (ADR-007)
