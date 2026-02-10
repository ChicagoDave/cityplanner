package analytics

import (
	"math"

	"github.com/ChicagoDave/cityplanner/pkg/spec"
)

const (
	// Land use fractions within the city area (inside outermost ring).
	residentialFraction = 0.60
	commercialFraction  = 0.15
	civicFraction       = 0.10
	greenPathsFraction  = 0.15

	// Density model constants.
	groundCoverage = 0.60  // building footprint / lot area
	avgUnitSizeM2  = 75.0  // average dwelling unit floor area in m²
	m2PerHa        = 10000 // square meters per hectare
)

// characterResidentialFraction returns the fraction of floor area that is
// residential for a given ring character. Inner rings are more commercial/civic;
// outer rings are predominantly residential.
func characterResidentialFraction(character string) float64 {
	switch character {
	case "civic_commercial":
		return 0.30
	case "high_density":
		return 0.65
	case "urban_midrise":
		return 0.60
	case "mixed_residential":
		return 0.70
	case "low_density":
		return 0.75
	default:
		return 0.60
	}
}

// characterAvgHouseholdSize returns the average household size for a ring
// character. Inner rings attract singles and couples (smaller households);
// outer rings attract families (larger households).
func characterAvgHouseholdSize(character string) float64 {
	switch character {
	case "civic_commercial":
		return 1.8
	case "high_density":
		return 1.8
	case "urban_midrise":
		return 2.2
	case "mixed_residential":
		return 3.0
	case "low_density":
		return 3.5
	default:
		return 2.5
	}
}

// resolveRings computes per-ring area, population, pod, and density data.
// Supports an arbitrary number of rings. Population is distributed proportionally
// to each ring's residential capacity weight (area × max_stories × residential_fraction).
// Inner rings with tall buildings get more people per pod; outer rings with
// family housing get fewer.
func resolveRings(s *spec.CitySpec, totalPop int) []RingData {
	outerRadius := s.CityZones.OuterRadius()
	totalCityAreaM2 := math.Pi * outerRadius * outerRadius
	podAreaM2 := math.Pi * s.Pods.WalkRadius * s.Pods.WalkRadius

	// First pass: compute pod counts (geometry-based) and capacity weights.
	type ringInfo struct {
		ring     spec.RingDef
		areaM2   float64
		areaHa   float64
		podCount int
		weight   float64 // residential capacity weight
		resFrac  float64 // residential fraction for this character
		avgHH    float64 // average household size for this character
	}

	infos := make([]ringInfo, len(s.CityZones.Rings))
	totalPodCount := 0
	totalWeight := 0.0

	for i, ring := range s.CityZones.Rings {
		areaM2 := math.Pi * (ring.RadiusTo*ring.RadiusTo - ring.RadiusFrom*ring.RadiusFrom)
		areaHa := areaM2 / m2PerHa
		podCount := int(math.Ceil(areaM2 / podAreaM2))
		if podCount < 1 {
			podCount = 1
		}
		resFrac := characterResidentialFraction(ring.Character)
		avgHH := characterAvgHouseholdSize(ring.Character)
		weight := areaHa * float64(ring.MaxStories) * resFrac

		infos[i] = ringInfo{
			ring:     ring,
			areaM2:   areaM2,
			areaHa:   areaHa,
			podCount: podCount,
			weight:   weight,
			resFrac:  resFrac,
			avgHH:    avgHH,
		}
		totalPodCount += podCount
		totalWeight += weight
	}

	// Second pass: distribute population by capacity weight.
	rings := make([]RingData, 0, len(infos))
	assignedPop := 0

	for i, info := range infos {
		fraction := info.areaM2 / totalCityAreaM2

		// Population proportional to residential capacity weight.
		ringPop := 0
		if totalWeight > 0 {
			if i == len(infos)-1 {
				ringPop = totalPop - assignedPop // last ring absorbs rounding
			} else {
				ringPop = int(math.Round(float64(totalPop) * info.weight / totalWeight))
			}
		}
		assignedPop += ringPop

		ringHH := int(math.Round(float64(ringPop) / info.avgHH))
		podPop := ringPop / info.podCount
		residentialHa := info.areaHa * info.resFrac

		requiredDensity := 0.0
		if residentialHa > 0 {
			requiredDensity = float64(ringHH) / residentialHa
		}

		achievableDensity := float64(info.ring.MaxStories) * groundCoverage * m2PerHa / avgUnitSizeM2

		rings = append(rings, RingData{
			Name:              info.ring.Name,
			RadiusFrom:        info.ring.RadiusFrom,
			RadiusTo:          info.ring.RadiusTo,
			AreaHa:            info.areaHa,
			AreaFraction:      fraction,
			Population:        ringPop,
			Households:        ringHH,
			PodCount:          info.podCount,
			PodPopulation:     podPop,
			MaxStories:        info.ring.MaxStories,
			AvgHouseholdSize:  info.avgHH,
			RequiredDensity:   requiredDensity,
			AchievableDensity: achievableDensity,
			ResidentialAreaHa: residentialHa,
		})
	}

	return rings
}

// resolveAreas computes the land-use area breakdown.
func resolveAreas(s *spec.CitySpec) AreaBreakdown {
	outerRadius := s.CityZones.OuterRadius()
	cityAreaM2 := math.Pi * outerRadius * outerRadius
	cityHa := cityAreaM2 / m2PerHa

	perimeterM2 := math.Pi * (s.CityZones.Perimeter.RadiusTo*s.CityZones.Perimeter.RadiusTo -
		s.CityZones.Perimeter.RadiusFrom*s.CityZones.Perimeter.RadiusFrom)
	perimeterHa := perimeterM2 / m2PerHa

	solarHa := s.CityZones.SolarRing.AreaHa
	if solarHa == 0 {
		// Compute from radii if area not specified
		solarM2 := math.Pi * (s.CityZones.SolarRing.RadiusTo*s.CityZones.SolarRing.RadiusTo -
			s.CityZones.SolarRing.RadiusFrom*s.CityZones.SolarRing.RadiusFrom)
		solarHa = solarM2 / m2PerHa
	}

	return AreaBreakdown{
		TotalCityHa:        cityHa,
		ResidentialHa:      cityHa * residentialFraction,
		CommercialHa:       cityHa * commercialFraction,
		CivicHa:            cityHa * civicFraction,
		GreenPathsHa:       cityHa * greenPathsFraction,
		PerimeterHa:        perimeterHa,
		SolarHa:            solarHa,
		TotalWithPerimeter: cityHa + perimeterHa + solarHa,
	}
}

// resolveEnergy computes the electrical supply/demand balance.
func resolveEnergy(s *spec.CitySpec) EnergyBalance {
	peakMW := float64(s.City.Population) * s.Infrastructure.Electrical.PeakDemandKWPer / 1000.0
	totalGen := s.Infrastructure.Electrical.SolarIntegratedAvgMW + s.Infrastructure.Electrical.SolarFarmAvgMW
	batteryMWh := s.Infrastructure.Electrical.BatteryCapacityMWh

	backupHours := 0.0
	if peakMW > 0 {
		backupHours = batteryMWh / peakMW
	}

	return EnergyBalance{
		PeakDemandMW:       peakMW,
		SolarIntegratedMW:  s.Infrastructure.Electrical.SolarIntegratedAvgMW,
		SolarFarmMW:        s.Infrastructure.Electrical.SolarFarmAvgMW,
		TotalGenerationMW:  totalGen,
		GridCapacityMW:     s.Infrastructure.Electrical.GridCapacityMW,
		BatteryCapacityMWh: batteryMWh,
		BackupHours:        backupHours,
	}
}
