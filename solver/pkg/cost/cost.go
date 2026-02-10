package cost

import (
	"math"

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
	Phase1            Breakdown `json:"phase_1"`
	Phase2            Breakdown `json:"phase_2"`
	Phase3            Breakdown `json:"phase_3"`
	PerimeterAndSolar Breakdown `json:"perimeter_and_solar"`
	Total             Breakdown `json:"total"`
}

// Report is the complete cost output.
type Report struct {
	Estimate *PhasedCost `json:"estimate"`
	Actual   *PhasedCost `json:"actual,omitempty"`

	Summary struct {
		TotalConstruction    float64 `json:"total_construction"`
		PerCapita            float64 `json:"per_capita"`
		AnnualDebtService    float64 `json:"annual_debt_service"`
		AnnualOperations     float64 `json:"annual_operations"`
		BreakEvenMonthlyRent float64 `json:"break_even_monthly_rent"`
	} `json:"summary"`
}

// Estimate computes Phase 1 aggregate cost estimate from analytical parameters.
func Estimate(s *spec.CitySpec, p *analytics.ResolvedParameters) *Report {
	report := &Report{}

	totalCityAreaM2 := p.Areas.TotalCityHa * M2PerHa

	// Construction phase area fractions (ADR-010)
	// Phase 1: 0-450m, Phase 2: 450-600m, Phase 3: 600-900m
	phase1AreaM2 := math.Pi * 450.0 * 450.0
	phase2AreaM2 := math.Pi * (600.0*600.0 - 450.0*450.0)
	phase3AreaM2 := math.Pi * (900.0*900.0 - 600.0*600.0)

	phase1Frac := phase1AreaM2 / totalCityAreaM2
	phase2Frac := phase2AreaM2 / totalCityAreaM2
	phase3Frac := phase3AreaM2 / totalCityAreaM2

	// Total costs
	excavation := p.ExcavationVolumeM3 * ExcavationCostPerM3
	structural := totalCityAreaM2 * float64(UndergroundLevels) * SlabCostPerM2

	// Building costs
	residentialFloorArea := float64(p.TotalHouseholds) * AvgUnitSizeM2
	commercialFloorArea := p.Areas.CommercialHa * M2PerHa * GroundCoverageRatio * AvgCommercialStories
	civicFloorArea := p.Areas.CivicHa * M2PerHa * GroundCoverageRatio * AvgCivicStories
	buildings := residentialFloorArea*ResidentialCostPerM2 +
		commercialFloorArea*CommercialCostPerM2 +
		civicFloorArea*CivicCostPerM2

	// Infrastructure: estimate network length
	edgeRadius := s.CityZones.OuterRadius()
	walkRadius := s.Pods.WalkRadius
	networkLenPerSystem := 2*edgeRadius + float64(p.PodCount)*2*walkRadius
	infrastructure := networkLenPerSystem * (InfraWaterCostPerM + InfraSewageCostPerM +
		InfraElectricalCostPerM + InfraTelecomCostPerM + InfraVehicleCostPerM)

	// Solar
	solarAreaM2 := p.Areas.SolarHa * M2PerHa
	solar := solarAreaM2 * SolarCostPerM2

	// Battery
	battery := s.Infrastructure.Electrical.BatteryCapacityMWh * BatteryCostPerMWh

	totalConstruction := excavation + structural + buildings + infrastructure + solar + battery

	// Phase breakdown
	report.Estimate = &PhasedCost{
		Phase1:            makeBreakdown(excavation*phase1Frac, structural*phase1Frac, buildings*phase1Frac, infrastructure*phase1Frac, 0, 0, 0),
		Phase2:            makeBreakdown(excavation*phase2Frac, structural*phase2Frac, buildings*phase2Frac, infrastructure*phase2Frac, 0, 0, 0),
		Phase3:            makeBreakdown(excavation*phase3Frac, structural*phase3Frac, buildings*phase3Frac, infrastructure*phase3Frac, 0, 0, 0),
		PerimeterAndSolar: makeBreakdown(0, 0, 0, 0, solar, battery, 0),
		Total:             makeBreakdown(excavation, structural, buildings, infrastructure, solar, battery, 0),
	}

	// Summary financials
	annualOps := s.Revenue.AnnualOpsCostM * 1_000_000.0
	annualDebt := computeAnnualDebtService(totalConstruction, s.Revenue.InterestRate, s.Revenue.DebtTermYears)

	breakEvenRent := 0.0
	if p.TotalHouseholds > 0 {
		breakEvenRent = (annualDebt + annualOps) / float64(p.TotalHouseholds) / 12.0
	}

	report.Summary.TotalConstruction = totalConstruction
	if s.City.Population > 0 {
		report.Summary.PerCapita = totalConstruction / float64(s.City.Population)
	}
	report.Summary.AnnualDebtService = annualDebt
	report.Summary.AnnualOperations = annualOps
	report.Summary.BreakEvenMonthlyRent = breakEvenRent

	return report
}

// computeAnnualDebtService uses the standard annuity formula.
// P * r(1+r)^n / ((1+r)^n - 1)
// At 0% interest, returns principal / term.
func computeAnnualDebtService(principal, rate float64, termYears int) float64 {
	if termYears <= 0 {
		return 0
	}
	if rate <= 0 {
		return principal / float64(termYears)
	}
	n := float64(termYears)
	factor := math.Pow(1+rate, n)
	return principal * rate * factor / (factor - 1)
}

func makeBreakdown(excavation, structural, buildings, infrastructure, solar, battery, other float64) Breakdown {
	return Breakdown{
		Excavation:     excavation,
		Structural:     structural,
		Buildings:      buildings,
		Infrastructure: infrastructure,
		Solar:          solar,
		Battery:        battery,
		Other:          other,
		Total:          excavation + structural + buildings + infrastructure + solar + battery + other,
	}
}

// Compute computes Phase 2 precise bottom-up cost from generated geometry.
func Compute(_ *spec.CitySpec, _ []layout.Building, _ []routing.Segment) *Report {
	// TODO: Implement Phase 2 cost model (ADR-010)
	return &Report{}
}
