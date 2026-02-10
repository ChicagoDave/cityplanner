# ADR-013: Persistence and Project Model

**Status:** Proposed
**Date:** 2026-02-09
**Deciders:** David Cornelson

---

## Context

The design tool operates in a loop: write spec, run solver, view in renderer, adjust spec, re-run. Users need to save and resume their work, compare different configurations, and possibly share designs. The tool produces several artifacts: the spec file, the scene graph, the cost report, and potentially saved camera positions or annotations.

We need to decide how projects are organized on disk, whether the scene graph is cached or always regenerated, and how versioning and undo/redo work.

## Decision Drivers

- The spec is the source of truth — the scene graph is derived from it
- Scene graph generation may take seconds to minutes — caching avoids redundant computation
- Users may want to compare two different specs side by side
- The tool should feel lightweight — not a heavyweight IDE with databases
- File-based storage is simplest and most portable
- Git-friendly formats enable version control of city designs

## Options

### Option A: Spec File Only — Always Regenerate

The project is a single YAML spec file. The scene graph is regenerated every time the renderer is opened. No caching.

**Pros:**
- Simplest possible persistence — one file is the entire project
- No cache invalidation problems
- Trivially git-friendly
- Spec is always in sync with what you see

**Cons:**
- Must wait for solver on every open — may be seconds to minutes
- Cannot open the renderer without running the solver
- No way to save camera positions, annotations, or bookmarks
- Comparison requires running the solver twice

### Option B: Project Directory with Cached Scene Graph

A project is a directory containing the spec file, the generated scene graph, a cost report, and optional user state (camera bookmarks, annotations). The scene graph is cached and only regenerated when the spec changes.

**Pros:**
- Fast reopening — load cached scene graph without re-solving
- User state (camera positions, notes) persists across sessions
- All artifacts in one directory — easy to zip, share, or git-track
- Cache invalidation is straightforward: hash the spec file, regenerate if hash changes
- Cost report is always available without re-solving

**Cons:**
- Cache can become stale if the solver code changes (same spec, different output)
- Directory structure is more to manage than a single file
- Scene graph files may be large (tens of MB)

### Option C: SQLite Project Database

A project is a single `.citydb` SQLite file containing the spec, scene graph, cost report, user state, and history.

**Pros:**
- Single file — easy to share and move
- Supports queries (e.g., "find all entities in pod 3 with cost > $X")
- Can store revision history within the database
- Atomic transactions for consistent state

**Cons:**
- Opaque binary format — not git-friendly, not human-readable
- Adds SQLite dependency
- Overkill for what is essentially a few JSON files
- Harder to debug than inspecting files in a directory
- Schema migrations needed as the tool evolves

### Option D: Project Directory with Manifest

A project directory with a manifest file that tracks contents and state:

```
my-city/
  city.yaml              — the spec (source of truth)
  city.scene.json        — cached scene graph
  city.cost.json         — cached cost report
  project.json           — manifest: spec hash, generation timestamp, user preferences
  bookmarks.json         — saved camera positions and annotations
```

The manifest tracks whether the cache is valid (spec hash match). The CLI checks the manifest before deciding whether to regenerate.

**Pros:**
- Human-readable files throughout — easy to inspect, debug, diff
- Git-friendly: spec and manifest are small text files; scene graph can be .gitignored
- Manifest enables smart caching without stale data
- Each artifact has a clear purpose and can be examined independently
- Easy to extend: add new files to the directory without breaking existing ones

**Cons:**
- Multiple files to manage (though the CLI handles this)
- Scene graph JSON may be large for git (solved by .gitignore)
- Directory must be kept consistent (manifest in sync with contents)

## Recommendation

**Option D: Project Directory with Manifest.**

This balances simplicity, debuggability, and Git compatibility. The spec file is the source of truth; everything else is derived or supplementary. The manifest enables smart caching. Large derived files (scene graph) are `.gitignore`-able while the spec and bookmarks are tracked.

### Project Lifecycle

```
cityplanner init my-city         → creates my-city/ with default city.yaml and project.json
cityplanner validate my-city     → runs Phase 1 validation, reports errors
cityplanner solve my-city        → runs full solver, writes scene graph and cost report
cityplanner view my-city         → opens renderer (re-solves if cache stale)
cityplanner cost my-city         → prints cost report (re-solves if needed)
```

### Manifest Structure

```json
{
  "project_name": "my-city",
  "spec_file": "city.yaml",
  "spec_hash": "sha256:abc123...",
  "solver_version": "0.1.0",
  "generated_at": "2026-02-09T14:30:00Z",
  "cache_valid": true,
  "scene_graph_file": "city.scene.json",
  "cost_report_file": "city.cost.json",
  "user_preferences": {
    "default_camera_mode": "bird_eye",
    "default_visible_layers": ["surface", "underground_3"]
  }
}
```

### Cache Invalidation

The cache (scene graph + cost report) is invalid when:
- Spec file hash doesn't match `spec_hash` in manifest
- Solver version doesn't match `solver_version` in manifest
- User explicitly requests regeneration (`--force`)

When the cache is invalid, `cityplanner view` or `cityplanner solve` regenerates automatically.

### Undo/Redo

Undo/redo operates at the spec level, not the scene graph level. Since the spec is a text file, undo/redo is best handled by the user's text editor or by Git. The tool does not implement its own undo stack — this avoids complexity and leverages existing tools.

For "what-if" comparison, users create spec variants (copy `city.yaml` to `city-variant.yaml`, adjust, solve both, compare cost reports).

## Consequences

- The CLI must manage project directories (init, validate, solve, view commands)
- `.gitignore` template should exclude `*.scene.json` (large, derived) but include `*.yaml`, `project.json`, `bookmarks.json`
- The scene graph serialization format (ADR-008) is written to disk as `city.scene.json`
- Bookmarks and annotations are a renderer concern but persisted in the project directory
- A `cityplanner watch` command could monitor spec file changes and auto-regenerate for a live design loop
