// internal/cli/watch.go
package cli

import (
	"fmt"
	"text/tabwriter"

	"github.com/ayushpatra11/netvisor/internal/netlinkwatcher"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// watchCmd returns the `netvisor watch` command.
// It streams live kernel events to stdout until Ctrl+C.
func watchCmd(watcher *netlinkwatcher.Watcher, logger *zap.Logger) *cobra.Command {
	return &cobra.Command{
		Use:   "watch",
		Short: "Stream live network events from the kernel",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(cmd.OutOrStdout(), "Watching for network events (Ctrl+C to stop)...")
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 3, ' ', 0)
			fmt.Fprintln(w, "INDEX\tNAME\tMTU\tSTATE\tFLAGS")

			for event := range watcher.Events() {

				fmt.Fprintf(w, "%d\t%s\t%d\t%s\t%s\n",
					event.Index,
					event.LinkName,
					event.MTU,
					event.OperState,
					event.Flags,
				)
				w.Flush()
			}

			return nil
		},
	}
}
