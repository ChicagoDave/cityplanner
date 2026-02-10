package analytics

import (
	"math"
	"testing"

	"github.com/ChicagoDave/cityplanner/pkg/spec"
)

func fullDefaultSpec() *spec.CitySpec {
	return &spec.CitySpec{
		SpecVersion: "0.1.0",
		City: spec.CityDef{
			Population:      50000,
			FootprintShape:  "circle",
			ExcavationDepth: 8,
			HeightProfile:   "bowl",
			MaxHeightCenter: 20,
			MaxHeightEdge:   4,
		},
		CityZones: spec.CityZones{
			Rings: []spec.RingDef{
				{Name: "center", Character: "civic_commercial", RadiusFrom: 0, RadiusTo: 300, MaxStories: 20},
				{Name: "middle", Character: "mixed", RadiusFrom: 300, RadiusTo: 600, MaxStories: 10},
				{Name: "edge", Character: "family", RadiusFrom: 600, RadiusTo: 900, MaxStories: 4},
			},
			Perimeter: spec.PerimeterDef{RadiusFrom: 900, RadiusTo: 1100},
			SolarRing: spec.SolarRingDef{RadiusFrom: 1100, RadiusTo: 1500, AreaHa: 250, CapacityMW: 500, AvgOutputMW: 100},
		},
		Pods: spec.PodsDef{
			WalkRadius: 400,
			RingAssignments: map[string]spec.PodRing{
				"center": {Character: "civic", RequiredServices: []string{"hospital", "performing_arts", "city_hall", "coworking_hub"}, MaxStories: 20},
				"middle": {Character: "mixed", RequiredServices: []string{"secondary_school", "coworking", "medical_clinic", "retail", "restaurant"}, MaxStories: 10},
				"edge":   {Character: "family", RequiredServices: []string{"elementary_school", "library", "grocery", "playground", "pediatric_clinic", "daycare"}, MaxStories: 4},
			},
		},
		Demographics: spec.Demographics{
			Singles: 0.15, Couples: 0.20, FamiliesYoung: 0.25,
			FamiliesTeen: 0.15, EmptyNest: 0.15, Retirees: 0.10,
		},
		Infrastructure: spec.Infrastructure{
			Water:  spec.WaterInfra{Source: "municipal_connection", CapacityGPDPer: 100},
			Sewage: spec.SewageInfra{Collection: "gravity_flow", CapacityGPDPer: 95},
			Electrical: spec.ElectricalInfra{
				SolarIntegratedAvgMW: 80, SolarFarmAvgMW: 100,
				BatteryCapacityMWh: 3000, GridCapacityMW: 150,
				PeakDemandKWPer: 2.5,
			},
			Telecom: spec.TelecomInfra{NodeSpacingM: 75},
		},
		Revenue: spec.Revenue{DebtTermYears: 30, InterestRate: 0.05, AnnualOpsCostM: 100},
		Site:    spec.SiteRequirements{MinAreaHa: 800, SolarIrradiance: 4.5},
	}
}

func TestResolveDefaultCity(t *testing.T) {
	s := fullDefaultSpec()
	params, report := Resolve(s)

	// Weighted average: 2.475
	if math.Abs(params.WeightedAvgHH-2.475) > 0.01 {
		t.Errorf("WeightedAvgHH = %v, want ~2.475", params.WeightedAvgHH)
	}

	// Total households: derived from per-ring avg household sizes.
	// With "civic_commercial" (1.8), "mixed" (default 2.5), "family" (default 2.5),
	// the city-wide effective avg is ~2.5, giving ~20,000 HH.
	if params.TotalHouseholds < 18000 || params.TotalHouseholds > 22000 {
		t.Errorf("TotalHouseholds = %d, want 18000-22000", params.TotalHouseholds)
	}

	// Population preserved
	if params.TotalPopulation != 50000 {
		t.Errorf("TotalPopulation = %d, want 50000", params.TotalPopulation)
	}

	// Pod count: 1+2+3 = 6
	if params.PodCount != 6 {
		t.Errorf("PodCount = %d, want 6", params.PodCount)
	}

	// Total city area: ~254 ha
	expectedArea := math.Pi * 900 * 900 / 10000
	if math.Abs(params.TotalAreaHa-expectedArea) > 1 {
		t.Errorf("TotalAreaHa = %.1f, want ~%.1f", params.TotalAreaHa, expectedArea)
	}

	// Excavation volume: ~20.36M m^3
	expectedExcav := expectedArea * 10000 * 8
	if math.Abs(params.ExcavationVolumeM3-expectedExcav) > 10000 {
		t.Errorf("ExcavationVolumeM3 = %.0f, want ~%.0f", params.ExcavationVolumeM3, expectedExcav)
	}

	// Energy peak demand: 125 MW
	if math.Abs(params.Energy.PeakDemandMW-125.0) > 0.1 {
		t.Errorf("PeakDemandMW = %.1f, want 125.0", params.Energy.PeakDemandMW)
	}

	// Should have cohorts
	if len(params.Cohorts) != 6 {
		t.Errorf("expected 6 cohorts, got %d", len(params.Cohorts))
	}

	// Should have rings
	if len(params.Rings) != 3 {
		t.Errorf("expected 3 rings, got %d", len(params.Rings))
	}

	// Should have services
	if len(params.Services) == 0 {
		t.Error("expected services to be populated")
	}

	// Dependency ratio should be reasonable
	if params.DependencyRatio < 0.2 || params.DependencyRatio > 1.0 {
		t.Errorf("DependencyRatio = %v, expected reasonable range", params.DependencyRatio)
	}

	// Report should be valid (no errors for default city)
	if !report.Valid {
		for _, e := range report.Errors {
			t.Logf("error: %s", e.Message)
		}
		t.Error("expected valid report for default city")
	}
}

func TestResolveInfeasibleDensity(t *testing.T) {
	s := fullDefaultSpec()
	s.City.Population = 500000 // 10x population, same area
	_, report := Resolve(s)

	// Should produce density errors with the new spec path format
	hasError := false
	for _, e := range report.Errors {
		if e.SpecPath == "city_zones.rings[0].max_stories" ||
			e.SpecPath == "city_zones.rings[1].max_stories" ||
			e.SpecPath == "city_zones.rings[2].max_stories" {
			hasError = true
			break
		}
	}
	if !hasError {
		t.Error("expected density feasibility error for 500K population")
	}
}

func TestResolveInsufficientBattery(t *testing.T) {
	s := fullDefaultSpec()
	s.Infrastructure.Electrical.BatteryCapacityMWh = 1000 // Not enough for 24h
	_, report := Resolve(s)

	hasWarning := false
	for _, w := range report.Warnings {
		if w.SpecPath == "infrastructure.electrical.battery_capacity_mwh" {
			hasWarning = true
			break
		}
	}
	if !hasWarning {
		t.Error("expected battery backup warning")
	}
}

func TestResolveInsufficientEnergy(t *testing.T) {
	s := fullDefaultSpec()
	s.Infrastructure.Electrical.SolarIntegratedAvgMW = 0
	s.Infrastructure.Electrical.SolarFarmAvgMW = 0
	s.Infrastructure.Electrical.GridCapacityMW = 50 // Only 50 MW vs 125 MW demand
	_, report := Resolve(s)

	if report.Valid {
		t.Error("expected invalid report for insufficient energy")
	}
}
