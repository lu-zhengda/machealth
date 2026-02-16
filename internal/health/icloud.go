package health

import (
	"os/exec"
	"regexp"
	"strings"
)

// CheckICloud collects iCloud sync state.
func CheckICloud() ICloud {
	ic := ICloud{Status: StatusGreen, CaughtUp: true}

	out, err := exec.Command("/usr/bin/brctl", "status", "com.apple.CloudDocs").Output()
	if err != nil {
		// brctl not available or failed — assume OK
		return ic
	}

	ic.Syncing, ic.CaughtUp, ic.LastSync = parseICloud(string(out))

	if !ic.CaughtUp {
		ic.Status = StatusYellow
	}

	return ic
}

var (
	// Strip ANSI escape codes from brctl output
	ansiRe    = regexp.MustCompile(`\x1b\[[0-9;]*m`)
	syncTimeRe = regexp.MustCompile(`last-sync:(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2})`)
)

func parseICloud(s string) (syncing, caughtUp bool, lastSync string) {
	// Strip ANSI codes
	s = ansiRe.ReplaceAllString(s, "")

	caughtUp = strings.Contains(s, "caught-up")
	syncing = strings.Contains(s, "client:needs-sync") || !caughtUp

	if m := syncTimeRe.FindStringSubmatch(s); len(m) >= 2 {
		lastSync = m[1]
	}

	return
}

// DiagnoseICloud returns diagnosis for iCloud issues.
func DiagnoseICloud(ic ICloud) *Diagnosis {
	if ic.Status == StatusGreen {
		return nil
	}
	d := &Diagnosis{
		Subsystem: "icloud",
		Severity:  ic.Status,
		Summary:   "iCloud Drive is actively syncing",
		Detail:    "CloudDocs container is not caught up.",
		Action:    "Wait for sync to complete, or pause iCloud Drive in System Settings > Apple Account > iCloud",
	}
	if ic.LastSync != "" {
		d.Detail += " Last sync: " + ic.LastSync + "."
	}
	return d
}
