package service

import (
	"github.com/stretchr/testify/assert"
	"github.com/threetoes/peeper/internal/config"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRegisterAndServe(t *testing.T) {
	svc := New(":9090")
	svc.RegisterEndpoint(&config.Endpoint{
		LocalPath:    "/testpath",
		RemotePath:   "http://localhost:9091/forwarded",
		LocalMethod:  "GET",
		RemoteMethod: "POST",
	})
	testSvc := httptest.NewUnstartedServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if !assert.Equal(t, http.MethodPost, request.Method) {
			writer.WriteHeader(http.StatusBadRequest)
			writer.Write([]byte("bad method in request"))
			return
		}
		if !assert.Equal(t, "/forwarded", request.URL.Path) {
			writer.WriteHeader(http.StatusBadRequest)
			writer.Write([]byte("bad path in request"))
			return
		}

		writer.WriteHeader(http.StatusOK)
		writer.Write([]byte("test success I guess"))
	}))

	listener, err := net.Listen("tcp", "localhost:9091")
	if err != nil {
		t.Errorf("couldn't listen: %v", err)
		return
	}
	testSvc.Listener.Close()
	testSvc.Listener = listener

	go func() {
		svc.Start()
	}()
	testSvc.Start()
	defer testSvc.Close()

	// give the server a couple of seconds to come up
	time.Sleep(2 * time.Second)
	client := &http.Client{}
	resp, err := client.Get("http://localhost:9090/testpath")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	assert.Equal(t, "test success I guess", string(body))
	svc.Stop()
}

func TestBasicAuthEndpoint(t *testing.T) {
	svc := New(":9090")
	svc.RegisterEndpoint(&config.Endpoint{
		LocalPath:    "/testpath",
		RemotePath:   "http://localhost:9091/forwarded",
		LocalMethod:  "GET",
		RemoteMethod: "POST",
		BasicAuth: &config.BasicAuthConfig{
			Username: "username1",
			Password: "passw0rd",
		},
	})
	testSvc := httptest.NewUnstartedServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		username, password, ok := request.BasicAuth()
		if !assert.True(t, ok) {
			writer.WriteHeader(http.StatusUnauthorized)
			writer.Write([]byte("basic auth returned an error"))
			return
		}
		assert.Equal(t, "username1", username)
		assert.Equal(t, "passw0rd", password)
		if !assert.Equal(t, http.MethodPost, request.Method) {
			writer.WriteHeader(http.StatusBadRequest)
			writer.Write([]byte("bad method in request"))
			return
		}
		if !assert.Equal(t, "/forwarded", request.URL.Path) {
			writer.WriteHeader(http.StatusBadRequest)
			writer.Write([]byte("bad path in request"))
			return
		}

		writer.WriteHeader(http.StatusOK)
		writer.Write([]byte("test success I guess"))
	}))

	listener, err := net.Listen("tcp", "localhost:9091")
	if err != nil {
		t.Errorf("couldn't listen: %v", err)
		return
	}
	testSvc.Listener.Close()
	testSvc.Listener = listener

	go func() {
		svc.Start()
	}()
	testSvc.Start()
	defer testSvc.Close()

	// give the server a couple of seconds to come up
	time.Sleep(2 * time.Second)
	client := &http.Client{}
	resp, err := client.Get("http://localhost:9090/testpath")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	assert.Equal(t, "test success I guess", string(body))
	svc.Stop()
}
