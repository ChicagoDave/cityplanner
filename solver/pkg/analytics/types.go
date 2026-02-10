package analytics

// CohortBreakdown holds per-cohort household and population data.
type CohortBreakdown struct {
	Name          string  `json:"name"`
	Ratio         float64 `json:"ratio"`
	HouseholdSize float64 `json:"household_size"`
	Households    int     `json:"households"`
	Population    int     `json:"population"`
	Adults        int     `json:"adults"`
	Children      int     `json:"children"`
}

// RingData holds computed data for one concentric ring.
type RingData struct {
	Name              string  `json:"name"`
	RadiusFrom        float64 `json:"radius_from_m"`
	RadiusTo          float64 `json:"radius_to_m"`
	AreaHa            float64 `json:"area_ha"`
	AreaFraction      float64 `json:"area_fraction"`
	Population        int     `json:"population"`
	Households        int     `json:"households"`
	PodCount          int     `json:"pod_count"`
	PodPopulation     int     `json:"pod_population"`
	MaxStories        int     `json:"max_stories"`
	AvgHouseholdSize  float64 `json:"avg_household_size"`
	RequiredDensity   float64 `json:"required_density_du_ha"`
	AchievableDensity float64 `json:"achievable_density_du_ha"`
	ResidentialAreaHa float64 `json:"residential_area_ha"`
}

// ServiceCount holds the required count for one service type.
type ServiceCount struct {
	Service    string `json:"service"`
	Threshold  int    `json:"threshold_per_unit"`
	Required   int    `json:"required_count"`
	Metric     string `json:"metric"`
	Population int    `json:"relevant_population"`
}

// AreaBreakdown holds the land-use allocation.
type AreaBreakdown struct {
	TotalCityHa        float64 `json:"total_city_ha"`
	ResidentialHa      float64 `json:"residential_ha"`
	CommercialHa       float64 `json:"commercial_ha"`
	CivicHa            float64 `json:"civic_ha"`
	GreenPathsHa       float64 `json:"green_paths_ha"`
	PerimeterHa        float64 `json:"perimeter_ha"`
	SolarHa            float64 `json:"solar_ha"`
	TotalWithPerimeter float64 `json:"total_with_perimeter_ha"`
}

// EnergyBalance holds electrical supply/demand analysis.
type EnergyBalance struct {
	PeakDemandMW       float64 `json:"peak_demand_mw"`
	SolarIntegratedMW  float64 `json:"solar_integrated_avg_mw"`
	SolarFarmMW        float64 `json:"solar_farm_avg_mw"`
	TotalGenerationMW  float64 `json:"total_generation_mw"`
	GridCapacityMW     float64 `json:"grid_capacity_mw"`
	BatteryCapacityMWh float64 `json:"battery_capacity_mwh"`
	BackupHours        float64 `json:"backup_hours"`
}
