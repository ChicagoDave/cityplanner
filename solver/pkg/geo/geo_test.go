package geo

import (
	"math"
	"testing"
)

const tolerance = 0.01

func approxEqual(a, b, tol float64) bool {
	return math.Abs(a-b) < tol
}

// --- Point2D tests ---

func TestPointDistance(t *testing.T) {
	a := Pt(0, 0)
	b := Pt(3, 4)
	if !approxEqual(a.Distance(b), 5.0, tolerance) {
		t.Errorf("expected distance 5.0, got %f", a.Distance(b))
	}
}

func TestPointAngle(t *testing.T) {
	p := Pt(1, 0)
	if !approxEqual(p.Angle(), 0, tolerance) {
		t.Errorf("expected angle 0, got %f", p.Angle())
	}
	p2 := Pt(0, 1)
	if !approxEqual(p2.Angle(), math.Pi/2, tolerance) {
		t.Errorf("expected angle pi/2, got %f", p2.Angle())
	}
}

func TestPointRotate(t *testing.T) {
	p := Pt(1, 0)
	r := p.Rotate(math.Pi / 2)
	if !approxEqual(r.X, 0, tolerance) || !approxEqual(r.Z, 1, tolerance) {
		t.Errorf("expected (0,1), got (%f,%f)", r.X, r.Z)
	}
}

func TestPointNormalize(t *testing.T) {
	p := Pt(3, 4)
	n := p.Normalize()
	if !approxEqual(n.Length(), 1.0, tolerance) {
		t.Errorf("expected unit length, got %f", n.Length())
	}
}

func TestPointLerp(t *testing.T) {
	a := Pt(0, 0)
	b := Pt(10, 10)
	mid := a.Lerp(b, 0.5)
	if !approxEqual(mid.X, 5, tolerance) || !approxEqual(mid.Z, 5, tolerance) {
		t.Errorf("expected (5,5), got (%f,%f)", mid.X, mid.Z)
	}
}

// --- Polygon tests ---

func TestPolygonAreaSquare(t *testing.T) {
	// 10x10 square
	sq := NewPolygon(Pt(0, 0), Pt(10, 0), Pt(10, 10), Pt(0, 10))
	area := sq.Area()
	if !approxEqual(area, 100, tolerance) {
		t.Errorf("expected area 100, got %f", area)
	}
}

func TestPolygonAreaTriangle(t *testing.T) {
	tri := NewPolygon(Pt(0, 0), Pt(10, 0), Pt(0, 10))
	area := tri.Area()
	if !approxEqual(area, 50, tolerance) {
		t.Errorf("expected area 50, got %f", area)
	}
}

func TestPolygonCentroid(t *testing.T) {
	sq := NewPolygon(Pt(0, 0), Pt(10, 0), Pt(10, 10), Pt(0, 10))
	c := sq.Centroid()
	if !approxEqual(c.X, 5, tolerance) || !approxEqual(c.Z, 5, tolerance) {
		t.Errorf("expected centroid (5,5), got (%f,%f)", c.X, c.Z)
	}
}

func TestPolygonContains(t *testing.T) {
	sq := NewPolygon(Pt(0, 0), Pt(10, 0), Pt(10, 10), Pt(0, 10))
	if !sq.Contains(Pt(5, 5)) {
		t.Error("expected (5,5) inside square")
	}
	if sq.Contains(Pt(15, 5)) {
		t.Error("expected (15,5) outside square")
	}
	if sq.Contains(Pt(-1, 5)) {
		t.Error("expected (-1,5) outside square")
	}
}

func TestPolygonBoundingBox(t *testing.T) {
	sq := NewPolygon(Pt(-5, -3), Pt(10, 0), Pt(7, 12))
	mn, mx := sq.BoundingBox()
	if !approxEqual(mn.X, -5, tolerance) || !approxEqual(mn.Z, -3, tolerance) {
		t.Errorf("expected min (-5,-3), got (%f,%f)", mn.X, mn.Z)
	}
	if !approxEqual(mx.X, 10, tolerance) || !approxEqual(mx.Z, 12, tolerance) {
		t.Errorf("expected max (10,12), got (%f,%f)", mx.X, mx.Z)
	}
}

