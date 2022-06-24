package routes

import (
	"fmt"
	"net/http"
)

type Router struct {
	methodHandlers map[string]func(w http.ResponseWriter, request *http.Request)
}

func (r *Router) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()
	if handlerFunc, ok := r.methodHandlers[request.Method]; ok {
		handlerFunc(writer, request)
	} else {
		writer.WriteHeader(http.StatusNotFound)
	}
}

func (r *Router) RegisterRoute(method string, handler func(w http.ResponseWriter, req *http.Request)) error {
	if _, ok := r.methodHandlers[method]; ok {
		return fmt.Errorf("could not register another handler for method '%s'", method)
	}
	r.methodHandlers[method] = handler
	return nil
}

func NewRouter() *Router {
	return &Router{
		methodHandlers: map[string]func(w http.ResponseWriter, request *http.Request){},
	}
}
