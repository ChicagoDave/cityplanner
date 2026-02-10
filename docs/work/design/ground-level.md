# Ground-Level 2D City Layout Renderer

## Goal

Build the first renderer in a multi-renderer stack: an interactive 2D top-down view of the entire city showing all 40 pods in their concentric rings, bike paths, pedestrian paths, shuttle stations, sports fields, and green space. This validates the city's macro layout before we invest in pod-level detail or 3D.

## What Exists Today

### Solver (reusable)
- **Pod placement** (`layout/pods.go`): Constrained Voronoi tessellation. Takes seed points along ring midlines, computes cells, clips to ring annuli. Outputs `Pod` structs with `Center`, `Boundary` polygon, `Ring`, `AreaHa`, `TargetPopulation`.
- **Path generation** (`layout/paths.go`): Spine + connector + inter-pod paths per pod. Outputs `PathSegment` with `Start`, `End`, `WidthM`, `Type`.
- **Zone allocation** (`layout/zones.go`): Concentric radial bands within each pod (commercial, civic, residential, green). Outputs `Zone` with `Polygon`, `Type`.
- **Building placement** (`layout/buildings.go`): Grid-based placement within zones. Outputs `Building` with footprint, stories, type.
- **Height envelope** (`layout/envelope.go`): Bowl profile — max stories by distance from center.
- **Analytics** (`analytics/`): Ring resolution — computes pod count, population, area, density per ring. 3 rings today (center/middle/edge).
- **Geometry** (`geo/`): Full 2D geometry library — Point2D, Polygon, Voronoi, clipping, area, centroid, containment.
- **Scene assembly** (`scene/assemble.go`): Converts layout data to 3D scene graph JSON.

### What Needs to Change
The current solver hardcodes 3 rings (center/middle/edge) with radii 0-300-600-900m. The new requirements call for **5 rings** (center/ring4/ring3/ring2/ring1) with different radii and **40 pods at ~1,600 people each**. The solver algorithms (Voronoi, paths, zones) are ring-count-agnostic — they just need updated configuration.

The current renderer is Three.js (3D). We need a new **2D renderer** that consumes a simpler output format.

---

## Architecture

```
city.yaml (updated) → Go solver → 2D scene JSON → HTML/SVG/Canvas renderer
```

### New Output: 2D Scene Format

The existing 3D scene graph schema is overkill for a 2D top-down view. Define a simpler 2D-specific JSON output:

```json
{
  "metadata": {
    "population": 64000,
    "pod_count": 40,
    "city_radius_m": 2200,
    "external_band_radius_m": 3200
  },
  "rings": [
    { "name": "center", "radius_from": 0, "radius_to": 250, "stories": 32, "pod_count": 3 },
    { "name": "ring4", "radius_from": 250, "radius_to": 500, "stories": 16, "pod_count": 5 },
    ...
  ],
  "pods": [
    {
      "id": "pod_center_0",
      "ring": "center",
      "center": [x, z],
      "boundary": [[x,z], ...],
      "population": 1600,
      "stories": 32,
      "zones": [
        { "type": "residential", "polygon": [[x,z], ...] },
        { "type": "commercial", "polygon": [[x,z], ...] },
        ...
      ]
    },
    ...
  ],
  "paths": {
    "pedestrian": [
      { "id": "...", "points": [[x,z], ...], "width": 4, "type": "spine|connector|inter_pod" },
      ...
    ],
    "bike": [
      { "id": "...", "points": [[x,z], ...], "width": 3, "elevated": true },
      ...
    ],
    "shuttle": [
      { "id": "...", "points": [[x,z], ...], "width": 3 },
      ...
    ]
  },
  "stations": [
    { "id": "...", "pod_id": "...", "position": [x, z], "type": "shuttle" },
    ...
  ],
  "sports": {
    "stadium": { "position": [x,z], "dimensions": [110, 75] },
    "soccer_fields": [ { "position": [x,z], "dimensions": [105, 68] }, ... ],
    "courts": [ { "position": [x,z], "type": "basketball|tennis|pickleball", "dimensions": [28,15] }, ... ]
  },
  "external_band": {
    "radius_from": 2200,
    "radius_to": 3200,
    "facilities": [
      { "type": "solar", "arc_from": 0, "arc_to": 180 },
      { "type": "water_treatment", "arc_from": 180, "arc_to": 220 },
      ...
    ]
  }
}
```

