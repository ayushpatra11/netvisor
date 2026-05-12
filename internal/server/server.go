package server

import (
	"context"

	"github.com/ayushpatra11/netvisor/internal/store"
	v1 "github.com/ayushpatra11/netvisor/proto/netvisor/v1"
)

type Server struct {
	v1.UnimplementedNetworkServiceServer

	serverStore *store.Store
}

func New(store *store.Store) *Server {
	return &Server{
		serverStore: store,
	}
}

func (s *Server) ListInterface(ctx context.Context, req *v1.ListInterfaceRequest) (*v1.ListInterfaceResponse, error) {
	var protoInterfaces []*v1.Interface

	for _, storeInterface := range s.serverStore.List() {
		protoInterfaces = append(protoInterfaces, &v1.Interface{
			Index:     int32(storeInterface.Index),
			LinkName:  storeInterface.LinkName,
			Mtu:       uint32(storeInterface.MTU),
			Flags:     uint32(storeInterface.Flags),
			OperState: uint32(storeInterface.OperState),
		})
	}
	return &v1.ListInterfaceResponse{Interfaces: protoInterfaces}, nil
}
