package analytics

import (
	"math"

	"github.com/ChicagoDave/cityplanner/pkg/spec"
)

// cohortDef defines the fixed household size and adult/child breakdown per cohort.
type cohortDef struct {
	name          string
	householdSize float64
	adultsPerHH   float64
	childrenPerHH float64
}

// Household sizes and compositions from the technical specification.
var cohortDefs = []cohortDef{
	{"singles", 1.0, 1.0, 0.0},
	{"couples", 2.0, 2.0, 0.0},
	{"families_young", 3.5, 2.0, 1.5},
	{"families_teen", 4.0, 2.0, 2.0},
	{"empty_nest", 2.0, 2.0, 0.0},
	{"retirees", 1.5, 1.5, 0.0},
}

// cohortRatio extracts the ratio for a cohort from the Demographics struct.
func cohortRatio(d *spec.Demographics, name string) float64 {
	switch name {
	case "singles":
		return d.Singles
	case "couples":
		return d.Couples
	case "families_young":
		return d.FamiliesYoung
	case "families_teen":
		return d.FamiliesTeen
	case "empty_nest":
		return d.EmptyNest
	case "retirees":
		return d.Retirees
	}
	return 0
}

// resolveDemographics computes cohort breakdowns and weighted average household size.
func resolveDemographics(s *spec.CitySpec) ([]CohortBreakdown, float64) {
	// Compute weighted average household size
	weightedAvg := 0.0
	for _, cd := range cohortDefs {
		ratio := cohortRatio(&s.Demographics, cd.name)
		weightedAvg += ratio * cd.householdSize
	}
	if weightedAvg == 0 {
		return nil, 0
	}

	totalHH := int(math.Round(float64(s.City.Population) / weightedAvg))

	cohorts := make([]CohortBreakdown, 0, len(cohortDefs))
	assignedHH := 0

	for i, cd := range cohortDefs {
		ratio := cohortRatio(&s.Demographics, cd.name)
		hh := int(math.Round(float64(totalHH) * ratio))

		// Last cohort absorbs rounding difference
		if i == len(cohortDefs)-1 {
			hh = totalHH - assignedHH
		}
		assignedHH += hh

		pop := int(math.Round(float64(hh) * cd.householdSize))
		adults := int(math.Round(float64(hh) * cd.adultsPerHH))
		children := int(math.Round(float64(hh) * cd.childrenPerHH))

		cohorts = append(cohorts, CohortBreakdown{
			Name:          cd.name,
			Ratio:         ratio,
			HouseholdSize: cd.householdSize,
			Households:    hh,
			Population:    pop,
			Adults:        adults,
			Children:      children,
		})
	}

	return cohorts, weightedAvg
}

// computeDependencyRatio computes dependents / working-age adults.
// Working age = all adults except retirees.
// Dependents = all children + retiree population.
func computeDependencyRatio(cohorts []CohortBreakdown) float64 {
	workingAge := 0
	dependents := 0
	for _, c := range cohorts {
		if c.Name == "retirees" {
			dependents += c.Population
		} else {
			workingAge += c.Adults
			dependents += c.Children
		}
	}
	if workingAge == 0 {
		return 0
	}
	return float64(dependents) / float64(workingAge)
}

// sumCohortTotals returns total adults, children, and estimated students.
func sumCohortTotals(cohorts []CohortBreakdown) (adults, children, students int) {
	for _, c := range cohorts {
		adults += c.Adults
		children += c.Children
	}
	// Students = all children (elementary + secondary age)
	students = children
	return
}
