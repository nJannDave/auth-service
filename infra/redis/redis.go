package inrd

import (
	"github.com/nJannDave/pkg/log"
	"github.com/nJannDave/pkg/log/structure"
	"go.uber.org/zap"

	"context"
	"os"

	"github.com/redis/go-redis/v9"
)

func ProviderCTX() context.Context {
	return context.TODO()
}

func Init(ctx context.Context) (*redis.Client, func()) {
	var logcfg = structure.LogConfig {
		Status: false,
		Service: "connect redis",
		Error: "",
	}
	rds := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("RA"),
		Password: os.Getenv("RP"),
		DB:       0,
	})
	if err := rds.Ping(ctx).Err(); err != nil {
		logcfg.Error = err.Error()
		log.ZapLog.Sugar().Fatalf("failed connect redis: %v", err)
	}
	return rds, func () {
		if err := rds.Close(); err != nil {
			logcfg.Error = err.Error()
			log.ZapLog.Fatal("failed stop redis", zap.Object("log_config: ", logcfg))
		}
	}
}