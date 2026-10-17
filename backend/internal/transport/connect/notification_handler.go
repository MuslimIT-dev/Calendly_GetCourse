package connect

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	notificationv1 "github.com/MuslimIT-dev/zaman/backend/gen/go/notification/v1"
	notificationv1connect "github.com/MuslimIT-dev/zaman/backend/gen/go/notification/v1/notificationv1connect"
	"github.com/MuslimIT-dev/zaman/backend/internal/domain"
	notificationuc "github.com/MuslimIT-dev/zaman/backend/internal/usecase/notification"
)

type NotificationHandler struct {
	listMyUC         *notificationuc.ListMyUseCase
	markReadUC       *notificationuc.MarkReadUseCase
	markAllReadUC    *notificationuc.MarkAllReadUseCase
	deleteUC         *notificationuc.DeleteUseCase
	unreadCountUC    *notificationuc.UnreadCountUseCase
	getSettingsUC    *notificationuc.GetSettingsUseCase
	updateSettingsUC *notificationuc.UpdateSettingsUseCase
	sendUC           *notificationuc.SendUseCase
}

func NewNotificationHandler(
	listMyUC *notificationuc.ListMyUseCase,
	markReadUC *notificationuc.MarkReadUseCase,
	markAllReadUC *notificationuc.MarkAllReadUseCase,
	deleteUC *notificationuc.DeleteUseCase,
	unreadCountUC *notificationuc.UnreadCountUseCase,
	getSettingsUC *notificationuc.GetSettingsUseCase,
	updateSettingsUC *notificationuc.UpdateSettingsUseCase,
	sendUC *notificationuc.SendUseCase,
) *NotificationHandler {
	return &NotificationHandler{
		listMyUC: listMyUC, markReadUC: markReadUC, markAllReadUC: markAllReadUC,
		deleteUC: deleteUC, unreadCountUC: unreadCountUC, getSettingsUC: getSettingsUC,
		updateSettingsUC: updateSettingsUC, sendUC: sendUC,
	}
}

var _ notificationv1connect.NotificationServiceHandler = (*NotificationHandler)(nil)

func (h *NotificationHandler) ListMyNotifications(ctx context.Context, req *connect.Request[notificationv1.ListMyNotificationsRequest]) (*connect.Response[notificationv1.ListMyNotificationsResponse], error) {
	var status *domain.NotificationStatus
	if req.Msg.StatusFilter != notificationv1.NotificationStatus_NOTIFICATION_STATUS_UNSPECIFIED {
		value := domain.NotificationStatus(req.Msg.StatusFilter)
		status = &value
	}
	out, err := h.listMyUC.Execute(ctx, notificationuc.ListMyInput{
		StatusFilter: status, PageSize: req.Msg.PageSize, PageToken: req.Msg.PageToken,
	})
	if err != nil {
		return nil, mapDomainError(err)
	}
	notifications := make([]*notificationv1.Notification, len(out.Notifications))
	for i, notification := range out.Notifications {
		notifications[i] = toProtoNotification(notification)
	}
	return connect.NewResponse(&notificationv1.ListMyNotificationsResponse{
		Notifications: notifications, NextPageToken: out.NextPageToken, UnreadCount: out.UnreadCount,
	}), nil
}

