package geo

import "math"

// Point2D represents a point in the XZ plane (Y is up in the 3D scene graph).
type Point2D struct {
	X float64 `json:"x"`
	Z float64 `json:"z"`
}

// Origin is the zero point.
var Origin = Point2D{0, 0}

// Pt is a shorthand constructor for Point2D.
func Pt(x, z float64) Point2D {
	return Point2D{X: x, Z: z}
}

// Add returns p + q.
func (p Point2D) Add(q Point2D) Point2D {
	return Point2D{p.X + q.X, p.Z + q.Z}
}

// Sub returns p - q.
func (p Point2D) Sub(q Point2D) Point2D {
	return Point2D{p.X - q.X, p.Z - q.Z}
}

// Scale returns p * s.
func (p Point2D) Scale(s float64) Point2D {
	return Point2D{p.X * s, p.Z * s}
}

// Length returns the Euclidean length of the vector.
func (p Point2D) Length() float64 {
	return math.Hypot(p.X, p.Z)
}

// Normalize returns the unit vector in the same direction.
// Returns zero vector if length is zero.
func (p Point2D) Normalize() Point2D {
	l := p.Length()
	if l < 1e-12 {
		return Point2D{}
	}
	return Point2D{p.X / l, p.Z / l}
}

// Dot returns the dot product of p and q.
func (p Point2D) Dot(q Point2D) float64 {
	return p.X*q.X + p.Z*q.Z
}

// Cross returns the 2D cross product (z-component of 3D cross).
func (p Point2D) Cross(q Point2D) float64 {
	return p.X*q.Z - p.Z*q.X
}

// Distance returns the Euclidean distance from p to q.
func (p Point2D) Distance(q Point2D) float64 {
	return p.Sub(q).Length()
}

// Angle returns the angle of the vector from the positive X axis in radians.
func (p Point2D) Angle() float64 {
	return math.Atan2(p.Z, p.X)
}

// AngleTo returns the angle from p to q relative to the positive X axis.
func (p Point2D) AngleTo(q Point2D) float64 {
	return q.Sub(p).Angle()
}

// Rotate returns p rotated by angle radians around the origin.
func (p Point2D) Rotate(angle float64) Point2D {
	c, s := math.Cos(angle), math.Sin(angle)
	return Point2D{
		X: p.X*c - p.Z*s,
		Z: p.X*s + p.Z*c,
	}
}

// RotateAround returns p rotated by angle radians around center.
func (p Point2D) RotateAround(center Point2D, angle float64) Point2D {
	return p.Sub(center).Rotate(angle).Add(center)
}

// Lerp returns the linear interpolation between p and q at t in [0,1].
func (p Point2D) Lerp(q Point2D, t float64) Point2D {
	return Point2D{
		X: p.X + (q.X-p.X)*t,
		Z: p.Z + (q.Z-p.Z)*t,
	}
}

// Perp returns a vector perpendicular to p (rotated 90 degrees counterclockwise).
func (p Point2D) Perp() Point2D {
	return Point2D{-p.Z, p.X}
}

// MidPoint returns the midpoint between p and q.
func MidPoint(p, q Point2D) Point2D {
	return p.Lerp(q, 0.5)
}
