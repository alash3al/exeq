package config

type HTTPServerConfig struct {
	ListenAddr       string `hcl:"listen"`
	EnableAccessLogs bool   `hcl:"access_logs"`
}
