package routing

import (
	"math"
	"testing"

	"github.com/ChicagoDave/cityplanner/pkg/analytics"
	"github.com/ChicagoDave/cityplanner/pkg/layout"
	"github.com/ChicagoDave/cityplanner/pkg/spec"
)

func defaultSpec() *spec.CitySpec {
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
		Vehicles: spec.Vehicles{
			ArterialWidthM:      6,
			ServiceBranchWidthM: 4,
			TotalFleet:          200,
		},
		Infrastructure: spec.Infrastructure{
			Telecom: spec.TelecomInfra{NodeSpacingM: 75},
		},
	}
}

func defaultParams() *analytics.ResolvedParameters {
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

func setupRouting(t *testing.T) ([]Segment, *spec.CitySpec, []layout.Pod) {
	t.Helper()
	s := defaultSpec()
	params := defaultParams()
	pods, _, podReport := layout.LayoutPods(s, params)
	if !podReport.Valid {
		t.Fatalf("pod layout failed: %v", podReport.Errors)
	}
	buildings, _, _ := layout.PlaceBuildings(s, pods, nil, params)
	segments, report := RouteInfrastructure(s, pods, buildings)
	if !report.Valid {
		t.Fatalf("routing report invalid: %v", report.Errors)
	}
	return segments, s, pods
}

func TestRouteInfrastructureProducesSegments(t *testing.T) {
	segments, _, _ := setupRouting(t)
	if len(segments) == 0 {
		t.Fatal("expected routing segments, got 0")
	}
	t.Logf("generated %d routing segments", len(segments))
}

func TestRouteInfrastructureAllNetworksPresent(t *testing.T) {
	segments, _, _ := setupRouting(t)
	netCounts := map[NetworkType]int{}
	for _, seg := range segments {
		netCounts[seg.Network]++
	}
	for _, net := range []NetworkType{NetworkSewage, NetworkWater, NetworkElectrical, NetworkTelecom, NetworkVehicle} {
		if netCounts[net] == 0 {
			t.Errorf("no segments for network %s", net)
		}
	}
	t.Logf("network segment counts: %v", netCounts)
}

func TestRouteInfrastructureLayerAssignment(t *testing.T) {
	segments, _, _ := setupRouting(t)
	for _, seg := range segments {
		switch seg.Network {
		case NetworkSewage, NetworkWater:
			if seg.Layer != 1 {
				t.Errorf("%s segment %s: expected layer 1, got %d", seg.Network, seg.ID, seg.Layer)
			}
		case NetworkElectrical, NetworkTelecom:
			if seg.Layer != 2 {
				t.Errorf("%s segment %s: expected layer 2, got %d", seg.Network, seg.ID, seg.Layer)
			}
		case NetworkVehicle:
			if seg.Layer != 3 {
				t.Errorf("%s segment %s: expected layer 3, got %d", seg.Network, seg.ID, seg.Layer)
			}
		}
	}
}

func TestRouteInfrastructureYOffsets(t *testing.T) {
	segments, _, _ := setupRouting(t)
	for _, seg := range segments {
		var expectedY float64
		switch seg.Layer {
		case 1:
			expectedY = yLayer1
		case 2:
			expectedY = yLayer2
		case 3:
			expectedY = yLayer3
		}
		if seg.Start[1] != expectedY || seg.End[1] != expectedY {
			t.Errorf("segment %s: expected Y=%.1f, got start Y=%.1f end Y=%.1f",
				seg.ID, expectedY, seg.Start[1], seg.End[1])
		}
	}
}

func TestRouteInfrastructureCapacityPositive(t *testing.T) {
	segments, _, _ := setupRouting(t)
	for _, seg := range segments {
		if seg.Capacity <= 0 {
			t.Errorf("segment %s (%s): expected positive capacity, got %.2f", seg.ID, seg.Network, seg.Capacity)
		}
	}
}

func TestRouteInfrastructureVehicleWidths(t *testing.T) {
	segments, s, _ := setupRouting(t)
	for _, seg := range segments {
		if seg.Network != NetworkVehicle {
			continue
		}
		if seg.IsTrunk && seg.WidthM != s.Vehicles.ArterialWidthM {
			t.Errorf("vehicle trunk %s: expected width %.1f, got %.1f", seg.ID, s.Vehicles.ArterialWidthM, seg.WidthM)
		}
		if !seg.IsTrunk && seg.WidthM != s.Vehicles.ServiceBranchWidthM {
			t.Errorf("vehicle branch %s: expected width %.1f, got %.1f", seg.ID, s.Vehicles.ServiceBranchWidthM, seg.WidthM)
		}
	}
}

func TestRouteInfrastructureTrunkAndBranch(t *testing.T) {
	segments, _, _ := setupRouting(t)
	trunkCount, branchCount := 0, 0
	for _, seg := range segments {
		if seg.IsTrunk {
			trunkCount++
		} else {
			branchCount++
		}
	}
	if trunkCount == 0 {
		t.Error("no trunk segments generated")
	}
	if branchCount == 0 {
		t.Error("no branch segments generated")
	}
	t.Logf("trunk: %d, branch: %d", trunkCount, branchCount)
}

func TestRouteInfrastructureEndpointsWithinBounds(t *testing.T) {
	segments, _, _ := setupRouting(t)
	maxR := 900.0 + 10.0 // perimeter + small tolerance
	for _, seg := range segments {
		for _, pt := range [][3]float64{seg.Start, seg.End} {
			r := math.Hypot(pt[0], pt[2])
			if r > maxR {
				t.Errorf("segment %s: endpoint (%.1f, %.1f) at radius %.1f exceeds bounds %.1f",
					seg.ID, pt[0], pt[2], r, maxR)
			}
		}
	}
}