func TestPolygonPerimeter(t *testing.T) {
	sq := NewPolygon(Pt(0, 0), Pt(10, 0), Pt(10, 10), Pt(0, 10))
	if !approxEqual(sq.Perimeter(), 40, tolerance) {
		t.Errorf("expected perimeter 40, got %f", sq.Perimeter())
	}
}

// --- Clipping tests ---

func TestApproximateCircleArea(t *testing.T) {
	circle := ApproximateCircle(Origin, 100, 128)
	expectedArea := math.Pi * 100 * 100
	if !approxEqual(circle.Area(), expectedArea, expectedArea*0.001) {
		t.Errorf("expected circle area ~%f, got %f", expectedArea, circle.Area())
	}
}

func TestClipToConvexSquareInsideSquare(t *testing.T) {
	outer := NewPolygon(Pt(0, 0), Pt(20, 0), Pt(20, 20), Pt(0, 20))
	inner := NewPolygon(Pt(5, 5), Pt(15, 5), Pt(15, 15), Pt(5, 15))
	clipped := ClipToConvex(inner, outer)
	// Inner is fully inside outer, so result should be identical.
	if !approxEqual(clipped.Area(), 100, tolerance) {
		t.Errorf("expected area 100, got %f", clipped.Area())
	}
}

func TestClipToConvexPartialOverlap(t *testing.T) {
	sq1 := NewPolygon(Pt(0, 0), Pt(10, 0), Pt(10, 10), Pt(0, 10))
	sq2 := NewPolygon(Pt(5, 5), Pt(15, 5), Pt(15, 15), Pt(5, 15))
	clipped := ClipToConvex(sq1, sq2)
	// Overlap is 5x5 = 25.
	if !approxEqual(clipped.Area(), 25, tolerance) {
		t.Errorf("expected area 25, got %f", clipped.Area())
	}
}

func TestClipToConvexNoOverlap(t *testing.T) {
	sq1 := NewPolygon(Pt(0, 0), Pt(5, 0), Pt(5, 5), Pt(0, 5))
	sq2 := NewPolygon(Pt(10, 10), Pt(20, 10), Pt(20, 20), Pt(10, 20))
	clipped := ClipToConvex(sq1, sq2)
	if !clipped.IsEmpty() {
		t.Error("expected empty polygon for non-overlapping squares")
	}
}

func TestClipToAnnulus(t *testing.T) {
	// A large square clipped to an annulus should have area ≈ π(R²-r²).
	sq := NewPolygon(Pt(-1000, -1000), Pt(1000, -1000), Pt(1000, 1000), Pt(-1000, 1000))
	clipped := ClipToAnnulus(sq, Origin, 100, 500)
	expectedArea := math.Pi * (500*500 - 100*100)
	// Allow 5% error due to polygon approximation.
	if !approxEqual(clipped.Area(), expectedArea, expectedArea*0.05) {
		t.Errorf("expected annulus area ~%f, got %f (error: %.1f%%)",
			expectedArea, clipped.Area(), math.Abs(clipped.Area()-expectedArea)/expectedArea*100)
	}
}

// --- Voronoi tests ---

func TestVoronoiTwoPoints(t *testing.T) {
	seeds := []Point2D{Pt(-5, 0), Pt(5, 0)}
	bounds := NewPolygon(Pt(-20, -20), Pt(20, -20), Pt(20, 20), Pt(-20, 20))
	cells := Voronoi(seeds, bounds)

	if len(cells) != 2 {
		t.Fatalf("expected 2 cells, got %d", len(cells))
	}
	// Each cell should have approximately half the area.
	totalArea := bounds.Area()
	for i, c := range cells {
		if c.Polygon.IsEmpty() {
			t.Errorf("cell %d is empty", i)
			continue
		}
		if !approxEqual(c.Polygon.Area(), totalArea/2, totalArea*0.05) {
			t.Errorf("cell %d area %f, expected ~%f", i, c.Polygon.Area(), totalArea/2)
		}
	}
}

