package layout

import (
	"fmt"
	"math"
	"sort"

	"github.com/ChicagoDave/cityplanner/pkg/geo"
	"github.com/ChicagoDave/cityplanner/pkg/spec"
	"github.com/ChicagoDave/cityplanner/pkg/validation"
)

// SportsField represents a placed sports facility.
type SportsField struct {
	ID         string     `json:"id"`
	Type       string     `json:"type"` // "stadium", "soccer", "basketball", "tennis", "pickleball"
	Position   geo.Point2D `json:"position"`
	Dimensions [2]float64 `json:"dimensions"` // [length, width] in meters
	Rotation   float64    `json:"rotation"`   // radians
	BufferID   string     `json:"buffer_id,omitempty"`
}

// BufferZone represents the green space between two adjacent pods.
type BufferZone struct {
	ID       string     `json:"id"`
	PodIDs   [2]string  `json:"pod_ids"`
	Centroid geo.Point2D `json:"centroid"`
	Length   float64    `json:"length"`
	Width    float64    `json:"width"`
	Rotation float64    `json:"rotation"` // angle of long axis
	Ring     string     `json:"ring"`
}

// PlaceSportsFields generates sports facilities in inter-pod buffer zones.
// Places 1 stadium, up to 10 soccer/cricket fields, and small courts.
func PlaceSportsFields(pods []Pod, adjacency map[string][]string, rings []spec.RingDef) ([]SportsField, *validation.Report) {
	report := validation.NewReport()

	buffers := IdentifyBufferZones(pods, adjacency)
	if len(buffers) == 0 {
		report.AddWarning(validation.Result{
			Level:   validation.LevelSpatial,
			Message: "no buffer zones identified for sports field placement",
		})
		return nil, report
	}

	consumed := make(map[string]bool)
	var fields []SportsField

	// 1. Place stadium near center (ring3/4 boundary).
	stadium, bufID, ok := placeStadium(buffers, rings)
	if ok {
		fields = append(fields, stadium)
		consumed[bufID] = true
	} else {
		report.AddWarning(validation.Result{
			Level:   validation.LevelSpatial,
			Message: "no buffer zone large enough for stadium (110x75m)",
		})
	}

	// 2. Place soccer/cricket fields.
	soccerFields := placeSoccerFields(buffers, consumed, len(fields))
	fields = append(fields, soccerFields...)

	// 3. Place small courts in remaining buffers.
	courts := placeSmallCourts(buffers, consumed, len(fields))
	fields = append(fields, courts...)

	report.AddInfo(validation.Result{
		Level: validation.LevelSpatial,
		Message: fmt.Sprintf("placed %d sports facilities (stadium=%v, soccer=%d, courts=%d)",
			len(fields), ok, len(soccerFields), len(courts)),
	})

	return fields, report
}

// IdentifyBufferZones finds the green space between adjacent pods.
func IdentifyBufferZones(pods []Pod, adjacency map[string][]string) []BufferZone {
	podMap := make(map[string]Pod, len(pods))
	for _, p := range pods {
		podMap[p.ID] = p
	}

	// Track processed pairs to avoid duplicates.
	seen := make(map[string]bool)
	var buffers []BufferZone
	bufIdx := 0

	for podID, neighbors := range adjacency {
		pod1, ok1 := podMap[podID]
		if !ok1 {
			continue
		}

		for _, adjID := range neighbors {
			// Create canonical pair key.
			pairKey := podID + "|" + adjID
			if podID > adjID {
				pairKey = adjID + "|" + podID
			}
			if seen[pairKey] {
				continue
			}
			seen[pairKey] = true

			pod2, ok2 := podMap[adjID]
			if !ok2 {
				continue
			}

			c1 := pod1.CenterPoint()
			c2 := pod2.CenterPoint()
			mid := geo.MidPoint(c1, c2)
			dir := c2.Sub(c1)
			dist := dir.Length()
			if dist < 1 {
				continue
			}

			// Buffer is oriented perpendicular to the line connecting pod centers.
			angle := math.Atan2(dir.Z, dir.X)
			// Estimate buffer dimensions: width = gap between pods (~50m),
			// length = perpendicular extent based on pod sizes.
			bufferWidth := 50.0
			bufferLength := dist * 0.4 // approximate perpendicular extent
			if bufferLength < 60 {
				bufferLength = 60
			}
			if bufferLength > 200 {
				bufferLength = 200
			}

			ring := pod1.Ring
			if pod2.Ring != pod1.Ring {
				ring = pod1.Ring + "/" + pod2.Ring
			}

			buffers = append(buffers, BufferZone{
				ID:       fmt.Sprintf("buffer_%d", bufIdx),
				PodIDs:   [2]string{podID, adjID},
				Centroid: mid,
				Length:   bufferLength,
				Width:    bufferWidth,
				Rotation: angle,
				Ring:     ring,
			})
			bufIdx++
		}
	}

	return buffers
}

