//go:generate go run github.com/golang/mock/mockgen@v1.6 -source=./credential_injector.go -destination=../mocks/auth/credential_injector.go
package auth

import "net/http"

type CredentialInjector interface {
	InjectCredentials(req *http.Request) error
}
