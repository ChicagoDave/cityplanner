package analytics

import "math"

// serviceThreshold defines the population threshold per unit for a service type.
type serviceThreshold struct {
	threshold int
	metric    string // "persons" or "students"
}

// ServiceThresholds maps service names to their population thresholds.
// From the technical specification.
var ServiceThresholds = map[string]serviceThreshold{
	"grocery":           {4000, "persons"},
	"elementary_school": {500, "students"},
	"secondary_school":  {800, "students"},
	"medical_clinic":    {10000, "persons"},
	"hospital":          {50000, "persons"},
	"library":           {15000, "persons"},
	"pharmacy":          {8000, "persons"},
	"dental_clinic":     {5000, "persons"},
	"pediatric_clinic":  {10000, "persons"},
	"daycare":           {5000, "persons"},
}

// resolveServices computes the required count for each service type.
func resolveServices(totalPop int, totalStudents int) []ServiceCount {
	services := make([]ServiceCount, 0, len(ServiceThresholds))

	for name, st := range ServiceThresholds {
		relevantPop := totalPop
		if st.metric == "students" {
			relevantPop = totalStudents
		}

		required := 0
		if relevantPop > 0 && st.threshold > 0 {
			required = int(math.Ceil(float64(relevantPop) / float64(st.threshold)))
		}

		services = append(services, ServiceCount{
			Service:    name,
			Threshold:  st.threshold,
			Required:   required,
			Metric:     st.metric,
			Population: relevantPop,
		})
	}

	return services
}
