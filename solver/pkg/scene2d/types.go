package scene2d

// Scene2D is the complete 2D scene output for an SVG top-down renderer.
type Scene2D struct {
	Metadata     Metadata        `json:"metadata"`
	Rings        []Ring          `json:"rings"`
	Pods         []Pod2D         `json:"pods"`
	Paths        PathCollection  `json:"paths"`
	Stations     []Station2D     `json:"stations"`
	Sports       SportsCollection `json:"sports"`
	Plazas       []Plaza2D       `json:"plazas"`
	Trees        TreeSummary     `json:"trees"`
	Buildings    BuildingSummary `json:"buildings"`
	ExternalBand *ExternalBand   `json:"external_band,omitempty"`
}

// Metadata holds city-level summary data.
type Metadata struct {
	Population          int     `json:"population"`
	PodCount            int     `json:"pod_count"`
	CityRadiusM         float64 `json:"city_radius_m"`
	ExternalBandRadiusM float64 `json:"external_band_radius_m,omitempty"`
	GeneratedAt         string  `json:"generated_at"`
}

// Ring describes a concentric ring zone.
type Ring struct {
	Name       string  `json:"name"`
	RadiusFrom float64 `json:"radius_from"`
	RadiusTo   float64 `json:"radius_to"`
	MaxStories int     `json:"max_stories"`
	Character  string  `json:"character"`
	PodCount   int     `json:"pod_count"`
	Population int     `json:"population"`
}

// Pod2D describes a single pod in the 2D view.
type Pod2D struct {
	ID         string       `json:"id"`
	Ring       string       `json:"ring"`
	Center     [2]float64   `json:"center"`
	Boundary   [][2]float64 `json:"boundary"`
	Population int          `json:"population"`
	MaxStories int          `json:"max_stories"`
	AreaHa     float64      `json:"area_ha"`
	Zones      []Zone2D     `json:"zones"`
}

// Zone2D describes a functional zone within a pod.
type Zone2D struct {
	Type    string       `json:"type"`
	Polygon [][2]float64 `json:"polygon"`
	AreaHa  float64      `json:"area_ha"`
}

// PathCollection groups all path types.
type PathCollection struct {
	Pedestrian []PedestrianPath2D `json:"pedestrian"`
	Bike       []BikePath2D       `json:"bike"`
	Shuttle    []ShuttlePath2D    `json:"shuttle"`
}

// PedestrianPath2D is a pedestrian path segment.
type PedestrianPath2D struct {
	ID    string     `json:"id"`
	Start [2]float64 `json:"start"`
	End   [2]float64 `json:"end"`
	Width float64    `json:"width"`
	Type  string     `json:"type"`
}

// BikePath2D is a bike path polyline.
type BikePath2D struct {
	ID       string       `json:"id"`
	Points   [][2]float64 `json:"points"`
	Width    float64      `json:"width"`
	Elevated float64      `json:"elevated"`
	Type     string       `json:"type"`
}

// ShuttlePath2D is a shuttle route polyline.
type ShuttlePath2D struct {
	ID     string       `json:"id"`
	Points [][2]float64 `json:"points"`
	Width  float64      `json:"width"`
	Type   string       `json:"type"`
}

// Station2D is a mobility hub.
type Station2D struct {
	ID       string     `json:"id"`
	PodID    string     `json:"pod_id"`
	Position [2]float64 `json:"position"`
	RouteID  string     `json:"route_id"`
}

// SportsCollection holds sports facilities.
type SportsCollection struct {
	Fields []SportsField2D `json:"fields"`
}

// SportsField2D is a single sports facility.
type SportsField2D struct {
	ID         string     `json:"id"`
	Type       string     `json:"type"`
	Position   [2]float64 `json:"position"`
	Dimensions [2]float64 `json:"dimensions"`
	Rotation   float64    `json:"rotation"`
}

// Plaza2D is a public gathering space.
type Plaza2D struct {
	ID       string     `json:"id"`
	PodID    string     `json:"pod_id"`
	Position [2]float64 `json:"position"`
	Width    float64    `json:"width"`
	Depth    float64    `json:"depth"`
	Rotation float64    `json:"rotation"`
}

// TreeSummary holds aggregate tree counts by context.
type TreeSummary struct {
	ParkCount  int `json:"park_count"`
	PathCount  int `json:"path_count"`
	PlazaCount int `json:"plaza_count"`
	Total      int `json:"total"`
}

// BuildingSummary holds aggregate building data.
type BuildingSummary struct {
	TotalBuildings int                       `json:"total_buildings"`
	TotalDU        int                       `json:"total_dwelling_units"`
	ByPod          map[string]PodBuildingSum `json:"by_pod"`
}

// PodBuildingSum is the building aggregate for one pod.
type PodBuildingSum struct {
	Residential   int      `json:"residential"`
	Commercial    int      `json:"commercial"`
	Civic         int      `json:"civic"`
	TotalUnits    int      `json:"total_units"`
	CommercialSqM float64  `json:"commercial_sqm"`
	ServiceTypes  []string `json:"service_types,omitempty"`
}

// ExternalBand describes the perimeter infrastructure zone.
type ExternalBand struct {
	RadiusFrom float64  `json:"radius_from"`
	RadiusTo   float64  `json:"radius_to"`
	Facilities []string `json:"facilities"`
}
