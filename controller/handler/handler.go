package handler

import (
	"auth/domain/entity"
	contract "auth/domain/interface"
	"auth/pb"
	"auth/pb/pbconnect"
	"errors"
	"time"
	"connectrpc.com/connect"
	"context"

	"github.com/nJannDave/pkg/log"
	utils "github.com/nJannDave/pkg/utils/handler"
	"github.com/nJannDave/pkg/const"
)

type Handler struct {
	proto pbconnect.UnimplementedAuthServiceHandler
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
	userData := entity.NewUserData(req.Msg.Nik, req.Msg.Name)
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