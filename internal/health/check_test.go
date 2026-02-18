package health

import (
	"testing"
	"time"
)

func TestCheck_Integration(t *testing.T) {
	start := time.Now()
	r := Check()
	elapsed := time.Since(start)

	// Must complete under 500ms
	if elapsed > 500*time.Millisecond {
		t.Errorf("Check() took %v, want < 500ms", elapsed)
	}

	// Timestamp should be recent
	if time.Since(r.Timestamp) > 5*time.Second {
		t.Error("timestamp should be recent")
	}

	// Score should be valid
	if r.Score.Status != StatusGreen && r.Score.Status != StatusYellow && r.Score.Status != StatusRed {
		t.Errorf("invalid score status: %s", r.Score.Status)
	}
	if r.Score.Value < 0 || r.Score.Value > 100 {
		t.Errorf("score value out of range: %d", r.Score.Value)
	}
	if r.Score.Reasons == nil {
		t.Error("reasons should not be nil")
	}
}

func TestComputeScore(t *testing.T) {
	tests := []struct {
		name       string
		report     Report
		wantStatus Status
		wantMin    int
		wantMax    int
	}{
		{
			name: "all green",
			report: Report{
				CPU:         CPU{Status: StatusGreen},
				Memory:      Memory{Status: StatusGreen},
				Disk:        Disk{Status: StatusGreen},
				Thermal:     Thermal{Status: StatusGreen},
				ICloud:      ICloud{Status: StatusGreen},
				Battery:     Battery{Status: StatusGreen},
				TimeMachine: TimeMachine{Status: StatusGreen},
				Network:     Network{Status: StatusGreen},
				Bluetooth:   Bluetooth{Status: StatusGreen},
			},
			wantStatus: StatusGreen,
			wantMin:    100, wantMax: 100,
		},
		{
			name: "one yellow",
			report: Report{
				CPU:         CPU{Status: StatusYellow},
				Memory:      Memory{Status: StatusGreen},
				Disk:        Disk{Status: StatusGreen},
				Thermal:     Thermal{Status: StatusGreen},
				ICloud:      ICloud{Status: StatusGreen},
				Battery:     Battery{Status: StatusGreen},
				TimeMachine: TimeMachine{Status: StatusGreen},
				Network:     Network{Status: StatusGreen},
				Bluetooth:   Bluetooth{Status: StatusGreen},
			},
			wantStatus: StatusYellow,
			wantMin:    80, wantMax: 99,
		},
		{
			name: "one red",
			report: Report{
				CPU:         CPU{Status: StatusGreen},
				Memory:      Memory{Status: StatusRed},
				Disk:        Disk{Status: StatusGreen},
				Thermal:     Thermal{Status: StatusGreen},
				ICloud:      ICloud{Status: StatusGreen},
				Battery:     Battery{Status: StatusGreen},
				TimeMachine: TimeMachine{Status: StatusGreen},
				Network:     Network{Status: StatusGreen},
				Bluetooth:   Bluetooth{Status: StatusGreen},
			},
			wantStatus: StatusRed,
			wantMin:    50, wantMax: 80,
		},
		{
			name: "all red",
			report: Report{
				CPU:         CPU{Status: StatusRed},
				Memory:      Memory{Status: StatusRed},
				Disk:        Disk{Status: StatusRed},
				Thermal:     Thermal{Status: StatusRed},
				ICloud:      ICloud{Status: StatusRed},
				Battery:     Battery{Status: StatusRed},
				TimeMachine: TimeMachine{Status: StatusRed},
				Network:     Network{Status: StatusRed},
				Bluetooth:   Bluetooth{Status: StatusRed},
			},
			wantStatus: StatusRed,
			wantMin:    0, wantMax: 0,
		},
		{
			name: "bluetooth red is capped at yellow in score",
			report: Report{
				CPU:         CPU{Status: StatusGreen},
				Memory:      Memory{Status: StatusGreen},
				Disk:        Disk{Status: StatusGreen},
				Thermal:     Thermal{Status: StatusGreen},
				ICloud:      ICloud{Status: StatusGreen},
				Battery:     Battery{Status: StatusGreen},
				TimeMachine: TimeMachine{Status: StatusGreen},
				Network:     Network{Status: StatusGreen},
				Bluetooth:   Bluetooth{Status: StatusRed},
			},
			// Bluetooth red is capped → overall status is yellow, not red.
			wantStatus: StatusYellow,
			wantMin:    80, wantMax: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := computeScore(tt.report)
			if score.Status != tt.wantStatus {
				t.Errorf("status = %s, want %s", score.Status, tt.wantStatus)
			}
			if score.Value < tt.wantMin || score.Value > tt.wantMax {
				t.Errorf("value = %d, want [%d, %d]", score.Value, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestExitCode(t *testing.T) {
	tests := []struct {
		status Status
		want   int
	}{
		{StatusGreen, 0},
		{StatusYellow, 1},
		{StatusRed, 2},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			r := Report{Score: Score{Status: tt.status}}
			if got := ExitCode(r); got != tt.want {
				t.Errorf("ExitCode() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestDiagnose_Integration(t *testing.T) {
	dr := Diagnose()

	if dr.Diagnoses == nil {
		t.Error("diagnoses should not be nil")
	}

	// If all green, diagnoses should be empty
	if dr.Score.Status == StatusGreen && len(dr.Diagnoses) != 0 {
		t.Errorf("expected 0 diagnoses for green status, got %d", len(dr.Diagnoses))
	}
}
