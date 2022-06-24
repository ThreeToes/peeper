package service

import (
	"context"
	"io/ioutil"
	"net/http"
	"peeper/internal/config"
)

type Service interface {
	RegisterEndpoint(e *config.Endpoint) error
	Start() error
	Stop() error
}

type NormalService struct {
	mux     *http.ServeMux
	httpSrv *http.Server
}

func (g *NormalService) RegisterEndpoint(e *config.Endpoint) error {
	g.mux.HandleFunc(e.LocalPath, func(respWriter http.ResponseWriter, req *http.Request) {
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
	return nil
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
		mux: mux,
		httpSrv: &http.Server{
			Addr:    addr,
			Handler: mux,
		},
	}

	return g
}
