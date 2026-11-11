// Package allnet contains the gRPC adapter that exposes the All.Net store-location
// domain client (port.AllNetClient) over the AllNetService gRPC API. It is a thin
// translation layer: it maps protobuf requests to client calls and domain
// entities back to protobuf responses, and performs no business logic of its own.
package allnet

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	allnetv1 "github.com/Duke40199/sega-idac-client/gen/allnetv1"
	"github.com/Duke40199/sega-idac-client/internal/domain/entity"
	"github.com/Duke40199/sega-idac-client/internal/domain/port"
)

// Handler implements allnetv1.AllNetServiceServer by delegating to the domain
// AllNetClient.
type Handler struct {
	allnetv1.UnimplementedAllNetServiceServer

	client port.AllNetClient
}

// NewHandler creates a new AllNetService gRPC handler backed by client.
func NewHandler(client port.AllNetClient) *Handler {
	return &Handler{client: client}
}

func (h *Handler) GetListStore(ctx context.Context, req *allnetv1.GetListStoreRequest) (*allnetv1.GetListStoreResponse, error) {
	stores, areaName, err := h.client.GetListStore(ctx, req.GetGameCode(), req.GetLanguageCode(), req.GetAreaCode())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get list store: %v", err)
	}
	return &allnetv1.GetListStoreResponse{
		Stores:   toProtoStoreLocations(stores),
		AreaName: areaName,
	}, nil
}

// ----- entity -> proto mappers -----

func toProtoStoreLocations(in []entity.StoreLocation) []*allnetv1.StoreLocation {
	out := make([]*allnetv1.StoreLocation, 0, len(in))
	for _, s := range in {
		store := &allnetv1.StoreLocation{
			Id:             s.ID,
			Name:           s.Name,
			Address:        s.Address,
			SegaAreaCode:   s.SegaAreaCode,
			AllNetAreaCode: s.AllNetAreaCode,
		}
		if !s.CreatedAt.IsZero() {
			store.CreatedAt = timestamppb.New(s.CreatedAt)
		}
		if !s.UpdatedAt.IsZero() {
			store.UpdatedAt = timestamppb.New(s.UpdatedAt)
		}
		out = append(out, store)
	}
	return out
}
