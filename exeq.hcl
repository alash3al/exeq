log_level = "debug"

http_server {
    listen = ":1215"
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
    mount "file.php" {
        content = ""
    }
}