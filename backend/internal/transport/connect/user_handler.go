package connect

import (
	"context"
	"strconv"

	"connectrpc.com/connect"

	userv1 "github.com/MuslimIT-dev/Calendly_GetCourse/backend/gen/go/user/v1"
	userv1connect "github.com/MuslimIT-dev/Calendly_GetCourse/backend/gen/go/user/v1/userv1connect"
	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
	useruc "github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/usecase/user"
)

type UserHandler struct {
	getUserUC    *useruc.GetUserUseCase
	getMeUC      *useruc.GetMeUseCase
	updateUserUC *useruc.UpdateUserUseCase
}

func NewUserHandler(
	getUserUC *useruc.GetUserUseCase,
	getMeUC *useruc.GetMeUseCase,
	updateUserUC *useruc.UpdateUserUseCase,
) *UserHandler {
	return &UserHandler{
		getUserUC:    getUserUC,
		getMeUC:      getMeUC,
		updateUserUC: updateUserUC,
	}
}

var _ userv1connect.UserServiceHandler = (*UserHandler)(nil)

func (h *UserHandler) GetUser(
	ctx context.Context,
	req *connect.Request[userv1.GetUserRequest],
) (*connect.Response[userv1.GetUserResponse], error) {
	id, err := strconv.ParseInt(req.Msg.Id, 10, 32)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	out, err := h.getUserUC.Execute(ctx, useruc.GetUserInput{UserID: int32(id)})
	if err != nil {
		return nil, mapDomainError(err)
	}

	return connect.NewResponse(&userv1.GetUserResponse{
		User: toProtoUser(out.User),
	}), nil
}

func (h *UserHandler) GetMe(
	ctx context.Context,
	req *connect.Request[userv1.GetMeRequest],
) (*connect.Response[userv1.GetMeResponse], error) {
	out, err := h.getMeUC.Execute(ctx, useruc.GetMeInput{})
	if err != nil {
		return nil, mapDomainError(err)
	}

	return connect.NewResponse(&userv1.GetMeResponse{
		User: toProtoUser(out.User),
	}), nil
}

func (h *UserHandler) UpdateUser(
	ctx context.Context,
	req *connect.Request[userv1.UpdateUserRequest],
) (*connect.Response[userv1.UpdateUserResponse], error) {
	in := useruc.UpdateUserInput{}

	if req.Msg.Name != "" {
		in.Name = &req.Msg.Name
	}
	if req.Msg.AvatarUrl != "" {
		in.AvatarURL = &req.Msg.AvatarUrl
	}
	if req.Msg.Timezone != "" {
		in.Timezone = &req.Msg.Timezone
	}

	out, err := h.updateUserUC.Execute(ctx, in)
	if err != nil {
		return nil, mapDomainError(err)
	}

	return connect.NewResponse(&userv1.UpdateUserResponse{
		User: toProtoUser(out.User),
	}), nil
}

func toProtoUser(u *domain.User) *userv1.User {
	roles := make([]userv1.Role, len(u.Roles))
	for i, r := range u.Roles {
		roles[i] = userv1.Role(r)
	}

	return &userv1.User{
		Id:            strconv.Itoa(int(u.ID)),
		Name:          u.Name,
		Email:         u.Email,
		AvatarUrl:     u.AvatarURL,
		Timezone:      u.Timezone,
		EmailVerified: u.EmailVerified,
		Roles:         roles,
	}
}