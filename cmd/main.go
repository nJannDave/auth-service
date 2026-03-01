package main

import (
	"github.com/joho/godotenv"
	"github.com/nJannDave/pkg/log"

	"auth/cmd/wire"
	"auth/controller/token"
	"os/signal"
	"syscall"

	"context"
	"net/http"
	"os"
	"time"
)


func main() {
	zapLog := log.InitLog()
	defer zapLog.Sync()
	if err := godotenv.Load(".env"); err != nil {
		zapLog.Error("failed open .env file")
		return
	}

	if err := token.Init(); err != nil {
		zapLog.Error("failed search privat and pub key")
		return
	}

	handler, cleanup, err := wiree.InitializeApp()
	if err != nil {
		zapLog.Sugar().Fatalf("error while initialize app: %v", err)
	}
	srv := wiree.WireHandler(handler)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed  {
			zapLog.Sugar().Fatalf("error while start server: %v", err)
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<- quit
	ctx, cancel := context.WithTimeout(context.TODO(), 2*time.Minute)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		zapLog.Sugar().Fatalf("error while shutdown server: %v", err)
	}
	cleanup()
}