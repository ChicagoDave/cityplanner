package scene

import (
	"testing"
)

func validGraph() *Graph {
	g := NewGraph()
	g.Entities = []Entity{
		{
			ID:         "bld-1",
			Type:       EntityBuilding,
			Position:   Vec3{X: 10, Y: 0, Z: 20},
			Dimensions: Vec3{X: 5, Y: 12, Z: 5},
			Rotation:   [4]float64{0, 0, 0, 1},
			Material:   "concrete",
			Pod:        "pod-0",
			Layer:      LayerSurface,
		},
		{
			ID:         "pipe-1",
			Type:       EntityPipe,
			Position:   Vec3{X: 10, Y: -7, Z: 20},
			Dimensions: Vec3{X: 0.4, Y: 1.5, Z: 50},
			Rotation:   [4]float64{0, 0, 0, 1},
			Material:   "steel",
			System:     SystemWater,
			Layer:      LayerUnderground1,
		},
	}
	g.Groups.Pods["pod-0"] = []string{"bld-1"}
	g.Groups.Systems[SystemWater] = []string{"pipe-1"}
	g.Groups.Layers[LayerSurface] = []string{"bld-1"}
	g.Groups.Layers[LayerUnderground1] = []string{"pipe-1"}
	g.Groups.EntityTypes[EntityBuilding] = []string{"bld-1"}
	g.Groups.EntityTypes[EntityPipe] = []string{"pipe-1"}
	g.Metadata = Metadata{
		SpecVersion: "0.1.0",
		CityBounds: BoundingBox{
			Min: Vec3{X: -100, Y: -10, Z: -100},
			Max: Vec3{X: 100, Y: 50, Z: 100},
		},
	}
	return g
}

func TestValidateGraph_Valid(t *testing.T) {
	r := ValidateGraph(validGraph())
	if !r.Valid {
		t.Errorf("expected valid, got %d errors", len(r.Errors))
		for _, e := range r.Errors {
			t.Logf("  error: %s", e.Message)
		}
	}
}

func TestValidateGraph_Nil(t *testing.T) {
	r := ValidateGraph(nil)
	if r.Valid {
		t.Error("expected invalid for nil graph")
	}
}

func TestValidateGraph_DuplicateID(t *testing.T) {
	g := validGraph()
	g.Entities = append(g.Entities, Entity{
		ID:         "bld-1",
		Type:       EntityBuilding,
		Position:   Vec3{X: 20, Y: 0, Z: 30},
		Dimensions: Vec3{X: 5, Y: 9, Z: 5},
		Rotation:   [4]float64{0, 0, 0, 1},
		Layer:      LayerSurface,
	})
	r := ValidateGraph(g)
	if r.Valid {
		t.Error("expected invalid for duplicate ID")
	}
}

func TestValidateGraph_OrphanedGroupReference(t *testing.T) {
	g := validGraph()
	g.Groups.Pods["pod-0"] = append(g.Groups.Pods["pod-0"], "nonexistent")
	r := ValidateGraph(g)
	if r.Valid {
		t.Error("expected invalid for orphaned group reference")
	}
}

func TestValidateGraph_MissingGroupMembership(t *testing.T) {
	g := validGraph()
	g.Groups.Layers[LayerSurface] = []string{}
	r := ValidateGraph(g)
	if r.Valid {
		t.Error("expected invalid for missing group membership")
	}
}

func TestValidateGraph_EmptyID(t *testing.T) {
	g := validGraph()
	g.Entities = append(g.Entities, Entity{
		ID:         "",
		Type:       EntityPath,
		Dimensions: Vec3{X: 2, Y: 0.1, Z: 10},
		Rotation:   [4]float64{0, 0, 0, 1},
		Layer:      LayerSurface,
	})
	r := ValidateGraph(g)
	if r.Valid {
		t.Error("expected invalid for empty ID")
	}
}

func TestValidateGraph_ZeroDimensionWarning(t *testing.T) {
	g := validGraph()
	g.Entities[0].Dimensions.Y = 0
	r := ValidateGraph(g)
	if len(r.Warnings) == 0 {
		t.Error("expected warning for zero dimension")
	}
}

func TestValidateGraph_RealGraph(t *testing.T) {
	g := assembleTestGraph(t)
	r := ValidateGraph(g)
	if !r.Valid {
		t.Errorf("real graph validation failed: %d errors", len(r.Errors))
		for _, e := range r.Errors {
			t.Logf("  error: %s", e.Message)
		}
	}
	t.Logf("validated %d entities: %s", len(g.Entities), r.Summary)
}