---

## Implementation Plan

### Phase 1: Update City Spec & Analytics (solver)

**Goal**: Support the new 5-ring, 40-pod, 64K-population configuration.

#### Step 1.1: Update `city.yaml`
Update `examples/default-city/city.yaml` with new ring structure:

```yaml
city:
  population: 64000
  footprint_shape: circle
  excavation_depth: 8
  height_profile: bowl
  max_height_center: 32
  max_height_edge: 2

city_zones:
  center:
    character: civic_commercial
    radius_from: 0
    radius_to: 250
    max_stories: 32
  ring4:
    character: high_density
    radius_from: 250
    radius_to: 500
    max_stories: 16
  ring3:
    character: urban_midrise
    radius_from: 500
    radius_to: 850
    max_stories: 8
  ring2:
    character: mixed_residential
    radius_from: 850
    radius_to: 1350
    max_stories: 4
  ring1:
    character: low_density
    radius_from: 1350
    radius_to: 2200
    max_stories: 2
  perimeter_infrastructure:
    radius_from: 2200
    radius_to: 2700
    contents: [solar, water_treatment, waste_treatment, highway_connection, airport_connection]
  solar_ring:
    radius_from: 2700
    radius_to: 3200
    area_ha: 800
    capacity_mw: 600
    avg_output_mw: 120
```

#### Step 1.2: Generalize Ring Resolution
The current `analytics/geometry.go:resolveRings()` hardcodes 3 zones (center/middle/edge). Refactor to:
- Read an arbitrary list of zones from the spec (requires `CityZones` to become a slice or ordered map)
- Compute pod count per ring proportional to area
- Constrain total pods to 40 (configurable)
- Distribute ~1,600 people per pod

**Key change to `spec/types.go`**: `CityZones` becomes `Rings []RingDef` (ordered list, innermost first). Breaking change accepted. Ring count is population-driven — the solver can auto-generate a default ring structure from population + max center/edge stories if rings aren't specified manually.

#### Step 1.3: Update Pod Target Population
Current analytics distributes population proportional to ring area. For the new model, population should be **uniform per pod** (~1,600 each) since every pod is a self-contained neighborhood regardless of ring. The ring only determines building height, not population.

Change: `PodPopulation = city.Population / totalPodCount` (not proportional to ring area).

#### Step 1.4: Update Height Envelope
Current `layout/envelope.go` uses hardcoded breakpoints at 300/600/900m. Replace with a lookup against the new ring boundaries:
- Distance → which ring → that ring's `max_stories`
- Linear interpolation between ring boundaries

### Phase 2: Generate Bike & Shuttle Paths (solver)

**Goal**: Add bike path and shuttle route generation alongside existing pedestrian paths.

#### Step 2.1: Bike Path Network
New function in `layout/paths.go` (or new file `layout/bike_paths.go`):

**Algorithm — organic spline network**:
1. **Ring corridors**: Bike paths loosely follow ring boundaries but use spline curves that weave between pods through the green buffer space. Not perfect circles — organic, landscape-architect feel.
2. **Radial connections**: Curved paths connecting center to edge, threading through inter-pod green space. Not straight spokes — follow natural corridors.
3. **Countryside extensions**: Paths continue past city edge to external band and beyond.
4. **Spline generation**: Use Catmull-Rom or cubic Bezier splines through waypoints placed in buffer zones. Waypoints avoid pod interiors.

All bike paths are elevated (+5m) and marked as such in the output.

Output: `[]BikePath` with spline control points, sampled polyline, width, and `elevated: true`.

#### Step 2.2: Shuttle Route Network
New file `layout/shuttle.go`:

**Algorithm**:
1. **Co-located with bike paths**: Shuttle routes run alongside bike corridors as a shared infrastructure ribbon through the green space. One set of corridors, two modes.
2. **Stations as mobility hubs**: One per pod, placed at nearest point on a shuttle route to the pod center. Each station includes bike racks — combined shuttle + bike hub.
3. **Route coverage**: Ring corridors + radial connections ensure every pod is within a short walk of a station.

Output: `[]ShuttleRoute` with polyline points, and `[]Station` with position and pod ID.

