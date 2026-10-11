package connect

import (
	"context"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/genproto/googleapis/type/dayofweek"
	"google.golang.org/genproto/googleapis/type/timeofday"
	"google.golang.org/genproto/googleapis/type/date"
	"google.golang.org/protobuf/types/known/timestamppb"

	schedulev1 "github.com/MuslimIT-dev/Calendly_GetCourse/backend/gen/go/schedule/v1"
	schedulev1connect "github.com/MuslimIT-dev/Calendly_GetCourse/backend/gen/go/schedule/v1/schedulev1connect"
	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
	scheduleuc "github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/usecase/schedule"
)

type ScheduleHandler struct {
	getMonthUC       *scheduleuc.GetMonthUseCase
	getDayUC         *scheduleuc.GetDayUseCase
	getNextUC        *scheduleuc.GetNextAvailableUseCase
	getRulesUC       *scheduleuc.GetRulesUseCase
	updateRulesUC    *scheduleuc.UpdateRulesUseCase
	listExceptionsUC *scheduleuc.ListExceptionsUseCase
	addExceptionUC   *scheduleuc.AddExceptionUseCase
	deleteExceptionUC *scheduleuc.DeleteExceptionUseCase
}

func NewScheduleHandler(
	getMonthUC *scheduleuc.GetMonthUseCase,
	getDayUC *scheduleuc.GetDayUseCase,
	getNextUC *scheduleuc.GetNextAvailableUseCase,
	getRulesUC *scheduleuc.GetRulesUseCase,
	updateRulesUC *scheduleuc.UpdateRulesUseCase,
	listExceptionsUC *scheduleuc.ListExceptionsUseCase,
	addExceptionUC *scheduleuc.AddExceptionUseCase,
	deleteExceptionUC *scheduleuc.DeleteExceptionUseCase,
) *ScheduleHandler {
	return &ScheduleHandler{
		getMonthUC:        getMonthUC,
		getDayUC:          getDayUC,
		getNextUC:         getNextUC,
		getRulesUC:        getRulesUC,
		updateRulesUC:     updateRulesUC,
		listExceptionsUC:  listExceptionsUC,
		addExceptionUC:    addExceptionUC,
		deleteExceptionUC: deleteExceptionUC,
	}
}

var _ schedulev1connect.ScheduleServiceHandler = (*ScheduleHandler)(nil)

func (h *ScheduleHandler) GetMonthSchedule(
	ctx context.Context,
	req *connect.Request[schedulev1.GetMonthScheduleRequest],
) (*connect.Response[schedulev1.GetMonthScheduleResponse], error) {
	masterID, _ := parseID(req.Msg.MasterId)
	serviceID, _ := parseID(req.Msg.ServiceId)
	locationID, _ := parseID(req.Msg.LocationId)

	out, err := h.getMonthUC.Execute(ctx, scheduleuc.GetMonthInput{
		MasterID:   masterID,
		ServiceID:  serviceID,
		LocationID: locationID,
		Year:       req.Msg.Year,
		Month:      req.Msg.Month,
		Timezone:   req.Msg.Timezone,
	})
	if err != nil {
		return nil, mapDomainError(err)
	}

	days := make([]*schedulev1.Day, len(out.Days))
	for i, d := range out.Days {
		days[i] = &schedulev1.Day{
			Date:       d.Date.Format("2006-01-02"),
			HasSlots:   d.HasSlots,
			SlotsCount: d.SlotsCount,
		}
	}
	return connect.NewResponse(&schedulev1.GetMonthScheduleResponse{Days: days}), nil
}

func (h *ScheduleHandler) GetDaySchedule(
	ctx context.Context,
	req *connect.Request[schedulev1.GetDayScheduleRequest],
) (*connect.Response[schedulev1.GetDayScheduleResponse], error) {
	masterID, _ := parseID(req.Msg.MasterId)
	serviceID, _ := parseID(req.Msg.ServiceId)
	locationID, _ := parseID(req.Msg.LocationId)

	date, err := time.Parse("2006-01-02", req.Msg.Date)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	out, err := h.getDayUC.Execute(ctx, scheduleuc.GetDayInput{
		MasterID:   masterID,
		ServiceID:  serviceID,
		LocationID: locationID,
		Date:       date,
		Timezone:   req.Msg.Timezone,
	})
	if err != nil {
		return nil, mapDomainError(err)
	}

	return connect.NewResponse(&schedulev1.GetDayScheduleResponse{
		Slots: toProtoSlots(out.Slots),
	}), nil
}

