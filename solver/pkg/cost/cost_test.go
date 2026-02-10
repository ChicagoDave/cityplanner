package cost

import (
	"math"
	"testing"

	"github.com/ChicagoDave/cityplanner/pkg/analytics"
	"github.com/ChicagoDave/cityplanner/pkg/spec"
)

func defaultCostSpec() *spec.CitySpec {
	return &spec.CitySpec{
		City: spec.CityDef{
			Population:      50000,
			ExcavationDepth: 8,
		},
		CityZones: spec.CityZones{
			Rings: []spec.RingDef{
				{Name: "center", RadiusFrom: 0, RadiusTo: 300, MaxStories: 20},
				{Name: "middle", RadiusFrom: 300, RadiusTo: 600, MaxStories: 10},
				{Name: "edge", RadiusFrom: 600, RadiusTo: 900, MaxStories: 4},
			},
			SolarRing: spec.SolarRingDef{AreaHa: 250},
		},
		Pods: spec.PodsDef{WalkRadius: 400},
		Infrastructure: spec.Infrastructure{
			Electrical: spec.ElectricalInfra{BatteryCapacityMWh: 3000},
		},
		Revenue: spec.Revenue{
			DebtTermYears:  30,
			InterestRate:   0.05,
			AnnualOpsCostM: 100,
		},
	}
}

func defaultParams() *analytics.ResolvedParameters {
	return &analytics.ResolvedParameters{
		TotalHouseholds:    20202,
		TotalPopulation:    50000,
		PodCount:           6,
		ExcavationVolumeM3: 20357520,
		Areas: analytics.AreaBreakdown{
			TotalCityHa:   254.47,
			ResidentialHa:  152.68,
			CommercialHa:   38.17,
			CivicHa:        25.45,
			GreenPathsHa:   38.17,
			SolarHa:        250.0,
		},
	}
}

func TestEstimateDefaultCity(t *testing.T) {
	s := defaultCostSpec()
	p := defaultParams()
	report := Estimate(s, p)

	if report.Estimate == nil {
		t.Fatal("expected non-nil estimate")
	}

	// Total construction should be in billions range
	total := report.Summary.TotalConstruction
	if total < 1_000_000_000 || total > 20_000_000_000 {
		t.Errorf("total construction = $%.0f, expected in $1B-$20B range", total)
	}

	// Per capita should be in $100K-$300K range
	perCapita := report.Summary.PerCapita
	if perCapita < 50_000 || perCapita > 500_000 {
		t.Errorf("per capita = $%.0f, expected $50K-$500K range", perCapita)
	}

	// Break-even rent should be positive
	if report.Summary.BreakEvenMonthlyRent <= 0 {
		t.Errorf("break-even rent = $%.0f, expected > 0", report.Summary.BreakEvenMonthlyRent)
	}

	// Annual debt service should be positive
	if report.Summary.AnnualDebtService <= 0 {
		t.Error("expected positive annual debt service")
	}

	// Annual operations = 100M
	if math.Abs(report.Summary.AnnualOperations-100_000_000) > 1 {
		t.Errorf("annual ops = $%.0f, want $100,000,000", report.Summary.AnnualOperations)
	}
}

func TestEstimatePhaseBreakdown(t *testing.T) {
	s := defaultCostSpec()
	p := defaultParams()
	report := Estimate(s, p)

	est := report.Estimate

	// Phase totals (excluding perimeter) should roughly sum to total minus solar/battery
	constructionPhases := est.Phase1.Total + est.Phase2.Total + est.Phase3.Total
	totalMinusSolarBattery := est.Total.Total - est.PerimeterAndSolar.Total

	// Allow 1% tolerance for floating point
	if math.Abs(constructionPhases-totalMinusSolarBattery)/totalMinusSolarBattery > 0.01 {
		t.Errorf("phase sum = $%.0f, total-solar-battery = $%.0f", constructionPhases, totalMinusSolarBattery)
	}

	// Phase 1 should be smaller than Phase 3 (larger area)
	if est.Phase1.Total >= est.Phase3.Total {
		t.Errorf("phase1 ($%.0f) should be < phase3 ($%.0f)", est.Phase1.Total, est.Phase3.Total)
	}

	// Perimeter should be solar + battery only
	if est.PerimeterAndSolar.Excavation != 0 {
		t.Error("perimeter should have zero excavation")
	}
	if est.PerimeterAndSolar.Solar <= 0 {
		t.Error("perimeter should have positive solar cost")
	}
	if est.PerimeterAndSolar.Battery <= 0 {
		t.Error("perimeter should have positive battery cost")
	}
}

func TestAnnuityFormula(t *testing.T) {
	// $1M at 5% for 30 years
	annual := computeAnnualDebtService(1_000_000, 0.05, 30)
	// Expected: ~$65,051 (standard amortization)
	if math.Abs(annual-65051) > 100 {
		t.Errorf("annuity = $%.0f, want ~$65,051", annual)
	}
}

func TestAnnuityZeroRate(t *testing.T) {
	annual := computeAnnualDebtService(1_000_000, 0, 30)
	expected := 1_000_000.0 / 30.0
	if math.Abs(annual-expected) > 1 {
		t.Errorf("annuity at 0%% = $%.0f, want $%.0f", annual, expected)
	}
}

func TestAnnuityZeroTerm(t *testing.T) {
	annual := computeAnnualDebtService(1_000_000, 0.05, 0)
	if annual != 0 {
		t.Errorf("annuity at 0 term = $%.0f, want $0", annual)
	}
}
