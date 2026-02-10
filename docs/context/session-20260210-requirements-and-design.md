# Session Summary: 2026-02-10 — City Requirements & Ground-Level Design Plan

## Status: Completed

## Goals
- Capture comprehensive city requirements through interactive discussion
- Work out population, pod count, geographic scale, and housing unit mix
- Design pod form factors for each ring density
- Plan the ground-level 2D city layout renderer

## Completed

### City Requirements Document (`docs/specification/city-requirements.md`)
Captured through iterative discussion and cataloged into a structured requirements doc:

**Design Principles**:
- Car-free city — no streets, no roads, no parking
- Three vertical layers of surface movement: elevated bike (+5m), ground pedestrian/shuttle, UG 1 underground

**Population & Scale**:
- ~64,000 people across 40 pods (~1,600 per pod)
- City diameter ~4.4 km, ~6 km with external band
- 5 concentric rings: 2/4/8/16/32 stories from edge to center

**Pods**:
- Self-contained neighborhoods with all daily life elements
- Sized to support a K-8 school (200 students)
- ~750 housing units per pod (40% 1BR, 33% 2BR, 18% 3BR, 7% 4BR, 2% 5BR)
- Form factors range from 360m diameter (2-story campus) to 90m (32-story vertical village)
- 50m green buffer between all non-commercial pods

**Education**: 40 K-8 schools (one per pod), 4 high schools (~800 students each), 1-2 community colleges

**Transport** (major pivot from maglev):
- Walking + biking as primary modes (city is small enough)
- Battery-powered automated shuttles (up to 10 passengers, on-demand + regular routes, with stations)
- Automated freight delivery on UG 2 (technology TBD)
- Maglev removed — overkill for 4.4 km city

**Pathways**:
- Elevated bike paths at ~5m with heated concrete (radiant heating, ice prevention)
- Bike paths connect into buildings at upper floors in taller rings
- Pedestrian paths with separated walk/run lanes

**Sports & Recreation**:
- 1 stadium/outdoor event field (centrally located)
- 10 soccer/cricket fields distributed across city
- Basketball, tennis, pickleball courts scattered in green buffers

**Underground**:
- UG 1: Public space connecting pods, pod basements open onto this level
- UG 2: Freight/delivery systems, mechanical/maintenance
- UG 3: Infrastructure (conduits, pipes, battery storage)

### Pod Form Factor Visualizations (`docs/specification/pod-form-factors.html`)
Created interactive HTML/SVG page with 6 diagrams:
- Ring 1 (2-story): sprawling campus with townhouse clusters
- Ring 2 (4-story): L/U-shaped perimeter blocks with garden courtyards
- Ring 3 (8-story): mixed-use mid-rise around central plaza
- Ring 4 (16-story): four towers on shared podium
- Center (32-story): vertical village — the pod IS the building
- City overview: all 40 pods in concentric rings

Deployed to `~/public_html/city/index.html`.

### Ground-Level 2D Renderer Design Plan (`docs/work/design/ground-level.md`)
6-phase implementation plan:

1. **Update spec & analytics** — Generalize from 3 rings to N (population-driven ring count), 64K/40 pods
2. **Bike & shuttle paths** — Organic spline curves through green space, co-located corridors, mobility hub stations with bike racks
3. **Sports field placement** — Stadium, soccer/cricket fields, small courts in 50m buffer zones
4. **2D scene assembly** — New `scene2d` package with lightweight JSON format
5. **2D SVG renderer** — Separate `renderer-2d/` Vite project, interactive pan/zoom/hover/click
6. **Deploy to public_html**

### Design Decisions Resolved
1. **Ring spec format**: Ordered list `rings: [...]`, ring count driven by population
2. **Bike paths**: Organic spline routing (landscape-architect aesthetic), not geometric circles
3. **Shuttle routes**: Co-located with bike paths, stations are combined mobility hubs
4. **Pod shapes**: Circular/elliptical (matching HTML designs), not angular Voronoi polygons
5. **External band**: Solar gets colored fill, water/waste as opaque blocks — simple, not detailed

## Key Decisions

### Maglev Replaced with Shuttles
The city's 4.4 km diameter makes maglev overkill. Edge to center is an 8-minute bike ride. Battery-powered automated shuttles (10 passengers, on-demand) handle accessibility and convenience. Freight delivery remains automated on UG 2 with technology left open for engineering.

### Pod Scale Anchored to K-8 Schools
200 students per K-8 school → ~1,600 people per pod → 40 pods for 64K population. This gives 4 high schools at ~800 students each and 1-2 community colleges.

### Population-Driven Ring Count
Ring count scales with population. A 64K city has 5 rings; a 150K city would have more. The solver should auto-generate ring structure from population + max stories parameters.

## Files Created
- `docs/specification/city-requirements.md` — Full city requirements document
- `docs/specification/pod-form-factors.html` — Interactive SVG form factor visualizations
- `docs/work/design/ground-level.md` — Ground-level 2D renderer implementation plan

## Next Steps
1. Begin Phase 1: Update city.yaml and generalize solver for N rings
2. Implement organic bike/shuttle path generation
3. Build the 2D SVG renderer
