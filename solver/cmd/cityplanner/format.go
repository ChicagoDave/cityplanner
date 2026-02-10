package main

import (
	"fmt"

	"github.com/ChicagoDave/cityplanner/pkg/cost"
	"github.com/ChicagoDave/cityplanner/pkg/validation"
)

func printValidationReport(r *validation.Report) {
	if len(r.Errors) > 0 {
		fmt.Printf("ERRORS (%d):\n", len(r.Errors))
		for _, e := range r.Errors {
			fmt.Printf("  [%s] %s\n", e.Level, e.Message)
			if e.SpecPath != "" {
				fmt.Printf("    -> %s = %v\n", e.SpecPath, e.ActualValue)
			}
			if e.Expected != "" {
				fmt.Printf("    expected: %s\n", e.Expected)
			}
			if e.ConflictWith != "" {
				fmt.Printf("    conflicts with: %s\n", e.ConflictWith)
			}
			for _, s := range e.Suggestions {
				fmt.Printf("    * %s\n", s)
			}
		}
		fmt.Println()
	}

	if len(r.Warnings) > 0 {
		fmt.Printf("WARNINGS (%d):\n", len(r.Warnings))
		for _, w := range r.Warnings {
			fmt.Printf("  [%s] %s\n", w.Level, w.Message)
			if w.SpecPath != "" {
				fmt.Printf("    -> %s = %v\n", w.SpecPath, w.ActualValue)
			}
			if w.Expected != "" {
				fmt.Printf("    expected: %s\n", w.Expected)
			}
			for _, s := range w.Suggestions {
				fmt.Printf("    * %s\n", s)
			}
		}
		fmt.Println()
	}

	if len(r.Info) > 0 {
		fmt.Printf("INFO (%d):\n", len(r.Info))
		for _, i := range r.Info {
			fmt.Printf("  [%s] %s\n", i.Level, i.Message)
		}
		fmt.Println()
	}

	if r.Valid {
		fmt.Printf("Result: VALID (%s)\n", r.Summary)
	} else {
		fmt.Printf("Result: INVALID (%s)\n", r.Summary)
	}
}

func printCostReport(r *cost.Report) {
	if r.Estimate == nil {
		fmt.Println("No cost estimate available.")
		return
	}

	fmt.Println("Cost Estimate (Phase 1 Analytical)")
	fmt.Println("===================================")
	fmt.Println()

	printBreakdownTable(r.Estimate)

	fmt.Println()
	fmt.Println("Summary")
	fmt.Println("-------")
	fmt.Printf("  Total construction:     $%s\n", formatMoney(r.Summary.TotalConstruction))
	fmt.Printf("  Per capita:             $%s\n", formatMoney(r.Summary.PerCapita))
	fmt.Printf("  Annual debt service:    $%s\n", formatMoney(r.Summary.AnnualDebtService))
	fmt.Printf("  Annual operations:      $%s\n", formatMoney(r.Summary.AnnualOperations))
	fmt.Printf("  Break-even rent/month:  $%s\n", formatMoney(r.Summary.BreakEvenMonthlyRent))
}

func printBreakdownTable(pc *cost.PhasedCost) {
	fmt.Printf("%-18s %14s %14s %14s %14s %14s\n",
		"Category", "Phase 1", "Phase 2", "Phase 3", "Perim+Solar", "Total")
	fmt.Printf("%-18s %14s %14s %14s %14s %14s\n",
		"------------------", "--------------", "--------------", "--------------", "--------------", "--------------")

	rows := []struct {
		label string
		vals  [5]float64
	}{
		{"Excavation", [5]float64{pc.Phase1.Excavation, pc.Phase2.Excavation, pc.Phase3.Excavation, pc.PerimeterAndSolar.Excavation, pc.Total.Excavation}},
		{"Structural", [5]float64{pc.Phase1.Structural, pc.Phase2.Structural, pc.Phase3.Structural, pc.PerimeterAndSolar.Structural, pc.Total.Structural}},
		{"Buildings", [5]float64{pc.Phase1.Buildings, pc.Phase2.Buildings, pc.Phase3.Buildings, pc.PerimeterAndSolar.Buildings, pc.Total.Buildings}},
		{"Infrastructure", [5]float64{pc.Phase1.Infrastructure, pc.Phase2.Infrastructure, pc.Phase3.Infrastructure, pc.PerimeterAndSolar.Infrastructure, pc.Total.Infrastructure}},
		{"Solar", [5]float64{pc.Phase1.Solar, pc.Phase2.Solar, pc.Phase3.Solar, pc.PerimeterAndSolar.Solar, pc.Total.Solar}},
		{"Battery", [5]float64{pc.Phase1.Battery, pc.Phase2.Battery, pc.Phase3.Battery, pc.PerimeterAndSolar.Battery, pc.Total.Battery}},
		{"TOTAL", [5]float64{pc.Phase1.Total, pc.Phase2.Total, pc.Phase3.Total, pc.PerimeterAndSolar.Total, pc.Total.Total}},
	}

	for _, row := range rows {
		fmt.Printf("%-18s", row.label)
		for _, v := range row.vals {
			fmt.Printf(" %14s", formatMoney(v))
		}
		fmt.Println()
	}
}

func formatMoney(v float64) string {
	if v >= 1_000_000_000 {
		return fmt.Sprintf("%.2fB", v/1_000_000_000)
	}
	if v >= 1_000_000 {
		return fmt.Sprintf("%.2fM", v/1_000_000)
	}
	if v >= 1_000 {
		return fmt.Sprintf("%.0fK", v/1_000)
	}
	return fmt.Sprintf("%.0f", v)
}
