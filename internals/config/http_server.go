package config

type HTTPServerConfig struct {
	ListenAddr string `hcl:"listen"`
}
