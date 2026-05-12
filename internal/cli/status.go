package cli

import (
	"fmt"

	"github.com/ayushpatra11/netvisor/internal/server"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func statusCmd(server *server.Server, grpcServer *grpc.Server, logger *zap.Logger) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Prints the status of the tool server",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(cmd.OutOrStdout(), "listening on port 50051...")
			return nil
		},
	}
}
