package geo

import "math"

// ApproximateCircle returns a polygon approximating a circle with the given
// center, radius, and number of segments. Vertices are in CCW order.
func ApproximateCircle(center Point2D, radius float64, segments int) Polygon {
	if segments < 3 {
		segments = 3
	}
	pts := make([]Point2D, segments)
	for i := 0; i < segments; i++ {
		angle := 2 * math.Pi * float64(i) / float64(segments)
		pts[i] = Point2D{
			X: center.X + radius*math.Cos(angle),
			Z: center.Z + radius*math.Sin(angle),
		}
	}
	return Polygon{Vertices: pts}
}

// circleSegments is the default resolution for circle approximation.
const circleSegments = 64

// ClipToConvex clips the subject polygon to a convex clip polygon using
// the Sutherland-Hodgman algorithm. Returns the intersection polygon.
func ClipToConvex(subject, clipper Polygon) Polygon {
	if subject.IsEmpty() || clipper.IsEmpty() {
		return Polygon{}
	}
	output := make([]Point2D, len(subject.Vertices))
	copy(output, subject.Vertices)

	clipN := len(clipper.Vertices)
	for i := 0; i < clipN; i++ {
		if len(output) == 0 {
			return Polygon{}
		}
		edgeStart := clipper.Vertices[i]
		edgeEnd := clipper.Vertices[(i+1)%clipN]
		input := output
		output = make([]Point2D, 0, len(input))

		for j := 0; j < len(input); j++ {
			current := input[j]
			next := input[(j+1)%len(input)]
			curInside := isInsideEdge(current, edgeStart, edgeEnd)
			nextInside := isInsideEdge(next, edgeStart, edgeEnd)

			if curInside && nextInside {
				output = append(output, next)
			} else if curInside && !nextInside {
				if ix, ok := lineIntersection(current, next, edgeStart, edgeEnd); ok {
					output = append(output, ix)
				}
			} else if !curInside && nextInside {
				if ix, ok := lineIntersection(current, next, edgeStart, edgeEnd); ok {
					output = append(output, ix)
				}
				output = append(output, next)
			}
		}
	}
	if len(output) < 3 {
		return Polygon{}
	}
	return Polygon{Vertices: output}
}

// ClipToOutsideConvex clips the subject polygon to the exterior of a convex
// clip polygon. This keeps the parts of subject that are OUTSIDE clipper.
// For simple convex subjects, this produces one polygon (or empty).
// This is used for subtracting the inner circle in annulus clipping.
//
// For our use case (subject is already clipped to outer circle, inner circle
// is contained within), this works by clipping to each half-plane defined
// by the inner polygon edges with inverted inside test.
func ClipToOutsideConvex(subject, inner Polygon) Polygon {
	if subject.IsEmpty() || inner.IsEmpty() {
		return subject
	}
	// Ensure inner is CCW so "outside" is to the right of each edge.
	inner = inner.EnsureCCW()

	// For each edge of the inner polygon, classify subject vertices.
	// Keep vertices that are outside (right side of) at least one edge.
	// This is an approximation that works well for convex-on-convex
	// where the inner polygon is fully contained.

	// Better approach for annulus: for each vertex of subject, if it's
	// inside the inner circle, project it to the inner circle boundary.
	// Then remove degenerate edges.
	center := inner.Centroid()

	// Compute inner radius from centroid to first vertex (approximate).
	innerR := center.Distance(inner.Vertices[0])

	result := make([]Point2D, 0, len(subject.Vertices)*2)
	n := len(subject.Vertices)
	for i := 0; i < n; i++ {
		curr := subject.Vertices[i]
		next := subject.Vertices[(i+1)%n]
		currDist := center.Distance(curr)
		nextDist := center.Distance(next)
		currInside := currDist < innerR-0.01
		nextInside := nextDist < innerR-0.01

		if !currInside && !nextInside {
			// Both outside inner circle — keep the edge endpoint.
			result = append(result, next)
		} else if !currInside && nextInside {
			// Crossing into inner circle — add intersection point.
			if pt, ok := lineCircleIntersectionNearest(curr, next, center, innerR); ok {
				result = append(result, pt)
			}
			// Walk along inner circle boundary to next exit point.
		} else if currInside && !nextInside {
			// Exiting inner circle — add intersection point, then next.
			if pt, ok := lineCircleIntersectionNearest(next, curr, center, innerR); ok {
				result = append(result, pt)
			}
			result = append(result, next)
		}
		// Both inside — skip (both will be projected).
	}

	if len(result) < 3 {
		return Polygon{}
	}
	return Polygon{Vertices: result}
}

