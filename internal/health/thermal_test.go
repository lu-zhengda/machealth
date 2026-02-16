package health

import "testing"

func TestParseThermal(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{
			name:  "no throttling recorded",
			input: "Note: No thermal warning level has been recorded\nNote: No CPU power status has been recorded\n",
			want:  100,
		},
		{
			name:  "throttled to 70%",
			input: "CPU_Scheduler_Limit  = 100\nCPU_Available_CPUs   = 8\nCPU_Speed_Limit      = 70\n",
			want:  70,
		},
		{
			name:  "throttled to 50%",
			input: "CPU_Speed_Limit      = 50\n",
			want:  50,
		},
		{
			name:  "full speed",
			input: "CPU_Speed_Limit      = 100\n",
			want:  100,
		},
		{
			name:  "empty",
			input: "",
			want:  100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseThermal(tt.input)
			if got != tt.want {
				t.Errorf("parseThermal() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestDiagnoseThermal(t *testing.T) {
	tests := []struct {
		name    string
		thermal Thermal
		wantNil bool
	}{
		{name: "green", thermal: Thermal{Status: StatusGreen, CPUSpeedLimit: 100}, wantNil: true},
		{name: "yellow", thermal: Thermal{Status: StatusYellow, CPUSpeedLimit: 85}, wantNil: false},
		{name: "red", thermal: Thermal{Status: StatusRed, CPUSpeedLimit: 60}, wantNil: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := DiagnoseThermal(tt.thermal)
			if tt.wantNil && d != nil {
				t.Error("expected nil")
			}
			if !tt.wantNil && d == nil {
				t.Error("expected non-nil")
			}
		})
	}
}
