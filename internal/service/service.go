package service

import (
	"context"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"peeper/internal/config"
)

type Service interface {
	RegisterEndpoint(e *config.Endpoint) error
	Start() error
	Stop() error
}

type GinService struct {
	engine  *gin.Engine
	httpSrv *http.Server
}

func (g *GinService) RegisterEndpoint(e *config.Endpoint) error {
	g.engine.Handle(e.LocalMethod, e.LocalPath, func(ctx *gin.Context) {
		forwardedReq, err := http.NewRequest(e.RemoteMethod, e.RemotePath, ctx.Request.Body)

		client := http.DefaultClient

		resp, err := client.Do(forwardedReq)
		if err != nil {
			ctx.String(http.StatusInternalServerError, "")
			return
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			ctx.String(http.StatusInternalServerError, "")
			return
		}
		ctx.String(resp.StatusCode, string(body))
	})
	return nil
}

func (g *GinService) Start() error {
	return g.httpSrv.ListenAndServe()
}

func (g *GinService) Stop() error {
	return g.httpSrv.Shutdown(context.Background())
}

func New(addr string) Service {
	gin.SetMode(gin.ReleaseMode)
	eng := gin.New()
	eng.Use(gin.Recovery())
	g := &GinService{
		engine: eng,
		httpSrv: &http.Server{
			Addr:    addr,
			Handler: eng,
		},
	}

	return g
}
