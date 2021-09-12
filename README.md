EXEQ (WIP)
===========
> Execute shell commands in queues via cli or http interface using redis as a broker.

Why
===
> I'm utilizing background jobs heavily in my projects especially crawling projects, let's say we have a scrapy project and we want to make its spiders run on a distributed queue workers without any complex setup, sod all what I need is to call `exeq` and pass scrapy commands to it either from a `cli` or its `http` api.

TODOs
========
[x]- Redis based task queue
[x]- No dependancies required
[x]- Macros (shortcuts for shell commands)
[x]- While using a macro, you can do something like k8s configmap which enables you to mount a text content as a specific filename, this is helpful in containerized environment where you need to keep only single file to be mapped instead of dealing with multiple files.
[x]- Initutive command line interface 
    [x]- `queue:work`
    [x]- `enqueue:cmd`
    [x]- `enqueue:macro -m MacroName -a ArgKey1=ArgVal1 ...`
    [ ]- `queue:stats` as json output
[ ]- Initutive http interface just like what cli provides.
