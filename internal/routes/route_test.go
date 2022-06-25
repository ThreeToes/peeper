package routes

import (
	"fmt"
	"github.com/threetoes/peeper/internal/auth"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRouter(t *testing.T) {
	tests := []struct {
		name string
		want *Router
	}{
		{
			name: "create new",
			want: &Router{
				methodHandlers: map[string]func(w http.ResponseWriter, request *http.Request){},
				credentials:    map[string]auth.CredentialInjector{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewRouter(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewRouter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRouter_RegisterRoute(t *testing.T) {
	type fields struct {
		methodHandlers map[string]func(w http.ResponseWriter, request *http.Request)
		credentials    map[string]auth.CredentialInjector
	}
	type args struct {
		localMethod  string
		remotePath   string
		remoteMethod string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "add success",
			fields: fields{
				methodHandlers: map[string]func(w http.ResponseWriter, request *http.Request){},
				credentials:    map[string]auth.CredentialInjector{},
			},
			args: args{
				localMethod:  "GET",
				remotePath:   "/",
				remoteMethod: "GET",
			},
			wantErr: false,
		},
		{
			name: "duplicate route",
			fields: fields{
				methodHandlers: map[string]func(w http.ResponseWriter, request *http.Request){
					"GET": func(w http.ResponseWriter, request *http.Request) {},
				},
				credentials: map[string]auth.CredentialInjector{},
			},
			args: args{
				localMethod:  "GET",
				remotePath:   "/",
				remoteMethod: "GET",
			},
			wantErr: true,
		},
		{
			name: "route with different method",
			fields: fields{
				methodHandlers: map[string]func(w http.ResponseWriter, request *http.Request){
					"POST": func(w http.ResponseWriter, request *http.Request) {},
				},
				credentials: map[string]auth.CredentialInjector{},
			},
			args: args{
				localMethod:  "GET",
				remotePath:   "/",
				remoteMethod: "GET",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Router{
				methodHandlers: tt.fields.methodHandlers,
				credentials:    tt.fields.credentials,
			}
			if err := r.RegisterRoute(tt.args.localMethod, tt.args.remotePath, tt.args.remoteMethod); (err != nil) != tt.wantErr {
				t.Errorf("RegisterRoute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRouter_ServeHTTP(t *testing.T) {
	getCalled := 0
	postCalled := 0
	getHandler := func(w http.ResponseWriter, request *http.Request) {
		getCalled++
	}
	postHandler := func(w http.ResponseWriter, request *http.Request) {
		postCalled++
	}
	type fields struct {
		methodHandlers map[string]func(w http.ResponseWriter, request *http.Request)
	}
	type args struct {
		writer  http.ResponseWriter
		request *http.Request
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		expectedGet  int
		expectedPost int
	}{
		{
			name: "call get",
			fields: fields{
				methodHandlers: map[string]func(w http.ResponseWriter, request *http.Request){
					"GET":  getHandler,
					"POST": postHandler,
				},
			},
			args: args{
				writer:  httptest.NewRecorder(),
				request: httptest.NewRequest("GET", "/", nil),
			},
			expectedGet:  1,
			expectedPost: 0,
		},
		{
			name: "call post",
			fields: fields{
				methodHandlers: map[string]func(w http.ResponseWriter, request *http.Request){
					"GET":  getHandler,
					"POST": postHandler,
				},
			},
			args: args{
				writer:  httptest.NewRecorder(),
				request: httptest.NewRequest("POST", "/", nil),
			},
			expectedGet:  0,
			expectedPost: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getCalled = 0
			postCalled = 0
			r := &Router{
				methodHandlers: tt.fields.methodHandlers,
			}
			r.ServeHTTP(tt.args.writer, tt.args.request)
			assert.Equal(t, tt.expectedGet, getCalled)
			assert.Equal(t, tt.expectedPost, postCalled)
		})
	}
}

func TestRouter_RegisterCredentials(t *testing.T) {
	type fields struct {
		methodHandlers map[string]func(w http.ResponseWriter, request *http.Request)
		credentials    map[string]auth.CredentialInjector
	}
	type args struct {
		method   string
		injector auth.CredentialInjector
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "normal",
			fields: fields{
				methodHandlers: map[string]func(w http.ResponseWriter, request *http.Request){},
				credentials:    map[string]auth.CredentialInjector{},
			},
			args: args{
				method:   "GET",
				injector: &auth.BasicAuth{},
			},
			wantErr: assert.NoError,
		},
		{
			name: "duplicate",
			fields: fields{
				methodHandlers: map[string]func(w http.ResponseWriter, request *http.Request){},
				credentials: map[string]auth.CredentialInjector{
					"GET": &auth.BasicAuth{},
				},
			},
			args: args{
				method:   "GET",
				injector: &auth.BasicAuth{},
			},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Router{
				methodHandlers: tt.fields.methodHandlers,
				credentials:    tt.fields.credentials,
			}
			tt.wantErr(t, r.RegisterCredentials(tt.args.method, tt.args.injector), fmt.Sprintf("RegisterCredentials(%v, %v)", tt.args.method, tt.args.injector))
		})
	}
}

func TestRegisteredRoutes(t *testing.T) {
	route := NewRouter()
	err := route.RegisterRoute(http.MethodPost, "http://localhost:9091/test", http.MethodPost)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	route.RegisterCredentials(http.MethodPost, auth.NewBasicAuth("username", "password"))
	testSvc := httptest.NewUnstartedServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if !assert.Equal(t, http.MethodPost, request.Method) {
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
		username, password, ok := request.BasicAuth()
		if !assert.True(t, ok) {
			writer.WriteHeader(http.StatusUnauthorized)
			return
		}
		if !assert.Equal(t, "username", username) || !assert.Equal(t, "password", password) {
			writer.WriteHeader(http.StatusForbidden)
			return
		}
		writer.WriteHeader(http.StatusOK)
		writer.Write([]byte("test success"))
	}))
	testSvc.Listener.Close()
	l, err := net.Listen("tcp", ":9091")
	if err != nil {
		t.Errorf("couldn't start listener")
		t.FailNow()
	}
	testSvc.Listener = l
	testSvc.Start()
	defer testSvc.Close()
	rw := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader("testbody"))
	route.ServeHTTP(rw, req)
	assert.Equal(t, 200, rw.Code)
	body, _ := ioutil.ReadAll(rw.Body)
	assert.Equal(t, "test success", string(body))
}
