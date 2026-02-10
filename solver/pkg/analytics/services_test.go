package analytics

import "testing"

func TestResolveServicesDefaultCity(t *testing.T) {
	services := resolveServices(50000, 7000)

	serviceMap := make(map[string]ServiceCount)
	for _, s := range services {
		serviceMap[s.Service] = s
	}

	// Grocery: ceil(50000/4000) = 13
	if g, ok := serviceMap["grocery"]; !ok {
		t.Error("missing grocery service")
	} else if g.Required != 13 {
		t.Errorf("grocery count = %d, want 13", g.Required)
	}

	// Hospital: ceil(50000/50000) = 1
	if h, ok := serviceMap["hospital"]; !ok {
		t.Error("missing hospital service")
	} else if h.Required != 1 {
		t.Errorf("hospital count = %d, want 1", h.Required)
	}

	// Elementary school: ceil(7000/500) = 14
	if e, ok := serviceMap["elementary_school"]; !ok {
		t.Error("missing elementary_school service")
	} else if e.Required != 14 {
		t.Errorf("elementary_school count = %d, want 14", e.Required)
	}

	// Secondary school: ceil(7000/800) = 9
	if s, ok := serviceMap["secondary_school"]; !ok {
		t.Error("missing secondary_school service")
	} else if s.Required != 9 {
		t.Errorf("secondary_school count = %d, want 9", s.Required)
	}

	// Medical clinic: ceil(50000/10000) = 5
	if m, ok := serviceMap["medical_clinic"]; !ok {
		t.Error("missing medical_clinic service")
	} else if m.Required != 5 {
		t.Errorf("medical_clinic count = %d, want 5", m.Required)
	}

	// Library: ceil(50000/15000) = 4
	if l, ok := serviceMap["library"]; !ok {
		t.Error("missing library service")
	} else if l.Required != 4 {
		t.Errorf("library count = %d, want 4 (ceil(50000/15000))", l.Required)
	}
}

func TestResolveServicesZeroPopulation(t *testing.T) {
	services := resolveServices(0, 0)
	for _, s := range services {
		if s.Required != 0 {
			t.Errorf("%s required = %d, want 0 for zero population", s.Service, s.Required)
		}
	}
}

func TestResolveServicesBelowThreshold(t *testing.T) {
	// Population below hospital threshold but above 0
	services := resolveServices(30000, 5000)
	serviceMap := make(map[string]ServiceCount)
	for _, s := range services {
		serviceMap[s.Service] = s
	}

	// Hospital: ceil(30000/50000) = 1
	if h := serviceMap["hospital"]; h.Required != 1 {
		t.Errorf("hospital count = %d, want 1 (ceil rounds up)", h.Required)
	}
}
