# ADR-008: Scene Graph Structure

**Status:** Proposed
**Date:** 2026-02-09
**Deciders:** David Cornelson

---

## Context

The solver produces spatial data for every entity in the city: buildings, paths, pipes, vehicle lanes, utility corridors, solar panels, parks, and more. This data must be organized into a structure that the renderer can traverse efficiently for visualization, that supports spatial queries (e.g., "what's at this coordinate?"), and that enables layer toggling and cross-section views.

The technical spec describes this as a "traversable 3D spatial structure" where "every pipe, wall, lane, building, and path has position, dimension, and material assignment."

This ADR also addresses the serialization format between solver and renderer (per ADR-001's consequence).

## Decision Drivers

- The renderer needs fast spatial queries for frustum culling, ray picking, and LOD
- Layer toggling requires filtering entities by system type (water, electrical, vehicles, etc.)
- Cross-section views require clipping geometry at arbitrary planes
- The scene graph must be serializable for the solver→renderer handoff (ADR-001)
- Entity count is large: thousands of buildings, kilometers of pipe segments, many thousands of path segments
- The structure must support both city-wide overview and individual pipe junction detail

## Options

### Option A: Flat Entity List with Spatial Index

All entities in a single flat array. Each entity has position, dimensions, material, type, layer, and pod assignment. A separate spatial index (octree or R-tree) enables fast spatial queries.

**Pros:**
- Simplest data structure — easy to serialize, easy to iterate
- Spatial index handles all query needs (frustum culling, ray picking, range queries)
- Filtering by type/layer is a simple array filter
- No hierarchy to maintain or traverse

**Cons:**
- No inherent grouping — "all buildings in pod 3" requires scanning or secondary indices
- Large flat arrays may be slow to deserialize and hold in memory
- No parent-child relationships (e.g., floors within a building, fixtures within a floor)

### Option B: Hierarchical Scene Tree

Entities organized in a tree: City → Rings → Pods → Buildings/Infrastructure → Components. Transforms are inherited down the tree. Each node has a bounding box for hierarchical culling.

**Pros:**
- Natural grouping: select a pod and get all its contents
- Hierarchical bounding volumes enable fast top-down culling
- Matches the conceptual structure (city contains rings, rings contain pods, pods contain buildings)
- Parent-child transforms simplify placement (building floors relative to building origin)

**Cons:**
- Infrastructure networks cross pod boundaries — where do trunk pipes live in the hierarchy?
- Deep trees add traversal overhead
- Serialization of tree structures is more complex than flat arrays
- Cross-cutting queries (e.g., "all water pipes everywhere") require full tree traversal

### Option C: Entity-Component System (ECS)

Entities are IDs. Components are data arrays (Position, Dimensions, Material, PodMembership, SystemType, MeshReference). Systems operate on entities matching component queries.

**Pros:**
- Maximum flexibility — any query is a component filter
- Cache-friendly data layout (arrays of components, not arrays of objects)
- Natural fit for "show only water system" (filter by SystemType component)
- Scales well to large entity counts
- Well-suited to renderer integration (Three.js can consume component arrays for instanced rendering)

**Cons:**
- More complex to implement from scratch
- ECS is a pattern, not a data format — serialization must be designed
- No inherent spatial hierarchy — spatial index is still needed separately
- Overhead may not be justified if entity counts are in the thousands (not millions)

### Option D: Layered Hybrid — Hierarchical Grouping with ECS-Style Components

Entities organized in a two-level hierarchy: **Groups** (by pod, by system, by layer) and **Entities** within groups. Each entity has typed component data. Groups can overlap (an entity belongs to both "Pod 3" and "Water System"). A spatial index sits alongside for fast spatial queries.

**Pros:**
- Group-level operations are fast (toggle all water, select all of Pod 3)
- Entity-level data is flat and efficient within groups
- Spatial index handles geometry queries orthogonally
- Serialization is tractable: groups are named arrays of entity references, entities are typed records
- Cross-cutting queries work via group intersection

**Cons:**
- Group membership management adds bookkeeping
- Overlapping groups mean entities are referenced multiple times (storage overhead)
- More complex than a flat list, less formal than full ECS

## Recommendation

**Option D: Layered Hybrid.**

The city has natural grouping axes that are used constantly: by pod (select a neighborhood), by system (show only water), by layer (underground level 2), and by entity type (all buildings). A hybrid approach with overlapping groups and a spatial index serves all these queries without forcing everything into a single hierarchy.

### Data Model

```typescript
interface SceneGraph {
  metadata: {
    spec_version: string;
    generated_at: string;
    city_bounds: BoundingBox;
  };
  entities: Entity[];           // flat array, indexed by entity ID
  groups: {
    pods: Map<PodId, EntityId[]>;
    systems: Map<SystemType, EntityId[]>;   // water, sewage, electrical, telecom, vehicle
    layers: Map<LayerType, EntityId[]>;     // underground_1, underground_2, underground_3, surface
    entity_types: Map<EntityType, EntityId[]>; // building, pipe, lane, path, panel, etc.
  };
  spatial_index: OctreeNode;    // built on load, not serialized
}

interface Entity {
  id: EntityId;
  type: EntityType;
  position: Vec3;
  dimensions: Vec3;
  rotation: Quaternion;
  material: MaterialId;
  system?: SystemType;
  pod?: PodId;
  layer: LayerType;
  metadata: Record<string, unknown>;  // capacity, flow_rate, unit_count, etc.
  children?: EntityId[];              // floors in a building, fixtures in a floor
}
```

### Serialization Format

The scene graph is serialized as a single JSON file for debuggability and simplicity. If file size becomes a problem (likely for full city scenes), a binary format (e.g., MessagePack or FlatBuffers) can be introduced as an optimization.

File structure: `{city_name}.scene.json`

### Spatial Index

The spatial index (octree) is **not** serialized — it is rebuilt on load from entity positions. This avoids serialization complexity and keeps the format simple. Octree construction for ~10,000 entities is sub-second.

## Consequences

- The solver outputs a `.scene.json` file containing all entities and group memberships
- The renderer loads this file, builds the octree, and creates Three.js objects from entities
- Group-based filtering (layer toggle, system toggle) operates on pre-computed entity ID lists — O(1) lookup
- Entity `children` support limited hierarchy (building → floors) without a full tree
- Material IDs reference a shared material palette (defined in the scene graph or spec)
- The `metadata` field on entities carries domain-specific data (pipe capacity, unit count, etc.) for inspector/tooltip display
- JSON serialization may produce files in the 10-100 MB range for a full city — acceptable for local file I/O, may need streaming for web delivery
