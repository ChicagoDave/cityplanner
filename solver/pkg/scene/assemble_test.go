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
	// Capacity-weighted: center gets more per pod (civic_commercial, tall),
	// edge gets less (family, low-rise).
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

func assembleTestGraph(t *testing.T) *Graph {
	t.Helper()
	s := testSpec()
	params := testParams()

	pods, adjacency, _ := layout.LayoutPods(s, params)
	buildings, paths, _ := layout.PlaceBuildings(s, pods, adjacency, params)
	segments, _ := routing.RouteInfrastructure(s, pods, buildings)
	greenZones := layout.CollectGreenZones(s, pods)
	bikePaths, _ := layout.GenerateBikePaths(pods, adjacency, s.CityZones.Rings)
	shuttleRoutes, stations, _ := layout.GenerateShuttleRoutes(bikePaths, pods)
	sportsFields, _ := layout.PlaceSportsFields(pods, adjacency, s.CityZones.Rings)

	return Assemble(s, pods, buildings, paths, segments, greenZones, bikePaths, shuttleRoutes, stations, sportsFields)
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
	for _, et := range []EntityType{EntityBuilding, EntityPath, EntityPipe, EntityLane, EntityPark, EntityPedway, EntityBikeTunnel, EntityBattery, EntityBikePath, EntityShuttleRoute, EntityStation, EntitySportsField} {
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
		t.Error("no underground_3 entities (vehicles should be in layer 3)")
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
	for _, sys := range []SystemType{SystemSewage, SystemWater, SystemElectrical, SystemTelecom, SystemVehicle, SystemPedestrian, SystemBicycle, SystemShuttle} {
		if len(g.Groups.Systems[sys]) == 0 {
			t.Errorf("no entities for system %s", sys)
		}
	}
}

func TestAssembleConnectivityInMetadata(t *testing.T) {
	g := assembleTestGraph(t)
	connCount := 0
	for _, e := range g.Entities {
		if e.Metadata != nil {
			if ct, ok := e.Metadata["connected_to"]; ok && ct != nil {
				if ids, ok := ct.([]string); ok && len(ids) > 0 {
					connCount++
				}
			}
		}
	}
	if connCount == 0 {
		t.Error("no entities have connected_to in metadata")
	}
	t.Logf("%d entities have connectivity data", connCount)
}
