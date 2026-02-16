package health

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// CheckThermal collects thermal throttling information.
func CheckThermal() Thermal {
	t := Thermal{Status: StatusGreen, CPUSpeedLimit: 100}

	out, err := exec.Command("/usr/bin/pmset", "-g", "therm").Output()
	if err != nil {
		return t
	}

	t.CPUSpeedLimit = parseThermal(string(out))
	t.Throttled = t.CPUSpeedLimit < 100

	switch {
	case t.CPUSpeedLimit < 80:
		t.Status = StatusRed
	case t.CPUSpeedLimit < 100:
		t.Status = StatusYellow
	}

	return t
}

var speedLimitRe = regexp.MustCompile(`CPU_Speed_Limit\s*=\s*(\d+)`)

func parseThermal(s string) int {
	// If "No thermal warning" is present, no throttling
	if strings.Contains(s, "No thermal warning level has been recorded") ||
		strings.Contains(s, "No CPU power status has been recorded") {
		return 100
	}
	m := speedLimitRe.FindStringSubmatch(s)
	if len(m) >= 2 {
		v, _ := strconv.Atoi(m[1])
		if v > 0 {
			return v
		}
	}
	return 100
}

// DiagnoseThermal returns diagnosis for thermal issues.
func DiagnoseThermal(t Thermal) *Diagnosis {
	if t.Status == StatusGreen {
		return nil
	}
	d := &Diagnosis{
		Subsystem: "thermal",
		Severity:  t.Status,
	}
	if t.Status == StatusRed {
		d.Summary = "CPU is severely thermally throttled"
		d.Detail = fmt.Sprintf("CPU speed limited to %d%% of maximum.", t.CPUSpeedLimit)
		d.Action = "Reduce workload, improve ventilation, or wait for the system to cool down"
	} else {
		d.Summary = "CPU is being thermally throttled"
		d.Detail = fmt.Sprintf("CPU speed limited to %d%% of maximum.", t.CPUSpeedLimit)
		d.Action = "Monitor thermal state. Avoid launching additional compute-heavy tasks"
	}
	return d
}
