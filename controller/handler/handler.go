package handler

import (
	"auth/domain/entity"
	contract "auth/domain/interface"

	"errors"
	"time"
	"connectrpc.com/connect"
	"context"

	"github.com/nJannDave/pkg/log"
	utils "github.com/nJannDave/pkg/utils/handler"
	"github.com/nJannDave/pkg/const"
	pb "github.com/nJannDave/pkg/pb/auth"
	pbc "github.com/nJannDave/pkg/pb/auth/authconnect"
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
			log.LogHSR(ctx, "internal server error", "register", req.Spec().Procedure, err.Error())
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
			log.LogHSR(ctx, "internal server error", "login", req.Spec().Procedure, err.Error())
		}
		return nil, errorr
	}
	utils.SetCookie(ctx, string(constt.AT), tkn.Access, time.Now().Add(3*time.Minute))
	utils.SetCookie(ctx, string(constt.RF), tkn.Refresh, time.Now().Add(24*3*time.Hour))
	return connect.NewResponse(&pb.Response{
		Status: true,
		Message: "successfully login",
	}), nil
}