# Charter City — Technical Specification

**Formulas, Data Model, and Engine Spec**

David Cornelson | February 2026

---

## Core Formulas

### Population Composition

The model works in households rather than individuals, since housing units map to households.

```
H_total = total households
H_cohort = H_total × r_cohort
P_total = Σ (H_cohort × avg_household_size_cohort)

Cohort ratios (must sum to 1.0):
  r_singles       = 0.15  (household size: 1.0)
  r_couples       = 0.20  (household size: 2.0)
  r_families_young = 0.25  (household size: 3.5)
  r_families_teen  = 0.15  (household size: 4.0)
  r_empty_nest    = 0.15  (household size: 2.0)
  r_retirees      = 0.10  (household size: 1.5)

Weighted average household size ≈ 2.5
For P_total = 50,000: H_total ≈ 20,000
```

### Dependency Ratio

```
working_age = P_singles + P_couples + P_fam_young_adults
            + P_fam_teen_adults + P_empty_nest
dependents  = P_children + P_retirees
dependency_ratio = dependents / working_age
Target: 0.5 to 0.6
```

### Service Thresholds

Each service type has a minimum population to sustain it. The generalized formula is:

```
service_count_i = ceil(P_relevant_i / threshold_i)

Service             | Population per unit | Relevant population
--------------------|---------------------|--------------------
Grocery store       | 4,000               | P_total
Elementary school   | 500 students        | P_school_age_elementary
Secondary school    | 800 students        | P_school_age_secondary
Medical clinic      | 10,000              | P_total
Hospital            | 50,000              | P_total
Library branch      | 15,000              | P_total
Pharmacy            | 8,000               | P_total
Dental clinic       | 5,000               | P_total
```

### Pod Sizing

Each pod must satisfy the proximity constraint: all residents within walking distance of essential services.

```
walk_radius = 400m (5-minute walk)
pod_area = π × walk_radius² ≈ 50 hectares (≈125 acres)
pod_population = pod_area × density
num_pods = ceil(P_total / pod_population)

Binding constraint:
  pod_population ≥ max(threshold_i) for all required services

If grocery needs 4,000/pod and density only supports 2,000/pod:
  → increase density, OR
  → increase walk_radius, OR
  → share services across adjacent pods
```

### Housing Density

```
dwelling_units_per_hectare = H_total / total_residential_area

Reference densities:
  Suburban sprawl:     ~15 du/ha
  Mid-rise walkable:   ~75-150 du/ha
  High-rise:           ~250+ du/ha

Without roads (+35% usable land):
  Effective mid-rise = 75 × 1.35 ≈ 100 du/ha

At 100 du/ha: 20,000 units / 100 = 200 ha residential
With commercial, parks, civic (×1.75): ~350 ha total
= ~865 acres = ~1.35 square miles
```

### Commercial Square Footage

```
retail_sqft = P_total × sqft_per_capita_retail
  sqft_per_capita_retail ≈ 15-25

P_working = working_age × labor_force_participation (≈0.63)
cowork_sqft = P_working × cowork_participation_rate × sqft_per_worker
  sqft_per_worker ≈ 50-75
  cowork_participation_rate = fraction working locally
```

### Key Feedback Loop

The central optimization loop:

```
P_total
  → service thresholds
  → min pod size
  → density requirement
  → housing mix
  → land area
  → excavation cost
  → per-capita cost
  → economic viability
  → back to P_total
```

---

## City Specification Schema

The following is the declarative specification that drives the generation engine. All values are adjustable; the engine validates constraints and generates the model.

### City Definition

```yaml
city:
  population: 50000
  footprint_shape: circle | square | irregular
  excavation_depth: 8m  # 3 underground layers
  height_profile: bowl
  max_height_center: 20 stories
  max_height_edge: 4 stories
```

### Zone Definitions

```yaml
city_zones:
  center:
    character: civic_commercial
    radius_from_center: 0-300m
    max_stories: 20

  middle:
    character: mixed_residential_commercial
    radius: 300-600m
    max_stories: 10

  edge:
    character: family_education
    radius: 600-900m
    max_stories: 4

  perimeter_infrastructure:
    radius: 900-1100m
    contents:
      - sewage_treatment
      - water_treatment
      - freight_staging
      - vehicle_maintenance
      - electrical_substation
      - grid_interconnect
    below_grade: true

  solar_ring:
    radius: 1100-1500m
    area_ha: 250
    capacity_mw: 500  # nameplate
    avg_output_mw: 100
```

