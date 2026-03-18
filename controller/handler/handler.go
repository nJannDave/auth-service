package handler

import (
	"auth/domain/entity"
	_ "auth/dto"
	contract "auth/domain/interface"

	"context"
	"errors"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/nJannDave/pkg/const"
	pb "github.com/nJannDave/pkg/pb/auth"
	pbc "github.com/nJannDave/pkg/pb/auth/authconnect"
	utils "github.com/nJannDave/pkg/utils/handler"
)

type Handler struct {
	proto pbc.UnimplementedAuthServiceHandler
	service contract.Service
}

func InitHandler(service contract.Service) *Handler {
	return &Handler{
		service: service,
	}
}

// Register godoc
// @Tags Auth
// @Summary make account or register
// @Description make account or register
// @Accept json
// @Produce json
// @Param Idempotency-Key header string true "idempotency key"
// @Param data body pb.UserData true "account data"
// @Success 200 {object} dto.ResponseRegister "success register"
// @Failure 500 {object} dto.ErrorInternal "if error internal"
// @Router /authService.AuthService/Register [post]
func (h *Handler) Register(
	ctx context.Context, 
	req *connect.Request[pb.UserData],
) (*connect.Response[pb.Response], error) {
	rCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := req.Msg.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	idempotencyKey := req.Header().Get(string(constt.IK))
	if idempotencyKey == "" {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("idempotency Key not found"))
	}
	userData := entity.NewUserData(req.Msg.Nik, req.Msg.Name, req.Msg.Password)
	residence := entity.NewResidence(req.Msg.Province, req.Msg.City)
	if err := h.service.Register(rCtx, *userData, *residence, idempotencyKey); err != nil {
		return nil, err
	}
	return connect.NewResponse(&pb.Response{
		Status: true,
		Message: "registered successfully",
	}), nil
}

// Login godoc
// @Tags Auth
// @Summary login for get access
// @Description login for get access
// @Accept json
// @Produce json
// @Param data body pb.LoginData true "account data"
// @Success 200 {object} dto.ResponseLogin "success login"
// @Failure 500 {object} dto.ErrorInternal "if error internal"
// @Router /authService.AuthService/Login [post]
func (h *Handler) Login(
	ctx context.Context,
	req *connect.Request[pb.LoginData],
) (*connect.Response[pb.Response], error) {
	rCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := req.Msg.Validate(); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	userData := entity.NewUserData(req.Msg.Nik, req.Msg.Name, req.Msg.Password)
	tkn, err := h.service.Login(rCtx, *userData)
	if err != nil {
		return nil, err
	}
	utils.SetCookie(ctx, string(constt.AT), tkn.Access, 3*time.Minute)
	utils.SetCookie(ctx, string(constt.RF), tkn.Refresh, 24*3*time.Hour)
	return connect.NewResponse(&pb.Response{
		Status: true,
		Message: "successfully login",
	}), nil
}

// Refresh godoc
// @Tags Auth
// @Summary for get a new token
// @Description for get a new token
// @Produce json
// @Security Refresh-Token
// @Param data body dto.Empty false "empty"
// @Success 200 {object} dto.ResponseRefresh "success login"
// @Failure 500 {object} dto.ErrorInternal "if error internal"
// @Router /authService.AuthService/Refresh [post]
func (h *Handler) Refresh(
	ctx context.Context,
	req *connect.Request[emptypb.Empty],
) (*connect.Response[pb.Response], error) {
	rCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	token, err := utils.GetCookie(req, string(constt.RF))
	if err != nil {
		return nil, err
	}
	tkn, err := h.service.Refresh(rCtx, token)
	if err != nil {
		return nil, err
	}
	utils.SetCookie(ctx, string(constt.AT), tkn.Access, 3*time.Minute)
	utils.SetCookie(ctx, string(constt.RF), tkn.Refresh, 24*3*time.Hour)
	return connect.NewResponse(&pb.Response{
		Status: true,
		Message: "successfully refresh token",
	}), nil
}

// Logout godoc
// @Tags Auth
// @Summary logout
// @Description logout
// @Produce json
// @Security Access-Token && Refresh-Token
// @Param data body dto.Empty false "empty"
// @Success 200 {object} dto.ResponseLogout "success login"
// @Failure 500 {object} dto.ErrorInternal "if error internal"
// @Router /authService.AuthService/Logout [post]
func (h *Handler) Logout(
	ctx context.Context,
	req *connect.Request[emptypb.Empty],
) (*connect.Response[pb.Response], error) {
	rCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	rTkn, err := utils.GetCookie(req, string(constt.RF))
	if err != nil {
		return nil, err
	}
	aTkn, err := utils.GetCookie(req, string(constt.AT))
	if err != nil {
		return nil, err
	}
	if rTkn == "" || aTkn == "" { return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("please login")) }
	if err := h.service.Logout(rCtx, aTkn, rTkn); err != nil {
		return nil, err
	}
	utils.SetCookie(ctx, string(constt.AT), "", 0)
	utils.SetCookie(ctx, string(constt.RF), "", 0)
	return connect.NewResponse(&pb.Response{
		Status: true,
		Message: "successfully logout",
	}), nil
}