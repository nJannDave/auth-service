package service

import (
	utils "github.com/nJannDave/pkg/utils/service"
	"github.com/nJannDave/pkg/const"

	"auth/controller/hash"
	"auth/controller/token"
	"auth/domain/entity"
	contract "auth/domain/interface"

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

func (s *service) Login(ctx context.Context, loginData entity.UserData) (*token.Token,error) {
	var service = "login"
	const role = "public"
	id, err := s.repo.GetNIK(ctx, loginData.NIK)
	if err != nil {
		return nil, utils.ValidateErrService(err, utils.WithService(service))
	}
	if id == 0 { return nil, errors.New("account not found") }
	pw, err := s.repo.GetPassword(ctx, id, loginData.NIK) 
	if err != nil {
		return nil, utils.ValidateErrService(err, utils.WithService(service))
	}
	if err := hash.UnHashPassword(loginData.Password, pw); err != nil {
		return nil, err
	}
	tkn, err := token.GenerateToken(id, role)
	if err != nil {
		return nil, err
	}
	if err := s.repo.RdsSet(ctx, string(constt.RF), tkn.Refresh, 24*3*time.Hour); err != nil {
		return nil, utils.ValidateErrService(err, utils.WithService(service))
	}
	return tkn, nil
}