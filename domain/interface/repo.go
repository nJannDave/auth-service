package contract

import (
	"auth/domain/entity"
	"context"
	"time"
)

type Repo interface {
	Setnx(ctx context.Context, key string, value string, ttl time.Duration) error
	RdsTX(ctx context.Context, fn func() error) error
	RdsSet(ctx context.Context, key string, value string, ttl time.Duration) error
	RdsGet(ctx context.Context, key string, name string) (any, error)
	RdsDel(ctx context.Context, key string) error

	Transactions(ctx context.Context, fn func(context.Context) error ) error

	GetNIK(ctx context.Context, nik string) (int, error)
	GetProvince(ctx context.Context, province string) (int, error)
	GetCity(ctx context.Context, city string) (int, error)
	AddAccount(ctx context.Context, userData entity.UserData) (int, error)
	AddJunction(ctx context.Context, data entity.JunctionData) error

	GetPassword(ctx context.Context, id int, nik string) (string, error)
}