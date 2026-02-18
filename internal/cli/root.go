package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// version is set via ldflags at build time.
	version = "dev"

	humanFlag bool
	jsonFlag  bool
)

var rootCmd = &cobra.Command{
	Use:   "machealth",
	Short: "macOS system health checker for AI agents",
	Long: `machealth is a unified macOS system health checker designed for AI agent
consumption. It provides a single-call health assessment including CPU load,
memory pressure, disk space, thermal state, iCloud sync, battery, Time Machine,
and network connectivity.

JSON output is the default. Use --human for human-readable output.
Exit codes: 0=healthy, 1=degraded, 2=critical.`,
	Version: version,
	RunE: func(cmd *cobra.Command, args []string) error {
		if shell, _ := cmd.Flags().GetString("generate-completion"); shell != "" {
			switch shell {
			case "bash":
				return cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				return cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				return cmd.Root().GenFishCompletion(os.Stdout, true)
			default:
				return fmt.Errorf("unsupported shell: %s (use bash, zsh, or fish)", shell)
			}
		}
		// Default action is check
		return checkCmd.RunE(cmd, args)
	},
}

// exitCode is set by commands to indicate health status.
var exitCode int

// Execute runs the root command and returns the exit code.
func Execute() (int, error) {
	err := rootCmd.Execute()
	return exitCode, err
}

func init() {
	rootCmd.SetVersionTemplate(fmt.Sprintf("machealth %s\n", version))
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.Flags().String("generate-completion", "", "Generate shell completion (bash, zsh, fish)")
	rootCmd.Flags().MarkHidden("generate-completion")
	rootCmd.PersistentFlags().BoolVar(&humanFlag, "human", false, "Output in human-readable format (default is JSON)")
	rootCmd.PersistentFlags().BoolVar(&jsonFlag, "json", false, "Output in JSON format (default; explicit form of the default)")
}
