package health

import "testing"

func TestParseTmutil(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantRunning bool
		wantPhase   string
		wantPercent float64
	}{
		{
			name: "not running",
			input: `Backup session status:
{
    ClientID = "com.apple.backupd";
    Percent = "-1";
    Running = 0;
}`,
			wantRunning: false, wantPhase: "", wantPercent: -1,
		},
		{
			name: "running copying",
			input: `Backup session status:
{
    BackupPhase = "Copying";
    ClientID = "com.apple.backupd";
    Percent = "0.45";
    Running = 1;
}`,
			wantRunning: true, wantPhase: "Copying", wantPercent: 45,
		},
		{
			name: "running thinning",
			input: `Backup session status:
{
    BackupPhase = "ThinningPreBackup";
    Percent = "0.10";
    Running = 1;
}`,
			wantRunning: true, wantPhase: "ThinningPreBackup", wantPercent: 10,
		},
		{
			name:        "empty",
			input:       "",
			wantRunning: false, wantPhase: "", wantPercent: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			running, phase, pct := parseTmutil(tt.input)
			if running != tt.wantRunning {
				t.Errorf("running = %v, want %v", running, tt.wantRunning)
			}
			if phase != tt.wantPhase {
				t.Errorf("phase = %q, want %q", phase, tt.wantPhase)
			}
			if abs(pct-tt.wantPercent) > 0.1 {
				t.Errorf("percent = %.2f, want %.2f", pct, tt.wantPercent)
			}
		})
	}
}
