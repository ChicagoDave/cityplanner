package layout

import (
	"fmt"
	"math"

	"github.com/ChicagoDave/cityplanner/pkg/geo"
)

// GeneratePaths creates the pedestrian/bicycle path network for a pod.
//
// Three path types:
//   - Spine: main path from pod center through each zone (4m wide)
//   - Connectors: perpendicular to spine at regular intervals (3m wide)
//   - Inter-pod: from pod center toward each adjacent pod center (4m wide)
func GeneratePaths(pod Pod, zones []Zone, adjacentCenters map[string]geo.Point2D) []PathSegment {
	var paths []PathSegment
	center := pod.CenterPoint()
	podPoly := pod.BoundaryPolygon()
	pathIdx := 0

	// 1. Spine paths: from pod center outward in the radial direction
	// and inward toward city center.
	outward := center.Normalize()
	if center.Length() < 1 {
		outward = geo.Pt(1, 0)
	}

	// Find extent of pod along outward direction.
	minProj := math.MaxFloat64
	maxProj := -math.MaxFloat64
	for _, v := range podPoly.Vertices {
		proj := v.Sub(center).Dot(outward)
		if proj < minProj {
			minProj = proj
		}
		if proj > maxProj {
			maxProj = proj
		}
	}

	// Outward spine.
	spineOut := center.Add(outward.Scale(maxProj * 0.95))
	paths = append(paths, PathSegment{
		ID:     fmt.Sprintf("%s_spine_%d", pod.ID, pathIdx),
		PodID:  pod.ID,
		Start:  center,
		End:    spineOut,
		WidthM: 4,
		Type:   "spine",
	})
	pathIdx++

	// Inward spine.
	spineIn := center.Add(outward.Scale(minProj * 0.95))
	paths = append(paths, PathSegment{
		ID:     fmt.Sprintf("%s_spine_%d", pod.ID, pathIdx),
		PodID:  pod.ID,
		Start:  center,
		End:    spineIn,
		WidthM: 4,
		Type:   "spine",
	})
	pathIdx++

	// 2. Perpendicular connectors along the spine at ~80m intervals.
	perp := outward.Perp()
	// Find perpendicular extent.
	minPerpProj := math.MaxFloat64
	maxPerpProj := -math.MaxFloat64
	for _, v := range podPoly.Vertices {
		proj := v.Sub(center).Dot(perp)
		if proj < minPerpProj {
			minPerpProj = proj
		}
		if proj > maxPerpProj {
			maxPerpProj = proj
		}
	}
	perpExtent := (maxPerpProj - minPerpProj) * 0.45

	connectorSpacing := 80.0
	spineLen := maxProj - minProj
	numConnectors := int(spineLen / connectorSpacing)
	for i := 1; i <= numConnectors; i++ {
		t := float64(i) / float64(numConnectors+1)
		pos := minProj + t*spineLen
		connCenter := center.Add(outward.Scale(pos))
		connStart := connCenter.Add(perp.Scale(-perpExtent))
		connEnd := connCenter.Add(perp.Scale(perpExtent))
		paths = append(paths, PathSegment{
			ID:     fmt.Sprintf("%s_conn_%d", pod.ID, pathIdx),
			PodID:  pod.ID,
			Start:  connStart,
			End:    connEnd,
			WidthM: 3,
			Type:   "connector",
		})
		pathIdx++
	}

	// 3. Inter-pod connectors: from pod center toward each adjacent pod center.
	for adjID, adjCenter := range adjacentCenters {
		dir := adjCenter.Sub(center)
		if dir.Length() < 1 {
			continue
		}
		dirNorm := dir.Normalize()
		// Extend from center to ~95% of the distance to the pod boundary.
		maxT := 0.0
		for _, v := range podPoly.Vertices {
			proj := v.Sub(center).Dot(dirNorm)
			if proj > maxT {
				maxT = proj
			}
		}
		endPt := center.Add(dirNorm.Scale(maxT * 0.95))
		paths = append(paths, PathSegment{
			ID:     fmt.Sprintf("%s_inter_%s_%d", pod.ID, adjID, pathIdx),
			PodID:  pod.ID,
			Start:  center,
			End:    endPt,
			WidthM: 4,
			Type:   "inter_pod",
		})
		pathIdx++
	}

	return paths
}
