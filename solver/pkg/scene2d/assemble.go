package scene2d

import (
	"time"

	"github.com/ChicagoDave/cityplanner/pkg/analytics"
	"github.com/ChicagoDave/cityplanner/pkg/geo"
	"github.com/ChicagoDave/cityplanner/pkg/layout"
	"github.com/ChicagoDave/cityplanner/pkg/spec"
)

// Assemble2D converts solver outputs into a 2D scene suitable for SVG rendering.
// Buildings and trees are summarized as aggregates; spatial data (pods, zones,
// paths, stations, sports, plazas) is preserved with 2D coordinates.
func Assemble2D(
	s *spec.CitySpec,
	params *analytics.ResolvedParameters,
	pods []layout.Pod,
	buildings []layout.Building,
	paths []layout.PathSegment,
	greenZones []layout.Zone,
	bikePaths []layout.BikePath,
	shuttleRoutes []layout.ShuttleRoute,
	stations []layout.Station,
	sportsFields []layout.SportsField,
	plazas []layout.Plaza,
	trees []layout.Tree,
) *Scene2D {
	return &Scene2D{
		Metadata:     assembleMetadata(s, params),
		Rings:        assembleRings(s, params),
		Pods:         assemblePods(s, pods),
		Paths:        assemblePaths(paths, bikePaths, shuttleRoutes),
		Stations:     assembleStations(stations),
		Sports:       assembleSports(sportsFields),
		Plazas:       assemblePlazas(plazas),
		Trees:        assembleTreeSummary(trees),
		Buildings:    assembleBuildingSummary(buildings),
		ExternalBand: assembleExternalBand(s),
	}
}

func assembleMetadata(s *spec.CitySpec, params *analytics.ResolvedParameters) Metadata {
	cityRadius := s.CityZones.OuterRadius()

	extRadius := 0.0
	if s.CityZones.SolarRing.RadiusTo > 0 {
		extRadius = s.CityZones.SolarRing.RadiusTo
	} else if s.CityZones.Perimeter.RadiusTo > 0 {
		extRadius = s.CityZones.Perimeter.RadiusTo
	}

	return Metadata{
		Population:          s.City.Population,
		PodCount:            params.PodCount,
		CityRadiusM:         cityRadius,
		ExternalBandRadiusM: extRadius,
		GeneratedAt:         time.Now().UTC().Format(time.RFC3339),
	}
}

func assembleRings(s *spec.CitySpec, params *analytics.ResolvedParameters) []Ring {
	rings := make([]Ring, 0, len(params.Rings))
	for _, rd := range params.Rings {
		character := ""
		if specRing := s.CityZones.RingByName(rd.Name); specRing != nil {
			character = specRing.Character
		}
		rings = append(rings, Ring{
			Name:       rd.Name,
			RadiusFrom: rd.RadiusFrom,
			RadiusTo:   rd.RadiusTo,
			MaxStories: rd.MaxStories,
			Character:  character,
			PodCount:   rd.PodCount,
			Population: rd.Population,
		})
	}
	return rings
}

func assemblePods(s *spec.CitySpec, pods []layout.Pod) []Pod2D {
	ringRadii := make(map[string][2]float64, len(s.CityZones.Rings))
	ringStories := make(map[string]int, len(s.CityZones.Rings))
	for _, ring := range s.CityZones.Rings {
		ringRadii[ring.Name] = [2]float64{ring.RadiusFrom, ring.RadiusTo}
		ringStories[ring.Name] = ring.MaxStories
	}

	result := make([]Pod2D, 0, len(pods))
	for _, pod := range pods {
		ringChar := ""
		if pr, ok := s.Pods.RingAssignments[pod.Ring]; ok {
			ringChar = pr.Character
		}
		radii := ringRadii[pod.Ring]
		zones := layout.AllocateZones(pod, ringChar, radii[0], radii[1])

		zones2d := make([]Zone2D, 0, len(zones))
		for _, z := range zones {
			zones2d = append(zones2d, Zone2D{
				Type:    string(z.Type),
				Polygon: polygonToCoords(z.Polygon),
				AreaHa:  z.AreaHa,
			})
		}

		result = append(result, Pod2D{
			ID:         pod.ID,
			Ring:       pod.Ring,
			Center:     pod.Center,
			Boundary:   pod.Boundary,
			Population: pod.TargetPopulation,
			MaxStories: ringStories[pod.Ring],
			AreaHa:     pod.AreaHa,
			Zones:      zones2d,
		})
	}
	return result
}

