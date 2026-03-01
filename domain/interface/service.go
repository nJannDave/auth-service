package contract

import (
	"auth/domain/entity"
	"auth/controller/token"
	"context"
)

type Service interface {
	Register(ctx context.Context, userData entity.UserData, residence entity.Residence, idempotencyKey string) error
	Login(ctx context.Context, loginData entity.UserData) (*token.Token,error)
}