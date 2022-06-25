// go:generate go run github.com/golang/mock/mockgen
package auth

import "net/http"

type CredentialInjector interface {
	InjectCredentials(req *http.Request) error
}
