# ADR-005: Pod Layout Algorithm

**Status:** Proposed
**Date:** 2026-02-09
**Deciders:** David Cornelson

---

## Context

The city is divided into approximately 13 neighborhood pods, each centered on a 400-meter walk radius. Pods are assigned to concentric rings (center, middle, edge) with different character, density, and service profiles. The technical spec identifies circle packing and Voronoi tessellation as candidate algorithms.

The pod layout is the foundational spatial decision — everything else (building placement, infrastructure routing, service distribution) depends on where pods are and what their boundaries look like.

## Decision Drivers

- Every resident must be within 400m walking distance of their pod's essential services
- Pods must respect ring boundaries (center 0-300m, middle 300-600m, edge 600-900m)
- The city footprint may be circular, square, or irregular
- Pod boundaries determine walkway routing, infrastructure branching, and service placement
- Pods in different rings have different population targets (3,000-6,000)
- Adjacent pods may share services when individual population is below a threshold

## Options

### Option A: Circle Packing

Place pods as non-overlapping circles of radius 400m within the city footprint. Each circle defines the walk-radius catchment. The interstitial space between circles becomes shared or transitional.

**Pros:**
- Directly models the walk-radius constraint — every point in a circle is within 400m of center
- Well-studied algorithmic problem with known solutions
- Produces natural, organic-feeling layouts
- Easy to reason about: "Am I in this pod?" = distance check

**Cons:**
- Circles don't tile — significant interstitial space is wasted or ambiguous
- Hard to enforce ring boundaries cleanly (circles straddle ring edges)
- Uneven population distribution — circle areas are fixed regardless of density needs
- Irregular footprints produce awkward packing with many gaps

### Option B: Voronoi Tessellation

Place pod center points within the city footprint, then compute Voronoi cells. Each cell defines the pod's territory. The walk-radius constraint is validated post-tessellation.

**Pros:**
- Complete coverage — every point in the city belongs to exactly one pod (no gaps)
- Flexible shapes adapt to ring boundaries and irregular footprints
- Pod centers can be placed strategically (e.g., along ring midlines)
- Voronoi cells naturally define "nearest pod" — useful for service assignment
- Cell areas can vary to match different density/population targets per ring

**Cons:**
- Voronoi cells can be elongated — some points in a cell may exceed 400m from center
- Must constrain or adjust cells to satisfy the walk-radius guarantee
- Less intuitive shapes (polygons vs. circles) for human reasoning
- Requires post-processing to clip cells to city footprint boundary

### Option C: Hexagonal Grid

Overlay a hexagonal grid on the city footprint. Each hexagon is a pod. Hex radius is set to satisfy the 400m walk constraint.

**Pros:**
- Uniform, predictable layout — every pod is the same shape and size
- Hexagons tile without gaps (complete coverage)
- Equal distance to all six neighbors — good for service sharing
- Simple to implement and reason about

**Cons:**
- Rigid — cannot adapt pod size to different ring densities
- Poor fit for non-hexagonal city footprints (many partial hexes at boundary)
- Ring boundaries don't align with hex grid lines
- Feels artificial and mechanical rather than responsive to city geometry

### Option D: Constrained Voronoi with Ring-Anchored Seeds

Hybrid approach: place seed points along ring midlines at spacing that respects the 400m constraint. Compute Voronoi tessellation. Clip cells to ring boundaries so no cell crosses rings. Post-validate that every point in each cell is within 400m of the pod center. Adjust seed positions iteratively if validation fails.

**Pros:**
- Complete coverage with no gaps
- Respects ring boundaries by construction
- Walk-radius constraint is explicitly validated and enforced
- Pod sizes vary naturally by ring (smaller dense center pods, larger edge pods)
- Seed placement along ring midlines produces regular, readable layouts

**Cons:**
- Most complex to implement
- Iterative adjustment may not converge for tight constraints
- Clipping Voronoi cells to ring arcs produces irregular polygon shapes

## Recommendation

**Option D: Constrained Voronoi with Ring-Anchored Seeds.**

This approach gives complete spatial coverage (every square meter belongs to a pod), respects the ring model by construction, and allows pod sizes to vary by ring to match density targets. The walk-radius constraint is validated explicitly rather than assumed from geometry.

### Algorithm Sketch

1. Define ring boundaries as concentric arcs (or bands for non-circular footprints)
2. For each ring, calculate the number of pods needed: `ring_population / target_pod_population`
3. Place seed points evenly along the ring's midline arc
4. Compute Voronoi tessellation of all seed points
5. Clip each Voronoi cell to its ring boundary
6. Validate: for each cell, confirm max distance from any point to seed ≤ 400m
7. If validation fails, adjust seed positions and re-tessellate
8. Output: list of pods with center point, boundary polygon, ring assignment, and area

## Consequences

- Pod boundaries are polygons — building placement, path routing, and infrastructure branching operate on polygonal regions
- The algorithm must handle non-circular city footprints (square, irregular) by adapting ring definitions
- Service-sharing logic needs adjacency information — Voronoi tessellation naturally provides this (shared edges = adjacent pods)
- The solver must include a Voronoi library or implement the algorithm (e.g., Fortune's algorithm or Delaunay-based dual)
- Pod layout is the first solver stage and feeds into all subsequent stages
