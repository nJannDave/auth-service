package handler

import (
	"auth/domain/entity"
	contract "auth/domain/interface"
	"strings"

	"context"
	"errors"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/nJannDave/pkg/const"
	"github.com/nJannDave/pkg/log"
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
		status, errorr := utils.ValidateErrHandler(err)
		if status == 500 {
			log.LogHSR(ctx, "failed register", "register", req.Spec().Procedure, err.Error())
		}
		return nil, errorr
	}
	return connect.NewResponse(&pb.Response{
		Status: true,
		Message: "registered successfully",
	}), nil
}

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
		status, errorr := utils.ValidateErrHandler(err)
		if status == 500 {
			log.LogHSR(ctx, "failed login", "login", req.Spec().Procedure, err.Error())
		}
		return nil, errorr
	}
	utils.SetCookie(ctx, string(constt.AT), tkn.Access, 3*time.Minute)
	utils.SetCookie(ctx, string(constt.RF), tkn.Refresh, 24*3*time.Hour)
	return connect.NewResponse(&pb.Response{
		Status: true,
		Message: "successfully login",
	}), nil
}

func (h *Handler) Refresh(
	ctx context.Context,
	req *connect.Request[emptypb.Empty],
) (*connect.Response[pb.Response], error) {
	rCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	token, err := utils.GetCookie(req, string(constt.RF))
	if err != nil {
		if strings.Contains(err.Error(), "internal server error: ") {
			log.LogHSR(ctx, "error while parsing cookie", "refresh", req.Spec().Procedure, err.Error())
			return nil, connect.NewError(connect.CodeInternal, errors.New("an error occured"))
		}
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}
	tkn, err := h.service.Refresh(rCtx, token)
	if err != nil {
		status, errorr := utils.ValidateErrHandler(err)
		if status == 500 {
			log.LogHSR(ctx, "failed refresh token", "refresh", req.Spec().Procedure, err.Error())
		}
		return nil, errorr
	}
	utils.SetCookie(ctx, string(constt.AT), tkn.Access, 3*time.Minute)
	utils.SetCookie(ctx, string(constt.RF), tkn.Refresh, 24*3*time.Hour)
	return connect.NewResponse(&pb.Response{
		Status: true,
		Message: "successfully refresh token",
	}), nil
}