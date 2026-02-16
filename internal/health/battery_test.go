package health

import "testing"

func TestParsePmsetBatt(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantSource  string
		wantPercent int
		wantCharge  bool
		wantFull    bool
	}{
		{
			name: "AC fully charged",
			input: `Now drawing from 'AC Power'
 -InternalBattery-0 (id=21168227)	100%; charged; 0:00 remaining present: true`,
			wantSource: "ac", wantPercent: 100, wantFull: true,
		},
		{
			name: "battery discharging",
			input: `Now drawing from 'Battery Power'
 -InternalBattery-0 (id=21168227)	75%; discharging; 3:45 remaining present: true`,
			wantSource: "battery", wantPercent: 75,
		},
		{
			name: "AC charging",
			input: `Now drawing from 'AC Power'
 -InternalBattery-0 (id=21168227)	45%; charging; 1:30 remaining present: true`,
			wantSource: "ac", wantPercent: 45, wantCharge: true,
		},
		{
			name: "not charging (optimized battery)",
			input: `Now drawing from 'AC Power'
 -InternalBattery-0 (id=21168227)	80%; not charging; present: true`,
			wantSource: "ac", wantPercent: 80, wantCharge: false, wantFull: false,
		},
		{
			name: "finishing charge",
			input: `Now drawing from 'AC Power'
 -InternalBattery-0 (id=21168227)	99%; finishing charge; present: true`,
			wantSource: "ac", wantPercent: 99, wantCharge: true, wantFull: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := Battery{TimeRemainingMin: -1, Percent: -1}
			parsePmsetBatt(tt.input, &b)
			if b.PowerSource != tt.wantSource {
				t.Errorf("PowerSource = %s, want %s", b.PowerSource, tt.wantSource)
			}
			if b.Percent != tt.wantPercent {
				t.Errorf("Percent = %d, want %d", b.Percent, tt.wantPercent)
			}
			if b.Charging != tt.wantCharge {
				t.Errorf("Charging = %v, want %v", b.Charging, tt.wantCharge)
			}
			if b.FullyCharged != tt.wantFull {
				t.Errorf("FullyCharged = %v, want %v", b.FullyCharged, tt.wantFull)
			}
		})
	}
}

func TestParseIoregBatt(t *testing.T) {
	input := `"BatteryInstalled" = Yes
"CycleCount" = 351
"DesignCapacity" = 6075
"NominalChargeCapacity" = 5225
"CurrentCapacity" = 100`

	b := Battery{}
	parseIoregBatt(input, &b)

	if !b.Installed {
		t.Error("Installed should be true")
	}
	if b.CycleCount != 351 {
		t.Errorf("CycleCount = %d, want 351", b.CycleCount)
	}
	if b.Percent != 100 {
		t.Errorf("Percent = %d, want 100", b.Percent)
	}
	expectedHealth := float64(5225) / float64(6075) * 100
	if abs(b.HealthPercent-expectedHealth) > 0.01 {
		t.Errorf("HealthPercent = %.2f, want %.2f", b.HealthPercent, expectedHealth)
	}
}

func TestParseIoregBatt_NoBattery(t *testing.T) {
	input := `"BatteryInstalled" = No`

	b := Battery{Percent: 50}
	parseIoregBatt(input, &b)

	if b.Installed {
		t.Error("Installed should be false")
	}
	// Percent should remain unchanged since we return early
	if b.Percent != 50 {
		t.Errorf("Percent should not be modified when no battery, got %d", b.Percent)
	}
}
