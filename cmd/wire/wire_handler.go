package wiree

import (
	"auth/controller/handler"
	pb "auth/pb/pbconnect"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func WireHandler(h *handler.Handler) *http.Server {
	router := gin.Default()
	h2s := &http2.Server{}

	path, ch := pb.NewAuthServiceHandler(h)
	router.Any(path+"*path", gin.WrapH(ch))
	server := &http.Server{
		Addr:    ":8080",
		Handler: h2c.NewHandler(router, h2s),
	}

	return server
}