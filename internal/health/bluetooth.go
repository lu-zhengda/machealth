package health

import (
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// CheckBluetooth collects Bluetooth controller state and connected-device info.
// It degrades gracefully: if system_profiler is unavailable or returns no useful
// data, Available is set to false and Status remains green (not a critical failure).
func CheckBluetooth() Bluetooth {
	bt := Bluetooth{Status: StatusGreen}

	out, err := exec.Command("/usr/sbin/system_profiler", "SPBluetoothDataType").Output()
	if err != nil || len(strings.TrimSpace(string(out))) == 0 {
		// system_profiler unavailable or returned nothing — degrade gracefully.
		bt.Available = false
		return bt
	}

	bt.Available = true
	parseBluetooth(string(out), &bt)

	// Status logic: Bluetooth is informational (read-only monitoring).
	// Never returns red to avoid false criticals.
	return bt
}

var (
	// macOS Sonoma/Sequoia: "State: On"
	btStateRe = regexp.MustCompile(`(?i)^\s+State:\s*(On|Off)\s*$`)
	// macOS Ventura and earlier: "Bluetooth Power: On"
	btPowerRe = regexp.MustCompile(`(?i)^\s+Bluetooth Power:\s*(On|Off)\s*$`)

	// Battery regexps covering multiple macOS key variants.
	btBatteryLineRe = regexp.MustCompile(`(?i)(?:Device batteryPercent|Battery Level|Left Battery Level|Right Battery Level|Case Battery Level):\s*(\d+)%`)
	// Connected property (old format): "Connected: Yes/No"
	btConnectedPropRe = regexp.MustCompile(`(?i)^\s+Connected:\s*(Yes|No)\s*$`)
)

// parseBluetooth fills bt from the text output of `system_profiler SPBluetoothDataType`.
// It handles two output layouts:
//   - Legacy (macOS ≤ Ventura): "Bluetooth Power: On" + "Devices (Paired...):" section
//     with per-device "Connected: Yes/No" property.
//   - Modern (macOS Sonoma+): "State: On" + top-level "Connected:" and "Not Connected:"
//     sections where connectivity is implied by section membership.
func parseBluetooth(s string, bt *Bluetooth) {
	lines := strings.Split(s, "\n")

	// ── Pass 1: detect power state ──────────────────────────────────────────
	for _, line := range lines {
		if m := btStateRe.FindStringSubmatch(line); len(m) >= 2 {
			bt.Enabled = strings.EqualFold(m[1], "on")
			break
		}
		if m := btPowerRe.FindStringSubmatch(line); len(m) >= 2 {
			bt.Enabled = strings.EqualFold(m[1], "on")
			break
		}
	}

	if !bt.Enabled {
		return
	}

	// ── Pass 2: detect layout ──────────────────────────────────────────────
	// Modern layout has top-level "Connected:" / "Not Connected:" section headers.
	hasModernSections := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "Connected:" || trimmed == "Not Connected:" {
			hasModernSections = true
			break
		}
	}

	if hasModernSections {
		parseBluetoothModern(lines, bt)
	} else {
		parseBluetoothLegacy(lines, bt)
	}
}

