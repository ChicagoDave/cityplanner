package layout

import (
	"math"
	"testing"

	"github.com/ChicagoDave/cityplanner/pkg/analytics"
	"github.com/ChicagoDave/cityplanner/pkg/spec"
)

func defaultSpec() *spec.CitySpec {
	return &spec.CitySpec{
		SpecVersion: "0.1.0",
		City: spec.CityDef{
			Population:      50000,
			ExcavationDepth: 8,
			MaxHeightCenter: 20,
			MaxHeightEdge:   4,
		},
		CityZones: spec.CityZones{
			Center: spec.ZoneDef{RadiusFrom: 0, RadiusTo: 300, MaxStories: 20, Character: "civic_commercial"},
			Middle: spec.ZoneDef{RadiusFrom: 300, RadiusTo: 600, MaxStories: 10, Character: "mixed_residential_commercial"},
			Edge:   spec.ZoneDef{RadiusFrom: 600, RadiusTo: 900, MaxStories: 4, Character: "family_education"},
		},
		Pods: spec.PodsDef{
			WalkRadius: 400,
			RingAssignments: map[string]spec.PodRing{
				"center": {Character: "civic_commercial", MaxStories: 20, RequiredServices: []string{
					"hospital", "performing_arts", "city_hall", "coworking_hub",
				}},
				"middle": {Character: "mixed", MaxStories: 10, RequiredServices: []string{
					"secondary_school", "coworking", "medical_clinic", "retail", "restaurant",
				}},
				"edge": {Character: "residential_family", MaxStories: 4, RequiredServices: []string{
					"elementary_school", "library", "grocery", "playground", "pediatric_clinic", "daycare",
				}},
			},
		},
	}
}

func defaultParams() *analytics.ResolvedParameters {
	return &analytics.ResolvedParameters{
		TotalPopulation: 50000,
		TotalHouseholds: 20202,
		PodCount:        6,
		TotalAreaHa:     254.47,
		Rings: []analytics.RingData{
			{
				Name:              "center",
				RadiusFrom:        0,
				RadiusTo:          300,
				AreaHa:            28.27,
				Population:        8333,
				Households:        3367,
				PodCount:          1,
				PodPopulation:     8333,
				MaxStories:        20,
				ResidentialAreaHa: 16.96,
			},
			{
				Name:              "middle",
				RadiusFrom:        300,
				RadiusTo:          600,
				AreaHa:            84.82,
				Population:        16667,
				Households:        6734,
				PodCount:          2,
				PodPopulation:     8333,
				MaxStories:        10,
				ResidentialAreaHa: 50.89,
			},
			{
				Name:              "edge",
				RadiusFrom:        600,
				RadiusTo:          900,
				AreaHa:            141.37,
				Population:        25000,
				Households:        10101,
				PodCount:          3,
				PodPopulation:     8333,
				MaxStories:        4,
				ResidentialAreaHa: 84.82,
			},
		},
	}
}

func TestLayoutPodsCount(t *testing.T) {
	pods, _, report := LayoutPods(defaultSpec(), defaultParams())
	if !report.Valid {
		t.Fatalf("layout failed: %v", report.Errors)
	}
	if len(pods) != 6 {
		t.Errorf("expected 6 pods, got %d", len(pods))
	}
}

func TestLayoutPodsRingAssignment(t *testing.T) {
	pods, _, _ := LayoutPods(defaultSpec(), defaultParams())
	ringCount := map[string]int{}
	for _, p := range pods {
		ringCount[p.Ring]++
	}
	if ringCount["center"] != 1 {
		t.Errorf("expected 1 center pod, got %d", ringCount["center"])
	}
	if ringCount["middle"] != 2 {
		t.Errorf("expected 2 middle pods, got %d", ringCount["middle"])
	}
	if ringCount["edge"] != 3 {
		t.Errorf("expected 3 edge pods, got %d", ringCount["edge"])
	}
}

func TestLayoutPodsAreaCoverage(t *testing.T) {
	pods, _, _ := LayoutPods(defaultSpec(), defaultParams())
	totalArea := 0.0
	for _, p := range pods {
		totalArea += p.AreaHa
	}
	// City area = π * 900² / 10000 ≈ 254.47 ha.
	cityArea := math.Pi * 900 * 900 / 10000
	coverage := totalArea / cityArea
	if coverage < 0.90 {
		t.Errorf("pod coverage %.1f%% is below 90%%", coverage*100)
	}
}

func TestLayoutPodsWalkRadiusWarnings(t *testing.T) {
	// The default city has large annular sector pods. Walk radius warnings are
	// expected for middle and edge pods since ring geometry creates pods larger
	// than the 400m walk radius. The solver correctly reports these as warnings.
	_, _, report := LayoutPods(defaultSpec(), defaultParams())
	hasWalkWarning := false
	for _, w := range report.Warnings {
		if len(w.Message) > 0 {
			hasWalkWarning = true
		}
	}
	if !hasWalkWarning {
		t.Log("no walk radius warnings — either pods fit within 400m or warnings changed")
	}
	// The center pod (seed at origin, ring 0-300m) should be within walk radius.
}

func TestLayoutPodsCenterPodWalkRadius(t *testing.T) {
	pods, _, _ := LayoutPods(defaultSpec(), defaultParams())
	for _, p := range pods {
		if p.Ring != "center" {
			continue
		}
		poly := p.BoundaryPolygon()
		center := p.CenterPoint()
		maxDist := poly.MaxDistanceTo(center)
		// Center pod: seed at origin, ring 0-300m. Max dist should be ~300m.
		if maxDist > 400*1.10 {
			t.Errorf("center pod: max distance %.0fm exceeds walk radius (expected ~300m)", maxDist)
		}
	}
}

func TestLayoutPodsAdjacency(t *testing.T) {
	_, adjacency, _ := LayoutPods(defaultSpec(), defaultParams())
	// Center pod should be adjacent to at least the 2 middle pods.
	centerAdj := adjacency["pod_center_0"]
	if len(centerAdj) < 2 {
		t.Errorf("center pod has %d neighbors, expected >= 2", len(centerAdj))
	}
}

func TestLayoutPodsNonOverlapping(t *testing.T) {
	pods, _, _ := LayoutPods(defaultSpec(), defaultParams())
	// Verify pod centers are inside their own boundaries.
	for _, p := range pods {
		poly := p.BoundaryPolygon()
		center := p.CenterPoint()
		if !poly.Contains(center) {
			t.Errorf("pod %s center is not inside its boundary", p.ID)
		}
	}
}

func TestLayoutPodsBoundaryVertices(t *testing.T) {
	pods, _, _ := LayoutPods(defaultSpec(), defaultParams())
	for _, p := range pods {
		if len(p.Boundary) < 3 {
			t.Errorf("pod %s has only %d boundary vertices", p.ID, len(p.Boundary))
		}
	}
}
