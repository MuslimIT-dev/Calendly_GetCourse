package connect

import (
	"context"
	"strconv"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	catalogv1 "github.com/MuslimIT-dev/Calendly_GetCourse/backend/gen/go/catalog/v1"
	catalogv1connect "github.com/MuslimIT-dev/Calendly_GetCourse/backend/gen/go/catalog/v1/catalogv1connect"
	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
	cataloguc "github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/usecase/catalog"
)

type CatalogHandler struct {
	listCategoriesUC *cataloguc.ListCategoriesUseCase
	getCategoryUC    *cataloguc.GetCategoryUseCase
	listServicesUC   *cataloguc.ListDefaultServicesUseCase
	getServiceUC     *cataloguc.GetDefaultServiceUseCase
}

func NewCatalogHandler(
	listCategoriesUC *cataloguc.ListCategoriesUseCase,
	getCategoryUC *cataloguc.GetCategoryUseCase,
	listServicesUC *cataloguc.ListDefaultServicesUseCase,
	getServiceUC *cataloguc.GetDefaultServiceUseCase,
) *CatalogHandler {
	return &CatalogHandler{
		listCategoriesUC: listCategoriesUC,
		getCategoryUC:    getCategoryUC,
		listServicesUC:   listServicesUC,
		getServiceUC:     getServiceUC,
	}
}

var _ catalogv1connect.CatalogServiceHandler = (*CatalogHandler)(nil)

func (h *CatalogHandler) ListCategories(
	ctx context.Context,
	req *connect.Request[catalogv1.ListCategoriesRequest],
) (*connect.Response[catalogv1.ListCategoriesResponse], error) {
	out, err := h.listCategoriesUC.Execute(ctx, cataloguc.ListCategoriesInput{})
	if err != nil {
		return nil, mapDomainError(err)
	}
	list := make([]*catalogv1.Category, len(out.Categories))
	for i, c := range out.Categories {
		list[i] = toProtoCategory(c)
	}
	return connect.NewResponse(&catalogv1.ListCategoriesResponse{Categories: list}), nil
}

func (h *CatalogHandler) ListDefaultServices(
	ctx context.Context,
	req *connect.Request[catalogv1.ListDefaultServicesRequest],
) (*connect.Response[catalogv1.ListDefaultServicesResponse], error) {
	var cid *int32
	if req.Msg.CategoryId != "" {
		v, err := parseID(req.Msg.CategoryId)
		if err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
		cid = &v
	}

	out, err := h.listServicesUC.Execute(ctx, cataloguc.ListDefaultServicesInput{CategoryID: cid})
	if err != nil {
		return nil, mapDomainError(err)
	}
	list := make([]*catalogv1.DefaultService, len(out.Services))
	for i, s := range out.Services {
		list[i] = toProtoDefaultService(s)
	}
	return connect.NewResponse(&catalogv1.ListDefaultServicesResponse{Services: list}), nil
}

