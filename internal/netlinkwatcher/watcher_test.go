package netlinkwatcher

import (
	"testing"

	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
)

// referred to :
// go doc github.com/vishvananda/netlink Dummy
// go doc github.com/vishvananda/netlink LinkAttrs

func TestToEvent(t *testing.T) {
	tests := []struct {
		name      string
		update    netlink.LinkUpdate
		wantType  EventType
		wantName  string
		wantIndex int
		wantMTU   int
	}{
		{
			name: "RTM_NEWLINK produces EventTypeAdded",
			update: netlink.LinkUpdate{
				Header: unix.NlMsghdr{
					Type: unix.RTM_NEWLINK,
				},
				Link: &netlink.Dummy{
					LinkAttrs: netlink.LinkAttrs{
						Name:  "eth0",
						Index: 2,
						MTU:   1500,
					},
				},
			},
			wantType:  EventTypeAdded,
			wantName:  "eth0",
			wantIndex: 2,
			wantMTU:   1500,
		},
		{
			name: "RTM_DELLINK produces EventTypeDeleted",
			update: netlink.LinkUpdate{
				Header: unix.NlMsghdr{
					Type: unix.RTM_DELLINK,
				},
				Link: &netlink.Dummy{
					LinkAttrs: netlink.LinkAttrs{
						Name:  "eth1",
						Index: 5,
						MTU:   1500,
					},
				},
			},
			wantType:  EventTypeDeleted,
			wantName:  "eth1",
			wantIndex: 5,
			wantMTU:   1500,
		},
		{
			name: "MTU is correctly translated",
			update: netlink.LinkUpdate{
				Header: unix.NlMsghdr{
					Type: unix.RTM_NEWLINK,
				},
				Link: &netlink.Dummy{
					LinkAttrs: netlink.LinkAttrs{
						Name:  "bond0",
						Index: 10,
						MTU:   9000,
					},
				},
			},
			wantType:  EventTypeAdded,
			wantName:  "bond0",
			wantIndex: 10,
			wantMTU:   9000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toEvent(tt.update)

			if got.Type != tt.wantType {
				t.Errorf("Type: got %q, want %q", got.Type, tt.wantType)
			}

			if got.LinkName != tt.wantName {
				t.Errorf("LinkName: got %q, want %q", got.LinkName, tt.wantName)
			}

			if got.Index != tt.wantIndex {
				t.Errorf("Index: got %d, want %d", got.Index, tt.wantIndex)
			}

			if got.MTU != tt.wantMTU {
				t.Errorf("MTU: got %d, want %d", got.MTU, tt.wantMTU)
			}
		})
	}
}
