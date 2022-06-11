package config

type AppOptions struct {
	Endpoints map[string]Endpoint `toml:"endpoints"`
}

type Endpoint struct {
	LocalPath    string `toml:"local_path"`
	RemotePath   string `toml:"target_path"`
	LocalMethod  string `toml:"local_method"`
	RemoteMethod string `toml:"remote_method"`
}
