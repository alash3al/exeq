package config

type QueueConfig struct {
	Driver        string `hcl:"driver"`
	DSN           string `hcl:"dsn"`
	WorkersCount  int    `hcl:"workers_count"`
	PollDuration  string `hcl:"poll_duration"`
	RetryAttempts int    `hcl:"retry_attempts"`
	History       struct {
		RetentionPeriod string `hcl:"retention_period"`
	} `hcl:"history,block"`
}
