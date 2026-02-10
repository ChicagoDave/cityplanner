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
}

// Resolve runs Phase 1 analytical resolution on the spec.
// It computes demographics, service counts, density, area, and cost estimates.
// Returns resolved parameters and a validation report.
func Resolve(_ *spec.CitySpec) (*ResolvedParameters, *validation.Report) {
	report := validation.NewReport()

	// TODO: Implement Phase 1 analytical resolution
	// - Compute household counts per cohort
	// - Calculate dependency ratio
	// - Determine service counts from thresholds
	// - Calculate pod count and target populations
	// - Compute required density per ring
	// - Estimate total area, excavation volume, and cost

	return &ResolvedParameters{}, report
}
