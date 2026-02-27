package inrd

import (
	"github.com/nJannDave/pkg/log"

	"context"
	"os"

	"github.com/redis/go-redis/v9"
)

func ProviderCTX() context.Context {
	return context.TODO()
}

func Init(ctx context.Context) (*redis.Client, func()) {
	rds := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("RA"),
		Password: os.Getenv("RP"),
		DB:       0,
	})
	if err := rds.Ping(ctx).Err(); err != nil {
		log.LogConfig("failed connect to redis", "connect_redis", err)
		panic(err)
	}
	return rds, func () {
		if err := rds.Close(); err != nil {
			panic(err)
		}
	}
}