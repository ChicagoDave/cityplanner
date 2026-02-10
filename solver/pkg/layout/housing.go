package layout

import (
	"math"

	"github.com/ChicagoDave/cityplanner/pkg/analytics"
)

// UnitMix holds the count of dwelling units by bedroom count.
type UnitMix struct {
	Studios  int `json:"studios"`
	OneBed   int `json:"one_bed"`
	TwoBed   int `json:"two_bed"`
	ThreeBed int `json:"three_bed"`
	FourBed  int `json:"four_bed"`
}

// Total returns the sum of all unit types.
func (u UnitMix) Total() int {
	return u.Studios + u.OneBed + u.TwoBed + u.ThreeBed + u.FourBed
}

// DistributeUnits maps demographic cohorts to unit types and computes
// the required unit counts.
//
// Mapping:
//
//	singles      → studio
//	couples      → 1-bed
//	families_young → 3-bed
//	families_teen  → 4-bed
//	empty_nest   → 2-bed
//	retirees     → 1-bed
func DistributeUnits(totalHH int, cohorts []analytics.CohortBreakdown) UnitMix {
	mix := UnitMix{}
	for _, c := range cohorts {
		switch c.Name {
		case "singles":
			mix.Studios += c.Households
		case "couples":
			mix.OneBed += c.Households
		case "families_young":
			mix.ThreeBed += c.Households
		case "families_teen":
			mix.FourBed += c.Households
		case "empty_nest":
			mix.TwoBed += c.Households
		case "retirees":
			mix.OneBed += c.Households
		}
	}
	// Reconcile rounding: adjust studios to match total.
	diff := totalHH - mix.Total()
	mix.Studios += diff
	return mix
}

// ScaleUnitMix returns a UnitMix scaled to a fraction of the original.
func ScaleUnitMix(mix UnitMix, fraction float64) UnitMix {
	return UnitMix{
		Studios:  int(math.Round(float64(mix.Studios) * fraction)),
		OneBed:   int(math.Round(float64(mix.OneBed) * fraction)),
		TwoBed:   int(math.Round(float64(mix.TwoBed) * fraction)),
		ThreeBed: int(math.Round(float64(mix.ThreeBed) * fraction)),
		FourBed:  int(math.Round(float64(mix.FourBed) * fraction)),
	}
}

// AvgUnitSizeM2 returns the weighted average unit size in square meters.
// Studio: 35m², 1-bed: 50m², 2-bed: 75m², 3-bed: 100m², 4-bed: 120m².
func AvgUnitSizeM2(mix UnitMix) float64 {
	total := mix.Total()
	if total == 0 {
		return 75
	}
	weighted := float64(mix.Studios)*35 + float64(mix.OneBed)*50 +
		float64(mix.TwoBed)*75 + float64(mix.ThreeBed)*100 +
		float64(mix.FourBed)*120
	return weighted / float64(total)
}
