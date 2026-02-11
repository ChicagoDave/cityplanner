package layout

import (
	"testing"
)

func TestGenerateShuttleRoutesProducesOutput(t *testing.T) {
	pods, adjacency, rings := bikeTestPods(t)
	bikePaths, _ := GenerateBikePaths(pods, adjacency, rings)
	routes, stations, report := GenerateShuttleRoutes(bikePaths, pods)

	if len(routes) == 0 {
		t.Fatal("expected shuttle routes to be generated")
	}
	if len(stations) == 0 {
		t.Fatal("expected stations to be generated")
	}
	if !report.Valid {
		t.Fatalf("report has errors: %s", report.Summary)
	}
	t.Logf("generated %d routes and %d stations", len(routes), len(stations))
}

func TestOneStationPerPod(t *testing.T) {
	pods, adjacency, rings := bikeTestPods(t)
	bikePaths, _ := GenerateBikePaths(pods, adjacency, rings)
	_, stations, _ := GenerateShuttleRoutes(bikePaths, pods)

	podStations := make(map[string]int)
	for _, st := range stations {
		podStations[st.PodID]++
	}

	for _, pod := range pods {
		if podStations[pod.ID] != 1 {
			t.Errorf("pod %s has %d stations, expected 1", pod.ID, podStations[pod.ID])
		}
	}
}

func TestShuttleRoutesParallelBikePaths(t *testing.T) {
	pods, adjacency, rings := bikeTestPods(t)
	bikePaths, _ := GenerateBikePaths(pods, adjacency, rings)
	routes, _, _ := GenerateShuttleRoutes(bikePaths, pods)

	if len(routes) != len(bikePaths) {
		t.Errorf("expected %d shuttle routes (one per bike path), got %d", len(bikePaths), len(routes))
	}
}

func TestStationsHaveRouteIDs(t *testing.T) {
	pods, adjacency, rings := bikeTestPods(t)
	bikePaths, _ := GenerateBikePaths(pods, adjacency, rings)
	_, stations, _ := GenerateShuttleRoutes(bikePaths, pods)

	for _, st := range stations {
		if st.RouteID == "" {
			t.Errorf("station %s has empty route ID", st.ID)
		}
	}
}
