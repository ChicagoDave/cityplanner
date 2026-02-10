package analytics

import (
	"github.com/ChicagoDave/cityplanner/pkg/spec"
	"github.com/ChicagoDave/cityplanner/pkg/validation"
)

// ResolvedParameters holds the computed values from Phase 1 analytical resolution.
type ResolvedParameters struct {
	TotalHouseholds     int     `json:"total_households"`
	TotalPopulation     int     `json:"total_population"`
	DependencyRatio     float64 `json:"dependency_ratio"`
	PodCount            int     `json:"pod_count"`
	RequiredDensityDUHa float64 `json:"required_density_du_ha"`
	TotalAreaHa         float64 `json:"total_area_ha"`
	ExcavationVolumeM3  float64 `json:"excavation_volume_m3"`
	PerCapitaCost       float64 `json:"per_capita_cost"`
	BreakEvenRent       float64 `json:"break_even_monthly_rent"`

	Cohorts       []CohortBreakdown `json:"cohorts"`
	Rings         []RingData        `json:"rings"`
	Services      []ServiceCount    `json:"services"`
	Areas         AreaBreakdown     `json:"areas"`
	Energy        EnergyBalance     `json:"energy"`
	TotalAdults   int               `json:"total_adults"`
	TotalChildren int               `json:"total_children"`
	TotalStudents int               `json:"total_students"`
	WeightedAvgHH float64           `json:"weighted_avg_household_size"`
}

// Resolve runs Phase 1 analytical resolution on the spec.
// It computes demographics, service counts, density, area, and cost estimates.
// Returns resolved parameters and a validation report.
func Resolve(s *spec.CitySpec) (*ResolvedParameters, *validation.Report) {
	report := validation.NewReport()

	// 1. Demographics
	cohorts, weightedAvg := resolveDemographics(s)
	depRatio := computeDependencyRatio(cohorts)
	adults, children, students := sumCohortTotals(cohorts)

	// 2. Areas
	areas := resolveAreas(s)

	// 3. Rings (capacity-weighted population distribution)
	rings := resolveRings(s, s.City.Population)

	// Total households derived from per-ring household counts.
	totalHH := 0
	for _, r := range rings {
		totalHH += r.Households
	}

	// 4. Pod count
	podCount := 0
	for _, r := range rings {
		podCount += r.PodCount
	}

	// 5. Services
	services := resolveServices(s.City.Population, students)

	// 6. Energy
	energy := resolveEnergy(s)

	// 7. Excavation volume
	excavVol := areas.TotalCityHa * m2PerHa * s.City.ExcavationDepth

	// 8. Overall required density
	requiredDensity := 0.0
	if areas.ResidentialHa > 0 {
		requiredDensity = float64(totalHH) / areas.ResidentialHa
	}

	params := &ResolvedParameters{
		TotalHouseholds:     totalHH,
		TotalPopulation:     s.City.Population,
		DependencyRatio:     depRatio,
		PodCount:            podCount,
		RequiredDensityDUHa: requiredDensity,
		TotalAreaHa:         areas.TotalCityHa,
		ExcavationVolumeM3:  excavVol,
		Cohorts:             cohorts,
		Rings:               rings,
		Services:            services,
		Areas:               areas,
		Energy:              energy,
		TotalAdults:         adults,
		TotalChildren:       children,
		TotalStudents:       students,
		WeightedAvgHH:       weightedAvg,
	}

	// 9. Analytical validation
	validateAnalytical(s, params, report)

	return params, report
}
