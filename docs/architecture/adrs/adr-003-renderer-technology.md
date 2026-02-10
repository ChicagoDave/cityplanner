# ADR-003: Renderer Technology

**Status:** Proposed
**Date:** 2026-02-09
**Deciders:** David Cornelson

---

## Context

The design engine requires a 3D renderer capable of first-person walkthrough, underground exploration, layer toggling, cross-section views, and level-of-detail from city-wide overview to individual pipe junctions. The technical spec identifies Three.js, Bevy, and Godot as candidates, noting that "Three.js is the pragmatic choice given web accessibility and TypeScript alignment."

This decision gates the deployment model, the frontend technology stack, and the level of effort required for the visualization features.

## Decision Drivers

- Must support first-person navigation, bird's-eye view, orbit camera, and underground exploration
- Must handle layer toggling and arbitrary vertical cross-sections
- Must render a city-scale scene (thousands of buildings, kilometers of pipe, vehicle lanes)
- Shareability matters — stakeholders should be able to view designs without installing software
- The solver is decoupled (ADR-001), so the renderer consumes a serialized scene graph
- Development velocity matters — this is a design tool, not a shipped game

## Options

### Option A: Three.js (Web, TypeScript)

Browser-based 3D engine. Mature, widely used, extensive ecosystem.

**Pros:**
- Zero-install sharing: send a URL, anyone can view the city
- TypeScript-native — same language as solver if solver is also TypeScript
- Huge ecosystem: loaders, controls, post-processing, spatial indexing libraries
- Community and hiring pool
- Incremental rendering features — start simple, add sophistication over time

**Cons:**
- Browser performance ceiling — WebGL/WebGPU has limits vs native GPU access
- Large scenes require careful optimization (instancing, LOD, frustum culling)
- No built-in physics or collision (must add for walkthrough navigation)
- Memory constrained by browser tab limits

### Option B: Bevy (Rust, Native)

Rust-based ECS game engine. High performance, modern architecture.

**Pros:**
- Excellent performance — native GPU access, ECS architecture handles large entity counts
- Rust aligns with a Rust solver (if ADR-004 chooses Rust)
- Strong type system and memory safety

**Cons:**
- Requires desktop install to view designs — no URL sharing
- Smaller ecosystem and community than Three.js
- Steeper learning curve
- Bevy is still pre-1.0; API churn is ongoing
- Cross-platform distribution adds build complexity

### Option C: Godot (GDScript/C#, Native + Web Export)

Full game engine with editor, scene system, and web export capability.

**Pros:**
- Rich built-in features: physics, navigation, UI, scripting
- Web export via Emscripten (browser viewing possible)
- Visual scene editor useful for debugging spatial layouts
- Large community, extensive documentation

**Cons:**
- Heaviest dependency — full game engine for what is a visualization tool
- GDScript is a niche language; C# adds a third language to the stack
- Web export produces large WASM bundles with long load times
- Opinionated scene/node system may fight against consuming an external scene graph
- Overkill for a tool that doesn't need game mechanics

### Option D: WebGPU Direct (TypeScript, Low-Level)

Build a custom renderer directly on the WebGPU API without a framework.

**Pros:**
- Maximum control over rendering pipeline
- WebGPU is the future standard, better performance than WebGL
- No framework overhead or abstractions to work around

**Cons:**
- Enormous development effort — rebuilding what Three.js provides
- WebGPU browser support is still incomplete (as of early 2026)
- No community resources for common problems
- Not practical for a design tool where visualization is a means, not the product

## Recommendation

**Option A: Three.js.**

The spec already identifies this as the pragmatic choice, and the reasoning holds. Zero-install web sharing is a significant advantage for a design tool that needs stakeholder review. Three.js handles city-scale scenes when properly optimized (instanced meshes, LOD, frustum culling, octree spatial indexing). TypeScript alignment simplifies the stack. The ecosystem provides ready-made solutions for camera controls, post-processing, and spatial queries.

Performance concerns are manageable at the expected scene complexity. The city is ~350 hectares with thousands of buildings and infrastructure segments — large but well within what optimized Three.js applications handle. If WebGPU support in Three.js matures (active development), it provides a performance upgrade path without changing application code.

### Key Implementation Notes

- Use `InstancedMesh` for repeated geometry (building types, pipe segments, solar panels)
- Implement LOD: simplified geometry at distance, full detail up close
- Spatial indexing (octree) for frustum culling and ray intersection
- Consider offscreen/worker-based scene graph processing for initial load

## Consequences

- The application is a web application — needs a dev server, bundler, and browser testing
- Must handle browser memory limits for large scenes (streaming, LOD, disposal)
- Camera/navigation system must be custom-built or assembled from Three.js addons
- Cross-section and layer toggling are shader/clipping-plane problems (see ADR-011)
- The renderer codebase is TypeScript, which aligns with or constrains ADR-004
