package config

type QueueConfig struct {
	DSN           string `hcl:"dsn"`
	WorkersCount  int    `hcl:"workers_count"`
	PollDuration  string `hcl:"poll_duration"`
	RetryAttempts int    `hcl:"retry_attempts"`
}
