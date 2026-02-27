package repo

import (
	utils "github.com/nJannDave/pkg/utils/repo"
	"github.com/nJannDave/pkg/const"

	"auth/domain/entity"
	"auth/domain/interface"
	"errors"
	"time"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"context"
)

type repo struct {
	db *gorm.DB
	rds *redis.Client
}

func InitRepo(db *gorm.DB, rds *redis.Client) contract.Repo {
	return &repo {
		db: db,
		rds: rds,
	}
}

func (r *repo) Setnx(ctx context.Context, key string, value string, ttl time.Duration) error {
	ok, err := r.rds.SetNX(ctx, key, value, ttl).Result()
	if err != nil {
		return utils.ValidateErrRedis(err, utils.WithFunc("redis setnx"))
	}
	if ok {
		return nil
	}
	return errors.New("data already exists")
}

func (r *repo) RdsDel(ctx context.Context, key string) error {
	if err := r.rds.Del(ctx, key).Err(); err != nil {
		return utils.ValidateErrRedis(err, utils.WithFunc("redis delete"))
	}
	return nil
} 

func (r *repo) Transactions(ctx context.Context, fn func(context.Context) error ) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		ctx = context.WithValue(ctx, constt.TX, tx)
		return fn(ctx)
	})
}

func (r *repo) GetNIK(ctx context.Context, nik string) (int, error) {
	var id int
	if err := r.db.WithContext(ctx).Table("account").Select("account_id").Where("nik = ?", nik).Limit(1).Scan(&id).Error; err != nil {
		return 0, utils.ValidateErrRepo(err, utils.WithName("nik"), utils.WithFunc("GetNIK"))
	}
	return id, nil
}

func (r *repo) GetProvince(ctx context.Context, province string) (int, error) {
	var id int
	if err := r.db.WithContext(ctx).Table("province").Select("province_id").Where("province = ?", province).Limit(1).Scan(&id).Error; err != nil {
		return 0, utils.ValidateErrRepo(err, utils.WithName("province"), utils.WithFunc("GetProvince"))
	}
	return id, nil
}

func (r *repo) GetCity(ctx context.Context, city string) (int, error) {
	var id int
	if err := r.db.WithContext(ctx).Table("city").Select("city_id").Where("city = ?", city).Limit(1).Scan(&id).Error; err != nil {
		return 0, utils.ValidateErrRepo(err, utils.WithName("city"), utils.WithFunc("GetCity"))
	}
	return id, nil
}

func (r *repo) AddAccount(ctx context.Context, userData entity.UserData) (int, error) {
	if err := r.db.WithContext(ctx).Clauses(clause.Returning{Columns: []clause.Column{{Name: "account_id"}}}).Create(&userData); err.Error != nil {
		return 0, utils.ValidateErrRepo(err.Error, utils.WithFunc("AddAccount"))
	}
	return userData.AccountId, nil
}

func (r *repo) AddJunction(ctx context.Context, data entity.JunctionData) error {
	if err := r.db.WithContext(ctx).Create(&data); err.Error != nil {
		return utils.ValidateErrRepo(err.Error, utils.WithFunc("AddJunction"))
	}
	return nil
}