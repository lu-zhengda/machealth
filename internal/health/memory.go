package health

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// CheckMemory collects memory pressure information.
func CheckMemory() Memory {
	m := Memory{Status: StatusGreen}

	// Get memory pressure level (free %)
	out, err := exec.Command("/usr/sbin/sysctl", "-n", "kern.memorystatus_level").Output()
	if err != nil {
		m.Error = fmt.Sprintf("failed to get memory pressure: %v", err)
	} else {
		m.PressurePercent, _ = strconv.Atoi(strings.TrimSpace(string(out)))
	}

	// Get swap usage
	out, err = exec.Command("/usr/sbin/sysctl", "vm.swapusage").Output()
	if err == nil {
		m.SwapTotalMB, m.SwapUsedMB = parseSwap(string(out))
	}

	// Determine status
	switch {
	case m.PressurePercent <= 10:
		m.Status = StatusRed
	case m.PressurePercent <= 25:
		m.Status = StatusYellow
	}

	return m
}

var swapRe = regexp.MustCompile(`total\s*=\s*([\d.]+)M\s+used\s*=\s*([\d.]+)M`)

// parseSwap parses "vm.swapusage: total = 0.00M  used = 0.00M  free = 0.00M".
func parseSwap(s string) (total, used float64) {
	matches := swapRe.FindStringSubmatch(s)
	if len(matches) >= 3 {
		total, _ = strconv.ParseFloat(matches[1], 64)
		used, _ = strconv.ParseFloat(matches[2], 64)
	}
	return
}

// DiagnoseMemory returns diagnosis for memory issues.
func DiagnoseMemory(m Memory) *Diagnosis {
	if m.Status == StatusGreen {
		return nil
	}
	d := &Diagnosis{
		Subsystem: "memory",
		Severity:  m.Status,
	}
	if m.Status == StatusRed {
		d.Summary = "System is under critical memory pressure"
		d.Detail = fmt.Sprintf("Only %d%% memory free. Swap used: %.1f MB.", m.PressurePercent, m.SwapUsedMB)
		d.Action = "Close memory-intensive applications immediately. Check with 'pstop' for top consumers"
	} else {
		d.Summary = "Memory pressure is elevated"
		d.Detail = fmt.Sprintf("%d%% memory free. Swap used: %.1f MB.", m.PressurePercent, m.SwapUsedMB)
		d.Action = "Monitor memory usage. Consider closing unused applications before launching heavy tasks"
	}
	return d
}
