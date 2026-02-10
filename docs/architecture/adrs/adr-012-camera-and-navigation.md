# ADR-012: Camera and Navigation Modes

**Status:** Proposed
**Date:** 2026-02-09
**Deciders:** David Cornelson

---

## Context

The technical spec describes several navigation modes for evaluating the generated city:

- **First-person walkthrough** at street level
- **Underground exploration** of vehicle lanes, pipe runs, utility corridors
- **Bird's-eye view** for city-wide overview
- **Orbit camera** for examining specific areas
- **Level-of-detail** from city-wide to individual pipe junctions

The camera system is how the user experiences and evaluates the design. It must handle transitions between surface and underground, support very different scales (city-wide vs. pipe junction), and feel responsive enough for fluid exploration.

## Decision Drivers

- Multiple camera modes with different control schemes (FPS, orbit, top-down)
- Must transition between surface and underground seamlessly
- City scale spans ~2km across — first-person movement must be fast enough to traverse
- Underground spaces have tight clearances (vehicle lanes ~3m ceiling) — collision/clipping matters
- LOD is required: simplified geometry at distance, full detail up close
- The tool is for design evaluation, not gaming — precision and clarity matter more than cinematic quality

## Options

### Option A: Single Unified Free Camera

One camera with adjustable speed. WASD/mouse movement. No mode switching — the user flies freely through the entire scene, including through walls and underground.

**Pros:**
- Simplest to implement
- Maximum freedom — no constraints on where the camera can go
- No mode-switching UI needed

**Cons:**
- Flying through walls is disorienting — no sense of place
- No ground reference — easy to lose orientation underground
- Speed that works for city-wide traversal is too fast for pipe-junction inspection
- Doesn't feel like walking through a city

### Option B: Discrete Camera Modes with Smooth Transitions

Three distinct modes, each with appropriate controls:

1. **First-person (street/underground):** WASD movement, mouse look, gravity, collision. Speed ~5 m/s (walking) to ~15 m/s (jogging). Toggleable between surface and underground layers.
2. **Orbit:** Click a point of interest, orbit around it at adjustable distance. Scroll to zoom. Good for inspecting a building, pod, or infrastructure junction.
3. **Bird's-eye:** Top-down orthographic or perspective view. Pan with drag, zoom with scroll. Shows the full city or a selected region.

Transitions between modes animate smoothly (camera interpolation over ~500ms).

**Pros:**
- Each mode has controls optimized for its purpose
- First-person gives a human-scale experience of the city
- Orbit is ideal for inspecting specific elements
- Bird's-eye provides the planning/overview perspective
- Smooth transitions prevent disorientation

**Cons:**
- Three control schemes to implement and test
- Mode switching adds UI complexity (toolbar buttons, keyboard shortcuts)
- First-person collision detection requires a physics/nav mesh layer
- Transition animation logic adds implementation effort

### Option C: Google Earth-Style Continuous Camera

A single camera that seamlessly transitions between overhead and ground-level. Scroll to zoom; at maximum zoom, the camera tilts to ground level. Double-click to fly to a location. No explicit mode switching.

**Pros:**
- Intuitive — familiar UX paradigm from Google Earth/Maps
- No mode switching — zoom level determines the experience
- Smooth, continuous transitions
- Works well for the overview-to-detail workflow

**Cons:**
- Complex to implement well (tilt, altitude, and zoom are coupled)
- Underground exploration doesn't fit this model naturally
- Not ideal for sustained first-person walkthrough (ground-level view)
- May fight with layer toggling and cross-section modes

## Recommendation

**Option B: Discrete Camera Modes with Smooth Transitions.**

Design evaluation requires distinct perspectives: walking through the city at human scale, inspecting a junction up close, and viewing the full layout from above. These are different tasks with different control needs. Discrete modes with keyboard shortcuts (1 = first-person, 2 = orbit, 3 = bird's-eye) provide clarity. Smooth camera interpolation during transitions prevents disorientation.

### Mode Specifications

#### First-Person Mode
- **Controls:** WASD movement, mouse look, space to jump/ascend, shift for speed boost
- **Physics:** Gravity-bound to surface or underground floor. Simple collision with walls and floors (bounding-sphere check against nearby geometry).
- **Speed:** Base 5 m/s, shift 15 m/s, configurable
- **Layer transition:** A UI toggle or hotkey switches between surface and underground. Camera is repositioned to the corresponding level at the same XZ coordinates. Transition animates vertically.
- **Underground:** Camera height locked to lane/corridor floor + eye height (1.7m). Collision prevents passing through walls.

#### Orbit Mode
- **Activation:** Double-click an entity or press "2" to orbit the current look-at point
- **Controls:** Left-drag to orbit, right-drag to pan, scroll to zoom
- **Center point:** World-space position, shown with a subtle indicator
- **Zoom range:** 1m (pipe junction detail) to 500m (pod overview)

#### Bird's-Eye Mode
- **Controls:** Click-drag to pan, scroll to zoom, right-drag to tilt angle
- **Projection:** Perspective with high altitude, transitioning toward orthographic feel at maximum zoom-out
- **Overlays:** Pod boundaries, ring boundaries, system color-coding visible in this mode
- **Zoom range:** Single pod to full city + solar ring

### Level of Detail (LOD)

LOD is orthogonal to camera mode but driven by camera distance:

| Distance from camera | Detail level |
|---------------------|-------------|
| < 50m | Full detail: individual windows, pipe joints, signage |
| 50-200m | Medium: building volumes with material colors, pipe runs as cylinders |
| 200-500m | Low: simplified block volumes, infrastructure as lines |
| > 500m | Minimal: colored blocks for buildings, no infrastructure |

LOD transitions use Three.js `LOD` objects with distance-based switching.

### Keyboard Shortcuts

```
1 — First-person mode
2 — Orbit mode
3 — Bird's-eye mode
U — Toggle underground/surface (first-person)
F — Fly to selected entity
R — Reset camera to default overview
```

## Consequences

- First-person mode requires basic collision detection — either a nav mesh or distance checks against nearby geometry using the spatial index (ADR-008)
- Underground exploration needs the layer visibility system (ADR-011) to hide the surface when viewing underground
- LOD requires multiple geometry representations per entity type — adds to scene graph complexity
- Camera state (position, mode, target) should be saveable/restorable as part of project state (ADR-013)
- A minimap or orientation indicator helps users maintain spatial awareness, especially underground
