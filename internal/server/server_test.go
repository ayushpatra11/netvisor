package server

import (
	"context"
	"testing"

	"net"

	"github.com/ayushpatra11/netvisor/internal/store"
	v1 "github.com/ayushpatra11/netvisor/proto/netvisor/v1"
	"github.com/vishvananda/netlink"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ayushpatra11/netvisor/internal/netlinkwatcher"
)

func makeStore(interfaces []netlinkwatcher.Event) *store.Store {
	s := store.New()
	for _, iface := range interfaces {
		s.Update(iface)
	}
	return s
}

func TestListInterfacesEmpty(t *testing.T) {
	s := makeStore(nil)
	srv := New(s)

	resp, err := srv.ListInterface(context.Background(), &v1.ListInterfaceRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Interfaces) != 0 {
		t.Errorf("expected 0 interfaces, got %d", len(resp.Interfaces))
	}
}

func TestListInterfacesReturnsAll(t *testing.T) {
	s := makeStore([]netlinkwatcher.Event{
		{Type: netlinkwatcher.EventTypeAdded, Index: 1, LinkName: "eth0", MTU: 1500, OperState: netlink.OperUp},
		{Type: netlinkwatcher.EventTypeAdded, Index: 2, LinkName: "lo", MTU: 65536, OperState: netlink.OperUnknown},
	})
	srv := New(s)

	resp, err := srv.ListInterface(context.Background(), &v1.ListInterfaceRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Interfaces) != 2 {
		t.Errorf("expected 2 interfaces, got %d", len(resp.Interfaces))
	}
}

func TestListInterfaceFieldMapping(t *testing.T) {
	s := makeStore([]netlinkwatcher.Event{
		{
			Type:      netlinkwatcher.EventTypeAdded,
			Index:     3,
			LinkName:  "wlan0",
			MTU:       1500,
			Flags:     net.FlagUp | net.FlagBroadcast,
			OperState: netlink.OperUp,
		},
	})
	srv := New(s)

	resp, err := srv.ListInterface(context.Background(), &v1.ListInterfaceRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	iface := resp.Interfaces[0]
	if iface.Index != 3 {
		t.Errorf("Index: got %d, want 3", iface.Index)
	}
	if iface.LinkName != "wlan0" {
		t.Errorf("LinkName: got %q, want wlan0", iface.LinkName)
	}
	if iface.Mtu != 1500 {
		t.Errorf("MTU: got %d, want 1500", iface.Mtu)
	}
}

func TestNewServer(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	_ = logger
	s := store.New()
	srv := New(s)
	if srv == nil {
		t.Error("expected non-nil server")
	}
	if srv.serverStore != s {
		t.Error("store not correctly assigned")
	}
}

func TestListInterfacesInvalidContext(t *testing.T) {
	s := makeStore(nil)
	srv := New(s)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	// should still return a valid response since we don't use context yet
	_, err := srv.ListInterface(ctx, &v1.ListInterfaceRequest{})
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.Canceled {
			return // acceptable
		}
		t.Fatalf("unexpected error: %v", err)
	}
}
