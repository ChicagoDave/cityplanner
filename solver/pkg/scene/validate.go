package scene

import (
	"fmt"

	"github.com/ChicagoDave/cityplanner/pkg/validation"
)

// ValidateGraph performs structural validation on a scene graph output.
// It checks entity integrity, group index consistency, and bounds enclosure.
func ValidateGraph(g *Graph) *validation.Report {
	r := validation.NewReport()

	if g == nil {
		r.AddError(validation.Result{
			Level:   validation.LevelSpatial,
			Message: "scene graph is nil",
		})
		return r
	}

	validateEntityIDs(g, r)
	validateGroupIndices(g, r)
	validateGroupMembership(g, r)
	validateBoundsEnclosure(g, r)
	validateEntityDimensions(g, r)

	return r
}

func validateEntityIDs(g *Graph, r *validation.Report) {
	seen := make(map[string]int, len(g.Entities))

	for i, e := range g.Entities {
		if e.ID == "" {
			r.AddError(validation.Result{
				Level:       validation.LevelSpatial,
				Message:     fmt.Sprintf("entity at index %d has empty ID", i),
				SpecPath:    fmt.Sprintf("entities[%d].id", i),
				ActualValue: "",
				Expected:    "non-empty string",
			})
			continue
		}
		if prev, exists := seen[e.ID]; exists {
			r.AddError(validation.Result{
				Level:       validation.LevelSpatial,
				Message:     fmt.Sprintf("duplicate entity ID %q at indices %d and %d", e.ID, prev, i),
				SpecPath:    fmt.Sprintf("entities[%d].id", i),
				ActualValue: e.ID,
			})
		}
		seen[e.ID] = i
	}
}

func validateGroupIndices(g *Graph, r *validation.Report) {
	entityIDs := make(map[string]bool, len(g.Entities))
	for _, e := range g.Entities {
		entityIDs[e.ID] = true
	}

	checkGroup := func(groupType, groupName string, ids []string) {
		for _, id := range ids {
			if !entityIDs[id] {
				r.AddError(validation.Result{
					Level:       validation.LevelSpatial,
					Message:     fmt.Sprintf("group %s.%s references non-existent entity %q", groupType, groupName, id),
					SpecPath:    fmt.Sprintf("groups.%s.%s", groupType, groupName),
					ActualValue: id,
					Expected:    "existing entity ID",
				})
			}
		}
	}

	for name, ids := range g.Groups.Pods {
		checkGroup("pods", name, ids)
	}
	for name, ids := range g.Groups.Systems {
		checkGroup("systems", string(name), ids)
	}
	for name, ids := range g.Groups.Layers {
		checkGroup("layers", string(name), ids)
	}
	for name, ids := range g.Groups.EntityTypes {
		checkGroup("entity_types", string(name), ids)
	}
}

