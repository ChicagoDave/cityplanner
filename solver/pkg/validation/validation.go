package validation

import "fmt"

// Level indicates which validation stage produced the result.
type Level string

const (
	LevelSchema     Level = "schema"
	LevelAnalytical Level = "analytical"
	LevelSpatial    Level = "spatial"
)

// Severity indicates how critical a validation result is.
type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
	SeverityInfo    Severity = "info"
)

// Result is a single validation finding.
type Result struct {
	Level          Level    `json:"level"`
	Severity       Severity `json:"severity"`
	Message        string   `json:"message"`
	SpecPath       string   `json:"spec_path"`
	ActualValue    any      `json:"actual_value,omitempty"`
	Expected       string   `json:"expected,omitempty"`
	ConflictWith   string   `json:"conflict_with,omitempty"`
	Suggestions    []string `json:"suggestions,omitempty"`
}

// Report is the complete validation output.
type Report struct {
	Valid    bool     `json:"valid"`
	Errors   []Result `json:"errors"`
	Warnings []Result `json:"warnings"`
	Info     []Result `json:"info"`
	Summary  string   `json:"summary"`
}

// NewReport creates an empty valid report.
func NewReport() *Report {
	return &Report{
		Valid:    true,
		Errors:   []Result{},
		Warnings: []Result{},
		Info:     []Result{},
	}
}

// AddError adds an error result and marks the report invalid.
func (r *Report) AddError(result Result) {
	result.Severity = SeverityError
	r.Errors = append(r.Errors, result)
	r.Valid = false
	r.updateSummary()
}

// AddWarning adds a warning result.
func (r *Report) AddWarning(result Result) {
	result.Severity = SeverityWarning
	r.Warnings = append(r.Warnings, result)
	r.updateSummary()
}

// AddInfo adds an informational result.
func (r *Report) AddInfo(result Result) {
	result.Severity = SeverityInfo
	r.Info = append(r.Info, result)
	r.updateSummary()
}

// Merge combines another report into this one.
func (r *Report) Merge(other *Report) {
	r.Errors = append(r.Errors, other.Errors...)
	r.Warnings = append(r.Warnings, other.Warnings...)
	r.Info = append(r.Info, other.Info...)
	if !other.Valid {
		r.Valid = false
	}
	r.updateSummary()
}

func (r *Report) updateSummary() {
	r.Summary = fmt.Sprintf("%d errors, %d warnings, %d info",
		len(r.Errors), len(r.Warnings), len(r.Info))
}
