// go:generate
package auth

import "net/http"

type CredentialInjector interface {
	InjectCredentials(req *http.Request) error
}
