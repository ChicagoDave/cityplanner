package layout

import "math"

// MaxStories computes the maximum allowed building height at a given distance
// from the city center, implementing the "bowl profile" height envelope.
//
// Height envelope (from ADR-006):
//
//	distance ≤ 300m:       maxCenter stories
//	300m < distance ≤ 600m: linear interpolation maxCenter → maxMiddle
//	600m < distance ≤ 900m: linear interpolation maxMiddle → maxEdge
//
// For the default city: center=20, middle=10, edge=4.
func MaxStories(distFromCenter float64, maxCenter, maxMiddle, maxEdge int) int {
	switch {
	case distFromCenter <= 300:
		return maxCenter
	case distFromCenter <= 600:
		t := (distFromCenter - 300) / 300
		stories := float64(maxCenter) + t*(float64(maxMiddle)-float64(maxCenter))
		return int(math.Max(1, math.Floor(stories)))
	case distFromCenter <= 900:
		t := (distFromCenter - 600) / 300
		stories := float64(maxMiddle) + t*(float64(maxEdge)-float64(maxMiddle))
		return int(math.Max(1, math.Floor(stories)))
	default:
		return maxEdge
	}
}

// MaxStoriesFromSpec computes maximum stories using the zone definitions
// from the city spec. Uses linear interpolation between ring boundaries.
func MaxStoriesFromSpec(distFromCenter float64, centerMax, middleMax, edgeMax int) int {
	return MaxStories(distFromCenter, centerMax, middleMax, edgeMax)
}
