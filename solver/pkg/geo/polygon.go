package geo

import "math"

// Polygon is a closed polygon defined by its vertices in order.
type Polygon struct {
	Vertices []Point2D
}

// NewPolygon creates a polygon from a list of vertices.
func NewPolygon(pts ...Point2D) Polygon {
	return Polygon{Vertices: pts}
}

// Len returns the number of vertices.
func (p Polygon) Len() int {
	return len(p.Vertices)
}

// IsEmpty returns true if the polygon has fewer than 3 vertices.
func (p Polygon) IsEmpty() bool {
	return len(p.Vertices) < 3
}

// Edge returns the i-th edge as (start, end). Wraps around.
func (p Polygon) Edge(i int) (Point2D, Point2D) {
	n := len(p.Vertices)
	return p.Vertices[i%n], p.Vertices[(i+1)%n]
}

// SignedArea returns the signed area using the shoelace formula.
// Positive for counterclockwise winding, negative for clockwise.
func (p Polygon) SignedArea() float64 {
	n := len(p.Vertices)
	if n < 3 {
		return 0
	}
	area := 0.0
	for i := 0; i < n; i++ {
		j := (i + 1) % n
		area += p.Vertices[i].X * p.Vertices[j].Z
		area -= p.Vertices[j].X * p.Vertices[i].Z
	}
	return area / 2
}

// Area returns the unsigned area of the polygon.
func (p Polygon) Area() float64 {
	return math.Abs(p.SignedArea())
}

// IsCounterClockwise returns true if vertices are in CCW order.
func (p Polygon) IsCounterClockwise() bool {
	return p.SignedArea() > 0
}

// EnsureCCW returns the polygon with vertices in counterclockwise order.
func (p Polygon) EnsureCCW() Polygon {
	if p.SignedArea() < 0 {
		return p.Reverse()
	}
	return p
}

// Reverse returns the polygon with reversed vertex order.
func (p Polygon) Reverse() Polygon {
	n := len(p.Vertices)
	rev := make([]Point2D, n)
	for i, v := range p.Vertices {
		rev[n-1-i] = v
	}
	return Polygon{Vertices: rev}
}

// Centroid returns the centroid of the polygon.
func (p Polygon) Centroid() Point2D {
	n := len(p.Vertices)
	if n == 0 {
		return Point2D{}
	}
	if n < 3 {
		// Average for degenerate case.
		sum := Point2D{}
		for _, v := range p.Vertices {
			sum = sum.Add(v)
		}
		return sum.Scale(1.0 / float64(n))
	}
	cx, cz := 0.0, 0.0
	a := p.SignedArea()
	if math.Abs(a) < 1e-12 {
		// Degenerate: return average.
		sum := Point2D{}
		for _, v := range p.Vertices {
			sum = sum.Add(v)
		}
		return sum.Scale(1.0 / float64(n))
	}
	for i := 0; i < n; i++ {
		j := (i + 1) % n
		cross := p.Vertices[i].X*p.Vertices[j].Z - p.Vertices[j].X*p.Vertices[i].Z
		cx += (p.Vertices[i].X + p.Vertices[j].X) * cross
		cz += (p.Vertices[i].Z + p.Vertices[j].Z) * cross
	}
	f := 1.0 / (6.0 * a)
	return Point2D{cx * f, cz * f}
}

// BoundingBox returns the axis-aligned bounding box as (min, max).
func (p Polygon) BoundingBox() (Point2D, Point2D) {
	if len(p.Vertices) == 0 {
		return Point2D{}, Point2D{}
	}
	minP := p.Vertices[0]
	maxP := p.Vertices[0]
	for _, v := range p.Vertices[1:] {
		if v.X < minP.X {
			minP.X = v.X
		}
		if v.Z < minP.Z {
			minP.Z = v.Z
		}
		if v.X > maxP.X {
			maxP.X = v.X
		}
		if v.Z > maxP.Z {
			maxP.Z = v.Z
		}
	}
	return minP, maxP
}

// Contains returns true if the point is inside the polygon using ray casting.
func (p Polygon) Contains(pt Point2D) bool {
	n := len(p.Vertices)
	if n < 3 {
		return false
	}
	inside := false
	j := n - 1
	for i := 0; i < n; i++ {
		vi := p.Vertices[i]
		vj := p.Vertices[j]
		if (vi.Z > pt.Z) != (vj.Z > pt.Z) &&
			pt.X < (vj.X-vi.X)*(pt.Z-vi.Z)/(vj.Z-vi.Z)+vi.X {
			inside = !inside
		}
		j = i
	}
	return inside
}

// Perimeter returns the total perimeter length.
func (p Polygon) Perimeter() float64 {
	n := len(p.Vertices)
	if n < 2 {
		return 0
	}
	total := 0.0
	for i := 0; i < n; i++ {
		j := (i + 1) % n
		total += p.Vertices[i].Distance(p.Vertices[j])
	}
	return total
}

// MaxDistanceTo returns the maximum distance from any vertex to the given point.
func (p Polygon) MaxDistanceTo(pt Point2D) float64 {
	maxDist := 0.0
	for _, v := range p.Vertices {
		d := v.Distance(pt)
		if d > maxDist {
			maxDist = d
		}
	}
	return maxDist
}

// FarthestVertexFrom returns the vertex farthest from the given point.
func (p Polygon) FarthestVertexFrom(pt Point2D) Point2D {
	maxDist := 0.0
	var farthest Point2D
	for _, v := range p.Vertices {
		d := v.Distance(pt)
		if d > maxDist {
			maxDist = d
			farthest = v
		}
	}
	return farthest
}
