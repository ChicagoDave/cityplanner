package routing

import (
	"fmt"
	"math"

	"github.com/ChicagoDave/cityplanner/pkg/layout"
	"github.com/ChicagoDave/cityplanner/pkg/spec"
	"github.com/ChicagoDave/cityplanner/pkg/validation"
)

// NetworkType identifies an infrastructure network.
type NetworkType string

const (
	NetworkSewage     NetworkType = "sewage"
	NetworkWater      NetworkType = "water"
	NetworkElectrical NetworkType = "electrical"
	NetworkTelecom    NetworkType = "telecom"
	NetworkVehicle    NetworkType = "vehicle"
	NetworkPedway     NetworkType = "pedway"
	NetworkBikeTunnel NetworkType = "bike_tunnel"
)

// Segment represents a routed infrastructure segment.
type Segment struct {
	ID          string      `json:"id"`
	Network     NetworkType `json:"network"`
	Layer       int         `json:"layer"` // 1=bottom, 2=middle, 3=top
	Start       [3]float64  `json:"start"`
	End         [3]float64  `json:"end"`
	WidthM      float64     `json:"width_m"`
	Capacity    float64     `json:"capacity"` // network-specific units
	IsTrunk     bool        `json:"is_trunk"`
	ConnectedTo []string    `json:"connected_to,omitempty"`
}

// Layer Y offsets (within 8m excavation depth).
const (
	yLayer1 = -7.0 // bottom: sewage + water
	yLayer2 = -4.5 // middle: electrical + telecom
	yLayer3 = -2.0 // top underground: vehicle lanes (3m clearance)
)

// backbone holds precomputed trunk geometry shared by all networks.
type backbone struct {
	numRadials   int
	radialAngles []float64
	perimeterR   float64
	ringRadii    []float64 // inner ring boundaries [300, 600]
	junctions    []junction
}

// junction is where a radial crosses a ring boundary.
type junction struct {
	radialIdx int
	ringR     float64
	x, z      float64
}

// networkDef defines routing parameters for a single network.
type networkDef struct {
	net       NetworkType
	layer     int
	yOffset   float64
	trunkW    float64
	branchW   float64
	capFactor float64 // per-capita capacity multiplier
}

// RouteInfrastructure generates all underground infrastructure routes
// using hierarchical trunk-and-branch topology (ADR-007).
func RouteInfrastructure(s *spec.CitySpec, pods []layout.Pod, _ []layout.Building) ([]Segment, *validation.Report) {
	report := validation.NewReport()

	if len(pods) == 0 {
		report.AddWarning(validation.Result{
			Level:   validation.LevelSpatial,
			Message: "no pods for infrastructure routing",
		})
		return nil, report
	}

	// Total population for capacity sizing.
	totalPop := 0
	for _, p := range pods {
		totalPop += p.TargetPopulation
	}

	// Compute backbone: radials + ring connectors.
	bb := computeBackbone(s, len(pods))

	// Vehicle widths from spec (with defaults).
	arterialW := s.Vehicles.ArterialWidthM
	if arterialW == 0 {
		arterialW = 6
	}
	branchW := s.Vehicles.ServiceBranchWidthM
	if branchW == 0 {
		branchW = 4
	}

	// Network definitions routed in dependency order per ADR-007.
	networks := []networkDef{
		{NetworkSewage, 1, yLayer1, 2.5, 1.5, 95},       // gpd per capita
		{NetworkWater, 1, yLayer1, 2.5, 1.5, 100},        // gpd per capita
		{NetworkElectrical, 2, yLayer2, 2.0, 1.0, 2.5},   // kW per capita
		{NetworkTelecom, 2, yLayer2, 1.5, 0.8, 0},         // special: node-based
		{NetworkVehicle, 3, yLayer3, arterialW, branchW, 0},    // special: fleet-based
		{NetworkPedway, 3, yLayer3, 3.0, 2.0, 0},               // underground pedestrian tunnels
		{NetworkBikeTunnel, 3, yLayer3, 2.5, 1.5, 0},           // underground bicycle tunnels
	}

	var allSegments []Segment
	idx := 0

	for _, nd := range networks {
		// Lateral offset to separate networks sharing a layer.
		lateralOffset := 0.0
		if nd.net == NetworkWater {
			lateralOffset = 3.0 // offset from sewage in layer 1
		}
		if nd.net == NetworkTelecom {
			lateralOffset = 2.5 // offset from electrical in layer 2
		}
		if nd.net == NetworkPedway {
			lateralOffset = 5.0 // offset from vehicle centerline in layer 3
		}
		if nd.net == NetworkBikeTunnel {
			lateralOffset = 8.0 // offset further from vehicle in layer 3
		}

		segs := routeNetwork(nd, bb, pods, totalPop, lateralOffset, &idx)
		allSegments = append(allSegments, segs...)
	}

	// Build connectivity graph and populate each segment.
	connMap := BuildConnectivity(allSegments)
	for i := range allSegments {
		if ids, ok := connMap[allSegments[i].ID]; ok {
			allSegments[i].ConnectedTo = ids
		}
	}

	// Validation summary.
	netCounts := map[NetworkType]int{}
	for _, seg := range allSegments {
		netCounts[seg.Network]++
	}
	for _, nd := range networks {
		if netCounts[nd.net] == 0 {
			report.AddWarning(validation.Result{
				Level:   validation.LevelSpatial,
				Message: fmt.Sprintf("no segments generated for %s network", nd.net),
			})
		}
	}

	report.AddInfo(validation.Result{
		Level: validation.LevelSpatial,
		Message: fmt.Sprintf("routed %d infrastructure segments: sewage=%d water=%d electrical=%d telecom=%d vehicle=%d pedway=%d bike_tunnel=%d",
			len(allSegments), netCounts[NetworkSewage], netCounts[NetworkWater],
			netCounts[NetworkElectrical], netCounts[NetworkTelecom], netCounts[NetworkVehicle],
			netCounts[NetworkPedway], netCounts[NetworkBikeTunnel]),
	})

	return allSegments, report
}

