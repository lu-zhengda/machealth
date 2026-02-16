package health

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// CheckBattery collects battery state.
func CheckBattery() Battery {
	b := Battery{Status: StatusGreen, TimeRemainingMin: -1, Percent: -1}

	// Quick check via pmset
	out, err := exec.Command("/usr/bin/pmset", "-g", "batt").Output()
	if err != nil {
		return b
	}
	parsePmsetBatt(string(out), &b)

	// Detailed info via ioreg
	out, err = exec.Command("/usr/sbin/ioreg", "-r", "-c", "AppleSmartBattery", "-w0").Output()
	if err == nil {
		parseIoregBatt(string(out), &b)
	}

	// No battery installed (desktop Mac) — always green
	if !b.Installed {
		b.Status = StatusGreen
		b.PowerSource = "ac"
		return b
	}

	// Determine status
	if b.PowerSource == "ac" {
		b.Status = StatusGreen
	} else {
		switch {
		case b.Percent <= 10:
			b.Status = StatusRed
		case b.Percent <= 20:
			b.Status = StatusYellow
		}
	}

	return b
}

var (
	powerSourceRe = regexp.MustCompile(`Now drawing from '([^']+)'`)
	battLineRe    = regexp.MustCompile(`(\d+)%;\s*(\w[\w\s]*)`)
	timeRemRe     = regexp.MustCompile(`(\d+):(\d+) remaining`)

	// Pre-compiled ioreg patterns for battery properties
	ioregCycleCountRe       = regexp.MustCompile(`"CycleCount"\s*=\s*(\d+)`)
	ioregDesignCapRe        = regexp.MustCompile(`"DesignCapacity"\s*=\s*(\d+)`)
	ioregNominalCapRe       = regexp.MustCompile(`"NominalChargeCapacity"\s*=\s*(\d+)`)
	ioregCurrentCapRe       = regexp.MustCompile(`"CurrentCapacity"\s*=\s*(\d+)`)
	ioregBattInstalledRe    = regexp.MustCompile(`"BatteryInstalled"\s*=\s*(\w+)`)
)

func parsePmsetBatt(s string, b *Battery) {
	if m := powerSourceRe.FindStringSubmatch(s); len(m) >= 2 {
		if strings.Contains(m[1], "AC") {
			b.PowerSource = "ac"
		} else {
			b.PowerSource = "battery"
		}
	}

	if m := battLineRe.FindStringSubmatch(s); len(m) >= 3 {
		b.Percent, _ = strconv.Atoi(m[1])
		b.Installed = true
		state := strings.TrimSpace(m[2])
		switch {
		case state == "charging" || state == "finishing charge":
			b.Charging = true
		case strings.Contains(state, "charged"):
			b.FullyCharged = true
		}
	}

	if m := timeRemRe.FindStringSubmatch(s); len(m) >= 3 {
		h, _ := strconv.Atoi(m[1])
		min, _ := strconv.Atoi(m[2])
		b.TimeRemainingMin = h*60 + min
	}
}

func parseIoregBatt(s string, b *Battery) {
	if m := ioregBattInstalledRe.FindStringSubmatch(s); len(m) >= 2 {
		b.Installed = strings.EqualFold(m[1], "yes") || strings.EqualFold(m[1], "true")
	}
	if !b.Installed {
		return
	}

	if m := ioregCycleCountRe.FindStringSubmatch(s); len(m) >= 2 {
		if v, err := strconv.Atoi(m[1]); err == nil {
			b.CycleCount = v
		}
	}

	var designCap, nominalCap int
	if m := ioregDesignCapRe.FindStringSubmatch(s); len(m) >= 2 {
		designCap, _ = strconv.Atoi(m[1])
	}
	if m := ioregNominalCapRe.FindStringSubmatch(s); len(m) >= 2 {
		nominalCap, _ = strconv.Atoi(m[1])
	}
	if designCap > 0 && nominalCap > 0 {
		b.HealthPercent = float64(nominalCap) / float64(designCap) * 100
	}

	if m := ioregCurrentCapRe.FindStringSubmatch(s); len(m) >= 2 {
		if v, err := strconv.Atoi(m[1]); err == nil {
			b.Percent = v
		}
	}
}

// DiagnoseBattery returns diagnosis for battery issues.
func DiagnoseBattery(b Battery) *Diagnosis {
	if b.Status == StatusGreen {
		return nil
	}
	d := &Diagnosis{
		Subsystem: "battery",
		Severity:  b.Status,
	}
	if b.Status == StatusRed {
		d.Summary = "Battery critically low"
		d.Detail = fmt.Sprintf("Battery at %d%% on battery power.", b.Percent)
		d.Action = "Connect to power immediately. System may shut down unexpectedly"
	} else {
		d.Summary = "Battery is low"
		d.Detail = fmt.Sprintf("Battery at %d%% on battery power.", b.Percent)
		if b.TimeRemainingMin > 0 {
			d.Detail += fmt.Sprintf(" Estimated %d minutes remaining.", b.TimeRemainingMin)
		}
		d.Action = "Connect to power before starting long-running tasks"
	}
	return d
}
