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
	bikePaths []layout.BikePath,
	shuttleRoutes []layout.ShuttleRoute,
	stations []layout.Station,
	sportsFields []layout.SportsField,
) *Graph {
	g := NewGraph()

	assembleBuildings(buildings, g)
	assemblePaths(paths, g)
	assembleRouting(segments, g)
	assembleParks(greenZones, g)
	assembleBatteries(s, pods, g)
	assembleBikePaths(bikePaths, g)
	assembleShuttleRoutes(shuttleRoutes, g)
	assembleStations(stations, g)
	assembleSportsFields(sportsFields, g)

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
				Y: 0.3,
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
		case routing.NetworkPedway:
			eType = EntityPedway
			mat = "paver"
			pipeH = 2.5
		case routing.NetworkBikeTunnel:
			eType = EntityBikeTunnel
			mat = "paver"
			pipeH = 2.5
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
				"capacity":     seg.Capacity,
				"is_trunk":     seg.IsTrunk,
				"network":      string(seg.Network),
				"connected_to": seg.ConnectedTo,
			},
		})
	}
}

func assembleParks(zones []layout.Zone, g *Graph) {
	for _, z := range zones {
		centroid := z.Polygon.Centroid()

		// Use actual polygon area to derive park dimensions.
		// Green zones are thin crescents whose bounding box is much larger
		// than their actual area. Approximate as a rectangle with 3:1 aspect.
		area := z.Polygon.Area()
		parkW := math.Sqrt(area * 3)
		parkD := area / parkW
		if parkW < parkD {
			parkW, parkD = parkD, parkW
		}

		// Orient the park along the tangent (perpendicular to radial direction).
		dist := math.Hypot(centroid.X, centroid.Z)
		angle := 0.0
		if dist > 1 {
			angle = math.Atan2(centroid.Z, centroid.X) + math.Pi/2
		}

		addEntity(g, Entity{
			ID:   fmt.Sprintf("%s_park", z.ID),
			Type: EntityPark,
			Position: Vec3{
				X: centroid.X,
				Y: 0,
				Z: centroid.Z,
			},
			Dimensions: Vec3{
				X: parkW,
				Y: 0.5,
				Z: parkD,
			},
			Rotation: yawQuat(angle),
			Material: "grass",
			Pod:      z.PodID,
			Layer:    LayerSurface,
			Metadata: map[string]any{"area_ha": z.AreaHa},
		})
	}
}

