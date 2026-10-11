package connect

import (
	"context"
	"strconv"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/genproto/googleapis/type/money"
	"google.golang.org/protobuf/types/known/timestamppb"

	paymentv1 "github.com/MuslimIT-dev/Calendly_GetCourse/backend/gen/go/payment/v1"
	paymentv1connect "github.com/MuslimIT-dev/Calendly_GetCourse/backend/gen/go/payment/v1/paymentv1connect"
	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
	paymentuc "github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/usecase/payment"
)

type PaymentHandler struct {
	createUC  *paymentuc.CreateUseCase
	getUC     *paymentuc.GetUseCase
	listMyUC  *paymentuc.ListMyUseCase
	listAllUC *paymentuc.ListAllUseCase
	refundUC  *paymentuc.RefundUseCase
	cancelUC  *paymentuc.CancelUseCase
	webhookUC *paymentuc.WebhookUseCase
}

func NewPaymentHandler(
	createUC *paymentuc.CreateUseCase,
	getUC *paymentuc.GetUseCase,
	listMyUC *paymentuc.ListMyUseCase,
	listAllUC *paymentuc.ListAllUseCase,
	refundUC *paymentuc.RefundUseCase,
	cancelUC *paymentuc.CancelUseCase,
	webhookUC *paymentuc.WebhookUseCase,
) *PaymentHandler {
	return &PaymentHandler{
		createUC:  createUC,
		getUC:     getUC,
		listMyUC:  listMyUC,
		listAllUC: listAllUC,
		refundUC:  refundUC,
		cancelUC:  cancelUC,
		webhookUC: webhookUC,
	}
}

var _ paymentv1connect.PaymentServiceHandler = (*PaymentHandler)(nil)

func (h *PaymentHandler) CreatePayment(
	ctx context.Context,
	req *connect.Request[paymentv1.CreatePaymentRequest],
) (*connect.Response[paymentv1.CreatePaymentResponse], error) {
	out, err := h.createUC.Execute(ctx, paymentuc.CreateInput{
		Purpose:     domain.PaymentPurpose(req.Msg.Purpose),
		Provider:    domain.PaymentProvider(req.Msg.Provider),
		ReferenceID: req.Msg.ReferenceId,
		ReturnURL:   req.Msg.ReturnUrl,
		Amount:      moneyToKopecks(req.Msg.Amount),
		Currency:    req.Msg.Amount.GetCurrencyCode(),
	})
	if err != nil {
		return nil, mapDomainError(err)
	}
	return connect.NewResponse(&paymentv1.CreatePaymentResponse{
		Payment: toProtoPayment(out.Payment),
	}), nil
}

