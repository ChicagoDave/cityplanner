package routing

import (
	"github.com/ChicagoDave/cityplanner/pkg/layout"
	"github.com/ChicagoDave/cityplanner/pkg/spec"
	"github.com/ChicagoDave/cityplanner/pkg/validation"
)

// NetworkType identifies an infrastructure network.
type NetworkType string

const (
	NetworkSewage     NetworkType = "sewage"
	NetworkWater      NetworkType = "water"
	NetworkElectrical NetworkType = "electrical"
	NetworkTelecom    NetworkType = "telecom"
	NetworkVehicle    NetworkType = "vehicle"
)

// Segment represents a routed infrastructure segment.
type Segment struct {
	ID        string      `json:"id"`
	Network   NetworkType `json:"network"`
	Layer     int         `json:"layer"` // 1=bottom, 2=middle, 3=top
	Start     [3]float64  `json:"start"`
	End       [3]float64  `json:"end"`
	WidthM    float64     `json:"width_m"`
	Capacity  float64     `json:"capacity"` // network-specific units
	IsTrunk   bool        `json:"is_trunk"`
}

// RouteInfrastructure generates all underground infrastructure routes.
func RouteInfrastructure(_ *spec.CitySpec, _ []layout.Pod, _ []layout.Building) ([]Segment, *validation.Report) {
	report := validation.NewReport()

	// TODO: Implement hierarchical trunk-and-branch routing (ADR-007)
	// Route in dependency order:
	// 1. Sewage (layer 1) — gravity-dependent, hardest constraint
	// 2. Water (layer 1) — pressurized, follows trunk backbone
	// 3. Electrical (layer 2) — from perimeter substation inward
	// 4. Telecom (layer 2) — fiber backbone with node breakouts
	// 5. Vehicle lanes (layer 3) — arterial + service branches

	return []Segment{}, report
}
