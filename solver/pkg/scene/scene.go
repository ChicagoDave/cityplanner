package scene

// SystemType identifies an infrastructure system.
type SystemType string

const (
	SystemWater      SystemType = "water"
	SystemSewage     SystemType = "sewage"
	SystemElectrical SystemType = "electrical"
	SystemTelecom    SystemType = "telecom"
	SystemVehicle    SystemType = "vehicle"
	SystemPedestrian SystemType = "pedestrian"
	SystemBicycle    SystemType = "bicycle"
	SystemShuttle    SystemType = "shuttle"
)

// LayerType identifies a vertical layer.
type LayerType string

const (
	LayerUnderground1 LayerType = "underground_1"
	LayerUnderground2 LayerType = "underground_2"
	LayerUnderground3 LayerType = "underground_3"
	LayerSurface      LayerType = "surface"
)

// EntityType identifies the kind of entity.
type EntityType string

const (
	EntityBuilding   EntityType = "building"
	EntityPipe       EntityType = "pipe"
	EntityLane       EntityType = "lane"
	EntityPath       EntityType = "path"
	EntityPanel      EntityType = "panel"
	EntityPark       EntityType = "park"
	EntityPedway        EntityType = "pedway"
	EntityBikeTunnel    EntityType = "bike_tunnel"
	EntityBattery       EntityType = "battery"
	EntityBikePath      EntityType = "bike_path"
	EntityShuttleRoute  EntityType = "shuttle_route"
	EntityStation       EntityType = "station"
	EntitySportsField   EntityType = "sports_field"
	EntityPlaza         EntityType = "plaza"
	EntityTree          EntityType = "tree"
)

// Vec3 is a 3D vector.
type Vec3 struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

// BoundingBox defines an axis-aligned bounding box.
type BoundingBox struct {
	Min Vec3 `json:"min"`
	Max Vec3 `json:"max"`
}

// Entity is a single element in the scene graph.
type Entity struct {
	ID         string            `json:"id"`
	Type       EntityType        `json:"type"`
	Position   Vec3              `json:"position"`
	Dimensions Vec3              `json:"dimensions"`
	Rotation   [4]float64        `json:"rotation"` // quaternion [x, y, z, w]
	Material   string            `json:"material"`
	System     SystemType        `json:"system,omitempty"`
	Pod        string            `json:"pod,omitempty"`
	Layer      LayerType         `json:"layer"`
	Metadata   map[string]any    `json:"metadata,omitempty"`
	Children   []string          `json:"children,omitempty"`
}

// Graph is the complete scene graph output from the solver.
type Graph struct {
	Metadata Metadata              `json:"metadata"`
	Entities []Entity              `json:"entities"`
	Groups   Groups                `json:"groups"`
}

// Metadata holds scene-level information.
type Metadata struct {
	SpecVersion string      `json:"spec_version"`
	GeneratedAt string      `json:"generated_at"`
	CityBounds  BoundingBox `json:"city_bounds"`
}

// Groups organizes entity IDs by various axes for fast filtering.
type Groups struct {
	Pods        map[string][]string     `json:"pods"`
	Systems     map[SystemType][]string `json:"systems"`
	Layers      map[LayerType][]string  `json:"layers"`
	EntityTypes map[EntityType][]string `json:"entity_types"`
}

// NewGraph creates an empty scene graph.
func NewGraph() *Graph {
	return &Graph{
		Entities: []Entity{},
		Groups: Groups{
			Pods:        make(map[string][]string),
			Systems:     make(map[SystemType][]string),
			Layers:      make(map[LayerType][]string),
			EntityTypes: make(map[EntityType][]string),
		},
	}
}
