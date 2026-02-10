package layout

import "github.com/ChicagoDave/cityplanner/pkg/spec"

// CollectGreenZones re-computes zone allocation for all pods and returns
// only the green zones. Used by scene graph assembly for park entities.
func CollectGreenZones(s *spec.CitySpec, pods []Pod) []Zone {
	ringRadii := make(map[string][2]float64, len(s.CityZones.Rings))
	for _, ring := range s.CityZones.Rings {
		ringRadii[ring.Name] = [2]float64{ring.RadiusFrom, ring.RadiusTo}
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
