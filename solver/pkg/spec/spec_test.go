package spec

import (
	"math"
	"testing"
)

func TestLoadProject(t *testing.T) {
	s, err := LoadProject("../../../examples/default-city")
	if err != nil {
		t.Fatalf("LoadProject failed: %v", err)
	}

	if s.SpecVersion != "0.2.0" {
		t.Errorf("spec_version = %q, want %q", s.SpecVersion, "0.2.0")
	}
	if s.City.Population != 64000 {
		t.Errorf("population = %d, want 64000", s.City.Population)
	}
	if s.City.FootprintShape != "circle" {
		t.Errorf("footprint_shape = %q, want %q", s.City.FootprintShape, "circle")
	}
	if s.City.ExcavationDepth != 8 {
		t.Errorf("excavation_depth = %v, want 8", s.City.ExcavationDepth)
	}
	if s.City.MaxHeightCenter != 32 {
		t.Errorf("max_height_center = %d, want 32", s.City.MaxHeightCenter)
	}
	if s.City.MaxHeightEdge != 2 {
		t.Errorf("max_height_edge = %d, want 2", s.City.MaxHeightEdge)
	}

	// Rings: 5 rings from center to ring1
	if len(s.CityZones.Rings) != 5 {
		t.Fatalf("expected 5 rings, got %d", len(s.CityZones.Rings))
	}
	if s.CityZones.Rings[0].Name != "center" || s.CityZones.Rings[0].RadiusTo != 250 {
		t.Errorf("ring[0] = %s radius_to=%v, want center/250", s.CityZones.Rings[0].Name, s.CityZones.Rings[0].RadiusTo)
	}
	if s.CityZones.Rings[4].Name != "ring1" || s.CityZones.Rings[4].RadiusTo != 2200 {
		t.Errorf("ring[4] = %s radius_to=%v, want ring1/2200", s.CityZones.Rings[4].Name, s.CityZones.Rings[4].RadiusTo)
	}

	// Demographics
	demoSum := s.Demographics.Singles + s.Demographics.Couples +
		s.Demographics.FamiliesYoung + s.Demographics.FamiliesTeen +
		s.Demographics.EmptyNest + s.Demographics.Retirees
	if math.Abs(demoSum-1.0) > 0.01 {
		t.Errorf("demographics sum = %v, want ~1.0", demoSum)
	}

	// Pods
	if s.Pods.WalkRadius != 400 {
		t.Errorf("walk_radius = %v, want 400", s.Pods.WalkRadius)
	}
	if len(s.Pods.RingAssignments) != 5 {
		t.Errorf("ring_assignments count = %d, want 5", len(s.Pods.RingAssignments))
	}
	ring1, ok := s.Pods.RingAssignments["ring1"]
	if !ok {
		t.Fatal("missing ring1 ring assignment")
	}
	if len(ring1.RequiredServices) != 5 {
		t.Errorf("ring1 required_services = %d, want 5", len(ring1.RequiredServices))
	}

	// Infrastructure
	if s.Infrastructure.Electrical.PeakDemandKWPer != 2.5 {
		t.Errorf("peak_demand_kw_per_capita = %v, want 2.5", s.Infrastructure.Electrical.PeakDemandKWPer)
	}
	if s.Infrastructure.Electrical.BatteryCapacityMWh != 3840 {
		t.Errorf("battery_capacity_mwh = %v, want 3840", s.Infrastructure.Electrical.BatteryCapacityMWh)
	}

	// Revenue
	if s.Revenue.DebtTermYears != 30 {
		t.Errorf("debt_term_years = %d, want 30", s.Revenue.DebtTermYears)
	}
	if s.Revenue.InterestRate != 0.05 {
		t.Errorf("interest_rate = %v, want 0.05", s.Revenue.InterestRate)
	}
}

func TestLoadProjectMissing(t *testing.T) {
	_, err := LoadProject("/nonexistent/path")
	if err == nil {
		t.Error("expected error for missing project directory")
	}
}
