package health

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// CheckDisk collects disk space information.
func CheckDisk() Disk {
	d := Disk{Status: StatusGreen}

	// Use diskutil for accurate APFS container free space
	out, err := exec.Command("/usr/sbin/diskutil", "info", "/").Output()
	if err == nil {
		d.TotalGB, d.AvailableGB = parseDiskutil(string(out))
	}

	// Fallback to df if diskutil didn't work
	if d.TotalGB == 0 {
		out, err = exec.Command("/bin/df", "-P", "/").Output()
		if err != nil {
			d.Error = fmt.Sprintf("failed to get disk info: %v", err)
		} else {
			d.TotalGB, d.AvailableGB = parseDf(string(out))
		}
	}

	if d.TotalGB > 0 {
		d.UsedPercent = (1 - d.AvailableGB/d.TotalGB) * 100
	}

	// Determine status based on available percentage
	availPct := 0.0
	if d.TotalGB > 0 {
		availPct = (d.AvailableGB / d.TotalGB) * 100
	}
	switch {
	case availPct <= 10:
		d.Status = StatusRed
	case availPct <= 20:
		d.Status = StatusYellow
	}

	return d
}

var (
	containerTotalRe = regexp.MustCompile(`Container Total Space:\s+[\d.]+ \w+ \((\d+) Bytes\)`)
	containerFreeRe  = regexp.MustCompile(`Container Free Space:\s+[\d.]+ \w+ \((\d+) Bytes\)`)
)

func parseDiskutil(s string) (total, available float64) {
	if m := containerTotalRe.FindStringSubmatch(s); len(m) >= 2 {
		b, _ := strconv.ParseFloat(m[1], 64)
		total = b / (1024 * 1024 * 1024)
	}
	if m := containerFreeRe.FindStringSubmatch(s); len(m) >= 2 {
		b, _ := strconv.ParseFloat(m[1], 64)
		available = b / (1024 * 1024 * 1024)
	}
	return
}

func parseDf(s string) (total, available float64) {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	if len(lines) < 2 {
		return
	}
	fields := strings.Fields(lines[1])
	if len(fields) < 4 {
		return
	}
	// df -P reports in 512-byte blocks
	totalBlocks, _ := strconv.ParseFloat(fields[1], 64)
	availBlocks, _ := strconv.ParseFloat(fields[3], 64)
	total = totalBlocks * 512 / (1024 * 1024 * 1024)
	available = availBlocks * 512 / (1024 * 1024 * 1024)
	return
}

// DiagnoseDisk returns diagnosis for disk issues.
func DiagnoseDisk(d Disk) *Diagnosis {
	if d.Status == StatusGreen {
		return nil
	}
	diag := &Diagnosis{
		Subsystem: "disk",
		Severity:  d.Status,
	}
	if d.Status == StatusRed {
		diag.Summary = "Disk space critically low"
		diag.Detail = fmt.Sprintf("Only %.1f GB available of %.1f GB (%.1f%% used).", d.AvailableGB, d.TotalGB, d.UsedPercent)
		diag.Action = "Free disk space immediately with 'macbroom'. Remove large unused files, empty trash, clear caches"
	} else {
		diag.Summary = "Disk space is getting low"
		diag.Detail = fmt.Sprintf("%.1f GB available of %.1f GB (%.1f%% used).", d.AvailableGB, d.TotalGB, d.UsedPercent)
		diag.Action = "Consider running 'macbroom' to identify cleanup opportunities"
	}
	return diag
}
