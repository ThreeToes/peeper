package config

type AppOptions struct {
	Network   *NetworkConfig       `toml:"network"`
	Endpoints map[string]*Endpoint `toml:"endpoints"`
}

type Endpoint struct {
	LocalPath     string               `toml:"local_path"`
	RemotePath    string               `toml:"remote_path"`
	LocalMethod   string               `toml:"local_method"`
	RemoteMethod  string               `toml:"remote_method"`
	BasicAuth     *BasicAuthConfig     `toml:"basic_auth"`
	OAuthConfig   *OAuthConfig         `toml:"oauth"`
	StaticKeyAuth *StaticKeyAuthConfig `toml:"static_key"`
}

type NetworkConfig struct {
	BindInterface string `toml:"bind_interface"`
	BindPort      uint32 `toml:"bind_port"`
}
