package scene

import (
	"fmt"
	"math"
	"time"

	"github.com/ChicagoDave/cityplanner/pkg/layout"
	"github.com/ChicagoDave/cityplanner/pkg/routing"
	"github.com/ChicagoDave/cityplanner/pkg/spec"
)

const floorHeight = 3.0 // meters per story

// Assemble converts all solver outputs into a scene graph matching the
// scene-graph.schema.json specification.
func Assemble(
	s *spec.CitySpec,
	pods []layout.Pod,
	buildings []layout.Building,
	paths []layout.PathSegment,
	segments []routing.Segment,
	greenZones []layout.Zone,
) *Graph {
	g := NewGraph()

	assembleBuildings(buildings, g)
	assemblePaths(paths, g)
	assembleRouting(segments, g)
	assembleParks(greenZones, g)

	g.Metadata = Metadata{
		SpecVersion: s.SpecVersion,
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		CityBounds:  computeBounds(g.Entities),
	}

	return g
}

func assembleBuildings(buildings []layout.Building, g *Graph) {
	for _, b := range buildings {
		mat := "concrete"
		meta := map[string]any{"stories": b.Stories}

		switch b.Type {
		case "commercial":
			mat = "glass"
			meta["commercial_sqm"] = b.CommercialSqM
		case "civic":
			mat = "brick"
			meta["service_type"] = b.ServiceType
		default: // residential
			meta["dwelling_units"] = b.DwellingUnits
		}

		addEntity(g, Entity{
			ID:   b.ID,
			Type: EntityBuilding,
			Position: Vec3{
				X: b.Position[0],
				Y: 0,
				Z: b.Position[2],
			},
			Dimensions: Vec3{
				X: b.Footprint[0],
				Y: float64(b.Stories) * floorHeight,
				Z: b.Footprint[1],
			},
			Rotation: identityQuat(),
			Material: mat,
			Pod:      b.PodID,
			Layer:    LayerSurface,
			Metadata: meta,
		})
	}
}

func assemblePaths(paths []layout.PathSegment, g *Graph) {
	for _, p := range paths {
		dx := p.End.X - p.Start.X
		dz := p.End.Z - p.Start.Z
		length := math.Hypot(dx, dz)
		midX := (p.Start.X + p.End.X) / 2
		midZ := (p.Start.Z + p.End.Z) / 2
		angle := math.Atan2(dz, dx)

		addEntity(g, Entity{
			ID:   p.ID,
			Type: EntityPath,
			Position: Vec3{
				X: midX,
				Y: 0,
				Z: midZ,
			},
			Dimensions: Vec3{
				X: p.WidthM,
				Y: 0.1,
				Z: length,
			},
			Rotation: yawQuat(angle),
			Material: "paver",
			Pod:      p.PodID,
			Layer:    LayerSurface,
			Metadata: map[string]any{"path_type": p.Type},
		})
	}
}

func assembleRouting(segments []routing.Segment, g *Graph) {
	for _, seg := range segments {
		dx := seg.End[0] - seg.Start[0]
		dz := seg.End[2] - seg.Start[2]
		length := math.Hypot(dx, dz)
		midX := (seg.Start[0] + seg.End[0]) / 2
		midZ := (seg.Start[2] + seg.End[2]) / 2
		angle := math.Atan2(dz, dx)

		eType := EntityPipe
		mat := "concrete"
		pipeH := 1.5

		switch seg.Network {
		case routing.NetworkSewage:
			mat = "concrete"
		case routing.NetworkWater:
			mat = "steel"
		case routing.NetworkElectrical:
			mat = "copper"
		case routing.NetworkTelecom:
			mat = "fiber"
		case routing.NetworkVehicle:
			eType = EntityLane
			mat = "asphalt"
			pipeH = 3.0
		}

		layer := layerFromInt(seg.Layer)
		sys := systemFromNetwork(seg.Network)

		addEntity(g, Entity{
			ID:   seg.ID,
			Type: eType,
			Position: Vec3{
				X: midX,
				Y: seg.Start[1],
				Z: midZ,
			},
			Dimensions: Vec3{
				X: seg.WidthM,
				Y: pipeH,
				Z: length,
			},
			Rotation: yawQuat(angle),
			Material: mat,
			System:   sys,
			Layer:    layer,
			Metadata: map[string]any{
				"capacity": seg.Capacity,
				"is_trunk": seg.IsTrunk,
				"network":  string(seg.Network),
			},
		})
	}
}

