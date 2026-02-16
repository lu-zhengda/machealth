package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/lu-zhengda/machealth/internal/health"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Run a one-shot system health check",
	Long:  "Runs all health checks in parallel and returns a composite health report.",
	RunE: func(cmd *cobra.Command, args []string) error {
		r := health.Check()
		exitCode = health.ExitCode(r)

		if humanFlag {
			printHumanReport(r)
			return nil
		}
		return printJSON(r)
	},
}

func init() {
	rootCmd.AddCommand(checkCmd)
}

func printHumanReport(r health.Report) {
	icon := statusIcon(r.Score.Status)
	fmt.Printf("%s System Health: %s (score: %d/100)\n\n", icon, strings.ToUpper(string(r.Score.Status)), r.Score.Value)

	printSubsystem("CPU", r.CPU.Status, fmt.Sprintf("Load: %.2f/%.2f/%.2f (%.2f per core, %d cores)",
		r.CPU.LoadAvg1m, r.CPU.LoadAvg5m, r.CPU.LoadAvg15m, r.CPU.LoadPerCore, r.CPU.LogicalCores))

	printSubsystem("Memory", r.Memory.Status, fmt.Sprintf("Free: %d%%, Swap: %.1f/%.1f MB",
		r.Memory.PressurePercent, r.Memory.SwapUsedMB, r.Memory.SwapTotalMB))

	printSubsystem("Disk", r.Disk.Status, fmt.Sprintf("Available: %.1f/%.1f GB (%.1f%% used)",
		r.Disk.AvailableGB, r.Disk.TotalGB, r.Disk.UsedPercent))

	printSubsystem("Thermal", r.Thermal.Status, fmt.Sprintf("CPU speed limit: %d%%", r.Thermal.CPUSpeedLimit))

	icloudDetail := "Caught up"
	if !r.ICloud.CaughtUp {
		icloudDetail = "Syncing"
	}
	if r.ICloud.LastSync != "" {
		icloudDetail += " (last: " + r.ICloud.LastSync + ")"
	}
	printSubsystem("iCloud", r.ICloud.Status, icloudDetail)

	battDetail := fmt.Sprintf("%d%%, %s", r.Battery.Percent, r.Battery.PowerSource)
	if !r.Battery.Installed {
		battDetail = "Not installed (desktop Mac)"
	} else if r.Battery.Charging {
		battDetail += ", charging"
	} else if r.Battery.FullyCharged {
		battDetail += ", fully charged"
	}
	printSubsystem("Battery", r.Battery.Status, battDetail)

	tmDetail := "Idle"
	if r.TimeMachine.Running {
		tmDetail = "Running"
		if r.TimeMachine.Phase != "" {
			tmDetail += " (" + r.TimeMachine.Phase + ")"
		}
		if r.TimeMachine.Percent >= 0 {
			tmDetail += fmt.Sprintf(" %.0f%%", r.TimeMachine.Percent)
		}
	}
	printSubsystem("Time Machine", r.TimeMachine.Status, tmDetail)

	netDetail := "Reachable"
	if !r.Network.Reachable {
		netDetail = "Unreachable"
	} else if r.Network.Interface != "" {
		netDetail += " via " + r.Network.Interface
		if r.Network.IP != "" {
			netDetail += " (" + r.Network.IP + ")"
		}
	}
	printSubsystem("Network", r.Network.Status, netDetail)

	if len(r.Score.Reasons) > 0 {
		fmt.Printf("\nReasons: %s\n", strings.Join(r.Score.Reasons, ", "))
	}
}

func printSubsystem(name string, status health.Status, detail string) {
	fmt.Printf("  %s %-14s %s\n", statusIcon(status), name, detail)
}

func statusIcon(s health.Status) string {
	switch s {
	case health.StatusGreen:
		return "[OK]"
	case health.StatusYellow:
		return "[!!]"
	case health.StatusRed:
		return "[XX]"
	default:
		return "[??]"
	}
}
