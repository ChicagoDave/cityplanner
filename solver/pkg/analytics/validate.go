package analytics

import (
	"fmt"
	"math"

	"github.com/ChicagoDave/cityplanner/pkg/spec"
	"github.com/ChicagoDave/cityplanner/pkg/validation"
)

// validateAnalytical runs Phase 1 analytical validation checks.
func validateAnalytical(s *spec.CitySpec, p *ResolvedParameters, report *validation.Report) {
	validateDensityFeasibility(p, report)
	validatePodServices(s, p, report)
	validateSiteArea(s, p, report)
	validateEnergyBalance(s, p, report)
	validateBatteryBackup(s, p, report)
	validateDependencyRatio(p, report)
}

func validateDensityFeasibility(p *ResolvedParameters, report *validation.Report) {
	for _, ring := range p.Rings {
		if ring.RequiredDensity > ring.AchievableDensity {
			neededStories := int(math.Ceil(ring.RequiredDensity * avgUnitSizeM2 / (groundCoverage * m2PerHa)))
			report.AddError(validation.Result{
				Level:       validation.LevelAnalytical,
				Message:     fmt.Sprintf("%s ring: required density %.0f du/ha exceeds achievable %.0f du/ha at %d stories", ring.Name, ring.RequiredDensity, ring.AchievableDensity, ring.MaxStories),
				SpecPath:    fmt.Sprintf("city_zones.%s.max_stories", ring.Name),
				ActualValue: ring.MaxStories,
				ConflictWith: fmt.Sprintf("required density %.0f du/ha from %d households in %.1f ha residential area",
					ring.RequiredDensity, ring.Households, ring.ResidentialAreaHa),
				Suggestions: []string{
					fmt.Sprintf("Increase %s max_stories to at least %d", ring.Name, neededStories),
					"Reduce population or increase zone radius",
				},
			})
		}
	}
}

func validatePodServices(s *spec.CitySpec, p *ResolvedParameters, report *validation.Report) {
	for _, ring := range p.Rings {
		podRing, exists := s.Pods.RingAssignments[ring.Name]
		if !exists {
			continue
		}
		for _, svc := range podRing.RequiredServices {
			st, ok := ServiceThresholds[svc]
			if !ok {
				continue
			}
			threshold := st.threshold
			if st.metric == "persons" && ring.PodPopulation < threshold {
				report.AddWarning(validation.Result{
					Level:       validation.LevelAnalytical,
					Message:     fmt.Sprintf("%s ring: pod population %d is below %s threshold of %d", ring.Name, ring.PodPopulation, svc, threshold),
					SpecPath:    fmt.Sprintf("pods.ring_assignments.%s.required_services", ring.Name),
					ActualValue: ring.PodPopulation,
					Expected:    fmt.Sprintf(">= %d for %s", threshold, svc),
					Suggestions: []string{
						"Adjacent pods may share this service",
						"Consider reducing walk_radius to increase pod count",
					},
				})
			}
		}
	}
}

func validateSiteArea(s *spec.CitySpec, p *ResolvedParameters, report *validation.Report) {
	if s.Site.MinAreaHa > 0 && p.Areas.TotalWithPerimeter > s.Site.MinAreaHa {
		report.AddWarning(validation.Result{
			Level:       validation.LevelAnalytical,
			Message:     fmt.Sprintf("computed total area (%.0f ha) exceeds site min_area_ha (%.0f ha)", p.Areas.TotalWithPerimeter, s.Site.MinAreaHa),
			SpecPath:    "site_requirements.min_area_ha",
			ActualValue: p.Areas.TotalWithPerimeter,
			Expected:    fmt.Sprintf("<= %.0f ha", s.Site.MinAreaHa),
			Suggestions: []string{
				fmt.Sprintf("Increase min_area_ha to at least %.0f", math.Ceil(p.Areas.TotalWithPerimeter)),
				"Reduce zone radii to decrease total city footprint",
			},
		})
	}
}

func validateEnergyBalance(s *spec.CitySpec, p *ResolvedParameters, report *validation.Report) {
	totalSupply := p.Energy.TotalGenerationMW + p.Energy.GridCapacityMW
	if p.Energy.PeakDemandMW > totalSupply {
		report.AddError(validation.Result{
			Level:       validation.LevelAnalytical,
			Message:     fmt.Sprintf("peak electrical demand %.0f MW exceeds total supply %.0f MW (generation %.0f + grid %.0f)", p.Energy.PeakDemandMW, totalSupply, p.Energy.TotalGenerationMW, p.Energy.GridCapacityMW),
			SpecPath:    "infrastructure.electrical",
			ActualValue: p.Energy.PeakDemandMW,
			Expected:    fmt.Sprintf("<= %.0f MW", totalSupply),
			Suggestions: []string{
				"Increase solar capacity or grid capacity",
				"Reduce peak_demand_kw_per_capita",
			},
		})
	}
}

func validateBatteryBackup(s *spec.CitySpec, p *ResolvedParameters, report *validation.Report) {
	if p.Energy.BackupHours < 24 {
		needed := p.Energy.PeakDemandMW * 24
		report.AddWarning(validation.Result{
			Level:       validation.LevelAnalytical,
			Message:     fmt.Sprintf("battery storage provides %.0f hours of backup (target: 24 hours)", p.Energy.BackupHours),
			SpecPath:    "infrastructure.electrical.battery_capacity_mwh",
			ActualValue: s.Infrastructure.Electrical.BatteryCapacityMWh,
			Expected:    fmt.Sprintf("%.0f MWh for 24-hour backup", needed),
			Suggestions: []string{fmt.Sprintf("Increase battery_capacity_mwh to %.0f", needed)},
		})
	}
}

func validateDependencyRatio(p *ResolvedParameters, report *validation.Report) {
	if p.DependencyRatio < 0.3 || p.DependencyRatio > 0.8 {
		report.AddError(validation.Result{
			Level:       validation.LevelAnalytical,
			Message:     fmt.Sprintf("dependency ratio %.2f is outside viable range (0.3-0.8)", p.DependencyRatio),
			SpecPath:    "demographics",
			ActualValue: p.DependencyRatio,
			Expected:    "0.3-0.8",
			Suggestions: []string{"Adjust demographic cohort ratios for a balanced population"},
		})
	} else if p.DependencyRatio < 0.5 || p.DependencyRatio > 0.6 {
		report.AddWarning(validation.Result{
			Level:       validation.LevelAnalytical,
			Message:     fmt.Sprintf("dependency ratio %.2f is outside ideal range (0.5-0.6)", p.DependencyRatio),
			SpecPath:    "demographics",
			ActualValue: p.DependencyRatio,
			Expected:    "0.5-0.6 (ideal)",
		})
	}
}