func assembleParks(zones []layout.Zone, g *Graph) {
	for _, z := range zones {
		centroid := z.Polygon.Centroid()
		bbMin, bbMax := z.Polygon.BoundingBox()

		addEntity(g, Entity{
			ID:   fmt.Sprintf("%s_park", z.ID),
			Type: EntityPark,
			Position: Vec3{
				X: centroid.X,
				Y: 0,
				Z: centroid.Z,
			},
			Dimensions: Vec3{
				X: bbMax.X - bbMin.X,
				Y: 0.5,
				Z: bbMax.Z - bbMin.Z,
			},
			Rotation: identityQuat(),
			Material: "grass",
			Pod:      z.PodID,
			Layer:    LayerSurface,
			Metadata: map[string]any{"area_ha": z.AreaHa},
		})
	}
}

// addEntity appends an entity and updates all group indices.
func addEntity(g *Graph, e Entity) {
	g.Entities = append(g.Entities, e)
	id := e.ID

	if e.Pod != "" {
		g.Groups.Pods[e.Pod] = append(g.Groups.Pods[e.Pod], id)
	}
	if e.System != "" {
		g.Groups.Systems[e.System] = append(g.Groups.Systems[e.System], id)
	}
	g.Groups.Layers[e.Layer] = append(g.Groups.Layers[e.Layer], id)
	g.Groups.EntityTypes[e.Type] = append(g.Groups.EntityTypes[e.Type], id)
}

// computeBounds calculates the AABB of all entities.
func computeBounds(entities []Entity) BoundingBox {
	if len(entities) == 0 {
		return BoundingBox{}
	}
	minV := Vec3{X: math.MaxFloat64, Y: math.MaxFloat64, Z: math.MaxFloat64}
	maxV := Vec3{X: -math.MaxFloat64, Y: -math.MaxFloat64, Z: -math.MaxFloat64}

	for _, e := range entities {
		halfX := e.Dimensions.X / 2
		halfZ := e.Dimensions.Z / 2

		loX := e.Position.X - halfX
		hiX := e.Position.X + halfX
		loY := e.Position.Y
		hiY := e.Position.Y + e.Dimensions.Y
		loZ := e.Position.Z - halfZ
		hiZ := e.Position.Z + halfZ

		if loX < minV.X {
			minV.X = loX
		}
		if hiX > maxV.X {
			maxV.X = hiX
		}
		if loY < minV.Y {
			minV.Y = loY
		}
		if hiY > maxV.Y {
			maxV.Y = hiY
		}
		if loZ < minV.Z {
			minV.Z = loZ
		}
		if hiZ > maxV.Z {
			maxV.Z = hiZ
		}
	}
	return BoundingBox{Min: minV, Max: maxV}
}

func identityQuat() [4]float64 {
	return [4]float64{0, 0, 0, 1}
}

func yawQuat(angle float64) [4]float64 {
	half := angle / 2
	return [4]float64{0, math.Sin(half), 0, math.Cos(half)}
}

func layerFromInt(l int) LayerType {
	switch l {
	case 1:
		return LayerUnderground1
	case 2:
		return LayerUnderground2
	case 3:
		return LayerUnderground3
	default:
		return LayerSurface
	}
}

func systemFromNetwork(n routing.NetworkType) SystemType {
	switch n {
	case routing.NetworkSewage:
		return SystemSewage
	case routing.NetworkWater:
		return SystemWater
	case routing.NetworkElectrical:
		return SystemElectrical
	case routing.NetworkTelecom:
		return SystemTelecom
	case routing.NetworkVehicle:
		return SystemVehicle
	default:
		return ""
	}
}
