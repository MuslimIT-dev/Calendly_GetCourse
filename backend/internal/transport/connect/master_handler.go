package connect

import (
	"context"
	"strconv"

	"connectrpc.com/connect"

	masterv1 "github.com/MuslimIT-dev/Calendly_GetCourse/backend/gen/go/master/v1"
	masterv1connect "github.com/MuslimIT-dev/Calendly_GetCourse/backend/gen/go/master/v1/masterv1connect"
	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
	masteruc "github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/usecase/master"
)

type MasterHandler struct {
	listUC       *masteruc.ListMastersUseCase
	getUC        *masteruc.GetMasterUseCase
	getMyUC      *masteruc.GetMyProfileUseCase
	updateUC     *masteruc.UpdateMasterUseCase
	updateSlugUC *masteruc.UpdateSlugUseCase
}

func NewMasterHandler(
	listUC *masteruc.ListMastersUseCase,
	getUC *masteruc.GetMasterUseCase,
	getMyUC *masteruc.GetMyProfileUseCase,
	updateUC *masteruc.UpdateMasterUseCase,
	updateSlugUC *masteruc.UpdateSlugUseCase,
) *MasterHandler {
	return &MasterHandler{
		listUC:       listUC,
		getUC:        getUC,
		getMyUC:      getMyUC,
		updateUC:     updateUC,
		updateSlugUC: updateSlugUC,
	}
}

var _ masterv1connect.MasterServiceHandler = (*MasterHandler)(nil)

func (h *MasterHandler) ListMasters(
	ctx context.Context,
	req *connect.Request[masterv1.ListMastersRequest],
) (*connect.Response[masterv1.ListMastersResponse], error) {
	in := masteruc.ListMastersInput{
		PageSize:  req.Msg.PageSize,
		PageToken: req.Msg.PageToken,
		Filter:    toDomainFilter(req.Msg.Filters),
	}

	out, err := h.listUC.Execute(ctx, in)
	if err != nil {
		return nil, mapDomainError(err)
	}

	cards := make([]*masterv1.MasterCard, len(out.Cards))
	for i, c := range out.Cards {
		cards[i] = toProtoMasterCard(c)
	}

	return connect.NewResponse(&masterv1.ListMastersResponse{
		Masters:       cards,
		NextPageToken: out.NextPageToken,
	}), nil
}

func (h *MasterHandler) GetMaster(
	ctx context.Context,
	req *connect.Request[masterv1.GetMasterRequest],
) (*connect.Response[masterv1.GetMasterResponse], error) {
	out, err := h.getUC.Execute(ctx, masteruc.GetMasterInput{Slug: req.Msg.Slug})
	if err != nil {
		return nil, mapDomainError(err)
	}
	return connect.NewResponse(&masterv1.GetMasterResponse{
		User:   toProtoUser(out.User),
		Master: toProtoProfile(out.Profile),
	}), nil
}

func (h *MasterHandler) GetMyMasterProfile(
	ctx context.Context,
	req *connect.Request[masterv1.GetMyMasterProfileRequest],
) (*connect.Response[masterv1.GetMasterResponse], error) {
	out, err := h.getMyUC.Execute(ctx, masteruc.GetMyProfileInput{})
	if err != nil {
		return nil, mapDomainError(err)
	}
	return connect.NewResponse(&masterv1.GetMasterResponse{
		Master: toProtoProfile(out.Profile),
	}), nil
}

func (h *MasterHandler) UpdateMaster(
	ctx context.Context,
	req *connect.Request[masterv1.UpdateMasterRequest],
) (*connect.Response[masterv1.UpdateMasterResponse], error) {
	in := masteruc.UpdateMasterInput{}

	if req.Msg.Bio != "" {
		in.Bio = &req.Msg.Bio
	}
	if req.Msg.Specialization != "" {
		in.Specialization = &req.Msg.Specialization
	}
	if req.Msg.YearsOfExperience != 0 {
		in.YearsOfExperience = &req.Msg.YearsOfExperience
	}
	if len(req.Msg.Languages) > 0 {
		in.HasLanguages = true
		in.Languages = toDomainLanguages(req.Msg.Languages)
	}
	if len(req.Msg.Certificates) > 0 {
		in.HasCertificates = true
		in.Certificates = toDomainCertificates(req.Msg.Certificates)
	}

	out, err := h.updateUC.Execute(ctx, in)
	if err != nil {
		return nil, mapDomainError(err)
	}
	return connect.NewResponse(&masterv1.UpdateMasterResponse{
		Master: toProtoProfile(out.Profile),
	}), nil
}

