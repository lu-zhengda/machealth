package health

import "testing"

func TestParseLoadAvg(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantL1     float64
		wantL5     float64
		wantL15    float64
	}{
		{
			name:   "normal output",
			input:  "{ 5.47 6.54 6.97 }",
			wantL1: 5.47, wantL5: 6.54, wantL15: 6.97,
		},
		{
			name:   "with trailing newline",
			input:  "{ 1.23 4.56 7.89 }\n",
			wantL1: 1.23, wantL5: 4.56, wantL15: 7.89,
		},
		{
			name:   "zero load",
			input:  "{ 0.00 0.00 0.00 }",
			wantL1: 0, wantL5: 0, wantL15: 0,
		},
		{
			name:   "empty input",
			input:  "",
			wantL1: 0, wantL5: 0, wantL15: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l1, l5, l15 := parseLoadAvg(tt.input)
			if l1 != tt.wantL1 || l5 != tt.wantL5 || l15 != tt.wantL15 {
				t.Errorf("parseLoadAvg(%q) = (%.2f, %.2f, %.2f), want (%.2f, %.2f, %.2f)",
					tt.input, l1, l5, l15, tt.wantL1, tt.wantL5, tt.wantL15)
			}
		})
	}
}

func TestCheckCPU_Status(t *testing.T) {
	// Integration test — requires macOS
	c := CheckCPU()

	if c.LogicalCores <= 0 {
		t.Error("LogicalCores should be > 0")
	}
	if c.LoadAvg1m < 0 {
		t.Error("LoadAvg1m should be >= 0")
	}
	if c.LoadPerCore < 0 {
		t.Error("LoadPerCore should be >= 0")
	}
	if c.Status != StatusGreen && c.Status != StatusYellow && c.Status != StatusRed {
		t.Errorf("unexpected status: %s", c.Status)
	}
}

func TestDiagnoseCPU(t *testing.T) {
	tests := []struct {
		name     string
		cpu      CPU
		wantNil  bool
		wantSev  Status
	}{
		{
			name:    "green returns nil",
			cpu:     CPU{Status: StatusGreen},
			wantNil: true,
		},
		{
			name:    "yellow returns diagnosis",
			cpu:     CPU{Status: StatusYellow, LoadPerCore: 1.5, LoadAvg1m: 15.0, LogicalCores: 10},
			wantNil: false,
			wantSev: StatusYellow,
		},
		{
			name:    "red returns diagnosis",
			cpu:     CPU{Status: StatusRed, LoadPerCore: 3.0, LoadAvg1m: 30.0, LogicalCores: 10},
			wantNil: false,
			wantSev: StatusRed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := DiagnoseCPU(tt.cpu)
			if tt.wantNil {
				if d != nil {
					t.Error("expected nil diagnosis for green status")
				}
				return
			}
			if d == nil {
				t.Fatal("expected non-nil diagnosis")
			}
			if d.Severity != tt.wantSev {
				t.Errorf("severity = %s, want %s", d.Severity, tt.wantSev)
			}
			if d.Subsystem != "cpu" {
				t.Errorf("subsystem = %s, want cpu", d.Subsystem)
			}
		})
	}
}
