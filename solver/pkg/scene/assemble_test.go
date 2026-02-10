package scene

import (
	"testing"

	"github.com/ChicagoDave/cityplanner/pkg/analytics"
	"github.com/ChicagoDave/cityplanner/pkg/layout"
	"github.com/ChicagoDave/cityplanner/pkg/routing"
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
			Center: spec.ZoneDef{RadiusFrom: 0, RadiusTo: 300, MaxStories: 20, Character: "civic_commercial"},
			Middle: spec.ZoneDef{RadiusFrom: 300, RadiusTo: 600, MaxStories: 10, Character: "mixed_residential_commercial"},
			Edge:   spec.ZoneDef{RadiusFrom: 600, RadiusTo: 900, MaxStories: 4, Character: "family_education"},
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
	}
}

func testParams() *analytics.ResolvedParameters {
	return &analytics.ResolvedParameters{
		TotalPopulation: 50000,
		TotalHouseholds: 20202,
		PodCount:        6,
		TotalAreaHa:     254.47,
		Rings: []analytics.RingData{
			{Name: "center", RadiusFrom: 0, RadiusTo: 300, AreaHa: 28.27, Population: 8333, Households: 3367, PodCount: 1, PodPopulation: 8333, MaxStories: 20, ResidentialAreaHa: 16.96},
			{Name: "middle", RadiusFrom: 300, RadiusTo: 600, AreaHa: 84.82, Population: 16667, Households: 6734, PodCount: 2, PodPopulation: 8333, MaxStories: 10, ResidentialAreaHa: 50.89},
			{Name: "edge", RadiusFrom: 600, RadiusTo: 900, AreaHa: 141.37, Population: 25000, Households: 10101, PodCount: 3, PodPopulation: 8333, MaxStories: 4, ResidentialAreaHa: 84.82},
		},
	}
}

func assembleTestGraph(t *testing.T) *Graph {
	t.Helper()
	s := testSpec()
	params := testParams()

	pods, adjacency, _ := layout.LayoutPods(s, params)
	buildings, paths, _ := layout.PlaceBuildings(s, pods, adjacency, params)
	segments, _ := routing.RouteInfrastructure(s, pods, buildings)
	greenZones := layout.CollectGreenZones(s, pods)

	return Assemble(s, pods, buildings, paths, segments, greenZones)
}

func TestAssembleProducesGraph(t *testing.T) {
	g := assembleTestGraph(t)
	if g == nil {
		t.Fatal("expected non-nil graph")
	}
	if len(g.Entities) == 0 {
		t.Fatal("expected entities")
	}
	t.Logf("scene graph: %d entities", len(g.Entities))
}

func TestAssembleHasAllEntityTypes(t *testing.T) {
	g := assembleTestGraph(t)
	for _, et := range []EntityType{EntityBuilding, EntityPath, EntityPipe, EntityLane, EntityPark} {
		if len(g.Groups.EntityTypes[et]) == 0 {
			t.Errorf("no entities of type %s", et)
		} else {
			t.Logf("%s: %d entities", et, len(g.Groups.EntityTypes[et]))
		}
	}
}

func TestAssembleGroupsPopulated(t *testing.T) {
	g := assembleTestGraph(t)

	if len(g.Groups.Pods) == 0 {
		t.Error("pods group is empty")
	}
	if len(g.Groups.Systems) == 0 {
		t.Error("systems group is empty")
	}
	if len(g.Groups.Layers) == 0 {
		t.Error("layers group is empty")
	}
	if len(g.Groups.EntityTypes) == 0 {
		t.Error("entity_types group is empty")
	}
	t.Logf("groups: %d pods, %d systems, %d layers, %d entity_types",
		len(g.Groups.Pods), len(g.Groups.Systems), len(g.Groups.Layers), len(g.Groups.EntityTypes))
}

func TestAssembleMetadata(t *testing.T) {
	g := assembleTestGraph(t)

	if g.Metadata.SpecVersion != "0.1.0" {
		t.Errorf("expected spec_version 0.1.0, got %s", g.Metadata.SpecVersion)
	}
	if g.Metadata.GeneratedAt == "" {
		t.Error("generated_at is empty")
	}
	if g.Metadata.CityBounds.Min.X >= g.Metadata.CityBounds.Max.X {
		t.Error("city_bounds min.x >= max.x")
	}
}

func TestAssembleBoundsEncloseEntities(t *testing.T) {
	g := assembleTestGraph(t)
	bounds := g.Metadata.CityBounds

	for _, e := range g.Entities {
		if e.Position.X < bounds.Min.X-1 || e.Position.X > bounds.Max.X+1 {
			t.Errorf("entity %s X=%.1f outside bounds [%.1f, %.1f]",
				e.ID, e.Position.X, bounds.Min.X, bounds.Max.X)
			break
		}
		if e.Position.Z < bounds.Min.Z-1 || e.Position.Z > bounds.Max.Z+1 {
			t.Errorf("entity %s Z=%.1f outside bounds [%.1f, %.1f]",
				e.ID, e.Position.Z, bounds.Min.Z, bounds.Max.Z)
			break
		}
	}
}

func TestAssembleLayerAssignment(t *testing.T) {
	g := assembleTestGraph(t)

	surfaceCount := len(g.Groups.Layers[LayerSurface])
	ug1Count := len(g.Groups.Layers[LayerUnderground1])
	ug2Count := len(g.Groups.Layers[LayerUnderground2])
	ug3Count := len(g.Groups.Layers[LayerUnderground3])

	if surfaceCount == 0 {
		t.Error("no surface entities")
	}
	if ug1Count == 0 {
		t.Error("no underground_1 entities")
	}
	if ug2Count == 0 {
		t.Error("no underground_2 entities")
	}
	if ug3Count == 0 {
		t.Error("no underground_3 entities")
	}
	t.Logf("layers: surface=%d ug1=%d ug2=%d ug3=%d", surfaceCount, ug1Count, ug2Count, ug3Count)
}

func TestAssembleUniqueEntityIDs(t *testing.T) {
	g := assembleTestGraph(t)
	seen := map[string]bool{}
	for _, e := range g.Entities {
		if seen[e.ID] {
			t.Errorf("duplicate entity ID: %s", e.ID)
		}
		seen[e.ID] = true
	}
}

func TestAssembleSystemsCoverAllNetworks(t *testing.T) {
	g := assembleTestGraph(t)
	for _, sys := range []SystemType{SystemSewage, SystemWater, SystemElectrical, SystemTelecom, SystemVehicle} {
		if len(g.Groups.Systems[sys]) == 0 {
			t.Errorf("no entities for system %s", sys)
		}
	}
}