#### Step 2.3: Integrate into Path Generation
Update the main layout pipeline to call bike and shuttle generation after pod placement. All path types share the same adjacency data.

### Phase 3: Place Sports Fields (solver)

**Goal**: Position 1 stadium, 10 soccer/cricket fields, and distributed small courts in the green buffer space between pods.

#### Step 3.1: Buffer Zone Identification
After pod placement, identify the inter-pod buffer zones (the 50m green strips). These are the Voronoi cell boundaries expanded outward — the "negative space" between pods.

**Algorithm**:
1. For each pair of adjacent pods, compute the midpoint corridor.
2. Buffer corridors are ~50m wide strips along pod boundary edges.
3. Find the centroid and dimensions of each buffer zone.

#### Step 3.2: Field Placement
- **Stadium**: Place near center (ring 3/4 boundary) in the largest available buffer zone. Size: 110m x 75m.
- **Soccer/cricket fields**: Distribute 10 across the city. Prefer larger buffer zones in outer rings (more space). Size: 105m x 68m. Only place in buffers where the field fits.
- **Small courts**: Fill remaining buffer zones with clusters of basketball (28x15m), tennis (12x24m), and pickleball (6x13m) courts.

### Phase 4: 2D Scene Assembly (solver)

**Goal**: New output format for the 2D renderer, separate from the existing 3D scene graph.

#### Step 4.1: New Package `pkg/scene2d/`
Create a dedicated 2D scene assembly package:

```go
package scene2d

type Scene2D struct {
    Metadata    Metadata        `json:"metadata"`
    Rings       []Ring          `json:"rings"`
    Pods        []Pod2D         `json:"pods"`
    Paths       PathNetwork     `json:"paths"`
    Stations    []Station       `json:"stations"`
    Sports      SportsLayout    `json:"sports"`
    ExternalBand ExternalBand   `json:"external_band"`
}
```

#### Step 4.2: Assembly Function
`Assemble2D()` takes layout results (pods, zones, paths, bike paths, shuttle routes, sports) and produces `Scene2D`. Pure data transformation — no new computation.

#### Step 4.3: New CLI Command
Add `cityplanner layout2d [project-path]` command that:
1. Loads spec
2. Runs analytics (Phase 1)
3. Runs pod layout (Phase 2 — pods, zones, paths)
4. Runs bike/shuttle path generation
5. Runs sports field placement
6. Assembles 2D scene
7. Writes JSON to stdout or file

Also add a `/api/scene2d` endpoint to the dev server.

### Phase 5: 2D HTML/SVG Renderer

**Goal**: Interactive browser-based 2D top-down city map.

#### Step 5.1: New Renderer Directory
Create `renderer-2d/` as a separate Vite project (or a new entry point in `renderer/`).

**Recommendation**: Separate directory `renderer-2d/` to avoid conflating 2D and 3D concerns. Lightweight — no Three.js dependency.

Tech stack:
- Vite + TypeScript (matches existing conventions)
- SVG for the map (scales cleanly, supports interaction, DOM-based click handling)
- Canvas fallback if SVG performance is an issue at 40 pods + paths (unlikely)

#### Step 5.2: Fetch & Parse
- Fetch `/api/scene2d` from the Go dev server
- TypeScript types matching the `Scene2D` JSON schema

#### Step 5.3: SVG Rendering Layers (bottom to top)

```
Layer 0: Background (dark/neutral)
Layer 1: External band (solar fields as colored fill, water/waste as opaque blocks)
Layer 2: Ring boundaries (dashed circles)
Layer 3: Green buffer zones (50m strips between pods)
Layer 4: Sports fields (rectangles in buffer zones)
Layer 5: Pod zone fills (residential, commercial, civic, green — colored polygons)
Layer 6: Pod boundaries (circular/elliptical outlines — soft, organic shapes)
Layer 7: Pedestrian paths (dashed lines)
Layer 8: Shuttle routes (solid lines with station dots)
Layer 9: Bike paths (colored lines, distinct style for elevated)
Layer 10: Labels (pod names, ring names, station names)
```