func TestVoronoiFourPointsSquare(t *testing.T) {
	seeds := []Point2D{Pt(-5, -5), Pt(5, -5), Pt(5, 5), Pt(-5, 5)}
	bounds := NewPolygon(Pt(-20, -20), Pt(20, -20), Pt(20, 20), Pt(-20, 20))
	cells := Voronoi(seeds, bounds)

	if len(cells) != 4 {
		t.Fatalf("expected 4 cells, got %d", len(cells))
	}
	totalArea := bounds.Area()
	for i, c := range cells {
		if c.Polygon.IsEmpty() {
			t.Errorf("cell %d is empty", i)
			continue
		}
		expectedArea := totalArea / 4
		if !approxEqual(c.Polygon.Area(), expectedArea, expectedArea*0.1) {
			t.Errorf("cell %d area %f, expected ~%f", i, c.Polygon.Area(), expectedArea)
		}
	}
	// Each cell should have 2-3 neighbors.
	for i, c := range cells {
		if len(c.Neighbors) < 2 {
			t.Errorf("cell %d has only %d neighbors, expected >= 2", i, len(c.Neighbors))
		}
	}
}

func TestVoronoiSinglePoint(t *testing.T) {
	seeds := []Point2D{Pt(0, 0)}
	bounds := ApproximateCircle(Origin, 100, 64)
	cells := Voronoi(seeds, bounds)

	if len(cells) != 1 {
		t.Fatalf("expected 1 cell, got %d", len(cells))
	}
	// Cell should be the entire bounds.
	if !approxEqual(cells[0].Polygon.Area(), bounds.Area(), bounds.Area()*0.01) {
		t.Errorf("single cell area %f, expected ~%f", cells[0].Polygon.Area(), bounds.Area())
	}
}

func TestVoronoiSixPointsCityLayout(t *testing.T) {
	// Simulate the city layout: 1 center + 2 middle + 3 edge seeds.
	seeds := []Point2D{
		Pt(0, 0),                                        // center
		Pt(450 * math.Cos(0), 450 * math.Sin(0)),        // middle 1
		Pt(450 * math.Cos(math.Pi), 450 * math.Sin(math.Pi)), // middle 2
		Pt(750 * math.Cos(0), 750 * math.Sin(0)),                          // edge 1
		Pt(750 * math.Cos(2*math.Pi/3), 750 * math.Sin(2*math.Pi/3)),     // edge 2
		Pt(750 * math.Cos(4*math.Pi/3), 750 * math.Sin(4*math.Pi/3)),     // edge 3
	}
	bounds := ApproximateCircle(Origin, 900, 128)
	cells := Voronoi(seeds, bounds)

	if len(cells) != 6 {
		t.Fatalf("expected 6 cells, got %d", len(cells))
	}
	// All cells should have non-zero area.
	totalArea := 0.0
	for i, c := range cells {
		if c.Polygon.IsEmpty() {
			t.Errorf("cell %d is empty", i)
			continue
		}
		if c.Polygon.Area() < 1000 {
			t.Errorf("cell %d has very small area: %f", i, c.Polygon.Area())
		}
		totalArea += c.Polygon.Area()
	}
	// Total area should approximately match bounds.
	boundsArea := bounds.Area()
	if !approxEqual(totalArea, boundsArea, boundsArea*0.1) {
		t.Errorf("total cell area %f, expected ~%f", totalArea, boundsArea)
	}
}

// --- Line-circle intersection tests ---

func TestLineCircleIntersections(t *testing.T) {
	// Horizontal line through center of circle.
	pts := lineCircleIntersections(Pt(-10, 0), Pt(10, 0), Origin, 5)
	if len(pts) != 2 {
		t.Fatalf("expected 2 intersections, got %d", len(pts))
	}
	// Should intersect at (-5,0) and (5,0).
	for _, p := range pts {
		if !approxEqual(p.Distance(Origin), 5, tolerance) {
			t.Errorf("intersection at distance %f from origin, expected 5", p.Distance(Origin))
		}
	}
}

func TestLineCircleNoIntersection(t *testing.T) {
	pts := lineCircleIntersections(Pt(-10, 10), Pt(10, 10), Origin, 5)
	if len(pts) != 0 {
		t.Errorf("expected 0 intersections, got %d", len(pts))
	}
}
