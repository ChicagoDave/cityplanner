package main

import (
	"os"

	"github.com/ChicagoDave/cityplanner/internal/server"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "cityplanner",
		Short: "Charter City specification-driven design engine",
	}

	rootCmd.AddCommand(solveCmd())
	rootCmd.AddCommand(validateCmd())
	rootCmd.AddCommand(costCmd())
	rootCmd.AddCommand(serveCmd())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func solveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "solve [project-path]",
		Short: "Run the full solver pipeline and generate a scene graph",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return runSolve(args[0])
		},
	}
}

func validateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate [project-path]",
		Short: "Validate a city spec without running the full solver",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return runValidate(args[0])
		},
	}
}

func costCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "cost [project-path]",
		Short: "Compute and display cost estimates",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			return runCost(args[0])
		},
	}
}

func serveCmd() *cobra.Command {
	var port int

	cmd := &cobra.Command{
		Use:   "serve [project-path]",
		Short: "Start the local dev server with interactive 3D renderer",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			srv := server.New(args[0], port)
			return srv.Start()
		},
	}

	cmd.Flags().IntVarP(&port, "port", "p", 3000, "HTTP server port")
	return cmd
}
