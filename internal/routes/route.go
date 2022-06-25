package routes

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"peeper/internal/auth"
)

type Router struct {
	methodHandlers map[string]func(w http.ResponseWriter, request *http.Request)
	credentials    map[string]auth.CredentialInjector
}

func (r *Router) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()
	if handlerFunc, ok := r.methodHandlers[request.Method]; ok {
		handlerFunc(writer, request)
	} else {
		writer.WriteHeader(http.StatusNotFound)
	}
}

func (r *Router) RegisterRoute(localMethod, remotePath, remoteMethod string) error {
	if _, ok := r.methodHandlers[localMethod]; ok {
		return fmt.Errorf("could not register another handler for method '%s'", localMethod)
	}
	r.methodHandlers[localMethod] = func(rw http.ResponseWriter, req *http.Request) {
		forwardedReq, err := http.NewRequest(remoteMethod, remotePath, req.Body)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
		if credentials, ok := r.credentials[localMethod]; ok {
			err = credentials.InjectCredentials(forwardedReq)
		}

		client := http.DefaultClient

		resp, err := client.Do(forwardedReq)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
		rw.WriteHeader(resp.StatusCode)
		_, err = rw.Write(body)
	}
	return nil
}

func (r *Router) RegisterCredentials(method string, injector auth.CredentialInjector) error {
	if _, ok := r.credentials[method]; ok {
		return fmt.Errorf("method %s already has a credential injector", method)
	}
	r.credentials[method] = injector
	return nil
}

func NewRouter() *Router {
	return &Router{
		methodHandlers: map[string]func(w http.ResponseWriter, request *http.Request){},
		credentials:    map[string]auth.CredentialInjector{},
	}
}
