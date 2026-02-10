# ADR-011: Layer Visibility and Cross-Section System

**Status:** Proposed
**Date:** 2026-02-09
**Deciders:** David Cornelson

---

## Context

The technical spec requires several visualization modes:

- **Layer toggling:** Show only water pipes, only electrical, only vehicles, etc.
- **Cross-section mode:** Slice the city vertically at any point to see underground layers
- **Underground exploration:** Navigate below the surface to inspect vehicle lanes, pipe runs, and utility corridors
- **Level-of-detail:** Zoom from city-wide overview to individual pipe junctions

These capabilities are essential for evaluating and debugging the generated city design. They must work performantly on a scene with thousands of entities across multiple underground layers and the surface.

## Decision Drivers

- Layer toggling must be instant (no perceptible delay)
- Cross-sections must work at arbitrary positions and angles
- Underground and surface must be viewable simultaneously (cross-section) or independently
- Must work within Three.js (per ADR-003)
- Large entity counts require efficient filtering, not brute-force show/hide
- These modes interact: a user may want "water system only, cross-section at X=200m"

## Options

### Option A: Scene Graph Filtering (Show/Hide Objects)

Each Three.js object has a `visible` flag. Layer toggling sets `visible = false` on all objects not matching the active filter. Cross-section is implemented by hiding all objects on one side of a plane.

**Pros:**
- Simplest to implement
- Uses built-in Three.js visibility
- No shader complexity

**Cons:**
- Toggling thousands of objects is O(n) per toggle — may cause frame hitches
- Cross-section by hiding objects produces jagged edges (whole objects appear or disappear)
- No smooth cut through buildings — a building is either fully visible or fully hidden
- Memory remains allocated for hidden objects

### Option B: GPU Clipping Planes (Shader-Based)

Use Three.js clipping planes to cut the scene. Each material is configured with clipping planes. Layer toggling uses a custom shader uniform that discards fragments not belonging to the active layer. Cross-section uses a world-space clipping plane.

**Pros:**
- Cross-section produces clean, smooth cuts through geometry
- Clipping is per-pixel — no jagged object-level boundaries
- Layer filtering via shader discard is fast (GPU-side, no CPU iteration)
- Multiple clipping planes can combine (cross-section + layer filter)
- Three.js has built-in clipping plane support in its material system

**Cons:**
- Custom shaders add complexity (must extend Three.js materials)
- Clipped geometry shows hollow interiors unless caps are generated
- All materials must support clipping — harder if using diverse material types
- Debugging shader issues is harder than debugging visibility flags

### Option C: Multiple Render Passes

Render each layer/system as a separate pass into a framebuffer, then composite. Layer toggling enables/disables passes. Cross-section applies a stencil or clip in a global pass.

**Pros:**
- Complete isolation between layers
- Each pass can have different rendering settings (e.g., wireframe for pipes, solid for buildings)
- Post-processing can be applied per-layer (e.g., highlight active system)

**Cons:**
- Multiple render passes multiply GPU work (draw calls × pass count)
- Complex compositing logic
- Overkill for the visualization needs described in the spec
- Harder to implement interactions between visible layers

### Option D: Hybrid — Clipping Planes for Cross-Section + Group Visibility for Layers

Cross-section uses GPU clipping planes (clean cuts, per-pixel). Layer toggling uses the scene graph groups (ADR-008) to set visibility on pre-organized entity groups. Since groups are pre-computed lists of entity IDs, toggling a layer is O(group size), and groups are already organized by system type.

**Pros:**
- Best tool for each job: clipping planes for spatial cuts, group visibility for categorical filtering
- Group-based toggling leverages the scene graph structure from ADR-008
- Clipping planes produce clean cross-sections
- Implementation complexity is moderate — no custom shaders for layer toggling, standard clipping for cross-section
- Modes compose naturally: toggle a layer, then clip

**Cons:**
- Two different mechanisms to maintain
- Group visibility toggling still touches many objects (though pre-indexed)
- Clipped geometry still shows hollow interiors without cap generation

## Recommendation

**Option D: Hybrid — Clipping Planes for Cross-Section + Group Visibility for Layers.**

This approach uses the right mechanism for each feature. The scene graph's group structure (ADR-008) pre-indexes entities by system, layer, pod, and type — toggling a group is a fast, pre-computed operation. Clipping planes are the standard approach for cross-sections in Three.js and produce clean, per-pixel cuts.

### Layer Toggling Implementation

```typescript
// Scene graph groups from ADR-008 map to Three.js object groups
function toggleSystem(system: SystemType, visible: boolean) {
  const entityIds = sceneGraph.groups.systems.get(system);
  for (const id of entityIds) {
    threeObjects.get(id).visible = visible;
  }
}

// Pre-built toggle presets
const presets = {
  surface_only: { underground_1: false, underground_2: false, underground_3: false, surface: true },
  water_only: { water: true, /* all others false */ },
  vehicles_only: { vehicle: true },
  all: { /* everything true */ },
};
```

### Cross-Section Implementation

```typescript
// Global clipping plane — all materials opt in
const clipPlane = new THREE.Plane(new THREE.Vector3(1, 0, 0), 0);
renderer.clippingPlanes = [clipPlane];

// User drags to move the cross-section plane
function setCrossSectionPosition(axis: 'x' | 'y' | 'z', position: number) {
  clipPlane.normal.set(axis === 'x' ? 1 : 0, axis === 'y' ? 1 : 0, axis === 'z' ? 1 : 0);
  clipPlane.constant = -position;
}
```

### Cap Generation for Cross-Sections

When clipping cuts through solid geometry (buildings, pipes), the interior is visible as a hollow shell. Cap generation adds a filled surface at the clipping plane intersection. This is a known Three.js problem with community solutions (stencil-buffer cap rendering). It is a polish feature, not required for initial implementation.

## Consequences

- All Three.js materials must have `clippingPlanes` enabled (a renderer-level setting)
- The UI needs controls for: system toggles, layer toggles, cross-section axis/position slider, and preset buttons
- Layer toggling and cross-section compose: a user can view "water system only, clipped at X=300m"
- Performance depends on group sizes — toggling the largest group (all buildings) may touch thousands of objects
- For initial implementation, batch visibility updates using `THREE.Group` containers per system/layer rather than individual object toggles
- Cap generation can be deferred to a later iteration