// ClipToAnnulus clips a polygon to the annular region between innerR and outerR
// centered at center. Returns the clipped polygon (may have curved sections
// approximated by line segments).
func ClipToAnnulus(subject Polygon, center Point2D, innerR, outerR float64) Polygon {
	if subject.IsEmpty() {
		return Polygon{}
	}
	// Step 1: clip to outer circle (keep inside).
	outerCircle := ApproximateCircle(center, outerR, circleSegments)
	result := ClipToConvex(subject, outerCircle)
	if result.IsEmpty() {
		return Polygon{}
	}

	// Step 2: if innerR > 0, remove inner circle (keep outside).
	if innerR > 0.01 {
		result = clipOutsideCircle(result, center, innerR)
	}

	return result
}

// clipOutsideCircle removes the interior of a circle from a polygon.
// It walks the polygon boundary, replacing segments inside the circle
// with arc segments along the circle.
func clipOutsideCircle(subject Polygon, center Point2D, radius float64) Polygon {
	if subject.IsEmpty() {
		return Polygon{}
	}
	n := len(subject.Vertices)
	result := make([]Point2D, 0, n*2)

	for i := 0; i < n; i++ {
		curr := subject.Vertices[i]
		next := subject.Vertices[(i+1)%n]
		currDist := center.Distance(curr)
		nextDist := center.Distance(next)
		currOutside := currDist >= radius-0.01
		nextOutside := nextDist >= radius-0.01

		if currOutside && nextOutside {
			// Check if the segment passes through the circle.
			if segmentIntersectsCircle(curr, next, center, radius) {
				// Find entry and exit points.
				pts := lineCircleIntersections(curr, next, center, radius)
				if len(pts) == 2 {
					result = append(result, pts[0])
					// Add arc along circle from pts[0] to pts[1].
					arcPts := arcBetween(center, radius, pts[0], pts[1])
					result = append(result, arcPts...)
					result = append(result, pts[1])
				}
			}
			result = append(result, next)
		} else if currOutside && !nextOutside {
			// Entering circle.
			if pt, ok := lineCircleIntersectionFirst(curr, next, center, radius); ok {
				result = append(result, pt)
			}
		} else if !currOutside && nextOutside {
			// Exiting circle — add arc from entry to exit, then next.
			if pt, ok := lineCircleIntersectionFirst(next, curr, center, radius); ok {
				// Add arc from previous entry to this exit.
				if len(result) > 0 {
					lastPt := result[len(result)-1]
					arcPts := arcBetween(center, radius, lastPt, pt)
					result = append(result, arcPts...)
				}
				result = append(result, pt)
			}
			result = append(result, next)
		}
		// Both inside: skip, will be replaced by arc.
	}

	if len(result) < 3 {
		return Polygon{}
	}
	return Polygon{Vertices: result}
}

// isInsideEdge returns true if the point is on the inside (left) of the
// directed edge from edgeStart to edgeEnd.
func isInsideEdge(p, edgeStart, edgeEnd Point2D) bool {
	return (edgeEnd.X-edgeStart.X)*(p.Z-edgeStart.Z)-
		(edgeEnd.Z-edgeStart.Z)*(p.X-edgeStart.X) >= 0
}