func (h *NotificationHandler) MarkAsRead(ctx context.Context, req *connect.Request[notificationv1.MarkAsReadRequest]) (*connect.Response[notificationv1.MarkAsReadResponse], error) {
	id, err := parseNotificationID(req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	if _, err := h.markReadUC.Execute(ctx, notificationuc.MarkReadInput{ID: id}); err != nil {
		return nil, mapDomainError(err)
	}
	return connect.NewResponse(&notificationv1.MarkAsReadResponse{}), nil
}

func (h *NotificationHandler) MarkAllAsRead(ctx context.Context, _ *connect.Request[notificationv1.MarkAllAsReadRequest]) (*connect.Response[notificationv1.MarkAllAsReadResponse], error) {
	out, err := h.markAllReadUC.Execute(ctx, notificationuc.MarkAllReadInput{})
	if err != nil {
		return nil, mapDomainError(err)
	}
	return connect.NewResponse(&notificationv1.MarkAllAsReadResponse{UpdatedCount: out.UpdatedCount}), nil
}

func (h *NotificationHandler) DeleteNotification(ctx context.Context, req *connect.Request[notificationv1.DeleteNotificationRequest]) (*connect.Response[notificationv1.DeleteNotificationResponse], error) {
	id, err := parseNotificationID(req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	if _, err := h.deleteUC.Execute(ctx, notificationuc.DeleteInput{ID: id}); err != nil {
		return nil, mapDomainError(err)
	}
	return connect.NewResponse(&notificationv1.DeleteNotificationResponse{}), nil
}

func (h *NotificationHandler) GetUnreadCount(ctx context.Context, _ *connect.Request[notificationv1.GetUnreadCountRequest]) (*connect.Response[notificationv1.GetUnreadCountResponse], error) {
	out, err := h.unreadCountUC.Execute(ctx, notificationuc.UnreadCountInput{})
	if err != nil {
		return nil, mapDomainError(err)
	}
	return connect.NewResponse(&notificationv1.GetUnreadCountResponse{Count: out.Count}), nil
}

func (h *NotificationHandler) GetNotificationSettings(ctx context.Context, _ *connect.Request[notificationv1.GetNotificationSettingsRequest]) (*connect.Response[notificationv1.GetNotificationSettingsResponse], error) {
	out, err := h.getSettingsUC.Execute(ctx, notificationuc.GetSettingsInput{})
	if err != nil {
		return nil, mapDomainError(err)
	}
	return connect.NewResponse(&notificationv1.GetNotificationSettingsResponse{Settings: toProtoNotificationSettings(out.Settings)}), nil
}

func (h *NotificationHandler) UpdateNotificationSettings(ctx context.Context, req *connect.Request[notificationv1.UpdateNotificationSettingsRequest]) (*connect.Response[notificationv1.UpdateNotificationSettingsResponse], error) {
	mutedTypes := make([]int32, len(req.Msg.MutedTypes))
	for i, notificationType := range req.Msg.MutedTypes {
		mutedTypes[i] = int32(notificationType)
	}
	out, err := h.updateSettingsUC.Execute(ctx, notificationuc.UpdateSettingsInput{
		EmailEnabled: req.Msg.EmailEnabled, TelegramEnabled: req.Msg.TelegramEnabled,
		SMSEnabled: req.Msg.SmsEnabled, PushEnabled: req.Msg.PushEnabled,
		MutedTypes: mutedTypes, Reminder24h: req.Msg.Reminder_24H,
		Reminder2h: req.Msg.Reminder_2H, Reminder30m: req.Msg.Reminder_30M,
	})
	if err != nil {
		return nil, mapDomainError(err)
	}
	return connect.NewResponse(&notificationv1.UpdateNotificationSettingsResponse{Settings: toProtoNotificationSettings(out.Settings)}), nil
}

func (h *NotificationHandler) SendNotification(ctx context.Context, req *connect.Request[notificationv1.SendNotificationRequest]) (*connect.Response[notificationv1.SendNotificationResponse], error) {
	userID, err := parseID(req.Msg.UserId)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	var metadata map[string]any
	if req.Msg.Metadata != "" {
		if err := json.Unmarshal([]byte(req.Msg.Metadata), &metadata); err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
	}
	out, err := h.sendUC.Execute(ctx, notificationuc.SendInput{
		UserID: userID, Type: domain.NotificationType(req.Msg.Type),
		Channel: domain.NotificationChannel(req.Msg.Channel), Title: req.Msg.Title,
		Body: req.Msg.Body, Link: req.Msg.Link, Metadata: metadata,
	})
	if err != nil {
		return nil, mapDomainError(err)
	}
	return connect.NewResponse(&notificationv1.SendNotificationResponse{Notification: toProtoNotification(out.Notification)}), nil
}

func parseNotificationID(value string) (int64, error) {
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid notification ID %q", value)
	}
	return id, nil
}

func toProtoNotification(n *domain.Notification) *notificationv1.Notification {
	if n == nil {
		return nil
	}
	var sentAt, readAt *timestamppb.Timestamp
	if n.SentAt != nil {
		sentAt = timestamppb.New(*n.SentAt)
	}
	if n.ReadAt != nil {
		readAt = timestamppb.New(*n.ReadAt)
	}
	return &notificationv1.Notification{
		Id: strconv.FormatInt(n.ID, 10), UserId: strconv.FormatInt(int64(n.UserID), 10),
		Type: notificationv1.NotificationType(n.Type), Channel: notificationv1.NotificationChannel(n.Channel),
		Status: notificationv1.NotificationStatus(n.Status), Title: n.Title, Body: n.Body,
		Link: n.Link, Metadata: string(n.Metadata), SentAt: sentAt, ReadAt: readAt,
		CreatedAt: timestamppb.New(n.CreatedAt),
	}
}

func toProtoNotificationSettings(settings *domain.NotificationSettings) *notificationv1.NotificationSettings {
	if settings == nil {
		return nil
	}
	mutedTypes := make([]notificationv1.NotificationType, len(settings.MutedTypes))
	for i, notificationType := range settings.MutedTypes {
		mutedTypes[i] = notificationv1.NotificationType(notificationType)
	}
	return &notificationv1.NotificationSettings{
		UserId:       strconv.FormatInt(int64(settings.UserID), 10),
		EmailEnabled: settings.EmailEnabled, TelegramEnabled: settings.TelegramEnabled,
		SmsEnabled: settings.SMSEnabled, PushEnabled: settings.PushEnabled,
		MutedTypes: mutedTypes, Reminder_24H: settings.Reminder24h,
		Reminder_2H: settings.Reminder2h, Reminder_30M: settings.Reminder30m,
		UpdatedAt: timestamppb.New(settings.UpdatedAt),
	}
}
