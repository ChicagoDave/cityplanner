package layout

import "github.com/ChicagoDave/cityplanner/pkg/spec"

// CollectGreenZones re-computes zone allocation for all pods and returns
// only the green zones. Used by scene graph assembly for park entities.
func CollectGreenZones(s *spec.CitySpec, pods []Pod) []Zone {
	ringRadii := map[string][2]float64{
		"center": {s.CityZones.Center.RadiusFrom, s.CityZones.Center.RadiusTo},
		"middle": {s.CityZones.Middle.RadiusFrom, s.CityZones.Middle.RadiusTo},
		"edge":   {s.CityZones.Edge.RadiusFrom, s.CityZones.Edge.RadiusTo},
	}

	var greens []Zone
	for _, pod := range pods {
		ringChar := ""
		if pr, ok := s.Pods.RingAssignments[pod.Ring]; ok {
			ringChar = pr.Character
		}
		radii := ringRadii[pod.Ring]
		zones := AllocateZones(pod, ringChar, radii[0], radii[1])
		for _, z := range zones {
			if z.Type == ZoneGreen {
				greens = append(greens, z)
			}
		}
	}
	return greens
}
