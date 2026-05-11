// Package cli wires together all cobra commands.
// main.go stays thin — all CLI logic lives here.
package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/ayushpatra11/netvisor/internal/netlinkwatcher"
	"github.com/ayushpatra11/netvisor/internal/store"
	"github.com/spf13/cobra"
	"github.com/vishvananda/netlink"
	"go.uber.org/zap"
)

// Run is the single entry point called by main.
// It returns an exit code so main can call os.Exit cleanly.
func Run(version string) int {
	// zap.NewDevelopment() gives human-readable output; Production() gives JSON.
	logger, err := zap.NewDevelopment()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to init logger: %v\n", err)
		return 1
	}
	// zap uses an internal buffer — Sync() flushes it before we exit.
	defer logger.Sync() //nolint:errcheck

	watcherInstance := netlinkwatcher.New(logger)
	storeInstance := store.New()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err = watcherInstance.Start(ctx); err != nil {
		logger.Error("failed to start watcher", zap.Error(err))
		return 1
	}

	// dispatcher — reads events from watcher, updates store
	go func() {
		for event := range watcherInstance.Events() {
			switch event.Type {
			case netlinkwatcher.EventTypeAdded:
				storeInstance.Update(event)
			case netlinkwatcher.EventTypeDeleted:
				storeInstance.Delete(event.Index)
			}
		}
	}()

	// populate store with current interfaces before handling commands
	links, err := netlink.LinkList()
	if err != nil {
		logger.Error("failed to list initial interfaces", zap.Error(err))
		return 1
	}
	for _, link := range links {
		attrs := link.Attrs()
		storeInstance.Update(netlinkwatcher.Event{
			Type:      netlinkwatcher.EventTypeAdded,
			Index:     attrs.Index,
			LinkName:  attrs.Name,
			MTU:       attrs.MTU,
			Flags:     attrs.Flags,
			OperState: attrs.OperState,
		})
	}

	root := buildRootCmd(version, logger, watcherInstance, storeInstance)

	if err := root.Execute(); err != nil {
		// cobra already prints the error; we just set the exit code.
		return 1
	}
	return 0
}

func buildRootCmd(version string, logger *zap.Logger, watcherInstance *netlinkwatcher.Watcher, storeInstance *store.Store) *cobra.Command {
	root := &cobra.Command{
		// Use is the name people type on the terminal.
		Use:   "netvisor",
		Short: "Network state observer for Linux kernel, OVS/OVN, LXD, and K8s",
		Long: `netvisor reads live network state from the Linux kernel via netlink
and correlates it with OVS/OVN bridges, LXD networks, and Kubernetes CNI.

Think of it as a unified control-plane lens across your entire network stack.`,
		// SilenceUsage stops cobra printing the full usage block on every error.
		// Without this, a typo floods the terminal — very annoying.
		SilenceUsage: true,
	}

	// Attach sub-commands
	root.AddCommand(versionCmd(version))
	root.AddCommand(listCmd(storeInstance, logger))
	root.AddCommand(watchCmd(watcherInstance, logger))

	return root
}

// versionCmd returns the `netvisor version` sub-command.
// Notice it's a plain function that returns *cobra.Command — this is the
// idiomatic Go pattern. Each command owns its own flags and logic.
func versionCmd(version string) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the netvisor version",
		// RunE returns an error; Run does not. Prefer RunE so you can propagate
		// errors up instead of calling os.Exit deep in business logic.
		RunE: func(cmd *cobra.Command, args []string) error {
			// Use cmd.OutOrStdout() instead of fmt.Printf directly.
			// In production this resolves to os.Stdout — identical behaviour.
			// In tests, cmd.SetOut(&buf) redirects it to our buffer,
			// so the test can actually capture and assert on the output.
			fmt.Fprintf(cmd.OutOrStdout(), "netvisor %s\n", version)
			return nil
		},
	}
}
