package validation

import "testing"

func TestNewReport(t *testing.T) {
	r := NewReport()
	if !r.Valid {
		t.Error("new report should be valid")
	}
	if len(r.Errors) != 0 || len(r.Warnings) != 0 || len(r.Info) != 0 {
		t.Error("new report should have empty slices")
	}
}

func TestAddError(t *testing.T) {
	r := NewReport()
	r.AddError(Result{
		Level:   LevelSchema,
		Message: "bad value",
	})
	if r.Valid {
		t.Error("report with error should be invalid")
	}
	if len(r.Errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(r.Errors))
	}
	if r.Errors[0].Severity != SeverityError {
		t.Error("AddError should set severity to error")
	}
	if r.Summary != "1 errors, 0 warnings, 0 info" {
		t.Errorf("unexpected summary: %s", r.Summary)
	}
}

func TestAddWarning(t *testing.T) {
	r := NewReport()
	r.AddWarning(Result{Level: LevelAnalytical, Message: "heads up"})
	if !r.Valid {
		t.Error("warnings should not invalidate report")
	}
	if len(r.Warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d", len(r.Warnings))
	}
	if r.Warnings[0].Severity != SeverityWarning {
		t.Error("AddWarning should set severity to warning")
	}
}

func TestAddInfo(t *testing.T) {
	r := NewReport()
	r.AddInfo(Result{Level: LevelAnalytical, Message: "fyi"})
	if !r.Valid {
		t.Error("info should not invalidate report")
	}
	if len(r.Info) != 1 {
		t.Fatalf("expected 1 info, got %d", len(r.Info))
	}
}

func TestMerge(t *testing.T) {
	r1 := NewReport()
	r1.AddWarning(Result{Level: LevelSchema, Message: "warn1"})

	r2 := NewReport()
	r2.AddError(Result{Level: LevelAnalytical, Message: "err1"})
	r2.AddWarning(Result{Level: LevelAnalytical, Message: "warn2"})
	r2.AddInfo(Result{Level: LevelAnalytical, Message: "info1"})

	r1.Merge(r2)

	if r1.Valid {
		t.Error("merged report should be invalid when other has errors")
	}
	if len(r1.Errors) != 1 {
		t.Errorf("expected 1 error, got %d", len(r1.Errors))
	}
	if len(r1.Warnings) != 2 {
		t.Errorf("expected 2 warnings, got %d", len(r1.Warnings))
	}
	if len(r1.Info) != 1 {
		t.Errorf("expected 1 info, got %d", len(r1.Info))
	}
	if r1.Summary != "1 errors, 2 warnings, 1 info" {
		t.Errorf("unexpected summary: %s", r1.Summary)
	}
}

func TestMergeValidIntoValid(t *testing.T) {
	r1 := NewReport()
	r2 := NewReport()
	r2.AddInfo(Result{Level: LevelSchema, Message: "note"})

	r1.Merge(r2)

	if !r1.Valid {
		t.Error("merging two valid reports should stay valid")
	}
	if len(r1.Info) != 1 {
		t.Errorf("expected 1 info, got %d", len(r1.Info))
	}
}
