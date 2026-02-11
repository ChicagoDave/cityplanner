package layout

import "testing"

func TestGeneratePlazasProducesOutput(t *testing.T) {
	pods, _, _ := bikeTestPods(t)
	s := defaultBikeSpec()
	plazas, report := GeneratePlazas(pods, s)

	if len(plazas) == 0 {
		t.Fatal("expected plazas to be generated")
	}
	if !report.Valid {
		t.Fatalf("report has errors: %s", report.Summary)
	}
	t.Logf("generated %d plazas", len(plazas))
}

func TestOnePlazaPerPod(t *testing.T) {
	pods, _, _ := bikeTestPods(t)
	s := defaultBikeSpec()
	plazas, _ := GeneratePlazas(pods, s)

	if len(plazas) != len(pods) {
		t.Errorf("expected %d plazas (one per pod), got %d", len(pods), len(plazas))
	}
}

func TestPlazaSizesVaryByCharacter(t *testing.T) {
	pods, _, _ := bikeTestPods(t)
	s := defaultBikeSpec()
	plazas, _ := GeneratePlazas(pods, s)

	sizes := make(map[float64]int)
	for _, p := range plazas {
		sizes[p.Width]++
	}
	if len(sizes) < 2 {
		t.Errorf("expected varied plaza sizes, got %d distinct widths", len(sizes))
	}
	t.Logf("plaza sizes: %v", sizes)
}

func TestPlazaDimensions(t *testing.T) {
	pods, _, _ := bikeTestPods(t)
	s := defaultBikeSpec()
	plazas, _ := GeneratePlazas(pods, s)

	for _, p := range plazas {
		if p.Width < 20 || p.Width > 40 {
			t.Errorf("plaza %s width %.1f outside expected range [20, 40]", p.ID, p.Width)
		}
		if p.Depth < 20 || p.Depth > 40 {
			t.Errorf("plaza %s depth %.1f outside expected range [20, 40]", p.ID, p.Depth)
		}
	}
}
