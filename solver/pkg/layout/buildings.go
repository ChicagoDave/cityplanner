package layout

import (
	"fmt"
	"math"

	"github.com/ChicagoDave/cityplanner/pkg/analytics"
	"github.com/ChicagoDave/cityplanner/pkg/geo"
	"github.com/ChicagoDave/cityplanner/pkg/spec"
	"github.com/ChicagoDave/cityplanner/pkg/validation"
)

// Building represents a placed building within a pod.
type Building struct {
	ID            string     `json:"id"`
	PodID         string     `json:"pod_id"`
	Type          string     `json:"type"` // residential, commercial, civic, service
	Position      [3]float64 `json:"position"`
	Footprint     [2]float64 `json:"footprint"` // [width, depth] in meters
	Stories       int        `json:"stories"`
	DwellingUnits int        `json:"dwelling_units,omitempty"`
	CommercialSqM float64    `json:"commercial_sqm,omitempty"`
	ServiceType   string     `json:"service_type,omitempty"`
}

// PathSegment represents a pedestrian or bicycle path within a pod.
type PathSegment struct {
	ID     string      `json:"id"`
	PodID  string      `json:"pod_id"`
	Start  geo.Point2D `json:"start"`
	End    geo.Point2D `json:"end"`
	WidthM float64     `json:"width_m"`
	Type   string      `json:"type"` // spine, connector, inter_pod
}

// serviceFootprints defines building dimensions for service types.
var serviceFootprints = map[string][2]float64{
	"hospital":          {60, 40},
	"elementary_school": {40, 30},
	"secondary_school":  {50, 35},
	"library":           {25, 20},
	"grocery":           {30, 25},
	"medical_clinic":    {25, 20},
	"performing_arts":   {40, 30},
	"city_hall":         {50, 35},
	"coworking_hub":     {30, 25},
	"coworking":         {25, 20},
	"retail":            {20, 15},
	"restaurant":        {15, 12},
	"playground":        {30, 30},
	"pediatric_clinic":  {25, 20},
	"daycare":           {25, 20},
}

// PlaceBuildings generates building placements within laid-out pods.
// Orchestrates: zones → paths → blocks → building placement.
// Returns buildings, path segments, and a validation report.
func PlaceBuildings(s *spec.CitySpec, pods []Pod, adjacency map[string][]string, params *analytics.ResolvedParameters) ([]Building, []PathSegment, *validation.Report) {
	report := validation.NewReport()

	// Global unit mix for the entire city.
	cityMix := DistributeUnits(params.TotalHouseholds, params.Cohorts)

	rings := s.CityZones.Rings

	var allBuildings []Building
	var allPaths []PathSegment
	buildingIdx := 0
	totalDU := 0

	// Build a pod center lookup for inter-pod paths.
	podCenterMap := make(map[string]geo.Point2D)
	for _, p := range pods {
		podCenterMap[p.ID] = p.CenterPoint()
	}

	// Build a ring radii lookup from spec rings.
	ringRadii := make(map[string][2]float64, len(rings))
	for _, ring := range rings {
		ringRadii[ring.Name] = [2]float64{ring.RadiusFrom, ring.RadiusTo}
	}

	for _, pod := range pods {
		// Determine ring character for zone proportions.
		ringChar := ""
		if pr, ok := s.Pods.RingAssignments[pod.Ring]; ok {
			ringChar = pr.Character
		}

		// 1. Zone allocation using radial bands.
		radii := ringRadii[pod.Ring]
		zones := AllocateZones(pod, ringChar, radii[0], radii[1])
		if len(zones) == 0 {
			report.AddWarning(validation.Result{
				Level:   validation.LevelSpatial,
				Message: fmt.Sprintf("pod %s: no zones allocated", pod.ID),
			})
			continue
		}

		// 2. Path network.
		adjCenters := make(map[string]geo.Point2D)
		for _, adjID := range adjacency[pod.ID] {
			if c, ok := podCenterMap[adjID]; ok {
				adjCenters[adjID] = c
			}
		}
		paths := GeneratePaths(pod, zones, adjCenters)
		allPaths = append(allPaths, paths...)

		// 3. Scale unit mix proportionally to this pod's population.
		popFraction := float64(pod.TargetPopulation) / float64(params.TotalPopulation)
		podMix := ScaleUnitMix(cityMix, popFraction)
		podDUTarget := podMix.Total()
		podDU := 0

		// 4. Process each zone.
		for _, zone := range zones {
			// Subdivide zone into blocks.
			blocks := SubdivideIntoBlocks(zone, pod.CenterPoint())

			switch zone.Type {
			case ZoneResidential:
				for _, block := range blocks {
					if podDU >= podDUTarget {
						break
					}
					buildings, du := placeResidentialOnBlock(block, pod, rings, &buildingIdx)
					allBuildings = append(allBuildings, buildings...)
					podDU += du
				}

			case ZoneCommercial:
				// Cap commercial: ~1 per 500 residents in the pod.
				comTarget := pod.TargetPopulation / 500
				if comTarget < 2 {
					comTarget = 2
				}
				comPlaced := 0
				for _, block := range blocks {
					if comPlaced >= comTarget {
						break
					}
					buildings := placeCommercialOnBlock(block, pod, rings, &buildingIdx)
					remaining := comTarget - comPlaced
					if len(buildings) > remaining {
						buildings = buildings[:remaining]
					}
					allBuildings = append(allBuildings, buildings...)
					comPlaced += len(buildings)
				}

			case ZoneCivic:
				// Place service buildings directly within the civic zone.
				// Bypass block subdivision since civic zones can be narrow.
				if pr, ok := s.Pods.RingAssignments[pod.Ring]; ok {
					for si, svc := range pr.RequiredServices {
						b := placeServiceAtZone(zone, pod, svc, si, rings, &buildingIdx)
						allBuildings = append(allBuildings, b)
					}
				}

			case ZoneGreen:
				// Green space: no buildings placed. Parks tracked as zones.
			}
		}

		totalDU += podDU
	}

	// Validation.
	duRatio := float64(totalDU) / float64(params.TotalHouseholds)
	if duRatio < 0.80 {
		report.AddWarning(validation.Result{
			Level:   validation.LevelSpatial,
			Message: fmt.Sprintf("dwelling unit shortfall: placed %d of %d target (%.0f%%)", totalDU, params.TotalHouseholds, duRatio*100),
		})
	}

	report.AddInfo(validation.Result{
		Level:   validation.LevelSpatial,
		Message: fmt.Sprintf("placed %d buildings (%d dwelling units) and %d path segments", len(allBuildings), totalDU, len(allPaths)),
	})

	return allBuildings, allPaths, report
}

