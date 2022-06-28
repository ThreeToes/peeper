package auth

import (
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"testing"
)

func TestStaticKeyInjector_InjectCredentials(t *testing.T) {
	headers := map[string]string{
		"x-api-key": "test",
	}
	keyInjector := NewStaticKeyInjector(headers)
	req := httptest.NewRequest("POST", "/test", nil)
	err := keyInjector.InjectCredentials(req)
	assert.NoError(t, err)
	assert.Equal(t, "test", req.Header.Get("x-api-key"))
}
