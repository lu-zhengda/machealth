package health

import "testing"

func TestParseICloud(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantSyncing  bool
		wantCaughtUp bool
		wantLastSync string
	}{
		{
			name:         "caught up",
			input:        `<com.apple.CloudDocs[1] foreground {client:idle server:full-sync sync:has-synced-down last-sync:2026-02-15 09:08:01.861, caught-up, token:abc}>`,
			wantSyncing:  false,
			wantCaughtUp: true,
			wantLastSync: "2026-02-15 09:08:01",
		},
		{
			name:         "needs sync",
			input:        `<com.apple.CloudDocs[1] foreground {client:needs-sync server:full-sync sync:oob-sync-ack last-sync:2026-02-14 12:00:00.000}>`,
			wantSyncing:  true,
			wantCaughtUp: false,
			wantLastSync: "2026-02-14 12:00:00",
		},
		{
			name:         "with ANSI codes",
			input:        "\033[0;1;32m<com.apple.CloudDocs[1] caught-up last-sync:2026-01-01 00:00:00.000>\033[0m",
			wantSyncing:  false,
			wantCaughtUp: true,
			wantLastSync: "2026-01-01 00:00:00",
		},
		{
			name:         "empty",
			input:        "",
			wantSyncing:  true,
			wantCaughtUp: false,
			wantLastSync: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			syncing, caughtUp, lastSync := parseICloud(tt.input)
			if syncing != tt.wantSyncing {
				t.Errorf("syncing = %v, want %v", syncing, tt.wantSyncing)
			}
			if caughtUp != tt.wantCaughtUp {
				t.Errorf("caughtUp = %v, want %v", caughtUp, tt.wantCaughtUp)
			}
			if lastSync != tt.wantLastSync {
				t.Errorf("lastSync = %q, want %q", lastSync, tt.wantLastSync)
			}
		})
	}
}
