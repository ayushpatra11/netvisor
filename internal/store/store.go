package store

import (
	"net"
	"sync"

	"github.com/ayushpatra11/netvisor/internal/netlinkwatcher"
	"github.com/vishvananda/netlink"
)

// Interface : representing the last known state of a network interface
// It is just a snapshot at a given moment, not a stream.
type Interface struct {
	// considering the events emitted from netlinkwatcher here.
	Index     int
	LinkName  string
	MTU       int
	Flags     net.Flags
	OperState netlink.LinkOperState
}

// Store : current state for all interfaces
// Safe for concurrent use.
type Store struct {
	// map :
	// 	key = Index
	//	value = Interface
	// have a Read Write lock as well.
	storeMap  map[int]Interface
	storeLock sync.RWMutex
}

// Acts as a constructor for the Store.
func New() *Store {
	// Initialising the store.
	return &Store{
		storeMap: make(map[int]Interface),
	}
}

// Allows the store to be updated whenever
// a new event is received.
func (s *Store) Update(event netlinkwatcher.Event) {
	eventIndex := event.Index
	s.storeLock.Lock()
	defer s.storeLock.Unlock()

	s.storeMap[eventIndex] = Interface{
		Index:     eventIndex,
		LinkName:  event.LinkName,
		MTU:       event.MTU,
		Flags:     event.Flags,
		OperState: event.OperState,
	}

}

// allows the store to delete an interface by index.
// In case the event has been deleted, we remove it
// from the store.
func (s *Store) Delete(index int) {
	s.storeLock.Lock()
	defer s.storeLock.Unlock()

	delete(s.storeMap, index)

}

// allows a snapshot of the all the interfaces
// to be listed and printed.
func (s *Store) List() []Interface {
	s.storeLock.RLock()
	defer s.storeLock.RUnlock()

	var interfaces []Interface

	for _, eventInterface := range s.storeMap {
		interfaces = append(interfaces, eventInterface)
	}

	return interfaces
}

// allows a particular interface to be fetched using
// the index.
func (s *Store) Get(index int) (Interface, bool) {
	s.storeLock.RLock()
	defer s.storeLock.RUnlock()

	iface, ok := s.storeMap[index]
	return iface, ok
}
