package service

import (
	"context"
	"github.com/threetoes/peeper/internal/auth"
	"github.com/threetoes/peeper/internal/config"
	"github.com/threetoes/peeper/internal/routes"
	"net/http"
)

type Service interface {
	RegisterEndpoint(e *config.Endpoint) error
	Start() error
	Stop() error
}

type NormalService struct {
	routes  map[string]*routes.Router
	mux     *http.ServeMux
	httpSrv *http.Server
}

func (g *NormalService) RegisterEndpoint(e *config.Endpoint) error {
	if _, ok := g.routes[e.LocalPath]; !ok {
		router := routes.NewRouter()
		g.routes[e.LocalPath] = router
		g.mux.HandleFunc(e.LocalPath, router.ServeHTTP)
	}
	if e.BasicAuth != nil && e.BasicAuth.Username != "" {
		err := g.routes[e.LocalPath].RegisterCredentials(e.LocalMethod, auth.NewBasicAuth(e.BasicAuth.Username, e.BasicAuth.Password))
		if err != nil {
			return err
		}
	} else if e.OAuthConfig != nil {
		conf := e.OAuthConfig
		injector := auth.NewOAuthInjector(conf.TokenEndpoint, conf.ClientId, conf.ClientSecret, conf.ExtraFormValues)
		err := g.routes[e.LocalPath].RegisterCredentials(e.LocalMethod, injector)
		if err != nil {
			return err
		}
	} else if e.StaticKeyAuth != nil {
		conf := e.StaticKeyAuth
		injector := auth.NewStaticKeyInjector(conf.Headers)
		if err := g.routes[e.LocalPath].RegisterCredentials(e.LocalMethod, injector); err != nil {
			return err
		}
	}
	return g.routes[e.LocalPath].RegisterRoute(e.LocalMethod, e.RemotePath, e.RemoteMethod)
}

func (g *NormalService) Start() error {
	return g.httpSrv.ListenAndServe()
}

func (g *NormalService) Stop() error {
	return g.httpSrv.Shutdown(context.Background())
}

func New(addr string) Service {
	mux := http.NewServeMux()
	g := &NormalService{
		mux:    mux,
		routes: map[string]*routes.Router{},
		httpSrv: &http.Server{
			Addr:    addr,
			Handler: mux,
		},
	}

	return g
}
