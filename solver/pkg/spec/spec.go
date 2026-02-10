package spec

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Load reads a city spec from a YAML file.
func Load(path string) (*CitySpec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading spec file: %w", err)
	}

	var spec CitySpec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("parsing spec YAML: %w", err)
	}

	return &spec, nil
}

// LoadProject loads a city spec from a project directory.
// It looks for city.yaml in the given directory.
func LoadProject(projectDir string) (*CitySpec, error) {
	specPath := filepath.Join(projectDir, "city.yaml")
	return Load(specPath)
}