### Pod Definitions

```yaml
pods:
  walk_radius: 400m
  ring_assignments:
    edge:
      character: residential_family
      required_services:
        - elementary_school
        - library
        - grocery
        - playground
        - pediatric_clinic
        - daycare
      max_stories: 4
    middle:
      character: mixed
      required_services:
        - secondary_school
        - coworking
        - medical_clinic
        - retail
        - restaurant
      max_stories: 10
    center:
      character: civic_commercial
      required_services:
        - hospital
        - performing_arts
        - city_hall
        - coworking_hub
      max_stories: 20
```

### Demographics

```yaml
demographics:
  singles: 0.15
  couples: 0.20
  families_young: 0.25
  families_teen: 0.15
  empty_nest: 0.15
  retirees: 0.10
```

---

## Infrastructure Specification

### Water Supply

```yaml
infrastructure.water:
  source: municipal_connection | well_field | reservoir
  treatment_plant:
    location: perimeter
    capacity_gpd: population * 100
  distribution: trunk_to_pod_branching
```

### Sewage

```yaml
infrastructure.sewage:
  collection: gravity_flow_to_perimeter
  treatment_plant:
    location: perimeter
    capacity_gpd: population * 95
    effluent: discharge | recycle_irrigation
```

### Stormwater

```yaml
infrastructure.stormwater:
  surface_capture: bioswale + drain_penetrations
  underground_retention_gallons: city_area * design_storm_depth
```

### Electrical

```yaml
infrastructure.electrical:
  generation:
    solar_integrated:
      coverage: all_upward_surfaces
      rooftop_area_ha: 150
      canopy_area_ha: 50
      avg_output_mw: 80
    solar_farm:
      location: solar_ring
      area_ha: 250
      avg_output_mw: 100
    total_avg_output_mw: 180

  storage:
    type: battery
    location: underground_distributed
    capacity_mwh: 3000  # 24 hours
    distribution: per_pod
    capacity_per_pod_mwh: 230

  grid_connection:
    type: bidirectional
    capacity_mw: 150
    mode: backup_import + surplus_export
    target_annual_import: 0

  distribution_model: induction_surface
  wired_segments:
    - substation_to_transformer
    - transformer_to_building
    - riser_to_floor
  induction_segments:
    - floor_to_devices

  peak_demand_mw: population * 0.0025  # 125 MW
  gas: none
```

### Telecom

```yaml
infrastructure.telecom:
  distribution_model: wireless_mesh
  backbone: fiber_to_node
  node_spacing_m: 75
  node_count: city_area / (75 * 75)
  last_mile: wifi6e + 5g_small_cell
  wired_to_unit: false
```

### Utility Corridors

```yaml
infrastructure.utility_corridors:
  width: 2.5m  # reduced from 3m via wireless/induction
  routing: parallel_to_vehicle_lanes
  access_points_per_pod: 2
```

---

## Vehicle and Logistics Specification

### Lane Network

```yaml
vehicles.lane_network:
  arterial_width: 6m
  service_branch_width: 4m
  routing: center_hub_radial | grid
```

### Vehicle Fleet

```yaml
vehicles.fleet:
  package_vehicles:
    type: small_autonomous
    capacity: 50 parcels
    size: golf_cart
    charging: inductive_lane
  freight_vehicles:
    type: large_autonomous_flatbed
    capacity: 4 pallets
    charging: inductive_lane
  total_fleet: ~200
```

### Storage and Maintenance

```yaml
vehicles.storage_depots:
  count: ceil(num_pods / 3)
  location: distributed
  capacity_per: 50 vehicles

vehicles.maintenance_facility:
  count: 2
  location: perimeter

vehicles.charging:
  type: inductive_lane
  lane_induction_coverage: 0.6  # 60% of lane surface
  depot_charging: false
```

### Freight Staging

