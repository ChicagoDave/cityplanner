package layout

import (
	"fmt"
	"math"

	"github.com/ChicagoDave/cityplanner/pkg/geo"
)

// Block represents a buildable city block within a zone.
type Block struct {
	ID       string      `json:"id"`
	PodID    string      `json:"pod_id"`
	ZoneType ZoneType    `json:"zone_type"`
	Polygon  geo.Polygon `json:"polygon"`
	AreaM2   float64     `json:"area_m2"`
}

// SubdivideIntoBlocks creates a grid of city blocks within a zone.
// Blocks are approximately blockW x blockD meters, with pathGap meter
// corridors between them. The grid is aligned to the zone's local
// orientation (radial + perpendicular to city center).
func SubdivideIntoBlocks(zone Zone, podCenter geo.Point2D) []Block {
	const (
		blockW  = 60.0 // block width along radial axis (m)
		blockD  = 40.0 // block depth along perpendicular axis (m)
		pathGap = 3.0  // path corridor between blocks (m)
		minArea = 200.0 // minimum block area (m²)
	)

	zonePoly := zone.Polygon
	if zonePoly.IsEmpty() {
		return nil
	}

	// Local coordinate system: outward from city center.
	centroid := zonePoly.Centroid()
	outward := centroid.Normalize()
	if centroid.Length() < 1 {
		outward = geo.Pt(1, 0)
	}
	perp := outward.Perp()

	// Bounding box in local coordinates.
	minU, maxU := math.MaxFloat64, -math.MaxFloat64
	minV, maxV := math.MaxFloat64, -math.MaxFloat64
	for _, v := range zonePoly.Vertices {
		rel := v.Sub(centroid)
		u := rel.Dot(outward)
		vv := rel.Dot(perp)
		if u < minU {
			minU = u
		}
		if u > maxU {
			maxU = u
		}
		if vv < minV {
			minV = vv
		}
		if vv > maxV {
			maxV = vv
		}
	}

	// Grid step sizes (block + gap).
	stepU := blockW + pathGap
	stepV := blockD + pathGap

	var blocks []Block
	blockIdx := 0

	// Generate grid of candidate blocks.
	for u := minU; u+blockW <= maxU; u += stepU {
		for v := minV; v+blockD <= maxV; v += stepV {
			// Block corners in world coordinates.
			p1 := centroid.Add(outward.Scale(u)).Add(perp.Scale(v))
			p2 := centroid.Add(outward.Scale(u + blockW)).Add(perp.Scale(v))
			p3 := centroid.Add(outward.Scale(u + blockW)).Add(perp.Scale(v + blockD))
			p4 := centroid.Add(outward.Scale(u)).Add(perp.Scale(v + blockD))
			blockPoly := geo.NewPolygon(p1, p2, p3, p4)

			// Check if block center is inside the zone.
			blockCenter := blockPoly.Centroid()
			if !zonePoly.Contains(blockCenter) {
				continue
			}
			area := blockPoly.Area()
			if area < minArea {
				continue
			}

			blocks = append(blocks, Block{
				ID:       fmt.Sprintf("%s_block_%d", zone.ID, blockIdx),
				PodID:    zone.PodID,
				ZoneType: zone.Type,
				Polygon:  blockPoly,
				AreaM2:   area,
			})
			blockIdx++
		}
	}

	return blocks
}