// computeBackbone builds the shared trunk geometry from spec.
func computeBackbone(s *spec.CitySpec, podCount int) backbone {
	numRadials := podCount
	if numRadials < 4 {
		numRadials = 4
	}

	perimeterR := s.CityZones.OuterRadius()
	if perimeterR == 0 {
		perimeterR = 900
	}

	// Ring boundaries are the RadiusTo of each ring except the outermost.
	var ringRadii []float64
	for i := 0; i < len(s.CityZones.Rings)-1; i++ {
		if s.CityZones.Rings[i].RadiusTo > 0 {
			ringRadii = append(ringRadii, s.CityZones.Rings[i].RadiusTo)
		}
	}

	angles := make([]float64, numRadials)
	for i := range angles {
		angles[i] = 2 * math.Pi * float64(i) / float64(numRadials)
	}

	// Compute junction points (radial × ring intersections).
	var junctions []junction
	for ri, a := range angles {
		for _, r := range ringRadii {
			junctions = append(junctions, junction{
				radialIdx: ri,
				ringR:     r,
				x:         r * math.Cos(a),
				z:         r * math.Sin(a),
			})
		}
	}

	return backbone{
		numRadials:   numRadials,
		radialAngles: angles,
		perimeterR:   perimeterR,
		ringRadii:    ringRadii,
		junctions:    junctions,
	}
}

