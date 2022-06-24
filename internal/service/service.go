package service

import (
	"context"
	"io/ioutil"
	"net/http"
	"peeper/internal/config"
	"peeper/internal/routes"
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

	return g.routes[e.LocalPath].RegisterRoute(e.LocalMethod, func(respWriter http.ResponseWriter, req *http.Request) {
		forwardedReq, err := http.NewRequest(e.RemoteMethod, e.RemotePath, req.Body)

		client := http.DefaultClient

		resp, err := client.Do(forwardedReq)
		if err != nil {
			respWriter.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			respWriter.WriteHeader(http.StatusInternalServerError)
			return
		}
		respWriter.WriteHeader(resp.StatusCode)
		_, err = respWriter.Write(body)
	})
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
