package connect

import (
	"context"
	"strconv"

	"connectrpc.com/connect"

	authv1 "github.com/MuslimIT-dev/Calendly_GetCourse/backend/gen/go/auth/v1"
	authv1connect "github.com/MuslimIT-dev/Calendly_GetCourse/backend/gen/go/auth/v1/authv1connect"
	userv1 "github.com/MuslimIT-dev/Calendly_GetCourse/backend/gen/go/user/v1"
	"github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/domain"
	authuc "github.com/MuslimIT-dev/Calendly_GetCourse/backend/internal/usecase/auth"
)

type AuthHandler struct {
	registerUC *authuc.RegisterUseCase
	loginUC    *authuc.LoginUseCase
}

func NewAuthHandler(
	registerUC *authuc.RegisterUseCase,
	loginUC *authuc.LoginUseCase,
) *AuthHandler {
	return &AuthHandler{
		registerUC: registerUC,
		loginUC:    loginUC,
	}
}

var _ authv1connect.AuthServiceHandler = (*AuthHandler)(nil)

func (h *AuthHandler) Register(
	ctx context.Context,
	req *connect.Request[authv1.RegisterRequest],
) (*connect.Response[authv1.RegisterResponse], error) {
	out, err := h.registerUC.Execute(ctx, authuc.RegisterInput{
		Name:     req.Msg.Name,
		Email:    req.Msg.Email,
		Password: req.Msg.Password,
		Role:     domain.Role(req.Msg.Role),
	})
	if err != nil {
		return nil, mapDomainError(err)
	}

	resp := &authv1.RegisterResponse{
		User: &userv1.User{
			Id:            strconv.Itoa(int(out.User.ID)),
			Name:          out.User.Name,
			Email:         out.User.Email,
			EmailVerified: out.User.EmailVerified,
		},
		Tokens: &authv1.TokenPair{
			AccessToken:  out.AccessToken,
			RefreshToken: out.RefreshToken,
			ExpiresIn:    out.ExpiresIn,
		},
	}

	return connect.NewResponse(resp), nil
}

func (h *AuthHandler) Login(
	ctx context.Context,
	req *connect.Request[authv1.LoginRequest],
) (*connect.Response[authv1.LoginResponse], error) {
	ip := req.Peer().Addr
	ua := req.Header().Get("User-Agent")

	out, err := h.loginUC.Execute(ctx, authuc.LoginInput{
		Email:     req.Msg.Email,
		Password:  req.Msg.Password,
		IPAddress: ip,
		UserAgent: ua,
	})
	if err != nil {
		return nil, mapDomainError(err)
	}

	resp := &authv1.LoginResponse{
		User: &userv1.User{
			Id:            strconv.Itoa(int(out.User.ID)),
			Name:          out.User.Name,
			Email:         out.User.Email,
			EmailVerified: out.User.EmailVerified,
			AvatarUrl:     out.User.AvatarURL,
			Timezone:      out.User.Timezone,
		},
		Tokens: &authv1.TokenPair{
			AccessToken:  out.AccessToken,
			RefreshToken: out.RefreshToken,
			ExpiresIn:    out.ExpiresIn,
		},
	}

	return connect.NewResponse(resp), nil
}

func (h *AuthHandler) RefreshToken(
	ctx context.Context,
	req *connect.Request[authv1.RefreshTokenRequest],
) (*connect.Response[authv1.RefreshTokenResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *AuthHandler) VerifyEmail(
	ctx context.Context,
	req *connect.Request[authv1.VerifyEmailRequest],
) (*connect.Response[authv1.VerifyEmailResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *AuthHandler) ForgotPassword(
	ctx context.Context,
	req *connect.Request[authv1.ForgotPasswordRequest],
) (*connect.Response[authv1.ForgotPasswordResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *AuthHandler) ResetPassword(
	ctx context.Context,
	req *connect.Request[authv1.ResetPasswordRequest],
) (*connect.Response[authv1.ResetPasswordResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *AuthHandler) Logout(
	ctx context.Context,
	req *connect.Request[authv1.LogoutRequest],
) (*connect.Response[authv1.LogoutResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *AuthHandler) LogoutAll(
	ctx context.Context,
	req *connect.Request[authv1.LogoutAllRequest],
) (*connect.Response[authv1.LogoutAllResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *AuthHandler) ChangePassword(
	ctx context.Context,
	req *connect.Request[authv1.ChangePasswordRequest],
) (*connect.Response[authv1.ChangePasswordResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}

func (h *AuthHandler) ListSessions(
	ctx context.Context,
	req *connect.Request[authv1.ListSessionsRequest],
) (*connect.Response[authv1.ListSessionsResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, nil)
}
