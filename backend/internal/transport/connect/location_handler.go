package connect

import (
	"context"
	"strconv"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	locationv1 "github.com/MuslimIT-dev/Calendly_GetCourse/backend/gen/go/location/v1"
	locationv1connect "github.com/MuslimIT-dev/Calendly_GetCourse/backend/gen/go/location/v1/locationv1connect"
	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
	locationuc "github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/usecase/location"
)

type LocationHandler struct {
	listUC   *locationuc.ListLocationsUseCase
	getUC    *locationuc.GetLocationUseCase
	createUC *locationuc.CreateLocationUseCase
	updateUC *locationuc.UpdateLocationUseCase
	deleteUC *locationuc.DeleteLocationUseCase
}

func NewLocationHandler(
	listUC *locationuc.ListLocationsUseCase,
	getUC *locationuc.GetLocationUseCase,
	createUC *locationuc.CreateLocationUseCase,
	updateUC *locationuc.UpdateLocationUseCase,
	deleteUC *locationuc.DeleteLocationUseCase,
) *LocationHandler {
	return &LocationHandler{
		listUC:   listUC,
		getUC:    getUC,
		createUC: createUC,
		updateUC: updateUC,
		deleteUC: deleteUC,
	}
}

var _ locationv1connect.LocationServiceHandler = (*LocationHandler)(nil)

func (h *LocationHandler) ListLocations(
	ctx context.Context,
	req *connect.Request[locationv1.ListLocationsRequest],
) (*connect.Response[locationv1.ListLocationsResponse], error) {
	masterID, err := parseID(req.Msg.MasterId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	out, err := h.listUC.Execute(ctx, locationuc.ListLocationsInput{
		MasterID:   masterID,
		OnlyActive: req.Msg.OnlyActive,
	})
	if err != nil {
		return nil, mapDomainError(err)
	}

	list := make([]*locationv1.Location, len(out.Locations))
	for i, l := range out.Locations {
		list[i] = toProtoLocation(l)
	}

	return connect.NewResponse(&locationv1.ListLocationsResponse{
		Locations: list,
	}), nil
}

func (h *LocationHandler) GetLocation(
	ctx context.Context,
	req *connect.Request[locationv1.GetLocationRequest],
) (*connect.Response[locationv1.GetLocationResponse], error) {
	id, err := parseID(req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	out, err := h.getUC.Execute(ctx, locationuc.GetLocationInput{ID: id})
	if err != nil {
		return nil, mapDomainError(err)
	}
	return connect.NewResponse(&locationv1.GetLocationResponse{
		Location: toProtoLocation(out.Location),
	}), nil
}

func (h *LocationHandler) CreateLocation(
	ctx context.Context,
	req *connect.Request[locationv1.CreateLocationRequest],
) (*connect.Response[locationv1.CreateLocationResponse], error) {
	out, err := h.createUC.Execute(ctx, locationuc.CreateLocationInput{
		Name:       req.Msg.Name,
		Address:    req.Msg.Address,
		Timezone:   req.Msg.Timezone,
		IsOnline:   req.Msg.IsOnline,
		MeetingURL: req.Msg.MeetingUrl,
	})
	if err != nil {
		return nil, mapDomainError(err)
	}
	return connect.NewResponse(&locationv1.CreateLocationResponse{
		Location: toProtoLocation(out.Location),
	}), nil
}

func (h *LocationHandler) UpdateLocation(
	ctx context.Context,
	req *connect.Request[locationv1.UpdateLocationRequest],
) (*connect.Response[locationv1.UpdateLocationResponse], error) {
	id, err := parseID(req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	in := locationuc.UpdateLocationInput{ID: id}

	if req.Msg.Name != "" {
		in.Name = &req.Msg.Name
	}
	if req.Msg.Address != "" {
		in.Address = &req.Msg.Address
	}
	if req.Msg.Timezone != "" {
		in.Timezone = &req.Msg.Timezone
	}
	in.IsOnline = &req.Msg.IsOnline
	if req.Msg.MeetingUrl != "" {
		in.MeetingURL = &req.Msg.MeetingUrl
	}
	in.IsActive = &req.Msg.IsActive

	out, err := h.updateUC.Execute(ctx, in)
	if err != nil {
		return nil, mapDomainError(err)
	}
	return connect.NewResponse(&locationv1.UpdateLocationResponse{
		Location: toProtoLocation(out.Location),
	}), nil
}

func (h *LocationHandler) DeleteLocation(
	ctx context.Context,
	req *connect.Request[locationv1.DeleteLocationRequest],
) (*connect.Response[locationv1.DeleteLocationResponse], error) {
	id, err := parseID(req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	_, err = h.deleteUC.Execute(ctx, locationuc.DeleteLocationInput{ID: id})
	if err != nil {
		return nil, mapDomainError(err)
	}
	return connect.NewResponse(&locationv1.DeleteLocationResponse{}), nil
}

func toProtoLocation(l *domain.Location) *locationv1.Location {
	return &locationv1.Location{
		Id:         strconv.Itoa(int(l.ID)),
		MasterId:   strconv.Itoa(int(l.MasterID)),
		Name:       l.Name,
		Address:    l.Address,
		Timezone:   l.Timezone,
		IsOnline:   l.IsOnline,
		MeetingUrl: l.MeetingURL,
		IsActive:   l.IsActive,
		CreatedAt:  timestamppb.New(l.CreatedAt),
		UpdatedAt:  timestamppb.New(l.UpdatedAt),
	}
}
