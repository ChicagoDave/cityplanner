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

	if s.SpecVersion != "0.1.0" {
		t.Errorf("spec_version = %q, want %q", s.SpecVersion, "0.1.0")
	}
	if s.City.Population != 50000 {
		t.Errorf("population = %d, want 50000", s.City.Population)
	}
	if s.City.FootprintShape != "circle" {
		t.Errorf("footprint_shape = %q, want %q", s.City.FootprintShape, "circle")
	}
	if s.City.ExcavationDepth != 8 {
		t.Errorf("excavation_depth = %v, want 8", s.City.ExcavationDepth)
	}
	if s.City.MaxHeightCenter != 20 {
		t.Errorf("max_height_center = %d, want 20", s.City.MaxHeightCenter)
	}
	if s.City.MaxHeightEdge != 4 {
		t.Errorf("max_height_edge = %d, want 4", s.City.MaxHeightEdge)
	}

	// Zones
	if s.CityZones.Center.RadiusTo != 300 {
		t.Errorf("center.radius_to = %v, want 300", s.CityZones.Center.RadiusTo)
	}
	if s.CityZones.Middle.RadiusFrom != 300 || s.CityZones.Middle.RadiusTo != 600 {
		t.Errorf("middle radius = %v-%v, want 300-600", s.CityZones.Middle.RadiusFrom, s.CityZones.Middle.RadiusTo)
	}
	if s.CityZones.Edge.RadiusTo != 900 {
		t.Errorf("edge.radius_to = %v, want 900", s.CityZones.Edge.RadiusTo)
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
	if len(s.Pods.RingAssignments) != 3 {
		t.Errorf("ring_assignments count = %d, want 3", len(s.Pods.RingAssignments))
	}
	edge, ok := s.Pods.RingAssignments["edge"]
	if !ok {
		t.Fatal("missing edge ring assignment")
	}
	if len(edge.RequiredServices) != 6 {
		t.Errorf("edge required_services = %d, want 6", len(edge.RequiredServices))
	}

	// Infrastructure
	if s.Infrastructure.Electrical.PeakDemandKWPer != 2.5 {
		t.Errorf("peak_demand_kw_per_capita = %v, want 2.5", s.Infrastructure.Electrical.PeakDemandKWPer)
	}
	if s.Infrastructure.Electrical.BatteryCapacityMWh != 3000 {
		t.Errorf("battery_capacity_mwh = %v, want 3000", s.Infrastructure.Electrical.BatteryCapacityMWh)
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
