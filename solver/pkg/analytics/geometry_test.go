package analytics

import (
	"math"
	"testing"

	"github.com/ChicagoDave/cityplanner/pkg/spec"
)

func defaultSpec() *spec.CitySpec {
	return &spec.CitySpec{
		City: spec.CityDef{
			Population:      50000,
			ExcavationDepth: 8,
			MaxHeightCenter: 20,
			MaxHeightEdge:   4,
		},
		CityZones: spec.CityZones{
			Center:    spec.ZoneDef{RadiusFrom: 0, RadiusTo: 300, MaxStories: 20},
			Middle:    spec.ZoneDef{RadiusFrom: 300, RadiusTo: 600, MaxStories: 10},
			Edge:      spec.ZoneDef{RadiusFrom: 600, RadiusTo: 900, MaxStories: 4},
			Perimeter: spec.PerimeterDef{RadiusFrom: 900, RadiusTo: 1100},
			SolarRing: spec.SolarRingDef{RadiusFrom: 1100, RadiusTo: 1500, AreaHa: 250},
		},
		Pods: spec.PodsDef{WalkRadius: 400},
		Infrastructure: spec.Infrastructure{
			Electrical: spec.ElectricalInfra{
				SolarIntegratedAvgMW: 80,
				SolarFarmAvgMW:       100,
				BatteryCapacityMWh:   3000,
				GridCapacityMW:       150,
				PeakDemandKWPer:      2.5,
			},
		},
	}
}

func TestResolveRingsAreas(t *testing.T) {
	s := defaultSpec()
	rings := resolveRings(s, 20000, 50000)

	if len(rings) != 3 {
		t.Fatalf("expected 3 rings, got %d", len(rings))
	}

	// Center: pi * 300^2 = 282743 m^2 = 28.27 ha
	center := rings[0]
	expectedCenterHa := math.Pi * 300 * 300 / 10000
	if math.Abs(center.AreaHa-expectedCenterHa) > 0.1 {
		t.Errorf("center area = %.2f ha, want %.2f", center.AreaHa, expectedCenterHa)
	}

	// Middle: pi * (600^2 - 300^2) = 848230 m^2 = 84.82 ha
	middle := rings[1]
	expectedMiddleHa := math.Pi * (600*600 - 300*300) / 10000
	if math.Abs(middle.AreaHa-expectedMiddleHa) > 0.1 {
		t.Errorf("middle area = %.2f ha, want %.2f", middle.AreaHa, expectedMiddleHa)
	}

	// Edge: pi * (900^2 - 600^2)
	edge := rings[2]
	expectedEdgeHa := math.Pi * (900*900 - 600*600) / 10000
	if math.Abs(edge.AreaHa-expectedEdgeHa) > 0.1 {
		t.Errorf("edge area = %.2f ha, want %.2f", edge.AreaHa, expectedEdgeHa)
	}

	// Area fractions should sum to 1.0
	fracSum := center.AreaFraction + middle.AreaFraction + edge.AreaFraction
	if math.Abs(fracSum-1.0) > 0.01 {
		t.Errorf("area fractions sum = %.4f, want 1.0", fracSum)
	}
}

func TestResolveRingsPodCounts(t *testing.T) {
	s := defaultSpec()
	rings := resolveRings(s, 20000, 50000)

	// Pod area = pi * 400^2 = 502655 m^2 = 50.27 ha
	// Center pods: ceil(28.27 / 50.27) = 1
	if rings[0].PodCount != 1 {
		t.Errorf("center pods = %d, want 1", rings[0].PodCount)
	}
	// Middle pods: ceil(84.82 / 50.27) = 2
	if rings[1].PodCount != 2 {
		t.Errorf("middle pods = %d, want 2", rings[1].PodCount)
	}
	// Edge pods: ceil(141.37 / 50.27) = 3
	if rings[2].PodCount != 3 {
		t.Errorf("edge pods = %d, want 3", rings[2].PodCount)
	}
}

func TestResolveRingsDensity(t *testing.T) {
	s := defaultSpec()
	rings := resolveRings(s, 20000, 50000)

	for _, ring := range rings {
		if ring.RequiredDensity <= 0 {
			t.Errorf("%s ring required density = %.2f, want > 0", ring.Name, ring.RequiredDensity)
		}
		if ring.AchievableDensity <= 0 {
			t.Errorf("%s ring achievable density = %.2f, want > 0", ring.Name, ring.AchievableDensity)
		}
		// For the default city, density should be feasible in all rings
		if ring.RequiredDensity > ring.AchievableDensity {
			t.Errorf("%s ring: required density %.0f > achievable %.0f",
				ring.Name, ring.RequiredDensity, ring.AchievableDensity)
		}
	}
}

func TestResolveAreas(t *testing.T) {
	s := defaultSpec()
	areas := resolveAreas(s)

	// Total city area: pi * 900^2 / 10000 = 254.47 ha
	expectedTotal := math.Pi * 900 * 900 / 10000
	if math.Abs(areas.TotalCityHa-expectedTotal) > 0.1 {
		t.Errorf("total city area = %.2f ha, want %.2f", areas.TotalCityHa, expectedTotal)
	}

	// Land use fractions should approximately sum to city area
	usageSum := areas.ResidentialHa + areas.CommercialHa + areas.CivicHa + areas.GreenPathsHa
	if math.Abs(usageSum-areas.TotalCityHa) > 0.1 {
		t.Errorf("land use sum = %.2f, total city = %.2f", usageSum, areas.TotalCityHa)
	}

	// Solar area should be 250 ha (from spec)
	if areas.SolarHa != 250 {
		t.Errorf("solar area = %.2f ha, want 250", areas.SolarHa)
	}

	// Total with perimeter should be greater than city area
	if areas.TotalWithPerimeter <= areas.TotalCityHa {
		t.Error("total with perimeter should exceed city area")
	}
}

func TestResolveEnergy(t *testing.T) {
	s := defaultSpec()
	e := resolveEnergy(s)

	// Peak demand: 50000 * 2.5 / 1000 = 125 MW
	if math.Abs(e.PeakDemandMW-125.0) > 0.1 {
		t.Errorf("peak demand = %.1f MW, want 125.0", e.PeakDemandMW)
	}

	// Total generation: 80 + 100 = 180 MW
	if math.Abs(e.TotalGenerationMW-180.0) > 0.1 {
		t.Errorf("total generation = %.1f MW, want 180.0", e.TotalGenerationMW)
	}

	// Backup hours: 3000 / 125 = 24
	if math.Abs(e.BackupHours-24.0) > 0.1 {
		t.Errorf("backup hours = %.1f, want 24.0", e.BackupHours)
	}
}
