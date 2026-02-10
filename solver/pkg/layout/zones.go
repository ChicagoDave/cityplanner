package layout

import (
	"fmt"

	"github.com/ChicagoDave/cityplanner/pkg/geo"
)

// ZoneType identifies the functional zone type within a pod.
type ZoneType string

const (
	ZoneResidential ZoneType = "residential"
	ZoneCommercial  ZoneType = "commercial"
	ZoneCivic       ZoneType = "civic"
	ZoneGreen       ZoneType = "green"
)

// Zone represents a functional zone within a pod.
type Zone struct {
	ID      string      `json:"id"`
	PodID   string      `json:"pod_id"`
	Type    ZoneType    `json:"type"`
	Polygon geo.Polygon `json:"polygon"`
	AreaHa  float64     `json:"area_ha"`
}

// zoneProportions returns the land use fractions for a given ring character.
func zoneProportions(ringChar string) map[ZoneType]float64 {
	switch ringChar {
	case "civic_commercial":
		return map[ZoneType]float64{
			ZoneResidential: 0.25,
			ZoneCommercial:  0.35,
			ZoneCivic:       0.25,
			ZoneGreen:       0.15,
		}
	default:
		return map[ZoneType]float64{
			ZoneResidential: 0.60,
			ZoneCommercial:  0.15,
			ZoneCivic:       0.10,
			ZoneGreen:       0.15,
		}
	}
}

// AllocateZones divides a pod into functional zones using concentric radial
// bands measured from the city center (origin).
//
// Band order from inner (nearest city center) to outer (city edge):
//
//	commercial → civic → residential → green
//
// Uses ring inner/outer radii from the spec so that band positions are
// independent of the pod polygon shape. Each band is clipped to the pod
// boundary via annulus clipping.
func AllocateZones(pod Pod, ringChar string, ringInnerR, ringOuterR float64) []Zone {
	podPoly := pod.BoundaryPolygon()
	if podPoly.IsEmpty() {
		return nil
	}

	extent := ringOuterR - ringInnerR
	if extent < 1 {
		return nil
	}

	proportions := zoneProportions(ringChar)

	commFrac := proportions[ZoneCommercial]
	civicFrac := proportions[ZoneCivic]
	greenFrac := proportions[ZoneGreen]

	cut1 := ringInnerR + commFrac*extent  // commercial → civic
	cut2 := cut1 + civicFrac*extent       // civic → residential
	cut3 := ringOuterR - greenFrac*extent // residential → green

	// Ensure proper ordering.
	if cut2 > cut3 {
		cut2 = (cut1 + cut3) / 2
	}

	type band struct {
		zoneType ZoneType
		innerR   float64
		outerR   float64
	}
	bands := []band{
		{ZoneCommercial, ringInnerR, cut1},
		{ZoneCivic, cut1, cut2},
		{ZoneResidential, cut2, cut3},
		{ZoneGreen, cut3, ringOuterR},
	}

	var zones []Zone
	for _, b := range bands {
		// Clip pod polygon to the radial band (annular region).
		zonePoly := geo.ClipToAnnulus(podPoly, geo.Origin, b.innerR, b.outerR)
		if zonePoly.IsEmpty() {
			continue
		}

		area := zonePoly.Area()
		if area < 100 { // skip tiny slivers (< 100 m²)
			continue
		}

		zones = append(zones, Zone{
			ID:      fmt.Sprintf("%s_%s", pod.ID, b.zoneType),
			PodID:   pod.ID,
			Type:    b.zoneType,
			Polygon: zonePoly,
			AreaHa:  area / 10000,
		})
	}

	return zones
}
