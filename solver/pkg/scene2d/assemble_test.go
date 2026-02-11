package scene2d

import (
	"testing"

	"github.com/ChicagoDave/cityplanner/pkg/analytics"
	"github.com/ChicagoDave/cityplanner/pkg/layout"
	"github.com/ChicagoDave/cityplanner/pkg/spec"
)

func testSpec() *spec.CitySpec {
	return &spec.CitySpec{
		SpecVersion: "0.1.0",
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
		Vehicles: spec.Vehicles{ArterialWidthM: 6, ServiceBranchWidthM: 4, TotalFleet: 200},
		Infrastructure: spec.Infrastructure{
			Electrical: spec.ElectricalInfra{BatteryCapacityMWh: 3000},
		},
	}
}

func testParams() *analytics.ResolvedParameters {
	return &analytics.ResolvedParameters{
		TotalPopulation: 50000,
		TotalHouseholds: 21296,
		PodCount:        6,
		TotalAreaHa:     254.47,
		Rings: []analytics.RingData{
			{Name: "center", RadiusFrom: 0, RadiusTo: 300, AreaHa: 28.27, Population: 8332, Households: 4629, PodCount: 1, PodPopulation: 8332, MaxStories: 20, AvgHouseholdSize: 1.8, ResidentialAreaHa: 8.48},
			{Name: "middle", RadiusFrom: 300, RadiusTo: 600, AreaHa: 84.82, Population: 24995, Households: 9998, PodCount: 2, PodPopulation: 12497, MaxStories: 10, AvgHouseholdSize: 2.5, ResidentialAreaHa: 50.89},
			{Name: "edge", RadiusFrom: 600, RadiusTo: 900, AreaHa: 141.37, Population: 16673, Households: 6669, PodCount: 3, PodPopulation: 5557, MaxStories: 4, AvgHouseholdSize: 2.5, ResidentialAreaHa: 84.82},
		},
	}
}

func assembleTestScene2D(t *testing.T) *Scene2D {
	t.Helper()
	s := testSpec()
	params := testParams()

	pods, adjacency, _ := layout.LayoutPods(s, params)
	buildings, paths, _ := layout.PlaceBuildings(s, pods, adjacency, params)
	greenZones := layout.CollectGreenZones(s, pods)
	bikePaths, _ := layout.GenerateBikePaths(pods, adjacency, s.CityZones.Rings)
	shuttleRoutes, stations, _ := layout.GenerateShuttleRoutes(bikePaths, pods)
	sportsFields, _ := layout.PlaceSportsFields(pods, adjacency, s.CityZones.Rings)
	plazas, _ := layout.GeneratePlazas(pods, s)
	trees, _ := layout.PlaceTrees(pods, greenZones, paths, bikePaths, plazas)

	return Assemble2D(s, params, pods, buildings, paths, greenZones,
		bikePaths, shuttleRoutes, stations, sportsFields, plazas, trees)
}

func TestAssemble2DProducesScene(t *testing.T) {
	sc := assembleTestScene2D(t)
	if sc == nil {
		t.Fatal("expected non-nil scene")
	}
	if sc.Metadata.Population == 0 {
		t.Error("expected non-zero population")
	}
	t.Logf("2D scene: %d pods, %d ped paths, %d bike paths, %d shuttle routes",
		len(sc.Pods), len(sc.Paths.Pedestrian), len(sc.Paths.Bike), len(sc.Paths.Shuttle))
}

func TestAssemble2DMetadata(t *testing.T) {
	sc := assembleTestScene2D(t)

	if sc.Metadata.Population != 50000 {
		t.Errorf("expected population 50000, got %d", sc.Metadata.Population)
	}
	if sc.Metadata.PodCount != 6 {
		t.Errorf("expected pod_count 6, got %d", sc.Metadata.PodCount)
	}
	if sc.Metadata.CityRadiusM != 900 {
		t.Errorf("expected city_radius_m 900, got %.1f", sc.Metadata.CityRadiusM)
	}
	if sc.Metadata.GeneratedAt == "" {
		t.Error("generated_at is empty")
	}
}

func TestAssemble2DRings(t *testing.T) {
	sc := assembleTestScene2D(t)

	if len(sc.Rings) != 3 {
		t.Fatalf("expected 3 rings, got %d", len(sc.Rings))
	}

	names := []string{"center", "middle", "edge"}
	for i, name := range names {
		if sc.Rings[i].Name != name {
			t.Errorf("ring %d: expected name %q, got %q", i, name, sc.Rings[i].Name)
		}
		if sc.Rings[i].MaxStories == 0 {
			t.Errorf("ring %s: max_stories is 0", name)
		}
		if sc.Rings[i].PodCount == 0 {
			t.Errorf("ring %s: pod_count is 0", name)
		}
		if sc.Rings[i].Population == 0 {
			t.Errorf("ring %s: population is 0", name)
		}
		if sc.Rings[i].Character == "" {
			t.Errorf("ring %s: character is empty", name)
		}
	}
}

