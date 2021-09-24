logging {
    // Available level names are: "disable" "fatal" "error" "warn" "info" "debug"
    log_level = "debug"

    // Sentry dsn
    sentry_dsn = ""
}

http {
    listen = ":1215"

    access_logs = true
}

queue {
    driver = "rmq"

    dsn = "redis://localhost:6379/1"

    workers_count = 5

    poll_duration = "1s"

    retry_attempts = 3

    history {
        retention_period = "48h"
    }
}

macro "crawl" {
    command = "scrapy crawl {{.Args.spider}} {{range $k, $v := .Args}} -a {{$k}}={{$v}} {{end}}"
    max_execution_time = ""
    mount "file.php" {
        content = ""
    }
}
