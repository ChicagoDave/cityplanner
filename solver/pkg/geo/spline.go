package geo

import "math"

// Polyline is an ordered sequence of points forming a path.
type Polyline struct {
	Points []Point2D
}

// NewPolyline creates a polyline from a list of points.
func NewPolyline(pts ...Point2D) Polyline {
	return Polyline{Points: pts}
}

// Length returns the total arc length of the polyline.
func (pl Polyline) Length() float64 {
	total := 0.0
	for i := 1; i < len(pl.Points); i++ {
		total += pl.Points[i-1].Distance(pl.Points[i])
	}
	return total
}

// PointAt returns the point at fraction t in [0,1] along the polyline length.
func (pl Polyline) PointAt(t float64) Point2D {
	if len(pl.Points) == 0 {
		return Point2D{}
	}
	if len(pl.Points) == 1 || t <= 0 {
		return pl.Points[0]
	}
	if t >= 1 {
		return pl.Points[len(pl.Points)-1]
	}

	totalLen := pl.Length()
	targetLen := t * totalLen
	walked := 0.0

	for i := 1; i < len(pl.Points); i++ {
		segLen := pl.Points[i-1].Distance(pl.Points[i])
		if walked+segLen >= targetLen {
			frac := (targetLen - walked) / segLen
			return pl.Points[i-1].Lerp(pl.Points[i], frac)
		}
		walked += segLen
	}
	return pl.Points[len(pl.Points)-1]
}

// NearestPoint returns the closest point on the polyline to p, and the distance.
func (pl Polyline) NearestPoint(p Point2D) (Point2D, float64) {
	if len(pl.Points) == 0 {
		return Point2D{}, math.MaxFloat64
	}
	if len(pl.Points) == 1 {
		d := p.Distance(pl.Points[0])
		return pl.Points[0], d
	}

	bestPt := pl.Points[0]
	bestDist := p.Distance(pl.Points[0])

	for i := 1; i < len(pl.Points); i++ {
		pt, dist := nearestPointOnSegment(p, pl.Points[i-1], pl.Points[i])
		if dist < bestDist {
			bestDist = dist
			bestPt = pt
		}
	}
	return bestPt, bestDist
}

// nearestPointOnSegment returns the closest point on segment ab to p.
func nearestPointOnSegment(p, a, b Point2D) (Point2D, float64) {
	ab := b.Sub(a)
	abLen2 := ab.Dot(ab)
	if abLen2 < 1e-12 {
		d := p.Distance(a)
		return a, d
	}
	t := p.Sub(a).Dot(ab) / abLen2
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	closest := a.Add(ab.Scale(t))
	return closest, p.Distance(closest)
}

// Offset returns a polyline offset by distance to the left (positive = left
// when walking along the polyline direction).
func (pl Polyline) Offset(distance float64) Polyline {
	n := len(pl.Points)
	if n < 2 {
		return pl
	}

	result := make([]Point2D, n)
	for i := 0; i < n; i++ {
		var normal Point2D
		if i == 0 {
			dir := pl.Points[1].Sub(pl.Points[0]).Normalize()
			normal = dir.Perp()
		} else if i == n-1 {
			dir := pl.Points[n-1].Sub(pl.Points[n-2]).Normalize()
			normal = dir.Perp()
		} else {
			dir1 := pl.Points[i].Sub(pl.Points[i-1]).Normalize()
			dir2 := pl.Points[i+1].Sub(pl.Points[i]).Normalize()
			avgDir := dir1.Add(dir2).Normalize()
			normal = avgDir.Perp()
		}
		result[i] = pl.Points[i].Add(normal.Scale(distance))
	}
	return Polyline{Points: result}
}

