package main

import (
	_ "auth/dto"
	_ "auth/controller/handler"
	
	"github.com/joho/godotenv"
	"github.com/nJannDave/pkg/log"
	utils "github.com/nJannDave/pkg/utils"

	"auth/cmd/wire"
	"auth/controller/token"

	"net/http"
	"time"
)

// @title Auth-Service API
// @version 1.0
// @description For login, logout, refresh token and make account
// @host localhost:7070
// @securityDefinitions.apikey Refresh-Token
// @in cookie
// @name RefreshToken
// @securityDefinitions.apikey Access-Token
// @in cookie
// @name AccessToken
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

	handler, cleanup, limiter, err := wiree.InitializeApp()
	if err != nil {
		zapLog.Sugar().Fatalf("error while initialize app: %v", err)
	}
	srv := wiree.WireHandler(handler, limiter)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed  {
			zapLog.Fatal("error while start server: " + err.Error())
		}
	}()
	if err := utils.GraceFShutD(srv, 4*time.Minute); err != nil {
		zapLog.Fatal("error while shutdown server: " + err.Error())
	}
	cleanup()
}