// placeStadium places the stadium near the ring3/4 boundary in the largest
// buffer zone that fits. Stadium: 110m x 75m.
func placeStadium(buffers []BufferZone, rings []spec.RingDef) (SportsField, string, bool) {
	stadiumL, stadiumW := 110.0, 75.0

	// Find the target radius: ring3/ring4 boundary or inner rings.
	targetRadius := 0.0
	if len(rings) >= 3 {
		targetRadius = rings[2].RadiusFrom // ring3 inner boundary
	} else if len(rings) >= 2 {
		targetRadius = rings[1].RadiusFrom
	}

	// Score buffers by proximity to target and size.
	type scored struct {
		idx   int
		score float64
	}
	var candidates []scored
	for i, buf := range buffers {
		if buf.Length < stadiumL || buf.Width < stadiumW {
			// Check rotated fit.
			if buf.Length < stadiumW || buf.Width < stadiumL {
				continue
			}
		}
		dist := buf.Centroid.Length()
		proximity := math.Abs(dist - targetRadius)
		candidates = append(candidates, scored{i, proximity})
	}

	if len(candidates) == 0 {
		return SportsField{}, "", false
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].score < candidates[j].score
	})

	buf := buffers[candidates[0].idx]
	return SportsField{
		ID:         "stadium_0",
		Type:       "stadium",
		Position:   buf.Centroid,
		Dimensions: [2]float64{stadiumL, stadiumW},
		Rotation:   buf.Rotation,
		BufferID:   buf.ID,
	}, buf.ID, true
}

// placeSoccerFields places up to 10 soccer/cricket fields (105x68m) in
// available buffer zones, preferring outer rings.
func placeSoccerFields(buffers []BufferZone, consumed map[string]bool, startIdx int) []SportsField {
	fieldL, fieldW := 105.0, 68.0
	maxFields := 10

	// Sort buffers by distance from center (descending = prefer outer).
	type scored struct {
		idx  int
		dist float64
	}
	var candidates []scored
	for i, buf := range buffers {
		if consumed[buf.ID] {
			continue
		}
		fits := (buf.Length >= fieldL && buf.Width >= fieldW) ||
			(buf.Length >= fieldW && buf.Width >= fieldL)
		if !fits {
			continue
		}
		candidates = append(candidates, scored{i, buf.Centroid.Length()})
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].dist > candidates[j].dist
	})

	var fields []SportsField
	for _, c := range candidates {
		if len(fields) >= maxFields {
			break
		}
		buf := buffers[c.idx]
		consumed[buf.ID] = true
		fields = append(fields, SportsField{
			ID:         fmt.Sprintf("soccer_%d", startIdx+len(fields)),
			Type:       "soccer",
			Position:   buf.Centroid,
			Dimensions: [2]float64{fieldL, fieldW},
			Rotation:   buf.Rotation,
			BufferID:   buf.ID,
		})
	}

	return fields
}

// placeSmallCourts packs basketball (28x15m), tennis (24x12m), and pickleball
// (13x6m) courts into remaining buffer zones.
func placeSmallCourts(buffers []BufferZone, consumed map[string]bool, startIdx int) []SportsField {
	type courtType struct {
		name   string
		length float64
		width  float64
	}
	courtTypes := []courtType{
		{"basketball", 28, 15},
		{"tennis", 24, 12},
		{"pickleball", 13, 6},
	}

	var fields []SportsField
	courtIdx := startIdx
	typeIdx := 0

	for _, buf := range buffers {
		if consumed[buf.ID] {
			continue
		}

		// Try to fit one court type per remaining buffer.
		ct := courtTypes[typeIdx%len(courtTypes)]
		fits := (buf.Length >= ct.length && buf.Width >= ct.width) ||
			(buf.Length >= ct.width && buf.Width >= ct.length)
		if !fits {
			// Try smaller court.
			placed := false
			for _, alt := range courtTypes {
				fits = (buf.Length >= alt.length && buf.Width >= alt.width) ||
					(buf.Length >= alt.width && buf.Width >= alt.length)
				if fits {
					ct = alt
					placed = true
					break
				}
			}
			if !placed {
				continue
			}
		}

		consumed[buf.ID] = true
		fields = append(fields, SportsField{
			ID:         fmt.Sprintf("court_%s_%d", ct.name, courtIdx),
			Type:       ct.name,
			Position:   buf.Centroid,
			Dimensions: [2]float64{ct.length, ct.width},
			Rotation:   buf.Rotation,
			BufferID:   buf.ID,
		})
		courtIdx++
		typeIdx++
	}

	return fields
}
