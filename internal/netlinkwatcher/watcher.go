// internal/netlinkwatcher/watcher.go
package netlinkwatcher

import (
	"context"
	"fmt"
	"net"

	"github.com/vishvananda/netlink"
	"go.uber.org/zap"
	"golang.org/x/sys/unix"
)

// EventType tells consumers what kind of change happened. (Similar to #define in cpp)
type EventType string

const (
	EventTypeAdded   EventType = "ADDED"
	EventTypeDeleted EventType = "DELETED"
	EventTypeChanged EventType = "CHANGED" // Reserved for later. 
)

// Event is what we send down the channel to consumers.
// There is a LinkUpdate-> Link -> Attrs():
// acquired using : go doc github.com/vishvananda/netlink LinkAttrs
type Event struct {
	Index     int
	Type      EventType
	LinkName  string
	MTU       int
	Flags     net.Flags
	OperState netlink.LinkOperState
}

// Watcher subscribes to kernel netlink events and fans them
// out over a channel.
type Watcher struct {
	events chan Event
	logger *zap.Logger
}

// New creates a new Watcher.
// In Go, constructors are just functions named New* that return
// a pointer to the struct.
func New(logger *zap.Logger) *Watcher {
	// create an "object" of type Watcher
	return &Watcher{
		events: make(chan Event, 64),
		logger: logger,
	}
}

// Start subscribes to netlink and begins streaming events.
// It is non-blocking — it spawns a goroutine and returns.
// The goroutine stops when ctx is cancelled.
func (w *Watcher) Start(ctx context.Context) error {
	// The loop should handle two cases:
	//   case update := <-updates:   translate to your Event, send to w.events
	//   case <-ctx.Done():          log that we're shutting down, return
	eventUpdates := make(chan netlink.LinkUpdate)
	eventDone := make(chan struct{})

	if err := netlink.LinkSubscribe(eventUpdates, eventDone); err != nil {
		return fmt.Errorf("netlinkwatcher: failed to subscribe: %w", err)
	}

	go func() {
		for {
			select {
			case update := <-eventUpdates:
				// whenever an update is received
				updatedEvent := toEvent(update)
				w.logger.Info("netlink event received",
					zap.String("link", updatedEvent.LinkName),
					zap.Int("index", updatedEvent.Index),
					zap.String("state", updatedEvent.OperState.String()),
				)
				select {
				case w.events <- updatedEvent:
					w.logger.Info("event dispatched", zap.String("link", updatedEvent.LinkName))
				default:
					w.logger.Warn("event channel full, dropping event", zap.String("link", updatedEvent.LinkName))
				}
			case <-ctx.Done():
				// done
				close(eventDone)
				w.logger.Info("Closing the Watcher...")
				return
			}
		}
	}()

	return nil
}

/*
 Events returns the read-only channel consumers listen on,
 since they should not be able to fake ones. Hence, READ-ONLY
 using goroutines here to have the watcher check for events
 in a non-blocking manner
 for my information:
 chan Event       // can send AND receive — read/write
 chan<- Event     // can only send INTO — write only
 <-chan Event     // can only receive FROM — read only
*/

func (w *Watcher) Events() <-chan Event {
	return w.events
}

// toEvent translates a raw kernel LinkUpdate into your Event type.
// Keeping translation logic in its own function makes it unit-testable
// without needing a real kernel.
func toEvent(update netlink.LinkUpdate) Event {
	linkAttr := update.Link.Attrs()
	eventType := EventTypeAdded
	if update.Header.Type == unix.RTM_DELLINK {
		eventType = EventTypeDeleted
	}
	return Event{
		Index:     linkAttr.Index,
		Type:      eventType,
		LinkName:  linkAttr.Name,
		MTU:       linkAttr.MTU,
		Flags:     linkAttr.Flags,
		OperState: linkAttr.OperState,
	}
}
