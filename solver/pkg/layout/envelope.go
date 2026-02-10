package layout

import (
	"math"

	"github.com/ChicagoDave/cityplanner/pkg/spec"
)

// MaxStoriesFromRings computes the maximum allowed building height at a given
// distance from the city center, using ring definitions for the height envelope.
//
// For distances within a ring, returns that ring's max_stories.
// For distances between ring boundaries, linearly interpolates.
// For distances beyond all rings, returns the outermost ring's max_stories.
func MaxStoriesFromRings(distFromCenter float64, rings []spec.RingDef) int {
	if len(rings) == 0 {
		return 1
	}

	// Find which ring the distance falls in.
	for _, ring := range rings {
		if distFromCenter >= ring.RadiusFrom && distFromCenter <= ring.RadiusTo {
			return ring.MaxStories
		}
	}

	// Between rings: interpolate between the two adjacent rings.
	for i := 0; i < len(rings)-1; i++ {
		if distFromCenter > rings[i].RadiusTo && distFromCenter < rings[i+1].RadiusFrom {
			t := (distFromCenter - rings[i].RadiusTo) / (rings[i+1].RadiusFrom - rings[i].RadiusTo)
			stories := float64(rings[i].MaxStories) + t*(float64(rings[i+1].MaxStories)-float64(rings[i].MaxStories))
			return int(math.Max(1, math.Floor(stories)))
		}
	}

	// Beyond all rings: use outermost ring's value.
	return rings[len(rings)-1].MaxStories
}

// MaxStories computes the maximum allowed building height using legacy
// center/middle/edge parameters. Kept for backward compatibility with tests.
func MaxStories(distFromCenter float64, maxCenter, maxMiddle, maxEdge int) int {
	rings := []spec.RingDef{
		{Name: "center", RadiusFrom: 0, RadiusTo: 300, MaxStories: maxCenter},
		{Name: "middle", RadiusFrom: 300, RadiusTo: 600, MaxStories: maxMiddle},
		{Name: "edge", RadiusFrom: 600, RadiusTo: 900, MaxStories: maxEdge},
	}
	return MaxStoriesFromRings(distFromCenter, rings)
}

// MaxStoriesFromSpec computes maximum stories using the zone definitions
// from the city spec. Uses linear interpolation between ring boundaries.
func MaxStoriesFromSpec(distFromCenter float64, centerMax, middleMax, edgeMax int) int {
	return MaxStories(distFromCenter, centerMax, middleMax, edgeMax)
}
