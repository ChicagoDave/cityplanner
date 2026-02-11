package layout

import (
	"testing"

	"github.com/ChicagoDave/cityplanner/pkg/analytics"
	"github.com/ChicagoDave/cityplanner/pkg/spec"
)

func bikeTestPods(t *testing.T) ([]Pod, map[string][]string, []spec.RingDef) {
	t.Helper()
	s := defaultBikeSpec()
	params := defaultBikeParams()
	pods, adjacency, _ := LayoutPods(s, params)
	return pods, adjacency, s.CityZones.Rings
}

func defaultBikeSpec() *spec.CitySpec {
	return &spec.CitySpec{
		SpecVersion: "0.2.0",
		City: spec.CityDef{
			Population:      50000,
			ExcavationDepth: 8,
			MaxHeightCenter: 20,
			MaxHeightEdge:   4,
		},
		CityZones: spec.CityZones{
			Rings: []spec.RingDef{
				{Name: "center", Character: "civic_commercial", RadiusFrom: 0, RadiusTo: 300, MaxStories: 20},
				{Name: "middle", Character: "mixed_residential_commercial", RadiusFrom: 300, RadiusTo: 600, MaxStories: 10},
				{Name: "edge", Character: "family_education", RadiusFrom: 600, RadiusTo: 900, MaxStories: 4},
			},
		},
		Pods: spec.PodsDef{
			WalkRadius: 400,
			RingAssignments: map[string]spec.PodRing{
				"center": {Character: "civic_commercial", MaxStories: 20},
				"middle": {Character: "mixed", MaxStories: 10},
				"edge":   {Character: "residential_family", MaxStories: 4},
			},
		},
	}
}

func defaultBikeParams() *analytics.ResolvedParameters {
	return &analytics.ResolvedParameters{
		TotalPopulation: 50000,
		TotalHouseholds: 21296,
		PodCount:        6,
		TotalAreaHa:     254.47,
		Rings: []analytics.RingData{
			{Name: "center", RadiusFrom: 0, RadiusTo: 300, PodCount: 1, PodPopulation: 8332, MaxStories: 20, AvgHouseholdSize: 1.8},
			{Name: "middle", RadiusFrom: 300, RadiusTo: 600, PodCount: 2, PodPopulation: 12497, MaxStories: 10, AvgHouseholdSize: 2.5},
			{Name: "edge", RadiusFrom: 600, RadiusTo: 900, PodCount: 3, PodPopulation: 5557, MaxStories: 4, AvgHouseholdSize: 2.5},
		},
	}
}

func TestGenerateBikePathsProducesOutput(t *testing.T) {
	pods, adjacency, rings := bikeTestPods(t)
	paths, report := GenerateBikePaths(pods, adjacency, rings)

	if len(paths) == 0 {
		t.Fatal("expected bike paths to be generated")
	}
	if !report.Valid {
		t.Fatalf("report has errors: %s", report.Summary)
	}
	t.Logf("generated %d bike paths", len(paths))
}

func TestBikePathsHaveRingCorridors(t *testing.T) {
	pods, adjacency, rings := bikeTestPods(t)
	paths, _ := GenerateBikePaths(pods, adjacency, rings)

	ringCount := 0
	for _, p := range paths {
		if p.Type == "ring_corridor" {
			ringCount++
		}
	}
	if ringCount == 0 {
		t.Error("expected at least one ring corridor bike path")
	}
	t.Logf("ring corridor paths: %d", ringCount)
}

func TestBikePathsHaveRadials(t *testing.T) {
	pods, adjacency, rings := bikeTestPods(t)
	paths, _ := GenerateBikePaths(pods, adjacency, rings)

	radialCount := 0
	for _, p := range paths {
		if p.Type == "radial" {
			radialCount++
		}
	}
	if radialCount == 0 {
		t.Error("expected at least one radial bike path")
	}
	t.Logf("radial paths: %d", radialCount)
}

func TestBikePathsAreElevated(t *testing.T) {
	pods, adjacency, rings := bikeTestPods(t)
	paths, _ := GenerateBikePaths(pods, adjacency, rings)

	for _, p := range paths {
		if p.ElevatedM < 4.0 {
			t.Errorf("bike path %s has elevation %.1fm, expected >= 4m", p.ID, p.ElevatedM)
		}
	}
}

func TestBikePathsHavePoints(t *testing.T) {
	pods, adjacency, rings := bikeTestPods(t)
	paths, _ := GenerateBikePaths(pods, adjacency, rings)

	for _, p := range paths {
		if len(p.Points) < 2 {
			t.Errorf("bike path %s has only %d points", p.ID, len(p.Points))
		}
	}
}
