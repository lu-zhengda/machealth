package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/lu-zhengda/machealth/internal/health"
)

var watchInterval time.Duration

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Continuously monitor system health",
	Long: `Continuously monitors system health at the specified interval.
Emits a JSON object per line (JSON Lines format) on each tick.
Use --human for a refreshing human-readable display.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

		enc := json.NewEncoder(os.Stdout)

		ticker := time.NewTicker(watchInterval)
		defer ticker.Stop()

		// Run immediately on start
		r := health.Check()
		exitCode = health.ExitCode(r)

		if humanFlag {
			fmt.Print("\033[2J\033[H")
			printHumanReport(r)
			fmt.Printf("\n(watching every %s, Ctrl+C to stop)\n", watchInterval)
		} else {
			if err := enc.Encode(r); err != nil {
				return fmt.Errorf("failed to encode JSON: %w", err)
			}
		}

		for {
			select {
			case <-sig:
				return nil
			case <-ticker.C:
				r = health.Check()
				exitCode = health.ExitCode(r)

				if humanFlag {
					fmt.Print("\033[2J\033[H")
					printHumanReport(r)
					fmt.Printf("\n(watching every %s, Ctrl+C to stop)\n", watchInterval)
				} else {
					if err := enc.Encode(r); err != nil {
						return fmt.Errorf("failed to encode JSON: %w", err)
					}
				}
			}
		}
	},
}

func init() {
	watchCmd.Flags().DurationVar(&watchInterval, "interval", 5*time.Second, "Check interval (e.g., 5s, 1m)")
	rootCmd.AddCommand(watchCmd)
}