// placeResidentialOnBlock places residential buildings on a block using a
// courtyard pattern: buildings around the perimeter with open center.
func placeResidentialOnBlock(block Block, pod Pod, rings []spec.RingDef, idx *int) ([]Building, int) {
	const (
		buildingW = 20.0 // width (m)
		buildingD = 15.0 // depth (m)
		spacing   = 2.0  // gap between buildings (m)
		setback   = 3.0  // setback from block edge (m)
		floorH    = 3.0  // story height (m)
		unitArea  = 75.0 // average unit area (m²)
	)

	centroid := block.Polygon.Centroid()
	dist := centroid.Distance(geo.Origin)
	stories := MaxStoriesFromRings(dist, rings)
	unitsPerFloor := int(math.Max(1, math.Floor(buildingW*buildingD/unitArea)))
	unitsPerBuilding := unitsPerFloor * stories

	// Determine local axes.
	outward := centroid.Normalize()
	if centroid.Length() < 1 {
		outward = geo.Pt(1, 0)
	}
	perp := outward.Perp()

	// Bounding box extent.
	bbMin, bbMax := block.Polygon.BoundingBox()
	extentU := bbMax.X - bbMin.X
	extentV := bbMax.Z - bbMin.Z

	// Place buildings in a grid within the block.
	stepU := buildingW + spacing
	stepV := buildingD + spacing
	availU := extentU - 2*setback
	availV := extentV - 2*setback

	var buildings []Building
	totalDU := 0

	numU := int(math.Max(1, math.Floor(availU/stepU)))
	numV := int(math.Max(1, math.Floor(availV/stepV)))

	for iu := 0; iu < numU; iu++ {
		for iv := 0; iv < numV; iv++ {
			bx := bbMin.X + setback + float64(iu)*stepU + buildingW/2
			bz := bbMin.Z + setback + float64(iv)*stepV + buildingD/2
			pos := geo.Pt(bx, bz)

			// Verify position is inside the block.
			if !block.Polygon.Contains(pos) {
				continue
			}

			_ = floorH
			_ = perp
			_ = outward

			buildings = append(buildings, Building{
				ID:            fmt.Sprintf("bldg_%05d", *idx),
				PodID:         pod.ID,
				Type:          "residential",
				Position:      [3]float64{pos.X, 0, pos.Z},
				Footprint:     [2]float64{buildingW, buildingD},
				Stories:       stories,
				DwellingUnits: unitsPerBuilding,
			})
			totalDU += unitsPerBuilding
			*idx++
		}
	}

	return buildings, totalDU
}

