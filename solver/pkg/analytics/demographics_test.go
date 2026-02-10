package analytics

import (
	"math"
	"testing"

	"github.com/ChicagoDave/cityplanner/pkg/spec"
)

func defaultDemographics() *spec.CitySpec {
	return &spec.CitySpec{
		City: spec.CityDef{Population: 50000},
		Demographics: spec.Demographics{
			Singles: 0.15, Couples: 0.20, FamiliesYoung: 0.25,
			FamiliesTeen: 0.15, EmptyNest: 0.15, Retirees: 0.10,
		},
	}
}

func TestResolveDemographicsWeightedAvg(t *testing.T) {
	s := defaultDemographics()
	_, avg := resolveDemographics(s)

	// 0.15*1 + 0.20*2 + 0.25*3.5 + 0.15*4 + 0.15*2 + 0.10*1.5 = 2.475
	expected := 2.475
	if math.Abs(avg-expected) > 0.001 {
		t.Errorf("weighted avg = %v, want %v", avg, expected)
	}
}

func TestResolveDemographicsTotalHouseholds(t *testing.T) {
	s := defaultDemographics()
	cohorts, avg := resolveDemographics(s)

	expectedHH := int(math.Round(50000.0 / avg))
	totalHH := 0
	for _, c := range cohorts {
		totalHH += c.Households
	}
	if totalHH != expectedHH {
		t.Errorf("total households = %d, want %d", totalHH, expectedHH)
	}
}

func TestResolveDemographicsCohortCounts(t *testing.T) {
	s := defaultDemographics()
	cohorts, _ := resolveDemographics(s)

	if len(cohorts) != 6 {
		t.Fatalf("expected 6 cohorts, got %d", len(cohorts))
	}

	// Spot check: singles should be ~15% of total HH
	singles := cohorts[0]
	if singles.Name != "singles" {
		t.Errorf("first cohort name = %q, want singles", singles.Name)
	}
	if singles.HouseholdSize != 1.0 {
		t.Errorf("singles household_size = %v, want 1.0", singles.HouseholdSize)
	}
	// Each single household = 1 adult, 0 children
	if singles.Children != 0 {
		t.Errorf("singles children = %d, want 0", singles.Children)
	}
	if singles.Adults != singles.Households {
		t.Errorf("singles adults = %d, want %d (= households)", singles.Adults, singles.Households)
	}

	// families_young: 2 adults + 1.5 children per HH
	fy := cohorts[2]
	if fy.Name != "families_young" {
		t.Errorf("third cohort name = %q, want families_young", fy.Name)
	}
	if fy.Children == 0 {
		t.Error("families_young should have children")
	}
}

func TestDependencyRatio(t *testing.T) {
	s := defaultDemographics()
	cohorts, _ := resolveDemographics(s)
	ratio := computeDependencyRatio(cohorts)

	// Should be in the 0.3-0.8 range
	if ratio < 0.2 || ratio > 1.0 {
		t.Errorf("dependency ratio = %v, expected reasonable range", ratio)
	}
}

func TestSumCohortTotals(t *testing.T) {
	s := defaultDemographics()
	cohorts, _ := resolveDemographics(s)
	adults, children, students := sumCohortTotals(cohorts)

	if adults <= 0 {
		t.Error("expected positive adult count")
	}
	if children <= 0 {
		t.Error("expected positive child count from family cohorts")
	}
	if students != children {
		t.Errorf("students = %d, expected = %d (same as children)", students, children)
	}
	// Total population should be close to adults + children
	totalPop := adults + children
	if math.Abs(float64(totalPop-50000)) > 500 {
		t.Errorf("total population from cohorts = %d, expected ~50000", totalPop)
	}
}

func TestAllSingles(t *testing.T) {
	s := &spec.CitySpec{
		City:         spec.CityDef{Population: 10000},
		Demographics: spec.Demographics{Singles: 1.0},
	}
	cohorts, avg := resolveDemographics(s)
	if math.Abs(avg-1.0) > 0.001 {
		t.Errorf("avg = %v, want 1.0", avg)
	}
	totalHH := 0
	for _, c := range cohorts {
		totalHH += c.Households
	}
	if totalHH != 10000 {
		t.Errorf("total households = %d, want 10000", totalHH)
	}

	ratio := computeDependencyRatio(cohorts)
	if ratio != 0 {
		t.Errorf("dependency ratio = %v, want 0 for all singles", ratio)
	}
}