func (h *CatalogHandler) GetDefaultService(
	ctx context.Context,
	req *connect.Request[catalogv1.GetDefaultServiceRequest],
) (*connect.Response[catalogv1.GetDefaultServiceResponse], error) {
	id, err := parseID(req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	out, err := h.getServiceUC.Execute(ctx, cataloguc.GetDefaultServiceInput{ID: id})
	if err != nil {
		return nil, mapDomainError(err)
	}
	return connect.NewResponse(&catalogv1.GetDefaultServiceResponse{
		Service: toProtoDefaultService(out.Service),
	}), nil
}

func toProtoCategory(c *domain.Category) *catalogv1.Category {
	return &catalogv1.Category{
		Id:        strconv.Itoa(int(c.ID)),
		Name:      c.Name,
		IsActive:  c.IsActive,
		CreatedAt: timestamppb.New(c.CreatedAt),
		UpdatedAt: timestamppb.New(c.UpdatedAt),
	}
}

func toProtoDefaultService(s *domain.DefaultService) *catalogv1.DefaultService {
	return &catalogv1.DefaultService{
		Id:           strconv.Itoa(int(s.ID)),
		Slug:         s.Slug,
		Name:         s.Name,
		Description:  s.Description,
		DurationMins: s.DurationMins,
		Price:        s.Price,
		CategoryId:   strconv.Itoa(int(s.CategoryID)),
		CreatedAt:    timestamppb.New(s.CreatedAt),
		UpdatedAt:    timestamppb.New(s.UpdatedAt),
	}
}

type AdminCatalogHandler struct {
	createCategoryUC *cataloguc.CreateCategoryUseCase
	updateCategoryUC *cataloguc.UpdateCategoryUseCase
	deleteCategoryUC *cataloguc.DeleteCategoryUseCase

	createServiceUC *cataloguc.CreateDefaultServiceUseCase
	updateServiceUC *cataloguc.UpdateDefaultServiceUseCase
	deleteServiceUC *cataloguc.DeleteDefaultServiceUseCase
}

func NewAdminCatalogHandler(
	createCategoryUC *cataloguc.CreateCategoryUseCase,
	updateCategoryUC *cataloguc.UpdateCategoryUseCase,
	deleteCategoryUC *cataloguc.DeleteCategoryUseCase,
	createServiceUC *cataloguc.CreateDefaultServiceUseCase,
	updateServiceUC *cataloguc.UpdateDefaultServiceUseCase,
	deleteServiceUC *cataloguc.DeleteDefaultServiceUseCase,
) *AdminCatalogHandler {
	return &AdminCatalogHandler{
		createCategoryUC: createCategoryUC,
		updateCategoryUC: updateCategoryUC,
		deleteCategoryUC: deleteCategoryUC,
		createServiceUC:  createServiceUC,
		updateServiceUC:  updateServiceUC,
		deleteServiceUC:  deleteServiceUC,
	}
}

var _ catalogv1connect.AdminCatalogServiceHandler = (*AdminCatalogHandler)(nil)

func (h *AdminCatalogHandler) AddCategory(
	ctx context.Context,
	req *connect.Request[catalogv1.AddCategoryRequest],
) (*connect.Response[catalogv1.AddCategoryResponse], error) {
	out, err := h.createCategoryUC.Execute(ctx, cataloguc.CreateCategoryInput{Name: req.Msg.Name})
	if err != nil {
		return nil, mapDomainError(err)
	}
	return connect.NewResponse(&catalogv1.AddCategoryResponse{
		Category: toProtoCategory(out.Category),
	}), nil
}

func (h *AdminCatalogHandler) UpdateCategory(
	ctx context.Context,
	req *connect.Request[catalogv1.UpdateCategoryRequest],
) (*connect.Response[catalogv1.UpdateCategoryResponse], error) {
	id, err := parseID(req.Msg.CategoryId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	out, err := h.updateCategoryUC.Execute(ctx, cataloguc.UpdateCategoryInput{ID: id, Name: req.Msg.Name})
	if err != nil {
		return nil, mapDomainError(err)
	}
	return connect.NewResponse(&catalogv1.UpdateCategoryResponse{
		Category: toProtoCategory(out.Category),
	}), nil
}

func (h *AdminCatalogHandler) DeleteCategory(
	ctx context.Context,
	req *connect.Request[catalogv1.DeleteCategoryRequest],
) (*connect.Response[catalogv1.DeleteCategoryResponse], error) {
	id, err := parseID(req.Msg.CategoryId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	_, err = h.deleteCategoryUC.Execute(ctx, cataloguc.DeleteCategoryInput{ID: id})
	if err != nil {
		return nil, mapDomainError(err)
	}
	return connect.NewResponse(&catalogv1.DeleteCategoryResponse{}), nil
}

func (h *AdminCatalogHandler) AddDefaultService(
	ctx context.Context,
	req *connect.Request[catalogv1.AddDefaultServiceRequest],
) (*connect.Response[catalogv1.AddDefaultServiceResponse], error) {
	categoryID, err := parseID(req.Msg.CategoryId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	out, err := h.createServiceUC.Execute(ctx, cataloguc.CreateDefaultServiceInput{
		CategoryID:   categoryID,
		Slug:         req.Msg.Slug,
		Name:         req.Msg.Name,
		Description:  req.Msg.Description,
		DurationMins: req.Msg.DurationMins,
		Price:        req.Msg.Price,
		Currency:     "RUB",
	})
	if err != nil {
		return nil, mapDomainError(err)
	}
	return connect.NewResponse(&catalogv1.AddDefaultServiceResponse{
		Service: toProtoDefaultService(out.Service),
	}), nil
}

func (h *AdminCatalogHandler) UpdateDefaultService(
	ctx context.Context,
	req *connect.Request[catalogv1.UpdateDefaultServiceRequest],
) (*connect.Response[catalogv1.UpdateDefaultServiceResponse], error) {
	id, err := parseID(req.Msg.ServiceId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	out, err := h.updateServiceUC.Execute(ctx, cataloguc.UpdateDefaultServiceInput{
		ID:           id,
		Slug:         req.Msg.Slug,
		Name:         req.Msg.Name,
		Description:  req.Msg.Description,
		DurationMins: req.Msg.DurationMins,
		Price:        req.Msg.Price,
	})
	if err != nil {
		return nil, mapDomainError(err)
	}
	return connect.NewResponse(&catalogv1.UpdateDefaultServiceResponse{
		Service: toProtoDefaultService(out.Service),
	}), nil
}

func (h *AdminCatalogHandler) DeleteDefaultService(
	ctx context.Context,
	req *connect.Request[catalogv1.DeleteDefaultServiceRequest],
) (*connect.Response[catalogv1.DeleteDefaultServiceResponse], error) {
	id, err := parseID(req.Msg.ServiceId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	_, err = h.deleteServiceUC.Execute(ctx, cataloguc.DeleteDefaultServiceInput{ID: id})
	if err != nil {
		return nil, mapDomainError(err)
	}
	return connect.NewResponse(&catalogv1.DeleteDefaultServiceResponse{}), nil
}
