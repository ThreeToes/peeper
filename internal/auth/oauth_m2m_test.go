package auth

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOAuthM2MCredentialInjector_InjectCredentials(t *testing.T) {
	t.Run("get token success", func(t *testing.T) {
		o := &OAuthM2MCredentialInjector{
			clientId:     "fakeId",
			clientSecret: "fakeSecret",
			oauthAddress: "http://localhost:9092/oauth",
			extraFormValues: map[string]string{
				"test-extra": "extra value",
			},
		}

		svc := httptest.NewUnstartedServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			err := req.ParseForm()
			if !assert.NoError(t, err) {
				rw.WriteHeader(400)
				rw.Write([]byte("form didn't parse"))
				return
			}
			if !assert.Equal(t, "fakeId", req.Form.Get("client_id")) {
				rw.WriteHeader(401)
				rw.Write([]byte("bad client_id"))
				return
			}
			if !assert.Equal(t, "client_credentials", req.Form.Get("grant_type")) {
				rw.WriteHeader(401)
				rw.Write([]byte("bad grant_type"))
				return
			}
			if !assert.Equal(t, "extra value", req.Form.Get("test-extra")) {
				rw.WriteHeader(400)
				rw.Write([]byte("extra form value not written correctly"))
				return
			}

			username, password, ok := req.BasicAuth()
			if !assert.True(t, ok) {
				rw.WriteHeader(401)
				rw.Write([]byte("basic auth not ok"))
				return
			}
			if !assert.Equal(t, "fakeId", username) {
				rw.WriteHeader(401)
				rw.Write([]byte("bad username"))
				return
			}
			if !assert.Equal(t, "fakeSecret", password) {
				rw.WriteHeader(401)
				rw.Write([]byte("bad password"))
				return
			}

			tok, err := json.Marshal(&token{
				AccessToken: "fakeaccesstoken",
				Scope:       "read:fake",
				TokenType:   "Bearer",
			})
			if !assert.NoError(t, err) {
				rw.WriteHeader(500)
				rw.Write([]byte("could not marshal token"))
				return
			}
			rw.WriteHeader(200)
			rw.Write(tok)
		}))

		l, err := net.Listen("tcp", ":9092")
		if !assert.NoError(t, err) {
			t.Errorf("Could not listen on port: %v", err)
			return
		}
		svc.Listener.Close()
		svc.Listener = l
		svc.Start()
		defer svc.Close()

		req, err := http.NewRequest(http.MethodPost, "http://test.com", nil)
		if !assert.NoError(t, err) {
			t.Errorf("could not build request object: %v", err)
		}
		err = o.InjectCredentials(req)
		assert.NoError(t, err)
		assert.Equal(t, "Bearer fakeaccesstoken", req.Header.Get("authorization"))
	})
	t.Run("get token failure", func(t *testing.T) {
		o := &OAuthM2MCredentialInjector{
			clientId:     "fakeId",
			clientSecret: "fakeSecret",
			oauthAddress: "http://localhost:9092/oauth",
			extraFormValues: map[string]string{
				"test-extra": "extra value",
			},
		}

		svc := httptest.NewUnstartedServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			rw.WriteHeader(401)
			rw.Write([]byte("no bots allowed"))
		}))

		l, err := net.Listen("tcp", ":9092")
		if !assert.NoError(t, err) {
			t.Errorf("Could not listen on port: %v", err)
			return
		}
		svc.Listener.Close()
		svc.Listener = l
		svc.Start()
		defer svc.Close()

		req, err := http.NewRequest(http.MethodPost, "http://test.com", nil)
		if !assert.NoError(t, err) {
			t.Errorf("could not build request object: %v", err)
		}
		err = o.InjectCredentials(req)
		assert.Error(t, err)
	})
}
