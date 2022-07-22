package config

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestClient(t *testing.T) {
	t.Run("no configs", func(t *testing.T) {
		httpConfig := &HttpConfig{}
		client, err := httpConfig.Client()
		assert.NoError(t, err)
		assert.NotNil(t, client)
	})
	t.Run("proxy set", func(t *testing.T) {
		httpConfig := &HttpConfig{
			ProxyServer: "http://localhost:8080",
		}
		client, err := httpConfig.Client()
		assert.NoError(t, err)
		if !assert.NotNil(t, client) {
			t.FailNow()
		}
		transport, ok := client.Transport.(*http.Transport)
		if !assert.True(t, ok) {
			t.FailNow()
		}
		assert.NotNil(t, transport.Proxy)
	})

	t.Run("proxy set", func(t *testing.T) {
		
		httpConfig := &HttpConfig{}
		client, err := httpConfig.Client()
		assert.NoError(t, err)
		if !assert.NotNil(t, client) {
			t.FailNow()
		}
		transport, ok := client.Transport.(*http.Transport)
		if !assert.True(t, ok) {
			t.FailNow()
		}

	})
}
