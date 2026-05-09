// internal/netlinkwatcher/watcher.go
package netlinkwatcher

import (
    "context"
    "fmt"

    "github.com/vishvananda/netlink"
    "go.uber.org/zap"
)

// EventType tells consumers what kind of change happened. (Similar to #define in cpp)
type EventType string

const (
    EventTypeAdded   EventType = "ADDED"
    EventTypeDeleted EventType = "DELETED"
    EventTypeChanged EventType = "CHANGED"
)

// Event is what we send down the channel to consumers.
type Event struct {
    Type      EventType
    LinkName  string

}

// Watcher subscribes to kernel netlink events and fans them
// out over a channel.
type Watcher struct {
    // TODO: need a channel to send events out on.
    //       Should it be buffered or unbuffered? Why?

    // TODO: need a logger

    // TODO: anything else?
}

// New creates a new Watcher.
// TODO: implement this.
// In Go, constructors are just functions named New* that return
// a pointer to the struct.
func New(logger *zap.Logger) *Watcher {

}

// Events returns the read-only channel consumers listen on.
// TODO: implement this — one line.
// Why does this return <-chan Event and not chan Event?
func (w *Watcher) Events() <-chan Event {

}

// Start subscribes to netlink and begins streaming events.
// It is non-blocking — it spawns a goroutine and returns.
// The goroutine stops when ctx is cancelled.
func (w *Watcher) Start(ctx context.Context) error {
    // The loop should handle two cases:
    //   case update := <-updates:   translate to your Event, send to w.events
    //   case <-ctx.Done():          log that we're shutting down, return

    // TODO: implement this
    return nil
}

// toEvent translates a raw kernel LinkUpdate into your Event type.
// Keeping translation logic in its own function makes it unit-testable
// without needing a real kernel.
// TODO: implement this
func toEvent(update netlink.LinkUpdate) Event {

}
