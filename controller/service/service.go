package service

import (
	utils "github.com/nJannDave/pkg/utils/service"
	// "github.com/nJannDave/pkg/const"

	"auth/domain/entity"
	contract "auth/domain/interface"
	"auth/controller/hash"

	"errors"
	// "fmt"
	"strings"
	"time"

	"context"
)

type service struct {
	repo contract.Repo
}

func InitService(repo contract.Repo) contract.Service {
	return &service{repo: repo}
}

func (s *service) Register(ctx context.Context, userData entity.UserData, residence entity.Residence, idempotencyKey string) error {
	var (
		service = "service-register"
		nikKey = "nik:" + userData.NIK
		msg = "data already exists"
		IsErr = false
	)
	defer func() {
		if IsErr {
			s.repo.RdsDel(ctx, idempotencyKey)
			s.repo.RdsDel(ctx, nikKey)
		}
	}()
	if err := s.repo.Setnx(ctx, idempotencyKey, idempotencyKey, 24*time.Hour); err != nil {
		IsErr = true
		if strings.Contains(err.Error(), msg) {
			return errors.New("duplicate request")
		}
		return utils.ValidateErrService(err, utils.WithService(service))
	}
	if err := s.repo.Setnx(ctx, nikKey, userData.NIK, 24*time.Hour); err != nil {
		IsErr = true
		if strings.Contains(err.Error(), msg) {
			return errors.New("nik already registered")
		}
		return utils.ValidateErrService(err, utils.WithService(service))	
	}
	if err := userData.ValidateNik(); err != nil {
		IsErr = true
		return err
	}
	hash, err := hash.HashPassword(userData.Password)
	if err != nil {
		return utils.ValidateErrService(err, utils.WithService(service))
	}
	userData.Password = hash
	id, err := s.repo.GetNIK(ctx, userData.NIK)
	if err != nil {
		IsErr = true
		if !strings.Contains(err.Error(), "not found") {
			return utils.ValidateErrService(err, utils.WithService(service))
		}
	}
	if id != 0 { return errors.New("duplicate nik. you already registered") }
	provinceId, err := s.repo.GetProvince(ctx, residence.Province)
	if err != nil {
		IsErr = true
		return utils.ValidateErrService(err, utils.WithService(service))
	}
	cityId, err := s.repo.GetCity(ctx, residence.City)
	if err != nil {
		IsErr = true
		return utils.ValidateErrService(err, utils.WithService(service))
	}	
	return s.repo.Transactions(ctx, func(ctx context.Context) error {
		accountId, err := s.repo.AddAccount(ctx, userData)
		if err != nil {
			IsErr = true
			return utils.ValidateErrService(err, utils.WithService(service))
		}
		junctionData := entity.NewJuctionData(provinceId, accountId, cityId)
		if err := s.repo.AddJunction(ctx, *junctionData); err != nil {
			IsErr = true
			return utils.ValidateErrService(err, utils.WithService(service))
		}
		return nil
	})
}