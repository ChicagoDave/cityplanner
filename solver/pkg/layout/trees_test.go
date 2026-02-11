package layout

import "testing"

func TestPlaceTreesProducesOutput(t *testing.T) {
	pods, adjacency, rings := bikeTestPods(t)
	s := defaultBikeSpec()

	greenZones := CollectGreenZones(s, pods)
	_, paths, _ := PlaceBuildings(s, pods, adjacency, defaultBikeParams())
	bikePaths, _ := GenerateBikePaths(pods, adjacency, rings)
	plazas, _ := GeneratePlazas(pods, s)

	trees, report := PlaceTrees(pods, greenZones, paths, bikePaths, plazas)

	if len(trees) == 0 {
		t.Fatal("expected trees to be placed")
	}
	if !report.Valid {
		t.Fatalf("report has errors: %s", report.Summary)
	}
	t.Logf("placed %d trees", len(trees))
}

func TestTreeContexts(t *testing.T) {
	pods, adjacency, rings := bikeTestPods(t)
	s := defaultBikeSpec()

	greenZones := CollectGreenZones(s, pods)
	_, paths, _ := PlaceBuildings(s, pods, adjacency, defaultBikeParams())
	bikePaths, _ := GenerateBikePaths(pods, adjacency, rings)
	plazas, _ := GeneratePlazas(pods, s)

	trees, _ := PlaceTrees(pods, greenZones, paths, bikePaths, plazas)

	contexts := make(map[string]int)
	for _, tr := range trees {
		contexts[tr.Context]++
	}
	t.Logf("tree contexts: %v", contexts)

	for _, ctx := range []string{"park", "path", "plaza"} {
		if contexts[ctx] == 0 {
			t.Errorf("expected trees with context %q", ctx)
		}
	}
}

func TestTreeDimensions(t *testing.T) {
	pods, adjacency, rings := bikeTestPods(t)
	s := defaultBikeSpec()

	greenZones := CollectGreenZones(s, pods)
	_, paths, _ := PlaceBuildings(s, pods, adjacency, defaultBikeParams())
	bikePaths, _ := GenerateBikePaths(pods, adjacency, rings)
	plazas, _ := GeneratePlazas(pods, s)

	trees, _ := PlaceTrees(pods, greenZones, paths, bikePaths, plazas)

	for _, tr := range trees {
		if tr.Height < 6 || tr.Height > 12 {
			t.Errorf("tree %s height %.1f outside range [6, 12]", tr.ID, tr.Height)
		}
		if tr.CanopyD < 4 || tr.CanopyD > 8 {
			t.Errorf("tree %s canopy %.1f outside range [4, 8]", tr.ID, tr.CanopyD)
		}
	}
}

func TestTreesAreDeterministic(t *testing.T) {
	pods, adjacency, rings := bikeTestPods(t)
	s := defaultBikeSpec()

	greenZones := CollectGreenZones(s, pods)
	_, paths, _ := PlaceBuildings(s, pods, adjacency, defaultBikeParams())
	bikePaths, _ := GenerateBikePaths(pods, adjacency, rings)
	plazas, _ := GeneratePlazas(pods, s)

	trees1, _ := PlaceTrees(pods, greenZones, paths, bikePaths, plazas)
	trees2, _ := PlaceTrees(pods, greenZones, paths, bikePaths, plazas)

	if len(trees1) != len(trees2) {
		t.Fatalf("non-deterministic: %d vs %d trees", len(trees1), len(trees2))
	}
	for i := range trees1 {
		if trees1[i].Position != trees2[i].Position {
			t.Errorf("tree %d position differs between runs", i)
			break
		}
	}
}