func assemblePaths(
	paths []layout.PathSegment,
	bikePaths []layout.BikePath,
	shuttleRoutes []layout.ShuttleRoute,
) PathCollection {
	pc := PathCollection{}

	pc.Pedestrian = make([]PedestrianPath2D, 0, len(paths))
	for _, p := range paths {
		pc.Pedestrian = append(pc.Pedestrian, PedestrianPath2D{
			ID:    p.ID,
			Start: [2]float64{p.Start.X, p.Start.Z},
			End:   [2]float64{p.End.X, p.End.Z},
			Width: p.WidthM,
			Type:  p.Type,
		})
	}

	pc.Bike = make([]BikePath2D, 0, len(bikePaths))
	for _, bp := range bikePaths {
		pc.Bike = append(pc.Bike, BikePath2D{
			ID:       bp.ID,
			Points:   pointsToCoords(bp.Points),
			Width:    bp.WidthM,
			Elevated: bp.ElevatedM,
			Type:     bp.Type,
		})
	}

	pc.Shuttle = make([]ShuttlePath2D, 0, len(shuttleRoutes))
	for _, sr := range shuttleRoutes {
		pc.Shuttle = append(pc.Shuttle, ShuttlePath2D{
			ID:     sr.ID,
			Points: pointsToCoords(sr.Points),
			Width:  sr.WidthM,
			Type:   sr.Type,
		})
	}

	return pc
}

func assembleStations(stations []layout.Station) []Station2D {
	result := make([]Station2D, 0, len(stations))
	for _, st := range stations {
		result = append(result, Station2D{
			ID:       st.ID,
			PodID:    st.PodID,
			Position: [2]float64{st.Position.X, st.Position.Z},
			RouteID:  st.RouteID,
		})
	}
	return result
}

func assembleSports(fields []layout.SportsField) SportsCollection {
	result := make([]SportsField2D, 0, len(fields))
	for _, f := range fields {
		result = append(result, SportsField2D{
			ID:         f.ID,
			Type:       f.Type,
			Position:   [2]float64{f.Position.X, f.Position.Z},
			Dimensions: f.Dimensions,
			Rotation:   f.Rotation,
		})
	}
	return SportsCollection{Fields: result}
}

func assemblePlazas(plazas []layout.Plaza) []Plaza2D {
	result := make([]Plaza2D, 0, len(plazas))
	for _, p := range plazas {
		result = append(result, Plaza2D{
			ID:       p.ID,
			PodID:    p.PodID,
			Position: [2]float64{p.Position.X, p.Position.Z},
			Width:    p.Width,
			Depth:    p.Depth,
			Rotation: p.Rotation,
		})
	}
	return result
}

func assembleTreeSummary(trees []layout.Tree) TreeSummary {
	var ts TreeSummary
	for _, t := range trees {
		switch t.Context {
		case "park":
			ts.ParkCount++
		case "path":
			ts.PathCount++
		case "plaza":
			ts.PlazaCount++
		}
	}
	ts.Total = ts.ParkCount + ts.PathCount + ts.PlazaCount
	return ts
}

func assembleBuildingSummary(buildings []layout.Building) BuildingSummary {
	bs := BuildingSummary{
		ByPod: make(map[string]PodBuildingSum),
	}
	for _, b := range buildings {
		bs.TotalBuildings++
		bs.TotalDU += b.DwellingUnits

		pbs := bs.ByPod[b.PodID]
		switch b.Type {
		case "residential":
			pbs.Residential++
			pbs.TotalUnits += b.DwellingUnits
		case "commercial":
			pbs.Commercial++
			pbs.CommercialSqM += b.CommercialSqM
		case "civic":
			pbs.Civic++
			if b.ServiceType != "" {
				found := false
				for _, st := range pbs.ServiceTypes {
					if st == b.ServiceType {
						found = true
						break
					}
				}
				if !found {
					pbs.ServiceTypes = append(pbs.ServiceTypes, b.ServiceType)
				}
			}
		}
		bs.ByPod[b.PodID] = pbs
	}
	return bs
}

func assembleExternalBand(s *spec.CitySpec) *ExternalBand {
	p := s.CityZones.Perimeter
	if p.RadiusFrom == 0 && p.RadiusTo == 0 {
		return nil
	}
	return &ExternalBand{
		RadiusFrom: p.RadiusFrom,
		RadiusTo:   p.RadiusTo,
		Facilities: p.Contents,
	}
}

// polygonToCoords converts a geo.Polygon to a [][2]float64 coordinate list.
func polygonToCoords(p geo.Polygon) [][2]float64 {
	coords := make([][2]float64, len(p.Vertices))
	for i, v := range p.Vertices {
		coords[i] = [2]float64{v.X, v.Z}
	}
	return coords
}

// pointsToCoords converts a []geo.Point2D to a [][2]float64 coordinate list.
func pointsToCoords(pts []geo.Point2D) [][2]float64 {
	coords := make([][2]float64, len(pts))
	for i, pt := range pts {
		coords[i] = [2]float64{pt.X, pt.Z}
	}
	return coords
}
