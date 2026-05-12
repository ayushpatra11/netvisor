// internal/cli/list.go
package cli

import (
	"fmt"
	"text/tabwriter"

	"github.com/ayushpatra11/netvisor/internal/netlinkwatcher"
	"github.com/ayushpatra11/netvisor/internal/store"
	"github.com/spf13/cobra"
	"github.com/vishvananda/netlink"
	"go.uber.org/zap"
)

// listCmd returns the `netvisor list interfaces` command.
func listCmd(_ *netlinkwatcher.Watcher, store *store.Store, logger *zap.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List current network interfaces",
		RunE: func(cmd *cobra.Command, args []string) error {
			// populate store with current interfaces before handling commands
			links, err := netlink.LinkList()
			if err != nil {
				logger.Error("failed to list initial interfaces", zap.Error(err))
				return err
			}
			for _, link := range links {
				attrs := link.Attrs()
				store.Update(netlinkwatcher.Event{
					Type:      netlinkwatcher.EventTypeAdded,
					Index:     attrs.Index,
					LinkName:  attrs.Name,
					MTU:       attrs.MTU,
					Flags:     attrs.Flags,
					OperState: attrs.OperState,
				})
			}

			interfaceList := store.List()

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 3, ' ', 0)

			fmt.Fprintln(w, "INDEX\tNAME\tMTU\tSTATE\tFLAGS")

			for _, interfaceInstance := range interfaceList {
				fmt.Fprintf(w, "%d\t%s\t%d\t%s\t%s\n",
					interfaceInstance.Index,
					interfaceInstance.LinkName,
					interfaceInstance.MTU,
					interfaceInstance.OperState,
					interfaceInstance.Flags,
				)
			}

			w.Flush()

			return nil
		},
	}
	return cmd
}
