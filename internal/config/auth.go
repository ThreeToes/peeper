package config

// BasicAuthConfig is the configuration for the BasicAuth struct
type BasicAuthConfig struct {
	Username string `toml:"username"`
	Password string `toml:"password"`
}

// OAuthConfig configures an OAuthM2MCredentialInjector
type OAuthConfig struct {
	// ClientId is the OAuth client app ID
	ClientId        string            `toml:"client_id"`
	ClientSecret    string            `toml:"client_secret"`
	TokenEndpoint   string            `toml:"token_endpoint"`
	ExtraFormValues map[string]string `toml:"extra_form_values"`
}