func (h *PaymentHandler) GetPayment(
	ctx context.Context,
	req *connect.Request[paymentv1.GetPaymentRequest],
) (*connect.Response[paymentv1.GetPaymentResponse], error) {
	id, err := parsePaymentID(req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	out, err := h.getUC.Execute(ctx, paymentuc.GetInput{ID: id})
	if err != nil {
		return nil, mapDomainError(err)
	}
	return connect.NewResponse(&paymentv1.GetPaymentResponse{
		Payment: toProtoPayment(out.Payment),
	}), nil
}

func (h *PaymentHandler) ListMyPayments(
	ctx context.Context,
	req *connect.Request[paymentv1.ListMyPaymentsRequest],
) (*connect.Response[paymentv1.ListMyPaymentsResponse], error) {
	var sf *domain.PaymentStatus
	if req.Msg.StatusFilter != paymentv1.PaymentStatus_PAYMENT_STATUS_UNSPECIFIED {
		v := domain.PaymentStatus(req.Msg.StatusFilter)
		sf = &v
	}
	var pf *domain.PaymentPurpose
	if req.Msg.PurposeFilter != paymentv1.PaymentPurpose_PAYMENT_PURPOSE_UNSPECIFIED {
		v := domain.PaymentPurpose(req.Msg.PurposeFilter)
		pf = &v
	}

	out, err := h.listMyUC.Execute(ctx, paymentuc.ListMyInput{
		StatusFilter:  sf,
		PurposeFilter: pf,
		PageSize:      req.Msg.PageSize,
		PageToken:     req.Msg.PageToken,
	})
	if err != nil {
		return nil, mapDomainError(err)
	}

	list := make([]*paymentv1.Payment, len(out.Payments))
	for i, p := range out.Payments {
		list[i] = toProtoPayment(p)
	}
	return connect.NewResponse(&paymentv1.ListMyPaymentsResponse{
		Payments:      list,
		NextPageToken: out.NextPageToken,
	}), nil
}

func (h *PaymentHandler) ListAllPayments(
	ctx context.Context,
	req *connect.Request[paymentv1.ListAllPaymentsRequest],
) (*connect.Response[paymentv1.ListAllPaymentsResponse], error) {
	var sf *domain.PaymentStatus
	if req.Msg.StatusFilter != paymentv1.PaymentStatus_PAYMENT_STATUS_UNSPECIFIED {
		v := domain.PaymentStatus(req.Msg.StatusFilter)
		sf = &v
	}
	var uid *int32
	if req.Msg.UserId != "" {
		id, err := parseID(req.Msg.UserId)
		if err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
		uid = &id
	}

	out, err := h.listAllUC.Execute(ctx, paymentuc.ListAllInput{
		StatusFilter: sf,
		UserID:       uid,
		FromDate:     req.Msg.FromDate,
		ToDate:       req.Msg.ToDate,
		PageSize:     req.Msg.PageSize,
		PageToken:    req.Msg.PageToken,
	})
	if err != nil {
		return nil, mapDomainError(err)
	}

	list := make([]*paymentv1.Payment, len(out.Payments))
	for i, p := range out.Payments {
		list[i] = toProtoPayment(p)
	}
	return connect.NewResponse(&paymentv1.ListAllPaymentsResponse{
		Payments:      list,
		NextPageToken: out.NextPageToken,
	}), nil
}

func (h *PaymentHandler) RefundPayment(
	ctx context.Context,
	req *connect.Request[paymentv1.RefundPaymentRequest],
) (*connect.Response[paymentv1.RefundPaymentResponse], error) {
	id, err := parsePaymentID(req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	out, err := h.refundUC.Execute(ctx, paymentuc.RefundInput{ID: id, Reason: req.Msg.Reason})
	if err != nil {
		return nil, mapDomainError(err)
	}
	return connect.NewResponse(&paymentv1.RefundPaymentResponse{
		Payment: toProtoPayment(out.Payment),
	}), nil
}

func (h *PaymentHandler) CancelPayment(
	ctx context.Context,
	req *connect.Request[paymentv1.CancelPaymentRequest],
) (*connect.Response[paymentv1.CancelPaymentResponse], error) {
	id, err := parsePaymentID(req.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	out, err := h.cancelUC.Execute(ctx, paymentuc.CancelInput{ID: id})
	if err != nil {
		return nil, mapDomainError(err)
	}
	return connect.NewResponse(&paymentv1.CancelPaymentResponse{
		Payment: toProtoPayment(out.Payment),
	}), nil
}

func (h *PaymentHandler) PaymentWebhook(
	ctx context.Context,
	req *connect.Request[paymentv1.PaymentWebhookRequest],
) (*connect.Response[paymentv1.PaymentWebhookResponse], error) {
	out, err := h.webhookUC.Execute(ctx, paymentuc.WebhookInput{
		Provider:  domain.PaymentProvider(req.Msg.Provider),
		Payload:   req.Msg.Payload,
		Signature: req.Msg.Signature,
	})
	if err != nil {
		return nil, mapDomainError(err)
	}
	return connect.NewResponse(&paymentv1.PaymentWebhookResponse{
		Accepted: out.Accepted,
	}), nil
}

func toProtoPayment(p *domain.Payment) *paymentv1.Payment {
	var paidAt, refundedAt *timestamppb.Timestamp
	if p.PaidAt != nil {
		paidAt = timestamppb.New(*p.PaidAt)
	}
	if p.RefundedAt != nil {
		refundedAt = timestamppb.New(*p.RefundedAt)
	}

	return &paymentv1.Payment{
		Id:            strconv.FormatInt(p.ID, 10),
		UserId:        strconv.Itoa(int(p.UserID)),
		Amount:        kopecksToMoney(p.Amount, p.Currency),
		Status:        paymentv1.PaymentStatus(p.Status),
		Provider:      paymentv1.PaymentProvider(p.Provider),
		Purpose:       paymentv1.PaymentPurpose(p.Purpose),
		ReferenceId:   p.ReferenceID,
		ExternalId:    p.ExternalID,
		PaymentUrl:    p.PaymentURL,
		FailureReason: p.FailureReason,
		ReceiptUrl:    p.ReceiptURL,
		PaidAt:        paidAt,
		RefundedAt:    refundedAt,
		ExpiresAt:     timestamppb.New(p.ExpiresAt),
		CreatedAt:     timestamppb.New(p.CreatedAt),
		UpdatedAt:     timestamppb.New(p.UpdatedAt),
	}
}

func kopecksToMoney(price int32, currency string) *money.Money {
	return &money.Money{
		CurrencyCode: currency,
		Units:        int64(price / 100),
		Nanos:        int32(price%100) * 10_000_000,
	}
}

func moneyToKopecks(m *money.Money) int32 {
	if m == nil {
		return 0
	}
	return int32(m.Units*100 + int64(m.Nanos)/10_000_000)
}

var _ = time.Now