func (h *MasterHandler) UpdateSlug(
	ctx context.Context,
	req *connect.Request[masterv1.UpdateSlugRequest],
) (*connect.Response[masterv1.UpdateSlugResponse], error) {
	out, err := h.updateSlugUC.Execute(ctx, masteruc.UpdateSlugInput{Slug: req.Msg.Slug})
	if err != nil {
		return nil, mapDomainError(err)
	}
	return connect.NewResponse(&masterv1.UpdateSlugResponse{
		Master: toProtoProfile(out.Profile),
	}), nil
}

func toProtoProfile(p *domain.MasterProfile) *masterv1.MasterProfile {
	if p == nil {
		return nil
	}

	langs := make([]*masterv1.Language, len(p.Languages))
	for i, l := range p.Languages {
		langs[i] = &masterv1.Language{
			Name:        l.Name,
			Proficiency: masterv1.ProficiencyType(l.Proficiency),
		}
	}

	certs := make([]*masterv1.Certificate, len(p.Certificates))
	for i, c := range p.Certificates {
		certs[i] = &masterv1.Certificate{
			Name:         c.Name,
			Organization: c.Organization,
			Year:         c.Year,
			FileUrl:      c.FileURL,
		}
	}

	defaultLocID := ""
	if p.DefaultLocationID != nil {
		defaultLocID = strconv.Itoa(int(*p.DefaultLocationID))
	}

	return &masterv1.MasterProfile{
		UserId:              strconv.Itoa(int(p.UserID)),
		Slug:                p.Slug,
		Bio:                 p.Bio,
		Specialization:      p.Specialization,
		YearsOfExperience:   p.YearsOfExperience,
		Languages:           langs,
		Certificates:        certs,
		IsAcceptingBookings: p.IsAcceptingBookings,
		DefaultLocationId:   defaultLocID,
		AvgRating:           p.AvgRating,
		ReviewsCount:        p.ReviewsCount,
	}
}

func toProtoMasterCard(c *domain.MasterCardData) *masterv1.MasterCard {
	return &masterv1.MasterCard{
		User:          toProtoUser(c.User),
		Master:        toProtoProfile(c.Profile),
		TotalServices: c.TotalServices,
		MinPrice:      c.MinPrice,
	}
}

func toDomainFilter(f *masterv1.Filter) domain.MasterFilter {
	if f == nil {
		return domain.MasterFilter{}
	}
	out := domain.MasterFilter{
		SortBy:         domain.SortType(f.SortBy),
		SortOrder:      domain.SortOrder(f.SortOrder),
		Specialization: f.Specialization,
		MasterTimezone: f.MasterTimezone,
	}
	if f.MinPrice > 0 {
		v := f.MinPrice
		out.MinPrice = &v
	}
	if f.MaxPrice > 0 {
		v := f.MaxPrice
		out.MaxPrice = &v
	}
	if f.MinExperience > 0 {
		v := f.MinExperience
		out.MinExperience = &v
	}
	if f.ServiceId != "" {
		if id, err := strconv.ParseInt(f.ServiceId, 10, 32); err == nil {
			v := int32(id)
			out.ServiceID = &v
		}
	}
	return out
}

func toDomainLanguages(in []string) []domain.Language {
	out := make([]domain.Language, len(in))
	for i, name := range in {
		out[i] = domain.Language{
			Name: name,
		}
	}
	return out
}

func toDomainCertificates(in []*masterv1.Certificate) []domain.Certificate {
	out := make([]domain.Certificate, len(in))
	for i, c := range in {
		out[i] = domain.Certificate{
			Name:         c.Name,
			Organization: c.Organization,
			Year:         c.Year,
			FileURL:      c.FileUrl,
		}
	}
	return out
}
