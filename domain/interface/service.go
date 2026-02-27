package contract

import (
	"auth/domain/entity"
	"context"
)

type Service interface {
	Register(ctx context.Context, userData entity.UserData, residence entity.Residence, idempotencyKey string) error
}