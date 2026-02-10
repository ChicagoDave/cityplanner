# Plan: Phase 2 (Bike & Shuttle Paths) + Phase 3 (Sports Fields)

## Context

Phase 1 (N-ring generalization) is complete. The solver supports 5 rings, 32 pods, 64K population with capacity-weighted distribution. Next steps from `docs/work/design/ground-level.md`: add surface-level bike paths, shuttle routes with stations, and sports fields in inter-pod buffer zones.

The geo package has no spline/curve support — all existing paths are straight-line segments. We need Catmull-Rom splines for the organic bike path feel the design doc calls for.

## Implementation Steps

### Step 1: Catmull-Rom splines + Polyline type

**New file**: `solver/pkg/geo/spline.go`

- `Polyline` type: `Points []Point2D`
- `CatmullRomSpline(controlPoints []Point2D, samplesPerSegment int, tension float64) Polyline` — evaluates spline through all control points. Phantom endpoints by reflecting first/last segments. Standard CR matrix formula applied to X and Z independently.
- `Polyline.Length() float64`
- `Polyline.PointAt(t float64) Point2D` — point at fraction of total length
- `Polyline.NearestPoint(p Point2D) (Point2D, float64)` — closest point + distance (needed for station placement)
- `Polyline.Offset(distance float64) Polyline` — parallel curve offset (needed for shuttle co-location)

**New file**: `solver/pkg/geo/spline_test.go`

- Spline passes through control points (within tolerance)
- Two-point degenerate case → linear interpolation
- Circular waypoints produce near-circular output
- Length, NearestPoint, Offset correctness

### Step 2: New entity/system types

**Modify**: `solver/pkg/scene/scene.go` (lines 29-39, 6-14)

Add constants:
```go
EntityBikePath     EntityType = "bike_path"
EntityShuttleRoute EntityType = "shuttle_route"
EntityStation      EntityType = "station"
EntitySportsField  EntityType = "sports_field"
SystemShuttle      SystemType = "shuttle"
```

### Step 3: Bike path generation

**New file**: `solver/pkg/layout/bike_paths.go`

Types:
```go
type BikePath struct {
    ID        string
    Type      string         // "ring_corridor" | "radial" | "countryside"
    Points    []geo.Point2D  // sampled polyline from spline
    WidthM    float64        // 3m
    ElevatedM float64        // 5m
    Ring      string         // for ring corridors
}
```

Functions:
- `GenerateBikePaths(pods []Pod, adjacency map[string][]string, rings []spec.RingDef) ([]BikePath, *validation.Report)`
- `generateRingCorridorWaypoints(pods []Pod, ringName string, ringInnerR, ringOuterR float64) []geo.Point2D` — sort ring pods by angle, place waypoints at inter-pod midpoints along ring midline radius with deterministic perturbation (hash of pod index, not random) for organic feel. Closed loop.
- `generateRadialWaypoints(pods []Pod, adjacency map[string][]string, rings []spec.RingDef) [][]geo.Point2D` — trace outward through adjacent pods from center to edge, waypoints at inter-pod gaps with intermediate S-curve points. ~8 radials for 32 pods.

Algorithm: waypoints → `CatmullRomSpline(waypoints, 10, 0.5)` → BikePath. Ring corridors for rings with 2+ pods. Radials from center to edge + 500m countryside extension.

**New file**: `solver/pkg/layout/bike_paths_test.go`

### Step 4: Shuttle routes + stations

**New file**: `solver/pkg/layout/shuttle.go`

Types:
```go
type ShuttleRoute struct {
    ID, Type string
    Points   []geo.Point2D
    WidthM   float64
}
type Station struct {
    ID, PodID string
    Position  geo.Point2D
    RouteID   string
}
```

Function: `GenerateShuttleRoutes(bikePaths []BikePath, pods []Pod) ([]ShuttleRoute, []Station, *validation.Report)`

- For each bike path, create parallel shuttle route via `Polyline.Offset(3.0)` — shared infrastructure ribbon
- For each pod, find nearest point on any shuttle route via `Polyline.NearestPoint()`, place station there
- Validate: every pod gets a station within 200m

**New file**: `solver/pkg/layout/shuttle_test.go`

### Step 5: Sports field placement

**New file**: `solver/pkg/layout/sports.go`

Types:
```go
type BufferZone struct {
    ID      string
    PodIDs  [2]string
    Polygon geo.Polygon
    AreaM2  float64
    Centroid geo.Point2D
    Length, Width float64  // oriented bounding box dimensions
    Ring    string
}
type SportsField struct {
    ID         string
    Type       string  // "stadium" | "soccer" | "basketball" | "tennis" | "pickleball"
    Position   geo.Point2D
    Dimensions [2]float64
    Rotation   float64
    BufferID   string
}
```