#### Step 5.4: Coordinate Transform
Solver outputs meters from city center (origin at 0,0). SVG needs pixel coordinates:
- Scale: fit city diameter (~6.4km with external band) to viewport
- Transform: `svgX = (meterX * scale) + viewportCenter`
- Pan/zoom: CSS transform or SVG viewBox manipulation

#### Step 5.5: Interaction
- **Pan & zoom**: Mouse drag to pan, scroll to zoom (or pinch on touch)
- **Hover**: Highlight pod on hover, show tooltip with pod info (population, ring, stories, area)
- **Click pod**: Show detail panel with zone breakdown, unit mix, services
- **Toggle layers**: Checkboxes to show/hide paths, zones, sports, rings, labels
- **Ring highlight**: Click ring label to highlight all pods in that ring

#### Step 5.6: Info Panel
Sidebar or overlay showing:
- City stats (population, pods, area, diameter)
- Selected pod details
- Legend (color key for zones, path types, etc.)

### Phase 6: Copy to public_html

Build step that copies the rendered output to `~/public_html/city/` for easy viewing alongside the existing form factor visualizations.

---

## File Changes Summary

### New Files
| File | Purpose |
|------|---------|
| `solver/pkg/layout/bike_paths.go` | Bike path network generation (ring loops + radial spokes) |
| `solver/pkg/layout/shuttle.go` | Shuttle route and station generation |
| `solver/pkg/layout/sports.go` | Sports field and court placement in buffer zones |
| `solver/pkg/scene2d/scene2d.go` | 2D scene types and assembly |
| `solver/cmd/cityplanner/layout2d.go` | CLI command for 2D layout output |
| `renderer-2d/` | New Vite+TS project for 2D SVG renderer |

### Modified Files
| File | Change |
|------|--------|
| `solver/pkg/spec/types.go` | Generalize `CityZones` to support N rings |
| `solver/pkg/analytics/geometry.go` | Generalize `resolveRings()` for N rings |
| `solver/pkg/analytics/types.go` | Update `RingData` if needed |
| `solver/pkg/layout/pods.go` | Works as-is (ring-count-agnostic) |
| `solver/pkg/layout/paths.go` | Minor — may add bike/shuttle integration |
| `solver/pkg/layout/envelope.go` | Update breakpoints for 5-ring model |
| `solver/pkg/layout/zones.go` | May need updated zone proportions per ring character |
| `examples/default-city/city.yaml` | New 5-ring, 64K population spec |
| `solver/cmd/cityplanner/serve.go` | Add `/api/scene2d` endpoint |

### Untouched
- `renderer/` (existing Three.js renderer stays as-is for now)
- `solver/pkg/routing/` (underground infrastructure — not needed for ground-level view)
- `solver/pkg/scene/` (3D scene assembly — kept for future use)

---

## Implementation Order

```
Phase 1 (spec + analytics)  ← Foundation, must come first
   ↓
Phase 2 (bike + shuttle)    ← Can start once pod placement works with new rings
   ↓
Phase 3 (sports fields)     ← Needs pod placement complete
   ↓
Phase 4 (2D scene assembly) ← Needs all solver output
   ↓
Phase 5 (2D renderer)       ← Needs JSON endpoint
   ↓
Phase 6 (deploy to public_html)
```

Phases 2 and 3 are independent of each other and can be done in parallel.

---

## Design Decisions (Resolved)

1. **Spec format**: Ordered list (`rings: [...]`). Breaking change accepted. Ring count is population-driven — a 64K city has 5 rings, a 150K city would have more. The solver should be able to auto-generate a default ring structure from population + max stories at center/edge.

2. **Bike path routing**: Organic, landscape-architect style. No perfect circles or straight radial spokes. Use spline curves that weave between pods, follow natural corridors through green space. This is an art/landscape design task — prioritize aesthetics over geometric simplicity.

3. **Shuttle routes**: Co-located with bike paths — one shared infrastructure ribbon through the green space. Shuttle stations include bike racks, functioning as combined mobility hubs. Reduces corridor sprawl through parkland.

4. **Pod boundary style**: Circular/elliptical forms (matching the HTML form factor designs). Voronoi is used internally for space partitioning, but rendered pod shapes should feel rounded and organic, not angular polygons.

5. **External band**: Simple rendering. Solar fields get a colored fill. Water treatment and waste facilities are opaque blocks. No internal detail — this is the city boundary, not the focus.
