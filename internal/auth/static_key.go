package auth

import (
	"net/http"
)

type StaticKeyInjector struct {
	headers map[string]string
}

func (s *StaticKeyInjector) InjectCredentials(req *http.Request) error {
	for k, v := range s.headers {
		req.Header.Set(k, v)
	}
	return nil
}

// NewStaticKeyInjector will return a pointer to a StaticKeyInjector. The map headers is a collection of key-value
// pairs, where the key is the header name and the value is what it should be set to
func NewStaticKeyInjector(headers map[string]string) *StaticKeyInjector {
	return &StaticKeyInjector{
		headers: headers,
	}
}