// parseBluetoothModern handles macOS Sonoma/Sequoia output where top-level section
// headers ("Connected:", "Not Connected:") determine device connectivity.
//
// Structure (indent guide):
//
//	      Bluetooth Controller:           ← 6 spaces
//	          State: On                   ← 10 spaces
//	      Connected:                      ← 6 spaces
//	          DeviceName:                 ← 10 spaces
//	              Left Battery Level: 80% ← 14 spaces
//	      Not Connected:                  ← 6 spaces
//	          OtherDevice:                ← 10 spaces
func parseBluetoothModern(lines []string, bt *Bluetooth) {
	type section int
	const (
		sectionOther        section = iota
		sectionConnected           // under "Connected:"
		sectionNotConnected        // under "Not Connected:"
	)

	currentSection := sectionOther
	var currentDevice *BluetoothDevice

	flushDevice := func() {
		if currentDevice != nil && currentDevice.Name != "" {
			bt.Devices = append(bt.Devices, *currentDevice)
			if currentDevice.Connected {
				bt.ConnectedDeviceCount++
			}
			currentDevice = nil
		}
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		indent := len(line) - len(strings.TrimLeft(line, " \t"))

		// Top-level section headers (indent ≈ 6)
		if indent <= 8 {
			switch trimmed {
			case "Connected:":
				flushDevice()
				currentSection = sectionConnected
				continue
			case "Not Connected:":
				flushDevice()
				currentSection = sectionNotConnected
				continue
			default:
				if strings.HasSuffix(trimmed, ":") {
					// e.g. "Bluetooth Controller:"
					flushDevice()
					currentSection = sectionOther
					continue
				}
			}
		}

		if currentSection == sectionOther {
			continue
		}

		// Device name lines (indent ≈ 10, ends with ":")
		if indent >= 9 && indent <= 13 && strings.HasSuffix(trimmed, ":") {
			flushDevice()
			name := strings.TrimSuffix(trimmed, ":")
			currentDevice = &BluetoothDevice{
				Name:           name,
				Connected:      currentSection == sectionConnected,
				BatteryPercent: -1,
			}
			continue
		}

		// Device property lines (indent ≥ 14)
		if currentDevice != nil && indent >= 14 {
			if m := btBatteryLineRe.FindStringSubmatch(line); len(m) >= 2 {
				if v, err := strconv.Atoi(m[1]); err == nil {
					// Keep highest battery value found (Left/Right/Case variants).
					if v > currentDevice.BatteryPercent {
						currentDevice.BatteryPercent = v
					}
				}
			}
		}
	}
	flushDevice()
}

// parseBluetoothLegacy handles macOS ≤ Ventura output where devices are listed under
// "Devices (Paired, Configured, etc.):" with a "Connected: Yes/No" property.
func parseBluetoothLegacy(lines []string, bt *Bluetooth) {
	inDevices := false
	var currentDevice *BluetoothDevice

	flushDevice := func() {
		if currentDevice != nil && currentDevice.Name != "" {
			bt.Devices = append(bt.Devices, *currentDevice)
			if currentDevice.Connected {
				bt.ConnectedDeviceCount++
			}
			currentDevice = nil
		}
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		indent := len(line) - len(strings.TrimLeft(line, " \t"))

		if strings.HasPrefix(trimmed, "Devices (") {
			inDevices = true
			continue
		}
		if !inDevices {
			continue
		}

		// New top-level section ends the devices block.
		if indent <= 6 && strings.HasSuffix(trimmed, ":") && !strings.HasPrefix(trimmed, "Devices") {
			break
		}

		// Device name lines (indent ≈ 8–12, ends with ":").
		if indent >= 8 && indent <= 12 && strings.HasSuffix(trimmed, ":") {
			flushDevice()
			name := strings.TrimSuffix(trimmed, ":")
			currentDevice = &BluetoothDevice{Name: name, BatteryPercent: -1}
			continue
		}

		// Device property lines.
		if currentDevice != nil {
			if m := btConnectedPropRe.FindStringSubmatch(line); len(m) >= 2 {
				currentDevice.Connected = strings.EqualFold(m[1], "yes")
			}
			if m := btBatteryLineRe.FindStringSubmatch(line); len(m) >= 2 {
				if v, err := strconv.Atoi(m[1]); err == nil {
					if v > currentDevice.BatteryPercent {
						currentDevice.BatteryPercent = v
					}
				}
			}
		}
	}
	flushDevice()
}

// DiagnoseBluetooth returns a diagnosis for Bluetooth issues.
// Currently Bluetooth is informational — it only produces diagnoses for
// yellow status; it never returns red severity to avoid false criticals.
func DiagnoseBluetooth(bt Bluetooth) *Diagnosis {
	if bt.Status == StatusGreen {
		return nil
	}
	return &Diagnosis{
		Subsystem: "bluetooth",
		Severity:  StatusYellow,
		Summary:   "Bluetooth status could not be fully determined",
		Detail:    bt.Error,
		Action:    "Check System Settings > Bluetooth for current state",
	}
}
