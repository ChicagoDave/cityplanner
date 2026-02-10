# ADR-014: Validation and Error Reporting

**Status:** Proposed
**Date:** 2026-02-09
**Deciders:** David Cornelson

---

## Context

The design tool must handle invalid, infeasible, and suboptimal specs gracefully. The user's workflow is iterative — they will frequently write specs that don't work and need clear feedback on what's wrong and how to fix it. Validation happens at multiple levels: spec syntax/schema, numerical feasibility (Phase 1 of ADR-009), and spatial feasibility (Phase 2).

The quality of error reporting directly affects how productive users are with the tool. A message like "solver failed" is useless; a message like "Pod 7 requires 3,200 residents but can only fit 2,400 at 4 stories within its 48-hectare boundary — increase max_height_edge to 6 or reduce edge ring population target" is actionable.

## Decision Drivers

- Errors occur at three levels: schema validation, analytical validation, spatial generation
- Users need actionable messages that point to specific spec parameters
- Warnings (suboptimal but valid) should be distinguished from errors (infeasible)
- The tool should suggest fixes where possible
- Validation should be runnable independently of the full solve (fast feedback)
- Error output must work in CLI (text) and renderer (visual overlay) contexts

## Options

### Option A: Simple Error Strings

Each validation check returns a string message on failure. Messages are collected and printed.

**Pros:**
- Simplest to implement
- No structured error format to design

**Cons:**
- Not machine-parseable — can't build UI features on error data
- No severity levels
- No structured reference to which spec parameters are involved
- Can't programmatically suggest fixes

### Option B: Structured Error Objects with Severity and Context

Each validation check returns a structured object containing severity, message, the spec path(s) involved, the conflicting values, and an optional suggested fix.

**Pros:**
- Machine-parseable — renderer can highlight problems in the UI
- Severity levels (error, warning, info) let users focus on what matters
- Spec path references (e.g., `city.population`, `pods.walk_radius`) enable click-to-fix in an editor
- Suggested fixes accelerate the design loop

**Cons:**
- More work to implement per validation check
- Fix suggestions require understanding the solution space for each constraint
- Structured format must be designed and maintained

### Option C: Structured Errors with Constraint Graph

Like Option B, but validation also outputs a constraint dependency graph showing which parameters affect each other. When a constraint fails, the graph shows the full chain of dependencies leading to the failure.

**Pros:**
- Maximum transparency — user sees exactly why something failed
- Enables sophisticated "what-if" analysis (change this parameter, see cascade)
- Useful for educational purposes (understanding city design tradeoffs)

**Cons:**
- Significant implementation complexity
- Constraint graph visualization is itself a UI challenge
- May overwhelm users with information
- Overkill for most validation scenarios

## Recommendation

**Option B: Structured Error Objects with Severity and Context.**

This provides actionable feedback without the complexity of a full constraint graph. Every validation check produces a structured result that includes enough context for both CLI display and renderer UI integration.

### Validation Levels

#### Level 1: Schema Validation (instant)

Validates the spec file against the JSON Schema (ADR-002). Catches:
- Missing required fields
- Wrong types (string where number expected)
- Values out of range (negative population, walk_radius < 0)
- Unknown fields (possible typos)

```typescript
{
  level: "schema",
  severity: "error",
  message: "demographics ratios must sum to 1.0, got 1.05",
  spec_path: "demographics",
  actual_value: 1.05,
  expected: "sum to 1.0",
  suggestion: "Reduce one or more cohort ratios by a total of 0.05"
}
```

#### Level 2: Analytical Validation (Phase 1, milliseconds)

Validates numerical feasibility after resolving the feedback loop. Catches:
- Density exceeds what the height envelope allows
- Population too low to sustain required services
- Pod population below service thresholds without adjacent pods to share
- Cost exceeds viability thresholds (if specified)
- Solar generation insufficient for demand (if site irradiance is low)

