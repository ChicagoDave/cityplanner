package layout

import (
	"fmt"
	"math"

	"github.com/ChicagoDave/cityplanner/pkg/analytics"
	"github.com/ChicagoDave/cityplanner/pkg/geo"
	"github.com/ChicagoDave/cityplanner/pkg/spec"
	"github.com/ChicagoDave/cityplanner/pkg/validation"
)

// Pod represents a laid-out neighborhood pod.
type Pod struct {
	ID               string       `json:"id"`
	Ring             string       `json:"ring"`
	Center           [2]float64   `json:"center"`   // [x, z] in meters
	Boundary         [][2]float64 `json:"boundary"` // polygon vertices
	AreaHa           float64      `json:"area_ha"`
	TargetPopulation int          `json:"target_population"`
}

// BoundaryPolygon returns the pod boundary as a geo.Polygon.
func (p Pod) BoundaryPolygon() geo.Polygon {
	pts := make([]geo.Point2D, len(p.Boundary))
	for i, b := range p.Boundary {
		pts[i] = geo.Pt(b[0], b[1])
	}
	return geo.NewPolygon(pts...)
}

// CenterPoint returns the pod center as a geo.Point2D.
func (p Pod) CenterPoint() geo.Point2D {
	return geo.Pt(p.Center[0], p.Center[1])
}

// LayoutPods computes pod placement using constrained Voronoi tessellation (ADR-005).
// Returns the pods, an adjacency map (pod ID → list of adjacent pod IDs), and
// a validation report.
func LayoutPods(s *spec.CitySpec, params *analytics.ResolvedParameters) ([]Pod, map[string][]string, *validation.Report) {
	report := validation.NewReport()

	// 1. Place seed points along ring midlines.
	var seeds []geo.Point2D
	type seedInfo struct {
		ring       string
		ringIndex  int
		podIndex   int
		population int
	}
	var seedMeta []seedInfo

	for ri, ring := range params.Rings {
		midR := (ring.RadiusFrom + ring.RadiusTo) / 2
		for pi := 0; pi < ring.PodCount; pi++ {
			var seed geo.Point2D
			if ring.PodCount == 1 && ring.RadiusFrom == 0 {
				// Center ring with 1 pod: seed at origin.
				seed = geo.Origin
			} else {
				angle := 2 * math.Pi * float64(pi) / float64(ring.PodCount)
				// Offset each ring's starting angle to stagger pods.
				angle += float64(ri) * math.Pi / 6
				seed = geo.Pt(midR*math.Cos(angle), midR*math.Sin(angle))
			}
			seeds = append(seeds, seed)
			seedMeta = append(seedMeta, seedInfo{
				ring:       ring.Name,
				ringIndex:  ri,
				podIndex:   pi,
				population: ring.PodPopulation,
			})
		}
	}

	if len(seeds) == 0 {
		report.AddError(validation.Result{Level: validation.LevelSpatial, Message: "no pods to lay out (zero pod count)"})
		return nil, nil, report
	}

	// 2. Compute Voronoi tessellation within city boundary.
	outerRadius := params.Rings[len(params.Rings)-1].RadiusTo
	cityBounds := geo.ApproximateCircle(geo.Origin, outerRadius, 128)
	cells := geo.Voronoi(seeds, cityBounds)

	// 3. Clip each cell to its ring boundary and validate walk radius.
	pods := make([]Pod, len(cells))
	walkRadius := s.Pods.WalkRadius

	for i, cell := range cells {
		meta := seedMeta[i]
		ring := params.Rings[meta.ringIndex]

		// Clip to ring annulus.
		clipped := geo.ClipToAnnulus(cell.Polygon, geo.Origin, ring.RadiusFrom, ring.RadiusTo)
		if clipped.IsEmpty() {
			report.AddError(validation.Result{
				Level:   validation.LevelSpatial,
				Message: fmt.Sprintf("pod %s_%d: Voronoi cell empty after ring clipping", meta.ring, meta.podIndex),
			})
			continue
		}

		// Validate walk radius: every vertex should be within walkRadius of the seed.
		maxDist := clipped.MaxDistanceTo(cell.Seed)
		if maxDist > walkRadius*1.05 { // 5% tolerance for polygon approximation
			report.AddWarning(validation.Result{
				Level:   validation.LevelSpatial,
				Message: fmt.Sprintf("pod %s_%d: max distance to boundary %.0fm exceeds walk radius %.0fm", meta.ring, meta.podIndex, maxDist, walkRadius),
			})
		}

		// Build boundary as [][2]float64.
		boundary := make([][2]float64, len(clipped.Vertices))
		for j, v := range clipped.Vertices {
			boundary[j] = [2]float64{v.X, v.Z}
		}

		pods[i] = Pod{
			ID:               fmt.Sprintf("pod_%s_%d", meta.ring, meta.podIndex),
			Ring:             meta.ring,
			Center:           [2]float64{cell.Seed.X, cell.Seed.Z},
			Boundary:         boundary,
			AreaHa:           clipped.Area() / 10000,
			TargetPopulation: meta.population,
		}
	}

	// 4. Build adjacency map from Voronoi neighbors.
	adjacency := make(map[string][]string)
	for i, cell := range cells {
		podID := pods[i].ID
		for _, ni := range cell.Neighbors {
			if ni >= 0 && ni < len(pods) {
				adjacency[podID] = append(adjacency[podID], pods[ni].ID)
			}
		}
	}

	// 5. Validation: check total area coverage.
	totalPodArea := 0.0
	for _, p := range pods {
		totalPodArea += p.AreaHa
	}
	cityAreaHa := math.Pi * outerRadius * outerRadius / 10000
	coverage := totalPodArea / cityAreaHa
	if coverage < 0.90 {
		report.AddWarning(validation.Result{
			Level:   validation.LevelSpatial,
			Message: fmt.Sprintf("pod coverage is only %.1f%% of city area (%.1f ha / %.1f ha)", coverage*100, totalPodArea, cityAreaHa),
		})
	}

	report.AddInfo(validation.Result{
		Level:   validation.LevelSpatial,
		Message: fmt.Sprintf("laid out %d pods across %d rings, total area %.1f ha (%.1f%% coverage)", len(pods), len(params.Rings), totalPodArea, coverage*100),
	})

	return pods, adjacency, report
}