```yaml
vehicles.freight_staging:
  location: perimeter
  count: 2-4
  connects_to: external_road_network
```

### Vertical Access

```yaml
vehicles.vertical_access:
  freight_elevators_per_pod: 1
  emergency_elevators_per_pod: 1
  surface_access_for_parks: as_needed
```

### Package Delivery

```yaml
logistics.package_delivery:
  entry: perimeter_freight_staging
  sortation: automated_by_pod
  underground_routing: arterial_to_pod_branch
  last_mile:
    pod_hub: locker_bank + pickup_points
    to_building: bot_via_freight_elevator
  daily_volume: population * 1.5  # ~75,000 packages/day
```

### Bulk Delivery

```yaml
logistics.bulk_delivery:
  entry: perimeter_freight_staging
  routing: arterial_to_commercial_loading
  recipients:
    - grocery (daily restocking)
    - retail (scheduled)
    - restaurants (daily)
    - construction/maintenance (as needed)
    - residential_large_items (on demand)
  underground_commercial_access: direct_from_below
```

### Waste and Recycling

```yaml
logistics.waste_and_recycling:
  collection: building_chutes_to_underground
  vehicles: shared_with_bulk (return trips)
  processing: perimeter_facility
  streams: [trash, recycling, compost, hazardous]
  frequency: daily_per_pod
```

---

## Excavation and Construction

### Excavation Calculations

```
excavation_volume = city_area * depth
  city_area = 350 hectares = 3,500,000 m²
  depth = 8m (3 underground layers)
  volume = 28,000,000 m³

excavation_cost = volume * cost_per_m³
  cost_per_m³ ≈ $20-50 (favorable soil)
  estimated: ~$980M at $35/m³

platform_slab_cost = city_area * slab_cost_per_m²

total_underground = excavation + slab + waterproofing
                  + MEP + vehicle_infrastructure
```

### Underground Layers

```
Layer 1 (bottom): Sewage, water mains (gravity-dependent)
Layer 2 (middle):  Utility corridors (electrical, telecom)
Layer 3 (top):     Vehicle lanes (max height clearance)
```

### Per-Capita Cost Analysis

```
Total construction: $5-8B estimated
Per capita: $100,000-160,000

Conventional comparison:
  Suburban per-capita infrastructure: $30,000-50,000
  PLUS ongoing: car ownership, road maintenance,
    parking structures, gas infrastructure

Charter city front-loads cost but eliminates:
  car payments, gas, insurance, road maintenance,
  utility fragmentation, delivery fees
```

---

## Ownership and Revenue Model

### Ownership Specification

```yaml
ownership:
  model: city_owned_all_property
  residential: lease_only
  commercial: lease_only
  no_purchase: true
  no_subletting: true
```

### Revenue Specification

```yaml
revenue.residential:
  unit_types:
    studio:     {count_pct: 0.15, monthly_rent: TBD}
    one_bed:    {count_pct: 0.20, monthly_rent: TBD}
    two_bed:    {count_pct: 0.30, monthly_rent: TBD}
    three_bed:  {count_pct: 0.25, monthly_rent: TBD}
    four_bed:   {count_pct: 0.10, monthly_rent: TBD}
  includes:
    - utilities_all
    - delivery_all
    - waste_removal
    - amenity_access
    - maintenance
    - data_connectivity

revenue.commercial:
  lease_model: per_sqm_monthly
  includes:
    - utilities_all
    - freight_delivery
    - waste_removal
    - maintenance
  city_curates_mix: true
```

### Financial Equations

```
total_annual_cost = debt_service + operations + maintenance + reserves
required_revenue = total_annual_cost
average_rent = required_revenue / total_units / 12

At $6B over 30yr at 5%:
  debt_service ≈ $385M/year
  operations   ≈ $100M/year
  20,000 units → ~$2,000/month average to break even
```

### Phased Construction

```yaml
population.phasing:
  phase_1:
    zones: center + inner_middle
    population: ~15,000
    units: ~6,000
    self_sustaining: true
  phase_2:
    zones: outer_middle
    population: +15,000
    funded_by: phase_1_revenue
  phase_3:
    zones: edge
    population: +20,000
    funded_by: phase_1_2_revenue
```

