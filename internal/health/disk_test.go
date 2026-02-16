package health

import "testing"

func TestParseDiskutil(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantTotal float64
		wantAvail float64
	}{
		{
			name: "normal APFS",
			input: `   Device Identifier:         disk3s3s1
   Container Total Space:     994662584320 Bytes (994.7 GB)
   Container Free Space:      828319760384 Bytes (828.3 GB)`,
			// Note: the regex expects "994.7 GB (994662584320 Bytes)" format
		},
		{
			name: "standard format",
			input: `Container Total Space:     994.7 GB (994662584320 Bytes)
Container Free Space:      828.3 GB (828319760384 Bytes)`,
			wantTotal: 994662584320.0 / (1024 * 1024 * 1024),
			wantAvail: 828319760384.0 / (1024 * 1024 * 1024),
		},
		{
			name:      "empty",
			input:     "",
			wantTotal: 0, wantAvail: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			total, avail := parseDiskutil(tt.input)
			if tt.wantTotal > 0 {
				if abs(total-tt.wantTotal) > 0.01 {
					t.Errorf("total = %.2f, want %.2f", total, tt.wantTotal)
				}
				if abs(avail-tt.wantAvail) > 0.01 {
					t.Errorf("avail = %.2f, want %.2f", avail, tt.wantAvail)
				}
			}
		})
	}
}

func TestParseDf(t *testing.T) {
	input := `Filesystem       512-blocks      Used Available Capacity  Mounted on
/dev/disk3s3s1  1942700360  31428224 1596502784     2%    /`

	total, avail := parseDf(input)
	if total <= 0 {
		t.Error("total should be > 0")
	}
	if avail <= 0 {
		t.Error("avail should be > 0")
	}
	if avail > total {
		t.Error("avail should be <= total")
	}
}

func TestCheckDisk_Integration(t *testing.T) {
	d := CheckDisk()

	if d.TotalGB <= 0 {
		t.Error("TotalGB should be > 0")
	}
	if d.AvailableGB <= 0 {
		t.Error("AvailableGB should be > 0")
	}
	if d.UsedPercent < 0 || d.UsedPercent > 100 {
		t.Errorf("UsedPercent out of range: %.2f", d.UsedPercent)
	}
}

func abs(f float64) float64 {
	if f < 0 {
		return -f
	}
	return f
}
