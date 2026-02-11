package layout

import (
	"fmt"
	"math"

	"github.com/ChicagoDave/cityplanner/pkg/geo"
	"github.com/ChicagoDave/cityplanner/pkg/validation"
)

// ShuttleRoute represents a ground-level automated shuttle route.
type ShuttleRoute struct {
	ID     string        `json:"id"`
	Type   string        `json:"type"` // "ring_corridor" or "radial"
	Points []geo.Point2D `json:"points"`
	WidthM float64       `json:"width_m"`
}

// Station represents a combined shuttle + bike mobility hub.
type Station struct {
	ID       string     `json:"id"`
	PodID    string     `json:"pod_id"`
	Position geo.Point2D `json:"position"`
	RouteID  string     `json:"route_id"`
}

// GenerateShuttleRoutes creates shuttle routes co-located with bike paths
// and places one station per pod at the nearest point on any shuttle route.
func GenerateShuttleRoutes(bikePaths []BikePath, pods []Pod) ([]ShuttleRoute, []Station, *validation.Report) {
	report := validation.NewReport()

	// Create shuttle routes by offsetting each bike path.
	var routes []ShuttleRoute
	for i, bp := range bikePaths {
		pl := geo.Polyline{Points: bp.Points}
		offset := pl.Offset(4.0) // 4m lateral offset from bike path centerline

		routes = append(routes, ShuttleRoute{
			ID:     fmt.Sprintf("shuttle_%s_%d", bp.Type, i),
			Type:   bp.Type,
			Points: offset.Points,
			WidthM: 3.0,
		})
	}

	// Place one station per pod at the nearest point on any shuttle route.
	var stations []Station
	for _, pod := range pods {
		center := pod.CenterPoint()
		bestDist := math.MaxFloat64
		var bestPt geo.Point2D
		bestRouteID := ""

		for _, route := range routes {
			pl := geo.Polyline{Points: route.Points}
			pt, dist := pl.NearestPoint(center)
			if dist < bestDist {
				bestDist = dist
				bestPt = pt
				bestRouteID = route.ID
			}
		}

		if bestRouteID == "" {
			report.AddWarning(validation.Result{
				Level:   validation.LevelSpatial,
				Message: fmt.Sprintf("pod %s: no shuttle route found for station placement", pod.ID),
			})
			continue
		}

		stations = append(stations, Station{
			ID:       fmt.Sprintf("station_%s", pod.ID),
			PodID:    pod.ID,
			Position: bestPt,
			RouteID:  bestRouteID,
		})

		if bestDist > 200 {
			report.AddWarning(validation.Result{
				Level:   validation.LevelSpatial,
				Message: fmt.Sprintf("pod %s: station is %.0fm from pod center (>200m)", pod.ID, bestDist),
			})
		}
	}

	report.AddInfo(validation.Result{
		Level:   validation.LevelSpatial,
		Message: fmt.Sprintf("generated %d shuttle routes and %d stations", len(routes), len(stations)),
	})

	return routes, stations, report
}
