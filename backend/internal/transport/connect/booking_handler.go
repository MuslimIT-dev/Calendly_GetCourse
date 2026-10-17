package connect

import (
	"context"
	"strconv"

	"connectrpc.com/connect"
	"google.golang.org/genproto/googleapis/type/money"
	"google.golang.org/protobuf/types/known/timestamppb"

	bookingv1 "github.com/MuslimIT-dev/zaman/backend/gen/go/booking/v1"
	bookingv1connect "github.com/MuslimIT-dev/zaman/backend/gen/go/booking/v1/bookingv1connect"
	"github.com/MuslimIT-dev/zaman/backend/internal/domain"
	bookinguc "github.com/MuslimIT-dev/zaman/backend/internal/usecase/booking"
)

type BookingHandler struct {
	createUC     *bookinguc.CreateUseCase
	getUC        *bookinguc.GetUseCase
	listMyUC     *bookinguc.ListMyUseCase
	listMasterUC *bookinguc.ListMasterUseCase
	cancelUC     *bookinguc.CancelUseCase
	rescheduleUC *bookinguc.RescheduleUseCase
	confirmUC    *bookinguc.ConfirmUseCase
	completeUC   *bookinguc.CompleteUseCase
	noShowUC     *bookinguc.NoShowUseCase
}

func NewBookingHandler(
	createUC *bookinguc.CreateUseCase,
	getUC *bookinguc.GetUseCase,
	listMyUC *bookinguc.ListMyUseCase,
	listMasterUC *bookinguc.ListMasterUseCase,
	cancelUC *bookinguc.CancelUseCase,
	rescheduleUC *bookinguc.RescheduleUseCase,
	confirmUC *bookinguc.ConfirmUseCase,
	completeUC *bookinguc.CompleteUseCase,
	noShowUC *bookinguc.NoShowUseCase,
) *BookingHandler {
	return &BookingHandler{
		createUC:     createUC,
		getUC:        getUC,
		listMyUC:     listMyUC,
		listMasterUC: listMasterUC,
		cancelUC:     cancelUC,
		rescheduleUC: rescheduleUC,
		confirmUC:    confirmUC,
		completeUC:   completeUC,
		noShowUC:     noShowUC,
	}
}

var _ bookingv1connect.BookingServiceHandler = (*BookingHandler)(nil)

