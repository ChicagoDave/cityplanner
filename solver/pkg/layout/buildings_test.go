package layout

import (
	"testing"
)

func TestPlaceBuildingsProducesResults(t *testing.T) {
	s := defaultSpec()
	params := defaultParams()
	pods, adjacency, _ := LayoutPods(s, params)
	buildings, paths, report := PlaceBuildings(s, pods, adjacency, params)

	if len(buildings) == 0 {
		t.Fatal("expected buildings to be placed")
	}
	if len(paths) == 0 {
		t.Fatal("expected paths to be generated")
	}
	if !report.Valid {
		t.Errorf("placement report invalid: %v", report.Errors)
	}
	t.Logf("placed %d buildings and %d paths", len(buildings), len(paths))
}

func TestPlaceBuildingsHasDwellingUnits(t *testing.T) {
	s := defaultSpec()
	params := defaultParams()
	pods, adjacency, _ := LayoutPods(s, params)
	buildings, _, _ := PlaceBuildings(s, pods, adjacency, params)

	totalDU := 0
	for _, b := range buildings {
		totalDU += b.DwellingUnits
	}
	if totalDU == 0 {
		t.Fatal("no dwelling units placed")
	}
	t.Logf("total dwelling units: %d (target: %d)", totalDU, params.TotalHouseholds)
}

func TestPlaceBuildingsHasBuildingTypes(t *testing.T) {
	s := defaultSpec()
	params := defaultParams()
	pods, adjacency, _ := LayoutPods(s, params)
	buildings, _, _ := PlaceBuildings(s, pods, adjacency, params)

	types := map[string]int{}
	for _, b := range buildings {
		types[b.Type]++
	}
	if types["residential"] == 0 {
		t.Error("no residential buildings placed")
	}
	if types["commercial"] == 0 {
		t.Error("no commercial buildings placed")
	}
	if types["civic"] == 0 {
		t.Error("no civic buildings placed")
	}
	t.Logf("building types: %v", types)
}

func TestPlaceBuildingsHeightEnvelope(t *testing.T) {
	s := defaultSpec()
	params := defaultParams()
	pods, adjacency, _ := LayoutPods(s, params)
	buildings, _, _ := PlaceBuildings(s, pods, adjacency, params)

	for _, b := range buildings {
		if b.Stories > s.City.MaxHeightCenter {
			t.Errorf("building %s has %d stories, exceeds max %d",
				b.ID, b.Stories, s.City.MaxHeightCenter)
		}
		if b.Stories < 1 {
			t.Errorf("building %s has %d stories, must be >= 1", b.ID, b.Stories)
		}
	}
}

func TestPlaceBuildingsAllPodsHaveBuildings(t *testing.T) {
	s := defaultSpec()
	params := defaultParams()
	pods, adjacency, _ := LayoutPods(s, params)
	buildings, _, _ := PlaceBuildings(s, pods, adjacency, params)

	podBuildings := map[string]int{}
	for _, b := range buildings {
		podBuildings[b.PodID]++
	}
	for _, p := range pods {
		if podBuildings[p.ID] == 0 {
			t.Errorf("pod %s has no buildings", p.ID)
		}
	}
}

func TestEnvelopeMaxStories(t *testing.T) {
	tests := []struct {
		dist     float64
		expected int
	}{
		{0, 20},
		{150, 20},
		{300, 20},
		{450, 15},
		{600, 10},
		{750, 7},
		{900, 4},
	}
	for _, tt := range tests {
		got := MaxStories(tt.dist, 20, 10, 4)
		if got != tt.expected {
			t.Errorf("MaxStories(%f) = %d, want %d", tt.dist, got, tt.expected)
		}
	}
}

func TestDistributeUnits(t *testing.T) {
	params := defaultParams()
	mix := DistributeUnits(params.TotalHouseholds, params.Cohorts)
	if mix.Total() != params.TotalHouseholds {
		t.Errorf("unit mix total %d != target %d", mix.Total(), params.TotalHouseholds)
	}
}