func assembleBatteries(s *spec.CitySpec, pods []layout.Pod, g *Graph) {
	totalMWh := s.Infrastructure.Electrical.BatteryCapacityMWh
	if totalMWh <= 0 {
		return
	}
	mwhPerPod := totalMWh / float64(len(pods))

	for i, pod := range pods {
		addEntity(g, Entity{
			ID:   fmt.Sprintf("battery_%03d", i),
			Type: EntityBattery,
			Position: Vec3{
				X: pod.Center[0],
				Y: -4.5, // Layer 2 depth
				Z: pod.Center[1],
			},
			Dimensions: Vec3{
				X: 20.0,
				Y: 3.0,
				Z: 20.0,
			},
			Rotation: identityQuat(),
			Material: "steel",
			System:   SystemElectrical,
			Pod:      pod.ID,
			Layer:    LayerUnderground2,
			Metadata: map[string]any{
				"capacity_mwh": mwhPerPod,
				"type":         "battery_storage",
			},
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

func assembleBikePaths(bikePaths []layout.BikePath, g *Graph) {
	segIdx := 0
	for _, bp := range bikePaths {
		// Segment each polyline into consecutive point pairs.
		for i := 1; i < len(bp.Points); i++ {
			p1 := bp.Points[i-1]
			p2 := bp.Points[i]
			dx := p2.X - p1.X
			dz := p2.Z - p1.Z
			length := math.Hypot(dx, dz)
			if length < 0.1 {
				continue
			}
			midX := (p1.X + p2.X) / 2
			midZ := (p1.Z + p2.Z) / 2
			angle := math.Atan2(dz, dx)

			addEntity(g, Entity{
				ID:   fmt.Sprintf("%s_seg_%d", bp.ID, segIdx),
				Type: EntityBikePath,
				Position: Vec3{
					X: midX,
					Y: bp.ElevatedM,
					Z: midZ,
				},
				Dimensions: Vec3{
					X: bp.WidthM,
					Y: 0.3,
					Z: length,
				},
				Rotation: yawQuat(angle),
				Material: "asphalt",
				System:   SystemBicycle,
				Layer:    LayerSurface,
				Metadata: map[string]any{
					"bike_path_id": bp.ID,
					"path_type":    bp.Type,
					"elevated_m":   bp.ElevatedM,
				},
			})
			segIdx++
		}
	}
}

func assembleShuttleRoutes(routes []layout.ShuttleRoute, g *Graph) {
	segIdx := 0
	for _, sr := range routes {
		for i := 1; i < len(sr.Points); i++ {
			p1 := sr.Points[i-1]
			p2 := sr.Points[i]
			dx := p2.X - p1.X
			dz := p2.Z - p1.Z
			length := math.Hypot(dx, dz)
			if length < 0.1 {
				continue
			}
			midX := (p1.X + p2.X) / 2
			midZ := (p1.Z + p2.Z) / 2
			angle := math.Atan2(dz, dx)

			addEntity(g, Entity{
				ID:   fmt.Sprintf("%s_seg_%d", sr.ID, segIdx),
				Type: EntityShuttleRoute,
				Position: Vec3{
					X: midX,
					Y: 0,
					Z: midZ,
				},
				Dimensions: Vec3{
					X: sr.WidthM,
					Y: 0.3,
					Z: length,
				},
				Rotation: yawQuat(angle),
				Material: "asphalt",
				System:   SystemShuttle,
				Layer:    LayerSurface,
				Metadata: map[string]any{
					"shuttle_route_id": sr.ID,
					"route_type":       sr.Type,
				},
			})
			segIdx++
		}
	}
}

func assembleStations(stations []layout.Station, g *Graph) {
	for _, st := range stations {
		// Station platform: 20m x 5m x 10m concrete platform.
		dist := math.Hypot(st.Position.X, st.Position.Z)
		angle := 0.0
		if dist > 1 {
			angle = math.Atan2(st.Position.Z, st.Position.X) + math.Pi/2
		}

		addEntity(g, Entity{
			ID:   st.ID,
			Type: EntityStation,
			Position: Vec3{
				X: st.Position.X,
				Y: 0,
				Z: st.Position.Z,
			},
			Dimensions: Vec3{
				X: 20.0,
				Y: 5.0,
				Z: 10.0,
			},
			Rotation: yawQuat(angle),
			Material: "concrete",
			System:   SystemShuttle,
			Pod:      st.PodID,
			Layer:    LayerSurface,
			Metadata: map[string]any{
				"route_id": st.RouteID,
				"type":     "mobility_hub",
			},
		})
	}
}

func assembleSportsFields(fields []layout.SportsField, g *Graph) {
	for _, sf := range fields {
		mat := "grass"
		height := 0.3
		switch sf.Type {
		case "stadium":
			mat = "grass"
			height = 15.0 // stadium walls
		case "basketball", "tennis", "pickleball":
			mat = "court"
			height = 0.2
		}

		addEntity(g, Entity{
			ID:   sf.ID,
			Type: EntitySportsField,
			Position: Vec3{
				X: sf.Position.X,
				Y: 0,
				Z: sf.Position.Z,
			},
			Dimensions: Vec3{
				X: sf.Dimensions[0],
				Y: height,
				Z: sf.Dimensions[1],
			},
			Rotation: yawQuat(sf.Rotation),
			Material: mat,
			Layer:    LayerSurface,
			Metadata: map[string]any{
				"field_type": sf.Type,
				"buffer_id":  sf.BufferID,
			},
		})
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
	case routing.NetworkPedway:
		return SystemPedestrian
	case routing.NetworkBikeTunnel:
		return SystemBicycle
	default:
		return ""
	}
}
