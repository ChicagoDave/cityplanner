package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/ChicagoDave/cityplanner/pkg/analytics"
	"github.com/ChicagoDave/cityplanner/pkg/cost"
	"github.com/ChicagoDave/cityplanner/pkg/layout"
	"github.com/ChicagoDave/cityplanner/pkg/routing"
	"github.com/ChicagoDave/cityplanner/pkg/scene"
	"github.com/ChicagoDave/cityplanner/pkg/spec"
	"github.com/ChicagoDave/cityplanner/pkg/validation"
)

// Server is the local development server for interactive design.
type Server struct {
	projectPath string
	port        int

	mu         sync.RWMutex
	citySpec   *spec.CitySpec
	params     *analytics.ResolvedParameters
	costReport *cost.Report
	valReport  *validation.Report
	sceneGraph *scene.Graph
}

// New creates a server for the given project directory.
func New(projectPath string, port int) *Server {
	return &Server{
		projectPath: projectPath,
		port:        port,
	}
}

// Start launches the HTTP server.
func (s *Server) Start() error {
	if err := s.loadAndSolve(); err != nil {
		log.Printf("Warning: initial solve failed: %v", err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/scene", s.handleScene)
	mux.HandleFunc("GET /api/cost", s.handleCost)
	mux.HandleFunc("GET /api/validation", s.handleValidation)
	mux.HandleFunc("POST /api/solve", s.handleSolve)
	mux.HandleFunc("GET /api/spec", s.handleSpec)
	mux.HandleFunc("GET /api/parameters", s.handleParameters)
	mux.HandleFunc("GET /", s.handleIndex)

	addr := fmt.Sprintf(":%d", s.port)
	log.Printf("CityPlanner server starting on http://localhost%s", addr)
	log.Printf("Project: %s", s.projectPath)

	return http.ListenAndServe(addr, mux)
}

func (s *Server) loadAndSolve() error {
	citySpec, err := spec.LoadProject(s.projectPath)
	if err != nil {
		return fmt.Errorf("loading spec: %w", err)
	}

	schemaReport := validation.ValidateSchema(citySpec)
	params, analyticsReport := analytics.Resolve(citySpec)
	schemaReport.Merge(analyticsReport)

	costReport := cost.Estimate(citySpec, params)
	params.PerCapitaCost = costReport.Summary.PerCapita
	params.BreakEvenRent = costReport.Summary.BreakEvenMonthlyRent

	// Phase 2: Spatial generation.
	pods, adjacency, podReport := layout.LayoutPods(citySpec, params)
	schemaReport.Merge(podReport)

	buildings, paths, buildReport := layout.PlaceBuildings(citySpec, pods, adjacency, params)
	schemaReport.Merge(buildReport)

	segments, routeReport := routing.RouteInfrastructure(citySpec, pods, buildings)
	schemaReport.Merge(routeReport)

	bikePaths, bikeReport := layout.GenerateBikePaths(pods, adjacency, citySpec.CityZones.Rings)
	schemaReport.Merge(bikeReport)

	shuttleRoutes, stations, shuttleReport := layout.GenerateShuttleRoutes(bikePaths, pods)
	schemaReport.Merge(shuttleReport)

	sportsFields, sportsReport := layout.PlaceSportsFields(pods, adjacency, citySpec.CityZones.Rings)
	schemaReport.Merge(sportsReport)

	greenZones := layout.CollectGreenZones(citySpec, pods)

	plazas, plazaReport := layout.GeneratePlazas(pods, citySpec)
	schemaReport.Merge(plazaReport)

	trees, treeReport := layout.PlaceTrees(pods, greenZones, paths, bikePaths, plazas)
	schemaReport.Merge(treeReport)

	graph := scene.Assemble(citySpec, pods, buildings, paths, segments, greenZones, bikePaths, shuttleRoutes, stations, sportsFields, plazas, trees)

	s.mu.Lock()
	defer s.mu.Unlock()
	s.citySpec = citySpec
	s.params = params
	s.costReport = costReport
	s.valReport = schemaReport
	s.sceneGraph = graph
	return nil
}

func (s *Server) handleIndex(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, `<!DOCTYPE html>
<html><head><title>CityPlanner</title></head>
<body style="margin:0;background:#111;color:#fff;font-family:system-ui;display:flex;align-items:center;justify-content:center;height:100vh">
<div style="text-align:center">
<h1>CityPlanner</h1>
<p>Renderer not yet embedded. Run <code>npm run dev</code> in renderer/ for development.</p>
<p>API endpoints: <a href="/api/spec">/api/spec</a> | <a href="/api/validation">/api/validation</a> | <a href="/api/cost">/api/cost</a> | <a href="/api/parameters">/api/parameters</a></p>
</div>
</body></html>`)
}

func (s *Server) handleScene(w http.ResponseWriter, _ *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	if s.sceneGraph == nil {
		http.Error(w, `{"error":"no scene graph available"}`, http.StatusServiceUnavailable)
		return
	}
	json.NewEncoder(w).Encode(s.sceneGraph)
}

func (s *Server) handleCost(w http.ResponseWriter, _ *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	if s.costReport == nil {
		http.Error(w, `{"error":"no cost data available"}`, http.StatusServiceUnavailable)
		return
	}
	json.NewEncoder(w).Encode(s.costReport)
}

func (s *Server) handleValidation(w http.ResponseWriter, _ *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	if s.valReport == nil {
		http.Error(w, `{"error":"no validation data available"}`, http.StatusServiceUnavailable)
		return
	}
	json.NewEncoder(w).Encode(s.valReport)
}

func (s *Server) handleSolve(w http.ResponseWriter, _ *http.Request) {
	if err := s.loadAndSolve(); err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"status":      "ok",
		"parameters":  s.params,
		"cost":        s.costReport,
		"validation":  s.valReport,
		"scene_graph": s.sceneGraph,
	})
}

func (s *Server) handleSpec(w http.ResponseWriter, _ *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	if s.citySpec == nil {
		http.Error(w, `{"error":"no spec loaded"}`, http.StatusServiceUnavailable)
		return
	}
	json.NewEncoder(w).Encode(s.citySpec)
}

func (s *Server) handleParameters(w http.ResponseWriter, _ *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	if s.params == nil {
		http.Error(w, `{"error":"no parameters available"}`, http.StatusServiceUnavailable)
		return
	}
	json.NewEncoder(w).Encode(s.params)
}
