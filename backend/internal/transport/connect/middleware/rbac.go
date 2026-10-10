package middleware

import (
	"context"
	"errors"

	"connectrpc.com/connect"

	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/appcontext"
	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
)

const (
	roleAdmin  = domain.RoleAdmin
	roleMaster = domain.RoleMaster
	roleClient = domain.RoleUser
)

var rolePermissions = map[string][]domain.Role{
	"/catalog.v1.AdminCatalogService/AddCategory":          {roleAdmin},
	"/catalog.v1.AdminCatalogService/UpdateCategory":       {roleAdmin},
	"/catalog.v1.AdminCatalogService/DeleteCategory":       {roleAdmin},
	"/catalog.v1.AdminCatalogService/AddDefaultService":    {roleAdmin},
	"/catalog.v1.AdminCatalogService/UpdateDefaultService": {roleAdmin},
	"/catalog.v1.AdminCatalogService/DeleteDefaultService": {roleAdmin},

	"/payment.v1.PaymentService/ListAllPayments": {roleAdmin},
	"/payment.v1.PaymentService/RefundPayment":   {roleAdmin},

	"/review.v1.ReviewService/HideReview": {roleAdmin},

	"/master.v1.MasterService/GetMyMasterProfile": {roleAdmin, roleMaster},
	"/master.v1.MasterService/UpdateMaster":       {roleAdmin, roleMaster},
	"/master.v1.MasterService/UpdateSlug":   	   {roleAdmin, roleMaster},

	"/service.v1.ServiceService/GetMyServices": {roleAdmin, roleMaster},
	"/service.v1.ServiceService/AddService":    {roleAdmin, roleMaster},
	"/service.v1.ServiceService/UpdateService": {roleAdmin, roleMaster},
	"/service.v1.ServiceService/DeleteService": {roleAdmin, roleMaster},

	"/schedule.v1.ScheduleService/GetScheduleRules":    {roleAdmin, roleMaster},
	"/schedule.v1.ScheduleService/UpdateScheduleRules": {roleAdmin, roleMaster},
	"/schedule.v1.ScheduleService/ListExceptions":      {roleAdmin, roleMaster},
	"/schedule.v1.ScheduleService/AddException":        {roleAdmin, roleMaster},
	"/schedule.v1.ScheduleService/DeleteException":     {roleAdmin, roleMaster},

	"/location.v1.LocationService/CreateLocation": {roleAdmin, roleMaster},
	"/location.v1.LocationService/UpdateLocation": {roleAdmin, roleMaster},
	"/location.v1.LocationService/DeleteLocation": {roleAdmin, roleMaster},

	"/user.v1.UserService/GetMe":      {roleAdmin, roleMaster, roleClient},
	"/user.v1.UserService/UpdateUser": {roleAdmin, roleMaster, roleClient},
	"/user.v1.UserService/GetUser":    {roleAdmin, roleClient},

	"/booking.v1.BookingService/CreateAppointment":      {roleAdmin, roleMaster, roleClient},
	"/booking.v1.BookingService/GetAppointment":         {roleAdmin, roleMaster, roleClient},
	"/booking.v1.BookingService/ListMyAppointments":     {roleAdmin, roleMaster, roleClient},
	"/booking.v1.BookingService/ListMasterAppointments": {roleAdmin, roleMaster},
	"/booking.v1.BookingService/CancelAppointment":      {roleAdmin, roleMaster, roleClient},
	"/booking.v1.BookingService/RescheduleAppointment":  {roleAdmin, roleMaster, roleClient},
	"/booking.v1.BookingService/ConfirmAppointment":     {roleAdmin, roleMaster},
	"/booking.v1.BookingService/CompleteAppointment":    {roleAdmin, roleMaster},
	"/booking.v1.BookingService/MarkNoShow":             {roleAdmin, roleMaster},

	"/review.v1.ReviewService/CreateReview":  {roleAdmin, roleClient},
	"/review.v1.ReviewService/UpdateReview":  {roleAdmin, roleClient},
	"/review.v1.ReviewService/DeleteReview":  {roleAdmin, roleClient},
	"/review.v1.ReviewService/ListMyReviews": {roleAdmin, roleClient},
	"/review.v1.ReviewService/ReplyToReview": {roleAdmin, roleMaster},

	"/payment.v1.PaymentService/CreatePayment":  {roleAdmin, roleClient},
	"/payment.v1.PaymentService/GetPayment":     {roleAdmin, roleClient},
	"/payment.v1.PaymentService/ListMyPayments": {roleAdmin, roleClient},
	"/payment.v1.PaymentService/CancelPayment":  {roleAdmin, roleClient},

	"/notification.v1.NotificationService/ListMyNotifications":        {roleAdmin, roleMaster, roleClient},
	"/notification.v1.NotificationService/MarkAsRead":                 {roleAdmin, roleMaster, roleClient},
	"/notification.v1.NotificationService/MarkAllAsRead":              {roleAdmin, roleMaster, roleClient},
	"/notification.v1.NotificationService/DeleteNotification":         {roleAdmin, roleMaster, roleClient},
	"/notification.v1.NotificationService/GetUnreadCount":             {roleAdmin, roleMaster, roleClient},
	"/notification.v1.NotificationService/GetNotificationSettings":    {roleAdmin, roleMaster, roleClient},
	"/notification.v1.NotificationService/UpdateNotificationSettings": {roleAdmin, roleMaster, roleClient},
}

func NewRBACInterceptor() connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			allowed, ok := rolePermissions[req.Spec().Procedure]
			if !ok {
				return next(ctx, req)
			}

			userRoles, ok := appcontext.Roles(ctx)
			if !ok {
				return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("no roles in context"))
			}

			for _, userRole := range userRoles {
				for _, allowedRole := range allowed {
					if domain.Role(userRole) == allowedRole {
						return next(ctx, req)
					}
				}
			}

			return nil, connect.NewError(connect.CodePermissionDenied, errors.New("forbidden"))
		}
	}
}