func (h *ScheduleHandler) GetNextAvailableSlot(
	ctx context.Context,
	req *connect.Request[schedulev1.GetNextAvailableSlotRequest],
) (*connect.Response[schedulev1.GetNextAvailableSlotResponse], error) {
	masterID, _ := parseID(req.Msg.MasterId)
	serviceID, _ := parseID(req.Msg.ServiceId)
	locationID, _ := parseID(req.Msg.LocationId)

	out, err := h.getNextUC.Execute(ctx, scheduleuc.GetNextAvailableInput{
		MasterID:   masterID,
		ServiceID:  serviceID,
		LocationID: locationID,
		Timezone:   req.Msg.Timezone,
	})
	if err != nil {
		return nil, mapDomainError(err)
	}
	return connect.NewResponse(&schedulev1.GetNextAvailableSlotResponse{
		Slots: toProtoSlots(out.Slots),
	}), nil
}

func (h *ScheduleHandler) GetScheduleRules(
	ctx context.Context,
	req *connect.Request[schedulev1.GetScheduleRulesRequest],
) (*connect.Response[schedulev1.GetScheduleRulesResponse], error) {
	locationID, _ := parseID(req.Msg.LocationId)

	out, err := h.getRulesUC.Execute(ctx, scheduleuc.GetRulesInput{LocationID: locationID})
	if err != nil {
		return nil, mapDomainError(err)
	}

	rules := make([]*schedulev1.ScheduleRule, len(out.Rules))
	for i, r := range out.Rules {
		rules[i] = toProtoRule(r)
	}
	return connect.NewResponse(&schedulev1.GetScheduleRulesResponse{ScheduleRules: rules}), nil
}

func (h *ScheduleHandler) UpdateScheduleRules(
	ctx context.Context,
	req *connect.Request[schedulev1.UpdateScheduleRulesRequest],
) (*connect.Response[schedulev1.UpdateScheduleRulesResponse], error) {
	locationID, _ := parseID(req.Msg.LocationId)

	rules := make([]*domain.ScheduleRule, len(req.Msg.Rules))
	for i, r := range req.Msg.Rules {
		rules[i] = fromProtoRule(r)
	}

	out, err := h.updateRulesUC.Execute(ctx, scheduleuc.UpdateRulesInput{
		LocationID: locationID,
		Rules:      rules,
	})
	if err != nil {
		return nil, mapDomainError(err)
	}

	protoRules := make([]*schedulev1.ScheduleRule, len(out.Rules))
	for i, r := range out.Rules {
		protoRules[i] = toProtoRule(r)
	}
	return connect.NewResponse(&schedulev1.UpdateScheduleRulesResponse{Rules: protoRules}), nil
}

func (h *ScheduleHandler) ListExceptions(
	ctx context.Context,
	req *connect.Request[schedulev1.ListExceptionsRequest],
) (*connect.Response[schedulev1.ListExceptionsResponse], error) {
	locationID, _ := parseID(req.Msg.LocationId)

	from, err := time.Parse("2006-01-02", req.Msg.FromDate)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	to, err := time.Parse("2006-01-02", req.Msg.ToDate)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	out, err := h.listExceptionsUC.Execute(ctx, scheduleuc.ListExceptionsInput{
		LocationID: locationID,
		From:       from,
		To:         to,
	})
	if err != nil {
		return nil, mapDomainError(err)
	}

	list := make([]*schedulev1.Exception, len(out.Exceptions))
	for i, e := range out.Exceptions {
		list[i] = toProtoException(e)
	}
	return connect.NewResponse(&schedulev1.ListExceptionsResponse{Exceptions: list}), nil
}