// lineIntersection returns the intersection point of lines (p1→p2) and (p3→p4).
func lineIntersection(p1, p2, p3, p4 Point2D) (Point2D, bool) {
	d := (p1.X-p2.X)*(p3.Z-p4.Z) - (p1.Z-p2.Z)*(p3.X-p4.X)
	if math.Abs(d) < 1e-12 {
		return Point2D{}, false
	}
	t := ((p1.X-p3.X)*(p3.Z-p4.Z) - (p1.Z-p3.Z)*(p3.X-p4.X)) / d
	return Point2D{
		X: p1.X + t*(p2.X-p1.X),
		Z: p1.Z + t*(p2.Z-p1.Z),
	}, true
}

// lineCircleIntersectionNearest returns the point on the line segment from a to b
// that is closest to the circle boundary and on the segment.
func lineCircleIntersectionNearest(a, b, center Point2D, radius float64) (Point2D, bool) {
	pts := lineCircleIntersections(a, b, center, radius)
	if len(pts) == 0 {
		return Point2D{}, false
	}
	// Return the point nearest to a.
	best := pts[0]
	bestDist := a.Distance(pts[0])
	for _, p := range pts[1:] {
		d := a.Distance(p)
		if d < bestDist {
			best = p
			bestDist = d
		}
	}
	return best, true
}

// lineCircleIntersectionFirst returns the first intersection of segment a→b with
// the circle, measured from a.
func lineCircleIntersectionFirst(a, b, center Point2D, radius float64) (Point2D, bool) {
	return lineCircleIntersectionNearest(a, b, center, radius)
}

// lineCircleIntersections returns all intersection points of the line segment
// from a to b with the circle at center with given radius.
func lineCircleIntersections(a, b, center Point2D, radius float64) []Point2D {
	d := b.Sub(a)
	f := a.Sub(center)

	aa := d.Dot(d)
	bb := 2 * f.Dot(d)
	cc := f.Dot(f) - radius*radius

	disc := bb*bb - 4*aa*cc
	if disc < 0 {
		return nil
	}

	var pts []Point2D
	sqrtDisc := math.Sqrt(disc)
	for _, sign := range []float64{-1, 1} {
		t := (-bb + sign*sqrtDisc) / (2 * aa)
		if t >= -0.001 && t <= 1.001 {
			t = math.Max(0, math.Min(1, t))
			pts = append(pts, a.Lerp(b, t))
		}
	}
	return pts
}

// segmentIntersectsCircle returns true if the line segment from a to b
// passes through the interior of the circle.
func segmentIntersectsCircle(a, b, center Point2D, radius float64) bool {
	// Find closest point on segment to center.
	d := b.Sub(a)
	lenSq := d.Dot(d)
	if lenSq < 1e-12 {
		return a.Distance(center) < radius
	}
	t := math.Max(0, math.Min(1, a.Sub(center).Scale(-1).Dot(d)/lenSq))
	closest := a.Lerp(b, t)
	return closest.Distance(center) < radius-0.01
}

// arcBetween returns intermediate points on an arc from p1 to p2 on a circle.
// The arc goes counterclockwise (shortest path). Points p1 and p2 should be
// on the circle.
func arcBetween(center Point2D, radius float64, p1, p2 Point2D) []Point2D {
	a1 := math.Atan2(p1.Z-center.Z, p1.X-center.X)
	a2 := math.Atan2(p2.Z-center.Z, p2.X-center.X)

	// Ensure we go counterclockwise from a1 to a2.
	diff := a2 - a1
	if diff < 0 {
		diff += 2 * math.Pi
	}
	if diff > 2*math.Pi {
		diff -= 2 * math.Pi
	}

	// Number of intermediate points based on arc length.
	arcLen := radius * diff
	numPts := int(math.Ceil(arcLen / 20.0)) // ~20m spacing for arc points
	if numPts < 1 {
		return nil
	}

	pts := make([]Point2D, 0, numPts)
	for i := 1; i < numPts; i++ {
		t := float64(i) / float64(numPts)
		angle := a1 + diff*t
		pts = append(pts, Point2D{
			X: center.X + radius*math.Cos(angle),
			Z: center.Z + radius*math.Sin(angle),
		})
	}
	return pts
}