// routeNetwork generates trunk + ring connector + branch segments for one network.
func routeNetwork(nd networkDef, bb backbone, pods []layout.Pod, totalPop int, lateralOffset float64, idx *int) []Segment {
	var segs []Segment

	// 1. Radial trunk segments: perimeter → ring boundaries → center.
	// Each radial is split at ring boundaries into sub-segments.
	for _, angle := range bb.radialAngles {
		// Build radial breakpoints from perimeter inward.
		breakpoints := []float64{bb.perimeterR}
		for i := len(bb.ringRadii) - 1; i >= 0; i-- {
			breakpoints = append(breakpoints, bb.ringRadii[i])
		}
		breakpoints = append(breakpoints, 10) // stop 10m from center (avoid singularity)

		cos, sin := math.Cos(angle), math.Sin(angle)
		perpX, perpZ := -sin, cos // perpendicular for lateral offset

		for i := 0; i < len(breakpoints)-1; i++ {
			outerR := breakpoints[i]
			innerR := breakpoints[i+1]

			// Downstream population: fraction of city inside this radius.
			downPop := downstreamPop(innerR, bb.perimeterR, totalPop, bb.numRadials)
			capacity := capacityForNetwork(nd, downPop, segLength(outerR, innerR))

			segs = append(segs, Segment{
				ID:      fmt.Sprintf("%s_trunk_%03d", nd.net, *idx),
				Network: nd.net,
				Layer:   nd.layer,
				Start: [3]float64{
					outerR*cos + lateralOffset*perpX,
					nd.yOffset,
					outerR*sin + lateralOffset*perpZ,
				},
				End: [3]float64{
					innerR*cos + lateralOffset*perpX,
					nd.yOffset,
					innerR*sin + lateralOffset*perpZ,
				},
				WidthM:   nd.trunkW,
				Capacity: capacity,
				IsTrunk:  true,
			})
			*idx++
		}
	}

	// 2. Ring connector segments: chords between adjacent radials at each ring radius.
	for _, ringR := range bb.ringRadii {
		for i := 0; i < bb.numRadials; i++ {
			a1 := bb.radialAngles[i]
			a2 := bb.radialAngles[(i+1)%bb.numRadials]

			capacity := capacityForNetwork(nd, totalPop/bb.numRadials, ringR*math.Abs(a2-a1))

			segs = append(segs, Segment{
				ID:      fmt.Sprintf("%s_ring_%03d", nd.net, *idx),
				Network: nd.net,
				Layer:   nd.layer,
				Start:   [3]float64{ringR * math.Cos(a1), nd.yOffset, ringR * math.Sin(a1)},
				End:     [3]float64{ringR * math.Cos(a2), nd.yOffset, ringR * math.Sin(a2)},
				WidthM:  nd.trunkW,
				Capacity: capacity,
				IsTrunk: true,
			})
			*idx++
		}
	}

	// 3. Branch segments: nearest junction → pod center.
	for _, pod := range pods {
		cx, cz := pod.Center[0], pod.Center[1]
		jx, jz := nearestJunction(cx, cz, bb)

		length := math.Hypot(cx-jx, cz-jz)
		if length < 1 {
			continue // pod center is already at a junction
		}

		capacity := capacityForNetwork(nd, pod.TargetPopulation, length)

		segs = append(segs, Segment{
			ID:      fmt.Sprintf("%s_branch_%03d", nd.net, *idx),
			Network: nd.net,
			Layer:   nd.layer,
			Start:   [3]float64{jx, nd.yOffset, jz},
			End:     [3]float64{cx, nd.yOffset, cz},
			WidthM:  nd.branchW,
			Capacity: capacity,
			IsTrunk: false,
		})
		*idx++
	}

	return segs
}

// nearestJunction finds the junction point closest to (x, z).
func nearestJunction(x, z float64, bb backbone) (float64, float64) {
	bestDist := math.MaxFloat64
	bestX, bestZ := 0.0, 0.0

	for _, j := range bb.junctions {
		d := math.Hypot(x-j.x, z-j.z)
		if d < bestDist {
			bestDist = d
			bestX = j.x
			bestZ = j.z
		}
	}
	// Also check radial points at the perimeter and center.
	for _, a := range bb.radialAngles {
		px, pz := bb.perimeterR*math.Cos(a), bb.perimeterR*math.Sin(a)
		if d := math.Hypot(x-px, z-pz); d < bestDist {
			bestDist = d
			bestX = px
			bestZ = pz
		}
	}
	return bestX, bestZ
}

// downstreamPop estimates population served downstream of a radial point.
// Uses area-proportional estimate: pop ∝ (innerR²/perimeterR²) shared across radials.
func downstreamPop(innerR, perimeterR float64, totalPop, numRadials int) int {
	if perimeterR < 1 {
		if numRadials < 1 {
			numRadials = 1
		}
		return totalPop / numRadials
	}
	fraction := 1 - (innerR*innerR)/(perimeterR*perimeterR)
	return int(math.Ceil(fraction * float64(totalPop) / float64(numRadials)))
}

// capacityForNetwork computes the capacity value for a segment.
func capacityForNetwork(nd networkDef, pop int, segLength float64) float64 {
	switch nd.net {
	case NetworkSewage, NetworkWater, NetworkElectrical:
		return float64(pop) * nd.capFactor
	case NetworkTelecom:
		if segLength < 75 {
			return 1
		}
		return math.Ceil(segLength / 75)
	case NetworkVehicle:
		return math.Max(1, float64(pop)/250) // ~1 vehicle per 250 people
	case NetworkPedway:
		return math.Max(1, float64(pop)/100) // ~1 pedestrian per 100 people peak
	case NetworkBikeTunnel:
		return math.Max(1, float64(pop)/200) // ~1 cyclist per 200 people peak
	default:
		return float64(pop)
	}
}

func segLength(outerR, innerR float64) float64 {
	return math.Abs(outerR - innerR)
}
