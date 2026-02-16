package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/lu-zhengda/machealth/internal/health"
)

var diagnoseCmd = &cobra.Command{
	Use:   "diagnose",
	Short: "Diagnose health issues with detailed explanations",
	Long:  "Runs all health checks and provides actionable diagnoses for any non-green subsystems.",
	RunE: func(cmd *cobra.Command, args []string) error {
		dr := health.Diagnose()
		exitCode = health.ExitCode(dr.Report)

		if humanFlag {
			printHumanDiagnose(dr)
			return nil
		}
		return printJSON(dr)
	},
}

func init() {
	rootCmd.AddCommand(diagnoseCmd)
}

func printHumanDiagnose(dr health.DiagnoseReport) {
	printHumanReport(dr.Report)

	if len(dr.Diagnoses) == 0 {
		fmt.Println("\nNo issues detected.")
		return
	}

	fmt.Printf("\n--- Diagnoses (%d) ---\n\n", len(dr.Diagnoses))
	for i, d := range dr.Diagnoses {
		fmt.Printf("%d. [%s] %s: %s\n", i+1, strings.ToUpper(string(d.Severity)), d.Subsystem, d.Summary)
		fmt.Printf("   Detail: %s\n", d.Detail)
		fmt.Printf("   Action: %s\n\n", d.Action)
	}
}
