package geo

import (
	"math"
	"testing"
)

func TestCatmullRomSplinePassesThroughControlPoints(t *testing.T) {
	pts := []Point2D{Pt(0, 0), Pt(100, 0), Pt(200, 100), Pt(300, 100)}
	spline := CatmullRomSpline(pts, 20, 0.5)

	// First point should be the first control point.
	if spline.Points[0].Distance(pts[0]) > 0.1 {
		t.Errorf("spline does not start at first control point: got %v", spline.Points[0])
	}
	// Last point should be the last control point.
	last := spline.Points[len(spline.Points)-1]
	if last.Distance(pts[len(pts)-1]) > 0.1 {
		t.Errorf("spline does not end at last control point: got %v", last)
	}

	// Interior control points should be within tolerance.
	for i := 1; i < len(pts)-1; i++ {
		pl := Polyline{Points: spline.Points}
		_, dist := pl.NearestPoint(pts[i])
		if dist > 5.0 {
			t.Errorf("control point %d is %.1fm from spline (>5m)", i, dist)
		}
	}
}

func TestCatmullRomSplineTwoPointsLinear(t *testing.T) {
	pts := []Point2D{Pt(0, 0), Pt(100, 0)}
	spline := CatmullRomSpline(pts, 10, 0.5)

	if len(spline.Points) != 11 {
		t.Fatalf("expected 11 points for 2-point spline with 10 samples, got %d", len(spline.Points))
	}

	// All points should be on the line y=0.
	for i, p := range spline.Points {
		if math.Abs(p.Z) > 0.01 {
			t.Errorf("point %d has Z=%.3f, expected 0 (linear)", i, p.Z)
		}
	}
}

func TestCatmullRomSplineClosedLoop(t *testing.T) {
	// Square waypoints.
	pts := []Point2D{Pt(100, 0), Pt(0, 100), Pt(-100, 0), Pt(0, -100)}
	spline := CatmullRomSplineClosed(pts, 10, 0.5)

	if len(spline.Points) < 40 {
		t.Fatalf("expected at least 40 points for closed loop, got %d", len(spline.Points))
	}

	// First and last points should be the same (closed loop).
	first := spline.Points[0]
	last := spline.Points[len(spline.Points)-1]
	if first.Distance(last) > 0.1 {
		t.Errorf("closed loop not closed: first=%v last=%v", first, last)
	}
}

func TestPolylineLength(t *testing.T) {
	pl := NewPolyline(Pt(0, 0), Pt(100, 0), Pt(100, 100))
	expected := 200.0
	if math.Abs(pl.Length()-expected) > 0.01 {
		t.Errorf("expected length %.1f, got %.1f", expected, pl.Length())
	}
}

func TestPolylinePointAt(t *testing.T) {
	pl := NewPolyline(Pt(0, 0), Pt(100, 0))

	mid := pl.PointAt(0.5)
	if mid.Distance(Pt(50, 0)) > 0.01 {
		t.Errorf("expected midpoint (50,0), got %v", mid)
	}

	start := pl.PointAt(0)
	if start.Distance(Pt(0, 0)) > 0.01 {
		t.Errorf("expected start (0,0), got %v", start)
	}

	end := pl.PointAt(1)
	if end.Distance(Pt(100, 0)) > 0.01 {
		t.Errorf("expected end (100,0), got %v", end)
	}
}

func TestPolylineNearestPoint(t *testing.T) {
	pl := NewPolyline(Pt(0, 0), Pt(100, 0))

	pt, dist := pl.NearestPoint(Pt(50, 10))
	if math.Abs(dist-10) > 0.01 {
		t.Errorf("expected distance 10, got %.2f", dist)
	}
	if pt.Distance(Pt(50, 0)) > 0.01 {
		t.Errorf("expected nearest (50,0), got %v", pt)
	}
}

func TestPolylineOffset(t *testing.T) {
	pl := NewPolyline(Pt(0, 0), Pt(100, 0), Pt(200, 0))
	offset := pl.Offset(10)

	if len(offset.Points) != 3 {
		t.Fatalf("expected 3 offset points, got %d", len(offset.Points))
	}

	// For a straight horizontal line, offset should shift in Z direction.
	for i, p := range offset.Points {
		if math.Abs(p.Z-10) > 0.5 {
			t.Errorf("offset point %d Z=%.2f, expected ~10", i, p.Z)
		}
	}
}

func TestPolylineEmpty(t *testing.T) {
	pl := Polyline{}
	if pl.Length() != 0 {
		t.Error("empty polyline should have zero length")
	}
	pt := pl.PointAt(0.5)
	if pt.X != 0 || pt.Z != 0 {
		t.Error("empty polyline PointAt should return zero")
	}
}
