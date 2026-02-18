package health

import "testing"

// Modern macOS (Sonoma/Sequoia) format with Connected: / Not Connected: sections.
const btSampleModern = `Bluetooth:

      Bluetooth Controller:
          Address: D0:11:E5:3A:9D:DF
          State: On
          Chipset: BCM_4388C2
          Discoverable: Off
          Firmware Version: 23.3.214.1296

      Connected:
          AirPods Pro:
              Address: 74:15:F5:4E:D0:50
              Left Battery Level: 88%
              Right Battery Level: 91%
              Case Battery Level: 74%

      Not Connected:
          Keychron K2:
              Address: DD:21:78:46:F3:D2
              Minor Type: Keyboard
          MX Master 3:
              Address: EE:F8:FA:CA:8A:28
              Minor Type: Mouse
`

// Legacy macOS (Ventura and earlier) format with "Devices (Paired...):" section.
const btSampleLegacy = `Bluetooth:

      Apple Bluetooth Software Version: 8.0.6d1
      Hardware, Features, and Settings:
          Address: F8:FF:C2:11:22:33
          Bluetooth Power: On
          Discoverable: Off

      Devices (Paired, Configured, etc.):
          AirPods Pro:
              Address: C4:CA:D9:AA:BB:CC
              Paired: Yes
              Connected: Yes
              Device batteryPercent: 79%
          MX Keys Mini:
              Address: D4:A3:3D:DD:EE:FF
              Paired: Yes
              Connected: Yes
          Magic Mouse:
              Address: 88:E9:FE:11:22:33
              Paired: Yes
              Connected: No
`

const btSampleDisabled = `Bluetooth:

      Bluetooth Controller:
          Address: D0:11:E5:3A:9D:DF
          State: Off
          Discoverable: Off
`

const btSampleLegacyDisabled = `Bluetooth:

      Hardware, Features, and Settings:
          Bluetooth Power: Off
          Discoverable: Off
`

func TestParseBluetooth_ModernEnabled(t *testing.T) {
	bt := Bluetooth{}
	parseBluetooth(btSampleModern, &bt)

	if !bt.Enabled {
		t.Error("Enabled should be true")
	}
	if bt.ConnectedDeviceCount != 1 {
		t.Errorf("ConnectedDeviceCount = %d, want 1", bt.ConnectedDeviceCount)
	}
	if len(bt.Devices) != 3 {
		t.Errorf("len(Devices) = %d, want 3", len(bt.Devices))
	}

	// AirPods Pro: connected, battery = max(88, 91, 74) = 91
	var airpods *BluetoothDevice
	for i := range bt.Devices {
		if bt.Devices[i].Name == "AirPods Pro" {
			airpods = &bt.Devices[i]
		}
	}
	if airpods == nil {
		t.Fatal("AirPods Pro not found in devices")
	}
	if !airpods.Connected {
		t.Error("AirPods Pro should be connected")
	}
	if airpods.BatteryPercent != 91 {
		t.Errorf("AirPods Pro battery = %d, want 91 (highest of 88/91/74)", airpods.BatteryPercent)
	}

	// Keychron K2: not connected, no battery
	var kbd *BluetoothDevice
	for i := range bt.Devices {
		if bt.Devices[i].Name == "Keychron K2" {
			kbd = &bt.Devices[i]
		}
	}
	if kbd == nil {
		t.Fatal("Keychron K2 not found in devices")
	}
	if kbd.Connected {
		t.Error("Keychron K2 should not be connected")
	}
	if kbd.BatteryPercent != -1 {
		t.Errorf("Keychron K2 battery should be -1 (not reported), got %d", kbd.BatteryPercent)
	}
}

func TestParseBluetooth_LegacyEnabled(t *testing.T) {
	bt := Bluetooth{}
	parseBluetooth(btSampleLegacy, &bt)

	if !bt.Enabled {
		t.Error("Enabled should be true")
	}
	if bt.ConnectedDeviceCount != 2 {
		t.Errorf("ConnectedDeviceCount = %d, want 2", bt.ConnectedDeviceCount)
	}
	if len(bt.Devices) != 3 {
		t.Errorf("len(Devices) = %d, want 3", len(bt.Devices))
	}

	// AirPods Pro: connected, battery 79%
	var airpods *BluetoothDevice
	for i := range bt.Devices {
		if bt.Devices[i].Name == "AirPods Pro" {
			airpods = &bt.Devices[i]
		}
	}
	if airpods == nil {
		t.Fatal("AirPods Pro not found in devices")
	}
	if !airpods.Connected {
		t.Error("AirPods Pro should be connected")
	}
	if airpods.BatteryPercent != 79 {
		t.Errorf("AirPods Pro battery = %d, want 79", airpods.BatteryPercent)
	}

	// MX Keys Mini: connected, no battery
	var keys *BluetoothDevice
	for i := range bt.Devices {
		if bt.Devices[i].Name == "MX Keys Mini" {
			keys = &bt.Devices[i]
		}
	}
	if keys == nil {
		t.Fatal("MX Keys Mini not found in devices")
	}
	if !keys.Connected {
		t.Error("MX Keys Mini should be connected")
	}
	if keys.BatteryPercent != -1 {
		t.Errorf("MX Keys Mini battery should be -1 (not reported), got %d", keys.BatteryPercent)
	}

	// Magic Mouse: not connected
	var mouse *BluetoothDevice
	for i := range bt.Devices {
		if bt.Devices[i].Name == "Magic Mouse" {
			mouse = &bt.Devices[i]
		}
	}
	if mouse == nil {
		t.Fatal("Magic Mouse not found in devices")
	}
	if mouse.Connected {
		t.Error("Magic Mouse should not be connected")
	}
}

