package cost

// Unit cost constants for Phase 1 estimation.
// These are baseline values from the technical specification.
const (
	ExcavationCostPerM3     = 35.0     // $/m³
	SlabCostPerM2           = 150.0    // $/m² per structural level
	ResidentialCostPerM2    = 2000.0   // $/m² floor area
	CommercialCostPerM2     = 2500.0   // $/m² floor area
	CivicCostPerM2          = 3000.0   // $/m² floor area
	SolarCostPerM2          = 200.0    // $/m² panel area
	BatteryCostPerMWh       = 300000.0 // $/MWh capacity
	InfraWaterCostPerM      = 500.0    // $/m pipe length
	InfraSewageCostPerM     = 600.0    // $/m pipe length
	InfraElectricalCostPerM = 400.0    // $/m conduit length
	InfraTelecomCostPerM    = 200.0    // $/m fiber length
	InfraVehicleCostPerM    = 1000.0   // $/m lane length

	AvgUnitSizeM2       = 75.0  // average dwelling unit floor area
	GroundCoverageRatio = 0.60  // building footprint / lot area
	UndergroundLevels   = 3     // number of structural slabs
	AvgCommercialStories = 6.0  // weighted average for commercial buildings
	AvgCivicStories      = 8.0  // weighted average for civic buildings
	M2PerHa              = 10000.0
)
