package health

import "testing"

func TestParseScutil(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantReachable bool
		wantIface     string
		wantIP        string
	}{
		{
			name: "reachable with en0",
			input: `Network information

IPv4 network interface information
     en0 : flags      : 0x7 (IPv4,IPv6,DNS)
           address    : 192.168.4.159
           reach      : 0x00000002 (Reachable)

   REACH : flags 0x00000002 (Reachable)`,
			wantReachable: true, wantIface: "en0", wantIP: "192.168.4.159",
		},
		{
			name: "not reachable",
			input: `Network information

   REACH : flags 0x00000000 (Not Reachable)`,
			wantReachable: false, wantIface: "", wantIP: "",
		},
		{
			name:          "empty",
			input:         "",
			wantReachable: false, wantIface: "", wantIP: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reachable, iface, ip := parseScutil(tt.input)
			if reachable != tt.wantReachable {
				t.Errorf("reachable = %v, want %v", reachable, tt.wantReachable)
			}
			if iface != tt.wantIface {
				t.Errorf("iface = %q, want %q", iface, tt.wantIface)
			}
			if ip != tt.wantIP {
				t.Errorf("ip = %q, want %q", ip, tt.wantIP)
			}
		})
	}
}

func TestCheckNetwork_Integration(t *testing.T) {
	n := CheckNetwork()
	// On a dev machine, network should generally be up
	if n.Status != StatusGreen && n.Status != StatusRed {
		t.Errorf("unexpected status: %s", n.Status)
	}
}
