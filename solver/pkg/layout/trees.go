package layout

import (
	"fmt"
	"math"

	"github.com/ChicagoDave/cityplanner/pkg/geo"
	"github.com/ChicagoDave/cityplanner/pkg/validation"
)

// Tree represents a placed tree entity in the scene.
type Tree struct {
	ID       string      `json:"id"`
	PodID    string      `json:"pod_id,omitempty"`
	Position geo.Point2D `json:"position"`
	CanopyD  float64     `json:"canopy_diameter"` // 4-8m
	Height   float64     `json:"height"`          // 6-12m
	Context  string      `json:"context"`         // "park", "path", "plaza"
}

const (
	parkTreeSpacing = 20.0 // meters between park trees (grid)
	pathTreeSpacing = 25.0 // meters between path trees
	pathTreeOffset  = 3.0  // perpendicular offset from path center
	plazaTreeOffset = 3.0  // meters outside plaza edge
)

// PlaceTrees generates trees in three contexts: parks, paths, and plazas.
func PlaceTrees(
	pods []Pod,
	greenZones []Zone,
	paths []PathSegment,
	bikePaths []BikePath,
	plazas []Plaza,
) ([]Tree, *validation.Report) {
	report := validation.NewReport()
	var trees []Tree
	idx := 0

	// 1. Park trees: grid fill within green zone polygons.
	for _, z := range greenZones {
		parkTrees := placeParkTrees(z, &idx)
		trees = append(trees, parkTrees...)
	}

	// 2. Path trees: along pedestrian paths (ground level only; bike paths
	//    are elevated so ground-level trees aren't placed beside them).
	for _, p := range paths {
		pathTrees := placePathTrees(p, &idx)
		trees = append(trees, pathTrees...)
	}

	// 3. Plaza perimeter trees.
	for _, pl := range plazas {
		plTrees := plazaPerimeterTrees(pl, &idx)
		trees = append(trees, plTrees...)
	}

	report.AddInfo(validation.Result{
		Level:   validation.LevelSpatial,
		Message: fmt.Sprintf("placed %d trees (park: green zones, path: %d segments, plaza: %d perimeters)",
			len(trees), len(paths), len(plazas)),
	})
	return trees, report
}

// placeParkTrees fills a green zone polygon with trees on a 10m grid.
func placeParkTrees(z Zone, idx *int) []Tree {
	minPt, maxPt := z.Polygon.BoundingBox()
	var trees []Tree

	for x := minPt.X; x <= maxPt.X; x += parkTreeSpacing {
		for zz := minPt.Z; zz <= maxPt.Z; zz += parkTreeSpacing {
			pt := geo.Point2D{X: x, Z: zz}
			if !z.Polygon.Contains(pt) {
				continue
			}
			h := 8.0 + 4.0*math.Abs(math.Sin(x*0.31+zz*0.47))
			c := 5.0 + 3.0*math.Abs(math.Sin(x*0.53+zz*0.29))

			trees = append(trees, Tree{
				ID:       fmt.Sprintf("tree_park_%05d", *idx),
				PodID:    z.PodID,
				Position: pt,
				CanopyD:  c,
				Height:   h,
				Context:  "park",
			})
			*idx++
		}
	}
	return trees
}

// placePathTrees places trees along a pedestrian path segment at regular intervals.
func placePathTrees(p PathSegment, idx *int) []Tree {
	dx := p.End.X - p.Start.X
	dz := p.End.Z - p.Start.Z
	length := math.Hypot(dx, dz)
	if length < pathTreeSpacing {
		return nil
	}

	// Unit direction and perpendicular.
	ux, uz := dx/length, dz/length
	px, pz := -uz, ux // perpendicular (left side)

	var trees []Tree
	for d := pathTreeSpacing / 2; d < length-pathTreeSpacing/2; d += pathTreeSpacing {
		t := d / length
		x := p.Start.X + dx*t + px*pathTreeOffset
		z := p.Start.Z + dz*t + pz*pathTreeOffset

		h := 6.0 + 4.0*math.Abs(math.Sin(x*0.37+z*0.41))
		c := 4.0 + 2.0*math.Abs(math.Sin(x*0.59+z*0.31))

		trees = append(trees, Tree{
			ID:       fmt.Sprintf("tree_path_%05d", *idx),
			PodID:    p.PodID,
			Position: geo.Point2D{X: x, Z: z},
			CanopyD:  c,
			Height:   h,
			Context:  "path",
		})
		*idx++
	}
	return trees
}

// plazaPerimeterTrees places 8 trees around a plaza (4 corners + 4 midpoints).
func plazaPerimeterTrees(pl Plaza, idx *int) []Tree {
	cos := math.Cos(pl.Rotation)
	sin := math.Sin(pl.Rotation)
	hw := pl.Width/2 + plazaTreeOffset
	hd := pl.Depth/2 + plazaTreeOffset

	// 8 positions: corners and midpoints of each side.
	offsets := [][2]float64{
		{-hw, -hd}, {hw, -hd}, {hw, hd}, {-hw, hd}, // corners
		{0, -hd}, {hw, 0}, {0, hd}, {-hw, 0}, // midpoints
	}

	var trees []Tree
	for _, off := range offsets {
		// Rotate offset by plaza rotation.
		rx := off[0]*cos - off[1]*sin
		rz := off[0]*sin + off[1]*cos
		x := pl.Position.X + rx
		z := pl.Position.Z + rz

		h := 8.0 + 2.0*math.Abs(math.Sin(x*0.33+z*0.51))
		c := 5.0 + 2.0*math.Abs(math.Sin(x*0.47+z*0.39))

		trees = append(trees, Tree{
			ID:       fmt.Sprintf("tree_plaza_%05d", *idx),
			PodID:    pl.PodID,
			Position: geo.Point2D{X: x, Z: z},
			CanopyD:  c,
			Height:   h,
			Context:  "plaza",
		})
		*idx++
	}
	return trees
}
