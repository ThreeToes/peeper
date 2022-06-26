package auth

// Referred to here for implementation https://datatracker.ietf.org/doc/html/rfc6749#section-4.4

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type token struct {
	AccessToken string `json:"access_token"`
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
}

// OAuthM2MCredentialInjector injects bearer tokens into the forwarded request. Only supports the client_credentials
// workflow
type OAuthM2MCredentialInjector struct {
	clientId        string
	clientSecret    string
	tokenEndpoint   string
	extraFormValues map[string]string
}

func (o *OAuthM2MCredentialInjector) InjectCredentials(req *http.Request) error {
	if tok, err := o.getToken(o.extraFormValues); err != nil {
		return err
	} else {
		switch tok.TokenType {
		case "Bearer":
			req.Header.Set("authorization", fmt.Sprintf("Bearer %s", tok.AccessToken))
		default:
			return fmt.Errorf("unknown token type '%s'", tok.TokenType)
		}
	}
	return nil
}

func (o *OAuthM2MCredentialInjector) getToken(extraFormValues map[string]string) (*token, error) {
	client := http.DefaultClient
	req, err := http.NewRequest(http.MethodPost, o.tokenEndpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(o.clientId, o.clientSecret)
	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("client_id", o.clientId)
	for k, v := range extraFormValues {
		form.Set(k, v)
	}
	req.Body = ioutil.NopCloser(strings.NewReader(form.Encode()))
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("received status code %d instead of 200", resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var tok token
	err = json.Unmarshal(body, &tok)
	if err != nil {
		return nil, err
	}

	return &tok, nil
}

func NewOAuthInjector(tokenEndpoint, clientId, clientSecret string, extraFormValues map[string]string) *OAuthM2MCredentialInjector {
	return &OAuthM2MCredentialInjector{
		clientId:        clientId,
		clientSecret:    clientSecret,
		tokenEndpoint:   tokenEndpoint,
		extraFormValues: extraFormValues,
	}
}
