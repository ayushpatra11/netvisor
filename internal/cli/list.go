// internal/cli/list.go
package cli

import (
	"fmt"
	"text/tabwriter"

	"github.com/ayushpatra11/netvisor/internal/store"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// listCmd returns the `netvisor list interfaces` command.
func listCmd(s *store.Store, logger *zap.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List current network interfaces",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: get all interfaces from the store

			interfaceList := s.List()

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