```typescript
{
  level: "analytical",
  severity: "error",
  message: "Edge ring density requires 120 du/ha but max 4 stories supports only 80 du/ha",
  spec_path: "city_zones.edge.max_stories",
  actual_value: 4,
  conflict_with: "demographics + population → 8,500 edge ring residents → 120 du/ha required",
  suggestions: [
    "Increase max_height_edge to 6 (supports ~120 du/ha)",
    "Reduce population to ~35,000 (reduces edge density to ~80 du/ha)",
    "Shift 10% of families to middle ring by adjusting pod ring assignments"
  ]
}
```

#### Level 3: Spatial Validation (Phase 2, seconds)

Validates that spatial generation succeeded. Catches:
- Pods that don't fit within the city footprint
- Buildings that can't be placed within a pod's boundary
- Infrastructure routes that exceed available corridor space
- Underground layer conflicts

```typescript
{
  level: "spatial",
  severity: "error",
  message: "Pod 7 (edge ring, NE sector): cannot place 42 buildings within boundary — 6 buildings overflow",
  spec_path: "city.footprint_shape",
  spatial_context: { pod_id: 7, ring: "edge", center: [720, 0, 340] },
  suggestions: [
    "Switch footprint_shape to 'circle' for more uniform pod areas",
    "Reduce edge pod population targets",
    "Increase city footprint area"
  ]
}
```

### Warning Examples (Non-Fatal)

```typescript
{
  level: "analytical",
  severity: "warning",
  message: "Battery storage provides only 18 hours of backup (target: 24 hours)",
  spec_path: "infrastructure.electrical.storage.capacity_mwh",
  actual_value: 2250,
  expected: 3000,
  suggestion: "Increase capacity_mwh to 3000 for full 24-hour backup"
}
```

```typescript
{
  level: "analytical",
  severity: "info",
  message: "Per-capita cost of $142,000 is within the $100K-160K target range but in the upper quartile",
  spec_path: "city.population",
  suggestion: "Increasing population to 60,000 would reduce per-capita cost to ~$125,000"
}
```

### Error Report Structure

```typescript
interface ValidationReport {
  valid: boolean;                    // true if no errors (warnings OK)
  errors: ValidationResult[];        // severity: error
  warnings: ValidationResult[];      // severity: warning
  info: ValidationResult[];          // severity: info
  summary: string;                   // "3 errors, 2 warnings, 1 info"
}

interface ValidationResult {
  level: "schema" | "analytical" | "spatial";
  severity: "error" | "warning" | "info";
  message: string;
  spec_path: string;
  actual_value?: unknown;
  expected?: string;
  conflict_with?: string;
  spatial_context?: { pod_id?: number; ring?: string; center?: number[] };
  suggestions?: string[];
}
```

### CLI Output

```
$ cityplanner validate my-city

  ERRORS (3):

  [schema] demographics ratios must sum to 1.0, got 1.05
    → demographics
    ⚡ Reduce one or more cohort ratios by a total of 0.05

  [analytical] Edge ring density requires 120 du/ha but max 4 stories supports only 80 du/ha
    → city_zones.edge.max_stories = 4
    ⚡ Increase max_height_edge to 6
    ⚡ Reduce population to ~35,000

  [spatial] Pod 7: cannot place 42 buildings — 6 overflow
    → city.footprint_shape
    ⚡ Switch to circle footprint for more uniform pod areas

  WARNINGS (1):

  [analytical] Battery storage provides only 18 hours of backup
    → infrastructure.electrical.storage.capacity_mwh = 2250
    ⚡ Increase to 3000 for 24-hour backup

  Result: INVALID (3 errors, 1 warning)
```

### Renderer Integration

In the renderer, spatial errors are visualized:
- Pods with errors are highlighted in red
- Clicking a highlighted pod shows the error detail and suggestions
- A validation panel lists all errors/warnings with links to the relevant spec lines

## Consequences

- Every validation check must produce a structured `ValidationResult` — more work per check but dramatically better UX
- Fix suggestions require domain knowledge embedded in the validators (e.g., knowing that 6 stories supports ~120 du/ha)
- The validation report is saved in the project directory (ADR-013) as `city.validation.json`
- Schema validation can run on every spec file save (sub-millisecond) for instant feedback
- Analytical validation runs as part of `cityplanner validate` (milliseconds)
- Spatial validation only runs as part of `cityplanner solve` (seconds)
- The renderer needs a validation panel component that consumes `ValidationResult[]`
