package routes

import (
	"net/http"
	"net/http/httptest"
	"reflect"
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
			want: &Router{methodHandlers: map[string]func(w http.ResponseWriter, request *http.Request){}},
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
	handler := func(w http.ResponseWriter, req *http.Request) {}
	type fields struct {
		methodHandlers map[string]func(w http.ResponseWriter, request *http.Request)
	}
	type args struct {
		method  string
		handler func(w http.ResponseWriter, req *http.Request)
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
			},
			args: args{
				method:  "GET",
				handler: handler,
			},
			wantErr: false,
		},
		{
			name: "duplicate route",
			fields: fields{
				methodHandlers: map[string]func(w http.ResponseWriter, request *http.Request){
					"GET": handler,
				},
			},
			args: args{
				method:  "GET",
				handler: handler,
			},
			wantErr: true,
		},
		{
			name: "route with different method",
			fields: fields{
				methodHandlers: map[string]func(w http.ResponseWriter, request *http.Request){
					"POST": handler,
				},
			},
			args: args{
				method:  "GET",
				handler: handler,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Router{
				methodHandlers: tt.fields.methodHandlers,
			}
			if err := r.RegisterRoute(tt.args.method, tt.args.handler); (err != nil) != tt.wantErr {
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
