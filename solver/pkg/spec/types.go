package spec

// CitySpec is the top-level specification for a charter city.
type CitySpec struct {
	SpecVersion string       `yaml:"spec_version" json:"spec_version"`
	City        CityDef      `yaml:"city" json:"city"`
	CityZones   CityZones    `yaml:"city_zones" json:"city_zones"`
	Pods        PodsDef      `yaml:"pods" json:"pods"`
	Demographics Demographics `yaml:"demographics" json:"demographics"`
	Infrastructure Infrastructure `yaml:"infrastructure" json:"infrastructure"`
	Vehicles    Vehicles     `yaml:"vehicles" json:"vehicles"`
	Logistics   Logistics    `yaml:"logistics" json:"logistics"`
	Ownership   Ownership    `yaml:"ownership" json:"ownership"`
	Revenue     Revenue      `yaml:"revenue" json:"revenue"`
	Site        SiteRequirements `yaml:"site_requirements" json:"site_requirements"`
}

type CityDef struct {
	Population      int     `yaml:"population" json:"population"`
	FootprintShape  string  `yaml:"footprint_shape" json:"footprint_shape"`
	ExcavationDepth float64 `yaml:"excavation_depth" json:"excavation_depth"`
	HeightProfile   string  `yaml:"height_profile" json:"height_profile"`
	MaxHeightCenter int     `yaml:"max_height_center" json:"max_height_center"`
	MaxHeightEdge   int     `yaml:"max_height_edge" json:"max_height_edge"`
}

type CityZones struct {
	Rings     []RingDef    `yaml:"rings" json:"rings"`
	Perimeter PerimeterDef `yaml:"perimeter_infrastructure" json:"perimeter_infrastructure"`
	SolarRing SolarRingDef `yaml:"solar_ring" json:"solar_ring"`
}

// OuterRadius returns the outermost ring's outer radius.
func (cz CityZones) OuterRadius() float64 {
	if len(cz.Rings) == 0 {
		return 0
	}
	return cz.Rings[len(cz.Rings)-1].RadiusTo
}

// RingByName returns the ring definition with the given name, or nil if not found.
func (cz CityZones) RingByName(name string) *RingDef {
	for i := range cz.Rings {
		if cz.Rings[i].Name == name {
			return &cz.Rings[i]
		}
	}
	return nil
}

// RingDef defines a concentric ring zone within the city.
type RingDef struct {
	Name       string  `yaml:"name" json:"name"`
	Character  string  `yaml:"character" json:"character"`
	RadiusFrom float64 `yaml:"radius_from" json:"radius_from"`
	RadiusTo   float64 `yaml:"radius_to" json:"radius_to"`
	MaxStories int     `yaml:"max_stories" json:"max_stories"`
}

type PerimeterDef struct {
	RadiusFrom float64  `yaml:"radius_from" json:"radius_from"`
	RadiusTo   float64  `yaml:"radius_to" json:"radius_to"`
	Contents   []string `yaml:"contents" json:"contents"`
	BelowGrade bool     `yaml:"below_grade" json:"below_grade"`
}

type SolarRingDef struct {
	RadiusFrom  float64 `yaml:"radius_from" json:"radius_from"`
	RadiusTo    float64 `yaml:"radius_to" json:"radius_to"`
	AreaHa      float64 `yaml:"area_ha" json:"area_ha"`
	CapacityMW  float64 `yaml:"capacity_mw" json:"capacity_mw"`
	AvgOutputMW float64 `yaml:"avg_output_mw" json:"avg_output_mw"`
}

type PodsDef struct {
	WalkRadius      float64            `yaml:"walk_radius" json:"walk_radius"`
	RingAssignments map[string]PodRing `yaml:"ring_assignments" json:"ring_assignments"`
}

type PodRing struct {
	Character        string   `yaml:"character" json:"character"`
	RequiredServices []string `yaml:"required_services" json:"required_services"`
	MaxStories       int      `yaml:"max_stories" json:"max_stories"`
}

type Demographics struct {
	Singles      float64 `yaml:"singles" json:"singles"`
	Couples      float64 `yaml:"couples" json:"couples"`
	FamiliesYoung float64 `yaml:"families_young" json:"families_young"`
	FamiliesTeen float64 `yaml:"families_teen" json:"families_teen"`
	EmptyNest    float64 `yaml:"empty_nest" json:"empty_nest"`
	Retirees     float64 `yaml:"retirees" json:"retirees"`
}

type Infrastructure struct {
	Water    WaterInfra    `yaml:"water" json:"water"`
	Sewage   SewageInfra   `yaml:"sewage" json:"sewage"`
	Electrical ElectricalInfra `yaml:"electrical" json:"electrical"`
	Telecom  TelecomInfra  `yaml:"telecom" json:"telecom"`
	UtilityCorridors UtilityCorridors `yaml:"utility_corridors" json:"utility_corridors"`
}

type WaterInfra struct {
	Source         string `yaml:"source" json:"source"`
	CapacityGPDPer int    `yaml:"capacity_gpd_per_capita" json:"capacity_gpd_per_capita"`
}

type SewageInfra struct {
	Collection     string `yaml:"collection" json:"collection"`
	CapacityGPDPer int    `yaml:"capacity_gpd_per_capita" json:"capacity_gpd_per_capita"`
	Effluent       string `yaml:"effluent" json:"effluent"`
}

type ElectricalInfra struct {
	SolarIntegratedAvgMW float64 `yaml:"solar_integrated_avg_mw" json:"solar_integrated_avg_mw"`
	SolarFarmAvgMW       float64 `yaml:"solar_farm_avg_mw" json:"solar_farm_avg_mw"`
	BatteryCapacityMWh   float64 `yaml:"battery_capacity_mwh" json:"battery_capacity_mwh"`
	GridCapacityMW       float64 `yaml:"grid_capacity_mw" json:"grid_capacity_mw"`
	PeakDemandKWPer      float64 `yaml:"peak_demand_kw_per_capita" json:"peak_demand_kw_per_capita"`
}

type TelecomInfra struct {
	NodeSpacingM int `yaml:"node_spacing_m" json:"node_spacing_m"`
}

type UtilityCorridors struct {
	WidthM            float64 `yaml:"width_m" json:"width_m"`
	AccessPointsPerPod int    `yaml:"access_points_per_pod" json:"access_points_per_pod"`
}

type Vehicles struct {
	ArterialWidthM     float64 `yaml:"arterial_width_m" json:"arterial_width_m"`
	ServiceBranchWidthM float64 `yaml:"service_branch_width_m" json:"service_branch_width_m"`
	TotalFleet         int     `yaml:"total_fleet" json:"total_fleet"`
}

type Logistics struct {
	DailyPackagesPerCapita float64 `yaml:"daily_packages_per_capita" json:"daily_packages_per_capita"`
}

type Ownership struct {
	Model string `yaml:"model" json:"model"`
}

type Revenue struct {
	DebtTermYears   int     `yaml:"debt_term_years" json:"debt_term_years"`
	InterestRate    float64 `yaml:"interest_rate" json:"interest_rate"`
	AnnualOpsCostM  float64 `yaml:"annual_ops_cost_m" json:"annual_ops_cost_m"`
}

type SiteRequirements struct {
	MinAreaHa        float64 `yaml:"min_area_ha" json:"min_area_ha"`
	SolarIrradiance  float64 `yaml:"solar_irradiance_kwh_m2_day" json:"solar_irradiance_kwh_m2_day"`
}
