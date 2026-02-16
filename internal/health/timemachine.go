package health

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
)

// CheckTimeMachine collects Time Machine backup state.
func CheckTimeMachine() TimeMachine {
	tm := TimeMachine{Status: StatusGreen, Percent: -1}

	out, err := exec.Command("/usr/bin/tmutil", "status").Output()
	if err != nil {
		return tm
	}

	tm.Running, tm.Phase, tm.Percent = parseTmutil(string(out))

	if tm.Running {
		tm.Status = StatusYellow
	}

	return tm
}

var (
	tmRunningRe = regexp.MustCompile(`Running\s*=\s*(\d)`)
	tmPhaseRe   = regexp.MustCompile(`BackupPhase\s*=\s*"([^"]+)"`)
	tmPercentRe = regexp.MustCompile(`Percent\s*=\s*"([\d.e-]+)"`)
)

func parseTmutil(s string) (running bool, phase string, percent float64) {
	percent = -1

	if m := tmRunningRe.FindStringSubmatch(s); len(m) >= 2 {
		running = m[1] == "1"
	}

	if m := tmPhaseRe.FindStringSubmatch(s); len(m) >= 2 {
		phase = m[1]
	}

	if m := tmPercentRe.FindStringSubmatch(s); len(m) >= 2 {
		v, err := strconv.ParseFloat(m[1], 64)
		if err == nil && v >= 0 {
			percent = v * 100 // Convert 0.0-1.0 to percentage
		}
	}

	// If not running, phase is clear
	if !running {
		phase = ""
	}

	return
}

// DiagnoseTimeMachine returns diagnosis for Time Machine issues.
func DiagnoseTimeMachine(tm TimeMachine) *Diagnosis {
	if tm.Status == StatusGreen {
		return nil
	}
	d := &Diagnosis{
		Subsystem: "timemachine",
		Severity:  tm.Status,
		Summary:   "Time Machine backup in progress",
	}

	detail := "A Time Machine backup is running"
	if tm.Phase != "" {
		detail += fmt.Sprintf(" (phase: %s)", tm.Phase)
	}
	if tm.Percent >= 0 {
		detail += fmt.Sprintf(", %.0f%% complete", tm.Percent)
	}
	detail += ". This may cause elevated disk I/O."
	d.Detail = detail
	d.Action = "Wait for backup to complete or defer heavy disk I/O tasks"

	return d
}
