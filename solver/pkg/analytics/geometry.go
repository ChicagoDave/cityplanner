package analytics

import (
	"math"

	"github.com/ChicagoDave/cityplanner/pkg/spec"
)

const (
	// Land use fractions within the city area (inside edge zone).
	residentialFraction = 0.60
	commercialFraction  = 0.15
	civicFraction       = 0.10
	greenPathsFraction  = 0.15

	// Density model constants.
	groundCoverage = 0.60  // building footprint / lot area
	avgUnitSizeM2  = 75.0  // average dwelling unit floor area in m²
	m2PerHa        = 10000 // square meters per hectare
)

// resolveRings computes per-ring area, population, pod, and density data.
func resolveRings(s *spec.CitySpec, totalHouseholds int, totalPop int) []RingData {
	zones := []struct {
		name       string
		from, to   float64
		maxStories int
	}{
		{"center", s.CityZones.Center.RadiusFrom, s.CityZones.Center.RadiusTo, s.CityZones.Center.MaxStories},
		{"middle", s.CityZones.Middle.RadiusFrom, s.CityZones.Middle.RadiusTo, s.CityZones.Middle.MaxStories},
		{"edge", s.CityZones.Edge.RadiusFrom, s.CityZones.Edge.RadiusTo, s.CityZones.Edge.MaxStories},
	}

	totalCityAreaM2 := math.Pi * s.CityZones.Edge.RadiusTo * s.CityZones.Edge.RadiusTo
	podAreaM2 := math.Pi * s.Pods.WalkRadius * s.Pods.WalkRadius

	rings := make([]RingData, 0, len(zones))
	for _, z := range zones {
		ringAreaM2 := math.Pi * (z.to*z.to - z.from*z.from)
		areaHa := ringAreaM2 / m2PerHa
		fraction := ringAreaM2 / totalCityAreaM2

		ringPop := int(math.Round(float64(totalPop) * fraction))
		ringHH := int(math.Round(float64(totalHouseholds) * fraction))

		podCount := int(math.Ceil(ringAreaM2 / podAreaM2))
		if podCount < 1 {
			podCount = 1
		}

		podPop := 0
		if podCount > 0 {
			podPop = ringPop / podCount
		}

		residentialHa := areaHa * residentialFraction

		// Required density: households per residential hectare
		requiredDensity := 0.0
		if residentialHa > 0 {
			requiredDensity = float64(ringHH) / residentialHa
		}

		// Achievable density at max stories:
		// stories * ground_coverage * m2PerHa / avg_unit_size
		achievableDensity := float64(z.maxStories) * groundCoverage * m2PerHa / avgUnitSizeM2

		rings = append(rings, RingData{
			Name:              z.name,
			RadiusFrom:        z.from,
			RadiusTo:          z.to,
			AreaHa:            areaHa,
			AreaFraction:      fraction,
			Population:        ringPop,
			Households:        ringHH,
			PodCount:          podCount,
			PodPopulation:     podPop,
			MaxStories:        z.maxStories,
			RequiredDensity:   requiredDensity,
			AchievableDensity: achievableDensity,
			ResidentialAreaHa: residentialHa,
		})
	}

	return rings
}

// resolveAreas computes the land-use area breakdown.
func resolveAreas(s *spec.CitySpec) AreaBreakdown {
	cityAreaM2 := math.Pi * s.CityZones.Edge.RadiusTo * s.CityZones.Edge.RadiusTo
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
