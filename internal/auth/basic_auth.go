package auth

import (
	"net/http"
)

type BasicAuth struct {
	username string
	password string
}

func (b *BasicAuth) InjectCredentials(req *http.Request) error {
	req.SetBasicAuth(b.username, b.password)
	return nil
}

func NewBasicAuth(username string, password string) *BasicAuth {
	return &BasicAuth{
		username: username,
		password: password,
	}
}
