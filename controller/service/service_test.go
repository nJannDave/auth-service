package service

import (
	"auth/controller/token"
	_ "auth/domain/entity"
	"auth/mocks"
	"context"
	"errors"
	"os"
	"testing"

	"github.com/go-openapi/testify/v2/require"
	"github.com/joho/godotenv"
	"github.com/nJannDave/pkg/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type errs struct {
	isErr bool
	message string
}

type generalInstructions struct {
	name string
	mocksFunc func()
	err errs
}

func TestMain(m *testing.M) {
	zapLog := log.InitLog(); defer zapLog.Sync()
	if err := godotenv.Load("../../.env"); err != nil { log.ZapLog.Sugar().Fatalf("failed test: error: %v", err) }
	if err := token.Init(); err != nil { log.ZapLog.Sugar().Fatalf("failed test: error: %v", err) }
	os.Exit(m.Run())
}

func TestRefresh(t *testing.T) {
	// --- Arrange ---
	const userId = 10
	tkn, err := token.GenerateToken(userId, "user")
	if err != nil { log.ZapLog.Sugar().Fatalf("failed test: error: %v", err) }

	ctx := context.Background()
	mocksDb := mocks.NewRepo(t)
	svc := InitService(mocksDb)

	var instructions = []generalInstructions {
		{
			name: "success refresh",
			mocksFunc: func() {
				mocksDb.On("RdsGet", ctx, mock.Anything, "refresh token").Return(nil,nil).Once()
				mocksDb.On("RdsTX", ctx, mock.Anything).Return(func(ctx context.Context, fn func(context.Context) error) error {
					return fn(ctx)
				}).Once()
				mocksDb.On("RdsSet", ctx, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				mocksDb.On("RdsDel", ctx, mock.Anything).Return(nil).Once()
			},
			err: errs{
				isErr: false,
				message: "",
			},
		},
		{
			name: "failed refresh - token not founds",
			mocksFunc: func() {
				mocksDb.On("RdsGet", ctx, mock.Anything, "refresh token").Return(nil, errors.New("refresh token not found")).Once()
			},
			err: errs{
				isErr: true,
				message: "please login",
			},
		},
		{
	        name: "failed refresh - transaction error",
	        mocksFunc: func() {
	            mocksDb.On("RdsGet", ctx, mock.Anything, "refresh token").Return(nil, nil).Once()
	            mocksDb.On("RdsTX", ctx, mock.Anything).Return(func(ctx context.Context, fn func(context.Context) error) error {
	                return errors.New("internal-server-error:error")
	            }).Once()
	        },
	        err: errs{isErr: true, message: "internal-server-error:error"},
	    },
	    {
	        name: "failed refresh - set new token error",
	        mocksFunc: func() {
	            mocksDb.On("RdsGet", ctx, mock.Anything, "refresh token").Return(nil, nil).Once()
	            mocksDb.On("RdsTX", ctx, mock.Anything).Return(func(ctx context.Context, fn func(context.Context) error) error {
	                return fn(ctx)
	            }).Once()
	            mocksDb.On("RdsSet", ctx, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("internal-server-error:redis set:error")).Once()
	        },
	        err: errs{isErr: true, message: "internal-server-error:redis set:error"},
	    },
	    {
	        name: "failed refresh - delete old token error",
	        mocksFunc: func() {
	            mocksDb.On("RdsGet", ctx, mock.Anything, "refresh token").Return(nil, nil).Once()
	            mocksDb.On("RdsTX", ctx, mock.Anything).Return(func(ctx context.Context, fn func(context.Context) error) error {
	                return fn(ctx)
	            }).Once()
	            mocksDb.On("RdsSet", ctx, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	            mocksDb.On("RdsDel", ctx, mock.Anything).Return(errors.New("internal-server-error:redis del:error")).Once()
	        },
	        err: errs{isErr: true, message: "internal-server-error:redis del:error"},
	    },
	}
	// --- Act & Assert ---
	for _, tc := range instructions {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			tc.mocksFunc()
			_, err := svc.Refresh(ctx, tkn.Refresh)
			if tc.err.isErr {
				require.Error(t, err)
				assert.ErrorContains(t, err, tc.err.message)
			} else {
				require.NoError(t, err)
			}
		}) 
	}
}

func TestLogout(t *testing.T) {
	// --- Arrange ---
	const userId = 10
	tkn, err := token.GenerateToken(userId, "user")
	if err != nil { log.ZapLog.Sugar().Fatalf("failed test: error: %v", err) }
	ctx := context.Background()
	mocksDb := mocks.NewRepo(t)
	svc := InitService(mocksDb)
	instructions := []generalInstructions{
		{
			name: "succes logout",
			mocksFunc: func() {
				mocksDb.On("GetId", ctx, userId).Return(userId, nil).Once()
				mocksDb.On("RdsTX", ctx, mock.Anything).Return(func(ctx context.Context, fn func(context.Context) error) error {
					return fn(ctx)
				}).Once()				
				mocksDb.On("RdsSet", ctx, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				mocksDb.On("RdsDel", ctx, mock.Anything).Return(nil).Once()
			},
			err: errs{
				isErr: false,
				message: "",
			},
		},
		{
			name: "failed logout - id not found",
			mocksFunc: func() {
				mocksDb.On("GetId", ctx, userId).Return(0, errors.New("id not found")).Once()
			},
			err: errs{
				isErr: true,
				message: "id not found",
			},
		},
		{
			name: "failed logout - internal server error",
			mocksFunc: func() {
				mocksDb.On("GetId", ctx, userId).Return(userId, nil).Once()
				mocksDb.On("RdsTX", ctx, mock.Anything).Return(func(ctx context.Context, fn func(context.Context) error) error {
					return fn(ctx)
				}).Once()
				mocksDb.On("RdsSet", ctx, mock.Anything, mock.Anything, mock.Anything).Return(errors.New("service:logout:internal server error:redis-set:error")).Once()
			},
			err: errs{
				isErr: true,
				message: "service:logout:internal server error:redis-set:error",
			},
		},
	}
	// --- Act & Assert ---
	for _, tc := range instructions {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			tc.mocksFunc()
			result := svc.Logout(ctx, tkn.Access, tkn.Refresh)

			if tc.err.isErr {
				require.Error(t, result)
				assert.ErrorContains(t, result, tc.err.message)
			} else {
				require.NoError(t, result)
			}
		})
	}
}