// placeCommercialOnBlock places commercial buildings on a block.
func placeCommercialOnBlock(block Block, pod Pod, rings []spec.RingDef, idx *int) []Building {
	const (
		buildingW     = 25.0
		buildingD     = 20.0
		spacing       = 2.0
		setback       = 3.0
		maxComStories = 6
	)

	centroid := block.Polygon.Centroid()
	dist := centroid.Distance(geo.Origin)
	stories := MaxStoriesFromRings(dist, rings)
	if stories > maxComStories {
		stories = maxComStories
	}

	bbMin, bbMax := block.Polygon.BoundingBox()
	availU := (bbMax.X - bbMin.X) - 2*setback
	availV := (bbMax.Z - bbMin.Z) - 2*setback

	stepU := buildingW + spacing
	stepV := buildingD + spacing

	var buildings []Building
	numU := int(math.Max(1, math.Floor(availU/stepU)))
	numV := int(math.Max(1, math.Floor(availV/stepV)))

	for iu := 0; iu < numU; iu++ {
		for iv := 0; iv < numV; iv++ {
			bx := bbMin.X + setback + float64(iu)*stepU + buildingW/2
			bz := bbMin.Z + setback + float64(iv)*stepV + buildingD/2
			pos := geo.Pt(bx, bz)
			if !block.Polygon.Contains(pos) {
				continue
			}
			sqm := buildingW * buildingD * float64(stories) * 0.80 // 80% usable
			buildings = append(buildings, Building{
				ID:            fmt.Sprintf("bldg_%05d", *idx),
				PodID:         pod.ID,
				Type:          "commercial",
				Position:      [3]float64{pos.X, 0, pos.Z},
				Footprint:     [2]float64{buildingW, buildingD},
				Stories:       stories,
				CommercialSqM: sqm,
			})
			*idx++
		}
	}

	return buildings
}

// placeServiceAtZone places a service building within a zone when no blocks
// are available, using the zone centroid with an offset for each service.
func placeServiceAtZone(zone Zone, pod Pod, serviceType string, index int, rings []spec.RingDef, idx *int) Building {
	fp, ok := serviceFootprints[serviceType]
	if !ok {
		fp = [2]float64{25, 20}
	}

	centroid := zone.Polygon.Centroid()
	// Offset each service building to avoid overlap.
	offset := float64(index) * 40.0
	outward := centroid.Normalize()
	if centroid.Length() < 1 {
		outward = geo.Pt(1, 0)
	}
	pos := centroid.Add(outward.Perp().Scale(offset - float64(index)*20))

	dist := pos.Distance(geo.Origin)
	stories := MaxStoriesFromRings(dist, rings)

	switch serviceType {
	case "hospital":
		// Use full height.
	case "elementary_school", "secondary_school":
		if stories > 3 {
			stories = 3
		}
	case "playground":
		stories = 1
	default:
		if stories > 4 {
			stories = 4
		}
	}

	b := Building{
		ID:          fmt.Sprintf("bldg_%05d", *idx),
		PodID:       pod.ID,
		Type:        "civic",
		Position:    [3]float64{pos.X, 0, pos.Z},
		Footprint:   [2]float64{fp[0], fp[1]},
		Stories:     stories,
		ServiceType: serviceType,
	}
	*idx++
	return b
}

// placeServiceBuilding places a civic/service building on a block.
func placeServiceBuilding(block Block, pod Pod, serviceType string, rings []spec.RingDef, idx *int) Building {
	fp, ok := serviceFootprints[serviceType]
	if !ok {
		fp = [2]float64{25, 20}
	}

	centroid := block.Polygon.Centroid()
	dist := centroid.Distance(geo.Origin)
	stories := MaxStoriesFromRings(dist, rings)

	// Service buildings are typically shorter.
	switch serviceType {
	case "hospital":
		// Use full height.
	case "elementary_school", "secondary_school":
		if stories > 3 {
			stories = 3
		}
	case "playground":
		stories = 1
	default:
		if stories > 4 {
			stories = 4
		}
	}

	b := Building{
		ID:          fmt.Sprintf("bldg_%05d", *idx),
		PodID:       pod.ID,
		Type:        "civic",
		Position:    [3]float64{centroid.X, 0, centroid.Z},
		Footprint:   [2]float64{fp[0], fp[1]},
		Stories:     stories,
		ServiceType: serviceType,
	}
	*idx++
	return b
}