func (h *BookingHandler) CreateAppointment(
	ctx context.Context,
	req *connect.Request[bookingv1.CreateAppointmentRequest],
) (*connect.Response[bookingv1.CreateAppointmentResponse], error) {
	masterID, err := parseID(req.Msg.MasterId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	serviceID, err := parseID(req.Msg.ServiceId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	locationID, err := parseID(req.Msg.LocationId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	out, err := h.createUC.Execute(ctx, bookinguc.CreateInput{
		MasterID:    masterID,
		ServiceID:   serviceID,
		LocationID:  locationID,
		StartTime:   req.Msg.StartTime.AsTime(),
		Timezone:    req.Msg.Timezone,
		ClientNotes: req.Msg.ClientNotes,
	})
	if err != nil {
		return nil, mapDomainError(err)
	}
	return connect.NewResponse(&bookingv1.CreateAppointmentResponse{
		Appointment: toProtoAppointment(out.Appointment),
	}), nil
}

func (h *BookingHandler) GetAppointment(
	ctx context.Context,
	req *connect.Request[bookingv1.GetAppointmentRequest],
) (*connect.Response[bookingv1.GetAppointmentResponse], error) {
	id, err := parseAppointmentID(req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	out, err := h.getUC.Execute(ctx, bookinguc.GetInput{ID: id})
	if err != nil {
		return nil, mapDomainError(err)
	}
	return connect.NewResponse(&bookingv1.GetAppointmentResponse{
		Appointment: toProtoAppointment(out.Appointment),
	}), nil
}

func (h *BookingHandler) ListMyAppointments(
	ctx context.Context,
	req *connect.Request[bookingv1.ListMyAppointmentsRequest],
) (*connect.Response[bookingv1.ListMyAppointmentsResponse], error) {
	var sf *domain.AppointmentStatus
	if req.Msg.StatusFilter != bookingv1.AppointmentStatus_APPOINTMENT_STATUS_UNSPECIFIED {
		v := domain.AppointmentStatus(req.Msg.StatusFilter)
		sf = &v
	}

	out, err := h.listMyUC.Execute(ctx, bookinguc.ListMyInput{
		StatusFilter: sf,
		PageSize:     req.Msg.PageSize,
		PageToken:    req.Msg.PageToken,
	})
	if err != nil {
		return nil, mapDomainError(err)
	}

	list := make([]*bookingv1.Appointment, len(out.Appointments))
	for i, a := range out.Appointments {
		list[i] = toProtoAppointment(a)
	}
	return connect.NewResponse(&bookingv1.ListMyAppointmentsResponse{
		Appointments:  list,
		NextPageToken: out.NextPageToken,
	}), nil
}

func (h *BookingHandler) ListMasterAppointments(
	ctx context.Context,
	req *connect.Request[bookingv1.ListMasterAppointmentsRequest],
) (*connect.Response[bookingv1.ListMasterAppointmentsResponse], error) {
	masterID, err := parseID(req.Msg.MasterId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	var locID *int32
	if req.Msg.LocationId != "" {
		id, err := parseID(req.Msg.LocationId)
		if err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
		locID = &id
	}

	var sf *domain.AppointmentStatus
	if req.Msg.StatusFilter != bookingv1.AppointmentStatus_APPOINTMENT_STATUS_UNSPECIFIED {
		v := domain.AppointmentStatus(req.Msg.StatusFilter)
		sf = &v
	}

	out, err := h.listMasterUC.Execute(ctx, bookinguc.ListMasterInput{
		MasterID:     masterID,
		LocationID:   locID,
		StatusFilter: sf,
		FromDate:     req.Msg.FromDate,
		ToDate:       req.Msg.ToDate,
		PageSize:     req.Msg.PageSize,
		PageToken:    req.Msg.PageToken,
	})
	if err != nil {
		return nil, mapDomainError(err)
	}

	list := make([]*bookingv1.Appointment, len(out.Appointments))
	for i, a := range out.Appointments {
		list[i] = toProtoAppointment(a)
	}
	return connect.NewResponse(&bookingv1.ListMasterAppointmentsResponse{
		Appointments:  list,
		NextPageToken: out.NextPageToken,
	}), nil
}

func (h *BookingHandler) CancelAppointment(
	ctx context.Context,
	req *connect.Request[bookingv1.CancelAppointmentRequest],
) (*connect.Response[bookingv1.CancelAppointmentResponse], error) {
	id, err := parseAppointmentID(req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	out, err := h.cancelUC.Execute(ctx, bookinguc.CancelInput{
		ID:     id,
		Reason: req.Msg.Reason,
	})
	if err != nil {
		return nil, mapDomainError(err)
	}
	return connect.NewResponse(&bookingv1.CancelAppointmentResponse{
		Appointment: toProtoAppointment(out.Appointment),
	}), nil
}

func (h *BookingHandler) RescheduleAppointment(
	ctx context.Context,
	req *connect.Request[bookingv1.RescheduleAppointmentRequest],
) (*connect.Response[bookingv1.RescheduleAppointmentResponse], error) {
	id, err := parseAppointmentID(req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	var locID *int32
	if req.Msg.LocationId != "" {
		v, err := parseID(req.Msg.LocationId)
		if err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
		locID = &v
	}

	out, err := h.rescheduleUC.Execute(ctx, bookinguc.RescheduleInput{
		ID:           id,
		LocationID:   locID,
		NewStartTime: req.Msg.NewStartTime.AsTime(),
		Timezone:     req.Msg.Timezone,
	})
	if err != nil {
		return nil, mapDomainError(err)
	}
	return connect.NewResponse(&bookingv1.RescheduleAppointmentResponse{
		Appointment: toProtoAppointment(out.Appointment),
	}), nil
}

func (h *BookingHandler) ConfirmAppointment(
	ctx context.Context,
	req *connect.Request[bookingv1.ConfirmAppointmentRequest],
) (*connect.Response[bookingv1.ConfirmAppointmentResponse], error) {
	id, err := parseAppointmentID(req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	out, err := h.confirmUC.Execute(ctx, bookinguc.ConfirmInput{ID: id})
	if err != nil {
		return nil, mapDomainError(err)
	}
	return connect.NewResponse(&bookingv1.ConfirmAppointmentResponse{
		Appointment: toProtoAppointment(out.Appointment),
	}), nil
}

func (h *BookingHandler) CompleteAppointment(
	ctx context.Context,
	req *connect.Request[bookingv1.CompleteAppointmentRequest],
) (*connect.Response[bookingv1.CompleteAppointmentResponse], error) {
	id, err := parseAppointmentID(req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	out, err := h.completeUC.Execute(ctx, bookinguc.CompleteInput{
		ID:          id,
		MasterNotes: req.Msg.MasterNotes,
	})
	if err != nil {
		return nil, mapDomainError(err)
	}
	return connect.NewResponse(&bookingv1.CompleteAppointmentResponse{
		Appointment: toProtoAppointment(out.Appointment),
	}), nil
}

func (h *BookingHandler) MarkNoShow(
	ctx context.Context,
	req *connect.Request[bookingv1.MarkNoShowRequest],
) (*connect.Response[bookingv1.MarkNoShowResponse], error) {
	id, err := parseAppointmentID(req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	out, err := h.noShowUC.Execute(ctx, bookinguc.NoShowInput{ID: id})
	if err != nil {
		return nil, mapDomainError(err)
	}
	return connect.NewResponse(&bookingv1.MarkNoShowResponse{
		Appointment: toProtoAppointment(out.Appointment),
	}), nil
}

func toProtoAppointment(card *domain.AppointmentCard) *bookingv1.Appointment {
	if card == nil || card.Appointment == nil {
		return nil
	}
	a := card.Appointment

	var cancelledByUserID string
	if a.CancelledBy != nil {
		cancelledByUserID = strconv.Itoa(int(*a.CancelledBy))
	}

	var createdBy string
	if a.CreatedBy != nil {
		createdBy = strconv.Itoa(int(*a.CreatedBy))
	}

	var cancelledAt *timestamppb.Timestamp
	if a.CancelledAt != nil {
		cancelledAt = timestamppb.New(*a.CancelledAt)
	}

	var cancelledByRole bookingv1.CancelledBy
	if a.CancelledByRole != nil {
		cancelledByRole = bookingv1.CancelledBy(*a.CancelledByRole)
	}

	return &bookingv1.Appointment{
		Id:                strconv.FormatInt(a.ID, 10),
		ClientId:          strconv.Itoa(int(a.ClientID)),
		MasterId:          strconv.Itoa(int(a.MasterID)),
		ServiceId:         strconv.Itoa(int(a.ServiceID)),
		LocationId:        strconv.Itoa(int(a.LocationID)),
		Slug:              card.MasterSlug + "/" + card.ServiceSlug,
		MasterName:        card.MasterName,
		ServiceName:       card.ServiceName,
		LocationName:      card.LocationName,
		Address:           card.Address,
		StartTime:         timestamppb.New(a.StartTime),
		EndTime:           timestamppb.New(a.EndTime),
		Timezone:          a.Timezone,
		Status:            bookingv1.AppointmentStatus(a.Status),
		Price:             toMoney(a.Price, a.Currency),
		PaymentStatus:     bookingv1.PaymentStatus(a.PaymentStatus),
		IsOnline:          a.IsOnline,
		MeetingUrl:        a.MeetingURL,
		ClientNotes:       a.ClientNotes,
		MasterNotes:       a.MasterNotes,
		CancelledAt:       cancelledAt,
		CancelledByUserId: cancelledByUserID,
		CancelledByRole:   cancelledByRole,
		CancelReason:      a.CancelReason,
		CreatedAt:         timestamppb.New(a.CreatedAt),
		UpdatedAt:         timestamppb.New(a.UpdatedAt),
		CreatedBy:         createdBy,
	}
}

func toMoney(price int32, currency string) *money.Money {
	return &money.Money{
		CurrencyCode: currency,
		Units:        int64(price / 100),
		Nanos:        int32(price%100) * 10_000_000,
	}
}
