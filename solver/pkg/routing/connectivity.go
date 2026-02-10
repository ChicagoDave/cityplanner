package routing

import (
	"math"
	"sort"
)

const connectTolerance = 1.0 // meters

// BuildConnectivity post-processes segments to build a connectivity graph.
// Two segments are connected if they share a Start or End point within
// tolerance and are on the same layer. Returns a map of segment ID to
// connected segment IDs, guaranteed bidirectional.
func BuildConnectivity(segments []Segment) map[string][]string {
	type endpoint struct {
		segIdx int
		isEnd  bool // false=Start, true=End
	}

	cellSize := connectTolerance * 2
	buckets := make(map[[2]int][]endpoint)

	cellKey := func(x, z float64) [2]int {
		return [2]int{int(math.Floor(x / cellSize)), int(math.Floor(z / cellSize))}
	}

	// Index all endpoints into spatial grid cells (plus neighbors for tolerance).
	for i, seg := range segments {
		for _, isEnd := range []bool{false, true} {
			pt := seg.Start
			if isEnd {
				pt = seg.End
			}
			key := cellKey(pt[0], pt[2])
			ep := endpoint{segIdx: i, isEnd: isEnd}
			for dx := -1; dx <= 1; dx++ {
				for dz := -1; dz <= 1; dz++ {
					bk := [2]int{key[0] + dx, key[1] + dz}
					buckets[bk] = append(buckets[bk], ep)
				}
			}
		}
	}

	// For each segment, find connections at both endpoints.
	conn := make(map[string]map[string]bool)
	for i, seg := range segments {
		for _, isEnd := range []bool{false, true} {
			pt := seg.Start
			if isEnd {
				pt = seg.End
			}
			key := cellKey(pt[0], pt[2])
			for _, ep := range buckets[key] {
				if ep.segIdx == i {
					continue
				}
				other := segments[ep.segIdx]
				// Must share the same layer for physical connectivity.
				if other.Layer != seg.Layer {
					continue
				}
				otherPt := other.Start
				if ep.isEnd {
					otherPt = other.End
				}
				dist := math.Hypot(pt[0]-otherPt[0], pt[2]-otherPt[2])
				if dist <= connectTolerance {
					// Bidirectional: add both directions.
					if conn[seg.ID] == nil {
						conn[seg.ID] = make(map[string]bool)
					}
					conn[seg.ID][other.ID] = true
					if conn[other.ID] == nil {
						conn[other.ID] = make(map[string]bool)
					}
					conn[other.ID][seg.ID] = true
				}
			}
		}
	}

	// Convert sets to sorted slices for deterministic output.
	result := make(map[string][]string, len(conn))
	for id, neighbors := range conn {
		ids := make([]string, 0, len(neighbors))
		for nid := range neighbors {
			ids = append(ids, nid)
		}
		sort.Strings(ids)
		result[id] = ids
	}
	return result
}
