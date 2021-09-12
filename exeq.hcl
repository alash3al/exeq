http_server {
    listen = ":1215"
}

queue {
    dsn = "redis://localhost:6379/1"

    workers_count = 5

    poll_duration = "50ms"

    retry_attempts = 3
}

macro "crawl" {
    command = "scrapy crawl {{.Args.spider}} {{range $k, $v := .Args}} -a {{$k}}={{$v}} {{end}}"
    mount "file.php" {
        content = ""
    }
}