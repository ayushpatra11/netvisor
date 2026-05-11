package store

import (
	"net"
	"sync"
	"testing"

	"github.com/ayushpatra11/netvisor/internal/netlinkwatcher"
	"github.com/vishvananda/netlink"
)

// helper to build a test event without repeating yourself
func makeEvent(index int, name string, mtu int) netlinkwatcher.Event {
	return netlinkwatcher.Event{
		Index:     index,
		LinkName:  name,
		MTU:       mtu,
		Flags:     net.FlagUp,
		OperState: netlink.OperUp,
	}
}

// Basic behaviour

func TestListEmptyStore(t *testing.T) {
	s := New()
	got := s.List()
	if len(got) != 0 {
		t.Errorf("expected empty list, got %d entries", len(got))
	}
}

func TestGetMissingIndex(t *testing.T) {
	s := New()
	_, ok := s.Get(99)
	if ok {
		t.Error("expected ok=false for missing index, got true")
	}
}

func TestUpsertAndGet(t *testing.T) {
	s := New()
	s.Update(makeEvent(2, "eth0", 1500))

	iface, ok := s.Get(2)
	if !ok {
		t.Fatal("expected ok=true after Update, got false")
	}
	if iface.LinkName != "eth0" {
		t.Errorf("LinkName: got %q, want %q", iface.LinkName, "eth0")
	}
	if iface.MTU != 1500 {
		t.Errorf("MTU: got %d, want %d", iface.MTU, 1500)
	}
}

func TestUpdateSameIndexTwice(t *testing.T) {
	s := New()
	s.Update(makeEvent(2, "eth0", 1500))
	s.Update(makeEvent(2, "eth0", 9000)) // same index, new MTU

	list := s.List()
	if len(list) != 1 {
		t.Errorf("expected 1 entry after updating same index twice, got %d", len(list))
	}
	if list[0].MTU != 9000 {
		t.Errorf("MTU: got %d, want 9000 after update", list[0].MTU)
	}
}

func TestDeleteRemovesInterface(t *testing.T) {
	s := New()
	s.Update(makeEvent(2, "eth0", 1500))
	s.Delete(2)

	_, ok := s.Get(2)
	if ok {
		t.Error("expected ok=false after Delete, got true")
	}
	if len(s.List()) != 0 {
		t.Errorf("expected empty list after Delete, got %d entries", len(s.List()))
	}
}

func TestListReturnsCopy(t *testing.T) {
	s := New()
	s.Update(makeEvent(2, "eth0", 1500))

	list := s.List()
	// mutate the returned slice
	list[0].LinkName = "hacked"

	// the store should be unaffected
	iface, _ := s.Get(2)
	if iface.LinkName == "hacked" {
		t.Error("List() returned a reference to internal state, not a copy")
	}
}

// Concurrency
// No assertions — the race detector (go test -race) catches any data races.

func TestStoreConcurrentAccess(t *testing.T) {
	s := New()
	var wg sync.WaitGroup

	// 10 writer goroutines
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				s.Update(makeEvent(id, "eth0", 1500))
			}
		}(i)
	}

	// 10 reader goroutines
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				s.List()
				s.Get(id)
			}
		}(i)
	}

	wg.Wait() // blocks until all 20 goroutines finish
}
