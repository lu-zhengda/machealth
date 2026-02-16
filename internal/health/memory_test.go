package health

import "testing"

func TestParseSwap(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantTotal float64
		wantUsed  float64
	}{
		{
			name:      "zero swap",
			input:     "vm.swapusage: total = 0.00M  used = 0.00M  free = 0.00M  (encrypted)",
			wantTotal: 0, wantUsed: 0,
		},
		{
			name:      "active swap",
			input:     "vm.swapusage: total = 2048.00M  used = 512.50M  free = 1535.50M  (encrypted)",
			wantTotal: 2048, wantUsed: 512.5,
		},
		{
			name:      "empty input",
			input:     "",
			wantTotal: 0, wantUsed: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			total, used := parseSwap(tt.input)
			if total != tt.wantTotal || used != tt.wantUsed {
				t.Errorf("parseSwap(%q) = (%.2f, %.2f), want (%.2f, %.2f)",
					tt.input, total, used, tt.wantTotal, tt.wantUsed)
			}
		})
	}
}

func TestCheckMemory_Status(t *testing.T) {
	m := CheckMemory()

	if m.PressurePercent < 0 || m.PressurePercent > 100 {
		t.Errorf("PressurePercent out of range: %d", m.PressurePercent)
	}
	if m.SwapUsedMB < 0 {
		t.Error("SwapUsedMB should be >= 0")
	}
}

func TestDiagnoseMemory(t *testing.T) {
	tests := []struct {
		name    string
		mem     Memory
		wantNil bool
	}{
		{name: "green", mem: Memory{Status: StatusGreen, PressurePercent: 90}, wantNil: true},
		{name: "yellow", mem: Memory{Status: StatusYellow, PressurePercent: 15}, wantNil: false},
		{name: "red", mem: Memory{Status: StatusRed, PressurePercent: 5}, wantNil: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := DiagnoseMemory(tt.mem)
			if tt.wantNil && d != nil {
				t.Error("expected nil")
			}
			if !tt.wantNil && d == nil {
				t.Error("expected non-nil")
			}
		})
	}
}
