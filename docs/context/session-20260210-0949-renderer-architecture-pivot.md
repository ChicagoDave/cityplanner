# Session Summary: 2026-02-10 - renderer-architecture-pivot (CST)

## Status: Paused - Awaiting Specification Rewrite

## Critical Finding

**User has identified fundamental architectural mismatch**: The current renderer approach produces "mathematical blobs" (primitive box geometry) when the requirement is CAD-like structural visualization with architectural detail. Session paused for user to write coherent rendering specification.

## Goals

- Continue renderer integration from previous session
- Add underground infrastructure support
- Implement network connectivity and route tracing
- Fix layer/system visibility logic

## Completed Work

### Phase 1: Underground Infrastructure Model Rework

**1. Fixed Layer Assignments per Technical Spec**
- Moved vehicle lanes back to Layer 3 (y=-2) to achieve "car-free surface" vision
- Surface (y=0): buildings, paths, parks
- Layer 1 (y=-1): sewage, water pipes
- Layer 2 (y=-1.5): electrical, telecom pipes, battery storage
- Layer 3 (y=-2): vehicle lanes, pedestrian tunnels, bike tunnels

**2. Added Two New Network Types**
- `pedway`: Underground pedestrian tunnels connecting all pods (36 segments for default city)
- `bike_tunnel`: Underground bicycle paths (36 segments for default city)
- Both routed using existing trunk-and-branch algorithm with lateral offsets
- All three Layer 3 systems run parallel with physical separation

**3. Battery Storage Integration**
- Added 6 battery room entities (one per pod, 500 MWh each per technical spec)
- Placed in Layer 2 alongside electrical infrastructure
- Type: `battery`, System: `electrical`

**4. Segment Connectivity Graph**
- **New file**: `solver/pkg/routing/connectivity.go` with spatial hash-based connectivity builder
- Algorithm: Build spatial hash of segment endpoints, find bidirectional neighbors within 0.1m tolerance
- **Result**: 252 segments (pipes + lanes + pedways + bike tunnels) now have `connected_to` metadata arrays
- Enables full network route tracing from any segment
- Added `ConnectedTo []string` field to `scene.Entity` metadata

### Phase 2: Extended Type System

**Entity Types** (8 total):
- `building`, `path`, `pipe`, `lane`, `park` (existing)
- `pedway`, `bike_tunnel`, `battery` (new)

**System Types** (7 total):
- `sewage`, `water`, `electrical`, `telecom`, `vehicle` (existing)
- `pedestrian`, `bicycle` (new)

### Phase 3: Renderer Visibility Logic Rewrite

**Fixed AND-logic for layer + system filtering**:
- Previous implementation: system toggles had no effect (bug)
- New implementation: Entity visible only if BOTH layer enabled AND system enabled
- Affects all 932 entities across 4 layers and 7 systems
- File: `renderer/src/ui/controls.ts` — complete rewrite of `updateVisibility()` logic

### Phase 4: Route Tracing Feature

**New file**: `renderer/src/ui/route-tracer.ts`
- Click any infrastructure segment to highlight its connected network
- Breadth-first search through `connected_to` graph
- Highlights in green, remembers previous selection for de-highlighting
- Works across all network types (pipes, lanes, pedways, bike tunnels)
- Integrated into `main.ts` raycaster click handler

### Phase 5: Testing and Validation

