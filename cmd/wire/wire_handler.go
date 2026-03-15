package wiree

import (
	"auth/controller/handler"

	"connectrpc.com/connect"
	"github.com/go-redis/redis_rate/v10"
	pb "github.com/nJannDave/pkg/pb/auth/authconnect"

	"net/http"

	middle "github.com/nJannDave/pkg/middleware"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func WireHandler(h *handler.Handler, rds *redis_rate.Limiter) *http.Server {
	router := gin.Default()
	h2s := &http2.Server{}

	rateLimiter := middle.NewRateLimiterInterceptor(rds)
	logging := middle.NewLoggingInterceptor()

	interceptor := connect.WithInterceptors(
		connect.UnaryInterceptorFunc(rateLimiter.WrapUnary),
		connect.UnaryInterceptorFunc(logging.WrapUnary),
	)

	path, ch := pb.NewAuthServiceHandler(h, interceptor)
	router.Any(path+"*path", gin.WrapH(ch))
	server := &http.Server{
		Addr:    ":8080",
		Handler: h2c.NewHandler(router, h2s),
	}

	return server
}