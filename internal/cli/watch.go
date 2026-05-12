// internal/cli/watch.go
package cli

import (
	"context"
	"fmt"
	"text/tabwriter"

	"github.com/ayushpatra11/netvisor/internal/netlinkwatcher"
	"github.com/ayushpatra11/netvisor/internal/store"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// watchCmd returns the `netvisor watch` command.
// It streams live kernel events to stdout until Ctrl+C.
func watchCmd(watcher *netlinkwatcher.Watcher, store *store.Store, logger *zap.Logger) *cobra.Command {
	return &cobra.Command{
		Use:   "watch",
		Short: "Stream live network events from the kernel",
		RunE: func(cmd *cobra.Command, args []string) error {

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			if err := watcher.Start(ctx); err != nil {
				logger.Error("failed to start watcher", zap.Error(err))
				return err
			}

			// dispatcher — reads events from watcher, updates store
			go func() {
				for event := range watcher.Events() {
					switch event.Type {
					case netlinkwatcher.EventTypeAdded:
						store.Update(event)
					case netlinkwatcher.EventTypeDeleted:
						store.Delete(event.Index)
					}
				}
			}()
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
