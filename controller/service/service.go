package service

import (
	"strconv"

	"github.com/nJannDave/pkg/const"
	utils "github.com/nJannDave/pkg/utils/service"

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

func (s *service) Login(ctx context.Context, loginData entity.UserData) (*token.Token, error) {
	key := "token:refresh:"
	var service = "login"
	const role = "public"
	id, err := s.repo.GetNIK(ctx, loginData.NIK)
	if err != nil {
		return nil, utils.ValidateErrService(err, utils.WithService(service))
	}
	key += strconv.Itoa(id)
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
	key = key + ":" + tkn.Refresh
	if err := s.repo.RdsSet(ctx, key, tkn.Refresh, 24*3*time.Hour); err != nil {
		return nil, utils.ValidateErrService(err, utils.WithService(service))
	}
	return tkn, nil
}

func (s *service) Refresh(ctx context.Context, tokenRefresh string) (*token.Token, error) {
	var (
		key = "token:refresh:id" 
		service = "refresh"
	)
	const role = "public"
	pbk, err := token.GetPublicKey()
	if err != nil {
		return nil, utils.ValidateErrService(err, utils.WithService(service))
	}
	tkn, err := token.ValidateToken(tokenRefresh, pbk)
	if err != nil {
		return nil, utils.ValidateErrService(err, utils.WithService(service))
	}
	key += strconv.Itoa(tkn.ID)
	key = key + ":" + tokenRefresh
	if _, err := s.repo.RdsGet(ctx, key, "refresh token"); err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, errors.New("please login")
		}
		return nil, utils.ValidateErrService(err, utils.WithService(service))
	}
	if tknNew, err := token.GenerateToken(tkn.ID, role); err != nil {
		return nil, utils.ValidateErrService(err, utils.WithService(service))
	} else {
		if err := s.repo.RdsTX(ctx, func() error {
			keySet := "token:refresh:id" + strconv.Itoa(tkn.ID) + ":" + tknNew.Refresh
			if err := s.repo.RdsSet(ctx, keySet, tknNew.Refresh, 24*3*time.Hour); err != nil {
				return utils.ValidateErrService(err, utils.WithService(service))
			}
			if err := s.repo.RdsDel(ctx, key); err != nil {
				return utils.ValidateErrService(err, utils.WithService(service))
			}
			return nil
		}); err != nil {
			return nil, err
		}
		return tknNew, nil
	}
}

func (s *service) Logout(ctx context.Context, id int, accessTkn string, refreshTkn string) error {
	var (
		key = "blacklist:token:access:" + accessTkn + "id:" + strconv.Itoa(id)
		service = "logout"
		keyDel = "token:refresh:id" + strconv.Itoa(id) + ":" + refreshTkn
	)
	return s.repo.RdsTX(ctx,
		func() error {
			if err := s.repo.RdsSet(ctx, key, accessTkn, 3*time.Minute); err != nil {
				return utils.ValidateErrService(err, utils.WithService(service))
			}
			if err := s.repo.RdsDel(ctx, keyDel); err != nil {
				return utils.ValidateErrService(err, utils.WithService(service))
			}
			return nil
		},
	)
} 