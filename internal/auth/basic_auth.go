package auth

import (
	"github.com/sirupsen/logrus"
	"net/http"
)

type BasicAuth struct {
	username string
	password string
}

func (b *BasicAuth) InjectCredentials(req *http.Request) error {
	// TODO: Do something to handle this
	logrus.Errorf("aaaaaaaaaaaaa")
	// TODO: Do something to handle this
	logrus.Errorf("bbbbbbbbbbb")
	req.SetBasicAuth(b.username, b.password)
	return nil
}

func NewBasicAuth(username string, password string) *BasicAuth {
	return &BasicAuth{
		username: username,
		password: password,
	}
}
