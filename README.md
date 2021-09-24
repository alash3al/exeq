EXEQ
======
> **DOCS STILL IN PROGRESS**
> Execute shell commands in queues via cli or http interface with modular queue engine.

Features
========
- Simple initutive tiny cli app.
- Modular queue backends (currently it supports redis as backend but we should support more in the future like `sqs`, `kafka`, `postgres`, ...).
- Powerful configurations thanks to HCL by hashicrop.
- OpenMetrics `/metrics` endpoint to inspect the queue via promethues.
- Error reporting via sentry and stdout/stderr logging.
- Easily create shortcuts for repeated shell commands (we name it Macros).
- Limit the job execution time.

Components
==========
- the configuration file which uses hcl as a configuration language.
- the binary itself which you can get from the releases page.

#### Config
```hcl:exeq.example.hcl

```

#### Exeq Binary
> exeq is an initutive cli app, you can just write `exeq help` from your shell and go with its help.
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
