EXEQ (WIP)
===========
> Execute shell commands in queues via cli or http interface with modular queue engine.

Why
===
> I'm utilizing background jobs heavily in my projects especially crawling projects, let's say we have a scrapy project and we want to make its spiders run on a distributed queue workers without any complex setup, so all what I need is to call `exeq` and pass scrapy commands to it either from a `cli` or its `http` api.

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

Examples
========