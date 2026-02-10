package cost

import (
	"github.com/ChicagoDave/cityplanner/pkg/analytics"
	"github.com/ChicagoDave/cityplanner/pkg/layout"
	"github.com/ChicagoDave/cityplanner/pkg/routing"
	"github.com/ChicagoDave/cityplanner/pkg/spec"
)

// Breakdown itemizes costs by category.
type Breakdown struct {
	Excavation     float64 `json:"excavation"`
	Structural     float64 `json:"structural"`
	Buildings      float64 `json:"buildings"`
	Infrastructure float64 `json:"infrastructure"`
	Solar          float64 `json:"solar"`
	Battery        float64 `json:"battery"`
	Other          float64 `json:"other"`
	Total          float64 `json:"total"`
}

// PhasedCost separates costs by construction phase.
type PhasedCost struct {
	Phase1           Breakdown `json:"phase_1"`
	Phase2           Breakdown `json:"phase_2"`
	Phase3           Breakdown `json:"phase_3"`
	PerimeterAndSolar Breakdown `json:"perimeter_and_solar"`
	Total            Breakdown `json:"total"`
}

// Report is the complete cost output.
type Report struct {
	Estimate *PhasedCost `json:"estimate"`
	Actual   *PhasedCost `json:"actual,omitempty"`

	Summary struct {
		TotalConstruction     float64 `json:"total_construction"`
		PerCapita             float64 `json:"per_capita"`
		AnnualDebtService     float64 `json:"annual_debt_service"`
		AnnualOperations      float64 `json:"annual_operations"`
		BreakEvenMonthlyRent  float64 `json:"break_even_monthly_rent"`
	} `json:"summary"`
}

// Estimate computes Phase 1 aggregate cost estimate from analytical parameters.
func Estimate(_ *spec.CitySpec, _ *analytics.ResolvedParameters) *Report {
	// TODO: Implement Phase 1 cost model (ADR-010)
	return &Report{}
}

// Compute computes Phase 2 precise bottom-up cost from generated geometry.
func Compute(_ *spec.CitySpec, _ []layout.Building, _ []routing.Segment) *Report {
	// TODO: Implement Phase 2 cost model (ADR-010)
	return &Report{}
}