func TestParseBluetooth_ModernDisabled(t *testing.T) {
	bt := Bluetooth{}
	parseBluetooth(btSampleDisabled, &bt)

	if bt.Enabled {
		t.Error("Enabled should be false")
	}
	if bt.ConnectedDeviceCount != 0 {
		t.Errorf("ConnectedDeviceCount = %d, want 0", bt.ConnectedDeviceCount)
	}
	if len(bt.Devices) != 0 {
		t.Errorf("len(Devices) = %d, want 0", len(bt.Devices))
	}
}

func TestParseBluetooth_LegacyDisabled(t *testing.T) {
	bt := Bluetooth{}
	parseBluetooth(btSampleLegacyDisabled, &bt)

	if bt.Enabled {
		t.Error("Enabled should be false")
	}
}

func TestParseBluetooth_BatteryVariant_BatteryLevel(t *testing.T) {
	// Some macOS versions use "Battery Level:" instead of "Device batteryPercent:".
	input := `Bluetooth:
      Hardware, Features, and Settings:
          Bluetooth Power: On
      Devices (Paired, Configured, etc.):
          Magic Keyboard:
              Connected: Yes
              Battery Level: 85%
`
	bt := Bluetooth{}
	parseBluetooth(input, &bt)

	if len(bt.Devices) == 0 {
		t.Fatal("expected at least one device")
	}
	found := false
	for _, d := range bt.Devices {
		if d.Name == "Magic Keyboard" {
			found = true
			if d.BatteryPercent != 85 {
				t.Errorf("BatteryPercent = %d, want 85", d.BatteryPercent)
			}
		}
	}
	if !found {
		t.Error("Magic Keyboard not found in devices")
	}
}

func TestCheckBluetooth_Integration(t *testing.T) {
	bt := CheckBluetooth()

	// Status must always be valid.
	if bt.Status != StatusGreen && bt.Status != StatusYellow && bt.Status != StatusRed {
		t.Errorf("unexpected status: %s", bt.Status)
	}

	// If unavailable, enabled must be false and connected count must be 0.
	if !bt.Available {
		if bt.Enabled {
			t.Error("Enabled should be false when not available")
		}
		if bt.ConnectedDeviceCount != 0 {
			t.Errorf("ConnectedDeviceCount should be 0 when not available, got %d", bt.ConnectedDeviceCount)
		}
	}

	// Bluetooth must never produce red — would cause false criticals.
	if bt.Status == StatusRed {
		t.Error("CheckBluetooth should never return StatusRed (would cause false criticals)")
	}
}

func TestDiagnoseBluetooth_Green(t *testing.T) {
	bt := Bluetooth{Status: StatusGreen, Available: true, Enabled: true}
	if d := DiagnoseBluetooth(bt); d != nil {
		t.Error("DiagnoseBluetooth should return nil for green status")
	}
}

func TestDiagnoseBluetooth_Unavailable(t *testing.T) {
	// Unavailable is still green — no diagnosis produced.
	bt := Bluetooth{Status: StatusGreen, Available: false}
	if d := DiagnoseBluetooth(bt); d != nil {
		t.Error("DiagnoseBluetooth should return nil for unavailable BT with green status")
	}
}

func TestDiagnoseBluetooth_Yellow(t *testing.T) {
	bt := Bluetooth{Status: StatusYellow, Error: "something went wrong"}
	d := DiagnoseBluetooth(bt)
	if d == nil {
		t.Fatal("DiagnoseBluetooth should return a diagnosis for yellow status")
	}
	if d.Severity != StatusYellow {
		t.Errorf("severity = %s, want yellow", d.Severity)
	}
	if d.Subsystem != "bluetooth" {
		t.Errorf("subsystem = %s, want bluetooth", d.Subsystem)
	}
}
