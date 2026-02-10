package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// Server is the local development server for interactive design.
type Server struct {
	projectPath string
	port        int
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
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/scene", s.handleScene)
	mux.HandleFunc("GET /api/cost", s.handleCost)
	mux.HandleFunc("GET /api/validation", s.handleValidation)
	mux.HandleFunc("POST /api/solve", s.handleSolve)
	mux.HandleFunc("GET /api/spec", s.handleSpec)
	mux.HandleFunc("GET /", s.handleIndex)

	addr := fmt.Sprintf(":%d", s.port)
	log.Printf("CityPlanner server starting on http://localhost%s", addr)
	log.Printf("Project: %s", s.projectPath)

	return http.ListenAndServe(addr, mux)
}

func (s *Server) handleIndex(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, `<!DOCTYPE html>
<html><head><title>CityPlanner</title></head>
<body style="margin:0;background:#111;color:#fff;font-family:system-ui;display:flex;align-items:center;justify-content:center;height:100vh">
<div style="text-align:center">
<h1>CityPlanner</h1>
<p>Renderer not yet embedded. Run <code>npm run dev</code> in renderer/ for development.</p>
</div>
</body></html>`)
}

func (s *Server) handleScene(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"metadata": map[string]any{
			"spec_version": "0.1.0",
			"generated_at": "",
		},
		"entities": []any{},
		"groups":   map[string]any{},
	})
}

func (s *Server) handleCost(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "not yet implemented"})
}

func (s *Server) handleValidation(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"valid":    true,
		"errors":   []any{},
		"warnings": []any{},
		"info":     []any{},
		"summary":  "0 errors, 0 warnings, 0 info",
	})
}

func (s *Server) handleSolve(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "solver not yet implemented"})
}

func (s *Server) handleSpec(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "not yet implemented"})
}
