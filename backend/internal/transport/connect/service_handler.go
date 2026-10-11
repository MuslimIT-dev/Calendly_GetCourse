package connect

import (
	"context"
	"strconv"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	servicev1 "github.com/MuslimIT-dev/Calendly_GetCourse/backend/gen/go/service/v1"
	servicev1connect "github.com/MuslimIT-dev/Calendly_GetCourse/backend/gen/go/service/v1/servicev1connect"
	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
	serviceuc "github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/usecase/service"
)

type ServiceHandler struct {
	getMyUC  *serviceuc.GetMyServicesUseCase
	addUC    *serviceuc.AddServiceUseCase
	updateUC *serviceuc.UpdateServiceUseCase
	deleteUC *serviceuc.DeleteServiceUseCase
}

func NewServiceHandler(
	getMyUC *serviceuc.GetMyServicesUseCase,
	addUC *serviceuc.AddServiceUseCase,
	updateUC *serviceuc.UpdateServiceUseCase,
	deleteUC *serviceuc.DeleteServiceUseCase,
) *ServiceHandler {
	return &ServiceHandler{
		getMyUC:  getMyUC,
		addUC:    addUC,
		updateUC: updateUC,
		deleteUC: deleteUC,
	}
}

var _ servicev1connect.ServiceServiceHandler = (*ServiceHandler)(nil)

func (h *ServiceHandler) GetMyServices(
	ctx context.Context,
	req *connect.Request[servicev1.GetMyServicesRequest],
) (*connect.Response[servicev1.GetMyServicesResponse], error) {
	out, err := h.getMyUC.Execute(ctx, serviceuc.GetMyServicesInput{})
	if err != nil {
		return nil, mapDomainError(err)
	}

	list := make([]*servicev1.Service, len(out.Services))
	for i, s := range out.Services {
		list[i] = toProtoService(s)
	}
	return connect.NewResponse(&servicev1.GetMyServicesResponse{Services: list}), nil
}

func (h *ServiceHandler) AddService(
	ctx context.Context,
	req *connect.Request[servicev1.AddServiceRequest],
) (*connect.Response[servicev1.AddServiceResponse], error) {
	defaultID, err := parseID(req.Msg.DefaultId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	out, err := h.addUC.Execute(ctx, serviceuc.AddServiceInput{
		DefaultID:      defaultID,
		NewDescription: req.Msg.NewDescription,
		DurationMins:   req.Msg.DurationMins,
		Price:          req.Msg.Price,
		Locations:      toDomainServiceLocations(req.Msg.Locations),
	})
	if err != nil {
		return nil, mapDomainError(err)
	}
	return connect.NewResponse(&servicev1.AddServiceResponse{
		Service: toProtoService(out.Service),
	}), nil
}

func (h *ServiceHandler) UpdateService(
	ctx context.Context,
	req *connect.Request[servicev1.UpdateServiceRequest],
) (*connect.Response[servicev1.UpdateServiceResponse], error) {
	id, err := parseID(req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	in := serviceuc.UpdateServiceInput{ID: id}

	if req.Msg.Description != "" {
		in.Description = &req.Msg.Description
	}
	if req.Msg.DurationMins > 0 {
		in.DurationMins = &req.Msg.DurationMins
	}
	if req.Msg.Price > 0 {
		in.Price = &req.Msg.Price
	}
	if len(req.Msg.Locations) > 0 {
		in.HasLocations = true
		in.Locations = toDomainServiceLocations(req.Msg.Locations)
	}

	out, err := h.updateUC.Execute(ctx, in)
	if err != nil {
		return nil, mapDomainError(err)
	}
	return connect.NewResponse(&servicev1.UpdateServiceResponse{
		Service: toProtoService(out.Service),
	}), nil
}

func (h *ServiceHandler) DeleteService(
	ctx context.Context,
	req *connect.Request[servicev1.DeleteServiceRequest],
) (*connect.Response[servicev1.DeleteServiceResponse], error) {
	id, err := parseID(req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	_, err = h.deleteUC.Execute(ctx, serviceuc.DeleteServiceInput{ID: id})
	if err != nil {
		return nil, mapDomainError(err)
	}
	return connect.NewResponse(&servicev1.DeleteServiceResponse{}), nil
}

func toProtoService(s *domain.Service) *servicev1.Service {
	locs := make([]*servicev1.ServiceLocation, len(s.Locations))
	for i, l := range s.Locations {
		locs[i] = &servicev1.ServiceLocation{
			LocationId:   strconv.Itoa(int(l.LocationID)),
			DurationMins: l.DurationMins,
			Price:        l.Price,
			IsActive:     l.IsActive,
		}
	}

	defaultID := ""
	if s.DefaultServiceID != nil {
		defaultID = strconv.Itoa(int(*s.DefaultServiceID))
	}

	return &servicev1.Service{
		Id:               strconv.Itoa(int(s.ID)),
		Slug:             s.Slug,
		Name:             s.Name,
		Description:      s.Description,
		DurationMins:     s.DurationMins,
		Price:            s.Price,
		Locations:        locs,
		IsActive:         s.IsActive,
		CreatedAt:        timestamppb.New(s.CreatedAt),
		UpdatedAt:        timestamppb.New(s.UpdatedAt),
		DefaultServiceId: defaultID,
		Currency:         s.Currency,
	}
}

func toDomainServiceLocations(in []*servicev1.ServiceLocation) []domain.ServiceLocation {
	out := make([]domain.ServiceLocation, 0, len(in))
	for _, l := range in {
		id, err := parseID(l.LocationId)
		if err != nil {
			continue
		}
		out = append(out, domain.ServiceLocation{
			LocationID:   id,
			DurationMins: l.DurationMins,
			Price:        l.Price,
			IsActive:     l.IsActive,
		})
	}
	return out
}