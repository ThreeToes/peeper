package config

// BasicAuthConfig is the configuration for the BasicAuth struct
type BasicAuthConfig struct {
	Username string `toml:"username"`
	Password string `toml:"password"`
}
