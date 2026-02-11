package layout

import (
	"testing"
)

func TestPlaceSportsFieldsProducesOutput(t *testing.T) {
	pods, adjacency, rings := bikeTestPods(t)
	fields, report := PlaceSportsFields(pods, adjacency, rings)

	if len(fields) == 0 {
		t.Fatal("expected sports fields to be placed")
	}
	if !report.Valid {
		t.Fatalf("report has errors: %s", report.Summary)
	}
	t.Logf("placed %d sports fields", len(fields))
}

func TestIdentifyBufferZones(t *testing.T) {
	pods, adjacency, _ := bikeTestPods(t)
	buffers := IdentifyBufferZones(pods, adjacency)

	if len(buffers) == 0 {
		t.Fatal("expected buffer zones between adjacent pods")
	}
	t.Logf("identified %d buffer zones", len(buffers))

	// Each buffer should have two pod IDs.
	for _, buf := range buffers {
		if buf.PodIDs[0] == "" || buf.PodIDs[1] == "" {
			t.Errorf("buffer %s has empty pod ID", buf.ID)
		}
	}
}

func TestSportsFieldTypes(t *testing.T) {
	pods, adjacency, rings := bikeTestPods(t)
	fields, _ := PlaceSportsFields(pods, adjacency, rings)

	types := make(map[string]int)
	for _, f := range fields {
		types[f.Type]++
	}

	t.Logf("field types: %v", types)

	// Should have at least some variety.
	if len(types) < 2 {
		t.Errorf("expected at least 2 field types, got %d", len(types))
	}
}

func TestSportsFieldDimensions(t *testing.T) {
	pods, adjacency, rings := bikeTestPods(t)
	fields, _ := PlaceSportsFields(pods, adjacency, rings)

	for _, f := range fields {
		if f.Dimensions[0] <= 0 || f.Dimensions[1] <= 0 {
			t.Errorf("field %s has invalid dimensions: %v", f.ID, f.Dimensions)
		}
	}
}
