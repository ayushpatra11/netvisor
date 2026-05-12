package cli

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/ayushpatra11/netvisor/internal/netlinkwatcher"
	"github.com/ayushpatra11/netvisor/internal/server"
	"github.com/ayushpatra11/netvisor/internal/store"
	v1 "github.com/ayushpatra11/netvisor/proto/netvisor/v1"
	"github.com/spf13/cobra"
	"github.com/vishvananda/netlink"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func serveCmd(watcher *netlinkwatcher.Watcher, store *store.Store, logger *zap.Logger) *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Stream live network events from the kernel while connected via tcp connection",
		RunE: func(cmd *cobra.Command, args []string) error {
			// populate store with current interfaces
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

			// start the watcher and the server:
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

			// this is for the gRPC api
			grpcServer := grpc.NewServer()
			netvisorServer := server.New(store)
			v1.RegisterNetworkServiceServer(grpcServer, netvisorServer)
			reflection.Register(grpcServer)
			lis, err := net.Listen("tcp", ":50051")
			if err != nil {
				logger.Error("failed to listen", zap.Error(err))
				return err
			}

			go grpcServer.Serve(lis) //nolint:errcheck

			fmt.Fprintln(cmd.OutOrStdout(), "listening on port 50051...")
			quit := make(chan os.Signal, 1)
			signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
			<-quit // blocks here

			logger.Info("shutting down...")
			grpcServer.GracefulStop()
			return nil
		},
	}
}