func TestAssemble2DPods(t *testing.T) {
	sc := assembleTestScene2D(t)

	if len(sc.Pods) != 6 {
		t.Fatalf("expected 6 pods, got %d", len(sc.Pods))
	}

	for _, pod := range sc.Pods {
		if pod.ID == "" {
			t.Error("pod has empty ID")
		}
		if len(pod.Boundary) == 0 {
			t.Errorf("pod %s has no boundary", pod.ID)
		}
		if pod.Population == 0 {
			t.Errorf("pod %s has zero population", pod.ID)
		}
		if pod.AreaHa == 0 {
			t.Errorf("pod %s has zero area", pod.ID)
		}
	}
}

func TestAssemble2DPodZones(t *testing.T) {
	sc := assembleTestScene2D(t)

	for _, pod := range sc.Pods {
		if len(pod.Zones) == 0 {
			t.Errorf("pod %s has no zones", pod.ID)
			continue
		}
		for _, z := range pod.Zones {
			if z.Type == "" {
				t.Errorf("pod %s: zone has empty type", pod.ID)
			}
			if len(z.Polygon) < 3 {
				t.Errorf("pod %s: zone %s has < 3 polygon vertices", pod.ID, z.Type)
			}
		}
		t.Logf("pod %s: %d zones", pod.ID, len(pod.Zones))
	}
}

func TestAssemble2DPaths(t *testing.T) {
	sc := assembleTestScene2D(t)

	if len(sc.Paths.Pedestrian) == 0 {
		t.Error("no pedestrian paths")
	}
	if len(sc.Paths.Bike) == 0 {
		t.Error("no bike paths")
	}
	if len(sc.Paths.Shuttle) == 0 {
		t.Error("no shuttle routes")
	}

	t.Logf("paths: %d pedestrian, %d bike, %d shuttle",
		len(sc.Paths.Pedestrian), len(sc.Paths.Bike), len(sc.Paths.Shuttle))
}

func TestAssemble2DStations(t *testing.T) {
	sc := assembleTestScene2D(t)

	if len(sc.Stations) != len(sc.Pods) {
		t.Errorf("expected %d stations (one per pod), got %d", len(sc.Pods), len(sc.Stations))
	}
	for _, st := range sc.Stations {
		if st.PodID == "" {
			t.Error("station has empty pod_id")
		}
		if st.RouteID == "" {
			t.Error("station has empty route_id")
		}
	}
}

func TestAssemble2DPlazas(t *testing.T) {
	sc := assembleTestScene2D(t)

	if len(sc.Plazas) != len(sc.Pods) {
		t.Errorf("expected %d plazas (one per pod), got %d", len(sc.Pods), len(sc.Plazas))
	}
	for _, p := range sc.Plazas {
		if p.PodID == "" {
			t.Error("plaza has empty pod_id")
		}
		if p.Width == 0 || p.Depth == 0 {
			t.Errorf("plaza %s has zero dimensions", p.ID)
		}
	}
}

func TestAssemble2DTreeSummary(t *testing.T) {
	sc := assembleTestScene2D(t)

	if sc.Trees.Total == 0 {
		t.Fatal("tree total is 0")
	}
	if sc.Trees.ParkCount == 0 {
		t.Error("park_count is 0")
	}
	if sc.Trees.PathCount == 0 {
		t.Error("path_count is 0")
	}
	if sc.Trees.PlazaCount == 0 {
		t.Error("plaza_count is 0")
	}
	sum := sc.Trees.ParkCount + sc.Trees.PathCount + sc.Trees.PlazaCount
	if sum != sc.Trees.Total {
		t.Errorf("park(%d) + path(%d) + plaza(%d) = %d, but total = %d",
			sc.Trees.ParkCount, sc.Trees.PathCount, sc.Trees.PlazaCount, sum, sc.Trees.Total)
	}
	t.Logf("trees: park=%d path=%d plaza=%d total=%d",
		sc.Trees.ParkCount, sc.Trees.PathCount, sc.Trees.PlazaCount, sc.Trees.Total)
}

func TestAssemble2DBuildingSummary(t *testing.T) {
	sc := assembleTestScene2D(t)

	if sc.Buildings.TotalBuildings == 0 {
		t.Fatal("total_buildings is 0")
	}
	if sc.Buildings.TotalDU == 0 {
		t.Error("total_dwelling_units is 0")
	}
	if len(sc.Buildings.ByPod) == 0 {
		t.Fatal("by_pod is empty")
	}

	for _, pod := range sc.Pods {
		if _, ok := sc.Buildings.ByPod[pod.ID]; !ok {
			t.Errorf("by_pod missing entry for pod %s", pod.ID)
		}
	}
	t.Logf("buildings: %d total, %d dwelling units, %d pods",
		sc.Buildings.TotalBuildings, sc.Buildings.TotalDU, len(sc.Buildings.ByPod))
}

func TestAssemble2DExternalBandNil(t *testing.T) {
	sc := assembleTestScene2D(t)

	// Default test spec has no perimeter defined, so external_band should be nil.
	if sc.ExternalBand != nil {
		t.Errorf("expected nil external_band, got %+v", sc.ExternalBand)
	}
}
