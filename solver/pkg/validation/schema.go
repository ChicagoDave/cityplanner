package validation

import (
	"fmt"
	"math"

	"github.com/ChicagoDave/cityplanner/pkg/spec"
)

// ValidateSchema performs Level 1 (schema) validation on a parsed CitySpec.
// It checks structural correctness before any computation.
func ValidateSchema(s *spec.CitySpec) *Report {
	r := NewReport()

	validatePopulation(s, r)
	validateDemographics(s, r)
	validateZones(s, r)
	validatePods(s, r)
	validateCity(s, r)
	validateRevenue(s, r)
	validateInfrastructure(s, r)

	return r
}

func validatePopulation(s *spec.CitySpec, r *Report) {
	if s.City.Population <= 0 {
		r.AddError(Result{
			Level:       LevelSchema,
			Message:     "population must be greater than 0",
			SpecPath:    "city.population",
			ActualValue: s.City.Population,
			Expected:    "> 0",
		})
	}
}

func validateDemographics(s *spec.CitySpec, r *Report) {
	d := s.Demographics
	ratios := map[string]float64{
		"singles":        d.Singles,
		"couples":        d.Couples,
		"families_young": d.FamiliesYoung,
		"families_teen":  d.FamiliesTeen,
		"empty_nest":     d.EmptyNest,
		"retirees":       d.Retirees,
	}

	for name, ratio := range ratios {
		if ratio < 0 {
			r.AddError(Result{
				Level:       LevelSchema,
				Message:     fmt.Sprintf("demographics.%s must be non-negative", name),
				SpecPath:    fmt.Sprintf("demographics.%s", name),
				ActualValue: ratio,
				Expected:    ">= 0",
			})
		}
	}

	sum := d.Singles + d.Couples + d.FamiliesYoung + d.FamiliesTeen + d.EmptyNest + d.Retirees
	if math.Abs(sum-1.0) > 0.01 {
		r.AddError(Result{
			Level:       LevelSchema,
			Message:     fmt.Sprintf("demographics ratios must sum to 1.0 (got %.4f)", sum),
			SpecPath:    "demographics",
			ActualValue: sum,
			Expected:    "1.0 (±0.01)",
			Suggestions: []string{"Adjust cohort ratios so they sum to 1.0"},
		})
	}
}

func validateZones(s *spec.CitySpec, r *Report) {
	rings := s.CityZones.Rings

	if len(rings) == 0 {
		r.AddError(Result{
			Level:    LevelSchema,
			Message:  "city_zones.rings must contain at least one ring",
			SpecPath: "city_zones.rings",
			Expected: "at least 1 ring",
		})
		return
	}

	for i, ring := range rings {
		if ring.RadiusFrom >= ring.RadiusTo {
			r.AddError(Result{
				Level:       LevelSchema,
				Message:     fmt.Sprintf("city_zones.rings[%d] (%s): radius_from (%.0f) must be less than radius_to (%.0f)", i, ring.Name, ring.RadiusFrom, ring.RadiusTo),
				SpecPath:    fmt.Sprintf("city_zones.rings[%d]", i),
				ActualValue: fmt.Sprintf("%.0f-%.0f", ring.RadiusFrom, ring.RadiusTo),
			})
		}

		if ring.MaxStories <= 0 {
			r.AddError(Result{
				Level:       LevelSchema,
				Message:     fmt.Sprintf("city_zones.rings[%d] (%s): max_stories must be > 0", i, ring.Name),
				SpecPath:    fmt.Sprintf("city_zones.rings[%d].max_stories", i),
				ActualValue: ring.MaxStories,
				Expected:    "> 0",
			})
		}
	}

	// Continuity: each ring's radius_to must equal the next ring's radius_from.
	for i := 0; i < len(rings)-1; i++ {
		if rings[i].RadiusTo != rings[i+1].RadiusFrom {
			r.AddError(Result{
				Level:       LevelSchema,
				Message:     fmt.Sprintf("zone gap: %s ends at %.0fm but %s starts at %.0fm", rings[i].Name, rings[i].RadiusTo, rings[i+1].Name, rings[i+1].RadiusFrom),
				SpecPath:    fmt.Sprintf("city_zones.rings[%d].radius_from", i+1),
				ActualValue: rings[i+1].RadiusFrom,
				Expected:    fmt.Sprintf("%.0f (matching %s.radius_to)", rings[i].RadiusTo, rings[i].Name),
			})
		}
	}
}

func validatePods(s *spec.CitySpec, r *Report) {
	if s.Pods.WalkRadius < 200 || s.Pods.WalkRadius > 800 {
		r.AddError(Result{
			Level:       LevelSchema,
			Message:     fmt.Sprintf("walk_radius %.0f is outside valid range (200-800m)", s.Pods.WalkRadius),
			SpecPath:    "pods.walk_radius",
			ActualValue: s.Pods.WalkRadius,
			Expected:    "200-800",
		})
	}
}

func validateCity(s *spec.CitySpec, r *Report) {
	if s.City.ExcavationDepth < 4 || s.City.ExcavationDepth > 15 {
		r.AddError(Result{
			Level:       LevelSchema,
			Message:     fmt.Sprintf("excavation_depth %.1f is outside valid range (4-15m)", s.City.ExcavationDepth),
			SpecPath:    "city.excavation_depth",
			ActualValue: s.City.ExcavationDepth,
			Expected:    "4-15",
		})
	}
}

func validateRevenue(s *spec.CitySpec, r *Report) {
	if s.Revenue.DebtTermYears <= 0 {
		r.AddError(Result{
			Level:       LevelSchema,
			Message:     "debt_term_years must be > 0",
			SpecPath:    "revenue.debt_term_years",
			ActualValue: s.Revenue.DebtTermYears,
			Expected:    "> 0",
		})
	}
	if s.Revenue.InterestRate < 0 || s.Revenue.InterestRate >= 1 {
		r.AddError(Result{
			Level:       LevelSchema,
			Message:     fmt.Sprintf("interest_rate %.4f must be >= 0 and < 1", s.Revenue.InterestRate),
			SpecPath:    "revenue.interest_rate",
			ActualValue: s.Revenue.InterestRate,
			Expected:    "0 <= rate < 1",
		})
	}
}

func validateInfrastructure(s *spec.CitySpec, r *Report) {
	if s.Infrastructure.Water.CapacityGPDPer <= 0 {
		r.AddError(Result{
			Level:    LevelSchema,
			Message:  "water capacity_gpd_per_capita must be > 0",
			SpecPath: "infrastructure.water.capacity_gpd_per_capita",
		})
	}
	if s.Infrastructure.Sewage.CapacityGPDPer <= 0 {
		r.AddError(Result{
			Level:    LevelSchema,
			Message:  "sewage capacity_gpd_per_capita must be > 0",
			SpecPath: "infrastructure.sewage.capacity_gpd_per_capita",
		})
	}
	if s.Infrastructure.Electrical.PeakDemandKWPer <= 0 {
		r.AddError(Result{
			Level:    LevelSchema,
			Message:  "peak_demand_kw_per_capita must be > 0",
			SpecPath: "infrastructure.electrical.peak_demand_kw_per_capita",
		})
	}
}
