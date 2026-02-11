package layout

import (
	"fmt"
	"math"

	"github.com/ChicagoDave/cityplanner/pkg/geo"
	"github.com/ChicagoDave/cityplanner/pkg/spec"
	"github.com/ChicagoDave/cityplanner/pkg/validation"
)

// Plaza represents a hard-paved public gathering space at a pod center.
type Plaza struct {
	ID       string      `json:"id"`
	PodID    string      `json:"pod_id"`
	Position geo.Point2D `json:"position"`
	Width    float64     `json:"width"`    // X dimension in meters
	Depth    float64     `json:"depth"`    // Z dimension in meters
	Rotation float64     `json:"rotation"` // radians
	RingChar string      `json:"ring_character"`
}

// GeneratePlazas creates one plaza per pod, placed at the pod center.
// Size varies by ring character.
func GeneratePlazas(pods []Pod, s *spec.CitySpec) ([]Plaza, *validation.Report) {
	report := validation.NewReport()

	var plazas []Plaza
	for _, pod := range pods {
		ringChar := ""
		if pr, ok := s.Pods.RingAssignments[pod.Ring]; ok {
			ringChar = pr.Character
		}

		size := plazaSize(ringChar)

		dist := math.Hypot(pod.Center[0], pod.Center[1])
		rotation := 0.0
		if dist > 1 {
			rotation = math.Atan2(pod.Center[1], pod.Center[0])
		}

		plazas = append(plazas, Plaza{
			ID:       fmt.Sprintf("plaza_%s", pod.ID),
			PodID:    pod.ID,
			Position: geo.Point2D{X: pod.Center[0], Z: pod.Center[1]},
			Width:    size,
			Depth:    size,
			Rotation: rotation,
			RingChar: ringChar,
		})
	}

	report.AddInfo(validation.Result{
		Level:   validation.LevelSpatial,
		Message: fmt.Sprintf("generated %d plazas", len(plazas)),
	})
	return plazas, report
}

// plazaSize returns the plaza side length in meters for a given ring character.
func plazaSize(ringChar string) float64 {
	switch ringChar {
	case "civic_commercial":
		return 40
	case "high_density":
		return 30
	case "urban_midrise", "mixed_residential":
		return 25
	case "low_density":
		return 20
	default:
		return 25
	}
}
