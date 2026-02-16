package health

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// CheckCPU collects CPU load information.
func CheckCPU() CPU {
	c := CPU{Status: StatusGreen}

	// Get logical core count
	out, err := exec.Command("/usr/sbin/sysctl", "-n", "hw.logicalcpu").Output()
	if err != nil {
		c.Error = fmt.Sprintf("failed to get cpu count: %v", err)
		c.LogicalCores = 1
	} else {
		c.LogicalCores, _ = strconv.Atoi(strings.TrimSpace(string(out)))
		if c.LogicalCores == 0 {
			c.LogicalCores = 1
		}
	}

	// Get load averages
	out, err = exec.Command("/usr/sbin/sysctl", "-n", "vm.loadavg").Output()
	if err != nil {
		c.Error = fmt.Sprintf("failed to get load averages: %v", err)
	} else {
		c.LoadAvg1m, c.LoadAvg5m, c.LoadAvg15m = parseLoadAvg(string(out))
	}

	c.LoadPerCore = c.LoadAvg1m / float64(c.LogicalCores)

	// Determine status based on load per core
	switch {
	case c.LoadPerCore >= 2.0:
		c.Status = StatusRed
	case c.LoadPerCore >= 0.8:
		c.Status = StatusYellow
	}

	return c
}

// parseLoadAvg parses "{ 5.47 6.54 6.97 }" format.
func parseLoadAvg(s string) (l1, l5, l15 float64) {
	s = strings.Trim(strings.TrimSpace(s), "{ }")
	fields := strings.Fields(s)
	if len(fields) >= 3 {
		l1, _ = strconv.ParseFloat(fields[0], 64)
		l5, _ = strconv.ParseFloat(fields[1], 64)
		l15, _ = strconv.ParseFloat(fields[2], 64)
	}
	return
}

// DiagnoseCPU returns diagnosis for CPU issues.
func DiagnoseCPU(c CPU) *Diagnosis {
	if c.Status == StatusGreen {
		return nil
	}
	d := &Diagnosis{
		Subsystem: "cpu",
		Severity:  c.Status,
	}
	if c.Status == StatusRed {
		d.Summary = "CPU is heavily loaded"
		d.Detail = fmt.Sprintf("Load per core is %.2f (1m avg: %.2f across %d cores). System may be unresponsive.", c.LoadPerCore, c.LoadAvg1m, c.LogicalCores)
		d.Action = "Check for runaway processes with 'pstop' and consider deferring heavy tasks"
	} else {
		d.Summary = "CPU load is elevated"
		d.Detail = fmt.Sprintf("Load per core is %.2f (1m avg: %.2f across %d cores).", c.LoadPerCore, c.LoadAvg1m, c.LogicalCores)
		d.Action = "Monitor load — it may settle. Avoid launching additional compute-heavy tasks"
	}
	return d
}