// CatmullRomSpline evaluates a Catmull-Rom spline through the given control
// points. It generates samplesPerSegment intermediate points per segment.
// Tension controls tightness (0.5 = centripetal, 0.0 = uniform).
// Returns a polyline of sampled points.
func CatmullRomSpline(controlPoints []Point2D, samplesPerSegment int, tension float64) Polyline {
	n := len(controlPoints)
	if n == 0 {
		return Polyline{}
	}
	if n == 1 {
		return NewPolyline(controlPoints[0])
	}
	if n == 2 {
		// Degenerate: linear interpolation.
		pts := make([]Point2D, samplesPerSegment+1)
		for i := 0; i <= samplesPerSegment; i++ {
			t := float64(i) / float64(samplesPerSegment)
			pts[i] = controlPoints[0].Lerp(controlPoints[1], t)
		}
		return Polyline{Points: pts}
	}

	if samplesPerSegment < 1 {
		samplesPerSegment = 1
	}

	// Build extended control point array with phantom endpoints.
	extended := make([]Point2D, n+2)
	// Phantom start: reflect first segment.
	extended[0] = controlPoints[0].Add(controlPoints[0].Sub(controlPoints[1]))
	copy(extended[1:], controlPoints)
	// Phantom end: reflect last segment.
	extended[n+1] = controlPoints[n-1].Add(controlPoints[n-1].Sub(controlPoints[n-2]))

	var pts []Point2D

	// For each segment between extended[i] and extended[i+1] where i goes from 1 to n-1
	// (i.e., between original control points).
	for i := 1; i < n; i++ {
		p0 := extended[i-1]
		p1 := extended[i]
		p2 := extended[i+1]
		p3 := extended[i+2]

		for j := 0; j < samplesPerSegment; j++ {
			t := float64(j) / float64(samplesPerSegment)
			pt := catmullRomPoint(p0, p1, p2, p3, t, tension)
			pts = append(pts, pt)
		}
	}
	// Add the last control point.
	pts = append(pts, controlPoints[n-1])

	return Polyline{Points: pts}
}

// CatmullRomSplineClosed evaluates a closed Catmull-Rom spline loop through
// the given control points. Returns a polyline where the last point connects
// back near the first.
func CatmullRomSplineClosed(controlPoints []Point2D, samplesPerSegment int, tension float64) Polyline {
	n := len(controlPoints)
	if n < 3 {
		return CatmullRomSpline(controlPoints, samplesPerSegment, tension)
	}

	if samplesPerSegment < 1 {
		samplesPerSegment = 1
	}

	var pts []Point2D
	for i := 0; i < n; i++ {
		p0 := controlPoints[(i-1+n)%n]
		p1 := controlPoints[i]
		p2 := controlPoints[(i+1)%n]
		p3 := controlPoints[(i+2)%n]

		for j := 0; j < samplesPerSegment; j++ {
			t := float64(j) / float64(samplesPerSegment)
			pt := catmullRomPoint(p0, p1, p2, p3, t, tension)
			pts = append(pts, pt)
		}
	}
	// Close the loop.
	pts = append(pts, pts[0])

	return Polyline{Points: pts}
}

// catmullRomPoint evaluates a single point on a Catmull-Rom spline segment.
func catmullRomPoint(p0, p1, p2, p3 Point2D, t, tension float64) Point2D {
	t2 := t * t
	t3 := t2 * t

	// Standard Catmull-Rom matrix with tension parameter.
	// tension=0.5 gives standard centripetal CR spline.
	s := tension

	x := 0.5 * ((-s*p0.X+(2-s)*p1.X+(s-2)*p2.X+s*p3.X)*t3 +
		(2*s*p0.X+(s-3)*p1.X+(3-2*s)*p2.X-s*p3.X)*t2 +
		(-s*p0.X+s*p2.X)*t +
		2*p1.X) / 1.0

	z := 0.5 * ((-s*p0.Z+(2-s)*p1.Z+(s-2)*p2.Z+s*p3.Z)*t3 +
		(2*s*p0.Z+(s-3)*p1.Z+(3-2*s)*p2.Z-s*p3.Z)*t2 +
		(-s*p0.Z+s*p2.Z)*t +
		2*p1.Z) / 1.0

	return Point2D{X: x, Z: z}
}
