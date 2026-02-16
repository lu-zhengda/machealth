package health

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// CheckNetwork collects network connectivity state.
func CheckNetwork() Network {
	n := Network{Status: StatusGreen}

	// Check network interface info via scutil
	out, err := exec.Command("/usr/sbin/scutil", "--nwi").Output()
	if err != nil {
		n.Status = StatusRed
		n.Error = fmt.Sprintf("failed to check network: %v", err)
		return n
	}

	n.Reachable, n.Interface, n.IP = parseScutil(string(out))

	if !n.Reachable {
		n.Status = StatusRed
	}

	return n
}

var (
	nwiInterfaceRe = regexp.MustCompile(`(\w+)\s*:\s*flags\s*:\s*\S+\s*\(.*IPv4`)
	nwiAddressRe   = regexp.MustCompile(`address\s*:\s*([\d.]+)`)
	nwiReachRe     = regexp.MustCompile(`REACH\s*:\s*flags\s*\S+\s*\((\w+)\)`)
)

func parseScutil(s string) (reachable bool, iface, ip string) {
	// Check global reachability
	if m := nwiReachRe.FindStringSubmatch(s); len(m) >= 2 {
		reachable = strings.Contains(m[1], "Reachable")
	}

	// Get primary interface
	if m := nwiInterfaceRe.FindStringSubmatch(s); len(m) >= 2 {
		iface = m[1]
	}

	// Get IP from the interface section
	if m := nwiAddressRe.FindStringSubmatch(s); len(m) >= 2 {
		ip = m[1]
	}

	return
}

// DiagnoseNetwork returns diagnosis for network issues.
func DiagnoseNetwork(n Network) *Diagnosis {
	if n.Status == StatusGreen {
		return nil
	}
	return &Diagnosis{
		Subsystem: "network",
		Severity:  StatusRed,
		Summary:   "Network is unreachable",
		Detail:    "No active network interface with reachability detected.",
		Action:    "Check Wi-Fi or Ethernet connection. Run 'netwhiz diagnose' for detailed network diagnostics",
	}
}