Functions:
- `PlaceSportsFields(pods []Pod, adjacency map[string][]string, rings []spec.RingDef) ([]SportsField, *validation.Report)`
- `IdentifyBufferZones(pods []Pod, adjacency map[string][]string) []BufferZone` — for each adjacent pair, compute midpoint, direction, perpendicular. Buffer = rectangular zone in the gap between pod boundaries.
- `PlaceStadium(buffers []BufferZone, rings []spec.RingDef) (SportsField, string, bool)` — near ring3/4 boundary, largest buffer fitting 110x75m
- `PlaceSoccerFields(buffers []BufferZone, consumed map[string]bool, rings []spec.RingDef) []SportsField` — 10 fields (105x68m), prefer outer rings
- `PlaceSmallCourts(buffers []BufferZone, consumed map[string]bool) []SportsField` — greedy packing of basketball (28x15), tennis (12x24), pickleball (6x13) in remaining buffers

**New file**: `solver/pkg/layout/sports_test.go`

### Step 6: Scene assembly integration

**Modify**: `solver/pkg/scene/assemble.go`

Extend `Assemble` signature:
```go
func Assemble(s *spec.CitySpec, pods []layout.Pod, buildings []layout.Building,
    paths []layout.PathSegment, bikePaths []layout.BikePath,
    shuttleRoutes []layout.ShuttleRoute, stations []layout.Station,
    segments []routing.Segment, greenZones []layout.Zone,
    sportsFields []layout.SportsField) *Graph
```

Add assembly functions following existing `assemblePaths` pattern (segment each polyline into consecutive point pairs → one Entity per segment with midpoint position, length, yaw rotation):
- `assembleBikePaths` — EntityBikePath, SystemBicycle, LayerSurface, Y=elevated, material "asphalt"
- `assembleShuttleRoutes` — EntityShuttleRoute, SystemShuttle, LayerSurface, Y=0, material "asphalt"
- `assembleStations` — EntityStation, SystemShuttle, LayerSurface, 20x5x10m concrete platform
- `assembleSportsFields` — EntitySportsField, LayerSurface, material "grass"/"court" by type

### Step 7: Pipeline integration

**Modify**: `solver/cmd/cityplanner/run.go` — `runSolve()` (after line 94, before line 97)

Insert:
```go
bikePaths, bikeReport := layout.GenerateBikePaths(pods, adjacency, citySpec.CityZones.Rings)
analyticsReport.Merge(bikeReport)

shuttleRoutes, stations, shuttleReport := layout.GenerateShuttleRoutes(bikePaths, pods)
analyticsReport.Merge(shuttleReport)

sportsFields, sportsReport := layout.PlaceSportsFields(pods, adjacency, citySpec.CityZones.Rings)
analyticsReport.Merge(sportsReport)
```

Update `scene.Assemble` call to pass new data.

### Step 8: Update existing tests

**Modify**: `solver/pkg/scene/assemble_test.go`

- Update `assembleTestGraph()` to call `GenerateBikePaths`, `GenerateShuttleRoutes`, `PlaceSportsFields` and pass results to `Assemble`
- Add new entity types to `TestAssembleHasAllEntityTypes` check list
- Add `SystemShuttle` to `TestAssembleSystemsCoverAllNetworks`

## File Summary

| New files | Purpose |
|-----------|---------|
| `pkg/geo/spline.go` | Polyline, CatmullRomSpline, Length/PointAt/NearestPoint/Offset |
| `pkg/geo/spline_test.go` | Spline and polyline tests |
| `pkg/layout/bike_paths.go` | BikePath type, ring corridor + radial waypoint generation |
| `pkg/layout/bike_paths_test.go` | Bike path generation tests |
| `pkg/layout/shuttle.go` | ShuttleRoute, Station types, co-location + station placement |
| `pkg/layout/shuttle_test.go` | Shuttle route and station tests |
| `pkg/layout/sports.go` | BufferZone, SportsField types, buffer identification + field placement |
| `pkg/layout/sports_test.go` | Sports field placement tests |

| Modified files | Changes |
|----------------|---------|
| `pkg/scene/scene.go` | 4 new EntityType + 1 SystemType constants |
| `pkg/scene/assemble.go` | Extended Assemble signature + 4 new assembly functions |
| `pkg/scene/assemble_test.go` | Updated assembleTestGraph + entity type checks |
| `cmd/cityplanner/run.go` | 3 new generation calls in runSolve pipeline |

## Go 1.19 Constraints

- No generics — use concrete types everywhere
- No builtin min/max — use `math.Min`/`math.Max` or manual `if`
- Benchmarks: `for i := 0; i < b.N; i++` (not `b.Loop()`)
- `sort.Slice` with explicit less functions (no `slices` package)

## Verification

```bash
export PATH=$PATH:/usr/local/go/bin && export TMPDIR=/home/dave/tmp

# After each step:
cd solver && go build ./...

# After all steps:
cd solver && go test ./...

# End-to-end:
cd solver && go build -o cityplanner ./cmd/cityplanner
./cityplanner solve ../examples/default-city | python3 -c "
import json, sys
d = json.load(sys.stdin)
sg = d['scene_graph']
types = {k: len(v) for k, v in sg['groups']['entity_types'].items()}
print('Entity types:', types)
for t in ['bike_path','shuttle_route','station','sports_field']:
    print(f'  {t}: {types.get(t, 0)}')
"
```
