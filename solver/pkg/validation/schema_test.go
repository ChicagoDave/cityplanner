package validation

import (
	"testing"

	"github.com/ChicagoDave/cityplanner/pkg/spec"
)

func validSpec() *spec.CitySpec {
	return &spec.CitySpec{
		SpecVersion: "0.1.0",
		City: spec.CityDef{
			Population:      50000,
			FootprintShape:  "circle",
			ExcavationDepth: 8,
			HeightProfile:   "bowl",
			MaxHeightCenter: 20,
			MaxHeightEdge:   4,
		},
		CityZones: spec.CityZones{
			Center: spec.ZoneDef{Character: "civic_commercial", RadiusFrom: 0, RadiusTo: 300, MaxStories: 20},
			Middle: spec.ZoneDef{Character: "mixed", RadiusFrom: 300, RadiusTo: 600, MaxStories: 10},
			Edge:   spec.ZoneDef{Character: "family", RadiusFrom: 600, RadiusTo: 900, MaxStories: 4},
			Perimeter: spec.PerimeterDef{RadiusFrom: 900, RadiusTo: 1100},
			SolarRing: spec.SolarRingDef{RadiusFrom: 1100, RadiusTo: 1500, AreaHa: 250, CapacityMW: 500, AvgOutputMW: 100},
		},
		Pods: spec.PodsDef{
			WalkRadius: 400,
			RingAssignments: map[string]spec.PodRing{
				"center": {Character: "civic", RequiredServices: []string{"hospital"}, MaxStories: 20},
				"middle": {Character: "mixed", RequiredServices: []string{"medical_clinic"}, MaxStories: 10},
				"edge":   {Character: "family", RequiredServices: []string{"grocery"}, MaxStories: 4},
			},
		},
		Demographics: spec.Demographics{
			Singles: 0.15, Couples: 0.20, FamiliesYoung: 0.25,
			FamiliesTeen: 0.15, EmptyNest: 0.15, Retirees: 0.10,
		},
		Infrastructure: spec.Infrastructure{
			Water:  spec.WaterInfra{CapacityGPDPer: 100},
			Sewage: spec.SewageInfra{CapacityGPDPer: 95},
			Electrical: spec.ElectricalInfra{
				SolarIntegratedAvgMW: 80, SolarFarmAvgMW: 100,
				BatteryCapacityMWh: 3000, GridCapacityMW: 150,
				PeakDemandKWPer: 2.5,
			},
		},
		Revenue: spec.Revenue{DebtTermYears: 30, InterestRate: 0.05, AnnualOpsCostM: 100},
		Site:    spec.SiteRequirements{MinAreaHa: 800, SolarIrradiance: 4.5},
	}
}

func TestValidateSchemaValid(t *testing.T) {
	r := ValidateSchema(validSpec())
	if !r.Valid {
		t.Errorf("expected valid report, got %d errors: %v", len(r.Errors), r.Errors)
	}
}

func TestValidateSchemaPopulationZero(t *testing.T) {
	s := validSpec()
	s.City.Population = 0
	r := ValidateSchema(s)
	if r.Valid {
		t.Error("expected invalid report for population=0")
	}
	assertHasError(t, r, "city.population")
}

func TestValidateSchemaDemographicsSum(t *testing.T) {
	s := validSpec()
	s.Demographics.Singles = 0.50 // sum now 1.35
	r := ValidateSchema(s)
	if r.Valid {
		t.Error("expected invalid report for demographics sum != 1.0")
	}
	assertHasError(t, r, "demographics")
}

func TestValidateSchemaNegativeRatio(t *testing.T) {
	s := validSpec()
	s.Demographics.Retirees = -0.10
	s.Demographics.Singles = 0.25 // keep sum at 1.0
	r := ValidateSchema(s)
	if r.Valid {
		t.Error("expected invalid for negative ratio")
	}
	assertHasError(t, r, "demographics.retirees")
}

func TestValidateSchemaZoneGap(t *testing.T) {
	s := validSpec()
	s.CityZones.Middle.RadiusFrom = 350 // gap between center(300) and middle(350)
	r := ValidateSchema(s)
	if r.Valid {
		t.Error("expected invalid for zone gap")
	}
	assertHasError(t, r, "city_zones.middle.radius_from")
}

func TestValidateSchemaZoneInverted(t *testing.T) {
	s := validSpec()
	s.CityZones.Center.RadiusFrom = 300
	s.CityZones.Center.RadiusTo = 0
	r := ValidateSchema(s)
	if r.Valid {
		t.Error("expected invalid for inverted zone radii")
	}
}

func TestValidateSchemaWalkRadius(t *testing.T) {
	s := validSpec()
	s.Pods.WalkRadius = 100 // below 200
	r := ValidateSchema(s)
	if r.Valid {
		t.Error("expected invalid for walk_radius out of range")
	}
	assertHasError(t, r, "pods.walk_radius")
}

func TestValidateSchemaExcavationDepth(t *testing.T) {
	s := validSpec()
	s.City.ExcavationDepth = 20 // above 15
	r := ValidateSchema(s)
	if r.Valid {
		t.Error("expected invalid for excavation_depth out of range")
	}
	assertHasError(t, r, "city.excavation_depth")
}

func TestValidateSchemaRevenue(t *testing.T) {
	s := validSpec()
	s.Revenue.DebtTermYears = 0
	r := ValidateSchema(s)
	if r.Valid {
		t.Error("expected invalid for debt_term_years=0")
	}
	assertHasError(t, r, "revenue.debt_term_years")
}

func TestValidateSchemaMaxStories(t *testing.T) {
	s := validSpec()
	s.CityZones.Edge.MaxStories = 0
	r := ValidateSchema(s)
	if r.Valid {
		t.Error("expected invalid for max_stories=0")
	}
}

func assertHasError(t *testing.T, r *Report, specPath string) {
	t.Helper()
	for _, e := range r.Errors {
		if e.SpecPath == specPath {
			return
		}
	}
	t.Errorf("expected error with spec_path %q, got errors: %v", specPath, r.Errors)
}
