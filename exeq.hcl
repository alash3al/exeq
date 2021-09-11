// Done
http_server {
    listen = ":1215"
}

// Done
queue {
    dsn = "redis://localhost:6379/1"

    workers_count = 5

    poll_duration = "5s"

    retry_attempts = 3
}