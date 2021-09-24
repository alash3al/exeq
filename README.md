EXEQ
======
> **DOCS STILL IN PROGRESS**.  

> Execute shell commands in queues via cli or http interface with modular queue engine.

Features
========
- Simple intuitive tiny cli app.
- Modular queue backends (currently it supports redis as backend but we should support more in the future like `sqs`, `kafka`, `postgres`, ...).
- Powerful configurations thanks to HCL by hashicrop.
- OpenMetrics `/metrics` endpoint to inspect the queue via promethues.
- Error reporting via sentry and stdout/stderr logging.
- Easily create shortcuts for repeated shell commands (we name it Macros).
- Limit the job execution time.

Components
==========
- the configuration file which uses hcl as a configuration language.
- the binary itself which you can get from the [releases page](https://github.com/alash3al/exeq/releases).

#### Config
```hcl
// here we define the logging related configs.
// kindly note that this config file is preprocessed at first with environment expander
// this means that you can easily use any ENV var in any value here i.e: ${HOME}
logging {
    // available level names are: "disable" "fatal" "error" "warn" "info" "debug".
    // NOTE: the value is case sensitive.
    log_level = "debug"

    // sentry dsn (for error reporting).
    // see https://sentry.io/
    // you can do this: sentry_dsn = ${SENTRY_DSN}, but you have to export that env
    // before running exeq.
    sentry_dsn = ""
}

// http server related configs.
http {
    // the address to start listening on.
    listen = ":1215"

    // whether to enable/disable access logs.
    access_logs = true
}

// queue related configs.
queue {
    // the driver used as queue backend.
    // the available drivers are: [rmq].
    driver = "rmq"

    // the data source name or the connection string for the selected driver.
    dsn = "redis://localhost:6379/1"

    // queue workers count.
    workers_count = 5

    // the duration between each poll from the queue
    // a duration string is a possibly signed sequence of decimal numbers, each with optional fraction and a unit suffix, 
    // such as "300ms", "-1.5h" or "2h45m". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
    poll_duration = "1s"

    // how many times a should we retry a failed job?
    retry_attempts = 3

    // each driver needs to keep jobs history, but you may not have to keep it for ever
    // so this history block helps you to configure the history feature.
    history {
        // for how long should we keep our history?
        // a duration string is a possibly signed sequence of decimal numbers, each with optional fraction and a unit suffix, 
        // such as "300ms", "-1.5h" or "2h45m". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
        retention_period = "48h"
    }
}

// here we define a macro which is considered an alias for a command.
// the macro name is "example1".
macro "example1" {
    // the command you want to execute when you call this macro.
    // you can pass arguments to the command and start using it in the command
    // you have to take a look at this first: https://pkg.go.dev/text/template
    // the main variable you can access is `{{.Args}}` which is a map of key=>value.
    // here we cat the data from hello.txt which we mounted before, then echo the value from
    // the arguments map exists in a key called message (just like $args['message']).
    command = "cat ./hello.txt && echo {{.Args.message}}"

    // after how long time should we kill the job?
    // a duration string is a possibly signed sequence of decimal numbers, each with optional fraction and a unit suffix, 
    // such as "300ms", "-1.5h" or "2h45m". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
    // empty means "no time limit".
    max_execution_time = ""

    // sometimes when you run this service within a cloud engine like k8s or docker container
    // you will have to define multiple configmaps/volumes mounts,
    // to simplify this job, we can only mount exeq configurations file and add any other file required by the job
    // into the following mounts block.
    mount "./hello.txt" {
        content = "hello world"
    }
}
```

#### Exeq Binary
> exeq is an intuitive cli app, you can just write `exeq help` from your shell and go with its help.
> the binary consists of the following subcommands
```shell
   queue:work     start the queue worker(s)
   enqueue:macro  submit macro to the queue
   enqueue:cmd    submit a raw shell command to the queue
   queue:jobs     list the jobs
   queue:stats    show queue stats
   serve:http     start the http server which enables you to enqueue macros and inspect the queue via /metrics endpoint with the help of promethues
   help, h        Shows a list of commands or help for one command
```

#### Steps
- Run a queue daemon via `exeq queue:work`
- Optional run the http server as per your needs.
- Start submitting commands either via:
    - HTTP API `POST /enqueue/{MACRO_NAME}`, this endpoint accepts a json message which will be passed to the underlying shell command as args,
    - CLI
        - `exeq enqueue:cmd echo hello world`
        - `exeq enqueue:macro MACRO_NAME -a k=v -a k2=v2`
- You may want to see the queue stats via:
    - CLI:
        - `exeq queue:stats`
        - `watch exeq queue:stats`
    - HTTP API:
        - `GET /`
        - `GET /metrics` (a promethues metrics endpoint)
- You may want to list jobs history `exeq queue:jobs`