**All tests passing**:
- Go: 86 tests total (solver/pkg/*)
- TypeScript: Compiles cleanly with strict mode
- Binary built successfully

**Default city (50K pop) now generates**:
- 932 entities (up from ~880 in previous session)
  - 553 buildings
  - 115 paths
  - 144 pipes
  - 36 vehicle lanes
  - 36 pedestrian tunnels
  - 36 bicycle tunnels
  - 6 battery storage rooms
  - 6 parks
- 4 active layers
- 7 infrastructure systems
- 252 segments with connectivity metadata

## Key Decisions

### 1. All Underground Transit in Layer 3
**Rationale**: Technical spec calls for "car-free surface" and "underground vehicle lanes." Placing vehicle, pedestrian, and bicycle networks at same depth (y=-2) simplifies vertical organization and aligns with vision of fully underground mobility infrastructure. Lateral separation (x/z offsets) prevents physical overlap.

### 2. Spatial Hash for Connectivity
**Rationale**: With 252+ segments, O(n²) endpoint comparison would be expensive. Spatial hash reduces to O(n) with 10m grid cells. 0.1m tolerance handles floating-point precision in segment endpoints.

### 3. Metadata-Driven Connectivity
**Rationale**: Storing `connected_to` arrays directly in scene graph JSON allows renderer to trace routes without recomputing topology. Clean separation between solver (builds graph) and renderer (visualizes graph).

### 4. AND-Logic Visibility
**Rationale**: User expects both layer AND system filters to work together (e.g., "show Layer 3 electrical" = only electrical pipes in Layer 3). Previous OR-logic was non-intuitive.

## Critical User Feedback

### Rendering Approach is Wrong

User stated the current renderer produces "mathematical blobs" when requirement is "CAD-like structural visualization." Specific feedback:

> "the rendering is nowhere near what I'm looking to create"
> "I don't want a mathematical blob"
> "I actually want to see structures like a CAD design"

**Implications**:
- Current approach uses primitive `BoxGeometry` for all entities (buildings = boxes, pipes = boxes, tunnels = boxes)
- Technical spec envisions "first-person walkthrough at street level" and "underground exploration of vehicle lanes, pipe runs, utility corridors"
- CAD-like visualization requires:
  - Proper architectural geometry for buildings (walls, windows, roofs, not just rectangular volumes)
  - Tunnel/corridor cross-sections with structural detail
  - Road surfaces with lane markings, curbs, texture
  - Infrastructure with realistic pipe/conduit geometry
  - Level-of-detail system for city-wide → street-level → interior navigation
  - Possibly procedural generation or 3D model assets
  - Sophisticated instancing for performance at scale

**Session paused** while user writes more coherent rendering specification.

## Files Modified

**Solver (Go)** (8 files):
- `pkg/routing/routing.go` — Layer 3 reassignment, pedway/bike_tunnel networks, connectivity integration
- `pkg/routing/connectivity.go` — NEW: spatial hash connectivity builder (140 lines)
- `pkg/routing/routing_test.go` — Tests for new networks + connectivity (added 3 tests)
- `pkg/scene/scene.go` — New entity/system type constants
- `pkg/scene/assemble.go` — `assembleBatteries()`, new entity type handling, layerFromInt fix
- `pkg/scene/assemble_test.go` — Updated assertions for 932 entities
- `pkg/scene/validate.go` — New entity type validation
- `pkg/scene/validate_test.go` — Test coverage for new types

**Renderer (TypeScript)** (5 files):
- `src/scene/loader.ts` — New entity types in mesh creation
- `src/scene/materials.ts` — Highlight material for route tracing
- `src/ui/controls.ts` — AND-logic visibility rewrite
- `src/ui/route-tracer.ts` — NEW: click-to-highlight route tracing (80 lines)
- `src/main.ts` — RouteTracer integration

## Architectural Notes

### Connectivity Graph Algorithm
The spatial hash approach in `connectivity.go` is elegant:
1. Hash all segment endpoints into 10m grid cells
2. For each segment, check its cell + 26 neighbor cells for potential connections
3. Distance threshold 0.1m handles floating-point precision
4. Bidirectional edges stored as string arrays in scene graph metadata

This scales linearly with segment count and produces clean metadata for renderer consumption.

### Layer Model Now Correct
Previous sessions had confusion about vehicle lane placement (was Layer 2, then surface, now correctly Layer 3). The model now matches technical spec vision:
- **Surface**: Human-scale pedestrian realm (buildings, parks, walking paths)
- **Layer 1**: Waste/water (gravity-fed systems)
- **Layer 2**: Energy/data (electrical, telecom, batteries)
- **Layer 3**: Mobility (vehicle lanes, pedestrian tunnels, bike tunnels)

This vertical stacking enables proper underground exploration visualization (when rendering is fixed).

### Scene Graph Entity Count
932 entities for 50K population city is reasonable scale. Breakdown:
- ~60% buildings (residential, commercial, civic)
- ~30% infrastructure segments (pipes, lanes, tunnels)
- ~10% public space (paths, parks)

This density will test renderer performance once proper geometry is implemented.

## Open Items

### Immediate (Blocking)
1. **User to write rendering specification** — must define:
   - Geometry approach (procedural? 3D models? hybrid?)
   - Level-of-detail strategy
   - Building appearance requirements
   - Infrastructure visualization style
   - Performance targets (FPS at what entity count?)
2. **Renderer architecture decision** — may require:
   - New Three.js scene organization (instancing? LOD groups?)
   - Asset pipeline for 3D models
   - Procedural geometry generators
   - Shader materials for architectural effects

### Short Term (After Spec)
1. Implement new rendering approach per user's spec
2. Add proper camera movement for street-level + underground exploration
3. Performance profiling and optimization
4. UI improvements for navigation (minimap? floor selector?)

### Long Term
1. **Animation system**: Spec mentions "simulate one day in 60 seconds with agent paths"
2. **Agent visualization**: People/vehicles moving through city
3. **Analytics overlay**: Heat maps, flow diagrams
4. **Interactive editing**: Modify city parameters and re-solve in browser

## Notes

**Session duration**: ~2.5 hours across two conversation threads (previous ran out of context)

**Approach**: Extended solver with underground infrastructure model, fixed visibility logic, added connectivity graph for route tracing. All technical implementations successful, but fundamental rendering approach misaligned with user vision.

**Previous Session Context**: Session before this one implemented initial renderer integration (materials, loader, camera controls), fixed 3 solver bugs (commercial building count, park dimensions, layer assignments), added scene validation and benchmarks. This session continued that work by adding underground networks and fixing visibility/tracing features.

**Key Insight**: The solver is producing correct geometric and topological data (positions, dimensions, connectivity). The problem is purely in the renderer's interpretation of that data as primitive boxes rather than architectural structures. The scene graph JSON is likely sufficient as-is — the issue is the mesh generation in `loader.ts`.

---

**Progressive update**: Session paused 2026-02-10 09:49 CST — awaiting rendering specification from user