func (h *ScheduleHandler) AddException(
	ctx context.Context,
	req *connect.Request[schedulev1.AddExceptionRequest],
) (*connect.Response[schedulev1.AddExceptionResponse], error) {
	locationID, err := parseID(req.Msg.LocationId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	date, err := time.Parse("2006-01-02", req.Msg.Date)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	out, err := h.addExceptionUC.Execute(ctx, scheduleuc.AddExceptionInput{
		LocationID:  locationID,
		Date:        date,
		Type:        domain.ExceptionType(req.Msg.Type),
		IsFullDay:   req.Msg.IsFullDay,
		CustomStart: fromProtoTimeOfDay(req.Msg.CustomStart),
		CustomEnd:   fromProtoTimeOfDay(req.Msg.CustomEnd),
		RestStart:   fromProtoTimeOfDay(req.Msg.RestStart),
		RestEnd:     fromProtoTimeOfDay(req.Msg.RestEnd),
		Reason:      req.Msg.Reason,
	})
	if err != nil {
		return nil, mapDomainError(err)
	}
	return connect.NewResponse(&schedulev1.AddExceptionResponse{
		Exception: toProtoException(out.Exception),
	}), nil
}

func (h *ScheduleHandler) DeleteException(
	ctx context.Context,
	req *connect.Request[schedulev1.DeleteExceptionRequest],
) (*connect.Response[schedulev1.DeleteExceptionResponse], error) {
	id, _ := parseID(req.Msg.ExceptionId)
	_, err := h.deleteExceptionUC.Execute(ctx, scheduleuc.DeleteExceptionInput{ExceptionID: id})
	if err != nil {
		return nil, mapDomainError(err)
	}
	return connect.NewResponse(&schedulev1.DeleteExceptionResponse{}), nil
}

func toProtoSlots(in []domain.TimeSlot) []*schedulev1.TimeSlot {
	out := make([]*schedulev1.TimeSlot, len(in))
	for i, s := range in {
		out[i] = &schedulev1.TimeSlot{
			Start: timestamppb.New(s.Start),
			End:   timestamppb.New(s.End),
		}
	}
	return out
}

func toProtoRule(r *domain.ScheduleRule) *schedulev1.ScheduleRule {
	return &schedulev1.ScheduleRule{
		Id:                    formatID(r.ID),
		LocationId:            formatID(r.LocationID),
		DayOfWeek:             dayofweek.DayOfWeek(r.DayOfWeek),
		IsWorkingDay:          r.IsWorkingDay,
		WorkStart:             toProtoTimeOfDay(r.WorkStart),
		WorkEnd:               toProtoTimeOfDay(r.WorkEnd),
		RestStart:             toProtoTimeOfDay(r.RestStart),
		RestEnd:               toProtoTimeOfDay(r.RestEnd),
		BufferBetweenSlotsMins: r.BufferBetweenSlotsMin,
		SlotStepMins:          r.SlotStepMins,
		CreatedAt:             timestamppb.New(r.CreatedAt),
		UpdatedAt:             timestamppb.New(r.UpdatedAt),
	}
}

func fromProtoRule(r *schedulev1.ScheduleRule) *domain.ScheduleRule {
	id, _ := parseID(r.Id)
	locationID, _ := parseID(r.LocationId)

	return &domain.ScheduleRule{
		ID:                    id,
		LocationID:            locationID,
		DayOfWeek:             int32(r.DayOfWeek),
		IsWorkingDay:          r.IsWorkingDay,
		WorkStart:             fromProtoTimeOfDay(r.WorkStart),
		WorkEnd:               fromProtoTimeOfDay(r.WorkEnd),
		RestStart:             fromProtoTimeOfDay(r.RestStart),
		RestEnd:               fromProtoTimeOfDay(r.RestEnd),
		BufferBetweenSlotsMin: r.BufferBetweenSlotsMins,
		SlotStepMins:          r.SlotStepMins,
	}
}

func toProtoException(e *domain.Exception) *schedulev1.Exception {
func toProtoException(e *domain.Exception) *schedulev1.Exception {
	return &schedulev1.Exception{
		Id:          formatID(e.ID),
		LocationId:  formatID(e.LocationID),
		Date:        toProtoDate(e.Date),
		IsFullDay:   e.IsFullDay,
		Type:        schedulev1.ExceptionType(e.Type),
		CustomStart: toProtoTimeOfDay(e.CustomStart),
		CustomEnd:   toProtoTimeOfDay(e.CustomEnd),
		RestStart:   toProtoTimeOfDay(e.RestStart),
		RestEnd:     toProtoTimeOfDay(e.RestEnd),
		Reason:      e.Reason,
		CreatedAt:   timestamppb.New(e.CreatedAt),
		UpdatedAt:   timestamppb.New(e.UpdatedAt),
	}
}

func toProtoDate(t time.Time) *date.Date {
	return &date.Date{
		Year:  int32(t.Year()),
		Month: int32(t.Month()),
		Day:   int32(t.Day()),
	}
}

func toProtoTimeOfDay(t *domain.TimeOfDay) *timeofday.TimeOfDay {
	if t == nil {
		return nil
	}
	return &timeofday.TimeOfDay{
		Hours:   t.Hours,
		Minutes: t.Minutes,
		Seconds: t.Seconds,
	}
}

func fromProtoTimeOfDay(t *timeofday.TimeOfDay) *domain.TimeOfDay {
	if t == nil {
		return nil
	}
	return &domain.TimeOfDay{
		Hours:   t.Hours,
		Minutes: t.Minutes,
		Seconds: t.Seconds,
	}
}