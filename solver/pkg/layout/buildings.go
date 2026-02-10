package layout

import (
	"github.com/ChicagoDave/cityplanner/pkg/spec"
	"github.com/ChicagoDave/cityplanner/pkg/validation"
)

// Building represents a placed building within a pod.
type Building struct {
	ID         string     `json:"id"`
	PodID      string     `json:"pod_id"`
	Type       string     `json:"type"` // residential, commercial, civic, service
	Position   [3]float64 `json:"position"`
	Footprint  [2]float64 `json:"footprint"` // [width, depth] in meters
	Stories    int        `json:"stories"`
	DwellingUnits int    `json:"dwelling_units,omitempty"`
	CommercialSqM float64 `json:"commercial_sqm,omitempty"`
}

// PlaceBuildings generates building placements within laid-out pods.
func PlaceBuildings(_ *spec.CitySpec, _ []Pod) ([]Building, *validation.Report) {
	report := validation.NewReport()

	// TODO: Implement hierarchical decomposition (ADR-006)
	// 1. Zone allocation within each pod
	// 2. Path network generation
	// 3. Block subdivision
	// 4. Building placement per block
	// 5. Validate dwelling units and commercial sqft targets

	return []Building{}, report
}
