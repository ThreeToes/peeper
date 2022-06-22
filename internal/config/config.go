package config

type AppOptions struct {
	Network   *NetworkConfig       `toml:"network"`
	Endpoints map[string]*Endpoint `toml:"routes"`
}

type Endpoint struct {
	LocalPath    string `toml:"local_path"`
	RemotePath   string `toml:"target_path"`
	LocalMethod  string `toml:"local_method"`
	RemoteMethod string `toml:"remote_method"`
}

type NetworkConfig struct {
	BindInterface string `toml:"bind_interface"`
	BindPort      uint32 `toml:"bind_port"`
}
