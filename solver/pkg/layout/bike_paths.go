package layout

import (
	"fmt"
	"math"
	"sort"

	"github.com/ChicagoDave/cityplanner/pkg/geo"
	"github.com/ChicagoDave/cityplanner/pkg/spec"
	"github.com/ChicagoDave/cityplanner/pkg/validation"
)

// BikePath represents a generated bike path through the city.
type BikePath struct {
	ID        string        `json:"id"`
	Type      string        `json:"type"` // "ring_corridor" or "radial"
	Points    []geo.Point2D `json:"points"`
	WidthM    float64       `json:"width_m"`
	ElevatedM float64       `json:"elevated_m"`
	Ring      string        `json:"ring,omitempty"`
}

// GenerateBikePaths creates the city-wide elevated bike path network.
// Ring corridors loop around each ring through inter-pod green space.
// Radial paths connect center to edge with countryside extensions.
func GenerateBikePaths(pods []Pod, adjacency map[string][]string, rings []spec.RingDef) ([]BikePath, *validation.Report) {
	report := validation.NewReport()
	var paths []BikePath

	// Group pods by ring.
	ringPods := make(map[string][]Pod)
	for _, pod := range pods {
		ringPods[pod.Ring] = append(ringPods[pod.Ring], pod)
	}

	// Generate ring corridor paths for rings with 2+ pods.
	pathIdx := 0
	for _, ring := range rings {
		rPods := ringPods[ring.Name]
		if len(rPods) < 2 {
			continue
		}

		waypoints := generateRingCorridorWaypoints(rPods, ring)
		if len(waypoints) < 3 {
			continue
		}

		spline := geo.CatmullRomSplineClosed(waypoints, 10, 0.5)
		paths = append(paths, BikePath{
			ID:        fmt.Sprintf("bike_ring_%s_%d", ring.Name, pathIdx),
			Type:      "ring_corridor",
			Points:    spline.Points,
			WidthM:    3.0,
			ElevatedM: 5.0,
			Ring:      ring.Name,
		})
		pathIdx++
	}

	// Generate radial paths from center to edge.
	radials := generateRadialWaypoints(pods, rings)
	for i, waypoints := range radials {
		if len(waypoints) < 2 {
			continue
		}

		spline := geo.CatmullRomSpline(waypoints, 10, 0.5)
		paths = append(paths, BikePath{
			ID:        fmt.Sprintf("bike_radial_%d", i),
			Type:      "radial",
			Points:    spline.Points,
			WidthM:    3.0,
			ElevatedM: 5.0,
		})
	}

	report.AddInfo(validation.Result{
		Level:   validation.LevelSpatial,
		Message: fmt.Sprintf("generated %d bike paths (%d ring corridors, %d radials)", len(paths), pathIdx, len(radials)),
	})

	return paths, report
}

// generateRingCorridorWaypoints creates waypoints for a closed loop bike path
// through the inter-pod green space of a ring.
func generateRingCorridorWaypoints(pods []Pod, ring spec.RingDef) []geo.Point2D {
	// Sort pods by angle from origin for consistent ordering.
	type podAngle struct {
		pod   Pod
		angle float64
	}
	sorted := make([]podAngle, len(pods))
	for i, p := range pods {
		sorted[i] = podAngle{
			pod:   p,
			angle: math.Atan2(p.Center[1], p.Center[0]),
		}
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].angle < sorted[j].angle
	})

	midRadius := (ring.RadiusFrom + ring.RadiusTo) / 2
	var waypoints []geo.Point2D

	for i := 0; i < len(sorted); i++ {
		cur := sorted[i].pod.CenterPoint()
		next := sorted[(i+1)%len(sorted)].pod.CenterPoint()

		// Place waypoint at the midpoint between consecutive pods,
		// projected to the ring midline radius.
		mid := geo.MidPoint(cur, next)
		dist := mid.Length()
		if dist < 1 {
			// Fallback: use midRadius on the bisector angle.
			angle := (sorted[i].angle + sorted[(i+1)%len(sorted)].angle) / 2
			waypoints = append(waypoints, geo.Pt(
				midRadius*math.Cos(angle),
				midRadius*math.Sin(angle),
			))
			continue
		}

		// Project to midRadius with deterministic perturbation for organic feel.
		// Use pod index as seed for consistent output.
		perturbation := 0.05 * math.Sin(float64(i)*2.3+0.7)
		targetR := midRadius * (1.0 + perturbation)
		wp := mid.Scale(targetR / dist)
		waypoints = append(waypoints, wp)
	}

	return waypoints
}

// generateRadialWaypoints creates waypoints for radial bike paths from center
// to edge with countryside extensions.
func generateRadialWaypoints(pods []Pod, rings []spec.RingDef) [][]geo.Point2D {
	if len(rings) == 0 {
		return nil
	}

	outerRadius := rings[len(rings)-1].RadiusTo
	countrysideExtension := 500.0

	// Determine number of radials based on pod count in outermost ring.
	// Find the outermost ring with pods.
	numRadials := 6 // default
	ringPods := make(map[string][]Pod)
	for _, pod := range pods {
		ringPods[pod.Ring] = append(ringPods[pod.Ring], pod)
	}
	for i := len(rings) - 1; i >= 0; i-- {
		rp := ringPods[rings[i].Name]
		if len(rp) >= 3 {
			numRadials = len(rp)
			if numRadials > 12 {
				numRadials = 12
			}
			break
		}
	}

	var radials [][]geo.Point2D

	for r := 0; r < numRadials; r++ {
		baseAngle := 2 * math.Pi * float64(r) / float64(numRadials)
		// Slight S-curve perturbation per radial.
		var waypoints []geo.Point2D

		// Start near center.
		startR := 50.0
		if len(rings) > 0 && rings[0].RadiusTo < startR {
			startR = rings[0].RadiusTo * 0.5
		}
		waypoints = append(waypoints, geo.Pt(
			startR*math.Cos(baseAngle),
			startR*math.Sin(baseAngle),
		))

		// Waypoint at each ring boundary, offset to thread through green space.
		for ri, ring := range rings {
			midR := (ring.RadiusFrom + ring.RadiusTo) / 2
			// Deterministic angular offset for S-curve feel.
			offset := 0.03 * math.Sin(float64(ri)*1.7+float64(r)*0.5)
			angle := baseAngle + offset
			waypoints = append(waypoints, geo.Pt(
				midR*math.Cos(angle),
				midR*math.Sin(angle),
			))
		}

		// Edge of city.
		waypoints = append(waypoints, geo.Pt(
			outerRadius*math.Cos(baseAngle),
			outerRadius*math.Sin(baseAngle),
		))

		// Countryside extension.
		extR := outerRadius + countrysideExtension
		waypoints = append(waypoints, geo.Pt(
			extR*math.Cos(baseAngle),
			extR*math.Sin(baseAngle),
		))

		radials = append(radials, waypoints)
	}

	return radials
}