func validateGroupMembership(g *Graph, r *validation.Report) {
	layerMembers := make(map[string]map[string]bool)
	for layer, ids := range g.Groups.Layers {
		m := make(map[string]bool, len(ids))
		for _, id := range ids {
			m[id] = true
		}
		layerMembers[string(layer)] = m
	}

	typeMembers := make(map[string]map[string]bool)
	for et, ids := range g.Groups.EntityTypes {
		m := make(map[string]bool, len(ids))
		for _, id := range ids {
			m[id] = true
		}
		typeMembers[string(et)] = m
	}

	systemMembers := make(map[string]map[string]bool)
	for sys, ids := range g.Groups.Systems {
		m := make(map[string]bool, len(ids))
		for _, id := range ids {
			m[id] = true
		}
		systemMembers[string(sys)] = m
	}

	podMembers := make(map[string]map[string]bool)
	for pod, ids := range g.Groups.Pods {
		m := make(map[string]bool, len(ids))
		for _, id := range ids {
			m[id] = true
		}
		podMembers[pod] = m
	}

	for _, e := range g.Entities {
		if e.ID == "" {
			continue
		}

		// Check layer group
		if lm, ok := layerMembers[string(e.Layer)]; ok {
			if !lm[e.ID] {
				r.AddError(validation.Result{
					Level:       validation.LevelSpatial,
					Message:     fmt.Sprintf("entity %q has layer %q but is not in layers group", e.ID, e.Layer),
					SpecPath:    fmt.Sprintf("groups.layers.%s", e.Layer),
					ActualValue: e.ID,
				})
			}
		} else if e.Layer != "" {
			r.AddError(validation.Result{
				Level:       validation.LevelSpatial,
				Message:     fmt.Sprintf("entity %q has layer %q but no such layer group exists", e.ID, e.Layer),
				SpecPath:    "groups.layers",
				ActualValue: string(e.Layer),
			})
		}

		// Check entity_type group
		if tm, ok := typeMembers[string(e.Type)]; ok {
			if !tm[e.ID] {
				r.AddError(validation.Result{
					Level:       validation.LevelSpatial,
					Message:     fmt.Sprintf("entity %q has type %q but is not in entity_types group", e.ID, e.Type),
					SpecPath:    fmt.Sprintf("groups.entity_types.%s", e.Type),
					ActualValue: e.ID,
				})
			}
		} else if e.Type != "" {
			r.AddError(validation.Result{
				Level:       validation.LevelSpatial,
				Message:     fmt.Sprintf("entity %q has type %q but no such entity_types group exists", e.ID, e.Type),
				SpecPath:    "groups.entity_types",
				ActualValue: string(e.Type),
			})
		}

		// Check system group (optional field)
		if e.System != "" {
			if sm, ok := systemMembers[string(e.System)]; ok {
				if !sm[e.ID] {
					r.AddError(validation.Result{
						Level:       validation.LevelSpatial,
						Message:     fmt.Sprintf("entity %q has system %q but is not in systems group", e.ID, e.System),
						SpecPath:    fmt.Sprintf("groups.systems.%s", e.System),
						ActualValue: e.ID,
					})
				}
			} else {
				r.AddError(validation.Result{
					Level:       validation.LevelSpatial,
					Message:     fmt.Sprintf("entity %q has system %q but no such systems group exists", e.ID, e.System),
					SpecPath:    "groups.systems",
					ActualValue: string(e.System),
				})
			}
		}

		// Check pod group (optional field)
		if e.Pod != "" {
			if pm, ok := podMembers[e.Pod]; ok {
				if !pm[e.ID] {
					r.AddError(validation.Result{
						Level:       validation.LevelSpatial,
						Message:     fmt.Sprintf("entity %q has pod %q but is not in pods group", e.ID, e.Pod),
						SpecPath:    fmt.Sprintf("groups.pods.%s", e.Pod),
						ActualValue: e.ID,
					})
				}
			} else {
				r.AddError(validation.Result{
					Level:       validation.LevelSpatial,
					Message:     fmt.Sprintf("entity %q has pod %q but no such pods group exists", e.ID, e.Pod),
					SpecPath:    "groups.pods",
					ActualValue: e.Pod,
				})
			}
		}
	}
}

func validateBoundsEnclosure(g *Graph, r *validation.Report) {
	bounds := g.Metadata.CityBounds
	tolerance := 1.0

	for _, e := range g.Entities {
		halfX := e.Dimensions.X / 2
		halfZ := e.Dimensions.Z / 2

		if e.Position.X-halfX < bounds.Min.X-tolerance || e.Position.X+halfX > bounds.Max.X+tolerance {
			r.AddWarning(validation.Result{
				Level:       validation.LevelSpatial,
				Message:     fmt.Sprintf("entity %q X extent [%.1f, %.1f] outside city bounds [%.1f, %.1f]", e.ID, e.Position.X-halfX, e.Position.X+halfX, bounds.Min.X, bounds.Max.X),
				SpecPath:    "metadata.city_bounds",
				ActualValue: e.Position.X,
			})
			break
		}
		if e.Position.Z-halfZ < bounds.Min.Z-tolerance || e.Position.Z+halfZ > bounds.Max.Z+tolerance {
			r.AddWarning(validation.Result{
				Level:       validation.LevelSpatial,
				Message:     fmt.Sprintf("entity %q Z extent [%.1f, %.1f] outside city bounds [%.1f, %.1f]", e.ID, e.Position.Z-halfZ, e.Position.Z+halfZ, bounds.Min.Z, bounds.Max.Z),
				SpecPath:    "metadata.city_bounds",
				ActualValue: e.Position.Z,
			})
			break
		}
	}
}

func validateEntityDimensions(g *Graph, r *validation.Report) {
	for _, e := range g.Entities {
		if e.Dimensions.X <= 0 || e.Dimensions.Y <= 0 || e.Dimensions.Z <= 0 {
			r.AddWarning(validation.Result{
				Level:       validation.LevelSpatial,
				Message:     fmt.Sprintf("entity %q has zero or negative dimension (%.2f, %.2f, %.2f)", e.ID, e.Dimensions.X, e.Dimensions.Y, e.Dimensions.Z),
				SpecPath:    fmt.Sprintf("entities.%s.dimensions", e.ID),
				ActualValue: fmt.Sprintf("%.2f x %.2f x %.2f", e.Dimensions.X, e.Dimensions.Y, e.Dimensions.Z),
				Expected:    "all dimensions > 0",
			})
		}
	}
}
