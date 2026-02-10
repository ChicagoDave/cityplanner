package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ChicagoDave/cityplanner/pkg/analytics"
	"github.com/ChicagoDave/cityplanner/pkg/cost"
	"github.com/ChicagoDave/cityplanner/pkg/layout"
	"github.com/ChicagoDave/cityplanner/pkg/routing"
	"github.com/ChicagoDave/cityplanner/pkg/scene"
	"github.com/ChicagoDave/cityplanner/pkg/spec"
	"github.com/ChicagoDave/cityplanner/pkg/validation"
)

// loadAndValidate loads the spec and runs schema validation.
func loadAndValidate(projectPath string) (*spec.CitySpec, *validation.Report, error) {
	citySpec, err := spec.LoadProject(projectPath)
	if err != nil {
		return nil, nil, fmt.Errorf("loading spec: %w", err)
	}
	schemaReport := validation.ValidateSchema(citySpec)
	return citySpec, schemaReport, nil
}

func runValidate(projectPath string) error {
	citySpec, schemaReport, err := loadAndValidate(projectPath)
	if err != nil {
		return err
	}

	// Run analytics for analytical validation
	_, analyticsReport := analytics.Resolve(citySpec)
	schemaReport.Merge(analyticsReport)

	printValidationReport(schemaReport)

	if !schemaReport.Valid {
		os.Exit(1)
	}
	return nil
}

func runCost(projectPath string) error {
	citySpec, schemaReport, err := loadAndValidate(projectPath)
	if err != nil {
		return err
	}
	if !schemaReport.Valid {
		printValidationReport(schemaReport)
		return fmt.Errorf("spec has validation errors; fix before computing cost")
	}

	params, analyticsReport := analytics.Resolve(citySpec)
	costReport := cost.Estimate(citySpec, params)

	params.PerCapitaCost = costReport.Summary.PerCapita
	params.BreakEvenRent = costReport.Summary.BreakEvenMonthlyRent

	printCostReport(costReport)

	if len(analyticsReport.Warnings) > 0 {
		fmt.Println()
		printValidationReport(analyticsReport)
	}
	return nil
}

func runSolve(projectPath string) error {
	citySpec, schemaReport, err := loadAndValidate(projectPath)
	if err != nil {
		return err
	}
	if !schemaReport.Valid {
		printValidationReport(schemaReport)
		return fmt.Errorf("spec has validation errors")
	}

	params, analyticsReport := analytics.Resolve(citySpec)
	if !analyticsReport.Valid {
		printValidationReport(analyticsReport)
		return fmt.Errorf("analytical validation failed")
	}

	costReport := cost.Estimate(citySpec, params)
	params.PerCapitaCost = costReport.Summary.PerCapita
	params.BreakEvenRent = costReport.Summary.BreakEvenMonthlyRent

	// Phase 2: Spatial generation.
	pods, adjacency, podReport := layout.LayoutPods(citySpec, params)
	analyticsReport.Merge(podReport)

	buildings, paths, buildReport := layout.PlaceBuildings(citySpec, pods, adjacency, params)
	analyticsReport.Merge(buildReport)

	segments, routeReport := routing.RouteInfrastructure(citySpec, pods, buildings)
	analyticsReport.Merge(routeReport)

	greenZones := layout.CollectGreenZones(citySpec, pods)
	graph := scene.Assemble(citySpec, pods, buildings, paths, segments, greenZones)

	output := map[string]any{
		"phase":       2,
		"parameters":  params,
		"cost":        costReport,
		"validation":  analyticsReport,
		"scene_graph": graph,
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(output)
}
