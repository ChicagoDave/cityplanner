package scene

import (
	"math"
	"testing"

	"github.com/ChicagoDave/cityplanner/pkg/analytics"
	"github.com/ChicagoDave/cityplanner/pkg/layout"
	"github.com/ChicagoDave/cityplanner/pkg/routing"
	"github.com/ChicagoDave/cityplanner/pkg/spec"
)

// specForPopulation creates a valid CitySpec scaled to the given population.
// Radii scale as sqrt(population/50000) relative to the default 50K city.
func specForPopulation(pop int) *spec.CitySpec {
	scale := math.Sqrt(float64(pop) / 50000.0)
	return &spec.CitySpec{
		SpecVersion: "0.1.0",
		City: spec.CityDef{
			Population:      pop,
			ExcavationDepth: 8,
			MaxHeightCenter: 20,
			MaxHeightEdge:   4,
		},
		Demographics: spec.Demographics{
			Singles:      0.15,
			Couples:      0.20,
			FamiliesYoung: 0.25,
			FamiliesTeen: 0.15,
			EmptyNest:    0.15,
			Retirees:     0.10,
		},
		CityZones: spec.CityZones{
			Center: spec.ZoneDef{RadiusFrom: 0, RadiusTo: 300 * scale, MaxStories: 20, Character: "civic_commercial"},
			Middle: spec.ZoneDef{RadiusFrom: 300 * scale, RadiusTo: 600 * scale, MaxStories: 10, Character: "mixed_residential_commercial"},
			Edge:   spec.ZoneDef{RadiusFrom: 600 * scale, RadiusTo: 900 * scale, MaxStories: 4, Character: "family_education"},
		},
		Pods: spec.PodsDef{
			WalkRadius: 400,
			RingAssignments: map[string]spec.PodRing{
				"center": {Character: "civic_commercial", MaxStories: 20, RequiredServices: []string{
					"hospital", "performing_arts", "city_hall", "coworking_hub",
				}},
				"middle": {Character: "mixed", MaxStories: 10, RequiredServices: []string{
					"secondary_school", "coworking", "medical_clinic", "retail", "restaurant",
				}},
				"edge": {Character: "residential_family", MaxStories: 4, RequiredServices: []string{
					"elementary_school", "library", "grocery", "playground", "pediatric_clinic", "daycare",
				}},
			},
		},
		Infrastructure: spec.Infrastructure{
			Electrical: spec.ElectricalInfra{
				SolarIntegratedAvgMW: 80 * scale * scale,
				SolarFarmAvgMW:       100 * scale * scale,
				BatteryCapacityMWh:   3000 * scale * scale,
				GridCapacityMW:       150 * scale * scale,
				PeakDemandKWPer:      2.5,
			},
		},
		Vehicles: spec.Vehicles{ArterialWidthM: 6, ServiceBranchWidthM: 4, TotalFleet: pop / 250},
		Revenue: spec.Revenue{
			DebtTermYears:  30,
			InterestRate:   0.05,
			AnnualOpsCostM: 100 * scale * scale,
		},
	}
}

func runFullPipeline(t testing.TB, pop int) *Graph {
	t.Helper()
	s := specForPopulation(pop)
	params, report := analytics.Resolve(s)
	if !report.Valid {
		t.Fatalf("analytics validation failed for %d pop: %s", pop, report.Summary)
	}

	pods, adjacency, layoutReport := layout.LayoutPods(s, params)
	if !layoutReport.Valid {
		t.Fatalf("layout validation failed for %d pop: %s", pop, layoutReport.Summary)
	}

	buildings, paths, buildReport := layout.PlaceBuildings(s, pods, adjacency, params)
	if !buildReport.Valid {
		t.Fatalf("building placement failed for %d pop: %s", pop, buildReport.Summary)
	}

	segments, routeReport := routing.RouteInfrastructure(s, pods, buildings)
	if !routeReport.Valid {
		t.Fatalf("routing failed for %d pop: %s", pop, routeReport.Summary)
	}

	greenZones := layout.CollectGreenZones(s, pods)
	return Assemble(s, pods, buildings, paths, segments, greenZones)
}

func TestLargeCity100K(t *testing.T) {
	g := runFullPipeline(t, 100000)
	if len(g.Entities) == 0 {
		t.Fatal("expected entities for 100K city")
	}
	t.Logf("100K city: %d entities, %d pods", len(g.Entities), len(g.Groups.Pods))

	for et, ids := range g.Groups.EntityTypes {
		t.Logf("  %s: %d", et, len(ids))
	}
}

func BenchmarkFullPipeline50K(b *testing.B) {
	for b.Loop() {
		runFullPipeline(b, 50000)
	}
}

func BenchmarkFullPipeline100K(b *testing.B) {
	for b.Loop() {
		runFullPipeline(b, 100000)
	}
}

func BenchmarkFullPipeline250K(b *testing.B) {
	for b.Loop() {
		runFullPipeline(b, 250000)
	}
}