---

## Community and Governance Specification

### Household Definition

```yaml
community.household_definition:
  max_generations: 3
  composition: direct_lineage_only
  max_occupants_per_unit: unit_bedroom_count + 1
```

### Clustering Prevention

```yaml
community.clustering_prevention:
  related_household_adjacency: not_guaranteed
  family_group_cap_per_pod: TBD  # needs legal review
  lease_holder: individual_household
  no_lease_transfer_to_family: true
```

### Governance (Open Items)

```yaml
governance:
  model: TBD  # requires social planning expert
  guardrails: TBD  # requires legal counsel
  principles:
    - demographic_diversity
    - no_cultural_enclaving
    - resident_voice_without_ownership
    - transparent_admission_criteria
  participation:
    - elected_pod_councils
    - city_council_elected_at_large
    - participatory_budgeting
    - referenda_on_major_decisions
  tenant_protections:
    - rent_increase_caps
    - eviction_protections
    - lease_renewal_rights
```

---

## Site Requirements Specification

```yaml
site_requirements:
  parcel:
    min_area_ha: 800
    terrain: flat
    water_table: below_excavation_depth
    soil: stable_non_bedrock
    seismic: low_risk

  regulatory:
    autonomous_governance: required
    property_model_freedom: required
    autonomous_vehicles: unrestricted_within_limits
    residency_criteria_control: required

  infrastructure_connections:
    electrical_grid: within_10km
    water_source: within_20km
    major_road: within_5km
    international_airport: within_100km

  climate:
    solar_irradiance: >4.5 kWh/m²/day
    flood_risk: low
    extreme_weather: minimal
```

### External Connections

The city connects to the outside world through five interfaces only:

- Electrical grid (bidirectional, backup import and surplus export)
- Water supply (municipal, well field, or reservoir)
- Sewage outfall (or closed-loop treatment)
- Data fiber (backbone connection)
- Physical road (freight staging at perimeter)

---

## Design Engine Architecture

The city is designed through specification, not drawing. A purpose-built engine reads the spec, solves constraints, and generates a navigable 3D model.

### Pipeline

**Stage 1 — Spec:** Declarative YAML/JSON input defining all city parameters.

**Stage 2 — Solver:** Constraint satisfaction engine. Generates pod layout within footprint, assigns rings by distance from center. Distributes population across pods respecting density and ring character. Validates every pod against service thresholds. Generates building footprints and heights within bowl envelope. Routes underground vehicle lanes, utility corridors, and pipe networks with capacity constraints. Calculates cost model.

**Stage 3 — Scene Graph:** Solver output organized into traversable 3D spatial structure. Every pipe, wall, lane, building, and path has position, dimension, and material assignment.

**Stage 4 — Renderer:** 3D visualization engine (comparable to game engines like Doom) consuming the scene graph for interactive exploration.

### Renderer Capabilities

- First-person walkthrough at street level
- Underground exploration of vehicle lanes, pipe runs, utility corridors
- Layer visibility toggling (water only, electrical only, vehicles only)
- Cross-section mode: slice the city vertically at any point
- Level-of-detail: zoom from city-wide overview to individual pipe junctions
- Bird's-eye and orbit camera modes

### Solver Algorithms

- Pod placement: circle packing or Voronoi tessellation within city footprint
- Building placement: procedural generation within pod boundaries respecting height envelope
- Infrastructure routing: capacity-constrained network flow for pipes and vehicle lanes
- Path routing: pedestrian/bike network connecting all pods with redundant routes
- Cost aggregation: bottom-up from materials and quantities to per-capita totals

### Technology Candidates

The renderer could be built on Three.js (web-based, TypeScript, shareable, no install required), Bevy (Rust, high performance), or Godot (full engine foundation). Three.js is the pragmatic choice given web accessibility and TypeScript alignment. The solver is pure computation with no graphics dependency and should be fully decoupled from the renderer, runnable headless for batch evaluation.

### Design Loop

```
1. Write/modify spec (YAML/JSON)
2. Run solver (validates, generates spatial data)
3. Load scene graph in renderer
4. Navigate and evaluate
5. Adjust spec values
6. Regenerate
```
