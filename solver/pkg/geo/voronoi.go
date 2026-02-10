package geo

import (
	"math"
	"sort"
)

// VoronoiCell represents one cell in a Voronoi diagram.
type VoronoiCell struct {
	SeedIndex int     // index into the original seed array
	Seed      Point2D // the seed point
	Polygon   Polygon // the cell boundary
	Neighbors []int   // indices of neighboring seed points
}

// Voronoi computes the Voronoi diagram of the given seed points,
// clipped to the given bounding polygon.
//
// Uses half-plane intersection for cell geometry (robust for small n)
// and Bowyer-Watson Delaunay for neighbor detection.
func Voronoi(seeds []Point2D, bounds Polygon) []VoronoiCell {
	n := len(seeds)
	if n == 0 {
		return nil
	}
	if n == 1 {
		return []VoronoiCell{{
			SeedIndex: 0,
			Seed:      seeds[0],
			Polygon:   bounds,
		}}
	}

	// Compute cell geometry via half-plane intersection.
	cells := make([]VoronoiCell, n)
	for i := 0; i < n; i++ {
		cells[i] = VoronoiCell{
			SeedIndex: i,
			Seed:      seeds[i],
			Polygon:   voronoiCellByHalfPlanes(i, seeds, bounds),
		}
	}

	// Compute neighbors via Bowyer-Watson Delaunay triangulation.
	neighbors := delaunayNeighbors(seeds, bounds)
	for i := 0; i < n; i++ {
		cells[i].Neighbors = neighbors[i]
	}

	return cells
}

// voronoiCellByHalfPlanes computes a Voronoi cell by intersecting half-planes.
// For each other seed, clip the bounds to the half-plane closer to seed[i].
func voronoiCellByHalfPlanes(seedIdx int, seeds []Point2D, bounds Polygon) Polygon {
	cell := bounds
	seed := seeds[seedIdx]
	for j, other := range seeds {
		if j == seedIdx {
			continue
		}
		mid := MidPoint(seed, other)
		dir := other.Sub(seed).Perp()
		cell = clipToHalfPlane(cell, mid, mid.Add(dir))
		if cell.IsEmpty() {
			break
		}
	}
	return cell
}

// clipToHalfPlane clips a polygon to the left side of the directed line from a to b.
func clipToHalfPlane(poly Polygon, a, b Point2D) Polygon {
	if poly.IsEmpty() {
		return Polygon{}
	}
	n := len(poly.Vertices)
	output := make([]Point2D, 0, n)
	for i := 0; i < n; i++ {
		curr := poly.Vertices[i]
		next := poly.Vertices[(i+1)%n]
		currInside := isInsideEdge(curr, a, b)
		nextInside := isInsideEdge(next, a, b)

		if currInside && nextInside {
			output = append(output, next)
		} else if currInside && !nextInside {
			if ix, ok := lineIntersection(curr, next, a, b); ok {
				output = append(output, ix)
			}
		} else if !currInside && nextInside {
			if ix, ok := lineIntersection(curr, next, a, b); ok {
				output = append(output, ix)
			}
			output = append(output, next)
		}
	}
	if len(output) < 3 {
		return Polygon{}
	}
	return Polygon{Vertices: output}
}

// delaunayNeighbors computes Delaunay triangulation and returns adjacency.
// neighbors[i] is a sorted list of seed indices adjacent to seed i.
func delaunayNeighbors(seeds []Point2D, bounds Polygon) [][]int {
	n := len(seeds)
	if n < 2 {
		return make([][]int, n)
	}

	// Jitter to avoid degeneracy.
	pts := make([]Point2D, n)
	for i, s := range seeds {
		pts[i] = Point2D{
			X: s.X + float64(i)*1e-8,
			Z: s.Z + float64(i)*1e-8,
		}
	}

	// Super-triangle.
	bbMin, bbMax := bounds.BoundingBox()
	dx := bbMax.X - bbMin.X
	dz := bbMax.Z - bbMin.Z
	maxD := math.Max(dx, dz) * 4

	superA := Point2D{bbMin.X - maxD, bbMin.Z - maxD}
	superB := Point2D{bbMax.X + maxD, bbMin.Z - maxD}
	superC := Point2D{(bbMin.X + bbMax.X) / 2, bbMax.Z + maxD}

	allPts := make([]Point2D, n+3)
	copy(allPts, pts)
	allPts[n] = superA
	allPts[n+1] = superB
	allPts[n+2] = superC

	type triangle struct{ v [3]int }
	triangles := []triangle{{v: [3]int{n, n + 1, n + 2}}}

	for pi := 0; pi < n; pi++ {
		p := allPts[pi]
		bad := make([]int, 0)
		for ti, t := range triangles {
			if inCircumcircle(p, allPts[t.v[0]], allPts[t.v[1]], allPts[t.v[2]]) {
				bad = append(bad, ti)
			}
		}

		type edge struct{ a, b int }
		edgeCount := make(map[edge]int)
		for _, ti := range bad {
			t := triangles[ti]
			for k := 0; k < 3; k++ {
				e := edge{t.v[k], t.v[(k+1)%3]}
				if e.a > e.b {
					e.a, e.b = e.b, e.a
				}
				edgeCount[e]++
			}
		}

		boundaryEdges := make([]edge, 0)
		for _, ti := range bad {
			t := triangles[ti]
			for k := 0; k < 3; k++ {
				e := edge{t.v[k], t.v[(k+1)%3]}
				eNorm := e
				if eNorm.a > eNorm.b {
					eNorm.a, eNorm.b = eNorm.b, eNorm.a
				}
				if edgeCount[eNorm] == 1 {
					boundaryEdges = append(boundaryEdges, e)
				}
			}
		}

		sort.Sort(sort.Reverse(sort.IntSlice(bad)))
		for _, ti := range bad {
			triangles[ti] = triangles[len(triangles)-1]
			triangles = triangles[:len(triangles)-1]
		}

		for _, e := range boundaryEdges {
			triangles = append(triangles, triangle{v: [3]int{e.a, e.b, pi}})
		}
	}

	// Extract neighbor map from non-super triangles.
	neighborSet := make([]map[int]bool, n)
	for i := range neighborSet {
		neighborSet[i] = make(map[int]bool)
	}
	for _, t := range triangles {
		if t.v[0] >= n || t.v[1] >= n || t.v[2] >= n {
			continue
		}
		for k := 0; k < 3; k++ {
			a, b := t.v[k], t.v[(k+1)%3]
			neighborSet[a][b] = true
			neighborSet[b][a] = true
		}
	}

	result := make([][]int, n)
	for i, ns := range neighborSet {
		keys := make([]int, 0, len(ns))
		for k := range ns {
			keys = append(keys, k)
		}
		sort.Ints(keys)
		result[i] = keys
	}
	return result
}

// inCircumcircle returns true if point p is inside the circumcircle of
// triangle (a,b,c). Uses the determinant test.
func inCircumcircle(p, a, b, c Point2D) bool {
	ax, az := a.X-p.X, a.Z-p.Z
	bx, bz := b.X-p.X, b.Z-p.Z
	cx, cz := c.X-p.X, c.Z-p.Z

	det := ax*(bz*(cx*cx+cz*cz)-cz*(bx*bx+bz*bz)) -
		az*(bx*(cx*cx+cz*cz)-cx*(bx*bx+bz*bz)) +
		(ax*ax+az*az)*(bx*cz-cx*bz)

	orient := (b.X-a.X)*(c.Z-a.Z) - (b.Z-a.Z)*(c.X-a.X)
	if orient < 0 {
		det = -det
	}
	return det > 0
}
