package auth

import (
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"testing"
)

func TestBasicAuth_InjectCredentials(t *testing.T) {
	b := NewBasicAuth("bigbos_1964", "sn@ke3ateR")
	req := httptest.NewRequest("POST", "/test", nil)
	err := b.InjectCredentials(req)
	assert.NoError(t, err)
	actualUsername, actualPassword, ok := req.BasicAuth()
	assert.True(t, ok)
	assert.Equal(t, b.username, actualUsername)
	assert.Equal(t, b.password, actualPassword)
}
